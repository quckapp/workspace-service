package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type TemplateRepository struct {
	db *sqlx.DB
}

func NewTemplateRepository(db *sqlx.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) Create(ctx context.Context, t *models.WorkspaceTemplate) error {
	query := `INSERT INTO workspace_templates (id, name, description, created_by, default_roles, default_channels, default_settings, is_public, use_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, t.ID, t.Name, t.Description, t.CreatedBy, t.DefaultRoles, t.DefaultChannels, t.DefaultSettings, t.IsPublic, t.UseCount, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *TemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WorkspaceTemplate, error) {
	var t models.WorkspaceTemplate
	err := r.db.GetContext(ctx, &t, "SELECT * FROM workspace_templates WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *TemplateRepository) ListPublic(ctx context.Context, page, perPage int) ([]*models.WorkspaceTemplate, int64, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM workspace_templates WHERE is_public = TRUE")
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	var templates []*models.WorkspaceTemplate
	err = r.db.SelectContext(ctx, &templates, "SELECT * FROM workspace_templates WHERE is_public = TRUE ORDER BY use_count DESC, created_at DESC LIMIT ? OFFSET ?", perPage, offset)
	return templates, total, err
}

func (r *TemplateRepository) ListByCreator(ctx context.Context, userID uuid.UUID) ([]*models.WorkspaceTemplate, error) {
	var templates []*models.WorkspaceTemplate
	err := r.db.SelectContext(ctx, &templates, "SELECT * FROM workspace_templates WHERE created_by = ? ORDER BY created_at DESC", userID)
	return templates, err
}

func (r *TemplateRepository) Update(ctx context.Context, t *models.WorkspaceTemplate) error {
	query := `UPDATE workspace_templates SET name = ?, description = ?, is_public = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, t.Name, t.Description, t.IsPublic, time.Now(), t.ID)
	return err
}

func (r *TemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workspace_templates WHERE id = ?", id)
	return err
}

func (r *TemplateRepository) IncrementUseCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE workspace_templates SET use_count = use_count + 1, updated_at = ? WHERE id = ?", time.Now(), id)
	return err
}
