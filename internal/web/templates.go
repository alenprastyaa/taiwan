package web

import (
	"net/http"
	"net/url"

	"university_agency/internal/dashboard"

	"github.com/go-chi/chi/v5"
)

func templatesPath(viewer dashboard.User) string {
	return "/" + currentPageRole(viewer.Role).String() + "/templates"
}

func textTemplateInputFromForm(r *http.Request) dashboard.TextTemplateInput {
	return dashboard.TextTemplateInput{
		Title:    r.FormValue("title"),
		Body:     r.FormValue("body"),
		Category: r.FormValue("category"),
	}
}

func (h Handler) createTextTemplate(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManageTemplates(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, templatesPath(viewer)+"?manage=1", http.StatusSeeOther)
		return
	}
	_, err := h.store.CreateTextTemplate(r.Context(), viewer, textTemplateInputFromForm(r))
	notice := "Template baru berhasil ditambahkan."
	if err != nil {
		notice = "Gagal menambahkan template. Judul dan isi wajib diisi."
	}
	http.Redirect(w, r, templatesPath(viewer)+"?manage=1&notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) updateTextTemplate(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManageTemplates(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, templatesPath(viewer)+"?manage=1", http.StatusSeeOther)
		return
	}
	_, err := h.store.UpdateTextTemplate(r.Context(), viewer, chi.URLParam(r, "templateID"), textTemplateInputFromForm(r))
	notice := "Template berhasil diperbarui."
	if err != nil {
		notice = "Gagal memperbarui template."
	}
	http.Redirect(w, r, templatesPath(viewer)+"?manage=1&notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) deleteTextTemplate(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManageTemplates(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	err := h.store.DeleteTextTemplate(r.Context(), viewer, chi.URLParam(r, "templateID"))
	notice := "Template dihapus."
	if err != nil {
		notice = "Gagal menghapus template."
	}
	http.Redirect(w, r, templatesPath(viewer)+"?manage=1&notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) moveTextTemplate(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if !canManageTemplates(viewer) {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, templatesPath(viewer)+"?manage=1", http.StatusSeeOther)
		return
	}
	err := h.store.ReorderTextTemplate(r.Context(), viewer, chi.URLParam(r, "templateID"), r.FormValue("direction"))
	notice := "?manage=1"
	if err != nil {
		notice = "?manage=1&notice=" + url.QueryEscape("Gagal mengubah urutan template.")
	}
	http.Redirect(w, r, templatesPath(viewer)+notice, http.StatusSeeOther)
}

func canManageTemplates(viewer dashboard.User) bool {
	return viewer.Role == dashboard.RoleOwner || viewer.Role == dashboard.RoleStaff
}
