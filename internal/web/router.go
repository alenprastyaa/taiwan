package web

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"university_agency/internal/auth"
	"university_agency/internal/dashboard"
	"university_agency/internal/security"
	"university_agency/internal/ui"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

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
		r.Post("/owner/orders/record-payment", h.recordOrderPayment)
		r.Post("/student/payments/proof", h.submitPaymentProof)
		r.Get("/student/payments/{orderCode}/invoice.pdf", h.studentInvoicePDF)
		r.Get("/owner/invoices/{orderCode}/invoice.pdf", h.studentInvoicePDF)
		r.Post("/student/documents/upload", h.uploadStudentDocument)
		r.Get("/student/documents/{documentID}/download", h.downloadStudentDocument)
		r.Post("/staff/documents/{documentID}/approve", h.approveDocument)
		r.Post("/staff/documents/{documentID}/reject", h.rejectDocument)
		r.Post("/staff/documents/bulk-approve", h.bulkApproveDocuments)
		r.Post("/staff/tasks/{taskID}/complete", h.completeTask)
		r.Post("/staff/tasks/create", h.createTask)
		r.Post("/staff/expenses/create", h.createExpense)
		r.Get("/staff/expenses/{expenseID}/receipt", h.downloadExpenseReceipt)
		r.Post("/staff/calendar/create", h.createSchedule)
		r.Get("/owner/clients/export", h.exportClients)
		r.Get("/staff/clients/export", h.exportClients)
		r.Post("/owner/clients/create", h.createClient)
		r.Post("/staff/clients/create", h.createClient)
		r.Get("/owner/finance/export", h.exportFinance)
		r.Get("/owner/reports/export", h.exportFinance)
		r.Post("/pipeline/stages/create", h.createPipelineStage)
		r.Post("/pipeline/stages/{stageID}/rename", h.renamePipelineStage)
		r.Post("/pipeline/stages/{stageID}/delete", h.deletePipelineStage)
		r.Post("/pipeline/stages/{stageID}/move", h.movePipelineStage)
		r.Post("/pipeline/clients/{clientID}/stage", h.updateClientStage)
		r.Post("/owner/services/create", h.createServicePackage)
		r.Post("/owner/services/{packageID}/update", h.updateServicePackage)
		r.Post("/owner/services/{packageID}/delete", h.deleteServicePackage)
		r.Post("/owner/services/{packageID}/move", h.moveServicePackage)
		r.Post("/templates/create", h.createTextTemplate)
		r.Post("/templates/{templateID}/update", h.updateTextTemplate)
		r.Post("/templates/{templateID}/delete", h.deleteTextTemplate)
		r.Post("/templates/{templateID}/move", h.moveTextTemplate)
		r.Post("/institutions/create", h.createInstitutionContact)
		r.Post("/institutions/{contactID}/update", h.updateInstitutionContact)
		r.Post("/institutions/{contactID}/delete", h.deleteInstitutionContact)
		r.Post("/institutions/{contactID}/move", h.moveInstitutionContact)
		r.Post("/intake/save", h.saveClientIntakeForm)
		r.Get("/owner/intake/export", h.exportClientIntakeForms)
		r.Get("/staff/intake/export", h.exportClientIntakeForms)
		r.Post("/activity/notes/create", h.createActivityNote)
		r.Post("/student/agreement/sign", h.signAgreement)
		r.Get("/student/agreement/pdf", h.downloadOwnAgreementPDF)
		r.Get("/owner/clients/{clientID}/agreement.pdf", h.downloadClientAgreementPDF)
		r.Get("/staff/clients/{clientID}/agreement.pdf", h.downloadClientAgreementPDF)
		r.Post("/owner/clients/{clientID}/reset-password", h.resetClientPassword)
		r.Post("/staff/clients/{clientID}/reset-password", h.resetClientPassword)
		r.Post("/owner/clients/{clientID}/update", h.updateClient)
		r.Post("/staff/clients/{clientID}/update", h.updateClient)
		r.Post("/owner/clients/{clientID}/toggle-active", h.toggleClientActive)
		r.Post("/owner/clients/{clientID}/delete", h.deleteClient)
		r.Post("/owner/invoices/create", h.createOrder)
		r.Post("/owner/invoices/{orderID}/update", h.updateOrder)
		r.Post("/owner/invoices/{orderID}/delete", h.deleteOrder)
		r.Post("/logistics/create", h.createShipment)
		r.Post("/logistics/{shipmentID}/received", h.markShipmentReceived)
		r.Post("/owner/settings/reset-data", h.resetTestData)
		r.Post("/owner/staff/create", h.createStaff)
		r.Post("/owner/staff/{staffID}/update", h.updateStaff)
		r.Post("/owner/staff/{staffID}/toggle-active", h.toggleStaffActive)
		r.Post("/owner/staff/{staffID}/reset-password", h.resetStaffPassword)
		r.Post("/profile/update", h.updateProfile)
		r.Post("/profile/change-password", h.changeOwnPassword)
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

	// Agreement gate: keyed off the viewer's own role, not the URL role param,
	// so an owner/staff impersonate-viewing /student/... (allowed via
	// CanViewRole) is never redirected — only an actual logged-in student
	// hitting their own pages is gated. The section exemption below prevents
	// a redirect loop on the gate page itself.
	viewer := currentUser(r)
	if viewer.Role == dashboard.RoleStudent && section != dashboard.SectionAgreement {
		signed, err := h.store.HasSignedAgreement(r.Context(), viewer)
		if err == nil && !signed {
			http.Redirect(w, r, "/student/agreement", http.StatusFound)
			return
		}
	}

	vm, err := h.dashboard.View(r.Context(), h.cfg.App.Name, h.cfg.App.URL, currentUser(r), role, section, dashboard.ViewOptions{
		ConversationID:   r.URL.Query().Get("conversation"),
		Flash:            r.URL.Query().Get("notice"),
		OrderCode:        r.URL.Query().Get("order"),
		InvoiceFilter:    r.URL.Query().Get("filter"),
		ClientStatus:     r.URL.Query().Get("status"),
		ClientPackage:    r.URL.Query().Get("package"),
		ClientPIC:        r.URL.Query().Get("pic"),
		ClientSearch:     r.URL.Query().Get("q"),
		ShowCreateForm:   r.URL.Query().Get("new") == "1",
		ShowStageManager: r.URL.Query().Get("manage") == "1",
		FilterDate:       r.URL.Query().Get("date"),
		EditClientID:     r.URL.Query().Get("edit"),
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
	})
}

