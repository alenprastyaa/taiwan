package dashboard

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

const clientAgreementSelect = `SELECT a.id, a.client_id, COALESCE(c.name, ''), a.agreement_version, a.agreement_text, a.full_name_typed, a.agreed_at, a.ip_address, a.user_agent FROM client_agreements a LEFT JOIN clients c ON c.id = a.client_id`

func scanClientAgreement(scanner rowScanner) (ClientAgreement, error) {
	var agreement ClientAgreement
	if err := scanner.Scan(&agreement.ID, &agreement.ClientID, &agreement.ClientName, &agreement.AgreementVersion, &agreement.AgreementText, &agreement.FullNameTyped, &agreement.AgreedAt, &agreement.IPAddress, &agreement.UserAgent); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ClientAgreement{}, ErrNotFound
		}
		return ClientAgreement{}, err
	}
	return agreement, nil
}

// ListClientAgreements is scoped like documents/intake forms: the student's
// own single row, staff's PIC clients, or all for owner — this is also how
// per-client agreement PDF downloads enforce RBAC (list-scoped-then-search).
func (r *SQLRepository) ListClientAgreements(ctx context.Context, viewer User, viewRole Role) ([]ClientAgreement, error) {
	clients, err := r.ListClients(ctx, viewer, viewRole)
	if err != nil {
		return nil, err
	}
	visible := make(map[string]bool, len(clients))
	for _, client := range clients {
		visible[client.ID] = true
	}

	rows, err := r.query(ctx, clientAgreementSelect+` ORDER BY a.agreed_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agreements []ClientAgreement
	for rows.Next() {
		agreement, err := scanClientAgreement(rows)
		if err != nil {
			return nil, err
		}
		if visible[agreement.ClientID] {
			agreements = append(agreements, agreement)
		}
	}
	return agreements, rows.Err()
}

// HasSignedAgreement is a cheap existence check keyed off the viewer's own
// client row — mirrors StudentHasPaymentAccess. Non-student viewers always
// pass (the gate itself only ever calls this for RoleStudent viewers, but
// this method stays safe to call for any role).
func (r *SQLRepository) HasSignedAgreement(ctx context.Context, viewer User) (bool, error) {
	if viewer.Role != RoleStudent {
		return true, nil
	}
	clients, err := r.ListClients(ctx, viewer, RoleStudent)
	if err != nil {
		return false, err
	}
	if len(clients) == 0 {
		return false, nil
	}
	var count int
	if err := r.queryRow(ctx, `SELECT COUNT(*) FROM client_agreements WHERE client_id = ?`, clients[0].ID).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// SignAgreement is idempotent: a double-submit (e.g. a slow double-click)
// just returns the already-signed record instead of erroring.
func (r *SQLRepository) SignAgreement(ctx context.Context, viewer User, fullNameTyped, ipAddress, userAgent string) (ClientAgreement, error) {
	if viewer.Role != RoleStudent {
		return ClientAgreement{}, ErrForbidden
	}
	fullNameTyped = strings.TrimSpace(fullNameTyped)
	if fullNameTyped == "" {
		return ClientAgreement{}, ErrInvalidInput
	}
	clients, err := r.ListClients(ctx, viewer, RoleStudent)
	if err != nil {
		return ClientAgreement{}, err
	}
	if len(clients) == 0 {
		return ClientAgreement{}, ErrForbidden
	}
	client := clients[0]

	existing, err := scanClientAgreement(r.queryRow(ctx, clientAgreementSelect+` WHERE a.client_id = ?`, client.ID))
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return ClientAgreement{}, err
	}

	id := newID("agreement")
	now := time.Now()
	if _, err := r.exec(ctx, `INSERT INTO client_agreements (id, client_id, agreement_version, agreement_text, full_name_typed, agreed_at, ip_address, user_agent) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, client.ID, CurrentAgreementVersion, AgreementText, fullNameTyped, now, strings.TrimSpace(ipAddress), strings.TrimSpace(userAgent)); err != nil {
		return ClientAgreement{}, err
	}
	return scanClientAgreement(r.queryRow(ctx, clientAgreementSelect+` WHERE a.id = ?`, id))
}
