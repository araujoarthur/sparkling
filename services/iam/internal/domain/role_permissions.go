package domain

import (
	"context"
	"fmt"

	"github.com/araujoarthur/intranetbackend/services/iam/internal/repository"
	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/google/uuid"
)

// RolePermissionService defines the business logic contract for managing
// the assignment of permissions to roles.
// Consumers should depend on this interface, never on the concrete implementation.
type RolePermissionService interface {
	// Assign adds a permission to a role.
	// Requires iam:roles:write permission.
	// Returns ErrForbidden if the caller lacks permission.
	// Returns ErrNotFound if the role or permission does not exist.
	Assign(ctx context.Context, callerID uuid.UUID, roleID, permissionID uuid.UUID) error

	// Remove revokes a permission from a role.
	// Requires iam:roles:write permission.
	// Returns ErrForbidden if the caller lacks permission.
	// Returns ErrNotFound if the assignment does not exist.
	Remove(ctx context.Context, callerID uuid.UUID, roleID, permissionID uuid.UUID) error

	// RoleHasPermission reports whether a role has been assigned a specific permission.
	RoleHasPermission(ctx context.Context, roleID, permissionID uuid.UUID) (bool, error)

	// ListRolesByPermission returns all roles that have been assigned the given permission,
	// ordered by name ascending.
	// Returns an empty slice if no roles have the permission.
	ListRolesByPermission(ctx context.Context, permissionID uuid.UUID) ([]repository.Role, error)
}

type rolePermissionService struct {
	store *repository.Store
}

func NewRolePermissionService(store *repository.Store) RolePermissionService {
	return &rolePermissionService{store: store}
}

func (s *rolePermissionService) Assign(ctx context.Context, callerID uuid.UUID, roleID, permissionID uuid.UUID) error {
	allowed, err := hasPermission(ctx, s.store, callerID, permissionIAMRolePermissionsAssign)
	if err != nil {
		return fmt.Errorf("RolePermissionService.Assign: %w", err)
	}

	if !allowed {
		return apierror.ErrForbidden
	}

	if err := s.store.RolePermissions.Assign(ctx, roleID, permissionID); err != nil {
		return fmt.Errorf("RolePermissionService.Assign: %w", err)
	}

	return nil
}

func (s *rolePermissionService) Remove(ctx context.Context, callerID uuid.UUID, roleID, permissionID uuid.UUID) error {
	allowed, err := hasPermission(ctx, s.store, callerID, permissionIAMRolePermissionsRevoke)
	if err != nil {
		return fmt.Errorf("RolePermissionService.Remove: %w", err)
	}

	if !allowed {
		return apierror.ErrForbidden
	}

	if err := s.store.RolePermissions.Remove(ctx, roleID, permissionID); err != nil {
		return fmt.Errorf("RolePermissionService.Remove: %w", err)
	}

	return nil
}

func (s *rolePermissionService) RoleHasPermission(ctx context.Context, roleID, permissionID uuid.UUID) (bool, error) {
	res, err := s.store.RolePermissions.RoleHasPermission(ctx, roleID, permissionID)
	if err != nil {
		return false, fmt.Errorf("RolePermissionService.RoleHasPermission: %w", err)
	}

	return res, nil
}

func (s *rolePermissionService) ListRolesByPermission(ctx context.Context, permissionID uuid.UUID) ([]repository.Role, error) {
	roles, err := s.store.RolePermissions.ListRolesByPermission(ctx, permissionID)
	if err != nil {
		return nil, fmt.Errorf("RolePermissionService.ListRolesByPermission: %w", err)
	}

	return roles, nil
}
