package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type BookmarkRepository struct {
	db *sqlx.DB
}

func NewBookmarkRepository(db *sqlx.DB) *BookmarkRepository {
	return &BookmarkRepository{db: db}
}

func (r *BookmarkRepository) Create(ctx context.Context, bookmark *models.WorkspaceBookmark) error {
	query := `
		INSERT INTO workspace_bookmarks (id, workspace_id, user_id, title, url, entity_type, entity_id, notes, folder_name, position, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, bookmark.ID, bookmark.WorkspaceID, bookmark.UserID, bookmark.Title, bookmark.URL, bookmark.EntityType, bookmark.EntityID, bookmark.Notes, bookmark.FolderName, bookmark.Position, bookmark.CreatedAt, bookmark.UpdatedAt)
	return err
}

func (r *BookmarkRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WorkspaceBookmark, error) {
	var bookmark models.WorkspaceBookmark
	query := `SELECT * FROM workspace_bookmarks WHERE id = ?`
	err := r.db.GetContext(ctx, &bookmark, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &bookmark, err
}

func (r *BookmarkRepository) ListByUser(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.WorkspaceBookmark, error) {
	var bookmarks []*models.WorkspaceBookmark
	query := `SELECT * FROM workspace_bookmarks WHERE workspace_id = ? AND user_id = ? ORDER BY folder_name ASC, position ASC, created_at DESC`
	err := r.db.SelectContext(ctx, &bookmarks, query, workspaceID, userID)
	return bookmarks, err
}

func (r *BookmarkRepository) ListByFolder(ctx context.Context, workspaceID, userID uuid.UUID, folderName string) ([]*models.WorkspaceBookmark, error) {
	var bookmarks []*models.WorkspaceBookmark
	query := `SELECT * FROM workspace_bookmarks WHERE workspace_id = ? AND user_id = ? AND folder_name = ? ORDER BY position ASC`
	err := r.db.SelectContext(ctx, &bookmarks, query, workspaceID, userID, folderName)
	return bookmarks, err
}

func (r *BookmarkRepository) Update(ctx context.Context, bookmark *models.WorkspaceBookmark) error {
	query := `UPDATE workspace_bookmarks SET title = ?, url = ?, notes = ?, folder_name = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, bookmark.Title, bookmark.URL, bookmark.Notes, bookmark.FolderName, bookmark.ID)
	return err
}

func (r *BookmarkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM workspace_bookmarks WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *BookmarkRepository) GetMaxPosition(ctx context.Context, workspaceID, userID uuid.UUID) (int, error) {
	var maxPos sql.NullInt64
	query := `SELECT MAX(position) FROM workspace_bookmarks WHERE workspace_id = ? AND user_id = ?`
	err := r.db.GetContext(ctx, &maxPos, query, workspaceID, userID)
	if err != nil || !maxPos.Valid {
		return 0, err
	}
	return int(maxPos.Int64), nil
}

func (r *BookmarkRepository) CountByUser(ctx context.Context, workspaceID, userID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM workspace_bookmarks WHERE workspace_id = ? AND user_id = ?`
	err := r.db.GetContext(ctx, &count, query, workspaceID, userID)
	return count, err
}

func (r *BookmarkRepository) ListFolders(ctx context.Context, workspaceID, userID uuid.UUID) ([]string, error) {
	var folders []string
	query := `SELECT DISTINCT folder_name FROM workspace_bookmarks WHERE workspace_id = ? AND user_id = ? AND folder_name IS NOT NULL ORDER BY folder_name ASC`
	err := r.db.SelectContext(ctx, &folders, query, workspaceID, userID)
	return folders, err
}
