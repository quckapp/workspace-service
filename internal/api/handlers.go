package api

import (
	"net/http"
	"strconv"
	"time"

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

// ── Workspace CRUD ──

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

// ── Workspace Stats ──

func (h *WorkspaceHandler) GetWorkspaceStats(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	stats, err := h.service.GetWorkspaceStats(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ── Workspace Settings ──

func (h *WorkspaceHandler) GetWorkspaceSettings(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	settings, err := h.service.GetWorkspaceSettings(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *WorkspaceHandler) UpdateWorkspaceSettings(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var settings models.JSON
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.UpdateWorkspaceSettings(c.Request.Context(), workspaceID, userID, settings)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ── Leave Workspace ──

func (h *WorkspaceHandler) LeaveWorkspace(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	if err := h.service.LeaveWorkspace(c.Request.Context(), workspaceID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Left workspace"})
}

// ── Ownership Transfer ──

func (h *WorkspaceHandler) TransferOwnership(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.TransferOwnershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newOwnerID, err := uuid.Parse(req.NewOwnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid new owner ID"})
		return
	}

	if err := h.service.TransferOwnership(c.Request.Context(), workspaceID, userID, newOwnerID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ownership transferred"})
}

// ── Member Management ──

func (h *WorkspaceHandler) GetMember(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	memberUserID, _ := uuid.Parse(c.Param("userId"))

	member, err := h.service.GetMember(c.Request.Context(), workspaceID, memberUserID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, member)
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

func (h *WorkspaceHandler) BulkInvite(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.BulkInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.BulkInvite(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
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

// ── Invite Management ──

func (h *WorkspaceHandler) ListInvites(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	invites, err := h.service.ListInvites(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"invites": invites})
}

func (h *WorkspaceHandler) RevokeInvite(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	inviteID, _ := uuid.Parse(c.Param("inviteId"))

	if err := h.service.RevokeInvite(c.Request.Context(), workspaceID, inviteID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invite revoked"})
}

// ── Invite Codes ──

func (h *WorkspaceHandler) CreateInviteCode(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.CreateInviteCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	code, err := h.service.CreateInviteCode(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, code)
}

func (h *WorkspaceHandler) JoinByCode(c *gin.Context) {
	userID := getUserID(c)

	var req models.JoinByCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspace, err := h.service.JoinByCode(c.Request.Context(), req.InviteCode, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, workspace)
}

func (h *WorkspaceHandler) ListInviteCodes(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	codes, err := h.service.ListInviteCodes(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"invite_codes": codes})
}

func (h *WorkspaceHandler) RevokeInviteCode(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	codeID, _ := uuid.Parse(c.Param("codeId"))

	if err := h.service.RevokeInviteCode(c.Request.Context(), workspaceID, codeID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invite code revoked"})
}

// ── Activity Log ──

func (h *WorkspaceHandler) GetActivityLog(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	result, err := h.service.GetActivityLog(c.Request.Context(), workspaceID, userID, page, perPage)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *WorkspaceHandler) GetActivityLogByActor(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	actorID, _ := uuid.Parse(c.Param("actorId"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	result, err := h.service.GetActivityLogByActor(c.Request.Context(), workspaceID, actorID, userID, page, perPage)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ── Member Profiles ──

func (h *WorkspaceHandler) GetMemberProfile(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))
	memberUserID, _ := uuid.Parse(c.Param("userId"))

	profile, err := h.service.GetMemberProfile(c.Request.Context(), workspaceID, memberUserID)
	if err != nil {
		handleError(c, err)
		return
	}
	if profile == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *WorkspaceHandler) UpdateMemberProfile(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.UpdateMemberProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.service.UpdateMemberProfile(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *WorkspaceHandler) SetOnlineStatus(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req struct {
		IsOnline bool `json:"is_online"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SetOnlineStatus(c.Request.Context(), workspaceID, userID, req.IsOnline); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated"})
}

// ── Custom Roles ──

func (h *WorkspaceHandler) CreateRole(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := h.service.CreateRole(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, role)
}

func (h *WorkspaceHandler) ListRoles(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))

	roles, err := h.service.ListRoles(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

func (h *WorkspaceHandler) UpdateRole(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	roleID, _ := uuid.Parse(c.Param("roleId"))

	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := h.service.UpdateRole(c.Request.Context(), workspaceID, roleID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, role)
}

func (h *WorkspaceHandler) DeleteRole(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	roleID, _ := uuid.Parse(c.Param("roleId"))

	if err := h.service.DeleteRole(c.Request.Context(), workspaceID, roleID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted"})
}

// ── Workspace Search ──

func (h *WorkspaceHandler) SearchWorkspaces(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	workspaces, total, err := h.service.SearchWorkspaces(c.Request.Context(), query, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workspaces": workspaces, "total": total, "page": page, "per_page": perPage})
}

// ── Workspace Analytics ──

func (h *WorkspaceHandler) GetAnalytics(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	analytics, err := h.service.GetAnalytics(c.Request.Context(), workspaceID, userID, days)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// ── Workspace Templates ──

func (h *WorkspaceHandler) CreateTemplateFromWorkspace(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.CreateTemplateFromWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.service.CreateTemplateFromWorkspace(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, template)
}

func (h *WorkspaceHandler) CreateWorkspaceFromTemplate(c *gin.Context) {
	userID := getUserID(c)

	var req models.CreateWorkspaceFromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspace, err := h.service.CreateWorkspaceFromTemplate(c.Request.Context(), userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, workspace)
}

func (h *WorkspaceHandler) ListTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	templates, total, err := h.service.ListTemplates(c.Request.Context(), page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates, "total": total, "page": page, "per_page": perPage})
}

func (h *WorkspaceHandler) GetTemplate(c *gin.Context) {
	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := h.service.GetTemplate(c.Request.Context(), templateID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, template)
}

func (h *WorkspaceHandler) UpdateTemplate(c *gin.Context) {
	userID := getUserID(c)
	templateID, _ := uuid.Parse(c.Param("templateId"))

	var req models.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.service.UpdateTemplate(c.Request.Context(), templateID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, template)
}

func (h *WorkspaceHandler) DeleteTemplate(c *gin.Context) {
	userID := getUserID(c)
	templateID, _ := uuid.Parse(c.Param("templateId"))

	if err := h.service.DeleteTemplate(c.Request.Context(), templateID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted"})
}

// ── Member Preferences ──

func (h *WorkspaceHandler) GetPreferences(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	pref, err := h.service.GetPreferences(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pref)
}

func (h *WorkspaceHandler) UpdatePreferences(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pref, err := h.service.UpdatePreferences(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pref)
}

func (h *WorkspaceHandler) ResetPreferences(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	if err := h.service.ResetPreferences(c.Request.Context(), workspaceID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Preferences reset to defaults"})
}

// ── Workspace Tags ──

func (h *WorkspaceHandler) CreateTag(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.service.CreateTag(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, tag)
}

func (h *WorkspaceHandler) ListTags(c *gin.Context) {
	workspaceID, _ := uuid.Parse(c.Param("id"))

	tags, err := h.service.ListTags(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

func (h *WorkspaceHandler) UpdateTag(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	tagID, _ := uuid.Parse(c.Param("tagId"))

	var req models.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.service.UpdateTag(c.Request.Context(), workspaceID, tagID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tag)
}

func (h *WorkspaceHandler) DeleteTag(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	tagID, _ := uuid.Parse(c.Param("tagId"))

	if err := h.service.DeleteTag(c.Request.Context(), workspaceID, tagID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tag deleted"})
}

// ── Workspace Moderation ──

func (h *WorkspaceHandler) BanMember(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	targetUserID, _ := uuid.Parse(c.Param("userId"))

	var req models.BanMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ban, err := h.service.BanMember(c.Request.Context(), workspaceID, targetUserID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, ban)
}

func (h *WorkspaceHandler) UnbanMember(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	targetUserID, _ := uuid.Parse(c.Param("userId"))

	if err := h.service.UnbanMember(c.Request.Context(), workspaceID, targetUserID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User unbanned"})
}

func (h *WorkspaceHandler) MuteMember(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	targetUserID, _ := uuid.Parse(c.Param("userId"))

	var req models.MuteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mute, err := h.service.MuteMember(c.Request.Context(), workspaceID, targetUserID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, mute)
}

func (h *WorkspaceHandler) UnmuteMember(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	targetUserID, _ := uuid.Parse(c.Param("userId"))

	if err := h.service.UnmuteMember(c.Request.Context(), workspaceID, targetUserID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User unmuted"})
}

func (h *WorkspaceHandler) GetModerationHistory(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	history, err := h.service.GetModerationHistory(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, history)
}

// ── Workspace Announcements ──

func (h *WorkspaceHandler) CreateAnnouncement(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	announcement, err := h.service.CreateAnnouncement(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, announcement)
}

func (h *WorkspaceHandler) ListAnnouncements(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	announcements, total, err := h.service.ListAnnouncements(c.Request.Context(), workspaceID, userID, page, perPage)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"announcements": announcements, "total": total, "page": page, "per_page": perPage})
}

func (h *WorkspaceHandler) UpdateAnnouncement(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	announcementID, _ := uuid.Parse(c.Param("announcementId"))

	var req models.UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	announcement, err := h.service.UpdateAnnouncement(c.Request.Context(), workspaceID, announcementID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, announcement)
}

func (h *WorkspaceHandler) PinAnnouncement(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	announcementID, _ := uuid.Parse(c.Param("announcementId"))

	var req models.PinAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.PinAnnouncement(c.Request.Context(), workspaceID, announcementID, userID, &req); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pin status updated"})
}

func (h *WorkspaceHandler) DeleteAnnouncement(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	announcementID, _ := uuid.Parse(c.Param("announcementId"))

	if err := h.service.DeleteAnnouncement(c.Request.Context(), workspaceID, announcementID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Announcement deleted"})
}

// ── Workspace Webhooks ──

func (h *WorkspaceHandler) CreateWebhook(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	webhook, err := h.service.CreateWebhook(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, webhook)
}

func (h *WorkspaceHandler) ListWebhooks(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	webhooks, err := h.service.ListWebhooks(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"webhooks": webhooks})
}

func (h *WorkspaceHandler) UpdateWebhook(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	webhookID, _ := uuid.Parse(c.Param("webhookId"))

	var req models.UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	webhook, err := h.service.UpdateWebhook(c.Request.Context(), workspaceID, webhookID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, webhook)
}

func (h *WorkspaceHandler) DeleteWebhook(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	webhookID, _ := uuid.Parse(c.Param("webhookId"))

	if err := h.service.DeleteWebhook(c.Request.Context(), workspaceID, webhookID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook deleted"})
}

func (h *WorkspaceHandler) TestWebhook(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	webhookID, _ := uuid.Parse(c.Param("webhookId"))

	if err := h.service.TestWebhook(c.Request.Context(), workspaceID, webhookID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook test successful"})
}

// ── Workspace Favorites ──

func (h *WorkspaceHandler) FavoriteWorkspace(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	fav, err := h.service.FavoriteWorkspace(c.Request.Context(), userID, workspaceID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, fav)
}

func (h *WorkspaceHandler) UnfavoriteWorkspace(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	if err := h.service.UnfavoriteWorkspace(c.Request.Context(), userID, workspaceID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workspace unfavorited"})
}

func (h *WorkspaceHandler) ListFavorites(c *gin.Context) {
	userID := getUserID(c)

	favs, err := h.service.ListFavorites(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list favorites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"favorites": favs})
}

func (h *WorkspaceHandler) ReorderFavorites(c *gin.Context) {
	userID := getUserID(c)

	var req models.ReorderFavoritesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ReorderFavorites(c.Request.Context(), userID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reorder favorites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Favorites reordered"})
}

// ── Audit Export ──

func (h *WorkspaceHandler) ExportAuditLog(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.AuditExportRequest
	req.Format = c.DefaultQuery("format", "json")
	req.ActionType = c.Query("action_type")

	if startStr := c.Query("start_date"); startStr != "" {
		t, err := time.Parse("2006-01-02", startStr)
		if err == nil {
			req.StartDate = &t
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		t, err := time.Parse("2006-01-02", endStr)
		if err == nil {
			req.EndDate = &t
		}
	}

	result, err := h.service.ExportAuditLog(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ── Member Notes ──

func (h *WorkspaceHandler) CreateMemberNote(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	targetID, _ := uuid.Parse(c.Param("userId"))

	var req models.CreateMemberNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := h.service.CreateMemberNote(c.Request.Context(), workspaceID, targetID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, note)
}

func (h *WorkspaceHandler) ListMemberNotes(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	targetID, _ := uuid.Parse(c.Param("userId"))

	notes, err := h.service.ListMemberNotes(c.Request.Context(), workspaceID, targetID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"notes": notes})
}

func (h *WorkspaceHandler) UpdateMemberNote(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	noteID, _ := uuid.Parse(c.Param("noteId"))

	var req models.UpdateMemberNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := h.service.UpdateMemberNote(c.Request.Context(), workspaceID, noteID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, note)
}

func (h *WorkspaceHandler) DeleteMemberNote(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	noteID, _ := uuid.Parse(c.Param("noteId"))

	if err := h.service.DeleteMemberNote(c.Request.Context(), workspaceID, noteID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted"})
}

// ── Scheduled Actions ──

func (h *WorkspaceHandler) CreateScheduledAction(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.CreateScheduledActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	action, err := h.service.CreateScheduledAction(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, action)
}

func (h *WorkspaceHandler) ListScheduledActions(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	actions, err := h.service.ListScheduledActions(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"scheduled_actions": actions})
}

func (h *WorkspaceHandler) UpdateScheduledAction(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	actionID, _ := uuid.Parse(c.Param("actionId"))

	var req models.UpdateScheduledActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	action, err := h.service.UpdateScheduledAction(c.Request.Context(), workspaceID, actionID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, action)
}

func (h *WorkspaceHandler) CancelScheduledAction(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	actionID, _ := uuid.Parse(c.Param("actionId"))

	if err := h.service.CancelScheduledAction(c.Request.Context(), workspaceID, actionID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheduled action cancelled"})
}

func (h *WorkspaceHandler) DeleteScheduledAction(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	actionID, _ := uuid.Parse(c.Param("actionId"))

	if err := h.service.DeleteScheduledAction(c.Request.Context(), workspaceID, actionID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheduled action deleted"})
}

// ── Usage Quotas ──

func (h *WorkspaceHandler) GetQuotaUsage(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	usage, err := h.service.GetQuotaUsage(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, usage)
}

func (h *WorkspaceHandler) UpdateQuota(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.UpdateQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	quota, err := h.service.UpdateQuota(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, quota)
}

// ── Workspace Archive / Restore ──

func (h *WorkspaceHandler) ArchiveWorkspace(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.ArchiveWorkspaceRequest
	c.ShouldBindJSON(&req) // optional body

	if err := h.service.ArchiveWorkspace(c.Request.Context(), workspaceID, userID, &req); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workspace archived"})
}

func (h *WorkspaceHandler) RestoreWorkspace(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	if err := h.service.RestoreWorkspace(c.Request.Context(), workspaceID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workspace restored"})
}

func (h *WorkspaceHandler) ListArchivedWorkspaces(c *gin.Context) {
	userID := getUserID(c)

	workspaces, err := h.service.ListArchivedWorkspaces(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, workspaces)
}

// ── Workspace Cloning ──

func (h *WorkspaceHandler) CloneWorkspace(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.CloneWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspace, err := h.service.CloneWorkspace(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, workspace)
}

// ── Pinned Items ──

func (h *WorkspaceHandler) CreatePinnedItem(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.CreatePinnedItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.CreatePinnedItem(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *WorkspaceHandler) ListPinnedItems(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	items, err := h.service.ListPinnedItems(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *WorkspaceHandler) UpdatePinnedItem(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	pinID, _ := uuid.Parse(c.Param("pinId"))

	var req models.UpdatePinnedItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.UpdatePinnedItem(c.Request.Context(), workspaceID, pinID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *WorkspaceHandler) DeletePinnedItem(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	pinID, _ := uuid.Parse(c.Param("pinId"))

	if err := h.service.DeletePinnedItem(c.Request.Context(), workspaceID, pinID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pinned item removed"})
}

func (h *WorkspaceHandler) ReorderPins(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.ReorderPinsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ReorderPins(c.Request.Context(), workspaceID, userID, &req); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pins reordered"})
}

// ── Member Groups / Teams ──

func (h *WorkspaceHandler) CreateGroup(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.service.CreateGroup(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, group)
}

func (h *WorkspaceHandler) ListGroups(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	groups, err := h.service.ListGroups(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, groups)
}

func (h *WorkspaceHandler) GetGroup(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	groupID, _ := uuid.Parse(c.Param("groupId"))

	group, err := h.service.GetGroup(c.Request.Context(), workspaceID, groupID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, group)
}

func (h *WorkspaceHandler) UpdateGroup(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	groupID, _ := uuid.Parse(c.Param("groupId"))

	var req models.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, err := h.service.UpdateGroup(c.Request.Context(), workspaceID, groupID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, group)
}

func (h *WorkspaceHandler) DeleteGroup(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	groupID, _ := uuid.Parse(c.Param("groupId"))

	if err := h.service.DeleteGroup(c.Request.Context(), workspaceID, groupID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted"})
}

func (h *WorkspaceHandler) AddGroupMembers(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	groupID, _ := uuid.Parse(c.Param("groupId"))

	var req models.AddGroupMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	added, err := h.service.AddGroupMembers(c.Request.Context(), workspaceID, groupID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"added": added})
}

func (h *WorkspaceHandler) RemoveGroupMember(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	groupID, _ := uuid.Parse(c.Param("groupId"))
	targetID, _ := uuid.Parse(c.Param("userId"))

	if err := h.service.RemoveGroupMember(c.Request.Context(), workspaceID, groupID, targetID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed from group"})
}

func (h *WorkspaceHandler) ListUserGroups(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	targetID, _ := uuid.Parse(c.Param("userId"))

	groups, err := h.service.ListUserGroups(c.Request.Context(), workspaceID, targetID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, groups)
}

// ── Custom Fields ──

func (h *WorkspaceHandler) CreateCustomField(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	var req models.CreateCustomFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	field, err := h.service.CreateCustomField(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, field)
}

func (h *WorkspaceHandler) ListCustomFields(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	fields, err := h.service.ListCustomFields(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, fields)
}

func (h *WorkspaceHandler) UpdateCustomField(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	fieldID, _ := uuid.Parse(c.Param("fieldId"))

	var req models.UpdateCustomFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	field, err := h.service.UpdateCustomField(c.Request.Context(), workspaceID, fieldID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, field)
}

func (h *WorkspaceHandler) DeleteCustomField(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	fieldID, _ := uuid.Parse(c.Param("fieldId"))

	if err := h.service.DeleteCustomField(c.Request.Context(), workspaceID, fieldID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Custom field deleted"})
}

func (h *WorkspaceHandler) SetCustomFieldValue(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))
	fieldID, _ := uuid.Parse(c.Param("fieldId"))

	var req models.SetCustomFieldValueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use workspace ID as the entity (workspace-level custom field value)
	value, err := h.service.SetCustomFieldValue(c.Request.Context(), workspaceID, fieldID, workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, value)
}

func (h *WorkspaceHandler) GetCustomFieldValues(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, _ := uuid.Parse(c.Param("id"))

	// Get workspace-level custom field values
	results, err := h.service.GetCustomFieldValues(c.Request.Context(), workspaceID, workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, results)
}

// ── Reactions ──

func (h *WorkspaceHandler) AddReaction(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	var req models.AddReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddReaction(c.Request.Context(), workspaceID, userID, &req); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Reaction added"})
}

func (h *WorkspaceHandler) RemoveReaction(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	entityType := c.Query("entity_type")
	entityID, err := uuid.Parse(c.Query("entity_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}
	emoji := c.Query("emoji")

	if err := h.service.RemoveReaction(c.Request.Context(), workspaceID, userID, entityType, entityID, emoji); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reaction removed"})
}

func (h *WorkspaceHandler) ListReactions(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	entityType := c.Query("entity_type")
	entityID, err := uuid.Parse(c.Query("entity_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	reactions, err := h.service.ListReactions(c.Request.Context(), workspaceID, userID, entityType, entityID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, reactions)
}

func (h *WorkspaceHandler) GetReactionSummary(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	entityType := c.Query("entity_type")
	entityID, err := uuid.Parse(c.Query("entity_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	summaries, err := h.service.GetReactionSummary(c.Request.Context(), workspaceID, userID, entityType, entityID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, summaries)
}

// ── Bookmarks ──

func (h *WorkspaceHandler) CreateBookmark(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	var req models.CreateBookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookmark, err := h.service.CreateBookmark(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, bookmark)
}

func (h *WorkspaceHandler) ListBookmarks(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	folder := c.Query("folder")
	if folder != "" {
		bookmarks, err := h.service.ListBookmarksByFolder(c.Request.Context(), workspaceID, userID, folder)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, bookmarks)
		return
	}

	bookmarks, err := h.service.ListBookmarks(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, bookmarks)
}

func (h *WorkspaceHandler) ListBookmarkFolders(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	folders, err := h.service.ListBookmarkFolders(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, folders)
}

func (h *WorkspaceHandler) UpdateBookmark(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	bookmarkID, err := uuid.Parse(c.Param("bookmarkId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bookmark ID"})
		return
	}

	var req models.UpdateBookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookmark, err := h.service.UpdateBookmark(c.Request.Context(), workspaceID, userID, bookmarkID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, bookmark)
}

func (h *WorkspaceHandler) DeleteBookmark(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	bookmarkID, err := uuid.Parse(c.Param("bookmarkId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bookmark ID"})
		return
	}

	if err := h.service.DeleteBookmark(c.Request.Context(), workspaceID, userID, bookmarkID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bookmark deleted"})
}

// ── Invitation History ──

func (h *WorkspaceHandler) ListInvitationHistory(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	records, total, err := h.service.ListInvitationHistory(c.Request.Context(), workspaceID, userID, page, perPage)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     records,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

func (h *WorkspaceHandler) GetInvitationStats(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	stats, err := h.service.GetInvitationStats(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ── Access Logs ──

func (h *WorkspaceHandler) ListAccessLogs(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	targetUser := c.Query("user_id")
	if targetUser != "" {
		targetUserID, err := uuid.Parse(targetUser)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		logs, total, err := h.service.ListAccessLogsByUser(c.Request.Context(), workspaceID, userID, targetUserID, page, perPage)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data":     logs,
			"total":    total,
			"page":     page,
			"per_page": perPage,
		})
		return
	}

	logs, total, err := h.service.ListAccessLogs(c.Request.Context(), workspaceID, userID, page, perPage)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     logs,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

func (h *WorkspaceHandler) GetAccessLogStats(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	stats, err := h.service.GetAccessLogStats(c.Request.Context(), workspaceID, userID, days)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ── Feature Flags ──

func (h *WorkspaceHandler) CreateFeatureFlag(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	var req models.CreateFeatureFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	flag, err := h.service.CreateFeatureFlag(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, flag)
}

func (h *WorkspaceHandler) ListFeatureFlags(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	flags, err := h.service.ListFeatureFlags(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, flags)
}

func (h *WorkspaceHandler) UpdateFeatureFlag(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	flagID, err := uuid.Parse(c.Param("flagId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid flag ID"})
		return
	}

	var req models.UpdateFeatureFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	flag, err := h.service.UpdateFeatureFlag(c.Request.Context(), workspaceID, userID, flagID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, flag)
}

func (h *WorkspaceHandler) DeleteFeatureFlag(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	flagID, err := uuid.Parse(c.Param("flagId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid flag ID"})
		return
	}

	if err := h.service.DeleteFeatureFlag(c.Request.Context(), workspaceID, userID, flagID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feature flag deleted"})
}

func (h *WorkspaceHandler) CheckFeatureFlag(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Feature flag key is required"})
		return
	}

	result, err := h.service.CheckFeatureFlag(c.Request.Context(), workspaceID, userID, key)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ── Integrations ──

func (h *WorkspaceHandler) CreateIntegration(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	var req models.CreateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	integration, err := h.service.CreateIntegration(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, integration)
}

func (h *WorkspaceHandler) ListIntegrations(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	integrations, err := h.service.ListIntegrations(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, integrations)
}

func (h *WorkspaceHandler) GetIntegration(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	integrationID, err := uuid.Parse(c.Param("integrationId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	integration, err := h.service.GetIntegration(c.Request.Context(), workspaceID, userID, integrationID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, integration)
}

func (h *WorkspaceHandler) UpdateIntegration(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	integrationID, err := uuid.Parse(c.Param("integrationId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	var req models.UpdateIntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	integration, err := h.service.UpdateIntegration(c.Request.Context(), workspaceID, userID, integrationID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, integration)
}

func (h *WorkspaceHandler) DeleteIntegration(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	integrationID, err := uuid.Parse(c.Param("integrationId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid integration ID"})
		return
	}

	if err := h.service.DeleteIntegration(c.Request.Context(), workspaceID, userID, integrationID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Integration deleted"})
}

// ── Labels ──

func (h *WorkspaceHandler) CreateLabel(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	var req models.CreateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	label, err := h.service.CreateLabel(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, label)
}

func (h *WorkspaceHandler) ListLabels(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	labels, err := h.service.ListLabels(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, labels)
}

func (h *WorkspaceHandler) UpdateLabel(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	labelID, err := uuid.Parse(c.Param("labelId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid label ID"})
		return
	}

	var req models.UpdateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	label, err := h.service.UpdateLabel(c.Request.Context(), workspaceID, userID, labelID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, label)
}

func (h *WorkspaceHandler) DeleteLabel(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	labelID, err := uuid.Parse(c.Param("labelId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid label ID"})
		return
	}

	if err := h.service.DeleteLabel(c.Request.Context(), workspaceID, userID, labelID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Label deleted"})
}

// ── Activity Streaks ──

func (h *WorkspaceHandler) RecordActivity(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	if err := h.service.RecordActivity(c.Request.Context(), workspaceID, userID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Activity recorded"})
}

func (h *WorkspaceHandler) GetMyStreak(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	streak, err := h.service.GetMyStreak(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, streak)
}

func (h *WorkspaceHandler) GetStreakLeaderboard(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	leaderboard, err := h.service.GetStreakLeaderboard(c.Request.Context(), workspaceID, userID, limit)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, leaderboard)
}

// ── Onboarding Checklists ──

func (h *WorkspaceHandler) CreateChecklist(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	var req models.CreateChecklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	checklist, err := h.service.CreateChecklist(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, checklist)
}

func (h *WorkspaceHandler) ListChecklists(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	checklists, err := h.service.ListChecklists(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, checklists)
}

func (h *WorkspaceHandler) GetChecklistWithSteps(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	checklistID, err := uuid.Parse(c.Param("checklistId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checklist ID"})
		return
	}

	checklist, err := h.service.GetChecklistWithSteps(c.Request.Context(), workspaceID, userID, checklistID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, checklist)
}

func (h *WorkspaceHandler) UpdateChecklist(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	checklistID, err := uuid.Parse(c.Param("checklistId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checklist ID"})
		return
	}

	var req models.UpdateChecklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	checklist, err := h.service.UpdateChecklist(c.Request.Context(), workspaceID, userID, checklistID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, checklist)
}

func (h *WorkspaceHandler) DeleteChecklist(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	checklistID, err := uuid.Parse(c.Param("checklistId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checklist ID"})
		return
	}

	if err := h.service.DeleteChecklist(c.Request.Context(), workspaceID, userID, checklistID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Checklist deleted"})
}

func (h *WorkspaceHandler) AddOnboardingStep(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	checklistID, err := uuid.Parse(c.Param("checklistId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid checklist ID"})
		return
	}

	var req models.AddStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	step, err := h.service.AddOnboardingStep(c.Request.Context(), workspaceID, userID, checklistID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, step)
}

func (h *WorkspaceHandler) DeleteOnboardingStep(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	stepID, err := uuid.Parse(c.Param("stepId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid step ID"})
		return
	}

	if err := h.service.DeleteOnboardingStep(c.Request.Context(), workspaceID, userID, stepID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Step deleted"})
}

func (h *WorkspaceHandler) CompleteOnboardingStep(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	stepID, err := uuid.Parse(c.Param("stepId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid step ID"})
		return
	}

	if err := h.service.CompleteOnboardingStep(c.Request.Context(), workspaceID, userID, stepID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Step completed"})
}

func (h *WorkspaceHandler) GetMyOnboardingStatus(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	statuses, err := h.service.GetMyOnboardingStatus(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, statuses)
}

// ── Compliance Policies ──

func (h *WorkspaceHandler) CreatePolicy(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	var req models.CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	policy, err := h.service.CreatePolicy(c.Request.Context(), workspaceID, userID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, policy)
}

func (h *WorkspaceHandler) ListPolicies(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	policies, err := h.service.ListPolicies(c.Request.Context(), workspaceID, userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, policies)
}

func (h *WorkspaceHandler) UpdatePolicy(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	policyID, err := uuid.Parse(c.Param("policyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid policy ID"})
		return
	}

	var req models.UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	policy, err := h.service.UpdatePolicy(c.Request.Context(), workspaceID, userID, policyID, &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, policy)
}

func (h *WorkspaceHandler) DeletePolicy(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	policyID, err := uuid.Parse(c.Param("policyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid policy ID"})
		return
	}

	if err := h.service.DeletePolicy(c.Request.Context(), workspaceID, userID, policyID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Policy deleted"})
}

func (h *WorkspaceHandler) AcknowledgePolicy(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	policyID, err := uuid.Parse(c.Param("policyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid policy ID"})
		return
	}

	if err := h.service.AcknowledgePolicy(c.Request.Context(), workspaceID, userID, policyID); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Policy acknowledged"})
}

func (h *WorkspaceHandler) GetPolicyComplianceStatus(c *gin.Context) {
	userID := getUserID(c)
	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	policyID, err := uuid.Parse(c.Param("policyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid policy ID"})
		return
	}

	status, err := h.service.GetPolicyComplianceStatus(c.Request.Context(), workspaceID, userID, policyID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, status)
}

// ── Helpers ──

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
	case service.ErrNotMember:
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a member of this workspace"})
	case service.ErrInviteNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Invite not found or expired"})
	case service.ErrAlreadyMember:
		c.JSON(http.StatusConflict, gin.H{"error": "Already a member"})
	case service.ErrInviteCodeNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Invite code not found or expired"})
	case service.ErrInviteCodeMaxUsed:
		c.JSON(http.StatusConflict, gin.H{"error": "Invite code has reached maximum uses"})
	case service.ErrCannotLeaveAsOwner:
		c.JSON(http.StatusConflict, gin.H{"error": "Owner cannot leave workspace, transfer ownership first"})
	case service.ErrRoleNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
	case service.ErrRoleNameExists:
		c.JSON(http.StatusConflict, gin.H{"error": "Role name already exists"})
	case service.ErrCannotDeleteDefault:
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete default role"})
	case service.ErrTemplateNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
	case service.ErrTagNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
	case service.ErrTagNameExists:
		c.JSON(http.StatusConflict, gin.H{"error": "Tag name already exists"})
	case service.ErrUserBanned:
		c.JSON(http.StatusForbidden, gin.H{"error": "User is banned from this workspace"})
	case service.ErrUserNotBanned:
		c.JSON(http.StatusNotFound, gin.H{"error": "User is not banned"})
	case service.ErrUserNotMuted:
		c.JSON(http.StatusNotFound, gin.H{"error": "User is not muted"})
	case service.ErrCannotBanOwner:
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot ban workspace owner"})
	case service.ErrCannotMuteOwner:
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot mute workspace owner"})
	case service.ErrAnnouncementNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Announcement not found"})
	case service.ErrWebhookNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
	case service.ErrAlreadyFavorited:
		c.JSON(http.StatusConflict, gin.H{"error": "Workspace already favorited"})
	case service.ErrNotFavorited:
		c.JSON(http.StatusNotFound, gin.H{"error": "Workspace is not favorited"})
	case service.ErrMemberNoteNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Member note not found"})
	case service.ErrScheduledActionNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Scheduled action not found"})
	case service.ErrScheduledActionPast:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Scheduled time must be in the future"})
	case service.ErrQuotaExceeded:
		c.JSON(http.StatusForbidden, gin.H{"error": "Workspace quota exceeded"})
	case service.ErrWorkspaceArchived:
		c.JSON(http.StatusConflict, gin.H{"error": "Workspace is archived"})
	case service.ErrWorkspaceNotArchived:
		c.JSON(http.StatusConflict, gin.H{"error": "Workspace is not archived"})
	case service.ErrPinnedItemNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Pinned item not found"})
	case service.ErrGroupNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
	case service.ErrGroupNameExists:
		c.JSON(http.StatusConflict, gin.H{"error": "Group name already exists"})
	case service.ErrAlreadyGroupMember:
		c.JSON(http.StatusConflict, gin.H{"error": "User is already a member of this group"})
	case service.ErrNotGroupMember:
		c.JSON(http.StatusNotFound, gin.H{"error": "User is not a member of this group"})
	case service.ErrCustomFieldNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Custom field not found"})
	case service.ErrCustomFieldNameExists:
		c.JSON(http.StatusConflict, gin.H{"error": "Custom field name already exists"})
	case service.ErrReactionExists:
		c.JSON(http.StatusConflict, gin.H{"error": "Reaction already exists"})
	case service.ErrBookmarkNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Bookmark not found"})
	case service.ErrBookmarkLimitReached:
		c.JSON(http.StatusForbidden, gin.H{"error": "Bookmark limit reached"})
	case service.ErrFeatureFlagNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature flag not found"})
	case service.ErrFeatureFlagKeyExists:
		c.JSON(http.StatusConflict, gin.H{"error": "Feature flag key already exists"})
	case service.ErrIntegrationNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
	case service.ErrLabelNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Label not found"})
	case service.ErrLabelNameExists:
		c.JSON(http.StatusConflict, gin.H{"error": "Label name already exists"})
	case service.ErrChecklistNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Checklist not found"})
	case service.ErrOnboardingStepNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Onboarding step not found"})
	case service.ErrPolicyNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Compliance policy not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
