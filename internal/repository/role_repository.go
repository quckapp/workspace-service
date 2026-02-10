package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type RoleRepository struct {
	db *sqlx.DB
}

func NewRoleRepository(db *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) Create(ctx context.Context, role *models.WorkspaceRole) error {
	query := `
		INSERT INTO workspace_roles (id, workspace_id, name, color, priority, permissions, is_default, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		role.ID, role.WorkspaceID, role.Name, role.Color, role.Priority,
		role.Permissions, role.IsDefault, role.CreatedBy,
		role.CreatedAt, role.UpdatedAt,
	)
	return err
}

func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WorkspaceRole, error) {
	var role models.WorkspaceRole
	query := `SELECT * FROM workspace_roles WHERE id = ?`
	err := r.db.GetContext(ctx, &role, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &role, err
}

func (r *RoleRepository) GetByName(ctx context.Context, workspaceID uuid.UUID, name string) (*models.WorkspaceRole, error) {
	var role models.WorkspaceRole
	query := `SELECT * FROM workspace_roles WHERE workspace_id = ? AND name = ?`
	err := r.db.GetContext(ctx, &role, query, workspaceID, name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &role, err
}

func (r *RoleRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceRole, error) {
	var roles []*models.WorkspaceRole
	query := `SELECT * FROM workspace_roles WHERE workspace_id = ? ORDER BY priority DESC, name ASC`
	err := r.db.SelectContext(ctx, &roles, query, workspaceID)
	return roles, err
}

func (r *RoleRepository) Update(ctx context.Context, role *models.WorkspaceRole) error {
	query := `
		UPDATE workspace_roles SET name = ?, color = ?, priority = ?, permissions = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, role.Name, role.Color, role.Priority, role.Permissions, time.Now(), role.ID)
	return err
}

func (r *RoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM workspace_roles WHERE id = ? AND is_default = FALSE`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *RoleRepository) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM workspace_roles WHERE workspace_id = ?`
	err := r.db.GetContext(ctx, &count, query, workspaceID)
	return count, err
}
