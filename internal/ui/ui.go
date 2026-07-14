package ui

import (
	"context"
	"html/template"
	"io"
	"strconv"
	"strings"
	"time"

	"university_agency/internal/dashboard"

	"github.com/a-h/templ"
)

const assetVersion = "20260712-student-payment-gate"

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
		description = "Dashboard operasional Taiwan Education Consulting untuk owner, staff, dan client."
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
	b.WriteString(`<div class="brand-block"><div class="brand-mark" aria-hidden="true"><span></span></div><div><p class="brand-title">TAIWAN</p><p class="brand-subtitle">Education Consulting</p><p class="brand-note">Study Abroad Taiwan</p></div></div>`)
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
	b.WriteString(`<p class="sidebar-footnote">&copy; 2026 Taiwan Education Consulting<br>v1.0.0</p>`)
	b.WriteString(`</aside>`)
	return b.String()
}

func topbarHTML(vm dashboard.ViewModel) string {
	var b strings.Builder
	b.WriteString(`<header class="topbar">`)
	b.WriteString(`<div class="min-w-0"><h1 class="page-title">` + esc(vm.Title) + `</h1><p class="page-subtitle">` + esc(vm.Subtitle) + `</p></div>`)
	b.WriteString(`<div class="topbar-actions">`)
	b.WriteString(`<div class="date-pill">` + toolIconHTML("calendar") + `<span>` + esc(vm.DateLabel) + `</span></div>`)
	b.WriteString(`<label class="search-box">` + toolIconHTML("search") + `<input type="search" placeholder="` + attr(vm.Search) + `" aria-label="Pencarian"></label>`)
	b.WriteString(`<button type="button" class="tool-button" aria-label="Notifikasi">` + toolIconHTML("bell") + `<span class="notif-dot">5</span></button>`)
	b.WriteString(`<button type="button" class="tool-button" aria-label="Pesan">` + toolIconHTML("mail") + `<span class="notif-dot notif-dot-red">3</span></button>`)
	b.WriteString(`<div class="user-chip"><div class="avatar-sm">` + esc(vm.UserInitials) + `</div><div class="hidden min-w-0 sm:block"><p>` + esc(vm.UserName) + `</p><span>` + esc(vm.UserRole) + `</span></div></div>`)
	b.WriteString(`<form method="post" action="/logout" class="logout-form"><button type="submit" class="outline-button small">Keluar</button></form>`)
	b.WriteString(`</div></header>`)
	return b.String()
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
		return ownerPipeline()
	case dashboard.SectionServices:
		return ownerServices()
	case dashboard.SectionInvoices:
		return ownerInvoices(vm)
	case dashboard.SectionReports:
		return ownerReports(vm)
	case dashboard.SectionChat:
		return chatPanel(vm, "Chat Operasional")
	case dashboard.SectionSettings:
		return settingsPanel("Owner", true)
	default:
		return ownerDashboard(vm)
	}
}

