package dashboard

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Repository interface {
	CompanySnapshot(ctx context.Context, viewer User, viewRole Role) (CompanySnapshot, error)
	ListClients(ctx context.Context, viewer User, viewRole Role) ([]ClientProfile, error)
	ListOrders(ctx context.Context, viewer User, viewRole Role) ([]Order, error)
	ListDocuments(ctx context.Context, viewer User, viewRole Role) ([]Document, error)
	ListProgressStages(ctx context.Context, viewer User, viewRole Role) ([]ProgressStage, error)
	ListSchedules(ctx context.Context, viewer User, viewRole Role) ([]ScheduleItem, error)
	ListTasks(ctx context.Context, viewer User, viewRole Role) ([]Task, error)
	ListExpenses(ctx context.Context, viewer User, viewRole Role) ([]Expense, error)
	ListConversations(ctx context.Context, viewer User) ([]ChatConversation, error)
	ListMessages(ctx context.Context, viewer User, conversationID string) ([]ChatMessage, error)
	ListPipelineStages(ctx context.Context) ([]PipelineStage, error)
	ListServicePackages(ctx context.Context) ([]ServicePackage, error)
}

type CompanySnapshot struct {
	ActiveClients   int
	CompletedOrders int
	OpenOrders      int
	UnpaidInvoices  int
	Revenue         int64
	Expenses        int64
	Profit          int64
}

type ViewOptions struct {
	ConversationID   string
	Flash            string
	OrderCode        string
	InvoiceFilter    string
	ClientStatus     string
	ClientPackage    string
	ClientPIC        string
	ClientSearch     string
	ShowCreateForm   bool
	ShowStageManager bool
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return Service{repository: repository}
}

func (s Service) View(ctx context.Context, appName, appURL string, viewer User, role Role, section Section, options ViewOptions) (ViewModel, error) {
	if !validSection(role, section) {
		return ViewModel{}, ErrNotFound
	}
	if !CanViewRole(viewer.Role, role) {
		return ViewModel{}, ErrForbidden
	}

	vm := ViewModel{
		AppName:          appName,
		AppURL:           appURL,
		Viewer:           viewer,
		Role:             role,
		Section:          section,
		DateLabel:        dateLabelID(time.Now()),
		NotificationText: "5 notifikasi baru",
		Navigation:       navigation(role, section),
		Flash:            options.Flash,
	}

	if s.repository != nil {
		snapshot, err := s.repository.CompanySnapshot(ctx, viewer, role)
		if err != nil {
			return ViewModel{}, fmt.Errorf("load company snapshot: %w", err)
		}
		vm.Snapshot = snapshot

		if vm.Clients, err = s.repository.ListClients(ctx, viewer, role); err != nil {
			return ViewModel{}, fmt.Errorf("load clients: %w", err)
		}
		if vm.Orders, err = s.repository.ListOrders(ctx, viewer, role); err != nil {
			return ViewModel{}, fmt.Errorf("load orders: %w", err)
		}
		vm.StudentFeatureAccess = role != RoleStudent || studentFeatureAccess(vm.Orders)
		if vm.Documents, err = s.repository.ListDocuments(ctx, viewer, role); err != nil {
			return ViewModel{}, fmt.Errorf("load documents: %w", err)
		}
		if vm.ProgressStages, err = s.repository.ListProgressStages(ctx, viewer, role); err != nil {
			return ViewModel{}, fmt.Errorf("load progress stages: %w", err)
		}
		if vm.Schedules, err = s.repository.ListSchedules(ctx, viewer, role); err != nil {
			return ViewModel{}, fmt.Errorf("load schedules: %w", err)
		}
		if vm.Tasks, err = s.repository.ListTasks(ctx, viewer, role); err != nil {
			return ViewModel{}, fmt.Errorf("load tasks: %w", err)
		}
		if vm.Expenses, err = s.repository.ListExpenses(ctx, viewer, role); err != nil {
			return ViewModel{}, fmt.Errorf("load expenses: %w", err)
		}
		if role == RoleOwner || role == RoleStaff {
			if vm.PipelineStages, err = s.repository.ListPipelineStages(ctx); err != nil {
				return ViewModel{}, fmt.Errorf("load pipeline stages: %w", err)
			}
		}
		if role == RoleOwner && section == SectionServices {
			if vm.ServicePackages, err = s.repository.ListServicePackages(ctx); err != nil {
				return ViewModel{}, fmt.Errorf("load service packages: %w", err)
			}
		}
		if section == SectionChat && vm.StudentFeatureAccess {
			if vm.Conversations, err = s.repository.ListConversations(ctx, viewer); err != nil {
				return ViewModel{}, fmt.Errorf("load conversations: %w", err)
			}
			vm.ActiveChatID = selectConversation(vm.Conversations, options.ConversationID)
			if vm.ActiveChatID != "" {
				if vm.Messages, err = s.repository.ListMessages(ctx, viewer, vm.ActiveChatID); err != nil {
					return ViewModel{}, fmt.Errorf("load messages: %w", err)
				}
			}
		}
	}

	vm.ActiveOrderCode = options.OrderCode
	vm.InvoiceFilter = options.InvoiceFilter
	vm.FilterStatus = options.ClientStatus
	vm.FilterPackage = options.ClientPackage
	vm.FilterPIC = options.ClientPIC
	vm.FilterSearch = options.ClientSearch
	vm.ShowCreateForm = options.ShowCreateForm
	vm.ShowStageManager = options.ShowStageManager

	vm.UserName = viewer.Name
	vm.UserRole = viewer.Role.Label()
	vm.UserInitials = initials(viewer.Name)

	switch role {
	case RoleOwner:
		vm.RoleBadge = "OWNER"
		vm.RoleBadgeClass = "bg-owner text-white shadow-[0_12px_26px_rgba(185,138,24,0.24)]"
		vm.Search = "Cari klien, invoice, laporan..."
	case RoleStaff:
		vm.RoleBadge = "STAFF"
		vm.RoleBadgeClass = "bg-violet-600 text-white shadow-[0_12px_26px_rgba(124,58,237,0.24)]"
		vm.Search = "Cari klien, order, dokumen..."
	case RoleStudent:
		vm.RoleBadge = "CLIENT"
		vm.RoleBadgeClass = "bg-emerald-600 text-white shadow-[0_12px_26px_rgba(5,150,105,0.24)]"
		vm.Search = "Cari dokumen, jadwal, invoice..."
	}

	vm.Title, vm.Subtitle, vm.Description = pageCopy(role, section)
	if role == RoleStudent && section == SectionDashboard {
		vm.Title = fmt.Sprintf("Halo, %s!", viewer.Name)
	}

	return vm, nil
}

