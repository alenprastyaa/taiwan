package dashboard

import (
	"context"
	"log/slog"
	"strings"
	"time"
)

const activityLogSelect = `SELECT a.id, a.staff_id, COALESCE(s.name, ''), a.client_id, COALESCE(c.name, ''), a.action_type, a.description, a.created_at FROM activity_log a LEFT JOIN users s ON s.id = a.staff_id LEFT JOIN clients c ON c.id = a.client_id`

// ListActivityLog returns the activity feed scoped to the viewer: owner sees
// every staff member's entries (optionally filtered to one day), staff sees
// only their own. Filtering by day happens in Go, matching this codebase's
// convention of aggregating in Go rather than SQL GROUP BY/WHERE on dates
// stored as free-text-adjacent timestamps.
func (r *SQLRepository) ListActivityLog(ctx context.Context, viewer User, dateFilter string) ([]ActivityLog, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return nil, ErrForbidden
	}
	rows, err := r.query(ctx, activityLogSelect+` ORDER BY a.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dateFilter = strings.TrimSpace(dateFilter)
	var entries []ActivityLog
	for rows.Next() {
		var entry ActivityLog
		if err := rows.Scan(&entry.ID, &entry.StaffID, &entry.StaffName, &entry.ClientID, &entry.ClientName, &entry.ActionType, &entry.Description, &entry.CreatedAt); err != nil {
			return nil, err
		}
		if viewer.Role == RoleStaff && entry.StaffID != viewer.ID {
			continue
		}
		if dateFilter != "" && entry.CreatedAt.Format("2006-01-02") != dateFilter {
			continue
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (r *SQLRepository) CreateActivityNote(ctx context.Context, viewer User, clientID, note string) (ActivityLog, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return ActivityLog{}, ErrForbidden
	}
	note = strings.TrimSpace(note)
	if note == "" {
		return ActivityLog{}, ErrInvalidInput
	}
	id := newID("activity")
	now := time.Now()
	if _, err := r.exec(ctx, `INSERT INTO activity_log (id, staff_id, client_id, action_type, description, created_at) VALUES (?, ?, ?, ?, ?, ?)`, id, viewer.ID, strings.TrimSpace(clientID), "manual_note", note, now); err != nil {
		return ActivityLog{}, err
	}
	entry, err := scanActivityLogByID(ctx, r, id)
	if err != nil {
		return ActivityLog{}, err
	}
	return entry, nil
}

func scanActivityLogByID(ctx context.Context, r *SQLRepository, id string) (ActivityLog, error) {
	var entry ActivityLog
	err := r.queryRow(ctx, activityLogSelect+` WHERE a.id = ?`, id).Scan(&entry.ID, &entry.StaffID, &entry.StaffName, &entry.ClientID, &entry.ClientName, &entry.ActionType, &entry.Description, &entry.CreatedAt)
	if err != nil {
		return ActivityLog{}, err
	}
	return entry, nil
}

// logActivity is best-effort, non-transactional logging: it runs as a
// follow-up call after a primary write already succeeded, and swallows
// (while logging) any insert error rather than failing the user-facing
// action or wrapping already-tested single-statement writes in new
// transactions.
func (r *SQLRepository) logActivity(ctx context.Context, staffID, clientID, actionType, description string) {
	if _, err := r.exec(ctx, `INSERT INTO activity_log (id, staff_id, client_id, action_type, description, created_at) VALUES (?, ?, ?, ?, ?, ?)`, newID("activity"), staffID, clientID, actionType, description, time.Now()); err != nil {
		slog.Default().Error("write activity log", "action", actionType, "error", err)
	}
}
