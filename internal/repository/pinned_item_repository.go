package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type PinnedItemRepository struct {
	db *sqlx.DB
}

func NewPinnedItemRepository(db *sqlx.DB) *PinnedItemRepository {
	return &PinnedItemRepository{db: db}
}

func (r *PinnedItemRepository) Create(ctx context.Context, item *models.WorkspacePinnedItem) error {
	query := `
		INSERT INTO workspace_pinned_items (id, workspace_id, item_type, item_id, title, content, url, pinned_by, position, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, item.ID, item.WorkspaceID, item.ItemType, item.ItemID, item.Title, item.Content, item.URL, item.PinnedBy, item.Position, item.CreatedAt, item.UpdatedAt)
	return err
}

func (r *PinnedItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WorkspacePinnedItem, error) {
	var item models.WorkspacePinnedItem
	query := `SELECT * FROM workspace_pinned_items WHERE id = ?`
	err := r.db.GetContext(ctx, &item, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &item, err
}

func (r *PinnedItemRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspacePinnedItem, error) {
	var items []*models.WorkspacePinnedItem
	query := `SELECT * FROM workspace_pinned_items WHERE workspace_id = ? ORDER BY position ASC, created_at DESC`
	err := r.db.SelectContext(ctx, &items, query, workspaceID)
	return items, err
}

func (r *PinnedItemRepository) Update(ctx context.Context, item *models.WorkspacePinnedItem) error {
	query := `UPDATE workspace_pinned_items SET title = ?, content = ?, url = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, item.Title, item.Content, item.URL, item.ID)
	return err
}

func (r *PinnedItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM workspace_pinned_items WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PinnedItemRepository) GetMaxPosition(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var maxPos sql.NullInt64
	query := `SELECT MAX(position) FROM workspace_pinned_items WHERE workspace_id = ?`
	err := r.db.GetContext(ctx, &maxPos, query, workspaceID)
	if err != nil || !maxPos.Valid {
		return 0, err
	}
	return int(maxPos.Int64), nil
}

func (r *PinnedItemRepository) UpdatePositions(ctx context.Context, workspaceID uuid.UUID, pinIDs []uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, pinID := range pinIDs {
		_, err := tx.ExecContext(ctx, `UPDATE workspace_pinned_items SET position = ? WHERE id = ? AND workspace_id = ?`, i, pinID, workspaceID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PinnedItemRepository) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM workspace_pinned_items WHERE workspace_id = ?`
	err := r.db.GetContext(ctx, &count, query, workspaceID)
	return count, err
}
