package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type SecurityRepository struct {
	db *sqlx.DB
}

func NewSecurityRepository(db *sqlx.DB) *SecurityRepository {
	return &SecurityRepository{db: db}
}

// IP Allowlist
func (r *SecurityRepository) AddIPEntry(ctx context.Context, entry *models.IPAllowlistEntry) error {
	query := `INSERT INTO workspace_ip_allowlist (id, workspace_id, ip_address, label, created_by, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, entry.ID, entry.WorkspaceID, entry.IPAddress, entry.Label, entry.CreatedBy, entry.IsActive, entry.CreatedAt, entry.UpdatedAt)
	return err
}

func (r *SecurityRepository) GetIPEntry(ctx context.Context, id uuid.UUID) (*models.IPAllowlistEntry, error) {
	var entry models.IPAllowlistEntry
	err := r.db.GetContext(ctx, &entry, "SELECT * FROM workspace_ip_allowlist WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &entry, err
}

func (r *SecurityRepository) ListIPEntries(ctx context.Context, workspaceID uuid.UUID) ([]*models.IPAllowlistEntry, error) {
	var entries []*models.IPAllowlistEntry
	err := r.db.SelectContext(ctx, &entries, "SELECT * FROM workspace_ip_allowlist WHERE workspace_id = ? ORDER BY created_at DESC", workspaceID)
	return entries, err
}

func (r *SecurityRepository) UpdateIPEntry(ctx context.Context, entry *models.IPAllowlistEntry) error {
	query := `UPDATE workspace_ip_allowlist SET ip_address = ?, label = ?, is_active = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, entry.IPAddress, entry.Label, entry.IsActive, time.Now(), entry.ID)
	return err
}

func (r *SecurityRepository) DeleteIPEntry(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM workspace_ip_allowlist WHERE id = ?", id)
	return err
}

func (r *SecurityRepository) CountIPEntries(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM workspace_ip_allowlist WHERE workspace_id = ? AND is_active = TRUE", workspaceID)
	return count, err
}

// Sessions
func (r *SecurityRepository) CreateSession(ctx context.Context, session *models.WorkspaceSession) error {
	query := `INSERT INTO workspace_sessions (id, workspace_id, user_id, session_token, ip_address, user_agent, device_type, location, is_active, last_active_at, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, session.ID, session.WorkspaceID, session.UserID, session.SessionToken, session.IPAddress, session.UserAgent, session.DeviceType, session.Location, session.IsActive, session.LastActiveAt, session.ExpiresAt, session.CreatedAt)
	return err
}

func (r *SecurityRepository) GetSession(ctx context.Context, id uuid.UUID) (*models.WorkspaceSession, error) {
	var session models.WorkspaceSession
	err := r.db.GetContext(ctx, &session, "SELECT * FROM workspace_sessions WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &session, err
}

func (r *SecurityRepository) ListUserSessions(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.WorkspaceSession, error) {
	var sessions []*models.WorkspaceSession
	err := r.db.SelectContext(ctx, &sessions, "SELECT * FROM workspace_sessions WHERE workspace_id = ? AND user_id = ? AND is_active = TRUE ORDER BY last_active_at DESC", workspaceID, userID)
	return sessions, err
}

func (r *SecurityRepository) ListAllSessions(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*models.WorkspaceSession, error) {
	var sessions []*models.WorkspaceSession
	err := r.db.SelectContext(ctx, &sessions, "SELECT * FROM workspace_sessions WHERE workspace_id = ? AND is_active = TRUE ORDER BY last_active_at DESC LIMIT ? OFFSET ?", workspaceID, limit, offset)
	return sessions, err
}

func (r *SecurityRepository) RevokeSession(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE workspace_sessions SET is_active = FALSE WHERE id = ?", id)
	return err
}

func (r *SecurityRepository) RevokeUserSessions(ctx context.Context, workspaceID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE workspace_sessions SET is_active = FALSE WHERE workspace_id = ? AND user_id = ?", workspaceID, userID)
	return err
}

func (r *SecurityRepository) RevokeAllSessions(ctx context.Context, workspaceID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE workspace_sessions SET is_active = FALSE WHERE workspace_id = ?", workspaceID)
	return err
}

func (r *SecurityRepository) CountActiveSessions(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM workspace_sessions WHERE workspace_id = ? AND is_active = TRUE", workspaceID)
	return count, err
}

func (r *SecurityRepository) CountUserSessions(ctx context.Context, workspaceID, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM workspace_sessions WHERE workspace_id = ? AND user_id = ? AND is_active = TRUE", workspaceID, userID)
	return count, err
}

// Security Policy
func (r *SecurityRepository) GetSecurityPolicy(ctx context.Context, workspaceID uuid.UUID) (*models.WorkspaceSecurityPolicy, error) {
	var policy models.WorkspaceSecurityPolicy
	err := r.db.GetContext(ctx, &policy, "SELECT * FROM workspace_security_policies WHERE workspace_id = ?", workspaceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &policy, err
}

func (r *SecurityRepository) CreateSecurityPolicy(ctx context.Context, policy *models.WorkspaceSecurityPolicy) error {
	query := `INSERT INTO workspace_security_policies (id, workspace_id, require_two_factor, session_timeout_minutes, max_sessions_per_user, password_min_length, require_special_chars, ip_allowlist_enabled, allow_guest_access, allow_external_sharing, data_retention_days, require_email_verification, allowed_domains, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, policy.ID, policy.WorkspaceID, policy.RequireTwoFactor, policy.SessionTimeoutMinutes, policy.MaxSessionsPerUser, policy.PasswordMinLength, policy.RequireSpecialChars, policy.IPAllowlistEnabled, policy.AllowGuestAccess, policy.AllowExternalSharing, policy.DataRetentionDays, policy.RequireEmailVerification, policy.AllowedDomains, policy.UpdatedBy, policy.CreatedAt, policy.UpdatedAt)
	return err
}

