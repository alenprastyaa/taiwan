package ui

import (
	"context"
	"html/template"
	"io"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"university_agency/internal/dashboard"

	"github.com/a-h/templ"
)

const assetVersion = "20260717-button-fix"

func Document(vm dashboard.ViewModel) templ.Component {
	return templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		_, err := io.WriteString(w, documentHTML(vm))
		return err
	})
}

func App(vm dashboard.ViewModel) templ.Component {
	return templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		_, err := io.WriteString(w, appHTML(vm))
		return err
	})
}

func documentHTML(vm dashboard.ViewModel) string {
	canonical := vm.AppURL + vm.Path()
	description := vm.Description
	if description == "" {
		description = "Dashboard operasional Formora Taiwan untuk owner, staff, dan client."
	}

	var b strings.Builder
	b.WriteString(`<!doctype html><html lang="id"><head><meta charset="utf-8">`)
	b.WriteString(`<meta name="viewport" content="width=device-width, initial-scale=1">`)
	b.WriteString(`<title>`)
	b.WriteString(esc(vm.Title + " - " + vm.AppName))
	b.WriteString(`</title>`)
	b.WriteString(`<meta name="description" content="`)
	b.WriteString(attr(description))
	b.WriteString(`">`)
	b.WriteString(`<link rel="canonical" href="`)
	b.WriteString(attr(canonical))
	b.WriteString(`">`)
	b.WriteString(`<meta property="og:type" content="website">`)
	b.WriteString(`<meta property="og:title" content="`)
	b.WriteString(attr(vm.Title + " - " + vm.AppName))
	b.WriteString(`">`)
	b.WriteString(`<meta property="og:description" content="`)
	b.WriteString(attr(description))
	b.WriteString(`">`)
	b.WriteString(`<meta property="og:url" content="`)
	b.WriteString(attr(canonical))
	b.WriteString(`">`)
	b.WriteString(`<meta name="twitter:card" content="summary_large_image">`)
	b.WriteString(`<meta name="theme-color" content="#071225">`)
	b.WriteString(`<link rel="manifest" href="/manifest.webmanifest">`)
	b.WriteString(`<link rel="icon" href="/favicon.svg" type="image/svg+xml">`)
	b.WriteString(`<link rel="stylesheet" href="/assets/app.css?v=` + assetVersion + `">`)
	b.WriteString(`<link rel="stylesheet" href="/assets/icons.css?v=` + assetVersion + `">`)
	b.WriteString(`<script src="/assets/app.js?v=` + assetVersion + `" defer></script>`)
	b.WriteString(`</head><body>`)
	b.WriteString(appHTML(vm))
	b.WriteString(`</body></html>`)
	return b.String()
}

func appHTML(vm dashboard.ViewModel) string {
	var b strings.Builder
	b.WriteString(`<div id="app" data-page-title="`)
	b.WriteString(attr(vm.Title + " - " + vm.AppName))
	b.WriteString(`" class="min-h-screen bg-surface text-slate-900">`)
	b.WriteString(`<div class="lg:hidden fixed inset-x-0 top-0 z-40 flex h-16 items-center justify-between border-b border-slate-200 bg-white/95 px-4 backdrop-blur">`)
	b.WriteString(`<button type="button" class="mobile-menu-button tool-button" aria-label="Buka menu" data-sidebar-toggle>` + toolIconHTML("menu") + `</button>`)
	b.WriteString(`<div><p class="text-sm font-semibold text-slate-900">` + esc(vm.AppName) + `</p><p class="text-xs text-slate-500">` + esc(vm.RoleBadge) + ` portal</p></div>`)
	b.WriteString(`<div class="avatar-sm">` + esc(vm.UserInitials) + `</div>`)
	b.WriteString(`</div>`)
	b.WriteString(`<div class="app-layout">`)
	b.WriteString(sidebarHTML(vm))
	b.WriteString(`<main class="content-shell">`)
	b.WriteString(topbarHTML(vm))
	b.WriteString(`<section class="content-wrap">`)
	b.WriteString(contentHTML(vm))
	b.WriteString(`</section>`)
	b.WriteString(`</main></div>`)
	b.WriteString(`<button type="button" class="sidebar-scrim" aria-label="Tutup menu" data-sidebar-close></button>`)
	b.WriteString(`</div>`)
	return b.String()
}

