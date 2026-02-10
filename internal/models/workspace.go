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

type WorkspaceInviteCode struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	Code        string     `json:"code" db:"code"`
	Role        string     `json:"role" db:"role"`
	MaxUses     int        `json:"max_uses" db:"max_uses"`
	UseCount    int        `json:"use_count" db:"use_count"`
	CreatedBy   uuid.UUID  `json:"created_by" db:"created_by"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateInviteCodeRequest struct {
	Role    string `json:"role" binding:"required,oneof=admin member guest"`
	MaxUses int    `json:"max_uses"`
}

// ── Activity Log ──

type ActivityLog struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	ActorID     uuid.UUID `json:"actor_id" db:"actor_id"`
	Action      string    `json:"action" db:"action"`
	EntityType  string    `json:"entity_type" db:"entity_type"`
	EntityID    string    `json:"entity_id" db:"entity_id"`
	Details     JSON      `json:"details" db:"details"`
	IPAddress   string    `json:"ip_address,omitempty" db:"ip_address"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type ActivityLogResponse struct {
	Activities []*ActivityLog `json:"activities"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
}

// ── Member Profile ──

type MemberProfile struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	DisplayName *string    `json:"display_name" db:"display_name"`
	Title       *string    `json:"title" db:"title"`
	StatusText  *string    `json:"status_text" db:"status_text"`
	StatusEmoji *string    `json:"status_emoji" db:"status_emoji"`
	Timezone    *string    `json:"timezone" db:"timezone"`
	IsOnline    bool       `json:"is_online" db:"is_online"`
	LastSeenAt  *time.Time `json:"last_seen_at" db:"last_seen_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type UpdateMemberProfileRequest struct {
	DisplayName *string `json:"display_name"`
	Title       *string `json:"title"`
	StatusText  *string `json:"status_text"`
	StatusEmoji *string `json:"status_emoji"`
	Timezone    *string `json:"timezone"`
}

type MemberWithProfile struct {
	WorkspaceMember
	Profile *MemberProfile `json:"profile,omitempty"`
}

// ── Custom Roles ──

type WorkspaceRole struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	Name        string    `json:"name" db:"name"`
	Color       *string   `json:"color" db:"color"`
	Priority    int       `json:"priority" db:"priority"`
	Permissions JSON      `json:"permissions" db:"permissions"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateRoleRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=50"`
	Color       *string `json:"color"`
	Priority    int     `json:"priority"`
	Permissions JSON    `json:"permissions" binding:"required"`
}

type UpdateRoleRequest struct {
	Name        *string `json:"name"`
	Color       *string `json:"color"`
	Priority    *int    `json:"priority"`
	Permissions JSON    `json:"permissions"`
}

// ── Workspace Analytics ──

type WorkspaceAnalytics struct {
	MemberGrowth     []DailyCount   `json:"member_growth"`
	ActiveMembers    int            `json:"active_members_30d"`
	TopContributors  []ContributorStat `json:"top_contributors"`
	RoleDistribution map[string]int `json:"role_distribution"`
	JoinMethodStats  map[string]int `json:"join_method_stats"`
}

type DailyCount struct {
	Date  string `json:"date" db:"date"`
	Count int    `json:"count" db:"count"`
}

type ContributorStat struct {
	UserID  uuid.UUID `json:"user_id" db:"user_id"`
	Actions int       `json:"actions" db:"actions"`
}

// ── Workspace Discovery ──

