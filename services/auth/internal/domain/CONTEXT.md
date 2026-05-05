# auth/internal/domain

Business logic layer for the auth service. Implements session management (registration, login, token refresh, logout) and defines the account management interface.

## Files

```
domain/
├── auth_domain.go      shared helpers: LoginResult, RefreshTokenDuration, generateToken, hashToken
├── session.go          SessionService interface + full implementation
├── account.go          AccountService interface + implementation
├── refresh_token.go    RefreshTokenService interface + implementation
└── service_token.go    ServiceTokenService interface + implementation
```

## Shared Helpers (`auth_domain.go`)

**Types:**
- `LoginResult { AccessToken string, RefreshToken string }` — returned by `Login` and `Refresh`

**Constants:**
- `RefreshTokenDuration = 7 * 24 * time.Hour` — expiry window for refresh tokens

**Functions:**
- `generateToken() (string, error)` — 32 bytes from `crypto/rand`, base64url-encoded. Used for refresh tokens.
- `hashToken(raw string) string` — SHA-256 hex digest. Only the hash is stored in the database; raw tokens are returned to clients.

## SessionService (`session.go`)

```go
type SessionService interface {
    Register(ctx, username, password string) (repository.Identity, error)
    Login(ctx, username, password string) (LoginResult, error)
    Refresh(ctx, rawRefreshToken string) (LoginResult, error)
    Logout(ctx, rawRefreshToken string) error
    LogoutAll(ctx, identityID uuid.UUID) error
}
```

**Constructor:** `NewSessionService(store *repository.Store, hasher hasher.Hasher, iamClient iamclient.IAMClient, privateKey *rsa.PrivateKey)`

### Method Details

| Method | Behaviour |
|---|---|
| `Register` | Hash password → create identity + credential in transaction → provision IAM principal (non-fatal on failure) |
| `Login` | Fetch credential by username → verify password → issue access token (15 min) + refresh token (7 days) → store hashed refresh token → update `last_used_at` (non-fatal) → return `LoginResult` |
| `Refresh` | Hash raw token → fetch by hash → issue new access token + new refresh token → store new token, revoke old in transaction → return `LoginResult` |
| `Logout` | Hash raw token → fetch by hash → revoke single token |
| `LogoutAll` | Revoke all refresh tokens for the identity |

### Error Mapping

- Unknown username → `apierror.ErrInvalidCredentials`
- Wrong password → `apierror.ErrInvalidCredentials`
- Unknown refresh token → `apierror.ErrInvalidCredentials`

### Non-Fatal Side Effects

- IAM provisioning failure during `Register` — logged via `fmt.Printf`, not returned
- `last_used_at` update failure during `Login` — logged, not returned

## AccountService (`account.go`)

```go
type AccountService interface {
    Delete(ctx context.Context, callerID uuid.UUID, identityID uuid.UUID) error
    ChangePassword(ctx context.Context, callerID uuid.UUID, identityID uuid.UUID, oldPassword, newPassword string) error
    RemoveCredential(ctx context.Context, callerID uuid.UUID, identityID uuid.UUID, credentialID uuid.UUID) error
}
```

Implemented in `account.go`.

| Method | Behaviour |
|---|---|
| `Delete` | Caller must be the owner or hold `auth:identities:delete`; deletes the identity through the repository cascade |
| `ChangePassword` | Caller must be the owner or hold `auth:credentials:edit`; owners verify the old password; updates the credential hash and revokes all refresh tokens in a transaction |
| `RemoveCredential` | Caller must be the owner or hold `auth:credentials:delete`; password credentials cannot be removed; credential must belong to the target identity |

### RefreshTokenService (`refresh_token.go`)

```go
type RefreshTokenService interface {
    DeleteAllExpired(ctx context.Context) error
}
```

Implemented cleanup service for expired refresh tokens.

### ServiceTokenService (`service_token.go`)

```go
type ServiceTokenService interface {
    Issue(ctx context.Context, identityID uuid.UUID) (string, error)
    Rotate(ctx context.Context) error
}
```

Implemented service-token issuing and rotation logic. `Issue` signs a non-expiring service token, revokes any existing active service tokens for that identity, and stores the new token in a transaction. `Rotate` reissues every active service token, logging and skipping individual failures.

## Dependencies

| Package | Used for |
|---|---|
| `services/auth/internal/repository` | `Store`, domain types (`Identity`, `Credential`, `RefreshToken`) |
| `shared/pkg/apierror` | `ErrNotFound`, `ErrInvalidCredentials` |
| `shared/pkg/hasher` | `Hasher` interface for password hashing/verification |
| `services/iam/client` | `IAMClient` interface for IAM principal provisioning and permission checks |
| `shared/pkg/token` | `Issue()` for signing access tokens, `Claims` |
| `shared/pkg/types` | `PrincipalTypeUser` |

## Current Gaps

- Auth REST handlers are still a stub outside this package.
- `cmd/authd/main.go` has not been created.
- `ServiceTokenService` has implementation methods but no constructor function.
