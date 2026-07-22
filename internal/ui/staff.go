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
func staffManagementPanel(vm dashboard.ViewModel) string {
	path := "/owner/staff"
	toolbar := `<div class="toolbar-row"><div><h2 class="text-lg font-semibold">Manajemen Staf</h2><p class="text-sm text-slate-500">Kelola akun staf konsultan: tambah, ubah data, reset password, atau nonaktifkan akses.</p></div><div class="toolbar-actions">` + toggleFormButton(vm.ShowCreateForm, "+ Tambah Staf", path) + `</div></div>`

	createPanel := ""
	if vm.ShowCreateForm {
		createForm := `<form method="post" action="` + path + `/create" class="package-edit-form mt-2"><div class="package-edit-grid"><label>Nama Lengkap<input name="name" required maxlength="191" placeholder="Nama staf"></label><label>Username<input name="username" required maxlength="64" placeholder="Untuk login staf"></label><label>Password<input name="password" type="password" minlength="8" required placeholder="Minimal 8 karakter"></label><label>Email<input name="email" type="email" maxlength="191" placeholder="staf@email.com"></label><label>No. WhatsApp<input name="phone" maxlength="64" placeholder="08xxxxxxxxxx"></label></div><div class="package-edit-actions"><button class="primary-button" type="submit">Simpan Staf</button></div></form>`
		createPanel = `<div class="panel stage-manager"><h3 class="text-sm font-semibold mb-1">Tambah Staf Baru</h3>` + createForm + `</div>`
	}

	var rows strings.Builder
	for _, staff := range vm.StaffAccounts {
		statusClass := "green"
		statusLabel := "Aktif"
		toggleLabel := "Nonaktifkan"
		if !staff.Active {
			statusClass = "red"
			statusLabel = "Nonaktif"
			toggleLabel = "Aktifkan"
		}
		resetConfirm := "Reset password " + staff.Name + "? Password baru akan digenerate dan ditampilkan sekali."
		toggleConfirm := toggleLabel + " akun " + staff.Name + "?"
		rows.WriteString(`<div class="package-manager-row">`)
		rows.WriteString(`<form method="post" action="` + path + `/` + attr(staff.ID) + `/update" class="package-edit-form">`)
		rows.WriteString(`<div class="package-edit-grid">`)
		rows.WriteString(`<label>Nama<input name="name" value="` + attr(staff.Name) + `" required maxlength="191"></label>`)
		rows.WriteString(`<label>Username<input value="` + attr(staff.Username) + `" disabled></label>`)
		rows.WriteString(`<label>Email<input name="email" type="email" value="` + attr(staff.Email) + `" maxlength="191"></label>`)
		rows.WriteString(`<label>No. WhatsApp<input name="phone" value="` + attr(staff.Phone) + `" maxlength="64" placeholder="08xxxxxxxxxx"></label>`)
		rows.WriteString(`</div>`)
		rows.WriteString(`<div class="package-edit-actions"><span class="status ` + statusClass + `">` + statusLabel + `</span><button class="primary-button small" type="submit">Simpan</button></div>`)
		rows.WriteString(`</form>`)
		rows.WriteString(`<div class="stage-manager-actions">`)
		rows.WriteString(`<form method="post" action="` + path + `/` + attr(staff.ID) + `/reset-password" class="inline-form" data-confirm="` + attr(resetConfirm) + `"><button type="submit" class="icon-action">Reset Password</button></form>`)
		rows.WriteString(`<form method="post" action="` + path + `/` + attr(staff.ID) + `/toggle-active" class="inline-form" data-confirm="` + attr(toggleConfirm) + `"><button type="submit" class="icon-action">` + toggleLabel + `</button></form>`)
		rows.WriteString(`</div></div>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<p class="empty-note">Belum ada akun staf. Klik "+ Tambah Staf" untuk menambahkan.</p>`)
	}
	listPanel := `<div class="panel stage-manager"><h3 class="text-sm font-semibold mb-1">Daftar Akun Staf</h3><p class="text-xs text-slate-500 mb-3">Ubah data, reset password, atau nonaktifkan akses staf.</p>` + rows.String() + `</div>`

	return `<div class="space-y-5">` + toolbar + createPanel + listPanel + `</div>`
}
