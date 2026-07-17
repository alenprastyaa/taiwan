package dashboard

import (
	"context"
	"strings"
	"time"
)

func canManageInstitutions(viewer User) bool {
	return viewer.Role == RoleOwner || viewer.Role == RoleStaff
}

func (r *SQLRepository) ListInstitutionContacts(ctx context.Context) ([]InstitutionContact, error) {
	rows, err := r.query(ctx, `SELECT id, name, category, phone, notes, position, created_at FROM institution_contacts ORDER BY position ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []InstitutionContact
	for rows.Next() {
		var contact InstitutionContact
		if err := rows.Scan(&contact.ID, &contact.Name, &contact.Category, &contact.Phone, &contact.Notes, &contact.Position, &contact.CreatedAt); err != nil {
			return nil, err
		}
		contacts = append(contacts, contact)
	}
	return contacts, rows.Err()
}

func validateInstitutionContactInput(input InstitutionContactInput) error {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Phone) == "" {
		return ErrInvalidInput
	}
	return nil
}

func (r *SQLRepository) CreateInstitutionContact(ctx context.Context, viewer User, input InstitutionContactInput) (InstitutionContact, error) {
	if !canManageInstitutions(viewer) {
		return InstitutionContact{}, ErrForbidden
	}
	if err := validateInstitutionContactInput(input); err != nil {
		return InstitutionContact{}, err
	}
	existing, err := r.ListInstitutionContacts(ctx)
	if err != nil {
		return InstitutionContact{}, err
	}
	contact := InstitutionContact{
		ID:        newID("inst"),
		Name:      strings.TrimSpace(input.Name),
		Category:  strings.TrimSpace(input.Category),
		Phone:     strings.TrimSpace(input.Phone),
		Notes:     strings.TrimSpace(input.Notes),
		Position:  len(existing),
		CreatedAt: time.Now(),
	}
	if _, err := r.exec(ctx, `INSERT INTO institution_contacts (id, name, category, phone, notes, position, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, contact.ID, contact.Name, contact.Category, contact.Phone, contact.Notes, contact.Position, contact.CreatedAt); err != nil {
		return InstitutionContact{}, err
	}
	return contact, nil
}

func (r *SQLRepository) UpdateInstitutionContact(ctx context.Context, viewer User, contactID string, input InstitutionContactInput) (InstitutionContact, error) {
	if !canManageInstitutions(viewer) {
		return InstitutionContact{}, ErrForbidden
	}
	if err := validateInstitutionContactInput(input); err != nil {
		return InstitutionContact{}, err
	}
	name := strings.TrimSpace(input.Name)
	category := strings.TrimSpace(input.Category)
	phone := strings.TrimSpace(input.Phone)
	notes := strings.TrimSpace(input.Notes)
	if _, err := r.exec(ctx, `UPDATE institution_contacts SET name = ?, category = ?, phone = ?, notes = ? WHERE id = ?`, name, category, phone, notes, contactID); err != nil {
		return InstitutionContact{}, err
	}
	var updated InstitutionContact
	if err := r.queryRow(ctx, `SELECT id, name, category, phone, notes, position, created_at FROM institution_contacts WHERE id = ?`, contactID).Scan(&updated.ID, &updated.Name, &updated.Category, &updated.Phone, &updated.Notes, &updated.Position, &updated.CreatedAt); err != nil {
		return InstitutionContact{}, err
	}
	return updated, nil
}

func (r *SQLRepository) DeleteInstitutionContact(ctx context.Context, viewer User, contactID string) error {
	if !canManageInstitutions(viewer) {
		return ErrForbidden
	}
	result, err := r.exec(ctx, `DELETE FROM institution_contacts WHERE id = ?`, contactID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SQLRepository) ReorderInstitutionContact(ctx context.Context, viewer User, contactID, direction string) error {
	if !canManageInstitutions(viewer) {
		return ErrForbidden
	}
	contacts, err := r.ListInstitutionContacts(ctx)
	if err != nil {
		return err
	}
	index := -1
	for i, contact := range contacts {
		if contact.ID == contactID {
			index = i
			break
		}
	}
	if index == -1 {
		return ErrNotFound
	}
	swapWith := -1
	switch direction {
	case "up":
		swapWith = index - 1
	case "down":
		swapWith = index + 1
	default:
		return ErrInvalidInput
	}
	if swapWith < 0 || swapWith >= len(contacts) {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := r.txExec(ctx, tx, `UPDATE institution_contacts SET position = ? WHERE id = ?`, contacts[swapWith].Position, contacts[index].ID); err != nil {
		return err
	}
	if _, err := r.txExec(ctx, tx, `UPDATE institution_contacts SET position = ? WHERE id = ?`, contacts[index].Position, contacts[swapWith].ID); err != nil {
		return err
	}
	return tx.Commit()
}
