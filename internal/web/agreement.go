package web

import (
	"net"
	"net/http"

	"university_agency/internal/dashboard"

	"github.com/go-chi/chi/v5"
)

// requireSignedAgreement closes the gap where a student could POST directly
// to an action endpoint (upload, payment proof, chat) before ever loading a
// gated page. Returns true if the request may proceed. Non-student viewers
// always pass.
func (h Handler) requireSignedAgreement(w http.ResponseWriter, r *http.Request) bool {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleStudent {
		return true
	}
	signed, err := h.store.HasSignedAgreement(r.Context(), viewer)
	if err != nil {
		http.Error(w, "gagal memverifikasi status perjanjian", http.StatusInternalServerError)
		return false
	}
	if signed {
		return true
	}
	http.Redirect(w, r, "/student/agreement", http.StatusFound)
	return false
}

func requestIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (h Handler) signAgreement(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleStudent {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/student/agreement", http.StatusSeeOther)
		return
	}
	if r.FormValue("agree") != "1" || r.FormValue("full_name") == "" {
		http.Redirect(w, r, "/student/agreement?notice=Centang+persetujuan+dan+isi+nama+lengkap+terlebih+dahulu.", http.StatusSeeOther)
		return
	}
	_, err := h.store.SignAgreement(r.Context(), viewer, r.FormValue("full_name"), requestIP(r), r.UserAgent())
	if err != nil {
		http.Redirect(w, r, "/student/agreement?notice=Gagal+menyimpan+persetujuan.", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/student", http.StatusSeeOther)
}

func (h Handler) downloadOwnAgreementPDF(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	agreements, err := h.store.ListClientAgreements(r.Context(), viewer, viewer.Role)
	if err != nil || len(agreements) == 0 {
		http.NotFound(w, r)
		return
	}
	pdf := renderAgreementPDF(h.cfg.App.Name, agreements[0])
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename="+strconvQuote("perjanjian-"+agreements[0].ClientName+".pdf"))
	_, _ = w.Write(pdf)
}

func (h Handler) downloadClientAgreementPDF(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	clientID := chi.URLParam(r, "clientID")
	agreements, err := h.store.ListClientAgreements(r.Context(), viewer, viewer.Role)
	if err != nil {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	for _, agreement := range agreements {
		if agreement.ClientID == clientID {
			pdf := renderAgreementPDF(h.cfg.App.Name, agreement)
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", "attachment; filename="+strconvQuote("perjanjian-"+agreement.ClientName+".pdf"))
			_, _ = w.Write(pdf)
			return
		}
	}
	http.NotFound(w, r)
}
