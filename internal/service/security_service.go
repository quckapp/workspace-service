package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/quckapp/workspace-service/internal/models"
	"github.com/quckapp/workspace-service/internal/repository"
	"github.com/sirupsen/logrus"
)

var (
	ErrIPEntryNotFound       = errors.New("IP allowlist entry not found")
	ErrSessionNotFound       = errors.New("session not found")
	ErrSecurityPolicyNotFound = errors.New("security policy not found")
)

type SecurityService struct {
	securityRepo *repository.SecurityRepository
	memberRepo   *repository.MemberRepository
	logger       *logrus.Logger
}

func NewSecurityService(securityRepo *repository.SecurityRepository, memberRepo *repository.MemberRepository, logger *logrus.Logger) *SecurityService {
	return &SecurityService{securityRepo: securityRepo, memberRepo: memberRepo, logger: logger}
}

// IP Allowlist
func (s *SecurityService) AddIPEntry(ctx context.Context, workspaceID, userID uuid.UUID, req *models.AddIPAllowlistRequest) (*models.IPAllowlistEntry, error) {
	member, err := s.memberRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
	if err != nil || member == nil {
		return nil, ErrNotMember
	}
	if member.Role != "owner" && member.Role != "admin" {
		return nil, ErrNotAuthorized
	}

	now := time.Now()
	entry := &models.IPAllowlistEntry{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		IPAddress:   req.IPAddress,
		Label:       req.Label,
		CreatedBy:   userID,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.securityRepo.AddIPEntry(ctx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *SecurityService) ListIPEntries(ctx context.Context, workspaceID uuid.UUID) ([]*models.IPAllowlistEntry, error) {
	return s.securityRepo.ListIPEntries(ctx, workspaceID)
}

func (s *SecurityService) UpdateIPEntry(ctx context.Context, workspaceID, userID, entryID uuid.UUID, req *models.UpdateIPAllowlistRequest) (*models.IPAllowlistEntry, error) {
	member, err := s.memberRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
	if err != nil || member == nil {
		return nil, ErrNotMember
	}
	if member.Role != "owner" && member.Role != "admin" {
		return nil, ErrNotAuthorized
	}

	entry, err := s.securityRepo.GetIPEntry(ctx, entryID)
	if err != nil || entry == nil {
		return nil, ErrIPEntryNotFound
	}

	if req.IPAddress != nil {
		entry.IPAddress = *req.IPAddress
	}
	if req.Label != nil {
		entry.Label = req.Label
	}
	if req.IsActive != nil {
		entry.IsActive = *req.IsActive
	}

	if err := s.securityRepo.UpdateIPEntry(ctx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *SecurityService) DeleteIPEntry(ctx context.Context, workspaceID, userID, entryID uuid.UUID) error {
	member, err := s.memberRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
	if err != nil || member == nil {
		return ErrNotMember
	}
	if member.Role != "owner" && member.Role != "admin" {
		return ErrNotAuthorized
	}
	return s.securityRepo.DeleteIPEntry(ctx, entryID)
}

// Sessions
func (s *SecurityService) ListUserSessions(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.WorkspaceSession, error) {
	return s.securityRepo.ListUserSessions(ctx, workspaceID, userID)
}

func (s *SecurityService) ListAllSessions(ctx context.Context, workspaceID, userID uuid.UUID, page, perPage int) ([]*models.WorkspaceSession, error) {
	member, err := s.memberRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
	if err != nil || member == nil {
		return nil, ErrNotMember
	}
	if member.Role != "owner" && member.Role != "admin" {
		return nil, ErrNotAuthorized
	}
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage
	return s.securityRepo.ListAllSessions(ctx, workspaceID, perPage, offset)
}

func (s *SecurityService) RevokeSession(ctx context.Context, workspaceID, userID, sessionID uuid.UUID) error {
	session, err := s.securityRepo.GetSession(ctx, sessionID)
	if err != nil || session == nil {
		return ErrSessionNotFound
	}

	if session.UserID != userID {
		member, err := s.memberRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
		if err != nil || member == nil || (member.Role != "owner" && member.Role != "admin") {
			return ErrNotAuthorized
		}
	}

	return s.securityRepo.RevokeSession(ctx, sessionID)
}

func (s *SecurityService) RevokeSessions(ctx context.Context, workspaceID, userID uuid.UUID, req *models.RevokeSessionsRequest) error {
	member, err := s.memberRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
	if err != nil || member == nil {
		return ErrNotMember
	}
	if member.Role != "owner" && member.Role != "admin" {
		return ErrNotAuthorized
	}

	if req.AllUsers {
		return s.securityRepo.RevokeAllSessions(ctx, workspaceID)
	}

	if req.UserID != nil {
		targetUserID, err := uuid.Parse(*req.UserID)
		if err != nil {
			return err
		}
		return s.securityRepo.RevokeUserSessions(ctx, workspaceID, targetUserID)
	}

	return nil
}

// Security Policy
func (s *SecurityService) GetSecurityPolicy(ctx context.Context, workspaceID uuid.UUID) (*models.WorkspaceSecurityPolicy, error) {
	policy, err := s.securityRepo.GetSecurityPolicy(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		now := time.Now()
		policy = &models.WorkspaceSecurityPolicy{
			ID:                     uuid.New(),
			WorkspaceID:            workspaceID,
			SessionTimeoutMinutes:  1440,
			MaxSessionsPerUser:     10,
			PasswordMinLength:      8,
			AllowGuestAccess:       true,
			AllowExternalSharing:   true,
			DataRetentionDays:      365,
			CreatedAt:              now,
			UpdatedAt:              now,
		}
		if err := s.securityRepo.CreateSecurityPolicy(ctx, policy); err != nil {
			return nil, err
		}
	}
	return policy, nil
}

func (s *SecurityService) UpdateSecurityPolicy(ctx context.Context, workspaceID, userID uuid.UUID, req *models.UpdateSecurityPolicyRequest) (*models.WorkspaceSecurityPolicy, error) {
	member, err := s.memberRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
	if err != nil || member == nil {
		return nil, ErrNotMember
	}
	if member.Role != "owner" && member.Role != "admin" {
		return nil, ErrNotAuthorized
	}

	policy, err := s.GetSecurityPolicy(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	if req.RequireTwoFactor != nil {
		policy.RequireTwoFactor = *req.RequireTwoFactor
	}
	if req.SessionTimeoutMinutes != nil {
		policy.SessionTimeoutMinutes = *req.SessionTimeoutMinutes
	}
	if req.MaxSessionsPerUser != nil {
		policy.MaxSessionsPerUser = *req.MaxSessionsPerUser
	}
	if req.PasswordMinLength != nil {
		policy.PasswordMinLength = *req.PasswordMinLength
	}
	if req.RequireSpecialChars != nil {
		policy.RequireSpecialChars = *req.RequireSpecialChars
	}
	if req.IPAllowlistEnabled != nil {
		policy.IPAllowlistEnabled = *req.IPAllowlistEnabled
	}
	if req.AllowGuestAccess != nil {
		policy.AllowGuestAccess = *req.AllowGuestAccess
	}
	if req.AllowExternalSharing != nil {
		policy.AllowExternalSharing = *req.AllowExternalSharing
	}
	if req.DataRetentionDays != nil {
		policy.DataRetentionDays = *req.DataRetentionDays
	}
	if req.RequireEmailVerification != nil {
		policy.RequireEmailVerification = *req.RequireEmailVerification
	}
	policy.UpdatedBy = userID

	if err := s.securityRepo.UpdateSecurityPolicy(ctx, policy); err != nil {
		return nil, err
	}
	return policy, nil
}

// Security Audit
func (s *SecurityService) ListSecurityAudit(ctx context.Context, workspaceID uuid.UUID, severity string, page, perPage int) ([]*models.SecurityAuditEntry, error) {
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage
	return s.securityRepo.ListAuditEntries(ctx, workspaceID, severity, perPage, offset)
}

// Security Overview
func (s *SecurityService) GetSecurityOverview(ctx context.Context, workspaceID uuid.UUID) (*models.SecurityOverview, error) {
	policy, err := s.GetSecurityPolicy(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	activeSessions, _ := s.securityRepo.CountActiveSessions(ctx, workspaceID)
	ipCount, _ := s.securityRepo.CountIPEntries(ctx, workspaceID)
	recentAlerts, _ := s.securityRepo.GetRecentAlerts(ctx, workspaceID, 5)

	riskLevel := "low"
	if !policy.RequireTwoFactor && policy.AllowExternalSharing {
		riskLevel = "medium"
	}
	if !policy.RequireTwoFactor && policy.AllowGuestAccess && policy.AllowExternalSharing && policy.PasswordMinLength < 8 {
		riskLevel = "high"
	}

	return &models.SecurityOverview{
		Policy:           policy,
		ActiveSessions:   activeSessions,
		IPAllowlistCount: ipCount,
		RecentAlerts:     recentAlerts,
		RiskLevel:        riskLevel,
	}, nil
}