func (h Handler) login(w http.ResponseWriter, r *http.Request) {
	loginError := func(status int, message string) {
		h.renderAuthStatus(w, r, ui.AuthPage{
			AppName:  h.cfg.App.Name,
			AppURL:   h.cfg.App.URL,
			Mode:     "login",
			Title:    "Login",
			Subtitle: "Masuk dengan username dan password.",
			Error:    message,
		}, status)
	}

	if err := r.ParseForm(); err != nil {
		loginError(http.StatusBadRequest, "Form tidak valid.")
		return
	}
	user, err := h.store.FindUserByUsername(r.Context(), r.FormValue("username"))
	if err != nil || !security.CheckPassword(user.PasswordHash, r.FormValue("password")) {
		loginError(http.StatusUnauthorized, "Username atau password salah.")
		return
	}
	session, err := h.sessions.Create(user)
	if err != nil {
		h.logger.Error("create session", "error", err)
		loginError(http.StatusInternalServerError, "Gagal membuat session.")
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
	})
}

func (h Handler) register(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManageStudents(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	registerError := func(status int, message string) {
		h.renderAuthStatus(w, r, ui.AuthPage{
			AppName:  h.cfg.App.Name,
			AppURL:   h.cfg.App.URL,
			Mode:     "register",
			Title:    "Register Client",
			Subtitle: "Buat akun client/student baru.",
			Error:    message,
		}, status)
	}

	if err := r.ParseForm(); err != nil {
		registerError(http.StatusBadRequest, "Form tidak valid.")
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
		registerError(http.StatusUnprocessableEntity, "Registrasi gagal. Username harus unik dan password minimal 8 karakter.")
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

func (h Handler) recordOrderPayment(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/owner/invoices?notice="+url.QueryEscape("Kode pesanan tidak valid."), http.StatusSeeOther)
		return
	}
	amount, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("amount")), 10, 64)
	order, err := h.store.RecordOrderPayment(r.Context(), currentUser(r), r.FormValue("order_code"), amount)
	if err != nil {
		notice := "Gagal mencatat pembayaran. Pastikan nominal lebih dari 0."
		if errors.Is(err, dashboard.ErrForbidden) || errors.Is(err, dashboard.ErrNotFound) {
			notice = "Order tidak ditemukan atau akses ditolak."
		}
		http.Redirect(w, r, "/owner/invoices?notice="+url.QueryEscape(notice), http.StatusSeeOther)
		return
	}
	notice := "Pembayaran untuk " + order.Code + " berhasil dicatat."
	if order.Status == dashboard.OrderPaid {
		notice = "Pembayaran untuk " + order.Code + " berhasil dicatat, order sudah lunas."
	}
	http.Redirect(w, r, "/owner/invoices?order="+url.QueryEscape(order.Code)+"&notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) submitPaymentProof(w http.ResponseWriter, r *http.Request) {
	if !h.requireSignedAgreement(w, r) {
		return
	}
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
	if !h.requireSignedAgreement(w, r) {
		return
	}
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

func (h Handler) downloadExpenseReceipt(w http.ResponseWriter, r *http.Request) {
	expenseID := chi.URLParam(r, "expenseID")
	viewer := currentUser(r)
	expenses, err := h.store.ListExpenses(r.Context(), viewer, viewer.Role)
	if err != nil {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	for _, expense := range expenses {
		if expense.ID == expenseID && expense.ReceiptStoragePath != "" {
			w.Header().Set("Content-Disposition", "attachment; filename="+strconvQuote(expense.ReceiptFileName))
			http.ServeFile(w, r, expense.ReceiptStoragePath)
			return
		}
	}
	http.NotFound(w, r)
}

func (h Handler) studentInvoicePDF(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "orderCode")
	viewer := currentUser(r)
	orders, err := h.store.ListOrders(r.Context(), viewer, viewer.Role)
	if err != nil {
		http.Error(w, "akses invoice ditolak", http.StatusForbidden)
		return
	}
	for _, order := range orders {
		if strings.EqualFold(order.Code, code) {
			pdf := renderInvoicePDF(h.cfg.App.Name, order)
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", "attachment; filename="+strconvQuote(order.Code+".pdf"))
			_, _ = w.Write(pdf)
			return
		}
	}
	http.NotFound(w, r)
}

func (h Handler) approveDocument(w http.ResponseWriter, r *http.Request) {
	h.reviewDocument(w, r, dashboard.DocumentApproved, "", "Dokumen disetujui.")
}

func (h Handler) rejectDocument(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	reason := r.FormValue("reason")
	h.reviewDocument(w, r, dashboard.DocumentRevision, reason, "Dokumen ditolak dan dikembalikan untuk revisi.")
}

func (h Handler) reviewDocument(w http.ResponseWriter, r *http.Request, status dashboard.DocumentStatus, note, notice string) {
	_, err := h.store.ReviewDocument(r.Context(), currentUser(r), chi.URLParam(r, "documentID"), status, note)
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

func (h Handler) bulkApproveDocuments(w http.ResponseWriter, r *http.Request) {
	count, err := h.store.BulkApproveDocuments(r.Context(), currentUser(r))
	notice := fmt.Sprintf("%d dokumen disetujui.", count)
	if err != nil {
		notice = "Aksi review massal gagal atau akses ditolak."
	}
	http.Redirect(w, r, "/staff/documents?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

// pipelinePath is only meaningful for owner/staff — students have no pipeline
// page to redirect back to, so callers must reject the request before ever
// building a URL from it.
func pipelinePath(viewer dashboard.User) string {
	return "/" + viewer.Role.String() + "/pipeline"
}

func canManagePipeline(viewer dashboard.User) bool {
	return viewer.Role == dashboard.RoleOwner || viewer.Role == dashboard.RoleStaff
}

func (h Handler) createPipelineStage(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManagePipeline(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, pipelinePath(viewer)+"?manage=1", http.StatusSeeOther)
		return
	}
	_, err := h.store.CreatePipelineStage(r.Context(), viewer, r.FormValue("name"))
	notice := "Tahap pipeline baru berhasil ditambahkan."
	if errors.Is(err, dashboard.ErrDuplicate) {
		notice = "Nama tahap sudah dipakai, pilih nama lain."
	} else if err != nil {
		notice = "Gagal menambahkan tahap. Nama tidak boleh kosong."
	}
	http.Redirect(w, r, pipelinePath(viewer)+"?manage=1&notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) renamePipelineStage(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManagePipeline(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, pipelinePath(viewer)+"?manage=1", http.StatusSeeOther)
		return
	}
	_, err := h.store.RenamePipelineStage(r.Context(), viewer, chi.URLParam(r, "stageID"), r.FormValue("name"))
	notice := "Nama tahap berhasil diperbarui."
	if errors.Is(err, dashboard.ErrDuplicate) {
		notice = "Nama tahap sudah dipakai, pilih nama lain."
	} else if err != nil {
		notice = "Gagal mengubah nama tahap."
	}
	http.Redirect(w, r, pipelinePath(viewer)+"?manage=1&notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) deletePipelineStage(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManagePipeline(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	err := h.store.DeletePipelineStage(r.Context(), viewer, chi.URLParam(r, "stageID"))
	notice := "Tahap pipeline dihapus. Client di tahap ini dipindah ke Belum Ditentukan."
	if err != nil {
		notice = "Gagal menghapus tahap."
	}
	http.Redirect(w, r, pipelinePath(viewer)+"?manage=1&notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) movePipelineStage(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManagePipeline(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, pipelinePath(viewer)+"?manage=1", http.StatusSeeOther)
		return
	}
	err := h.store.ReorderPipelineStage(r.Context(), viewer, chi.URLParam(r, "stageID"), r.FormValue("direction"))
	notice := "?manage=1"
	if err != nil {
		notice = "?manage=1&notice=" + url.QueryEscape("Gagal mengubah urutan tahap.")
	}
	http.Redirect(w, r, pipelinePath(viewer)+notice, http.StatusSeeOther)
}

func (h Handler) updateClientStage(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManagePipeline(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, pipelinePath(viewer), http.StatusSeeOther)
		return
	}
	_, err := h.store.UpdateClientStage(r.Context(), viewer, chi.URLParam(r, "clientID"), r.FormValue("stage_name"))
	notice := "Tahap client berhasil diperbarui."
	if err != nil {
		notice = "Gagal memindahkan client ke tahap ini atau akses ditolak."
	}
	http.Redirect(w, r, pipelinePath(viewer)+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func servicePackageInputFromForm(r *http.Request) dashboard.ServicePackageInput {
	price, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("price")), 10, 64)
	return dashboard.ServicePackageInput{
		Name:        r.FormValue("name"),
		Category:    r.FormValue("category"),
		Description: r.FormValue("description"),
		Price:       price,
		PriceIsFrom: r.FormValue("price_is_from") == "1",
		Highlights:  r.FormValue("highlights"),
	}
}

func (h Handler) createServicePackage(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/owner/services?manage=1", http.StatusSeeOther)
		return
	}
	_, err := h.store.CreateServicePackage(r.Context(), viewer, servicePackageInputFromForm(r))
	notice := "Paket baru berhasil ditambahkan."
	if errors.Is(err, dashboard.ErrDuplicate) {
		notice = "Nama paket sudah dipakai, pilih nama lain."
	} else if err != nil {
		notice = "Gagal menambahkan paket. Nama dan harga wajib diisi."
	}
	http.Redirect(w, r, "/owner/services?manage=1&notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) updateServicePackage(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/owner/services?manage=1", http.StatusSeeOther)
		return
	}
	_, err := h.store.UpdateServicePackage(r.Context(), viewer, chi.URLParam(r, "packageID"), servicePackageInputFromForm(r))
	notice := "Paket berhasil diperbarui."
	if errors.Is(err, dashboard.ErrDuplicate) {
		notice = "Nama paket sudah dipakai, pilih nama lain."
	} else if err != nil {
		notice = "Gagal memperbarui paket."
	}
	http.Redirect(w, r, "/owner/services?manage=1&notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) deleteServicePackage(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	err := h.store.DeleteServicePackage(r.Context(), viewer, chi.URLParam(r, "packageID"))
	notice := "Paket dihapus."
	if err != nil {
		notice = "Gagal menghapus paket."
	}
	http.Redirect(w, r, "/owner/services?manage=1&notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) moveServicePackage(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/owner/services?manage=1", http.StatusSeeOther)
		return
	}
	err := h.store.ReorderServicePackage(r.Context(), viewer, chi.URLParam(r, "packageID"), r.FormValue("direction"))
	notice := "?manage=1"
	if err != nil {
		notice = "?manage=1&notice=" + url.QueryEscape("Gagal mengubah urutan paket.")
	}
	http.Redirect(w, r, "/owner/services"+notice, http.StatusSeeOther)
}

func (h Handler) createTask(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/staff/tasks?notice="+url.QueryEscape("Form tugas tidak valid."), http.StatusSeeOther)
		return
	}
	_, err := h.store.CreateTask(r.Context(), currentUser(r), dashboard.CreateTaskInput{
		ClientID:  r.FormValue("client_id"),
		TimeLabel: r.FormValue("time_label"),
		Title:     r.FormValue("title"),
		Priority:  r.FormValue("priority"),
	})
	notice := "Task baru berhasil ditambahkan."
	if err != nil {
		notice = "Gagal menambahkan task. Pastikan client dan judul tugas terisi."
	}
	http.Redirect(w, r, "/staff/tasks?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) createExpense(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil && err != http.ErrNotMultipart {
		http.Redirect(w, r, "/staff/expenses?notice="+url.QueryEscape("Form pengeluaran tidak valid."), http.StatusSeeOther)
		return
	}
	amount, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("amount")), 10, 64)
	var fileName, storagePath string
	if saved, err := h.saveUploadedFile(r, "receipt_file", "expenses"); err == nil && saved != "" {
		fileName = filepath.Base(saved)
		storagePath = saved
	} else if err != nil && !errors.Is(err, http.ErrMissingFile) {
		http.Redirect(w, r, "/staff/expenses?notice="+url.QueryEscape("File bukti pengeluaran tidak valid."), http.StatusSeeOther)
		return
	}
	_, err := h.store.CreateExpense(r.Context(), currentUser(r), dashboard.CreateExpenseInput{
		ClientID:    r.FormValue("client_id"),
		Need:        r.FormValue("need"),
		Category:    r.FormValue("category"),
		Amount:      amount,
		DateLabel:   r.FormValue("date_label"),
		Description: r.FormValue("description"),
		FileName:    fileName,
		StoragePath: storagePath,
	})
	notice := "Pengeluaran baru berhasil dicatat."
	if err != nil {
		notice = "Gagal mencatat pengeluaran. Pastikan client, keperluan, dan nominal terisi."
	}
	http.Redirect(w, r, "/staff/expenses?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) createSchedule(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/staff/calendar?notice="+url.QueryEscape("Form jadwal tidak valid."), http.StatusSeeOther)
		return
	}
	_, err := h.store.CreateSchedule(r.Context(), currentUser(r), dashboard.CreateScheduleInput{
		ClientID:  r.FormValue("client_id"),
		Title:     r.FormValue("title"),
		DateLabel: r.FormValue("date_label"),
		TimeLabel: r.FormValue("time_label"),
		Location:  r.FormValue("location"),
	})
	notice := "Jadwal baru berhasil ditambahkan."
	if err != nil {
		notice = "Gagal menambahkan jadwal. Pastikan client dan judul jadwal terisi."
	}
	http.Redirect(w, r, "/staff/calendar?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) exportClients(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	clients, err := h.store.ListClients(r.Context(), viewer, viewer.Role)
	if err != nil {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=clients.csv")
	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"Nama", "Email", "Paket", "PIC", "Status", "Progress", "Jadwal Terakhir"})
	for _, client := range clients {
		_ = writer.Write([]string{client.Name, client.Email, client.PackageName, client.PICName, client.Status, strconv.Itoa(client.Progress) + "%", client.LastSchedule})
	}
	writer.Flush()
}

