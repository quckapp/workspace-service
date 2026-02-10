package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/quckapp/workspace-service/internal/db"
	"github.com/quckapp/workspace-service/internal/models"
	"github.com/quckapp/workspace-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var (
	ErrWorkspaceNotFound  = errors.New("workspace not found")
	ErrSlugExists         = errors.New("slug already exists")
	ErrNotMember          = errors.New("not a member of this workspace")
	ErrNotAuthorized      = errors.New("not authorized")
	ErrInviteNotFound     = errors.New("invite not found or expired")
	ErrAlreadyMember      = errors.New("already a member")
	ErrInviteCodeNotFound = errors.New("invite code not found or expired")
	ErrInviteCodeMaxUsed  = errors.New("invite code has reached max uses")
	ErrCannotLeaveAsOwner = errors.New("owner cannot leave workspace, transfer ownership first")
	ErrRoleNotFound        = errors.New("role not found")
	ErrRoleNameExists      = errors.New("role name already exists in this workspace")
	ErrCannotDeleteDefault = errors.New("cannot delete default role")
	ErrTemplateNotFound    = errors.New("template not found")
	ErrTagNotFound         = errors.New("tag not found")
	ErrTagNameExists       = errors.New("tag name already exists in this workspace")
	ErrUserBanned          = errors.New("user is banned from this workspace")
	ErrUserNotBanned       = errors.New("user is not banned")
	ErrUserNotMuted        = errors.New("user is not muted")
	ErrCannotBanOwner      = errors.New("cannot ban workspace owner")
	ErrCannotMuteOwner     = errors.New("cannot mute workspace owner")
	ErrAnnouncementNotFound    = errors.New("announcement not found")
	ErrWebhookNotFound         = errors.New("webhook not found")
	ErrAlreadyFavorited        = errors.New("workspace already favorited")
	ErrNotFavorited            = errors.New("workspace is not favorited")
	ErrMemberNoteNotFound      = errors.New("member note not found")
	ErrScheduledActionNotFound = errors.New("scheduled action not found")
	ErrScheduledActionPast     = errors.New("scheduled time must be in the future")
	ErrQuotaExceeded           = errors.New("workspace quota exceeded")
	ErrWorkspaceArchived       = errors.New("workspace is archived")
	ErrWorkspaceNotArchived    = errors.New("workspace is not archived")
	ErrPinnedItemNotFound      = errors.New("pinned item not found")
	ErrGroupNotFound           = errors.New("group not found")
	ErrGroupNameExists         = errors.New("group name already exists in this workspace")
	ErrAlreadyGroupMember      = errors.New("user is already a member of this group")
	ErrNotGroupMember          = errors.New("user is not a member of this group")
	ErrCustomFieldNotFound     = errors.New("custom field not found")
	ErrCustomFieldNameExists   = errors.New("custom field name already exists in this workspace")
	ErrReactionExists          = errors.New("reaction already exists")
	ErrBookmarkNotFound        = errors.New("bookmark not found")
	ErrBookmarkLimitReached    = errors.New("bookmark limit reached")
	ErrFeatureFlagNotFound     = errors.New("feature flag not found")
	ErrFeatureFlagKeyExists    = errors.New("feature flag key already exists in this workspace")
	ErrIntegrationNotFound     = errors.New("integration not found")
	ErrLabelNotFound           = errors.New("label not found")
	ErrLabelNameExists         = errors.New("label name already exists in this workspace")
	ErrChecklistNotFound       = errors.New("checklist not found")
	ErrOnboardingStepNotFound  = errors.New("onboarding step not found")
	ErrPolicyNotFound          = errors.New("compliance policy not found")
)

const (
	cacheTTL             = 15 * time.Minute
	cacheKeyWorkspace    = "workspace:%s"
	cacheKeyMembers      = "workspace:%s:members"
	cacheKeyStats        = "workspace:%s:stats"
	cacheKeyUserWsList   = "user:%s:workspaces"
)

type WorkspaceService struct {
	workspaceRepo    *repository.WorkspaceRepository
	memberRepo       *repository.MemberRepository
	inviteRepo       *repository.InviteRepository
	inviteCodeRepo   *repository.InviteCodeRepository
	activityRepo     *repository.ActivityRepository
	profileRepo      *repository.ProfileRepository
	roleRepo         *repository.RoleRepository
	templateRepo     *repository.TemplateRepository
	preferenceRepo   *repository.PreferenceRepository
	tagRepo          *repository.TagRepository
	moderationRepo   *repository.ModerationRepository
	announcementRepo *repository.AnnouncementRepository
	webhookRepo        *repository.WebhookRepository
	favoriteRepo       *repository.FavoriteRepository
	memberNoteRepo     *repository.MemberNoteRepository
	scheduledActionRepo *repository.ScheduledActionRepository
	quotaRepo          *repository.QuotaRepository
	pinnedItemRepo         *repository.PinnedItemRepository
	groupRepo              *repository.GroupRepository
	customFieldRepo        *repository.CustomFieldRepository
	reactionRepo           *repository.ReactionRepository
	bookmarkRepo           *repository.BookmarkRepository
	invitationHistoryRepo  *repository.InvitationHistoryRepository
	accessLogRepo          *repository.AccessLogRepository
	featureFlagRepo        *repository.FeatureFlagRepository
	integrationRepo        *repository.IntegrationRepository
	labelRepo              *repository.LabelRepository
	streakRepo             *repository.StreakRepository
	onboardingRepo         *repository.OnboardingRepository
	complianceRepo         *repository.ComplianceRepository
	redis                  *redis.Client
	kafka                  *db.KafkaProducer
	logger                 *logrus.Logger
}

func NewWorkspaceService(
	workspaceRepo *repository.WorkspaceRepository,
	memberRepo *repository.MemberRepository,
	inviteRepo *repository.InviteRepository,
	inviteCodeRepo *repository.InviteCodeRepository,
	activityRepo *repository.ActivityRepository,
	profileRepo *repository.ProfileRepository,
	roleRepo *repository.RoleRepository,
	templateRepo *repository.TemplateRepository,
	preferenceRepo *repository.PreferenceRepository,
	tagRepo *repository.TagRepository,
	moderationRepo *repository.ModerationRepository,
	announcementRepo *repository.AnnouncementRepository,
	webhookRepo *repository.WebhookRepository,
	favoriteRepo *repository.FavoriteRepository,
	memberNoteRepo *repository.MemberNoteRepository,
	scheduledActionRepo *repository.ScheduledActionRepository,
	quotaRepo *repository.QuotaRepository,
	pinnedItemRepo *repository.PinnedItemRepository,
	groupRepo *repository.GroupRepository,
	customFieldRepo *repository.CustomFieldRepository,
	reactionRepo *repository.ReactionRepository,
	bookmarkRepo *repository.BookmarkRepository,
	invitationHistoryRepo *repository.InvitationHistoryRepository,
	accessLogRepo *repository.AccessLogRepository,
	featureFlagRepo *repository.FeatureFlagRepository,
	integrationRepo *repository.IntegrationRepository,
	labelRepo *repository.LabelRepository,
	streakRepo *repository.StreakRepository,
	onboardingRepo *repository.OnboardingRepository,
	complianceRepo *repository.ComplianceRepository,
	redis *redis.Client,
	kafka *db.KafkaProducer,
	logger *logrus.Logger,
) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo:         workspaceRepo,
		memberRepo:            memberRepo,
		inviteRepo:            inviteRepo,
		inviteCodeRepo:        inviteCodeRepo,
		activityRepo:          activityRepo,
		profileRepo:           profileRepo,
		roleRepo:              roleRepo,
		templateRepo:          templateRepo,
		preferenceRepo:        preferenceRepo,
		tagRepo:               tagRepo,
		moderationRepo:        moderationRepo,
		announcementRepo:      announcementRepo,
		webhookRepo:           webhookRepo,
		favoriteRepo:          favoriteRepo,
		memberNoteRepo:        memberNoteRepo,
		scheduledActionRepo:   scheduledActionRepo,
		quotaRepo:             quotaRepo,
		pinnedItemRepo:        pinnedItemRepo,
		groupRepo:             groupRepo,
		customFieldRepo:       customFieldRepo,
		reactionRepo:          reactionRepo,
		bookmarkRepo:          bookmarkRepo,
		invitationHistoryRepo: invitationHistoryRepo,
		accessLogRepo:         accessLogRepo,
		featureFlagRepo:       featureFlagRepo,
		integrationRepo:       integrationRepo,
		labelRepo:             labelRepo,
		streakRepo:            streakRepo,
		onboardingRepo:        onboardingRepo,
		complianceRepo:        complianceRepo,
		redis:                 redis,
		kafka:                 kafka,
		logger:                logger,
	}
}

// ── Workspace CRUD ──

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, ownerID uuid.UUID, req *models.CreateWorkspaceRequest) (*models.Workspace, error) {
	existing, _ := s.workspaceRepo.GetBySlug(ctx, req.Slug)
	if existing != nil {
		return nil, ErrSlugExists
	}

	workspace := &models.Workspace{
		ID:          uuid.New(),
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		OwnerID:     ownerID,
		Plan:        "free",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.workspaceRepo.Create(ctx, workspace); err != nil {
		return nil, err
	}

	member := &models.WorkspaceMember{
		ID:          uuid.New(),
		WorkspaceID: workspace.ID,
		UserID:      ownerID,
		Role:        "owner",
		JoinedAt:    time.Now(),
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	s.memberRepo.Create(ctx, member)

	s.invalidateUserWorkspaces(ctx, ownerID)
	s.publishEvent(ctx, "workspace-events", workspace.ID.String(), "workspace.created", map[string]interface{}{
		"workspace": workspace,
	})

	return workspace, nil
}

func (s *WorkspaceService) GetWorkspace(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.WorkspaceResponse, error) {
	// Try cache
	if cached, err := s.getCachedWorkspaceResponse(ctx, id, userID); err == nil && cached != nil {
		return cached, nil
	}

	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil || workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	memberCount, _ := s.workspaceRepo.GetMemberCount(ctx, id)
	role, _ := s.memberRepo.GetRole(ctx, id, userID)

	resp := &models.WorkspaceResponse{
		Workspace:   workspace,
		MemberCount: memberCount,
		MyRole:      role,
	}

	s.cacheWorkspace(ctx, id, workspace)
	return resp, nil
}

func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *models.UpdateWorkspaceRequest) (*models.Workspace, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil || workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	role, _ := s.memberRepo.GetRole(ctx, id, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	if req.Name != nil {
		workspace.Name = *req.Name
	}
	if req.Description != nil {
		workspace.Description = req.Description
	}
	if req.IconURL != nil {
		workspace.IconURL = req.IconURL
	}
	if req.Settings != nil {
		workspace.Settings = req.Settings
	}

	if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
		return nil, err
	}

	s.invalidateWorkspace(ctx, id)
	s.publishEvent(ctx, "workspace-events", id.String(), "workspace.updated", map[string]interface{}{
		"workspace": workspace,
		"updated_by": userID,
	})

	return workspace, nil
}

func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil || workspace == nil {
		return ErrWorkspaceNotFound
	}

	if workspace.OwnerID != userID {
		return ErrNotAuthorized
	}

	if err := s.workspaceRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.invalidateWorkspace(ctx, id)
	s.invalidateUserWorkspaces(ctx, userID)
	s.publishEvent(ctx, "workspace-events", id.String(), "workspace.deleted", map[string]interface{}{
		"workspace_id": id,
		"deleted_by":   userID,
	})

	return nil
}

func (s *WorkspaceService) ListWorkspaces(ctx context.Context, userID uuid.UUID, page, perPage int) (*models.WorkspacesListResponse, error) {
	workspaces, total, err := s.workspaceRepo.ListByUserID(ctx, userID, page, perPage)
	if err != nil {
		return nil, err
	}

	var responses []*models.WorkspaceResponse
	for _, w := range workspaces {
		memberCount, _ := s.workspaceRepo.GetMemberCount(ctx, w.ID)
		role, _ := s.memberRepo.GetRole(ctx, w.ID, userID)
		responses = append(responses, &models.WorkspaceResponse{
			Workspace:   w,
			MemberCount: memberCount,
			MyRole:      role,
		})
	}

	return &models.WorkspacesListResponse{
		Workspaces: responses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
	}, nil
}

// ── Workspace Stats ──

func (s *WorkspaceService) GetWorkspaceStats(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID) (*models.WorkspaceStats, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil || workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	// Try cache
	if cached, err := s.getCachedStats(ctx, workspaceID); err == nil && cached != nil {
		return cached, nil
	}

	memberCount, _ := s.workspaceRepo.GetMemberCount(ctx, workspaceID)
	inviteCount, _ := s.inviteRepo.GetPendingCount(ctx, workspaceID)
	roleCounts, _ := s.workspaceRepo.GetRoleCounts(ctx, workspaceID)

	stats := &models.WorkspaceStats{
		MemberCount: memberCount,
		InviteCount: inviteCount,
		RoleCounts:  roleCounts,
		CreatedAt:   workspace.CreatedAt,
		Plan:        workspace.Plan,
	}

	s.cacheStats(ctx, workspaceID, stats)
	return stats, nil
}

// ── Workspace Settings ──

func (s *WorkspaceService) GetWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID) (models.JSON, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil || workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	if workspace.Settings == nil {
		return models.JSON{}, nil
	}
	return workspace.Settings, nil
}

func (s *WorkspaceService) UpdateWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, settings models.JSON) (models.JSON, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil || workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	workspace.Settings = settings
	if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
		return nil, err
	}

	s.invalidateWorkspace(ctx, workspaceID)
	return settings, nil
}

// ── Leave Workspace ──

