package models

import (
	"time"

	"github.com/google/uuid"
)

// ── Custom Emoji ──

type CustomEmoji struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	Name        string    `json:"name" db:"name"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	Category    *string   `json:"category" db:"category"`
	AliasFor    *string   `json:"alias_for" db:"alias_for"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	IsAnimated  bool      `json:"is_animated" db:"is_animated"`
	IsGlobal    bool      `json:"is_global" db:"is_global"`
	UsageCount  int       `json:"usage_count" db:"usage_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateEmojiRequest struct {
	Name       string  `json:"name" binding:"required,min=2,max=50"`
	ImageURL   string  `json:"image_url" binding:"required,url"`
	Category   *string `json:"category"`
	AliasFor   *string `json:"alias_for"`
	IsAnimated bool    `json:"is_animated"`
}

type UpdateEmojiRequest struct {
	Name     *string `json:"name"`
	Category *string `json:"category"`
	AliasFor *string `json:"alias_for"`
}

type EmojiCategory struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type EmojiSearchParams struct {
	Query    string `form:"q"`
	Category string `form:"category"`
	Page     int    `form:"page,default=1"`
	PerPage  int    `form:"per_page,default=50"`
}

type BulkDeleteEmojiRequest struct {
	EmojiIDs []string `json:"emoji_ids" binding:"required,min=1,max=50"`
}

type EmojiPackRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=100"`
	Description *string  `json:"description"`
	EmojiIDs    []string `json:"emoji_ids" binding:"required,min=1"`
}

type EmojiPack struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	EmojiCount  int       `json:"emoji_count" db:"emoji_count"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type EmojiPackMapping struct {
	ID       uuid.UUID `json:"id" db:"id"`
	PackID   uuid.UUID `json:"pack_id" db:"pack_id"`
	EmojiID  uuid.UUID `json:"emoji_id" db:"emoji_id"`
	Position int       `json:"position" db:"position"`
}

type EmojiStats struct {
	TotalEmojis    int             `json:"total_emojis"`
	AnimatedCount  int             `json:"animated_count"`
	TotalPacks     int             `json:"total_packs"`
	TopEmojis      []*CustomEmoji  `json:"top_emojis"`
	Categories     []EmojiCategory `json:"categories"`
}
