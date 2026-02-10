package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type InvitationHistoryRepository struct {
	db *sqlx.DB
}

func NewInvitationHistoryRepository(db *sqlx.DB) *InvitationHistoryRepository {
	return &InvitationHistoryRepository{db: db}
}

func (r *InvitationHistoryRepository) Create(ctx context.Context, record *models.InvitationHistory) error {
	query := `
		INSERT INTO workspace_invitation_history (id, workspace_id, inviter_id, invitee_email, invitee_id, method, role, status, accepted_at, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, record.ID, record.WorkspaceID, record.InviterID, record.InviteeEmail, record.InviteeID, record.Method, record.Role, record.Status, record.AcceptedAt, record.ExpiresAt, record.CreatedAt)
	return err
}

func (r *InvitationHistoryRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, page, perPage int) ([]*models.InvitationHistory, int64, error) {
	var records []*models.InvitationHistory
	var total int64
	offset := (page - 1) * perPage

	countQuery := `SELECT COUNT(*) FROM workspace_invitation_history WHERE workspace_id = ?`
	r.db.GetContext(ctx, &total, countQuery, workspaceID)

	query := `SELECT * FROM workspace_invitation_history WHERE workspace_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	err := r.db.SelectContext(ctx, &records, query, workspaceID, perPage, offset)
	return records, total, err
}

func (r *InvitationHistoryRepository) ListByInviter(ctx context.Context, workspaceID, inviterID uuid.UUID) ([]*models.InvitationHistory, error) {
	var records []*models.InvitationHistory
	query := `SELECT * FROM workspace_invitation_history WHERE workspace_id = ? AND inviter_id = ? ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &records, query, workspaceID, inviterID)
	return records, err
}

func (r *InvitationHistoryRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE workspace_invitation_history SET status = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *InvitationHistoryRepository) GetStats(ctx context.Context, workspaceID uuid.UUID) (*models.InvitationStats, error) {
	stats := &models.InvitationStats{ByMethod: make(map[string]int)}

	// Total counts by status
	type statusCount struct {
		Status string `db:"status"`
		Count  int    `db:"count"`
	}
	var statusCounts []statusCount
	statusQuery := `SELECT status, COUNT(*) as count FROM workspace_invitation_history WHERE workspace_id = ? GROUP BY status`
	r.db.SelectContext(ctx, &statusCounts, statusQuery, workspaceID)

	for _, sc := range statusCounts {
		switch sc.Status {
		case "accepted":
			stats.TotalAccepted = sc.Count
		case "pending":
			stats.TotalPending = sc.Count
		case "expired":
			stats.TotalExpired = sc.Count
		}
		stats.TotalSent += sc.Count
	}

	// By method
	type methodCount struct {
		Method string `db:"method"`
		Count  int    `db:"count"`
	}
	var methodCounts []methodCount
	methodQuery := `SELECT method, COUNT(*) as count FROM workspace_invitation_history WHERE workspace_id = ? GROUP BY method`
	r.db.SelectContext(ctx, &methodCounts, methodQuery, workspaceID)

	for _, mc := range methodCounts {
		stats.ByMethod[mc.Method] = mc.Count
	}

	return stats, nil
}
