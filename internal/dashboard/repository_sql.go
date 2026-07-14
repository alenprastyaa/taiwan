package dashboard

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"university_agency/internal/security"
)

type sqlDialect string

const (
	dialectPostgres sqlDialect = "postgres"
	dialectMySQL    sqlDialect = "mysql"
)

type SQLRepository struct {
	db      *sql.DB
	dialect sqlDialect
}

type SQLRepositoryOptions struct {
	SeedDemoData bool
}

type rowScanner interface {
	Scan(dest ...any) error
}

func NewSQLRepository(ctx context.Context, db *sql.DB, driver string, opts SQLRepositoryOptions) (*SQLRepository, error) {
	if db == nil {
		return nil, ErrInvalidInput
	}
	dialect, err := parseSQLDialect(driver)
	if err != nil {
		return nil, err
	}
	repo := &SQLRepository{db: db, dialect: dialect}
	if err := repo.migrate(ctx); err != nil {
		return nil, err
	}
	if opts.SeedDemoData {
		if err := repo.seed(ctx); err != nil {
			return nil, err
		}
		if err := repo.seedDemoStudentDetails(ctx); err != nil {
			return nil, err
		}
	}
	if err := repo.ensureUsable(ctx); err != nil {
		return nil, err
	}
	if err := repo.ensureStudentProgressDefaults(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func parseSQLDialect(driver string) (sqlDialect, error) {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "postgres", "postgresql", "pgx":
		return dialectPostgres, nil
	case "mysql", "mariadb":
		return dialectMySQL, nil
	default:
		return "", fmt.Errorf("unsupported database driver %q", driver)
	}
}

func (r *SQLRepository) q(query string) string {
	if r.dialect != dialectPostgres {
		return query
	}
	var b strings.Builder
	index := 1
	for _, ch := range query {
		if ch == '?' {
			b.WriteString(fmt.Sprintf("$%d", index))
			index++
			continue
		}
		b.WriteRune(ch)
	}
	return b.String()
}

func (r *SQLRepository) exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return r.db.ExecContext(ctx, r.q(query), args...)
}

func (r *SQLRepository) query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return r.db.QueryContext(ctx, r.q(query), args...)
}

func (r *SQLRepository) queryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return r.db.QueryRowContext(ctx, r.q(query), args...)
}

func (r *SQLRepository) txExec(ctx context.Context, tx *sql.Tx, query string, args ...any) (sql.Result, error) {
	return tx.ExecContext(ctx, r.q(query), args...)
}

func (r *SQLRepository) txQueryRow(ctx context.Context, tx *sql.Tx, query string, args ...any) *sql.Row {
	return tx.QueryRowContext(ctx, r.q(query), args...)
}

func (r *SQLRepository) FindUserByUsername(ctx context.Context, username string) (User, error) {
	user, err := r.scanUser(r.queryRow(ctx, `SELECT id, username, name, email, phone, role, password_hash, active, created_at FROM users WHERE username = ?`, usernameKey(username)))
	if err != nil {
		return User{}, err
	}
	if !user.Active {
		return User{}, ErrForbidden
	}
	return user, nil
}

