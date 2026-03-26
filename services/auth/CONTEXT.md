# Auth Service

The auth service is the single source of truth for identity and credential management. It is the **only** service that issues JWTs. All other services validate tokens but never create them.

**Module:** `github.com/araujoarthur/intranetbackend/services/auth`
**Default port:** not yet set (handler layer in progress)
**Entry point:** `cmd/authd/main.go` (not yet created)

## Directory Structure

```
services/auth/
├── internal/
│   ├── domain/
│   │   ├── auth_domain.go      shared domain helpers (generateToken)
│   │   ├── session.go          SessionService interface (in progress)
│   │   ├── identity.go         stub
│   │   ├── credential.go       stub
│   │   ├── refresh_token.go    stub
│   │   └── service_token.go    stub
│   └── repository/
│       ├── auth_repository.go  domain types + mappers
│       ├── store.go            Store — single entry point to all repos
│       ├── identity_repository.go
│       ├── credential_repository.go
│       ├── refresh_token_repository.go
│       ├── service_token_repository.go
│       └── sqlc/
│           └── generated/      sqlc-generated query code
```

## Responsibilities

- Create and manage **identities** — stable UUID anchors for any entity (user or service account)
- Manage **credentials** — password (hashed) or service_token type, each linked to an identity
- Issue **access tokens** (15 min, RS256 JWT) and **refresh tokens** (7-day expiry)
- Issue and rotate **service tokens** (non-expiring JWTs; revocation-controlled)
- Provision an IAM principal automatically when a new identity registers (via `provisioner.PrincipalProvisioner`)
- Run a daily background job to rotate service tokens

## Database Schema (`auth` schema)

| Table | Columns | Notes |
|---|---|---|
| `auth.identities` | `id uuid PK`, `created_at` | Stable anchor; no profile data |
| `auth.credentials` | `id`, `identity_id FK`, `type` (enum), `identifier`, `secret_hash`, `created_at`, `last_used_at` | One row per auth method per identity |
| `auth.refresh_tokens` | `id`, `identity_id FK`, `token_hash`, `expires_at`, `revoked_at`, `created_at` | Hashed; 7-day expiry |
| `auth.service_tokens` | `id`, `identity_id FK`, `token` (JWT string), `issued_at`, `revoked_at` | Full JWT stored; no expiry |

## Repository Layer

### Domain Types (`auth_repository.go`)

| Type | Fields |
|---|---|
| `Identity` | `ID uuid`, `CreatedAt time.Time` |
| `Credential` | `ID`, `IdentityID`, `Type CredentialType`, `Identifier string`, `SecretHash string`, `CreatedAt`, `LastUsedAt *time.Time` |
| `RefreshToken` | `ID`, `IdentityID`, `TokenHash string`, `ExpiresAt`, `RevokedAt *time.Time`, `CreatedAt` |
| `ServiceToken` | `ID`, `IdentityID`, `Token string`, `IssuedAt`, `RevokedAt *time.Time` |

`CredentialType` is a string enum: `"password"` / `"service_token"`.

Mappers (`toIdentity`, `toCredential`, `toRefreshToken`, `toServiceToken`) translate sqlc-generated types to domain types using `helpers.FromNullableTime` for nullable timestamps.

### Store (`store.go`)

Single entry point to all repositories. Never instantiate repositories directly.

```go
store := repository.NewStore(pool)
store.Identities    // IdentityRepository
store.Credentials   // CredentialRepository
store.RefreshTokens // RefreshTokenRepository
store.ServiceTokens // ServiceTokenRepository
```

`store.WithTx(ctx, func(tx *Store) error)` — wraps operations in a transaction. The tx-scoped store re-creates all repositories from the transaction executor.

### IdentityRepository

| Method | Description |
|---|---|
| `Create(ctx)` | Insert new identity; returns `Identity` |
| `GetByID(ctx, id)` | Fetch by UUID; `ErrNotFound` if missing |
| `Delete(ctx, id)` | Hard delete; cascades to credentials, tokens |

### CredentialRepository

| Method | Description |
|---|---|
| `Create(ctx, identityID, type, identifier, secretHash)` | `ErrConflict` on duplicate type+identifier |
| `GetByID(ctx, credentialID)` | `ErrNotFound` if missing |
| `GetByTypeAndIdentifier(ctx, type, identifier)` | Primary login lookup |
| `GetByIdentity(ctx, identityID)` | All credentials for an identity |
| `GetByIdentityAndType(ctx, identityID, type)` | Single credential by type |
| `UpdateLastUsed(ctx, credentialID)` | Sets `last_used_at = now()` |
| `UpdateSecret(ctx, credentialID, secretHash)` | Password change / token rotation |
| `Delete(ctx, credentialID)` | Hard delete single credential |
| `DeleteByIdentity(ctx, identityID)` | Bulk delete all credentials |

### RefreshTokenRepository

| Method | Description |
|---|---|
| `Create(ctx, identityID, tokenHash, expiresAt)` | Store hashed refresh token |
| `GetByID(ctx, tokenID)` | Fetch by UUID |
| `GetByHash(ctx, hash)` | Login lookup by token hash |
| `GetActiveByIdentity(ctx, identityID)` | All non-revoked, non-expired tokens |
| `Revoke(ctx, tokenID)` | Mark single token revoked |
| `RevokeAllByIdentity(ctx, identityID)` | Logout / password change |
| `DeleteAllExpired(ctx)` | Cleanup job |
| `DeleteAllByIdentity(ctx, identityID)` | Called on identity delete |

### ServiceTokenRepository

| Method | Description |
|---|---|
| `Create(ctx, identityID, token)` | Store signed JWT string |
| `GetByID(ctx, tokenID)` | Fetch by UUID |
| `GetActiveByIdentity(ctx, identityID)` | Current active token |
| `GetByToken(ctx, token)` | Validate incoming service token |
| `Revoke(ctx, tokenID)` | Revoke single token |
| `RevokeAllByIdentity(ctx, identityID)` | Called before rotation |
| `ListActive(ctx)` | All active tokens; used by rotation job |

## Domain Layer (In Progress)

### `auth_domain.go`

`generateToken() (string, error)` — generates a 32-byte cryptographically secure random token encoded as base64 URL. Used for refresh tokens (only the hash is stored, not the raw token).

### `session.go` — `SessionService` interface

```go
type SessionService interface {
    Register(ctx context.Context, username, password string) (repository.Identity, error)
}
```

Currently just the interface definition. Implementation stubs (`identity.go`, `credential.go`, `refresh_token.go`, `service_token.go`) are empty — the full domain layer is in progress.

## Status

- Repository layer: **complete** — all four repositories implemented via interface
- Domain layer: **in progress** — `SessionService` interface defined, implementations are stubs
- Handler layer: **not started**
- `main.go`: **not started**

## Shared Packages Used

| Package | Purpose |
|---|---|
| `shared/pkg/apierror` | Sentinel errors returned from repository |
| `shared/pkg/helpers` | `MapError`, `PgxText`, `FromNullableTime` |
| `shared/pkg/database` | `WithTx` via `store.WithTx` |
| `shared/pkg/token` | `Issue` for signing JWTs (domain layer, when implemented) |
| `shared/pkg/hasher` | Password hashing (domain layer, when implemented) |
| `shared/pkg/provisioner` | `PrincipalProvisioner` (domain layer, when implemented) |
| `shared/pkg/keyprovider` | RSA key loading (main, when implemented) |