func (s *WorkspaceService) LeaveWorkspace(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID) error {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil || workspace == nil {
		return ErrWorkspaceNotFound
	}

	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return ErrNotMember
	}

	if workspace.OwnerID == userID {
		return ErrCannotLeaveAsOwner
	}

	if err := s.memberRepo.Remove(ctx, workspaceID, userID); err != nil {
		return err
	}

	s.invalidateWorkspace(ctx, workspaceID)
	s.invalidateUserWorkspaces(ctx, userID)
	s.publishEvent(ctx, "workspace-events", workspaceID.String(), "member.left", map[string]interface{}{
		"workspace_id": workspaceID,
		"user_id":      userID,
	})

	return nil
}

// ── Get Member ──

func (s *WorkspaceService) GetMember(ctx context.Context, workspaceID, memberUserID uuid.UUID) (*models.WorkspaceMember, error) {
	member, err := s.memberRepo.GetByID(ctx, workspaceID, memberUserID)
	if err != nil || member == nil {
		return nil, ErrNotMember
	}
	return member, nil
}

// ── Ownership Transfer ──

func (s *WorkspaceService) TransferOwnership(ctx context.Context, workspaceID, currentOwnerID, newOwnerID uuid.UUID) error {
	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil || workspace == nil {
		return ErrWorkspaceNotFound
	}

	if workspace.OwnerID != currentOwnerID {
		return ErrNotAuthorized
	}

	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, newOwnerID)
	if !isMember {
		return ErrNotMember
	}

	if err := s.workspaceRepo.TransferOwnership(ctx, workspaceID, newOwnerID); err != nil {
		return err
	}

	// Update roles
	s.memberRepo.UpdateRole(ctx, workspaceID, newOwnerID, "owner")
	s.memberRepo.UpdateRole(ctx, workspaceID, currentOwnerID, "admin")

	s.invalidateWorkspace(ctx, workspaceID)
	s.publishEvent(ctx, "workspace-events", workspaceID.String(), "ownership.transferred", map[string]interface{}{
		"workspace_id":   workspaceID,
		"previous_owner": currentOwnerID,
		"new_owner":      newOwnerID,
	})

	return nil
}

// ── Member Management ──

func (s *WorkspaceService) InviteMember(ctx context.Context, workspaceID uuid.UUID, inviterID uuid.UUID, req *models.InviteMemberRequest) (*models.WorkspaceInvite, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, inviterID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	existing, _ := s.inviteRepo.GetPendingByEmail(ctx, workspaceID, req.Email)
	if existing != nil {
		return existing, nil
	}

	token := generateToken()
	invite := &models.WorkspaceInvite{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Email:       req.Email,
		Role:        req.Role,
		Token:       token,
		InvitedBy:   inviterID,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:   time.Now(),
	}

	if err := s.inviteRepo.Create(ctx, invite); err != nil {
		return nil, err
	}

	s.publishEvent(ctx, "notification-events", invite.ID.String(), "workspace.invite", map[string]interface{}{
		"invite": invite,
	})

	return invite, nil
}

func (s *WorkspaceService) BulkInvite(ctx context.Context, workspaceID uuid.UUID, inviterID uuid.UUID, req *models.BulkInviteRequest) (*models.BulkInviteResponse, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, inviterID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	resp := &models.BulkInviteResponse{}
	for _, inv := range req.Invites {
		_, err := s.InviteMember(ctx, workspaceID, inviterID, &inv)
		if err != nil {
			resp.Failed = append(resp.Failed, struct {
				Email  string `json:"email"`
				Reason string `json:"reason"`
			}{Email: inv.Email, Reason: err.Error()})
		} else {
			resp.Successful = append(resp.Successful, inv.Email)
		}
	}

	return resp, nil
}

func (s *WorkspaceService) AcceptInvite(ctx context.Context, token string, userID uuid.UUID) (*models.Workspace, error) {
	invite, err := s.inviteRepo.GetByToken(ctx, token)
	if err != nil || invite == nil {
		return nil, ErrInviteNotFound
	}

	// Check if user is banned
	isBanned, _ := s.moderationRepo.IsUserBanned(ctx, invite.WorkspaceID, userID)
	if isBanned {
		return nil, ErrUserBanned
	}

	isMember, _ := s.memberRepo.IsMember(ctx, invite.WorkspaceID, userID)
	if isMember {
		return nil, ErrAlreadyMember
	}

	member := &models.WorkspaceMember{
		ID:          uuid.New(),
		WorkspaceID: invite.WorkspaceID,
		UserID:      userID,
		Role:        invite.Role,
		JoinedAt:    time.Now(),
		InvitedBy:   &invite.InvitedBy,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	s.inviteRepo.MarkAccepted(ctx, invite.ID)
	s.invalidateWorkspace(ctx, invite.WorkspaceID)
	s.invalidateUserWorkspaces(ctx, userID)
	s.publishEvent(ctx, "workspace-events", invite.WorkspaceID.String(), "member.joined", map[string]interface{}{
		"workspace_id": invite.WorkspaceID,
		"user_id":      userID,
		"role":         invite.Role,
	})

	return s.workspaceRepo.GetByID(ctx, invite.WorkspaceID)
}

func (s *WorkspaceService) RemoveMember(ctx context.Context, workspaceID, memberUserID, requestorID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, requestorID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	memberRole, _ := s.memberRepo.GetRole(ctx, workspaceID, memberUserID)
	if memberRole == "owner" {
		return ErrNotAuthorized
	}

	if err := s.memberRepo.Remove(ctx, workspaceID, memberUserID); err != nil {
		return err
	}

	s.invalidateWorkspace(ctx, workspaceID)
	s.invalidateUserWorkspaces(ctx, memberUserID)
	s.publishEvent(ctx, "workspace-events", workspaceID.String(), "member.removed", map[string]interface{}{
		"workspace_id": workspaceID,
		"user_id":      memberUserID,
		"removed_by":   requestorID,
	})

	return nil
}

func (s *WorkspaceService) UpdateMemberRole(ctx context.Context, workspaceID, memberUserID, requestorID uuid.UUID, newRole string) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, requestorID)
	if role != "owner" {
		return ErrNotAuthorized
	}

	if err := s.memberRepo.UpdateRole(ctx, workspaceID, memberUserID, newRole); err != nil {
		return err
	}

	s.invalidateWorkspace(ctx, workspaceID)
	s.publishEvent(ctx, "workspace-events", workspaceID.String(), "member.role_updated", map[string]interface{}{
		"workspace_id": workspaceID,
		"user_id":      memberUserID,
		"new_role":     newRole,
		"updated_by":   requestorID,
	})

	return nil
}

func (s *WorkspaceService) ListMembers(ctx context.Context, workspaceID uuid.UUID, page, perPage int) ([]*models.WorkspaceMember, int64, error) {
	return s.memberRepo.ListByWorkspace(ctx, workspaceID, page, perPage)
}

// ── Invite Management ──

func (s *WorkspaceService) ListInvites(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID) ([]*models.WorkspaceInvite, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	return s.inviteRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *WorkspaceService) RevokeInvite(ctx context.Context, workspaceID uuid.UUID, inviteID uuid.UUID, userID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	invite, err := s.inviteRepo.GetByID(ctx, inviteID)
	if err != nil || invite == nil {
		return ErrInviteNotFound
	}

	if invite.WorkspaceID != workspaceID {
		return ErrNotAuthorized
	}

	return s.inviteRepo.Delete(ctx, inviteID)
}

// ── Invite Code System ──

func (s *WorkspaceService) CreateInviteCode(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, req *models.CreateInviteCodeRequest) (*models.WorkspaceInviteCode, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	code := generateInviteCode()
	inviteCode := &models.WorkspaceInviteCode{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Code:        code,
		Role:        req.Role,
		MaxUses:     req.MaxUses,
		UseCount:    0,
		CreatedBy:   userID,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.inviteCodeRepo.Create(ctx, inviteCode); err != nil {
		return nil, err
	}

	return inviteCode, nil
}

func (s *WorkspaceService) JoinByCode(ctx context.Context, code string, userID uuid.UUID) (*models.Workspace, error) {
	inviteCode, err := s.inviteCodeRepo.GetByCode(ctx, code)
	if err != nil || inviteCode == nil {
		return nil, ErrInviteCodeNotFound
	}

	if inviteCode.MaxUses > 0 && inviteCode.UseCount >= inviteCode.MaxUses {
		return nil, ErrInviteCodeMaxUsed
	}

	// Check if user is banned
	isBanned, _ := s.moderationRepo.IsUserBanned(ctx, inviteCode.WorkspaceID, userID)
	if isBanned {
		return nil, ErrUserBanned
	}

	isMember, _ := s.memberRepo.IsMember(ctx, inviteCode.WorkspaceID, userID)
	if isMember {
		return nil, ErrAlreadyMember
	}

	member := &models.WorkspaceMember{
		ID:          uuid.New(),
		WorkspaceID: inviteCode.WorkspaceID,
		UserID:      userID,
		Role:        inviteCode.Role,
		JoinedAt:    time.Now(),
		InvitedBy:   &inviteCode.CreatedBy,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	s.inviteCodeRepo.IncrementUseCount(ctx, inviteCode.ID)
	s.invalidateWorkspace(ctx, inviteCode.WorkspaceID)
	s.invalidateUserWorkspaces(ctx, userID)
	s.publishEvent(ctx, "workspace-events", inviteCode.WorkspaceID.String(), "member.joined_by_code", map[string]interface{}{
		"workspace_id": inviteCode.WorkspaceID,
		"user_id":      userID,
		"invite_code":  code,
	})

	return s.workspaceRepo.GetByID(ctx, inviteCode.WorkspaceID)
}

func (s *WorkspaceService) ListInviteCodes(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID) ([]*models.WorkspaceInviteCode, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	return s.inviteCodeRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *WorkspaceService) RevokeInviteCode(ctx context.Context, workspaceID uuid.UUID, codeID uuid.UUID, userID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	return s.inviteCodeRepo.Deactivate(ctx, codeID)
}

// ── Activity Log ──

func (s *WorkspaceService) LogActivity(ctx context.Context, workspaceID, actorID uuid.UUID, action, entityType, entityID string, details models.JSON) {
	log := &models.ActivityLog{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Details:     details,
		CreatedAt:   time.Now(),
	}
	if err := s.activityRepo.Create(ctx, log); err != nil {
		s.logger.WithError(err).Warn("Failed to log activity")
	}
}

func (s *WorkspaceService) GetActivityLog(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, page, perPage int) (*models.ActivityLogResponse, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	activities, total, err := s.activityRepo.ListByWorkspace(ctx, workspaceID, page, perPage)
	if err != nil {
		return nil, err
	}

	return &models.ActivityLogResponse{
		Activities: activities,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
	}, nil
}

func (s *WorkspaceService) GetActivityLogByActor(ctx context.Context, workspaceID, actorID, userID uuid.UUID, page, perPage int) (*models.ActivityLogResponse, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	activities, total, err := s.activityRepo.ListByActor(ctx, workspaceID, actorID, page, perPage)
	if err != nil {
		return nil, err
	}

	return &models.ActivityLogResponse{
		Activities: activities,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
	}, nil
}

// ── Member Profiles ──

func (s *WorkspaceService) GetMemberProfile(ctx context.Context, workspaceID, memberUserID uuid.UUID) (*models.MemberProfile, error) {
	return s.profileRepo.GetByWorkspaceAndUser(ctx, workspaceID, memberUserID)
}

func (s *WorkspaceService) UpdateMemberProfile(ctx context.Context, workspaceID, userID uuid.UUID, req *models.UpdateMemberProfileRequest) (*models.MemberProfile, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	existing, _ := s.profileRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
	now := time.Now()

	profile := &models.MemberProfile{
		WorkspaceID: workspaceID,
		UserID:      userID,
		UpdatedAt:   now,
	}

	if existing != nil {
		profile.ID = existing.ID
		profile.CreatedAt = existing.CreatedAt
		profile.DisplayName = existing.DisplayName
		profile.Title = existing.Title
		profile.StatusText = existing.StatusText
		profile.StatusEmoji = existing.StatusEmoji
		profile.Timezone = existing.Timezone
		profile.IsOnline = existing.IsOnline
		profile.LastSeenAt = existing.LastSeenAt
	} else {
		profile.ID = uuid.New()
		profile.CreatedAt = now
	}

	if req.DisplayName != nil {
		profile.DisplayName = req.DisplayName
	}
	if req.Title != nil {
		profile.Title = req.Title
	}
	if req.StatusText != nil {
		profile.StatusText = req.StatusText
	}
	if req.StatusEmoji != nil {
		profile.StatusEmoji = req.StatusEmoji
	}
	if req.Timezone != nil {
		profile.Timezone = req.Timezone
	}

	if err := s.profileRepo.Upsert(ctx, profile); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "profile.updated", "member", userID.String(), nil)
	return profile, nil
}

func (s *WorkspaceService) SetOnlineStatus(ctx context.Context, workspaceID, userID uuid.UUID, isOnline bool) error {
	return s.profileRepo.UpdateOnlineStatus(ctx, workspaceID, userID, isOnline)
}

// ── Custom Roles ──

func (s *WorkspaceService) CreateRole(ctx context.Context, workspaceID, userID uuid.UUID, req *models.CreateRoleRequest) (*models.WorkspaceRole, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" {
		return nil, ErrNotAuthorized
	}

	existing, _ := s.roleRepo.GetByName(ctx, workspaceID, req.Name)
	if existing != nil {
		return nil, ErrRoleNameExists
	}

	newRole := &models.WorkspaceRole{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Color:       req.Color,
		Priority:    req.Priority,
		Permissions: req.Permissions,
		IsDefault:   false,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.roleRepo.Create(ctx, newRole); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "role.created", "role", newRole.ID.String(), models.JSON{"name": req.Name})
	return newRole, nil
}

