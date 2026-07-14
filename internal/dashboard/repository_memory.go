package dashboard

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"university_agency/internal/security"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrForbidden    = errors.New("forbidden")
	ErrDuplicate    = errors.New("duplicate")
	ErrInvalidInput = errors.New("invalid input")
)

type MemoryRepository struct {
	mu            sync.RWMutex
	users         map[string]User
	usersByName   map[string]string
	clients       map[string]ClientProfile
	orders        map[string]Order
	orderByCode   map[string]string
	documents     map[string]Document
	stages        map[string]ProgressStage
	schedules     map[string]ScheduleItem
	tasks         map[string]Task
	expenses      map[string]Expense
	conversations map[string]ChatConversation
	messages      map[string][]ChatMessage
	next          int64
}

func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		users:         make(map[string]User),
		usersByName:   make(map[string]string),
		clients:       make(map[string]ClientProfile),
		orders:        make(map[string]Order),
		orderByCode:   make(map[string]string),
		documents:     make(map[string]Document),
		stages:        make(map[string]ProgressStage),
		schedules:     make(map[string]ScheduleItem),
		tasks:         make(map[string]Task),
		expenses:      make(map[string]Expense),
		conversations: make(map[string]ChatConversation),
		messages:      make(map[string][]ChatMessage),
		next:          100,
	}
	repo.seed()
	return repo
}