func (r *SecurityRepository) UpdateSecurityPolicy(ctx context.Context, policy *models.WorkspaceSecurityPolicy) error {
	query := `UPDATE workspace_security_policies SET require_two_factor = ?, session_timeout_minutes = ?, max_sessions_per_user = ?, password_min_length = ?, require_special_chars = ?, ip_allowlist_enabled = ?, allow_guest_access = ?, allow_external_sharing = ?, data_retention_days = ?, require_email_verification = ?, allowed_domains = ?, updated_by = ?, updated_at = ? WHERE workspace_id = ?`
	_, err := r.db.ExecContext(ctx, query, policy.RequireTwoFactor, policy.SessionTimeoutMinutes, policy.MaxSessionsPerUser, policy.PasswordMinLength, policy.RequireSpecialChars, policy.IPAllowlistEnabled, policy.AllowGuestAccess, policy.AllowExternalSharing, policy.DataRetentionDays, policy.RequireEmailVerification, policy.AllowedDomains, policy.UpdatedBy, time.Now(), policy.WorkspaceID)
	return err
}

// Security Audit
func (r *SecurityRepository) CreateAuditEntry(ctx context.Context, entry *models.SecurityAuditEntry) error {
	query := `INSERT INTO workspace_security_audit (id, workspace_id, user_id, event_type, description, ip_address, user_agent, severity, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, entry.ID, entry.WorkspaceID, entry.UserID, entry.EventType, entry.Description, entry.IPAddress, entry.UserAgent, entry.Severity, entry.Metadata, entry.CreatedAt)
	return err
}

func (r *SecurityRepository) ListAuditEntries(ctx context.Context, workspaceID uuid.UUID, severity string, limit, offset int) ([]*models.SecurityAuditEntry, error) {
	var entries []*models.SecurityAuditEntry
	q := "SELECT * FROM workspace_security_audit WHERE workspace_id = ?"
	args := []interface{}{workspaceID}

	if severity != "" {
		q += " AND severity = ?"
		args = append(args, severity)
	}
	q += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &entries, q, args...)
	return entries, err
}

func (r *SecurityRepository) GetRecentAlerts(ctx context.Context, workspaceID uuid.UUID, limit int) ([]*models.SecurityAuditEntry, error) {
	var entries []*models.SecurityAuditEntry
	err := r.db.SelectContext(ctx, &entries, "SELECT * FROM workspace_security_audit WHERE workspace_id = ? AND severity IN ('warning', 'critical') ORDER BY created_at DESC LIMIT ?", workspaceID, limit)
	return entries, err
}
