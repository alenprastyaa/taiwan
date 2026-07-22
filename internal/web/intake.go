package web

import (
	"net/http"
	"net/url"

	"university_agency/internal/dashboard"

	"github.com/xuri/excelize/v2"
)

func (h Handler) saveClientIntakeForm(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleStudent && viewer.Role != dashboard.RoleOwner && viewer.Role != dashboard.RoleStaff {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	basePath := "/student/intake"
	if viewer.Role != dashboard.RoleStudent {
		basePath = "/" + currentPageRole(viewer.Role).String() + "/intake"
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, basePath, http.StatusSeeOther)
		return
	}
	clientID := r.FormValue("client_id")
	input := dashboard.ClientIntakeFormInput{
		Email:          r.FormValue("email"),
		FullNameEn:     r.FormValue("full_name_en"),
		Gender:         r.FormValue("gender"),
		DateOfBirth:    r.FormValue("date_of_birth"),
		PlaceOfBirth:   r.FormValue("place_of_birth"),
		PassportNumber: r.FormValue("passport_number"),
		PhoneNumber:    r.FormValue("phone_number"),
		Address:        r.FormValue("address"),
		PostalCode:     r.FormValue("postal_code"),
		FatherName:     r.FormValue("father_name"),
		FatherDOB:      r.FormValue("father_dob"),
		FatherPhone:    r.FormValue("father_phone"),
		MotherName:     r.FormValue("mother_name"),
		MotherDOB:      r.FormValue("mother_dob"),
		MotherPhone:    r.FormValue("mother_phone"),
		SchoolName:     r.FormValue("school_name"),
		SchoolLocation: r.FormValue("school_location"),
		DatesEnrolled:  r.FormValue("dates_enrolled"),
		DatesGraduate:  r.FormValue("dates_graduate"),
		SocialMediaIG:  r.FormValue("social_media_ig"),
	}
	_, err := h.store.SaveClientIntakeForm(r.Context(), viewer, clientID, input)
	notice := "Formulir berhasil disimpan."
	if err != nil {
		notice = "Gagal menyimpan formulir. Email dan nama lengkap wajib diisi."
	}
	http.Redirect(w, r, basePath+"?notice="+url.QueryEscape(notice), http.StatusSeeOther)
}

func (h Handler) exportClientIntakeForms(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer.Role != dashboard.RoleOwner && viewer.Role != dashboard.RoleStaff {
		http.Error(w, "akses ditolak", http.StatusForbidden)
		return
	}
	forms, err := h.store.ListClientIntakeForms(r.Context(), viewer, viewer.Role)
	if err != nil {
		http.Error(w, "gagal memuat data formulir", http.StatusInternalServerError)
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Formulir"
	f.SetSheetName(f.GetSheetName(0), sheet)

	headers := []string{
		"Nama Client", "Email", "Nama Lengkap (EN)", "Gender", "Tanggal Lahir", "Tempat Lahir",
		"Nomor Paspor", "No. Telepon", "Alamat", "Kode Pos",
		"Nama Ayah", "Tanggal Lahir Ayah", "No. Telepon Ayah",
		"Nama Ibu", "Tanggal Lahir Ibu", "No. Telepon Ibu",
		"Nama Sekolah/Institusi", "Lokasi Sekolah", "Tanggal Masuk", "Tanggal Lulus", "Instagram",
		"Terakhir Diperbarui",
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
	}

	for i, form := range forms {
		row := i + 2
		values := []any{
			form.ClientName, form.Email, form.FullNameEn, form.Gender, form.DateOfBirth, form.PlaceOfBirth,
			form.PassportNumber, form.PhoneNumber, form.Address, form.PostalCode,
			form.FatherName, form.FatherDOB, form.FatherPhone,
			form.MotherName, form.MotherDOB, form.MotherPhone,
			form.SchoolName, form.SchoolLocation, form.DatesEnrolled, form.DatesGraduate, form.SocialMediaIG,
			form.UpdatedAt.Format("2006-01-02 15:04"),
		}
		for col, value := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, row)
			f.SetCellValue(sheet, cell, value)
		}
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=data-client-formulir.xlsx")
	if err := f.Write(w); err != nil {
		h.logger.Error("write intake export", "error", err)
	}
}
