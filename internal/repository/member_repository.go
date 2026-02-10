package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type MemberRepository struct {
	db *sqlx.DB
}

func NewMemberRepository(db *sqlx.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) Create(ctx context.Context, m *models.WorkspaceMember) error {
	query := `
		INSERT INTO workspace_members (id, workspace_id, user_id, role, joined_at, invited_by, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, m.ID, m.WorkspaceID, m.UserID, m.Role, m.JoinedAt, m.InvitedBy, m.IsActive, m.CreatedAt, m.UpdatedAt)
	return err
}

func (r *MemberRepository) GetByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) (*models.WorkspaceMember, error) {
	var m models.WorkspaceMember
	query := `SELECT * FROM workspace_members WHERE workspace_id = ? AND user_id = ?`
	err := r.db.GetContext(ctx, &m, query, workspaceID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &m, err
}

func (r *MemberRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, page, perPage int) ([]*models.WorkspaceMember, int64, error) {
	var members []*models.WorkspaceMember
	var total int64
	offset := (page - 1) * perPage

	countQuery := `SELECT COUNT(*) FROM workspace_members WHERE workspace_id = ? AND is_active = TRUE`
	r.db.GetContext(ctx, &total, countQuery, workspaceID)

	query := `
		SELECT * FROM workspace_members WHERE workspace_id = ? AND is_active = TRUE
		ORDER BY joined_at ASC LIMIT ? OFFSET ?
	`
	err := r.db.SelectContext(ctx, &members, query, workspaceID, perPage, offset)
	return members, total, err
}

func (r *MemberRepository) UpdateRole(ctx context.Context, workspaceID, userID uuid.UUID, role string) error {
	query := `UPDATE workspace_members SET role = ?, updated_at = ? WHERE workspace_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, role, time.Now(), workspaceID, userID)
	return err
}

func (r *MemberRepository) Remove(ctx context.Context, workspaceID, userID uuid.UUID) error {
	query := `UPDATE workspace_members SET is_active = FALSE, updated_at = ? WHERE workspace_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now(), workspaceID, userID)
	return err
}

func (r *MemberRepository) IsMember(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM workspace_members WHERE workspace_id = ? AND user_id = ? AND is_active = TRUE`
	err := r.db.GetContext(ctx, &count, query, workspaceID, userID)
	return count > 0, err
}

func (r *MemberRepository) GetRole(ctx context.Context, workspaceID, userID uuid.UUID) (string, error) {
	var role string
	query := `SELECT role FROM workspace_members WHERE workspace_id = ? AND user_id = ? AND is_active = TRUE`
	err := r.db.GetContext(ctx, &role, query, workspaceID, userID)
	return role, err
}

func (r *MemberRepository) GetByID(ctx context.Context, workspaceID, userID uuid.UUID) (*models.WorkspaceMember, error) {
	var m models.WorkspaceMember
	query := `SELECT * FROM workspace_members WHERE workspace_id = ? AND user_id = ? AND is_active = TRUE`
	err := r.db.GetContext(ctx, &m, query, workspaceID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &m, err
}
