package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type FavoriteRepository struct {
	db *sqlx.DB
}

func NewFavoriteRepository(db *sqlx.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

func (r *FavoriteRepository) Create(ctx context.Context, fav *models.WorkspaceFavorite) error {
	query := `INSERT INTO workspace_favorites (id, user_id, workspace_id, position, created_at)
		VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, fav.ID, fav.UserID, fav.WorkspaceID, fav.Position, fav.CreatedAt)
	return err
}

func (r *FavoriteRepository) Delete(ctx context.Context, userID, workspaceID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workspace_favorites WHERE user_id = ? AND workspace_id = ?", userID, workspaceID)
	return err
}

func (r *FavoriteRepository) GetByUserAndWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) (*models.WorkspaceFavorite, error) {
	var fav models.WorkspaceFavorite
	err := r.db.GetContext(ctx, &fav, "SELECT * FROM workspace_favorites WHERE user_id = ? AND workspace_id = ?", userID, workspaceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &fav, err
}

func (r *FavoriteRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.WorkspaceFavorite, error) {
	var favs []*models.WorkspaceFavorite
	err := r.db.SelectContext(ctx, &favs, "SELECT * FROM workspace_favorites WHERE user_id = ? ORDER BY position ASC, created_at ASC", userID)
	return favs, err
}

func (r *FavoriteRepository) GetMaxPosition(ctx context.Context, userID uuid.UUID) (int, error) {
	var pos sql.NullInt64
	err := r.db.GetContext(ctx, &pos, "SELECT MAX(position) FROM workspace_favorites WHERE user_id = ?", userID)
	if err != nil || !pos.Valid {
		return 0, err
	}
	return int(pos.Int64), nil
}

func (r *FavoriteRepository) UpdatePositions(ctx context.Context, userID uuid.UUID, workspaceIDs []uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, wsID := range workspaceIDs {
		_, err := tx.ExecContext(ctx, "UPDATE workspace_favorites SET position = ? WHERE user_id = ? AND workspace_id = ?", i, userID, wsID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *FavoriteRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM workspace_favorites WHERE user_id = ?", userID)
	return count, err
}
