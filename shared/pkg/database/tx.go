package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxFn func(tx pgx.Tx) error

// WithTx executes fn inside a database transaction.
//
// It begins a transaction, passes it to fn, and commits if fn returns nil.
// If fn returns an error, or if the commit fails, the transaction is rolled back.
//
// If both the original error and the rollback fail, both errors are returned
// wrapped together so neither is lost.
//
// The caller is responsible for using the provided tx inside fn for all
// database operations that should be part of the transaction.
//
// Example:
//
//	err := database.WithTx(ctx, pool, func(tx pgx.Tx) error {
//	    if err := assignRole(ctx, tx, principalID, roleID); err != nil {
//	        return err
//	    }
//	    return logAuditEntry(ctx, tx, principalID, roleID)
//	})
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn TxFn) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx failed: %w, rollback failed: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("commit failed: %w, rollback failed: %w", err, rbErr)
		}
		return fmt.Errorf("failed to commit transaction, rolled back: %w", err)
	}

	return nil
}