type WorkspaceSearchParams struct {
	Query   string `form:"q"`
	Page    int    `form:"page,default=1"`
	PerPage int    `form:"per_page,default=20"`
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

type TransferOwnershipRequest struct {
	NewOwnerID string `json:"new_owner_id" binding:"required"`
}

type JoinByCodeRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

type BulkInviteRequest struct {
	Invites []InviteMemberRequest `json:"invites" binding:"required,min=1,max=50"`
}

// Response DTOs

type WorkspaceResponse struct {
	Workspace    *Workspace `json:"workspace"`
	MemberCount  int        `json:"member_count"`
	ChannelCount int        `json:"channel_count"`
	MyRole       string     `json:"my_role,omitempty"`
}

type WorkspacesListResponse struct {
	Workspaces []*WorkspaceResponse `json:"workspaces"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PerPage    int                  `json:"per_page"`
}

type WorkspaceStats struct {
	MemberCount  int            `json:"member_count"`
	ChannelCount int            `json:"channel_count"`
	InviteCount  int            `json:"invite_count"`
	RoleCounts   map[string]int `json:"role_counts"`
	CreatedAt    time.Time      `json:"created_at"`
	Plan         string         `json:"plan"`
}

type BulkInviteResponse struct {
	Successful []string `json:"successful"`
	Failed     []struct {
		Email  string `json:"email"`
		Reason string `json:"reason"`
	} `json:"failed"`
}

// ── Workspace Templates ──

type WorkspaceTemplate struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Description     *string   `json:"description" db:"description"`
	CreatedBy       uuid.UUID `json:"created_by" db:"created_by"`
	DefaultRoles    JSON      `json:"default_roles" db:"default_roles"`
	DefaultChannels JSON      `json:"default_channels" db:"default_channels"`
	DefaultSettings JSON      `json:"default_settings" db:"default_settings"`
	IsPublic        bool      `json:"is_public" db:"is_public"`
	UseCount        int       `json:"use_count" db:"use_count"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type CreateTemplateFromWorkspaceRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=100"`
	Description *string `json:"description"`
	IsPublic    bool    `json:"is_public"`
}

type CreateWorkspaceFromTemplateRequest struct {
	TemplateID string  `json:"template_id" binding:"required"`
	Name       string  `json:"name" binding:"required,min=2,max=100"`
	Slug       string  `json:"slug" binding:"required,min=2,max=50"`
}

type UpdateTemplateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsPublic    *bool   `json:"is_public"`
}

// ── Member Preferences ──

type WorkspaceMemberPreference struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	WorkspaceID        uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	UserID             uuid.UUID  `json:"user_id" db:"user_id"`
	NotificationLevel  string     `json:"notification_level" db:"notification_level"`
	EmailNotifications bool       `json:"email_notifications" db:"email_notifications"`
	MuteUntil          *time.Time `json:"mute_until" db:"mute_until"`
	SidebarPosition    int        `json:"sidebar_position" db:"sidebar_position"`
	Theme              *string    `json:"theme" db:"theme"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

type UpdatePreferencesRequest struct {
	NotificationLevel  *string    `json:"notification_level" binding:"omitempty,oneof=all mentions none"`
	EmailNotifications *bool      `json:"email_notifications"`
	MuteUntil          *time.Time `json:"mute_until"`
	SidebarPosition    *int       `json:"sidebar_position"`
	Theme              *string    `json:"theme"`
}

// ── Workspace Tags ──

type WorkspaceTag struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	Name        string    `json:"name" db:"name"`
	Color       *string   `json:"color" db:"color"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateTagRequest struct {
	Name  string  `json:"name" binding:"required,min=1,max=50"`
	Color *string `json:"color"`
}

type UpdateTagRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

// ── Workspace Moderation ──

type WorkspaceBan struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	BannedBy    uuid.UUID  `json:"banned_by" db:"banned_by"`
	Reason      *string    `json:"reason" db:"reason"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	IsPermanent bool       `json:"is_permanent" db:"is_permanent"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

type WorkspaceMute struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	MutedBy     uuid.UUID  `json:"muted_by" db:"muted_by"`
	Reason      *string    `json:"reason" db:"reason"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

type BanMemberRequest struct {
	Reason      *string    `json:"reason"`
	ExpiresAt   *time.Time `json:"expires_at"`
	IsPermanent bool       `json:"is_permanent"`
}

type MuteMemberRequest struct {
	Reason    *string    `json:"reason"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type ModerationHistoryResponse struct {
	Bans  []*WorkspaceBan  `json:"bans"`
	Mutes []*WorkspaceMute `json:"mutes"`
}

// ── Workspace Announcements ──

