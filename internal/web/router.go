package web

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"university_agency/internal/auth"
	"university_agency/internal/config"
	"university_agency/internal/dashboard"
	"university_agency/internal/security"
	"university_agency/internal/ui"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Dependencies struct {
	Config    config.Config
	Logger    *slog.Logger
	Dashboard dashboard.Service
	Store     Store
	Sessions  *auth.SessionStore
}

type Store interface {
	dashboard.Repository
	FindUserByUsername(ctx context.Context, username string) (dashboard.User, error)
	FindUserByID(ctx context.Context, id string) (dashboard.User, error)
	CreateStudent(ctx context.Context, input dashboard.CreateStudentInput) (dashboard.User, error)
	MarkOrderPaidByCode(ctx context.Context, viewer dashboard.User, code string) (dashboard.Order, error)
	SubmitPaymentProof(ctx context.Context, viewer dashboard.User, code, note, fileName, storagePath string) (dashboard.Order, error)
	StudentHasPaymentAccess(ctx context.Context, viewer dashboard.User) (bool, error)
	UploadStudentDocument(ctx context.Context, viewer dashboard.User, documentName, fileName, storagePath string) (dashboard.Document, error)
	ReviewDocument(ctx context.Context, viewer dashboard.User, documentID string, status dashboard.DocumentStatus) (dashboard.Document, error)
	CompleteTask(ctx context.Context, viewer dashboard.User, taskID string) (dashboard.Task, error)
	SaveMessage(ctx context.Context, viewer dashboard.User, conversationID, body string) (dashboard.ChatMessage, error)
}

type Handler struct {
	cfg       config.Config
	logger    *slog.Logger
	dashboard dashboard.Service
	store     Store
	sessions  *auth.SessionStore
	chatHub   *ChatHub
}

