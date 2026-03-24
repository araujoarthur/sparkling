// refresh_tokens_repository.go implements RefreshTokenRepository, providing data access
// for auth refresh tokens backed by PostgreSQL via sqlc-generated queries.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/araujoarthur/intranetbackend/services/auth/internal/repository/sqlc/generated"
	"github.com/araujoarthur/intranetbackend/shared/pkg/helpers"
	"github.com/google/uuid"
)

// RefreshTokenRepository defines the data access contract for auth refresh tokens.
// A refresh token is a long-lived token issued alongside an access token that
// allows the client to obtain a new access token without re-authenticating.
// Consumers should depend on this interface, never on the concrete implementation.
type RefreshTokenRepository interface {
	// Create inserts a new refresh token for the given identity and returns it.
	Create(ctx context.Context, identityID uuid.UUID, tokenHash string, expiresAt time.Time) (RefreshToken, error)

	// GetByID retrieves a refresh token by its internal UUID.
	// Returns apierror.ErrNotFound if no token exists with the given ID.
	GetByID(ctx context.Context, tokenID uuid.UUID) (RefreshToken, error)

	// GetByHash retrieves an active, non-expired refresh token by its hash.
	// Returns apierror.ErrNotFound if no matching active token exists.
	GetByHash(ctx context.Context, tokenHash string) (RefreshToken, error)

	// GetActiveByIdentity returns all active, non-expired refresh tokens
	// for the given identity, ordered by creation date descending.
	GetActiveByIdentity(ctx context.Context, identityID uuid.UUID) ([]RefreshToken, error)

	// Revoke marks a refresh token as revoked.
	// A revoked token cannot be used to issue new access tokens.
	// Returns apierror.ErrNotFound if no token exists with the given ID.
	Revoke(ctx context.Context, tokenID uuid.UUID) error

	// RevokeAllByIdentity revokes all active refresh tokens for the given identity.
	// Called on logout or password change to invalidate all active sessions.
	RevokeAllByIdentity(ctx context.Context, identityID uuid.UUID) error

	// DeleteAllExpired permanently removes all expired refresh tokens.
	// Called by a cleanup job to prevent unbounded table growth.
	DeleteAllExpired(ctx context.Context) error

	// DeleteAllByIdentity permanently removes all refresh tokens for the given identity.
	// Called when an identity is deleted.
	DeleteAllByIdentity(ctx context.Context, identityID uuid.UUID) error
}

// refreshTokenRepository is the concrete implementation of RefreshTokenRepository.
// It wraps sqlc-generated queries and translates between database and domain types.
// Instantiated exclusively by NewStore — never directly.
type refreshTokenRepository struct {
	q *generated.Queries
}

func (r *refreshTokenRepository) Create(ctx context.Context, identityID uuid.UUID, tokenHash string, expiresAt time.Time) (RefreshToken, error) {
	row, err := r.q.CreateRefreshToken(ctx, &generated.CreateRefreshTokenParams{
		IdentityID: identityID,
		TokenHash:  tokenHash,
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		return RefreshToken{}, fmt.Errorf("RefreshTokenRepository.Create: %w", helpers.MapError(err))
	}

	return toRefreshToken(row), nil
}

func (r *refreshTokenRepository) GetByID(ctx context.Context, tokenID uuid.UUID) (RefreshToken, error) {
	row, err := r.q.GetRefreshTokenByID(ctx, tokenID)
	if err != nil {
		return RefreshToken{}, fmt.Errorf("RefreshTokenRepository.GetByID: %w", helpers.MapError(err))
	}

	return toRefreshToken(row), nil
}

func (r *refreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (RefreshToken, error) {
	row, err := r.q.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return RefreshToken{}, fmt.Errorf("RefreshTokenRepository.GetByHash: %w", helpers.MapError(err))
	}

	return toRefreshToken(row), nil
}

func (r *refreshTokenRepository) GetActiveByIdentity(ctx context.Context, identityID uuid.UUID) ([]RefreshToken, error) {
	rows, err := r.q.GetActiveRefreshTokensByIdentity(ctx, identityID)
	if err != nil {
		return nil, fmt.Errorf("RefreshTokenRepository.GetActiveByIdentity: %w", helpers.MapError(err))
	}

	refreshTokens := make([]RefreshToken, len(rows))
	for i, token := range rows {
		refreshTokens[i] = toRefreshToken(token)
	}

	return refreshTokens, nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, tokenID uuid.UUID) error {
	if err := r.q.RevokeRefreshToken(ctx, tokenID); err != nil {
		return fmt.Errorf("RefreshTokenRepository.Revoke: %w", helpers.MapError(err))
	}

	return nil
}

func (r *refreshTokenRepository) RevokeAllByIdentity(ctx context.Context, identityID uuid.UUID) error {
	if err := r.q.RevokeAllRefreshTokensByIdentity(ctx, identityID); err != nil {
		return fmt.Errorf("RefreshTokenRepository.RevokeAllByIdentity: %w", helpers.MapError(err))
	}

	return nil
}

func (r *refreshTokenRepository) DeleteAllExpired(ctx context.Context) error {
	if err := r.q.DeleteExpiredRefreshTokens(ctx); err != nil {
		return fmt.Errorf("RefreshTokenRepository.DeleteAllExpired: %w", helpers.MapError(err))
	}

	return nil
}

func (r *refreshTokenRepository) DeleteAllByIdentity(ctx context.Context, identityID uuid.UUID) error {
	if err := r.q.DeleteAllRefreshTokensByIdentity(ctx, identityID); err != nil {
		return fmt.Errorf("RefreshTokenRepository.DeleteAllByIdentity: %w", helpers.MapError(err))
	}

	return nil
}
