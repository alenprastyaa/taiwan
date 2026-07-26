package dashboard

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"
)

func scanOrderPayment(scanner rowScanner) (OrderPayment, error) {
	var payment OrderPayment
	var status string
	var verifiedAt sql.NullTime
	if err := scanner.Scan(&payment.ID, &payment.OrderID, &payment.Amount, &payment.Note, &payment.ProofFileName, &payment.ProofStoragePath, &status, &payment.SubmittedBy, &payment.SubmittedAt, &verifiedAt, &payment.RejectReason); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return OrderPayment{}, ErrNotFound
		}
		return OrderPayment{}, err
	}
	payment.Status = PaymentStatus(status)
	if verifiedAt.Valid {
		payment.VerifiedAt = &verifiedAt.Time
	}
	return payment, nil
}

const orderPaymentColumns = `id, order_id, amount, note, proof_file_name, proof_storage_path, status, submitted_by, submitted_at, verified_at, reject_reason`

// recomputeOrderPaidState is the single place that turns the order_payments
// ledger into the cached orders.paid/status/paid_at fields every other part
// of the app reads. Every write to the ledger (a new pending submission
// aside — that alone can't move money) ends by calling this, so orders.paid
// can never drift from "sum of verified cicilan."
func (r *SQLRepository) recomputeOrderPaidState(ctx context.Context, orderID string) (Order, error) {
	order, err := r.findOrderByID(ctx, orderID)
	if err != nil {
		return Order{}, err
	}
	var paid int64
	if err := r.queryRow(ctx, `SELECT COALESCE(SUM(amount), 0) FROM order_payments WHERE order_id = ? AND status = ?`, orderID, PaymentVerified).Scan(&paid); err != nil {
		return Order{}, err
	}
	if paid < 0 {
		paid = 0
	}
	if paid > order.Total {
		paid = order.Total
	}
	status := OrderUnpaid
	var paidAt *time.Time
	if order.Total > 0 && paid >= order.Total {
		status = OrderPaid
		paidAt = order.PaidAt
		if paidAt == nil {
			now := time.Now()
			paidAt = &now
		}
	}
	if _, err := r.exec(ctx, `UPDATE orders SET paid = ?, status = ?, paid_at = ? WHERE id = ?`, paid, status, paidAt, orderID); err != nil {
		return Order{}, err
	}
	return r.findOrderByID(ctx, orderID)
}

// insertOrderPayment is the shared row-writer every ledger mutation uses.
func (r *SQLRepository) insertOrderPayment(ctx context.Context, payment OrderPayment) error {
	_, err := r.exec(ctx, `INSERT INTO order_payments (`+orderPaymentColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		payment.ID, payment.OrderID, payment.Amount, payment.Note, payment.ProofFileName, payment.ProofStoragePath, payment.Status, payment.SubmittedBy, payment.SubmittedAt, payment.VerifiedAt, payment.RejectReason)
	return err
}

// VerifyOrderPayment confirms a client-submitted (pending) cicilan: owner
// or the client's PIC staff only. Recomputes orders.paid/status afterward.
func (r *SQLRepository) VerifyOrderPayment(ctx context.Context, viewer User, paymentID string) (Order, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return Order{}, ErrForbidden
	}
	payment, err := scanOrderPayment(r.queryRow(ctx, `SELECT `+orderPaymentColumns+` FROM order_payments WHERE id = ?`, paymentID))
	if err != nil {
		return Order{}, err
	}
	order, err := r.findOrderByID(ctx, payment.OrderID)
	if err != nil {
		return Order{}, err
	}
	if viewer.Role == RoleStaff {
		client, err := r.findClientByID(ctx, order.ClientID)
		if err != nil {
			return Order{}, err
		}
		if client.PICStaffID != viewer.ID {
			return Order{}, ErrForbidden
		}
	}
	now := time.Now()
	result, err := r.exec(ctx, `UPDATE order_payments SET status = ?, verified_at = ? WHERE id = ? AND status = ?`, PaymentVerified, now, paymentID, PaymentPending)
	if err != nil {
		return Order{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return Order{}, ErrConflict
	}
	updated, err := r.recomputeOrderPaidState(ctx, order.ID)
	if err != nil {
		return Order{}, err
	}
	r.logActivity(ctx, viewer.ID, order.ClientID, "payment_verified", "Cicilan Rp"+formatAmount(payment.Amount)+" untuk "+order.Code+" diverifikasi")
	return updated, nil
}

// RejectOrderPayment declines a pending cicilan without touching paid — if
// no other pending entries remain for the order, it drops the order back
// out of "menunggu verifikasi".
func (r *SQLRepository) RejectOrderPayment(ctx context.Context, viewer User, paymentID, reason string) (Order, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return Order{}, ErrForbidden
	}
	payment, err := scanOrderPayment(r.queryRow(ctx, `SELECT `+orderPaymentColumns+` FROM order_payments WHERE id = ?`, paymentID))
	if err != nil {
		return Order{}, err
	}
	order, err := r.findOrderByID(ctx, payment.OrderID)
	if err != nil {
		return Order{}, err
	}
	if viewer.Role == RoleStaff {
		client, err := r.findClientByID(ctx, order.ClientID)
		if err != nil {
			return Order{}, err
		}
		if client.PICStaffID != viewer.ID {
			return Order{}, ErrForbidden
		}
	}
	result, err := r.exec(ctx, `UPDATE order_payments SET status = ?, reject_reason = ? WHERE id = ? AND status = ?`, PaymentRejected, strings.TrimSpace(reason), paymentID, PaymentPending)
	if err != nil {
		return Order{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return Order{}, ErrConflict
	}
	if order.Status == OrderWaitingVerification {
		var pendingCount int
		if err := r.queryRow(ctx, `SELECT COUNT(*) FROM order_payments WHERE order_id = ? AND status = ?`, order.ID, PaymentPending).Scan(&pendingCount); err != nil {
			return Order{}, err
		}
		if pendingCount == 0 {
			if _, err := r.exec(ctx, `UPDATE orders SET status = ? WHERE id = ?`, OrderUnpaid, order.ID); err != nil {
				return Order{}, err
			}
		}
	}
	r.logActivity(ctx, viewer.ID, order.ClientID, "payment_rejected", "Cicilan Rp"+formatAmount(payment.Amount)+" untuk "+order.Code+" ditolak")
	return r.findOrderByID(ctx, order.ID)
}

