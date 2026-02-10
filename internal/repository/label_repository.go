package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type LabelRepository struct {
	db *sqlx.DB
}

func NewLabelRepository(db *sqlx.DB) *LabelRepository {
	return &LabelRepository{db: db}
}

func (r *LabelRepository) Create(ctx context.Context, label *models.WorkspaceLabel) error {
	query := `INSERT INTO workspace_labels (id, workspace_id, name, color, description, position, usage_count, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		label.ID, label.WorkspaceID, label.Name, label.Color, label.Description,
		label.Position, label.UsageCount, label.CreatedBy, label.CreatedAt, label.UpdatedAt)
	return err
}

func (r *LabelRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WorkspaceLabel, error) {
	var label models.WorkspaceLabel
	query := `SELECT * FROM workspace_labels WHERE id = ?`
	err := r.db.GetContext(ctx, &label, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &label, err
}

func (r *LabelRepository) GetByName(ctx context.Context, workspaceID uuid.UUID, name string) (*models.WorkspaceLabel, error) {
	var label models.WorkspaceLabel
	query := `SELECT * FROM workspace_labels WHERE workspace_id = ? AND name = ?`
	err := r.db.GetContext(ctx, &label, query, workspaceID, name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &label, err
}

func (r *LabelRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceLabel, error) {
	var labels []*models.WorkspaceLabel
	query := `SELECT * FROM workspace_labels WHERE workspace_id = ? ORDER BY position ASC, name ASC`
	err := r.db.SelectContext(ctx, &labels, query, workspaceID)
	return labels, err
}

func (r *LabelRepository) Update(ctx context.Context, label *models.WorkspaceLabel) error {
	query := `UPDATE workspace_labels SET name = ?, color = ?, description = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, label.Name, label.Color, label.Description, label.ID)
	return err
}

func (r *LabelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM workspace_labels WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *LabelRepository) GetMaxPosition(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var pos sql.NullInt64
	query := `SELECT MAX(position) FROM workspace_labels WHERE workspace_id = ?`
	err := r.db.GetContext(ctx, &pos, query, workspaceID)
	if err != nil || !pos.Valid {
		return 0, err
	}
	return int(pos.Int64), nil
}

func (r *LabelRepository) IncrementUsageCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE workspace_labels SET usage_count = usage_count + 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *LabelRepository) DecrementUsageCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE workspace_labels SET usage_count = GREATEST(usage_count - 1, 0) WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *LabelRepository) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM workspace_labels WHERE workspace_id = ?`
	err := r.db.GetContext(ctx, &count, query, workspaceID)
	return count, err
}
