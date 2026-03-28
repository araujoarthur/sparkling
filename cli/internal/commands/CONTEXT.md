# cli/internal/commands

Implements all Cobra subcommands for `inetbctl`. Each file exports a single command constructor.

## Files

```
commands/
├── db.go          DBCmd() — parent command for all database operations
├── bootstrap.go   bootstrapCmd() — schema/role/privilege initialization
├── migrate.go     migrateCmd() — service migration management (up/down/status)
├── seed.go        seedCmd() — seed data management (up/down/status)
└── keys.go        KeysCmd() — RSA key pair generation
```

## Commands

### `db bootstrap`

Runs goose migrations from `db/migrations/global` using `OWNER_DSN`. Creates schemas (`auth`, `iam`, `global`), database roles, and default privileges. `--down` reverses the bootstrap.

### `db migrate (up|down|status) [service]`

Runs goose migrations for service-specific schemas. Uses `MIGRATIONS_DIR` env var (default: `db/migrations`).

**Service discovery:** When no service argument is given, `discoverServices(root)` reads the migrations directory and returns all subdirectory names except `global`.

**Schema isolation:** Each service gets its own goose version table: `{service}.goose_db_version`.

### `db seed (up|down|status) [service]`

Identical to `db migrate` but targets `SEEDS_DIR` (default: `db/seeds`). Uses a separate goose table: `{service}.goose_seeds_version`.

### `keys generate`

Generates a 2048-bit RSA key pair:
- `private.pem` — PKCS#1 format (`RSA PRIVATE KEY` header)
- `public.pem` — PKIX format (`PUBLIC KEY` header)

`--out` flag sets the output directory (default: `./keys`).

## Shared Helpers

`discoverServices(root string) ([]string, error)` — reads a directory and returns all subdirectory names except `global`. Used by both `migrate` and `seed` commands.

`runMigration(args, fn)` and `runSeed(args, fn)` — shared runners that handle DSN loading, DB connection, service discovery, and per-service goose table scoping.

## Environment Variables

| Variable | Required by | Default |
|---|---|---|
| `OWNER_DSN` | all `db` commands | (none — error if missing) |
| `MIGRATIONS_DIR` | `db migrate` | `db/migrations` |
| `SEEDS_DIR` | `db seed` | `db/seeds` |
