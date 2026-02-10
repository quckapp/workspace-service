package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type BillingRepository struct {
	db *sqlx.DB
}

func NewBillingRepository(db *sqlx.DB) *BillingRepository {
	return &BillingRepository{db: db}
}

// Plan methods
func (r *BillingRepository) GetPlan(ctx context.Context, workspaceID uuid.UUID) (*models.WorkspacePlan, error) {
	var plan models.WorkspacePlan
	err := r.db.GetContext(ctx, &plan, "SELECT * FROM workspace_plans WHERE workspace_id = ?", workspaceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &plan, err
}

func (r *BillingRepository) CreatePlan(ctx context.Context, plan *models.WorkspacePlan) error {
	query := `INSERT INTO workspace_plans (id, workspace_id, plan_type, status, billing_cycle, seat_count, seat_limit, storage_limit_mb, storage_used_mb, price_per_seat, currency, trial_ends_at, current_period_start, current_period_end, canceled_at, external_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, plan.ID, plan.WorkspaceID, plan.PlanType, plan.Status, plan.BillingCycle, plan.SeatCount, plan.SeatLimit, plan.StorageLimitMB, plan.StorageUsedMB, plan.PricePerSeat, plan.Currency, plan.TrialEndsAt, plan.CurrentPeriodStart, plan.CurrentPeriodEnd, plan.CanceledAt, plan.ExternalID, plan.CreatedAt, plan.UpdatedAt)
	return err
}

func (r *BillingRepository) UpdatePlan(ctx context.Context, plan *models.WorkspacePlan) error {
	query := `UPDATE workspace_plans SET plan_type = ?, status = ?, billing_cycle = ?, seat_count = ?, seat_limit = ?, storage_limit_mb = ?, storage_used_mb = ?, price_per_seat = ?, current_period_start = ?, current_period_end = ?, canceled_at = ?, updated_at = ? WHERE workspace_id = ?`
	_, err := r.db.ExecContext(ctx, query, plan.PlanType, plan.Status, plan.BillingCycle, plan.SeatCount, plan.SeatLimit, plan.StorageLimitMB, plan.StorageUsedMB, plan.PricePerSeat, plan.CurrentPeriodStart, plan.CurrentPeriodEnd, plan.CanceledAt, time.Now(), plan.WorkspaceID)
	return err
}

// Invoice methods
func (r *BillingRepository) CreateInvoice(ctx context.Context, invoice *models.BillingInvoice) error {
	query := `INSERT INTO workspace_invoices (id, workspace_id, invoice_number, amount, currency, status, description, period_start, period_end, paid_at, due_date, external_id, pdf_url, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, invoice.ID, invoice.WorkspaceID, invoice.InvoiceNumber, invoice.Amount, invoice.Currency, invoice.Status, invoice.Description, invoice.PeriodStart, invoice.PeriodEnd, invoice.PaidAt, invoice.DueDate, invoice.ExternalID, invoice.PDFURL, invoice.CreatedAt)
	return err
}

func (r *BillingRepository) GetInvoice(ctx context.Context, id uuid.UUID) (*models.BillingInvoice, error) {
	var invoice models.BillingInvoice
	err := r.db.GetContext(ctx, &invoice, "SELECT * FROM workspace_invoices WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &invoice, err
}

func (r *BillingRepository) ListInvoices(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*models.BillingInvoice, error) {
	var invoices []*models.BillingInvoice
	err := r.db.SelectContext(ctx, &invoices, "SELECT * FROM workspace_invoices WHERE workspace_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?", workspaceID, limit, offset)
	return invoices, err
}

func (r *BillingRepository) UpdateInvoiceStatus(ctx context.Context, id uuid.UUID, status string, paidAt *time.Time) error {
	_, err := r.db.ExecContext(ctx, "UPDATE workspace_invoices SET status = ?, paid_at = ? WHERE id = ?", status, paidAt, id)
	return err
}

// Payment Method methods
func (r *BillingRepository) CreatePaymentMethod(ctx context.Context, pm *models.PaymentMethod) error {
	query := `INSERT INTO workspace_payment_methods (id, workspace_id, type, brand, last4, exp_month, exp_year, is_default, external_id, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, pm.ID, pm.WorkspaceID, pm.Type, pm.Brand, pm.Last4, pm.ExpMonth, pm.ExpYear, pm.IsDefault, pm.ExternalID, pm.CreatedBy, pm.CreatedAt, pm.UpdatedAt)
	return err
}

func (r *BillingRepository) GetPaymentMethod(ctx context.Context, id uuid.UUID) (*models.PaymentMethod, error) {
	var pm models.PaymentMethod
	err := r.db.GetContext(ctx, &pm, "SELECT * FROM workspace_payment_methods WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &pm, err
}

func (r *BillingRepository) ListPaymentMethods(ctx context.Context, workspaceID uuid.UUID) ([]*models.PaymentMethod, error) {
	var methods []*models.PaymentMethod
	err := r.db.SelectContext(ctx, &methods, "SELECT * FROM workspace_payment_methods WHERE workspace_id = ? ORDER BY is_default DESC, created_at DESC", workspaceID)
	return methods, err
}

func (r *BillingRepository) GetDefaultPaymentMethod(ctx context.Context, workspaceID uuid.UUID) (*models.PaymentMethod, error) {
	var pm models.PaymentMethod
	err := r.db.GetContext(ctx, &pm, "SELECT * FROM workspace_payment_methods WHERE workspace_id = ? AND is_default = TRUE LIMIT 1", workspaceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &pm, err
}

func (r *BillingRepository) SetDefaultPaymentMethod(ctx context.Context, workspaceID, methodID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE workspace_payment_methods SET is_default = FALSE WHERE workspace_id = ?", workspaceID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "UPDATE workspace_payment_methods SET is_default = TRUE WHERE id = ?", methodID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *BillingRepository) DeletePaymentMethod(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workspace_payment_methods WHERE id = ?", id)
	return err
}

// Billing Events
func (r *BillingRepository) CreateEvent(ctx context.Context, event *models.BillingEvent) error {
	query := `INSERT INTO workspace_billing_events (id, workspace_id, event_type, description, metadata, actor_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, event.ID, event.WorkspaceID, event.EventType, event.Description, event.Metadata, event.ActorID, event.CreatedAt)
	return err
}

func (r *BillingRepository) ListEvents(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*models.BillingEvent, error) {
	var events []*models.BillingEvent
	err := r.db.SelectContext(ctx, &events, "SELECT * FROM workspace_billing_events WHERE workspace_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?", workspaceID, limit, offset)
	return events, err
}

// Usage tracking
func (r *BillingRepository) UpdateStorageUsage(ctx context.Context, workspaceID uuid.UUID, storageMB int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE workspace_plans SET storage_used_mb = ?, updated_at = ? WHERE workspace_id = ?", storageMB, time.Now(), workspaceID)
	return err
}

func (r *BillingRepository) UpdateSeatCount(ctx context.Context, workspaceID uuid.UUID, seatCount int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE workspace_plans SET seat_count = ?, updated_at = ? WHERE workspace_id = ?", seatCount, time.Now(), workspaceID)
	return err
}
