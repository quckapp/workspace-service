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

type WorkspaceHandler struct {
	service *service.WorkspaceService
	logger  *logrus.Logger
}

func NewWorkspaceHandler(svc *service.WorkspaceService, logger *logrus.Logger) *WorkspaceHandler {
	return &WorkspaceHandler{service: svc, logger: logger}
}

func (h *WorkspaceHandler) CreateWorkspace(c *gin.Context) {
	userID := getUserID(c)
	var req models.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspace, err := h.service.CreateWorkspace(c.Request.Context(), userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, workspace)
}

func (h *WorkspaceHandler) GetWorkspace(c *gin.Context) {
	userID := getUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	response, err := h.service.GetWorkspace(c.Request.Context(), id, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *WorkspaceHandler) UpdateWorkspace(c *gin.Context) {
	userID := getUserID(c)
	id, _ := uuid.Parse(c.Param("id"))

	var req models.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspace, err := h.service.UpdateWorkspace(c.Request.Context(), id, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, workspace)
}

func (h *WorkspaceHandler) DeleteWorkspace(c *gin.Context) {
	userID := getUserID(c)
	id, _ := uuid.Parse(c.Param("id"))

	if err := h.service.DeleteWorkspace(c.Request.Context(), id, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *WorkspaceHandler) ListWorkspaces(c *gin.Context) {
	userID := getUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	response, err := h.service.ListWorkspaces(c.Request.Context(), userID, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list workspaces"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *WorkspaceHandler) InviteMember(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invite, err := h.service.InviteMember(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, invite)
}

func (h *WorkspaceHandler) AcceptInvite(c *gin.Context) {
	userID := getUserID(c)
	token := c.Param("token")

	workspace, err := h.service.AcceptInvite(c.Request.Context(), token, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, workspace)
}

func (h *WorkspaceHandler) RemoveMember(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	memberUserID, _ := uuid.Parse(c.Param("userId"))

	if err := h.service.RemoveMember(c.Request.Context(), workspaceID, memberUserID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *WorkspaceHandler) UpdateMemberRole(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	memberUserID, _ := uuid.Parse(c.Param("userId"))

	var req models.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateMemberRole(c.Request.Context(), workspaceID, memberUserID, userID, req.Role); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role updated"})
}

func (h *WorkspaceHandler) ListMembers(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	members, total, err := h.service.ListMembers(c.Request.Context(), workspaceID, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list members"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": members, "total": total})
}

func getUserID(c *gin.Context) uuid.UUID {
	userIDStr, _ := c.Get("user_id")
	userID, _ := uuid.Parse(userIDStr.(string))
	return userID
}

func handleError(c *gin.Context, err error) {
	switch err {
	case service.ErrWorkspaceNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Workspace not found"})
	case service.ErrSlugExists:
		c.JSON(http.StatusConflict, gin.H{"error": "Slug already exists"})
	case service.ErrNotAuthorized:
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
	case service.ErrInviteNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Invite not found or expired"})
	case service.ErrAlreadyMember:
		c.JSON(http.StatusConflict, gin.H{"error": "Already a member"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