func NewRouter(deps Dependencies) http.Handler {
	h := Handler{
		cfg:       deps.Config,
		logger:    deps.Logger,
		dashboard: deps.Dashboard,
		store:     deps.Store,
		sessions:  deps.Sessions,
		chatHub:   NewChatHub(),
	}
	if h.sessions == nil {
		h.sessions = auth.NewSessionStore(12 * time.Hour)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(h.securityHeaders)
	r.Use(h.originGuard)
	r.Use(h.loadSession)

	fileServer := http.FileServer(http.FS(staticFiles()))
	r.Handle("/assets/*", fileServer)
	r.Handle("/icons/*", fileServer)
	r.Handle("/favicon.svg", fileServer)
	r.Handle("/manifest.webmanifest", fileServer)
	r.Handle("/sw.js", fileServer)

	r.Get("/login", h.loginPage)
	r.Post("/login", h.login)

	r.Group(func(r chi.Router) {
		r.Use(h.requireAuth)
		r.Get("/", h.home)
		r.Get("/register", h.registerPage)
		r.Post("/register", h.register)
		r.Post("/logout", h.logout)
		r.Post("/owner/orders/mark-paid", h.markOrderPaid)
		r.Post("/student/payments/proof", h.submitPaymentProof)
		r.Get("/student/payments/{orderCode}/invoice.pdf", h.studentInvoicePDF)
		r.Post("/student/documents/upload", h.uploadStudentDocument)
		r.Get("/student/documents/{documentID}/download", h.downloadStudentDocument)
		r.Post("/staff/documents/{documentID}/approve", h.approveDocument)
		r.Post("/staff/documents/{documentID}/reject", h.rejectDocument)
		r.Post("/staff/tasks/{taskID}/complete", h.completeTask)
		r.Post("/chat/{conversationID}/messages", h.postChatMessage)
		r.Get("/ws/chat/{conversationID}", h.chatWebSocket)
		r.Get("/client", h.clientAlias)
		r.Get("/client/{section}", h.clientAlias)
		r.Get("/{role}", h.page)
		r.Get("/{role}/{section}", h.page)
	})

	return r
}

func (h Handler) home(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	http.Redirect(w, r, defaultPath(viewer.Role), http.StatusFound)
}

func (h Handler) clientAlias(w http.ResponseWriter, r *http.Request) {
	section := chi.URLParam(r, "section")
	target := "/student"
	if section != "" {
		target += "/" + section
	}
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func (h Handler) page(w http.ResponseWriter, r *http.Request) {
	role, ok := dashboard.ParseRole(strings.ToLower(chi.URLParam(r, "role")))
	if !ok {
		http.NotFound(w, r)
		return
	}

	section := dashboard.ParseSection(strings.ToLower(chi.URLParam(r, "section")))
	vm, err := h.dashboard.View(r.Context(), h.cfg.App.Name, h.cfg.App.URL, currentUser(r), role, section, dashboard.ViewOptions{
		ConversationID: r.URL.Query().Get("conversation"),
		Flash:          r.URL.Query().Get("notice"),
	})
	if err != nil {
		if errors.Is(err, dashboard.ErrForbidden) {
			http.Error(w, "akses ditolak", http.StatusForbidden)
			return
		}
		if errors.Is(err, dashboard.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("render dashboard", "error", err)
		http.Error(w, "failed to render dashboard", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if isHTMX(r) {
		if err := ui.App(vm).Render(r.Context(), w); err != nil {
			h.logger.Error("render app partial", "error", err)
		}
		return
	}

	if err := ui.Document(vm).Render(r.Context(), w); err != nil {
		h.logger.Error("render document", "error", err)
	}
}

func (h Handler) loginPage(w http.ResponseWriter, r *http.Request) {
	if user, ok := userFromContext(r.Context()); ok {
		http.Redirect(w, r, defaultPath(user.Role), http.StatusFound)
		return
	}
	h.renderAuth(w, r, ui.AuthPage{
		AppName:  h.cfg.App.Name,
		AppURL:   h.cfg.App.URL,
		Mode:     "login",
		Title:    "Login",
		Subtitle: "Masuk dengan username dan password.",
		Notice:   r.URL.Query().Get("notice"),
		Error:    r.URL.Query().Get("error"),
	})
}

func (h Handler) login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/login?error="+url.QueryEscape("Form tidak valid."), http.StatusSeeOther)
		return
	}
	user, err := h.store.FindUserByUsername(r.Context(), r.FormValue("username"))
	if err != nil || !security.CheckPassword(user.PasswordHash, r.FormValue("password")) {
		http.Redirect(w, r, "/login?error="+url.QueryEscape("Username atau password salah."), http.StatusSeeOther)
		return
	}
	session, err := h.sessions.Create(user)
	if err != nil {
		h.logger.Error("create session", "error", err)
		http.Redirect(w, r, "/login?error="+url.QueryEscape("Gagal membuat session."), http.StatusSeeOther)
		return
	}
	h.setSessionCookie(w, r, session)
	http.Redirect(w, r, defaultPath(user.Role), http.StatusSeeOther)
}

func (h Handler) registerPage(w http.ResponseWriter, r *http.Request) {
	if !canManageStudents(currentUser(r)) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	h.renderAuth(w, r, ui.AuthPage{
		AppName:  h.cfg.App.Name,
		AppURL:   h.cfg.App.URL,
		Mode:     "register",
		Title:    "Register Client",
		Subtitle: "Buat akun client/student baru.",
		Error:    r.URL.Query().Get("error"),
	})
}

func (h Handler) register(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManageStudents(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/register?error="+url.QueryEscape("Form tidak valid."), http.StatusSeeOther)
		return
	}
	user, err := h.store.CreateStudent(r.Context(), dashboard.CreateStudentInput{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Phone:    r.FormValue("phone"),
	})
	if err != nil {
		message := "Registrasi gagal. Username harus unik dan password minimal 8 karakter."
		http.Redirect(w, r, "/register?error="+url.QueryEscape(message), http.StatusSeeOther)
		return
	}
	target := "/" + currentPageRole(viewer.Role).String() + "/clients?notice=" + url.QueryEscape("Akun client "+user.Username+" sudah dibuat.")
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func (h Handler) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		h.sessions.Delete(cookie.Value)
	}
	http.SetCookie(w, expiredSessionCookie())
	http.Redirect(w, r, "/login?notice="+url.QueryEscape("Kamu sudah logout."), http.StatusSeeOther)
}

func (h Handler) markOrderPaid(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/owner/invoices?notice="+url.QueryEscape("Kode pesanan tidak valid."), http.StatusSeeOther)
		return
	}
	order, err := h.store.MarkOrderPaidByCode(r.Context(), currentUser(r), r.FormValue("order_code"))
	if err != nil {
		http.Redirect(w, r, "/owner/invoices?notice="+url.QueryEscape("Order tidak ditemukan atau akses ditolak."), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/owner/invoices?notice="+url.QueryEscape("Order "+order.Code+" sudah ditandai lunas."), http.StatusSeeOther)
}

func (h Handler) submitPaymentProof(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil && err != http.ErrNotMultipart {
		http.Redirect(w, r, "/student/payments?notice="+url.QueryEscape("Bukti bayar tidak valid."), http.StatusSeeOther)
		return
	}
	note := r.FormValue("note")
	var proofFileName, proofStoragePath string
	if saved, err := h.saveUploadedFile(r, "proof_file", "payments"); err == nil && saved != "" {
		proofFileName = filepath.Base(saved)
		proofStoragePath = saved
	} else if err != nil && !errors.Is(err, http.ErrMissingFile) {
		http.Redirect(w, r, "/student/payments?notice="+url.QueryEscape("File bukti bayar tidak valid."), http.StatusSeeOther)
		return
	}
	if strings.TrimSpace(note) == "" && proofStoragePath == "" {
		http.Redirect(w, r, "/student/payments?notice="+url.QueryEscape("Catatan atau file bukti bayar wajib diisi."), http.StatusSeeOther)
		return
	}
	order, err := h.store.SubmitPaymentProof(r.Context(), currentUser(r), r.FormValue("order_code"), note, proofFileName, proofStoragePath)
	if err != nil {
		http.Redirect(w, r, "/student/payments?notice="+url.QueryEscape("Gagal mengirim bukti bayar."), http.StatusSeeOther)
		return
	}
	notice := "Bukti bayar untuk " + order.Code + " menunggu verifikasi."
	if order.Status == dashboard.OrderPaid {
		notice = "Order " + order.Code + " sudah lunas."
	}
	http.Redirect(w, r, "/student/payments?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) uploadStudentDocument(w http.ResponseWriter, r *http.Request) {
	allowed, err := h.store.StudentHasPaymentAccess(r.Context(), currentUser(r))
	if err != nil || !allowed {
		http.Redirect(w, r, "/student/payments?notice="+url.QueryEscape("Upload dokumen dibuka setelah bukti pembayaran dikirim atau invoice lunas."), http.StatusSeeOther)
		return
	}
	if err := r.ParseMultipartForm(15 << 20); err != nil {
		http.Redirect(w, r, "/student/documents?notice="+url.QueryEscape("Upload dokumen tidak valid."), http.StatusSeeOther)
		return
	}
	documentName := r.FormValue("document_name")
	storagePath, err := h.saveUploadedFile(r, "document_file", "documents")
	if err != nil {
		http.Redirect(w, r, "/student/documents?notice="+url.QueryEscape("File dokumen wajib dipilih."), http.StatusSeeOther)
		return
	}
	_, err = h.store.UploadStudentDocument(r.Context(), currentUser(r), documentName, filepath.Base(storagePath), storagePath)
	if err != nil {
		http.Redirect(w, r, "/student/documents?notice="+url.QueryEscape("Upload gagal atau akses ditolak."), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/student/documents?notice="+url.QueryEscape("Dokumen terkirim dan menunggu review."), http.StatusSeeOther)
}

func (h Handler) downloadStudentDocument(w http.ResponseWriter, r *http.Request) {
	documentID := chi.URLParam(r, "documentID")
	documents, err := h.store.ListDocuments(r.Context(), currentUser(r), dashboard.RoleStudent)
	if err != nil {
		http.Error(w, "akses dokumen ditolak", http.StatusForbidden)
		return
	}
	for _, document := range documents {
		if document.ID == documentID && document.StoragePath != "" {
			w.Header().Set("Content-Disposition", "attachment; filename="+strconvQuote(document.FileName))
			http.ServeFile(w, r, document.StoragePath)
			return
		}
	}
	http.NotFound(w, r)
}

func (h Handler) studentInvoicePDF(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "orderCode")
	orders, err := h.store.ListOrders(r.Context(), currentUser(r), dashboard.RoleStudent)
	if err != nil {
		http.Error(w, "akses invoice ditolak", http.StatusForbidden)
		return
	}
	for _, order := range orders {
		if strings.EqualFold(order.Code, code) {
			body := []string{
				"Taiwan Education Consulting",
				"Invoice " + order.Code,
				"Client: " + order.ClientName,
				"Paket: " + order.PackageName,
				"Total: " + formatIDR(order.Total),
				"Sudah Dibayar: " + formatIDR(order.Paid),
				"Sisa: " + formatIDR(maxInt64(order.Total-order.Paid, 0)),
				"Status: " + string(order.Status),
			}
			pdf := simplePDF(body)
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", "attachment; filename="+strconvQuote(order.Code+".pdf"))
			_, _ = w.Write(pdf)
			return
		}
	}
	http.NotFound(w, r)
}

func (h Handler) approveDocument(w http.ResponseWriter, r *http.Request) {
	h.reviewDocument(w, r, dashboard.DocumentApproved, "Dokumen disetujui.")
}

func (h Handler) rejectDocument(w http.ResponseWriter, r *http.Request) {
	h.reviewDocument(w, r, dashboard.DocumentRevision, "Dokumen dikembalikan untuk revisi.")
}

func (h Handler) reviewDocument(w http.ResponseWriter, r *http.Request, status dashboard.DocumentStatus, notice string) {
	_, err := h.store.ReviewDocument(r.Context(), currentUser(r), chi.URLParam(r, "documentID"), status)
	if err != nil {
		notice = "Aksi dokumen gagal atau akses ditolak."
	}
	http.Redirect(w, r, "/staff/documents?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) completeTask(w http.ResponseWriter, r *http.Request) {
	_, err := h.store.CompleteTask(r.Context(), currentUser(r), chi.URLParam(r, "taskID"))
	notice := "Task sudah ditandai selesai."
	if err != nil {
		notice = "Task gagal diperbarui atau akses ditolak."
	}
	http.Redirect(w, r, "/staff/tasks?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) postChatMessage(w http.ResponseWriter, r *http.Request) {
	if currentUser(r).Role == dashboard.RoleStudent {
		allowed, err := h.store.StudentHasPaymentAccess(r.Context(), currentUser(r))
		if err != nil || !allowed {
			http.Redirect(w, r, "/student/payments?notice="+url.QueryEscape("Chat konsultan dibuka setelah bukti pembayaran dikirim atau invoice lunas."), http.StatusSeeOther)
			return
		}
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, defaultPath(currentUser(r).Role), http.StatusSeeOther)
		return
	}
	conversationID := chi.URLParam(r, "conversationID")
	message, err := h.store.SaveMessage(r.Context(), currentUser(r), conversationID, r.FormValue("body"))
	if err == nil {
		h.chatHub.Broadcast(conversationID, messagePayload(message))
	}
	target := "/" + currentPageRole(currentUser(r).Role).String() + "/chat?conversation=" + url.QueryEscape(conversationID)
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func (h Handler) renderAuth(w http.ResponseWriter, r *http.Request, vm ui.AuthPage) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := ui.AuthDocument(vm).Render(r.Context(), w); err != nil {
		h.logger.Error("render auth", "error", err)
	}
}

func (h Handler) saveUploadedFile(r *http.Request, fieldName, category string) (string, error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	cleanName := safeUploadName(header.Filename)
	if cleanName == "" {
		return "", http.ErrMissingFile
	}
	dir := filepath.Join("data", "uploads", category, currentUser(r).ID)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	target := filepath.Join(dir, time.Now().Format("20060102150405")+"-"+cleanName)
	out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, io.LimitReader(file, 15<<20)); err != nil {
		return "", err
	}
	return target, nil
}

func safeUploadName(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '.', r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, name)
	return strings.Trim(name, ".- _")
}

func strconvQuote(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, "") + `"`
}

func simplePDF(lines []string) []byte {
	var content bytes.Buffer
	content.WriteString("BT\n/F1 12 Tf\n72 760 Td\n")
	for i, line := range lines {
		if i > 0 {
			content.WriteString("0 -22 Td\n")
		}
		content.WriteString("(" + pdfText(line) + ") Tj\n")
	}
	content.WriteString("ET\n")

	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", content.Len(), content.String()),
	}
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, len(objects)+1)
	offsets = append(offsets, 0)
	for i, object := range objects {
		offsets = append(offsets, out.Len())
		out.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", i+1, object))
	}
	xref := out.Len()
	out.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", len(objects)+1))
	for i := 1; i < len(offsets); i++ {
		out.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	out.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref))
	return out.Bytes()
}

