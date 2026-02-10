package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type ReactionRepository struct {
	db *sqlx.DB
}

func NewReactionRepository(db *sqlx.DB) *ReactionRepository {
	return &ReactionRepository{db: db}
}

func (r *ReactionRepository) Create(ctx context.Context, reaction *models.WorkspaceReaction) error {
	query := `
		INSERT INTO workspace_reactions (id, entity_type, entity_id, user_id, emoji, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, reaction.ID, reaction.EntityType, reaction.EntityID, reaction.UserID, reaction.Emoji, reaction.CreatedAt)
	return err
}

func (r *ReactionRepository) Delete(ctx context.Context, entityType string, entityID, userID uuid.UUID, emoji string) error {
	query := `DELETE FROM workspace_reactions WHERE entity_type = ? AND entity_id = ? AND user_id = ? AND emoji = ?`
	_, err := r.db.ExecContext(ctx, query, entityType, entityID, userID, emoji)
	return err
}

func (r *ReactionRepository) Exists(ctx context.Context, entityType string, entityID, userID uuid.UUID, emoji string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM workspace_reactions WHERE entity_type = ? AND entity_id = ? AND user_id = ? AND emoji = ?`
	err := r.db.GetContext(ctx, &count, query, entityType, entityID, userID, emoji)
	return count > 0, err
}

func (r *ReactionRepository) ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.WorkspaceReaction, error) {
	var reactions []*models.WorkspaceReaction
	query := `SELECT * FROM workspace_reactions WHERE entity_type = ? AND entity_id = ? ORDER BY created_at ASC`
	err := r.db.SelectContext(ctx, &reactions, query, entityType, entityID)
	return reactions, err
}

func (r *ReactionRepository) GetSummary(ctx context.Context, entityType string, entityID uuid.UUID) ([]models.ReactionSummary, error) {
	var summaries []models.ReactionSummary
	query := `
		SELECT emoji, COUNT(*) as count FROM workspace_reactions
		WHERE entity_type = ? AND entity_id = ?
		GROUP BY emoji ORDER BY count DESC
	`
	err := r.db.SelectContext(ctx, &summaries, query, entityType, entityID)
	return summaries, err
}

func (r *ReactionRepository) DeleteAllByEntity(ctx context.Context, entityType string, entityID uuid.UUID) error {
	query := `DELETE FROM workspace_reactions WHERE entity_type = ? AND entity_id = ?`
	_, err := r.db.ExecContext(ctx, query, entityType, entityID)
	return err
}
