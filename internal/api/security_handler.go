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

type SecurityHandler struct {
	service *service.SecurityService
	logger  *logrus.Logger
}

func NewSecurityHandler(svc *service.SecurityService, logger *logrus.Logger) *SecurityHandler {
	return &SecurityHandler{service: svc, logger: logger}
}

// IP Allowlist
func (h *SecurityHandler) AddIPEntry(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	var req models.AddIPAllowlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry, err := h.service.AddIPEntry(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		securityHandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, entry)
}

func (h *SecurityHandler) ListIPEntries(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	entries, err := h.service.ListIPEntries(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list IP entries"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

func (h *SecurityHandler) UpdateIPEntry(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	entryID, _ := uuid.Parse(c.Param("entryId"))
	var req models.UpdateIPAllowlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry, err := h.service.UpdateIPEntry(c.Request.Context(), workspaceID, userID, entryID, &req)
	if err != nil {
		securityHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, entry)
}

func (h *SecurityHandler) DeleteIPEntry(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	entryID, _ := uuid.Parse(c.Param("entryId"))
	if err := h.service.DeleteIPEntry(c.Request.Context(), workspaceID, userID, entryID); err != nil {
		securityHandleError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// Sessions
func (h *SecurityHandler) ListMySessions(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	sessions, err := h.service.ListUserSessions(c.Request.Context(), workspaceID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list sessions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func (h *SecurityHandler) ListAllSessions(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	sessions, err := h.service.ListAllSessions(c.Request.Context(), workspaceID, userID, page, perPage)
	if err != nil {
		securityHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func (h *SecurityHandler) RevokeSession(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	sessionID, _ := uuid.Parse(c.Param("sessionId"))
	if err := h.service.RevokeSession(c.Request.Context(), workspaceID, userID, sessionID); err != nil {
		securityHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Session revoked"})
}

func (h *SecurityHandler) RevokeSessions(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	var req models.RevokeSessionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.RevokeSessions(c.Request.Context(), workspaceID, userID, &req); err != nil {
		securityHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Sessions revoked"})
}

// Security Policy
func (h *SecurityHandler) GetSecurityPolicy(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	policy, err := h.service.GetSecurityPolicy(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get security policy"})
		return
	}
	c.JSON(http.StatusOK, policy)
}

func (h *SecurityHandler) UpdateSecurityPolicy(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	var req models.UpdateSecurityPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	policy, err := h.service.UpdateSecurityPolicy(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		securityHandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, policy)
}

// Audit
func (h *SecurityHandler) ListSecurityAudit(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	severity := c.Query("severity")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	entries, err := h.service.ListSecurityAudit(c.Request.Context(), workspaceID, severity, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list security audit"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

// Overview
func (h *SecurityHandler) GetSecurityOverview(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	overview, err := h.service.GetSecurityOverview(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get security overview"})
		return
	}
	c.JSON(http.StatusOK, overview)
}

func securityHandleError(c *gin.Context, err error) {
	switch err {
	case service.ErrNotMember:
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
	case service.ErrNotAuthorized:
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
	case service.ErrIPEntryNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "IP allowlist entry not found"})
	case service.ErrSessionNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
