package dashboard

import (
	"fmt"
	"time"
)

type Role string

const (
	RoleOwner   Role = "owner"
	RoleStaff   Role = "staff"
	RoleStudent Role = "student"
)

type Section string

const (
	SectionDashboard Section = "dashboard"
	SectionFinance   Section = "finance"
	SectionClients   Section = "clients"
	SectionPipeline  Section = "pipeline"
	SectionServices  Section = "services"
	SectionTasks     Section = "tasks"
	SectionDocuments Section = "documents"
	SectionInvoices  Section = "invoices"
	SectionCalendar  Section = "calendar"
	SectionReports   Section = "reports"
	SectionSettings  Section = "settings"
	SectionProgress  Section = "progress"
	SectionPayments  Section = "payments"
	SectionChat      Section = "chat"
	SectionExpenses     Section = "expenses"
	SectionOrders       Section = "orders"
	SectionTemplates    Section = "templates"
	SectionInstitutions Section = "institutions"
	SectionIntake       Section = "intake"
	SectionActivity     Section = "activity"
	SectionAgreement    Section = "agreement"
	SectionLogistics    Section = "logistics"
	SectionStaff        Section = "staff"
)

// CurrentAgreementVersion tags every signature with the agreement text
// version in force at signing time — bump this if AgreementText changes so
// past signatures remain a faithful record of what was actually agreed to.
const CurrentAgreementVersion = "v1"

// AgreementText is the perjanjian kerjasama shown to every client/student
// before they can use the rest of their dashboard. Not owner-editable in
// this pass — only the gating/signing/PDF flow was requested.
const AgreementText = `SURAT PERJANJIAN KERJASAMA LAYANAN KONSULTASI PENDIDIKAN

Dengan menyetujui perjanjian ini, Client menyatakan bahwa:

1. Client telah membaca dan memahami seluruh ketentuan layanan yang diberikan oleh agensi, termasuk namun tidak terbatas pada: pendampingan aplikasi kampus, pengurusan dokumen, penerjemahan, legalisir, dan pengurusan visa.
2. Client bertanggung jawab atas keaslian dan kebenaran seluruh dokumen serta data yang diserahkan kepada agensi.
3. Biaya yang telah dibayarkan mengikuti ketentuan paket layanan yang dipilih dan tidak dapat dikembalikan setelah proses aplikasi dimulai, kecuali diatur lain secara tertulis.
4. Agensi akan menjaga kerahasiaan data pribadi Client dan hanya menggunakannya untuk keperluan proses aplikasi pendidikan dan visa.
5. Estimasi waktu proses dapat berubah sewaktu-waktu mengikuti kebijakan kampus tujuan, kedutaan, atau lembaga terkait di luar kendali agensi.
6. Client wajib melakukan follow up aktif dan responsif terhadap permintaan dokumen atau informasi tambahan dari agensi.

Perjanjian ini berlaku sejak Client menyetujui secara elektronik melalui sistem ini, dan menjadi bukti sah kesepakatan kedua belah pihak.`

type ViewModel struct {
	AppName              string
	AppURL               string
	Viewer               User
	Role                 Role
	Section              Section
	Title                string
	Subtitle             string
	Description          string
	DateLabel            string
	Search               string
	UserName             string
	UserRole             string
	UserInitials         string
	RoleBadge            string
	RoleBadgeClass       string
	NotificationText     string
	Navigation           []NavItem
	RoleSwitch           []RoleSwitchItem
	Snapshot             CompanySnapshot
	Clients              []ClientProfile
	Orders               []Order
	Documents            []Document
	ProgressStages       []ProgressStage
	PipelineStages       []PipelineStage
	ServicePackages      []ServicePackage
	Schedules            []ScheduleItem
	Tasks                []Task
	Expenses             []Expense
	ExpenseCategories    []string
	Shipments            []Shipment
	ShipmentCouriers     []string
	Staff                []User
	StaffAccounts        []User
	TextTemplates        []TextTemplate
	InstitutionContacts  []InstitutionContact
	IntakeForm           ClientIntakeForm
	HasIntakeForm        bool
	IntakeForms          []ClientIntakeForm
	EditClientID         string
	ActivityLog          []ActivityLog
	FilterDate           string
	Agreement            ClientAgreement
	HasSignedAgreement   bool
	Conversations        []ChatConversation
	Messages             []ChatMessage
	ActiveChatID         string
	Flash                string
	StudentFeatureAccess bool
	ActiveOrderCode      string
	InvoiceFilter        string
	FilterStatus         string
	FilterPackage        string
	FilterPIC            string
	FilterSearch         string
	ShowCreateForm       bool
	ShowStageManager     bool
}

