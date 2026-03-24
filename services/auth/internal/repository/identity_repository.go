// identity_repository.go implements IdentityRepository, providing data access
// for auth identities backed by PostgreSQL via sqlc-generated queries.
package repository

import (
	"context"
	"fmt"

	"github.com/araujoarthur/intranetbackend/services/auth/internal/repository/sqlc/generated"
	"github.com/araujoarthur/intranetbackend/shared/pkg/helpers"
	"github.com/google/uuid"
)

// IdentityRepository defines the data access contract for auth identities.
// An identity is the stable, singular record for any entity in the system.
// It carries no credentials or profile data — it is simply a UUID anchor
// that all other services reference via external_id.
// Consumers should depend on this interface, never on the concrete implementation.
type IdentityRepository interface {
	// Create inserts a new identity and returns it.
	Create(ctx context.Context) (Identity, error)

	// GetByID retrieves an identity by its internal UUID.
	// Returns apierror.ErrNotFound if no identity exists with the given ID.
	GetByID(ctx context.Context, id uuid.UUID) (Identity, error)

	// Delete permanently removes an identity and cascades to all its credentials,
	// refresh tokens and service tokens.
	// Returns apierror.ErrNotFound if no identity exists with the given ID.
	Delete(ctx context.Context, id uuid.UUID) error
}

// identityRepository is the concrete implementation of IdentityRepository.
// It wraps sqlc-generated queries and translates between database and domain types.
// Instantiated exclusively by NewStore — never directly.
type identityRepository struct {
	q *generated.Queries
}

func (r *identityRepository) Create(ctx context.Context) (Identity, error) {
	row, err := r.q.CreateIdentity(ctx)
	if err != nil {
		return Identity{}, fmt.Errorf("IdentityRepository.Create: %w", helpers.MapError(err))
	}

	return toIdentity(row), nil
}

func (r *identityRepository) GetByID(ctx context.Context, id uuid.UUID) (Identity, error) {
	row, err := r.q.GetIdentityByID(ctx, id)
	if err != nil {
		return Identity{}, fmt.Errorf("IdentityRepository.GetByID: %w", helpers.MapError(err))
	}

	return toIdentity(row), nil
}

func (r *identityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeleteIdentity(ctx, id); err != nil {
		return fmt.Errorf("IdentityRepository.Delete: %w", helpers.MapError(err))
	}

	return nil
}