func (r *MemoryRepository) seed() {
	now := time.Now()
	owner := r.mustUser("user-owner", "owner", "owner12345", "Amanda Wijaya", "owner@taiwanedu.test", "", RoleOwner, now)
	staff := r.mustUser("user-staff-rina", "staff", "staff12345", "Rina Amalia Putri", "rina@taiwanedu.test", "", RoleStaff, now)
	bayu := r.mustUser("user-staff-bayu", "bayu", "bayu12345", "Bayu Pratama Wijaya", "bayu@taiwanedu.test", "", RoleStaff, now)
	student := r.mustUser("user-student-budi", "student", "student12345", "Budi Santoso", "budi@example.test", "081234567890", RoleStudent, now)
	_ = owner
	_ = bayu
	_ = student

	clients := []ClientProfile{
		{ID: "client-budi", UserID: "user-student-budi", Name: "Budi Santoso", Email: "budi@example.test", Phone: "081234567890", Country: "Taiwan", Campus: "National Taiwan University", PackageName: "Paket Profesional", PICStaffID: staff.ID, PICName: staff.Name, Status: "Persiapan Dokumen", Progress: 45, LastSchedule: "20 Jun 2026", CurrentStage: "Persiapan Dokumen", CreatedAt: now.AddDate(0, -2, 0)},
		{ID: "client-siti", Name: "Siti Aisyah", Email: "siti@example.test", Phone: "081222333444", Country: "Taiwan", Campus: "National Cheng Kung University", PackageName: "Paket Basic", PICStaffID: staff.ID, PICName: staff.Name, Status: "Persiapan Dokumen", Progress: 60, LastSchedule: "22 Jun 2026", CurrentStage: "Translate Dokumen", CreatedAt: now.AddDate(0, -1, -6)},
		{ID: "client-dewi", Name: "Dewi Lestari", Email: "dewi@example.test", Phone: "081987654321", Country: "Taiwan", Campus: "NTU", PackageName: "Paket Profesional", PICStaffID: staff.ID, PICName: staff.Name, Status: "LOA", Progress: 80, LastSchedule: "18 Jun 2026", CurrentStage: "LOA", CreatedAt: now.AddDate(0, -3, 0)},
		{ID: "client-ricky", Name: "Ricky Pratama", Email: "ricky@example.test", Phone: "081111222333", Country: "Taiwan", Campus: "Taipei Medical University", PackageName: "Layanan Satuan (Visa)", PICStaffID: staff.ID, PICName: staff.Name, Status: "Visa", Progress: 70, LastSchedule: "25 Jun 2026", CurrentStage: "Visa", CreatedAt: now.AddDate(0, -1, -12)},
	}
	for _, client := range clients {
		r.clients[client.ID] = client
	}

	r.addOrder(Order{ID: "order-0102", Code: "ORD-2026-0102", ClientID: "client-budi", ClientName: "Budi Santoso", PackageName: "Paket Profesional", Total: 7500000, Paid: 0, Status: OrderUnpaid, DueDate: now.AddDate(0, 0, 7), CreatedAt: now.AddDate(0, 0, -5)})
	r.addOrder(Order{ID: "order-0103", Code: "ORD-2026-0103", ClientID: "client-budi", ClientName: "Budi Santoso", PackageName: "Paket Profesional", Total: 7500000, Paid: 7500000, Status: OrderPaid, DueDate: now.AddDate(0, 0, -5), CreatedAt: now.AddDate(0, -1, 0), PaidAt: ptrTime(now.AddDate(0, 0, -2))})
	r.addOrder(Order{ID: "order-0104", Code: "ORD-2026-0104", ClientID: "client-siti", ClientName: "Siti Aisyah", PackageName: "Paket Basic", Total: 4500000, Paid: 0, Status: OrderWaitingVerification, ProofNote: "Transfer BCA 25 Juni", DueDate: now.AddDate(0, 0, 3), CreatedAt: now.AddDate(0, 0, -3)})
	r.addOrder(Order{ID: "order-0105", Code: "ORD-2026-0105", ClientID: "client-dewi", ClientName: "Dewi Lestari", PackageName: "Paket Profesional", Total: 10000000, Paid: 10000000, Status: OrderPaid, DueDate: now.AddDate(0, 0, -9), CreatedAt: now.AddDate(0, -2, 0), PaidAt: ptrTime(now.AddDate(0, -1, -20))})

	for _, doc := range []Document{
		{ID: "doc-budi-passport", ClientID: "client-budi", ClientName: "Budi Santoso", Name: "Passport", Status: DocumentApproved, Reviewer: staff.Name, UpdatedAt: now.AddDate(0, 0, -4)},
		{ID: "doc-budi-ijazah", ClientID: "client-budi", ClientName: "Budi Santoso", Name: "Ijazah Terakhir", Status: DocumentApproved, Reviewer: staff.Name, UpdatedAt: now.AddDate(0, 0, -5)},
		{ID: "doc-budi-foto", ClientID: "client-budi", ClientName: "Budi Santoso", Name: "Foto Background Putih 4x6", Status: DocumentReview, Reviewer: staff.Name, UpdatedAt: now},
		{ID: "doc-siti-transkrip", ClientID: "client-siti", ClientName: "Siti Aisyah", Name: "Transkrip Nilai", Status: DocumentReview, Reviewer: staff.Name, UpdatedAt: now.Add(-3 * time.Hour)},
		{ID: "doc-ricky-passport", ClientID: "client-ricky", ClientName: "Ricky Pratama", Name: "Passport", Status: DocumentReview, Reviewer: staff.Name, UpdatedAt: now.Add(-5 * time.Hour)},
	} {
		r.documents[doc.ID] = doc
	}
	for _, stage := range []ProgressStage{
		{ID: "stage-budi-1", ClientID: "client-budi", Step: 1, Title: "Konsultasi Awal", Description: "Kebutuhan studi dan target intake dikunci.", Status: ProgressStageDone, Progress: 100, DueLabel: "Selesai", PICName: staff.Name, UpdatedAt: now.AddDate(0, -2, 0)},
		{ID: "stage-budi-2", ClientID: "client-budi", Step: 2, Title: "Persiapan Dokumen", Description: "Dokumen identitas dan akademik dikumpulkan.", Status: ProgressStageActive, Progress: 60, DueLabel: "28 Mei 2026", PICName: staff.Name, UpdatedAt: now.AddDate(0, 0, -2)},
		{ID: "stage-budi-3", ClientID: "client-budi", Step: 3, Title: "Translate & Legalisir", Description: "Dokumen diterjemahkan dan dilegalisir.", Status: ProgressStagePending, Progress: 0, DueLabel: "-", PICName: staff.Name, UpdatedAt: now.AddDate(0, 0, -2)},
	} {
		r.stages[stage.ID] = stage
	}
	for _, schedule := range []ScheduleItem{
		{ID: "schedule-budi-1", ClientID: "client-budi", Title: "Review Form Data Diri", DateLabel: "26 Jun 2026", TimeLabel: "10:00", Location: "Online Meeting", Status: "Terjadwal", CreatedAt: now.AddDate(0, 0, 1)},
		{ID: "schedule-budi-2", ClientID: "client-budi", Title: "Deadline Upload Surat Rekomendasi", DateLabel: "28 Jun 2026", TimeLabel: "23:59", Location: "Portal Client", Status: "Menunggu Dokumen", CreatedAt: now.AddDate(0, 0, 3)},
	} {
		r.schedules[schedule.ID] = schedule
	}

	for _, task := range []Task{
		{ID: "task-followup", StaffID: staff.ID, ClientID: "client-budi", ClientName: "Budi Santoso", TimeLabel: "09:00", Title: "Follow up dokumen", Priority: "Tinggi", Status: TaskOpen},
		{ID: "task-translate", StaffID: staff.ID, ClientID: "client-siti", ClientName: "Siti Aisyah", TimeLabel: "10:00", Title: "Translate Transkrip Nilai", Priority: "Sedang", Status: TaskOpen},
		{ID: "task-appointment", StaffID: staff.ID, ClientID: "client-dewi", ClientName: "Dewi Lestari", TimeLabel: "11:00", Title: "Appointment Kedutaan", Priority: "Tinggi", Status: TaskOpen},
		{ID: "task-passport", StaffID: staff.ID, ClientID: "client-ricky", ClientName: "Ricky Pratama", TimeLabel: "14:00", Title: "Upload Passport", Priority: "Rendah", Status: TaskDone},
	} {
		r.tasks[task.ID] = task
	}

	for _, expense := range []Expense{
		{ID: "expense-1", StaffID: staff.ID, ClientID: "client-budi", ClientName: "Budi Santoso", Need: "Legalisir Kemenkumham", Category: "Legalisir", Amount: 450000, Status: ExpenseWaiting, DateLabel: "25 Jun 2026"},
		{ID: "expense-2", StaffID: staff.ID, ClientID: "client-siti", ClientName: "Siti Aisyah", Need: "Translate Transkrip Nilai", Category: "Translate", Amount: 1200000, Status: ExpenseWaiting, DateLabel: "24 Jun 2026"},
		{ID: "expense-3", StaffID: staff.ID, ClientID: "client-ricky", ClientName: "Ricky Pratama", Need: "Pengiriman Dokumen DHL", Category: "Lainnya", Amount: 300000, Status: ExpenseRecorded, DateLabel: "22 Jun 2026"},
	} {
		r.expenses[expense.ID] = expense
	}

	r.addConversation(ChatConversation{ID: "chat-budi", ClientID: "client-budi", ClientName: "Budi Santoso", StaffID: staff.ID, StaffName: staff.Name, LastMessage: "Form data diri masih perlu dilengkapi.", UpdatedAt: now})
	r.addMessage(ChatMessage{ID: "msg-1", ConversationID: "chat-budi", SenderID: staff.ID, SenderName: staff.Name, SenderRole: RoleStaff, Body: "Halo, dokumen ijazah sudah masuk. Saya cek dulu ya.", CreatedAt: now.Add(-2 * time.Hour)})
	r.addMessage(ChatMessage{ID: "msg-2", ConversationID: "chat-budi", SenderID: "user-student-budi", SenderName: "Budi Santoso", SenderRole: RoleStudent, Body: "Baik kak, terima kasih.", CreatedAt: now.Add(-90 * time.Minute)})
	r.addMessage(ChatMessage{ID: "msg-3", ConversationID: "chat-budi", SenderID: staff.ID, SenderName: staff.Name, SenderRole: RoleStaff, Body: "Form data diri masih perlu dilengkapi bagian alamat Taiwan.", CreatedAt: now.Add(-35 * time.Minute)})
}

