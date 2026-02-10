package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type IntegrationRepository struct {
	db *sqlx.DB
}

func NewIntegrationRepository(db *sqlx.DB) *IntegrationRepository {
	return &IntegrationRepository{db: db}
}

func (r *IntegrationRepository) Create(ctx context.Context, integration *models.WorkspaceIntegration) error {
	query := `INSERT INTO workspace_integrations (id, workspace_id, provider, name, status, config, credentials, webhook_url, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		integration.ID, integration.WorkspaceID, integration.Provider, integration.Name,
		integration.Status, integration.Config, integration.Credentials, integration.WebhookURL,
		integration.CreatedBy, integration.CreatedAt, integration.UpdatedAt)
	return err
}

func (r *IntegrationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WorkspaceIntegration, error) {
	var integration models.WorkspaceIntegration
	query := `SELECT * FROM workspace_integrations WHERE id = ?`
	err := r.db.GetContext(ctx, &integration, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &integration, err
}

func (r *IntegrationRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceIntegration, error) {
	var integrations []*models.WorkspaceIntegration
	query := `SELECT * FROM workspace_integrations WHERE workspace_id = ? ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &integrations, query, workspaceID)
	return integrations, err
}

func (r *IntegrationRepository) ListByProvider(ctx context.Context, workspaceID uuid.UUID, provider string) ([]*models.WorkspaceIntegration, error) {
	var integrations []*models.WorkspaceIntegration
	query := `SELECT * FROM workspace_integrations WHERE workspace_id = ? AND provider = ? ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &integrations, query, workspaceID, provider)
	return integrations, err
}

func (r *IntegrationRepository) Update(ctx context.Context, integration *models.WorkspaceIntegration) error {
	query := `UPDATE workspace_integrations SET name = ?, status = ?, config = ?, credentials = ?, webhook_url = ?, error_message = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query,
		integration.Name, integration.Status, integration.Config, integration.Credentials,
		integration.WebhookURL, integration.ErrorMessage, integration.ID)
	return err
}

func (r *IntegrationRepository) UpdateSyncTime(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE workspace_integrations SET last_sync_at = NOW(), updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *IntegrationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM workspace_integrations WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *IntegrationRepository) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM workspace_integrations WHERE workspace_id = ?`
	err := r.db.GetContext(ctx, &count, query, workspaceID)
	return count, err
}