type NavItem struct {
	Label  string
	Href   string
	Icon   string
	Badge  string
	Active bool
}

type RoleSwitchItem struct {
	Label  string
	Href   string
	Active bool
}

type User struct {
	ID           string
	Username     string
	Name         string
	Email        string
	Phone        string
	Role         Role
	PasswordHash []byte
	Active       bool
	CreatedAt    time.Time
}

type CreateStudentInput struct {
	Username     string
	Password     string
	Name         string
	Email        string
	Phone        string
	PackageName  string
	Country      string
	Campus       string
	PICStaffID   string
	InvoiceTotal int64
	AmountPaid   int64
}

type CreateStaffInput struct {
	Username string
	Password string
	Name     string
	Email    string
	Phone    string
}

type UpdateStaffInput struct {
	Name  string
	Email string
	Phone string
}

// UpdateOwnProfileInput is the self-service counterpart to UpdateStaffInput —
// any logged-in user (owner, staff, or client) editing their own name,
// email, or phone, as opposed to an owner editing someone else's staff
// account.
type UpdateOwnProfileInput struct {
	Name  string
	Email string
	Phone string
}

type CreateTaskInput struct {
	ClientID  string
	TimeLabel string
	Title     string
	Priority  string
}

type CreateExpenseInput struct {
	ClientID    string
	Need        string
	Category    string
	Amount      int64
	DateLabel   string
	Description string
	FileName    string
	StoragePath string
}

type CreateScheduleInput struct {
	ClientID  string
	Title     string
	DateLabel string
	TimeLabel string
	Location  string
}

type ClientProfile struct {
	ID           string
	UserID       string
	Name         string
	Email        string
	Phone        string
	Country      string
	Campus       string
	PackageName  string
	PICStaffID   string
	PICName      string
	Status       string
	Progress     int
	LastSchedule string
	CurrentStage string
	CreatedAt    time.Time
	Active       bool
}

// UpdateClientInput edits an existing client's profile fields. Username and
// password aren't included here — those go through ResetStudentPassword,
// mirroring how UpdateStaffInput leaves credentials to a separate action.
type UpdateClientInput struct {
	Name        string
	Email       string
	Phone       string
	PackageName string
	Country     string
	Campus      string
	PICStaffID  string
	// InvoiceTotal is owner-only, mirroring CreateStudentInput's field of the
	// same name: 0 means "leave the client's order alone." A positive value
	// updates the client's latest order's total (or opens a new order if
	// they don't have one yet) — the same total-only edit UpdateOrderInput
	// allows, just reachable from the client form too.
	InvoiceTotal int64
}

type OrderStatus string

const (
	OrderUnpaid              OrderStatus = "unpaid"
	OrderWaitingVerification OrderStatus = "waiting_verification"
	OrderPaid                OrderStatus = "paid"
)

type Order struct {
	ID               string
	Code             string
	ClientID         string
	ClientName       string
	PackageName      string
	Total            int64
	Paid             int64
	Status           OrderStatus
	DueDate          time.Time
	ProofNote        string
	ProofFileName    string
	ProofStoragePath string
	CreatedAt        time.Time
	PaidAt           *time.Time
}

// CreateOrderInput opens a new invoice for an existing client — the
// standalone counterpart to the InvoiceTotal/AmountPaid fields on
// CreateStudentInput, for clients who already have an account and need a
// new/renewal order.
type CreateOrderInput struct {
	ClientID    string
	PackageName string
	Total       int64
	Paid        int64
	DueDate     time.Time
}

// UpdateOrderInput edits an order's package/total/due date. Paid amount is
// deliberately excluded — that goes through RecordOrderPayment/
// MarkOrderPaidByCode so the paid/status invariant stays in one place.
type UpdateOrderInput struct {
	PackageName string
	Total       int64
	DueDate     time.Time
}

type DocumentStatus string

