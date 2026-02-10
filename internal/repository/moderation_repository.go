package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type ModerationRepository struct {
	db *sqlx.DB
}

func NewModerationRepository(db *sqlx.DB) *ModerationRepository {
	return &ModerationRepository{db: db}
}

// ── Ban Operations ──

func (r *ModerationRepository) CreateBan(ctx context.Context, ban *models.WorkspaceBan) error {
	query := `INSERT INTO workspace_bans (id, workspace_id, user_id, banned_by, reason, expires_at, is_permanent, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, ban.ID, ban.WorkspaceID, ban.UserID, ban.BannedBy, ban.Reason, ban.ExpiresAt, ban.IsPermanent, ban.CreatedAt)
	return err
}

func (r *ModerationRepository) RemoveBan(ctx context.Context, workspaceID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workspace_bans WHERE workspace_id = ? AND user_id = ?", workspaceID, userID)
	return err
}

func (r *ModerationRepository) IsUserBanned(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM workspace_bans WHERE workspace_id = ? AND user_id = ? AND (is_permanent = TRUE OR expires_at IS NULL OR expires_at > NOW())", workspaceID, userID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ModerationRepository) ListBans(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceBan, error) {
	var bans []*models.WorkspaceBan
	err := r.db.SelectContext(ctx, &bans, "SELECT * FROM workspace_bans WHERE workspace_id = ? ORDER BY created_at DESC", workspaceID)
	return bans, err
}

func (r *ModerationRepository) GetBan(ctx context.Context, workspaceID, userID uuid.UUID) (*models.WorkspaceBan, error) {
	var ban models.WorkspaceBan
	err := r.db.GetContext(ctx, &ban, "SELECT * FROM workspace_bans WHERE workspace_id = ? AND user_id = ?", workspaceID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &ban, err
}

// ── Mute Operations ──

func (r *ModerationRepository) CreateMute(ctx context.Context, mute *models.WorkspaceMute) error {
	query := `INSERT INTO workspace_mutes (id, workspace_id, user_id, muted_by, reason, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, mute.ID, mute.WorkspaceID, mute.UserID, mute.MutedBy, mute.Reason, mute.ExpiresAt, mute.CreatedAt)
	return err
}

func (r *ModerationRepository) RemoveMute(ctx context.Context, workspaceID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workspace_mutes WHERE workspace_id = ? AND user_id = ?", workspaceID, userID)
	return err
}

func (r *ModerationRepository) IsUserMuted(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM workspace_mutes WHERE workspace_id = ? AND user_id = ? AND (expires_at IS NULL OR expires_at > NOW())", workspaceID, userID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ModerationRepository) ListMutes(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceMute, error) {
	var mutes []*models.WorkspaceMute
	err := r.db.SelectContext(ctx, &mutes, "SELECT * FROM workspace_mutes WHERE workspace_id = ? ORDER BY created_at DESC", workspaceID)
	return mutes, err
}

func (r *ModerationRepository) GetMute(ctx context.Context, workspaceID, userID uuid.UUID) (*models.WorkspaceMute, error) {
	var mute models.WorkspaceMute
	err := r.db.GetContext(ctx, &mute, "SELECT * FROM workspace_mutes WHERE workspace_id = ? AND user_id = ?", workspaceID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &mute, err
}
