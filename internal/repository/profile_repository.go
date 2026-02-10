package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type ProfileRepository struct {
	db *sqlx.DB
}

func NewProfileRepository(db *sqlx.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) Upsert(ctx context.Context, profile *models.MemberProfile) error {
	query := `
		INSERT INTO workspace_member_profiles (id, workspace_id, user_id, display_name, title, status_text, status_emoji, timezone, is_online, last_seen_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			display_name = VALUES(display_name),
			title = VALUES(title),
			status_text = VALUES(status_text),
			status_emoji = VALUES(status_emoji),
			timezone = VALUES(timezone),
			updated_at = VALUES(updated_at)
	`
	_, err := r.db.ExecContext(ctx, query,
		profile.ID, profile.WorkspaceID, profile.UserID,
		profile.DisplayName, profile.Title, profile.StatusText, profile.StatusEmoji,
		profile.Timezone, profile.IsOnline, profile.LastSeenAt,
		profile.CreatedAt, profile.UpdatedAt,
	)
	return err
}

func (r *ProfileRepository) GetByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) (*models.MemberProfile, error) {
	var profile models.MemberProfile
	query := `SELECT * FROM workspace_member_profiles WHERE workspace_id = ? AND user_id = ?`
	err := r.db.GetContext(ctx, &profile, query, workspaceID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &profile, err
}

func (r *ProfileRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*models.MemberProfile, error) {
	var profiles []*models.MemberProfile
	query := `SELECT * FROM workspace_member_profiles WHERE workspace_id = ?`
	err := r.db.SelectContext(ctx, &profiles, query, workspaceID)
	return profiles, err
}

func (r *ProfileRepository) UpdateOnlineStatus(ctx context.Context, workspaceID, userID uuid.UUID, isOnline bool) error {
	now := time.Now()
	query := `
		UPDATE workspace_member_profiles SET is_online = ?, last_seen_at = ?, updated_at = ?
		WHERE workspace_id = ? AND user_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, isOnline, now, now, workspaceID, userID)
	return err
}

func (r *ProfileRepository) Delete(ctx context.Context, workspaceID, userID uuid.UUID) error {
	query := `DELETE FROM workspace_member_profiles WHERE workspace_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, workspaceID, userID)
	return err
}
