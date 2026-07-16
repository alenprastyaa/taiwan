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
	SectionExpenses  Section = "expenses"
	SectionOrders    Section = "orders"
)

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
	Username string
	Password string
	Name     string
	Email    string
	Phone    string
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