func (s *WorkspaceService) ListRoles(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceRole, error) {
	return s.roleRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *WorkspaceService) UpdateRole(ctx context.Context, workspaceID, roleID, userID uuid.UUID, req *models.UpdateRoleRequest) (*models.WorkspaceRole, error) {
	memberRole, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if memberRole != "owner" {
		return nil, ErrNotAuthorized
	}

	existingRole, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil || existingRole == nil {
		return nil, ErrRoleNotFound
	}

	if existingRole.WorkspaceID != workspaceID {
		return nil, ErrNotAuthorized
	}

	if req.Name != nil {
		dup, _ := s.roleRepo.GetByName(ctx, workspaceID, *req.Name)
		if dup != nil && dup.ID != roleID {
			return nil, ErrRoleNameExists
		}
		existingRole.Name = *req.Name
	}
	if req.Color != nil {
		existingRole.Color = req.Color
	}
	if req.Priority != nil {
		existingRole.Priority = *req.Priority
	}
	if req.Permissions != nil {
		existingRole.Permissions = req.Permissions
	}

	if err := s.roleRepo.Update(ctx, existingRole); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "role.updated", "role", roleID.String(), models.JSON{"name": existingRole.Name})
	return existingRole, nil
}

func (s *WorkspaceService) DeleteRole(ctx context.Context, workspaceID, roleID, userID uuid.UUID) error {
	memberRole, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if memberRole != "owner" {
		return ErrNotAuthorized
	}

	existingRole, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil || existingRole == nil {
		return ErrRoleNotFound
	}

	if existingRole.IsDefault {
		return ErrCannotDeleteDefault
	}

	if err := s.roleRepo.Delete(ctx, roleID); err != nil {
		return err
	}

	s.LogActivity(ctx, workspaceID, userID, "role.deleted", "role", roleID.String(), nil)
	return nil
}

// ── Workspace Search ──

func (s *WorkspaceService) SearchWorkspaces(ctx context.Context, query string, page, perPage int) ([]*models.Workspace, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.workspaceRepo.Search(ctx, query, page, perPage)
}

// ── Workspace Analytics ──

func (s *WorkspaceService) GetAnalytics(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, days int) (*models.WorkspaceAnalytics, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	if days < 1 || days > 365 {
		days = 30
	}

	memberGrowth, _ := s.workspaceRepo.GetMemberGrowth(ctx, workspaceID, days)
	roleCounts, _ := s.workspaceRepo.GetRoleCounts(ctx, workspaceID)
	joinMethodStats, _ := s.workspaceRepo.GetJoinMethodStats(ctx, workspaceID)

	topContributors, _ := s.activityRepo.GetTopContributors(ctx, workspaceID, time.Now().AddDate(0, 0, -days), 10)

	// Count active members from activity log in last 30 days
	allActivities, _, _ := s.activityRepo.ListByWorkspace(ctx, workspaceID, 1, 1)
	activeCount := 0
	if allActivities != nil {
		activeCount = len(allActivities)
	}

	return &models.WorkspaceAnalytics{
		MemberGrowth:     memberGrowth,
		ActiveMembers:    activeCount,
		TopContributors:  topContributors,
		RoleDistribution: roleCounts,
		JoinMethodStats:  joinMethodStats,
	}, nil
}

// ── Workspace Templates ──

func (s *WorkspaceService) CreateTemplateFromWorkspace(ctx context.Context, workspaceID, userID uuid.UUID, req *models.CreateTemplateFromWorkspaceRequest) (*models.WorkspaceTemplate, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil || workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	roles, _ := s.roleRepo.ListByWorkspace(ctx, workspaceID)
	var rolesData []map[string]interface{}
	for _, r := range roles {
		rolesData = append(rolesData, map[string]interface{}{
			"name": r.Name, "color": r.Color, "priority": r.Priority, "permissions": r.Permissions, "is_default": r.IsDefault,
		})
	}

	template := &models.WorkspaceTemplate{
		ID:              uuid.New(),
		Name:            req.Name,
		Description:     req.Description,
		CreatedBy:       userID,
		DefaultRoles:    models.JSON{"roles": rolesData},
		DefaultChannels: models.JSON{},
		DefaultSettings: models.JSON(workspace.Settings),
		IsPublic:        req.IsPublic,
		UseCount:        0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.templateRepo.Create(ctx, template); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "template.created", "template", template.ID.String(), models.JSON{"name": req.Name})
	return template, nil
}

func (s *WorkspaceService) CreateWorkspaceFromTemplate(ctx context.Context, userID uuid.UUID, req *models.CreateWorkspaceFromTemplateRequest) (*models.Workspace, error) {
	templateID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		return nil, ErrTemplateNotFound
	}

	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil || template == nil {
		return nil, ErrTemplateNotFound
	}

	existing, _ := s.workspaceRepo.GetBySlug(ctx, req.Slug)
	if existing != nil {
		return nil, ErrSlugExists
	}

	var settings models.JSON
	if template.DefaultSettings != nil {
		settings = template.DefaultSettings
	}

	workspace := &models.Workspace{
		ID:        uuid.New(),
		Name:      req.Name,
		Slug:      req.Slug,
		OwnerID:   userID,
		Plan:      "free",
		Settings:  settings,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.workspaceRepo.Create(ctx, workspace); err != nil {
		return nil, err
	}

	member := &models.WorkspaceMember{
		ID:          uuid.New(),
		WorkspaceID: workspace.ID,
		UserID:      userID,
		Role:        "owner",
		JoinedAt:    time.Now(),
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	s.memberRepo.Create(ctx, member)

	// Create roles from template
	if template.DefaultRoles != nil {
		if rolesRaw, ok := template.DefaultRoles["roles"]; ok {
			if rolesSlice, ok := rolesRaw.([]interface{}); ok {
				for _, roleRaw := range rolesSlice {
					if roleMap, ok := roleRaw.(map[string]interface{}); ok {
						name, _ := roleMap["name"].(string)
						if name == "" {
							continue
						}
						priority := 0
						if p, ok := roleMap["priority"].(float64); ok {
							priority = int(p)
						}
						var color *string
						if c, ok := roleMap["color"].(string); ok {
							color = &c
						}
						isDefault := false
						if d, ok := roleMap["is_default"].(bool); ok {
							isDefault = d
						}
						var permissions models.JSON
						if p, ok := roleMap["permissions"].(map[string]interface{}); ok {
							permissions = models.JSON(p)
						}
						newRole := &models.WorkspaceRole{
							ID:          uuid.New(),
							WorkspaceID: workspace.ID,
							Name:        name,
							Color:       color,
							Priority:    priority,
							Permissions: permissions,
							IsDefault:   isDefault,
							CreatedBy:   userID,
							CreatedAt:   time.Now(),
							UpdatedAt:   time.Now(),
						}
						s.roleRepo.Create(ctx, newRole)
					}
				}
			}
		}
	}

	s.templateRepo.IncrementUseCount(ctx, templateID)
	s.invalidateUserWorkspaces(ctx, userID)
	s.publishEvent(ctx, "workspace-events", workspace.ID.String(), "workspace.created_from_template", map[string]interface{}{
		"workspace":   workspace,
		"template_id": templateID,
	})

	return workspace, nil
}

func (s *WorkspaceService) ListTemplates(ctx context.Context, page, perPage int) ([]*models.WorkspaceTemplate, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.templateRepo.ListPublic(ctx, page, perPage)
}

func (s *WorkspaceService) GetTemplate(ctx context.Context, templateID uuid.UUID) (*models.WorkspaceTemplate, error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil || template == nil {
		return nil, ErrTemplateNotFound
	}
	return template, nil
}

func (s *WorkspaceService) UpdateTemplate(ctx context.Context, templateID, userID uuid.UUID, req *models.UpdateTemplateRequest) (*models.WorkspaceTemplate, error) {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil || template == nil {
		return nil, ErrTemplateNotFound
	}

	if template.CreatedBy != userID {
		return nil, ErrNotAuthorized
	}

	if req.Name != nil {
		template.Name = *req.Name
	}
	if req.Description != nil {
		template.Description = req.Description
	}
	if req.IsPublic != nil {
		template.IsPublic = *req.IsPublic
	}

	if err := s.templateRepo.Update(ctx, template); err != nil {
		return nil, err
	}

	return template, nil
}

func (s *WorkspaceService) DeleteTemplate(ctx context.Context, templateID, userID uuid.UUID) error {
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil || template == nil {
		return ErrTemplateNotFound
	}

	if template.CreatedBy != userID {
		return ErrNotAuthorized
	}

	return s.templateRepo.Delete(ctx, templateID)
}

// ── Member Preferences ──

func (s *WorkspaceService) GetPreferences(ctx context.Context, workspaceID, userID uuid.UUID) (*models.WorkspaceMemberPreference, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	pref, _ := s.preferenceRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
	if pref == nil {
		// Return defaults
		return &models.WorkspaceMemberPreference{
			WorkspaceID:        workspaceID,
			UserID:             userID,
			NotificationLevel:  "all",
			EmailNotifications: true,
			SidebarPosition:    0,
		}, nil
	}
	return pref, nil
}

func (s *WorkspaceService) UpdatePreferences(ctx context.Context, workspaceID, userID uuid.UUID, req *models.UpdatePreferencesRequest) (*models.WorkspaceMemberPreference, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	existing, _ := s.preferenceRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
	now := time.Now()

	pref := &models.WorkspaceMemberPreference{
		WorkspaceID:        workspaceID,
		UserID:             userID,
		NotificationLevel:  "all",
		EmailNotifications: true,
		UpdatedAt:          now,
	}

	if existing != nil {
		pref.ID = existing.ID
		pref.CreatedAt = existing.CreatedAt
		pref.NotificationLevel = existing.NotificationLevel
		pref.EmailNotifications = existing.EmailNotifications
		pref.MuteUntil = existing.MuteUntil
		pref.SidebarPosition = existing.SidebarPosition
		pref.Theme = existing.Theme
	} else {
		pref.ID = uuid.New()
		pref.CreatedAt = now
	}

	if req.NotificationLevel != nil {
		pref.NotificationLevel = *req.NotificationLevel
	}
	if req.EmailNotifications != nil {
		pref.EmailNotifications = *req.EmailNotifications
	}
	if req.MuteUntil != nil {
		pref.MuteUntil = req.MuteUntil
	}
	if req.SidebarPosition != nil {
		pref.SidebarPosition = *req.SidebarPosition
	}
	if req.Theme != nil {
		pref.Theme = req.Theme
	}

	if err := s.preferenceRepo.Upsert(ctx, pref); err != nil {
		return nil, err
	}

	return pref, nil
}

func (s *WorkspaceService) ResetPreferences(ctx context.Context, workspaceID, userID uuid.UUID) error {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return ErrNotMember
	}

	return s.preferenceRepo.Delete(ctx, workspaceID, userID)
}

// ── Workspace Tags ──

func (s *WorkspaceService) CreateTag(ctx context.Context, workspaceID, userID uuid.UUID, req *models.CreateTagRequest) (*models.WorkspaceTag, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	existing, _ := s.tagRepo.GetByName(ctx, workspaceID, req.Name)
	if existing != nil {
		return nil, ErrTagNameExists
	}

	tag := &models.WorkspaceTag{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Color:       req.Color,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.tagRepo.Create(ctx, tag); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "tag.created", "tag", tag.ID.String(), models.JSON{"name": req.Name})
	return tag, nil
}

func (s *WorkspaceService) ListTags(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceTag, error) {
	return s.tagRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *WorkspaceService) UpdateTag(ctx context.Context, workspaceID, tagID, userID uuid.UUID, req *models.UpdateTagRequest) (*models.WorkspaceTag, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	tag, err := s.tagRepo.GetByID(ctx, tagID)
	if err != nil || tag == nil {
		return nil, ErrTagNotFound
	}

	if tag.WorkspaceID != workspaceID {
		return nil, ErrNotAuthorized
	}

	if req.Name != nil {
		dup, _ := s.tagRepo.GetByName(ctx, workspaceID, *req.Name)
		if dup != nil && dup.ID != tagID {
			return nil, ErrTagNameExists
		}
		tag.Name = *req.Name
	}
	if req.Color != nil {
		tag.Color = req.Color
	}

	if err := s.tagRepo.Update(ctx, tag); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "tag.updated", "tag", tagID.String(), models.JSON{"name": tag.Name})
	return tag, nil
}

func (s *WorkspaceService) DeleteTag(ctx context.Context, workspaceID, tagID, userID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	tag, err := s.tagRepo.GetByID(ctx, tagID)
	if err != nil || tag == nil {
		return ErrTagNotFound
	}

	if tag.WorkspaceID != workspaceID {
		return ErrNotAuthorized
	}

	if err := s.tagRepo.Delete(ctx, tagID); err != nil {
		return err
	}

	s.LogActivity(ctx, workspaceID, userID, "tag.deleted", "tag", tagID.String(), nil)
	return nil
}

// ── Workspace Moderation ──