type WorkspaceAnnouncement struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	Title       string     `json:"title" db:"title"`
	Content     string     `json:"content" db:"content"`
	Priority    string     `json:"priority" db:"priority"`
	AuthorID    uuid.UUID  `json:"author_id" db:"author_id"`
	IsPinned    bool       `json:"is_pinned" db:"is_pinned"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateAnnouncementRequest struct {
	Title     string     `json:"title" binding:"required,min=1,max=200"`
	Content   string     `json:"content" binding:"required,min=1"`
	Priority  string     `json:"priority" binding:"required,oneof=normal important urgent"`
	IsPinned  bool       `json:"is_pinned"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type UpdateAnnouncementRequest struct {
	Title     *string    `json:"title"`
	Content   *string    `json:"content"`
	Priority  *string    `json:"priority" binding:"omitempty,oneof=normal important urgent"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type PinAnnouncementRequest struct {
	IsPinned bool `json:"is_pinned"`
}

// ── Workspace Webhooks ──

type WorkspaceWebhook struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	WorkspaceID     uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	Name            string     `json:"name" db:"name"`
	URL             string     `json:"url" db:"url"`
	Secret          string     `json:"secret,omitempty" db:"secret"`
	Events          JSON       `json:"events" db:"events"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	CreatedBy       uuid.UUID  `json:"created_by" db:"created_by"`
	LastTriggeredAt *time.Time `json:"last_triggered_at" db:"last_triggered_at"`
	FailureCount    int        `json:"failure_count" db:"failure_count"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateWebhookRequest struct {
	Name   string   `json:"name" binding:"required,min=1,max=100"`
	URL    string   `json:"url" binding:"required,url"`
	Events []string `json:"events" binding:"required,min=1"`
}

type UpdateWebhookRequest struct {
	Name     *string  `json:"name"`
	URL      *string  `json:"url"`
	Events   []string `json:"events"`
	IsActive *bool    `json:"is_active"`
}

// ── Workspace Favorites ──

type WorkspaceFavorite struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	Position    int       `json:"position" db:"position"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type ReorderFavoritesRequest struct {
	WorkspaceIDs []string `json:"workspace_ids" binding:"required,min=1"`
}

// ── Member Notes ──

type MemberNote struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	TargetID    uuid.UUID `json:"target_id" db:"target_id"`
	AuthorID    uuid.UUID `json:"author_id" db:"author_id"`
	Content     string    `json:"content" db:"content"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateMemberNoteRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

type UpdateMemberNoteRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

// ── Scheduled Actions ──

type ScheduledAction struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	ActionType  string     `json:"action_type" db:"action_type"`
	Payload     JSON       `json:"payload" db:"payload"`
	ScheduledAt time.Time  `json:"scheduled_at" db:"scheduled_at"`
	ExecutedAt  *time.Time `json:"executed_at" db:"executed_at"`
	Status      string     `json:"status" db:"status"`
	CreatedBy   uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateScheduledActionRequest struct {
	ActionType  string    `json:"action_type" binding:"required,oneof=archive unarchive lock unlock send_reminder"`
	Payload     JSON      `json:"payload"`
	ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
}

type UpdateScheduledActionRequest struct {
	Payload     JSON       `json:"payload"`
	ScheduledAt *time.Time `json:"scheduled_at"`
}

// ── Usage Quotas ──

type WorkspaceQuota struct {
	ID              uuid.UUID `json:"id" db:"id"`
	WorkspaceID     uuid.UUID `json:"workspace_id" db:"workspace_id"`
	MaxMembers      int       `json:"max_members" db:"max_members"`
	MaxChannels     int       `json:"max_channels" db:"max_channels"`
	MaxStorageMB    int       `json:"max_storage_mb" db:"max_storage_mb"`
	MaxInviteCodes  int       `json:"max_invite_codes" db:"max_invite_codes"`
	MaxWebhooks     int       `json:"max_webhooks" db:"max_webhooks"`
	MaxRoles        int       `json:"max_roles" db:"max_roles"`
	CurrentMembers  int       `json:"current_members" db:"current_members"`
	CurrentChannels int       `json:"current_channels" db:"current_channels"`
	CurrentStorageMB int      `json:"current_storage_mb" db:"current_storage_mb"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type UpdateQuotaRequest struct {
	MaxMembers     *int `json:"max_members"`
	MaxChannels    *int `json:"max_channels"`
	MaxStorageMB   *int `json:"max_storage_mb"`
	MaxInviteCodes *int `json:"max_invite_codes"`
	MaxWebhooks    *int `json:"max_webhooks"`
	MaxRoles       *int `json:"max_roles"`
}

type QuotaUsageResponse struct {
	Quota   *WorkspaceQuota `json:"quota"`
	Usage   map[string]int  `json:"usage"`
	Limits  map[string]int  `json:"limits"`
	Percent map[string]int  `json:"percent_used"`
}

// ── Audit Export ──

type AuditExportRequest struct {
	StartDate  *time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate    *time.Time `form:"end_date" time_format:"2006-01-02"`
	Format     string     `form:"format,default=json"`
	ActionType string     `form:"action_type"`
}

type AuditExportResponse struct {
	Activities []*ActivityLog `json:"activities"`
	Total      int64          `json:"total"`
	StartDate  *time.Time     `json:"start_date,omitempty"`
	EndDate    *time.Time     `json:"end_date,omitempty"`
	ExportedAt time.Time      `json:"exported_at"`
}

// ── Workspace Archive / Restore ──

type ArchiveWorkspaceRequest struct {
	Reason string `json:"reason"`
}

// ── Workspace Cloning ──

type CloneWorkspaceRequest struct {
	Name            string `json:"name" binding:"required,min=2,max=100"`
	Slug            string `json:"slug" binding:"required,min=2,max=50"`
	IncludeRoles    bool   `json:"include_roles"`
	IncludeSettings bool   `json:"include_settings"`
	IncludeTags     bool   `json:"include_tags"`
}

// ── Workspace Pinned Items ──

type WorkspacePinnedItem struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	ItemType    string    `json:"item_type" db:"item_type"` // message, file, link, note
	ItemID      *string   `json:"item_id" db:"item_id"`
	Title       string    `json:"title" db:"title"`
	Content     *string   `json:"content" db:"content"`
	URL         *string   `json:"url" db:"url"`
	PinnedBy    uuid.UUID `json:"pinned_by" db:"pinned_by"`
	Position    int       `json:"position" db:"position"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreatePinnedItemRequest struct {
	ItemType string  `json:"item_type" binding:"required,oneof=message file link note"`
	ItemID   *string `json:"item_id"`
	Title    string  `json:"title" binding:"required,min=1,max=200"`
	Content  *string `json:"content"`
	URL      *string `json:"url"`
}

type UpdatePinnedItemRequest struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
	URL     *string `json:"url"`
}

type ReorderPinsRequest struct {
	PinIDs []string `json:"pin_ids" binding:"required,min=1"`
}

// ── Member Groups / Teams ──

type MemberGroup struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	Color       *string   `json:"color" db:"color"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	MemberCount int       `json:"member_count" db:"member_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type MemberGroupMembership struct {
	ID        uuid.UUID `json:"id" db:"id"`
	GroupID   uuid.UUID `json:"group_id" db:"group_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	AddedBy   uuid.UUID `json:"added_by" db:"added_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateGroupRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=100"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
}

type UpdateGroupRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
}

type AddGroupMembersRequest struct {
	UserIDs []string `json:"user_ids" binding:"required,min=1"`
}

type MemberGroupWithMembers struct {
	MemberGroup
	Members []uuid.UUID `json:"members"`
}

// ── Workspace Custom Fields ──

type WorkspaceCustomField struct {
	ID           uuid.UUID `json:"id" db:"id"`
	WorkspaceID  uuid.UUID `json:"workspace_id" db:"workspace_id"`
	Name         string    `json:"name" db:"name"`
	FieldType    string    `json:"field_type" db:"field_type"` // text, number, date, boolean, select
	Options      JSON      `json:"options" db:"options"`       // for select type
	DefaultValue *string   `json:"default_value" db:"default_value"`
	IsRequired   bool      `json:"is_required" db:"is_required"`
	Position     int       `json:"position" db:"position"`
	CreatedBy    uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type WorkspaceCustomFieldValue struct {
	ID        uuid.UUID `json:"id" db:"id"`
	FieldID   uuid.UUID `json:"field_id" db:"field_id"`
	EntityID  uuid.UUID `json:"entity_id" db:"entity_id"` // workspace_id or member_id
	Value     string    `json:"value" db:"value"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreateCustomFieldRequest struct {
	Name         string  `json:"name" binding:"required,min=1,max=100"`
	FieldType    string  `json:"field_type" binding:"required,oneof=text number date boolean select"`
	Options      JSON    `json:"options"`
	DefaultValue *string `json:"default_value"`
	IsRequired   bool    `json:"is_required"`
}

type UpdateCustomFieldRequest struct {
	Name         *string `json:"name"`
	Options      JSON    `json:"options"`
	DefaultValue *string `json:"default_value"`
	IsRequired   *bool   `json:"is_required"`
}

type SetCustomFieldValueRequest struct {
	Value string `json:"value" binding:"required"`
}

type CustomFieldWithValue struct {
	WorkspaceCustomField
	Value *string `json:"value,omitempty"`
}

// ── Workspace Reactions ──

type WorkspaceReaction struct {
	ID         uuid.UUID `json:"id" db:"id"`
	EntityType string    `json:"entity_type" db:"entity_type"` // announcement, pin, note
	EntityID   uuid.UUID `json:"entity_id" db:"entity_id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	Emoji      string    `json:"emoji" db:"emoji"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type AddReactionRequest struct {
	EntityType string `json:"entity_type" binding:"required,oneof=announcement pin note"`
	EntityID   string `json:"entity_id" binding:"required"`
	Emoji      string `json:"emoji" binding:"required,min=1,max=10"`
}

type ReactionSummary struct {
	Emoji string      `json:"emoji" db:"emoji"`
	Count int         `json:"count" db:"count"`
	Users []uuid.UUID `json:"users,omitempty"`
}

// ── Workspace Bookmarks ──

type WorkspaceBookmark struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Title       string    `json:"title" db:"title"`
	URL         *string   `json:"url" db:"url"`
	EntityType  *string   `json:"entity_type" db:"entity_type"` // message, channel, file
	EntityID    *string   `json:"entity_id" db:"entity_id"`
	Notes       *string   `json:"notes" db:"notes"`
	FolderName  *string   `json:"folder_name" db:"folder_name"`
	Position    int       `json:"position" db:"position"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateBookmarkRequest struct {
	Title      string  `json:"title" binding:"required,min=1,max=200"`
	URL        *string `json:"url"`
	EntityType *string `json:"entity_type" binding:"omitempty,oneof=message channel file"`
	EntityID   *string `json:"entity_id"`
	Notes      *string `json:"notes"`
	FolderName *string `json:"folder_name"`
}

type UpdateBookmarkRequest struct {
	Title      *string `json:"title"`
	URL        *string `json:"url"`
	Notes      *string `json:"notes"`
	FolderName *string `json:"folder_name"`
}

// ── Invitation Tracking ──

type InvitationHistory struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	InviterID   uuid.UUID  `json:"inviter_id" db:"inviter_id"`
	InviteeEmail string   `json:"invitee_email" db:"invitee_email"`
	InviteeID   *uuid.UUID `json:"invitee_id" db:"invitee_id"`
	Method      string     `json:"method" db:"method"` // email, code, link
	Role        string     `json:"role" db:"role"`
	Status      string     `json:"status" db:"status"` // pending, accepted, expired, revoked
	AcceptedAt  *time.Time `json:"accepted_at" db:"accepted_at"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

type InvitationHistoryResponse struct {
	Invitations []*InvitationHistory `json:"invitations"`
	Total       int64                `json:"total"`
	Page        int                  `json:"page"`
	PerPage     int                  `json:"per_page"`
}

type InvitationStats struct {
	TotalSent     int            `json:"total_sent"`
	TotalAccepted int            `json:"total_accepted"`
	TotalPending  int            `json:"total_pending"`
	TotalExpired  int            `json:"total_expired"`
	ByMethod      map[string]int `json:"by_method"`
}

// ── Workspace Access Logs ──

type WorkspaceAccessLog struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Action      string    `json:"action" db:"action"` // view, api_call, settings_access
	Resource    string    `json:"resource" db:"resource"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	UserAgent   *string   `json:"user_agent" db:"user_agent"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type AccessLogResponse struct {
	Logs    []*WorkspaceAccessLog `json:"logs"`
	Total   int64                 `json:"total"`
	Page    int                   `json:"page"`
	PerPage int                   `json:"per_page"`
}

type AccessLogStats struct {
	TotalAccesses   int            `json:"total_accesses"`
	UniqueUsers     int            `json:"unique_users"`
	TopResources    []ResourceStat `json:"top_resources"`
	AccessesByDay   []DailyCount   `json:"accesses_by_day"`
}

type ResourceStat struct {
	Resource string `json:"resource" db:"resource"`
	Count    int    `json:"count" db:"count"`
}

// ── Workspace Feature Flags ──

type WorkspaceFeatureFlag struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	Key         string     `json:"key" db:"key"`
	Enabled     bool       `json:"enabled" db:"enabled"`
	Description *string    `json:"description" db:"description"`
	Metadata    JSON       `json:"metadata" db:"metadata"`
	CreatedBy   uuid.UUID  `json:"created_by" db:"created_by"`
	UpdatedBy   *uuid.UUID `json:"updated_by" db:"updated_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateFeatureFlagRequest struct {
	Key         string  `json:"key" binding:"required,min=1,max=100"`
	Enabled     bool    `json:"enabled"`
	Description *string `json:"description"`
	Metadata    JSON    `json:"metadata"`
}

type UpdateFeatureFlagRequest struct {
	Enabled     *bool   `json:"enabled"`
	Description *string `json:"description"`
	Metadata    JSON    `json:"metadata"`
}

type FeatureFlagCheckResponse struct {
	Key     string `json:"key"`
	Enabled bool   `json:"enabled"`
}

// ── Workspace Integrations ──

type WorkspaceIntegration struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	WorkspaceID  uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	Provider     string     `json:"provider" db:"provider"` // slack, github, jira, discord, etc.
	Name         string     `json:"name" db:"name"`
	Status       string     `json:"status" db:"status"` // active, inactive, error
	Config       JSON       `json:"config" db:"config"`
	Credentials  *string    `json:"-" db:"credentials"` // encrypted, never exposed in JSON
	WebhookURL   *string    `json:"webhook_url" db:"webhook_url"`
	LastSyncAt   *time.Time `json:"last_sync_at" db:"last_sync_at"`
	ErrorMessage *string    `json:"error_message" db:"error_message"`
	CreatedBy    uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateIntegrationRequest struct {
	Provider string  `json:"provider" binding:"required,oneof=slack github jira discord linear notion"`
	Name     string  `json:"name" binding:"required,min=1,max=100"`
	Config   JSON    `json:"config"`
	Credentials *string `json:"credentials"`
}

type UpdateIntegrationRequest struct {
	Name        *string `json:"name"`
	Status      *string `json:"status" binding:"omitempty,oneof=active inactive"`
	Config      JSON    `json:"config"`
	Credentials *string `json:"credentials"`
}

// ── Workspace Labels ──

type WorkspaceLabel struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	Name        string    `json:"name" db:"name"`
	Color       string    `json:"color" db:"color"` // hex color code
	Description *string   `json:"description" db:"description"`
	Position    int       `json:"position" db:"position"`
	UsageCount  int       `json:"usage_count" db:"usage_count"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateLabelRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=50"`
	Color       string  `json:"color" binding:"required,min=4,max=7"`
	Description *string `json:"description"`
}