func (r *MemoryRepository) mustUser(id, username, password, name, email, phone string, role Role, now time.Time) User {
	hash, err := security.HashPassword(password)
	if err != nil {
		panic(err)
	}
	user := User{ID: id, Username: usernameKey(username), Name: name, Email: email, Phone: phone, Role: role, PasswordHash: hash, Active: true, CreatedAt: now}
	r.users[user.ID] = user
	r.usersByName[user.Username] = user.ID
	return user
}

func (r *MemoryRepository) addOrder(order Order) {
	r.orders[order.ID] = order
	r.orderByCode[strings.ToUpper(order.Code)] = order.ID
}

func (r *MemoryRepository) addConversation(conversation ChatConversation) {
	r.conversations[conversation.ID] = conversation
}

func (r *MemoryRepository) addMessage(message ChatMessage) {
	r.messages[message.ConversationID] = append(r.messages[message.ConversationID], message)
	if conversation, ok := r.conversations[message.ConversationID]; ok {
		conversation.LastMessage = message.Body
		conversation.UpdatedAt = message.CreatedAt
		r.conversations[message.ConversationID] = conversation
	}
}

func (r *MemoryRepository) FindUserByUsername(ctx context.Context, username string) (User, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.usersByName[usernameKey(username)]
	if !ok {
		return User{}, ErrNotFound
	}
	user := r.users[id]
	if !user.Active {
		return User{}, ErrForbidden
	}
	return user, nil
}

