package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type InviteCodeRepository struct {
	db *sqlx.DB
}

func NewInviteCodeRepository(db *sqlx.DB) *InviteCodeRepository {
	return &InviteCodeRepository{db: db}
}

func (r *InviteCodeRepository) Create(ctx context.Context, ic *models.WorkspaceInviteCode) error {
	query := `
		INSERT INTO workspace_invite_codes (id, workspace_id, code, role, max_uses, use_count, created_by, expires_at, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, ic.ID, ic.WorkspaceID, ic.Code, ic.Role, ic.MaxUses, ic.UseCount, ic.CreatedBy, ic.ExpiresAt, ic.IsActive, ic.CreatedAt, ic.UpdatedAt)
	return err
}

func (r *InviteCodeRepository) GetByCode(ctx context.Context, code string) (*models.WorkspaceInviteCode, error) {
	var ic models.WorkspaceInviteCode
	query := `SELECT * FROM workspace_invite_codes WHERE code = ? AND is_active = TRUE AND (expires_at IS NULL OR expires_at > NOW())`
	err := r.db.GetContext(ctx, &ic, query, code)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &ic, err
}

func (r *InviteCodeRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceInviteCode, error) {
	var codes []*models.WorkspaceInviteCode
	query := `SELECT * FROM workspace_invite_codes WHERE workspace_id = ? ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &codes, query, workspaceID)
	return codes, err
}

func (r *InviteCodeRepository) IncrementUseCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE workspace_invite_codes SET use_count = use_count + 1, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *InviteCodeRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE workspace_invite_codes SET is_active = FALSE, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *InviteCodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM workspace_invite_codes WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
