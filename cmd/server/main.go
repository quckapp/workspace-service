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
	logger.Info("Repositories initialized")

	// Initialize service
	workspaceService := service.NewWorkspaceService(
		workspaceRepo,
		memberRepo,
		inviteRepo,
		redisClient,
		kafkaProducer,
		logger,
	)
	logger.Info("Service layer initialized")

	// Initialize router
	router := api.NewRouter(workspaceService, cfg, logger)
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
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	return nil
}