func (r *MemoryRepository) FindUserByID(ctx context.Context, id string) (User, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok || !user.Active {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (r *MemoryRepository) CreateStudent(ctx context.Context, input CreateStudentInput) (User, error) {
	_ = ctx
	username := usernameKey(input.Username)
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(input.Email)
	phone := strings.TrimSpace(input.Phone)
	if username == "" || len(input.Password) < 8 || name == "" {
		return User{}, ErrInvalidInput
	}

	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return User{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.usersByName[username]; exists {
		return User{}, ErrDuplicate
	}

	now := time.Now()
	r.next++
	userID := fmt.Sprintf("user-student-%d", r.next)
	clientID := fmt.Sprintf("client-%d", r.next)
	user := User{
		ID:           userID,
		Username:     username,
		Name:         name,
		Email:        email,
		Phone:        phone,
		Role:         RoleStudent,
		PasswordHash: hash,
		Active:       true,
		CreatedAt:    now,
	}
	r.users[user.ID] = user
	r.usersByName[user.Username] = user.ID

	staff := r.users["user-staff-rina"]
	client := ClientProfile{
		ID:           clientID,
		UserID:       user.ID,
		Name:         user.Name,
		Email:        user.Email,
		Phone:        user.Phone,
		Country:      "Taiwan",
		Campus:       "Belum dipilih",
		PackageName:  "Belum memilih paket",
		PICStaffID:   staff.ID,
		PICName:      staff.Name,
		Status:       "Registrasi",
		Progress:     0,
		LastSchedule: "-",
		CurrentStage: "Registrasi akun",
		CreatedAt:    now,
	}
	r.clients[client.ID] = client
	r.stages[fmt.Sprintf("stage-%d", r.next)] = ProgressStage{
		ID:          fmt.Sprintf("stage-%d", r.next),
		ClientID:    client.ID,
		Step:        1,
		Title:       "Registrasi akun",
		Description: "Akun client sudah dibuat dan menunggu pemilihan paket serta jadwal konsultasi.",
		Status:      ProgressStageActive,
		Progress:    client.Progress,
		DueLabel:    "-",
		PICName:     staff.Name,
		UpdatedAt:   now,
	}

	conversation := ChatConversation{
		ID:         fmt.Sprintf("chat-%d", r.next),
		ClientID:   client.ID,
		ClientName: client.Name,
		StaffID:    staff.ID,
		StaffName:  staff.Name,
		UpdatedAt:  now,
	}
	r.conversations[conversation.ID] = conversation

	return user, nil
}

func (r *MemoryRepository) CompanySnapshot(ctx context.Context, viewer User, viewRole Role) (CompanySnapshot, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()

	var snapshot CompanySnapshot
	visibleClients := r.visibleClientIDsLocked(viewer, viewRole)
	for _, client := range r.clients {
		if !visibleClients[client.ID] {
			continue
		}
		if client.Progress >= 100 || strings.EqualFold(client.Status, "Selesai") {
			snapshot.CompletedOrders++
		} else {
			snapshot.ActiveClients++
		}
	}
	for _, order := range r.orders {
		if !visibleClients[order.ClientID] {
			continue
		}
		if order.Status == OrderPaid {
			snapshot.Revenue += order.Paid
		} else {
			snapshot.OpenOrders++
			snapshot.UnpaidInvoices++
		}
	}
	for _, expense := range r.expenses {
		if visibleClients[expense.ClientID] {
			snapshot.Expenses += expense.Amount
		}
	}
	snapshot.Profit = snapshot.Revenue - snapshot.Expenses
	return snapshot, nil
}

func (r *MemoryRepository) ListClients(ctx context.Context, viewer User, viewRole Role) ([]ClientProfile, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()

	visible := r.visibleClientIDsLocked(viewer, viewRole)
	clients := make([]ClientProfile, 0, len(visible))
	for _, client := range r.clients {
		if visible[client.ID] {
			clients = append(clients, client)
		}
	}
	sort.Slice(clients, func(i, j int) bool { return clients[i].CreatedAt.After(clients[j].CreatedAt) })
	return clients, nil
}

func (r *MemoryRepository) ListOrders(ctx context.Context, viewer User, viewRole Role) ([]Order, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()

	visible := r.visibleClientIDsLocked(viewer, viewRole)
	orders := make([]Order, 0, len(r.orders))
	for _, order := range r.orders {
		if visible[order.ClientID] {
			orders = append(orders, order)
		}
	}
	sort.Slice(orders, func(i, j int) bool { return orders[i].CreatedAt.After(orders[j].CreatedAt) })
	return orders, nil
}

func (r *MemoryRepository) ListDocuments(ctx context.Context, viewer User, viewRole Role) ([]Document, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()

	visible := r.visibleClientIDsLocked(viewer, viewRole)
	documents := make([]Document, 0, len(r.documents))
	for _, document := range r.documents {
		if visible[document.ClientID] {
			documents = append(documents, document)
		}
	}
	sort.Slice(documents, func(i, j int) bool { return documents[i].UpdatedAt.After(documents[j].UpdatedAt) })
	return documents, nil
}

func (r *MemoryRepository) ListProgressStages(ctx context.Context, viewer User, viewRole Role) ([]ProgressStage, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()

	visible := r.visibleClientIDsLocked(viewer, viewRole)
	stages := make([]ProgressStage, 0, len(r.stages))
	for _, stage := range r.stages {
		if visible[stage.ClientID] {
			stages = append(stages, stage)
		}
	}
	sort.Slice(stages, func(i, j int) bool { return stages[i].Step < stages[j].Step })
	return stages, nil
}

func (r *MemoryRepository) ListSchedules(ctx context.Context, viewer User, viewRole Role) ([]ScheduleItem, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()

	visible := r.visibleClientIDsLocked(viewer, viewRole)
	schedules := make([]ScheduleItem, 0, len(r.schedules))
	for _, schedule := range r.schedules {
		if visible[schedule.ClientID] {
			schedules = append(schedules, schedule)
		}
	}
	sort.Slice(schedules, func(i, j int) bool { return schedules[i].CreatedAt.Before(schedules[j].CreatedAt) })
	return schedules, nil
}

func (r *MemoryRepository) ListTasks(ctx context.Context, viewer User, viewRole Role) ([]Task, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		if viewer.Role == RoleOwner || task.StaffID == viewer.ID {
			tasks = append(tasks, task)
		}
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].TimeLabel < tasks[j].TimeLabel })
	return tasks, nil
}

