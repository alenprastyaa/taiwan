package dashboard

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
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
	if err := repo.ensurePipelineStagesSeeded(ctx); err != nil {
		return nil, err
	}
	if err := repo.ensureServicePackagesSeeded(ctx); err != nil {
		return nil, err
	}
	if err := repo.ensureTextTemplatesSeeded(ctx); err != nil {
		return nil, err
	}
	if err := repo.ensureInstitutionContactsSeeded(ctx); err != nil {
		return nil, err
	}
	if err := repo.ensureExpenseCategoriesSeeded(ctx); err != nil {
		return nil, err
	}
	if err := repo.ensureShipmentCouriersSeeded(ctx); err != nil {
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
	if picID := strings.TrimSpace(input.PICStaffID); picID != "" {
		if picked, err := r.findStaffByID(ctx, picID); err == nil {
			staff = picked
		}
	}

	packageName := strings.TrimSpace(input.PackageName)
	if packageName == "" {
		packageName = "Belum memilih paket"
	}
	country := strings.TrimSpace(input.Country)
	if country == "" {
		country = "Taiwan"
	}
	campus := strings.TrimSpace(input.Campus)
	if campus == "" {
		campus = "Belum dipilih"
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
		Country:      country,
		Campus:       campus,
		PackageName:  packageName,
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
	if _, err := r.txExec(ctx, tx, `INSERT INTO activity_log (id, staff_id, client_id, action_type, description, created_at) VALUES (?, ?, ?, ?, ?, ?)`, newID("activity"), staff.ID, client.ID, "client_created", "Client baru ditambahkan: "+client.Name, now); err != nil {
		return User{}, err
	}
	if input.InvoiceTotal > 0 {
		paid := input.AmountPaid
		if paid < 0 {
			paid = 0
		}
		if paid > input.InvoiceTotal {
			paid = input.InvoiceTotal
		}
		status := OrderUnpaid
		var paidAt *time.Time
		if paid >= input.InvoiceTotal {
			status = OrderPaid
			paidAt = &now
		}
		order := Order{
			ID:          newID("order"),
			Code:        newOrderCode(now),
			ClientID:    client.ID,
			PackageName: client.PackageName,
			Total:       input.InvoiceTotal,
			Paid:        paid,
			Status:      status,
			DueDate:     now.AddDate(0, 0, 7),
			CreatedAt:   now,
			PaidAt:      paidAt,
		}
		if _, err := r.txExec(ctx, tx, `INSERT INTO orders (id, code, client_id, package_name, total, paid, status, due_date, created_at, paid_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, order.ID, order.Code, order.ClientID, order.PackageName, order.Total, order.Paid, order.Status, order.DueDate, order.CreatedAt, order.PaidAt); err != nil {
			return User{}, err
		}
		description := fmt.Sprintf("Invoice %s dibuat: total Rp%d, uang masuk Rp%d", order.Code, order.Total, order.Paid)
		if _, err := r.txExec(ctx, tx, `INSERT INTO activity_log (id, staff_id, client_id, action_type, description, created_at) VALUES (?, ?, ?, ?, ?, ?)`, newID("activity"), staff.ID, client.ID, "invoice_created", description, now); err != nil {
			return User{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return User{}, err
	}
	return user, nil
}

// newOrderCode mints a human-readable, effectively-unique invoice code
// (orders.code has a UNIQUE constraint) without needing a sequence table.
func newOrderCode(now time.Time) string {
	buf := make([]byte, 3)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("ORD-%d-%06d", now.Year(), now.UnixNano()%1000000)
	}
	return fmt.Sprintf("ORD-%d-%s", now.Year(), strings.ToUpper(hex.EncodeToString(buf)))
}

// ResetTestData permanently wipes every client-linked and transactional row
// (clients, orders, documents, tasks, expenses, shipments, chat, intake
// forms, activity log, signed agreements) so an owner can clear out
// demo/seed data before going live with real clients. Owner/staff accounts
// and business configuration (pipeline stages, service packages, text
// templates, institution contacts, expense categories, shipment couriers)
// are deliberately left untouched — those are real setup, not test data.
// Callers (the HTTP handler) are expected to have already re-verified the
// owner's password before invoking this; there is no undo.
func (r *SQLRepository) ResetTestData(ctx context.Context, viewer User) error {
	if viewer.Role != RoleOwner {
		return ErrForbidden
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	tables := []string{
		"chat_messages",
		"chat_conversations",
		"client_agreements",
		"client_intake_forms",
		"shipments",
		"expenses",
		"tasks",
		"documents",
		"orders",
		"schedules",
		"progress_stages",
		"activity_log",
		"clients",
	}
	for _, table := range tables {
		if _, err := r.txExec(ctx, tx, `DELETE FROM `+table); err != nil {
			return fmt.Errorf("clear %s: %w", table, err)
		}
	}
	if _, err := r.txExec(ctx, tx, `DELETE FROM users WHERE role = ?`, RoleStudent); err != nil {
		return fmt.Errorf("clear student users: %w", err)
	}
	return tx.Commit()
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
	rows, err := r.query(ctx, `SELECT c.id, COALESCE(c.user_id, ''), c.name, c.email, c.phone, c.country, c.campus, c.package_name, c.pic_staff_id, COALESCE(s.name, ''), c.status, c.progress, c.last_schedule, c.current_stage, c.created_at, COALESCE(u.active, TRUE) FROM clients c LEFT JOIN users s ON s.id = c.pic_staff_id LEFT JOIN users u ON u.id = c.user_id ORDER BY c.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []ClientProfile
	for rows.Next() {
		var client ClientProfile
		if err := rows.Scan(&client.ID, &client.UserID, &client.Name, &client.Email, &client.Phone, &client.Country, &client.Campus, &client.PackageName, &client.PICStaffID, &client.PICName, &client.Status, &client.Progress, &client.LastSchedule, &client.CurrentStage, &client.CreatedAt, &client.Active); err != nil {
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

	rows, err := r.query(ctx, `SELECT d.id, d.client_id, COALESCE(c.name, ''), d.name, d.status, d.reviewer, COALESCE(d.review_note, ''), COALESCE(d.file_name, ''), COALESCE(d.storage_path, ''), d.updated_at FROM documents d LEFT JOIN clients c ON c.id = d.client_id ORDER BY d.updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var document Document
		var status string
		if err := rows.Scan(&document.ID, &document.ClientID, &document.ClientName, &document.Name, &status, &document.Reviewer, &document.ReviewNote, &document.FileName, &document.StoragePath, &document.UpdatedAt); err != nil {
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

	rows, err := r.query(ctx, `SELECT e.id, e.staff_id, e.client_id, COALESCE(c.name, ''), e.need, e.category, e.amount, e.status, e.date_label, e.description, COALESCE(e.receipt_file_name, ''), COALESCE(e.receipt_storage_path, '') FROM expenses e LEFT JOIN clients c ON c.id = e.client_id ORDER BY e.date_label DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []Expense
	for rows.Next() {
		var expense Expense
		var status string
		if err := rows.Scan(&expense.ID, &expense.StaffID, &expense.ClientID, &expense.ClientName, &expense.Need, &expense.Category, &expense.Amount, &status, &expense.DateLabel, &expense.Description, &expense.ReceiptFileName, &expense.ReceiptStoragePath); err != nil {
			return nil, err
		}
		expense.Status = ExpenseStatus(status)
		if visible[expense.ClientID] && (viewer.Role == RoleOwner || expense.StaffID == viewer.ID) {
			expenses = append(expenses, expense)
		}
	}
	return expenses, rows.Err()
}

func (r *SQLRepository) ListExpenseCategories(ctx context.Context) ([]string, error) {
	rows, err := r.query(ctx, `SELECT name FROM expense_categories ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		categories = append(categories, name)
	}
	return categories, rows.Err()
}

// upsertExpenseCategory remembers a newly typed category so it becomes an
// autocomplete suggestion next time. It's a best-effort side effect: callers
// should not fail expense creation if this errors.
func (r *SQLRepository) upsertExpenseCategory(ctx context.Context, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	var exists int
	if err := r.queryRow(ctx, `SELECT COUNT(*) FROM expense_categories WHERE name = ?`, name).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	_, err := r.exec(ctx, `INSERT INTO expense_categories (id, name, created_at) VALUES (?, ?, ?)`, newID("expcat"), name, time.Now())
	return err
}

// ListShipments returns shipments scoped to what the viewer may see — same
// visibility rule as ListDocuments/ListExpenses: resolve visible clients via
// ListClients (already scoped per role) and filter by ClientID, so owner,
// PIC staff, and the client themselves each see the right slice.
func (r *SQLRepository) ListShipments(ctx context.Context, viewer User, viewRole Role) ([]Shipment, error) {
	clients, err := r.ListClients(ctx, viewer, viewRole)
	if err != nil {
		return nil, err
	}
	visible := make(map[string]bool, len(clients))
	for _, client := range clients {
		visible[client.ID] = true
	}

	rows, err := r.query(ctx, `SELECT s.id, s.client_id, COALESCE(c.name, ''), s.staff_id, s.direction, s.courier, s.tracking_number, s.contents, s.sender_address, s.recipient_address, s.status, s.shipped_date_label, s.received_date_label, s.created_at FROM shipments s LEFT JOIN clients c ON c.id = s.client_id ORDER BY s.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shipments []Shipment
	for rows.Next() {
		var shipment Shipment
		var direction, status string
		if err := rows.Scan(&shipment.ID, &shipment.ClientID, &shipment.ClientName, &shipment.StaffID, &direction, &shipment.Courier, &shipment.TrackingNumber, &shipment.Contents, &shipment.SenderAddress, &shipment.RecipientAddress, &status, &shipment.ShippedDateLabel, &shipment.ReceivedDateLabel, &shipment.CreatedAt); err != nil {
			return nil, err
		}
		shipment.Direction = ShipmentDirection(direction)
		shipment.Status = ShipmentStatus(status)
		if visible[shipment.ClientID] {
			shipments = append(shipments, shipment)
		}
	}
	return shipments, rows.Err()
}

func (r *SQLRepository) ListShipmentCouriers(ctx context.Context) ([]string, error) {
	rows, err := r.query(ctx, `SELECT name FROM shipment_couriers ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var couriers []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		couriers = append(couriers, name)
	}
	return couriers, rows.Err()
}

// upsertShipmentCourier remembers a newly typed courier so it becomes an
// autocomplete suggestion next time — same best-effort shape as
// upsertExpenseCategory.
func (r *SQLRepository) upsertShipmentCourier(ctx context.Context, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	var exists int
	if err := r.queryRow(ctx, `SELECT COUNT(*) FROM shipment_couriers WHERE name = ?`, name).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	_, err := r.exec(ctx, `INSERT INTO shipment_couriers (id, name, created_at) VALUES (?, ?, ?)`, newID("courier"), name, time.Now())
	return err
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
	order, err := r.findOrderByCode(ctx, normalized)
	if err != nil {
		return Order{}, err
	}
	r.logActivity(ctx, viewer.ID, order.ClientID, "order_marked_paid", "Order "+order.Code+" ditandai lunas")
	return order, nil
}

// RecordOrderPayment adds an incoming payment to an existing invoice (e.g. a
// cicilan/installment) instead of marking it fully paid outright. It clamps
// the running total at the invoice total and flips the order to OrderPaid
// once fully covered.
func (r *SQLRepository) RecordOrderPayment(ctx context.Context, viewer User, code string, amount int64) (Order, error) {
	if viewer.Role != RoleOwner {
		return Order{}, ErrForbidden
	}
	if amount <= 0 {
		return Order{}, ErrInvalidInput
	}
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if normalized == "" {
		return Order{}, ErrInvalidInput
	}
	order, err := r.findOrderByCode(ctx, normalized)
	if err != nil {
		return Order{}, err
	}
	if order.Status == OrderPaid {
		return order, nil
	}
	newPaid := order.Paid + amount
	if newPaid >= order.Total {
		newPaid = order.Total
		now := time.Now()
		if _, err := r.exec(ctx, `UPDATE orders SET paid = ?, status = ?, paid_at = ? WHERE id = ?`, newPaid, OrderPaid, now, order.ID); err != nil {
			return Order{}, err
		}
	} else {
		if _, err := r.exec(ctx, `UPDATE orders SET paid = ? WHERE id = ?`, newPaid, order.ID); err != nil {
			return Order{}, err
		}
	}
	order, err = r.findOrderByCode(ctx, normalized)
	if err != nil {
		return Order{}, err
	}
	r.logActivity(ctx, viewer.ID, order.ClientID, "order_payment_recorded", fmt.Sprintf("Pembayaran Rp%d dicatat untuk order %s", amount, order.Code))
	return order, nil
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

// CreateOrder opens a new invoice for an already-existing client — the
// standalone counterpart to the InvoiceTotal/AmountPaid fields baked into
// CreateStudent, for a renewal order or a client whose first order wasn't
// created at signup.
func (r *SQLRepository) CreateOrder(ctx context.Context, viewer User, input CreateOrderInput) (Order, error) {
	if viewer.Role != RoleOwner {
		return Order{}, ErrForbidden
	}
	if input.Total <= 0 {
		return Order{}, ErrInvalidInput
	}
	client, err := r.findClientByID(ctx, strings.TrimSpace(input.ClientID))
	if err != nil {
		return Order{}, err
	}
	paid := input.Paid
	if paid < 0 {
		paid = 0
	}
	if paid > input.Total {
		paid = input.Total
	}
	packageName := strings.TrimSpace(input.PackageName)
	if packageName == "" {
		packageName = client.PackageName
	}
	now := time.Now()
	dueDate := input.DueDate
	if dueDate.IsZero() {
		dueDate = now.AddDate(0, 0, 7)
	}
	status := OrderUnpaid
	var paidAt *time.Time
	if paid >= input.Total {
		status = OrderPaid
		paidAt = &now
	}
	order := Order{
		ID:          newID("order"),
		Code:        newOrderCode(now),
		ClientID:    client.ID,
		PackageName: packageName,
		Total:       input.Total,
		Paid:        paid,
		Status:      status,
		DueDate:     dueDate,
		CreatedAt:   now,
		PaidAt:      paidAt,
	}
	if _, err := r.exec(ctx, `INSERT INTO orders (id, code, client_id, package_name, total, paid, status, due_date, created_at, paid_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, order.ID, order.Code, order.ClientID, order.PackageName, order.Total, order.Paid, order.Status, order.DueDate, order.CreatedAt, order.PaidAt); err != nil {
		return Order{}, err
	}
	description := fmt.Sprintf("Invoice %s dibuat: total Rp%d, uang masuk Rp%d", order.Code, order.Total, order.Paid)
	r.logActivity(ctx, viewer.ID, client.ID, "invoice_created", description)
	return r.findOrderByCode(ctx, order.Code)
}

// UpdateOrder edits an order's package/total/due date. The paid amount is
// deliberately out of scope — RecordOrderPayment/MarkOrderPaidByCode own
// that invariant — but changing the total can flip the derived status back
// to unpaid if it now exceeds what's already been paid in.
func (r *SQLRepository) UpdateOrder(ctx context.Context, viewer User, orderID string, input UpdateOrderInput) (Order, error) {
	if viewer.Role != RoleOwner {
		return Order{}, ErrForbidden
	}
	order, err := r.findOrderByID(ctx, orderID)
	if err != nil {
		return Order{}, err
	}
	if input.Total <= 0 || input.Total < order.Paid {
		return Order{}, ErrInvalidInput
	}
	packageName := strings.TrimSpace(input.PackageName)
	if packageName == "" {
		packageName = order.PackageName
	}
	dueDate := input.DueDate
	if dueDate.IsZero() {
		dueDate = order.DueDate
	}
	status := order.Status
	var paidAt *time.Time = order.PaidAt
	if order.Paid >= input.Total {
		status = OrderPaid
		if paidAt == nil {
			now := time.Now()
			paidAt = &now
		}
	} else if order.Status == OrderPaid {
		status = OrderUnpaid
		paidAt = nil
	}
	if _, err := r.exec(ctx, `UPDATE orders SET package_name = ?, total = ?, due_date = ?, status = ?, paid_at = ? WHERE id = ?`, packageName, input.Total, dueDate, status, paidAt, order.ID); err != nil {
		return Order{}, err
	}
	r.logActivity(ctx, viewer.ID, order.ClientID, "invoice_updated", "Invoice "+order.Code+" diperbarui")
	return r.findOrderByID(ctx, order.ID)
}

// DeleteOrder removes an invoice outright — unlike clients/staff, orders
// aren't referenced by other tables, so a hard delete is safe and matches
// what "hapus order" should actually do.
func (r *SQLRepository) DeleteOrder(ctx context.Context, viewer User, orderID string) error {
	if viewer.Role != RoleOwner {
		return ErrForbidden
	}
	order, err := r.findOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	result, err := r.exec(ctx, `DELETE FROM orders WHERE id = ?`, order.ID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	r.logActivity(ctx, viewer.ID, order.ClientID, "invoice_deleted", "Invoice "+order.Code+" dihapus")
	return nil
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
		// Re-upload resets the document back to review and clears the previous reviewer/note.
		if _, err := r.exec(ctx, `UPDATE documents SET status = ?, reviewer = ?, review_note = ?, file_name = ?, storage_path = ?, updated_at = ? WHERE id = ?`, DocumentReview, "", "", fileName, storagePath, now, existing.ID); err != nil {
			return Document{}, err
		}
		return r.findDocumentByID(ctx, existing.ID)
	}
	if !errors.Is(err, ErrNotFound) {
		return Document{}, err
	}
	id := newID("doc")
	if _, err := r.exec(ctx, `INSERT INTO documents (id, client_id, name, status, reviewer, review_note, file_name, storage_path, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, id, client.ID, documentName, DocumentReview, "", "", fileName, storagePath, now); err != nil {
		return Document{}, err
	}
	return r.findDocumentByID(ctx, id)
}

func (r *SQLRepository) ReviewDocument(ctx context.Context, viewer User, documentID string, status DocumentStatus, note string) (Document, error) {
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
	// Only documents that are actually waiting for review can be approved/rejected.
	if document.Status != DocumentReview {
		return Document{}, ErrInvalidInput
	}
	client, err := r.findClientByID(ctx, document.ClientID)
	if err != nil {
		return Document{}, err
	}
	if viewer.Role == RoleStaff && client.PICStaffID != viewer.ID {
		return Document{}, ErrForbidden
	}
	note = strings.TrimSpace(note)
	if status == DocumentApproved {
		note = ""
	} else if note == "" {
		note = "Dokumen perlu diperbaiki. Silakan unggah ulang sesuai ketentuan."
	}
	now := time.Now()
	if _, err := r.exec(ctx, `UPDATE documents SET status = ?, reviewer = ?, review_note = ?, updated_at = ? WHERE id = ?`, status, viewer.Name, note, now, documentID); err != nil {
		return Document{}, err
	}
	updated, err := r.findDocumentByID(ctx, documentID)
	if err != nil {
		return Document{}, err
	}
	r.logActivity(ctx, viewer.ID, updated.ClientID, "document_reviewed", "Dokumen "+updated.Name+" -> "+string(updated.Status))
	return updated, nil
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
	updated, err := r.findTaskByID(ctx, taskID)
	if err != nil {
		return Task{}, err
	}
	r.logActivity(ctx, viewer.ID, updated.ClientID, "task_completed", "Task selesai: "+updated.Title)
	return updated, nil
}

func (r *SQLRepository) CreateTask(ctx context.Context, viewer User, input CreateTaskInput) (Task, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return Task{}, ErrForbidden
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return Task{}, ErrInvalidInput
	}
	client, err := r.findClientByID(ctx, input.ClientID)
	if err != nil {
		return Task{}, err
	}
	if viewer.Role == RoleStaff && client.PICStaffID != viewer.ID {
		return Task{}, ErrForbidden
	}
	staffID := client.PICStaffID
	if staffID == "" {
		staffID = viewer.ID
	}
	timeLabel := strings.TrimSpace(input.TimeLabel)
	if timeLabel == "" {
		timeLabel = time.Now().Format("15:04")
	}
	priority := strings.TrimSpace(input.Priority)
	if priority == "" {
		priority = "Sedang"
	}
	id := newID("task")
	if _, err := r.exec(ctx, `INSERT INTO tasks (id, staff_id, client_id, time_label, title, priority, status) VALUES (?, ?, ?, ?, ?, ?, ?)`, id, staffID, client.ID, timeLabel, title, priority, TaskOpen); err != nil {
		return Task{}, err
	}
	return r.findTaskByID(ctx, id)
}

func (r *SQLRepository) CreateExpense(ctx context.Context, viewer User, input CreateExpenseInput) (Expense, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return Expense{}, ErrForbidden
	}
	need := strings.TrimSpace(input.Need)
	if need == "" || input.Amount <= 0 {
		return Expense{}, ErrInvalidInput
	}
	client, err := r.findClientByID(ctx, input.ClientID)
	if err != nil {
		return Expense{}, err
	}
	if viewer.Role == RoleStaff && client.PICStaffID != viewer.ID {
		return Expense{}, ErrForbidden
	}
	dateLabel := strings.TrimSpace(input.DateLabel)
	if dateLabel == "" {
		dateLabel = time.Now().Format("02/01/2006")
	}
	category := strings.TrimSpace(input.Category)
	if category == "" {
		category = "Lainnya"
	}
	status := ExpenseWaiting
	if viewer.Role == RoleOwner {
		status = ExpenseRecorded
	}
	staffID := client.PICStaffID
	if staffID == "" {
		staffID = viewer.ID
	}
	id := newID("expense")
	if _, err := r.exec(ctx, `INSERT INTO expenses (id, staff_id, client_id, need, category, amount, status, date_label, description, receipt_file_name, receipt_storage_path) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, id, staffID, client.ID, need, category, input.Amount, status, dateLabel, strings.TrimSpace(input.Description), strings.TrimSpace(input.FileName), strings.TrimSpace(input.StoragePath)); err != nil {
		return Expense{}, err
	}
	_ = r.upsertExpenseCategory(ctx, category)
	expense, err := r.findExpenseByID(ctx, id)
	if err != nil {
		return Expense{}, err
	}
	r.logActivity(ctx, viewer.ID, client.ID, "expense_recorded", need+" - Rp"+strconv.FormatInt(input.Amount, 10))
	return expense, nil
}

func (r *SQLRepository) CreateShipment(ctx context.Context, viewer User, input CreateShipmentInput) (Shipment, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return Shipment{}, ErrForbidden
	}
	client, err := r.findClientByID(ctx, input.ClientID)
	if err != nil {
		return Shipment{}, err
	}
	if viewer.Role == RoleStaff && client.PICStaffID != viewer.ID {
		return Shipment{}, ErrForbidden
	}
	direction := ShipmentDirection(input.Direction)
	if direction != ShipmentOutgoing && direction != ShipmentIncoming {
		return Shipment{}, ErrInvalidInput
	}
	courier := strings.TrimSpace(input.Courier)
	if courier == "" {
		return Shipment{}, ErrInvalidInput
	}
	shippedDateLabel := strings.TrimSpace(input.ShippedDateLabel)
	if shippedDateLabel == "" {
		shippedDateLabel = time.Now().Format("02/01/2006")
	}
	staffID := client.PICStaffID
	if staffID == "" {
		staffID = viewer.ID
	}
	id := newID("shipment")
	now := time.Now()
	if _, err := r.exec(ctx, `INSERT INTO shipments (id, client_id, staff_id, direction, courier, tracking_number, contents, sender_address, recipient_address, status, shipped_date_label, received_date_label, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, client.ID, staffID, direction, courier, strings.TrimSpace(input.TrackingNumber), strings.TrimSpace(input.Contents), strings.TrimSpace(input.SenderAddress), strings.TrimSpace(input.RecipientAddress), ShipmentShipped, shippedDateLabel, "", now); err != nil {
		return Shipment{}, err
	}
	_ = r.upsertShipmentCourier(ctx, courier)
	shipment, err := r.findShipmentByID(ctx, id)
	if err != nil {
		return Shipment{}, err
	}
	r.logActivity(ctx, viewer.ID, client.ID, "shipment_created", "Kirim dokumen via "+courier+" ("+strings.TrimSpace(input.TrackingNumber)+")")
	return shipment, nil
}

func (r *SQLRepository) MarkShipmentReceived(ctx context.Context, viewer User, shipmentID string) (Shipment, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return Shipment{}, ErrForbidden
	}
	shipment, err := r.findShipmentByID(ctx, shipmentID)
	if err != nil {
		return Shipment{}, err
	}
	if viewer.Role == RoleStaff && shipment.StaffID != viewer.ID {
		return Shipment{}, ErrForbidden
	}
	receivedDateLabel := time.Now().Format("02/01/2006")
	if _, err := r.exec(ctx, `UPDATE shipments SET status = ?, received_date_label = ? WHERE id = ?`, ShipmentDelivered, receivedDateLabel, shipmentID); err != nil {
		return Shipment{}, err
	}
	updated, err := r.findShipmentByID(ctx, shipmentID)
	if err != nil {
		return Shipment{}, err
	}
	r.logActivity(ctx, viewer.ID, updated.ClientID, "shipment_received", "Dokumen diterima via "+updated.Courier+" ("+updated.TrackingNumber+")")
	return updated, nil
}

func (r *SQLRepository) findShipmentByID(ctx context.Context, id string) (Shipment, error) {
	var shipment Shipment
	var direction, status string
	if err := r.queryRow(ctx, `SELECT s.id, s.client_id, COALESCE(c.name, ''), s.staff_id, s.direction, s.courier, s.tracking_number, s.contents, s.sender_address, s.recipient_address, s.status, s.shipped_date_label, s.received_date_label, s.created_at FROM shipments s LEFT JOIN clients c ON c.id = s.client_id WHERE s.id = ?`, id).Scan(&shipment.ID, &shipment.ClientID, &shipment.ClientName, &shipment.StaffID, &direction, &shipment.Courier, &shipment.TrackingNumber, &shipment.Contents, &shipment.SenderAddress, &shipment.RecipientAddress, &status, &shipment.ShippedDateLabel, &shipment.ReceivedDateLabel, &shipment.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Shipment{}, ErrNotFound
		}
		return Shipment{}, err
	}
	shipment.Direction = ShipmentDirection(direction)
	shipment.Status = ShipmentStatus(status)
	return shipment, nil
}

func (r *SQLRepository) CreateSchedule(ctx context.Context, viewer User, input CreateScheduleInput) (ScheduleItem, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return ScheduleItem{}, ErrForbidden
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return ScheduleItem{}, ErrInvalidInput
	}
	client, err := r.findClientByID(ctx, input.ClientID)
	if err != nil {
		return ScheduleItem{}, err
	}
	if viewer.Role == RoleStaff && client.PICStaffID != viewer.ID {
		return ScheduleItem{}, ErrForbidden
	}
	id := newID("schedule")
	now := time.Now()
	if _, err := r.exec(ctx, `INSERT INTO schedules (id, client_id, title, date_label, time_label, location, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, id, client.ID, title, strings.TrimSpace(input.DateLabel), strings.TrimSpace(input.TimeLabel), strings.TrimSpace(input.Location), "Terjadwal", now); err != nil {
		return ScheduleItem{}, err
	}
	return r.findScheduleByID(ctx, id)
}

func (r *SQLRepository) BulkApproveDocuments(ctx context.Context, viewer User) (int, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return 0, ErrForbidden
	}
	documents, err := r.ListDocuments(ctx, viewer, viewer.Role)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	count := 0
	for _, document := range documents {
		if document.Status != DocumentReview {
			continue
		}
		if _, err := r.exec(ctx, `UPDATE documents SET status = ?, reviewer = ?, updated_at = ? WHERE id = ?`, DocumentApproved, viewer.Name, now, document.ID); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

var pipelineTonePalette = []string{"mint", "rose", "emerald", "sky", "violet", "lime", "amber", "blue", "teal"}

func (r *SQLRepository) ListPipelineStages(ctx context.Context) ([]PipelineStage, error) {
	rows, err := r.query(ctx, `SELECT id, name, position, tone, created_at FROM pipeline_stages ORDER BY position ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stages []PipelineStage
	for rows.Next() {
		var stage PipelineStage
		if err := rows.Scan(&stage.ID, &stage.Name, &stage.Position, &stage.Tone, &stage.CreatedAt); err != nil {
			return nil, err
		}
		stages = append(stages, stage)
	}
	return stages, rows.Err()
}

func (r *SQLRepository) CreatePipelineStage(ctx context.Context, viewer User, name string) (PipelineStage, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return PipelineStage{}, ErrForbidden
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return PipelineStage{}, ErrInvalidInput
	}
	existing, err := r.ListPipelineStages(ctx)
	if err != nil {
		return PipelineStage{}, err
	}
	for _, stage := range existing {
		if strings.EqualFold(stage.Name, name) {
			return PipelineStage{}, ErrDuplicate
		}
	}
	stage := PipelineStage{
		ID:        newID("pstage"),
		Name:      name,
		Position:  len(existing),
		Tone:      pipelineTonePalette[len(existing)%len(pipelineTonePalette)],
		CreatedAt: time.Now(),
	}
	if _, err := r.exec(ctx, `INSERT INTO pipeline_stages (id, name, position, tone, created_at) VALUES (?, ?, ?, ?, ?)`, stage.ID, stage.Name, stage.Position, stage.Tone, stage.CreatedAt); err != nil {
		return PipelineStage{}, err
	}
	return stage, nil
}

func (r *SQLRepository) RenamePipelineStage(ctx context.Context, viewer User, stageID, name string) (PipelineStage, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return PipelineStage{}, ErrForbidden
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return PipelineStage{}, ErrInvalidInput
	}
	stages, err := r.ListPipelineStages(ctx)
	if err != nil {
		return PipelineStage{}, err
	}
	var target PipelineStage
	found := false
	for _, stage := range stages {
		if stage.ID == stageID {
			target = stage
			found = true
			continue
		}
		if strings.EqualFold(stage.Name, name) {
			return PipelineStage{}, ErrDuplicate
		}
	}
	if !found {
		return PipelineStage{}, ErrNotFound
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return PipelineStage{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := r.txExec(ctx, tx, `UPDATE pipeline_stages SET name = ? WHERE id = ?`, name, stageID); err != nil {
		return PipelineStage{}, err
	}
	// Clients tracked by the old free-text name follow the stage to its new name.
	if _, err := r.txExec(ctx, tx, `UPDATE clients SET current_stage = ? WHERE current_stage = ?`, name, target.Name); err != nil {
		return PipelineStage{}, err
	}
	if err := tx.Commit(); err != nil {
		return PipelineStage{}, err
	}
	target.Name = name
	return target, nil
}

func (r *SQLRepository) DeletePipelineStage(ctx context.Context, viewer User, stageID string) error {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return ErrForbidden
	}
	stages, err := r.ListPipelineStages(ctx)
	if err != nil {
		return err
	}
	var target PipelineStage
	found := false
	for _, stage := range stages {
		if stage.ID == stageID {
			target = stage
			found = true
			break
		}
	}
	if !found {
		return ErrNotFound
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	// Clients sitting on the deleted stage become unassigned rather than
	// silently vanishing from the board; they show up in the "Belum
	// Ditentukan" bucket until someone moves them to a real stage.
	if _, err := r.txExec(ctx, tx, `UPDATE clients SET current_stage = '' WHERE current_stage = ?`, target.Name); err != nil {
		return err
	}
	if _, err := r.txExec(ctx, tx, `DELETE FROM pipeline_stages WHERE id = ?`, stageID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SQLRepository) ReorderPipelineStage(ctx context.Context, viewer User, stageID, direction string) error {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return ErrForbidden
	}
	stages, err := r.ListPipelineStages(ctx)
	if err != nil {
		return err
	}
	index := -1
	for i, stage := range stages {
		if stage.ID == stageID {
			index = i
			break
		}
	}
	if index == -1 {
		return ErrNotFound
	}
	swapWith := -1
	switch direction {
	case "up":
		swapWith = index - 1
	case "down":
		swapWith = index + 1
	default:
		return ErrInvalidInput
	}
	if swapWith < 0 || swapWith >= len(stages) {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := r.txExec(ctx, tx, `UPDATE pipeline_stages SET position = ? WHERE id = ?`, stages[swapWith].Position, stages[index].ID); err != nil {
		return err
	}
	if _, err := r.txExec(ctx, tx, `UPDATE pipeline_stages SET position = ? WHERE id = ?`, stages[index].Position, stages[swapWith].ID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SQLRepository) UpdateClientStage(ctx context.Context, viewer User, clientID, stageName string) (ClientProfile, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return ClientProfile{}, ErrForbidden
	}
	client, err := r.findClientByID(ctx, clientID)
	if err != nil {
		return ClientProfile{}, err
	}
	if viewer.Role == RoleStaff && client.PICStaffID != viewer.ID {
		return ClientProfile{}, ErrForbidden
	}
	stageName = strings.TrimSpace(stageName)
	status := stageName
	progress := client.Progress
	if strings.EqualFold(stageName, "Selesai") {
		progress = 100
	}
	if _, err := r.exec(ctx, `UPDATE clients SET current_stage = ?, status = ?, progress = ? WHERE id = ?`, stageName, status, progress, clientID); err != nil {
		return ClientProfile{}, err
	}
	return r.findClientByID(ctx, clientID)
}

// ResetStudentPassword generates a brand-new password for a client's login
// account and returns it once so owner/staff can hand it to the client
// directly (e.g. via chat/WhatsApp). Nothing is ever persisted in plaintext:
// only the bcrypt hash is written to the users table.
func (r *SQLRepository) ResetStudentPassword(ctx context.Context, viewer User, clientID string) (User, string, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return User{}, "", ErrForbidden
	}
	client, err := r.findClientByID(ctx, clientID)
	if err != nil {
		return User{}, "", err
	}
	if viewer.Role == RoleStaff && client.PICStaffID != viewer.ID {
		return User{}, "", ErrForbidden
	}
	if client.UserID == "" {
		return User{}, "", ErrNotFound
	}
	newPassword, err := security.GenerateRandomPassword()
	if err != nil {
		return User{}, "", err
	}
	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return User{}, "", err
	}
	if _, err := r.exec(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, hash, client.UserID); err != nil {
		return User{}, "", err
	}
	user, err := r.FindUserByID(ctx, client.UserID)
	if err != nil {
		return User{}, "", err
	}
	return user, newPassword, nil
}

// UpdateClient edits a client's profile fields (name/contact/package/PIC).
// Login credentials go through ResetStudentPassword instead, mirroring how
// UpdateStaff leaves the username/password out of scope.
func (r *SQLRepository) UpdateClient(ctx context.Context, viewer User, clientID string, input UpdateClientInput) (ClientProfile, error) {
	if viewer.Role != RoleOwner && viewer.Role != RoleStaff {
		return ClientProfile{}, ErrForbidden
	}
	client, err := r.findClientByID(ctx, clientID)
	if err != nil {
		return ClientProfile{}, err
	}
	if viewer.Role == RoleStaff && client.PICStaffID != viewer.ID {
		return ClientProfile{}, ErrForbidden
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return ClientProfile{}, ErrInvalidInput
	}
	picStaffID := strings.TrimSpace(input.PICStaffID)
	if picStaffID == "" {
		picStaffID = client.PICStaffID
	} else if _, err := r.findStaffByID(ctx, picStaffID); err != nil {
		return ClientProfile{}, ErrInvalidInput
	}
	packageName := strings.TrimSpace(input.PackageName)
	if packageName == "" {
		packageName = "Belum memilih paket"
	}
	country := strings.TrimSpace(input.Country)
	if country == "" {
		country = "Taiwan"
	}
	campus := strings.TrimSpace(input.Campus)
	if campus == "" {
		campus = "Belum dipilih"
	}
	email := strings.TrimSpace(input.Email)
	phone := strings.TrimSpace(input.Phone)
	if _, err := r.exec(ctx, `UPDATE clients SET name = ?, email = ?, phone = ?, country = ?, campus = ?, package_name = ?, pic_staff_id = ? WHERE id = ?`, name, email, phone, country, campus, packageName, picStaffID, client.ID); err != nil {
		return ClientProfile{}, err
	}
	if client.UserID != "" {
		if _, err := r.exec(ctx, `UPDATE users SET name = ?, email = ?, phone = ? WHERE id = ?`, name, email, phone, client.UserID); err != nil {
			return ClientProfile{}, err
		}
	}
	if viewer.Role == RoleOwner && input.InvoiceTotal > 0 {
		if err := r.upsertClientInvoiceTotal(ctx, viewer, client, packageName, input.InvoiceTotal); err != nil {
			return ClientProfile{}, err
		}
	}
	return r.findClientByID(ctx, client.ID)
}

// upsertClientInvoiceTotal backs the Total Tagihan field on the client edit
// form. It only ever touches total — paid/status stay owned by
// RecordOrderPayment/MarkOrderPaidByCode, same invariant UpdateOrder
// enforces. If the client has no order yet, one is opened at this total
// (paid 0), matching what CreateStudent does at signup time.
func (r *SQLRepository) upsertClientInvoiceTotal(ctx context.Context, viewer User, client ClientProfile, packageName string, total int64) error {
	order, err := r.findLatestOrderByClientID(ctx, client.ID)
	if errors.Is(err, ErrNotFound) {
		_, err := r.CreateOrder(ctx, viewer, CreateOrderInput{ClientID: client.ID, PackageName: packageName, Total: total})
		return err
	}
	if err != nil {
		return err
	}
	_, err = r.UpdateOrder(ctx, viewer, order.ID, UpdateOrderInput{PackageName: order.PackageName, Total: total, DueDate: order.DueDate})
	return err
}

// ToggleClientActive flips a client's login access on/off without touching
// their client/order/document history — the same soft-delete pattern as
// ToggleStaffActive, since orders and documents reference the clients row
// and a hard delete would orphan that data. Owner-only: unlike edit/reset-
// password, blocking a client's access is a business decision, not routine
// account admin a PIC staff should be able to do unilaterally.
func (r *SQLRepository) ToggleClientActive(ctx context.Context, viewer User, clientID string) error {
	if viewer.Role != RoleOwner {
		return ErrForbidden
	}
	client, err := r.findClientByID(ctx, clientID)
	if err != nil {
		return err
	}
	if client.UserID == "" {
		return ErrNotFound
	}
	_, err = r.exec(ctx, `UPDATE users SET active = ? WHERE id = ?`, !client.Active, client.UserID)
	return err
}

// DeleteClient permanently removes a client. It's blocked (ErrConflict) if
// the client still has real business records — orders, documents, tasks,
// expenses, shipments, or schedules — since wiping those would destroy
// financial/operational history with no undo; ToggleClientActive is the
// tool for that case. What's safe to cascade away is the incidental data
// every client accumulates just by existing (chat, signed agreement,
// intake form, onboarding progress checklist, activity log) — none of
// that has value once the client itself is gone.
func (r *SQLRepository) DeleteClient(ctx context.Context, viewer User, clientID string) error {
	if viewer.Role != RoleOwner {
		return ErrForbidden
	}
	client, err := r.findClientByID(ctx, clientID)
	if err != nil {
		return err
	}
	blockingTables := []string{"orders", "documents", "tasks", "expenses", "shipments", "schedules"}
	for _, table := range blockingTables {
		var count int
		if err := r.queryRow(ctx, `SELECT COUNT(*) FROM `+table+` WHERE client_id = ?`, client.ID).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			return ErrConflict
		}
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := r.txExec(ctx, tx, `DELETE FROM chat_messages WHERE conversation_id IN (SELECT id FROM chat_conversations WHERE client_id = ?)`, client.ID); err != nil {
		return err
	}
	cascadeTables := []string{"chat_conversations", "client_agreements", "client_intake_forms", "progress_stages", "activity_log"}
	for _, table := range cascadeTables {
		if _, err := r.txExec(ctx, tx, `DELETE FROM `+table+` WHERE client_id = ?`, client.ID); err != nil {
			return fmt.Errorf("clear %s: %w", table, err)
		}
	}
	if _, err := r.txExec(ctx, tx, `DELETE FROM clients WHERE id = ?`, client.ID); err != nil {
		return err
	}
	if client.UserID != "" {
		if _, err := r.txExec(ctx, tx, `DELETE FROM users WHERE id = ?`, client.UserID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ListAllStaff returns every staff account regardless of active status, so
// the owner-facing management page can also see (and re-activate) disabled
// accounts — unlike ListStaff, which only surfaces active staff for PIC
// dropdowns elsewhere.
func (r *SQLRepository) ListAllStaff(ctx context.Context) ([]User, error) {
	rows, err := r.query(ctx, `SELECT id, username, name, email, phone, role, password_hash, active, created_at FROM users WHERE role = ? ORDER BY name ASC`, RoleStaff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var staff []User
	for rows.Next() {
		user, err := r.scanUser(rows)
		if err != nil {
			return nil, err
		}
		staff = append(staff, user)
	}
	return staff, rows.Err()
}

func (r *SQLRepository) findStaffAccountByID(ctx context.Context, id string) (User, error) {
	return r.scanUser(r.queryRow(ctx, `SELECT id, username, name, email, phone, role, password_hash, active, created_at FROM users WHERE id = ? AND role = ?`, id, RoleStaff))
}

func (r *SQLRepository) CreateStaff(ctx context.Context, viewer User, input CreateStaffInput) (User, error) {
	if viewer.Role != RoleOwner {
		return User{}, ErrForbidden
	}
	username := usernameKey(input.Username)
	name := strings.TrimSpace(input.Name)
	if username == "" || len(input.Password) < 8 || name == "" {
		return User{}, ErrInvalidInput
	}
	if _, err := r.scanUser(r.queryRow(ctx, `SELECT id, username, name, email, phone, role, password_hash, active, created_at FROM users WHERE username = ?`, username)); err == nil {
		return User{}, ErrDuplicate
	} else if !errors.Is(err, ErrNotFound) {
		return User{}, err
	}
	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return User{}, err
	}
	user := User{
		ID:           newID("user-staff"),
		Username:     username,
		Name:         name,
		Email:        strings.TrimSpace(input.Email),
		Phone:        strings.TrimSpace(input.Phone),
		Role:         RoleStaff,
		PasswordHash: hash,
		Active:       true,
		CreatedAt:    time.Now(),
	}
	if _, err := r.exec(ctx, `INSERT INTO users (id, username, name, email, phone, role, password_hash, active, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, user.ID, user.Username, user.Name, user.Email, user.Phone, user.Role, user.PasswordHash, user.Active, user.CreatedAt); err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *SQLRepository) UpdateStaff(ctx context.Context, viewer User, staffID string, input UpdateStaffInput) (User, error) {
	if viewer.Role != RoleOwner {
		return User{}, ErrForbidden
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return User{}, ErrInvalidInput
	}
	if _, err := r.findStaffAccountByID(ctx, staffID); err != nil {
		return User{}, err
	}
	if _, err := r.exec(ctx, `UPDATE users SET name = ?, email = ?, phone = ? WHERE id = ? AND role = ?`, name, strings.TrimSpace(input.Email), strings.TrimSpace(input.Phone), staffID, RoleStaff); err != nil {
		return User{}, err
	}
	return r.findStaffAccountByID(ctx, staffID)
}

// ToggleStaffActive flips a staff account between active and disabled —
// used as a soft delete since staff IDs are referenced elsewhere (PIC on
// clients, tasks, expenses, activity log), so a hard delete would orphan
// that history.
func (r *SQLRepository) ToggleStaffActive(ctx context.Context, viewer User, staffID string) error {
	if viewer.Role != RoleOwner {
		return ErrForbidden
	}
	user, err := r.findStaffAccountByID(ctx, staffID)
	if err != nil {
		return err
	}
	_, err = r.exec(ctx, `UPDATE users SET active = ? WHERE id = ?`, !user.Active, user.ID)
	return err
}

// ResetStaffPassword mirrors ResetStudentPassword's one-time-display
// pattern, scoped to staff accounts and owner-only.
func (r *SQLRepository) ResetStaffPassword(ctx context.Context, viewer User, staffID string) (User, string, error) {
	if viewer.Role != RoleOwner {
		return User{}, "", ErrForbidden
	}
	user, err := r.findStaffAccountByID(ctx, staffID)
	if err != nil {
		return User{}, "", err
	}
	newPassword, err := security.GenerateRandomPassword()
	if err != nil {
		return User{}, "", err
	}
	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return User{}, "", err
	}
	if _, err := r.exec(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, hash, user.ID); err != nil {
		return User{}, "", err
	}
	return user, newPassword, nil
}

// UpdateOwnProfile lets any logged-in user (owner, staff, or client) edit
// their own name/email/phone — unlike UpdateStaff, there's no role check
// beyond "must be editing yourself" since viewer.ID is the target.
func (r *SQLRepository) UpdateOwnProfile(ctx context.Context, viewer User, input UpdateOwnProfileInput) (User, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return User{}, ErrInvalidInput
	}
	if _, err := r.exec(ctx, `UPDATE users SET name = ?, email = ?, phone = ? WHERE id = ?`, name, strings.TrimSpace(input.Email), strings.TrimSpace(input.Phone), viewer.ID); err != nil {
		return User{}, err
	}
	return r.FindUserByID(ctx, viewer.ID)
}

// ChangeOwnPassword requires the caller's current password, unlike
// ResetStudentPassword/ResetStaffPassword (owner/staff resetting someone
// else's password without knowing the old one).
func (r *SQLRepository) ChangeOwnPassword(ctx context.Context, viewer User, currentPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrInvalidInput
	}
	if !security.CheckPassword(viewer.PasswordHash, currentPassword) {
		return ErrForbidden
	}
	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return err
	}
	_, err = r.exec(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, hash, viewer.ID)
	return err
}

func (r *SQLRepository) ListServicePackages(ctx context.Context) ([]ServicePackage, error) {
	rows, err := r.query(ctx, `SELECT id, name, category, description, price, price_is_from, highlights, position, created_at FROM service_packages ORDER BY position ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []ServicePackage
	for rows.Next() {
		var pkg ServicePackage
		if err := rows.Scan(&pkg.ID, &pkg.Name, &pkg.Category, &pkg.Description, &pkg.Price, &pkg.PriceIsFrom, &pkg.Highlights, &pkg.Position, &pkg.CreatedAt); err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}
	return packages, rows.Err()
}

func validateServicePackageInput(input ServicePackageInput) error {
	if strings.TrimSpace(input.Name) == "" || input.Price < 0 {
		return ErrInvalidInput
	}
	return nil
}

func (r *SQLRepository) CreateServicePackage(ctx context.Context, viewer User, input ServicePackageInput) (ServicePackage, error) {
	if viewer.Role != RoleOwner {
		return ServicePackage{}, ErrForbidden
	}
	if err := validateServicePackageInput(input); err != nil {
		return ServicePackage{}, err
	}
	existing, err := r.ListServicePackages(ctx)
	if err != nil {
		return ServicePackage{}, err
	}
	for _, pkg := range existing {
		if strings.EqualFold(pkg.Name, input.Name) {
			return ServicePackage{}, ErrDuplicate
		}
	}
	pkg := ServicePackage{
		ID:          newID("svcpkg"),
		Name:        strings.TrimSpace(input.Name),
		Category:    strings.TrimSpace(input.Category),
		Description: strings.TrimSpace(input.Description),
		Price:       input.Price,
		PriceIsFrom: input.PriceIsFrom,
		Highlights:  strings.TrimSpace(input.Highlights),
		Position:    len(existing),
		CreatedAt:   time.Now(),
	}
	if _, err := r.exec(ctx, `INSERT INTO service_packages (id, name, category, description, price, price_is_from, highlights, position, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, pkg.ID, pkg.Name, pkg.Category, pkg.Description, pkg.Price, pkg.PriceIsFrom, pkg.Highlights, pkg.Position, pkg.CreatedAt); err != nil {
		return ServicePackage{}, err
	}
	return pkg, nil
}

func (r *SQLRepository) UpdateServicePackage(ctx context.Context, viewer User, packageID string, input ServicePackageInput) (ServicePackage, error) {
	if viewer.Role != RoleOwner {
		return ServicePackage{}, ErrForbidden
	}
	if err := validateServicePackageInput(input); err != nil {
		return ServicePackage{}, err
	}
	packages, err := r.ListServicePackages(ctx)
	if err != nil {
		return ServicePackage{}, err
	}
	found := false
	for _, pkg := range packages {
		if pkg.ID == packageID {
			found = true
			continue
		}
		if strings.EqualFold(pkg.Name, input.Name) {
			return ServicePackage{}, ErrDuplicate
		}
	}
	if !found {
		return ServicePackage{}, ErrNotFound
	}
	name := strings.TrimSpace(input.Name)
	if _, err := r.exec(ctx, `UPDATE service_packages SET name = ?, category = ?, description = ?, price = ?, price_is_from = ?, highlights = ? WHERE id = ?`, name, strings.TrimSpace(input.Category), strings.TrimSpace(input.Description), input.Price, input.PriceIsFrom, strings.TrimSpace(input.Highlights), packageID); err != nil {
		return ServicePackage{}, err
	}
	var updated ServicePackage
	if err := r.queryRow(ctx, `SELECT id, name, category, description, price, price_is_from, highlights, position, created_at FROM service_packages WHERE id = ?`, packageID).Scan(&updated.ID, &updated.Name, &updated.Category, &updated.Description, &updated.Price, &updated.PriceIsFrom, &updated.Highlights, &updated.Position, &updated.CreatedAt); err != nil {
		return ServicePackage{}, err
	}
	return updated, nil
}

func (r *SQLRepository) DeleteServicePackage(ctx context.Context, viewer User, packageID string) error {
	if viewer.Role != RoleOwner {
		return ErrForbidden
	}
	result, err := r.exec(ctx, `DELETE FROM service_packages WHERE id = ?`, packageID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SQLRepository) ReorderServicePackage(ctx context.Context, viewer User, packageID, direction string) error {
	if viewer.Role != RoleOwner {
		return ErrForbidden
	}
	packages, err := r.ListServicePackages(ctx)
	if err != nil {
		return err
	}
	index := -1
	for i, pkg := range packages {
		if pkg.ID == packageID {
			index = i
			break
		}
	}
	if index == -1 {
		return ErrNotFound
	}
	swapWith := -1
	switch direction {
	case "up":
		swapWith = index - 1
	case "down":
		swapWith = index + 1
	default:
		return ErrInvalidInput
	}
	if swapWith < 0 || swapWith >= len(packages) {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := r.txExec(ctx, tx, `UPDATE service_packages SET position = ? WHERE id = ?`, packages[swapWith].Position, packages[index].ID); err != nil {
		return err
	}
	if _, err := r.txExec(ctx, tx, `UPDATE service_packages SET position = ? WHERE id = ?`, packages[index].Position, packages[swapWith].ID); err != nil {
		return err
	}
	return tx.Commit()
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

func (r *SQLRepository) findOrderByID(ctx context.Context, id string) (Order, error) {
	return scanOrder(r.queryRow(ctx, `SELECT o.id, o.code, o.client_id, COALESCE(c.name, ''), o.package_name, o.total, o.paid, o.status, o.due_date, o.proof_note, o.proof_file_name, o.proof_storage_path, o.created_at, o.paid_at FROM orders o LEFT JOIN clients c ON c.id = o.client_id WHERE o.id = ?`, id))
}

// findLatestOrderByClientID backs the Total Tagihan field on the client edit
// form: a client can have several orders (renewals), so "the client's
// invoice" means the most recently opened one.
func (r *SQLRepository) findLatestOrderByClientID(ctx context.Context, clientID string) (Order, error) {
	return scanOrder(r.queryRow(ctx, `SELECT o.id, o.code, o.client_id, COALESCE(c.name, ''), o.package_name, o.total, o.paid, o.status, o.due_date, o.proof_note, o.proof_file_name, o.proof_storage_path, o.created_at, o.paid_at FROM orders o LEFT JOIN clients c ON c.id = o.client_id WHERE o.client_id = ? ORDER BY o.created_at DESC LIMIT 1`, clientID))
}

func (r *SQLRepository) findClientByID(ctx context.Context, id string) (ClientProfile, error) {
	var client ClientProfile
	if err := r.queryRow(ctx, `SELECT c.id, COALESCE(c.user_id, ''), c.name, c.email, c.phone, c.country, c.campus, c.package_name, c.pic_staff_id, COALESCE(s.name, ''), c.status, c.progress, c.last_schedule, c.current_stage, c.created_at, COALESCE(u.active, TRUE) FROM clients c LEFT JOIN users s ON s.id = c.pic_staff_id LEFT JOIN users u ON u.id = c.user_id WHERE c.id = ?`, id).Scan(&client.ID, &client.UserID, &client.Name, &client.Email, &client.Phone, &client.Country, &client.Campus, &client.PackageName, &client.PICStaffID, &client.PICName, &client.Status, &client.Progress, &client.LastSchedule, &client.CurrentStage, &client.CreatedAt, &client.Active); err != nil {
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
	if err := r.queryRow(ctx, `SELECT d.id, d.client_id, COALESCE(c.name, ''), d.name, d.status, d.reviewer, COALESCE(d.review_note, ''), COALESCE(d.file_name, ''), COALESCE(d.storage_path, ''), d.updated_at FROM documents d LEFT JOIN clients c ON c.id = d.client_id WHERE d.id = ?`, id).Scan(&document.ID, &document.ClientID, &document.ClientName, &document.Name, &status, &document.Reviewer, &document.ReviewNote, &document.FileName, &document.StoragePath, &document.UpdatedAt); err != nil {
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
	if err := r.queryRow(ctx, `SELECT d.id, d.client_id, COALESCE(c.name, ''), d.name, d.status, d.reviewer, COALESCE(d.review_note, ''), COALESCE(d.file_name, ''), COALESCE(d.storage_path, ''), d.updated_at FROM documents d LEFT JOIN clients c ON c.id = d.client_id WHERE d.client_id = ? AND LOWER(d.name) = LOWER(?)`, clientID, name).Scan(&document.ID, &document.ClientID, &document.ClientName, &document.Name, &status, &document.Reviewer, &document.ReviewNote, &document.FileName, &document.StoragePath, &document.UpdatedAt); err != nil {
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

func (r *SQLRepository) findExpenseByID(ctx context.Context, id string) (Expense, error) {
	var expense Expense
	var status string
	if err := r.queryRow(ctx, `SELECT e.id, e.staff_id, e.client_id, COALESCE(c.name, ''), e.need, e.category, e.amount, e.status, e.date_label, e.description, COALESCE(e.receipt_file_name, ''), COALESCE(e.receipt_storage_path, '') FROM expenses e LEFT JOIN clients c ON c.id = e.client_id WHERE e.id = ?`, id).Scan(&expense.ID, &expense.StaffID, &expense.ClientID, &expense.ClientName, &expense.Need, &expense.Category, &expense.Amount, &status, &expense.DateLabel, &expense.Description, &expense.ReceiptFileName, &expense.ReceiptStoragePath); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Expense{}, ErrNotFound
		}
		return Expense{}, err
	}
	expense.Status = ExpenseStatus(status)
	return expense, nil
}

func (r *SQLRepository) findScheduleByID(ctx context.Context, id string) (ScheduleItem, error) {
	var item ScheduleItem
	if err := r.queryRow(ctx, `SELECT id, client_id, title, date_label, time_label, location, status, created_at FROM schedules WHERE id = ?`, id).Scan(&item.ID, &item.ClientID, &item.Title, &item.DateLabel, &item.TimeLabel, &item.Location, &item.Status, &item.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ScheduleItem{}, ErrNotFound
		}
		return ScheduleItem{}, err
	}
	return item, nil
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
