package ui

import (
	"strings"

	"university_agency/internal/dashboard"
)

// staffManagementPanel renders the owner-only "Manajemen Staf" page — the
// only place staff login accounts can be created, edited, password-reset,
// or disabled. Accounts are soft-deleted (Active toggle) rather than hard
// deleted since staff IDs are referenced elsewhere (PIC on clients, tasks,
// expenses, activity log).
//
// Mirrors ownerIntakeTable's table/edit-panel split: reuses the generic
// vm.EditClientID (populated from the page's `?edit=` query param for any
// section, not just intake) to swap the whole panel into a single-staff
// edit form instead of adding a dedicated field for the same plumbing.
func staffManagementPanel(vm dashboard.ViewModel) string {
	path := "/owner/staff"
	if vm.EditClientID != "" {
		return staffEditPanel(vm, path)
	}

	toolbar := `<div class="panel-head"><div><h2 class="text-lg font-semibold">Manajemen Staf</h2><p class="text-sm text-slate-500">Kelola akun staf konsultan: tambah, ubah data, reset password, atau nonaktifkan akses.</p></div>` + toggleFormButton(vm.ShowCreateForm, "+ Tambah Staf", path) + `</div>`

	createPanel := ""
	if vm.ShowCreateForm {
		createForm := `<form method="post" action="` + path + `/create" class="package-edit-form mt-2"><div class="package-edit-grid"><label>Nama Lengkap<input name="name" required maxlength="191" placeholder="Nama staf"></label><label>Username<input name="username" required maxlength="64" placeholder="Untuk login staf"></label><label>Password<input name="password" type="password" minlength="8" required placeholder="Minimal 8 karakter"></label><label>Email<input name="email" type="email" maxlength="191" placeholder="staf@email.com"></label><label>No. WhatsApp<input name="phone" maxlength="64" placeholder="08xxxxxxxxxx"></label></div><div class="package-edit-actions"><button class="primary-button" type="submit">Simpan Staf</button></div></form>`
		createPanel = `<div class="panel stage-manager"><h3 class="text-sm font-semibold mb-1">Tambah Staf Baru</h3>` + createForm + `</div>`
	}

	var rows strings.Builder
	for i, staff := range vm.StaffAccounts {
		statusClass := "green"
		statusLabel := "Aktif"
		toggleLabel := "Nonaktifkan"
		if !staff.Active {
			statusClass = "red"
			statusLabel = "Nonaktif"
			toggleLabel = "Aktifkan"
		}
		email := staff.Email
		if email == "" {
			email = "-"
		}
		phone := staff.Phone
		if phone == "" {
			phone = "-"
		}
		editHref := path + "?edit=" + attr(staff.ID)
		resetConfirm := "Reset password " + staff.Name + "? Password baru akan digenerate dan ditampilkan sekali."
		toggleConfirm := toggleLabel + " akun " + staff.Name + "?"
		rows.WriteString(`<tr><td>` + intText(i+1) + `</td><td><strong>` + esc(staff.Name) + `</strong><span>` + esc(staff.Username) + `</span></td><td>` + esc(email) + `</td><td>` + esc(phone) + `</td><td><span class="status ` + statusClass + `">` + statusLabel + `</span></td><td>` + esc(dateLabel(staff.CreatedAt)) + `</td><td><a class="icon-action" href="` + attr(editHref) + `" hx-get="` + attr(editHref) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Edit</a> <form method="post" action="` + path + `/` + attr(staff.ID) + `/reset-password" class="inline-form" data-confirm="` + attr(resetConfirm) + `"><button type="submit" class="icon-action">Reset Password</button></form> <form method="post" action="` + path + `/` + attr(staff.ID) + `/toggle-active" class="inline-form" data-confirm="` + attr(toggleConfirm) + `"><button type="submit" class="icon-action">` + toggleLabel + `</button></form></td></tr>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<tr><td colspan="7">Belum ada akun staf. Klik "+ Tambah Staf" untuk menambahkan.</td></tr>`)
	}
	table := `<div class="panel table-panel"><table class="data-table"><thead><tr><th>No.</th><th>Nama &amp; Username</th><th>Email</th><th>No. WhatsApp</th><th>Status</th><th>Dibuat</th><th>Aksi</th></tr></thead><tbody>` + rows.String() + `</tbody></table></div>`

	return `<div class="space-y-5"><div class="panel">` + toolbar + createPanel + `</div>` + table + `</div>`
}

// staffEditPanel is the single-staff edit form shown in place of the table
// when vm.EditClientID matches a row's "Edit" link — see ownerIntakeTable /
// intakeEditPanel for the identical pattern applied to client biodata.
func staffEditPanel(vm dashboard.ViewModel, basePath string) string {
	var target dashboard.User
	found := false
	for _, staff := range vm.StaffAccounts {
		if staff.ID == vm.EditClientID {
			target = staff
			found = true
			break
		}
	}
	if !found {
		return `<div class="panel"><p class="empty-note">Akun staf tidak ditemukan.</p><a class="outline-button mt-3" href="` + attr(basePath) + `" hx-get="` + attr(basePath) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Kembali</a></div>`
	}
	return `<div class="panel">
  <div class="panel-head"><div><h2 class="text-lg font-semibold">Edit Staf — ` + esc(target.Name) + `</h2><p class="text-sm text-slate-500">Username: ` + esc(target.Username) + ` (tidak bisa diubah).</p></div><a class="outline-button" href="` + attr(basePath) + `" hx-get="` + attr(basePath) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Batal</a></div>
  <form method="post" action="` + basePath + `/` + attr(target.ID) + `/update" class="student-upload-form">
    <label>Nama Lengkap<input name="name" value="` + attr(target.Name) + `" required maxlength="191"></label>
    <label>Email<input name="email" type="email" value="` + attr(target.Email) + `" maxlength="191"></label>
    <label>No. WhatsApp<input name="phone" value="` + attr(target.Phone) + `" maxlength="64" placeholder="08xxxxxxxxxx"></label>
    <button class="primary-button" type="submit">Simpan Perubahan</button>
  </form>
</div>`
}
