# backendv2 — Project Context

## What this is

Microservices backend for an intranet platform. Two active services (auth, iam), a stub (webapp), and a CLI tool (inetbctl) live together as a Go Workspace monorepo. Shared infrastructure lives in `shared/pkg/`.

The workspace (`go.work`) includes `cli`, `services/auth`, `services/iam`, and `shared`. The `services/webapp` directory exists but is **not** in the workspace.

**Go module root:** `github.com/araujoarthur/intranetbackend`
**Go version:** 1.26.1

---

## Directory Layout

```
backendv2/
├── cli/                        # inetbctl — DB management and key generation CLI
├── db/
│   ├── migrations/
│   │   ├── auth/               # Auth service schema migrations
│   │   ├── iam/                # IAM service schema migrations
│   │   └── global/             # Cross-service schema (schemas, roles, privileges)
│   └── seeds/
│       └── iam/                # IAM seed data (built-in roles and permissions)
├── services/
│   ├── auth/                   # Authentication service
│   ├── iam/                    # Identity & Access Management service
│   └── webapp/                 # Web application service (stub, not implemented)
└── shared/
    └── pkg/
        ├── apierror/           # Sentinel errors + error code mapping
        ├── database/           # pgx connection pool + transaction helper
        ├── hasher/             # Password hashing (Argon2id / bcrypt)
        ├── helpers/            # pgx-level utilities and error mapper
        ├── keyprovider/        # RSA key loading (file, env, vault)
        ├── response/           # HTTP response envelope helpers
        ├── token/              # JWT issuance, validation, middleware
        └── types/              # Shared type definitions (PrincipalType)
```

---

## Services

### Auth (`services/auth`)

Single source of truth for identity and credential management. The **only** service that issues JWTs.

**Responsibilities:**
- Create and manage identities (stable UUID anchors for any entity — user or service account)
- Manage credentials: `password` type (hashed) and `service_token` type
- Issue access tokens (15 min expiry, RS256) and refresh tokens (7-day expiry)
- Issue and rotate service tokens (non-expiring; validity controlled by revocation)
- Provision an IAM principal automatically when a new identity is registered
- Provide service-token rotation logic; the background scheduler is not wired yet

**Key database tables (schema `auth`):**
- `auth.identities` — id, created_at
- `auth.credentials` — id, identity_id, type (enum: password/service_token), secret_hash, created_at, last_used_at
- `auth.refresh_tokens` — id, identity_id, token_hash, expires_at, revoked_at, created_at
- `auth.service_tokens` — id, identity_id, token, issued_at, revoked_at

**Status:** Repository layer complete. Domain layer largely implemented — `SessionService`, `AccountService`, `RefreshTokenService`, and `ServiceTokenService` are implemented. Handler layer is only a stub and `cmd/authd/main.go` has not been created.

---

### IAM (`services/iam`)

Centralised RBAC system. Manages roles, permissions, and principals; enforces access control on all write operations.

**Responsibilities:**
- CRUD for roles (lowercase names, e.g. `admin`, `super-admin`)
- CRUD for permissions (format: `scope:resource:action`, e.g. `iam:roles:write`)
- Assign/revoke permissions on roles
- Manage principals (users and service accounts) and their active/inactive state
- Assign/revoke roles on principals; enforces grant permission checks
- Expose computed effective permissions for principals

**Scope conventions in permission names:** `auth`, `iam`, `profile`, `webapp`, `global` — these are naming conventions, not enforced by code. The regex validates the `scope:resource:action` shape but accepts any lowercase scope.

**Auto-generated artefacts:**
- Creating a role automatically creates an `iam:role-{rolename}:grant` permission

**Built-in permissions (seeded on bootstrap):**
```
iam:roles:write          iam:roles:delete
iam:permissions:write    iam:permissions:delete
iam:permissions:assign   iam:permissions:revoke
iam:principals:write     iam:principals:delete
iam:role-{name}:grant    (one per role, auto-created)
```

**Key database tables (schema `iam`):**
- `iam.roles` — id, name (unique), description, is_system, created_at, updated_at
- `iam.permissions` — id, name (unique), description, created_at
- `iam.principals` — id, external_id, principal_type (enum: user/service), is_active, created_at, updated_at
- `iam.role_permissions` — (role_id, permission_id) composite PK
- `iam.principal_roles` — principal_id, role_id, granted_by, created_at

