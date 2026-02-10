package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type AccessLogRepository struct {
	db *sqlx.DB
}

func NewAccessLogRepository(db *sqlx.DB) *AccessLogRepository {
	return &AccessLogRepository{db: db}
}

func (r *AccessLogRepository) Create(ctx context.Context, log *models.WorkspaceAccessLog) error {
	query := `
		INSERT INTO workspace_access_logs (id, workspace_id, user_id, action, resource, ip_address, user_agent, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, log.ID, log.WorkspaceID, log.UserID, log.Action, log.Resource, log.IPAddress, log.UserAgent, log.CreatedAt)
	return err
}

func (r *AccessLogRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, page, perPage int) ([]*models.WorkspaceAccessLog, int64, error) {
	var logs []*models.WorkspaceAccessLog
	var total int64
	offset := (page - 1) * perPage

	countQuery := `SELECT COUNT(*) FROM workspace_access_logs WHERE workspace_id = ?`
	r.db.GetContext(ctx, &total, countQuery, workspaceID)

	query := `SELECT * FROM workspace_access_logs WHERE workspace_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	err := r.db.SelectContext(ctx, &logs, query, workspaceID, perPage, offset)
	return logs, total, err
}

func (r *AccessLogRepository) ListByUser(ctx context.Context, workspaceID, userID uuid.UUID, page, perPage int) ([]*models.WorkspaceAccessLog, int64, error) {
	var logs []*models.WorkspaceAccessLog
	var total int64
	offset := (page - 1) * perPage

	countQuery := `SELECT COUNT(*) FROM workspace_access_logs WHERE workspace_id = ? AND user_id = ?`
	r.db.GetContext(ctx, &total, countQuery, workspaceID, userID)

	query := `SELECT * FROM workspace_access_logs WHERE workspace_id = ? AND user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	err := r.db.SelectContext(ctx, &logs, query, workspaceID, userID, perPage, offset)
	return logs, total, err
}

func (r *AccessLogRepository) GetStats(ctx context.Context, workspaceID uuid.UUID, days int) (*models.AccessLogStats, error) {
	stats := &models.AccessLogStats{}

	// Total accesses
	r.db.GetContext(ctx, &stats.TotalAccesses, `SELECT COUNT(*) FROM workspace_access_logs WHERE workspace_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)`, workspaceID, days)

	// Unique users
	r.db.GetContext(ctx, &stats.UniqueUsers, `SELECT COUNT(DISTINCT user_id) FROM workspace_access_logs WHERE workspace_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)`, workspaceID, days)

	// Top resources
	resourceQuery := `
		SELECT resource, COUNT(*) as count FROM workspace_access_logs
		WHERE workspace_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
		GROUP BY resource ORDER BY count DESC LIMIT 10
	`
	r.db.SelectContext(ctx, &stats.TopResources, resourceQuery, workspaceID, days)

	// Daily counts
	dailyQuery := `
		SELECT DATE(created_at) as date, COUNT(*) as count FROM workspace_access_logs
		WHERE workspace_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
		GROUP BY DATE(created_at) ORDER BY date ASC
	`
	r.db.SelectContext(ctx, &stats.AccessesByDay, dailyQuery, workspaceID, days)

	return stats, nil
}
