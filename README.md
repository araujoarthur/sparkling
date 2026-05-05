# Intranet Backend

Microservices backend for an intranet platform, structured as a Go Workspace monorepo.

Two active services (**auth** and **iam**), a shared package library, and a management CLI live in a single repository. A third service (webapp) is planned but not yet implemented.

## Services

| Service | Purpose | Status |
|---------|---------|--------|
| **auth** (`services/auth`) | Identity and credential management; the only service that issues JWTs | Repository layer complete, domain and handler layers in progress |
| **iam** (`services/iam`) | Role-based access control (RBAC) — roles, permissions, principals | Fully implemented |
| **webapp** (`services/webapp`) | Web application service | Stub, not in workspace |

## Project Structure

```
backendv2/
├── cli/                    inetbctl management CLI
├── db/
│   ├── migrations/         Schema migrations (global, auth, iam)
│   └── seeds/              Seed data (iam)
├── services/
│   ├── auth/               Authentication service
│   ├── iam/                Identity & Access Management service
│   └── webapp/             (stub)
└── shared/pkg/             Shared packages used across all services
```

Detailed documentation for each package and service lives in `CONTEXT.md` files throughout the tree. These files are written for LLMs and coding agents — they provide the structured context needed for AI-assisted development. The root `CONTEXT.md` provides the full architectural overview.

## Service Boundary

`auth` and `iam` are internal backend services. Application services such as the future `webapp` call them with service tokens; browsers and end users should not call them directly.

Except for `/api/v1/health`, IAM routes are protected by service-token middleware. Auth routes should follow the same boundary: registration, login, refresh, logout, and account-management operations are requested by a trusted service using its own service token. When a request is made on behalf of a human user, the caller passes that user in `X-Principal-ID`.

## Prerequisites

- **Go** >= 1.26.1
- **PostgreSQL** >= 18
- **sqlc** (code generation from SQL queries)

## Go Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| [chi](https://github.com/go-chi/chi) | v5.2.5 | HTTP routing |
| [pgx](https://github.com/jackc/pgx) | v5.8.0 | PostgreSQL driver and connection pool |
| [golang-jwt](https://github.com/golang-jwt/jwt) | v5.3.1 | JWT signing and validation (RS256) |
| [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) | v0.49.0 | Argon2id and bcrypt password hashing |
| [google/uuid](https://github.com/google/uuid) | v1.6.0 | UUID generation |
| [goose](https://github.com/pressly/goose) | v3.27.0 | Database migrations and seed management |
| [cobra](https://github.com/spf13/cobra) | v1.10.2 | CLI command framework |
| [godotenv](https://github.com/joho/godotenv) | v1.5.1 | `.env` file loading |
| [sqlc](https://sqlc.dev/) | (codegen) | Type-safe SQL-to-Go code generation |

## Getting Started

### 1. Generate RSA keys

```bash
go run ./cli/cmd/inetbctl keys generate --out ./keys
```

This produces `private.pem` and `public.pem` in the `./keys` directory.

### 2. Configure environment

Create a `.env` file in the project root:

```env
OWNER_DSN=postgres://owner:password@localhost:5432/intranet
IAM_DSN=postgres://iam_user:password@localhost:5432/intranet
AUTH_DSN=postgres://auth_user:password@localhost:5432/intranet

PUBLIC_KEY_PATH=./keys/public.pem
PRIVATE_KEY_PATH=./keys/private.pem

IAM_ADDR=:8081
AUTH_HASHER=argon2
```

### 3. Bootstrap the database

```bash
go run ./cli/cmd/inetbctl db bootstrap
go run ./cli/cmd/inetbctl db migrate up
go run ./cli/cmd/inetbctl db seed up
```

### 4. Run the IAM service

```bash
go run ./services/iam/cmd/iamd
```

## CLI Reference (inetbctl)

```
inetbctl db bootstrap [--down]       Initialize (or reverse) schemas, roles, privileges
inetbctl db migrate up|down [svc]    Run migrations for a service or all services
inetbctl db migrate status [svc]     Show migration status
inetbctl db seed up|down [svc]       Apply or rollback seed data
inetbctl db seed status [svc]        Show seed status
inetbctl keys generate [--out dir]   Generate RSA key pair (default: ./keys)
```

## Environment Variables

| Variable | Used by | Purpose |
|----------|---------|---------|
| `OWNER_DSN` | inetbctl | Admin DB connection for bootstrap and migrations |
| `IAM_DSN` | IAM service | IAM service DB connection |
| `AUTH_DSN` | Auth service | Auth service DB connection |
| `PUBLIC_KEY_PATH` | all services | Path to RSA public key PEM file |
| `PRIVATE_KEY_PATH` | auth, inetbctl | Path to RSA private key PEM file |
| `IAM_ADDR` | IAM service | Listen address (default `:8081`) |
| `AUTH_HASHER` | Auth service | Hash algorithm: `argon2` or `bcrypt` |
| `MIGRATIONS_DIR` | inetbctl | Migrations root (default `db/migrations`) |
| `SEEDS_DIR` | inetbctl | Seeds root (default `db/seeds`) |

## Further Reading

The `CONTEXT.md` files are designed for LLMs and coding agents to quickly understand the codebase without reading every source file. They document types, interfaces, function signatures, business rules, and known issues.

- `CONTEXT.md` — Full project architecture, service details, shared package reference
- `docs/error-handling.md` — Cross-package error translation rules
- `docs/response-package.md` — Response envelope and handler conventions
- `services/iam/CONTEXT.md` — IAM service internals, route table, domain rules
- `services/auth/CONTEXT.md` — Auth service internals, repository interfaces
- `shared/pkg/*/CONTEXT.md` — Per-package documentation
