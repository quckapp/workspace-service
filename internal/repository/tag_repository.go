package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type TagRepository struct {
	db *sqlx.DB
}

func NewTagRepository(db *sqlx.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) Create(ctx context.Context, tag *models.WorkspaceTag) error {
	query := `INSERT INTO workspace_tags (id, workspace_id, name, color, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, tag.ID, tag.WorkspaceID, tag.Name, tag.Color, tag.CreatedBy, tag.CreatedAt, tag.UpdatedAt)
	return err
}

func (r *TagRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WorkspaceTag, error) {
	var tag models.WorkspaceTag
	err := r.db.GetContext(ctx, &tag, "SELECT * FROM workspace_tags WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tag, err
}

func (r *TagRepository) GetByName(ctx context.Context, workspaceID uuid.UUID, name string) (*models.WorkspaceTag, error) {
	var tag models.WorkspaceTag
	err := r.db.GetContext(ctx, &tag, "SELECT * FROM workspace_tags WHERE workspace_id = ? AND name = ?", workspaceID, name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tag, err
}

func (r *TagRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceTag, error) {
	var tags []*models.WorkspaceTag
	err := r.db.SelectContext(ctx, &tags, "SELECT * FROM workspace_tags WHERE workspace_id = ? ORDER BY name ASC", workspaceID)
	return tags, err
}

func (r *TagRepository) Update(ctx context.Context, tag *models.WorkspaceTag) error {
	query := `UPDATE workspace_tags SET name = ?, color = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, tag.Name, tag.Color, time.Now(), tag.ID)
	return err
}

func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workspace_tags WHERE id = ?", id)
	return err
}
