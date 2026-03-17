// permissions.go implements PermissionService, providing business logic
// for IAM permission management backed by the repository layer.
package domain

import (
	"context"
	"fmt"

	"github.com/araujoarthur/intranetbackend/services/iam/internal/repository"
	"github.com/google/uuid"
)

// PermissionService defines the business logic contract for IAM permissions.
// It enforces permission name format validation and access control on write operations.
// Consumers should depend on this interface, never on the concrete implementation.
type PermissionService interface {
	// GetByID retrieves a permission by its internal UUID.
	// Returns ErrNotFound if no permission exists with the given ID.
	GetByID(ctx context.Context, id uuid.UUID) (repository.Permission, error)

	// GetByName retrieves a permission by its unique name.
	// Returns ErrNotFound if no permission exists with the given name.
	GetByName(ctx context.Context, name string) (repository.Permission, error)

	// List returns all permissions ordered by name ascending.
	// Returns an empty slice if no permissions exist.
	List(ctx context.Context) ([]repository.Permission, error)

	// ListByRole returns all permissions assigned to the given role,
	// ordered by name ascending.
	// Returns an empty slice if the role exists but has no permissions.
	// Returns ErrNotFound if the role does not exist.
	ListByRole(ctx context.Context, roleID uuid.UUID) ([]repository.Permission, error)

	// Create validates and inserts a new permission.
	// Permission names must follow the scope:resource:action format.
	// Requires iam:permissions:write permission.
	// Returns ErrForbidden if the caller lacks permission.
	// Returns ErrInvalidArgument if the name format is invalid.
	// Returns ErrConflict if a permission with the same name already exists.
	Create(ctx context.Context, callerID uuid.UUID, name, description string) (repository.Permission, error)

	// Delete removes a permission and cascades to all role assignments.
	// Requires iam:permissions:delete permission.
	// Returns ErrForbidden if the caller lacks permission.
	// Returns ErrNotFound if the permission does not exist.
	Delete(ctx context.Context, callerID uuid.UUID, id uuid.UUID) error
}

type permissionService struct {
	store *repository.Store
}

func NewPermissionService(store *repository.Store) PermissionService {
	return &permissionService{store: store}
}

func (s *permissionService) Create(ctx context.Context, callerID uuid.UUID, name, description string) (repository.Permission, error) {
	allowed, err := hasPermission(ctx, s.store, callerID, permissionIAMPermissionsWrite)
	if err != nil {
		return repository.Permission{}, fmt.Errorf("PermissionService.Create: %w", err)
	}

	if !allowed {
		return repository.Permission{}, repository.ErrForbidden
	}

	if err := validatePermissionName(name); err != nil {
		return repository.Permission{}, fmt.Errorf("PermissionService.Create: %w", err)
	}

	created, err := s.store.Permissions.Create(ctx, name, description)
	if err != nil {
		return repository.Permission{}, fmt.Errorf("PermissionService.Create: %w", err)
	}

	return created, nil
}

func (s *permissionService) Delete(ctx context.Context, callerID uuid.UUID, id uuid.UUID) error {
	allowed, err := hasPermission(ctx, s.store, callerID, permissionIAMPermissionsDelete)
	if err != nil {
		return fmt.Errorf("PermissionService.Delete: %w", err)
	}

	if !allowed {
		return repository.ErrForbidden
	}

	err = s.store.Permissions.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("PermissionService.Delete: %w", err)
	}

	return nil
}

func (s *permissionService) GetByID(ctx context.Context, id uuid.UUID) (repository.Permission, error) {
	perm, err := s.store.Permissions.GetByID(ctx, id)
	if err != nil {
		return repository.Permission{}, fmt.Errorf("PermissionService.GetByID: %w", err)
	}

	return perm, nil
}

func (s *permissionService) GetByName(ctx context.Context, name string) (repository.Permission, error) {
	perm, err := s.store.Permissions.GetByName(ctx, name)
	if err != nil {
		return repository.Permission{}, fmt.Errorf("PermissionService.GetByName: %w", err)
	}

	return perm, nil
}

func (s *permissionService) List(ctx context.Context) ([]repository.Permission, error) {
	perms, err := s.store.Permissions.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("PermissionService.List: %w", err)
	}

	return perms, nil
}

func (s *permissionService) ListByRole(ctx context.Context, roleID uuid.UUID) ([]repository.Permission, error) {
	perms, err := s.store.Permissions.ListByRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("PermissionService.ListByRole: %w", err)
	}

	return perms, nil
}