// CreateManualInstallment is the owner's "+ Cicilan Manual" — a
// pre-verified entry, since the owner is the one attesting it happened.
func (r *SQLRepository) CreateManualInstallment(ctx context.Context, viewer User, orderID string, amount int64, note string) (Order, error) {
	if viewer.Role != RoleOwner {
		return Order{}, ErrForbidden
	}
	if amount == 0 {
		return Order{}, ErrInvalidInput
	}
	order, err := r.findOrderByID(ctx, orderID)
	if err != nil {
		return Order{}, err
	}
	now := time.Now()
	note = strings.TrimSpace(note)
	if note == "" {
		note = "Cicilan manual"
	}
	if err := r.insertOrderPayment(ctx, OrderPayment{
		ID: newID("payment"), OrderID: order.ID, Amount: amount, Note: note,
		Status: PaymentVerified, SubmittedBy: viewer.ID, SubmittedAt: now, VerifiedAt: &now,
	}); err != nil {
		return Order{}, err
	}
	updated, err := r.recomputeOrderPaidState(ctx, order.ID)
	if err != nil {
		return Order{}, err
	}
	r.logActivity(ctx, viewer.ID, order.ClientID, "payment_manual", "Cicilan manual Rp"+formatAmount(amount)+" ditambahkan untuk "+order.Code)
	return updated, nil
}

// ListAllOrderPayments loads every payment entry the viewer is allowed to
// see, following the same "load broad, filter in the template" convention
// ListOrders/ListClients already use — the UI filters this down to the
// active order's rows.
func (r *SQLRepository) ListAllOrderPayments(ctx context.Context, viewer User, viewRole Role) ([]OrderPayment, error) {
	clients, err := r.ListClients(ctx, viewer, viewRole)
	if err != nil {
		return nil, err
	}
	visible := make(map[string]bool, len(clients))
	for _, client := range clients {
		visible[client.ID] = true
	}
	rows, err := r.query(ctx, `SELECT p.id, p.order_id, p.amount, p.note, p.proof_file_name, p.proof_storage_path, p.status, p.submitted_by, p.submitted_at, p.verified_at, p.reject_reason, o.client_id FROM order_payments p JOIN orders o ON o.id = p.order_id ORDER BY p.submitted_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []OrderPayment
	for rows.Next() {
		var payment OrderPayment
		var status string
		var verifiedAt sql.NullTime
		var clientID string
		if err := rows.Scan(&payment.ID, &payment.OrderID, &payment.Amount, &payment.Note, &payment.ProofFileName, &payment.ProofStoragePath, &status, &payment.SubmittedBy, &payment.SubmittedAt, &verifiedAt, &payment.RejectReason, &clientID); err != nil {
			return nil, err
		}
		if !visible[clientID] {
			continue
		}
		payment.Status = PaymentStatus(status)
		if verifiedAt.Valid {
			payment.VerifiedAt = &verifiedAt.Time
		}
		payments = append(payments, payment)
	}
	return payments, rows.Err()
}

// ensureOrderPaymentsBackfilled seeds one verified ledger row per existing
// order whose paid > 0 predates the order_payments table — without this,
// the very first recompute on such an order would sum zero ledger rows and
// silently wipe out its real paid amount.
func (r *SQLRepository) ensureOrderPaymentsBackfilled(ctx context.Context) error {
	rows, err := r.query(ctx, `SELECT o.id, o.paid, o.paid_at, o.created_at FROM orders o WHERE o.paid > 0 AND NOT EXISTS (SELECT 1 FROM order_payments p WHERE p.order_id = o.id)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type missingOrder struct {
		id        string
		paid      int64
		paidAt    sql.NullTime
		createdAt time.Time
	}
	var missing []missingOrder
	for rows.Next() {
		var item missingOrder
		if err := rows.Scan(&item.id, &item.paid, &item.paidAt, &item.createdAt); err != nil {
			return err
		}
		missing = append(missing, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, item := range missing {
		at := item.createdAt
		if item.paidAt.Valid {
			at = item.paidAt.Time
		}
		if err := r.insertOrderPayment(ctx, OrderPayment{
			ID: newID("payment"), OrderID: item.id, Amount: item.paid,
			Note: "Saldo awal (migrasi riwayat cicilan)", Status: PaymentVerified,
			SubmittedAt: at, VerifiedAt: &at,
		}); err != nil {
			return err
		}
	}
	return nil
}

func formatAmount(amount int64) string {
	return strconv.FormatInt(amount, 10)
}
