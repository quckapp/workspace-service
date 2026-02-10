package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type MemberNoteRepository struct {
	db *sqlx.DB
}

func NewMemberNoteRepository(db *sqlx.DB) *MemberNoteRepository {
	return &MemberNoteRepository{db: db}
}

func (r *MemberNoteRepository) Create(ctx context.Context, note *models.MemberNote) error {
	query := `INSERT INTO workspace_member_notes (id, workspace_id, target_id, author_id, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, note.ID, note.WorkspaceID, note.TargetID, note.AuthorID, note.Content, note.CreatedAt, note.UpdatedAt)
	return err
}

func (r *MemberNoteRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.MemberNote, error) {
	var note models.MemberNote
	err := r.db.GetContext(ctx, &note, "SELECT * FROM workspace_member_notes WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &note, err
}

func (r *MemberNoteRepository) ListByTarget(ctx context.Context, workspaceID, targetID uuid.UUID) ([]*models.MemberNote, error) {
	var notes []*models.MemberNote
	err := r.db.SelectContext(ctx, &notes,
		"SELECT * FROM workspace_member_notes WHERE workspace_id = ? AND target_id = ? ORDER BY created_at DESC",
		workspaceID, targetID)
	return notes, err
}

func (r *MemberNoteRepository) ListByAuthor(ctx context.Context, workspaceID, authorID uuid.UUID) ([]*models.MemberNote, error) {
	var notes []*models.MemberNote
	err := r.db.SelectContext(ctx, &notes,
		"SELECT * FROM workspace_member_notes WHERE workspace_id = ? AND author_id = ? ORDER BY created_at DESC",
		workspaceID, authorID)
	return notes, err
}

func (r *MemberNoteRepository) Update(ctx context.Context, note *models.MemberNote) error {
	query := `UPDATE workspace_member_notes SET content = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, note.Content, time.Now(), note.ID)
	return err
}

func (r *MemberNoteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workspace_member_notes WHERE id = ?", id)
	return err
}