func CanViewRole(actor, target Role) bool {
	switch actor {
	case RoleOwner:
		return target == RoleOwner || target == RoleStaff || target == RoleStudent
	case RoleStaff:
		return target == RoleStaff || target == RoleStudent
	case RoleStudent:
		return target == RoleStudent
	default:
		return false
	}
}

func selectConversation(conversations []ChatConversation, requested string) string {
	if requested != "" {
		for _, conversation := range conversations {
			if conversation.ID == requested {
				return requested
			}
		}
	}
	if len(conversations) == 0 {
		return ""
	}
	return conversations[0].ID
}

func initials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "UA"
	}
	if len(parts) == 1 {
		return strings.ToUpper(parts[0][:1])
	}
	return strings.ToUpper(parts[0][:1] + parts[1][:1])
}

func validSection(role Role, section Section) bool {
	allowed := map[Role]map[Section]bool{
		RoleOwner: {
			SectionDashboard: true, SectionFinance: true, SectionClients: true, SectionPipeline: true,
			SectionServices: true, SectionReports: true, SectionSettings: true, SectionOrders: true,
			SectionInvoices: true, SectionChat: true,
		},
		RoleStaff: {
			SectionDashboard: true, SectionClients: true, SectionPipeline: true, SectionTasks: true,
			SectionDocuments: true, SectionExpenses: true, SectionCalendar: true, SectionChat: true,
			SectionSettings: true,
		},
		RoleStudent: {
			SectionDashboard: true, SectionProgress: true, SectionDocuments: true, SectionPayments: true,
			SectionCalendar: true, SectionChat: true,
		},
	}
	return allowed[role][section]
}

