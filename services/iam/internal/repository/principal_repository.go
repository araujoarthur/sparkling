package repository

import (
	"context"
	"fmt"

	"github.com/araujoarthur/intranetbackend/services/iam/internal/repository/sqlc/generated"
	"github.com/araujoarthur/intranetbackend/shared/pkg/helpers"
	"github.com/google/uuid"
)

// PrincipalRepository defines the data access contract for IAM principals.
// A principal is any entity that can be assigned roles — either a human user
// or a service account. Principals reference identities owned by the auth service
// via ExternalID and are never the source of truth for credentials.
// Consumers should depend on this interface, never on the concrete implementation.
type PrincipalRepository interface {
	// GetByID retrieves a principal by its internal IAM UUID.
	// Returns ErrNotFound if no principal exists with the given ID.
	GetByID(ctx context.Context, id uuid.UUID) (Principal, error)

	// GetByExternalID retrieves a principal by the ID issued by the auth service
	// and its type. The combination of ExternalID and PrincipalType is unique.
	// Returns ErrNotFound if no matching principal exists.
	GetByExternalID(ctx context.Context, externalID uuid.UUID, principalType PrincipalType) (Principal, error)

	// List returns all principals ordered by creation date ascending.
	List(ctx context.Context) ([]Principal, error)

	// ListByType returns all principals of the given type ordered by creation date ascending.
	ListByType(ctx context.Context, principalType PrincipalType) ([]Principal, error)

	// Create registers a new principal in IAM for the given external identity.
	// The externalID is the ID issued by the auth service for this entity.
	// Returns ErrConflict if a principal with the same ExternalID and PrincipalType already exists.
	Create(ctx context.Context, externalID uuid.UUID, principalType PrincipalType) (Principal, error)

	// Activate marks a principal as active, allowing it to be granted roles and permissions.
	// Returns ErrNotFound if no principal exists with the given ID.
	Activate(ctx context.Context, id uuid.UUID) (Principal, error)

	// Deactivate marks a principal as inactive, effectively revoking all access
	// without removing role assignments. The principal can be reactivated later.
	// Returns ErrNotFound if no principal exists with the given ID.
	Deactivate(ctx context.Context, id uuid.UUID) (Principal, error)

	// Delete permanently removes a principal and cascades to all its role assignments.
	// Returns ErrNotFound if no principal exists with the given ID.
	Delete(ctx context.Context, id uuid.UUID) error

	// GetPermissions returns the full flat list of permissions for a principal,
	// resolved across all assigned roles. Only returns results if the principal
	// is active. Returns an empty slice if the principal has no roles or permissions.
	GetPermissions(ctx context.Context, id uuid.UUID) ([]Permission, error)
}

// principalRepository is the concrete implementation of PrincipalRepository.
// It wraps sqlc-generated queries and translates between database and domain types.
// Instantiated exclusively by NewStore — never directly.
type principalRepository struct {
	q *generated.Queries
}

//--------------------------
// Concrete Implementations
// -------------------------

func (r *principalRepository) GetByID(ctx context.Context, id uuid.UUID) (Principal, error) {
	row, err := r.q.GetPrincipalByID(ctx, id)
	if err != nil {
		return Principal{}, fmt.Errorf("PrincipalRepository.GetByID: %w", helpers.MapError(err))
	}

	return toPrincipal(row), nil
}

func (r *principalRepository) GetByExternalID(ctx context.Context, externalID uuid.UUID, principalType PrincipalType) (Principal, error) {
	row, err := r.q.GetPrincipalByExternalID(ctx, &generated.GetPrincipalByExternalIDParams{
		ExternalID:    externalID,
		PrincipalType: generated.IamPrincipalType(principalType),
	})
	if err != nil {
		return Principal{}, fmt.Errorf("PrincipalRepository.GetByExternalID: %w", helpers.MapError(err))
	}

	return toPrincipal(row), nil
}

func (r *principalRepository) List(ctx context.Context) ([]Principal, error) {
	rows, err := r.q.ListPrincipals(ctx)
	if err != nil {
		return nil, fmt.Errorf("PrincipalRepository.List: %w", helpers.MapError(err))
	}

	principals := make([]Principal, len(rows))

	for i, row := range rows {
		principals[i] = toPrincipal(row)
	}

	return principals, nil
}

func (r *principalRepository) ListByType(ctx context.Context, principalType PrincipalType) ([]Principal, error) {
	rows, err := r.q.ListPrincipalsByType(ctx, generated.IamPrincipalType(principalType))
	if err != nil {
		return nil, fmt.Errorf("PrincipalRepository.ListByType: %w", helpers.MapError(err))
	}

	principals := make([]Principal, len(rows))

	for i, row := range rows {
		principals[i] = toPrincipal(row)
	}

	return principals, nil
}

func (r *principalRepository) Create(ctx context.Context, externalID uuid.UUID, principalType PrincipalType) (Principal, error) {
	row, err := r.q.CreatePrincipal(ctx, &generated.CreatePrincipalParams{
		ExternalID:    externalID,
		PrincipalType: generated.IamPrincipalType(principalType),
	})
	if err != nil {
		return Principal{}, fmt.Errorf("PrincipalRepository.Create: %w", helpers.MapError(err))
	}

	return toPrincipal(row), nil
}

func (r *principalRepository) Activate(ctx context.Context, id uuid.UUID) (Principal, error) {
	row, err := r.q.ActivatePrincipal(ctx, id)
	if err != nil {
		return Principal{}, fmt.Errorf("PrincipalRepository.Activate: %w", helpers.MapError(err))
	}

	return toPrincipal(row), nil
}

func (r *principalRepository) Deactivate(ctx context.Context, id uuid.UUID) (Principal, error) {
	row, err := r.q.DeactivatePrincipal(ctx, id)
	if err != nil {
		return Principal{}, fmt.Errorf("PrincipalRepository.Activate: %w", helpers.MapError(err))
	}

	return toPrincipal(row), nil
}

func (r *principalRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.q.DeletePrincipal(ctx, id)
	if err != nil {
		return fmt.Errorf("PrincipalRepository.Delete: %w", helpers.MapError(err))
	}

	return nil
}

func (r *principalRepository) GetPermissions(ctx context.Context, id uuid.UUID) ([]Permission, error) {
	rows, err := r.q.GetPrincipalPermissions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("PrincipalRepository.GetPermissions: %w", helpers.MapError(err))
	}

	permissions := make([]Permission, len(rows))

	for i, row := range rows {
		permissions[i] = toPermission(row)
	}

	return permissions, nil
}