func (h Handler) exportFinance(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	orders, err := h.store.ListOrders(r.Context(), viewer, viewer.Role)
	if err != nil {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=finance.csv")
	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"Kode Order", "Client", "Paket", "Total", "Sudah Dibayar", "Sisa", "Status", "Jatuh Tempo"})
	for _, order := range orders {
		due := ""
		if !order.DueDate.IsZero() {
			due = order.DueDate.Format("02 Jan 2006")
		}
		_ = writer.Write([]string{order.Code, order.ClientName, order.PackageName, formatIDR(order.Total), formatIDR(order.Paid), formatIDR(maxInt64(order.Total-order.Paid, 0)), string(order.Status), due})
	}
	writer.Flush()
}

func (h Handler) postChatMessage(w http.ResponseWriter, r *http.Request) {
	if currentUser(r).Role == dashboard.RoleStudent {
		if !h.requireSignedAgreement(w, r) {
			return
		}
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
	h.renderAuthStatus(w, r, vm, http.StatusOK)
}

func (h Handler) renderAuthStatus(w http.ResponseWriter, r *http.Request, vm ui.AuthPage, status int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
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

func formatIDR(value int64) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	digits := strconv.FormatInt(value, 10)
	var parts []string
	for len(digits) > 3 {
		parts = append([]string{digits[len(digits)-3:]}, parts...)
		digits = digits[:len(digits)-3]
	}
	parts = append([]string{digits}, parts...)
	return sign + "Rp " + strings.Join(parts, ".")
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
