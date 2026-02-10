package models

import (
	"time"

	"github.com/google/uuid"
)

// ── Workspace Discovery & Recommendations ──

type WorkspaceDirectoryEntry struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	IsListed    bool      `json:"is_listed" db:"is_listed"`
	Category    *string   `json:"category" db:"category"` // engineering, marketing, sales, support, etc.
	Tags        JSON      `json:"tags" db:"tags"`
	Description *string   `json:"description" db:"description"`
	MemberCount int       `json:"member_count" db:"member_count"`
	IconURL     *string   `json:"icon_url" db:"icon_url"`
	BannerURL   *string   `json:"banner_url" db:"banner_url"`
	Website     *string   `json:"website" db:"website"`
	IsVerified  bool      `json:"is_verified" db:"is_verified"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type UpdateDirectoryEntryRequest struct {
	IsListed    *bool    `json:"is_listed"`
	Category    *string  `json:"category" binding:"omitempty,oneof=engineering marketing sales support design product hr finance legal operations other"`
	Tags        []string `json:"tags"`
	Description *string  `json:"description"`
	BannerURL   *string  `json:"banner_url"`
	Website     *string  `json:"website"`
}

type DirectorySearchParams struct {
	Query    string `form:"q"`
	Category string `form:"category"`
	SortBy   string `form:"sort_by,default=member_count"` // member_count, name, created_at
	Page     int    `form:"page,default=1"`
	PerPage  int    `form:"per_page,default=20"`
}

type WorkspaceRecommendation struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	WorkspaceID     uuid.UUID `json:"workspace_id" db:"workspace_id"`
	Reason          string    `json:"reason" db:"reason"` // similar_members, popular, category_match, invited
	Score           float64   `json:"score" db:"score"`
	IsDismissed     bool      `json:"is_dismissed" db:"is_dismissed"`
	DismissedAt     *time.Time `json:"dismissed_at" db:"dismissed_at"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type WorkspaceRecommendationResponse struct {
	Workspace *WorkspaceResponse       `json:"workspace"`
	Reason    string                   `json:"reason"`
	Score     float64                  `json:"score"`
}

type TrendingWorkspace struct {
	Workspace    *Workspace `json:"workspace"`
	MemberCount  int        `json:"member_count" db:"member_count"`
	GrowthRate   float64    `json:"growth_rate" db:"growth_rate"` // percentage growth last 7 days
	ActiveUsers  int        `json:"active_users" db:"active_users"`
}

type WorkspaceCategory struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type DismissRecommendationRequest struct {
	WorkspaceID string `json:"workspace_id" binding:"required"`
}
