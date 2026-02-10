package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type FeatureFlagRepository struct {
	db *sqlx.DB
}

func NewFeatureFlagRepository(db *sqlx.DB) *FeatureFlagRepository {
	return &FeatureFlagRepository{db: db}
}

func (r *FeatureFlagRepository) Create(ctx context.Context, flag *models.WorkspaceFeatureFlag) error {
	query := `
		INSERT INTO workspace_feature_flags (id, workspace_id, ` + "`key`" + `, enabled, description, metadata, created_by, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, flag.ID, flag.WorkspaceID, flag.Key, flag.Enabled, flag.Description, flag.Metadata, flag.CreatedBy, flag.UpdatedBy, flag.CreatedAt, flag.UpdatedAt)
	return err
}

func (r *FeatureFlagRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WorkspaceFeatureFlag, error) {
	var flag models.WorkspaceFeatureFlag
	query := "SELECT * FROM workspace_feature_flags WHERE id = ?"
	err := r.db.GetContext(ctx, &flag, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &flag, err
}

func (r *FeatureFlagRepository) GetByKey(ctx context.Context, workspaceID uuid.UUID, key string) (*models.WorkspaceFeatureFlag, error) {
	var flag models.WorkspaceFeatureFlag
	query := "SELECT * FROM workspace_feature_flags WHERE workspace_id = ? AND `key` = ?"
	err := r.db.GetContext(ctx, &flag, query, workspaceID, key)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &flag, err
}

func (r *FeatureFlagRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceFeatureFlag, error) {
	var flags []*models.WorkspaceFeatureFlag
	query := "SELECT * FROM workspace_feature_flags WHERE workspace_id = ? ORDER BY `key` ASC"
	err := r.db.SelectContext(ctx, &flags, query, workspaceID)
	return flags, err
}

func (r *FeatureFlagRepository) Update(ctx context.Context, flag *models.WorkspaceFeatureFlag) error {
	query := "UPDATE workspace_feature_flags SET enabled = ?, description = ?, metadata = ?, updated_by = ?, updated_at = NOW() WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, flag.Enabled, flag.Description, flag.Metadata, flag.UpdatedBy, flag.ID)
	return err
}

func (r *FeatureFlagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM workspace_feature_flags WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *FeatureFlagRepository) IsEnabled(ctx context.Context, workspaceID uuid.UUID, key string) (bool, error) {
	var enabled bool
	query := "SELECT enabled FROM workspace_feature_flags WHERE workspace_id = ? AND `key` = ?"
	err := r.db.GetContext(ctx, &enabled, query, workspaceID, key)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return enabled, err
}
