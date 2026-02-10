package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/api"
	"github.com/quckapp/workspace-service/internal/config"
	"github.com/quckapp/workspace-service/internal/db"
	"github.com/quckapp/workspace-service/internal/repository"
	"github.com/quckapp/workspace-service/internal/service"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting workspace service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	if cfg.Environment == "development" {
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.WithFields(logrus.Fields{
		"port":        cfg.Port,
		"environment": cfg.Environment,
	}).Info("Configuration loaded")

	// Initialize MySQL database
	mysqlDB, err := db.NewMySQL(cfg.DatabaseURL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to MySQL")
	}
	defer mysqlDB.Close()
	logger.Info("Connected to MySQL database")

	// Run database migrations
	if err := runMigrations(mysqlDB); err != nil {
		logger.WithError(err).Warn("Failed to run migrations")
	}

	// Initialize Redis
	redisClient, err := db.NewRedis(cfg.RedisURL)
	if err != nil {
		logger.WithError(err).Warn("Failed to connect to Redis, continuing without cache")
		redisClient = nil
	} else {
		defer redisClient.Close()
		logger.Info("Connected to Redis")
	}

	// Initialize Kafka producer
	var kafkaProducer *db.KafkaProducer
	if len(cfg.KafkaBrokers) > 0 && cfg.KafkaBrokers[0] != "" {
		kafkaProducer, err = db.NewKafkaProducer(cfg.KafkaBrokers)
		if err != nil {
			logger.WithError(err).Warn("Failed to connect to Kafka, continuing without events")
			kafkaProducer = nil
		} else {
			defer kafkaProducer.Close()
			logger.WithField("brokers", cfg.KafkaBrokers).Info("Connected to Kafka")
		}
	}

	// Initialize repositories
	workspaceRepo := repository.NewWorkspaceRepository(mysqlDB)
	memberRepo := repository.NewMemberRepository(mysqlDB)
	inviteRepo := repository.NewInviteRepository(mysqlDB)
	inviteCodeRepo := repository.NewInviteCodeRepository(mysqlDB)
	activityRepo := repository.NewActivityRepository(mysqlDB)
	profileRepo := repository.NewProfileRepository(mysqlDB)
	roleRepo := repository.NewRoleRepository(mysqlDB)
	templateRepo := repository.NewTemplateRepository(mysqlDB)
	preferenceRepo := repository.NewPreferenceRepository(mysqlDB)
	tagRepo := repository.NewTagRepository(mysqlDB)
	moderationRepo := repository.NewModerationRepository(mysqlDB)
	announcementRepo := repository.NewAnnouncementRepository(mysqlDB)
	webhookRepo := repository.NewWebhookRepository(mysqlDB)
	favoriteRepo := repository.NewFavoriteRepository(mysqlDB)
	memberNoteRepo := repository.NewMemberNoteRepository(mysqlDB)
	scheduledActionRepo := repository.NewScheduledActionRepository(mysqlDB)
	quotaRepo := repository.NewQuotaRepository(mysqlDB)
	pinnedItemRepo := repository.NewPinnedItemRepository(mysqlDB)
	groupRepo := repository.NewGroupRepository(mysqlDB)
	customFieldRepo := repository.NewCustomFieldRepository(mysqlDB)
	reactionRepo := repository.NewReactionRepository(mysqlDB)
	bookmarkRepo := repository.NewBookmarkRepository(mysqlDB)
	invitationHistoryRepo := repository.NewInvitationHistoryRepository(mysqlDB)
	accessLogRepo := repository.NewAccessLogRepository(mysqlDB)
	featureFlagRepo := repository.NewFeatureFlagRepository(mysqlDB)
	integrationRepo := repository.NewIntegrationRepository(mysqlDB)
	labelRepo := repository.NewLabelRepository(mysqlDB)
	streakRepo := repository.NewStreakRepository(mysqlDB)
	onboardingRepo := repository.NewOnboardingRepository(mysqlDB)
	complianceRepo := repository.NewComplianceRepository(mysqlDB)
	emojiRepo := repository.NewEmojiRepository(mysqlDB)
	billingRepo := repository.NewBillingRepository(mysqlDB)
	securityRepo := repository.NewSecurityRepository(mysqlDB)
	discoveryRepo := repository.NewDiscoveryRepository(mysqlDB)
	logger.Info("Repositories initialized")

	// Initialize service
	workspaceService := service.NewWorkspaceService(
		workspaceRepo,
		memberRepo,
		inviteRepo,
		inviteCodeRepo,
		activityRepo,
		profileRepo,
		roleRepo,
		templateRepo,
		preferenceRepo,
		tagRepo,
		moderationRepo,
		announcementRepo,
		webhookRepo,
		favoriteRepo,
		memberNoteRepo,
		scheduledActionRepo,
		quotaRepo,
		pinnedItemRepo,
		groupRepo,
		customFieldRepo,
		reactionRepo,
		bookmarkRepo,
		invitationHistoryRepo,
		accessLogRepo,
		featureFlagRepo,
		integrationRepo,
		labelRepo,
		streakRepo,
		onboardingRepo,
		complianceRepo,
		redisClient,
		kafkaProducer,
		logger,
	)
	emojiService := service.NewEmojiService(emojiRepo, memberRepo, logger)
	billingService := service.NewBillingService(billingRepo, memberRepo, logger)
	securityService := service.NewSecurityService(securityRepo, memberRepo, logger)
	discoveryService := service.NewDiscoveryService(discoveryRepo, workspaceRepo, memberRepo, logger)
	logger.Info("Service layer initialized")

	// Initialize router
	router := api.NewRouter(workspaceService, emojiService, billingService, securityService, discoveryService, cfg, logger)
	logger.Info("HTTP router initialized")

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.WithField("port", cfg.Port).Info("Workspace service listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down workspace service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Workspace service stopped")
}

