package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type ActivityRepository struct {
	db *sqlx.DB
}

func NewActivityRepository(db *sqlx.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) Create(ctx context.Context, log *models.ActivityLog) error {
	query := `
		INSERT INTO workspace_activity_log (id, workspace_id, actor_id, action, entity_type, entity_id, details, ip_address, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, log.ID, log.WorkspaceID, log.ActorID, log.Action, log.EntityType, log.EntityID, log.Details, log.IPAddress, log.CreatedAt)
	return err
}

func (r *ActivityRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, page, perPage int) ([]*models.ActivityLog, int64, error) {
	var activities []*models.ActivityLog
	var total int64
	offset := (page - 1) * perPage

	countQuery := `SELECT COUNT(*) FROM workspace_activity_log WHERE workspace_id = ?`
	r.db.GetContext(ctx, &total, countQuery, workspaceID)

	query := `
		SELECT * FROM workspace_activity_log WHERE workspace_id = ?
		ORDER BY created_at DESC LIMIT ? OFFSET ?
	`
	err := r.db.SelectContext(ctx, &activities, query, workspaceID, perPage, offset)
	return activities, total, err
}

func (r *ActivityRepository) ListByActor(ctx context.Context, workspaceID, actorID uuid.UUID, page, perPage int) ([]*models.ActivityLog, int64, error) {
	var activities []*models.ActivityLog
	var total int64
	offset := (page - 1) * perPage

	countQuery := `SELECT COUNT(*) FROM workspace_activity_log WHERE workspace_id = ? AND actor_id = ?`
	r.db.GetContext(ctx, &total, countQuery, workspaceID, actorID)

	query := `
		SELECT * FROM workspace_activity_log WHERE workspace_id = ? AND actor_id = ?
		ORDER BY created_at DESC LIMIT ? OFFSET ?
	`
	err := r.db.SelectContext(ctx, &activities, query, workspaceID, actorID, perPage, offset)
	return activities, total, err
}

func (r *ActivityRepository) ListByAction(ctx context.Context, workspaceID uuid.UUID, action string, page, perPage int) ([]*models.ActivityLog, int64, error) {
	var activities []*models.ActivityLog
	var total int64
	offset := (page - 1) * perPage

	countQuery := `SELECT COUNT(*) FROM workspace_activity_log WHERE workspace_id = ? AND action = ?`
	r.db.GetContext(ctx, &total, countQuery, workspaceID, action)

	query := `
		SELECT * FROM workspace_activity_log WHERE workspace_id = ? AND action = ?
		ORDER BY created_at DESC LIMIT ? OFFSET ?
	`
	err := r.db.SelectContext(ctx, &activities, query, workspaceID, action, perPage, offset)
	return activities, total, err
}

func (r *ActivityRepository) GetTopContributors(ctx context.Context, workspaceID uuid.UUID, since time.Time, limit int) ([]models.ContributorStat, error) {
	var stats []models.ContributorStat
	query := `
		SELECT actor_id as user_id, COUNT(*) as actions FROM workspace_activity_log
		WHERE workspace_id = ? AND created_at >= ?
		GROUP BY actor_id ORDER BY actions DESC LIMIT ?
	`
	err := r.db.SelectContext(ctx, &stats, query, workspaceID, since, limit)
	return stats, err
}

func (r *ActivityRepository) ListByDateRange(ctx context.Context, workspaceID uuid.UUID, startDate, endDate *time.Time, actionType string) ([]*models.ActivityLog, int64, error) {
	var activities []*models.ActivityLog
	var total int64

	query := "SELECT * FROM workspace_activity_log WHERE workspace_id = ?"
	countQuery := "SELECT COUNT(*) FROM workspace_activity_log WHERE workspace_id = ?"
	args := []interface{}{workspaceID}

	if startDate != nil {
		query += " AND created_at >= ?"
		countQuery += " AND created_at >= ?"
		args = append(args, *startDate)
	}
	if endDate != nil {
		query += " AND created_at <= ?"
		countQuery += " AND created_at <= ?"
		args = append(args, *endDate)
	}
	if actionType != "" {
		query += " AND action = ?"
		countQuery += " AND action = ?"
		args = append(args, actionType)
	}

	r.db.GetContext(ctx, &total, countQuery, args...)

	query += " ORDER BY created_at DESC"
	err := r.db.SelectContext(ctx, &activities, query, args...)
	return activities, total, err
}

func (r *ActivityRepository) GetDailyActionCounts(ctx context.Context, workspaceID uuid.UUID, days int) ([]models.DailyCount, error) {
	var counts []models.DailyCount
	query := `
		SELECT DATE(created_at) as date, COUNT(*) as count FROM workspace_activity_log
		WHERE workspace_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
		GROUP BY DATE(created_at) ORDER BY date ASC
	`
	err := r.db.SelectContext(ctx, &counts, query, workspaceID, days)
	return counts, err
}