func staffContent(vm dashboard.ViewModel) string {
	switch vm.Section {
	case dashboard.SectionClients:
		return staffClients(vm)
	case dashboard.SectionPipeline:
		return staffPipeline()
	case dashboard.SectionTasks:
		return staffTasks(vm)
	case dashboard.SectionDocuments:
		return staffDocuments(vm)
	case dashboard.SectionExpenses:
		return staffExpenses(vm)
	case dashboard.SectionCalendar:
		return calendarPanel("Jadwal Staff")
	case dashboard.SectionChat:
		return chatPanel(vm, "Chat Klien")
	case dashboard.SectionSettings:
		return settingsPanel("Staff", false)
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
	case dashboard.SectionPayments:
		return studentPayments(vm)
	case dashboard.SectionCalendar:
		return studentCalendar(vm)
	case dashboard.SectionChat:
		if !vm.StudentFeatureAccess {
			return lockedStudentFeature("Chat dibuka setelah pembayaran", "Kirim bukti pembayaran atau tunggu invoice ditandai lunas agar chat konsultan aktif.")
		}
		return chatPanel(vm, "Chat Konsultan")
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

    <div class="panel">
      <div class="panel-head"><h2>Pipeline Keseluruhan</h2><a href="/owner/pipeline" hx-get="/owner/pipeline" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Lihat Detail</a></div>
      <div class="pipeline-bars">
        <div><span>Konsultasi</span><div><i class="w-[24%] bg-violet-500"></i></div><b>32</b><em>24%</em></div>
        <div><span>Persiapan Dokumen</span><div><i class="w-[21%] bg-blue-500"></i></div><b>28</b><em>21%</em></div>
        <div><span>Apply / Proses</span><div><i class="w-[19%] bg-sky-500"></i></div><b>26</b><em>19%</em></div>
        <div><span>LOA</span><div><i class="w-[12%] bg-teal-500"></i></div><b>16</b><em>12%</em></div>
        <div><span>Visa</span><div><i class="w-[15%] bg-emerald-500"></i></div><b>20</b><em>15%</em></div>
        <div><span>Keberangkatan</span><div><i class="w-[7%] bg-lime-500"></i></div><b>9</b><em>7%</em></div>
        <div><span>Selesai</span><div><i class="w-full bg-amber-500"></i></div><b>58</b><em>100%</em></div>
      </div>
    </div>

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
    <div class="panel">
      <div class="panel-head"><h2>Recent Activity</h2></div>
      <div class="activity-grid">
        <div><span class="avatar-mini bg-blue-100 text-blue-700">RP</span><p><strong>Ricky Pratama</strong>Order #LS-2026-0123 selesai</p><em>2 jam lalu</em></div>
        <div><span class="avatar-mini bg-emerald-100 text-emerald-700">SA</span><p><strong>Siti Aisyah</strong>DP diterima Rp 7.500.000</p><em>3 jam lalu</em></div>
        <div><span class="avatar-mini bg-violet-100 text-violet-700">DL</span><p><strong>Dewi Lestari</strong>LOA diterima dari NTU</p><em>5 jam lalu</em></div>
        <div><span class="avatar-mini bg-amber-100 text-amber-700">AW</span><p><strong>Andi Wijaya</strong>Pembayaran tahap 2</p><em>1 hari lalu</em></div>
      </div>
    </div>
    <div class="panel">
      <div class="panel-head"><h2>Reminder</h2><a href="/owner/clients" hx-get="/owner/clients" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Lihat Semua</a></div>
      <ul class="reminder-list">
        <li><span class="alert-dot amber"></span>3 invoice jatuh tempo hari ini</li>
        <li><span class="alert-dot red"></span>5 pembayaran belum diverifikasi</li>
        <li><span class="alert-dot red"></span>2 dokumen mahasiswa kedaluwarsa</li>
      </ul>
    </div>
  </section>
</div>`
}

func ownerFinance(vm dashboard.ViewModel) string {
	return `
<div class="space-y-5">
  <div class="toolbar-row"><div class="tabs"><button class="tab-active">Ringkasan</button><button>Pemasukan</button><button>Pengeluaran</button><button>Piutang</button><button>Arus Kas</button></div><div class="toolbar-actions"><button class="outline-button">Juni 2026</button><button class="primary-button">Export</button></div></div>
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

func ownerPipeline() string {
	return `
<div class="space-y-5">
  <div class="toolbar-row"><div class="filter-group"><button class="outline-button">Semua Paket</button><button class="outline-button">Semua Staff</button></div><button class="primary-button">Export</button></div>
  <section class="kanban-board">
    ` + kanbanColumn("Konsultasi", "5", "mint", []string{"Budi Santoso|Paket Profesional|Rina", "Siti Aisyah|Paket Basic|Dika", "Fajar Ramadhan|Layanan Satuan|Kevin"}) + `
    ` + kanbanColumn("Persiapan Dokumen", "28", "rose", []string{"Dewi Lestari|Paket Profesional|Rina", "Andi Wijaya|Paket Basic|Dika", "Clara Monica|Paket Basic|Kevin"}) + `
    ` + kanbanColumn("Apply / Proses", "26", "emerald", []string{"Ricky Pratama|Paket Profesional|Rina", "Kevin Santoso|Layanan Satuan|Sari", "Nabila Putri|Layanan Satuan|Kevin"}) + `
    ` + kanbanColumn("LOA", "16", "sky", []string{"Jason Tanu|Paket Basic|Rina", "Michelle Chen|Layanan Satuan|Dika", "William K.|Paket Basic|Kevin"}) + `
    ` + kanbanColumn("Visa", "20", "violet", []string{"Putri Amanda|Paket Profesional|Rina", "Natasha Lee|Paket Basic|Dika", "Budi Setyawan|Layanan Satuan|Sari"}) + `
    ` + kanbanColumn("Keberangkatan", "9", "lime", []string{"Dimas Aditya|Paket Profesional|Kevin", "Yuan Zo|Paket Basic|Dika", "Siti Nurhaliza|Layanan Satuan|Rina"}) + `
    ` + kanbanColumn("Selesai", "58", "amber", []string{"Budi Santoso|Paket Profesional|Rina", "Dewi Lestari|Paket Profesional|Rina", "Andi Wijaya|Paket Basic|Dika"}) + `
  </section>
</div>`
}

func clientsPanel(vm dashboard.ViewModel, includeStaffFilter bool) string {
	filter := `<button class="outline-button">Semua Paket</button><button class="outline-button">Semua Status</button>`
	if includeStaffFilter {
		filter += `<button class="outline-button">Semua Staff</button>`
	}
	var rows strings.Builder
	for i, client := range vm.Clients {
		rows.WriteString(`<tr><td>` + intText(i+1) + `</td><td><strong>` + esc(client.Name) + `</strong><span>` + esc(client.Email) + `</span></td><td>` + esc(client.PackageName) + `</td><td>` + esc(client.PICName) + `</td><td>` + esc(client.Status) + `</td><td><span class="progress-line"><i style="width:` + percentStyle(client.Progress) + `"></i></span>` + intText(client.Progress) + `%</td><td>` + esc(client.LastSchedule) + `</td><td><a class="icon-action" href="/` + attr(vm.Role.String()) + `/chat" hx-get="/` + attr(vm.Role.String()) + `/chat" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Chat</a></td></tr>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<tr><td colspan="8">Belum ada client yang bisa ditampilkan.</td></tr>`)
	}
	return `
<div class="panel table-panel">
  <div class="toolbar-row mb-4"><label class="search-box wide">` + toolIconHTML("search") + `<input type="search" placeholder="Cari nama, email, atau WA" aria-label="Cari client"></label><div class="filter-group">` + filter + `<button class="primary-button">Export</button></div></div>
  <table class="data-table"><thead><tr><th>No.</th><th>Nama Klien</th><th>Paket / Layanan</th><th>PIC</th><th>Status Terakhir</th><th>Progress</th><th>Jadwal Terakhir</th><th>Aksi</th></tr></thead><tbody>
  ` + rows.String() + `</tbody></table>
  <div class="pagination"><span>Menampilkan ` + intText(len(vm.Clients)) + ` data</span><div><button disabled>&lt;</button><button class="active">1</button><button disabled>&gt;</button></div></div>
</div>`
}

func ownerServices() string {
	return `
<div class="grid gap-5 xl:grid-cols-3">
  <article class="panel service-card"><span class="service-badge bg-violet-100 text-violet-700">Paket</span><h2>Paket Profesional</h2><p>Full cover untuk universitas, dokumen, visa, dan keberangkatan.</p><strong>Rp 25.000.000</strong><ul><li>28 client dalam proses</li><li>Margin sehat 38%</li><li>Checklist lengkap</li></ul></article>
  <article class="panel service-card"><span class="service-badge bg-blue-100 text-blue-700">Paket</span><h2>Paket Basic</h2><p>Pendampingan inti untuk dokumen dan aplikasi kampus.</p><strong>Rp 15.000.000</strong><ul><li>31 client dalam proses</li><li>Margin sehat 32%</li><li>Dokumen prioritas</li></ul></article>
  <article class="panel service-card"><span class="service-badge bg-amber-100 text-amber-700">Satuan</span><h2>Layanan Satuan</h2><p>Visa, legalisir, translate, TOEFL, medical, dan apply terpisah.</p><strong>Mulai Rp 450.000</strong><ul><li>72 order dalam proses</li><li>Perlu tracking bukti</li><li>Approval fleksibel</li></ul></article>
</div>`
}

func ownerInvoices(vm dashboard.ViewModel) string {
	var side strings.Builder
	side.WriteString(`<button class="active">Belum Dibayar <span>` + intText(countOpenOrders(vm.Orders)) + `</span></button>`)
	side.WriteString(`<button>Menunggu Verifikasi <span>` + intText(countWaitingOrders(vm.Orders)) + `</span></button>`)
	side.WriteString(`<button>Lunas <span>` + intText(countPaidOrders(vm.Orders)) + `</span></button>`)
	active := firstOrder(vm.Orders)
	proofDetail := `<div><span>Belum ada bukti pembayaran</span><strong>-</strong></div>`
	if active.ProofFileName != "" || active.ProofNote != "" {
		fileName := active.ProofFileName
		if fileName == "" {
			fileName = "-"
		}
		note := active.ProofNote
		if note == "" {
			note = "-"
		}
		proofDetail = `<div><span>Bukti Pembayaran</span><strong>` + esc(fileName) + `</strong></div><div><span>Catatan Transfer</span><strong>` + esc(note) + `</strong></div>`
	}
	statusClass := orderStatusClass(active.Status)
	return `
<div class="grid gap-5 xl:grid-cols-[320px_minmax(0,1fr)]">
  <aside class="panel"><h2 class="mb-4 text-base font-semibold">Semua Invoice</h2><div class="side-filter">` + side.String() + `</div><form method="post" action="/owner/orders/mark-paid" class="code-form mt-5"><label>Kode Pesanan</label><input name="order_code" value="` + attr(active.Code) + `" placeholder="ORD-2026-0102" required><button class="primary-button" type="submit">Tandai Lunas</button></form></aside>
  <section class="panel">
    <div class="invoice-head"><div><p>Invoice <strong>#` + esc(active.Code) + `</strong></p><span>` + esc(orderStatusLabel(active.Status)) + `</span></div><div><p>Jatuh Tempo</p><strong>` + esc(dateLabel(active.DueDate)) + `</strong></div></div>
    <div class="invoice-layout">
      <div><dl class="detail-grid"><div><dt>Nama Klien</dt><dd>` + esc(active.ClientName) + `</dd></div><div><dt>Paket / Layanan</dt><dd>` + esc(active.PackageName) + `</dd></div><div><dt>Kode Pesanan</dt><dd>` + esc(active.Code) + `</dd></div><div><dt>Total Tagihan</dt><dd>` + money(active.Total) + `</dd></div><div><dt>Status Pembayaran</dt><dd><span class="status ` + statusClass + `">` + esc(orderStatusLabel(active.Status)) + `</span></dd></div></dl><div class="invoice-box"><h3>Rincian Tagihan</h3><div><span>` + esc(active.PackageName) + `</span><strong>` + money(active.Total) + `</strong></div><div><span>Sudah Dibayar</span><strong>` + money(active.Paid) + `</strong></div><div class="total"><span>Sisa</span><strong>` + money(active.Total-active.Paid) + `</strong></div></div><div class="invoice-box"><h3>Bukti Pembayaran</h3>` + proofDetail + `</div></div>
      <div class="action-stack"><button class="success-button" type="button">Kirim Invoice (WA)</button><button class="primary-button" type="button">Kirim Invoice (Email)</button><button class="outline-button" type="button">Download PDF</button><form method="post" action="/owner/orders/mark-paid"><input type="hidden" name="order_code" value="` + attr(active.Code) + `"><button class="outline-button" type="submit">Tandai Lunas</button></form></div>
    </div>
  </section>
</div>`
}

func ownerReports(vm dashboard.ViewModel) string {
	return `
<div class="space-y-5">
  <div class="toolbar-row"><div class="tabs"><button class="tab-active">Omzet</button><button>Profit</button><button>Pemasukan</button><button>Pengeluaran</button><button>Piutang</button></div><div class="toolbar-actions"><button class="outline-button">Tahun Ini</button><button class="primary-button">Export</button></div></div>
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
    <div class="panel border-0 shadow-none"><div class="panel-head"><h2>Pipeline Saya</h2><a href="/staff/pipeline" hx-get="/staff/pipeline" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Lihat Semua</a></div><div class="mini-kanban">` + miniPipeline() + `</div></div>
    <div class="right-rail"><div class="panel border-0 shadow-none"><div class="panel-head"><h2>Pengingat</h2><a href="#">Lihat Semua</a></div><ul class="reminder-list clean"><li>3 appointment kedutaan minggu ini</li><li>5 dokumen menunggu review</li><li>2 pengeluaran menunggu dicatat</li><li>4 dokumen akan kedaluwarsa</li><li>1 jadwal interview besok</li></ul></div>` + quickSchedule() + recentDocs() + `</div>
  </section>
  <section class="panel border-0 shadow-none">
    <div class="toolbar-row mb-4"><div><h2 class="text-lg font-semibold">Pengeluaran Operasional</h2><p class="text-sm text-slate-500">Catat biaya pengeluaran untuk setiap berkas atau keperluan klien.</p></div><button class="dark-button">+ Catat Pengeluaran</button></div>
    ` + expensesTable(vm.Expenses) + `
  </section>
</div>`
}

func staffClients(vm dashboard.ViewModel) string {
	return `<div class="space-y-5"><section class="metric-grid metric-grid-four"><article class="metric-card"><p>Klien Aktif</p><strong>` + intText(len(vm.Clients)) + `</strong><span class="trend-up">Dalam akses staff</span></article><article class="metric-card"><p>Butuh Follow Up</p><strong>` + intText(countOpenTasks(vm.Tasks)) + `</strong><span class="text-amber-600">Prioritas</span></article><article class="metric-card"><p>Dokumen Review</p><strong>` + intText(countReviewDocs(vm.Documents)) + `</strong><span class="text-red-600">Harus dicek</span></article><article class="metric-card"><p>Selesai Bulan Ini</p><strong>` + intText(countPaidOrders(vm.Orders)) + `</strong><span class="trend-up">Order lunas</span></article></section>` + clientsPanel(vm, false) + `</div>`
}

func staffPipeline() string {
	return `<div class="panel border-0 shadow-none"><div class="mini-kanban large">` + miniPipeline() + `</div></div>`
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
	return `<div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(0,.45fr)]"><div class="panel table-panel"><div class="panel-head"><h2>Task & To Do List</h2><button class="primary-button" type="button">Tambah Task</button></div><table class="data-table"><thead><tr><th>Waktu</th><th>Tugas</th><th>Klien</th><th>Prioritas</th><th>Status</th><th>Aksi</th></tr></thead><tbody>` + rows.String() + `</tbody></table></div><div class="right-rail">` + quickSchedule() + recentDocs() + `</div></div>`
}

func staffDocuments(vm dashboard.ViewModel) string {
	return documentsPanel(vm.Documents, true)
}

func staffExpenses(vm dashboard.ViewModel) string {
	return `<div class="panel"><div class="toolbar-row mb-4"><div><h2 class="text-lg font-semibold">Pengeluaran Operasional</h2><p class="text-sm text-slate-500">Input kategori, nominal, bukti, dan keterangan untuk owner approval.</p></div><button class="dark-button" type="button">+ Catat Pengeluaran</button></div>` + expensesTable(vm.Expenses) + `</div>`
}

func studentDashboard(vm dashboard.ViewModel) string {
	client := firstClient(vm.Clients)
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
	docAction := lockedStudentAction(vm, "Upload", "/student/documents")
	return `
<div class="space-y-5">
  <section class="grid gap-5 xl:grid-cols-[minmax(0,1.1fr)_minmax(0,.9fr)_minmax(0,.8fr)]">
    <div class="panel student-hero"><span class="service-badge bg-emerald-100 text-emerald-700">` + esc(client.PackageName) + `</span><h2>` + esc(client.Name) + `</h2><p>` + esc(client.Campus) + `</p><div class="student-progress"><strong>` + intText(client.Progress) + `%</strong><span>` + esc(client.CurrentStage) + `</span></div></div>
    <div class="panel"><div class="panel-head"><h2>Progress Keseluruhan</h2><a href="/student/progress" hx-get="/student/progress" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Detail</a></div><div class="progress-ring" style="` + progressRingStyle(client.Progress) + `"><span>` + intText(client.Progress) + `%</span></div><p class="mt-4 text-sm text-slate-500">Tahap saat ini: ` + esc(client.CurrentStage) + `. ` + intText(countDoneStages(vm.ProgressStages)) + ` dari ` + intText(len(vm.ProgressStages)) + ` tahap selesai.</p></div>
    <div class="panel"><div class="panel-head"><h2>Konsultan</h2></div><div class="consultant-card"><span class="avatar-mini bg-emerald-100 text-emerald-700">` + initials(client.PICName) + `</span><div><strong>` + esc(client.PICName) + `</strong><p>Staff Konsultan</p></div></div><div class="action-grid">` + chatAction + `<button class="outline-button" type="button">Email</button></div></div>
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
	active := activeStage(vm.ProgressStages)
	var steps strings.Builder
	for _, stage := range vm.ProgressStages {
		steps.WriteString(`<span class="` + attr(string(stage.Status)) + `">` + intText(stage.Step) + `</span>`)
	}
	var rows strings.Builder
	for _, stage := range vm.ProgressStages {
		rows.WriteString(`<tr><td>` + intText(stage.Step) + `. ` + esc(stage.Title) + `</td><td>` + esc(stage.Description) + `</td><td><span class="status ` + progressStatusClass(stage.Status) + `">` + esc(progressStatusLabel(stage.Status)) + `</span></td><td>` + esc(stage.DueLabel) + `</td><td>` + esc(stage.PICName) + `</td></tr>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<tr><td colspan="5">Belum ada data progress.</td></tr>`)
	}
	completed := countDoneStages(vm.ProgressStages)
	totalStages := len(vm.ProgressStages)
	return `
<div class="panel">
  <div class="stepper">` + steps.String() + `</div>
  <div class="panel-head mt-6"><div><h2>Detail Tahap: ` + esc(active.Title) + `</h2><p>Progress keseluruhan ` + intText(client.Progress) + `% - ` + intText(completed) + ` dari ` + intText(totalStages) + ` tahap selesai</p></div>` + lockedStudentAction(vm, "Upload Dokumen", "/student/documents") + `</div>
  <div class="student-onboarding compact"><strong>` + esc(active.Title) + ` sedang berjalan</strong><span>` + esc(active.Description) + `</span></div>
  <div class="stage-progress"><i style="width:` + percentStyle(client.Progress) + `"></i></div>
  <table class="data-table mt-5"><thead><tr><th>Tahap</th><th>Deskripsi</th><th>Status</th><th>Deadline</th><th>PIC</th></tr></thead><tbody>` + rows.String() + `</tbody></table>
</div>`
}

func studentDocuments(vm dashboard.ViewModel) string {
	if !vm.StudentFeatureAccess {
		return lockedStudentFeature("Dokumen dibuka setelah pembayaran", "Kirim bukti pembayaran atau tunggu invoice ditandai lunas sebelum mengunggah dokumen.")
	}
	return `<div class="grid gap-5 xl:grid-cols-[minmax(0,.8fr)_minmax(0,1.2fr)]"><section class="panel"><div class="panel-head"><div><h2>Upload Dokumen</h2><p>Unggah file sesuai permintaan konsultan.</p></div></div><form method="post" action="/student/documents/upload" enctype="multipart/form-data" class="student-upload-form"><label>Jenis dokumen<select name="document_name" required><option>Passport</option><option>Ijazah Terakhir</option><option>Transkrip Nilai</option><option>Foto Background Putih 4x6</option><option>KTP</option><option>KK</option><option>Akta Kelahiran</option><option>Autobiografi</option><option>Mutasi Rekening</option><option>Surat Rekomendasi</option></select></label><label>File dokumen<input name="document_file" type="file" required></label><p>Format PDF, JPG, atau PNG. File yang masuk akan tampil sebagai status review.</p><button class="primary-button" type="submit">Upload Dokumen</button></form></section><section class="panel table-panel"><div class="panel-head"><h2>Dokumen Saya</h2></div>` + documentsTable(vm.Documents, false) + `</section></div>`
}

func studentPayments(vm dashboard.ViewModel) string {
	if len(vm.Orders) == 0 {
		return `<div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(0,.75fr)]"><section class="panel empty-state"><span>Invoice</span><h2>Belum ada invoice aktif</h2><p>Pembayaran akan muncul setelah owner atau staff mengonfirmasi paket dan membuat kode pesanan kamu.</p></section><aside class="panel"><div class="panel-head"><h2>Status Pembayaran</h2></div><div class="summary-list compact"><div><span>Total Tagihan</span><strong>Rp 0</strong></div><div><span>Sudah Dibayar</span><strong>Rp 0</strong></div><div><span>Sisa Tagihan</span><strong>Rp 0</strong></div></div></aside></div>`
	}
	var invoiceList strings.Builder
	for i, order := range vm.Orders {
		className := ""
		if i == 0 {
			className = ` class="active"`
		}
		invoiceList.WriteString(`<button` + className + ` type="button">` + esc(order.Code) + ` <span>` + esc(orderStatusLabel(order.Status)) + `</span><strong>` + money(order.Total) + `</strong></button>`)
	}
	active := firstOrder(vm.Orders)
	remaining := active.Total - active.Paid
	if remaining < 0 {
		remaining = 0
	}
	paymentAction := `<form method="post" action="/student/payments/proof" enctype="multipart/form-data" class="payment-proof-form mt-5"><input type="hidden" name="order_code" value="` + attr(active.Code) + `"><label>Catatan transfer manual</label><textarea name="note" rows="3" placeholder="Contoh: Transfer bank atas nama kamu, tanggal transfer"></textarea><label>Bukti transfer<input name="proof_file" type="file"></label><div class="action-grid mt-5"><a class="outline-button" href="/student/payments/` + attr(active.Code) + `/invoice.pdf">Download Invoice PDF</a><button class="success-button" type="submit">Kirim Bukti Bayar</button></div></form>`
	if remaining == 0 || active.Status == dashboard.OrderPaid {
		paymentAction = `<div class="student-onboarding compact"><strong>Pembayaran sudah lunas</strong><span>Invoice ini sudah tercatat lunas di sistem.</span><a href="/student/payments/` + attr(active.Code) + `/invoice.pdf">Download Invoice PDF</a></div>`
	} else if active.Status == dashboard.OrderWaitingVerification {
		paymentAction = `<div class="student-onboarding compact"><strong>Bukti bayar sedang diverifikasi</strong><span>Tim akan memperbarui status setelah pembayaran cocok dengan kode pesanan.</span><a href="/student/payments/` + attr(active.Code) + `/invoice.pdf">Download Invoice PDF</a></div>` + paymentAction
	}
	return `<div class="grid gap-5 xl:grid-cols-[360px_minmax(0,1fr)]"><aside class="panel"><h2 class="mb-4 text-base font-semibold">Daftar Invoice</h2><div class="invoice-list">` + invoiceList.String() + `</div></aside><section class="panel"><div class="invoice-head"><div><p>Detail Invoice</p><strong>` + esc(active.Code) + `</strong></div><strong class="text-2xl">` + money(active.Total) + `</strong></div><dl class="detail-grid"><div><dt>Client</dt><dd>` + esc(active.ClientName) + `</dd></div><div><dt>Paket</dt><dd>` + esc(active.PackageName) + `</dd></div><div><dt>Status</dt><dd><span class="status ` + orderStatusClass(active.Status) + `">` + esc(orderStatusLabel(active.Status)) + `</span></dd></div><div><dt>Jatuh Tempo</dt><dd>` + esc(dateLabel(active.DueDate)) + `</dd></div></dl><div class="invoice-box"><h3>Rincian</h3><div><span>Subtotal</span><strong>` + money(active.Total) + `</strong></div><div><span>Sudah Dibayar</span><strong>` + money(active.Paid) + `</strong></div><div class="total"><span>Sisa</span><strong>` + money(remaining) + `</strong></div></div>` + paymentAction + `</section></div>`
}

