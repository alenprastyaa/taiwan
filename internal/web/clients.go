package web

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"university_agency/internal/dashboard"
)

func (h Handler) createClient(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManageStudents(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form tidak valid", http.StatusBadRequest)
		return
	}
	basePath := "/" + currentPageRole(viewer.Role).String() + "/clients"
	input := dashboard.CreateStudentInput{
		Username:    r.FormValue("username"),
		Password:    r.FormValue("password"),
		Name:        r.FormValue("name"),
		Email:       r.FormValue("email"),
		Phone:       r.FormValue("phone"),
		PackageName: r.FormValue("package_name"),
		Country:     r.FormValue("country"),
		Campus:      r.FormValue("campus"),
		PICStaffID:  r.FormValue("pic_staff_id"),
	}
	if viewer.Role == dashboard.RoleOwner {
		input.InvoiceTotal, _ = strconv.ParseInt(strings.TrimSpace(r.FormValue("invoice_total")), 10, 64)
		input.AmountPaid, _ = strconv.ParseInt(strings.TrimSpace(r.FormValue("amount_paid")), 10, 64)
	}
	user, err := h.store.CreateStudent(r.Context(), input)
	if err != nil {
		http.Redirect(w, r, basePath+"?new=1&notice="+url.QueryEscape("Gagal menambah client. Username harus unik dan password minimal 8 karakter."), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, basePath+"?notice="+url.QueryEscape("Client "+user.Username+" berhasil ditambahkan."), http.StatusSeeOther)
}

func (h Handler) updateClient(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManageStudents(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	basePath := "/" + currentPageRole(viewer.Role).String() + "/clients"
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form tidak valid", http.StatusBadRequest)
		return
	}
	clientID := chi.URLParam(r, "clientID")
	_, err := h.store.UpdateClient(r.Context(), viewer, clientID, dashboard.UpdateClientInput{
		Name:        r.FormValue("name"),
		Email:       r.FormValue("email"),
		Phone:       r.FormValue("phone"),
		PackageName: r.FormValue("package_name"),
		Country:     r.FormValue("country"),
		Campus:      r.FormValue("campus"),
		PICStaffID:  r.FormValue("pic_staff_id"),
	})
	notice := "Data client berhasil diperbarui."
	if err != nil {
		notice = "Gagal memperbarui data client. Nama wajib diisi."
		if errors.Is(err, dashboard.ErrForbidden) {
			notice = "Kamu tidak punya akses untuk mengubah client ini."
		}
	}
	http.Redirect(w, r, basePath+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) toggleClientActive(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	basePath := "/" + currentPageRole(viewer.Role).String() + "/clients"
	err := h.store.ToggleClientActive(r.Context(), viewer, chi.URLParam(r, "clientID"))
	notice := "Status client berhasil diperbarui."
	if err != nil {
		notice = "Gagal mengubah status client."
	}
	http.Redirect(w, r, basePath+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) deleteClient(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	basePath := "/" + currentPageRole(viewer.Role).String() + "/clients"
	err := h.store.DeleteClient(r.Context(), viewer, chi.URLParam(r, "clientID"))
	notice := "Client berhasil dihapus permanen."
	if errors.Is(err, dashboard.ErrConflict) {
		notice = "Client tidak bisa dihapus karena masih punya order, dokumen, task, expense, shipment, atau jadwal. Nonaktifkan saja jika tidak dipakai lagi."
	} else if err != nil {
		notice = "Gagal menghapus client."
	}
	http.Redirect(w, r, basePath+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) resetClientPassword(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManageStudents(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	basePath := "/" + currentPageRole(viewer.Role).String() + "/clients"
	clientID := chi.URLParam(r, "clientID")
	user, newPassword, err := h.store.ResetStudentPassword(r.Context(), viewer, clientID)
	if err != nil {
		notice := "Gagal mereset password client."
		if errors.Is(err, dashboard.ErrForbidden) {
			notice = "Kamu tidak punya akses untuk mereset password client ini."
		}
		http.Redirect(w, r, basePath+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
		return
	}
	notice := "Password baru untuk " + user.Username + ": " + newPassword + " (catat sekarang, tidak akan ditampilkan lagi)"
	http.Redirect(w, r, basePath+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}
