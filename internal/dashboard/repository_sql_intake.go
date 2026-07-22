package dashboard

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

func scanClientIntakeForm(scanner rowScanner) (ClientIntakeForm, error) {
	var form ClientIntakeForm
	if err := scanner.Scan(&form.ID, &form.ClientID, &form.ClientName, &form.Email, &form.FullNameEn, &form.Gender, &form.DateOfBirth, &form.PlaceOfBirth, &form.PassportNumber, &form.PhoneNumber, &form.Address, &form.PostalCode, &form.FatherName, &form.FatherDOB, &form.FatherPhone, &form.MotherName, &form.MotherDOB, &form.MotherPhone, &form.SchoolName, &form.SchoolLocation, &form.DatesEnrolled, &form.DatesGraduate, &form.SocialMediaIG, &form.SubmittedAt, &form.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ClientIntakeForm{}, ErrNotFound
		}
		return ClientIntakeForm{}, err
	}
	return form, nil
}

const intakeFormSelect = `SELECT f.id, f.client_id, COALESCE(c.name, ''), f.email, f.full_name_en, f.gender, f.date_of_birth, f.place_of_birth, f.passport_number, f.phone_number, f.address, f.postal_code, f.father_name, f.father_dob, f.father_phone, f.mother_name, f.mother_dob, f.mother_phone, f.school_name, f.school_location, f.dates_enrolled, f.dates_graduate, f.social_media_ig, f.submitted_at, f.updated_at FROM client_intake_forms f LEFT JOIN clients c ON c.id = f.client_id`

// ListClientIntakeForms returns intake forms scoped to what the viewer may
// see — the student's own single form, staff's PIC clients, or all for owner.
func (r *SQLRepository) ListClientIntakeForms(ctx context.Context, viewer User, viewRole Role) ([]ClientIntakeForm, error) {
	clients, err := r.ListClients(ctx, viewer, viewRole)
	if err != nil {
		return nil, err
	}
	visible := make(map[string]bool, len(clients))
	for _, client := range clients {
		visible[client.ID] = true
	}

	rows, err := r.query(ctx, intakeFormSelect+` ORDER BY f.updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var forms []ClientIntakeForm
	for rows.Next() {
		form, err := scanClientIntakeForm(rows)
		if err != nil {
			return nil, err
		}
		if visible[form.ClientID] {
			forms = append(forms, form)
		}
	}
	return forms, rows.Err()
}

func (r *SQLRepository) findIntakeFormByClientID(ctx context.Context, clientID string) (ClientIntakeForm, error) {
	form, err := scanClientIntakeForm(r.queryRow(ctx, intakeFormSelect+` WHERE f.client_id = ?`, clientID))
	if err != nil {
		return ClientIntakeForm{}, err
	}
	return form, nil
}

func validateClientIntakeFormInput(input ClientIntakeFormInput) error {
	if strings.TrimSpace(input.Email) == "" || strings.TrimSpace(input.FullNameEn) == "" {
		return ErrInvalidInput
	}
	return nil
}

// SaveClientIntakeForm upserts a client's biodata by find-then-update-or-insert
// (no ON CONFLICT, to stay dialect-agnostic). A student always saves their own
// single form; owner/staff can save on a specific client's behalf (e.g. to
// add missing fields or fix a typo), gated the same way as other client-scoped
// writes — owner sees everyone, staff only their own PIC clients.
func (r *SQLRepository) SaveClientIntakeForm(ctx context.Context, viewer User, clientID string, input ClientIntakeFormInput) (ClientIntakeForm, error) {
	if viewer.Role != RoleStudent && viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return ClientIntakeForm{}, ErrForbidden
	}
	if err := validateClientIntakeFormInput(input); err != nil {
		return ClientIntakeForm{}, err
	}
	var client ClientProfile
	if viewer.Role == RoleStudent {
		clients, err := r.ListClients(ctx, viewer, RoleStudent)
		if err != nil {
			return ClientIntakeForm{}, err
		}
		if len(clients) == 0 {
			return ClientIntakeForm{}, ErrForbidden
		}
		client = clients[0]
	} else {
		found, err := r.findClientByID(ctx, clientID)
		if err != nil {
			return ClientIntakeForm{}, err
		}
		if viewer.Role == RoleStaff && found.PICStaffID != viewer.ID {
			return ClientIntakeForm{}, ErrForbidden
		}
		client = found
	}
	now := time.Now()

	existing, err := r.findIntakeFormByClientID(ctx, client.ID)
	if err == nil {
		if _, err := r.exec(ctx, `UPDATE client_intake_forms SET email = ?, full_name_en = ?, gender = ?, date_of_birth = ?, place_of_birth = ?, passport_number = ?, phone_number = ?, address = ?, postal_code = ?, father_name = ?, father_dob = ?, father_phone = ?, mother_name = ?, mother_dob = ?, mother_phone = ?, school_name = ?, school_location = ?, dates_enrolled = ?, dates_graduate = ?, social_media_ig = ?, updated_at = ? WHERE id = ?`,
			strings.TrimSpace(input.Email), strings.TrimSpace(input.FullNameEn), strings.TrimSpace(input.Gender), strings.TrimSpace(input.DateOfBirth), strings.TrimSpace(input.PlaceOfBirth), strings.TrimSpace(input.PassportNumber), strings.TrimSpace(input.PhoneNumber), strings.TrimSpace(input.Address), strings.TrimSpace(input.PostalCode), strings.TrimSpace(input.FatherName), strings.TrimSpace(input.FatherDOB), strings.TrimSpace(input.FatherPhone), strings.TrimSpace(input.MotherName), strings.TrimSpace(input.MotherDOB), strings.TrimSpace(input.MotherPhone), strings.TrimSpace(input.SchoolName), strings.TrimSpace(input.SchoolLocation), strings.TrimSpace(input.DatesEnrolled), strings.TrimSpace(input.DatesGraduate), strings.TrimSpace(input.SocialMediaIG), now, existing.ID); err != nil {
			return ClientIntakeForm{}, err
		}
		return r.findIntakeFormByClientID(ctx, client.ID)
	}
	if !errors.Is(err, ErrNotFound) {
		return ClientIntakeForm{}, err
	}

	id := newID("intake")
	if _, err := r.exec(ctx, `INSERT INTO client_intake_forms (id, client_id, email, full_name_en, gender, date_of_birth, place_of_birth, passport_number, phone_number, address, postal_code, father_name, father_dob, father_phone, mother_name, mother_dob, mother_phone, school_name, school_location, dates_enrolled, dates_graduate, social_media_ig, submitted_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, client.ID, strings.TrimSpace(input.Email), strings.TrimSpace(input.FullNameEn), strings.TrimSpace(input.Gender), strings.TrimSpace(input.DateOfBirth), strings.TrimSpace(input.PlaceOfBirth), strings.TrimSpace(input.PassportNumber), strings.TrimSpace(input.PhoneNumber), strings.TrimSpace(input.Address), strings.TrimSpace(input.PostalCode), strings.TrimSpace(input.FatherName), strings.TrimSpace(input.FatherDOB), strings.TrimSpace(input.FatherPhone), strings.TrimSpace(input.MotherName), strings.TrimSpace(input.MotherDOB), strings.TrimSpace(input.MotherPhone), strings.TrimSpace(input.SchoolName), strings.TrimSpace(input.SchoolLocation), strings.TrimSpace(input.DatesEnrolled), strings.TrimSpace(input.DatesGraduate), strings.TrimSpace(input.SocialMediaIG), now, now); err != nil {
		return ClientIntakeForm{}, err
	}
	return r.findIntakeFormByClientID(ctx, client.ID)
}
