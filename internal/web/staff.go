package web

import (
	"errors"
	"net/http"
	"net/url"

	"university_agency/internal/dashboard"

	"github.com/go-chi/chi/v5"
)

const staffManagementPath = "/owner/staff"

func (h Handler) createStaff(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, staffManagementPath+"?new=1", http.StatusSeeOther)
		return
	}
	user, err := h.store.CreateStaff(r.Context(), viewer, dashboard.CreateStaffInput{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Phone:    r.FormValue("phone"),
	})
	if err != nil {
		notice := "Gagal membuat akun staf. Username harus unik dan password minimal 8 karakter."
		http.Redirect(w, r, staffManagementPath+"?new=1&notice="+url.QueryEscape(notice), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, staffManagementPath+"?notice="+url.QueryEscape("Akun staf "+user.Username+" berhasil dibuat."), http.StatusSeeOther)
}

func (h Handler) updateStaff(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, staffManagementPath, http.StatusSeeOther)
		return
	}
	_, err := h.store.UpdateStaff(r.Context(), viewer, chi.URLParam(r, "staffID"), dashboard.UpdateStaffInput{
		Name:  r.FormValue("name"),
		Email: r.FormValue("email"),
		Phone: r.FormValue("phone"),
	})
	notice := "Data staf berhasil diperbarui."
	if err != nil {
		notice = "Gagal memperbarui data staf. Nama wajib diisi."
	}
	http.Redirect(w, r, staffManagementPath+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) toggleStaffActive(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	err := h.store.ToggleStaffActive(r.Context(), viewer, chi.URLParam(r, "staffID"))
	notice := "Status staf berhasil diperbarui."
	if err != nil {
		notice = "Gagal mengubah status staf."
	}
	http.Redirect(w, r, staffManagementPath+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) deleteStaff(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	err := h.store.DeleteStaff(r.Context(), viewer, chi.URLParam(r, "staffID"))
	notice := "Akun staf berhasil dihapus permanen."
	if errors.Is(err, dashboard.ErrConflict) {
		notice = "Staf tidak bisa dihapus karena masih jadi PIC client, atau punya task/pengeluaran/pengiriman tercatat. Nonaktifkan saja jika sudah tidak dipakai."
	} else if err != nil {
		notice = "Gagal menghapus akun staf."
	}
	http.Redirect(w, r, staffManagementPath+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) resetStaffPassword(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	user, newPassword, err := h.store.ResetStaffPassword(r.Context(), viewer, chi.URLParam(r, "staffID"))
	if err != nil {
		http.Redirect(w, r, staffManagementPath+"?notice="+url.QueryEscape("Gagal mereset password staf."), http.StatusSeeOther)
		return
	}
	notice := "Password baru untuk " + user.Username + ": " + newPassword + " (catat sekarang, tidak akan ditampilkan lagi)"
	http.Redirect(w, r, staffManagementPath+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}
