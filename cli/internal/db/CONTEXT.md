# cli/internal/db

Database connection helper for the CLI. Returns both a `pgx.Conn` and a `database/sql.DB` from the same connection config.

## Files

```
db/
└── connect.go    Connect function
```

## Functions

### `Connect(dsn string) (*pgx.Conn, *sql.DB, error)`

Connects to PostgreSQL using pgx, then opens a `*sql.DB` from the same connection config via `pgx/stdlib.OpenDB`.

**Why two handles:**
- `*pgx.Conn` — for any pgx-native operations
- `*sql.DB` — required by goose, which depends on `database/sql`

**Callers must close both** — `conn.Close(ctx)` and `db.Close()`.

**Error cases:**
- Empty DSN → `"DSN is not set"`
- Connection failure → raw pgx error
