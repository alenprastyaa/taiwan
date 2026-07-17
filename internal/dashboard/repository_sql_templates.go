package dashboard

import (
	"context"
	"strings"
	"time"
)

func canManageTemplates(viewer User) bool {
	return viewer.Role == RoleOwner || viewer.Role == RoleStaff
}

func (r *SQLRepository) ListTextTemplates(ctx context.Context) ([]TextTemplate, error) {
	rows, err := r.query(ctx, `SELECT id, title, body, category, position, created_at FROM text_templates ORDER BY position ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []TextTemplate
	for rows.Next() {
		var tpl TextTemplate
		if err := rows.Scan(&tpl.ID, &tpl.Title, &tpl.Body, &tpl.Category, &tpl.Position, &tpl.CreatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, tpl)
	}
	return templates, rows.Err()
}

func validateTextTemplateInput(input TextTemplateInput) error {
	if strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Body) == "" {
		return ErrInvalidInput
	}
	return nil
}

func (r *SQLRepository) CreateTextTemplate(ctx context.Context, viewer User, input TextTemplateInput) (TextTemplate, error) {
	if !canManageTemplates(viewer) {
		return TextTemplate{}, ErrForbidden
	}
	if err := validateTextTemplateInput(input); err != nil {
		return TextTemplate{}, err
	}
	existing, err := r.ListTextTemplates(ctx)
	if err != nil {
		return TextTemplate{}, err
	}
	tpl := TextTemplate{
		ID:        newID("tpl"),
		Title:     strings.TrimSpace(input.Title),
		Body:      strings.TrimSpace(input.Body),
		Category:  strings.TrimSpace(input.Category),
		Position:  len(existing),
		CreatedAt: time.Now(),
	}
	if _, err := r.exec(ctx, `INSERT INTO text_templates (id, title, body, category, position, created_at) VALUES (?, ?, ?, ?, ?, ?)`, tpl.ID, tpl.Title, tpl.Body, tpl.Category, tpl.Position, tpl.CreatedAt); err != nil {
		return TextTemplate{}, err
	}
	return tpl, nil
}

func (r *SQLRepository) UpdateTextTemplate(ctx context.Context, viewer User, templateID string, input TextTemplateInput) (TextTemplate, error) {
	if !canManageTemplates(viewer) {
		return TextTemplate{}, ErrForbidden
	}
	if err := validateTextTemplateInput(input); err != nil {
		return TextTemplate{}, err
	}
	title := strings.TrimSpace(input.Title)
	body := strings.TrimSpace(input.Body)
	category := strings.TrimSpace(input.Category)
	if _, err := r.exec(ctx, `UPDATE text_templates SET title = ?, body = ?, category = ? WHERE id = ?`, title, body, category, templateID); err != nil {
		return TextTemplate{}, err
	}
	var updated TextTemplate
	if err := r.queryRow(ctx, `SELECT id, title, body, category, position, created_at FROM text_templates WHERE id = ?`, templateID).Scan(&updated.ID, &updated.Title, &updated.Body, &updated.Category, &updated.Position, &updated.CreatedAt); err != nil {
		return TextTemplate{}, err
	}
	return updated, nil
}

func (r *SQLRepository) DeleteTextTemplate(ctx context.Context, viewer User, templateID string) error {
	if !canManageTemplates(viewer) {
		return ErrForbidden
	}
	result, err := r.exec(ctx, `DELETE FROM text_templates WHERE id = ?`, templateID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SQLRepository) ReorderTextTemplate(ctx context.Context, viewer User, templateID, direction string) error {
	if !canManageTemplates(viewer) {
		return ErrForbidden
	}
	templates, err := r.ListTextTemplates(ctx)
	if err != nil {
		return err
	}
	index := -1
	for i, tpl := range templates {
		if tpl.ID == templateID {
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
	if swapWith < 0 || swapWith >= len(templates) {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := r.txExec(ctx, tx, `UPDATE text_templates SET position = ? WHERE id = ?`, templates[swapWith].Position, templates[index].ID); err != nil {
		return err
	}
	if _, err := r.txExec(ctx, tx, `UPDATE text_templates SET position = ? WHERE id = ?`, templates[index].Position, templates[swapWith].ID); err != nil {
		return err
	}
	return tx.Commit()
}
