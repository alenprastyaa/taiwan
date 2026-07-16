package dashboard

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"university_agency/internal/security"
)

func (r *SQLRepository) seed(ctx context.Context) error {
	var count int
	if err := r.queryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now()
	owner, err := seedUser("user-owner", "owner", "owner12345", "Amanda Wijaya", "owner@taiwanedu.test", "", RoleOwner, now)
	if err != nil {
		return err
	}
	staff, err := seedUser("user-staff-rina", "staff", "staff12345", "Rina Amalia Putri", "rina@taiwanedu.test", "", RoleStaff, now)
	if err != nil {
		return err
	}
	bayu, err := seedUser("user-staff-bayu", "bayu", "bayu12345", "Bayu Pratama Wijaya", "bayu@taiwanedu.test", "", RoleStaff, now)
	if err != nil {
		return err
	}
	student, err := seedUser("user-student-budi", "student", "student12345", "Budi Santoso", "budi@example.test", "081234567890", RoleStudent, now)
	if err != nil {
		return err
	}

	for _, user := range []User{owner, staff, bayu, student} {
		if _, err := r.txExec(ctx, tx, `INSERT INTO users (id, username, name, email, phone, role, password_hash, active, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, user.ID, user.Username, user.Name, user.Email, user.Phone, user.Role, user.PasswordHash, user.Active, user.CreatedAt); err != nil {
			return fmt.Errorf("seed user %s: %w", user.Username, err)
		}
	}

	clients := []ClientProfile{
		{ID: "client-budi", UserID: student.ID, Name: "Budi Santoso", Email: "budi@example.test", Phone: "081234567890", Country: "Taiwan", Campus: "National Taiwan University", PackageName: "Paket Profesional", PICStaffID: staff.ID, Status: "Persiapan Dokumen", Progress: 45, LastSchedule: "20 Jun 2026", CurrentStage: "Persiapan Dokumen", CreatedAt: now.AddDate(0, -2, 0)},
		{ID: "client-siti", Name: "Siti Aisyah", Email: "siti@example.test", Phone: "081222333444", Country: "Taiwan", Campus: "National Cheng Kung University", PackageName: "Paket Basic", PICStaffID: staff.ID, Status: "Persiapan Dokumen", Progress: 60, LastSchedule: "22 Jun 2026", CurrentStage: "Translate Dokumen", CreatedAt: now.AddDate(0, -1, -6)},
		{ID: "client-dewi", Name: "Dewi Lestari", Email: "dewi@example.test", Phone: "081987654321", Country: "Taiwan", Campus: "NTU", PackageName: "Paket Profesional", PICStaffID: staff.ID, Status: "LOA", Progress: 80, LastSchedule: "18 Jun 2026", CurrentStage: "LOA", CreatedAt: now.AddDate(0, -3, 0)},
		{ID: "client-ricky", Name: "Ricky Pratama", Email: "ricky@example.test", Phone: "081111222333", Country: "Taiwan", Campus: "Taipei Medical University", PackageName: "Layanan Satuan (Visa)", PICStaffID: staff.ID, Status: "Visa", Progress: 70, LastSchedule: "25 Jun 2026", CurrentStage: "Visa", CreatedAt: now.AddDate(0, -1, -12)},
	}
	for _, client := range clients {
		userID := sql.NullString{String: client.UserID, Valid: client.UserID != ""}
		if _, err := r.txExec(ctx, tx, `INSERT INTO clients (id, user_id, name, email, phone, country, campus, package_name, pic_staff_id, status, progress, last_schedule, current_stage, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, client.ID, userID, client.Name, client.Email, client.Phone, client.Country, client.Campus, client.PackageName, client.PICStaffID, client.Status, client.Progress, client.LastSchedule, client.CurrentStage, client.CreatedAt); err != nil {
			return fmt.Errorf("seed client %s: %w", client.ID, err)
		}
	}

	orders := []Order{
		{ID: "order-0102", Code: "ORD-2026-0102", ClientID: "client-budi", PackageName: "Paket Profesional", Total: 7500000, Paid: 0, Status: OrderUnpaid, DueDate: now.AddDate(0, 0, 7), CreatedAt: now.AddDate(0, 0, -5)},
		{ID: "order-0103", Code: "ORD-2026-0103", ClientID: "client-budi", PackageName: "Paket Profesional", Total: 7500000, Paid: 7500000, Status: OrderPaid, DueDate: now.AddDate(0, 0, -5), CreatedAt: now.AddDate(0, -1, 0), PaidAt: ptrTime(now.AddDate(0, 0, -2))},
		{ID: "order-0104", Code: "ORD-2026-0104", ClientID: "client-siti", PackageName: "Paket Basic", Total: 4500000, Paid: 0, Status: OrderWaitingVerification, ProofNote: "Transfer BCA 25 Juni", DueDate: now.AddDate(0, 0, 3), CreatedAt: now.AddDate(0, 0, -3)},
		{ID: "order-0105", Code: "ORD-2026-0105", ClientID: "client-dewi", PackageName: "Paket Profesional", Total: 10000000, Paid: 10000000, Status: OrderPaid, DueDate: now.AddDate(0, 0, -9), CreatedAt: now.AddDate(0, -2, 0), PaidAt: ptrTime(now.AddDate(0, -1, -20))},
	}
	for _, order := range orders {
		var paidAt any
		if order.PaidAt != nil {
			paidAt = *order.PaidAt
		}
		if _, err := r.txExec(ctx, tx, `INSERT INTO orders (id, code, client_id, package_name, total, paid, status, due_date, proof_note, created_at, paid_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, order.ID, order.Code, order.ClientID, order.PackageName, order.Total, order.Paid, order.Status, order.DueDate, order.ProofNote, order.CreatedAt, paidAt); err != nil {
			return fmt.Errorf("seed order %s: %w", order.Code, err)
		}
	}

	documents := []Document{
		{ID: "doc-budi-passport", ClientID: "client-budi", Name: "Passport", Status: DocumentApproved, Reviewer: staff.Name, UpdatedAt: now.AddDate(0, 0, -4)},
		{ID: "doc-budi-ijazah", ClientID: "client-budi", Name: "Ijazah Terakhir", Status: DocumentApproved, Reviewer: staff.Name, UpdatedAt: now.AddDate(0, 0, -5)},
		{ID: "doc-budi-foto", ClientID: "client-budi", Name: "Foto Background Putih 4x6", Status: DocumentReview, Reviewer: staff.Name, UpdatedAt: now},
		{ID: "doc-siti-transkrip", ClientID: "client-siti", Name: "Transkrip Nilai", Status: DocumentReview, Reviewer: staff.Name, UpdatedAt: now.Add(-3 * time.Hour)},
		{ID: "doc-ricky-passport", ClientID: "client-ricky", Name: "Passport", Status: DocumentReview, Reviewer: staff.Name, UpdatedAt: now.Add(-5 * time.Hour)},
	}
	for _, document := range documents {
		if _, err := r.txExec(ctx, tx, `INSERT INTO documents (id, client_id, name, status, reviewer, updated_at) VALUES (?, ?, ?, ?, ?, ?)`, document.ID, document.ClientID, document.Name, document.Status, document.Reviewer, document.UpdatedAt); err != nil {
			return fmt.Errorf("seed document %s: %w", document.ID, err)
		}
	}

	stages := []ProgressStage{
		{ID: "stage-budi-01", ClientID: "client-budi", Step: 1, Title: "Konsultasi Awal", Description: "Kebutuhan studi dan target intake dikunci.", Status: ProgressStageDone, Progress: 100, DueLabel: "Selesai", PICName: staff.Name, UpdatedAt: now.AddDate(0, -2, 0)},
		{ID: "stage-budi-02", ClientID: "client-budi", Step: 2, Title: "Persiapan Dokumen", Description: "Dokumen identitas dan akademik dikumpulkan.", Status: ProgressStageActive, Progress: 60, DueLabel: "28 Mei 2026", PICName: staff.Name, UpdatedAt: now.AddDate(0, 0, -2)},
		{ID: "stage-budi-03", ClientID: "client-budi", Step: 3, Title: "Translate & Legalisir", Description: "Dokumen diterjemahkan dan dilegalisir.", Status: ProgressStagePending, Progress: 0, DueLabel: "-", PICName: staff.Name, UpdatedAt: now.AddDate(0, 0, -2)},
		{ID: "stage-budi-04", ClientID: "client-budi", Step: 4, Title: "Apply Kampus", Description: "Berkas dikirim ke kampus tujuan.", Status: ProgressStagePending, Progress: 0, DueLabel: "-", PICName: staff.Name, UpdatedAt: now.AddDate(0, 0, -2)},
		{ID: "stage-budi-05", ClientID: "client-budi", Step: 5, Title: "LOA", Description: "Menunggu hasil admission dan LOA.", Status: ProgressStagePending, Progress: 0, DueLabel: "-", PICName: staff.Name, UpdatedAt: now.AddDate(0, 0, -2)},
		{ID: "stage-budi-06", ClientID: "client-budi", Step: 6, Title: "Visa & Keberangkatan", Description: "Visa, tiket, dan final briefing.", Status: ProgressStagePending, Progress: 0, DueLabel: "-", PICName: staff.Name, UpdatedAt: now.AddDate(0, 0, -2)},
	}
	for _, stage := range stages {
		if _, err := r.txExec(ctx, tx, `INSERT INTO progress_stages (id, client_id, step, title, description, status, progress, due_label, pic_name, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, stage.ID, stage.ClientID, stage.Step, stage.Title, stage.Description, stage.Status, stage.Progress, stage.DueLabel, stage.PICName, stage.UpdatedAt); err != nil {
			return fmt.Errorf("seed progress stage %s: %w", stage.ID, err)
		}
	}

	schedules := []ScheduleItem{
		{ID: "schedule-budi-01", ClientID: "client-budi", Title: "Review Form Data Diri", DateLabel: "26 Jun 2026", TimeLabel: "10:00", Location: "Online Meeting", Status: "Terjadwal", CreatedAt: now.AddDate(0, 0, 1)},
		{ID: "schedule-budi-02", ClientID: "client-budi", Title: "Deadline Upload Surat Rekomendasi", DateLabel: "28 Jun 2026", TimeLabel: "23:59", Location: "Portal Client", Status: "Menunggu Dokumen", CreatedAt: now.AddDate(0, 0, 3)},
		{ID: "schedule-budi-03", ClientID: "client-budi", Title: "Konsultasi Pilihan Kampus", DateLabel: "2 Jul 2026", TimeLabel: "14:00", Location: "Google Meet", Status: "Terjadwal", CreatedAt: now.AddDate(0, 0, 7)},
	}
	for _, schedule := range schedules {
		if _, err := r.txExec(ctx, tx, `INSERT INTO schedules (id, client_id, title, date_label, time_label, location, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, schedule.ID, schedule.ClientID, schedule.Title, schedule.DateLabel, schedule.TimeLabel, schedule.Location, schedule.Status, schedule.CreatedAt); err != nil {
			return fmt.Errorf("seed schedule %s: %w", schedule.ID, err)
		}
	}

	tasks := []Task{
		{ID: "task-followup", StaffID: staff.ID, ClientID: "client-budi", TimeLabel: "09:00", Title: "Follow up dokumen", Priority: "Tinggi", Status: TaskOpen},
		{ID: "task-translate", StaffID: staff.ID, ClientID: "client-siti", TimeLabel: "10:00", Title: "Translate Transkrip Nilai", Priority: "Sedang", Status: TaskOpen},
		{ID: "task-appointment", StaffID: staff.ID, ClientID: "client-dewi", TimeLabel: "11:00", Title: "Appointment Kedutaan", Priority: "Tinggi", Status: TaskOpen},
		{ID: "task-passport", StaffID: staff.ID, ClientID: "client-ricky", TimeLabel: "14:00", Title: "Upload Passport", Priority: "Rendah", Status: TaskDone},
	}
	for _, task := range tasks {
		if _, err := r.txExec(ctx, tx, `INSERT INTO tasks (id, staff_id, client_id, time_label, title, priority, status) VALUES (?, ?, ?, ?, ?, ?, ?)`, task.ID, task.StaffID, task.ClientID, task.TimeLabel, task.Title, task.Priority, task.Status); err != nil {
			return fmt.Errorf("seed task %s: %w", task.ID, err)
		}
	}

	expenses := []Expense{
		{ID: "expense-1", StaffID: staff.ID, ClientID: "client-budi", Need: "Legalisir Kemenkumham", Category: "Legalisir", Amount: 450000, Status: ExpenseWaiting, DateLabel: "25 Jun 2026"},
		{ID: "expense-2", StaffID: staff.ID, ClientID: "client-siti", Need: "Translate Transkrip Nilai", Category: "Translate", Amount: 1200000, Status: ExpenseWaiting, DateLabel: "24 Jun 2026"},
		{ID: "expense-3", StaffID: staff.ID, ClientID: "client-ricky", Need: "Pengiriman Dokumen DHL", Category: "Lainnya", Amount: 300000, Status: ExpenseRecorded, DateLabel: "22 Jun 2026"},
	}
	for _, expense := range expenses {
		if _, err := r.txExec(ctx, tx, `INSERT INTO expenses (id, staff_id, client_id, need, category, amount, status, date_label, description) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, expense.ID, expense.StaffID, expense.ClientID, expense.Need, expense.Category, expense.Amount, expense.Status, expense.DateLabel, expense.Description); err != nil {
			return fmt.Errorf("seed expense %s: %w", expense.ID, err)
		}
	}

	conversation := ChatConversation{ID: "chat-budi", ClientID: "client-budi", StaffID: staff.ID, LastMessage: "Form data diri masih perlu dilengkapi.", UpdatedAt: now}
	if _, err := r.txExec(ctx, tx, `INSERT INTO chat_conversations (id, client_id, staff_id, last_message, updated_at) VALUES (?, ?, ?, ?, ?)`, conversation.ID, conversation.ClientID, conversation.StaffID, conversation.LastMessage, conversation.UpdatedAt); err != nil {
		return fmt.Errorf("seed conversation: %w", err)
	}
	messages := []ChatMessage{
		{ID: "msg-1", ConversationID: "chat-budi", SenderID: staff.ID, Body: "Halo, dokumen ijazah sudah masuk. Saya cek dulu ya.", CreatedAt: now.Add(-2 * time.Hour)},
		{ID: "msg-2", ConversationID: "chat-budi", SenderID: student.ID, Body: "Baik kak, terima kasih.", CreatedAt: now.Add(-90 * time.Minute)},
		{ID: "msg-3", ConversationID: "chat-budi", SenderID: staff.ID, Body: "Form data diri masih perlu dilengkapi bagian alamat Taiwan.", CreatedAt: now.Add(-35 * time.Minute)},
	}
	for _, message := range messages {
		if _, err := r.txExec(ctx, tx, `INSERT INTO chat_messages (id, conversation_id, sender_id, body, created_at) VALUES (?, ?, ?, ?, ?)`, message.ID, message.ConversationID, message.SenderID, message.Body, message.CreatedAt); err != nil {
			return fmt.Errorf("seed message %s: %w", message.ID, err)
		}
	}

	return tx.Commit()
}

func (r *SQLRepository) ensureUsable(ctx context.Context) error {
	var count int
	if err := r.queryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("database has no users; run an explicit bootstrap before starting without demo seed")
	}
	return nil
}

func (r *SQLRepository) seedDemoStudentDetails(ctx context.Context) error {
	var stageCount int
	if err := r.queryRow(ctx, `SELECT COUNT(*) FROM progress_stages WHERE client_id = ?`, "client-budi").Scan(&stageCount); err != nil {
		return err
	}
	var scheduleCount int
	if err := r.queryRow(ctx, `SELECT COUNT(*) FROM schedules WHERE client_id = ?`, "client-budi").Scan(&scheduleCount); err != nil {
		return err
	}
	if stageCount > 0 && scheduleCount > 0 {
		return nil
	}

	staff, err := r.firstStaff(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	if stageCount == 0 {
		stages := []ProgressStage{
			{ID: newID("stage"), ClientID: "client-budi", Step: 1, Title: "Konsultasi Awal", Description: "Kebutuhan studi dan target intake dikunci.", Status: ProgressStageDone, Progress: 100, DueLabel: "Selesai", PICName: staff.Name, UpdatedAt: now.AddDate(0, -2, 0)},
			{ID: newID("stage"), ClientID: "client-budi", Step: 2, Title: "Persiapan Dokumen", Description: "Dokumen identitas dan akademik dikumpulkan.", Status: ProgressStageActive, Progress: 60, DueLabel: "28 Mei 2026", PICName: staff.Name, UpdatedAt: now.AddDate(0, 0, -2)},
			{ID: newID("stage"), ClientID: "client-budi", Step: 3, Title: "Translate & Legalisir", Description: "Dokumen diterjemahkan dan dilegalisir.", Status: ProgressStagePending, Progress: 0, DueLabel: "-", PICName: staff.Name, UpdatedAt: now.AddDate(0, 0, -2)},
			{ID: newID("stage"), ClientID: "client-budi", Step: 4, Title: "Apply Kampus", Description: "Berkas dikirim ke kampus tujuan.", Status: ProgressStagePending, Progress: 0, DueLabel: "-", PICName: staff.Name, UpdatedAt: now.AddDate(0, 0, -2)},
			{ID: newID("stage"), ClientID: "client-budi", Step: 5, Title: "LOA", Description: "Menunggu hasil admission dan LOA.", Status: ProgressStagePending, Progress: 0, DueLabel: "-", PICName: staff.Name, UpdatedAt: now.AddDate(0, 0, -2)},
			{ID: newID("stage"), ClientID: "client-budi", Step: 6, Title: "Visa & Keberangkatan", Description: "Visa, tiket, dan final briefing.", Status: ProgressStagePending, Progress: 0, DueLabel: "-", PICName: staff.Name, UpdatedAt: now.AddDate(0, 0, -2)},
		}
		for _, stage := range stages {
			if _, err := r.exec(ctx, `INSERT INTO progress_stages (id, client_id, step, title, description, status, progress, due_label, pic_name, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, stage.ID, stage.ClientID, stage.Step, stage.Title, stage.Description, stage.Status, stage.Progress, stage.DueLabel, stage.PICName, stage.UpdatedAt); err != nil {
				return err
			}
		}
	}
	if scheduleCount == 0 {
		schedules := []ScheduleItem{
			{ID: newID("schedule"), ClientID: "client-budi", Title: "Review Form Data Diri", DateLabel: "26 Jun 2026", TimeLabel: "10:00", Location: "Online Meeting", Status: "Terjadwal", CreatedAt: now.AddDate(0, 0, 1)},
			{ID: newID("schedule"), ClientID: "client-budi", Title: "Deadline Upload Surat Rekomendasi", DateLabel: "28 Jun 2026", TimeLabel: "23:59", Location: "Portal Client", Status: "Menunggu Dokumen", CreatedAt: now.AddDate(0, 0, 3)},
			{ID: newID("schedule"), ClientID: "client-budi", Title: "Konsultasi Pilihan Kampus", DateLabel: "2 Jul 2026", TimeLabel: "14:00", Location: "Google Meet", Status: "Terjadwal", CreatedAt: now.AddDate(0, 0, 7)},
		}
		for _, schedule := range schedules {
			if _, err := r.exec(ctx, `INSERT INTO schedules (id, client_id, title, date_label, time_label, location, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, schedule.ID, schedule.ClientID, schedule.Title, schedule.DateLabel, schedule.TimeLabel, schedule.Location, schedule.Status, schedule.CreatedAt); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *SQLRepository) ensureStudentProgressDefaults(ctx context.Context) error {
	rows, err := r.query(ctx, `SELECT c.id, c.progress, c.current_stage, COALESCE(s.name, ''), COUNT(ps.id) FROM clients c LEFT JOIN users s ON s.id = c.pic_staff_id LEFT JOIN progress_stages ps ON ps.client_id = c.id WHERE COALESCE(c.user_id, '') <> '' GROUP BY c.id, c.progress, c.current_stage, s.name`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type missingClient struct {
		id           string
		currentStage string
		picName      string
	}
	var missing []missingClient
	for rows.Next() {
		var item missingClient
		var count int
		var progress int
		if err := rows.Scan(&item.id, &progress, &item.currentStage, &item.picName, &count); err != nil {
			return err
		}
		if count == 0 {
			missing = append(missing, item)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	now := time.Now()
	for _, item := range missing {
		title := strings.TrimSpace(item.currentStage)
		if title == "" {
			title = "Registrasi akun"
		}
		if item.picName == "" {
			item.picName = "-"
		}
		if _, err := r.exec(ctx, `INSERT INTO progress_stages (id, client_id, step, title, description, status, progress, due_label, pic_name, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, newID("stage"), item.id, 1, title, "Tahap awal client sudah tercatat di sistem.", ProgressStageActive, 0, "-", item.picName, now); err != nil {
			return err
		}
	}
	return nil
}

// defaultPipelineStages seeds the pipeline board with the stages this agency
// used before stage management became configurable, so upgrading doesn't
// change anyone's board. Owners/staff can rename, reorder, add, or delete
// freely afterwards.
func defaultPipelineStages() []string {
	return []string{"Konsultasi", "Persiapan Dokumen", "Apply / Proses", "LOA", "Visa", "Keberangkatan", "Selesai"}
}

func defaultPipelineTones() []string {
	return []string{"mint", "rose", "emerald", "sky", "violet", "lime", "amber"}
}

func (r *SQLRepository) ensurePipelineStagesSeeded(ctx context.Context) error {
	var count int
	if err := r.queryRow(ctx, `SELECT COUNT(*) FROM pipeline_stages`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	now := time.Now()
	names := defaultPipelineStages()
	tones := defaultPipelineTones()
	for i, name := range names {
		if _, err := r.exec(ctx, `INSERT INTO pipeline_stages (id, name, position, tone, created_at) VALUES (?, ?, ?, ?, ?)`, newID("pstage"), name, i, tones[i%len(tones)], now); err != nil {
			return err
		}
	}
	return nil
}

// defaultServicePackages seeds the price list this agency used before it
// became owner-editable, so upgrading doesn't change what's on the page.
func defaultServicePackages() []ServicePackage {
	return []ServicePackage{
		{
			Name:        "Paket Profesional",
			Category:    "Paket",
			Description: "Full cover untuk universitas, dokumen, visa, dan keberangkatan.",
			Price:       25000000,
			Highlights:  "Checklist lengkap",
		},
		{
			Name:        "Paket Basic",
			Category:    "Paket",
			Description: "Pendampingan inti untuk dokumen dan aplikasi kampus.",
			Price:       15000000,
			Highlights:  "Dokumen prioritas",
		},
		{
			Name:        "Layanan Satuan",
			Category:    "Satuan",
			Description: "Visa, legalisir, translate, TOEFL, medical, dan apply terpisah.",
			Price:       450000,
			PriceIsFrom: true,
			Highlights:  "Perlu tracking bukti\nApproval fleksibel",
		},
	}
}

func (r *SQLRepository) ensureServicePackagesSeeded(ctx context.Context) error {
	var count int
	if err := r.queryRow(ctx, `SELECT COUNT(*) FROM service_packages`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	now := time.Now()
	for i, pkg := range defaultServicePackages() {
		if _, err := r.exec(ctx, `INSERT INTO service_packages (id, name, category, description, price, price_is_from, highlights, position, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, newID("svcpkg"), pkg.Name, pkg.Category, pkg.Description, pkg.Price, pkg.PriceIsFrom, pkg.Highlights, i, now); err != nil {
			return err
		}
	}
	return nil
}

func seedUser(id, username, password, name, email, phone string, role Role, now time.Time) (User, error) {
	hash, err := security.HashPassword(password)
	if err != nil {
		return User{}, err
	}
	return User{
		ID:           id,
		Username:     usernameKey(username),
		Name:         name,
		Email:        email,
		Phone:        phone,
		Role:         role,
		PasswordHash: hash,
		Active:       true,
		CreatedAt:    now,
	}, nil
}
