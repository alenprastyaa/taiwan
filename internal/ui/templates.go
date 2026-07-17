package ui

import (
	"strings"

	"university_agency/internal/dashboard"
)

// templatesPanel renders the "Template Teks" library — reusable canned
// replies staff/owner can copy-paste to clients instead of retyping the
// same answers. Data lives in vm.TextTemplates (owner/staff editable table),
// not hardcoded copy.
func templatesPanel(vm dashboard.ViewModel) string {
	path := templatesBasePath(vm.Role)
	toggleLabel := "Kelola Template"
	toggleTarget := path + "?manage=1"
	if vm.ShowStageManager {
		toggleLabel = "Tutup Pengaturan Template"
		toggleTarget = path
	}
	toolbar := `<div class="toolbar-row"><div><h2 class="text-lg font-semibold">Template Teks</h2><p class="text-sm text-slate-500">Simpan jawaban siap pakai untuk pertanyaan client yang sering berulang.</p></div><div class="toolbar-actions"><a class="outline-button" href="` + attr(toggleTarget) + `" hx-get="` + attr(toggleTarget) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">` + toolIconHTML("settings-slider") + ` ` + toggleLabel + `</a></div></div>`

	managePanel := ""
	if vm.ShowStageManager {
		managePanel = templateManagerPanel(vm.TextTemplates)
	}

	var cards strings.Builder
	if len(vm.TextTemplates) == 0 {
		cards.WriteString(`<p class="empty-note">Belum ada template. Klik "Kelola Template" untuk menambahkan.</p>`)
	}
	for _, tpl := range vm.TextTemplates {
		category := tpl.Category
		if category == "" {
			category = "Umum"
		}
		cards.WriteString(`<article class="panel service-card"><span class="service-badge bg-violet-100 text-violet-700">` + esc(category) + `</span><h2>` + esc(tpl.Title) + `</h2><p>` + esc(tpl.Body) + `</p><button type="button" class="outline-button small" data-copy-text="` + attr(tpl.Body) + `">Salin Teks</button></article>`)
	}

	return `<div class="space-y-5">` + toolbar + managePanel + `<div class="grid gap-5 xl:grid-cols-3">` + cards.String() + `</div></div>`
}

func templateManagerPanel(templates []dashboard.TextTemplate) string {
	var rows strings.Builder
	for i, tpl := range templates {
		upAttr := ""
		if i == 0 {
			upAttr = " disabled"
		}
		downAttr := ""
		if i == len(templates)-1 {
			downAttr = " disabled"
		}
		confirmMsg := "Hapus template \"" + tpl.Title + "\"?"
		rows.WriteString(`<div class="package-manager-row">`)
		rows.WriteString(`<form method="post" action="/templates/` + attr(tpl.ID) + `/update" class="package-edit-form">`)
		rows.WriteString(`<div class="package-edit-grid">`)
		rows.WriteString(`<label>Judul<input name="title" value="` + attr(tpl.Title) + `" required maxlength="120"></label>`)
		rows.WriteString(`<label>Kategori<input name="category" value="` + attr(tpl.Category) + `" maxlength="60" placeholder="Pembayaran / Dokumen / Timeline"></label>`)
		rows.WriteString(`</div>`)
		rows.WriteString(`<label>Isi Teks<textarea name="body" rows="3" maxlength="1000" required>` + esc(tpl.Body) + `</textarea></label>`)
		rows.WriteString(`<div class="package-edit-actions"><button class="primary-button small" type="submit">Simpan</button></div>`)
		rows.WriteString(`</form>`)
		rows.WriteString(`<div class="stage-manager-actions">`)
		rows.WriteString(`<form method="post" action="/templates/` + attr(tpl.ID) + `/move" class="inline-form"><input type="hidden" name="direction" value="up"><button type="submit"` + upAttr + ` title="Naikkan urutan" aria-label="Naikkan urutan">` + svgIcon("arrow-up") + `</button></form>`)
		rows.WriteString(`<form method="post" action="/templates/` + attr(tpl.ID) + `/move" class="inline-form"><input type="hidden" name="direction" value="down"><button type="submit"` + downAttr + ` title="Turunkan urutan" aria-label="Turunkan urutan">` + svgIcon("arrow-down") + `</button></form>`)
		rows.WriteString(`<form method="post" action="/templates/` + attr(tpl.ID) + `/delete" class="inline-form" data-confirm="` + attr(confirmMsg) + `"><button type="submit" class="stage-delete" title="Hapus template" aria-label="Hapus template">` + svgIcon("close") + `</button></form>`)
		rows.WriteString(`</div></div>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<p class="empty-note">Belum ada template. Tambahkan template pertama di bawah.</p>`)
	}
	addForm := `<form method="post" action="/templates/create" class="package-edit-form mt-2"><div class="package-edit-grid"><label>Judul<input name="title" required maxlength="120" placeholder="Contoh: Cara Pembayaran"></label><label>Kategori<input name="category" maxlength="60" placeholder="Pembayaran / Dokumen / Timeline"></label></div><label>Isi Teks<textarea name="body" rows="3" maxlength="1000" required placeholder="Isi jawaban siap pakai"></textarea></label><div class="package-edit-actions"><button class="primary-button" type="submit">+ Tambah Template</button></div></form>`
	return `<div class="panel stage-manager"><h3 class="text-sm font-semibold mb-1">Kelola Template Teks</h3><p class="text-xs text-slate-500 mb-3">Tambah, ubah, urutkan, atau hapus template jawaban untuk client.</p>` + rows.String() + addForm + `</div>`
}

func templatesBasePath(role dashboard.Role) string {
	return "/" + role.String() + "/templates"
}