func (r *MemoryRepository) ListExpenses(ctx context.Context, viewer User, viewRole Role) ([]Expense, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()

	visible := r.visibleClientIDsLocked(viewer, viewRole)
	expenses := make([]Expense, 0, len(r.expenses))
	for _, expense := range r.expenses {
		if visible[expense.ClientID] && (viewer.Role == RoleOwner || expense.StaffID == viewer.ID) {
			expenses = append(expenses, expense)
		}
	}
	return expenses, nil
}

func (r *MemoryRepository) MarkOrderPaidByCode(ctx context.Context, viewer User, code string) (Order, error) {
	_ = ctx
	if viewer.Role != RoleOwner {
		return Order{}, ErrForbidden
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id, ok := r.orderByCode[strings.ToUpper(strings.TrimSpace(code))]
	if !ok {
		return Order{}, ErrNotFound
	}
	order := r.orders[id]
	now := time.Now()
	order.Paid = order.Total
	order.Status = OrderPaid
	order.PaidAt = &now
	r.orders[id] = order
	return order, nil
}

func (r *MemoryRepository) SubmitPaymentProof(ctx context.Context, viewer User, code, note, fileName, storagePath string) (Order, error) {
	_ = ctx
	if viewer.Role != RoleStudent {
		return Order{}, ErrForbidden
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id, ok := r.orderByCode[strings.ToUpper(strings.TrimSpace(code))]
	if !ok {
		return Order{}, ErrNotFound
	}
	order := r.orders[id]
	client, ok := r.clients[order.ClientID]
	if !ok || client.UserID != viewer.ID {
		return Order{}, ErrForbidden
	}
	if order.Status != OrderPaid {
		order.Status = OrderWaitingVerification
		order.ProofNote = strings.TrimSpace(note)
		order.ProofFileName = strings.TrimSpace(fileName)
		order.ProofStoragePath = strings.TrimSpace(storagePath)
		r.orders[id] = order
	}
	return order, nil
}

func (r *MemoryRepository) StudentHasPaymentAccess(ctx context.Context, viewer User) (bool, error) {
	_ = ctx
	if viewer.Role != RoleStudent {
		return true, nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, client := range r.clients {
		if client.UserID != viewer.ID {
			continue
		}
		for _, order := range r.orders {
			if order.ClientID == client.ID && (order.Status == OrderWaitingVerification || order.Status == OrderPaid) {
				return true, nil
			}
		}
		return false, nil
	}
	return false, nil
}

func (r *MemoryRepository) UploadStudentDocument(ctx context.Context, viewer User, documentName, fileName, storagePath string) (Document, error) {
	_ = ctx
	if viewer.Role != RoleStudent {
		return Document{}, ErrForbidden
	}
	allowed, err := r.StudentHasPaymentAccess(ctx, viewer)
	if err != nil {
		return Document{}, err
	}
	if !allowed {
		return Document{}, ErrForbidden
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	var client ClientProfile
	for _, item := range r.clients {
		if item.UserID == viewer.ID {
			client = item
			break
		}
	}
	if client.ID == "" {
		return Document{}, ErrForbidden
	}
	for _, document := range r.documents {
		if document.ClientID == client.ID && strings.EqualFold(document.Name, documentName) {
			document.Status = DocumentReview
			document.Reviewer = ""
			document.FileName = fileName
			document.StoragePath = storagePath
			document.UpdatedAt = time.Now()
			r.documents[document.ID] = document
			return document, nil
		}
	}
	r.next++
	document := Document{ID: fmt.Sprintf("doc-%d", r.next), ClientID: client.ID, ClientName: client.Name, Name: documentName, Status: DocumentReview, FileName: fileName, StoragePath: storagePath, UpdatedAt: time.Now()}
	r.documents[document.ID] = document
	return document, nil
}

func (r *MemoryRepository) ReviewDocument(ctx context.Context, viewer User, documentID string, status DocumentStatus) (Document, error) {
	_ = ctx
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return Document{}, ErrForbidden
	}
	if status != DocumentApproved && status != DocumentRevision {
		return Document{}, ErrInvalidInput
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	document, ok := r.documents[documentID]
	if !ok {
		return Document{}, ErrNotFound
	}
	if viewer.Role == RoleStaff {
		client, ok := r.clients[document.ClientID]
		if !ok || client.PICStaffID != viewer.ID {
			return Document{}, ErrForbidden
		}
	}
	document.Status = status
	document.Reviewer = viewer.Name
	document.UpdatedAt = time.Now()
	r.documents[document.ID] = document
	return document, nil
}

func (r *MemoryRepository) CompleteTask(ctx context.Context, viewer User, taskID string) (Task, error) {
	_ = ctx
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return Task{}, ErrForbidden
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[taskID]
	if !ok {
		return Task{}, ErrNotFound
	}
	if viewer.Role == RoleStaff && task.StaffID != viewer.ID {
		return Task{}, ErrForbidden
	}
	task.Status = TaskDone
	r.tasks[task.ID] = task
	return task, nil
}

func (r *MemoryRepository) ListConversations(ctx context.Context, viewer User) ([]ChatConversation, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()

	conversations := make([]ChatConversation, 0, len(r.conversations))
	for _, conversation := range r.conversations {
		if r.canAccessConversationLocked(viewer, conversation) {
			conversations = append(conversations, conversation)
		}
	}
	sort.Slice(conversations, func(i, j int) bool { return conversations[i].UpdatedAt.After(conversations[j].UpdatedAt) })
	return conversations, nil
}

func (r *MemoryRepository) ListMessages(ctx context.Context, viewer User, conversationID string) ([]ChatMessage, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()

	conversation, ok := r.conversations[conversationID]
	if !ok {
		return nil, ErrNotFound
	}
	if !r.canAccessConversationLocked(viewer, conversation) {
		return nil, ErrForbidden
	}
	messages := append([]ChatMessage(nil), r.messages[conversationID]...)
	return messages, nil
}

func (r *MemoryRepository) SaveMessage(ctx context.Context, viewer User, conversationID, body string) (ChatMessage, error) {
	_ = ctx
	body = strings.TrimSpace(body)
	if body == "" || len(body) > 2000 {
		return ChatMessage{}, ErrInvalidInput
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	conversation, ok := r.conversations[conversationID]
	if !ok {
		return ChatMessage{}, ErrNotFound
	}
	if !r.canAccessConversationLocked(viewer, conversation) {
		return ChatMessage{}, ErrForbidden
	}

	r.next++
	message := ChatMessage{
		ID:             fmt.Sprintf("msg-%d", r.next),
		ConversationID: conversationID,
		SenderID:       viewer.ID,
		SenderName:     viewer.Name,
		SenderRole:     viewer.Role,
		Body:           body,
		CreatedAt:      time.Now(),
	}
	r.addMessage(message)
	return message, nil
}

func (r *MemoryRepository) visibleClientIDsLocked(viewer User, viewRole Role) map[string]bool {
	visible := make(map[string]bool)
	for _, client := range r.clients {
		switch viewer.Role {
		case RoleOwner:
			visible[client.ID] = true
		case RoleStaff:
			if client.PICStaffID == viewer.ID {
				visible[client.ID] = true
			}
		case RoleStudent:
			if client.UserID == viewer.ID {
				visible[client.ID] = true
			}
		}
	}
	return visible
}

func (r *MemoryRepository) canAccessConversationLocked(viewer User, conversation ChatConversation) bool {
	switch viewer.Role {
	case RoleOwner:
		return true
	case RoleStaff:
		return conversation.StaffID == viewer.ID
	case RoleStudent:
		client, ok := r.clients[conversation.ClientID]
		return ok && client.UserID == viewer.ID
	default:
		return false
	}
}

func usernameKey(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
