package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/quckchat/workspace-service/internal/db"
	"github.com/quckchat/workspace-service/internal/models"
	"github.com/quckchat/workspace-service/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrSlugExists        = errors.New("slug already exists")
	ErrNotMember         = errors.New("not a member of this workspace")
	ErrNotAuthorized     = errors.New("not authorized")
	ErrInviteNotFound    = errors.New("invite not found or expired")
	ErrAlreadyMember     = errors.New("already a member")
)

type WorkspaceService struct {
	workspaceRepo *repository.WorkspaceRepository
	memberRepo    *repository.MemberRepository
	inviteRepo    *repository.InviteRepository
	redis         *redis.Client
	kafka         *db.KafkaProducer
	logger        *logrus.Logger
}

func NewWorkspaceService(
	workspaceRepo *repository.WorkspaceRepository,
	memberRepo *repository.MemberRepository,
	inviteRepo *repository.InviteRepository,
	redis *redis.Client,
	kafka *db.KafkaProducer,
	logger *logrus.Logger,
) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo: workspaceRepo,
		memberRepo:    memberRepo,
		inviteRepo:    inviteRepo,
		redis:         redis,
		kafka:         kafka,
		logger:        logger,
	}
}

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

	// Add owner as member
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

	// Publish event
	if s.kafka != nil {
		s.kafka.Publish(ctx, "workspace-events", workspace.ID.String(), map[string]interface{}{
			"type":      "workspace.created",
			"workspace": workspace,
			"timestamp": time.Now(),
		})
	}

	return workspace, nil
}

func (s *WorkspaceService) GetWorkspace(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.WorkspaceResponse, error) {
	workspace, err := s.workspaceRepo.GetByID(ctx, id)
	if err != nil || workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	memberCount, _ := s.workspaceRepo.GetMemberCount(ctx, id)
	role, _ := s.memberRepo.GetRole(ctx, id, userID)

	return &models.WorkspaceResponse{
		Workspace:   workspace,
		MemberCount: memberCount,
		MyRole:      role,
	}, nil
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

	return s.workspaceRepo.Delete(ctx, id)
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

func (s *WorkspaceService) InviteMember(ctx context.Context, workspaceID uuid.UUID, inviterID uuid.UUID, req *models.InviteMemberRequest) (*models.WorkspaceInvite, error) {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, inviterID)
	if role != "owner" && role != "admin" {
		return nil, ErrNotAuthorized
	}

	// Check if already invited
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

	// Publish event for notification service
	if s.kafka != nil {
		s.kafka.Publish(ctx, "notification-events", invite.ID.String(), map[string]interface{}{
			"type":   "workspace.invite",
			"invite": invite,
		})
	}

	return invite, nil
}

func (s *WorkspaceService) AcceptInvite(ctx context.Context, token string, userID uuid.UUID) (*models.Workspace, error) {
	invite, err := s.inviteRepo.GetByToken(ctx, token)
	if err != nil || invite == nil {
		return nil, ErrInviteNotFound
	}

	// Check if already member
	isMember, _ := s.memberRepo.IsMember(ctx, invite.WorkspaceID, userID)
	if isMember {
		return nil, ErrAlreadyMember
	}

	// Add member
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

	return s.memberRepo.Remove(ctx, workspaceID, memberUserID)
}

func (s *WorkspaceService) UpdateMemberRole(ctx context.Context, workspaceID, memberUserID, requestorID uuid.UUID, newRole string) error {
	role, _ := s.memberRepo.GetRole(ctx, workspaceID, requestorID)
	if role != "owner" {
		return ErrNotAuthorized
	}

	return s.memberRepo.UpdateRole(ctx, workspaceID, memberUserID, newRole)
}

func (s *WorkspaceService) ListMembers(ctx context.Context, workspaceID uuid.UUID, page, perPage int) ([]*models.WorkspaceMember, int64, error) {
	return s.memberRepo.ListByWorkspace(ctx, workspaceID, page, perPage)
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