func sidebarHTML(vm dashboard.ViewModel) string {
	var b strings.Builder
	b.WriteString(`<aside class="sidebar" data-sidebar>`)
	b.WriteString(`<div class="brand-block"><div class="brand-mark" aria-hidden="true"><span></span></div><div><p class="brand-title">FORMORA</p><p class="brand-subtitle">Taiwan</p><p class="brand-note">Study Abroad Taiwan</p></div></div>`)
	if vm.Role != dashboard.RoleStudent {
		b.WriteString(`<div class="px-4 pb-4"><span class="role-badge ` + attr(vm.RoleBadgeClass) + `">` + esc(vm.RoleBadge) + `</span></div>`)
	}
	b.WriteString(`<nav class="nav-list" aria-label="Navigasi utama">`)
	for _, item := range vm.Navigation {
		className := "nav-link"
		if item.Active {
			className += " nav-link-active"
		}
		b.WriteString(`<a class="` + className + `" href="` + attr(item.Href) + `" hx-get="` + attr(item.Href) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">`)
		b.WriteString(iconHTML(item.Icon))
		b.WriteString(`<span>` + esc(item.Label) + `</span>`)
		if item.Badge != "" {
			b.WriteString(`<span class="nav-badge">` + esc(item.Badge) + `</span>`)
		}
		b.WriteString(`</a>`)
	}
	b.WriteString(`</nav>`)
	b.WriteString(`<div class="sidebar-help"><p class="text-sm font-semibold text-white">Butuh Bantuan?</p><p class="mt-1 text-xs text-slate-400">Hubungi tim support kami.</p><a href="/` + attr(vm.Role.String()) + `/chat" class="help-button" hx-get="/` + attr(vm.Role.String()) + `/chat" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Hubungi Sekarang</a></div>`)
	b.WriteString(`<p class="sidebar-footnote">&copy; 2026 Formora Taiwan<br>v1.0.0</p>`)
	b.WriteString(`</aside>`)
	return b.String()
}

func topbarHTML(vm dashboard.ViewModel) string {
	clientsPath := "/" + vm.Role.String() + "/clients"

	var b strings.Builder
	b.WriteString(`<header class="topbar">`)
	b.WriteString(`<div class="min-w-0"><h1 class="page-title">` + esc(vm.Title) + `</h1><p class="page-subtitle">` + esc(vm.Subtitle) + `</p></div>`)
	b.WriteString(`<div class="topbar-actions">`)
	b.WriteString(`<div class="date-pill">` + toolIconHTML("calendar") + `<span>` + esc(vm.DateLabel) + `</span></div>`)
	if vm.Role != dashboard.RoleStudent {
		b.WriteString(`<form method="get" action="` + attr(clientsPath) + `" hx-get="` + attr(clientsPath) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true" class="search-box">` + toolIconHTML("search") + `<input type="search" name="q" placeholder="` + attr(vm.Search) + `" aria-label="Pencarian"></form>`)
	} else {
		b.WriteString(`<label class="search-box">` + toolIconHTML("search") + `<input type="search" placeholder="` + attr(vm.Search) + `" aria-label="Pencarian" disabled></label>`)
	}
	bellHref, mailHref := topbarLinks(vm)
	b.WriteString(`<a class="tool-button" aria-label="Notifikasi" title="Lihat pengingat" href="` + attr(bellHref) + `" hx-get="` + attr(bellHref) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">` + toolIconHTML("bell") + `</a>`)
	b.WriteString(`<a class="tool-button" aria-label="Pesan" title="Buka chat" href="` + attr(mailHref) + `" hx-get="` + attr(mailHref) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">` + toolIconHTML("mail") + `</a>`)
	b.WriteString(`<div class="user-chip"><div class="avatar-sm">` + esc(vm.UserInitials) + `</div><div class="hidden min-w-0 sm:block"><p>` + esc(vm.UserName) + `</p><span>` + esc(vm.UserRole) + `</span></div></div>`)
	b.WriteString(`<form method="post" action="/logout" class="logout-form"><button type="submit" class="outline-button small">Keluar</button></form>`)
	b.WriteString(`</div></header>`)
	return b.String()
}

func topbarLinks(vm dashboard.ViewModel) (bellHref, mailHref string) {
	switch vm.Role {
	case dashboard.RoleOwner:
		bellHref = "/owner/invoices"
	case dashboard.RoleStaff:
		bellHref = "/staff/tasks"
	default:
		bellHref = "/student/payments"
	}
	mailHref = "/" + vm.Role.String() + "/chat"
	return bellHref, mailHref
}

func contentHTML(vm dashboard.ViewModel) string {
	content := ""
	switch vm.Role {
	case dashboard.RoleOwner:
		content = ownerContent(vm)
	case dashboard.RoleStaff:
		content = staffContent(vm)
	case dashboard.RoleStudent:
		content = studentContent(vm)
	default:
		content = `<div class="panel">Dashboard tidak ditemukan.</div>`
	}
	if vm.Flash != "" {
		return `<div class="flash-banner">` + esc(vm.Flash) + `</div>` + content
	}
	return content
}

func ownerContent(vm dashboard.ViewModel) string {
	switch vm.Section {
	case dashboard.SectionFinance:
		return ownerFinance(vm)
	case dashboard.SectionClients:
		return clientsPanel(vm, true)
	case dashboard.SectionPipeline:
		return ownerPipeline(vm)
	case dashboard.SectionServices:
		return ownerServices(vm)
	case dashboard.SectionInvoices:
		return ownerInvoices(vm)
	case dashboard.SectionReports:
		return ownerReports(vm)
	case dashboard.SectionTemplates:
		return templatesPanel(vm)
	case dashboard.SectionInstitutions:
		return institutionsPanel(vm)
	case dashboard.SectionIntake:
		return ownerIntakeTable(vm)
	case dashboard.SectionActivity:
		return ownerActivity(vm)
	case dashboard.SectionLogistics:
		return staffLogistics(vm)
	case dashboard.SectionStaff:
		return staffManagementPanel(vm)
	case dashboard.SectionChat:
		return chatPanel(vm, "Chat Operasional")
	case dashboard.SectionSettings:
		return settingsPanel(vm, "Owner", true)
	default:
		return ownerDashboard(vm)
	}
}

func staffContent(vm dashboard.ViewModel) string {
	switch vm.Section {
	case dashboard.SectionClients:
		return staffClients(vm)
	case dashboard.SectionPipeline:
		return staffPipeline(vm)
	case dashboard.SectionTasks:
		return staffTasks(vm)
	case dashboard.SectionDocuments:
		return staffDocuments(vm)
	case dashboard.SectionExpenses:
		return staffExpenses(vm)
	case dashboard.SectionCalendar:
		return calendarPanel(vm, "Jadwal Staff")
	case dashboard.SectionTemplates:
		return templatesPanel(vm)
	case dashboard.SectionInstitutions:
		return institutionsPanel(vm)
	case dashboard.SectionIntake:
		return ownerIntakeTable(vm)
	case dashboard.SectionActivity:
		return staffActivity(vm)
	case dashboard.SectionLogistics:
		return staffLogistics(vm)
	case dashboard.SectionChat:
		return chatPanel(vm, "Chat Klien")
	case dashboard.SectionSettings:
		return settingsPanel(vm, "Staff", false)
	default:
		return staffDashboard(vm)
	}
}

func studentContent(vm dashboard.ViewModel) string {
	switch vm.Section {
	case dashboard.SectionProgress:
		return studentProgress(vm)
	case dashboard.SectionDocuments:
		return studentDocuments(vm)
	case dashboard.SectionIntake:
		return studentIntakeForm(vm)
	case dashboard.SectionAgreement:
		return studentAgreementPage(vm)
	case dashboard.SectionLogistics:
		return studentLogistics(vm)
	case dashboard.SectionPayments:
		return studentPayments(vm)
	case dashboard.SectionCalendar:
		return studentCalendar(vm)
	case dashboard.SectionChat:
		if !vm.StudentFeatureAccess {
			return lockedStudentFeature("Chat dibuka setelah pembayaran", "Kirim bukti pembayaran atau tunggu invoice ditandai lunas agar chat konsultan aktif.")
		}
		return chatPanel(vm, "Chat Konsultan")
	case dashboard.SectionSettings:
		return settingsPanel(vm, "Client", false)
	default:
		return studentDashboard(vm)
	}
}

func ownerDashboard(vm dashboard.ViewModel) string {
	revenue := vm.Snapshot.Revenue
	profit := vm.Snapshot.Profit
	activeClients := vm.Snapshot.ActiveClients
	openOrders := vm.Snapshot.OpenOrders
	completedOrders := vm.Snapshot.CompletedOrders
	unpaid := unpaidAmount(vm.Orders)
	return `
<div class="space-y-5">
  <section class="metric-grid metric-grid-owner">
    <article class="metric-card">` + metricIconHTML("revenue", "tone-violet") + `<div><p>Total Omzet (IDR)</p><strong>` + money(revenue) + `</strong><span class="trend-up">Dari order lunas</span></div></article>
    <article class="metric-card">` + metricIconHTML("profit", "tone-emerald") + `<div><p>Total Profit (IDR)</p><strong>` + money(profit) + `</strong><span class="trend-up">Omzet dikurangi pengeluaran</span></div></article>
    <article class="metric-card">` + metricIconHTML("users", "tone-blue") + `<div><p>Klien Aktif</p><strong>` + intText(activeClients) + `</strong><span class="trend-up">Data aktif di sistem</span></div></article>
    <article class="metric-card">` + metricIconHTML("orders", "tone-amber") + `<div><p>Order Dalam Proses</p><strong>` + intText(openOrders) + `</strong><span class="trend-up">Belum lunas/verifikasi</span></div></article>
    <article class="metric-card">` + metricIconHTML("check", "tone-teal") + `<div><p>Order Selesai</p><strong>` + intText(completedOrders) + `</strong><span class="trend-up">Sudah lunas</span></div></article>
  </section>

  <section class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(0,1.1fr)_minmax(0,1fr)]">
    <div class="panel">
      <div class="panel-head"><h2>Pendapatan per Kategori</h2></div>
      <div class="donut-wrap">
        <div class="donut donut-owner" aria-label="Diagram paket layanan"></div>
        <div class="legend-list">
          <div><span class="legend-dot bg-violet-600"></span><p>Order lunas</p><strong>` + money(revenue) + `</strong><em>` + intText(countPaidOrders(vm.Orders)) + ` order</em></div>
          <div><span class="legend-dot bg-blue-500"></span><p>Belum lunas</p><strong>` + money(unpaid) + `</strong><em>` + intText(countOpenOrders(vm.Orders)) + ` order</em></div>
          <div><span class="legend-dot bg-amber-500"></span><p>Pengeluaran</p><strong>` + money(vm.Snapshot.Expenses) + `</strong><em>DB</em></div>
        </div>
      </div>
      <div class="total-row"><span>Total omzet lunas</span><strong>` + money(revenue) + `</strong></div>
    </div>

    ` + ownerPipelineSummaryPanel(vm.Clients, vm.PipelineStages) + `

    <div class="panel">
      <div class="panel-head"><h2>Keuangan Bulan Ini</h2><a href="/owner/finance" hx-get="/owner/finance" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Lihat Detail</a></div>
      <div class="finance-mini">
        <div><span>Pemasukan</span><strong>` + money(revenue) + `</strong><div><i class="w-[80%]"></i></div><em>DB</em></div>
        <div><span>Pengeluaran</span><strong>` + money(vm.Snapshot.Expenses) + `</strong><div><i class="w-[73%]"></i></div><em>DB</em></div>
        <div><span>Profit</span><strong class="text-emerald-700">` + money(profit) + `</strong><div><i class="w-[72%]"></i></div><em>DB</em></div>
        <div><span>Piutang</span><strong>` + money(unpaid) + `</strong><div><i class="w-[75%]"></i></div><em>DB</em></div>
      </div>
    </div>
  </section>

  <section class="grid gap-5 xl:grid-cols-[minmax(0,1.3fr)_minmax(0,.9fr)]">
    ` + ownerRecentActivityPanel(vm.ActivityLog) + `
    ` + ownerReminderPanel(vm.Orders, vm.Documents) + `
  </section>
</div>`
}

// ownerPipelineSummaryPanel renders real client-per-stage counts instead of
// placeholder numbers, reusing each stage's configured tone color
// (toneHex) so it stays in sync with the pipeline board itself.
func ownerPipelineSummaryPanel(clients []dashboard.ClientProfile, stages []dashboard.PipelineStage) string {
	counts := make(map[string]int, len(stages))
	for _, client := range clients {
		counts[client.CurrentStage]++
	}
	total := len(clients)
	var bars strings.Builder
	for _, stage := range stages {
		count := counts[stage.Name]
		percent := 0
		if total > 0 {
			percent = count * 100 / total
		}
		bars.WriteString(`<div><span>` + esc(stage.Name) + `</span><div><i style="width:` + intText(percent) + `%;background:` + toneHex(stage.Tone) + `"></i></div><b>` + intText(count) + `</b><em>` + intText(percent) + `%</em></div>`)
	}
	if len(stages) == 0 {
		bars.WriteString(`<p class="empty-note">Belum ada tahap pipeline yang dikonfigurasi.</p>`)
	}
	return `<div class="panel">
      <div class="panel-head"><h2>Pipeline Keseluruhan</h2><a href="/owner/pipeline" hx-get="/owner/pipeline" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Lihat Detail</a></div>
      <div class="pipeline-bars">` + bars.String() + `</div>
    </div>`
}

var activityAvatarTones = []string{"bg-blue-100 text-blue-700", "bg-emerald-100 text-emerald-700", "bg-violet-100 text-violet-700", "bg-amber-100 text-amber-700"}

// ownerRecentActivityPanel shows the real activity log (most recent first)
// instead of fabricated names, matching the feed already used on the
// dedicated Activity page.
func ownerRecentActivityPanel(entries []dashboard.ActivityLog) string {
	const maxShown = 5
	var rows strings.Builder
	shown := 0
	for _, entry := range entries {
		if shown >= maxShown {
			break
		}
		actor := entry.StaffName
		if actor == "" {
			actor = "Sistem"
		}
		tone := activityAvatarTones[shown%len(activityAvatarTones)]
		rows.WriteString(`<div><span class="avatar-mini ` + tone + `">` + esc(initials(actor)) + `</span><p><strong>` + esc(actor) + `</strong>` + esc(entry.Description) + `</p><em>` + timeAgoLabel(entry.CreatedAt) + `</em></div>`)
		shown++
	}
	if shown == 0 {
		rows.WriteString(`<p class="empty-note">Belum ada aktivitas tercatat.</p>`)
	}
	return `<div class="panel">
      <div class="panel-head"><h2>Recent Activity</h2><a href="/owner/activity" hx-get="/owner/activity" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Lihat Semua</a></div>
      <div class="activity-grid">` + rows.String() + `</div>
    </div>`
}

// ownerReminderPanel surfaces real, actionable counts instead of static
// copy: invoices due today or overdue, unverified payments, and documents
// awaiting revision.
func ownerReminderPanel(orders []dashboard.Order, documents []dashboard.Document) string {
	endOfToday := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
	dueCount := 0
	for _, order := range orders {
		if order.Status != dashboard.OrderPaid && !order.DueDate.IsZero() && order.DueDate.Before(endOfToday) {
			dueCount++
		}
	}
	waitingCount := countWaitingOrders(orders)
	revisionCount := 0
	for _, document := range documents {
		if document.Status == dashboard.DocumentRevision {
			revisionCount++
		}
	}
	items := []struct {
		tone  string
		label string
	}{
		{"amber", intText(dueCount) + " invoice jatuh tempo/terlambat"},
		{"red", intText(waitingCount) + " pembayaran belum diverifikasi"},
		{"amber", intText(revisionCount) + " dokumen client perlu revisi"},
	}
	var lis strings.Builder
	for _, item := range items {
		lis.WriteString(`<li><span class="alert-dot ` + item.tone + `"></span>` + esc(item.label) + `</li>`)
	}
	return `<div class="panel">
      <div class="panel-head"><h2>Reminder</h2><a href="/owner/clients" hx-get="/owner/clients" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Lihat Semua</a></div>
      <ul class="reminder-list">` + lis.String() + `</ul>
    </div>`
}

func ownerFinance(vm dashboard.ViewModel) string {
	return `
<div class="space-y-5">
  <div class="toolbar-row"><div class="tabs"><span class="tab-active">Ringkasan</span><span>Pemasukan</span><span>Pengeluaran</span><span>Piutang</span><span>Arus Kas</span></div><div class="toolbar-actions"><span class="date-pill">` + esc(vm.DateLabel) + `</span><a class="primary-button" href="/owner/finance/export">Export</a></div></div>
  <section class="metric-grid metric-grid-four">
    <article class="metric-card"><div><p>Total Pemasukan</p><strong>` + money(vm.Snapshot.Revenue) + `</strong></div></article>
    <article class="metric-card"><div><p>Total Pengeluaran</p><strong>` + money(vm.Snapshot.Expenses) + `</strong></div></article>
    <article class="metric-card"><div><p>Profit</p><strong class="text-emerald-700">` + money(vm.Snapshot.Profit) + `</strong></div></article>
    <article class="metric-card"><div><p>Piutang</p><strong>` + money(unpaidAmount(vm.Orders)) + `</strong></div></article>
  </section>
  <section class="grid gap-5 xl:grid-cols-[minmax(0,1.1fr)_minmax(0,.9fr)]">
    <div class="panel">
      <div class="panel-head"><h2>Grafik Pemasukan vs Pengeluaran</h2></div>
      <div class="line-chart" aria-label="Grafik pemasukan dan pengeluaran">
        <div class="bar h-20"><i class="income h-16"></i><i class="expense h-9"></i></div>
        <div class="bar h-24"><i class="income h-20"></i><i class="expense h-12"></i></div>
        <div class="bar h-24"><i class="income h-18"></i><i class="expense h-11"></i></div>
        <div class="bar h-28"><i class="income h-24"></i><i class="expense h-14"></i></div>
        <div class="bar h-24"><i class="income h-20"></i><i class="expense h-12"></i></div>
        <div class="bar h-32"><i class="income h-28"></i><i class="expense h-16"></i></div>
        <div class="bar h-28"><i class="income h-24"></i><i class="expense h-14"></i></div>
        <div class="bar h-36"><i class="income h-32"></i><i class="expense h-18"></i></div>
      </div>
      <div class="chart-legend"><span><i class="bg-violet-600"></i>Pemasukan</span><span><i class="bg-amber-500"></i>Pengeluaran</span></div>
    </div>
    <div class="panel">
      <div class="panel-head"><h2>Pemasukan per Kategori</h2></div>
      <div class="category-list">
        <div><span><i class="bg-violet-600"></i>Order lunas</span><strong>` + money(vm.Snapshot.Revenue) + `</strong><div><b class="w-[54%] bg-violet-600"></b></div><em>` + intText(countPaidOrders(vm.Orders)) + ` order</em></div>
        <div><span><i class="bg-blue-500"></i>Belum lunas</span><strong>` + money(unpaidAmount(vm.Orders)) + `</strong><div><b class="w-[17%] bg-blue-500"></b></div><em>` + intText(countOpenOrders(vm.Orders)) + ` order</em></div>
        <div><span><i class="bg-amber-500"></i>Pengeluaran</span><strong>` + money(vm.Snapshot.Expenses) + `</strong><div><b class="w-[29%] bg-amber-500"></b></div><em>DB</em></div>
      </div>
    </div>
  </section>
  <div class="panel table-panel">
    <div class="panel-head"><h2>Transaksi Terbaru</h2><a href="#">Lihat Semua</a></div>
    <table class="data-table"><thead><tr><th>Tanggal</th><th>Deskripsi</th><th>Kategori</th><th>Masuk (IDR)</th><th>Keluar (IDR)</th><th>Saldo (IDR)</th></tr></thead><tbody>
      <tr><td>-</td><td>Order lunas</td><td><span class="status green">Pemasukan</span></td><td>` + money(vm.Snapshot.Revenue) + `</td><td>-</td><td>` + money(vm.Snapshot.Profit) + `</td></tr>
      <tr><td>-</td><td>Pengeluaran tercatat</td><td><span class="status red">Pengeluaran</span></td><td>-</td><td>` + money(vm.Snapshot.Expenses) + `</td><td>` + money(vm.Snapshot.Profit) + `</td></tr>
      <tr><td>-</td><td>Piutang order</td><td><span class="status amber">Piutang</span></td><td>` + money(unpaidAmount(vm.Orders)) + `</td><td>-</td><td>` + money(vm.Snapshot.Profit) + `</td></tr>
    </tbody></table>
  </div>
</div>`
}

func ownerPipeline(vm dashboard.ViewModel) string {
	return `
<div class="space-y-5">
  ` + pipelineToolbar(vm, "/owner/pipeline") + `
  ` + pipelineColumns(vm) + `
</div>`
}

func clientsPanel(vm dashboard.ViewModel, includeStaffFilter bool) string {
	path := "/" + vm.Role.String() + "/clients"
	if vm.EditClientID != "" {
		return clientEditPanel(vm, path)
	}
	clients := filterClients(vm.Clients, vm.FilterStatus, vm.FilterPackage, vm.FilterPIC, vm.FilterSearch)

	selectField := func(name, label string, options []string, selected string) string {
		var opts strings.Builder
		opts.WriteString(`<option value="">` + esc(label) + `</option>`)
		for _, option := range options {
			sel := ""
			if option == selected {
				sel = " selected"
			}
			opts.WriteString(`<option value="` + attr(option) + `"` + sel + `>` + esc(option) + `</option>`)
		}
		return `<select name="` + name + `" aria-label="` + attr(label) + `">` + opts.String() + `</select>`
	}

	filterFields := selectField("package", "Semua Paket", distinctClientValues(vm.Clients, func(c dashboard.ClientProfile) string { return c.PackageName }), vm.FilterPackage)
	filterFields += selectField("status", "Semua Status", distinctClientValues(vm.Clients, func(c dashboard.ClientProfile) string { return c.Status }), vm.FilterStatus)
	if includeStaffFilter {
		filterFields += selectField("pic", "Semua Staff", distinctClientValues(vm.Clients, func(c dashboard.ClientProfile) string { return c.PICName }), vm.FilterPIC)
	}

	var rows strings.Builder
	for i, client := range clients {
		resetConfirm := "Reset password " + client.Name + "? Password baru akan digenerate dan ditampilkan sekali."
		editHref := path + "?edit=" + attr(client.ID)
		activeBadge := `<span class="status green">Aktif</span>`
		toggleAction := ""
		deleteAction := ""
		if !client.Active {
			activeBadge = `<span class="status red">Nonaktif</span>`
		}
		if vm.Role == dashboard.RoleOwner {
			toggleLabel := "Nonaktifkan"
			toggleConfirm := "Nonaktifkan akses login " + client.Name + "? Data client tetap tersimpan, hanya login yang diblokir."
			if !client.Active {
				toggleLabel = "Aktifkan"
				toggleConfirm = "Aktifkan kembali akses login " + client.Name + "?"
			}
			toggleAction = ` <form method="post" action="` + attr(path) + `/` + attr(client.ID) + `/toggle-active" class="inline-form" data-confirm="` + attr(toggleConfirm) + `"><button type="submit" class="icon-action">` + toggleLabel + `</button></form>`
			deleteConfirm := "Hapus permanen client " + client.Name + "? Hanya bisa jika belum ada order, dokumen, task, expense, shipment, atau jadwal. Tindakan ini tidak bisa dibatalkan."
			deleteAction = ` <form method="post" action="` + attr(path) + `/` + attr(client.ID) + `/delete" class="inline-form" data-confirm="` + attr(deleteConfirm) + `"><button type="submit" class="icon-action icon-action-danger">Hapus</button></form>`
		}
		rows.WriteString(`<tr><td>` + intText(i+1) + `</td><td><strong>` + esc(client.Name) + `</strong><span>` + esc(client.Email) + `</span></td><td>` + esc(client.PackageName) + `</td><td>` + esc(client.PICName) + `</td><td>` + activeBadge + ` ` + esc(client.Status) + `</td><td><span class="progress-line"><i style="width:` + percentStyle(client.Progress) + `"></i></span>` + intText(client.Progress) + `%</td><td>` + esc(client.LastSchedule) + `</td><td><a class="icon-action" href="` + attr(editHref) + `" hx-get="` + attr(editHref) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Edit</a> <a class="icon-action" href="/` + attr(vm.Role.String()) + `/chat" hx-get="/` + attr(vm.Role.String()) + `/chat" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Chat</a> <a class="icon-action" href="/` + attr(vm.Role.String()) + `/clients/` + attr(client.ID) + `/agreement.pdf">Perjanjian</a> <form method="post" action="/` + attr(vm.Role.String()) + `/clients/` + attr(client.ID) + `/reset-password" class="inline-form" data-confirm="` + attr(resetConfirm) + `"><button type="submit" class="icon-action">Reset Password</button></form>` + toggleAction + deleteAction + `</td></tr>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<tr><td colspan="8">Belum ada client yang cocok dengan filter ini.</td></tr>`)
	}
	addModal := ""
	if vm.ShowCreateForm {
		var packageOptions strings.Builder
		for _, pkg := range vm.ServicePackages {
			packageOptions.WriteString(`<option value="` + attr(pkg.Name) + `">` + esc(pkg.Name) + `</option>`)
		}
		var staffOptions strings.Builder
		for _, staff := range vm.Staff {
			staffOptions.WriteString(`<option value="` + attr(staff.ID) + `">` + esc(staff.Name) + `</option>`)
		}
		invoiceFields := ""
		if vm.Role == dashboard.RoleOwner {
			invoiceFields = `<label>Total Tagihan (IDR)<input name="invoice_total" type="number" min="0" placeholder="7500000" data-order-total></label><label>Uang Masuk (IDR)<input name="amount_paid" type="number" min="0" placeholder="0" data-order-paid></label><label class="span-2">Kekurangan (IDR)<input type="text" readonly tabindex="-1" data-order-remaining value="0"></label>`
		}
		addModal = `<div class="modal-overlay"><div class="modal-card modal-card-wide">
  <div class="panel-head"><h3>Tambah Client</h3><a class="outline-button" href="` + attr(path) + `" hx-get="` + attr(path) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Batal</a></div>
  <form method="post" action="` + attr(path) + `/create" class="student-upload-form modal-form-grid" autocomplete="off"><label>Nama Lengkap<input name="name" required placeholder="Nama client" autocomplete="off"></label><label>Username<input name="username" required placeholder="Untuk login client" autocomplete="off"></label><label>Password<input name="password" type="password" minlength="8" required placeholder="Minimal 8 karakter" autocomplete="new-password"></label><label>Email<input name="email" type="email" placeholder="client@email.com" autocomplete="off"></label><label>No. WhatsApp<input name="phone" placeholder="08xxxxxxxxxx"></label><label>Paket<select name="package_name"><option value="">Belum dipilih</option>` + packageOptions.String() + `</select></label><label>Negara<input name="country" placeholder="Taiwan"></label><label>Kampus<input name="campus" placeholder="Nama kampus"></label><label>PIC Staff<select name="pic_staff_id"><option value="">Otomatis</option>` + staffOptions.String() + `</select></label>` + invoiceFields + `<div class="span-2 package-edit-actions"><button class="primary-button" type="submit">Simpan Client</button></div></form>
</div></div>`
	}
	return `
<div class="panel table-panel">
  <div class="panel-head"><h2>Data Client</h2>` + toggleFormButton(vm.ShowCreateForm, "+ Tambah Client", path) + `</div>
  ` + addModal + `
  <form method="get" action="` + attr(path) + `" hx-get="` + attr(path) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true" class="toolbar-row mb-4"><label class="search-box wide">` + toolIconHTML("search") + `<input type="search" name="q" value="` + attr(vm.FilterSearch) + `" placeholder="Cari nama atau email" aria-label="Cari client"></label><div class="filter-group">` + filterFields + `<button class="outline-button" type="submit">Terapkan</button><a class="outline-button" href="` + attr(path) + `" hx-get="` + attr(path) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Reset</a><a class="primary-button" href="` + attr(path) + `/export">Export</a></div></form>
  <table class="data-table"><thead><tr><th>No.</th><th>Nama Klien</th><th>Paket / Layanan</th><th>PIC</th><th>Status Terakhir</th><th>Progress</th><th>Jadwal Terakhir</th><th>Aksi</th></tr></thead><tbody>
  ` + rows.String() + `</tbody></table>
  <div class="pagination"><span>Menampilkan ` + intText(len(clients)) + ` dari ` + intText(len(vm.Clients)) + ` data</span><div><button disabled>&lt;</button><button class="active">1</button><button disabled>&gt;</button></div></div>
</div>`
}

// clientEditPanel lets owner/staff edit an existing client's profile fields
// in place of the table, mirroring staffEditPanel/intakeEditPanel's
// ?edit=<id> pattern. Credentials aren't editable here — that's still
// ResetStudentPassword, kept as a separate, explicit action.
func clientEditPanel(vm dashboard.ViewModel, basePath string) string {
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
	var packageOptions strings.Builder
	for _, pkg := range vm.ServicePackages {
		sel := ""
		if pkg.Name == target.PackageName {
			sel = " selected"
		}
		packageOptions.WriteString(`<option value="` + attr(pkg.Name) + `"` + sel + `>` + esc(pkg.Name) + `</option>`)
	}
	var staffOptions strings.Builder
	for _, staff := range vm.Staff {
		sel := ""
		if staff.ID == target.PICStaffID {
			sel = " selected"
		}
		staffOptions.WriteString(`<option value="` + attr(staff.ID) + `"` + sel + `>` + esc(staff.Name) + `</option>`)
	}
	invoiceFields := ""
	if vm.Role == dashboard.RoleOwner {
		order, hasOrder := latestOrderForClient(vm.Orders, target.ID)
		totalValue := ""
		if hasOrder {
			totalValue = strconv.FormatInt(order.Total, 10)
		}
		invoiceFields = `<label>Total Tagihan (IDR)<input name="invoice_total" type="number" min="0" placeholder="7500000" value="` + attr(totalValue) + `" data-order-total></label><label>Uang Masuk (IDR)<input name="amount_paid" type="number" min="0" value="` + strconv.FormatInt(order.Paid, 10) + `" data-order-paid></label><label class="span-2">Kekurangan (IDR)<input type="text" readonly tabindex="-1" data-order-remaining value="0"></label>`
	}
	return `<div class="panel">
  <div class="panel-head"><div><h2 class="text-lg font-semibold">Edit Client — ` + esc(target.Name) + `</h2><p class="text-sm text-slate-500">Password login diubah lewat "Reset Password" di tabel, bukan di sini.</p></div><a class="outline-button" href="` + attr(basePath) + `" hx-get="` + attr(basePath) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Batal</a></div>
  <form method="post" action="` + basePath + `/` + attr(target.ID) + `/update" class="student-upload-form modal-form-grid">
    <label>Nama Lengkap<input name="name" value="` + attr(target.Name) + `" required></label>
    <label>Email<input name="email" type="email" value="` + attr(target.Email) + `"></label>
    <label>No. WhatsApp<input name="phone" value="` + attr(target.Phone) + `" placeholder="08xxxxxxxxxx"></label>
    <label>Paket<select name="package_name"><option value="">Belum dipilih</option>` + packageOptions.String() + `</select></label>
    <label>Negara<input name="country" value="` + attr(target.Country) + `"></label>
    <label>Kampus<input name="campus" value="` + attr(target.Campus) + `"></label>
    <label>PIC Staff<select name="pic_staff_id"><option value="">Tidak diubah</option>` + staffOptions.String() + `</select></label>` + invoiceFields + `
    <div class="span-2 package-edit-actions"><button class="primary-button" type="submit">Simpan Perubahan</button></div>
  </form>
</div>`
}

func distinctClientValues(clients []dashboard.ClientProfile, selector func(dashboard.ClientProfile) string) []string {
	seen := make(map[string]bool)
	var values []string
	for _, client := range clients {
		value := selector(client)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}

func filterClients(clients []dashboard.ClientProfile, status, packageName, pic, search string) []dashboard.ClientProfile {
	search = strings.ToLower(strings.TrimSpace(search))
	filtered := make([]dashboard.ClientProfile, 0, len(clients))
	for _, client := range clients {
		if status != "" && client.Status != status {
			continue
		}
		if packageName != "" && client.PackageName != packageName {
			continue
		}
		if pic != "" && client.PICName != pic {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(client.Name), search) && !strings.Contains(strings.ToLower(client.Email), search) {
			continue
		}
		filtered = append(filtered, client)
	}
	return filtered
}

// ownerServices renders the price list from vm.ServicePackages — owner-editable
// catalog data — instead of three hardcoded cards with made-up numbers. Client
// counts and revenue per package are computed live from vm.Clients/vm.Orders,
// so the stats shown are real, not decorative.
func ownerServices(vm dashboard.ViewModel) string {
	toggleLabel := "Kelola Paket"
	toggleTarget := "/owner/services?manage=1"
	if vm.ShowStageManager {
		toggleLabel = "Tutup Pengaturan Paket"
		toggleTarget = "/owner/services"
	}
	toolbar := `<div class="toolbar-row"><div><h2 class="text-lg font-semibold">Paket &amp; Layanan</h2><p class="text-sm text-slate-500">Daftar harga dikelola sendiri lewat "Kelola Paket" — bukan hardcode.</p></div><div class="toolbar-actions"><a class="outline-button" href="` + attr(toggleTarget) + `" hx-get="` + attr(toggleTarget) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">` + toolIconHTML("settings-slider") + ` ` + toggleLabel + `</a></div></div>`
	managePanel := ""
	if vm.ShowStageManager {
		managePanel = packageManagerPanel(vm.ServicePackages)
	}

	var cards strings.Builder
	if len(vm.ServicePackages) == 0 {
		cards.WriteString(`<p class="empty-note">Belum ada paket. Klik "Kelola Paket" untuk menambahkan.</p>`)
	}
	for i, pkg := range vm.ServicePackages {
		bg, text := serviceBadgeTone(i)
		category := pkg.Category
		if category == "" {
			category = "Paket"
		}
		priceLabel := money(pkg.Price)
		if pkg.PriceIsFrom {
			priceLabel = "Mulai " + priceLabel
		}
		inProgress, revenue, orders := packageStats(vm, pkg.Name)
		var items strings.Builder
		items.WriteString(`<li>` + intText(inProgress) + ` client dalam proses</li>`)
		items.WriteString(`<li>` + intText(orders) + ` order tercatat &middot; ` + money(revenue) + ` pendapatan lunas</li>`)
		for _, line := range strings.Split(pkg.Highlights, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				items.WriteString(`<li>` + esc(line) + `</li>`)
			}
		}
		cards.WriteString(`<article class="panel service-card"><span class="service-badge ` + bg + ` ` + text + `">` + esc(category) + `</span><h2>` + esc(pkg.Name) + `</h2><p>` + esc(pkg.Description) + `</p><strong>` + priceLabel + `</strong><ul>` + items.String() + `</ul></article>`)
	}

	return `<div class="space-y-5">` + toolbar + managePanel + `<div class="grid gap-5 xl:grid-cols-3">` + cards.String() + `</div></div>`
}

func serviceBadgeTone(position int) (string, string) {
	palette := [][2]string{
		{"bg-violet-100", "text-violet-700"},
		{"bg-blue-100", "text-blue-700"},
		{"bg-amber-100", "text-amber-700"},
		{"bg-emerald-100", "text-emerald-700"},
		{"bg-rose-100", "text-rose-700"},
		{"bg-sky-100", "text-sky-700"},
	}
	pair := palette[position%len(palette)]
	return pair[0], pair[1]
}

// packageStats computes real, live numbers from already-loaded client/order
// data instead of the placeholder counts the static cards used to show.
func packageStats(vm dashboard.ViewModel, packageName string) (inProgress int, revenue int64, orders int) {
	for _, client := range vm.Clients {
		if client.PackageName == packageName && client.Progress < 100 {
			inProgress++
		}
	}
	for _, order := range vm.Orders {
		if order.PackageName != packageName {
			continue
		}
		orders++
		if order.Status == dashboard.OrderPaid {
			revenue += order.Paid
		}
	}
	return
}

// packageManagerPanel lets the owner add, edit, reorder, or delete packages
// themselves — the price list changes with the business, not with a code deploy.
func packageManagerPanel(packages []dashboard.ServicePackage) string {
	var rows strings.Builder
	for i, pkg := range packages {
		upAttr := ""
		if i == 0 {
			upAttr = " disabled"
		}
		downAttr := ""
		if i == len(packages)-1 {
			downAttr = " disabled"
		}
		checked := ""
		if pkg.PriceIsFrom {
			checked = " checked"
		}
		confirmMsg := "Hapus paket \"" + pkg.Name + "\"? Data client/order yang sudah memakai nama paket ini tidak akan berubah."
		rows.WriteString(`<div class="package-manager-row">`)
		rows.WriteString(`<form method="post" action="/owner/services/` + attr(pkg.ID) + `/update" class="package-edit-form">`)
		rows.WriteString(`<div class="package-edit-grid">`)
		rows.WriteString(`<label>Nama<input name="name" value="` + attr(pkg.Name) + `" required maxlength="120"></label>`)
		rows.WriteString(`<label>Kategori<input name="category" value="` + attr(pkg.Category) + `" maxlength="60" placeholder="Paket / Satuan"></label>`)
		rows.WriteString(`<label>Harga (Rp)<input name="price" type="number" min="0" value="` + intText(int(pkg.Price)) + `" required></label>`)
		rows.WriteString(`<label class="checkbox-label"><input type="checkbox" name="price_is_from" value="1"` + checked + `> Harga mulai dari</label>`)
		rows.WriteString(`</div>`)
		rows.WriteString(`<label>Deskripsi<textarea name="description" rows="2" maxlength="300">` + esc(pkg.Description) + `</textarea></label>`)
		rows.WriteString(`<label>Highlight (satu baris = satu poin)<textarea name="highlights" rows="2" maxlength="300">` + esc(pkg.Highlights) + `</textarea></label>`)
		rows.WriteString(`<div class="package-edit-actions"><button class="primary-button small" type="submit">Simpan</button></div>`)
		rows.WriteString(`</form>`)
		rows.WriteString(`<div class="stage-manager-actions">`)
		rows.WriteString(`<form method="post" action="/owner/services/` + attr(pkg.ID) + `/move" class="inline-form"><input type="hidden" name="direction" value="up"><button type="submit"` + upAttr + ` title="Naikkan urutan" aria-label="Naikkan urutan">` + svgIcon("arrow-up") + `</button></form>`)
		rows.WriteString(`<form method="post" action="/owner/services/` + attr(pkg.ID) + `/move" class="inline-form"><input type="hidden" name="direction" value="down"><button type="submit"` + downAttr + ` title="Turunkan urutan" aria-label="Turunkan urutan">` + svgIcon("arrow-down") + `</button></form>`)
		rows.WriteString(`<form method="post" action="/owner/services/` + attr(pkg.ID) + `/delete" class="inline-form" data-confirm="` + attr(confirmMsg) + `"><button type="submit" class="stage-delete" title="Hapus paket" aria-label="Hapus paket">` + svgIcon("close") + `</button></form>`)
		rows.WriteString(`</div></div>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<p class="empty-note">Belum ada paket. Tambahkan paket pertama di bawah.</p>`)
	}
	addForm := `<form method="post" action="/owner/services/create" class="package-edit-form mt-2"><div class="package-edit-grid"><label>Nama<input name="name" required maxlength="120" placeholder="Contoh: Paket Express"></label><label>Kategori<input name="category" maxlength="60" placeholder="Paket / Satuan"></label><label>Harga (Rp)<input name="price" type="number" min="0" required placeholder="15000000"></label><label class="checkbox-label"><input type="checkbox" name="price_is_from" value="1"> Harga mulai dari</label></div><label>Deskripsi<textarea name="description" rows="2" maxlength="300" placeholder="Ringkasan paket"></textarea></label><label>Highlight (satu baris = satu poin)<textarea name="highlights" rows="2" maxlength="300" placeholder="Contoh: Prioritas review dokumen"></textarea></label><div class="package-edit-actions"><button class="primary-button" type="submit">+ Tambah Paket</button></div></form>`
	return `<div class="panel stage-manager"><h3 class="text-sm font-semibold mb-1">Kelola Paket &amp; Layanan</h3><p class="text-xs text-slate-500 mb-3">Tambah, ubah, urutkan, atau hapus paket sesuai kebutuhan bisnis kamu — tidak perlu ubah kode program.</p>` + rows.String() + addForm + `</div>`
}

func ownerInvoices(vm dashboard.ViewModel) string {
	const invoicesPath = "/owner/invoices"
	if vm.EditClientID != "" {
		return orderEditPanel(vm, invoicesPath)
	}

	addTrigger := toggleFormButton(vm.ShowCreateForm, "+ Tambah Order", invoicesPath)
	addModal := ""
	if vm.ShowCreateForm {
		addModal = orderCreateModal(vm, invoicesPath)
	}

	if len(vm.Orders) == 0 {
		return `<div class="space-y-5">` + addModal + `<div class="panel empty-state"><div class="panel-head"><div><span>Invoice</span><h2>Belum ada invoice</h2><p>Invoice akan muncul di sini setelah owner atau staff membuat order untuk client.</p></div>` + addTrigger + `</div></div></div>`
	}

	filter := vm.InvoiceFilter
	filterLink := func(label, value string, count int) string {
		className := ""
		if filter == value {
			className = ` class="active"`
		}
		href := "/owner/invoices"
		if value != "" {
			href += "?filter=" + value
		}
		return `<a` + className + ` href="` + attr(href) + `" hx-get="` + attr(href) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">` + esc(label) + ` <span>` + intText(count) + `</span></a>`
	}
	var side strings.Builder
	side.WriteString(filterLink("Semua", "", len(vm.Orders)))
	side.WriteString(filterLink("Belum Dibayar", "unpaid", len(filterOrdersByStatus(vm.Orders, "unpaid"))))
	side.WriteString(filterLink("Menunggu Verifikasi", "waiting", len(filterOrdersByStatus(vm.Orders, "waiting"))))
	side.WriteString(filterLink("Lunas", "paid", len(filterOrdersByStatus(vm.Orders, "paid"))))

	filtered := filterOrdersByStatus(vm.Orders, filter)
	pickFrom := filtered
	if len(pickFrom) == 0 {
		pickFrom = vm.Orders
	}
	active := selectOrder(pickFrom, vm.ActiveOrderCode)

	var invoiceList strings.Builder
	for _, order := range filtered {
		className := ""
		if order.Code == active.Code {
			className = ` class="active"`
		}
		href := "/owner/invoices?order=" + url.QueryEscape(order.Code)
		if filter != "" {
			href += "&filter=" + url.QueryEscape(filter)
		}
		pct := orderPaymentPercent(order)
		invoiceList.WriteString(`<a` + className + ` href="` + attr(href) + `" hx-get="` + attr(href) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true"><strong>` + esc(order.ClientName) + `</strong><span>` + esc(order.Code) + ` &middot; ` + esc(orderStatusLabel(order.Status)) + `</span><span><span class="progress-line"><i style="width:` + percentStyle(pct) + `"></i></span>` + intText(pct) + `% &middot; Sisa ` + money(orderRemaining(order)) + `</span><strong>` + money(order.Total) + `</strong></a>`)
	}
	if invoiceList.Len() == 0 {
		invoiceList.WriteString(`<p class="empty-note">Tidak ada invoice pada filter ini.</p>`)
	}

	var paymentRows strings.Builder
	for _, payment := range vm.OrderPayments {
		if payment.OrderID != active.ID {
			continue
		}
		note := payment.Note
		if note == "" {
			note = "-"
		}
		proofCell := "-"
		if payment.ProofStoragePath != "" {
			proofCell = `<a href="/payments/` + attr(payment.ID) + `/proof" target="_blank" rel="noopener">Lihat</a>`
		}
		actions := "-"
		if payment.Status == dashboard.PaymentPending {
			actions = `<form method="post" action="/owner/invoices/payments/` + attr(payment.ID) + `/verify" class="inline-form"><button type="submit" class="icon-action">Verifikasi</button></form> <form method="post" action="/owner/invoices/payments/` + attr(payment.ID) + `/reject" class="inline-form" data-confirm="Tolak cicilan Rp` + attr(money(payment.Amount)) + ` ini?"><button type="submit" class="icon-action icon-action-danger">Tolak</button></form>`
		} else if payment.Status == dashboard.PaymentRejected && payment.RejectReason != "" {
			actions = esc(payment.RejectReason)
		}
		paymentRows.WriteString(`<tr><td>` + esc(dateLabel(payment.SubmittedAt)) + `<span class="block text-xs text-slate-500">` + esc(note) + `</span></td><td>` + money(payment.Amount) + `</td><td><span class="status ` + paymentStatusClass(payment.Status) + `">` + paymentStatusLabel(payment.Status) + `</span></td><td>` + proofCell + `</td><td>` + actions + `</td></tr>`)
	}
	if paymentRows.Len() == 0 {
		paymentRows.WriteString(`<tr><td colspan="5">Belum ada cicilan.</td></tr>`)
	}
	paymentsTable := `<table class="data-table mt-3"><thead><tr><th>Tanggal &amp; Catatan</th><th>Jumlah</th><th>Status</th><th>Bukti</th><th>Aksi</th></tr></thead><tbody>` + paymentRows.String() + `</tbody></table>`
	statusClass := orderStatusClass(active.Status)
	client := findClientByID(vm.Clients, active.ClientID)
	waHref := waLink(client.Phone, "Halo "+active.ClientName+", ini invoice "+active.Code+" sebesar "+money(active.Total)+" dari Formora Taiwan.")
	mailHref := mailtoLink(client.Email, "Invoice "+active.Code, "Halo "+active.ClientName+",\n\nBerikut invoice "+active.Code+" sebesar "+money(active.Total)+".\n\nTerima kasih.")
	pdfHref := "/owner/invoices/" + url.QueryEscape(active.Code) + "/invoice.pdf"
	editHref := invoicesPath + "?edit=" + url.QueryEscape(active.ID) + "&order=" + url.QueryEscape(active.Code)
	deleteConfirm := "Hapus order " + active.Code + "? Tindakan ini tidak bisa dibatalkan."
	commActions := `<div class="action-stack"><a class="success-button" href="` + attr(waHref) + `" target="_blank" rel="noopener">Kirim Invoice (WA)</a><a class="primary-button" href="` + attr(mailHref) + `">Kirim Invoice (Email)</a><a class="outline-button" href="` + attr(pdfHref) + `">Download PDF</a><a class="outline-button" href="` + attr(editHref) + `" hx-get="` + attr(editHref) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Edit Order</a><form method="post" action="` + invoicesPath + `/` + attr(active.ID) + `/delete" data-confirm="` + attr(deleteConfirm) + `"><button type="submit" class="danger-button" style="width:100%">Hapus Order</button></form></div>`

	manualInstallmentForm := `<form method="post" action="/owner/invoices/` + attr(active.ID) + `/payments/create" class="student-upload-form mt-3"><label>Tambah Cicilan Manual (IDR)<input name="amount" type="number" min="1" max="` + strconv.FormatInt(active.Total-active.Paid, 10) + `" required placeholder="Contoh: 2000000"></label><label>Catatan<input name="note" placeholder="Contoh: Transfer BCA a.n. client"></label><button class="outline-button" type="submit">Tambah Cicilan</button></form>`

	var statusAction string
	switch active.Status {
	case dashboard.OrderPaid:
		statusAction = `<div class="invoice-status-note is-paid"><strong>Invoice sudah lunas</strong><p>Pembayaran sudah tercatat penuh di sistem.</p></div>`
	case dashboard.OrderWaitingVerification:
		statusAction = `<div class="invoice-status-note"><strong>Menunggu verifikasi</strong><p>Client sudah mengirim cicilan baru. Verifikasi atau tolak dari daftar Riwayat Cicilan di bawah, atau tandai lunas langsung kalau sudah pasti.</p><form method="post" action="/owner/orders/mark-paid"><input type="hidden" name="order_code" value="` + attr(active.Code) + `"><button class="success-button" type="submit">Tandai Lunas</button></form>` + manualInstallmentForm + `</div>`
	default:
		statusAction = `<div class="invoice-status-note"><strong>Belum ada pembayaran</strong><p>Catat cicilan yang sudah masuk, atau tandai lunas manual jika client sudah transfer penuh di luar sistem.</p><form method="post" action="/owner/orders/mark-paid"><input type="hidden" name="order_code" value="` + attr(active.Code) + `"><button class="outline-button" type="submit">Tandai Lunas Manual</button></form>` + manualInstallmentForm + `</div>`
	}

	return `<div class="space-y-5">` + addModal + `
<div class="grid gap-5 xl:grid-cols-[320px_minmax(0,1fr)]">
  <aside class="panel"><div class="panel-head"><div><h2 class="text-base font-semibold">Semua Invoice</h2><p class="text-sm text-slate-500">Klik salah satu invoice untuk lihat detail dan tindak lanjut.</p></div>` + addTrigger + `</div><div class="side-filter">` + side.String() + `</div><div class="invoice-list mt-4">` + invoiceList.String() + `</div></aside>
  <section class="panel">
    <div class="invoice-head"><div><p>Invoice <strong>#` + esc(active.Code) + `</strong></p><span>` + esc(active.ClientName) + ` &middot; ` + esc(active.PackageName) + `</span></div><div><p>Jatuh Tempo</p><strong>` + esc(dateLabel(active.DueDate)) + `</strong></div></div>
    <div class="invoice-layout">
      <div><dl class="detail-grid"><div><dt>Nama Klien</dt><dd>` + esc(active.ClientName) + `</dd></div><div><dt>Paket / Layanan</dt><dd>` + esc(active.PackageName) + `</dd></div><div><dt>Kode Pesanan</dt><dd>` + esc(active.Code) + `</dd></div><div><dt>Total Tagihan</dt><dd>` + money(active.Total) + `</dd></div><div><dt>Status Pembayaran</dt><dd><span class="status ` + statusClass + `">` + esc(orderStatusLabel(active.Status)) + `</span></dd></div></dl><div class="invoice-box"><h3>Rincian Tagihan</h3><div><span>` + esc(active.PackageName) + `</span><strong>` + money(active.Total) + `</strong></div><div><span>Sudah Dibayar (Uang Masuk)</span><strong>` + money(active.Paid) + `</strong></div><div class="total"><span>Kekurangan</span><strong>` + money(active.Total-active.Paid) + `</strong></div></div><div class="invoice-box"><h3>Riwayat Cicilan</h3>` + paymentsTable + `</div></div>
      <div class="invoice-rail">` + statusAction + commActions + `</div>
    </div>
  </section>
</div></div>`
}

// orderCreateModal is the "+ Tambah Order" form for opening a new invoice
// against an existing client — the standalone counterpart to the
// invoice_total/amount_paid fields on the client creation form, for a
// renewal order or a client whose first order wasn't set up at signup.
func orderCreateModal(vm dashboard.ViewModel, basePath string) string {
	var clientOptions strings.Builder
	for _, client := range vm.Clients {
		clientOptions.WriteString(`<option value="` + attr(client.ID) + `">` + esc(client.Name) + `</option>`)
	}
	var packageOptions strings.Builder
	for _, pkg := range vm.ServicePackages {
		packageOptions.WriteString(`<option value="` + attr(pkg.Name) + `">` + esc(pkg.Name) + `</option>`)
	}
	return `<div class="modal-overlay"><div class="modal-card modal-card-wide">
  <div class="panel-head"><h3>Tambah Order</h3><a class="outline-button" href="` + attr(basePath) + `" hx-get="` + attr(basePath) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Batal</a></div>
  <form method="post" action="` + attr(basePath) + `/create" class="student-upload-form modal-form-grid">
    <label class="span-2">Client<select name="client_id" required><option value="">Pilih client</option>` + clientOptions.String() + `</select></label>
    <label class="span-2">Paket<select name="package_name"><option value="">Sama seperti profil client</option>` + packageOptions.String() + `</select></label>
    <label>Total Tagihan (IDR)<input name="total" type="number" min="1" required placeholder="7500000" data-order-total></label>
    <label>Uang Masuk (IDR)<input name="paid" type="number" min="0" placeholder="0" data-order-paid></label>
    <label class="span-2">Kekurangan (IDR)<input type="text" readonly tabindex="-1" data-order-remaining value="0"></label>
    <label>Jatuh Tempo<input name="due_date" type="date"></label>
    <div class="span-2 package-edit-actions"><button class="primary-button" type="submit">Simpan Order</button></div>
  </form>
</div></div>`
}

// orderEditPanel edits an existing order's package/total/due date, reusing
// the ?edit=<id> full-panel-swap pattern from clientEditPanel/
// staffEditPanel. Paid amount stays out of this form — RecordOrderPayment/
// MarkOrderPaidByCode own that value so status stays consistent.
func orderEditPanel(vm dashboard.ViewModel, basePath string) string {
	var target dashboard.Order
	found := false
	for _, order := range vm.Orders {
		if order.ID == vm.EditClientID {
			target = order
			found = true
			break
		}
	}
	if !found {
		return `<div class="panel"><p class="empty-note">Order tidak ditemukan.</p><a class="outline-button mt-3" href="` + attr(basePath) + `" hx-get="` + attr(basePath) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Kembali</a></div>`
	}
	var packageOptions strings.Builder
	matchedPackage := false
	for _, pkg := range vm.ServicePackages {
		sel := ""
		if pkg.Name == target.PackageName {
			sel = " selected"
			matchedPackage = true
		}
		packageOptions.WriteString(`<option value="` + attr(pkg.Name) + `"` + sel + `>` + esc(pkg.Name) + `</option>`)
	}
	if !matchedPackage && target.PackageName != "" {
		packageOptions.WriteString(`<option value="` + attr(target.PackageName) + `" selected>` + esc(target.PackageName) + `</option>`)
	}
	dueValue := ""
	if !target.DueDate.IsZero() {
		dueValue = target.DueDate.Format("2006-01-02")
	}
	return `<div class="panel">
  <div class="panel-head"><div><h2 class="text-lg font-semibold">Edit Order — ` + esc(target.Code) + `</h2><p class="text-sm text-slate-500">` + esc(target.ClientName) + `</p></div><a class="outline-button" href="` + attr(basePath) + `" hx-get="` + attr(basePath) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Batal</a></div>
  <form method="post" action="` + basePath + `/` + attr(target.ID) + `/update" class="student-upload-form modal-form-grid">
    <label class="span-2">Paket<select name="package_name">` + packageOptions.String() + `</select></label>
    <label>Total Tagihan (IDR)<input name="total" type="number" min="1" required value="` + strconv.FormatInt(target.Total, 10) + `" data-order-total></label>
    <label>Uang Masuk (IDR)<input name="paid" type="number" min="0" value="` + strconv.FormatInt(target.Paid, 10) + `" data-order-paid></label>
    <label class="span-2">Kekurangan (IDR)<input type="text" readonly tabindex="-1" data-order-remaining value="0"></label>
    <label>Jatuh Tempo<input name="due_date" type="date" value="` + dueValue + `"></label>
    <div class="span-2 package-edit-actions"><button class="primary-button" type="submit">Simpan Perubahan</button></div>
  </form>
</div>`
}

func filterOrdersByStatus(orders []dashboard.Order, filter string) []dashboard.Order {
	if filter == "" {
		return orders
	}
	var status dashboard.OrderStatus
	switch filter {
	case "unpaid":
		status = dashboard.OrderUnpaid
	case "waiting":
		status = dashboard.OrderWaitingVerification
	case "paid":
		status = dashboard.OrderPaid
	default:
		return orders
	}
	filtered := make([]dashboard.Order, 0, len(orders))
	for _, order := range orders {
		if order.Status == status {
			filtered = append(filtered, order)
		}
	}
	return filtered
}

func selectOrder(orders []dashboard.Order, code string) dashboard.Order {
	if code != "" {
		for _, order := range orders {
			if strings.EqualFold(order.Code, code) {
				return order
			}
		}
	}
	return firstOrder(orders)
}

func findClientByID(clients []dashboard.ClientProfile, id string) dashboard.ClientProfile {
	for _, client := range clients {
		if client.ID == id {
			return client
		}
	}
	return dashboard.ClientProfile{}
}

func waLink(phone, message string) string {
	digits := make([]byte, 0, len(phone))
	for i := 0; i < len(phone); i++ {
		if phone[i] >= '0' && phone[i] <= '9' {
			digits = append(digits, phone[i])
		}
	}
	number := string(digits)
	if strings.HasPrefix(number, "0") {
		number = "62" + number[1:]
	}
	return "https://wa.me/" + number + "?text=" + url.QueryEscape(message)
}

func mailtoLink(email, subject, body string) string {
	return "mailto:" + email + "?subject=" + url.QueryEscape(subject) + "&body=" + url.QueryEscape(body)
}

func ownerReports(vm dashboard.ViewModel) string {
	return `
<div class="space-y-5">
  <div class="toolbar-row"><div class="tabs"><span class="tab-active">Omzet</span><span>Profit</span><span>Pemasukan</span><span>Pengeluaran</span><span>Piutang</span></div><div class="toolbar-actions"><span class="date-pill">Tahun ` + intText(time.Now().Year()) + `</span><a class="primary-button" href="/owner/reports/export">Export</a></div></div>
  <section class="grid gap-5 xl:grid-cols-[minmax(0,1.4fr)_minmax(0,.8fr)]">
    <div class="panel"><div class="panel-head"><h2>Grafik Omzet (Tahun 2026)</h2></div><div class="monthly-chart"><i class="h-20"></i><i class="h-24"></i><i class="h-32"></i><i class="h-40"></i><i class="h-48"></i><i class="h-56"></i><i class="h-40"></i><i class="h-44"></i><i class="h-36"></i><i class="h-32"></i><i class="h-28"></i><i class="h-24"></i></div><div class="month-labels"><span>Jan</span><span>Feb</span><span>Mar</span><span>Apr</span><span>Mei</span><span>Jun</span><span>Jul</span><span>Agu</span><span>Sep</span><span>Okt</span><span>Nov</span><span>Des</span></div></div>
    <div class="panel"><div class="panel-head"><h2>Ringkasan Omzet</h2></div><div class="summary-list"><div><span>Total Omzet Lunas</span><strong>` + money(vm.Snapshot.Revenue) + `</strong></div><div><span>Total Pengeluaran</span><strong>` + money(vm.Snapshot.Expenses) + `</strong></div><div><span>Profit</span><strong>` + money(vm.Snapshot.Profit) + `</strong></div><div><span>Piutang</span><strong>` + money(unpaidAmount(vm.Orders)) + `</strong></div></div></div>
  </section>
  <div class="panel table-panel"><div class="panel-head"><h2>Detail Omzet</h2></div><table class="data-table"><thead><tr><th>Sumber</th><th>Jumlah Order</th><th>Omzet Lunas</th><th>Piutang</th><th>Profit</th></tr></thead><tbody><tr><td>Database saat ini</td><td>` + intText(len(vm.Orders)) + `</td><td>` + money(vm.Snapshot.Revenue) + `</td><td>` + money(unpaidAmount(vm.Orders)) + `</td><td>` + money(vm.Snapshot.Profit) + `</td></tr></tbody></table></div>
</div>`
}

func staffDashboard(vm dashboard.ViewModel) string {
	return `
<div class="staff-layout">
  <section class="metric-grid metric-grid-staff">
    <article class="metric-card clean">` + metricSymbolHTML("users") + `<div><p>Klien Saya</p><strong>` + intText(len(vm.Clients)) + `</strong></div></article>
    <article class="metric-card clean">` + metricSymbolHTML("briefcase") + `<div><p>Tugas Hari Ini</p><strong>` + intText(len(vm.Tasks)) + `</strong><span>` + intText(countOpenTasks(vm.Tasks)) + ` belum selesai</span></div></article>
    <article class="metric-card clean">` + metricSymbolHTML("file") + `<div><p>Dokumen Review</p><strong>` + intText(countReviewDocs(vm.Documents)) + `</strong><span>Dari client aktif</span></div></article>
    <article class="metric-card clean">` + metricSymbolHTML("wallet") + `<div><p>Pengeluaran Menunggu</p><strong>` + intText(countWaitingExpenses(vm.Expenses)) + `</strong><span>Perlu dicatat</span></div></article>
    <article class="metric-card clean">` + metricSymbolHTML("check") + `<div><p>Tugas Selesai</p><strong>` + intText(countDoneTasks(vm.Tasks)) + `</strong><span>Hari ini</span></div></article>
  </section>
  <section class="grid gap-5 xl:grid-cols-[minmax(0,1.05fr)_minmax(0,1.6fr)_minmax(0,.85fr)]">
    <div class="panel border-0 shadow-none"><div class="panel-head"><h2>Tugas Hari Ini</h2><a href="/staff/tasks" hx-get="/staff/tasks" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Lihat Semua</a></div>` + taskListCompact(vm.Tasks) + `<a class="link-row" href="/staff/tasks" hx-get="/staff/tasks" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Lihat Semua Tugas</a></div>
    <div class="panel border-0 shadow-none"><div class="panel-head"><h2>Pipeline Saya</h2><a href="/staff/pipeline" hx-get="/staff/pipeline" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Lihat Semua</a></div><div class="mini-kanban">` + pipelineColumns(vm) + `</div></div>
    <div class="right-rail"><div class="panel border-0 shadow-none"><div class="panel-head"><h2>Pengingat</h2><a href="/staff/tasks" hx-get="/staff/tasks" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Lihat Semua</a></div>` + staffReminderList(vm) + `</div>` + quickSchedule(vm) + recentDocs(vm) + `</div>
  </section>
  <section class="panel border-0 shadow-none">
    <div class="toolbar-row mb-4"><div><h2 class="text-lg font-semibold">Pengeluaran Operasional</h2><p class="text-sm text-slate-500">Catat biaya pengeluaran untuk setiap berkas atau keperluan klien.</p></div><a class="dark-button" href="/staff/expenses?new=1" hx-get="/staff/expenses?new=1" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">+ Catat Pengeluaran</a></div>
    ` + expensesTable(vm.Expenses) + `
  </section>
</div>`
}

func staffClients(vm dashboard.ViewModel) string {
	return `<div class="space-y-5"><section class="metric-grid metric-grid-four"><article class="metric-card"><p>Klien Aktif</p><strong>` + intText(len(vm.Clients)) + `</strong><span class="trend-up">Dalam akses staff</span></article><article class="metric-card"><p>Butuh Follow Up</p><strong>` + intText(countOpenTasks(vm.Tasks)) + `</strong><span class="text-amber-600">Prioritas</span></article><article class="metric-card"><p>Dokumen Review</p><strong>` + intText(countReviewDocs(vm.Documents)) + `</strong><span class="text-red-600">Harus dicek</span></article><article class="metric-card"><p>Selesai Bulan Ini</p><strong>` + intText(countPaidOrders(vm.Orders)) + `</strong><span class="trend-up">Order lunas</span></article></section>` + clientsPanel(vm, false) + `</div>`
}

func staffPipeline(vm dashboard.ViewModel) string {
	return `<div class="space-y-5">
  ` + pipelineToolbar(vm, "/staff/pipeline") + `
  <div class="panel border-0 shadow-none">` + pipelineColumns(vm) + `</div>
</div>`
}

// pipelineToolbar renders the header row above the kanban board, including
// the "Kelola Tahap" toggle that reveals stageManagerPanel — the entry point
// for owners/staff to define pipeline stages themselves instead of relying on
// a fixed set baked into the code.
func pipelineToolbar(vm dashboard.ViewModel, basePath string) string {
	toggleLabel := "Kelola Tahap"
	toggleTarget := basePath + "?manage=1"
	if vm.ShowStageManager {
		toggleLabel = "Tutup Pengaturan Tahap"
		toggleTarget = basePath
	}
	exportPath := "/" + vm.Role.String() + "/clients/export"
	toolbar := `<div class="toolbar-row"><div><h2 class="text-lg font-semibold">Pipeline Keseluruhan</h2><p class="text-sm text-slate-500">Tahap pipeline dikonfigurasi manual lewat "Kelola Tahap" — bukan hardcode, jadi bisa disesuaikan kapan saja.</p></div><div class="toolbar-actions"><a class="outline-button" href="` + attr(toggleTarget) + `" hx-get="` + attr(toggleTarget) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">` + toolIconHTML("settings-slider") + ` ` + toggleLabel + `</a><a class="primary-button" href="` + attr(exportPath) + `">Export</a></div></div>`
	if vm.ShowStageManager {
		toolbar += stageManagerPanel(vm.PipelineStages)
	}
	return toolbar
}

func clientSelectOptions(clients []dashboard.ClientProfile) string {
	var opts strings.Builder
	for _, client := range clients {
		opts.WriteString(`<option value="` + attr(client.ID) + `">` + esc(client.Name) + `</option>`)
	}
	return opts.String()
}

func expenseCategoryOptions(categories []string) string {
	var opts strings.Builder
	for _, category := range categories {
		opts.WriteString(`<option value="` + attr(category) + `">`)
	}
	return opts.String()
}

func toggleFormButton(showForm bool, label, basePath string) string {
	if showForm {
		return `<a class="outline-button" href="` + attr(basePath) + `" hx-get="` + attr(basePath) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Batal</a>`
	}
	newPath := basePath + "?new=1"
	return `<a class="primary-button" href="` + attr(newPath) + `" hx-get="` + attr(newPath) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">` + esc(label) + `</a>`
}

func staffTasks(vm dashboard.ViewModel) string {
	var rows strings.Builder
	for _, task := range vm.Tasks {
		action := `<button class="icon-action" disabled>Selesai</button>`
		if task.Status != dashboard.TaskDone {
			action = `<form method="post" action="/staff/tasks/` + attr(task.ID) + `/complete"><button class="icon-action" type="submit">Selesai</button></form>`
		}
		rows.WriteString(`<tr><td>` + esc(task.TimeLabel) + `</td><td>` + esc(task.Title) + `</td><td>` + esc(task.ClientName) + `</td><td><span class="status ` + priorityClass(task.Priority) + `">` + esc(task.Priority) + `</span></td><td><span class="status ` + taskStatusClass(task.Status) + `">` + esc(taskStatusLabel(task.Status)) + `</span></td><td>` + action + `</td></tr>`)
	}
	form := ""
	if vm.ShowCreateForm {
		form = `<form method="post" action="/staff/tasks/create" class="student-upload-form"><label>Client<select name="client_id" required><option value="">Pilih client</option>` + clientSelectOptions(vm.Clients) + `</select></label><label>Judul Tugas<input name="title" required placeholder="Contoh: Follow up dokumen"></label><label>Waktu<input name="time_label" placeholder="10:00"></label><label>Prioritas<select name="priority"><option>Rendah</option><option selected>Sedang</option><option>Tinggi</option></select></label><button class="primary-button" type="submit">Simpan Task</button></form>`
	}
	return `<div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(0,.45fr)]"><div class="panel table-panel"><div class="panel-head"><h2>Task & To Do List</h2>` + toggleFormButton(vm.ShowCreateForm, "Tambah Task", "/staff/tasks") + `</div>` + form + `<table class="data-table"><thead><tr><th>Waktu</th><th>Tugas</th><th>Klien</th><th>Prioritas</th><th>Status</th><th>Aksi</th></tr></thead><tbody>` + rows.String() + `</tbody></table></div><div class="right-rail">` + quickSchedule(vm) + recentDocs(vm) + `</div></div>`
}

func staffDocuments(vm dashboard.ViewModel) string {
	return documentsPanel(vm.Documents, true)
}

func staffExpenses(vm dashboard.ViewModel) string {
	form := ""
	if vm.ShowCreateForm {
		form = `<form method="post" action="/staff/expenses/create" enctype="multipart/form-data" class="student-upload-form"><label>Client<select name="client_id" required><option value="">Pilih client</option>` + clientSelectOptions(vm.Clients) + `</select></label><label>Keperluan / Berkas<input name="need" required placeholder="Contoh: Translate ijazah"></label><label>Kategori<input name="category" list="expense-category-list" placeholder="Pilih atau ketik kategori baru" autocomplete="off"><datalist id="expense-category-list">` + expenseCategoryOptions(vm.ExpenseCategories) + `</datalist></label><label>Nominal (IDR)<input name="amount" type="number" min="1" required placeholder="450000"></label><label>Tanggal<input name="date_label" placeholder="25/06/2026"></label><label>Keterangan<input name="description" placeholder="Catatan tambahan"></label><label>Bukti pengeluaran<input name="receipt_file" type="file"></label><button class="primary-button" type="submit">Simpan Pengeluaran</button></form>`
	}
	return `<div class="panel"><div class="toolbar-row mb-4"><div><h2 class="text-lg font-semibold">Pengeluaran Operasional</h2><p class="text-sm text-slate-500">Input kategori, nominal, bukti, dan keterangan untuk owner approval.</p></div>` + toggleFormButton(vm.ShowCreateForm, "+ Catat Pengeluaran", "/staff/expenses") + `</div>` + form + expensesTable(vm.Expenses) + `</div>`
}

func studentDashboard(vm dashboard.ViewModel) string {
	client := firstClient(vm.Clients)
	timeline := buildClientTimeline(vm.PipelineStages, vm.ClientStageHistory, client.CurrentStage)
	timelineDone := countTimelineDone(timeline)
	timelineTotal := len(timeline)
	timelinePct := timelinePercent(timeline)
	total, paid := orderTotals(vm.Orders)
	remaining := total - paid
	if remaining < 0 {
		remaining = 0
	}
	nextSchedule := firstSchedule(vm.Schedules)
	onboarding := ""
	if len(vm.Orders) == 0 || len(vm.Schedules) == 0 {
		onboarding = `<section class="student-onboarding"><strong>Akun client aktif</strong><span>Fitur dokumen dan chat dibuka setelah bukti pembayaran dikirim atau invoice lunas. Jadwal dan tahapan berikutnya akan diatur oleh konsultan.</span><a href="/student/payments" hx-get="/student/payments" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Cek Pembayaran</a></section>`
	}
	chatAction := lockedStudentAction(vm, "Chat", "/student/chat")
	docAction := studentActionLink("Upload", "/student/documents")
	return `
<div class="space-y-5">
  <section class="grid gap-5 xl:grid-cols-[minmax(0,1.1fr)_minmax(0,.9fr)_minmax(0,.8fr)]">
    <div class="panel student-hero"><span class="service-badge bg-emerald-100 text-emerald-700">` + esc(client.PackageName) + `</span><h2>` + esc(client.Name) + `</h2><p>` + esc(client.Campus) + `</p><div class="student-progress"><strong>` + intText(timelinePct) + `%</strong><span>` + esc(client.CurrentStage) + `</span></div></div>
    <div class="panel"><div class="panel-head"><h2>Progress Keseluruhan</h2><a href="/student/progress" hx-get="/student/progress" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Detail</a></div><div class="progress-ring" style="` + progressRingStyle(timelinePct) + `"><span>` + intText(timelinePct) + `%</span></div><p class="mt-4 text-sm text-slate-500">Tahap saat ini: ` + esc(client.CurrentStage) + `. ` + intText(timelineDone) + ` dari ` + intText(timelineTotal) + ` tahap selesai.</p></div>
    <div class="panel"><div class="panel-head"><h2>Konsultan</h2></div><div class="consultant-card"><span class="avatar-mini bg-emerald-100 text-emerald-700">` + initials(client.PICName) + `</span><div><strong>` + esc(client.PICName) + `</strong><p>Staff Konsultan</p></div></div><div class="action-grid">` + chatAction + `</div></div>
  </section>
  <section class="metric-grid metric-grid-four">
    <article class="metric-card"><div><p>Total Tagihan</p><strong>` + money(total) + `</strong><span>Semua invoice milik kamu</span></div></article>
    <article class="metric-card"><div><p>Sudah Dibayar</p><strong>` + money(paid) + `</strong><span>Order lunas/tercatat</span></div></article>
    <article class="metric-card"><div><p>Sisa Tagihan</p><strong>` + money(remaining) + `</strong><span>` + intText(countOpenOrders(vm.Orders)) + ` invoice belum lunas</span></div></article>
    <article class="metric-card"><div><p>Dokumen Review</p><strong>` + intText(countReviewDocs(vm.Documents)) + `</strong><span>` + intText(countApprovedDocs(vm.Documents)) + ` disetujui</span></div></article>
  </section>
  ` + onboarding + `
  <section class="grid gap-5 xl:grid-cols-[minmax(0,1.2fr)_minmax(0,.8fr)]">
    <div class="panel table-panel"><div class="panel-head"><h2>Dokumen Saya</h2>` + docAction + `</div>` + documentsTable(vm.Documents, false) + `</div>
    <div class="right-rail">` + studentPaymentCard(vm) + studentNextSchedule(nextSchedule) + `</div>
  </section>
</div>`
}

func studentProgress(vm dashboard.ViewModel) string {
	client := firstClient(vm.Clients)
	timeline := buildClientTimeline(vm.PipelineStages, vm.ClientStageHistory, client.CurrentStage)
	active := activeTimelineEntry(timeline)
	completed := countTimelineDone(timeline)
	percent := timelinePercent(timeline)

	var items strings.Builder
	for _, entry := range timeline {
		statusLine := stageTimelineLabel(entry.Status)
		if entry.Status == "selesai" && entry.DateLabel != "" {
			statusLine += " &middot; " + esc(entry.DateLabel)
		}
		items.WriteString(`<div class="stage-timeline-item ` + attr(entry.Status) + `"><span class="stage-timeline-dot"></span><div><strong>` + esc(entry.Name) + `</strong><span>` + statusLine + `</span></div></div>`)
	}
	if items.Len() == 0 {
		items.WriteString(`<p class="empty-note">Belum ada tahap yang ditentukan. Owner/staff bisa mengatur daftar tahap lewat Pipeline &gt; Kelola Tahap.</p>`)
	}

	statusNote := `<div class="student-onboarding compact mt-5"><strong>` + esc(active.Name) + ` sedang berjalan</strong><span>Tahap ini akan otomatis ditandai selesai saat owner/staff memindahkan kartu kamu ke tahap berikutnya di Pipeline.</span></div>`
	if len(timeline) > 0 && percent == 100 {
		statusNote = `<div class="student-onboarding compact mt-5"><strong>Semua tahap selesai</strong><span>Perjalanan kamu di Pipeline sudah mencapai tahap terakhir.</span></div>`
	}
	return `
<div class="panel">
  <div class="panel-head"><div><h2>Progress keseluruhan</h2><p>` + intText(percent) + `% &middot; ` + intText(completed) + ` dari ` + intText(len(timeline)) + ` tahap selesai</p></div>` + studentActionLink("Upload Dokumen", "/student/documents") + `</div>
  <div class="stage-progress"><i style="width:` + percentStyle(percent) + `"></i></div>
  <div class="stage-timeline">` + items.String() + `</div>
  ` + statusNote + `
</div>`
}

func studentDocuments(vm dashboard.ViewModel) string {
	return `<div class="grid gap-5 xl:grid-cols-[minmax(0,.8fr)_minmax(0,1.2fr)]"><section class="panel"><div class="panel-head"><div><h2>Upload Dokumen</h2><p>Unggah file sesuai permintaan konsultan.</p></div></div><form method="post" action="/student/documents/upload" enctype="multipart/form-data" class="student-upload-form"><label>Jenis dokumen<select name="document_name" required><option>Passport</option><option>Ijazah Terakhir</option><option>Transkrip Nilai</option><option>Foto Background Putih 4x6</option><option>KTP</option><option>KK</option><option>Akta Kelahiran</option><option>Autobiografi</option><option>Mutasi Rekening</option><option>Surat Rekomendasi</option></select></label><label>File dokumen<input name="document_file" type="file" required></label><p>Format PDF, JPG, atau PNG. File yang masuk akan tampil sebagai status review.</p><button class="primary-button" type="submit">Upload Dokumen</button></form></section><section class="panel table-panel"><div class="panel-head"><h2>Dokumen Saya</h2></div>` + documentsTable(vm.Documents, false) + `</section></div>`
}

func studentPayments(vm dashboard.ViewModel) string {
	if len(vm.Orders) == 0 {
		return `<div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(0,.75fr)]"><section class="panel empty-state"><span>Invoice</span><h2>Belum ada invoice aktif</h2><p>Pembayaran akan muncul setelah owner atau staff mengonfirmasi paket dan membuat kode pesanan kamu.</p></section><aside class="panel"><div class="panel-head"><h2>Status Pembayaran</h2></div><div class="summary-list compact"><div><span>Total Tagihan</span><strong>Rp 0</strong></div><div><span>Sudah Dibayar</span><strong>Rp 0</strong></div><div><span>Sisa Tagihan</span><strong>Rp 0</strong></div></div></aside></div>`
	}
	active := selectOrder(vm.Orders, vm.ActiveOrderCode)
	var invoiceList strings.Builder
	for _, order := range vm.Orders {
		className := ""
		if order.Code == active.Code {
			className = ` class="active"`
		}
		href := "/student/payments?order=" + url.QueryEscape(order.Code)
		invoiceList.WriteString(`<a` + className + ` href="` + attr(href) + `" hx-get="` + attr(href) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">` + esc(order.Code) + ` <span>` + esc(orderStatusLabel(order.Status)) + `</span><strong>` + money(order.Total) + `</strong></a>`)
	}
	remaining := active.Total - active.Paid
	if remaining < 0 {
		remaining = 0
	}
	paymentAction := `<form method="post" action="/student/payments/proof" enctype="multipart/form-data" class="payment-proof-form mt-5"><input type="hidden" name="order_code" value="` + attr(active.Code) + `"><label>Jumlah ditransfer (IDR)<input name="amount" type="number" min="1" required placeholder="Contoh: 2000000"></label><label>Catatan transfer manual</label><textarea name="note" rows="3" placeholder="Contoh: Transfer bank atas nama kamu, tanggal transfer"></textarea><label>Bukti transfer<input name="proof_file" type="file"></label><div class="action-grid mt-5"><a class="outline-button" href="/student/payments/` + attr(active.Code) + `/invoice.pdf">Download Invoice PDF</a><button class="success-button" type="submit">Kirim Bukti Bayar</button></div></form>`
	if remaining == 0 || active.Status == dashboard.OrderPaid {
		paymentAction = `<div class="student-onboarding compact"><strong>Pembayaran sudah lunas</strong><span>Invoice ini sudah tercatat lunas di sistem.</span><a href="/student/payments/` + attr(active.Code) + `/invoice.pdf">Download Invoice PDF</a></div>`
	} else if active.Status == dashboard.OrderWaitingVerification {
		paymentAction = `<div class="student-onboarding compact"><strong>Cicilan sedang diverifikasi</strong><span>Tim akan memperbarui status setelah cicilan yang kamu kirim dicek.</span><a href="/student/payments/` + attr(active.Code) + `/invoice.pdf">Download Invoice PDF</a></div>` + paymentAction
	}

	var paymentRows strings.Builder
	for _, payment := range vm.OrderPayments {
		if payment.OrderID != active.ID {
			continue
		}
		note := payment.Note
		if note == "" {
			note = "-"
		}
		proofCell := "-"
		if payment.ProofStoragePath != "" {
			proofCell = `<a href="/payments/` + attr(payment.ID) + `/proof" target="_blank" rel="noopener">Lihat</a>`
		}
		paymentRows.WriteString(`<tr><td>` + esc(dateLabel(payment.SubmittedAt)) + `<span class="block text-xs text-slate-500">` + esc(note) + `</span></td><td>` + money(payment.Amount) + `</td><td><span class="status ` + paymentStatusClass(payment.Status) + `">` + paymentStatusLabel(payment.Status) + `</span></td><td>` + proofCell + `</td></tr>`)
	}
	if paymentRows.Len() == 0 {
		paymentRows.WriteString(`<tr><td colspan="4">Belum ada cicilan.</td></tr>`)
	}
	paymentsTable := `<div class="invoice-box"><h3>Riwayat Cicilan</h3><table class="data-table mt-3"><thead><tr><th>Tanggal &amp; Catatan</th><th>Jumlah</th><th>Status</th><th>Bukti</th></tr></thead><tbody>` + paymentRows.String() + `</tbody></table></div>`
	paidPercent := orderPaymentPercent(active)
	progressBlock := `<div class="invoice-box"><h3>Progress Pembayaran</h3><div class="stage-progress"><i style="width:` + percentStyle(paidPercent) + `"></i></div><p class="text-xs text-slate-500 mt-1">` + intText(paidPercent) + `% dari total tagihan sudah dibayar</p></div>`

	return `<div class="grid gap-5 xl:grid-cols-[360px_minmax(0,1fr)]"><aside class="panel"><h2 class="mb-4 text-base font-semibold">Daftar Invoice</h2><div class="invoice-list">` + invoiceList.String() + `</div></aside><section class="panel"><div class="invoice-head"><div><p>Detail Invoice</p><strong>` + esc(active.Code) + `</strong></div><strong class="text-2xl">` + money(active.Total) + `</strong></div>` + progressBlock + `<dl class="detail-grid"><div><dt>Client</dt><dd>` + esc(active.ClientName) + `</dd></div><div><dt>Paket</dt><dd>` + esc(active.PackageName) + `</dd></div><div><dt>Status</dt><dd><span class="status ` + orderStatusClass(active.Status) + `">` + esc(orderStatusLabel(active.Status)) + `</span></dd></div><div><dt>Jatuh Tempo</dt><dd>` + esc(dateLabel(active.DueDate)) + `</dd></div></dl><div class="invoice-box"><h3>Rincian</h3><div><span>Subtotal</span><strong>` + money(active.Total) + `</strong></div><div><span>Sudah Dibayar</span><strong>` + money(active.Paid) + `</strong></div><div class="total"><span>Sisa</span><strong>` + money(remaining) + `</strong></div></div>` + paymentsTable + paymentAction + `</section></div>`
}

func settingsPanel(vm dashboard.ViewModel, role string, includeFinance bool) string {
	finance := ""
	if includeFinance {
		finance = `<div><span>Keuangan</span><strong>Owner only</strong><em>Aktif</em></div>`
	}
	dangerZone := ""
	if role == "Owner" {
		dangerZone = `<div class="panel" style="border-color:#fecaca">
  <h2 class="mb-2 text-lg font-semibold text-red-600">Reset Data Test</h2>
  <p class="text-sm leading-6 text-slate-600 mb-4">Menghapus permanen: akun client, order/invoice, dokumen, progress, jadwal, task, pengeluaran, pengiriman logistik, chat, formulir data diri, aktivitas, dan perjanjian yang sudah ditandatangani. <strong>Tidak</strong> menghapus akun owner/staff maupun konfigurasi (paket layanan, tahap pipeline, template teks, direktori institusi, kategori pengeluaran, daftar ekspedisi). Aksi ini permanen dan tidak bisa dibatalkan — pastikan sudah backup database sebelum lanjut.</p>
  <form method="post" action="/owner/settings/reset-data" class="student-upload-form" data-confirm="Yakin reset semua data client &amp; transaksi test? Aksi ini permanen dan tidak bisa dibatalkan.">
    <label>Ketik &quot;RESET DATA&quot; untuk konfirmasi<input name="confirm_phrase" required placeholder="RESET DATA"></label>
    <label>Password Owner<input name="password" type="password" required></label>
    <button class="primary-button" type="submit" style="background:#dc2626;border-color:#dc2626">Reset Data Test</button>
  </form>
</div>`
	}

	viewer := vm.Viewer
	profilePanel := `<div class="panel"><h2 class="mb-4 text-lg font-semibold">Edit Profil</h2><form method="post" action="/profile/update" class="student-upload-form"><label>Nama Lengkap<input name="name" value="` + attr(viewer.Name) + `" required maxlength="191"></label><label>Username<input value="` + attr(viewer.Username) + `" disabled></label><label>Email<input name="email" type="email" value="` + attr(viewer.Email) + `" maxlength="191"></label><label>No. WhatsApp<input name="phone" value="` + attr(viewer.Phone) + `" maxlength="64" placeholder="08xxxxxxxxxx"></label><button class="primary-button" type="submit">Simpan Profil</button></form></div>`
	passwordPanel := `<div class="panel"><h2 class="mb-4 text-lg font-semibold">Ubah Password</h2><form method="post" action="/profile/change-password" class="student-upload-form"><label>Password Saat Ini<input name="current_password" type="password" autocomplete="current-password" required></label><label>Password Baru<input name="new_password" type="password" autocomplete="new-password" minlength="8" required></label><label>Konfirmasi Password Baru<input name="confirm_password" type="password" autocomplete="new-password" minlength="8" required></label><button class="primary-button" type="submit">Ubah Password</button></form></div>`

	return `<div class="space-y-5">
  <div style="display:grid;gap:1.25rem;grid-template-columns:repeat(auto-fit,minmax(320px,1fr))">` + profilePanel + passwordPanel + `</div>
  <div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(0,.8fr)]"><div class="panel"><h2 class="mb-4 text-lg font-semibold">Preferensi ` + esc(role) + `</h2><div class="settings-list"><div><span>Autentikasi</span><strong>Password + session aman</strong><em>Aktif</em></div><div><span>Notifikasi</span><strong>Email dan WhatsApp reminder</strong><em>Aktif</em></div><div><span>PWA</span><strong>Installable di Android</strong><em>Aktif</em></div>` + finance + `<div><span>SEO</span><strong>Metadata per halaman</strong><em>Aktif</em></div></div></div><div class="panel"><h2 class="mb-4 text-lg font-semibold">Keamanan</h2><p class="text-sm leading-6 text-slate-600">Aplikasi menggunakan security headers, konfigurasi via environment, timeout server, dan graceful shutdown. Hak akses disiapkan per role agar staff tidak melihat laporan profit dan client hanya melihat data miliknya sendiri.</p></div></div>
  ` + dangerZone + `
</div>`
}

func calendarPanel(vm dashboard.ViewModel, title string) string {
	form := ""
	if vm.ShowCreateForm {
		form = `<form method="post" action="/staff/calendar/create" class="student-upload-form"><label>Client<select name="client_id" required><option value="">Pilih client</option>` + clientSelectOptions(vm.Clients) + `</select></label><label>Judul Jadwal<input name="title" required placeholder="Contoh: Interview Visa"></label><label>Tanggal<input name="date_label" placeholder="26 Jun 2026"></label><label>Waktu<input name="time_label" placeholder="10:00"></label><label>Lokasi<input name="location" placeholder="Kedutaan Taiwan"></label><button class="primary-button" type="submit">Simpan Jadwal</button></form>`
	}
	var rows strings.Builder
	for _, item := range vm.Schedules {
		rows.WriteString(`<div><strong>` + esc(item.TimeLabel) + `</strong><span>` + esc(item.Title) + `</span><em>` + esc(item.DateLabel) + ` - ` + esc(item.Location) + `</em></div>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<p class="empty-note">Belum ada jadwal.</p>`)
	}
	return `<div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(0,.75fr)]"><div class="panel"><div class="panel-head"><h2>` + esc(title) + `</h2>` + toggleFormButton(vm.ShowCreateForm, "Tambah Jadwal", "/staff/calendar") + `</div>` + form + `<div class="calendar-grid">` + rows.String() + `</div></div>` + quickSchedule(vm) + `</div>`
}

func studentCalendar(vm dashboard.ViewModel) string {
	if len(vm.Schedules) == 0 {
		return `<div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(0,.75fr)]"><section class="panel"><div class="panel-head"><div><h2>Alur Penjadwalan</h2><p>Agenda akan aktif setelah konsultan menetapkan tanggal.</p></div></div><div class="schedule-steps"><div><strong>1</strong><span>Konsultasi awal</span><em>Validasi target kampus dan intake.</em></div><div><strong>2</strong><span>Deadline dokumen</span><em>Tenggat upload berkas yang dibutuhkan.</em></div><div><strong>3</strong><span>Appointment / interview</span><em>Jadwal kampus, medical, visa, atau kedutaan.</em></div><div><strong>4</strong><span>Keberangkatan</span><em>Final briefing, tiket, dan persiapan Taiwan.</em></div></div></section>` + studentScheduleList(vm.Schedules) + `</div>`
	}
	var rows strings.Builder
	for _, item := range vm.Schedules {
		rows.WriteString(`<div><strong>` + esc(item.TimeLabel) + `</strong><span>` + esc(item.Title) + `</span><em>` + esc(item.DateLabel) + ` - ` + esc(item.Location) + `</em></div>`)
	}
	return `<div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(0,.75fr)]"><div class="panel"><div class="panel-head"><h2>Jadwal Saya</h2></div><div class="calendar-grid">` + rows.String() + `</div></div>` + studentScheduleList(vm.Schedules) + `</div>`
}

func lockedStudentAction(vm dashboard.ViewModel, label, href string) string {
	if vm.StudentFeatureAccess {
		return studentActionLink(label, href)
	}
	return `<a class="outline-button locked-action" href="/student/payments" hx-get="/student/payments" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">` + esc(label) + `</a>`
}

// studentActionLink is the always-unlocked variant of lockedStudentAction —
// document upload isn't gated behind payment status the way chat is, so a
// client who hasn't paid or submitted a cicilan yet can still get their
// documents in early.
func studentActionLink(label, href string) string {
	return `<a class="success-button" href="` + attr(href) + `" hx-get="` + attr(href) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">` + esc(label) + `</a>`
}

func lockedStudentFeature(title, body string) string {
	return `<section class="panel empty-state"><span>Akses terkunci</span><h2>` + esc(title) + `</h2><p>` + esc(body) + `</p><a class="success-button" href="/student/payments" hx-get="/student/payments" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Kirim Bukti Bayar</a></section>`
}

func chatPanel(vm dashboard.ViewModel, title string) string {
	if vm.ActiveChatID == "" {
		return `<div class="panel"><h2 class="mb-4 text-base font-semibold">` + esc(title) + `</h2><p class="text-sm text-slate-500">Belum ada percakapan aktif.</p></div>`
	}
	var conversations strings.Builder
	for _, conversation := range vm.Conversations {
		className := ""
		if conversation.ID == vm.ActiveChatID {
			className = ` class="active"`
		}
		href := "/" + vm.Role.String() + "/chat?conversation=" + conversation.ID
		conversations.WriteString(`<a` + className + ` href="` + attr(href) + `" hx-get="` + attr(href) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true"><strong>` + esc(conversation.ClientName) + `</strong><span>` + esc(conversation.LastMessage) + `</span></a>`)
	}
	var messages strings.Builder
	for _, message := range vm.Messages {
		className := "other"
		if message.SenderID == vm.Viewer.ID {
			className = "self"
		}
		messages.WriteString(`<div class="message ` + className + `" data-message-id="` + attr(message.ID) + `"><strong>` + esc(message.SenderName) + `</strong><span>` + esc(message.Body) + `</span></div>`)
	}
	return `<div class="grid gap-5 xl:grid-cols-[340px_minmax(0,1fr)]" data-chat data-conversation-id="` + attr(vm.ActiveChatID) + `" data-user-id="` + attr(vm.Viewer.ID) + `"><aside class="panel"><h2 class="mb-4 text-base font-semibold">` + esc(title) + `</h2><div class="chat-list">` + conversations.String() + `</div></aside><section class="panel"><div class="chat-window" data-chat-window>` + messages.String() + `</div><form class="chat-input" method="post" action="/chat/` + attr(vm.ActiveChatID) + `/messages" data-chat-form><input name="body" autocomplete="off" placeholder="Tulis pesan..." aria-label="Tulis pesan" required maxlength="2000"><button class="primary-button" type="submit">Kirim</button></form></section></div>`
}

func documentsPanel(documents []dashboard.Document, actionable bool) string {
	bulkAction := ""
	extras := ""
	if actionable {
		if countReviewDocs(documents) > 0 {
			bulkAction = `<form method="post" action="/staff/documents/bulk-approve" class="inline-form"><button class="primary-button" type="submit">Review Massal (` + intText(countReviewDocs(documents)) + `)</button></form>`
		}
		extras = rejectModal()
	}
	return `<div class="panel table-panel"><div class="panel-head"><h2>Dokumen Menunggu Review</h2>` + bulkAction + `</div>` + documentsTable(documents, actionable) + `</div>` + extras
}

func rejectModal() string {
	return `<div class="modal-overlay" data-reject-modal hidden><div class="modal-card"><h3>Tolak Dokumen</h3><p class="modal-sub" data-reject-target></p><form method="post" data-reject-form><label>Alasan penolakan</label><textarea name="reason" rows="3" maxlength="200" required placeholder="Contoh: Foto buram, mohon scan ulang yang lebih jelas"></textarea><div class="modal-actions"><button type="button" class="outline-button" data-reject-cancel>Batal</button><button type="submit" class="danger-button">Tolak Dokumen</button></div></form></div></div>`
}

func documentsTable(documents []dashboard.Document, actionable bool) string {
	var rows strings.Builder
	for _, document := range documents {
		action := `<span class="doc-nofile">Belum ada file</span>`
		if actionable && document.Status == dashboard.DocumentReview {
			viewFile := ""
			if document.StoragePath != "" {
				viewFile = `<a class="icon-btn icon-btn-view" href="/student/documents/` + attr(document.ID) + `/download" title="Lihat file" aria-label="Lihat file">` + svgIcon("eye") + `</a>`
			}
			rejectName := document.ClientName + " - " + document.Name
			action = `<div class="doc-review-actions"><form method="post" action="/staff/documents/` + attr(document.ID) + `/approve" class="inline-form"><button class="icon-btn icon-btn-approve" type="submit" title="Approve" aria-label="Approve">` + svgIcon("check-plain") + `</button></form>` + viewFile + `<button type="button" class="icon-btn icon-btn-reject" data-reject-url="/staff/documents/` + attr(document.ID) + `/reject" data-reject-name="` + attr(rejectName) + `" title="Reject" aria-label="Reject">` + svgIcon("close") + `</button></div>`
		} else if document.StoragePath != "" {
			action = `<a class="icon-btn icon-btn-view" href="/student/documents/` + attr(document.ID) + `/download" title="Download" aria-label="Download">` + svgIcon("download") + `</a>`
		}
		fileLabel := document.FileName
		if fileLabel == "" {
			fileLabel = "-"
		}
		docCell := esc(document.Name) + `<span>` + esc(fileLabel) + `</span>`
		if document.Status == dashboard.DocumentRevision && document.ReviewNote != "" {
			docCell += `<span class="doc-note">Alasan revisi: ` + esc(document.ReviewNote) + `</span>`
		}
		rows.WriteString(`<tr><td>` + esc(document.ClientName) + `</td><td>` + docCell + `</td><td>` + esc(dateLabel(document.UpdatedAt)) + `</td><td><span class="status ` + documentStatusClass(document.Status) + `">` + esc(documentStatusLabel(document.Status)) + `</span></td><td>` + esc(document.Reviewer) + `</td><td>` + action + `</td></tr>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<tr><td colspan="6">Belum ada dokumen.</td></tr>`)
	}
	return `<table class="data-table"><thead><tr><th>Client</th><th>Dokumen</th><th>Update</th><th>Status</th><th>Reviewer</th><th>Aksi</th></tr></thead><tbody>` + rows.String() + `</tbody></table>`
}

func studentPaymentCard(vm dashboard.ViewModel) string {
	total, paid := orderTotals(vm.Orders)
	remaining := total - paid
	if remaining < 0 {
		remaining = 0
	}
	if len(vm.Orders) == 0 {
		return `<div class="panel"><div class="panel-head"><h2>Pembayaran</h2><a href="/student/payments" hx-get="/student/payments" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Detail</a></div><div class="summary-list compact"><div><span>Total Tagihan</span><strong>Rp 0</strong></div><div><span>Sudah Dibayar</span><strong>Rp 0</strong></div><div><span>Sisa Tagihan</span><strong>Rp 0</strong></div></div><p class="empty-note">Belum ada invoice aktif.</p></div>`
	}
	if remaining == 0 {
		return `<div class="panel"><div class="panel-head"><h2>Pembayaran</h2><a href="/student/payments" hx-get="/student/payments" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Detail</a></div><div class="summary-list compact"><div><span>Total Tagihan</span><strong>` + money(total) + `</strong></div><div><span>Sudah Dibayar</span><strong>` + money(paid) + `</strong></div><div><span>Sisa Tagihan</span><strong>` + money(remaining) + `</strong></div></div><p class="empty-note">Pembayaran kamu sudah lunas.</p></div>`
	}
	return `<div class="panel"><div class="panel-head"><h2>Pembayaran</h2><a href="/student/payments" hx-get="/student/payments" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Detail</a></div><div class="summary-list compact"><div><span>Total Tagihan</span><strong>` + money(total) + `</strong></div><div><span>Sudah Dibayar</span><strong>` + money(paid) + `</strong></div><div><span>Sisa Tagihan</span><strong>` + money(remaining) + `</strong></div></div><a href="/student/payments" hx-get="/student/payments" hx-target="#app" hx-swap="outerHTML" hx-push-url="true" class="success-button w-full mt-4">Upload Bukti Bayar</a></div>`
}

func studentAnnouncement() string {
	return `<div class="panel"><div class="panel-head"><h2>Pengumuman</h2></div><ul class="reminder-list clean"><li>Update Visa: dokumen apply dibuka awal Juli.</li><li>Intake Februari 2027 masih menerima pendaftar.</li><li>NTU meminta autobiografi versi final.</li></ul></div>`
}

func studentNextSchedule(item dashboard.ScheduleItem) string {
	if item.ID == "" {
		return `<div class="panel"><div class="panel-head"><h2>Jadwal Terdekat</h2><a href="/student/calendar" hx-get="/student/calendar" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Detail</a></div><p class="empty-note">Belum ada jadwal aktif. Agenda akan tampil setelah staff menambahkannya.</p></div>`
	}
	return `<div class="panel"><div class="panel-head"><h2>Jadwal Terdekat</h2><a href="/student/calendar" hx-get="/student/calendar" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Detail</a></div><div class="summary-list compact"><div><span>` + esc(item.DateLabel) + `</span><strong>` + esc(item.Title) + `</strong></div><div><span>Jam</span><strong>` + esc(item.TimeLabel) + `</strong></div><div><span>Lokasi</span><strong>` + esc(item.Location) + `</strong></div></div></div>`
}

func studentScheduleList(schedules []dashboard.ScheduleItem) string {
	var rows strings.Builder
	for _, item := range schedules {
		rows.WriteString(`<div><span>` + esc(item.DateLabel) + `</span><strong>` + esc(item.Title) + `</strong><em>` + esc(item.TimeLabel) + ` - ` + esc(item.Status) + `</em></div>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<div><span>-</span><strong>Belum ada jadwal</strong><em>Agenda akan tampil setelah staff menambahkan jadwal.</em></div>`)
	}
	return `<aside class="panel"><div class="panel-head"><h2>Agenda</h2></div><div class="summary-list">` + rows.String() + `</div></aside>`
}

func firstSchedule(schedules []dashboard.ScheduleItem) dashboard.ScheduleItem {
	if len(schedules) == 0 {
		return dashboard.ScheduleItem{}
	}
	return schedules[0]
}

// stageTimelineEntry is the client-facing render shape for one pipeline
// stage: unlike the old ProgressStage table, nothing here is stored —
// Status and DateLabel are derived fresh every render from PipelineStages
// (the owner's "Kelola Tahap" list) plus ClientStageHistory, so the client
// timeline can never drift out of sync with the owner's Pipeline board the
// way progress_stages used to.
type stageTimelineEntry struct {
	Name      string
	Status    string // "selesai" | "berjalan" | "belum"
	DateLabel string // formatted completion date, or "" if not recorded
}

// buildClientTimeline derives every stage's status for one client by
// comparing its position in the shared stage order to the client's current
// stage — done if earlier, active if it matches, pending otherwise. A
// "selesai" stage's date is when the client left it (the entered_at of the
// next stage they reached); it's blank for transitions that happened
// before client_stage_history existed (see
// ensureClientStageHistoryBackfilled), which is an honest gap rather than
// a guessed date.
func buildClientTimeline(stages []dashboard.PipelineStage, history []dashboard.ClientStageEvent, currentStage string) []stageTimelineEntry {
	currentIndex := -1
	for i, stage := range stages {
		if stage.Name == currentStage {
			currentIndex = i
			break
		}
	}
	entries := make([]stageTimelineEntry, 0, len(stages))
	for i, stage := range stages {
		entry := stageTimelineEntry{Name: stage.Name, Status: "belum"}
		switch {
		case currentIndex < 0:
			// unassigned client — every stage stays "belum"
		case i < currentIndex, i == currentIndex && i == len(stages)-1:
			// being on the very last configured stage counts as fully
			// done, same special case UpdateClientStage persists into
			// clients.progress for the same reason
			entry.Status = "selesai"
			entry.DateLabel = nextStageEntryDate(history, stage.Name)
		case i == currentIndex:
			entry.Status = "berjalan"
		}
		entries = append(entries, entry)
	}
	return entries
}

// nextStageEntryDate finds when a client left stageName: the entered_at of
// whichever history row comes right after the one for stageName.
func nextStageEntryDate(history []dashboard.ClientStageEvent, stageName string) string {
	for i, event := range history {
		if event.StageName == stageName && i+1 < len(history) {
			return dateLabel(history[i+1].EnteredAt)
		}
	}
	return ""
}

func activeTimelineEntry(entries []stageTimelineEntry) stageTimelineEntry {
	for _, entry := range entries {
		if entry.Status == "berjalan" {
			return entry
		}
	}
	return stageTimelineEntry{Name: "Belum ada tahap", Status: "belum"}
}

func countTimelineDone(entries []stageTimelineEntry) int {
	count := 0
	for _, entry := range entries {
		if entry.Status == "selesai" {
			count++
		}
	}
	return count
}

// timelinePercent mirrors the persisted clients.progress formula in
// UpdateClientStage (index/total, 100% on the last stage) so the ring %
// shown here always agrees with what's stored — computed live from the
// same timeline being rendered right below it, rather than trusting the
// stored column, so the two numbers on this page can never disagree with
// each other.
func timelinePercent(entries []stageTimelineEntry) int {
	if len(entries) == 0 {
		return 0
	}
	return countTimelineDone(entries) * 100 / len(entries)
}

func stageTimelineLabel(status string) string {
	switch status {
	case "selesai":
		return "Selesai"
	case "berjalan":
		return "Sedang Berjalan"
	default:
		return "Belum Dimulai"
	}
}

func stageTimelineClass(status string) string {
	switch status {
	case "selesai":
		return "green"
	case "berjalan":
		return "blue"
	default:
		return "slate"
	}
}

func taskListCompact(tasks []dashboard.Task) string {
	var b strings.Builder
	b.WriteString(`<div class="task-list">`)
	for _, task := range tasks {
		b.WriteString(`<div><strong>` + esc(task.TimeLabel) + `</strong><span>` + esc(task.Title) + `</span><em>` + esc(task.ClientName) + `</em><b>` + esc(task.Priority) + `</b></div>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}

// pipelineColumns renders the kanban board from vm.PipelineStages — the set of
// stages owners/staff configure themselves (see stageManagerPanel) — rather
// than a fixed list baked into the code. Clients whose CurrentStage doesn't
// match any configured stage (e.g. the stage was renamed or deleted) land in
// an implicit "Belum Ditentukan" column instead of disappearing from the board.
func pipelineColumns(vm dashboard.ViewModel) string {
	stages := vm.PipelineStages
	stageNames := make(map[string]bool, len(stages))
	for _, stage := range stages {
		stageNames[stage.Name] = true
	}
	grouped := make(map[string][]dashboard.ClientProfile, len(stages))
	var unassigned []dashboard.ClientProfile
	for _, client := range vm.Clients {
		if client.CurrentStage != "" && stageNames[client.CurrentStage] {
			grouped[client.CurrentStage] = append(grouped[client.CurrentStage], client)
			continue
		}
		unassigned = append(unassigned, client)
	}
	moreHref := "/" + vm.Role.String() + "/clients"

	var b strings.Builder
	b.WriteString(`<section class="kanban-board">`)
	if len(unassigned) > 0 {
		b.WriteString(kanbanColumn(dashboard.PipelineStage{Name: "Belum Ditentukan", Tone: "slate"}, unassigned, stages, moreHref))
	}
	for _, stage := range stages {
		b.WriteString(kanbanColumn(stage, grouped[stage.Name], stages, moreHref))
	}
	if len(stages) == 0 && len(unassigned) == 0 {
		b.WriteString(`<p class="empty-note">Belum ada tahap pipeline yang dikonfigurasi.</p>`)
	}
	b.WriteString(`</section>`)
	return b.String()
}

func kanbanColumn(stage dashboard.PipelineStage, clients []dashboard.ClientProfile, allStages []dashboard.PipelineStage, moreHref string) string {
	const maxShown = 4
	shown := clients
	remaining := 0
	if len(clients) > maxShown {
		shown = clients[:maxShown]
		remaining = len(clients) - maxShown
	}
	tone := stage.Tone
	if tone == "" {
		tone = "slate"
	}
	var b strings.Builder
	b.WriteString(`<div class="kanban-column tone-` + attr(tone) + `"><div class="kanban-title"><h3>` + esc(stage.Name) + `</h3><span>` + intText(len(clients)) + ` Klien</span></div><div class="kanban-cards">`)
	for _, client := range shown {
		b.WriteString(`<article><span class="avatar-mini bg-slate-100 text-slate-700">` + initials(client.Name) + `</span><div><strong>` + esc(client.Name) + `</strong><p>` + esc(client.PackageName) + `</p><em>` + esc(client.PICName) + `</em>` + stageMoveSelect(client, allStages) + `</div></article>`)
	}
	if len(clients) == 0 {
		b.WriteString(`<p class="empty-note">Belum ada client di tahap ini.</p>`)
	} else if remaining > 0 {
		b.WriteString(`<a class="more-card" href="` + attr(moreHref) + `" hx-get="` + attr(moreHref) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">+ ` + intText(remaining) + ` lainnya</a>`)
	}
	b.WriteString(`</div></div>`)
	return b.String()
}

func stageMoveSelect(client dashboard.ClientProfile, stages []dashboard.PipelineStage) string {
	if len(stages) == 0 {
		return ""
	}
	var opts strings.Builder
	opts.WriteString(`<option value="" disabled` + selectedIf(client.CurrentStage == "") + `>Pindah ke tahap...</option>`)
	for _, stage := range stages {
		opts.WriteString(`<option value="` + attr(stage.Name) + `"` + selectedIf(stage.Name == client.CurrentStage) + `>` + esc(stage.Name) + `</option>`)
	}
	action := "/pipeline/clients/" + attr(client.ID) + "/stage"
	return `<form method="post" action="` + action + `" class="stage-move-form"><select name="stage_name" data-autosubmit aria-label="Pindahkan tahap ` + attr(client.Name) + `">` + opts.String() + `</select></form>`
}

func selectedIf(cond bool) string {
	if cond {
		return " selected"
	}
	return ""
}

func toneHex(tone string) string {
	switch tone {
	case "mint":
		return "#22c55e"
	case "rose":
		return "#f43f5e"
	case "emerald":
		return "#10b981"
	case "sky":
		return "#0ea5e9"
	case "violet":
		return "#8b5cf6"
	case "lime":
		return "#84cc16"
	case "amber":
		return "#f59e0b"
	case "blue":
		return "#3b82f6"
	case "teal":
		return "#14b8a6"
	default:
		return "#94a3b8"
	}
}

// stageManagerPanel lets an owner or staff member define the pipeline stages
// themselves — add, rename, reorder, or delete — with no code change and no
// developer involvement, addressing pipelines whose shape isn't known up front.
func stageManagerPanel(stages []dashboard.PipelineStage) string {
	var rows strings.Builder
	for i, stage := range stages {
		upAttr := ""
		if i == 0 {
			upAttr = " disabled"
		}
		downAttr := ""
		if i == len(stages)-1 {
			downAttr = " disabled"
		}
		confirmMsg := "Hapus tahap \"" + stage.Name + "\"? Client di tahap ini akan dipindah ke Belum Ditentukan."
		rows.WriteString(`<div class="stage-manager-row">`)
		rows.WriteString(`<span class="stage-swatch" style="background:` + toneHex(stage.Tone) + `"></span>`)
		rows.WriteString(`<form method="post" action="/pipeline/stages/` + attr(stage.ID) + `/rename" class="rename-form"><input name="name" value="` + attr(stage.Name) + `" required maxlength="80" aria-label="Nama tahap"><button class="outline-button small" type="submit">Simpan</button></form>`)
		rows.WriteString(`<div class="stage-manager-actions">`)
		rows.WriteString(`<form method="post" action="/pipeline/stages/` + attr(stage.ID) + `/move" class="inline-form"><input type="hidden" name="direction" value="up"><button type="submit"` + upAttr + ` title="Naikkan urutan" aria-label="Naikkan urutan">` + svgIcon("arrow-up") + `</button></form>`)
		rows.WriteString(`<form method="post" action="/pipeline/stages/` + attr(stage.ID) + `/move" class="inline-form"><input type="hidden" name="direction" value="down"><button type="submit"` + downAttr + ` title="Turunkan urutan" aria-label="Turunkan urutan">` + svgIcon("arrow-down") + `</button></form>`)
		rows.WriteString(`<form method="post" action="/pipeline/stages/` + attr(stage.ID) + `/delete" class="inline-form" data-confirm="` + attr(confirmMsg) + `"><button type="submit" class="stage-delete" title="Hapus tahap" aria-label="Hapus tahap">` + svgIcon("close") + `</button></form>`)
		rows.WriteString(`</div></div>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<p class="empty-note">Belum ada tahap. Tambahkan tahap pertama di bawah.</p>`)
	}
	return `<div class="panel stage-manager"><h3 class="text-sm font-semibold mb-1">Kelola Tahap Pipeline</h3><p class="text-xs text-slate-500 mb-3">Tambah, ubah nama, urutkan, atau hapus tahap sesuai kebutuhan operasional kamu — tidak perlu ubah kode program.</p>` + rows.String() + `<form method="post" action="/pipeline/stages/create" class="stage-manager-add mt-2"><input name="name" placeholder="Nama tahap baru, contoh: Medical Check Up" required maxlength="80" aria-label="Nama tahap baru"><button class="primary-button" type="submit">+ Tambah Tahap</button></form></div>`
}

func staffReminderList(vm dashboard.ViewModel) string {
	var items []string
	if n := countOpenTasks(vm.Tasks); n > 0 {
		items = append(items, intText(n)+" tugas belum selesai")
	}
	if n := countReviewDocs(vm.Documents); n > 0 {
		items = append(items, intText(n)+" dokumen menunggu review")
	}
	if n := countWaitingExpenses(vm.Expenses); n > 0 {
		items = append(items, intText(n)+" pengeluaran menunggu dicatat")
	}
	if n := len(vm.Schedules); n > 0 {
		items = append(items, intText(n)+" jadwal aktif")
	}
	if len(items) == 0 {
		return `<ul class="reminder-list clean"><li>Tidak ada pengingat. Semua tugas beres 🎉</li></ul>`
	}
	var b strings.Builder
	b.WriteString(`<ul class="reminder-list clean">`)
	for _, item := range items {
		b.WriteString(`<li>` + esc(item) + `</li>`)
	}
	b.WriteString(`</ul>`)
	return b.String()
}

func quickSchedule(vm dashboard.ViewModel) string {
	var rows strings.Builder
	shown := 0
	for _, item := range vm.Schedules {
		if shown >= 4 {
			break
		}
		time := item.TimeLabel
		if time == "" {
			time = item.DateLabel
		}
		rows.WriteString(`<div><strong>` + esc(time) + `</strong><span>` + esc(item.Title) + `</span><em>` + esc(item.Location) + `</em></div>`)
		shown++
	}
	if rows.Len() == 0 {
		rows.WriteString(`<p class="empty-note">Belum ada jadwal. Tambahkan lewat menu Kalender.</p>`)
	}
	return `<div class="panel border-0 shadow-none"><div class="panel-head"><h2>Jadwal Terdekat</h2><a href="/staff/calendar" hx-get="/staff/calendar" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Lihat Kalender</a></div><div class="schedule-list">` + rows.String() + `</div></div>`
}

func recentDocs(vm dashboard.ViewModel) string {
	var rows strings.Builder
	shown := 0
	for _, document := range vm.Documents {
		if shown >= 4 {
			break
		}
		rows.WriteString(`<div><strong>` + esc(document.ClientName) + ` - ` + esc(document.Name) + `</strong><span>` + esc(dateLabel(document.UpdatedAt)) + `</span><em>` + esc(documentStatusLabel(document.Status)) + `</em></div>`)
		shown++
	}
	if rows.Len() == 0 {
		rows.WriteString(`<p class="empty-note">Belum ada dokumen dari client.</p>`)
	}
	return `<div class="panel border-0 shadow-none"><div class="panel-head"><h2>Dokumen Terbaru</h2><a href="/staff/documents" hx-get="/staff/documents" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Lihat Semua</a></div><div class="doc-list">` + rows.String() + `</div></div>`
}

func expensesTable(expenses []dashboard.Expense) string {
	var rows strings.Builder
	for _, expense := range expenses {
		action := `<button class="icon-action small" type="button" disabled>Belum ada bukti</button>`
		if expense.ReceiptStoragePath != "" {
			action = `<a class="icon-action" href="/staff/expenses/` + attr(expense.ID) + `/receipt">Lihat</a>`
		}
		rows.WriteString(`<tr><td>` + esc(expense.DateLabel) + `</td><td><strong>` + esc(expense.ClientName) + `</strong></td><td>` + esc(expense.Need) + `</td><td>` + esc(expense.Category) + `</td><td>` + money(expense.Amount) + `</td><td><span class="receipt-thumb"></span></td><td>` + esc(expenseStatusLabel(expense.Status)) + `</td><td>` + action + `</td></tr>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<tr><td colspan="8">Belum ada pengeluaran.</td></tr>`)
	}
	return `<table class="data-table"><thead><tr><th>Tanggal</th><th>Klien</th><th>Berkas / Keperluan</th><th>Kategori</th><th>Nominal</th><th>Bukti</th><th>Status</th><th>Aksi</th></tr></thead><tbody>` + rows.String() + `</tbody></table><div class="pagination"><span>Menampilkan ` + intText(len(expenses)) + ` data</span><div><button disabled>&lt;</button><button class="active">1</button><button disabled>&gt;</button></div></div>`
}

func firstClient(clients []dashboard.ClientProfile) dashboard.ClientProfile {
	if len(clients) > 0 {
		return clients[0]
	}
	return dashboard.ClientProfile{Name: "Client", PackageName: "-", PICName: "-", Progress: 0, CurrentStage: "-", Campus: "-"}
}

func firstOrder(orders []dashboard.Order) dashboard.Order {
	if len(orders) > 0 {
		return orders[0]
	}
	return dashboard.Order{Code: "-", ClientName: "-", PackageName: "-", Status: dashboard.OrderUnpaid}
}

// latestOrderForClient finds a client's most recent order — vm.Orders is
// already sorted newest-first by the repository, so the first match wins.
// Used to prefill the Total Tagihan/Uang Masuk fields on the client edit
// form; found is false for a client who doesn't have an order yet.
func latestOrderForClient(orders []dashboard.Order, clientID string) (order dashboard.Order, found bool) {
	for _, o := range orders {
		if o.ClientID == clientID {
			return o, true
		}
	}
	return dashboard.Order{}, false
}

func orderTotals(orders []dashboard.Order) (int64, int64) {
	var total, paid int64
	for _, order := range orders {
		total += order.Total
		paid += order.Paid
	}
	return total, paid
}

func countOpenOrders(orders []dashboard.Order) int {
	count := 0
	for _, order := range orders {
		if order.Status != dashboard.OrderPaid {
			count++
		}
	}
	return count
}

func countWaitingOrders(orders []dashboard.Order) int {
	count := 0
	for _, order := range orders {
		if order.Status == dashboard.OrderWaitingVerification {
			count++
		}
	}
	return count
}

func countPaidOrders(orders []dashboard.Order) int {
	count := 0
	for _, order := range orders {
		if order.Status == dashboard.OrderPaid {
			count++
		}
	}
	return count
}

func unpaidAmount(orders []dashboard.Order) int64 {
	var total int64
	for _, order := range orders {
		remaining := order.Total - order.Paid
		if remaining > 0 {
			total += remaining
		}
	}
	return total
}

func countReviewDocs(documents []dashboard.Document) int {
	count := 0
	for _, document := range documents {
		if document.Status == dashboard.DocumentReview {
			count++
		}
	}
	return count
}

func countApprovedDocs(documents []dashboard.Document) int {
	count := 0
	for _, document := range documents {
		if document.Status == dashboard.DocumentApproved {
			count++
		}
	}
	return count
}

func countOpenTasks(tasks []dashboard.Task) int {
	count := 0
	for _, task := range tasks {
		if task.Status != dashboard.TaskDone {
			count++
		}
	}
	return count
}

func countDoneTasks(tasks []dashboard.Task) int {
	count := 0
	for _, task := range tasks {
		if task.Status == dashboard.TaskDone {
			count++
		}
	}
	return count
}

func countWaitingExpenses(expenses []dashboard.Expense) int {
	count := 0
	for _, expense := range expenses {
		if expense.Status == dashboard.ExpenseWaiting {
			count++
		}
	}
	return count
}

func money(value int64) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	digits := strconv.FormatInt(value, 10)
	var parts []string
	for len(digits) > 3 {
		parts = append([]string{digits[len(digits)-3:]}, parts...)
		digits = digits[:len(digits)-3]
	}
	parts = append([]string{digits}, parts...)
	return sign + "Rp " + strings.Join(parts, ".")
}

func intText(value int) string {
	return strconv.Itoa(value)
}

func percentStyle(value int) string {
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	return strconv.Itoa(value) + "%"
}

func progressRingStyle(value int) string {
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	percent := strconv.Itoa(value) + "%"
	return "background:conic-gradient(#059669 0 " + percent + ",#e2e8f0 " + percent + " 100%)"
}

func dateLabel(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.Format("02 Jan 2006")
}

func timeAgoLabel(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	elapsed := time.Since(value)
	switch {
	case elapsed < time.Minute:
		return "Baru saja"
	case elapsed < time.Hour:
		return intText(int(elapsed.Minutes())) + " menit lalu"
	case elapsed < 24*time.Hour:
		return intText(int(elapsed.Hours())) + " jam lalu"
	case elapsed < 7*24*time.Hour:
		return intText(int(elapsed.Hours()/24)) + " hari lalu"
	default:
		return dateLabel(value)
	}
}

func orderPaymentPercent(order dashboard.Order) int {
	if order.Total <= 0 {
		return 0
	}
	return int(order.Paid * 100 / order.Total)
}

func orderRemaining(order dashboard.Order) int64 {
	remaining := order.Total - order.Paid
	if remaining < 0 {
		return 0
	}
	return remaining
}

func orderStatusLabel(status dashboard.OrderStatus) string {
	switch status {
	case dashboard.OrderPaid:
		return "Lunas"
	case dashboard.OrderWaitingVerification:
		return "Menunggu Verifikasi"
	default:
		return "Belum Dibayar"
	}
}

func orderStatusClass(status dashboard.OrderStatus) string {
	switch status {
	case dashboard.OrderPaid:
		return "green"
	case dashboard.OrderWaitingVerification:
		return "amber"
	default:
		return "red"
	}
}

func paymentStatusLabel(status dashboard.PaymentStatus) string {
	switch status {
	case dashboard.PaymentVerified:
		return "Terverifikasi"
	case dashboard.PaymentRejected:
		return "Ditolak"
	default:
		return "Menunggu"
	}
}

func paymentStatusClass(status dashboard.PaymentStatus) string {
	switch status {
	case dashboard.PaymentVerified:
		return "green"
	case dashboard.PaymentRejected:
		return "red"
	default:
		return "amber"
	}
}

func documentStatusLabel(status dashboard.DocumentStatus) string {
	switch status {
	case dashboard.DocumentApproved:
		return "Disetujui"
	case dashboard.DocumentReview:
		return "Review"
	case dashboard.DocumentRevision:
		return "Revisi"
	default:
		return "Belum Upload"
	}
}

func documentStatusClass(status dashboard.DocumentStatus) string {
	switch status {
	case dashboard.DocumentApproved:
		return "green"
	case dashboard.DocumentReview:
		return "blue"
	case dashboard.DocumentRevision:
		return "amber"
	default:
		return "slate"
	}
}

func taskStatusLabel(status dashboard.TaskStatus) string {
	if status == dashboard.TaskDone {
		return "Selesai"
	}
	return "Belum Selesai"
}

func taskStatusClass(status dashboard.TaskStatus) string {
	if status == dashboard.TaskDone {
		return "green"
	}
	return "red"
}

func priorityClass(priority string) string {
	switch strings.ToLower(priority) {
	case "tinggi":
		return "red"
	case "sedang":
		return "amber"
	default:
		return "green"
	}
}

func expenseStatusLabel(status dashboard.ExpenseStatus) string {
	if status == dashboard.ExpenseRecorded {
		return "Tercatat"
	}
	return "Menunggu"
}

func iconHTML(name string) string {
	return `<span class="nav-icon nav-icon-` + attr(name) + `" aria-hidden="true">` + svgIcon(name) + `</span>`
}

func toolIconHTML(name string) string {
	return `<span class="tool-icon ` + attr(name) + `" aria-hidden="true">` + svgIcon(name) + `</span>`
}

func metricIconHTML(name, tone string) string {
	return `<div class="metric-icon ` + attr(tone) + `" aria-hidden="true">` + svgIcon(name) + `</div>`
}

func metricSymbolHTML(name string) string {
	return `<div class="metric-symbol ` + attr(name) + `" aria-hidden="true">` + svgIcon(name) + `</div>`
}

func svgIcon(name string) string {
	switch name {
	case "grid":
		return svg(`<rect x="3" y="3" width="7" height="7" rx="1.5"></rect><rect x="14" y="3" width="7" height="7" rx="1.5"></rect><rect x="14" y="14" width="7" height="7" rx="1.5"></rect><rect x="3" y="14" width="7" height="7" rx="1.5"></rect>`)
	case "wallet":
		return svg(`<path d="M20 7V6a2 2 0 0 0-2-2H5a3 3 0 0 0 0 6h15v8a2 2 0 0 1-2 2H5a3 3 0 0 1-3-3V7"></path><path d="M16 13h4"></path>`)
	case "users":
		return svg(`<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M22 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path>`)
	case "kanban":
		return svg(`<rect x="3" y="4" width="5" height="16" rx="1.5"></rect><rect x="10" y="4" width="5" height="10" rx="1.5"></rect><rect x="17" y="4" width="4" height="13" rx="1.5"></rect>`)
	case "package":
		return svg(`<path d="m21 8-9-5-9 5 9 5 9-5Z"></path><path d="M3 8v8l9 5 9-5V8"></path><path d="M12 13v8"></path>`)
	case "receipt":
		return svg(`<path d="M4 2v20l3-2 3 2 3-2 3 2 4-2V2l-4 2-3-2-3 2-3-2-3 2Z"></path><path d="M8 7h8"></path><path d="M8 12h8"></path><path d="M8 17h5"></path>`)
	case "chart":
		return svg(`<path d="M3 3v18h18"></path><path d="M8 17V9"></path><path d="M13 17V5"></path><path d="M18 17v-6"></path>`)
	case "settings":
		return svg(`<circle cx="12" cy="12" r="3"></circle><path d="M12 2v3"></path><path d="M12 19v3"></path><path d="m4.93 4.93 2.12 2.12"></path><path d="m16.95 16.95 2.12 2.12"></path><path d="M2 12h3"></path><path d="M19 12h3"></path><path d="m4.93 19.07 2.12-2.12"></path><path d="m16.95 7.05 2.12-2.12"></path>`)
	case "calendar":
		return svg(`<rect x="3" y="4" width="18" height="17" rx="2"></rect><path d="M16 2v4"></path><path d="M8 2v4"></path><path d="M3 10h18"></path><path d="M8 14h.01"></path><path d="M12 14h.01"></path><path d="M16 14h.01"></path>`)
	case "chat":
		return svg(`<path d="M21 15a4 4 0 0 1-4 4H8l-5 3V7a4 4 0 0 1 4-4h10a4 4 0 0 1 4 4Z"></path><path d="M8 9h8"></path><path d="M8 13h5"></path>`)
	case "file":
		return svg(`<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8Z"></path><path d="M14 2v6h6"></path><path d="M8 13h8"></path><path d="M8 17h5"></path>`)
	case "briefcase":
		return svg(`<rect x="3" y="7" width="18" height="13" rx="2"></rect><path d="M8 7V5a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><path d="M3 12h18"></path><path d="M12 12v2"></path>`)
	case "check":
		return svg(`<circle cx="12" cy="12" r="9"></circle><path d="m8 12 3 3 5-6"></path>`)
	case "menu":
		return svg(`<path d="M4 6h16"></path><path d="M4 12h16"></path><path d="M4 18h16"></path>`)
	case "search":
		return svg(`<circle cx="11" cy="11" r="7"></circle><path d="m16 16 4 4"></path>`)
	case "bell":
		return svg(`<path d="M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9"></path><path d="M10 21h4"></path>`)
	case "mail":
		return svg(`<rect x="3" y="5" width="18" height="14" rx="2"></rect><path d="m3 7 9 6 9-6"></path>`)
	case "revenue":
		return svg(`<circle cx="12" cy="12" r="9"></circle><path d="M12 7v10"></path><path d="M16 9.5A3.5 3.5 0 0 0 12.5 7H11a2 2 0 0 0 0 4h2a2 2 0 0 1 0 4h-1.5A3.5 3.5 0 0 1 8 12.5"></path>`)
	case "profit":
		return svg(`<path d="M3 17 9 11l4 4 8-8"></path><path d="M14 7h7v7"></path>`)
	case "orders":
		return svg(`<rect x="5" y="4" width="14" height="17" rx="2"></rect><path d="M9 4a3 3 0 0 1 6 0"></path><path d="M9 10h6"></path><path d="M9 14h6"></path><path d="M9 18h4"></path>`)
	case "check-plain":
		return svg(`<path d="M20 6 9 17l-5-5"></path>`)
	case "eye":
		return svg(`<path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7Z"></path><circle cx="12" cy="12" r="3"></circle>`)
	case "close":
		return svg(`<path d="M18 6 6 18"></path><path d="m6 6 12 12"></path>`)
	case "download":
		return svg(`<path d="M12 3v12"></path><path d="m7 10 5 5 5-5"></path><path d="M5 21h14"></path>`)
	case "arrow-up":
		return svg(`<path d="M12 19V5"></path><path d="m5 12 7-7 7 7"></path>`)
	case "arrow-down":
		return svg(`<path d="M12 5v14"></path><path d="m19 12-7 7-7-7"></path>`)
	case "plus":
		return svg(`<path d="M12 5v14"></path><path d="M5 12h14"></path>`)
	case "settings-slider":
		return svg(`<line x1="4" x2="4" y1="21" y2="14"></line><line x1="4" x2="4" y1="10" y2="3"></line><line x1="12" x2="12" y1="21" y2="12"></line><line x1="12" x2="12" y1="8" y2="3"></line><line x1="20" x2="20" y1="21" y2="16"></line><line x1="20" x2="20" y1="12" y2="3"></line><line x1="2" x2="6" y1="14" y2="14"></line><line x1="10" x2="14" y1="8" y2="8"></line><line x1="18" x2="22" y1="16" y2="16"></line>`)
	default:
		return svg(`<circle cx="12" cy="12" r="9"></circle>`)
	}
}

func svg(paths string) string {
	return `<svg viewBox="0 0 24 24" aria-hidden="true" focusable="false" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">` + paths + `</svg>`
}

func initials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "NA"
	}
	if len(parts) == 1 {
		return esc(strings.ToUpper(parts[0][:1]))
	}
	return esc(strings.ToUpper(parts[0][:1] + parts[1][:1]))
}

func esc(value string) string {
	return template.HTMLEscapeString(value)
}

func attr(value string) string {
	return template.HTMLEscapeString(value)
}