func pdfText(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `(`, `\(`)
	value = strings.ReplaceAll(value, `)`, `\)`)
	return value
}

func formatIDR(value int64) string {
	return fmt.Sprintf("Rp %d", value)
}

func maxInt64(value, floor int64) int64 {
	if value < floor {
		return floor
	}
	return value
}

func isHTMX(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("HX-Request"), "true")
}

const sessionCookieName = "ua_session"

type userContextKey struct{}

func (h Handler) loadSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		session, ok := h.sessions.Find(cookie.Value)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		user, err := h.store.FindUserByID(r.Context(), session.UserID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h Handler) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := userFromContext(r.Context()); !ok {
			if r.URL.Path == "/register" {
				http.NotFound(w, r)
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func userFromContext(ctx context.Context) (dashboard.User, bool) {
	user, ok := ctx.Value(userContextKey{}).(dashboard.User)
	return user, ok
}

func currentUser(r *http.Request) dashboard.User {
	user, _ := userFromContext(r.Context())
	return user
}

func (h Handler) setSessionCookie(w http.ResponseWriter, r *http.Request, session auth.Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		MaxAge:   int(time.Until(session.ExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   isSecureRequest(r, h.cfg.App.URL),
		SameSite: http.SameSiteLaxMode,
	})
}

func expiredSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func defaultPath(role dashboard.Role) string {
	switch role {
	case dashboard.RoleOwner:
		return "/owner"
	case dashboard.RoleStaff:
		return "/staff"
	case dashboard.RoleStudent:
		return "/student"
	default:
		return "/login"
	}
}

func currentPageRole(role dashboard.Role) dashboard.Role {
	switch role {
	case dashboard.RoleOwner, dashboard.RoleStaff, dashboard.RoleStudent:
		return role
	default:
		return dashboard.RoleStudent
	}
}

func canManageStudents(user dashboard.User) bool {
	return user.Role == dashboard.RoleOwner || user.Role == dashboard.RoleStaff
}

func isSecureRequest(r *http.Request, appURL string) bool {
	if r.TLS != nil {
		return true
	}
	if strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		return true
	}
	return strings.HasPrefix(strings.ToLower(appURL), "https://")
}

func (h Handler) originGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if !sameOrigin(r) {
			http.Error(w, "invalid origin", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	parsed, err := url.Parse(origin)
	return err == nil && strings.EqualFold(parsed.Host, r.Host)
}

func (h Handler) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; base-uri 'self'; object-src 'none'; frame-ancestors 'none'; form-action 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; connect-src 'self'; manifest-src 'self'")
		next.ServeHTTP(w, r)
	})
}
