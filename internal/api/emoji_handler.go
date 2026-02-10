package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/quckapp/workspace-service/internal/models"
	"github.com/quckapp/workspace-service/internal/service"
	"github.com/sirupsen/logrus"
)

type EmojiHandler struct {
	service *service.EmojiService
	logger  *logrus.Logger
}

func NewEmojiHandler(svc *service.EmojiService, logger *logrus.Logger) *EmojiHandler {
	return &EmojiHandler{service: svc, logger: logger}
}

func (h *EmojiHandler) CreateEmoji(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	var req models.CreateEmojiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	emoji, err := h.service.CreateEmoji(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		emojiHandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, emoji)
}

func (h *EmojiHandler) GetEmoji(c *gin.Context) {
	emojiID, _ := uuid.Parse(c.Param("emojiId"))
	emoji, err := h.service.GetEmoji(c.Request.Context(), emojiID)
	if err != nil {
		emojiHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, emoji)
}

func (h *EmojiHandler) ListEmojis(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	emojis, err := h.service.ListEmojis(c.Request.Context(), workspaceID, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list emojis"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"emojis": emojis})
}

func (h *EmojiHandler) SearchEmojis(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	var params models.EmojiSearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	emojis, err := h.service.SearchEmojis(c.Request.Context(), workspaceID, &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search emojis"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"emojis": emojis})
}

func (h *EmojiHandler) UpdateEmoji(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	emojiID, _ := uuid.Parse(c.Param("emojiId"))
	var req models.UpdateEmojiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	emoji, err := h.service.UpdateEmoji(c.Request.Context(), workspaceID, userID, emojiID, &req)
	if err != nil {
		emojiHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, emoji)
}

func (h *EmojiHandler) DeleteEmoji(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	emojiID, _ := uuid.Parse(c.Param("emojiId"))
	if err := h.service.DeleteEmoji(c.Request.Context(), workspaceID, userID, emojiID); err != nil {
		emojiHandleError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *EmojiHandler) BulkDeleteEmojis(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	var req models.BulkDeleteEmojiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.BulkDeleteEmojis(c.Request.Context(), workspaceID, userID, req.EmojiIDs); err != nil {
		emojiHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Emojis deleted successfully"})
}

func (h *EmojiHandler) IncrementUsage(c *gin.Context) {
	emojiID, _ := uuid.Parse(c.Param("emojiId"))
	if err := h.service.IncrementUsage(c.Request.Context(), emojiID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to increment usage"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Usage recorded"})
}

func (h *EmojiHandler) GetCategories(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	categories, err := h.service.GetCategories(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get categories"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

func (h *EmojiHandler) GetEmojiStats(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	stats, err := h.service.GetEmojiStats(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get emoji stats"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// Packs
func (h *EmojiHandler) CreatePack(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	var req models.EmojiPackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pack, err := h.service.CreatePack(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		emojiHandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, pack)
}

func (h *EmojiHandler) ListPacks(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	packs, err := h.service.ListPacks(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list packs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"packs": packs})
}

func (h *EmojiHandler) GetPackEmojis(c *gin.Context) {
	packID, _ := uuid.Parse(c.Param("packId"))
	emojis, err := h.service.GetPackEmojis(c.Request.Context(), packID)
	if err != nil {
		emojiHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"emojis": emojis})
}

func (h *EmojiHandler) DeletePack(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	packID, _ := uuid.Parse(c.Param("packId"))
	if err := h.service.DeletePack(c.Request.Context(), workspaceID, userID, packID); err != nil {
		emojiHandleError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func emojiHandleError(c *gin.Context, err error) {
	switch err {
	case service.ErrNotMember:
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
	case service.ErrNotAuthorized:
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
	case service.ErrEmojiNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Emoji not found"})
	case service.ErrEmojiNameExists:
		c.JSON(http.StatusConflict, gin.H{"error": "Emoji name already exists"})
	case service.ErrEmojiPackNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Emoji pack not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
