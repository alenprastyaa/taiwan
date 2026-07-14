package auth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"university_agency/internal/dashboard"
)

type Session struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type SessionStore struct {
	mu       sync.RWMutex
	ttl      time.Duration
	sessions map[string]Session
}

func NewSessionStore(ttl time.Duration) *SessionStore {
	return &SessionStore{
		ttl:      ttl,
		sessions: make(map[string]Session),
	}
}

func (s *SessionStore) Create(user dashboard.User) (Session, error) {
	token, err := randomToken()
	if err != nil {
		return Session{}, err
	}

	now := time.Now()
	session := Session{
		Token:     token,
		UserID:    user.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(s.ttl),
	}

	s.mu.Lock()
	s.sessions[token] = session
	s.mu.Unlock()

	return session, nil
}

func (s *SessionStore) Find(token string) (Session, bool) {
	if token == "" {
		return Session{}, false
	}

	s.mu.RLock()
	session, ok := s.sessions[token]
	s.mu.RUnlock()
	if !ok {
		return Session{}, false
	}
	if time.Now().After(session.ExpiresAt) {
		s.Delete(token)
		return Session{}, false
	}
	return session, true
}

func (s *SessionStore) Delete(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
