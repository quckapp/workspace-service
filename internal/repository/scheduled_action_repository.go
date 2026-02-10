package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type ScheduledActionRepository struct {
	db *sqlx.DB
}

func NewScheduledActionRepository(db *sqlx.DB) *ScheduledActionRepository {
	return &ScheduledActionRepository{db: db}
}

func (r *ScheduledActionRepository) Create(ctx context.Context, action *models.ScheduledAction) error {
	query := `INSERT INTO workspace_scheduled_actions (id, workspace_id, action_type, payload, scheduled_at, status, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, action.ID, action.WorkspaceID, action.ActionType, action.Payload, action.ScheduledAt, action.Status, action.CreatedBy, action.CreatedAt, action.UpdatedAt)
	return err
}

func (r *ScheduledActionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ScheduledAction, error) {
	var action models.ScheduledAction
	err := r.db.GetContext(ctx, &action, "SELECT * FROM workspace_scheduled_actions WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &action, err
}

func (r *ScheduledActionRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*models.ScheduledAction, error) {
	var actions []*models.ScheduledAction
	err := r.db.SelectContext(ctx, &actions,
		"SELECT * FROM workspace_scheduled_actions WHERE workspace_id = ? ORDER BY scheduled_at ASC",
		workspaceID)
	return actions, err
}

func (r *ScheduledActionRepository) ListPending(ctx context.Context, workspaceID uuid.UUID) ([]*models.ScheduledAction, error) {
	var actions []*models.ScheduledAction
	err := r.db.SelectContext(ctx, &actions,
		"SELECT * FROM workspace_scheduled_actions WHERE workspace_id = ? AND status = 'pending' ORDER BY scheduled_at ASC",
		workspaceID)
	return actions, err
}

func (r *ScheduledActionRepository) ListDue(ctx context.Context) ([]*models.ScheduledAction, error) {
	var actions []*models.ScheduledAction
	err := r.db.SelectContext(ctx, &actions,
		"SELECT * FROM workspace_scheduled_actions WHERE status = 'pending' AND scheduled_at <= NOW() ORDER BY scheduled_at ASC")
	return actions, err
}

func (r *ScheduledActionRepository) Update(ctx context.Context, action *models.ScheduledAction) error {
	query := `UPDATE workspace_scheduled_actions SET payload = ?, scheduled_at = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, action.Payload, action.ScheduledAt, time.Now(), action.ID)
	return err
}

func (r *ScheduledActionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE workspace_scheduled_actions SET status = ?, executed_at = ?, updated_at = ? WHERE id = ?", status, time.Now(), time.Now(), id)
	return err
}

func (r *ScheduledActionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workspace_scheduled_actions WHERE id = ?", id)
	return err
}

func (r *ScheduledActionRepository) CancelPending(ctx context.Context, workspaceID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE workspace_scheduled_actions SET status = 'cancelled', updated_at = ? WHERE workspace_id = ? AND status = 'pending'", time.Now(), workspaceID)
	return err
}
