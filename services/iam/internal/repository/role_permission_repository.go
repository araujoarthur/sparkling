package repository

import (
	"context"
	"fmt"

	"github.com/araujoarthur/intranetbackend/services/iam/internal/repository/sqlc/generated"
	"github.com/google/uuid"
)

// RolePermissionRepository defines the data access contract for managing
// the assignment of permissions to roles.
// Consumers should depend on this interface, never on the concrete implementation.
type RolePermissionRepository interface {
	// Assign adds a permission to a role.
	// If the assignment already exists it is silently ignored (idempotent).
	Assign(ctx context.Context, roleID, permissionID uuid.UUID) error

	// Remove revokes a permission from a role.
	// Returns ErrNotFound if the assignment does not exist.
	Remove(ctx context.Context, roleID, permissionID uuid.UUID) error

	// RoleHasPermission reports whether a role has been assigned a specific permission.
	RoleHasPermission(ctx context.Context, roleID, permissionID uuid.UUID) (bool, error)

	// ListRolesByPermission returns all roles that have been assigned the given permission,
	// ordered by name ascending.
	// Returns an empty slice if no roles have the permission.
	ListRolesByPermission(ctx context.Context, permissionID uuid.UUID) ([]Role, error)
}

// rolePermissionRepository is the concrete implementation of RolePermissionRepository.
// It wraps sqlc-generated queries and translates between database and domain types.
// Instantiated exclusively by NewStore — never directly.
type rolePermissionRepository struct {
	q *generated.Queries
}

//--------------------------
// Concrete Implementations
// -------------------------

func (r *rolePermissionRepository) Assign(ctx context.Context, roleID, permissionID uuid.UUID) error {
	err := r.q.AssignPermissionToRole(ctx, &generated.AssignPermissionToRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})

	if err != nil {
		return fmt.Errorf("RolePermissionRepository.Assign: %w", mapError(err))
	}

	return nil
}

func (r *rolePermissionRepository) Remove(ctx context.Context, roleID, permissionID uuid.UUID) error {
	err := r.q.RemovePermissionFromRole(ctx, &generated.RemovePermissionFromRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})

	if err != nil {
		return fmt.Errorf("RolePermissionRepository.Remove: %w", mapError(err))
	}

	return nil
}

func (r *rolePermissionRepository) RoleHasPermission(ctx context.Context, roleID, permissionID uuid.UUID) (bool, error) {
	res, err := r.q.RoleHasPermission(ctx, &generated.RoleHasPermissionParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})

	if err != nil {
		return false, fmt.Errorf("RolePermissionRepository.RoleHasPermission: %w", mapError(err))
	}

	return res, nil
}

func (r *rolePermissionRepository) ListRolesByPermission(ctx context.Context, permissionID uuid.UUID) ([]Role, error) {
	rows, err := r.q.ListRolesByPermission(ctx, permissionID)
	if err != nil {
		return nil, fmt.Errorf("RolePermissionRepository.ListRolesByPermission: %w", mapError(err))
	}

	roles := make([]Role, len(rows))
	for i, row := range rows {
		roles[i] = toRole(row)
	}

	return roles, nil
}
