package web

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"university_agency/internal/dashboard"

	"github.com/go-chi/chi/v5"
)

const webSocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type ChatHub struct {
	mu      sync.RWMutex
	clients map[string]map[*wsClient]struct{}
}

type wsClient struct {
	conversationID string
	conn           net.Conn
	reader         *bufio.Reader
	mu             sync.Mutex
}

type incomingChatMessage struct {
	Body string `json:"body"`
}

type outgoingChatPayload struct {
	Type    string      `json:"type"`
	Message chatMessage `json:"message"`
}

type chatMessage struct {
	ID         string `json:"id"`
	SenderID   string `json:"senderId"`
	SenderName string `json:"senderName"`
	SenderRole string `json:"senderRole"`
	Body       string `json:"body"`
	CreatedAt  string `json:"createdAt"`
}

func NewChatHub() *ChatHub {
	return &ChatHub{clients: make(map[string]map[*wsClient]struct{})}
}

func (h *ChatHub) Add(client *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[client.conversationID] == nil {
		h.clients[client.conversationID] = make(map[*wsClient]struct{})
	}
	h.clients[client.conversationID][client] = struct{}{}
}

func (h *ChatHub) Remove(client *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients := h.clients[client.conversationID]; clients != nil {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.clients, client.conversationID)
		}
	}
}

func (h *ChatHub) Broadcast(conversationID string, payload outgoingChatPayload) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	h.mu.RLock()
	clients := make([]*wsClient, 0, len(h.clients[conversationID]))
	for client := range h.clients[conversationID] {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		if err := client.writeText(data); err != nil {
			h.Remove(client)
			_ = client.conn.Close()
		}
	}
}

func (h Handler) chatWebSocket(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		http.Error(w, "invalid origin", http.StatusForbidden)
		return
	}

	viewer := currentUser(r)
	conversationID := chi.URLParam(r, "conversationID")
	if conversationID == "" {
		http.NotFound(w, r)
		return
	}
	if _, err := h.store.ListMessages(r.Context(), viewer, conversationID); err != nil {
		http.Error(w, "akses chat ditolak", http.StatusForbidden)
		return
	}

	client, err := upgradeWebSocket(w, r, conversationID)
	if err != nil {
		h.logger.Warn("websocket upgrade failed", "error", err)
		return
	}
	h.chatHub.Add(client)
	defer func() {
		h.chatHub.Remove(client)
		_ = client.conn.Close()
	}()

	for {
		opcode, data, err := client.readFrame()
		if err != nil {
			return
		}
		switch opcode {
		case 0x1:
			var incoming incomingChatMessage
			if json.Unmarshal(data, &incoming) != nil {
				continue
			}
			message, err := h.store.SaveMessage(context.Background(), viewer, conversationID, incoming.Body)
			if err != nil {
				continue
			}
			h.chatHub.Broadcast(conversationID, messagePayload(message))
		case 0x8:
			_ = client.writeClose()
			return
		case 0x9:
			_ = client.writeFrame(0xA, data)
		}
	}
}

func upgradeWebSocket(w http.ResponseWriter, r *http.Request, conversationID string) (*wsClient, error) {
	if !strings.EqualFold(r.Header.Get("Connection"), "upgrade") && !strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade") {
		http.Error(w, "upgrade required", http.StatusBadRequest)
		return nil, errors.New("missing connection upgrade")
	}
	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		http.Error(w, "upgrade required", http.StatusBadRequest)
		return nil, errors.New("missing websocket upgrade")
	}
	key := strings.TrimSpace(r.Header.Get("Sec-WebSocket-Key"))
	if key == "" {
		http.Error(w, "bad websocket key", http.StatusBadRequest)
		return nil, errors.New("missing websocket key")
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "websocket unsupported", http.StatusInternalServerError)
		return nil, errors.New("hijacker unsupported")
	}
	conn, rw, err := hijacker.Hijack()
	if err != nil {
		return nil, err
	}

	accept := websocketAccept(key)
	_, _ = rw.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
	_, _ = rw.WriteString("Upgrade: websocket\r\n")
	_, _ = rw.WriteString("Connection: Upgrade\r\n")
	_, _ = rw.WriteString("Sec-WebSocket-Accept: " + accept + "\r\n")
	_, _ = rw.WriteString("\r\n")
	if err := rw.Flush(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return &wsClient{conversationID: conversationID, conn: conn, reader: rw.Reader}, nil
}

func websocketAccept(key string) string {
	sum := sha1.Sum([]byte(key + webSocketGUID))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func (c *wsClient) readFrame() (byte, []byte, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(c.reader, header); err != nil {
		return 0, nil, err
	}
	opcode := header[0] & 0x0F
	masked := header[1]&0x80 != 0
	length := uint64(header[1] & 0x7F)
	switch length {
	case 126:
		ext := make([]byte, 2)
		if _, err := io.ReadFull(c.reader, ext); err != nil {
			return 0, nil, err
		}
		length = uint64(binary.BigEndian.Uint16(ext))
	case 127:
		ext := make([]byte, 8)
		if _, err := io.ReadFull(c.reader, ext); err != nil {
			return 0, nil, err
		}
		length = binary.BigEndian.Uint64(ext)
	}
	if length > 4096 {
		return 0, nil, errors.New("websocket payload too large")
	}
	mask := make([]byte, 4)
	if masked {
		if _, err := io.ReadFull(c.reader, mask); err != nil {
			return 0, nil, err
		}
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(c.reader, payload); err != nil {
		return 0, nil, err
	}
	if masked {
		for i := range payload {
			payload[i] ^= mask[i%4]
		}
	}
	return opcode, payload, nil
}

func (c *wsClient) writeText(payload []byte) error {
	return c.writeFrame(0x1, payload)
}

func (c *wsClient) writeClose() error {
	return c.writeFrame(0x8, nil)
}

func (c *wsClient) writeFrame(opcode byte, payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_ = c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	frame := []byte{0x80 | opcode}
	length := len(payload)
	switch {
	case length < 126:
		frame = append(frame, byte(length))
	case length <= 65535:
		frame = append(frame, 126, byte(length>>8), byte(length))
	default:
		frame = append(frame, 127, 0, 0, 0, 0, byte(length>>24), byte(length>>16), byte(length>>8), byte(length))
	}
	frame = append(frame, payload...)
	_, err := c.conn.Write(frame)
	return err
}

func messagePayload(message dashboard.ChatMessage) outgoingChatPayload {
	return outgoingChatPayload{
		Type: "message",
		Message: chatMessage{
			ID:         message.ID,
			SenderID:   message.SenderID,
			SenderName: message.SenderName,
			SenderRole: message.SenderRole.String(),
			Body:       message.Body,
			CreatedAt:  message.CreatedAt.Format(time.RFC3339),
		},
	}
}