func settingsPanel(role string, includeFinance bool) string {
	finance := ""
	if includeFinance {
		finance = `<div><span>Keuangan</span><strong>Owner only</strong><em>Aktif</em></div>`
	}
	return `<div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(0,.8fr)]"><div class="panel"><h2 class="mb-4 text-lg font-semibold">Preferensi ` + esc(role) + `</h2><div class="settings-list"><div><span>Autentikasi</span><strong>Password + session aman</strong><em>Aktif</em></div><div><span>Notifikasi</span><strong>Email dan WhatsApp reminder</strong><em>Aktif</em></div><div><span>PWA</span><strong>Installable di Android</strong><em>Aktif</em></div>` + finance + `<div><span>SEO</span><strong>Metadata per halaman</strong><em>Aktif</em></div></div></div><div class="panel"><h2 class="mb-4 text-lg font-semibold">Keamanan</h2><p class="text-sm leading-6 text-slate-600">Aplikasi menggunakan security headers, konfigurasi via environment, timeout server, dan graceful shutdown. Hak akses disiapkan per role agar staff tidak melihat laporan profit dan client hanya melihat data miliknya sendiri.</p></div></div>`
}

func calendarPanel(title string) string {
	return `<div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(0,.75fr)]"><div class="panel"><div class="panel-head"><h2>` + esc(title) + `</h2><button class="primary-button">Tambah Jadwal</button></div><div class="calendar-grid"><div><strong>10:00</strong><span>Appointment Kedutaan</span><em>Dewi Lestari</em></div><div><strong>13:00</strong><span>Interview Visa</span><em>Jason Tanu</em></div><div><strong>15:00</strong><span>Medical Check Up</span><em>Putri Amanda</em></div><div><strong>26 Jun</strong><span>Deadline dokumen visa</span><em>Ricky Pratama</em></div></div></div>` + quickSchedule() + `</div>`
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
		return `<a class="success-button" href="` + attr(href) + `" hx-get="` + attr(href) + `" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">` + esc(label) + `</a>`
	}
	return `<a class="outline-button locked-action" href="/student/payments" hx-get="/student/payments" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">` + esc(label) + `</a>`
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
	return `<div class="panel table-panel"><div class="panel-head"><h2>Dokumen Menunggu Review</h2><button class="primary-button" type="button">Review Massal</button></div>` + documentsTable(documents, actionable) + `</div>`
}

