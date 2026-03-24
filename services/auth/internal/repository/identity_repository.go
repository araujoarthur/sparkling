package repository

import (
	"context"

	"github.com/araujoarthur/intranetbackend/services/auth/internal/repository/sqlc/generated"
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