func (s *WorkspaceService) BanMember(ctx context.Context, workspaceID, targetUserID, actorID uuid.UUID, req *models.BanMemberRequest) (*models.WorkspaceBan, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, actorID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	targetRole, _ := s.memberRepo.GetRole(ctx, workspaceID, targetUserID)
	if targetRole == "owner" {
		return nil, ErrCannotBanOwner
	}

	existingBan, _ := s.moderationRepo.GetBan(ctx, workspaceID, targetUserID)
	if existingBan != nil {
		s.moderationRepo.RemoveBan(ctx, workspaceID, targetUserID)
	}

	ban := &models.WorkspaceBan{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		UserID:      targetUserID,
		BannedBy:    actorID,
		Reason:      req.Reason,
		ExpiresAt:   req.ExpiresAt,
		IsPermanent: req.IsPermanent,
		CreatedAt:   time.Now(),
	}

	if err := s.moderationRepo.CreateBan(ctx, ban); err != nil {
		return nil, err
	}

	// Remove member from workspace
	s.memberRepo.Remove(ctx, workspaceID, targetUserID)
	s.invalidateWorkspace(ctx, workspaceID)
	s.invalidateUserWorkspaces(ctx, targetUserID)

	s.LogActivity(ctx, workspaceID, actorID, "member.banned", "member", targetUserID.String(), models.JSON{"reason": req.Reason, "is_permanent": req.IsPermanent})
	s.publishEvent(ctx, "workspace-events", workspaceID.String(), "member.banned", map[string]interface{}{
		"workspace_id": workspaceID,
		"user_id":      targetUserID,
		"banned_by":    actorID,
	})

	return ban, nil
}

func (s *WorkspaceService) UnbanMember(ctx context.Context, workspaceID, targetUserID, actorID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, actorID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	ban, _ := s.moderationRepo.GetBan(ctx, workspaceID, targetUserID)
	if ban == nil {
		return ErrUserNotBanned
	}

	if err := s.moderationRepo.RemoveBan(ctx, workspaceID, targetUserID); err != nil {
		return err
	}

	s.LogActivity(ctx, workspaceID, actorID, "member.unbanned", "member", targetUserID.String(), nil)
	s.publishEvent(ctx, "workspace-events", workspaceID.String(), "member.unbanned", map[string]interface{}{
		"workspace_id": workspaceID,
		"user_id":      targetUserID,
		"unbanned_by":  actorID,
	})

	return nil
}

func (s *WorkspaceService) MuteMember(ctx context.Context, workspaceID, targetUserID, actorID uuid.UUID, req *models.MuteMemberRequest) (*models.WorkspaceMute, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, actorID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	targetRole, _ := s.memberRepo.GetRole(ctx, workspaceID, targetUserID)
	if targetRole == "owner" {
		return nil, ErrCannotMuteOwner
	}

	existingMute, _ := s.moderationRepo.GetMute(ctx, workspaceID, targetUserID)
	if existingMute != nil {
		s.moderationRepo.RemoveMute(ctx, workspaceID, targetUserID)
	}

	mute := &models.WorkspaceMute{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		UserID:      targetUserID,
		MutedBy:     actorID,
		Reason:      req.Reason,
		ExpiresAt:   req.ExpiresAt,
		CreatedAt:   time.Now(),
	}

	if err := s.moderationRepo.CreateMute(ctx, mute); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, actorID, "member.muted", "member", targetUserID.String(), models.JSON{"reason": req.Reason})
	s.publishEvent(ctx, "workspace-events", workspaceID.String(), "member.muted", map[string]interface{}{
		"workspace_id": workspaceID,
		"user_id":      targetUserID,
		"muted_by":     actorID,
	})

	return mute, nil
}

func (s *WorkspaceService) UnmuteMember(ctx context.Context, workspaceID, targetUserID, actorID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, actorID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	mute, _ := s.moderationRepo.GetMute(ctx, workspaceID, targetUserID)
	if mute == nil {
		return ErrUserNotMuted
	}

	if err := s.moderationRepo.RemoveMute(ctx, workspaceID, targetUserID); err != nil {
		return err
	}

	s.LogActivity(ctx, workspaceID, actorID, "member.unmuted", "member", targetUserID.String(), nil)
	return nil
}

func (s *WorkspaceService) GetModerationHistory(ctx context.Context, workspaceID, userID uuid.UUID) (*models.ModerationHistoryResponse, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	bans, _ := s.moderationRepo.ListBans(ctx, workspaceID)
	mutes, _ := s.moderationRepo.ListMutes(ctx, workspaceID)

	return &models.ModerationHistoryResponse{
		Bans:  bans,
		Mutes: mutes,
	}, nil
}

// ── Workspace Announcements ──

func (s *WorkspaceService) CreateAnnouncement(ctx context.Context, workspaceID, userID uuid.UUID, req *models.CreateAnnouncementRequest) (*models.WorkspaceAnnouncement, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	announcement := &models.WorkspaceAnnouncement{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Title:       req.Title,
		Content:     req.Content,
		Priority:    req.Priority,
		AuthorID:    userID,
		IsPinned:    req.IsPinned,
		ExpiresAt:   req.ExpiresAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.announcementRepo.Create(ctx, announcement); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "announcement.created", "announcement", announcement.ID.String(), models.JSON{"title": req.Title})
	s.publishEvent(ctx, "workspace-events", workspaceID.String(), "workspace.announcement.created", map[string]interface{}{
		"announcement": announcement,
	})

	return announcement, nil
}

func (s *WorkspaceService) ListAnnouncements(ctx context.Context, workspaceID, userID uuid.UUID, page, perPage int) ([]*models.WorkspaceAnnouncement, int64, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, 0, ErrNotMember
	}

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	return s.announcementRepo.ListByWorkspace(ctx, workspaceID, page, perPage)
}

func (s *WorkspaceService) UpdateAnnouncement(ctx context.Context, workspaceID, announcementID, userID uuid.UUID, req *models.UpdateAnnouncementRequest) (*models.WorkspaceAnnouncement, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	announcement, err := s.announcementRepo.GetByID(ctx, announcementID)
	if err != nil || announcement == nil {
		return nil, ErrAnnouncementNotFound
	}

	if announcement.WorkspaceID != workspaceID {
		return nil, ErrNotAuthorized
	}

	if req.Title != nil {
		announcement.Title = *req.Title
	}
	if req.Content != nil {
		announcement.Content = *req.Content
	}
	if req.Priority != nil {
		announcement.Priority = *req.Priority
	}
	if req.ExpiresAt != nil {
		announcement.ExpiresAt = req.ExpiresAt
	}

	if err := s.announcementRepo.Update(ctx, announcement); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "announcement.updated", "announcement", announcementID.String(), nil)
	return announcement, nil
}

func (s *WorkspaceService) PinAnnouncement(ctx context.Context, workspaceID, announcementID, userID uuid.UUID, req *models.PinAnnouncementRequest) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	announcement, err := s.announcementRepo.GetByID(ctx, announcementID)
	if err != nil || announcement == nil {
		return ErrAnnouncementNotFound
	}

	if announcement.WorkspaceID != workspaceID {
		return ErrNotAuthorized
	}

	if err := s.announcementRepo.UpdatePinStatus(ctx, announcementID, req.IsPinned); err != nil {
		return err
	}

	action := "announcement.pinned"
	if !req.IsPinned {
		action = "announcement.unpinned"
	}
	s.LogActivity(ctx, workspaceID, userID, action, "announcement", announcementID.String(), nil)
	return nil
}

func (s *WorkspaceService) DeleteAnnouncement(ctx context.Context, workspaceID, announcementID, userID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	announcement, err := s.announcementRepo.GetByID(ctx, announcementID)
	if err != nil || announcement == nil {
		return ErrAnnouncementNotFound
	}

	if announcement.WorkspaceID != workspaceID {
		return ErrNotAuthorized
	}

	if err := s.announcementRepo.Delete(ctx, announcementID); err != nil {
		return err
	}

	s.LogActivity(ctx, workspaceID, userID, "announcement.deleted", "announcement", announcementID.String(), nil)
	return nil
}

// ── Workspace Webhooks ──

func (s *WorkspaceService) CreateWebhook(ctx context.Context, workspaceID, userID uuid.UUID, req *models.CreateWebhookRequest) (*models.WorkspaceWebhook, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	eventsJSON := models.JSON{"events": req.Events}

	webhook := &models.WorkspaceWebhook{
		ID:           uuid.New(),
		WorkspaceID:  workspaceID,
		Name:         req.Name,
		URL:          req.URL,
		Secret:       generateToken(),
		Events:       eventsJSON,
		IsActive:     true,
		CreatedBy:    userID,
		FailureCount: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.webhookRepo.Create(ctx, webhook); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "webhook.created", "webhook", webhook.ID.String(), models.JSON{"name": req.Name, "url": req.URL})
	return webhook, nil
}

func (s *WorkspaceService) ListWebhooks(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.WorkspaceWebhook, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	return s.webhookRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *WorkspaceService) UpdateWebhook(ctx context.Context, workspaceID, webhookID, userID uuid.UUID, req *models.UpdateWebhookRequest) (*models.WorkspaceWebhook, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	webhook, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil || webhook == nil {
		return nil, ErrWebhookNotFound
	}

	if webhook.WorkspaceID != workspaceID {
		return nil, ErrNotAuthorized
	}

	if req.Name != nil {
		webhook.Name = *req.Name
	}
	if req.URL != nil {
		webhook.URL = *req.URL
	}
	if req.Events != nil {
		webhook.Events = models.JSON{"events": req.Events}
	}
	if req.IsActive != nil {
		webhook.IsActive = *req.IsActive
	}

	if err := s.webhookRepo.Update(ctx, webhook); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "webhook.updated", "webhook", webhookID.String(), nil)
	return webhook, nil
}

func (s *WorkspaceService) DeleteWebhook(ctx context.Context, workspaceID, webhookID, userID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	webhook, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil || webhook == nil {
		return ErrWebhookNotFound
	}

	if webhook.WorkspaceID != workspaceID {
		return ErrNotAuthorized
	}

	if err := s.webhookRepo.Delete(ctx, webhookID); err != nil {
		return err
	}

	s.LogActivity(ctx, workspaceID, userID, "webhook.deleted", "webhook", webhookID.String(), nil)
	return nil
}

func (s *WorkspaceService) TestWebhook(ctx context.Context, workspaceID, webhookID, userID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	webhook, err := s.webhookRepo.GetByID(ctx, webhookID)
	if err != nil || webhook == nil {
		return ErrWebhookNotFound
	}

	if webhook.WorkspaceID != workspaceID {
		return ErrNotAuthorized
	}

	payload := map[string]interface{}{
		"type":         "webhook.test",
		"workspace_id": workspaceID,
		"timestamp":    time.Now(),
	}

	if err := s.sendWebhookRequest(webhook.URL, webhook.Secret, payload); err != nil {
		s.webhookRepo.IncrementFailureCount(ctx, webhookID)
		return fmt.Errorf("webhook test failed: %w", err)
	}

	s.webhookRepo.UpdateLastTriggered(ctx, webhookID)
	s.webhookRepo.ResetFailureCount(ctx, webhookID)
	return nil
}

func (s *WorkspaceService) TriggerWebhooks(ctx context.Context, workspaceID uuid.UUID, eventType string, payload map[string]interface{}) {
	webhooks, err := s.webhookRepo.ListActiveByEvent(ctx, workspaceID, eventType)
	if err != nil || len(webhooks) == 0 {
		return
	}

	for _, webhook := range webhooks {
		go func(w *models.WorkspaceWebhook) {
			if err := s.sendWebhookRequest(w.URL, w.Secret, payload); err != nil {
				s.webhookRepo.IncrementFailureCount(ctx, w.ID)
				s.logger.WithError(err).WithField("webhook_id", w.ID).Warn("Failed to trigger webhook")
			} else {
				s.webhookRepo.UpdateLastTriggered(ctx, w.ID)
				s.webhookRepo.ResetFailureCount(ctx, w.ID)
			}
		}(webhook)
	}
}

func (s *WorkspaceService) sendWebhookRequest(url, secret string, payload map[string]interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// ── Workspace Favorites ──

func (s *WorkspaceService) FavoriteWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) (*models.WorkspaceFavorite, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	existing, _ := s.favoriteRepo.GetByUserAndWorkspace(ctx, userID, workspaceID)
	if existing != nil {
		return nil, ErrAlreadyFavorited
	}

	maxPos, _ := s.favoriteRepo.GetMaxPosition(ctx, userID)

	fav := &models.WorkspaceFavorite{
		ID:          uuid.New(),
		UserID:      userID,
		WorkspaceID: workspaceID,
		Position:    maxPos + 1,
		CreatedAt:   time.Now(),
	}

	if err := s.favoriteRepo.Create(ctx, fav); err != nil {
		return nil, err
	}

	return fav, nil
}

func (s *WorkspaceService) UnfavoriteWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) error {
	existing, _ := s.favoriteRepo.GetByUserAndWorkspace(ctx, userID, workspaceID)
	if existing == nil {
		return ErrNotFavorited
	}

	return s.favoriteRepo.Delete(ctx, userID, workspaceID)
}

func (s *WorkspaceService) ListFavorites(ctx context.Context, userID uuid.UUID) ([]*models.WorkspaceFavorite, error) {
	return s.favoriteRepo.ListByUser(ctx, userID)
}

func (s *WorkspaceService) ReorderFavorites(ctx context.Context, userID uuid.UUID, req *models.ReorderFavoritesRequest) error {
	var wsIDs []uuid.UUID
	for _, id := range req.WorkspaceIDs {
		parsed, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		wsIDs = append(wsIDs, parsed)
	}

	return s.favoriteRepo.UpdatePositions(ctx, userID, wsIDs)
}

// ── Audit Export ──

func (s *WorkspaceService) ExportAuditLog(ctx context.Context, workspaceID, userID uuid.UUID, req *models.AuditExportRequest) (*models.AuditExportResponse, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	activities, total, err := s.activityRepo.ListByDateRange(ctx, workspaceID, req.StartDate, req.EndDate, req.ActionType)
	if err != nil {
		return nil, err
	}

	return &models.AuditExportResponse{
		Activities: activities,
		Total:      total,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		ExportedAt: time.Now(),
	}, nil
}

