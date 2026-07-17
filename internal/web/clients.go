package web

import (
	"net/http"
	"net/url"

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
	user, err := h.store.CreateStudent(r.Context(), dashboard.CreateStudentInput{
		Username:    r.FormValue("username"),
		Password:    r.FormValue("password"),
		Name:        r.FormValue("name"),
		Email:       r.FormValue("email"),
		Phone:       r.FormValue("phone"),
		PackageName: r.FormValue("package_name"),
		Country:     r.FormValue("country"),
		Campus:      r.FormValue("campus"),
		PICStaffID:  r.FormValue("pic_staff_id"),
	})
	if err != nil {
		http.Redirect(w, r, basePath+"?new=1&notice="+url.QueryEscape("Gagal menambah client. Username harus unik dan password minimal 8 karakter."), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, basePath+"?notice="+url.QueryEscape("Client "+user.Username+" berhasil ditambahkan."), http.StatusSeeOther)
}
