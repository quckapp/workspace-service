package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type AnnouncementRepository struct {
	db *sqlx.DB
}

func NewAnnouncementRepository(db *sqlx.DB) *AnnouncementRepository {
	return &AnnouncementRepository{db: db}
}

func (r *AnnouncementRepository) Create(ctx context.Context, a *models.WorkspaceAnnouncement) error {
	query := `INSERT INTO workspace_announcements (id, workspace_id, title, content, priority, author_id, is_pinned, expires_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, a.ID, a.WorkspaceID, a.Title, a.Content, a.Priority, a.AuthorID, a.IsPinned, a.ExpiresAt, a.CreatedAt, a.UpdatedAt)
	return err
}

func (r *AnnouncementRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WorkspaceAnnouncement, error) {
	var a models.WorkspaceAnnouncement
	err := r.db.GetContext(ctx, &a, "SELECT * FROM workspace_announcements WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}

func (r *AnnouncementRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, page, perPage int) ([]*models.WorkspaceAnnouncement, int64, error) {
	var total int64
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM workspace_announcements WHERE workspace_id = ? AND (expires_at IS NULL OR expires_at > NOW())", workspaceID)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	var announcements []*models.WorkspaceAnnouncement
	err = r.db.SelectContext(ctx, &announcements,
		`SELECT * FROM workspace_announcements WHERE workspace_id = ? AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY is_pinned DESC, FIELD(priority, 'urgent', 'important', 'normal'), created_at DESC
		LIMIT ? OFFSET ?`, workspaceID, perPage, offset)
	return announcements, total, err
}

func (r *AnnouncementRepository) Update(ctx context.Context, a *models.WorkspaceAnnouncement) error {
	query := `UPDATE workspace_announcements SET title = ?, content = ?, priority = ?, expires_at = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, a.Title, a.Content, a.Priority, a.ExpiresAt, time.Now(), a.ID)
	return err
}

func (r *AnnouncementRepository) UpdatePinStatus(ctx context.Context, id uuid.UUID, isPinned bool) error {
	_, err := r.db.ExecContext(ctx, "UPDATE workspace_announcements SET is_pinned = ?, updated_at = ? WHERE id = ?", isPinned, time.Now(), id)
	return err
}

func (r *AnnouncementRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workspace_announcements WHERE id = ?", id)
	return err
}

func (r *AnnouncementRepository) CountActive(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM workspace_announcements WHERE workspace_id = ? AND (expires_at IS NULL OR expires_at > NOW())", workspaceID)
	return count, err
}
