package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type WebhookRepository struct {
	db *sqlx.DB
}

func NewWebhookRepository(db *sqlx.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

func (r *WebhookRepository) Create(ctx context.Context, w *models.WorkspaceWebhook) error {
	query := `INSERT INTO workspace_webhooks (id, workspace_id, name, url, secret, events, is_active, created_by, failure_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, w.ID, w.WorkspaceID, w.Name, w.URL, w.Secret, w.Events, w.IsActive, w.CreatedBy, w.FailureCount, w.CreatedAt, w.UpdatedAt)
	return err
}

func (r *WebhookRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WorkspaceWebhook, error) {
	var w models.WorkspaceWebhook
	err := r.db.GetContext(ctx, &w, "SELECT * FROM workspace_webhooks WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &w, err
}

func (r *WebhookRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceWebhook, error) {
	var webhooks []*models.WorkspaceWebhook
	err := r.db.SelectContext(ctx, &webhooks, "SELECT * FROM workspace_webhooks WHERE workspace_id = ? ORDER BY created_at DESC", workspaceID)
	return webhooks, err
}

func (r *WebhookRepository) Update(ctx context.Context, w *models.WorkspaceWebhook) error {
	query := `UPDATE workspace_webhooks SET name = ?, url = ?, events = ?, is_active = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, w.Name, w.URL, w.Events, w.IsActive, time.Now(), w.ID)
	return err
}

func (r *WebhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workspace_webhooks WHERE id = ?", id)
	return err
}

func (r *WebhookRepository) ListActiveByEvent(ctx context.Context, workspaceID uuid.UUID, eventType string) ([]*models.WorkspaceWebhook, error) {
	var webhooks []*models.WorkspaceWebhook
	jsonEvent := fmt.Sprintf(`"%s"`, eventType)
	err := r.db.SelectContext(ctx, &webhooks,
		"SELECT * FROM workspace_webhooks WHERE workspace_id = ? AND is_active = TRUE AND JSON_CONTAINS(events, ?)",
		workspaceID, jsonEvent)
	return webhooks, err
}

func (r *WebhookRepository) UpdateLastTriggered(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE workspace_webhooks SET last_triggered_at = ?, updated_at = ? WHERE id = ?", time.Now(), time.Now(), id)
	return err
}

func (r *WebhookRepository) IncrementFailureCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE workspace_webhooks SET failure_count = failure_count + 1, updated_at = ? WHERE id = ?", time.Now(), id)
	return err
}

func (r *WebhookRepository) ResetFailureCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE workspace_webhooks SET failure_count = 0, updated_at = ? WHERE id = ?", time.Now(), id)
	return err
}
