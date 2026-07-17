package dashboard

import "context"

func (r *SQLRepository) ListStaff(ctx context.Context) ([]User, error) {
	rows, err := r.query(ctx, `SELECT id, username, name, email, phone, role, password_hash, active, created_at FROM users WHERE role = ? AND active = ? ORDER BY name ASC`, RoleStaff, true)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var staff []User
	for rows.Next() {
		user, err := r.scanUser(rows)
		if err != nil {
			return nil, err
		}
		staff = append(staff, user)
	}
	return staff, rows.Err()
}

func (r *SQLRepository) findStaffByID(ctx context.Context, id string) (User, error) {
	user, err := r.scanUser(r.queryRow(ctx, `SELECT id, username, name, email, phone, role, password_hash, active, created_at FROM users WHERE id = ? AND role = ? AND active = ?`, id, RoleStaff, true))
	if err != nil {
		return User{}, err
	}
	return user, nil
}
