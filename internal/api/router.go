package api

import (
	"github.com/gin-gonic/gin"
	"github.com/quckapp/workspace-service/internal/config"
	"github.com/quckapp/workspace-service/internal/middleware"
	"github.com/quckapp/workspace-service/internal/service"
	"github.com/sirupsen/logrus"
)

func NewRouter(
	workspaceService *service.WorkspaceService,
	emojiService *service.EmojiService,
	billingService *service.BillingService,
	securityService *service.SecurityService,
	discoveryService *service.DiscoveryService,
	cfg *config.Config,
	logger *logrus.Logger,
) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy", "service": "workspace-service"})
	})

	api := r.Group("/api/v1")
	{
		handler := NewWorkspaceHandler(workspaceService, logger)
		emojiHandler := NewEmojiHandler(emojiService, logger)
		billingHandler := NewBillingHandler(billingService, logger)
		securityHandler := NewSecurityHandler(securityService, logger)
		discoveryHandler := NewDiscoveryHandler(discoveryService, logger)

		workspaces := api.Group("/workspaces")
		workspaces.Use(middleware.Auth(cfg.JWTSecret))
		{
			// Workspace CRUD
			workspaces.POST("", handler.CreateWorkspace)
			workspaces.GET("", handler.ListWorkspaces)
			workspaces.GET("/:id", handler.GetWorkspace)
			workspaces.PUT("/:id", handler.UpdateWorkspace)
			workspaces.DELETE("/:id", handler.DeleteWorkspace)

			// Workspace extras
			workspaces.GET("/:id/stats", handler.GetWorkspaceStats)
			workspaces.GET("/:id/settings", handler.GetWorkspaceSettings)
			workspaces.PUT("/:id/settings", handler.UpdateWorkspaceSettings)
			workspaces.POST("/:id/leave", handler.LeaveWorkspace)
			workspaces.POST("/:id/transfer-ownership", handler.TransferOwnership)
			workspaces.GET("/:id/analytics", handler.GetAnalytics)

			// Members
			workspaces.GET("/:id/members", handler.ListMembers)
			workspaces.GET("/:id/members/:userId", handler.GetMember)
			workspaces.POST("/:id/members/invite", handler.InviteMember)
			workspaces.POST("/:id/members/bulk-invite", handler.BulkInvite)
			workspaces.DELETE("/:id/members/:userId", handler.RemoveMember)
			workspaces.PUT("/:id/members/:userId/role", handler.UpdateMemberRole)

			// Member Profiles
			workspaces.GET("/:id/members/:userId/profile", handler.GetMemberProfile)
			workspaces.PUT("/:id/profile", handler.UpdateMemberProfile)
			workspaces.PUT("/:id/online-status", handler.SetOnlineStatus)

			// Invites
			workspaces.GET("/:id/invites", handler.ListInvites)
			workspaces.DELETE("/:id/invites/:inviteId", handler.RevokeInvite)

			// Invite Codes
			workspaces.POST("/:id/invite-codes", handler.CreateInviteCode)
			workspaces.GET("/:id/invite-codes", handler.ListInviteCodes)
			workspaces.DELETE("/:id/invite-codes/:codeId", handler.RevokeInviteCode)

			// Activity Log
			workspaces.GET("/:id/activity", handler.GetActivityLog)
			workspaces.GET("/:id/activity/actor/:actorId", handler.GetActivityLogByActor)

			// Custom Roles
			workspaces.POST("/:id/roles", handler.CreateRole)
			workspaces.GET("/:id/roles", handler.ListRoles)
			workspaces.PUT("/:id/roles/:roleId", handler.UpdateRole)
			workspaces.DELETE("/:id/roles/:roleId", handler.DeleteRole)

			// Templates (per-workspace)
			workspaces.POST("/:id/template", handler.CreateTemplateFromWorkspace)

			// Member Preferences
			workspaces.GET("/:id/preferences", handler.GetPreferences)
			workspaces.PUT("/:id/preferences", handler.UpdatePreferences)
			workspaces.DELETE("/:id/preferences", handler.ResetPreferences)

			// Tags
			workspaces.POST("/:id/tags", handler.CreateTag)
			workspaces.GET("/:id/tags", handler.ListTags)
			workspaces.PUT("/:id/tags/:tagId", handler.UpdateTag)
			workspaces.DELETE("/:id/tags/:tagId", handler.DeleteTag)

			// Moderation
			workspaces.POST("/:id/members/:userId/ban", handler.BanMember)
			workspaces.DELETE("/:id/members/:userId/ban", handler.UnbanMember)
			workspaces.POST("/:id/members/:userId/mute", handler.MuteMember)
			workspaces.DELETE("/:id/members/:userId/mute", handler.UnmuteMember)
			workspaces.GET("/:id/moderation", handler.GetModerationHistory)

			// Announcements
			workspaces.POST("/:id/announcements", handler.CreateAnnouncement)
			workspaces.GET("/:id/announcements", handler.ListAnnouncements)
			workspaces.PUT("/:id/announcements/:announcementId", handler.UpdateAnnouncement)
			workspaces.DELETE("/:id/announcements/:announcementId", handler.DeleteAnnouncement)
			workspaces.PUT("/:id/announcements/:announcementId/pin", handler.PinAnnouncement)

			// Webhooks
			workspaces.POST("/:id/webhooks", handler.CreateWebhook)
			workspaces.GET("/:id/webhooks", handler.ListWebhooks)
			workspaces.PUT("/:id/webhooks/:webhookId", handler.UpdateWebhook)
			workspaces.DELETE("/:id/webhooks/:webhookId", handler.DeleteWebhook)
			workspaces.POST("/:id/webhooks/:webhookId/test", handler.TestWebhook)

			// Favorites
			workspaces.POST("/:id/favorite", handler.FavoriteWorkspace)
			workspaces.DELETE("/:id/favorite", handler.UnfavoriteWorkspace)

			// Audit Export
			workspaces.GET("/:id/audit-export", handler.ExportAuditLog)

			// Member Notes
			workspaces.POST("/:id/members/:userId/notes", handler.CreateMemberNote)
			workspaces.GET("/:id/members/:userId/notes", handler.ListMemberNotes)
			workspaces.PUT("/:id/notes/:noteId", handler.UpdateMemberNote)
			workspaces.DELETE("/:id/notes/:noteId", handler.DeleteMemberNote)

			// Scheduled Actions
			workspaces.POST("/:id/scheduled-actions", handler.CreateScheduledAction)
			workspaces.GET("/:id/scheduled-actions", handler.ListScheduledActions)
			workspaces.PUT("/:id/scheduled-actions/:actionId", handler.UpdateScheduledAction)
			workspaces.POST("/:id/scheduled-actions/:actionId/cancel", handler.CancelScheduledAction)
			workspaces.DELETE("/:id/scheduled-actions/:actionId", handler.DeleteScheduledAction)

			// Usage Quotas
			workspaces.GET("/:id/quota", handler.GetQuotaUsage)
			workspaces.PUT("/:id/quota", handler.UpdateQuota)

			// Archive / Restore
			workspaces.POST("/:id/archive", handler.ArchiveWorkspace)
			workspaces.POST("/:id/restore", handler.RestoreWorkspace)

			// Cloning
			workspaces.POST("/:id/clone", handler.CloneWorkspace)

			// Pinned Items
			workspaces.POST("/:id/pins", handler.CreatePinnedItem)
			workspaces.GET("/:id/pins", handler.ListPinnedItems)
			workspaces.PUT("/:id/pins/:pinId", handler.UpdatePinnedItem)
			workspaces.DELETE("/:id/pins/:pinId", handler.DeletePinnedItem)
			workspaces.PUT("/:id/pins/reorder", handler.ReorderPins)

			// Member Groups / Teams
			workspaces.POST("/:id/groups", handler.CreateGroup)
			workspaces.GET("/:id/groups", handler.ListGroups)
			workspaces.GET("/:id/groups/:groupId", handler.GetGroup)
			workspaces.PUT("/:id/groups/:groupId", handler.UpdateGroup)
			workspaces.DELETE("/:id/groups/:groupId", handler.DeleteGroup)
			workspaces.POST("/:id/groups/:groupId/members", handler.AddGroupMembers)
			workspaces.DELETE("/:id/groups/:groupId/members/:userId", handler.RemoveGroupMember)
			workspaces.GET("/:id/members/:userId/groups", handler.ListUserGroups)

			// Custom Fields
			workspaces.POST("/:id/custom-fields", handler.CreateCustomField)
			workspaces.GET("/:id/custom-fields", handler.ListCustomFields)
			workspaces.PUT("/:id/custom-fields/:fieldId", handler.UpdateCustomField)
			workspaces.DELETE("/:id/custom-fields/:fieldId", handler.DeleteCustomField)
			workspaces.PUT("/:id/custom-fields/:fieldId/value", handler.SetCustomFieldValue)
			workspaces.GET("/:id/custom-fields/values", handler.GetCustomFieldValues)

			// Reactions
			workspaces.POST("/:id/reactions", handler.AddReaction)
			workspaces.DELETE("/:id/reactions", handler.RemoveReaction)
			workspaces.GET("/:id/reactions", handler.ListReactions)
			workspaces.GET("/:id/reactions/summary", handler.GetReactionSummary)

			// Bookmarks
			workspaces.POST("/:id/bookmarks", handler.CreateBookmark)
			workspaces.GET("/:id/bookmarks", handler.ListBookmarks)
			workspaces.GET("/:id/bookmarks/folders", handler.ListBookmarkFolders)
			workspaces.PUT("/:id/bookmarks/:bookmarkId", handler.UpdateBookmark)
			workspaces.DELETE("/:id/bookmarks/:bookmarkId", handler.DeleteBookmark)

			// Invitation History
			workspaces.GET("/:id/invitation-history", handler.ListInvitationHistory)
			workspaces.GET("/:id/invitation-stats", handler.GetInvitationStats)

			// Access Logs
			workspaces.GET("/:id/access-logs", handler.ListAccessLogs)
			workspaces.GET("/:id/access-logs/stats", handler.GetAccessLogStats)

			// Feature Flags
			workspaces.POST("/:id/feature-flags", handler.CreateFeatureFlag)
			workspaces.GET("/:id/feature-flags", handler.ListFeatureFlags)
			workspaces.PUT("/:id/feature-flags/:flagId", handler.UpdateFeatureFlag)
			workspaces.DELETE("/:id/feature-flags/:flagId", handler.DeleteFeatureFlag)
			workspaces.GET("/:id/feature-flags/:key/check", handler.CheckFeatureFlag)

			// Integrations
			workspaces.POST("/:id/integrations", handler.CreateIntegration)
			workspaces.GET("/:id/integrations", handler.ListIntegrations)
			workspaces.GET("/:id/integrations/:integrationId", handler.GetIntegration)
			workspaces.PUT("/:id/integrations/:integrationId", handler.UpdateIntegration)
			workspaces.DELETE("/:id/integrations/:integrationId", handler.DeleteIntegration)

			// Labels
			workspaces.POST("/:id/labels", handler.CreateLabel)
			workspaces.GET("/:id/labels", handler.ListLabels)
			workspaces.PUT("/:id/labels/:labelId", handler.UpdateLabel)
			workspaces.DELETE("/:id/labels/:labelId", handler.DeleteLabel)

			// Activity Streaks
			workspaces.POST("/:id/streaks/record", handler.RecordActivity)
			workspaces.GET("/:id/streaks/me", handler.GetMyStreak)
			workspaces.GET("/:id/streaks/leaderboard", handler.GetStreakLeaderboard)

			// Onboarding Checklists
			workspaces.POST("/:id/onboarding", handler.CreateChecklist)
			workspaces.GET("/:id/onboarding", handler.ListChecklists)
			workspaces.GET("/:id/onboarding/:checklistId", handler.GetChecklistWithSteps)
			workspaces.PUT("/:id/onboarding/:checklistId", handler.UpdateChecklist)
			workspaces.DELETE("/:id/onboarding/:checklistId", handler.DeleteChecklist)
			workspaces.POST("/:id/onboarding/:checklistId/steps", handler.AddOnboardingStep)
			workspaces.DELETE("/:id/onboarding/steps/:stepId", handler.DeleteOnboardingStep)
			workspaces.POST("/:id/onboarding/steps/:stepId/complete", handler.CompleteOnboardingStep)
			workspaces.GET("/:id/onboarding/status", handler.GetMyOnboardingStatus)

			// Compliance Policies
			workspaces.POST("/:id/policies", handler.CreatePolicy)
			workspaces.GET("/:id/policies", handler.ListPolicies)
			workspaces.PUT("/:id/policies/:policyId", handler.UpdatePolicy)
			workspaces.DELETE("/:id/policies/:policyId", handler.DeletePolicy)
			workspaces.POST("/:id/policies/:policyId/acknowledge", handler.AcknowledgePolicy)
			workspaces.GET("/:id/policies/:policyId/compliance", handler.GetPolicyComplianceStatus)

			// ── NEW: Custom Emojis ──
			workspaces.POST("/:id/emojis", emojiHandler.CreateEmoji)
			workspaces.GET("/:id/emojis", emojiHandler.ListEmojis)
			workspaces.GET("/:id/emojis/search", emojiHandler.SearchEmojis)
			workspaces.GET("/:id/emojis/categories", emojiHandler.GetCategories)
			workspaces.GET("/:id/emojis/stats", emojiHandler.GetEmojiStats)
			workspaces.GET("/:id/emojis/:emojiId", emojiHandler.GetEmoji)
			workspaces.PUT("/:id/emojis/:emojiId", emojiHandler.UpdateEmoji)
			workspaces.DELETE("/:id/emojis/:emojiId", emojiHandler.DeleteEmoji)
			workspaces.POST("/:id/emojis/:emojiId/usage", emojiHandler.IncrementUsage)
			workspaces.POST("/:id/emojis/bulk-delete", emojiHandler.BulkDeleteEmojis)
			workspaces.POST("/:id/emoji-packs", emojiHandler.CreatePack)
			workspaces.GET("/:id/emoji-packs", emojiHandler.ListPacks)
			workspaces.GET("/:id/emoji-packs/:packId/emojis", emojiHandler.GetPackEmojis)
			workspaces.DELETE("/:id/emoji-packs/:packId", emojiHandler.DeletePack)

			// ── NEW: Billing & Plans ──
			workspaces.GET("/:id/billing", billingHandler.GetBillingOverview)
			workspaces.GET("/:id/billing/plan", billingHandler.GetPlan)
			workspaces.PUT("/:id/billing/plan", billingHandler.ChangePlan)
			workspaces.DELETE("/:id/billing/plan", billingHandler.CancelPlan)
			workspaces.POST("/:id/billing/seats/add", billingHandler.AddSeats)
			workspaces.POST("/:id/billing/seats/remove", billingHandler.RemoveSeats)
			workspaces.GET("/:id/billing/invoices", billingHandler.ListInvoices)
			workspaces.GET("/:id/billing/invoices/:invoiceId", billingHandler.GetInvoice)
			workspaces.POST("/:id/billing/payment-methods", billingHandler.AddPaymentMethod)
			workspaces.GET("/:id/billing/payment-methods", billingHandler.ListPaymentMethods)
			workspaces.PUT("/:id/billing/payment-methods/:methodId/default", billingHandler.SetDefaultPaymentMethod)
			workspaces.DELETE("/:id/billing/payment-methods/:methodId", billingHandler.DeletePaymentMethod)
			workspaces.GET("/:id/billing/events", billingHandler.ListBillingEvents)

			// ── NEW: Security ──
			workspaces.GET("/:id/security", securityHandler.GetSecurityOverview)
			workspaces.GET("/:id/security/policy", securityHandler.GetSecurityPolicy)
			workspaces.PUT("/:id/security/policy", securityHandler.UpdateSecurityPolicy)
			workspaces.POST("/:id/security/ip-allowlist", securityHandler.AddIPEntry)
			workspaces.GET("/:id/security/ip-allowlist", securityHandler.ListIPEntries)
			workspaces.PUT("/:id/security/ip-allowlist/:entryId", securityHandler.UpdateIPEntry)
			workspaces.DELETE("/:id/security/ip-allowlist/:entryId", securityHandler.DeleteIPEntry)
			workspaces.GET("/:id/security/sessions", securityHandler.ListAllSessions)
			workspaces.GET("/:id/security/sessions/me", securityHandler.ListMySessions)
			workspaces.DELETE("/:id/security/sessions/:sessionId", securityHandler.RevokeSession)
			workspaces.POST("/:id/security/sessions/revoke", securityHandler.RevokeSessions)
			workspaces.GET("/:id/security/audit", securityHandler.ListSecurityAudit)

			// ── NEW: Directory ──
			workspaces.GET("/:id/directory", discoveryHandler.GetDirectoryEntry)
			workspaces.PUT("/:id/directory", discoveryHandler.UpdateDirectoryEntry)
		}

		// Search (auth required)
		api.GET("/search/workspaces", middleware.Auth(cfg.JWTSecret), handler.SearchWorkspaces)

		// Join by invite code (auth required)
		api.POST("/join", middleware.Auth(cfg.JWTSecret), handler.JoinByCode)

		// Invite acceptance (auth required)
		api.POST("/invites/:token/accept", middleware.Auth(cfg.JWTSecret), handler.AcceptInvite)

		// Templates (standalone routes to avoid :id path collision)
		api.POST("/workspaces/from-template", middleware.Auth(cfg.JWTSecret), handler.CreateWorkspaceFromTemplate)
		api.GET("/templates", middleware.Auth(cfg.JWTSecret), handler.ListTemplates)
		api.GET("/templates/:templateId", middleware.Auth(cfg.JWTSecret), handler.GetTemplate)
		api.PUT("/templates/:templateId", middleware.Auth(cfg.JWTSecret), handler.UpdateTemplate)
		api.DELETE("/templates/:templateId", middleware.Auth(cfg.JWTSecret), handler.DeleteTemplate)

		// Favorites (standalone, not workspace-scoped)
		api.GET("/favorites", middleware.Auth(cfg.JWTSecret), handler.ListFavorites)
		api.PUT("/favorites/reorder", middleware.Auth(cfg.JWTSecret), handler.ReorderFavorites)

		// Archived workspaces (standalone)
		api.GET("/workspaces/archived", middleware.Auth(cfg.JWTSecret), handler.ListArchivedWorkspaces)

		// ── NEW: Billing Plans (standalone) ──
		api.GET("/plans", middleware.Auth(cfg.JWTSecret), billingHandler.GetAvailablePlans)
		api.GET("/plans/:planType", middleware.Auth(cfg.JWTSecret), billingHandler.GetPlanFeatures)

		// ── NEW: Discovery (standalone) ──
		discovery := api.Group("/discovery")
		discovery.Use(middleware.Auth(cfg.JWTSecret))
		{
			discovery.GET("/search", discoveryHandler.SearchDirectory)
			discovery.GET("/categories", discoveryHandler.GetCategories)
			discovery.GET("/trending", discoveryHandler.GetTrending)
			discovery.GET("/recommendations", discoveryHandler.GetRecommendations)
			discovery.POST("/recommendations/dismiss", discoveryHandler.DismissRecommendation)
		}
	}

	return r
}
