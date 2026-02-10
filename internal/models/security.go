package models

import (
	"time"

	"github.com/google/uuid"
)

// ── Advanced Security ──

type IPAllowlistEntry struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	IPAddress   string    `json:"ip_address" db:"ip_address"` // CIDR notation supported
	Label       *string   `json:"label" db:"label"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type AddIPAllowlistRequest struct {
	IPAddress string  `json:"ip_address" binding:"required"`
	Label     *string `json:"label"`
}

type UpdateIPAllowlistRequest struct {
	IPAddress *string `json:"ip_address"`
	Label     *string `json:"label"`
	IsActive  *bool   `json:"is_active"`
}

type WorkspaceSession struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	WorkspaceID  uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	SessionToken string     `json:"session_token" db:"session_token"`
	IPAddress    string     `json:"ip_address" db:"ip_address"`
	UserAgent    *string    `json:"user_agent" db:"user_agent"`
	DeviceType   *string    `json:"device_type" db:"device_type"` // desktop, mobile, tablet, api
	Location     *string    `json:"location" db:"location"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	LastActiveAt time.Time  `json:"last_active_at" db:"last_active_at"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

type WorkspaceSecurityPolicy struct {
	ID                     uuid.UUID `json:"id" db:"id"`
	WorkspaceID            uuid.UUID `json:"workspace_id" db:"workspace_id"`
	RequireTwoFactor       bool      `json:"require_two_factor" db:"require_two_factor"`
	SessionTimeoutMinutes  int       `json:"session_timeout_minutes" db:"session_timeout_minutes"`
	MaxSessionsPerUser     int       `json:"max_sessions_per_user" db:"max_sessions_per_user"`
	PasswordMinLength      int       `json:"password_min_length" db:"password_min_length"`
	RequireSpecialChars    bool      `json:"require_special_chars" db:"require_special_chars"`
	IPAllowlistEnabled     bool      `json:"ip_allowlist_enabled" db:"ip_allowlist_enabled"`
	AllowGuestAccess       bool      `json:"allow_guest_access" db:"allow_guest_access"`
	AllowExternalSharing   bool      `json:"allow_external_sharing" db:"allow_external_sharing"`
	DataRetentionDays      int       `json:"data_retention_days" db:"data_retention_days"`
	RequireEmailVerification bool    `json:"require_email_verification" db:"require_email_verification"`
	AllowedDomains         JSON      `json:"allowed_domains" db:"allowed_domains"` // restrict signups to certain email domains
	UpdatedBy              uuid.UUID `json:"updated_by" db:"updated_by"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
}

type UpdateSecurityPolicyRequest struct {
	RequireTwoFactor       *bool   `json:"require_two_factor"`
	SessionTimeoutMinutes  *int    `json:"session_timeout_minutes" binding:"omitempty,min=5,max=43200"`
	MaxSessionsPerUser     *int    `json:"max_sessions_per_user" binding:"omitempty,min=1,max=100"`
	PasswordMinLength      *int    `json:"password_min_length" binding:"omitempty,min=6,max=128"`
	RequireSpecialChars    *bool   `json:"require_special_chars"`
	IPAllowlistEnabled     *bool   `json:"ip_allowlist_enabled"`
	AllowGuestAccess       *bool   `json:"allow_guest_access"`
	AllowExternalSharing   *bool   `json:"allow_external_sharing"`
	DataRetentionDays      *int    `json:"data_retention_days" binding:"omitempty,min=0,max=3650"`
	RequireEmailVerification *bool `json:"require_email_verification"`
	AllowedDomains         []string `json:"allowed_domains"`
}

type SecurityAuditEntry struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	EventType   string    `json:"event_type" db:"event_type"` // login, logout, failed_login, policy_change, session_revoked, ip_blocked
	Description string    `json:"description" db:"description"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	UserAgent   *string   `json:"user_agent" db:"user_agent"`
	Severity    string    `json:"severity" db:"severity"` // info, warning, critical
	Metadata    JSON      `json:"metadata" db:"metadata"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type SecurityOverview struct {
	Policy           *WorkspaceSecurityPolicy `json:"policy"`
	ActiveSessions   int                      `json:"active_sessions"`
	IPAllowlistCount int                      `json:"ip_allowlist_count"`
	TwoFactorAdoption float64                 `json:"two_factor_adoption"` // percentage
	RecentAlerts     []*SecurityAuditEntry    `json:"recent_alerts"`
	RiskLevel        string                   `json:"risk_level"` // low, medium, high
}

type RevokeSessionsRequest struct {
	UserID    *string `json:"user_id"`
	AllUsers  bool    `json:"all_users"`
	ExceptCurrent bool `json:"except_current"`
}
