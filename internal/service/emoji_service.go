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
	ErrEmojiNotFound    = errors.New("custom emoji not found")
	ErrEmojiNameExists  = errors.New("emoji name already exists in this workspace")
	ErrEmojiPackNotFound = errors.New("emoji pack not found")
)

type EmojiService struct {
	emojiRepo *repository.EmojiRepository
	memberRepo *repository.MemberRepository
	logger    *logrus.Logger
}

func NewEmojiService(emojiRepo *repository.EmojiRepository, memberRepo *repository.MemberRepository, logger *logrus.Logger) *EmojiService {
	return &EmojiService{emojiRepo: emojiRepo, memberRepo: memberRepo, logger: logger}
}

func (s *EmojiService) CreateEmoji(ctx context.Context, workspaceID, userID uuid.UUID, req *models.CreateEmojiRequest) (*models.CustomEmoji, error) {
	member, err := s.memberRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
	if err != nil || member == nil {
		return nil, ErrNotMember
	}

	existing, _ := s.emojiRepo.GetByName(ctx, workspaceID, req.Name)
	if existing != nil {
		return nil, ErrEmojiNameExists
	}

	now := time.Now()
	emoji := &models.CustomEmoji{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        req.Name,
		ImageURL:    req.ImageURL,
		Category:    req.Category,
		AliasFor:    req.AliasFor,
		CreatedBy:   userID,
		IsAnimated:  req.IsAnimated,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.emojiRepo.Create(ctx, emoji); err != nil {
		return nil, err
	}
	return emoji, nil
}

func (s *EmojiService) GetEmoji(ctx context.Context, id uuid.UUID) (*models.CustomEmoji, error) {
	emoji, err := s.emojiRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if emoji == nil {
		return nil, ErrEmojiNotFound
	}
	return emoji, nil
}

func (s *EmojiService) ListEmojis(ctx context.Context, workspaceID uuid.UUID, page, perPage int) ([]*models.CustomEmoji, error) {
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage
	return s.emojiRepo.ListByWorkspace(ctx, workspaceID, perPage, offset)
}

func (s *EmojiService) SearchEmojis(ctx context.Context, workspaceID uuid.UUID, params *models.EmojiSearchParams) ([]*models.CustomEmoji, error) {
	if params.PerPage > 100 {
		params.PerPage = 100
	}
	offset := (params.Page - 1) * params.PerPage
	return s.emojiRepo.Search(ctx, workspaceID, params.Query, params.Category, params.PerPage, offset)
}

func (s *EmojiService) UpdateEmoji(ctx context.Context, workspaceID, userID, emojiID uuid.UUID, req *models.UpdateEmojiRequest) (*models.CustomEmoji, error) {
	member, err := s.memberRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
	if err != nil || member == nil {
		return nil, ErrNotMember
	}

	emoji, err := s.emojiRepo.GetByID(ctx, emojiID)
	if err != nil || emoji == nil {
		return nil, ErrEmojiNotFound
	}

	if emoji.CreatedBy != userID && member.Role != "owner" && member.Role != "admin" {
		return nil, ErrNotAuthorized
	}

	if req.Name != nil {
		existing, _ := s.emojiRepo.GetByName(ctx, workspaceID, *req.Name)
		if existing != nil && existing.ID != emojiID {
			return nil, ErrEmojiNameExists
		}
		emoji.Name = *req.Name
	}
	if req.Category != nil {
		emoji.Category = req.Category
	}
	if req.AliasFor != nil {
		emoji.AliasFor = req.AliasFor
	}

	if err := s.emojiRepo.Update(ctx, emoji); err != nil {
		return nil, err
	}
	return emoji, nil
}

func (s *EmojiService) DeleteEmoji(ctx context.Context, workspaceID, userID, emojiID uuid.UUID) error {
	member, err := s.memberRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
	if err != nil || member == nil {
		return ErrNotMember
	}

	emoji, err := s.emojiRepo.GetByID(ctx, emojiID)
	if err != nil || emoji == nil {
		return ErrEmojiNotFound
	}

	if emoji.CreatedBy != userID && member.Role != "owner" && member.Role != "admin" {
		return ErrNotAuthorized
	}

	return s.emojiRepo.Delete(ctx, emojiID)
}

func (s *EmojiService) BulkDeleteEmojis(ctx context.Context, workspaceID, userID uuid.UUID, emojiIDs []string) error {
	member, err := s.memberRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
	if err != nil || member == nil {
		return ErrNotMember
	}
	if member.Role != "owner" && member.Role != "admin" {
		return ErrNotAuthorized
	}

	ids := make([]uuid.UUID, 0, len(emojiIDs))
	for _, idStr := range emojiIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return s.emojiRepo.BulkDelete(ctx, ids)
}

func (s *EmojiService) IncrementUsage(ctx context.Context, emojiID uuid.UUID) error {
	return s.emojiRepo.IncrementUsage(ctx, emojiID)
}

func (s *EmojiService) GetCategories(ctx context.Context, workspaceID uuid.UUID) ([]models.EmojiCategory, error) {
	return s.emojiRepo.GetCategories(ctx, workspaceID)
}

func (s *EmojiService) GetEmojiStats(ctx context.Context, workspaceID uuid.UUID) (*models.EmojiStats, error) {
	total, err := s.emojiRepo.CountByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	animated, err := s.emojiRepo.CountAnimated(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	packCount, err := s.emojiRepo.CountPacks(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	topEmojis, err := s.emojiRepo.GetTopEmojis(ctx, workspaceID, 10)
	if err != nil {
		return nil, err
	}
	categories, err := s.emojiRepo.GetCategories(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	return &models.EmojiStats{
		TotalEmojis:   total,
		AnimatedCount: animated,
		TotalPacks:    packCount,
		TopEmojis:     topEmojis,
		Categories:    categories,
	}, nil
}

// Pack operations
func (s *EmojiService) CreatePack(ctx context.Context, workspaceID, userID uuid.UUID, req *models.EmojiPackRequest) (*models.EmojiPack, error) {
	member, err := s.memberRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
	if err != nil || member == nil {
		return nil, ErrNotMember
	}

	now := time.Now()
	pack := &models.EmojiPack{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
		EmojiCount:  len(req.EmojiIDs),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.emojiRepo.CreatePack(ctx, pack); err != nil {
		return nil, err
	}

	for i, idStr := range req.EmojiIDs {
		emojiID, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		mapping := &models.EmojiPackMapping{
			ID:       uuid.New(),
			PackID:   pack.ID,
			EmojiID:  emojiID,
			Position: i,
		}
		if err := s.emojiRepo.AddEmojiToPack(ctx, mapping); err != nil {
			s.logger.WithError(err).Warn("Failed to add emoji to pack")
		}
	}

	return pack, nil
}

func (s *EmojiService) ListPacks(ctx context.Context, workspaceID uuid.UUID) ([]*models.EmojiPack, error) {
	return s.emojiRepo.ListPacks(ctx, workspaceID)
}

func (s *EmojiService) GetPackEmojis(ctx context.Context, packID uuid.UUID) ([]*models.CustomEmoji, error) {
	pack, err := s.emojiRepo.GetPackByID(ctx, packID)
	if err != nil || pack == nil {
		return nil, ErrEmojiPackNotFound
	}
	return s.emojiRepo.ListPackEmojis(ctx, packID)
}

func (s *EmojiService) DeletePack(ctx context.Context, workspaceID, userID, packID uuid.UUID) error {
	member, err := s.memberRepo.GetByWorkspaceAndUser(ctx, workspaceID, userID)
	if err != nil || member == nil {
		return ErrNotMember
	}
	if member.Role != "owner" && member.Role != "admin" {
		return ErrNotAuthorized
	}
	return s.emojiRepo.DeletePack(ctx, packID)
}
