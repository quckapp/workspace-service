package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
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

func (r *WorkspaceRepository) TransferOwnership(ctx context.Context, workspaceID, newOwnerID uuid.UUID) error {
	query := `UPDATE workspaces SET owner_id = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, newOwnerID, time.Now(), workspaceID)
	return err
}

func (r *WorkspaceRepository) Search(ctx context.Context, query string, page, perPage int) ([]*models.Workspace, int64, error) {
	var workspaces []*models.Workspace
	var total int64
	offset := (page - 1) * perPage
	searchTerm := "%" + query + "%"

	countQuery := `SELECT COUNT(*) FROM workspaces WHERE (name LIKE ? OR slug LIKE ? OR description LIKE ?) AND deleted_at IS NULL AND is_active = TRUE`
	r.db.GetContext(ctx, &total, countQuery, searchTerm, searchTerm, searchTerm)

	q := `
		SELECT * FROM workspaces WHERE (name LIKE ? OR slug LIKE ? OR description LIKE ?)
		AND deleted_at IS NULL AND is_active = TRUE
		ORDER BY name ASC LIMIT ? OFFSET ?
	`
	err := r.db.SelectContext(ctx, &workspaces, q, searchTerm, searchTerm, searchTerm, perPage, offset)
	return workspaces, total, err
}

func (r *WorkspaceRepository) GetMemberGrowth(ctx context.Context, workspaceID uuid.UUID, days int) ([]models.DailyCount, error) {
	var counts []models.DailyCount
	query := `
		SELECT DATE(joined_at) as date, COUNT(*) as count FROM workspace_members
		WHERE workspace_id = ? AND joined_at >= DATE_SUB(NOW(), INTERVAL ? DAY) AND is_active = TRUE
		GROUP BY DATE(joined_at) ORDER BY date ASC
	`
	err := r.db.SelectContext(ctx, &counts, query, workspaceID, days)
	return counts, err
}

func (r *WorkspaceRepository) GetJoinMethodStats(ctx context.Context, workspaceID uuid.UUID) (map[string]int, error) {
	type methodCount struct {
		Method string `db:"method"`
		Count  int    `db:"count"`
	}
	var counts []methodCount
	query := `
		SELECT CASE WHEN invited_by IS NULL THEN 'direct' ELSE 'invited' END as method, COUNT(*) as count
		FROM workspace_members WHERE workspace_id = ? AND is_active = TRUE GROUP BY method
	`
	err := r.db.SelectContext(ctx, &counts, query, workspaceID)
	if err != nil {
		return nil, err
	}
	result := make(map[string]int)
	for _, mc := range counts {
		result[mc.Method] = mc.Count
	}
	return result, nil
}

func (r *WorkspaceRepository) GetRoleCounts(ctx context.Context, workspaceID uuid.UUID) (map[string]int, error) {
	type roleCount struct {
		Role  string `db:"role"`
		Count int    `db:"count"`
	}
	var counts []roleCount
	query := `SELECT role, COUNT(*) as count FROM workspace_members WHERE workspace_id = ? AND is_active = TRUE GROUP BY role`
	err := r.db.SelectContext(ctx, &counts, query, workspaceID)
	if err != nil {
		return nil, err
	}
	result := make(map[string]int)
	for _, rc := range counts {
		result[rc.Role] = rc.Count
	}
	return result, nil
}

func (r *WorkspaceRepository) ListArchivedByUser(ctx context.Context, userID uuid.UUID) ([]*models.Workspace, error) {
	var workspaces []*models.Workspace
	query := `
		SELECT w.* FROM workspaces w
		INNER JOIN workspace_members m ON w.id = m.workspace_id
		WHERE m.user_id = ? AND m.is_active = TRUE AND w.deleted_at IS NOT NULL
		ORDER BY w.deleted_at DESC
	`
	err := r.db.SelectContext(ctx, &workspaces, query, userID)
	return workspaces, err
}
