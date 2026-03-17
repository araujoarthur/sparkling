// roles_repository.go implements RoleRepository, providing data access
// for IAM roles backed by PostgreSQL via sqlc-generated queries.
package repository

import (
	"context"
	"fmt"

	"github.com/araujoarthur/intranetbackend/services/iam/internal/repository/sqlc/generated"
	"github.com/google/uuid"
)

// RoleRepository defines the data access contract for IAM roles.
// Consumers should depend on this interface, never on the concrete implementation.
type RoleRepository interface {
	// GetByID retrieves a role by its internal UUID.
	// Returns ErrNotFound if no role exists with the given ID.
	GetByID(ctx context.Context, id uuid.UUID) (Role, error)

	// GetByName retrieves a role by its unique name.
	// Returns ErrNotFound if no role exists with the given name.
	GetByName(ctx context.Context, name string) (Role, error)

	// List returns all roles ordered by name ascending.
	List(ctx context.Context) ([]Role, error)

	// Create inserts a new role and returns it.
	// isSystem marks the role as built-in, preventing mutation or deletion.
	Create(ctx context.Context, name, description string, isSystem bool) (Role, error)

	// Update modifies the name and description of a non-system role.
	// Returns ErrNotFound if the role does not exist.
	// System roles cannot be updated and the query will return ErrNotFound for them.
	Update(ctx context.Context, id uuid.UUID, name, description string) (Role, error)

	// Delete removes a non-system role and cascades to all its assignments.
	// Returns ErrNotFound if the role does not exist.
	// System roles cannot be deleted and the query will return ErrNotFound for them.
	Delete(ctx context.Context, id uuid.UUID) error
}

// roleRepository is the concrete implementation of RoleRepository.
// It wraps sqlc-generated queries and translates between database and domain types.
// Instantiated exclusively by NewStore — never directly.
type roleRepository struct {
	q *generated.Queries
}


//--------------------------
// Concrete Implementations
// -------------------------

// GetByID retrieves a role by its internal UUID.
// Returns ErrNotFound if no role exists with the given ID.
func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (Role, error) {
	row, err := r.q.GetRoleByID(ctx, id)
	if err != nil {
		return Role{}, fmt.Errorf("RoleRepository.GetByID: %w", mapError(err))
	}

	return toRole(row), nil
}

// GetByName retrieves a role by its unique name.
// Returns ErrNotFound if no role exists with the given name.
func (r *roleRepository) GetByName(ctx context.Context, name string) (Role, error) {
	row, err := r.q.GetRoleByName(ctx, name)
	if err != nil {
		return Role{}, fmt.Errorf("RoleRepository.GetByName: %w", mapError(err))
	}

	return toRole(row), nil
}

// List returns all roles ordered by name ascending.
// Returns an empty slice if no roles exist.
func (r *roleRepository) List(ctx context.Context) ([]Role, error) {
	rows, err := r.q.ListRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("RoleRepository.List: %w", mapError(err))
	}

	roles := make([]Role, len(rows))

	for i, row := range rows {
		roles[i] = toRole(row)
	}

	return roles, nil
}

// Create inserts a new role and returns it.
// isSystem marks the role as built-in, preventing mutation or deletion.
// Returns ErrConflict if a role with the same name already exists.
func (r *roleRepository) Create(ctx context.Context, name, description string, isSystem bool) (Role, error) {
	row, err := r.q.CreateRole(ctx, &generated.CreateRoleParams{
		Name:        name,
		Description: pgxText(description),
		IsSystem:    isSystem,
	})

	if err != nil {
		return Role{}, fmt.Errorf("RoleRepository.Create: %w", mapError(err))
	}

	return toRole(row), nil
}

// Update modifies the name and description of a non-system role.
// Returns ErrNotFound if the role does not exist.
// System roles cannot be updated and the query will return ErrNotFound for them.
func (r *roleRepository) Update(ctx context.Context, id uuid.UUID, name, description string) (Role, error) {
	row, err := r.q.UpdateRole(ctx, &generated.UpdateRoleParams{
		ID:          id,
		Name:        name,
		Description: pgxText(description),
	})

	if err != nil {
		return Role{}, fmt.Errorf("RoleRepository.Update: %w", mapError(err))
	}

	return toRole(row), nil
}

// Delete removes a non-system role and cascades to all its assignments.
// Returns ErrNotFound if the role does not exist.
// System roles cannot be deleted and the query will return ErrNotFound for them.
func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteRole(ctx, id); err != nil {
		return fmt.Errorf("RoleRepository.Delete: %w", mapError(err))
	}

	return nil
}