const (
	DocumentMissing  DocumentStatus = "missing"
	DocumentReview   DocumentStatus = "review"
	DocumentRevision DocumentStatus = "revision"
	DocumentApproved DocumentStatus = "approved"
)

type Document struct {
	ID          string
	ClientID    string
	ClientName  string
	Name        string
	Status      DocumentStatus
	Reviewer    string
	ReviewNote  string
	FileName    string
	StoragePath string
	UpdatedAt   time.Time
}

type ProgressStageStatus string

const (
	ProgressStageDone    ProgressStageStatus = "done"
	ProgressStageActive  ProgressStageStatus = "active"
	ProgressStagePending ProgressStageStatus = "pending"
)

type ProgressStage struct {
	ID          string
	ClientID    string
	Step        int
	Title       string
	Description string
	Status      ProgressStageStatus
	Progress    int
	DueLabel    string
	PICName     string
	UpdatedAt   time.Time
}

type ScheduleItem struct {
	ID        string
	ClientID  string
	Title     string
	DateLabel string
	TimeLabel string
	Location  string
	Status    string
	CreatedAt time.Time
}

// PipelineStage is an owner/staff-defined column on the pipeline kanban board.
// Unlike ProgressStage (a per-client onboarding checklist), this is shared,
// company-wide configuration: the set of stages every client can be placed into.
type PipelineStage struct {
	ID        string
	Name      string
	Position  int
	Tone      string
	CreatedAt time.Time
}

// ServicePackage is an owner-defined catalog entry on the "Paket & Layanan"
// page — the price list clients are sold, not tied to any one client/order.
type ServicePackage struct {
	ID          string
	Name        string
	Category    string
	Description string
	Price       int64
	PriceIsFrom bool
	Highlights  string
	Position    int
	CreatedAt   time.Time
}

type ServicePackageInput struct {
	Name        string
	Category    string
	Description string
	Price       int64
	PriceIsFrom bool
	Highlights  string
}

// TextTemplate is a staff/owner-managed canned reply — reusable copy-paste
// text for answering repeated client questions (payment, docs, timeline, etc.).
type TextTemplate struct {
	ID        string
	Title     string
	Body      string
	Category  string
	Position  int
	CreatedAt time.Time
}

type TextTemplateInput struct {
	Title    string
	Body     string
	Category string
}

// InstitutionContact is an owner/staff-managed directory entry for partner
// institutions (translators, legalization offices, etc.) with a WhatsApp
// number for quick follow-up via click-to-chat.
type InstitutionContact struct {
	ID        string
	Name      string
	Category  string
	Phone     string
	Notes     string
	Position  int
	CreatedAt time.Time
}

type InstitutionContactInput struct {
	Name     string
	Category string
	Phone    string
	Notes    string
}

// ClientIntakeForm is the client-filled biodata form (mirrors the agency's
// Google Form) — one row per client, editable anytime by the client, and
// aggregated live for owner/staff on the "Data Client / Formulir" page.
// Owner/staff can also fill in or correct it on the client's behalf (e.g.
// missing fields, a typo caught later) via the same save path.
type ClientIntakeForm struct {
	ID             string
	ClientID       string
	ClientName     string
	Email          string
	FullNameEn     string
	Gender         string
	DateOfBirth    string
	PlaceOfBirth   string
	PassportNumber string
	PhoneNumber    string
	Address        string
	PostalCode     string
	FatherName     string
	FatherDOB      string
	FatherPhone    string
	MotherName     string
	MotherDOB      string
	MotherPhone    string
	SchoolName     string
	SchoolLocation string
	DatesEnrolled  string
	DatesGraduate  string
	SocialMediaIG  string
	SubmittedAt    time.Time
	UpdatedAt      time.Time
}

// ActivityLog is an immutable, append-only event stream of staff actions
// (auto-logged from existing write methods) plus manual notes — the "what
// did staff do today" report the owner can review per staff per day.
type ActivityLog struct {
	ID          string
	StaffID     string
	StaffName   string
	ClientID    string
	ClientName  string
	ActionType  string
	Description string
	CreatedAt   time.Time
}

