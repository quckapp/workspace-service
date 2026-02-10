package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type PreferenceRepository struct {
	db *sqlx.DB
}

func NewPreferenceRepository(db *sqlx.DB) *PreferenceRepository {
	return &PreferenceRepository{db: db}
}

func (r *PreferenceRepository) Upsert(ctx context.Context, p *models.WorkspaceMemberPreference) error {
	query := `INSERT INTO workspace_member_preferences (id, workspace_id, user_id, notification_level, email_notifications, mute_until, sidebar_position, theme, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			notification_level = VALUES(notification_level),
			email_notifications = VALUES(email_notifications),
			mute_until = VALUES(mute_until),
			sidebar_position = VALUES(sidebar_position),
			theme = VALUES(theme),
			updated_at = VALUES(updated_at)`
	_, err := r.db.ExecContext(ctx, query, p.ID, p.WorkspaceID, p.UserID, p.NotificationLevel, p.EmailNotifications, p.MuteUntil, p.SidebarPosition, p.Theme, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *PreferenceRepository) GetByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) (*models.WorkspaceMemberPreference, error) {
	var p models.WorkspaceMemberPreference
	err := r.db.GetContext(ctx, &p, "SELECT * FROM workspace_member_preferences WHERE workspace_id = ? AND user_id = ?", workspaceID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (r *PreferenceRepository) Delete(ctx context.Context, workspaceID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workspace_member_preferences WHERE workspace_id = ? AND user_id = ?", workspaceID, userID)
	return err
}
