package web

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"university_agency/internal/dashboard"
)

const invoicesPath = "/owner/invoices"

func parseOrderAmount(raw string) int64 {
	value, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if value < 0 {
		return 0
	}
	return value
}

func parseOrderDueDate(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func (h Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form tidak valid", http.StatusBadRequest)
		return
	}
	order, err := h.store.CreateOrder(r.Context(), viewer, dashboard.CreateOrderInput{
		ClientID:    r.FormValue("client_id"),
		PackageName: r.FormValue("package_name"),
		Total:       parseOrderAmount(r.FormValue("total")),
		Paid:        parseOrderAmount(r.FormValue("paid")),
		DueDate:     parseOrderDueDate(r.FormValue("due_date")),
	})
	if err != nil {
		http.Redirect(w, r, invoicesPath+"?new=1&notice="+url.QueryEscape("Gagal membuat order. Pilih client dan isi total tagihan."), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, invoicesPath+"?order="+url.QueryEscape(order.Code)+"&notice="+url.QueryEscape("Order "+order.Code+" berhasil dibuat."), http.StatusSeeOther)
}

func (h Handler) updateOrder(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form tidak valid", http.StatusBadRequest)
		return
	}
	orderID := chi.URLParam(r, "orderID")
	order, err := h.store.UpdateOrder(r.Context(), viewer, orderID, dashboard.UpdateOrderInput{
		PackageName: r.FormValue("package_name"),
		Total:       parseOrderAmount(r.FormValue("total")),
		Paid:        parseOrderAmount(r.FormValue("paid")),
		DueDate:     parseOrderDueDate(r.FormValue("due_date")),
	})
	if err != nil {
		http.Redirect(w, r, invoicesPath+"?notice="+url.QueryEscape("Gagal memperbarui order. Isi total tagihan."), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, invoicesPath+"?order="+url.QueryEscape(order.Code)+"&notice="+url.QueryEscape("Order "+order.Code+" berhasil diperbarui."), http.StatusSeeOther)
}

func (h Handler) deleteOrder(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	orderID := chi.URLParam(r, "orderID")
	err := h.store.DeleteOrder(r.Context(), viewer, orderID)
	notice := "Order berhasil dihapus."
	if err != nil {
		notice = "Gagal menghapus order."
	}
	http.Redirect(w, r, invoicesPath+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}
