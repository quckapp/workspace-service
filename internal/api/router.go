package api

import (
	"github.com/gin-gonic/gin"
	"github.com/quckapp/workspace-service/internal/config"
	"github.com/quckapp/workspace-service/internal/middleware"
	"github.com/quckapp/workspace-service/internal/service"
	"github.com/sirupsen/logrus"
)

func NewRouter(workspaceService *service.WorkspaceService, cfg *config.Config, logger *logrus.Logger) *gin.Engine {
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

		workspaces := api.Group("/workspaces")
		workspaces.Use(middleware.Auth(cfg.JWTSecret))
		{
			workspaces.POST("", handler.CreateWorkspace)
			workspaces.GET("", handler.ListWorkspaces)
			workspaces.GET("/:id", handler.GetWorkspace)
			workspaces.PUT("/:id", handler.UpdateWorkspace)
			workspaces.DELETE("/:id", handler.DeleteWorkspace)

			// Members
			workspaces.GET("/:id/members", handler.ListMembers)
			workspaces.POST("/:id/members/invite", handler.InviteMember)
			workspaces.DELETE("/:id/members/:userId", handler.RemoveMember)
			workspaces.PUT("/:id/members/:userId/role", handler.UpdateMemberRole)
		}

		// Invite acceptance (no auth required)
		api.POST("/invites/:token/accept", middleware.Auth(cfg.JWTSecret), handler.AcceptInvite)
	}

	return r
}
