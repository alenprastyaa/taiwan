package web

import (
	"net/http"
	"net/url"
	"strings"

	"university_agency/internal/dashboard"
	"university_agency/internal/security"
)

// resetTestDataPhrase is the fixed text an owner must type before a data
// reset is allowed to run, on top of re-entering their own password — two
// independent checks for an action that has no undo.
const resetTestDataPhrase = "RESET DATA"

func (h Handler) resetTestData(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/owner/settings?notice="+url.QueryEscape("Form tidak valid."), http.StatusSeeOther)
		return
	}
	if strings.TrimSpace(r.FormValue("confirm_phrase")) != resetTestDataPhrase {
		http.Redirect(w, r, "/owner/settings?notice="+url.QueryEscape("Frasa konfirmasi tidak sesuai. Reset dibatalkan."), http.StatusSeeOther)
		return
	}
	if !security.CheckPassword(viewer.PasswordHash, r.FormValue("password")) {
		http.Redirect(w, r, "/owner/settings?notice="+url.QueryEscape("Password salah. Reset dibatalkan."), http.StatusSeeOther)
		return
	}
	if err := h.store.ResetTestData(r.Context(), viewer); err != nil {
		http.Redirect(w, r, "/owner/settings?notice="+url.QueryEscape("Gagal mereset data."), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/owner/settings?notice="+url.QueryEscape("Semua data client & transaksi test berhasil direset. Akun owner/staff dan konfigurasi tetap aman."), http.StatusSeeOther)
}
