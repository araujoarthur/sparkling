package domain

import (
	"context"
	"fmt"

	"github.com/araujoarthur/intranetbackend/services/iam/internal/repository"
	"github.com/google/uuid"
)

// PrincipalService defines the business logic contract for IAM principals.
// Consumers should depend on this interface, never on the concrete implementation.
type PrincipalService interface {
	// GetByID retrieves a principal by its internal IAM UUID.
	// Returns ErrNotFound if no principal exists with the given ID.
	GetByID(ctx context.Context, id uuid.UUID) (repository.Principal, error)

	// GetByExternalID retrieves a principal by the ID issued by the auth service.
	// Returns ErrNotFound if no matching principal exists.
	GetByExternalID(ctx context.Context, externalID uuid.UUID, principalType repository.PrincipalType) (repository.Principal, error)

	// List returns all principals ordered by creation date ascending.
	List(ctx context.Context) ([]repository.Principal, error)

	// ListByType returns all principals of the given type ordered by creation date ascending.
	ListByType(ctx context.Context, principalType repository.PrincipalType) ([]repository.Principal, error)

	// Create registers a new principal in IAM for the given external identity.
	// Called automatically when a user registers in the auth service.
	// Returns ErrConflict if a principal with the same ExternalID and PrincipalType already exists.
	Create(ctx context.Context, externalID uuid.UUID, principalType repository.PrincipalType) (repository.Principal, error)

	// Activate marks a principal as active.
	// Requires iam:principals:write permission.
	// Returns ErrForbidden if the caller lacks permission.
	// Returns ErrNotFound if no principal exists with the given ID.
	Activate(ctx context.Context, callerID uuid.UUID, id uuid.UUID) (repository.Principal, error)

	// Deactivate marks a principal as inactive.
	// Requires iam:principals:write permission.
	// Returns ErrForbidden if the caller lacks permission.
	// Returns ErrNotFound if no principal exists with the given ID.
	Deactivate(ctx context.Context, callerID uuid.UUID, id uuid.UUID) (repository.Principal, error)

	// Delete permanently removes a principal and cascades to all its role assignments.
	// Requires iam:principals:delete permission.
	// Returns ErrForbidden if the caller lacks permission.
	// Returns ErrNotFound if no principal exists with the given ID.
	Delete(ctx context.Context, callerID uuid.UUID, id uuid.UUID) error

	// GetPermissions returns the full flat list of permissions for a principal,
	// resolved across all assigned roles.
	GetPermissions(ctx context.Context, id uuid.UUID) ([]repository.Permission, error)
}

type principalService struct {
	store *repository.Store
}

func NewPrincipalService(store *repository.Store) PrincipalService {
	return &principalService{store: store}
}

func (s *principalService) Create(ctx context.Context, externalID uuid.UUID, principalType repository.PrincipalType) (repository.Principal, error) {
	// has no caller ID because it's designed to be called by auth when the credential is created.
	principal, err := s.store.Principals.Create(ctx, externalID, principalType)
	if err != nil {
		return repository.Principal{}, fmt.Errorf("PrincipalService.Create: %w", err)
	}

	return principal, nil
}

func (s *principalService) Delete(ctx context.Context, callerID uuid.UUID, id uuid.UUID) error {
	allowed, err := hasPermission(ctx, s.store, callerID, permissionIAMPrincipalsDelete)
	if err != nil {
		return fmt.Errorf("PrincipalService.Delete: %w", err)
	}

	if !allowed {
		return repository.ErrForbidden
	}

	if err := s.store.Principals.Delete(ctx, id); err != nil {
		return fmt.Errorf("PrincipalService.Delete: %w", err)
	}

	return nil
}

func (s *principalService) Activate(ctx context.Context, callerID uuid.UUID, id uuid.UUID) (repository.Principal, error) {
	allowed, err := hasPermission(ctx, s.store, callerID, permissionIAMPrincipalsWrite)
	if err != nil {
		return repository.Principal{}, fmt.Errorf("PrincipalService.Activate: %w", err)
	}

	if !allowed {
		return repository.Principal{}, repository.ErrForbidden
	}

	principal, err := s.store.Principals.Activate(ctx, id)
	if err != nil {
		return repository.Principal{}, fmt.Errorf("PrincipalService.Activate: %w", err)
	}

	return principal, nil
}

func (s *principalService) Deactivate(ctx context.Context, callerID uuid.UUID, id uuid.UUID) (repository.Principal, error) {
	allowed, err := hasPermission(ctx, s.store, callerID, permissionIAMPrincipalsWrite)
	if err != nil {
		return repository.Principal{}, fmt.Errorf("PrincipalService.Deactivate: %w", err)
	}

	if !allowed {
		return repository.Principal{}, repository.ErrForbidden
	}

	principal, err := s.store.Principals.Deactivate(ctx, id)
	if err != nil {
		return repository.Principal{}, fmt.Errorf("PrincipalService.Deactivate: %w", err)
	}

	return principal, nil
}

func (s *principalService) GetByID(ctx context.Context, id uuid.UUID) (repository.Principal, error) {
	principal, err := s.store.Principals.GetByID(ctx, id)
	if err != nil {
		return repository.Principal{}, fmt.Errorf("PrincipalService.GetByID: %w", err)
	}

	return principal, nil
}

func (s *principalService) GetByExternalID(ctx context.Context, externalID uuid.UUID, principalType repository.PrincipalType) (repository.Principal, error) {
	principal, err := s.store.Principals.GetByExternalID(ctx, externalID, principalType)
	if err != nil {
		return repository.Principal{}, fmt.Errorf("PrincipalService.GetByExternalID: %w", err)
	}

	return principal, nil
}

func (s *principalService) List(ctx context.Context) ([]repository.Principal, error) {
	list, err := s.store.Principals.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("PrincipalService.List: %w", err)
	}

	return list, nil
}

func (s *principalService) ListByType(ctx context.Context, principalType repository.PrincipalType) ([]repository.Principal, error) {
	list, err := s.store.Principals.ListByType(ctx, principalType)
	if err != nil {
		return nil, fmt.Errorf("PrincipalService.ListByType: %w", err)
	}

	return list, nil
}

func (s *principalService) GetPermissions(ctx context.Context, id uuid.UUID) ([]repository.Permission, error) {
	perms, err := s.store.Principals.GetPermissions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("PrincipalService.GetPermissions: %w", err)
	}

	return perms, nil
}
