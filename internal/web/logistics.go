package web

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"university_agency/internal/dashboard"
)

func (h Handler) createShipment(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	basePath := "/" + currentPageRole(viewer.Role).String() + "/logistics"
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, basePath+"?notice="+url.QueryEscape("Form pengiriman tidak valid."), http.StatusSeeOther)
		return
	}
	_, err := h.store.CreateShipment(r.Context(), viewer, dashboard.CreateShipmentInput{
		ClientID:         r.FormValue("client_id"),
		Direction:        r.FormValue("direction"),
		Courier:          r.FormValue("courier"),
		TrackingNumber:   r.FormValue("tracking_number"),
		Contents:         r.FormValue("contents"),
		SenderAddress:    r.FormValue("sender_address"),
		RecipientAddress: r.FormValue("recipient_address"),
		ShippedDateLabel: r.FormValue("shipped_date_label"),
	})
	notice := "Pengiriman baru berhasil dicatat."
	if err != nil {
		notice = "Gagal mencatat pengiriman. Pastikan client dan ekspedisi terisi."
	}
	http.Redirect(w, r, basePath+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) markShipmentReceived(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	basePath := "/" + currentPageRole(viewer.Role).String() + "/logistics"
	_, err := h.store.MarkShipmentReceived(r.Context(), viewer, chi.URLParam(r, "shipmentID"))
	notice := "Pengiriman sudah ditandai diterima."
	if err != nil {
		notice = "Gagal memperbarui status atau akses ditolak."
	}
	http.Redirect(w, r, basePath+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}
