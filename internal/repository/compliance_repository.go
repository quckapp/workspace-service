package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type ComplianceRepository struct {
	db *sqlx.DB
}

func NewComplianceRepository(db *sqlx.DB) *ComplianceRepository {
	return &ComplianceRepository{db: db}
}

func (r *ComplianceRepository) Create(ctx context.Context, policy *models.CompliancePolicy) error {
	query := `INSERT INTO compliance_policies (id, workspace_id, name, description, policy_type, rules, severity, is_enforced, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		policy.ID, policy.WorkspaceID, policy.Name, policy.Description,
		policy.PolicyType, policy.Rules, policy.Severity, policy.IsEnforced,
		policy.CreatedBy, policy.CreatedAt, policy.UpdatedAt)
	return err
}

func (r *ComplianceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CompliancePolicy, error) {
	var policy models.CompliancePolicy
	query := `SELECT * FROM compliance_policies WHERE id = ?`
	err := r.db.GetContext(ctx, &policy, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &policy, err
}

func (r *ComplianceRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*models.CompliancePolicy, error) {
	var policies []*models.CompliancePolicy
	query := `SELECT * FROM compliance_policies WHERE workspace_id = ? ORDER BY severity DESC, name ASC`
	err := r.db.SelectContext(ctx, &policies, query, workspaceID)
	return policies, err
}

func (r *ComplianceRepository) ListByType(ctx context.Context, workspaceID uuid.UUID, policyType string) ([]*models.CompliancePolicy, error) {
	var policies []*models.CompliancePolicy
	query := `SELECT * FROM compliance_policies WHERE workspace_id = ? AND policy_type = ? ORDER BY name ASC`
	err := r.db.SelectContext(ctx, &policies, query, workspaceID, policyType)
	return policies, err
}

func (r *ComplianceRepository) Update(ctx context.Context, policy *models.CompliancePolicy) error {
	query := `UPDATE compliance_policies SET name = ?, description = ?, rules = ?, severity = ?, is_enforced = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query,
		policy.Name, policy.Description, policy.Rules, policy.Severity, policy.IsEnforced, policy.ID)
	return err
}

func (r *ComplianceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM compliance_policies WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ── Acknowledgements ──

func (r *ComplianceRepository) Acknowledge(ctx context.Context, ack *models.PolicyAcknowledgement) error {
	query := `INSERT INTO policy_acknowledgements (id, policy_id, user_id, acked_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE acked_at = VALUES(acked_at)`
	_, err := r.db.ExecContext(ctx, query, ack.ID, ack.PolicyID, ack.UserID, ack.AckedAt)
	return err
}

func (r *ComplianceRepository) HasAcknowledged(ctx context.Context, policyID, userID uuid.UUID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM policy_acknowledgements WHERE policy_id = ? AND user_id = ?`
	err := r.db.GetContext(ctx, &count, query, policyID, userID)
	return count > 0, err
}

func (r *ComplianceRepository) GetAcknowledgementCount(ctx context.Context, policyID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM policy_acknowledgements WHERE policy_id = ?`
	err := r.db.GetContext(ctx, &count, query, policyID)
	return count, err
}

func (r *ComplianceRepository) ListAcknowledgements(ctx context.Context, policyID uuid.UUID) ([]*models.PolicyAcknowledgement, error) {
	var acks []*models.PolicyAcknowledgement
	query := `SELECT * FROM policy_acknowledgements WHERE policy_id = ? ORDER BY acked_at DESC`
	err := r.db.SelectContext(ctx, &acks, query, policyID)
	return acks, err
}

func (r *ComplianceRepository) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM compliance_policies WHERE workspace_id = ?`
	err := r.db.GetContext(ctx, &count, query, workspaceID)
	return count, err
}
