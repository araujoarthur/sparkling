// service_tokens_repository.go implements ServiceTokenRepository, providing data access
// for auth service tokens backed by PostgreSQL via sqlc-generated queries.
package repository

import (
	"context"
	"fmt"

	"github.com/araujoarthur/intranetbackend/services/auth/internal/repository/sqlc/generated"
	"github.com/araujoarthur/intranetbackend/shared/pkg/helpers"
	"github.com/google/uuid"
)

// ServiceTokenRepository defines the data access contract for auth service tokens.
// A service token is a non-expiring token issued to service accounts.
// Its validity is controlled by revocation rather than expiry.
// Only one active service token exists per identity at any time.
// Consumers should depend on this interface, never on the concrete implementation.
type ServiceTokenRepository interface {
	// Create inserts a new service token for the given identity and returns it.
	// The token string should be a pre-signed JWT issued by the token package.
	Create(ctx context.Context, identityID uuid.UUID, token string) (ServiceToken, error)

	// GetByID retrieves a service token by its internal UUID.
	// Returns apierror.ErrNotFound if no token exists with the given ID.
	GetByID(ctx context.Context, tokenID uuid.UUID) (ServiceToken, error)

	// GetActiveByIdentity retrieves the current active service token for the given identity.
	// Returns apierror.ErrNotFound if no active token exists.
	GetActiveByIdentity(ctx context.Context, identityID uuid.UUID) (ServiceToken, error)

	// GetByToken retrieves an active service token by its JWT string.
	// Used to validate incoming service token requests.
	// Returns apierror.ErrNotFound if no matching active token exists.
	GetByToken(ctx context.Context, token string) (ServiceToken, error)

	// Revoke marks a service token as revoked.
	// Returns apierror.ErrNotFound if no token exists with the given ID.
	Revoke(ctx context.Context, tokenID uuid.UUID) error

	// RevokeAllByIdentity revokes all active service tokens for the given identity.
	// Called before issuing a new service token during rotation.
	RevokeAllByIdentity(ctx context.Context, identityID uuid.UUID) error

	// ListActive returns all active service tokens across all identities,
	// ordered by issued_at descending.
	// Used by the daily rotation background job.
	ListActive(ctx context.Context) ([]ServiceToken, error)
}

// serviceTokenRepository is the concrete implementation of ServiceTokenRepository.
// It wraps sqlc-generated queries and translates between database and domain types.
// Instantiated exclusively by NewStore — never directly.
type serviceTokenRepository struct {
	q *generated.Queries
}

func (r *serviceTokenRepository) Create(ctx context.Context, identityID uuid.UUID, token string) (ServiceToken, error) {
	row, err := r.q.CreateServiceToken(ctx, &generated.CreateServiceTokenParams{
		IdentityID: identityID,
		Token:      token,
	})
	if err != nil {
		return ServiceToken{}, fmt.Errorf("ServiceTokenRepository.Create: %w", helpers.MapError(err))
	}

	return toServiceToken(row), nil
}

func (r *serviceTokenRepository) GetByID(ctx context.Context, tokenID uuid.UUID) (ServiceToken, error) {
	row, err := r.q.GetServiceTokenByID(ctx, tokenID)
	if err != nil {
		return ServiceToken{}, fmt.Errorf("ServiceTokenRepository.GetByID: %w", helpers.MapError(err))
	}

	return toServiceToken(row), nil
}

func (r *serviceTokenRepository) GetActiveByIdentity(ctx context.Context, identityID uuid.UUID) (ServiceToken, error) {
	row, err := r.q.GetActiveServiceTokenByIdentity(ctx, identityID)
	if err != nil {
		return ServiceToken{}, fmt.Errorf("ServiceTokenRepository.GetActiveByIdentity: %w", helpers.MapError(err))
	}

	return toServiceToken(row), nil
}

func (r *serviceTokenRepository) GetByToken(ctx context.Context, token string) (ServiceToken, error) {
	row, err := r.q.GetServiceTokenByToken(ctx, token)
	if err != nil {
		return ServiceToken{}, fmt.Errorf("ServiceTokenRepository.GetByToken: %w", helpers.MapError(err))
	}

	return toServiceToken(row), nil
}

func (r *serviceTokenRepository) Revoke(ctx context.Context, tokenID uuid.UUID) error {
	if err := r.q.RevokeServiceToken(ctx, tokenID); err != nil {
		return fmt.Errorf("ServiceTokenRepository.Revoke: %w", helpers.MapError(err))
	}

	return nil
}

func (r *serviceTokenRepository) RevokeAllByIdentity(ctx context.Context, identityID uuid.UUID) error {
	if err := r.q.RevokeAllServiceTokensByIdentity(ctx, identityID); err != nil {
		return fmt.Errorf("ServiceTokenRepository.RevokeAllByIdentity: %w", helpers.MapError(err))
	}

	return nil
}

func (r *serviceTokenRepository) ListActive(ctx context.Context) ([]ServiceToken, error) {
	rows, err := r.q.ListActiveServiceTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("ServiceTokenRepository.ListActive: %w", helpers.MapError(err))
	}

	serviceTokens := make([]ServiceToken, len(rows))
	for i, serviceToken := range rows {
		serviceTokens[i] = toServiceToken(serviceToken)
	}

	return serviceTokens, nil
}
