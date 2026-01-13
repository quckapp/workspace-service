package models

import (
	"time"

	"github.com/google/uuid"
)

type Workspace struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Slug        string     `json:"slug" db:"slug"`
	Description *string    `json:"description" db:"description"`
	IconURL     *string    `json:"icon_url" db:"icon_url"`
	OwnerID     uuid.UUID  `json:"owner_id" db:"owner_id"`
	Plan        string     `json:"plan" db:"plan"` // free, pro, enterprise
	Settings    JSON       `json:"settings" db:"settings"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type WorkspaceMember struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	Role        string     `json:"role" db:"role"` // owner, admin, member, guest
	JoinedAt    time.Time  `json:"joined_at" db:"joined_at"`
	InvitedBy   *uuid.UUID `json:"invited_by" db:"invited_by"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type WorkspaceInvite struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	Email       string     `json:"email" db:"email"`
	Role        string     `json:"role" db:"role"`
	Token       string     `json:"token" db:"token"`
	InvitedBy   uuid.UUID  `json:"invited_by" db:"invited_by"`
	ExpiresAt   time.Time  `json:"expires_at" db:"expires_at"`
	AcceptedAt  *time.Time `json:"accepted_at" db:"accepted_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

type JSON map[string]interface{}

// DTOs
type CreateWorkspaceRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=100"`
	Slug        string  `json:"slug" binding:"required,min=2,max=50"`
	Description *string `json:"description"`
}

type UpdateWorkspaceRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IconURL     *string `json:"icon_url"`
	Settings    JSON    `json:"settings"`
}

type InviteMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin member guest"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member guest"`
}

type WorkspaceResponse struct {
	Workspace    *Workspace          `json:"workspace"`
	MemberCount  int                 `json:"member_count"`
	ChannelCount int                 `json:"channel_count"`
	MyRole       string              `json:"my_role,omitempty"`
}

type WorkspacesListResponse struct {
	Workspaces []*WorkspaceResponse `json:"workspaces"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PerPage    int                  `json:"per_page"`
}
