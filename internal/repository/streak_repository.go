package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type StreakRepository struct {
	db *sqlx.DB
}

func NewStreakRepository(db *sqlx.DB) *StreakRepository {
	return &StreakRepository{db: db}
}

func (r *StreakRepository) Upsert(ctx context.Context, streak *models.MemberActivityStreak) error {
	query := `INSERT INTO member_activity_streaks (id, workspace_id, user_id, current_streak, longest_streak, total_active_days, activity_score, last_active_date, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		current_streak = VALUES(current_streak),
		longest_streak = VALUES(longest_streak),
		total_active_days = VALUES(total_active_days),
		activity_score = VALUES(activity_score),
		last_active_date = VALUES(last_active_date),
		updated_at = VALUES(updated_at)`
	_, err := r.db.ExecContext(ctx, query,
		streak.ID, streak.WorkspaceID, streak.UserID,
		streak.CurrentStreak, streak.LongestStreak, streak.TotalActiveDays,
		streak.ActivityScore, streak.LastActiveDate, streak.UpdatedAt)
	return err
}

func (r *StreakRepository) GetByUserID(ctx context.Context, workspaceID, userID uuid.UUID) (*models.MemberActivityStreak, error) {
	var streak models.MemberActivityStreak
	query := `SELECT * FROM member_activity_streaks WHERE workspace_id = ? AND user_id = ?`
	err := r.db.GetContext(ctx, &streak, query, workspaceID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &streak, err
}

func (r *StreakRepository) GetLeaderboard(ctx context.Context, workspaceID uuid.UUID, limit int) ([]models.StreakLeaderboard, error) {
	var leaderboard []models.StreakLeaderboard
	query := `SELECT user_id, current_streak, longest_streak, activity_score
		FROM member_activity_streaks
		WHERE workspace_id = ?
		ORDER BY activity_score DESC, current_streak DESC
		LIMIT ?`
	err := r.db.SelectContext(ctx, &leaderboard, query, workspaceID, limit)
	return leaderboard, err
}

func (r *StreakRepository) RecordDailyActivity(ctx context.Context, workspaceID, userID uuid.UUID) error {
	today := time.Now().Format("2006-01-02")

	existing, err := r.GetByUserID(ctx, workspaceID, userID)
	if err != nil {
		return err
	}

	if existing == nil {
		streak := &models.MemberActivityStreak{
			ID:              uuid.New(),
			WorkspaceID:     workspaceID,
			UserID:          userID,
			CurrentStreak:   1,
			LongestStreak:   1,
			TotalActiveDays: 1,
			ActivityScore:   1.0,
			LastActiveDate:  today,
			UpdatedAt:       time.Now(),
		}
		return r.Upsert(ctx, streak)
	}

	// Already logged today
	if existing.LastActiveDate == today {
		return nil
	}

	// Check if yesterday was active (continue streak)
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if existing.LastActiveDate == yesterday {
		existing.CurrentStreak++
	} else {
		existing.CurrentStreak = 1
	}

	if existing.CurrentStreak > existing.LongestStreak {
		existing.LongestStreak = existing.CurrentStreak
	}

	existing.TotalActiveDays++
	existing.ActivityScore = float64(existing.TotalActiveDays) * (1.0 + float64(existing.CurrentStreak)*0.1)
	existing.LastActiveDate = today
	existing.UpdatedAt = time.Now()

	return r.Upsert(ctx, existing)
}

func (r *StreakRepository) ResetStreak(ctx context.Context, workspaceID, userID uuid.UUID) error {
	query := `UPDATE member_activity_streaks SET current_streak = 0, activity_score = 0, updated_at = NOW() WHERE workspace_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, workspaceID, userID)
	return err
}
