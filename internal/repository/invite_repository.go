package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type InviteRepository struct {
	db *sqlx.DB
}

func NewInviteRepository(db *sqlx.DB) *InviteRepository {
	return &InviteRepository{db: db}
}

func (r *InviteRepository) Create(ctx context.Context, inv *models.WorkspaceInvite) error {
	query := `
		INSERT INTO workspace_invites (id, workspace_id, email, role, token, invited_by, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, inv.ID, inv.WorkspaceID, inv.Email, inv.Role, inv.Token, inv.InvitedBy, inv.ExpiresAt, inv.CreatedAt)
	return err
}

func (r *InviteRepository) GetByToken(ctx context.Context, token string) (*models.WorkspaceInvite, error) {
	var inv models.WorkspaceInvite
	query := `SELECT * FROM workspace_invites WHERE token = ? AND accepted_at IS NULL AND expires_at > NOW()`
	err := r.db.GetContext(ctx, &inv, query, token)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &inv, err
}

func (r *InviteRepository) GetPendingByEmail(ctx context.Context, workspaceID uuid.UUID, email string) (*models.WorkspaceInvite, error) {
	var inv models.WorkspaceInvite
	query := `SELECT * FROM workspace_invites WHERE workspace_id = ? AND email = ? AND accepted_at IS NULL AND expires_at > NOW()`
	err := r.db.GetContext(ctx, &inv, query, workspaceID, email)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &inv, err
}

func (r *InviteRepository) MarkAccepted(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE workspace_invites SET accepted_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *InviteRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceInvite, error) {
	var invites []*models.WorkspaceInvite
	query := `SELECT * FROM workspace_invites WHERE workspace_id = ? AND accepted_at IS NULL AND expires_at > NOW() ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &invites, query, workspaceID)
	return invites, err
}

func (r *InviteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM workspace_invites WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *InviteRepository) GetPendingCount(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM workspace_invites WHERE workspace_id = ? AND accepted_at IS NULL AND expires_at > NOW()`
	err := r.db.GetContext(ctx, &count, query, workspaceID)
	return count, err
}

func (r *InviteRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WorkspaceInvite, error) {
	var inv models.WorkspaceInvite
	query := `SELECT * FROM workspace_invites WHERE id = ?`
	err := r.db.GetContext(ctx, &inv, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &inv, err
}
