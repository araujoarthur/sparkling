// permissions_repository.go implements PermissionRepository, providing data access
// for IAM permissions backed by PostgreSQL via sqlc-generated queries.
package repository

import (
	"context"
	"fmt"

	"github.com/araujoarthur/intranetbackend/services/iam/internal/repository/sqlc/generated"
	"github.com/araujoarthur/intranetbackend/shared/pkg/helpers"
	"github.com/google/uuid"
)

// PermissionRepository defines the data access contract for IAM permissions.
// Consumers should depend on this interface, never on the concrete implementation.
type PermissionRepository interface {
	// GetByID retrieves a permission by its internal UUID.
	// Returns ErrNotFound if no permission exists with the given ID.
	GetByID(ctx context.Context, id uuid.UUID) (Permission, error)

	// GetByName retrieves a permission by its unique name.
	// Returns ErrNotFound if no permission exists with the given name.
	GetByName(ctx context.Context, name string) (Permission, error)

	// List returns all permissions ordered by name ascending.
	List(ctx context.Context) ([]Permission, error)

	// ListByRole returns all permissions assigned to the given role,
	// ordered by name ascending.
	// Returns an empty slice if the role exists but has no permissions.
	// Returns ErrNotFound if the role does not exist.
	ListByRole(ctx context.Context, roleID uuid.UUID) ([]Permission, error)

	// Create inserts a new permission and returns it.
	// Permission names should follow the resource:action convention (e.g. users:read).
	// Returns ErrConflict if a permission with the same name already exists.
	Create(ctx context.Context, name, description string) (Permission, error)

	// Delete removes a permission and cascades to all role assignments.
	// Returns ErrNotFound if the permission does not exist.
	Delete(ctx context.Context, id uuid.UUID) error
}

// permissionRepository is the concrete implementation of PermissionRepository.
// It wraps sqlc-generated queries and translates between database and domain types.
// Instantiated exclusively by NewStore — never directly.
type permissionRepository struct {
	q *generated.Queries
}

// GetByID retrieves a permission by its internal UUID.
// Returns ErrNotFound if no permission exists with the given ID.
func (r *permissionRepository) GetByID(ctx context.Context, id uuid.UUID) (Permission, error) {
	row, err := r.q.GetPermissionByID(ctx, id)
	if err != nil {
		return Permission{}, fmt.Errorf("PermissionRepository.GetByID: %w", helpers.MapError(err))
	}

	return toPermission(row), nil
}

// GetByName retrieves a permission by its unique name.
// Returns ErrNotFound if no permission exists with the given name.
func (r *permissionRepository) GetByName(ctx context.Context, name string) (Permission, error) {
	row, err := r.q.GetPermissionByName(ctx, name)
	if err != nil {
		return Permission{}, fmt.Errorf("PermissionRepository.GetByName: %w", helpers.MapError(err))
	}

	return toPermission(row), nil
}

// List returns all permissions ordered by name ascending.
// Returns an empty slice if no permissions exist.
func (r *permissionRepository) List(ctx context.Context) ([]Permission, error) {
	rows, err := r.q.ListPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("PermissionRepository.List: %w", helpers.MapError(err))
	}

	permissions := make([]Permission, len(rows))

	for i, row := range rows {
		permissions[i] = toPermission(row)
	}

	return permissions, nil
}

// ListByRole returns all permissions assigned to the given role,
// ordered by name ascending.
// Returns an empty slice if the role exists but has no permissions.
// Returns ErrNotFound if the role does not exist.
func (r *permissionRepository) ListByRole(ctx context.Context, roleID uuid.UUID) ([]Permission, error) {
	rows, err := r.q.ListPermissionsByRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("PermissionRepository.ListByRole: %w", helpers.MapError(err))
	}

	permissions := make([]Permission, len(rows))

	for i, row := range rows {
		permissions[i] = toPermission(row)
	}

	return permissions, nil
}

// Create inserts a new permission and returns it.
// Permission names should follow the resource:action convention (e.g. users:read).
// Returns ErrConflict if a permission with the same name already exists.
func (r *permissionRepository) Create(ctx context.Context, name, description string) (Permission, error) {
	row, err := r.q.CreatePermission(ctx, &generated.CreatePermissionParams{
		Name:        name,
		Description: helpers.PgxText(description),
	})

	if err != nil {
		return Permission{}, fmt.Errorf("PermissionRepository.Create: %w", helpers.MapError(err))
	}

	return toPermission(row), nil
}

// Delete removes a permission and cascades to all role assignments.
// Returns ErrNotFound if the permission does not exist.
func (r *permissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeletePermission(ctx, id); err != nil {
		return fmt.Errorf("PermissionRepository.Delete: %w", helpers.MapError(err))
	}

	return nil
}
