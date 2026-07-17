package ui

import (
	"strings"

	"university_agency/internal/dashboard"
)

// institutionsPanel renders the "Direktori Institusi" WhatsApp contact
// directory — partner institutions (translators, legalization offices, etc.)
// owner/staff can manage and click-to-chat for quick follow-up.
func institutionsPanel(vm dashboard.ViewModel) string {
	path := institutionsBasePath(vm.Role)
	toggleLabel := "Kelola Kontak"
	toggleTarget := path + "?manage=1"
	if vm.ShowStageManager {
		toggleLabel = "Tutup Pengaturan Kontak"
		toggleTarget = path
	}
	toolbar := `<div class="toolbar-row"><div><h2 class="text-lg font-semibold">Direktori Institusi</h2><p class="text-sm text-slate-500">Kontak lembaga mitra untuk follow up cepat lewat WhatsApp.</p></div><div class="toolbar-actions"><a class="outline-button" href="` + attr(toggleTarget) + `" hx-get="` + attr(toggleTarget) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">` + toolIconHTML("settings-slider") + ` ` + toggleLabel + `</a></div></div>`

	managePanel := ""
	if vm.ShowStageManager {
		managePanel = institutionManagerPanel(vm.InstitutionContacts)
	}

	var rows strings.Builder
	for _, contact := range vm.InstitutionContacts {
		category := contact.Category
		if category == "" {
			category = "Umum"
		}
		waHref := waLink(contact.Phone, "Halo "+contact.Name+", saya dari "+vm.AppName+" ingin follow up terkait proses client.")
		rows.WriteString(`<tr><td><strong>` + esc(contact.Name) + `</strong><span>` + esc(contact.Notes) + `</span></td><td>` + esc(category) + `</td><td>` + esc(contact.Phone) + `</td><td><a class="primary-button small" href="` + attr(waHref) + `" target="_blank" rel="noopener">Chat WhatsApp</a></td></tr>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<tr><td colspan="4">Belum ada kontak institusi. Klik "Kelola Kontak" untuk menambahkan.</td></tr>`)
	}

	table := `<div class="panel table-panel"><table class="data-table"><thead><tr><th>Nama</th><th>Kategori</th><th>WhatsApp</th><th>Aksi</th></tr></thead><tbody>` + rows.String() + `</tbody></table></div>`

	return `<div class="space-y-5">` + toolbar + managePanel + table + `</div>`
}

func institutionManagerPanel(contacts []dashboard.InstitutionContact) string {
	var rows strings.Builder
	for i, contact := range contacts {
		upAttr := ""
		if i == 0 {
			upAttr = " disabled"
		}
		downAttr := ""
		if i == len(contacts)-1 {
			downAttr = " disabled"
		}
		confirmMsg := "Hapus kontak \"" + contact.Name + "\"?"
		rows.WriteString(`<div class="package-manager-row">`)
		rows.WriteString(`<form method="post" action="/institutions/` + attr(contact.ID) + `/update" class="package-edit-form">`)
		rows.WriteString(`<div class="package-edit-grid">`)
		rows.WriteString(`<label>Nama<input name="name" value="` + attr(contact.Name) + `" required maxlength="191"></label>`)
		rows.WriteString(`<label>Kategori<input name="category" value="` + attr(contact.Category) + `" maxlength="96" placeholder="Penerjemah / Legalisir / Notaris"></label>`)
		rows.WriteString(`<label>No. WhatsApp<input name="phone" value="` + attr(contact.Phone) + `" required maxlength="64" placeholder="08xxxxxxxxxx"></label>`)
		rows.WriteString(`</div>`)
		rows.WriteString(`<label>Catatan<textarea name="notes" rows="2" maxlength="300">` + esc(contact.Notes) + `</textarea></label>`)
		rows.WriteString(`<div class="package-edit-actions"><button class="primary-button small" type="submit">Simpan</button></div>`)
		rows.WriteString(`</form>`)
		rows.WriteString(`<div class="stage-manager-actions">`)
		rows.WriteString(`<form method="post" action="/institutions/` + attr(contact.ID) + `/move" class="inline-form"><input type="hidden" name="direction" value="up"><button type="submit"` + upAttr + ` title="Naikkan urutan" aria-label="Naikkan urutan">` + svgIcon("arrow-up") + `</button></form>`)
		rows.WriteString(`<form method="post" action="/institutions/` + attr(contact.ID) + `/move" class="inline-form"><input type="hidden" name="direction" value="down"><button type="submit"` + downAttr + ` title="Turunkan urutan" aria-label="Turunkan urutan">` + svgIcon("arrow-down") + `</button></form>`)
		rows.WriteString(`<form method="post" action="/institutions/` + attr(contact.ID) + `/delete" class="inline-form" data-confirm="` + attr(confirmMsg) + `"><button type="submit" class="stage-delete" title="Hapus kontak" aria-label="Hapus kontak">` + svgIcon("close") + `</button></form>`)
		rows.WriteString(`</div></div>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<p class="empty-note">Belum ada kontak. Tambahkan kontak pertama di bawah.</p>`)
	}
	addForm := `<form method="post" action="/institutions/create" class="package-edit-form mt-2"><div class="package-edit-grid"><label>Nama<input name="name" required maxlength="191" placeholder="Contoh: Penerjemah Tersumpah"></label><label>Kategori<input name="category" maxlength="96" placeholder="Penerjemah / Legalisir / Notaris"></label><label>No. WhatsApp<input name="phone" required maxlength="64" placeholder="08xxxxxxxxxx"></label></div><label>Catatan<textarea name="notes" rows="2" maxlength="300" placeholder="Layanan yang ditangani"></textarea></label><div class="package-edit-actions"><button class="primary-button" type="submit">+ Tambah Kontak</button></div></form>`
	return `<div class="panel stage-manager"><h3 class="text-sm font-semibold mb-1">Kelola Direktori Institusi</h3><p class="text-xs text-slate-500 mb-3">Tambah, ubah, urutkan, atau hapus kontak lembaga mitra.</p>` + rows.String() + addForm + `</div>`
}

func institutionsBasePath(role dashboard.Role) string {
	return "/" + role.String() + "/institutions"
}