func navigation(role Role, active Section) []NavItem {
	items := map[Role][]NavItem{
		RoleOwner: {
			{Label: "Dashboard", Href: "/owner", Icon: "grid"},
			{Label: "Keuangan", Href: "/owner/finance", Icon: "wallet"},
			{Label: "Klien & Order", Href: "/owner/clients", Icon: "users"},
			{Label: "Pipeline", Href: "/owner/pipeline", Icon: "kanban"},
			{Label: "Paket & Layanan", Href: "/owner/services", Icon: "package"},
			{Label: "Invoice", Href: "/owner/invoices", Icon: "receipt"},
			{Label: "Laporan", Href: "/owner/reports", Icon: "chart"},
			{Label: "Chat", Href: "/owner/chat", Icon: "chat"},
			{Label: "Pengaturan", Href: "/owner/settings", Icon: "settings"},
		},
		RoleStaff: {
			{Label: "Dashboard", Href: "/staff", Icon: "grid"},
			{Label: "Klien", Href: "/staff/clients", Icon: "users"},
			{Label: "Pipeline", Href: "/staff/pipeline", Icon: "kanban"},
			{Label: "Tugas Saya", Href: "/staff/tasks", Icon: "calendar", Badge: "6"},
			{Label: "Dokumen", Href: "/staff/documents", Icon: "file"},
			{Label: "Pengeluaran", Href: "/staff/expenses", Icon: "wallet"},
			{Label: "Kalender", Href: "/staff/calendar", Icon: "calendar"},
			{Label: "Chat", Href: "/staff/chat", Icon: "chat"},
			{Label: "Pengaturan", Href: "/staff/settings", Icon: "settings"},
		},
		RoleStudent: {
			{Label: "Dashboard", Href: "/student", Icon: "grid"},
			{Label: "Progress Saya", Href: "/student/progress", Icon: "kanban"},
			{Label: "Dokumen Saya", Href: "/student/documents", Icon: "file"},
			{Label: "Pembayaran", Href: "/student/payments", Icon: "receipt"},
			{Label: "Jadwal", Href: "/student/calendar", Icon: "calendar"},
			{Label: "Chat Konsultan", Href: "/student/chat", Icon: "chat"},
		},
	}

	for i := range items[role] {
		if sectionFromHref(items[role][i].Href) == active {
			items[role][i].Active = true
		}
	}
	return items[role]
}

func sectionFromHref(href string) Section {
	switch href {
	case "/owner", "/staff", "/student":
		return SectionDashboard
	case "/owner/finance":
		return SectionFinance
	case "/owner/clients", "/staff/clients":
		return SectionClients
	case "/owner/pipeline", "/staff/pipeline":
		return SectionPipeline
	case "/owner/services":
		return SectionServices
	case "/owner/invoices":
		return SectionInvoices
	case "/owner/reports":
		return SectionReports
	case "/owner/chat":
		return SectionChat
	case "/staff/tasks":
		return SectionTasks
	case "/staff/documents", "/student/documents":
		return SectionDocuments
	case "/staff/expenses":
		return SectionExpenses
	case "/staff/calendar", "/student/calendar":
		return SectionCalendar
	case "/staff/chat", "/student/chat":
		return SectionChat
	case "/owner/settings", "/staff/settings", "/student/settings":
		return SectionSettings
	case "/student/progress":
		return SectionProgress
	case "/student/payments":
		return SectionPayments
	default:
		return SectionDashboard
	}
}

