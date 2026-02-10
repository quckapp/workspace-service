package service

import (
	"context"
	"errors"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/quckapp/workspace-service/internal/models"
	"github.com/quckapp/workspace-service/internal/repository"
	"github.com/sirupsen/logrus"
)

var (
	ErrDirectoryEntryNotFound = errors.New("directory entry not found")
)

type DiscoveryService struct {
	discoveryRepo *repository.DiscoveryRepository
	workspaceRepo *repository.WorkspaceRepository
	memberRepo    *repository.MemberRepository
	logger        *logrus.Logger
}

func NewDiscoveryService(discoveryRepo *repository.DiscoveryRepository, workspaceRepo *repository.WorkspaceRepository, memberRepo *repository.MemberRepository, logger *logrus.Logger) *DiscoveryService {
	return &DiscoveryService{discoveryRepo: discoveryRepo, workspaceRepo: workspaceRepo, memberRepo: memberRepo, logger: logger}
}

func (s *DiscoveryService) GetDirectoryEntry(ctx context.Context, workspaceID uuid.UUID) (*models.WorkspaceDirectoryEntry, error) {
	entry, err := s.discoveryRepo.GetDirectoryEntry(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, ErrDirectoryEntryNotFound
	}
	return entry, nil
}

func (s *DiscoveryService) UpdateDirectoryEntry(ctx context.Context, workspaceID, userID uuid.UUID, req *models.UpdateDirectoryEntryRequest) (*models.WorkspaceDirectoryEntry, error) {
	member, err := s.memberRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
	if err != nil || member == nil {
		return nil, ErrNotMember
	}
	if member.Role != "owner" && member.Role != "admin" {
		return nil, ErrNotAuthorized
	}

	workspace, err := s.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil || workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	now := time.Now()
	entry, _ := s.discoveryRepo.GetDirectoryEntry(ctx, workspaceID)
	if entry == nil {
		entry = &models.WorkspaceDirectoryEntry{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			IconURL:     workspace.IconURL,
			CreatedAt:   now,
		}
	}

	if req.IsListed != nil {
		entry.IsListed = *req.IsListed
	}
	if req.Category != nil {
		entry.Category = req.Category
	}
	if req.Tags != nil {
		tagsJSON, _ := json.Marshal(req.Tags)
		tags := models.JSON{}
		json.Unmarshal(tagsJSON, &tags)
		entry.Tags = tags
	}
	if req.Description != nil {
		entry.Description = req.Description
	}
	if req.BannerURL != nil {
		entry.BannerURL = req.BannerURL
	}
	if req.Website != nil {
		entry.Website = req.Website
	}
	entry.UpdatedAt = now

	if err := s.discoveryRepo.UpsertDirectoryEntry(ctx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *DiscoveryService) SearchDirectory(ctx context.Context, params *models.DirectorySearchParams) ([]*models.WorkspaceDirectoryEntry, error) {
	if params.PerPage > 50 {
		params.PerPage = 50
	}
	offset := (params.Page - 1) * params.PerPage
	return s.discoveryRepo.SearchDirectory(ctx, params.Query, params.Category, params.SortBy, params.PerPage, offset)
}

func (s *DiscoveryService) GetCategories(ctx context.Context) ([]models.WorkspaceCategory, error) {
	return s.discoveryRepo.GetCategories(ctx)
}

func (s *DiscoveryService) GetRecommendations(ctx context.Context, userID uuid.UUID, limit int) ([]*models.WorkspaceRecommendation, error) {
	if limit > 20 {
		limit = 20
	}
	return s.discoveryRepo.GetRecommendations(ctx, userID, limit)
}

func (s *DiscoveryService) DismissRecommendation(ctx context.Context, userID uuid.UUID, req *models.DismissRecommendationRequest) error {
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return err
	}
	return s.discoveryRepo.DismissRecommendation(ctx, userID, workspaceID)
}

func (s *DiscoveryService) GetTrending(ctx context.Context, limit int) ([]*models.TrendingWorkspace, error) {
	if limit > 20 {
		limit = 20
	}
	return s.discoveryRepo.GetTrending(ctx, limit)
}
