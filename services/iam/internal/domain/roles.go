package domain

import (
	"context"
	"fmt"

	"github.com/araujoarthur/intranetbackend/services/iam/internal/repository"
	"github.com/google/uuid"
)

type RoleService interface {
	// GetByID retrieves a role from its internal UUID
	GetByID(ctx context.Context, id uuid.UUID) (repository.Role, error)

	// GetByName retrieves a role from its internal UUID
	GetByName(ctx context.Context, name string) (repository.Role, error)

	// List returns all roles
	List(ctx context.Context) ([]repository.Role, error)

	// Create validates and creates a new role, and automatically creates
	// its grant permission. Requires iam:roles:write.
	Create(ctx context.Context, callerID uuid.UUID, name, description string) (repository.Role, error)

	// Update modifies a non-system role. Requires iam:roles:write.
	Update(ctx context.Context, callerID uuid.UUID, id uuid.UUID, name, description string) (repository.Role, error)

	// Delete removes a non-system role. Requires iam:roles:delete.
	Delete(ctx context.Context, callerID uuid.UUID, id uuid.UUID) error
}

type roleService struct {
	store *repository.Store
}

func NewRoleService(store *repository.Store) RoleService {
	return &roleService{store: store}
}

// Create validates and creates a new role, and automatically creates its grant
// permission (iam:role-{name}:grant) in the same transaction.
// Requires iam:roles:write permission.
// Returns ErrForbidden if the caller lacks permission.
// Returns ErrInvalidArgument if the name format is invalid.
func (s *roleService) Create(ctx context.Context, callerID uuid.UUID, name, description string) (repository.Role, error) {
	allowed, err := hasPermission(ctx, s.store, callerID, permissionIAMRolesWrite)
	if err != nil {
		return repository.Role{}, fmt.Errorf("RoleService.Create: %w", err)
	}

	if !allowed {
		return repository.Role{}, repository.ErrForbidden
	}

	if err := validateRoleName(name); err != nil {
		return repository.Role{}, fmt.Errorf("RoleService.Create: %w", err)
	}

	var role repository.Role

	err = s.store.WithTx(ctx, func(tx *repository.Store) error {
		// creates the role
		created, err := tx.Roles.Create(ctx, name, description, false)
		if err != nil {
			return fmt.Errorf("creating role: %w", err)
		}

		// creates the grant permissions
		_, err = tx.Permissions.Create(ctx, fmt.Sprintf("iam:role-%s:grant", name), fmt.Sprintf("Grants the %s role to a principal", name))
		if err != nil {
			return fmt.Errorf("creating grant permission: %w", err)
		}

		// copies the role to the outter variable
		role = created

		return nil
	})

	if err != nil {
		return repository.Role{}, fmt.Errorf("RoleService.Create: %w", err)
	}

	return role, nil
}

// Delete removes a non-system role and its automatically created grant permission
// in the same transaction.
// Requires iam:roles:delete permission.
// Returns ErrForbidden if the caller lacks permission.
// Returns ErrNotFound if the role does not exist.
func (s *roleService) Delete(ctx context.Context, callerID uuid.UUID, id uuid.UUID) error {
	allowed, err := hasPermission(ctx, s.store, callerID, permissionIAMRolesDelete)
	if err != nil {
		return fmt.Errorf("RoleService.Delete: %w", err)
	}

	if !allowed {
		return repository.ErrForbidden
	}

	// fetch role first to get its name for the grant permission lookup
	role, err := s.store.Roles.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("RoleService.Delete: %w", err)
	}

	err = s.store.WithTx(ctx, func(tx *repository.Store) error {
		grantPermission, err := tx.Permissions.GetByName(ctx, fmt.Sprintf("iam:role-%s:grant", role.Name))
		if err != nil {
			return fmt.Errorf("fetching grant permission: %w", err)
		}

		if err := tx.Permissions.Delete(ctx, grantPermission.ID); err != nil {
			return fmt.Errorf("deleting grant permission: %w", err)
		}

		if err := tx.Roles.Delete(ctx, id); err != nil {
			return fmt.Errorf("deleting role: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("RoleService.Delete: %w", err)
	}

	return nil
}

// Update modifies the name and description of a non-system role.
// Requires iam:roles:write permission.
// Returns ErrForbidden if the caller lacks permission.
// Returns ErrNotFound if the role does not exist.
// Returns ErrInvalidArgument if the name format is invalid.
func (s *roleService) Update(ctx context.Context, callerID uuid.UUID, id uuid.UUID, name, description string) (repository.Role, error) {
	allowed, err := hasPermission(ctx, s.store, callerID, permissionIAMRolesWrite)
	if err != nil {
		return repository.Role{}, fmt.Errorf("RoleService.Update: %w", err)
	}

	if !allowed {
		return repository.Role{}, repository.ErrForbidden
	}

	err = validateRoleName(name)
	if err != nil {
		return repository.Role{}, fmt.Errorf("RoleService.Update: %w", err)
	}

	role, err := s.store.Roles.Update(ctx, id, name, description)
	if err != nil {
		return repository.Role{}, fmt.Errorf("RoleService.Update: %w", err)
	}

	return role, nil
}

// GetByID retrieves a role by its internal UUID.
// Returns ErrNotFound if no role exists with the given ID.
func (s *roleService) GetByID(ctx context.Context, id uuid.UUID) (repository.Role, error) {
	role, err := s.store.Roles.GetByID(ctx, id)
	if err != nil {
		return repository.Role{}, fmt.Errorf("RoleService.GetByID: %w", err)
	}

	return role, nil
}

// GetByName retrieves a role by its unique name.
// Returns ErrNotFound if no role exists with the given name.
func (s *roleService) GetByName(ctx context.Context, name string) (repository.Role, error) {
	role, err := s.store.Roles.GetByName(ctx, name)
	if err != nil {
		return repository.Role{}, fmt.Errorf("RoleService.GetByName: %w", err)
	}

	return role, nil
}

// List returns all roles ordered by name ascending.
// Returns an empty slice if no roles exist.
func (s *roleService) List(ctx context.Context) ([]repository.Role, error) {
	roles, err := s.store.Roles.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("RoleService.List: %w", err)
	}

	return roles, nil
}
