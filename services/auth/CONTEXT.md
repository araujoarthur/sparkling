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
│   │   ├── auth_domain.go      shared helpers (LoginResult, RefreshTokenDuration, generateToken, hashToken)
│   │   ├── session.go          SessionService interface + full implementation
│   │   ├── account.go          AccountService interface + implementation
│   │   ├── refresh_token.go    RefreshTokenService interface + implementation
│   │   └── service_token.go    ServiceTokenService interface + implementation
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
- Provision an IAM principal automatically when a new identity registers (via `iamclient.IAMClient`)
- Provide service-token rotation logic; the background scheduler is not wired yet

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

---

## Domain Layer

### Shared Helpers (`auth_domain.go`)

**Types:**
- `LoginResult` — `{ AccessToken string, RefreshToken string }` — returned by `Login` and `Refresh`

**Constants:**
- `RefreshTokenDuration = 7 * 24 * time.Hour` — used when storing refresh tokens

**Functions:**
- `generateToken() (string, error)` — 32 bytes from `crypto/rand`, base64url-encoded. Used for refresh tokens (only the SHA-256 hash is stored).
- `hashToken(raw string) string` — SHA-256 hex digest of a raw token string.

### SessionService (`session.go`)

```go
type SessionService interface {
    Register(ctx, username, password string) (repository.Identity, error)
    Login(ctx, username, password string) (LoginResult, error)
    Refresh(ctx, rawRefreshToken string) (LoginResult, error)
    Logout(ctx, rawRefreshToken string) error
    LogoutAll(ctx, identityID uuid.UUID) error
}
```

**Constructor:** `NewSessionService(store, hasher, iamClient, privateKey)`

| Method | Behaviour |
|---|---|
| `Register` | Hashes password → creates identity + credential in a transaction → provisions IAM principal (non-fatal on failure) |
| `Login` | Fetches credential by username → verifies password hash → issues access token + refresh token → stores hashed refresh token with 7-day expiry → updates `last_used_at` (non-fatal) → returns `LoginResult` |
| `Refresh` | Hashes raw token → fetches by hash → issues new access token + new refresh token → stores new, revokes old in a transaction → returns `LoginResult` |
| `Logout` | Hashes raw token → fetches by hash → revokes the token |
| `LogoutAll` | Revokes all refresh tokens for the identity |

**Error mapping:**
- Unknown username → `apierror.ErrInvalidCredentials`
- Wrong password → `apierror.ErrInvalidCredentials`
- Unknown refresh token → `apierror.ErrInvalidCredentials`

**Non-fatal side effects:**
- IAM provisioning failure during `Register` — logged, not returned
- `last_used_at` update failure during `Login` — logged, not returned

### AccountService (`account.go`)

```go
type AccountService interface {
    Delete(ctx context.Context, callerID uuid.UUID, identityID uuid.UUID) error
    ChangePassword(ctx context.Context, callerID uuid.UUID, identityID uuid.UUID, oldPassword, newPassword string) error
    RemoveCredential(ctx context.Context, callerID uuid.UUID, identityID uuid.UUID, credentialID uuid.UUID) error
}
```

Implemented. `Delete` requires the caller to be the identity owner or hold `auth:identities:delete`, then deletes the identity through the repository cascade. `ChangePassword` requires ownership or `auth:credentials:edit`; owners must verify the old password, admins with permission skip that check, and all refresh tokens are revoked after the password changes. `RemoveCredential` requires ownership or `auth:credentials:delete`, blocks removal of password credentials, and verifies the credential belongs to the target identity.

### RefreshTokenService (`refresh_token.go`)

```go
type RefreshTokenService interface {
    DeleteAllExpired(ctx context.Context) error
}
```

Implemented cleanup service for deleting expired refresh tokens through the repository.

### ServiceTokenService (`service_token.go`)

```go
type ServiceTokenService interface {
    Issue(ctx context.Context, identityID uuid.UUID) (string, error)
    Rotate(ctx context.Context) error
}
```

Implemented service-token issuing and rotation logic. `Issue` signs a non-expiring service JWT, revokes existing active service tokens for the identity, and stores the new token in one transaction. `Rotate` lists active service tokens and reissues them one identity at a time, logging per-identity failures and continuing.

---

## Status

- Repository layer: **complete** — all four repositories implemented
- Domain layer: **mostly implemented** — `SessionService`, `AccountService`, `RefreshTokenService`, and `ServiceTokenService` are implemented. `ServiceTokenService` currently has no constructor function.
- Handler layer: **stub only** — `internal/handler/rest/server.go` currently defines an empty `Server` struct.
- `cmd/authd/main.go`: **not started**

## Shared Packages Used

| Package | Purpose |
|---|---|
| `shared/pkg/apierror` | Sentinel errors returned from repository and domain |
| `shared/pkg/helpers` | `MapError`, `PgxText`, `FromNullableTime` |
| `shared/pkg/database` | `WithTx` via `store.WithTx` |
| `shared/pkg/token` | `Issue` for signing access tokens (RS256) |
| `shared/pkg/hasher` | Password hashing (`Argon2Hasher` / `BcryptHasher`) |
| `shared/pkg/types` | `PrincipalType` constants |
| `services/iam/client` | `IAMClient` interface — IAM principal provisioning and permission checks |
| `shared/pkg/keyprovider` | RSA key loading (main, when implemented) |
