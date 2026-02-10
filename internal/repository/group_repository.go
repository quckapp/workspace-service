package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/quckapp/workspace-service/internal/models"
)

type GroupRepository struct {
	db *sqlx.DB
}

func NewGroupRepository(db *sqlx.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) Create(ctx context.Context, group *models.MemberGroup) error {
	query := `
		INSERT INTO workspace_member_groups (id, workspace_id, name, description, color, created_by, member_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, group.ID, group.WorkspaceID, group.Name, group.Description, group.Color, group.CreatedBy, group.MemberCount, group.CreatedAt, group.UpdatedAt)
	return err
}

func (r *GroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.MemberGroup, error) {
	var group models.MemberGroup
	query := `SELECT * FROM workspace_member_groups WHERE id = ?`
	err := r.db.GetContext(ctx, &group, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &group, err
}

func (r *GroupRepository) GetByName(ctx context.Context, workspaceID uuid.UUID, name string) (*models.MemberGroup, error) {
	var group models.MemberGroup
	query := `SELECT * FROM workspace_member_groups WHERE workspace_id = ? AND name = ?`
	err := r.db.GetContext(ctx, &group, query, workspaceID, name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &group, err
}

func (r *GroupRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*models.MemberGroup, error) {
	var groups []*models.MemberGroup
	query := `SELECT * FROM workspace_member_groups WHERE workspace_id = ? ORDER BY name ASC`
	err := r.db.SelectContext(ctx, &groups, query, workspaceID)
	return groups, err
}

func (r *GroupRepository) Update(ctx context.Context, group *models.MemberGroup) error {
	query := `UPDATE workspace_member_groups SET name = ?, description = ?, color = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, group.Name, group.Description, group.Color, group.ID)
	return err
}

func (r *GroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM workspace_member_groups WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *GroupRepository) AddMember(ctx context.Context, membership *models.MemberGroupMembership) error {
	query := `
		INSERT INTO workspace_member_group_memberships (id, group_id, user_id, added_by, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, membership.ID, membership.GroupID, membership.UserID, membership.AddedBy, membership.CreatedAt)
	return err
}

func (r *GroupRepository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	query := `DELETE FROM workspace_member_group_memberships WHERE group_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, query, groupID, userID)
	return err
}

func (r *GroupRepository) ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	query := `SELECT user_id FROM workspace_member_group_memberships WHERE group_id = ? ORDER BY created_at ASC`
	err := r.db.SelectContext(ctx, &userIDs, query, groupID)
	return userIDs, err
}

func (r *GroupRepository) IsMemberOfGroup(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM workspace_member_group_memberships WHERE group_id = ? AND user_id = ?`
	err := r.db.GetContext(ctx, &count, query, groupID, userID)
	return count > 0, err
}

func (r *GroupRepository) IncrementMemberCount(ctx context.Context, groupID uuid.UUID) error {
	query := `UPDATE workspace_member_groups SET member_count = member_count + 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, groupID)
	return err
}

func (r *GroupRepository) DecrementMemberCount(ctx context.Context, groupID uuid.UUID) error {
	query := `UPDATE workspace_member_groups SET member_count = GREATEST(member_count - 1, 0) WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, groupID)
	return err
}

func (r *GroupRepository) ListGroupsByUser(ctx context.Context, workspaceID, userID uuid.UUID) ([]*models.MemberGroup, error) {
	var groups []*models.MemberGroup
	query := `
		SELECT g.* FROM workspace_member_groups g
		INNER JOIN workspace_member_group_memberships m ON g.id = m.group_id
		WHERE g.workspace_id = ? AND m.user_id = ?
		ORDER BY g.name ASC
	`
	err := r.db.SelectContext(ctx, &groups, query, workspaceID, userID)
	return groups, err
}

func (r *GroupRepository) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM workspace_member_groups WHERE workspace_id = ?`
	err := r.db.GetContext(ctx, &count, query, workspaceID)
	return count, err
}
