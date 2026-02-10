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

type DiscoveryHandler struct {
	service *service.DiscoveryService
	logger  *logrus.Logger
}

func NewDiscoveryHandler(svc *service.DiscoveryService, logger *logrus.Logger) *DiscoveryHandler {
	return &DiscoveryHandler{service: svc, logger: logger}
}

func (h *DiscoveryHandler) GetDirectoryEntry(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	entry, err := h.service.GetDirectoryEntry(c.Request.Context(), workspaceID)
	if err != nil {
		discoveryHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, entry)
}

func (h *DiscoveryHandler) UpdateDirectoryEntry(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	var req models.UpdateDirectoryEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry, err := h.service.UpdateDirectoryEntry(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		discoveryHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, entry)
}

func (h *DiscoveryHandler) SearchDirectory(c *gin.Context) {
	var params models.DirectorySearchParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entries, err := h.service.SearchDirectory(c.Request.Context(), &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search directory"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"workspaces": entries})
}

func (h *DiscoveryHandler) GetCategories(c *gin.Context) {
	categories, err := h.service.GetCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get categories"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

func (h *DiscoveryHandler) GetRecommendations(c *gin.Context) {
	userID := getUserID(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	recs, err := h.service.GetRecommendations(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recommendations"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"recommendations": recs})
}

func (h *DiscoveryHandler) DismissRecommendation(c *gin.Context) {
	userID := getUserID(c)
	var req models.DismissRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.DismissRecommendation(c.Request.Context(), userID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to dismiss recommendation"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Recommendation dismissed"})
}

func (h *DiscoveryHandler) GetTrending(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	trending, err := h.service.GetTrending(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trending"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"trending": trending})
}

func discoveryHandleError(c *gin.Context, err error) {
	switch err {
	case service.ErrNotMember:
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
	case service.ErrNotAuthorized:
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
	case service.ErrDirectoryEntryNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Directory entry not found"})
	case service.ErrWorkspaceNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Workspace not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