// ClientAgreement is the signed record of a client accepting the agency's
// perjanjian: a checkbox + typed full name + timestamp + IP address, used
// as the legal proof exported to PDF for both parties.
type ClientAgreement struct {
	ID               string
	ClientID         string
	ClientName       string
	AgreementVersion string
	AgreementText    string
	FullNameTyped    string
	AgreedAt         time.Time
	IPAddress        string
	UserAgent        string
}

type ClientIntakeFormInput struct {
	Email          string
	FullNameEn     string
	Gender         string
	DateOfBirth    string
	PlaceOfBirth   string
	PassportNumber string
	PhoneNumber    string
	Address        string
	PostalCode     string
	FatherName     string
	FatherDOB      string
	FatherPhone    string
	MotherName     string
	MotherDOB      string
	MotherPhone    string
	SchoolName     string
	SchoolLocation string
	DatesEnrolled  string
	DatesGraduate  string
	SocialMediaIG  string
}

type TaskStatus string

const (
	TaskOpen TaskStatus = "open"
	TaskDone TaskStatus = "done"
)

type Task struct {
	ID         string
	StaffID    string
	ClientID   string
	ClientName string
	TimeLabel  string
	Title      string
	Priority   string
	Status     TaskStatus
}

type ExpenseStatus string

const (
	ExpenseWaiting  ExpenseStatus = "waiting"
	ExpenseRecorded ExpenseStatus = "recorded"
)

type Expense struct {
	ID                 string
	StaffID            string
	ClientID           string
	ClientName         string
	Need               string
	Category           string
	Amount             int64
	Status             ExpenseStatus
	DateLabel          string
	Description        string
	ReceiptFileName    string
	ReceiptStoragePath string
}

type ShipmentDirection string

const (
	// ShipmentOutgoing is agency -> client/institution (e.g. legalized
	// documents mailed back to the client, or forwarded to an embassy).
	ShipmentOutgoing ShipmentDirection = "outgoing"
	// ShipmentIncoming is client/institution -> agency (e.g. the client
	// couriers original documents to the office).
	ShipmentIncoming ShipmentDirection = "incoming"
)

type ShipmentStatus string

const (
	ShipmentShipped   ShipmentStatus = "shipped"
	ShipmentDelivered ShipmentStatus = "delivered"
)

// Shipment tracks a single courier shipment of physical documents in either
// direction, so owner/staff/client can all see what was sent, its tracking
// number, and whether it has arrived — mirrors Expense's shape (client-linked,
// staff-recorded, simple status).
type Shipment struct {
	ID                string
	ClientID          string
	ClientName        string
	StaffID           string
	Direction         ShipmentDirection
	Courier           string
	TrackingNumber    string
	Contents          string
	SenderAddress     string
	RecipientAddress  string
	Status            ShipmentStatus
	ShippedDateLabel  string
	ReceivedDateLabel string
	CreatedAt         time.Time
}

type CreateShipmentInput struct {
	ClientID         string
	Direction        string
	Courier          string
	TrackingNumber   string
	Contents         string
	SenderAddress    string
	RecipientAddress string
	ShippedDateLabel string
}

type ChatConversation struct {
	ID          string
	ClientID    string
	ClientName  string
	StaffID     string
	StaffName   string
	LastMessage string
	UpdatedAt   time.Time
}

type ChatMessage struct {
	ID             string
	ConversationID string
	SenderID       string
	SenderName     string
	SenderRole     Role
	Body           string
	CreatedAt      time.Time
}

func ParseRole(value string) (Role, bool) {
	switch Role(value) {
	case RoleOwner, RoleStaff, RoleStudent:
		return Role(value), true
	case "client":
		return RoleStudent, true
	default:
		return "", false
	}
}

func ParseSection(value string) Section {
	if value == "" {
		return SectionDashboard
	}
	return Section(value)
}

func (r Role) String() string {
	return string(r)
}

func (s Section) String() string {
	return string(s)
}

func (vm ViewModel) Path() string {
	if vm.Section == SectionDashboard {
		return fmt.Sprintf("/%s", vm.Role)
	}
	return fmt.Sprintf("/%s/%s", vm.Role, vm.Section)
}

func (r Role) Label() string {
	switch r {
	case RoleOwner:
		return "Owner"
	case RoleStaff:
		return "Staff Konsultan"
	case RoleStudent:
		return "Client"
	default:
		return string(r)
	}
}
