# inetbctl (CLI)

`inetbctl` is the management CLI for the intranet backend. It handles database bootstrapping, migrations, seed data, and RSA key generation. Built with Cobra.

**Module:** `github.com/araujoarthur/intranetbackend/cli`
**Binary name:** `inetbctl`
**Entry point:** `cmd/inetbctl/main.go`

## Directory Structure

```
cli/
├── cmd/inetbctl/main.go          root command + command registration
└── internal/
    ├── db/
    │   └── connect.go            DB connection helper (pgx + database/sql)
    └── commands/
        ├── db.go                 DBCmd parent command
        ├── bootstrap.go          bootstrap subcommand
        ├── migrate.go            migrate subcommands (up/down/status)
        ├── seed.go               seed subcommands (up/down/status)
        └── keys.go               keys generate subcommand
```

## Command Tree

```
inetbctl
├── db
│   ├── bootstrap [--down]
│   ├── migrate
│   │   ├── up [service]
│   │   ├── down [service]
│   │   └── status [service]
│   └── seed
│       ├── up [service]
│       ├── down [service]
│       └── status [service]
└── keys
    └── generate [--out <dir>]
```

`.env` is automatically loaded at startup via `godotenv.Load()` on the root command's `PersistentPreRunE`.

---

## Commands

### `db bootstrap`

**Env required:** `OWNER_DSN`

Runs goose migrations from `db/migrations/global` using the owner (admin) database connection. This creates schemas, roles, and privileges that must exist before any service-specific migrations run.

`--down` flag reverses the bootstrap.

### `db migrate`

**Env required:** `OWNER_DSN`
**Optional:** `MIGRATIONS_DIR` (default: `db/migrations`)

Runs goose migrations for services. Each service has its own subfolder under the migrations root.

**Service discovery:** When no service argument is given, `discoverServices(root)` reads the migrations directory and returns all subdirectory names except `global`. Migration tracking table is scoped per service: `{service}.goose_db_version`.

**Subcommands:**
- `up [service]` — apply all pending migrations
- `down [service]` — roll back one migration
- `status [service]` — show applied/pending status

### `db seed`

**Env required:** `OWNER_DSN`
**Optional:** `SEEDS_DIR` (default: `db/seeds`)

Works identically to `db migrate` but targets the `db/seeds` directory. Goose table: `{service}.goose_seeds_version`.

**Subcommands:** `up`, `down`, `status`

### `keys generate`

Generates a 2048-bit RSA key pair as PEM files.

`--out <dir>` — output directory (default: `./keys`)

**Output files:**
- `private.pem` — PKCS#1 format (`RSA PRIVATE KEY` header) — for `FileKeyProvider.PrivateKey` and `EnvKeyProvider`
- `public.pem` — PKIX format (`PUBLIC KEY` header) — for `FileKeyProvider.PublicKey` and `EnvKeyProvider`

For production (`EnvKeyProvider`), base64-encode these files before setting the environment variables.

---

## Database Connection (`internal/db/connect.go`)

`Connect(dsn string) (*pgx.Conn, *sql.DB, error)` — returns both a `*pgx.Conn` and a `*sql.DB` opened from the same connection config via `pgx/stdlib.OpenDB`. The `*sql.DB` is needed by goose, which requires `database/sql`. Callers must close both. Returns an error if the DSN is empty or the connection fails.

---

## Key Design Decisions

- **Owner DSN** — all `db` subcommands use `OWNER_DSN`, the admin database user. This user has the privileges to create schemas and assign roles. Service-specific DSNs are not used here.
- **Goose for migrations and seeds** — uses the same goose engine for both, with separate versioning tables, allowing seeds to be independently applied and reversed.
- **Service isolation** — per-service goose version tables (`{service}.goose_db_version`) mean migration state for one service does not affect another.
- **No token rotation command** — `inetbctl token rotate` is referenced in README but not yet implemented in code.
