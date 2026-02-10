package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type CustomFieldRepository struct {
	db *sqlx.DB
}

func NewCustomFieldRepository(db *sqlx.DB) *CustomFieldRepository {
	return &CustomFieldRepository{db: db}
}

func (r *CustomFieldRepository) Create(ctx context.Context, field *models.WorkspaceCustomField) error {
	query := `
		INSERT INTO workspace_custom_fields (id, workspace_id, name, field_type, options, default_value, is_required, position, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, field.ID, field.WorkspaceID, field.Name, field.FieldType, field.Options, field.DefaultValue, field.IsRequired, field.Position, field.CreatedBy, field.CreatedAt, field.UpdatedAt)
	return err
}

func (r *CustomFieldRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.WorkspaceCustomField, error) {
	var field models.WorkspaceCustomField
	query := `SELECT * FROM workspace_custom_fields WHERE id = ?`
	err := r.db.GetContext(ctx, &field, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &field, err
}

func (r *CustomFieldRepository) GetByName(ctx context.Context, workspaceID uuid.UUID, name string) (*models.WorkspaceCustomField, error) {
	var field models.WorkspaceCustomField
	query := `SELECT * FROM workspace_custom_fields WHERE workspace_id = ? AND name = ?`
	err := r.db.GetContext(ctx, &field, query, workspaceID, name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &field, err
}

func (r *CustomFieldRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*models.WorkspaceCustomField, error) {
	var fields []*models.WorkspaceCustomField
	query := `SELECT * FROM workspace_custom_fields WHERE workspace_id = ? ORDER BY position ASC, name ASC`
	err := r.db.SelectContext(ctx, &fields, query, workspaceID)
	return fields, err
}

func (r *CustomFieldRepository) Update(ctx context.Context, field *models.WorkspaceCustomField) error {
	query := `UPDATE workspace_custom_fields SET name = ?, options = ?, default_value = ?, is_required = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, field.Name, field.Options, field.DefaultValue, field.IsRequired, field.ID)
	return err
}

func (r *CustomFieldRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Delete field values first, then the field itself
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM workspace_custom_field_values WHERE field_id = ?`, id)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM workspace_custom_fields WHERE id = ?`, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *CustomFieldRepository) GetMaxPosition(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var maxPos sql.NullInt64
	query := `SELECT MAX(position) FROM workspace_custom_fields WHERE workspace_id = ?`
	err := r.db.GetContext(ctx, &maxPos, query, workspaceID)
	if err != nil || !maxPos.Valid {
		return 0, err
	}
	return int(maxPos.Int64), nil
}

// ── Field Values ──

func (r *CustomFieldRepository) SetValue(ctx context.Context, value *models.WorkspaceCustomFieldValue) error {
	query := `
		INSERT INTO workspace_custom_field_values (id, field_id, entity_id, value, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE value = VALUES(value), updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query, value.ID, value.FieldID, value.EntityID, value.Value, value.CreatedAt, value.UpdatedAt)
	return err
}

func (r *CustomFieldRepository) GetValue(ctx context.Context, fieldID, entityID uuid.UUID) (*models.WorkspaceCustomFieldValue, error) {
	var value models.WorkspaceCustomFieldValue
	query := `SELECT * FROM workspace_custom_field_values WHERE field_id = ? AND entity_id = ?`
	err := r.db.GetContext(ctx, &value, query, fieldID, entityID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &value, err
}

func (r *CustomFieldRepository) ListValuesByEntity(ctx context.Context, entityID uuid.UUID) ([]*models.WorkspaceCustomFieldValue, error) {
	var values []*models.WorkspaceCustomFieldValue
	query := `SELECT * FROM workspace_custom_field_values WHERE entity_id = ? ORDER BY created_at ASC`
	err := r.db.SelectContext(ctx, &values, query, entityID)
	return values, err
}

func (r *CustomFieldRepository) DeleteValue(ctx context.Context, fieldID, entityID uuid.UUID) error {
	query := `DELETE FROM workspace_custom_field_values WHERE field_id = ? AND entity_id = ?`
	_, err := r.db.ExecContext(ctx, query, fieldID, entityID)
	return err
}

func (r *CustomFieldRepository) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM workspace_custom_fields WHERE workspace_id = ?`
	err := r.db.GetContext(ctx, &count, query, workspaceID)
	return count, err
}