func runMigrations(db *sqlx.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS workspaces (
			id CHAR(36) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			slug VARCHAR(50) NOT NULL UNIQUE,
			description TEXT,
			icon_url VARCHAR(500),
			owner_id CHAR(36) NOT NULL,
			plan VARCHAR(20) DEFAULT 'free',
			settings JSON,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP NULL,
			INDEX idx_owner_id (owner_id),
			INDEX idx_slug (slug),
			INDEX idx_deleted_at (deleted_at)
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_members (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'member',
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			invited_by CHAR(36),
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_user (workspace_id, user_id),
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_user_id (user_id),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_invites (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			email VARCHAR(255) NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'member',
			token VARCHAR(64) NOT NULL UNIQUE,
			invited_by CHAR(36) NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			accepted_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_token (token),
			INDEX idx_email (email),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_invite_codes (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			code VARCHAR(20) NOT NULL UNIQUE,
			role VARCHAR(20) NOT NULL DEFAULT 'member',
			max_uses INT DEFAULT 0,
			use_count INT DEFAULT 0,
			created_by CHAR(36) NOT NULL,
			expires_at TIMESTAMP NULL,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_code (code),
			INDEX idx_workspace_id (workspace_id),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_activity_log (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			actor_id CHAR(36) NOT NULL,
			action VARCHAR(50) NOT NULL,
			entity_type VARCHAR(50) NOT NULL,
			entity_id VARCHAR(36),
			details JSON,
			ip_address VARCHAR(45),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_actor_id (actor_id),
			INDEX idx_action (action),
			INDEX idx_created_at (created_at),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_member_profiles (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			display_name VARCHAR(100),
			title VARCHAR(100),
			status_text VARCHAR(255),
			status_emoji VARCHAR(10),
			timezone VARCHAR(50),
			is_online BOOLEAN DEFAULT FALSE,
			last_seen_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_user_profile (workspace_id, user_id),
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_user_id (user_id),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_roles (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			name VARCHAR(50) NOT NULL,
			color VARCHAR(7),
			priority INT DEFAULT 0,
			permissions JSON,
			is_default BOOLEAN DEFAULT FALSE,
			created_by CHAR(36) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_role_name (workspace_id, name),
			INDEX idx_workspace_id (workspace_id),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_templates (
			id CHAR(36) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			created_by CHAR(36) NOT NULL,
			default_roles JSON,
			default_channels JSON,
			default_settings JSON,
			is_public BOOLEAN DEFAULT FALSE,
			use_count INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_created_by (created_by),
			INDEX idx_is_public (is_public)
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_member_preferences (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			notification_level VARCHAR(20) DEFAULT 'all',
			email_notifications BOOLEAN DEFAULT TRUE,
			mute_until TIMESTAMP NULL,
			sidebar_position INT DEFAULT 0,
			theme VARCHAR(50),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_user_pref (workspace_id, user_id),
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_user_id (user_id),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_tags (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			name VARCHAR(50) NOT NULL,
			color VARCHAR(7),
			created_by CHAR(36) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_tag_name (workspace_id, name),
			INDEX idx_workspace_id (workspace_id),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_bans (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			banned_by CHAR(36) NOT NULL,
			reason TEXT,
			expires_at TIMESTAMP NULL,
			is_permanent BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_ban_user (workspace_id, user_id),
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_user_id (user_id),
			INDEX idx_expires_at (expires_at),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_mutes (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			muted_by CHAR(36) NOT NULL,
			reason TEXT,
			expires_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_mute_user (workspace_id, user_id),
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_user_id (user_id),
			INDEX idx_expires_at (expires_at),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_announcements (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			title VARCHAR(200) NOT NULL,
			content TEXT NOT NULL,
			priority VARCHAR(20) DEFAULT 'normal',
			author_id CHAR(36) NOT NULL,
			is_pinned BOOLEAN DEFAULT FALSE,
			expires_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_author_id (author_id),
			INDEX idx_priority (priority),
			INDEX idx_is_pinned (is_pinned),
			INDEX idx_expires_at (expires_at),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_webhooks (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			name VARCHAR(100) NOT NULL,
			url VARCHAR(500) NOT NULL,
			secret VARCHAR(64) NOT NULL,
			events JSON,
			is_active BOOLEAN DEFAULT TRUE,
			created_by CHAR(36) NOT NULL,
			last_triggered_at TIMESTAMP NULL,
			failure_count INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_is_active (is_active),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_favorites (
			id CHAR(36) PRIMARY KEY,
			user_id CHAR(36) NOT NULL,
			workspace_id CHAR(36) NOT NULL,
			position INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_user_workspace_fav (user_id, workspace_id),
			INDEX idx_user_id (user_id),
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_position (position),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_member_notes (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			target_id CHAR(36) NOT NULL,
			author_id CHAR(36) NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_target_id (target_id),
			INDEX idx_author_id (author_id),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_scheduled_actions (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			action_type VARCHAR(50) NOT NULL,
			payload JSON,
			scheduled_at TIMESTAMP NOT NULL,
			executed_at TIMESTAMP NULL,
			status VARCHAR(20) DEFAULT 'pending',
			created_by CHAR(36) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_status (status),
			INDEX idx_scheduled_at (scheduled_at),
			INDEX idx_created_by (created_by),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_quotas (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			max_members INT DEFAULT 100,
			max_channels INT DEFAULT 50,
			max_storage_mb INT DEFAULT 5120,
			max_invite_codes INT DEFAULT 10,
			max_webhooks INT DEFAULT 5,
			max_roles INT DEFAULT 10,
			current_members INT DEFAULT 0,
			current_channels INT DEFAULT 0,
			current_storage_mb INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_quota (workspace_id),
			INDEX idx_workspace_id (workspace_id),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_pinned_items (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			item_type VARCHAR(20) NOT NULL,
			item_id VARCHAR(36),
			title VARCHAR(200) NOT NULL,
			content TEXT,
			url VARCHAR(500),
			pinned_by CHAR(36) NOT NULL,
			position INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_item_type (item_type),
			INDEX idx_pinned_by (pinned_by),
			INDEX idx_position (position),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_member_groups (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			color VARCHAR(7),
			created_by CHAR(36) NOT NULL,
			member_count INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_group_name (workspace_id, name),
			INDEX idx_workspace_id (workspace_id),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_member_group_memberships (
			id CHAR(36) PRIMARY KEY,
			group_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			added_by CHAR(36) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_group_user (group_id, user_id),
			INDEX idx_group_id (group_id),
			INDEX idx_user_id (user_id),
			FOREIGN KEY (group_id) REFERENCES workspace_member_groups(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_custom_fields (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			name VARCHAR(100) NOT NULL,
			field_type VARCHAR(20) NOT NULL,
			options JSON,
			default_value VARCHAR(500),
			is_required BOOLEAN DEFAULT FALSE,
			position INT DEFAULT 0,
			created_by CHAR(36) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_field_name (workspace_id, name),
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_field_type (field_type),
			INDEX idx_position (position),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_custom_field_values (
			id CHAR(36) PRIMARY KEY,
			field_id CHAR(36) NOT NULL,
			entity_id CHAR(36) NOT NULL,
			value TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_field_entity (field_id, entity_id),
			INDEX idx_field_id (field_id),
			INDEX idx_entity_id (entity_id),
			FOREIGN KEY (field_id) REFERENCES workspace_custom_fields(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_reactions (
			id CHAR(36) PRIMARY KEY,
			entity_type VARCHAR(50) NOT NULL,
			entity_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			emoji VARCHAR(50) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_entity_user_emoji (entity_type, entity_id, user_id, emoji),
			INDEX idx_entity (entity_type, entity_id),
			INDEX idx_user_id (user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_bookmarks (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			title VARCHAR(200) NOT NULL,
			url VARCHAR(500),
			entity_type VARCHAR(50),
			entity_id CHAR(36),
			notes TEXT,
			folder_name VARCHAR(100),
			position INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_workspace_user (workspace_id, user_id),
			INDEX idx_folder (workspace_id, user_id, folder_name),
			INDEX idx_position (position),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_invitation_history (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			inviter_id CHAR(36) NOT NULL,
			invitee_email VARCHAR(255) NOT NULL,
			invitee_id CHAR(36),
			method VARCHAR(20) NOT NULL,
			role VARCHAR(20) NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			accepted_at TIMESTAMP NULL,
			expires_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_inviter_id (inviter_id),
			INDEX idx_status (status),
			INDEX idx_created_at (created_at),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_access_logs (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			action VARCHAR(50) NOT NULL,
			resource VARCHAR(200),
			ip_address VARCHAR(45),
			user_agent VARCHAR(500),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_user_id (user_id),
			INDEX idx_action (action),
			INDEX idx_created_at (created_at),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_feature_flags (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			` + "`key`" + ` VARCHAR(100) NOT NULL,
			enabled BOOLEAN DEFAULT FALSE,
			description TEXT,
			metadata JSON,
			created_by CHAR(36) NOT NULL,
			updated_by CHAR(36) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_flag_key (workspace_id, ` + "`key`" + `),
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_key (` + "`key`" + `),
			INDEX idx_enabled (enabled),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_integrations (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			provider VARCHAR(50) NOT NULL,
			name VARCHAR(100) NOT NULL,
			status VARCHAR(20) DEFAULT 'active',
			config JSON,
			credentials TEXT,
			webhook_url VARCHAR(500),
			last_sync_at TIMESTAMP NULL,
			error_message TEXT,
			created_by CHAR(36) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_provider (provider),
			INDEX idx_status (status),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_labels (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			name VARCHAR(50) NOT NULL,
			color VARCHAR(7) NOT NULL,
			description TEXT,
			position INT DEFAULT 0,
			usage_count INT DEFAULT 0,
			created_by CHAR(36) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_label_name (workspace_id, name),
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_position (position),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS member_activity_streaks (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			current_streak INT DEFAULT 0,
			longest_streak INT DEFAULT 0,
			total_active_days INT DEFAULT 0,
			activity_score DOUBLE DEFAULT 0,
			last_active_date VARCHAR(10),
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_user_streak (workspace_id, user_id),
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_activity_score (activity_score),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS onboarding_checklists (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			title VARCHAR(200) NOT NULL,
			description TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			created_by CHAR(36) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_is_active (is_active),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS onboarding_steps (
			id CHAR(36) PRIMARY KEY,
			checklist_id CHAR(36) NOT NULL,
			title VARCHAR(200) NOT NULL,
			description TEXT,
			action_type VARCHAR(20) NOT NULL,
			action_data TEXT,
			position INT DEFAULT 0,
			is_required BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_checklist_id (checklist_id),
			INDEX idx_position (position),
			FOREIGN KEY (checklist_id) REFERENCES onboarding_checklists(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS onboarding_progress (
			id CHAR(36) PRIMARY KEY,
			step_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			completed_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_step_user (step_id, user_id),
			INDEX idx_step_id (step_id),
			INDEX idx_user_id (user_id),
			FOREIGN KEY (step_id) REFERENCES onboarding_steps(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS compliance_policies (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			name VARCHAR(200) NOT NULL,
			description TEXT,
			policy_type VARCHAR(30) NOT NULL,
			rules JSON,
			severity VARCHAR(20) DEFAULT 'info',
			is_enforced BOOLEAN DEFAULT FALSE,
			created_by CHAR(36) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_policy_type (policy_type),
			INDEX idx_severity (severity),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS policy_acknowledgements (
			id CHAR(36) PRIMARY KEY,
			policy_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			acked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_policy_user (policy_id, user_id),
			INDEX idx_policy_id (policy_id),
			INDEX idx_user_id (user_id),
			FOREIGN KEY (policy_id) REFERENCES compliance_policies(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_custom_emojis (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			name VARCHAR(100) NOT NULL,
			image_url VARCHAR(500) NOT NULL,
			category VARCHAR(50),
			created_by CHAR(36) NOT NULL,
			is_animated BOOLEAN DEFAULT FALSE,
			alias_for VARCHAR(100),
			usage_count INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_emoji_name (workspace_id, name),
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_category (category),
			INDEX idx_usage_count (usage_count),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_emoji_packs (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			created_by CHAR(36) NOT NULL,
			is_default BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_pack_name (workspace_id, name),
			INDEX idx_workspace_id (workspace_id),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_emoji_pack_mappings (
			id CHAR(36) PRIMARY KEY,
			pack_id CHAR(36) NOT NULL,
			emoji_id CHAR(36) NOT NULL,
			position INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_pack_emoji (pack_id, emoji_id),
			INDEX idx_pack_id (pack_id),
			INDEX idx_emoji_id (emoji_id),
			FOREIGN KEY (pack_id) REFERENCES workspace_emoji_packs(id) ON DELETE CASCADE,
			FOREIGN KEY (emoji_id) REFERENCES workspace_custom_emojis(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_plans (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			plan_type VARCHAR(20) NOT NULL DEFAULT 'free',
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			seats INT DEFAULT 5,
			price_per_seat INT DEFAULT 0,
			billing_cycle VARCHAR(20) DEFAULT 'monthly',
			current_period_start TIMESTAMP NULL,
			current_period_end TIMESTAMP NULL,
			canceled_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_plan (workspace_id),
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_plan_type (plan_type),
			INDEX idx_status (status),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_invoices (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			amount INT NOT NULL,
			currency VARCHAR(3) DEFAULT 'USD',
			status VARCHAR(20) DEFAULT 'pending',
			description TEXT,
			period_start TIMESTAMP NULL,
			period_end TIMESTAMP NULL,
			paid_at TIMESTAMP NULL,
			due_date TIMESTAMP NULL,
			pdf_url VARCHAR(500),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_status (status),
			INDEX idx_created_at (created_at),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_payment_methods (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			method_type VARCHAR(20) NOT NULL,
			last_four VARCHAR(4),
			brand VARCHAR(20),
			exp_month INT,
			exp_year INT,
			is_default BOOLEAN DEFAULT FALSE,
			billing_email VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_is_default (is_default),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_billing_events (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			event_type VARCHAR(50) NOT NULL,
			description TEXT,
			amount INT,
			metadata JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_event_type (event_type),
			INDEX idx_created_at (created_at),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_ip_allowlist (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			ip_address VARCHAR(45) NOT NULL,
			label VARCHAR(100),
			added_by CHAR(36) NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_ip (workspace_id, ip_address),
			INDEX idx_workspace_id (workspace_id),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_sessions (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			user_id CHAR(36) NOT NULL,
			ip_address VARCHAR(45),
			user_agent VARCHAR(500),
			device_type VARCHAR(20),
			location VARCHAR(200),
			is_active BOOLEAN DEFAULT TRUE,
			last_active_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_user_id (user_id),
			INDEX idx_is_active (is_active),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_security_policies (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			two_factor_required BOOLEAN DEFAULT FALSE,
			session_timeout_minutes INT DEFAULT 480,
			password_min_length INT DEFAULT 8,
			ip_allowlist_enabled BOOLEAN DEFAULT FALSE,
			max_sessions_per_user INT DEFAULT 5,
			allow_external_sharing BOOLEAN DEFAULT TRUE,
			require_email_verification BOOLEAN DEFAULT TRUE,
			updated_by CHAR(36),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_security_policy (workspace_id),
			INDEX idx_workspace_id (workspace_id),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_security_audit (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			event_type VARCHAR(50) NOT NULL,
			severity VARCHAR(20) DEFAULT 'info',
			actor_id CHAR(36),
			description TEXT,
			ip_address VARCHAR(45),
			metadata JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_workspace_id (workspace_id),
			INDEX idx_event_type (event_type),
			INDEX idx_severity (severity),
			INDEX idx_created_at (created_at),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_directory (
			id CHAR(36) PRIMARY KEY,
			workspace_id CHAR(36) NOT NULL,
			is_listed BOOLEAN DEFAULT FALSE,
			short_description VARCHAR(200),
			category VARCHAR(50),
			tags JSON,
			member_count INT DEFAULT 0,
			website_url VARCHAR(500),
			verified BOOLEAN DEFAULT FALSE,
			featured BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_workspace_directory (workspace_id),
			INDEX idx_is_listed (is_listed),
			INDEX idx_category (category),
			INDEX idx_featured (featured),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_recommendations (
			id CHAR(36) PRIMARY KEY,
			user_id CHAR(36) NOT NULL,
			workspace_id CHAR(36) NOT NULL,
			score DOUBLE DEFAULT 0,
			reason VARCHAR(200),
			is_dismissed BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_user_workspace_rec (user_id, workspace_id),
			INDEX idx_user_id (user_id),
			INDEX idx_is_dismissed (is_dismissed),
			INDEX idx_score (score),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	return nil
}