type UpdateLabelRequest struct {
	Name        *string `json:"name"`
	Color       *string `json:"color"`
	Description *string `json:"description"`
}

// ── Member Activity Streaks ──

type MemberActivityStreak struct {
	ID              uuid.UUID `json:"id" db:"id"`
	WorkspaceID     uuid.UUID `json:"workspace_id" db:"workspace_id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	CurrentStreak   int       `json:"current_streak" db:"current_streak"`
	LongestStreak   int       `json:"longest_streak" db:"longest_streak"`
	TotalActiveDays int       `json:"total_active_days" db:"total_active_days"`
	ActivityScore   float64   `json:"activity_score" db:"activity_score"`
	LastActiveDate  string    `json:"last_active_date" db:"last_active_date"` // YYYY-MM-DD
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type StreakLeaderboard struct {
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	CurrentStreak int       `json:"current_streak" db:"current_streak"`
	LongestStreak int       `json:"longest_streak" db:"longest_streak"`
	ActivityScore float64   `json:"activity_score" db:"activity_score"`
}

// ── Onboarding Checklists ──

type OnboardingChecklist struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	Title       string    `json:"title" db:"title"`
	Description *string   `json:"description" db:"description"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type OnboardingStep struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ChecklistID uuid.UUID `json:"checklist_id" db:"checklist_id"`
	Title       string    `json:"title" db:"title"`
	Description *string   `json:"description" db:"description"`
	ActionType  string    `json:"action_type" db:"action_type"` // link, task, acknowledgement
	ActionData  *string   `json:"action_data" db:"action_data"`
	Position    int       `json:"position" db:"position"`
	IsRequired  bool      `json:"is_required" db:"is_required"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type OnboardingProgress struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	StepID      uuid.UUID  `json:"step_id" db:"step_id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

type CreateChecklistRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=200"`
	Description *string `json:"description"`
}

type UpdateChecklistRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

type AddStepRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=200"`
	Description *string `json:"description"`
	ActionType  string  `json:"action_type" binding:"required,oneof=link task acknowledgement"`
	ActionData  *string `json:"action_data"`
	IsRequired  bool    `json:"is_required"`
}

type ChecklistWithSteps struct {
	OnboardingChecklist
	Steps []OnboardingStep `json:"steps"`
}

type UserOnboardingStatus struct {
	Checklist      OnboardingChecklist `json:"checklist"`
	Steps          []StepWithProgress  `json:"steps"`
	CompletedCount int                 `json:"completed_count"`
	TotalSteps     int                 `json:"total_steps"`
	IsComplete     bool                `json:"is_complete"`
}

type StepWithProgress struct {
	OnboardingStep
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completed_at"`
}

// ── Compliance Policies ──

type CompliancePolicy struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	PolicyType  string    `json:"policy_type" db:"policy_type"` // data_retention, access_control, content_policy, privacy
	Rules       JSON      `json:"rules" db:"rules"`
	Severity    string    `json:"severity" db:"severity"` // info, warning, critical
	IsEnforced  bool      `json:"is_enforced" db:"is_enforced"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreatePolicyRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=200"`
	Description *string `json:"description"`
	PolicyType  string  `json:"policy_type" binding:"required,oneof=data_retention access_control content_policy privacy"`
	Rules       JSON    `json:"rules"`
	Severity    string  `json:"severity" binding:"required,oneof=info warning critical"`
	IsEnforced  bool    `json:"is_enforced"`
}

type UpdatePolicyRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Rules       JSON    `json:"rules"`
	Severity    *string `json:"severity" binding:"omitempty,oneof=info warning critical"`
	IsEnforced  *bool   `json:"is_enforced"`
}

type PolicyAcknowledgement struct {
	ID        uuid.UUID `json:"id" db:"id"`
	PolicyID  uuid.UUID `json:"policy_id" db:"policy_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	AckedAt   time.Time `json:"acked_at" db:"acked_at"`
}

type PolicyComplianceStatus struct {
	Policy           CompliancePolicy `json:"policy"`
	TotalMembers     int              `json:"total_members"`
	AcknowledgedCount int             `json:"acknowledged_count"`
	ComplianceRate   float64          `json:"compliance_rate"`
}
