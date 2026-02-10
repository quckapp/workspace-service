package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type DiscoveryRepository struct {
	db *sqlx.DB
}

func NewDiscoveryRepository(db *sqlx.DB) *DiscoveryRepository {
	return &DiscoveryRepository{db: db}
}

// Directory
func (r *DiscoveryRepository) GetDirectoryEntry(ctx context.Context, workspaceID uuid.UUID) (*models.WorkspaceDirectoryEntry, error) {
	var entry models.WorkspaceDirectoryEntry
	err := r.db.GetContext(ctx, &entry, "SELECT * FROM workspace_directory WHERE workspace_id = ?", workspaceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &entry, err
}

func (r *DiscoveryRepository) UpsertDirectoryEntry(ctx context.Context, entry *models.WorkspaceDirectoryEntry) error {
	query := `INSERT INTO workspace_directory (id, workspace_id, is_listed, category, tags, description, member_count, icon_url, banner_url, website, is_verified, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE is_listed = VALUES(is_listed), category = VALUES(category), tags = VALUES(tags), description = VALUES(description), member_count = VALUES(member_count), icon_url = VALUES(icon_url), banner_url = VALUES(banner_url), website = VALUES(website), updated_at = VALUES(updated_at)`
	_, err := r.db.ExecContext(ctx, query, entry.ID, entry.WorkspaceID, entry.IsListed, entry.Category, entry.Tags, entry.Description, entry.MemberCount, entry.IconURL, entry.BannerURL, entry.Website, entry.IsVerified, entry.CreatedAt, entry.UpdatedAt)
	return err
}

func (r *DiscoveryRepository) SearchDirectory(ctx context.Context, query string, category string, sortBy string, limit, offset int) ([]*models.WorkspaceDirectoryEntry, error) {
	var entries []*models.WorkspaceDirectoryEntry
	q := "SELECT * FROM workspace_directory WHERE is_listed = TRUE"
	args := []interface{}{}

	if query != "" {
		q += " AND (description LIKE ? OR category LIKE ?)"
		args = append(args, "%"+query+"%", "%"+query+"%")
	}
	if category != "" {
		q += " AND category = ?"
		args = append(args, category)
	}

	switch sortBy {
	case "name":
		q += " ORDER BY description ASC"
	case "created_at":
		q += " ORDER BY created_at DESC"
	default:
		q += " ORDER BY member_count DESC"
	}

	q += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &entries, q, args...)
	return entries, err
}

func (r *DiscoveryRepository) GetCategories(ctx context.Context) ([]models.WorkspaceCategory, error) {
	var categories []models.WorkspaceCategory
	err := r.db.SelectContext(ctx, &categories, "SELECT COALESCE(category, 'other') as name, COUNT(*) as count FROM workspace_directory WHERE is_listed = TRUE GROUP BY category ORDER BY count DESC")
	return categories, err
}

func (r *DiscoveryRepository) CountListed(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM workspace_directory WHERE is_listed = TRUE")
	return count, err
}

// Recommendations
func (r *DiscoveryRepository) CreateRecommendation(ctx context.Context, rec *models.WorkspaceRecommendation) error {
	query := `INSERT INTO workspace_recommendations (id, user_id, workspace_id, reason, score, is_dismissed, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, rec.ID, rec.UserID, rec.WorkspaceID, rec.Reason, rec.Score, rec.IsDismissed, rec.CreatedAt)
	return err
}

func (r *DiscoveryRepository) GetRecommendations(ctx context.Context, userID uuid.UUID, limit int) ([]*models.WorkspaceRecommendation, error) {
	var recs []*models.WorkspaceRecommendation
	err := r.db.SelectContext(ctx, &recs, "SELECT * FROM workspace_recommendations WHERE user_id = ? AND is_dismissed = FALSE ORDER BY score DESC LIMIT ?", userID, limit)
	return recs, err
}

func (r *DiscoveryRepository) DismissRecommendation(ctx context.Context, userID, workspaceID uuid.UUID) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, "UPDATE workspace_recommendations SET is_dismissed = TRUE, dismissed_at = ? WHERE user_id = ? AND workspace_id = ?", now, userID, workspaceID)
	return err
}

func (r *DiscoveryRepository) GetTrending(ctx context.Context, limit int) ([]*models.TrendingWorkspace, error) {
	var trending []*models.TrendingWorkspace
	query := `SELECT w.*, d.member_count, 0 as growth_rate, 0 as active_users
		FROM workspaces w
		JOIN workspace_directory d ON w.id = d.workspace_id
		WHERE d.is_listed = TRUE AND w.is_active = TRUE AND w.deleted_at IS NULL
		ORDER BY d.member_count DESC LIMIT ?`
	err := r.db.SelectContext(ctx, &trending, query, limit)
	return trending, err
}
