package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type QuotaRepository struct {
	db *sqlx.DB
}

func NewQuotaRepository(db *sqlx.DB) *QuotaRepository {
	return &QuotaRepository{db: db}
}

func (r *QuotaRepository) Upsert(ctx context.Context, quota *models.WorkspaceQuota) error {
	query := `INSERT INTO workspace_quotas (id, workspace_id, max_members, max_channels, max_storage_mb, max_invite_codes, max_webhooks, max_roles, current_members, current_channels, current_storage_mb, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		max_members = VALUES(max_members), max_channels = VALUES(max_channels), max_storage_mb = VALUES(max_storage_mb),
		max_invite_codes = VALUES(max_invite_codes), max_webhooks = VALUES(max_webhooks), max_roles = VALUES(max_roles),
		current_members = VALUES(current_members), current_channels = VALUES(current_channels), current_storage_mb = VALUES(current_storage_mb),
		updated_at = VALUES(updated_at)`
	_, err := r.db.ExecContext(ctx, query, quota.ID, quota.WorkspaceID, quota.MaxMembers, quota.MaxChannels, quota.MaxStorageMB, quota.MaxInviteCodes, quota.MaxWebhooks, quota.MaxRoles, quota.CurrentMembers, quota.CurrentChannels, quota.CurrentStorageMB, quota.CreatedAt, quota.UpdatedAt)
	return err
}

func (r *QuotaRepository) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID) (*models.WorkspaceQuota, error) {
	var quota models.WorkspaceQuota
	err := r.db.GetContext(ctx, &quota, "SELECT * FROM workspace_quotas WHERE workspace_id = ?", workspaceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &quota, err
}

func (r *QuotaRepository) UpdateUsage(ctx context.Context, workspaceID uuid.UUID, field string, value int) error {
	query := "UPDATE workspace_quotas SET " + field + " = ?, updated_at = ? WHERE workspace_id = ?"
	_, err := r.db.ExecContext(ctx, query, value, time.Now(), workspaceID)
	return err
}

func (r *QuotaRepository) IncrementUsage(ctx context.Context, workspaceID uuid.UUID, field string) error {
	query := "UPDATE workspace_quotas SET " + field + " = " + field + " + 1, updated_at = ? WHERE workspace_id = ?"
	_, err := r.db.ExecContext(ctx, query, time.Now(), workspaceID)
	return err
}

func (r *QuotaRepository) DecrementUsage(ctx context.Context, workspaceID uuid.UUID, field string) error {
	query := "UPDATE workspace_quotas SET " + field + " = GREATEST(" + field + " - 1, 0), updated_at = ? WHERE workspace_id = ?"
	_, err := r.db.ExecContext(ctx, query, time.Now(), workspaceID)
	return err
}

func (r *QuotaRepository) Delete(ctx context.Context, workspaceID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workspace_quotas WHERE workspace_id = ?", workspaceID)
	return err
}
