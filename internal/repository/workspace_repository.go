package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckchat/workspace-service/internal/models"
)

type WorkspaceRepository struct {
	db *sqlx.DB
}

func NewWorkspaceRepository(db *sqlx.DB) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

func (r *WorkspaceRepository) Create(ctx context.Context, w *models.Workspace) error {
	query := `
		INSERT INTO workspaces (id, name, slug, description, icon_url, owner_id, plan, settings, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, w.ID, w.Name, w.Slug, w.Description, w.IconURL, w.OwnerID, w.Plan, w.Settings, w.IsActive, w.CreatedAt, w.UpdatedAt)
	return err
}

func (r *WorkspaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Workspace, error) {
	var w models.Workspace
	query := `SELECT * FROM workspaces WHERE id = ? AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &w, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &w, err
}

func (r *WorkspaceRepository) GetBySlug(ctx context.Context, slug string) (*models.Workspace, error) {
	var w models.Workspace
	query := `SELECT * FROM workspaces WHERE slug = ? AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &w, query, slug)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &w, err
}

func (r *WorkspaceRepository) Update(ctx context.Context, w *models.Workspace) error {
	query := `
		UPDATE workspaces SET name = ?, description = ?, icon_url = ?, settings = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, w.Name, w.Description, w.IconURL, w.Settings, time.Now(), w.ID)
	return err
}

func (r *WorkspaceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE workspaces SET deleted_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *WorkspaceRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, perPage int) ([]*models.Workspace, int64, error) {
	var workspaces []*models.Workspace
	var total int64
	offset := (page - 1) * perPage

	countQuery := `
		SELECT COUNT(*) FROM workspaces w
		INNER JOIN workspace_members m ON w.id = m.workspace_id
		WHERE m.user_id = ? AND w.deleted_at IS NULL AND m.is_active = TRUE
	`
	r.db.GetContext(ctx, &total, countQuery, userID)

	query := `
		SELECT w.* FROM workspaces w
		INNER JOIN workspace_members m ON w.id = m.workspace_id
		WHERE m.user_id = ? AND w.deleted_at IS NULL AND m.is_active = TRUE
		ORDER BY w.created_at DESC
		LIMIT ? OFFSET ?
	`
	err := r.db.SelectContext(ctx, &workspaces, query, userID, perPage, offset)
	return workspaces, total, err
}

func (r *WorkspaceRepository) GetMemberCount(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM workspace_members WHERE workspace_id = ? AND is_active = TRUE`
	err := r.db.GetContext(ctx, &count, query, workspaceID)
	return count, err
}
