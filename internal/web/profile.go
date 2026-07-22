package web

import (
	"errors"
	"net/http"
	"net/url"

	"university_agency/internal/dashboard"
)

// profilePath sends the user back to their own role's Settings page, where
// the profile/password forms live — the only place these actions are
// reachable from.
func profilePath(viewer dashboard.User) string {
	return "/" + currentPageRole(viewer.Role).String() + "/settings"
}

func (h Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, profilePath(viewer), http.StatusSeeOther)
		return
	}
	_, err := h.store.UpdateOwnProfile(r.Context(), viewer, dashboard.UpdateOwnProfileInput{
		Name:  r.FormValue("name"),
		Email: r.FormValue("email"),
		Phone: r.FormValue("phone"),
	})
	notice := "Profil berhasil diperbarui."
	if err != nil {
		notice = "Gagal memperbarui profil. Nama wajib diisi."
	}
	http.Redirect(w, r, profilePath(viewer)+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) changeOwnPassword(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, profilePath(viewer), http.StatusSeeOther)
		return
	}
	newPassword := r.FormValue("new_password")
	if newPassword != r.FormValue("confirm_password") {
		http.Redirect(w, r, profilePath(viewer)+"?notice="+url.QueryEscape("Konfirmasi password baru tidak cocok."), http.StatusSeeOther)
		return
	}
	err := h.store.ChangeOwnPassword(r.Context(), viewer, r.FormValue("current_password"), newPassword)
	notice := "Password berhasil diubah."
	if errors.Is(err, dashboard.ErrForbidden) {
		notice = "Password saat ini salah."
	} else if err != nil {
		notice = "Gagal mengubah password. Password baru minimal 8 karakter."
	}
	http.Redirect(w, r, profilePath(viewer)+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}
