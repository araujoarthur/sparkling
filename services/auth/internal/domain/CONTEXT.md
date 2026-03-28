# auth/internal/domain

Business logic layer for the auth service. Implements session management (registration, login, token refresh, logout) and defines the account management interface.

## Files

```
domain/
├── auth_domain.go      shared helpers: LoginResult, RefreshTokenDuration, generateToken, hashToken
├── session.go          SessionService interface + full implementation
├── account.go          AccountService interface (not yet implemented)
├── refresh_token.go    placeholder (empty)
└── service_token.go    placeholder (empty)
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

**Constructor:** `NewSessionService(store *repository.Store, hasher hasher.Hasher, provisioner provisioner.PrincipalProvisioner, privateKey *rsa.PrivateKey)`

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
    Delete(ctx, identityID uuid.UUID) error
    ChangePassword(ctx, identityID uuid.UUID, oldPassword, newPassword string) error
}
```

Interface only — no implementation yet. `Delete` cascades to credentials, refresh tokens, and service tokens. `ChangePassword` verifies the old password and revokes all active refresh tokens.

## Dependencies

| Package | Used for |
|---|---|
| `services/auth/internal/repository` | `Store`, domain types (`Identity`, `Credential`, `RefreshToken`) |
| `shared/pkg/apierror` | `ErrNotFound`, `ErrInvalidCredentials` |
| `shared/pkg/hasher` | `Hasher` interface for password hashing/verification |
| `shared/pkg/provisioner` | `PrincipalProvisioner` interface (note: package deleted, pending migration to `iamclient.IAMClient`) |
| `shared/pkg/token` | `Issue()` for signing access tokens, `Claims` |
| `shared/pkg/types` | `PrincipalTypeUser` |

## Known Issues

- `session.go` still imports `shared/pkg/provisioner` which has been deleted. Needs to be updated to use `iamclient.IAMClient` from `services/iam/client`.
