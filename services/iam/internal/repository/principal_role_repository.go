// principal_roles_repository.go implements PrincipalRoleRepository, providing data access
// for IAM principal role assignments backed by PostgreSQL via sqlc-generated queries.
package repository

import (
	"context"
	"fmt"

	"github.com/araujoarthur/intranetbackend/services/iam/internal/repository/sqlc/generated"
	"github.com/google/uuid"
)

// PrincipalRoleRepository defines the data access contract for managing
// the assignment of roles to principals.
// Consumers should depend on this interface, never on the concrete implementation.
type PrincipalRoleRepository interface {
	// Assign grants a role to a principal, recording who made the assignment.
	// grantedBy must be the internal IAM UUID of the principal performing the action.
	// Returns ErrConflict if the principal already has the role.
	// Returns ErrNotFound if the principal, role, or granter does not exist.
	Assign(ctx context.Context, principalID, roleID, grantedBy uuid.UUID) (PrincipalRole, error)

	// Remove revokes a role from a principal.
	// Returns ErrNotFound if the assignment does not exist.
	Remove(ctx context.Context, principalID, roleID uuid.UUID) error

	// ListRolesByPrincipal returns all roles assigned to the given principal,
	// ordered by name ascending.
	// Returns an empty slice if the principal has no roles.
	ListRolesByPrincipal(ctx context.Context, principalID uuid.UUID) ([]Role, error)

	// ListPrincipalsByRole returns all active principals assigned the given role,
	// ordered by creation date ascending.
	// Returns an empty slice if no active principals have the role.
	ListPrincipalsByRole(ctx context.Context, roleID uuid.UUID) ([]Principal, error)

	// PrincipalHasRole reports whether a principal has been assigned a specific role.
	PrincipalHasRole(ctx context.Context, principalID, roleID uuid.UUID) (bool, error)
}

// principalRoleRepository is the concrete implementation of PrincipalRoleRepository.
// It wraps sqlc-generated queries and translates between database and domain types.
// Instantiated exclusively by NewStore — never directly.
type principalRoleRepository struct {
	q *generated.Queries
}

func (r *principalRoleRepository) Assign(ctx context.Context, principalID, roleID, grantedBy uuid.UUID) (PrincipalRole, error) {
	row, err := r.q.AssignRoleToPrincipal(ctx, &generated.AssignRoleToPrincipalParams{
		PrincipalID: principalID,
		RoleID:      roleID,
		GrantedBy:   grantedBy,
	})

	if err != nil {
		return PrincipalRole{}, fmt.Errorf("PrincipalRoleRepository.Assign: %w", mapError(err))
	}

	return toPrincipalRole(row), nil
}

func (r *principalRoleRepository) Remove(ctx context.Context, principalID, roleID uuid.UUID) error {
	err := r.q.RemoveRoleFromPrincipal(ctx, &generated.RemoveRoleFromPrincipalParams{
		PrincipalID: principalID,
		RoleID:      roleID,
	})

	if err != nil {
		return fmt.Errorf("PrincipalRoleRepository.Remove: %w", mapError(err))
	}

	return nil
}

func (r *principalRoleRepository) ListRolesByPrincipal(ctx context.Context, principalID uuid.UUID) ([]Role, error) {
	rows, err := r.q.ListRolesByPrincipal(ctx, principalID)

	if err != nil {
		return nil, fmt.Errorf("PrincipalRoleRepository.ListRolesByPrincipal: %w", mapError(err))
	}

	roles := make([]Role, len(rows))

	for i, row := range rows {
		roles[i] = toRole(row)
	}

	return roles, nil
}

func (r *principalRoleRepository) ListPrincipalsByRole(ctx context.Context, roleID uuid.UUID) ([]Principal, error) {
	rows, err := r.q.ListPrincipalsByRole(ctx, roleID)

	if err != nil {
		return nil, fmt.Errorf("PrincipalRoleRepository.ListPrincipalsByRole: %w", mapError(err))
	}

	principals := make([]Principal, len(rows))

	for i, row := range rows {
		principals[i] = toPrincipal(row)
	}

	return principals, nil
}

func (r *principalRoleRepository) PrincipalHasRole(ctx context.Context, principalID, roleID uuid.UUID) (bool, error) {
	res, err := r.q.PrincipalHasRole(ctx, &generated.PrincipalHasRoleParams{
		PrincipalID: principalID,
		RoleID:      roleID,
	})

	if err != nil {
		return false, fmt.Errorf("PrincipalRoleRepository.PrincipalHasRole: %w", mapError(err))
	}

	return res, nil
}
