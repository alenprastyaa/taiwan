package web

import (
	"context"
	"log/slog"

	"university_agency/internal/auth"
	"university_agency/internal/config"
	"university_agency/internal/dashboard"
)

type Dependencies struct {
	Config    config.Config
	Logger    *slog.Logger
	Dashboard dashboard.Service
	Store     Store
	Sessions  *auth.SessionStore
}

type Store interface {
	dashboard.Repository
	FindUserByUsername(ctx context.Context, username string) (dashboard.User, error)
	FindUserByID(ctx context.Context, id string) (dashboard.User, error)
	CreateStudent(ctx context.Context, input dashboard.CreateStudentInput) (dashboard.User, error)
	MarkOrderPaidByCode(ctx context.Context, viewer dashboard.User, code string) (dashboard.Order, error)
	SubmitPaymentProof(ctx context.Context, viewer dashboard.User, code, note, fileName, storagePath string) (dashboard.Order, error)
	StudentHasPaymentAccess(ctx context.Context, viewer dashboard.User) (bool, error)
	UploadStudentDocument(ctx context.Context, viewer dashboard.User, documentName, fileName, storagePath string) (dashboard.Document, error)
	ReviewDocument(ctx context.Context, viewer dashboard.User, documentID string, status dashboard.DocumentStatus, note string) (dashboard.Document, error)
	CompleteTask(ctx context.Context, viewer dashboard.User, taskID string) (dashboard.Task, error)
	SaveMessage(ctx context.Context, viewer dashboard.User, conversationID, body string) (dashboard.ChatMessage, error)
	CreateTask(ctx context.Context, viewer dashboard.User, input dashboard.CreateTaskInput) (dashboard.Task, error)
	CreateExpense(ctx context.Context, viewer dashboard.User, input dashboard.CreateExpenseInput) (dashboard.Expense, error)
	CreateSchedule(ctx context.Context, viewer dashboard.User, input dashboard.CreateScheduleInput) (dashboard.ScheduleItem, error)
	BulkApproveDocuments(ctx context.Context, viewer dashboard.User) (int, error)
	CreatePipelineStage(ctx context.Context, viewer dashboard.User, name string) (dashboard.PipelineStage, error)
	RenamePipelineStage(ctx context.Context, viewer dashboard.User, stageID, name string) (dashboard.PipelineStage, error)
	DeletePipelineStage(ctx context.Context, viewer dashboard.User, stageID string) error
	ReorderPipelineStage(ctx context.Context, viewer dashboard.User, stageID, direction string) error
	UpdateClientStage(ctx context.Context, viewer dashboard.User, clientID, stageName string) (dashboard.ClientProfile, error)
	CreateServicePackage(ctx context.Context, viewer dashboard.User, input dashboard.ServicePackageInput) (dashboard.ServicePackage, error)
	UpdateServicePackage(ctx context.Context, viewer dashboard.User, packageID string, input dashboard.ServicePackageInput) (dashboard.ServicePackage, error)
	DeleteServicePackage(ctx context.Context, viewer dashboard.User, packageID string) error
	ReorderServicePackage(ctx context.Context, viewer dashboard.User, packageID, direction string) error
	CreateTextTemplate(ctx context.Context, viewer dashboard.User, input dashboard.TextTemplateInput) (dashboard.TextTemplate, error)
	UpdateTextTemplate(ctx context.Context, viewer dashboard.User, templateID string, input dashboard.TextTemplateInput) (dashboard.TextTemplate, error)
	DeleteTextTemplate(ctx context.Context, viewer dashboard.User, templateID string) error
	ReorderTextTemplate(ctx context.Context, viewer dashboard.User, templateID, direction string) error
	CreateInstitutionContact(ctx context.Context, viewer dashboard.User, input dashboard.InstitutionContactInput) (dashboard.InstitutionContact, error)
	UpdateInstitutionContact(ctx context.Context, viewer dashboard.User, contactID string, input dashboard.InstitutionContactInput) (dashboard.InstitutionContact, error)
	DeleteInstitutionContact(ctx context.Context, viewer dashboard.User, contactID string) error
	ReorderInstitutionContact(ctx context.Context, viewer dashboard.User, contactID, direction string) error
	SaveClientIntakeForm(ctx context.Context, viewer dashboard.User, input dashboard.ClientIntakeFormInput) (dashboard.ClientIntakeForm, error)
	CreateActivityNote(ctx context.Context, viewer dashboard.User, clientID, note string) (dashboard.ActivityLog, error)
	HasSignedAgreement(ctx context.Context, viewer dashboard.User) (bool, error)
	SignAgreement(ctx context.Context, viewer dashboard.User, fullNameTyped, ipAddress, userAgent string) (dashboard.ClientAgreement, error)
}

type Handler struct {
	cfg       config.Config
	logger    *slog.Logger
	dashboard dashboard.Service
	store     Store
	sessions  *auth.SessionStore
	chatHub   *ChatHub
}
