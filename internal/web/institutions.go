package web

import (
	"net/http"
	"net/url"

	"university_agency/internal/dashboard"

	"github.com/go-chi/chi/v5"
)

func institutionsPath(viewer dashboard.User) string {
	return "/" + currentPageRole(viewer.Role).String() + "/institutions"
}

func institutionContactInputFromForm(r *http.Request) dashboard.InstitutionContactInput {
	return dashboard.InstitutionContactInput{
		Name:     r.FormValue("name"),
		Category: r.FormValue("category"),
		Phone:    r.FormValue("phone"),
		Notes:    r.FormValue("notes"),
	}
}

func canManageInstitutions(viewer dashboard.User) bool {
	return viewer.Role == dashboard.RoleOwner || viewer.Role == dashboard.RoleStaff
}

func (h Handler) createInstitutionContact(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManageInstitutions(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, institutionsPath(viewer)+"?manage=1", http.StatusSeeOther)
		return
	}
	_, err := h.store.CreateInstitutionContact(r.Context(), viewer, institutionContactInputFromForm(r))
	notice := "Kontak institusi baru berhasil ditambahkan."
	if err != nil {
		notice = "Gagal menambahkan kontak. Nama dan nomor WhatsApp wajib diisi."
	}
	http.Redirect(w, r, institutionsPath(viewer)+"?manage=1&notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) updateInstitutionContact(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManageInstitutions(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, institutionsPath(viewer)+"?manage=1", http.StatusSeeOther)
		return
	}
	_, err := h.store.UpdateInstitutionContact(r.Context(), viewer, chi.URLParam(r, "contactID"), institutionContactInputFromForm(r))
	notice := "Kontak institusi berhasil diperbarui."
	if err != nil {
		notice = "Gagal memperbarui kontak."
	}
	http.Redirect(w, r, institutionsPath(viewer)+"?manage=1&notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) deleteInstitutionContact(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManageInstitutions(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	err := h.store.DeleteInstitutionContact(r.Context(), viewer, chi.URLParam(r, "contactID"))
	notice := "Kontak institusi dihapus."
	if err != nil {
		notice = "Gagal menghapus kontak."
	}
	http.Redirect(w, r, institutionsPath(viewer)+"?manage=1&notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) moveInstitutionContact(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManageInstitutions(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, institutionsPath(viewer)+"?manage=1", http.StatusSeeOther)
		return
	}
	err := h.store.ReorderInstitutionContact(r.Context(), viewer, chi.URLParam(r, "contactID"), r.FormValue("direction"))
	notice := "?manage=1"
	if err != nil {
		notice = "?manage=1&notice=" + url.QueryEscape("Gagal mengubah urutan kontak.")
	}
	http.Redirect(w, r, institutionsPath(viewer)+notice, http.StatusSeeOther)
}
