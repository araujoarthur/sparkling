# database

Package `database` provides the pgx connection pool factory and a transaction helper used by every service.

## Files

```
database/
├── database.go    NewPool — connection pool factory
└── tx.go          WithTx — transaction helper
```

## Types

### `Config`

```go
type Config struct {
    DSN string
}
```

Single-field config struct. `DSN` must be a non-empty PostgreSQL connection string. `NewPool` returns an error immediately if `DSN == ""`.

### `TxFn`

```go
type TxFn func(tx pgx.Tx) error
```

The function type accepted by `WithTx`. Receives a `pgx.Tx` that must be used for all DB operations within the transaction.

## Functions

### `NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error)`

Creates a `pgxpool.Pool` and pings the database to verify connectivity. Fails fast at startup — if the DB is unreachable the service should not start.

**Error cases:**
- `cfg.DSN == ""` → returns `"database DSN is not set"`
- `pgxpool.New` fails → returns wrapped error
- `pool.Ping` fails → returns `"failed to reach the database: ..."`

### `WithTx(ctx context.Context, pool *pgxpool.Pool, fn TxFn) error`

Executes `fn` inside a transaction.

**Lifecycle:**
1. `pool.Begin(ctx)` — starts transaction
2. `fn(tx)` — runs caller-provided logic
3. If `fn` returns `nil` → `tx.Commit(ctx)`
4. If `fn` returns an error → `tx.Rollback(ctx)`
5. If commit fails → `tx.Rollback(ctx)`

**Error handling:**
- If both `fn` and rollback fail, both errors are wrapped together: `"tx failed: %w, rollback failed: %w"` so neither is lost.
- If commit fails but rollback succeeds: `"failed to commit transaction, rolled back: %w"`

**Usage — services wrap `WithTx` inside their own Store:**

```go
// Each service's Store.WithTx re-creates the sqlc Queries with the tx executor
func (s *Store) WithTx(ctx context.Context, fn func(store *Store) error) error {
    return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
        q := generated.New(tx)
        txStore := &Store{ /* all repos using q */ }
        return fn(txStore)
    })
}
```

In practice, services call `database.WithTx` through their own `Store.WithTx` wrapper, which handles re-creating sqlc queries from the transaction executor. Direct use of `database.WithTx` is possible but bypasses the Store's repository setup.
