package domain

import (
	"context"
	"fmt"

	"github.com/araujoarthur/intranetbackend/services/iam/internal/repository"
	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/google/uuid"
)

// PrincipalRoleService defines the business logic contract for managing
// the assignment of roles to principals.
// Consumers should depend on this interface, never on the concrete implementation.
type PrincipalRoleService interface {
	// Assign grants a role to a principal.
	// Requires iam:role-{rolename}:grant permission.
	// Returns ErrForbidden if the caller lacks the specific grant permission.
	// Returns ErrNotFound if the principal, role, or granter does not exist.
	// Returns ErrConflict if the principal already has the role.
	Assign(ctx context.Context, callerID uuid.UUID, principalID, roleID uuid.UUID) error

	// Remove revokes a role from a principal.
	// A principal may revoke their own roles freely.
	// Revoking another principal's role requires iam:role-{rolename}:grant.
	// Returns ErrForbidden if the caller lacks permission.
	// Returns ErrNotFound if the assignment does not exist.
	Remove(ctx context.Context, callerID uuid.UUID, principalID, roleID uuid.UUID) error

	// ListRolesByPrincipal returns all roles assigned to the given principal,
	// ordered by name ascending.
	// Returns an empty slice if the principal has no roles.
	ListRolesByPrincipal(ctx context.Context, principalID uuid.UUID) ([]repository.Role, error)

	// ListPrincipalsByRole returns all active principals assigned the given role,
	// ordered by creation date ascending.
	// Returns an empty slice if no active principals have the role.
	ListPrincipalsByRole(ctx context.Context, roleID uuid.UUID) ([]repository.Principal, error)

	// PrincipalHasRole reports whether a principal has been assigned a specific role.
	PrincipalHasRole(ctx context.Context, principalID, roleID uuid.UUID) (bool, error)
}

type principalRoleService struct {
	store *repository.Store
}

// Constructor

func NewPrincipalRoleService(store *repository.Store) PrincipalRoleService {
	return &principalRoleService{store: store}
}

// Concrete Implementations

func (s *principalRoleService) Assign(ctx context.Context, callerID uuid.UUID, principalID, roleID uuid.UUID) error {
	role, err := s.store.Roles.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("PrincipalRoleService.Assign: %w", err)
	}

	allowed, err := hasPermission(ctx, s.store, callerID, GetGrantRolePermission(role.Name))
	if err != nil {
		return fmt.Errorf("PrincipalRoleService.Assign: %w", err)
	}

	if !allowed {
		return apierror.ErrForbidden
	}

	_, err = s.store.PrincipalRoles.Assign(ctx, principalID, roleID, callerID)
	if err != nil {
		return fmt.Errorf("PrincipalRoleService.Assign: %w", err)
	}

	return nil
}

func (s *principalRoleService) Remove(ctx context.Context, callerID uuid.UUID, principalID, roleID uuid.UUID) error {
	// Self-revokation rule: An user can revoke its own permissions.
	if callerID != principalID {
		role, err := s.store.Roles.GetByID(ctx, roleID)
		if err != nil {
			return fmt.Errorf("PrincipalRoleService.Remove: %w", err)
		}

		allowed, err := hasPermission(ctx, s.store, callerID, GetGrantRolePermission(role.Name))
		if err != nil {
			return fmt.Errorf("PrincipalRoleService.Remove: %w", err)
		}

		if !allowed {
			return apierror.ErrForbidden
		}
	}

	if err := s.store.PrincipalRoles.Remove(ctx, principalID, roleID); err != nil {
		return fmt.Errorf("PrincipalRoleService.Remove: %w", err)
	}

	return nil
}

func (s *principalRoleService) ListRolesByPrincipal(ctx context.Context, principalID uuid.UUID) ([]repository.Role, error) {
	list, err := s.store.PrincipalRoles.ListRolesByPrincipal(ctx, principalID)
	if err != nil {
		return nil, fmt.Errorf("PrincipalRoleService.ListRolesByPrincipal: %w", err)
	}

	return list, nil
}

func (s *principalRoleService) ListPrincipalsByRole(ctx context.Context, roleID uuid.UUID) ([]repository.Principal, error) {
	list, err := s.store.PrincipalRoles.ListPrincipalsByRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("PrincipalRoleService.ListPrincipalsByRole: %w", err)
	}

	return list, nil
}

func (s *principalRoleService) PrincipalHasRole(ctx context.Context, principalID, roleID uuid.UUID) (bool, error) {
	has, err := s.store.PrincipalRoles.PrincipalHasRole(ctx, principalID, roleID)
	if err != nil {
		return false, fmt.Errorf("PrincipalRoleService.PrincipalHasRole: %w", err)
	}

	return has, nil
}
