package dashboard

import (
	"context"
	"strings"
	"time"
)

// ListClientStageHistory returns a client's stage-transition history,
// oldest first, so the client-facing timeline (buildClientTimeline in
// ui.go) can look up when they left each stage. Visibility mirrors
// UpdateClient: owner sees any client, staff only their own PIC clients,
// students only themselves.
func (r *SQLRepository) ListClientStageHistory(ctx context.Context, viewer User, clientID string) ([]ClientStageEvent, error) {
	client, err := r.findClientByID(ctx, clientID)
	if err != nil {
		return nil, err
	}
	switch viewer.Role {
	case RoleOwner:
	case RoleStaff:
		if client.PICStaffID != viewer.ID {
			return nil, ErrForbidden
		}
	case RoleStudent:
		if client.UserID != viewer.ID {
			return nil, ErrForbidden
		}
	default:
		return nil, ErrForbidden
	}
	rows, err := r.query(ctx, `SELECT id, client_id, stage_name, entered_at FROM client_stage_history WHERE client_id = ? ORDER BY entered_at ASC`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []ClientStageEvent
	for rows.Next() {
		var event ClientStageEvent
		if err := rows.Scan(&event.ID, &event.ClientID, &event.StageName, &event.EnteredAt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

// ensureClientStageHistoryBackfilled seeds one history row per client who
// has a current_stage but no history yet — either client_stage_history is
// new on an existing database, or the client was created before this table
// existed. Stages before this baseline row won't have real transition
// dates (the data was never recorded); this only guarantees the *current*
// stage has an anchor to build forward from.
func (r *SQLRepository) ensureClientStageHistoryBackfilled(ctx context.Context) error {
	rows, err := r.query(ctx, `SELECT c.id, c.current_stage, c.created_at FROM clients c WHERE c.current_stage <> '' AND NOT EXISTS (SELECT 1 FROM client_stage_history h WHERE h.client_id = c.id)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type missingClient struct {
		id        string
		stageName string
		createdAt time.Time
	}
	var missing []missingClient
	for rows.Next() {
		var item missingClient
		if err := rows.Scan(&item.id, &item.stageName, &item.createdAt); err != nil {
			return err
		}
		missing = append(missing, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, item := range missing {
		if _, err := r.exec(ctx, `INSERT INTO client_stage_history (id, client_id, stage_name, entered_at) VALUES (?, ?, ?, ?)`, newID("stagehist"), item.id, strings.TrimSpace(item.stageName), item.createdAt); err != nil {
			return err
		}
	}
	return nil
}
