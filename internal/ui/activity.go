package ui

import (
	"sort"
	"strings"

	"university_agency/internal/dashboard"
)

func activityActionLabel(actionType string) string {
	switch actionType {
	case "document_reviewed":
		return "Dokumen Direview"
	case "task_completed":
		return "Task Selesai"
	case "expense_recorded":
		return "Pengeluaran Dicatat"
	case "order_marked_paid":
		return "Invoice Lunas"
	case "client_created":
		return "Client Baru"
	case "manual_note":
		return "Catatan Manual"
	default:
		return "Aktivitas"
	}
}

// staffActivity renders the staff's own reverse-chronological activity feed
// (auto-logged from existing actions) plus a manual-note form, following the
// same ShowCreateForm/?new=1 inline-form pattern as staffTasks.
func staffActivity(vm dashboard.ViewModel) string {
	var rows strings.Builder
	for _, entry := range vm.ActivityLog {
		client := entry.ClientName
		if client == "" {
			client = "-"
		}
		rows.WriteString(`<tr><td>` + dateLabel(entry.CreatedAt) + `</td><td><span class="status">` + esc(activityActionLabel(entry.ActionType)) + `</span></td><td>` + esc(client) + `</td><td>` + esc(entry.Description) + `</td></tr>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<tr><td colspan="4">Belum ada aktivitas tercatat.</td></tr>`)
	}
	form := ""
	if vm.ShowCreateForm {
		form = `<form method="post" action="/activity/notes/create" class="student-upload-form"><label>Client (opsional)<select name="client_id"><option value="">Umum</option>` + clientSelectOptions(vm.Clients) + `</select></label><label>Catatan<input name="note" required placeholder="Contoh: Follow up client via WhatsApp"></label><button class="primary-button" type="submit">Simpan Catatan</button></form>`
	}
	return `<div class="panel table-panel"><div class="panel-head"><h2>Aktivitas Saya</h2>` + toggleFormButton(vm.ShowCreateForm, "+ Tambah Catatan", "/staff/activity") + `</div>` + form + `<table class="data-table"><thead><tr><th>Waktu</th><th>Jenis</th><th>Client</th><th>Keterangan</th></tr></thead><tbody>` + rows.String() + `</tbody></table></div>`
}

// ownerActivity renders a rollup of every staff member's activity, grouped
// by staff then aggregated in Go (matching this codebase's convention for
// date-based grouping), with a simple date filter.
func ownerActivity(vm dashboard.ViewModel) string {
	filterForm := `<form method="get" action="/owner/activity" hx-get="/owner/activity" hx-target="#app" hx-swap="outerHTML" hx-push-url="true" class="toolbar-row mb-4"><label>Tanggal<input type="date" name="date" value="` + attr(vm.FilterDate) + `"></label><button class="outline-button" type="submit">Terapkan</button><a class="outline-button" href="/owner/activity" hx-get="/owner/activity" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Reset</a></form>`

	grouped := make(map[string][]dashboard.ActivityLog)
	var staffOrder []string
	for _, entry := range vm.ActivityLog {
		name := entry.StaffName
		if name == "" {
			name = "Tidak diketahui"
		}
		if _, ok := grouped[name]; !ok {
			staffOrder = append(staffOrder, name)
		}
		grouped[name] = append(grouped[name], entry)
	}
	sort.Strings(staffOrder)

	var groups strings.Builder
	for _, name := range staffOrder {
		entries := grouped[name]
		var rows strings.Builder
		for _, entry := range entries {
			client := entry.ClientName
			if client == "" {
				client = "-"
			}
			rows.WriteString(`<tr><td>` + dateLabel(entry.CreatedAt) + `</td><td><span class="status">` + esc(activityActionLabel(entry.ActionType)) + `</span></td><td>` + esc(client) + `</td><td>` + esc(entry.Description) + `</td></tr>`)
		}
		groups.WriteString(`<div class="panel table-panel mt-4"><div class="panel-head"><h2>` + esc(name) + `</h2><span class="status">` + intText(len(entries)) + ` aktivitas</span></div><table class="data-table"><thead><tr><th>Waktu</th><th>Jenis</th><th>Client</th><th>Keterangan</th></tr></thead><tbody>` + rows.String() + `</tbody></table></div>`)
	}
	if groups.Len() == 0 {
		groups.WriteString(`<div class="panel empty-state"><span>Aktivitas</span><h2>Belum ada aktivitas</h2><p>Aktivitas staff akan muncul di sini secara otomatis.</p></div>`)
	}

	return `<div class="space-y-2"><div class="panel"><h2 class="text-lg font-semibold mb-2">Aktivitas Staff</h2><p class="text-sm text-slate-500 mb-3">Rangkuman otomatis semua staff, bisa difilter per tanggal.</p>` + filterForm + `</div>` + groups.String() + `</div>`
}
