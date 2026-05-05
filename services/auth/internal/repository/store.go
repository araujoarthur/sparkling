// store.go provides the Store type, which is the single entry point into the auth
// repository layer. It initializes all repositories from a shared pool and supports
// atomic operations via WithTx.
package repository

import (
	"context"

	"github.com/araujoarthur/intranetbackend/services/auth/internal/repository/sqlc/generated"
	"github.com/araujoarthur/intranetbackend/shared/pkg/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store is the single entry point into the Auth repository layer.
// All repositories are accessible through it. Use NewStore to instantiate.
// Use WithTx to wrap multiple repository calls in a single atomic transaction.
type Store struct {
	Identities    IdentityRepository
	Credentials   CredentialRepository
	RefreshTokens RefreshTokenRepository
	ServiceTokens ServiceTokenRepository

	pool *pgxpool.Pool

	// WithTxFn, when set, replaces the default transaction behavior in WithTx.
	// Used in unit tests to avoid requiring a real database connection.
	WithTxFn func(ctx context.Context, fn func(store *Store) error) error
}

func NewStore(pool *pgxpool.Pool) *Store {
	q := generated.New(pool)
	return &Store{
		Identities:    &identityRepository{q: q},
		Credentials:   &credentialRepository{q: q},
		RefreshTokens: &refreshTokenRepository{q: q},
		ServiceTokens: &serviceTokenRepository{q: q},
		pool:          pool,
	}
}

// WithTx runs fn inside a transaction. All repository calls inside fn
// should use the tx-scoped store returned to the callback.
func (s *Store) WithTx(ctx context.Context, fn func(store *Store) error) error {
	if s.WithTxFn != nil {
		return s.WithTxFn(ctx, fn)
	}
	return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := generated.New(tx)
		txStore := &Store{
			Identities:    &identityRepository{q: q},
			Credentials:   &credentialRepository{q: q},
			RefreshTokens: &refreshTokenRepository{q: q},
			ServiceTokens: &serviceTokenRepository{q: q},
			pool:          s.pool,
		}

		return fn(txStore)
	})
}