**REST API base:** `/api/v1`

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | `/health` | none | Health check |
| * | `/roles`, `/permissions`, `/principals` | service token | Full CRUD + assignment operations |

IAM is an internal service. Except for `/health`, all IAM routes are protected by `token.Middleware`, which accepts only service tokens. Auth calls `POST /principals` through the IAM client using the auth service token; users and browsers should not call IAM directly.

**Status:** Fully implemented — repository, domain, and handler layers complete.

**Client library:** `services/iam/client` — defines `IAMClient` interface (`Provision` + `HasPermission`) and its HTTP implementation (`Client`). Used by the auth service to create IAM principals on identity registration and to check permissions.

---

### WebApp (`services/webapp`)

Empty stub. The directory exists but is **not** included in `go.work`. No implementation yet.

---

## Shared Packages

### `apierror`

Single source of truth for error classification. No service should define its own sentinel errors for conditions covered here.

**Sentinel errors → HTTP status:**
| Error | Status |
|-------|--------|
| `ErrNotFound` | 404 |
| `ErrConflict` | 409 |
| `ErrForbidden` | 403 |
| `ErrInvalidArgument` | 400 |
| `ErrUnauthorized` | 401 |
| `ErrInvalidCredentials` | 401 |
| `ErrInternal` | 500 |

**Functions:** `HTTPStatus(err) int`, `Code(err) ErrorCode`

---

### `response`

Standard JSON response envelopes. All handlers write responses through this package, never directly to `ResponseWriter`.

```
{ "data": ... }                             // JSON / Paginated
{ "data": ..., "meta": { page, per_page, total, total_pages } }
{ "error": { "code": "...", "message": "..." } }
```

**Functions:** `JSON(w, status, data)`, `Paginated(w, status, data, page, perPage, total)`, `Error(w, err, message)`

See `docs/response-package.md` for handler usage conventions.

---

### `token`

JWT handling (RS256 / asymmetric).

- **Issue** — only auth-domain code currently calls token issuance; `inetbctl` generates RSA keys but does not issue JWTs
- **Parse** — any service calls `Parse()` with the public key
- **Middleware** — Chi middleware that validates the Bearer token; only accepts service tokens (user tokens are rejected with 401); injects claims + acting principal into request context
- **FromContext / ActingPrincipalFromContext** — read injected values in handlers

**Token lifetimes:**
- User tokens: 15 minutes (refreshed via refresh token flow)
- Service tokens: non-expiring; revocation managed in auth service

---

### `keyprovider`

Interface-based RSA key loading; implementations are swappable per environment.

| Implementation | Source | When to use |
|---|---|---|
| `FileKeyProvider` | PEM files on disk | Dev / CI |
| `EnvKeyProvider` | Base64-encoded PEM in env vars | Production / containers |
| `VaultKeyProvider` | HashiCorp Vault (placeholder) | Future |

Services depend on `PublicKeyProvider` / `PrivateKeyProvider` interfaces, never concrete types.

---

### `database`

pgx connection pool factory and transaction helper.

- `NewPool(ctx, cfg)` — creates and pings pool; fails fast if DB is unreachable
- `WithTx(ctx, pool, fn)` — begins a transaction, calls `fn`, commits on success, rolls back on error

---

### `hasher`

Password hashing abstraction.

| Implementation | Algorithm | Notes |
|---|---|---|
| `Argon2Hasher` | Argon2id | Default; 64 MB memory, 3 iterations, 2-way parallel; salt embedded in hash |
| `BcryptHasher` | bcrypt | Legacy; minimum cost 12 |

Both implement `Hasher { Hash(plaintext string) (string, error); Verify(plaintext, hash string) (bool, error) }`.

---

### `helpers`

Low-level pgx utilities:
- `MapError(err)` — translates `pgx.ErrNoRows` → `apierror.ErrNotFound`, unique violation → `apierror.ErrConflict`
- `PgxText(s)` — converts Go string to `pgtype.Text`; empty string becomes NULL
- `FromNullableTime(t)` — converts `pgtype.Timestamptz` to `*time.Time`

---

### `types`

Shared type definitions with no business logic; exists solely to avoid import cycles.

