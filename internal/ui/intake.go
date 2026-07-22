package ui

import (
	"strings"

	"university_agency/internal/dashboard"
)

// studentIntakeForm renders the client biodata form (mirrors the agency's
// Google Form) — client fills/edits it themselves anytime, no lock state.
func studentIntakeForm(vm dashboard.ViewModel) string {
	form := vm.IntakeForm
	notice := ""
	if vm.HasIntakeForm {
		notice = `<div class="auth-notice">Formulir tersimpan, terakhir diperbarui ` + dateLabel(form.UpdatedAt) + `.</div>`
	}
	return `<div class="panel">
  <div class="panel-head"><div><h2 class="text-lg font-semibold">Formulir Data Diri</h2><p class="text-sm text-slate-500">Lengkapi biodata untuk keperluan aplikasi kampus dan visa. Bisa diubah kapan saja.</p></div></div>
  ` + notice + `
  <form method="post" action="/intake/save" class="student-upload-form intake-form">` + intakeFormFields(form) + `
    <button class="primary-button" type="submit">Simpan Formulir</button>
  </form>
</div>`
}

// intakeFormFields renders the shared ~20-field biodata block used both by
// the client's own form and by owner/staff editing a client's form on their
// behalf (e.g. to add data the client missed or correct a typo).
func intakeFormFields(form dashboard.ClientIntakeForm) string {
	genderOption := func(value, label string) string {
		selected := ""
		if form.Gender == value {
			selected = " selected"
		}
		return `<option value="` + attr(value) + `"` + selected + `>` + esc(label) + `</option>`
	}
	return `
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
`
}

// ownerIntakeTable renders the live aggregated table of every client for
// owner/staff, with an Excel export action and an per-client edit action so
// staff can fill in missing biodata or correct it later — not just view it.
func ownerIntakeTable(vm dashboard.ViewModel) string {
	path := "/" + vm.Role.String() + "/intake"

	if vm.EditClientID != "" {
		return intakeEditPanel(vm, path)
	}

	formsByClient := make(map[string]dashboard.ClientIntakeForm, len(vm.IntakeForms))
	for _, form := range vm.IntakeForms {
		formsByClient[form.ClientID] = form
	}

	var rows strings.Builder
	for i, client := range vm.Clients {
		form, filled := formsByClient[client.ID]
		statusLabel := "Belum Diisi"
		statusClass := "slate"
		updated := "-"
		if filled {
			statusLabel = "Terisi"
			statusClass = "green"
			updated = dateLabel(form.UpdatedAt)
		}
		editHref := path + "?edit=" + attr(client.ID)
		rows.WriteString(`<tr><td>` + intText(i+1) + `</td><td><strong>` + esc(client.Name) + `</strong><span>` + esc(client.Email) + `</span></td><td>` + esc(form.FullNameEn) + `</td><td>` + esc(form.Gender) + `</td><td>` + esc(form.PassportNumber) + `</td><td>` + esc(form.SchoolName) + `</td><td><span class="status ` + statusClass + `">` + statusLabel + `</span></td><td>` + updated + `</td><td><a class="icon-action" href="` + attr(editHref) + `" hx-get="` + attr(editHref) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Isi/Edit</a></td></tr>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<tr><td colspan="9">Belum ada client.</td></tr>`)
	}
	return `<div class="panel table-panel">
  <div class="panel-head"><div><h2 class="text-lg font-semibold">Data Client / Formulir</h2><p class="text-sm text-slate-500">Terisi otomatis saat client mengisi formulir mereka sendiri, dan bisa diisi/diedit staff bila ada data yang kurang.</p></div><a class="primary-button" href="` + attr(path) + `/export">Export Excel</a></div>
  <table class="data-table"><thead><tr><th>No.</th><th>Client</th><th>Nama Lengkap (EN)</th><th>Gender</th><th>No. Paspor</th><th>Sekolah/Institusi</th><th>Status</th><th>Diperbarui</th><th>Aksi</th></tr></thead><tbody>` + rows.String() + `</tbody></table>
</div>`
}

// intakeEditPanel lets owner/staff fill in or correct a specific client's
// biodata form, prefilled from their existing submission if any.
func intakeEditPanel(vm dashboard.ViewModel, basePath string) string {
	var target dashboard.ClientProfile
	found := false
	for _, client := range vm.Clients {
		if client.ID == vm.EditClientID {
			target = client
			found = true
			break
		}
	}
	if !found {
		return `<div class="panel"><p class="empty-note">Client tidak ditemukan.</p><a class="outline-button mt-3" href="` + attr(basePath) + `" hx-get="` + attr(basePath) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Kembali</a></div>`
	}
	form := dashboard.ClientIntakeForm{ClientID: target.ID, ClientName: target.Name, Email: target.Email}
	for _, existing := range vm.IntakeForms {
		if existing.ClientID == target.ID {
			form = existing
			break
		}
	}
	return `<div class="panel">
  <div class="panel-head"><div><h2 class="text-lg font-semibold">Formulir Data Diri — ` + esc(target.Name) + `</h2><p class="text-sm text-slate-500">Isi atau perbarui biodata client ini bila ada data yang kurang.</p></div><a class="outline-button" href="` + attr(basePath) + `" hx-get="` + attr(basePath) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Batal</a></div>
  <form method="post" action="/intake/save" class="student-upload-form intake-form">
    <input type="hidden" name="client_id" value="` + attr(target.ID) + `">` + intakeFormFields(form) + `
    <button class="primary-button" type="submit">Simpan Formulir</button>
  </form>
</div>`
}