// ── Member Notes ──

func (s *WorkspaceService) CreateMemberNote(ctx context.Context, workspaceID, targetID, authorID uuid.UUID, req *models.CreateMemberNoteRequest) (*models.MemberNote, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, authorID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, targetID)
	if !isMember {
		return nil, ErrNotMember
	}

	note := &models.MemberNote{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		TargetID:    targetID,
		AuthorID:    authorID,
		Content:     req.Content,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.memberNoteRepo.Create(ctx, note); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, authorID, "member_note.created", "member_note", note.ID.String(), models.JSON{"target_id": targetID})
	return note, nil
}

func (s *WorkspaceService) ListMemberNotes(ctx context.Context, workspaceID, targetID, userID uuid.UUID) ([]*models.MemberNote, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	return s.memberNoteRepo.ListByTarget(ctx, workspaceID, targetID)
}

func (s *WorkspaceService) UpdateMemberNote(ctx context.Context, workspaceID, noteID, userID uuid.UUID, req *models.UpdateMemberNoteRequest) (*models.MemberNote, error) {
	note, err := s.memberNoteRepo.GetByID(ctx, noteID)
	if err != nil || note == nil {
		return nil, ErrMemberNoteNotFound
	}

	if note.WorkspaceID != workspaceID {
		return nil, ErrNotAuthorized
	}

	if note.AuthorID != userID {
		return nil, ErrNotAuthorized
	}

	note.Content = req.Content

	if err := s.memberNoteRepo.Update(ctx, note); err != nil {
		return nil, err
	}

	return note, nil
}

func (s *WorkspaceService) DeleteMemberNote(ctx context.Context, workspaceID, noteID, userID uuid.UUID) error {
	note, err := s.memberNoteRepo.GetByID(ctx, noteID)
	if err != nil || note == nil {
		return ErrMemberNoteNotFound
	}

	if note.WorkspaceID != workspaceID {
		return ErrNotAuthorized
	}

	// Only author or owner can delete
	if note.AuthorID != userID {
		role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
		if role != "owner" {
			return ErrNotAuthorized
		}
	}

	return s.memberNoteRepo.Delete(ctx, noteID)
}

// ── Scheduled Actions ──

func (s *WorkspaceService) CreateScheduledAction(ctx context.Context, workspaceID, userID uuid.UUID, req *models.CreateScheduledActionRequest) (*models.ScheduledAction, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	if req.ScheduledAt.Before(time.Now()) {
		return nil, ErrScheduledActionPast
	}

	action := &models.ScheduledAction{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		ActionType:  req.ActionType,
		Payload:     req.Payload,
		ScheduledAt: req.ScheduledAt,
		Status:      "pending",
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.scheduledActionRepo.Create(ctx, action); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "scheduled_action.created", "scheduled_action", action.ID.String(), models.JSON{"action_type": req.ActionType, "scheduled_at": req.ScheduledAt})
	return action, nil
}

func (s *WorkspaceService) ListScheduledActions(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.ScheduledAction, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	return s.scheduledActionRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *WorkspaceService) UpdateScheduledAction(ctx context.Context, workspaceID, actionID, userID uuid.UUID, req *models.UpdateScheduledActionRequest) (*models.ScheduledAction, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	action, err := s.scheduledActionRepo.GetByID(ctx, actionID)
	if err != nil || action == nil {
		return nil, ErrScheduledActionNotFound
	}

	if action.WorkspaceID != workspaceID {
		return nil, ErrNotAuthorized
	}

	if action.Status != "pending" {
		return nil, fmt.Errorf("cannot update action with status: %s", action.Status)
	}

	if req.Payload != nil {
		action.Payload = req.Payload
	}
	if req.ScheduledAt != nil {
		if req.ScheduledAt.Before(time.Now()) {
			return nil, ErrScheduledActionPast
		}
		action.ScheduledAt = *req.ScheduledAt
	}

	if err := s.scheduledActionRepo.Update(ctx, action); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "scheduled_action.updated", "scheduled_action", actionID.String(), nil)
	return action, nil
}

func (s *WorkspaceService) CancelScheduledAction(ctx context.Context, workspaceID, actionID, userID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	action, err := s.scheduledActionRepo.GetByID(ctx, actionID)
	if err != nil || action == nil {
		return ErrScheduledActionNotFound
	}

	if action.WorkspaceID != workspaceID {
		return ErrNotAuthorized
	}

	if action.Status != "pending" {
		return fmt.Errorf("cannot cancel action with status: %s", action.Status)
	}

	if err := s.scheduledActionRepo.UpdateStatus(ctx, actionID, "cancelled"); err != nil {
		return err
	}

	s.LogActivity(ctx, workspaceID, userID, "scheduled_action.cancelled", "scheduled_action", actionID.String(), nil)
	return nil
}

func (s *WorkspaceService) DeleteScheduledAction(ctx context.Context, workspaceID, actionID, userID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	action, err := s.scheduledActionRepo.GetByID(ctx, actionID)
	if err != nil || action == nil {
		return ErrScheduledActionNotFound
	}

	if action.WorkspaceID != workspaceID {
		return ErrNotAuthorized
	}

	if err := s.scheduledActionRepo.Delete(ctx, actionID); err != nil {
		return err
	}

	s.LogActivity(ctx, workspaceID, userID, "scheduled_action.deleted", "scheduled_action", actionID.String(), nil)
	return nil
}

// ── Usage Quotas ──

func (s *WorkspaceService) GetQuotaUsage(ctx context.Context, workspaceID, userID uuid.UUID) (*models.QuotaUsageResponse, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	quota, _ := s.quotaRepo.GetByWorkspace(ctx, workspaceID)
	if quota == nil {
		// Return default quotas for free plan
		quota = &models.WorkspaceQuota{
			WorkspaceID:    workspaceID,
			MaxMembers:     100,
			MaxChannels:    50,
			MaxStorageMB:   5120,
			MaxInviteCodes: 10,
			MaxWebhooks:    5,
			MaxRoles:       10,
		}
	}

	// Calculate current usage
	memberCount, _ := s.workspaceRepo.GetMemberCount(ctx, workspaceID)
	inviteCodeCount := 0
	codes, _ := s.inviteCodeRepo.ListByWorkspace(ctx, workspaceID)
	if codes != nil {
		inviteCodeCount = len(codes)
	}
	webhooks, _ := s.webhookRepo.ListByWorkspace(ctx, workspaceID)
	webhookCount := 0
	if webhooks != nil {
		webhookCount = len(webhooks)
	}
	roles, _ := s.roleRepo.ListByWorkspace(ctx, workspaceID)
	roleCount := 0
	if roles != nil {
		roleCount = len(roles)
	}

	usage := map[string]int{
		"members":      memberCount,
		"invite_codes": inviteCodeCount,
		"webhooks":     webhookCount,
		"roles":        roleCount,
	}

	limits := map[string]int{
		"members":      quota.MaxMembers,
		"channels":     quota.MaxChannels,
		"storage_mb":   quota.MaxStorageMB,
		"invite_codes": quota.MaxInviteCodes,
		"webhooks":     quota.MaxWebhooks,
		"roles":        quota.MaxRoles,
	}

	percent := map[string]int{}
	for key, used := range usage {
		if limit, ok := limits[key]; ok && limit > 0 {
			percent[key] = (used * 100) / limit
		}
	}

	return &models.QuotaUsageResponse{
		Quota:   quota,
		Usage:   usage,
		Limits:  limits,
		Percent: percent,
	}, nil
}

func (s *WorkspaceService) UpdateQuota(ctx context.Context, workspaceID, userID uuid.UUID, req *models.UpdateQuotaRequest) (*models.WorkspaceQuota, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" {
		return nil, ErrNotAuthorized
	}

	existing, _ := s.quotaRepo.GetByWorkspace(ctx, workspaceID)
	now := time.Now()

	quota := &models.WorkspaceQuota{
		WorkspaceID:    workspaceID,
		MaxMembers:     100,
		MaxChannels:    50,
		MaxStorageMB:   5120,
		MaxInviteCodes: 10,
		MaxWebhooks:    5,
		MaxRoles:       10,
		UpdatedAt:      now,
	}

	if existing != nil {
		quota.ID = existing.ID
		quota.CreatedAt = existing.CreatedAt
		quota.MaxMembers = existing.MaxMembers
		quota.MaxChannels = existing.MaxChannels
		quota.MaxStorageMB = existing.MaxStorageMB
		quota.MaxInviteCodes = existing.MaxInviteCodes
		quota.MaxWebhooks = existing.MaxWebhooks
		quota.MaxRoles = existing.MaxRoles
		quota.CurrentMembers = existing.CurrentMembers
		quota.CurrentChannels = existing.CurrentChannels
		quota.CurrentStorageMB = existing.CurrentStorageMB
	} else {
		quota.ID = uuid.New()
		quota.CreatedAt = now
	}

	if req.MaxMembers != nil {
		quota.MaxMembers = *req.MaxMembers
	}
	if req.MaxChannels != nil {
		quota.MaxChannels = *req.MaxChannels
	}
	if req.MaxStorageMB != nil {
		quota.MaxStorageMB = *req.MaxStorageMB
	}
	if req.MaxInviteCodes != nil {
		quota.MaxInviteCodes = *req.MaxInviteCodes
	}
	if req.MaxWebhooks != nil {
		quota.MaxWebhooks = *req.MaxWebhooks
	}
	if req.MaxRoles != nil {
		quota.MaxRoles = *req.MaxRoles
	}

	if err := s.quotaRepo.Upsert(ctx, quota); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "quota.updated", "quota", workspaceID.String(), nil)
	return quota, nil
}

// ── Workspace Archive / Restore ──

func (s *WorkspaceService) ArchiveWorkspace(ctx context.Context, workspaceID, userID uuid.UUID, req *models.ArchiveWorkspaceRequest) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" {
		return ErrNotAuthorized
	}

	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil || workspace == nil {
		return ErrWorkspaceNotFound
	}

	if workspace.DeletedAt != nil {
		return ErrWorkspaceArchived
	}

	now := time.Now()
	workspace.DeletedAt = &now
	workspace.IsActive = false

	if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
		return err
	}

	s.invalidateWorkspace(ctx, workspaceID)
	s.LogActivity(ctx, workspaceID, userID, "workspace.archived", "workspace", workspaceID.String(), models.JSON{"reason": req.Reason})
	s.publishEvent(ctx, "workspace.events", workspaceID.String(), "workspace.archived", map[string]interface{}{"workspace_id": workspaceID, "archived_by": userID})
	return nil
}

func (s *WorkspaceService) RestoreWorkspace(ctx context.Context, workspaceID, userID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" {
		return ErrNotAuthorized
	}

	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil || workspace == nil {
		return ErrWorkspaceNotFound
	}

	if workspace.DeletedAt == nil {
		return ErrWorkspaceNotArchived
	}

	workspace.DeletedAt = nil
	workspace.IsActive = true

	if err := s.workspaceRepo.Update(ctx, workspace); err != nil {
		return err
	}

	s.invalidateWorkspace(ctx, workspaceID)
	s.LogActivity(ctx, workspaceID, userID, "workspace.restored", "workspace", workspaceID.String(), nil)
	s.publishEvent(ctx, "workspace.events", workspaceID.String(), "workspace.restored", map[string]interface{}{"workspace_id": workspaceID, "restored_by": userID})
	return nil
}

func (s *WorkspaceService) ListArchivedWorkspaces(ctx context.Context, userID uuid.UUID) ([]*models.Workspace, error) {
	return s.workspaceRepo.ListArchivedByUser(ctx, userID)
}

// ── Workspace Cloning ──