func (r *SQLRepository) FindUserByID(ctx context.Context, id string) (User, error) {
	user, err := r.scanUser(r.queryRow(ctx, `SELECT id, username, name, email, phone, role, password_hash, active, created_at FROM users WHERE id = ?`, id))
	if err != nil {
		return User{}, err
	}
	if !user.Active {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (r *SQLRepository) CreateStudent(ctx context.Context, input CreateStudentInput) (User, error) {
	username := usernameKey(input.Username)
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(input.Email)
	phone := strings.TrimSpace(input.Phone)
	if username == "" || len(input.Password) < 8 || name == "" {
		return User{}, ErrInvalidInput
	}

	if _, err := r.FindUserByUsername(ctx, username); err == nil {
		return User{}, ErrDuplicate
	} else if !errors.Is(err, ErrNotFound) {
		return User{}, err
	}

	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return User{}, err
	}
	staff, err := r.firstStaff(ctx)
	if err != nil {
		return User{}, fmt.Errorf("load default staff: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now()
	user := User{
		ID:           newID("user-student"),
		Username:     username,
		Name:         name,
		Email:        email,
		Phone:        phone,
		Role:         RoleStudent,
		PasswordHash: hash,
		Active:       true,
		CreatedAt:    now,
	}
	client := ClientProfile{
		ID:           newID("client"),
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
	conversation := ChatConversation{
		ID:        newID("chat"),
		ClientID:  client.ID,
		StaffID:   staff.ID,
		UpdatedAt: now,
	}

	if _, err := r.txExec(ctx, tx, `INSERT INTO users (id, username, name, email, phone, role, password_hash, active, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, user.ID, user.Username, user.Name, user.Email, user.Phone, user.Role, user.PasswordHash, user.Active, user.CreatedAt); err != nil {
		return User{}, err
	}
	if _, err := r.txExec(ctx, tx, `INSERT INTO clients (id, user_id, name, email, phone, country, campus, package_name, pic_staff_id, status, progress, last_schedule, current_stage, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, client.ID, client.UserID, client.Name, client.Email, client.Phone, client.Country, client.Campus, client.PackageName, client.PICStaffID, client.Status, client.Progress, client.LastSchedule, client.CurrentStage, client.CreatedAt); err != nil {
		return User{}, err
	}
	if _, err := r.txExec(ctx, tx, `INSERT INTO chat_conversations (id, client_id, staff_id, last_message, updated_at) VALUES (?, ?, ?, ?, ?)`, conversation.ID, conversation.ClientID, conversation.StaffID, "", conversation.UpdatedAt); err != nil {
		return User{}, err
	}
	if _, err := r.txExec(ctx, tx, `INSERT INTO progress_stages (id, client_id, step, title, description, status, progress, due_label, pic_name, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, newID("stage"), client.ID, 1, "Registrasi akun", "Akun client sudah dibuat dan menunggu pemilihan paket serta jadwal konsultasi.", ProgressStageActive, client.Progress, "-", staff.Name, now); err != nil {
		return User{}, err
	}
	if err := tx.Commit(); err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *SQLRepository) CompanySnapshot(ctx context.Context, viewer User, viewRole Role) (CompanySnapshot, error) {
	clients, err := r.ListClients(ctx, viewer, viewRole)
	if err != nil {
		return CompanySnapshot{}, err
	}
	orders, err := r.ListOrders(ctx, viewer, viewRole)
	if err != nil {
		return CompanySnapshot{}, err
	}
	expenses, err := r.ListExpenses(ctx, viewer, viewRole)
	if err != nil {
		return CompanySnapshot{}, err
	}

	var snapshot CompanySnapshot
	for _, client := range clients {
		if client.Progress >= 100 || strings.EqualFold(client.Status, "Selesai") {
			snapshot.CompletedOrders++
		} else {
			snapshot.ActiveClients++
		}
	}
	for _, order := range orders {
		if order.Status == OrderPaid {
			snapshot.Revenue += order.Paid
		} else {
			snapshot.OpenOrders++
			snapshot.UnpaidInvoices++
		}
	}
	for _, expense := range expenses {
		snapshot.Expenses += expense.Amount
	}
	snapshot.Profit = snapshot.Revenue - snapshot.Expenses
	return snapshot, nil
}

func (r *SQLRepository) ListClients(ctx context.Context, viewer User, viewRole Role) ([]ClientProfile, error) {
	rows, err := r.query(ctx, `SELECT c.id, COALESCE(c.user_id, ''), c.name, c.email, c.phone, c.country, c.campus, c.package_name, c.pic_staff_id, COALESCE(s.name, ''), c.status, c.progress, c.last_schedule, c.current_stage, c.created_at FROM clients c LEFT JOIN users s ON s.id = c.pic_staff_id ORDER BY c.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []ClientProfile
	for rows.Next() {
		var client ClientProfile
		if err := rows.Scan(&client.ID, &client.UserID, &client.Name, &client.Email, &client.Phone, &client.Country, &client.Campus, &client.PackageName, &client.PICStaffID, &client.PICName, &client.Status, &client.Progress, &client.LastSchedule, &client.CurrentStage, &client.CreatedAt); err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	visible := visibleClientIDs(clients, viewer, viewRole)
	filtered := clients[:0]
	for _, client := range clients {
		if visible[client.ID] {
			filtered = append(filtered, client)
		}
	}
	return filtered, nil
}

func (r *SQLRepository) ListOrders(ctx context.Context, viewer User, viewRole Role) ([]Order, error) {
	clients, err := r.ListClients(ctx, viewer, viewRole)
	if err != nil {
		return nil, err
	}
	visible := make(map[string]bool, len(clients))
	for _, client := range clients {
		visible[client.ID] = true
	}

	rows, err := r.query(ctx, `SELECT o.id, o.code, o.client_id, COALESCE(c.name, ''), o.package_name, o.total, o.paid, o.status, o.due_date, o.proof_note, o.proof_file_name, o.proof_storage_path, o.created_at, o.paid_at FROM orders o LEFT JOIN clients c ON c.id = o.client_id ORDER BY o.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		if visible[order.ClientID] {
			orders = append(orders, order)
		}
	}
	return orders, rows.Err()
}

func (r *SQLRepository) ListDocuments(ctx context.Context, viewer User, viewRole Role) ([]Document, error) {
	clients, err := r.ListClients(ctx, viewer, viewRole)
	if err != nil {
		return nil, err
	}
	visible := make(map[string]bool, len(clients))
	for _, client := range clients {
		visible[client.ID] = true
	}

	rows, err := r.query(ctx, `SELECT d.id, d.client_id, COALESCE(c.name, ''), d.name, d.status, d.reviewer, COALESCE(d.file_name, ''), COALESCE(d.storage_path, ''), d.updated_at FROM documents d LEFT JOIN clients c ON c.id = d.client_id ORDER BY d.updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var document Document
		var status string
		if err := rows.Scan(&document.ID, &document.ClientID, &document.ClientName, &document.Name, &status, &document.Reviewer, &document.FileName, &document.StoragePath, &document.UpdatedAt); err != nil {
			return nil, err
		}
		document.Status = DocumentStatus(status)
		if visible[document.ClientID] {
			documents = append(documents, document)
		}
	}
	return documents, rows.Err()
}

func (r *SQLRepository) ListProgressStages(ctx context.Context, viewer User, viewRole Role) ([]ProgressStage, error) {
	clients, err := r.ListClients(ctx, viewer, viewRole)
	if err != nil {
		return nil, err
	}
	visible := make(map[string]bool, len(clients))
	for _, client := range clients {
		visible[client.ID] = true
	}
	rows, err := r.query(ctx, `SELECT id, client_id, step, title, description, status, progress, due_label, pic_name, updated_at FROM progress_stages ORDER BY step ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stages []ProgressStage
	for rows.Next() {
		var stage ProgressStage
		var status string
		if err := rows.Scan(&stage.ID, &stage.ClientID, &stage.Step, &stage.Title, &stage.Description, &status, &stage.Progress, &stage.DueLabel, &stage.PICName, &stage.UpdatedAt); err != nil {
			return nil, err
		}
		stage.Status = ProgressStageStatus(status)
		if visible[stage.ClientID] {
			stages = append(stages, stage)
		}
	}
	return stages, rows.Err()
}

func (r *SQLRepository) ListSchedules(ctx context.Context, viewer User, viewRole Role) ([]ScheduleItem, error) {
	clients, err := r.ListClients(ctx, viewer, viewRole)
	if err != nil {
		return nil, err
	}
	visible := make(map[string]bool, len(clients))
	for _, client := range clients {
		visible[client.ID] = true
	}
	rows, err := r.query(ctx, `SELECT id, client_id, title, date_label, time_label, location, status, created_at FROM schedules ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []ScheduleItem
	for rows.Next() {
		var item ScheduleItem
		if err := rows.Scan(&item.ID, &item.ClientID, &item.Title, &item.DateLabel, &item.TimeLabel, &item.Location, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		if visible[item.ClientID] {
			schedules = append(schedules, item)
		}
	}
	return schedules, rows.Err()
}

func (r *SQLRepository) ListTasks(ctx context.Context, viewer User, viewRole Role) ([]Task, error) {
	rows, err := r.query(ctx, `SELECT t.id, t.staff_id, t.client_id, COALESCE(c.name, ''), t.time_label, t.title, t.priority, t.status FROM tasks t LEFT JOIN clients c ON c.id = t.client_id ORDER BY t.time_label ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var status string
		if err := rows.Scan(&task.ID, &task.StaffID, &task.ClientID, &task.ClientName, &task.TimeLabel, &task.Title, &task.Priority, &status); err != nil {
			return nil, err
		}
		task.Status = TaskStatus(status)
		if viewer.Role == RoleOwner || task.StaffID == viewer.ID {
			tasks = append(tasks, task)
		}
	}
	return tasks, rows.Err()
}

func (r *SQLRepository) ListExpenses(ctx context.Context, viewer User, viewRole Role) ([]Expense, error) {
	clients, err := r.ListClients(ctx, viewer, viewRole)
	if err != nil {
		return nil, err
	}
	visible := make(map[string]bool, len(clients))
	for _, client := range clients {
		visible[client.ID] = true
	}

	rows, err := r.query(ctx, `SELECT e.id, e.staff_id, e.client_id, COALESCE(c.name, ''), e.need, e.category, e.amount, e.status, e.date_label, e.description FROM expenses e LEFT JOIN clients c ON c.id = e.client_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []Expense
	for rows.Next() {
		var expense Expense
		var status string
		if err := rows.Scan(&expense.ID, &expense.StaffID, &expense.ClientID, &expense.ClientName, &expense.Need, &expense.Category, &expense.Amount, &status, &expense.DateLabel, &expense.Description); err != nil {
			return nil, err
		}
		expense.Status = ExpenseStatus(status)
		if visible[expense.ClientID] && (viewer.Role == RoleOwner || expense.StaffID == viewer.ID) {
			expenses = append(expenses, expense)
		}
	}
	return expenses, rows.Err()
}

func (r *SQLRepository) MarkOrderPaidByCode(ctx context.Context, viewer User, code string) (Order, error) {
	if viewer.Role != RoleOwner {
		return Order{}, ErrForbidden
	}
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if normalized == "" {
		return Order{}, ErrInvalidInput
	}
	now := time.Now()
	result, err := r.exec(ctx, `UPDATE orders SET paid = total, status = ?, paid_at = ? WHERE UPPER(code) = ?`, OrderPaid, now, normalized)
	if err != nil {
		return Order{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return Order{}, ErrNotFound
	}
	return r.findOrderByCode(ctx, normalized)
}

func (r *SQLRepository) SubmitPaymentProof(ctx context.Context, viewer User, code, note, fileName, storagePath string) (Order, error) {
	if viewer.Role != RoleStudent {
		return Order{}, ErrForbidden
	}
	order, err := r.findOrderByCode(ctx, strings.ToUpper(strings.TrimSpace(code)))
	if err != nil {
		return Order{}, err
	}
	client, err := r.findClientByID(ctx, order.ClientID)
	if err != nil {
		return Order{}, err
	}
	if client.UserID != viewer.ID {
		return Order{}, ErrForbidden
	}
	if order.Status == OrderPaid {
		return order, nil
	}
	if _, err := r.exec(ctx, `UPDATE orders SET status = ?, proof_note = ?, proof_file_name = ?, proof_storage_path = ? WHERE id = ?`, OrderWaitingVerification, strings.TrimSpace(note), strings.TrimSpace(fileName), strings.TrimSpace(storagePath), order.ID); err != nil {
		return Order{}, err
	}
	return r.findOrderByCode(ctx, order.Code)
}

func (r *SQLRepository) StudentHasPaymentAccess(ctx context.Context, viewer User) (bool, error) {
	if viewer.Role != RoleStudent {
		return true, nil
	}
	clients, err := r.ListClients(ctx, viewer, RoleStudent)
	if err != nil {
		return false, err
	}
	if len(clients) == 0 {
		return false, nil
	}
	var count int
	if err := r.queryRow(ctx, `SELECT COUNT(*) FROM orders WHERE client_id = ? AND status IN (?, ?)`, clients[0].ID, OrderWaitingVerification, OrderPaid).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *SQLRepository) UploadStudentDocument(ctx context.Context, viewer User, documentName, fileName, storagePath string) (Document, error) {
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
	documentName = strings.TrimSpace(documentName)
	fileName = strings.TrimSpace(fileName)
	storagePath = strings.TrimSpace(storagePath)
	if documentName == "" || fileName == "" || storagePath == "" {
		return Document{}, ErrInvalidInput
	}
	clients, err := r.ListClients(ctx, viewer, RoleStudent)
	if err != nil {
		return Document{}, err
	}
	if len(clients) == 0 {
		return Document{}, ErrForbidden
	}
	client := clients[0]
	now := time.Now()
	existing, err := r.findDocumentByClientAndName(ctx, client.ID, documentName)
	if err == nil {
		if _, err := r.exec(ctx, `UPDATE documents SET status = ?, reviewer = ?, file_name = ?, storage_path = ?, updated_at = ? WHERE id = ?`, DocumentReview, "", fileName, storagePath, now, existing.ID); err != nil {
			return Document{}, err
		}
		return r.findDocumentByID(ctx, existing.ID)
	}
	if !errors.Is(err, ErrNotFound) {
		return Document{}, err
	}
	id := newID("doc")
	if _, err := r.exec(ctx, `INSERT INTO documents (id, client_id, name, status, reviewer, file_name, storage_path, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, id, client.ID, documentName, DocumentReview, "", fileName, storagePath, now); err != nil {
		return Document{}, err
	}
	return r.findDocumentByID(ctx, id)
}

func (r *SQLRepository) ReviewDocument(ctx context.Context, viewer User, documentID string, status DocumentStatus) (Document, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return Document{}, ErrForbidden
	}
	if status != DocumentApproved && status != DocumentRevision {
		return Document{}, ErrInvalidInput
	}
	document, err := r.findDocumentByID(ctx, documentID)
	if err != nil {
		return Document{}, err
	}
	client, err := r.findClientByID(ctx, document.ClientID)
	if err != nil {
		return Document{}, err
	}
	if viewer.Role == RoleStaff && client.PICStaffID != viewer.ID {
		return Document{}, ErrForbidden
	}
	now := time.Now()
	if _, err := r.exec(ctx, `UPDATE documents SET status = ?, reviewer = ?, updated_at = ? WHERE id = ?`, status, viewer.Name, now, documentID); err != nil {
		return Document{}, err
	}
	return r.findDocumentByID(ctx, documentID)
}

func (r *SQLRepository) CompleteTask(ctx context.Context, viewer User, taskID string) (Task, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return Task{}, ErrForbidden
	}
	task, err := r.findTaskByID(ctx, taskID)
	if err != nil {
		return Task{}, err
	}
	if viewer.Role == RoleStaff && task.StaffID != viewer.ID {
		return Task{}, ErrForbidden
	}
	if _, err := r.exec(ctx, `UPDATE tasks SET status = ? WHERE id = ?`, TaskDone, taskID); err != nil {
		return Task{}, err
	}
	return r.findTaskByID(ctx, taskID)
}

func (r *SQLRepository) ListConversations(ctx context.Context, viewer User) ([]ChatConversation, error) {
	clients, err := r.ListClients(ctx, viewer, viewer.Role)
	if err != nil {
		return nil, err
	}
	visible := make(map[string]ClientProfile, len(clients))
	for _, client := range clients {
		visible[client.ID] = client
	}

	rows, err := r.query(ctx, `SELECT cc.id, cc.client_id, COALESCE(c.name, ''), cc.staff_id, COALESCE(s.name, ''), cc.last_message, cc.updated_at FROM chat_conversations cc LEFT JOIN clients c ON c.id = cc.client_id LEFT JOIN users s ON s.id = cc.staff_id ORDER BY cc.updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []ChatConversation
	for rows.Next() {
		var conversation ChatConversation
		if err := rows.Scan(&conversation.ID, &conversation.ClientID, &conversation.ClientName, &conversation.StaffID, &conversation.StaffName, &conversation.LastMessage, &conversation.UpdatedAt); err != nil {
			return nil, err
		}
		if viewer.Role == RoleOwner || (viewer.Role == RoleStaff && conversation.StaffID == viewer.ID) || visible[conversation.ClientID].ID != "" {
			conversations = append(conversations, conversation)
		}
	}
	return conversations, rows.Err()
}

func (r *SQLRepository) ListMessages(ctx context.Context, viewer User, conversationID string) ([]ChatMessage, error) {
	if err := r.ensureConversationAccess(ctx, viewer, conversationID); err != nil {
		return nil, err
	}
	rows, err := r.query(ctx, `SELECT m.id, m.conversation_id, m.sender_id, COALESCE(u.name, ''), COALESCE(u.role, ''), m.body, m.created_at FROM chat_messages m LEFT JOIN users u ON u.id = m.sender_id WHERE m.conversation_id = ? ORDER BY m.created_at ASC`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var message ChatMessage
		var role string
		if err := rows.Scan(&message.ID, &message.ConversationID, &message.SenderID, &message.SenderName, &role, &message.Body, &message.CreatedAt); err != nil {
			return nil, err
		}
		message.SenderRole = Role(role)
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (r *SQLRepository) SaveMessage(ctx context.Context, viewer User, conversationID, body string) (ChatMessage, error) {
	body = strings.TrimSpace(body)
	if body == "" || len(body) > 2000 {
		return ChatMessage{}, ErrInvalidInput
	}
	if err := r.ensureConversationAccess(ctx, viewer, conversationID); err != nil {
		return ChatMessage{}, err
	}
	now := time.Now()
	message := ChatMessage{
		ID:             newID("msg"),
		ConversationID: conversationID,
		SenderID:       viewer.ID,
		SenderName:     viewer.Name,
		SenderRole:     viewer.Role,
		Body:           body,
		CreatedAt:      now,
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ChatMessage{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := r.txExec(ctx, tx, `INSERT INTO chat_messages (id, conversation_id, sender_id, body, created_at) VALUES (?, ?, ?, ?, ?)`, message.ID, message.ConversationID, message.SenderID, message.Body, message.CreatedAt); err != nil {
		return ChatMessage{}, err
	}
	if _, err := r.txExec(ctx, tx, `UPDATE chat_conversations SET last_message = ?, updated_at = ? WHERE id = ?`, message.Body, message.CreatedAt, conversationID); err != nil {
		return ChatMessage{}, err
	}
	if err := tx.Commit(); err != nil {
		return ChatMessage{}, err
	}
	return message, nil
}

func (r *SQLRepository) scanUser(scanner rowScanner) (User, error) {
	var user User
	var role string
	if err := scanner.Scan(&user.ID, &user.Username, &user.Name, &user.Email, &user.Phone, &role, &user.PasswordHash, &user.Active, &user.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	user.Role = Role(role)
	return user, nil
}

func scanOrder(scanner rowScanner) (Order, error) {
	var order Order
	var status string
	var paidAt sql.NullTime
	if err := scanner.Scan(&order.ID, &order.Code, &order.ClientID, &order.ClientName, &order.PackageName, &order.Total, &order.Paid, &status, &order.DueDate, &order.ProofNote, &order.ProofFileName, &order.ProofStoragePath, &order.CreatedAt, &paidAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Order{}, ErrNotFound
		}
		return Order{}, err
	}
	order.Status = OrderStatus(status)
	if paidAt.Valid {
		order.PaidAt = &paidAt.Time
	}
	return order, nil
}

func (r *SQLRepository) findOrderByCode(ctx context.Context, code string) (Order, error) {
	return scanOrder(r.queryRow(ctx, `SELECT o.id, o.code, o.client_id, COALESCE(c.name, ''), o.package_name, o.total, o.paid, o.status, o.due_date, o.proof_note, o.proof_file_name, o.proof_storage_path, o.created_at, o.paid_at FROM orders o LEFT JOIN clients c ON c.id = o.client_id WHERE UPPER(o.code) = ?`, strings.ToUpper(strings.TrimSpace(code))))
}

func (r *SQLRepository) findClientByID(ctx context.Context, id string) (ClientProfile, error) {
	var client ClientProfile
	if err := r.queryRow(ctx, `SELECT c.id, COALESCE(c.user_id, ''), c.name, c.email, c.phone, c.country, c.campus, c.package_name, c.pic_staff_id, COALESCE(s.name, ''), c.status, c.progress, c.last_schedule, c.current_stage, c.created_at FROM clients c LEFT JOIN users s ON s.id = c.pic_staff_id WHERE c.id = ?`, id).Scan(&client.ID, &client.UserID, &client.Name, &client.Email, &client.Phone, &client.Country, &client.Campus, &client.PackageName, &client.PICStaffID, &client.PICName, &client.Status, &client.Progress, &client.LastSchedule, &client.CurrentStage, &client.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ClientProfile{}, ErrNotFound
		}
		return ClientProfile{}, err
	}
	return client, nil
}

func (r *SQLRepository) findDocumentByID(ctx context.Context, id string) (Document, error) {
	var document Document
	var status string
	if err := r.queryRow(ctx, `SELECT d.id, d.client_id, COALESCE(c.name, ''), d.name, d.status, d.reviewer, COALESCE(d.file_name, ''), COALESCE(d.storage_path, ''), d.updated_at FROM documents d LEFT JOIN clients c ON c.id = d.client_id WHERE d.id = ?`, id).Scan(&document.ID, &document.ClientID, &document.ClientName, &document.Name, &status, &document.Reviewer, &document.FileName, &document.StoragePath, &document.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Document{}, ErrNotFound
		}
		return Document{}, err
	}
	document.Status = DocumentStatus(status)
	return document, nil
}

func (r *SQLRepository) findDocumentByClientAndName(ctx context.Context, clientID, name string) (Document, error) {
	var document Document
	var status string
	if err := r.queryRow(ctx, `SELECT d.id, d.client_id, COALESCE(c.name, ''), d.name, d.status, d.reviewer, COALESCE(d.file_name, ''), COALESCE(d.storage_path, ''), d.updated_at FROM documents d LEFT JOIN clients c ON c.id = d.client_id WHERE d.client_id = ? AND LOWER(d.name) = LOWER(?)`, clientID, name).Scan(&document.ID, &document.ClientID, &document.ClientName, &document.Name, &status, &document.Reviewer, &document.FileName, &document.StoragePath, &document.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Document{}, ErrNotFound
		}
		return Document{}, err
	}
	document.Status = DocumentStatus(status)
	return document, nil
}

func (r *SQLRepository) findTaskByID(ctx context.Context, id string) (Task, error) {
	var task Task
	var status string
	if err := r.queryRow(ctx, `SELECT t.id, t.staff_id, t.client_id, COALESCE(c.name, ''), t.time_label, t.title, t.priority, t.status FROM tasks t LEFT JOIN clients c ON c.id = t.client_id WHERE t.id = ?`, id).Scan(&task.ID, &task.StaffID, &task.ClientID, &task.ClientName, &task.TimeLabel, &task.Title, &task.Priority, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, ErrNotFound
		}
		return Task{}, err
	}
	task.Status = TaskStatus(status)
	return task, nil
}

func (r *SQLRepository) ensureConversationAccess(ctx context.Context, viewer User, conversationID string) error {
	var conversation ChatConversation
	if err := r.queryRow(ctx, `SELECT cc.id, cc.client_id, COALESCE(c.name, ''), cc.staff_id, COALESCE(s.name, ''), cc.last_message, cc.updated_at FROM chat_conversations cc LEFT JOIN clients c ON c.id = cc.client_id LEFT JOIN users s ON s.id = cc.staff_id WHERE cc.id = ?`, conversationID).Scan(&conversation.ID, &conversation.ClientID, &conversation.ClientName, &conversation.StaffID, &conversation.StaffName, &conversation.LastMessage, &conversation.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if viewer.Role == RoleOwner || (viewer.Role == RoleStaff && conversation.StaffID == viewer.ID) {
		return nil
	}
	client, err := r.findClientByID(ctx, conversation.ClientID)
	if err != nil {
		return err
	}
	if viewer.Role == RoleStudent && client.UserID == viewer.ID {
		allowed, err := r.StudentHasPaymentAccess(ctx, viewer)
		if err != nil {
			return err
		}
		if !allowed {
			return ErrForbidden
		}
		return nil
	}
	return ErrForbidden
}

func (r *SQLRepository) firstStaff(ctx context.Context) (User, error) {
	return r.scanUser(r.queryRow(ctx, `SELECT id, username, name, email, phone, role, password_hash, active, created_at FROM users WHERE role = ? AND active = ? ORDER BY created_at ASC`, RoleStaff, true))
}

func visibleClientIDs(clients []ClientProfile, viewer User, viewRole Role) map[string]bool {
	visible := make(map[string]bool)
	for _, client := range clients {
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

func newID(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(buf)
}

func sortClientsByCreatedAt(clients []ClientProfile) {
	sort.Slice(clients, func(i, j int) bool {
		return clients[i].CreatedAt.After(clients[j].CreatedAt)
	})
}
