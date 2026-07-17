package web

import (
	"net/http"
	"net/url"

	"university_agency/internal/dashboard"
)

func (h Handler) createActivityNote(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner && viewer.Role != dashboard.RoleStaff {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/"+currentPageRole(viewer.Role).String()+"/activity?new=1", http.StatusSeeOther)
		return
	}
	basePath := "/" + currentPageRole(viewer.Role).String() + "/activity"
	_, err := h.store.CreateActivityNote(r.Context(), viewer, r.FormValue("client_id"), r.FormValue("note"))
	notice := "Catatan aktivitas berhasil ditambahkan."
	if err != nil {
		notice = "Gagal menambahkan catatan. Isi catatan tidak boleh kosong."
	}
	http.Redirect(w, r, basePath+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}