func documentsTable(documents []dashboard.Document, actionable bool) string {
	var rows strings.Builder
	for _, document := range documents {
		action := `<button class="outline-button small" type="button" disabled>Belum ada file</button>`
		if actionable && document.Status == dashboard.DocumentReview {
			action = `<form method="post" action="/staff/documents/` + attr(document.ID) + `/approve" class="inline-form"><button class="success-button small" type="submit">Approve</button></form><form method="post" action="/staff/documents/` + attr(document.ID) + `/reject" class="inline-form"><button class="danger-button small" type="submit">Reject</button></form>`
		} else if document.StoragePath != "" {
			action = `<a class="outline-button small" href="/student/documents/` + attr(document.ID) + `/download">Download</a>`
		}
		fileLabel := document.FileName
		if fileLabel == "" {
			fileLabel = "-"
		}
		rows.WriteString(`<tr><td>` + esc(document.ClientName) + `</td><td>` + esc(document.Name) + `<span>` + esc(fileLabel) + `</span></td><td>` + esc(dateLabel(document.UpdatedAt)) + `</td><td><span class="status ` + documentStatusClass(document.Status) + `">` + esc(documentStatusLabel(document.Status)) + `</span></td><td>` + esc(document.Reviewer) + `</td><td>` + action + `</td></tr>`)
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

func activeStage(stages []dashboard.ProgressStage) dashboard.ProgressStage {
	for _, stage := range stages {
		if stage.Status == dashboard.ProgressStageActive {
			return stage
		}
	}
	if len(stages) > 0 {
		return stages[0]
	}
	return dashboard.ProgressStage{Title: "Belum ada tahap", Status: dashboard.ProgressStagePending}
}

func countDoneStages(stages []dashboard.ProgressStage) int {
	count := 0
	for _, stage := range stages {
		if stage.Status == dashboard.ProgressStageDone {
			count++
		}
	}
	return count
}

func progressStatusLabel(status dashboard.ProgressStageStatus) string {
	switch status {
	case dashboard.ProgressStageDone:
		return "Selesai"
	case dashboard.ProgressStageActive:
		return "Proses"
	default:
		return "Belum Mulai"
	}
}

func progressStatusClass(status dashboard.ProgressStageStatus) string {
	switch status {
	case dashboard.ProgressStageDone:
		return "green"
	case dashboard.ProgressStageActive:
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

func miniPipeline() string {
	return kanbanColumn("Konsultasi", "5", "mint", []string{"Budi Santoso|Paket Profesional|Rina", "Siti Aisyah|Paket Basic|Rina", "Fajar Ramadhan|Layanan Satuan|Rina", "Clara Monica|Paket Basic|Rina"}) +
		kanbanColumn("Persiapan Dokumen", "7", "rose", []string{"Dewi Lestari|Paket Profesional|Rina", "Andi Wijaya|Paket Basic|Rina", "Nabila Putri|Layanan Satuan|Rina", "Jason Tanu|Layanan Satuan|Rina"}) +
		kanbanColumn("Apply / Proses", "6", "sky", []string{"Ricky Pratama|Paket Profesional|Rina", "Kevin Santoso|Paket Basic|Rina", "Putri Amanda|Layanan Satuan|Rina", "Michelle Chen|Layanan Satuan|Rina"}) +
		kanbanColumn("Visa & Imigrasi", "4", "violet", []string{"Natasha Lee|Paket Profesional|Rina", "William K.|Paket Basic|Rina", "Budi Setyawan|Layanan Satuan|Rina", "Clara Monica|Layanan Satuan|Rina"}) +
		kanbanColumn("Keberangkatan", "2", "mint", []string{"Dimas Aditya|Paket Profesional|Rina", "Yuan Zhe|Paket Basic|Rina"})
}

func kanbanColumn(title, count, tone string, cards []string) string {
	var b strings.Builder
	b.WriteString(`<div class="kanban-column tone-` + attr(tone) + `"><div class="kanban-title"><h3>` + esc(title) + `</h3><span>` + esc(count) + ` Klien</span></div><div class="kanban-cards">`)
	for _, card := range cards {
		parts := strings.Split(card, "|")
		name, packageName, pic := parts[0], "", ""
		if len(parts) > 1 {
			packageName = parts[1]
		}
		if len(parts) > 2 {
			pic = parts[2]
		}
		b.WriteString(`<article><span class="avatar-mini bg-slate-100 text-slate-700">` + initials(name) + `</span><div><strong>` + esc(name) + `</strong><p>` + esc(packageName) + `</p><em>` + esc(pic) + `</em></div></article>`)
	}
	b.WriteString(`<button class="more-card">+ ` + esc(count) + ` lainnya</button></div></div>`)
	return b.String()
}

func quickSchedule() string {
	return `<div class="panel border-0 shadow-none"><div class="panel-head"><h2>Jadwal Hari Ini</h2><a href="#">Lihat Kalender</a></div><div class="schedule-list"><div><strong>10:00</strong><span>Appointment Kedutaan</span><em>Dewi Lestari</em></div><div><strong>13:00</strong><span>Interview Visa</span><em>Jason Tanu</em></div><div><strong>15:00</strong><span>Medical Check Up</span><em>Putri Amanda</em></div></div></div>`
}

func recentDocs() string {
	return `<div class="panel border-0 shadow-none"><div class="panel-head"><h2>Dokumen Terbaru</h2><a href="/staff/documents" hx-get="/staff/documents" hx-target="#app" hx-swap="outerHTML" hx-push-url="true">Lihat Semua</a></div><div class="doc-list"><div><strong>Budi Santoso - Ijazah</strong><span>Diupload 2 jam lalu</span><em>Selesai</em></div><div><strong>Siti Aisyah - Transkrip Nilai</strong><span>Diupload 3 jam lalu</span><em>Review</em></div><div><strong>Ricky Pratama - Passport</strong><span>Diupload 5 jam lalu</span><em>Review</em></div><div><strong>Dewi Lestari - Form Visa</strong><span>Diupload 6 jam lalu</span><em>Selesai</em></div></div></div>`
}

func expensesTable(expenses []dashboard.Expense) string {
	var rows strings.Builder
	for _, expense := range expenses {
		rows.WriteString(`<tr><td>` + esc(expense.DateLabel) + `</td><td><strong>` + esc(expense.ClientName) + `</strong></td><td>` + esc(expense.Need) + `</td><td>` + esc(expense.Category) + `</td><td>` + money(expense.Amount) + `</td><td><span class="receipt-thumb"></span></td><td>` + esc(expenseStatusLabel(expense.Status)) + `</td><td><button class="icon-action" type="button">Lihat</button></td></tr>`)
	}
	if rows.Len() == 0 {
		rows.WriteString(`<tr><td colspan="8">Belum ada pengeluaran.</td></tr>`)
	}
	return `<div class="toolbar-row mb-4"><div class="filter-group"><button class="outline-button">Semua Klien</button><button class="outline-button">Semua Kategori</button><button class="outline-button">Semua Status</button><button class="outline-button">01/06/2026 - 25/06/2026</button></div></div><table class="data-table"><thead><tr><th>Tanggal</th><th>Klien</th><th>Berkas / Keperluan</th><th>Kategori</th><th>Nominal</th><th>Bukti</th><th>Status</th><th>Aksi</th></tr></thead><tbody>` + rows.String() + `</tbody></table><div class="pagination"><span>Menampilkan ` + intText(len(expenses)) + ` data</span><div><button disabled>&lt;</button><button class="active">1</button><button disabled>&gt;</button></div></div>`
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
