package ui

import (
	"strings"

	"university_agency/internal/dashboard"
)

// studentIntakeForm renders the client biodata form (mirrors the agency's
// Google Form) — client fills/edits it themselves anytime, no lock state.
func studentIntakeForm(vm dashboard.ViewModel) string {
	form := vm.IntakeForm
	genderOption := func(value, label string) string {
		selected := ""
		if form.Gender == value {
			selected = " selected"
		}
		return `<option value="` + attr(value) + `"` + selected + `>` + esc(label) + `</option>`
	}
	notice := ""
	if vm.HasIntakeForm {
		notice = `<div class="auth-notice">Formulir tersimpan, terakhir diperbarui ` + dateLabel(form.UpdatedAt) + `.</div>`
	}
	return `<div class="panel">
  <div class="panel-head"><div><h2 class="text-lg font-semibold">Formulir Data Diri</h2><p class="text-sm text-slate-500">Lengkapi biodata untuk keperluan aplikasi kampus dan visa. Bisa diubah kapan saja.</p></div></div>
  ` + notice + `
  <form method="post" action="/student/intake/save" class="student-upload-form intake-form">
    <p class="intake-section-title">Data Pribadi</p>
    <label class="span-2">Email<input name="email" type="email" required value="` + attr(form.Email) + `"></label>
    <label class="span-2">Nama Lengkap (Bahasa Inggris)<input name="full_name_en" required value="` + attr(form.FullNameEn) + `"></label>
    <label>Gender<select name="gender"><option value="">Pilih</option>` + genderOption("Male", "Male") + genderOption("Female", "Female") + `</select></label>
    <label>Tanggal Lahir<input name="date_of_birth" placeholder="DD/MM/YYYY" value="` + attr(form.DateOfBirth) + `"></label>
    <label>Tempat Lahir<input name="place_of_birth" value="` + attr(form.PlaceOfBirth) + `"></label>
    <label>Nomor Paspor<input name="passport_number" value="` + attr(form.PassportNumber) + `"></label>
    <label>No. Telepon<input name="phone_number" value="` + attr(form.PhoneNumber) + `"></label>
    <label>Kode Pos<input name="postal_code" value="` + attr(form.PostalCode) + `"></label>
    <label class="span-2">Alamat<input name="address" value="` + attr(form.Address) + `"></label>
    <label class="span-2">Social Media (Instagram)<input name="social_media_ig" value="` + attr(form.SocialMediaIG) + `"></label>

    <p class="intake-section-title">Data Orang Tua</p>
    <label>Nama Ayah (Bahasa Inggris)<input name="father_name" value="` + attr(form.FatherName) + `"></label>
    <label>Tanggal Lahir Ayah<input name="father_dob" placeholder="DD/MM/YYYY" value="` + attr(form.FatherDOB) + `"></label>
    <label>No. Telepon Ayah<input name="father_phone" value="` + attr(form.FatherPhone) + `"></label>
    <label>Nama Ibu (Bahasa Inggris)<input name="mother_name" value="` + attr(form.MotherName) + `"></label>
    <label>Tanggal Lahir Ibu<input name="mother_dob" placeholder="DD/MM/YYYY" value="` + attr(form.MotherDOB) + `"></label>
    <label>No. Telepon Ibu<input name="mother_phone" value="` + attr(form.MotherPhone) + `"></label>

    <p class="intake-section-title">Data Sekolah</p>
    <label class="span-2">Nama Sekolah (SMA/K) / Institusi<input name="school_name" value="` + attr(form.SchoolName) + `"></label>
    <label>Lokasi Sekolah<input name="school_location" value="` + attr(form.SchoolLocation) + `"></label>
    <label>Tanggal Masuk<input name="dates_enrolled" placeholder="DD/MM/YYYY" value="` + attr(form.DatesEnrolled) + `"></label>
    <label>Tanggal Lulus<input name="dates_graduate" placeholder="DD/MM/YYYY" value="` + attr(form.DatesGraduate) + `"></label>

    <button class="primary-button" type="submit">Simpan Formulir</button>
  </form>
</div>`
}

// ownerIntakeTable renders the live aggregated table of all client
// submissions for owner/staff, with an Excel export action.
func ownerIntakeTable(vm dashboard.ViewModel) string {
	path := "/" + vm.Role.String() + "/intake"
	var rows strings.Builder
	for i, form := range vm.IntakeForms {
		rows.WriteString(`<tr><td>` + intText(i+1) + `</td><td><strong>` + esc(form.ClientName) + `</strong><span>` + esc(form.Email) + `</span></td><td>` + esc(form.FullNameEn) + `</td><td>` + esc(form.Gender) + `</td><td>` + esc(form.PassportNumber) + `</td><td>` + esc(form.SchoolName) + `</td><td>` + dateLabel(form.UpdatedAt) + `</td></tr>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<tr><td colspan="7">Belum ada client yang mengisi formulir.</td></tr>`)
	}
	return `<div class="panel table-panel">
  <div class="panel-head"><div><h2 class="text-lg font-semibold">Data Client / Formulir</h2><p class="text-sm text-slate-500">Terisi otomatis saat client mengisi formulir mereka sendiri.</p></div><a class="primary-button" href="` + attr(path) + `/export">Export Excel</a></div>
  <table class="data-table"><thead><tr><th>No.</th><th>Client</th><th>Nama Lengkap (EN)</th><th>Gender</th><th>No. Paspor</th><th>Sekolah/Institusi</th><th>Diperbarui</th></tr></thead><tbody>` + rows.String() + `</tbody></table>
</div>`
}
