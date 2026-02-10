package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type OnboardingRepository struct {
	db *sqlx.DB
}

func NewOnboardingRepository(db *sqlx.DB) *OnboardingRepository {
	return &OnboardingRepository{db: db}
}

// ── Checklists ──

func (r *OnboardingRepository) CreateChecklist(ctx context.Context, checklist *models.OnboardingChecklist) error {
	query := `INSERT INTO onboarding_checklists (id, workspace_id, title, description, is_active, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		checklist.ID, checklist.WorkspaceID, checklist.Title, checklist.Description,
		checklist.IsActive, checklist.CreatedBy, checklist.CreatedAt, checklist.UpdatedAt)
	return err
}

func (r *OnboardingRepository) GetChecklistByID(ctx context.Context, id uuid.UUID) (*models.OnboardingChecklist, error) {
	var checklist models.OnboardingChecklist
	query := `SELECT * FROM onboarding_checklists WHERE id = ?`
	err := r.db.GetContext(ctx, &checklist, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &checklist, err
}

func (r *OnboardingRepository) ListChecklists(ctx context.Context, workspaceID uuid.UUID) ([]*models.OnboardingChecklist, error) {
	var checklists []*models.OnboardingChecklist
	query := `SELECT * FROM onboarding_checklists WHERE workspace_id = ? ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &checklists, query, workspaceID)
	return checklists, err
}

func (r *OnboardingRepository) UpdateChecklist(ctx context.Context, checklist *models.OnboardingChecklist) error {
	query := `UPDATE onboarding_checklists SET title = ?, description = ?, is_active = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, checklist.Title, checklist.Description, checklist.IsActive, checklist.ID)
	return err
}

func (r *OnboardingRepository) DeleteChecklist(ctx context.Context, id uuid.UUID) error {
	// Steps and progress are cascade-deleted via FK
	query := `DELETE FROM onboarding_checklists WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ── Steps ──

func (r *OnboardingRepository) AddStep(ctx context.Context, step *models.OnboardingStep) error {
	query := `INSERT INTO onboarding_steps (id, checklist_id, title, description, action_type, action_data, position, is_required, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		step.ID, step.ChecklistID, step.Title, step.Description,
		step.ActionType, step.ActionData, step.Position, step.IsRequired, step.CreatedAt)
	return err
}

func (r *OnboardingRepository) ListSteps(ctx context.Context, checklistID uuid.UUID) ([]models.OnboardingStep, error) {
	var steps []models.OnboardingStep
	query := `SELECT * FROM onboarding_steps WHERE checklist_id = ? ORDER BY position ASC`
	err := r.db.SelectContext(ctx, &steps, query, checklistID)
	return steps, err
}

func (r *OnboardingRepository) DeleteStep(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM onboarding_steps WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *OnboardingRepository) GetStepByID(ctx context.Context, id uuid.UUID) (*models.OnboardingStep, error) {
	var step models.OnboardingStep
	query := `SELECT * FROM onboarding_steps WHERE id = ?`
	err := r.db.GetContext(ctx, &step, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &step, err
}

func (r *OnboardingRepository) GetMaxStepPosition(ctx context.Context, checklistID uuid.UUID) (int, error) {
	var pos sql.NullInt64
	query := `SELECT MAX(position) FROM onboarding_steps WHERE checklist_id = ?`
	err := r.db.GetContext(ctx, &pos, query, checklistID)
	if err != nil || !pos.Valid {
		return 0, err
	}
	return int(pos.Int64), nil
}

// ── Progress ──

func (r *OnboardingRepository) CompleteStep(ctx context.Context, progress *models.OnboardingProgress) error {
	query := `INSERT INTO onboarding_progress (id, step_id, user_id, completed_at, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE completed_at = VALUES(completed_at)`
	_, err := r.db.ExecContext(ctx, query,
		progress.ID, progress.StepID, progress.UserID, progress.CompletedAt, progress.CreatedAt)
	return err
}

func (r *OnboardingRepository) GetProgress(ctx context.Context, checklistID, userID uuid.UUID) ([]*models.OnboardingProgress, error) {
	var progress []*models.OnboardingProgress
	query := `SELECT p.* FROM onboarding_progress p
		JOIN onboarding_steps s ON s.id = p.step_id
		WHERE s.checklist_id = ? AND p.user_id = ?`
	err := r.db.SelectContext(ctx, &progress, query, checklistID, userID)
	return progress, err
}

func (r *OnboardingRepository) UncompleteStep(ctx context.Context, stepID, userID uuid.UUID) error {
	query := `DELETE FROM onboarding_progress WHERE step_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, stepID, userID)
	return err
}