func (s *WorkspaceService) CloneWorkspace(ctx context.Context, sourceID, userID uuid.UUID, req *models.CloneWorkspaceRequest) (*models.Workspace, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, sourceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	existing, _ := s.workspaceRepo.GetBySlug(ctx, req.Slug)
	if existing != nil {
		return nil, ErrSlugExists
	}

	source, err := s.workspaceRepo.GetByID(ctx, sourceID)
	if err != nil || source == nil {
		return nil, ErrWorkspaceNotFound
	}

	newWorkspace := &models.Workspace{
		ID:          uuid.New(),
		Name:        req.Name,
		Slug:        req.Slug,
		Description: source.Description,
		OwnerID:     userID,
		Plan:        "free",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if req.IncludeSettings {
		newWorkspace.Settings = source.Settings
	}

	if err := s.workspaceRepo.Create(ctx, newWorkspace); err != nil {
		return nil, err
	}

	// Add creator as owner
	member := &models.WorkspaceMember{
		ID:          uuid.New(),
		WorkspaceID: newWorkspace.ID,
		UserID:      userID,
		Role:        "owner",
		JoinedAt:    time.Now(),
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	s.memberRepo.Create(ctx, member)

	// Clone roles
	if req.IncludeRoles {
		roles, _ := s.roleRepo.ListByWorkspace(ctx, sourceID)
		for _, r := range roles {
			newRole := &models.WorkspaceRole{
				ID:          uuid.New(),
				WorkspaceID: newWorkspace.ID,
				Name:        r.Name,
				Color:       r.Color,
				Priority:    r.Priority,
				Permissions: r.Permissions,
				IsDefault:   r.IsDefault,
				CreatedBy:   userID,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			s.roleRepo.Create(ctx, newRole)
		}
	}

	// Clone tags
	if req.IncludeTags {
		tags, _ := s.tagRepo.ListByWorkspace(ctx, sourceID)
		for _, t := range tags {
			newTag := &models.WorkspaceTag{
				ID:          uuid.New(),
				WorkspaceID: newWorkspace.ID,
				Name:        t.Name,
				Color:       t.Color,
				CreatedBy:   userID,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			s.tagRepo.Create(ctx, newTag)
		}
	}

	s.LogActivity(ctx, newWorkspace.ID, userID, "workspace.cloned", "workspace", newWorkspace.ID.String(), models.JSON{"source_id": sourceID})
	s.publishEvent(ctx, "workspace.events", newWorkspace.ID.String(), "workspace.created", map[string]interface{}{"workspace_id": newWorkspace.ID, "cloned_from": sourceID})
	return newWorkspace, nil
}

// ── Pinned Items ──

func (s *WorkspaceService) CreatePinnedItem(ctx context.Context, workspaceID, userID uuid.UUID, req *models.CreatePinnedItemRequest) (*models.WorkspacePinnedItem, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role == "" {
		return nil, ErrNotMember
	}

	maxPos, _ := s.pinnedItemRepo.GetMaxPosition(ctx, workspaceID)

	item := &models.WorkspacePinnedItem{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		ItemType:    req.ItemType,
		ItemID:      req.ItemID,
		Title:       req.Title,
		Content:     req.Content,
		URL:         req.URL,
		PinnedBy:    userID,
		Position:    maxPos + 1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.pinnedItemRepo.Create(ctx, item); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "pin.created", "pinned_item", item.ID.String(), models.JSON{"item_type": req.ItemType, "title": req.Title})
	return item, nil
}

func (s *WorkspaceService) ListPinnedItems(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.WorkspacePinnedItem, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	return s.pinnedItemRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *WorkspaceService) UpdatePinnedItem(ctx context.Context, workspaceID, pinID, userID uuid.UUID, req *models.UpdatePinnedItemRequest) (*models.WorkspacePinnedItem, error) {
	item, err := s.pinnedItemRepo.GetByID(ctx, pinID)
	if err != nil || item == nil {
		return nil, ErrPinnedItemNotFound
	}

	if item.WorkspaceID != workspaceID {
		return nil, ErrNotAuthorized
	}

	// Only pinner or admin/owner can update
	if item.PinnedBy != userID {
		role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
		if role != "owner" && role != "admin" {
			return nil, ErrNotAuthorized
		}
	}

	if req.Title != nil {
		item.Title = *req.Title
	}
	if req.Content != nil {
		item.Content = req.Content
	}
	if req.URL != nil {
		item.URL = req.URL
	}

	if err := s.pinnedItemRepo.Update(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *WorkspaceService) DeletePinnedItem(ctx context.Context, workspaceID, pinID, userID uuid.UUID) error {
	item, err := s.pinnedItemRepo.GetByID(ctx, pinID)
	if err != nil || item == nil {
		return ErrPinnedItemNotFound
	}

	if item.WorkspaceID != workspaceID {
		return ErrNotAuthorized
	}

	// Only pinner or admin/owner can delete
	if item.PinnedBy != userID {
		role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
		if role != "owner" && role != "admin" {
			return ErrNotAuthorized
		}
	}

	if err := s.pinnedItemRepo.Delete(ctx, pinID); err != nil {
		return err
	}

	s.LogActivity(ctx, workspaceID, userID, "pin.deleted", "pinned_item", pinID.String(), nil)
	return nil
}

func (s *WorkspaceService) ReorderPins(ctx context.Context, workspaceID, userID uuid.UUID, req *models.ReorderPinsRequest) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	var pinIDs []uuid.UUID
	for _, id := range req.PinIDs {
		parsed, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		pinIDs = append(pinIDs, parsed)
	}

	return s.pinnedItemRepo.UpdatePositions(ctx, workspaceID, pinIDs)
}

// ── Member Groups / Teams ──

func (s *WorkspaceService) CreateGroup(ctx context.Context, workspaceID, userID uuid.UUID, req *models.CreateGroupRequest) (*models.MemberGroup, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	existing, _ := s.groupRepo.GetByName(ctx, workspaceID, req.Name)
	if existing != nil {
		return nil, ErrGroupNameExists
	}

	group := &models.MemberGroup{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		CreatedBy:   userID,
		MemberCount: 0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "group.created", "group", group.ID.String(), models.JSON{"name": req.Name})
	return group, nil
}

func (s *WorkspaceService) ListGroups(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.MemberGroup, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	return s.groupRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *WorkspaceService) GetGroup(ctx context.Context, workspaceID, groupID, userID uuid.UUID) (*models.MemberGroupWithMembers, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil || group == nil {
		return nil, ErrGroupNotFound
	}

	if group.WorkspaceID != workspaceID {
		return nil, ErrNotAuthorized
	}

	members, _ := s.groupRepo.ListGroupMembers(ctx, groupID)

	return &models.MemberGroupWithMembers{
		MemberGroup: *group,
		Members:     members,
	}, nil
}

func (s *WorkspaceService) UpdateGroup(ctx context.Context, workspaceID, groupID, userID uuid.UUID, req *models.UpdateGroupRequest) (*models.MemberGroup, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil || group == nil {
		return nil, ErrGroupNotFound
	}

	if group.WorkspaceID != workspaceID {
		return nil, ErrNotAuthorized
	}

	if req.Name != nil {
		// Check for name conflict
		existing, _ := s.groupRepo.GetByName(ctx, workspaceID, *req.Name)
		if existing != nil && existing.ID != groupID {
			return nil, ErrGroupNameExists
		}
		group.Name = *req.Name
	}
	if req.Description != nil {
		group.Description = req.Description
	}
	if req.Color != nil {
		group.Color = req.Color
	}

	if err := s.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "group.updated", "group", groupID.String(), nil)
	return group, nil
}

func (s *WorkspaceService) DeleteGroup(ctx context.Context, workspaceID, groupID, userID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil || group == nil {
		return ErrGroupNotFound
	}

	if group.WorkspaceID != workspaceID {
		return ErrNotAuthorized
	}

	if err := s.groupRepo.Delete(ctx, groupID); err != nil {
		return err
	}

	s.LogActivity(ctx, workspaceID, userID, "group.deleted", "group", groupID.String(), models.JSON{"name": group.Name})
	return nil
}

func (s *WorkspaceService) AddGroupMembers(ctx context.Context, workspaceID, groupID, userID uuid.UUID, req *models.AddGroupMembersRequest) ([]uuid.UUID, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil || group == nil {
		return nil, ErrGroupNotFound
	}

	if group.WorkspaceID != workspaceID {
		return nil, ErrNotAuthorized
	}

	var added []uuid.UUID
	for _, uidStr := range req.UserIDs {
		uid, err := uuid.Parse(uidStr)
		if err != nil {
			continue
		}

		// Verify they are workspace members
		isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, uid)
		if !isMember {
			continue
		}

		// Check not already in group
		inGroup, _ := s.groupRepo.IsMemberOfGroup(ctx, groupID, uid)
		if inGroup {
			continue
		}

		membership := &models.MemberGroupMembership{
			ID:        uuid.New(),
			GroupID:   groupID,
			UserID:    uid,
			AddedBy:   userID,
			CreatedAt: time.Now(),
		}

		if err := s.groupRepo.AddMember(ctx, membership); err == nil {
			s.groupRepo.IncrementMemberCount(ctx, groupID)
			added = append(added, uid)
		}
	}

	s.LogActivity(ctx, workspaceID, userID, "group.members_added", "group", groupID.String(), models.JSON{"added_count": len(added)})
	return added, nil
}

func (s *WorkspaceService) RemoveGroupMember(ctx context.Context, workspaceID, groupID, targetID, userID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil || group == nil {
		return ErrGroupNotFound
	}

	if group.WorkspaceID != workspaceID {
		return ErrNotAuthorized
	}

	inGroup, _ := s.groupRepo.IsMemberOfGroup(ctx, groupID, targetID)
	if !inGroup {
		return ErrNotGroupMember
	}

	if err := s.groupRepo.RemoveMember(ctx, groupID, targetID); err != nil {
		return err
	}

	s.groupRepo.DecrementMemberCount(ctx, groupID)
	s.LogActivity(ctx, workspaceID, userID, "group.member_removed", "group", groupID.String(), models.JSON{"removed_user": targetID})
	return nil
}

func (s *WorkspaceService) ListUserGroups(ctx context.Context, workspaceID, targetID, userID uuid.UUID) ([]*models.MemberGroup, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	return s.groupRepo.ListGroupsByUser(ctx, workspaceID, targetID)
}

// ── Custom Fields ──

func (s *WorkspaceService) CreateCustomField(ctx context.Context, workspaceID, userID uuid.UUID, req *models.CreateCustomFieldRequest) (*models.WorkspaceCustomField, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	existing, _ := s.customFieldRepo.GetByName(ctx, workspaceID, req.Name)
	if existing != nil {
		return nil, ErrCustomFieldNameExists
	}

	maxPos, _ := s.customFieldRepo.GetMaxPosition(ctx, workspaceID)

	field := &models.WorkspaceCustomField{
		ID:           uuid.New(),
		WorkspaceID:  workspaceID,
		Name:         req.Name,
		FieldType:    req.FieldType,
		Options:      req.Options,
		DefaultValue: req.DefaultValue,
		IsRequired:   req.IsRequired,
		Position:     maxPos + 1,
		CreatedBy:    userID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.customFieldRepo.Create(ctx, field); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "custom_field.created", "custom_field", field.ID.String(), models.JSON{"name": req.Name, "type": req.FieldType})
	return field, nil
}

func (s *WorkspaceService) ListCustomFields(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.WorkspaceCustomField, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	return s.customFieldRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *WorkspaceService) UpdateCustomField(ctx context.Context, workspaceID, fieldID, userID uuid.UUID, req *models.UpdateCustomFieldRequest) (*models.WorkspaceCustomField, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	field, err := s.customFieldRepo.GetByID(ctx, fieldID)
	if err != nil || field == nil {
		return nil, ErrCustomFieldNotFound
	}

	if field.WorkspaceID != workspaceID {
		return nil, ErrNotAuthorized
	}

	if req.Name != nil {
		existing, _ := s.customFieldRepo.GetByName(ctx, workspaceID, *req.Name)
		if existing != nil && existing.ID != fieldID {
			return nil, ErrCustomFieldNameExists
		}
		field.Name = *req.Name
	}
	if req.Options != nil {
		field.Options = req.Options
	}
	if req.DefaultValue != nil {
		field.DefaultValue = req.DefaultValue
	}
	if req.IsRequired != nil {
		field.IsRequired = *req.IsRequired
	}

	if err := s.customFieldRepo.Update(ctx, field); err != nil {
		return nil, err
	}

	s.LogActivity(ctx, workspaceID, userID, "custom_field.updated", "custom_field", fieldID.String(), nil)
	return field, nil
}

func (s *WorkspaceService) DeleteCustomField(ctx context.Context, workspaceID, fieldID, userID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	field, err := s.customFieldRepo.GetByID(ctx, fieldID)
	if err != nil || field == nil {
		return ErrCustomFieldNotFound
	}

	if field.WorkspaceID != workspaceID {
		return ErrNotAuthorized
	}

	if err := s.customFieldRepo.Delete(ctx, fieldID); err != nil {
		return err
	}

	s.LogActivity(ctx, workspaceID, userID, "custom_field.deleted", "custom_field", fieldID.String(), models.JSON{"name": field.Name})
	return nil
}

func (s *WorkspaceService) SetCustomFieldValue(ctx context.Context, workspaceID, fieldID, entityID, userID uuid.UUID, req *models.SetCustomFieldValueRequest) (*models.WorkspaceCustomFieldValue, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	field, err := s.customFieldRepo.GetByID(ctx, fieldID)
	if err != nil || field == nil {
		return nil, ErrCustomFieldNotFound
	}

	if field.WorkspaceID != workspaceID {
		return nil, ErrNotAuthorized
	}

	value := &models.WorkspaceCustomFieldValue{
		ID:        uuid.New(),
		FieldID:   fieldID,
		EntityID:  entityID,
		Value:     req.Value,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.customFieldRepo.SetValue(ctx, value); err != nil {
		return nil, err
	}

	return value, nil
}

func (s *WorkspaceService) GetCustomFieldValues(ctx context.Context, workspaceID, entityID, userID uuid.UUID) ([]*models.CustomFieldWithValue, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	fields, _ := s.customFieldRepo.ListByWorkspace(ctx, workspaceID)
	values, _ := s.customFieldRepo.ListValuesByEntity(ctx, entityID)

	valueMap := make(map[uuid.UUID]string)
	for _, v := range values {
		valueMap[v.FieldID] = v.Value
	}

	var results []*models.CustomFieldWithValue
	for _, f := range fields {
		item := &models.CustomFieldWithValue{
			WorkspaceCustomField: *f,
		}
		if val, ok := valueMap[f.ID]; ok {
			item.Value = &val
		}
		results = append(results, item)
	}

	return results, nil
}

// ── Reactions ──

func (s *WorkspaceService) AddReaction(ctx context.Context, workspaceID, userID uuid.UUID, req *models.AddReactionRequest) error {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return ErrNotMember
	}

	entityID, err := uuid.Parse(req.EntityID)
	if err != nil {
		return fmt.Errorf("invalid entity ID")
	}

	exists, err := s.reactionRepo.Exists(ctx, req.EntityType, entityID, userID, req.Emoji)
	if err != nil {
		return err
	}
	if exists {
		return ErrReactionExists
	}

	reaction := &models.WorkspaceReaction{
		ID:         uuid.New(),
		EntityType: req.EntityType,
		EntityID:   entityID,
		UserID:     userID,
		Emoji:      req.Emoji,
		CreatedAt:  time.Now(),
	}

	return s.reactionRepo.Create(ctx, reaction)
}

func (s *WorkspaceService) RemoveReaction(ctx context.Context, workspaceID, userID uuid.UUID, entityType string, entityID uuid.UUID, emoji string) error {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return ErrNotMember
	}
	return s.reactionRepo.Delete(ctx, entityType, entityID, userID, emoji)
}

func (s *WorkspaceService) ListReactions(ctx context.Context, workspaceID, userID uuid.UUID, entityType string, entityID uuid.UUID) ([]*models.WorkspaceReaction, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	return s.reactionRepo.ListByEntity(ctx, entityType, entityID)
}

func (s *WorkspaceService) GetReactionSummary(ctx context.Context, workspaceID, userID uuid.UUID, entityType string, entityID uuid.UUID) ([]models.ReactionSummary, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	return s.reactionRepo.GetSummary(ctx, entityType, entityID)
}

// ── Bookmarks ──

func (s *WorkspaceService) CreateBookmark(ctx context.Context, workspaceID, userID uuid.UUID, req *models.CreateBookmarkRequest) (*models.WorkspaceBookmark, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	count, err := s.bookmarkRepo.CountByUser(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if count >= 100 {
		return nil, ErrBookmarkLimitReached
	}

	maxPos, _ := s.bookmarkRepo.GetMaxPosition(ctx, workspaceID, userID)

	bookmark := &models.WorkspaceBookmark{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		UserID:      userID,
		Title:       req.Title,
		URL:         req.URL,
		EntityType:  req.EntityType,
		EntityID:    req.EntityID,
		Notes:       req.Notes,
		FolderName:  req.FolderName,
		Position:    maxPos + 1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.bookmarkRepo.Create(ctx, bookmark); err != nil {
		return nil, err
	}
	return bookmark, nil
}

func (s *WorkspaceService) ListBookmarks(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.WorkspaceBookmark, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	return s.bookmarkRepo.ListByUser(ctx, workspaceID, userID)
}

func (s *WorkspaceService) ListBookmarksByFolder(ctx context.Context, workspaceID, userID uuid.UUID, folderName string) ([]*models.WorkspaceBookmark, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	return s.bookmarkRepo.ListByFolder(ctx, workspaceID, userID, folderName)
}

func (s *WorkspaceService) ListBookmarkFolders(ctx context.Context, workspaceID, userID uuid.UUID) ([]string, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	return s.bookmarkRepo.ListFolders(ctx, workspaceID, userID)
}

func (s *WorkspaceService) UpdateBookmark(ctx context.Context, workspaceID, userID, bookmarkID uuid.UUID, req *models.UpdateBookmarkRequest) (*models.WorkspaceBookmark, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	bookmark, err := s.bookmarkRepo.GetByID(ctx, bookmarkID)
	if err != nil {
		return nil, err
	}
	if bookmark == nil || bookmark.UserID != userID {
		return nil, ErrBookmarkNotFound
	}

	if req.Title != nil {
		bookmark.Title = *req.Title
	}
	if req.URL != nil {
		bookmark.URL = req.URL
	}
	if req.Notes != nil {
		bookmark.Notes = req.Notes
	}
	if req.FolderName != nil {
		bookmark.FolderName = req.FolderName
	}

	if err := s.bookmarkRepo.Update(ctx, bookmark); err != nil {
		return nil, err
	}
	return bookmark, nil
}

func (s *WorkspaceService) DeleteBookmark(ctx context.Context, workspaceID, userID, bookmarkID uuid.UUID) error {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return ErrNotMember
	}

	bookmark, err := s.bookmarkRepo.GetByID(ctx, bookmarkID)
	if err != nil {
		return err
	}
	if bookmark == nil || bookmark.UserID != userID {
		return ErrBookmarkNotFound
	}

	return s.bookmarkRepo.Delete(ctx, bookmarkID)
}

// ── Invitation History ──

func (s *WorkspaceService) RecordInvitation(ctx context.Context, workspaceID, inviterID uuid.UUID, inviteeEmail string, inviteeID *uuid.UUID, method, role string, expiresAt *time.Time) error {
	record := &models.InvitationHistory{
		ID:           uuid.New(),
		WorkspaceID:  workspaceID,
		InviterID:    inviterID,
		InviteeEmail: inviteeEmail,
		InviteeID:    inviteeID,
		Method:       method,
		Role:         role,
		Status:       "pending",
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}
	return s.invitationHistoryRepo.Create(ctx, record)
}

func (s *WorkspaceService) ListInvitationHistory(ctx context.Context, workspaceID, userID uuid.UUID, page, perPage int) ([]*models.InvitationHistory, int64, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, 0, ErrNotAuthorized
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.invitationHistoryRepo.ListByWorkspace(ctx, workspaceID, page, perPage)
}

func (s *WorkspaceService) GetInvitationStats(ctx context.Context, workspaceID, userID uuid.UUID) (*models.InvitationStats, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}
	return s.invitationHistoryRepo.GetStats(ctx, workspaceID)
}

// ── Access Logs ──

func (s *WorkspaceService) LogAccess(ctx context.Context, workspaceID, userID uuid.UUID, action, resource, ipAddress, userAgent string) error {
	var ua *string
	if userAgent != "" {
		ua = &userAgent
	}
	log := &models.WorkspaceAccessLog{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		UserID:      userID,
		Action:      action,
		Resource:    resource,
		IPAddress:   ipAddress,
		UserAgent:   ua,
		CreatedAt:   time.Now(),
	}
	return s.accessLogRepo.Create(ctx, log)
}

func (s *WorkspaceService) ListAccessLogs(ctx context.Context, workspaceID, userID uuid.UUID, page, perPage int) ([]*models.WorkspaceAccessLog, int64, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, 0, ErrNotAuthorized
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.accessLogRepo.ListByWorkspace(ctx, workspaceID, page, perPage)
}

func (s *WorkspaceService) ListAccessLogsByUser(ctx context.Context, workspaceID, requesterID, targetUserID uuid.UUID, page, perPage int) ([]*models.WorkspaceAccessLog, int64, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, requesterID)
	if role != "owner" && role != "admin" {
		return nil, 0, ErrNotAuthorized
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.accessLogRepo.ListByUser(ctx, workspaceID, targetUserID, page, perPage)
}

func (s *WorkspaceService) GetAccessLogStats(ctx context.Context, workspaceID, userID uuid.UUID, days int) (*models.AccessLogStats, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}
	if days < 1 || days > 90 {
		days = 30
	}
	return s.accessLogRepo.GetStats(ctx, workspaceID, days)
}

