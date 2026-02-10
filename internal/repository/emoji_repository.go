package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type EmojiRepository struct {
	db *sqlx.DB
}

func NewEmojiRepository(db *sqlx.DB) *EmojiRepository {
	return &EmojiRepository{db: db}
}

func (r *EmojiRepository) Create(ctx context.Context, emoji *models.CustomEmoji) error {
	query := `INSERT INTO workspace_custom_emojis (id, workspace_id, name, image_url, category, alias_for, created_by, is_animated, is_global, usage_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, emoji.ID, emoji.WorkspaceID, emoji.Name, emoji.ImageURL, emoji.Category, emoji.AliasFor, emoji.CreatedBy, emoji.IsAnimated, emoji.IsGlobal, emoji.UsageCount, emoji.CreatedAt, emoji.UpdatedAt)
	return err
}

func (r *EmojiRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CustomEmoji, error) {
	var emoji models.CustomEmoji
	err := r.db.GetContext(ctx, &emoji, "SELECT * FROM workspace_custom_emojis WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &emoji, err
}

func (r *EmojiRepository) GetByName(ctx context.Context, workspaceID uuid.UUID, name string) (*models.CustomEmoji, error) {
	var emoji models.CustomEmoji
	err := r.db.GetContext(ctx, &emoji, "SELECT * FROM workspace_custom_emojis WHERE workspace_id = ? AND name = ?", workspaceID, name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &emoji, err
}

func (r *EmojiRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*models.CustomEmoji, error) {
	var emojis []*models.CustomEmoji
	err := r.db.SelectContext(ctx, &emojis, "SELECT * FROM workspace_custom_emojis WHERE workspace_id = ? ORDER BY name ASC LIMIT ? OFFSET ?", workspaceID, limit, offset)
	return emojis, err
}

func (r *EmojiRepository) Search(ctx context.Context, workspaceID uuid.UUID, query string, category string, limit, offset int) ([]*models.CustomEmoji, error) {
	var emojis []*models.CustomEmoji
	q := "SELECT * FROM workspace_custom_emojis WHERE workspace_id = ?"
	args := []interface{}{workspaceID}

	if query != "" {
		q += " AND name LIKE ?"
		args = append(args, "%"+query+"%")
	}
	if category != "" {
		q += " AND category = ?"
		args = append(args, category)
	}
	q += " ORDER BY usage_count DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &emojis, q, args...)
	return emojis, err
}

func (r *EmojiRepository) Update(ctx context.Context, emoji *models.CustomEmoji) error {
	query := `UPDATE workspace_custom_emojis SET name = ?, category = ?, alias_for = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, emoji.Name, emoji.Category, emoji.AliasFor, time.Now(), emoji.ID)
	return err
}

func (r *EmojiRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workspace_custom_emojis WHERE id = ?", id)
	return err
}

func (r *EmojiRepository) IncrementUsage(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE workspace_custom_emojis SET usage_count = usage_count + 1 WHERE id = ?", id)
	return err
}

func (r *EmojiRepository) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM workspace_custom_emojis WHERE workspace_id = ?", workspaceID)
	return count, err
}

func (r *EmojiRepository) GetCategories(ctx context.Context, workspaceID uuid.UUID) ([]models.EmojiCategory, error) {
	var categories []models.EmojiCategory
	err := r.db.SelectContext(ctx, &categories, "SELECT COALESCE(category, 'uncategorized') as name, COUNT(*) as count FROM workspace_custom_emojis WHERE workspace_id = ? GROUP BY category ORDER BY count DESC", workspaceID)
	return categories, err
}

func (r *EmojiRepository) GetTopEmojis(ctx context.Context, workspaceID uuid.UUID, limit int) ([]*models.CustomEmoji, error) {
	var emojis []*models.CustomEmoji
	err := r.db.SelectContext(ctx, &emojis, "SELECT * FROM workspace_custom_emojis WHERE workspace_id = ? ORDER BY usage_count DESC LIMIT ?", workspaceID, limit)
	return emojis, err
}

func (r *EmojiRepository) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	query, args, err := sqlx.In("DELETE FROM workspace_custom_emojis WHERE id IN (?)", ids)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *EmojiRepository) CountAnimated(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM workspace_custom_emojis WHERE workspace_id = ? AND is_animated = TRUE", workspaceID)
	return count, err
}

// Emoji Pack methods
func (r *EmojiRepository) CreatePack(ctx context.Context, pack *models.EmojiPack) error {
	query := `INSERT INTO workspace_emoji_packs (id, workspace_id, name, description, created_by, emoji_count, is_public, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, pack.ID, pack.WorkspaceID, pack.Name, pack.Description, pack.CreatedBy, pack.EmojiCount, pack.IsPublic, pack.CreatedAt, pack.UpdatedAt)
	return err
}

func (r *EmojiRepository) GetPackByID(ctx context.Context, id uuid.UUID) (*models.EmojiPack, error) {
	var pack models.EmojiPack
	err := r.db.GetContext(ctx, &pack, "SELECT * FROM workspace_emoji_packs WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &pack, err
}

func (r *EmojiRepository) ListPacks(ctx context.Context, workspaceID uuid.UUID) ([]*models.EmojiPack, error) {
	var packs []*models.EmojiPack
	err := r.db.SelectContext(ctx, &packs, "SELECT * FROM workspace_emoji_packs WHERE workspace_id = ? ORDER BY name ASC", workspaceID)
	return packs, err
}

func (r *EmojiRepository) DeletePack(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workspace_emoji_packs WHERE id = ?", id)
	return err
}

func (r *EmojiRepository) AddEmojiToPack(ctx context.Context, mapping *models.EmojiPackMapping) error {
	query := `INSERT INTO workspace_emoji_pack_mappings (id, pack_id, emoji_id, position) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, mapping.ID, mapping.PackID, mapping.EmojiID, mapping.Position)
	return err
}

func (r *EmojiRepository) RemoveEmojiFromPack(ctx context.Context, packID, emojiID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workspace_emoji_pack_mappings WHERE pack_id = ? AND emoji_id = ?", packID, emojiID)
	return err
}

func (r *EmojiRepository) ListPackEmojis(ctx context.Context, packID uuid.UUID) ([]*models.CustomEmoji, error) {
	var emojis []*models.CustomEmoji
	err := r.db.SelectContext(ctx, &emojis, `SELECT e.* FROM workspace_custom_emojis e
		JOIN workspace_emoji_pack_mappings m ON e.id = m.emoji_id
		WHERE m.pack_id = ? ORDER BY m.position ASC`, packID)
	return emojis, err
}

func (r *EmojiRepository) CountPacks(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM workspace_emoji_packs WHERE workspace_id = ?", workspaceID)
	return count, err
}
