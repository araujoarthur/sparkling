// store.go provides the Store type, which is the single entry point into the IAM
// repository layer. It initializes all repositories from a shared pool and supports
// atomic operations via WithTx.
package repository

import (
	"context"

	"github.com/araujoarthur/intranetbackend/services/iam/internal/repository/sqlc/generated"
	"github.com/araujoarthur/intranetbackend/shared/pkg/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store is the single entry point into the IAM repository layer.
// All repositories are accessible through it. Use NewStore to instantiate.
// Use WithTx to wrap multiple repository calls in a single atomic transaction.
type Store struct {
	Roles           RoleRepository
	Permissions     PermissionRepository
	RolePermissions RolePermissionRepository
	Principals      PrincipalRepository
	PrincipalRoles  PrincipalRoleRepository

	pool *pgxpool.Pool // underlying connection pool, used for transaction management
}

// NewStore constructs a Store with all repositories initialized.
func NewStore(pool *pgxpool.Pool) *Store {
	q := generated.New(pool)
	return &Store{
		Roles:           &roleRepository{q: q},
		Permissions:     &permissionRepository{q: q},
		RolePermissions: &rolePermissionRepository{q: q},
		Principals:      &principalRepository{q: q},
		PrincipalRoles:  &principalRoleRepository{q: q},
		pool:            pool,
	}
}

// WithTx runs fn inside a transaction. All repository calls inside fn
// should use the tx-scoped store returned to the callback.
//
// Example:
//
//	err := store.WithTx(ctx, func(tx *Store) error {
//	    principal, err := tx.Principals.Create(ctx, externalID, repository.PrincipalTypeUser)
//	    if err != nil {
//	        return err
//	    }
//	    _, err = tx.PrincipalRoles.Assign(ctx, principal.ID, roleID, grantedBy)
//	    return err
//	})
func (s *Store) WithTx(ctx context.Context, fn func(store *Store) error) error {
	return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := generated.New(tx)
		txStore := &Store{
			Roles:           &roleRepository{q: q},
			Permissions:     &permissionRepository{q: q},
			RolePermissions: &rolePermissionRepository{q: q},
			Principals:      &principalRepository{q: q},
			PrincipalRoles:  &principalRoleRepository{q: q},
			pool:            s.pool,
		}
		return fn(txStore)
	})
}