// ── Feature Flags ──

func (s *WorkspaceService) CreateFeatureFlag(ctx context.Context, workspaceID, userID uuid.UUID, req *models.CreateFeatureFlagRequest) (*models.WorkspaceFeatureFlag, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	existing, _ := s.featureFlagRepo.GetByKey(ctx, workspaceID, req.Key)
	if existing != nil {
		return nil, ErrFeatureFlagKeyExists
	}

	flag := &models.WorkspaceFeatureFlag{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Key:         req.Key,
		Enabled:     req.Enabled,
		Description: req.Description,
		Metadata:    req.Metadata,
		CreatedBy:   userID,
		UpdatedBy:   &userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.featureFlagRepo.Create(ctx, flag); err != nil {
		return nil, err
	}
	return flag, nil
}

func (s *WorkspaceService) ListFeatureFlags(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.WorkspaceFeatureFlag, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	return s.featureFlagRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *WorkspaceService) UpdateFeatureFlag(ctx context.Context, workspaceID, userID, flagID uuid.UUID, req *models.UpdateFeatureFlagRequest) (*models.WorkspaceFeatureFlag, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	flag, err := s.featureFlagRepo.GetByID(ctx, flagID)
	if err != nil {
		return nil, err
	}
	if flag == nil || flag.WorkspaceID != workspaceID {
		return nil, ErrFeatureFlagNotFound
	}

	if req.Enabled != nil {
		flag.Enabled = *req.Enabled
	}
	if req.Description != nil {
		flag.Description = req.Description
	}
	if req.Metadata != nil {
		flag.Metadata = req.Metadata
	}
	flag.UpdatedBy = &userID

	if err := s.featureFlagRepo.Update(ctx, flag); err != nil {
		return nil, err
	}
	return flag, nil
}

func (s *WorkspaceService) DeleteFeatureFlag(ctx context.Context, workspaceID, userID, flagID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	flag, err := s.featureFlagRepo.GetByID(ctx, flagID)
	if err != nil {
		return err
	}
	if flag == nil || flag.WorkspaceID != workspaceID {
		return ErrFeatureFlagNotFound
	}

	return s.featureFlagRepo.Delete(ctx, flagID)
}

func (s *WorkspaceService) CheckFeatureFlag(ctx context.Context, workspaceID, userID uuid.UUID, key string) (*models.FeatureFlagCheckResponse, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	flag, err := s.featureFlagRepo.GetByKey(ctx, workspaceID, key)
	if err != nil {
		return nil, err
	}
	if flag == nil {
		return &models.FeatureFlagCheckResponse{Key: key, Enabled: false}, nil
	}

	return &models.FeatureFlagCheckResponse{Key: key, Enabled: flag.Enabled}, nil
}

// ── Integrations ──

func (s *WorkspaceService) CreateIntegration(ctx context.Context, workspaceID, userID uuid.UUID, req *models.CreateIntegrationRequest) (*models.WorkspaceIntegration, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	integration := &models.WorkspaceIntegration{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Provider:    req.Provider,
		Name:        req.Name,
		Status:      "active",
		Config:      req.Config,
		Credentials: req.Credentials,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.integrationRepo.Create(ctx, integration); err != nil {
		return nil, err
	}
	return integration, nil
}

func (s *WorkspaceService) ListIntegrations(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.WorkspaceIntegration, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	return s.integrationRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *WorkspaceService) GetIntegration(ctx context.Context, workspaceID, userID, integrationID uuid.UUID) (*models.WorkspaceIntegration, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	integration, err := s.integrationRepo.GetByID(ctx, integrationID)
	if err != nil {
		return nil, err
	}
	if integration == nil || integration.WorkspaceID != workspaceID {
		return nil, ErrIntegrationNotFound
	}
	return integration, nil
}

func (s *WorkspaceService) UpdateIntegration(ctx context.Context, workspaceID, userID, integrationID uuid.UUID, req *models.UpdateIntegrationRequest) (*models.WorkspaceIntegration, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	integration, err := s.integrationRepo.GetByID(ctx, integrationID)
	if err != nil {
		return nil, err
	}
	if integration == nil || integration.WorkspaceID != workspaceID {
		return nil, ErrIntegrationNotFound
	}

	if req.Name != nil {
		integration.Name = *req.Name
	}
	if req.Status != nil {
		integration.Status = *req.Status
	}
	if req.Config != nil {
		integration.Config = req.Config
	}
	if req.Credentials != nil {
		integration.Credentials = req.Credentials
	}

	if err := s.integrationRepo.Update(ctx, integration); err != nil {
		return nil, err
	}
	return integration, nil
}

func (s *WorkspaceService) DeleteIntegration(ctx context.Context, workspaceID, userID, integrationID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	integration, err := s.integrationRepo.GetByID(ctx, integrationID)
	if err != nil {
		return err
	}
	if integration == nil || integration.WorkspaceID != workspaceID {
		return ErrIntegrationNotFound
	}

	return s.integrationRepo.Delete(ctx, integrationID)
}

// ── Labels ──

func (s *WorkspaceService) CreateLabel(ctx context.Context, workspaceID, userID uuid.UUID, req *models.CreateLabelRequest) (*models.WorkspaceLabel, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	existing, _ := s.labelRepo.GetByName(ctx, workspaceID, req.Name)
	if existing != nil {
		return nil, ErrLabelNameExists
	}

	maxPos, _ := s.labelRepo.GetMaxPosition(ctx, workspaceID)

	label := &models.WorkspaceLabel{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Color:       req.Color,
		Description: req.Description,
		Position:    maxPos + 1,
		UsageCount:  0,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.labelRepo.Create(ctx, label); err != nil {
		return nil, err
	}
	return label, nil
}

func (s *WorkspaceService) ListLabels(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.WorkspaceLabel, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	return s.labelRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *WorkspaceService) UpdateLabel(ctx context.Context, workspaceID, userID, labelID uuid.UUID, req *models.UpdateLabelRequest) (*models.WorkspaceLabel, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	label, err := s.labelRepo.GetByID(ctx, labelID)
	if err != nil {
		return nil, err
	}
	if label == nil || label.WorkspaceID != workspaceID {
		return nil, ErrLabelNotFound
	}

	if req.Name != nil {
		existing, _ := s.labelRepo.GetByName(ctx, workspaceID, *req.Name)
		if existing != nil && existing.ID != labelID {
			return nil, ErrLabelNameExists
		}
		label.Name = *req.Name
	}
	if req.Color != nil {
		label.Color = *req.Color
	}
	if req.Description != nil {
		label.Description = req.Description
	}

	if err := s.labelRepo.Update(ctx, label); err != nil {
		return nil, err
	}
	return label, nil
}

func (s *WorkspaceService) DeleteLabel(ctx context.Context, workspaceID, userID, labelID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	label, err := s.labelRepo.GetByID(ctx, labelID)
	if err != nil {
		return err
	}
	if label == nil || label.WorkspaceID != workspaceID {
		return ErrLabelNotFound
	}

	return s.labelRepo.Delete(ctx, labelID)
}

// ── Activity Streaks ──

func (s *WorkspaceService) RecordActivity(ctx context.Context, workspaceID, userID uuid.UUID) error {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return ErrNotMember
	}
	return s.streakRepo.RecordDailyActivity(ctx, workspaceID, userID)
}

func (s *WorkspaceService) GetMyStreak(ctx context.Context, workspaceID, userID uuid.UUID) (*models.MemberActivityStreak, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	streak, err := s.streakRepo.GetByUserID(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if streak == nil {
		// Return empty streak
		return &models.MemberActivityStreak{
			WorkspaceID: workspaceID,
			UserID:      userID,
		}, nil
	}
	return streak, nil
}

func (s *WorkspaceService) GetStreakLeaderboard(ctx context.Context, workspaceID, userID uuid.UUID, limit int) ([]models.StreakLeaderboard, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.streakRepo.GetLeaderboard(ctx, workspaceID, limit)
}

// ── Onboarding Checklists ──

func (s *WorkspaceService) CreateChecklist(ctx context.Context, workspaceID, userID uuid.UUID, req *models.CreateChecklistRequest) (*models.OnboardingChecklist, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	checklist := &models.OnboardingChecklist{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Title:       req.Title,
		Description: req.Description,
		IsActive:    true,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.onboardingRepo.CreateChecklist(ctx, checklist); err != nil {
		return nil, err
	}
	return checklist, nil
}

func (s *WorkspaceService) ListChecklists(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.OnboardingChecklist, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	return s.onboardingRepo.ListChecklists(ctx, workspaceID)
}

func (s *WorkspaceService) GetChecklistWithSteps(ctx context.Context, workspaceID, userID, checklistID uuid.UUID) (*models.ChecklistWithSteps, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	checklist, err := s.onboardingRepo.GetChecklistByID(ctx, checklistID)
	if err != nil {
		return nil, err
	}
	if checklist == nil || checklist.WorkspaceID != workspaceID {
		return nil, ErrChecklistNotFound
	}

	steps, _ := s.onboardingRepo.ListSteps(ctx, checklistID)

	return &models.ChecklistWithSteps{
		OnboardingChecklist: *checklist,
		Steps:               steps,
	}, nil
}

func (s *WorkspaceService) UpdateChecklist(ctx context.Context, workspaceID, userID, checklistID uuid.UUID, req *models.UpdateChecklistRequest) (*models.OnboardingChecklist, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	checklist, err := s.onboardingRepo.GetChecklistByID(ctx, checklistID)
	if err != nil {
		return nil, err
	}
	if checklist == nil || checklist.WorkspaceID != workspaceID {
		return nil, ErrChecklistNotFound
	}

	if req.Title != nil {
		checklist.Title = *req.Title
	}
	if req.Description != nil {
		checklist.Description = req.Description
	}
	if req.IsActive != nil {
		checklist.IsActive = *req.IsActive
	}

	if err := s.onboardingRepo.UpdateChecklist(ctx, checklist); err != nil {
		return nil, err
	}
	return checklist, nil
}

func (s *WorkspaceService) DeleteChecklist(ctx context.Context, workspaceID, userID, checklistID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	checklist, err := s.onboardingRepo.GetChecklistByID(ctx, checklistID)
	if err != nil {
		return err
	}
	if checklist == nil || checklist.WorkspaceID != workspaceID {
		return ErrChecklistNotFound
	}

	return s.onboardingRepo.DeleteChecklist(ctx, checklistID)
}

func (s *WorkspaceService) AddOnboardingStep(ctx context.Context, workspaceID, userID, checklistID uuid.UUID, req *models.AddStepRequest) (*models.OnboardingStep, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	checklist, err := s.onboardingRepo.GetChecklistByID(ctx, checklistID)
	if err != nil {
		return nil, err
	}
	if checklist == nil || checklist.WorkspaceID != workspaceID {
		return nil, ErrChecklistNotFound
	}

	maxPos, _ := s.onboardingRepo.GetMaxStepPosition(ctx, checklistID)

	step := &models.OnboardingStep{
		ID:          uuid.New(),
		ChecklistID: checklistID,
		Title:       req.Title,
		Description: req.Description,
		ActionType:  req.ActionType,
		ActionData:  req.ActionData,
		Position:    maxPos + 1,
		IsRequired:  req.IsRequired,
		CreatedAt:   time.Now(),
	}

	if err := s.onboardingRepo.AddStep(ctx, step); err != nil {
		return nil, err
	}
	return step, nil
}

func (s *WorkspaceService) DeleteOnboardingStep(ctx context.Context, workspaceID, userID, stepID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	step, err := s.onboardingRepo.GetStepByID(ctx, stepID)
	if err != nil {
		return err
	}
	if step == nil {
		return ErrOnboardingStepNotFound
	}

	// Verify the step belongs to this workspace
	checklist, err := s.onboardingRepo.GetChecklistByID(ctx, step.ChecklistID)
	if err != nil || checklist == nil || checklist.WorkspaceID != workspaceID {
		return ErrOnboardingStepNotFound
	}

	return s.onboardingRepo.DeleteStep(ctx, stepID)
}

func (s *WorkspaceService) CompleteOnboardingStep(ctx context.Context, workspaceID, userID, stepID uuid.UUID) error {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return ErrNotMember
	}

	step, err := s.onboardingRepo.GetStepByID(ctx, stepID)
	if err != nil {
		return err
	}
	if step == nil {
		return ErrOnboardingStepNotFound
	}

	now := time.Now()
	progress := &models.OnboardingProgress{
		ID:          uuid.New(),
		StepID:      stepID,
		UserID:      userID,
		CompletedAt: &now,
		CreatedAt:   now,
	}

	return s.onboardingRepo.CompleteStep(ctx, progress)
}

func (s *WorkspaceService) GetMyOnboardingStatus(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.UserOnboardingStatus, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}

	checklists, err := s.onboardingRepo.ListChecklists(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var statuses []*models.UserOnboardingStatus
	for _, cl := range checklists {
		if !cl.IsActive {
			continue
		}

		steps, _ := s.onboardingRepo.ListSteps(ctx, cl.ID)
		progress, _ := s.onboardingRepo.GetProgress(ctx, cl.ID, userID)

		progressMap := make(map[uuid.UUID]*models.OnboardingProgress)
		for _, p := range progress {
			progressMap[p.StepID] = p
		}

		var stepsWithProgress []models.StepWithProgress
		completedCount := 0
		for _, step := range steps {
			swp := models.StepWithProgress{
				OnboardingStep: step,
			}
			if p, ok := progressMap[step.ID]; ok {
				swp.Completed = true
				swp.CompletedAt = p.CompletedAt
				completedCount++
			}
			stepsWithProgress = append(stepsWithProgress, swp)
		}

		statuses = append(statuses, &models.UserOnboardingStatus{
			Checklist:      *cl,
			Steps:          stepsWithProgress,
			CompletedCount: completedCount,
			TotalSteps:     len(steps),
			IsComplete:     completedCount >= len(steps),
		})
	}

	return statuses, nil
}

// ── Compliance Policies ──

func (s *WorkspaceService) CreatePolicy(ctx context.Context, workspaceID, userID uuid.UUID, req *models.CreatePolicyRequest) (*models.CompliancePolicy, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	policy := &models.CompliancePolicy{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Description: req.Description,
		PolicyType:  req.PolicyType,
		Rules:       req.Rules,
		Severity:    req.Severity,
		IsEnforced:  req.IsEnforced,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.complianceRepo.Create(ctx, policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *WorkspaceService) ListPolicies(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.CompliancePolicy, error) {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return nil, ErrNotMember
	}
	return s.complianceRepo.ListByWorkspace(ctx, workspaceID)
}

func (s *WorkspaceService) UpdatePolicy(ctx context.Context, workspaceID, userID, policyID uuid.UUID, req *models.UpdatePolicyRequest) (*models.CompliancePolicy, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	policy, err := s.complianceRepo.GetByID(ctx, policyID)
	if err != nil {
		return nil, err
	}
	if policy == nil || policy.WorkspaceID != workspaceID {
		return nil, ErrPolicyNotFound
	}

	if req.Name != nil {
		policy.Name = *req.Name
	}
	if req.Description != nil {
		policy.Description = req.Description
	}
	if req.Rules != nil {
		policy.Rules = req.Rules
	}
	if req.Severity != nil {
		policy.Severity = *req.Severity
	}
	if req.IsEnforced != nil {
		policy.IsEnforced = *req.IsEnforced
	}

	if err := s.complianceRepo.Update(ctx, policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *WorkspaceService) DeletePolicy(ctx context.Context, workspaceID, userID, policyID uuid.UUID) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return ErrNotAuthorized
	}

	policy, err := s.complianceRepo.GetByID(ctx, policyID)
	if err != nil {
		return err
	}
	if policy == nil || policy.WorkspaceID != workspaceID {
		return ErrPolicyNotFound
	}

	return s.complianceRepo.Delete(ctx, policyID)
}

func (s *WorkspaceService) AcknowledgePolicy(ctx context.Context, workspaceID, userID, policyID uuid.UUID) error {
	isMember, _ := s.memberRepo.IsMember(ctx, workspaceID, userID)
	if !isMember {
		return ErrNotMember
	}

	policy, err := s.complianceRepo.GetByID(ctx, policyID)
	if err != nil {
		return err
	}
	if policy == nil || policy.WorkspaceID != workspaceID {
		return ErrPolicyNotFound
	}

	ack := &models.PolicyAcknowledgement{
		ID:       uuid.New(),
		PolicyID: policyID,
		UserID:   userID,
		AckedAt:  time.Now(),
	}

	return s.complianceRepo.Acknowledge(ctx, ack)
}

func (s *WorkspaceService) GetPolicyComplianceStatus(ctx context.Context, workspaceID, userID, policyID uuid.UUID) (*models.PolicyComplianceStatus, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, userID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	policy, err := s.complianceRepo.GetByID(ctx, policyID)
	if err != nil {
		return nil, err
	}
	if policy == nil || policy.WorkspaceID != workspaceID {
		return nil, ErrPolicyNotFound
	}

	// Get total member count
	_, total, _ := s.memberRepo.ListByWorkspace(ctx, workspaceID, 1, 1)
	totalMembers := int(total)

	ackedCount, _ := s.complianceRepo.GetAcknowledgementCount(ctx, policyID)

	complianceRate := float64(0)
	if totalMembers > 0 {
		complianceRate = float64(ackedCount) / float64(totalMembers) * 100
	}

	return &models.PolicyComplianceStatus{
		Policy:            *policy,
		TotalMembers:      totalMembers,
		AcknowledgedCount: ackedCount,
		ComplianceRate:    complianceRate,
	}, nil
}

// ── Redis Cache Helpers ──

func (s *WorkspaceService) cacheWorkspace(ctx context.Context, id uuid.UUID, workspace *models.Workspace) {
	if s.redis == nil {
		return
	}
	data, err := json.Marshal(workspace)
	if err != nil {
		return
	}
	key := fmt.Sprintf(cacheKeyWorkspace, id.String())
	s.redis.Set(ctx, key, data, cacheTTL)
}

func (s *WorkspaceService) getCachedWorkspaceResponse(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.WorkspaceResponse, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("no redis")
	}
	key := fmt.Sprintf(cacheKeyWorkspace, id.String())
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var workspace models.Workspace
	if err := json.Unmarshal(data, &workspace); err != nil {
		return nil, err
	}
	memberCount, _ := s.workspaceRepo.GetMemberCount(ctx, id)
	role, _ := s.memberRepo.GetRole(ctx, id, userID)
	return &models.WorkspaceResponse{
		Workspace:   &workspace,
		MemberCount: memberCount,
		MyRole:      role,
	}, nil
}

func (s *WorkspaceService) cacheStats(ctx context.Context, workspaceID uuid.UUID, stats *models.WorkspaceStats) {
	if s.redis == nil {
		return
	}
	data, err := json.Marshal(stats)
	if err != nil {
		return
	}
	key := fmt.Sprintf(cacheKeyStats, workspaceID.String())
	s.redis.Set(ctx, key, data, cacheTTL)
}

func (s *WorkspaceService) getCachedStats(ctx context.Context, workspaceID uuid.UUID) (*models.WorkspaceStats, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("no redis")
	}
	key := fmt.Sprintf(cacheKeyStats, workspaceID.String())
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var stats models.WorkspaceStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

func (s *WorkspaceService) invalidateWorkspace(ctx context.Context, workspaceID uuid.UUID) {
	if s.redis == nil {
		return
	}
	keys := []string{
		fmt.Sprintf(cacheKeyWorkspace, workspaceID.String()),
		fmt.Sprintf(cacheKeyMembers, workspaceID.String()),
		fmt.Sprintf(cacheKeyStats, workspaceID.String()),
	}
	s.redis.Del(ctx, keys...)
}

func (s *WorkspaceService) invalidateUserWorkspaces(ctx context.Context, userID uuid.UUID) {
	if s.redis == nil {
		return
	}
	key := fmt.Sprintf(cacheKeyUserWsList, userID.String())
	s.redis.Del(ctx, key)
}

// ── Kafka Event Helpers ──

func (s *WorkspaceService) publishEvent(ctx context.Context, topic, key, eventType string, data map[string]interface{}) {
	if s.kafka == nil {
		return
	}
	data["type"] = eventType
	data["timestamp"] = time.Now()
	if err := s.kafka.Publish(ctx, topic, key, data); err != nil {
		s.logger.WithError(err).WithField("event_type", eventType).Warn("Failed to publish event")
	}
}

// ── Token/Code Generators ──

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateInviteCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var sb strings.Builder
	for i := 0; i < 8; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		sb.WriteByte(charset[n.Int64()])
	}
	return sb.String()
}