```go
type PrincipalType string
const (
    PrincipalTypeUser    PrincipalType = "user"
    PrincipalTypeService PrincipalType = "service"
)
```

---

## Architecture

### Layer Model

Every service follows a strict three-layer model:

```
Handler layer   — HTTP, request parsing, response writing
    ↓
Domain layer    — business rules, validation, permission checks
    ↓
Repository layer — data access, sqlc wrappers, error mapping
    ↓
Database (PostgreSQL)
```

Dependencies flow strictly downward. Handlers depend on domain service interfaces (e.g. `domain.RoleService`). Domain services receive the concrete `*repository.Store` but interact with its fields, which are repository interfaces (e.g. `RoleRepository`, `PrincipalRepository`). No layer knows about the layers above it.

### Error Flow

```
pgx error
  → helpers.MapError()        (in repository)
  → apierror sentinel
  → apierror.HTTPStatus/Code  (in handler via response.Error)
  → JSON error response
```

### Repository Store Pattern

All repositories are accessed through a single `Store` struct, never instantiated individually. Transactions are supported transparently:

```go
store := repository.NewStore(pool)

// Direct access
identity, err := store.Identities.Create(ctx)

// Transactional
err := store.WithTx(ctx, func(tx *Store) error {
    identity, _ := tx.Identities.Create(ctx)
    _, _ = tx.Credentials.Create(ctx, identity.ID, ...)
    return nil
})
```

### Type-Safe SQL (sqlc)

Schema migrations define the authoritative database structure. SQL query files are hand-written; `go generate ./...` produces type-safe Go code from them. Repository structs wrap the generated code and map to domain types via mapper functions.

### Service-to-Service Authentication

Auth and IAM are internal services. Browser clients and end users should not call them directly. They are called by trusted application services, such as the future `webapp`, using service-to-service authentication.

1. Each calling service has a service account identity in auth.
2. Auth issues a non-expiring service token for that service identity.
3. The calling service attaches that token as `Authorization: Bearer <token>`.
4. The receiving service validates the token's RS256 signature against the shared public key.
5. `token.Middleware` accepts service tokens only; user tokens are rejected.
6. When the service is acting for a human user, `X-Principal-ID` carries that user's principal UUID.
7. If `X-Principal-ID` is absent, handlers treat the calling service principal as the actor.

Only health checks are intentionally unauthenticated. Registration, login, refresh, logout, IAM principal creation, and IAM management routes are service-token-protected internal operations.

### Dependency Injection

Services are wired at startup in `main.go`: key provider → DB pool → store → domain services → HTTP server. No global state; everything is passed explicitly.

---

## CLI (inetbctl)

Database management and key generation. Built with Cobra.

```
inetbctl db bootstrap              # Initialize schemas, roles, privileges
inetbctl db migrate up|down [svc]  # Run migrations for a service
inetbctl db seed up|down [svc]     # Apply/rollback seed data
inetbctl keys generate [--out dir] # Generate RSA key pair for token signing
```

Note: `inetbctl token rotate` is referenced in older READMEs but is not yet implemented in code.

---

## Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/go-chi/chi/v5` | HTTP routing |
| `github.com/golang-jwt/jwt/v5` | JWT signing and validation |
| `github.com/jackc/pgx/v5` | PostgreSQL driver |
| `golang.org/x/crypto` | Argon2id and bcrypt |
| `github.com/google/uuid` | UUID generation |
| `github.com/pressly/goose/v3` | Database migrations |
| `github.com/spf13/cobra` | CLI framework |
| `github.com/joho/godotenv` | `.env` file loading |
| `sqlc` (codegen) | Type-safe SQL generation |

---

## Environment Variables

| Variable | Used by | Purpose |
|----------|---------|---------|
| `OWNER_DSN` | inetbctl | Admin DB connection for bootstrap |
| `{SERVICE}_DSN` | each service | Service-specific DB connection |
| `PUBLIC_KEY_PATH` / `PRIVATE_KEY_PATH` | services | PEM key file paths; IAM currently loads only the public key but passes both paths to `FileKeyProvider` |
| `{SERVICE}_ADDR` | each service | Listen address |
| `AUTH_HASHER` | auth | Hash algorithm: `argon2` or `bcrypt` |