func pageCopy(role Role, section Section) (string, string, string) {
	switch role {
	case RoleOwner:
		switch section {
		case SectionFinance:
			return "Detail Keuangan", "Ringkasan arus kas dan performa pembayaran.", "Pantau pemasukan, pengeluaran, profit, piutang, dan transaksi terbaru."
		case SectionClients:
			return "Daftar Klien", "Kelola semua client, PIC, paket, status, dan progress.", "Tabel operasional seluruh client dengan filter paket, status, staff, dan aksi cepat."
		case SectionPipeline:
			return "Pipeline Keseluruhan", "Pantau tahapan konsultasi sampai selesai.", "Kanban pipeline untuk melihat beban proses per tahap dan bottleneck operasional."
		case SectionServices:
			return "Paket & Layanan", "Konfigurasi paket profesional, basic, dan layanan satuan.", "Daftar layanan, harga, dan volume order yang dapat dikelola owner."
		case SectionInvoices:
			return "Invoice & Pembayaran", "Pantau tagihan, sisa pembayaran, dan follow up.", "Detail invoice dengan aksi kirim WhatsApp, email, download PDF, dan tandai lunas."
		case SectionReports:
			return "Laporan Keuangan", "Analisis pendapatan, pengeluaran, profit, dan piutang.", "Grafik serta ringkasan bulanan untuk ekspor PDF dan Excel."
		case SectionChat:
			return "Chat Operasional", "Pantau percakapan staff dan client.", "Akses owner untuk memantau komunikasi dan eskalasi client."
		case SectionSettings:
			return "Pengaturan", "Kelola preferensi perusahaan dan akses role.", "Kontrol keamanan, brand, notifikasi, dan konfigurasi operasional."
		default:
			return "Dashboard Owner", "Ringkasan bisnis per 25 Juni 2026.", "Tampilan utama untuk kondisi perusahaan, pipeline, keuangan, aktivitas, dan reminder."
		}
	case RoleStaff:
		switch section {
		case SectionClients:
			return "Klien Saya", "Daftar client yang sedang ditangani staff.", "Update progress, cek dokumen, dan tindak lanjuti order client."
		case SectionPipeline:
			return "Pipeline Saya", "Fokus pada tahapan client milik staff.", "Kanban operasional untuk pekerjaan konsultasi, dokumen, visa, dan keberangkatan."
		case SectionTasks:
			return "Tugas Saya", "Task list harian berdasarkan prioritas.", "Kelola follow up dokumen, appointment, pembayaran, dan upload berkas."
		case SectionDocuments:
			return "Dokumen Review", "Dokumen client yang perlu diperiksa.", "Approve, reject, atau minta revisi dokumen dari client."
		case SectionExpenses:
			return "Pengeluaran Operasional", "Catat biaya untuk setiap berkas atau kebutuhan client.", "Input kategori, nominal, bukti, dan status approval tanpa melihat profit perusahaan."
		case SectionCalendar:
			return "Kalender", "Jadwal interview, medical, appointment, dan follow up.", "Agenda staff untuk memastikan tidak ada jadwal client yang terlewat."
		case SectionChat:
			return "Chat Klien", "Percakapan prioritas dengan client aktif.", "Akses cepat ke pesan yang butuh balasan staff."
		case SectionSettings:
			return "Pengaturan Staff", "Preferensi akun, notifikasi, dan keamanan.", "Atur akun staff tanpa akses ke data keuangan perusahaan."
		default:
			return "Dashboard Staff", "Ringkasan aktivitas dan prioritas hari ini.", "Dashboard staff untuk tugas, pipeline, reminder, jadwal, dokumen, dan pengeluaran operasional."
		}
	case RoleStudent:
		switch section {
		case SectionProgress:
			return "Progress Saya", "Lihat posisi proses dan estimasi selesai.", "Progress bar setiap tahapan mulai konsultasi sampai keberangkatan."
		case SectionDocuments:
			return "Dokumen Saya", "Upload dan pantau status review dokumen.", "Passport, ijazah, transkrip, foto, KTP, KK, akta, autobiografi, dan mutasi rekening."
		case SectionPayments:
			return "Pembayaran", "Invoice, total tagihan, sisa tagihan, dan bukti bayar.", "Download invoice dan upload bukti pembayaran untuk diverifikasi."
		case SectionCalendar:
			return "Jadwal Saya", "Interview, medical, appointment, dan keberangkatan.", "Kalender personal untuk semua agenda proses keberangkatan."
		case SectionChat:
			return "Chat Konsultan", "Hubungi staff untuk update proses.", "Percakapan langsung dengan konsultan yang menangani proses kamu."
		case SectionSettings:
			return "Pengaturan Client", "Profil, preferensi notifikasi, dan keamanan akun.", "Kelola data kontak dan preferensi komunikasi."
		default:
			return "Halo!", "Pantau progress studi Taiwan kamu dari satu tempat.", "Dashboard client untuk progress, biodata, pembayaran, dokumen, jadwal, dan pengumuman."
		}
	default:
		return "Dashboard", "Ringkasan aktivitas.", "University agency dashboard."
	}
}

func dateLabelID(value time.Time) string {
	months := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	return fmt.Sprintf("%d %s %d", value.Day(), months[int(value.Month())-1], value.Year())
}

func studentFeatureAccess(orders []Order) bool {
	for _, order := range orders {
		if order.Status == OrderWaitingVerification || order.Status == OrderPaid {
			return true
		}
	}
	return false
}
