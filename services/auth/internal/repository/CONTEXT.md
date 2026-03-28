# auth/internal/repository

Data access layer for the auth service. Wraps sqlc-generated queries with domain types, error mapping, and transaction support.

## Files

```
repository/
├── auth_repository.go           domain types (Identity, Credential, RefreshToken, ServiceToken) + mappers
├── store.go                     Store — single entry point to all repos + WithTx
├── identity_repository.go       IdentityRepository interface + implementation
├── credential_repository.go     CredentialRepository interface + implementation
├── refresh_token_repository.go  RefreshTokenRepository interface + implementation
├── service_token_repository.go  ServiceTokenRepository interface + implementation
└── sqlc/
    ├── queries/                 hand-written SQL query files
    └── generated/               sqlc-generated Go code (do not edit)
```

## Domain Types (`auth_repository.go`)

| Type | Fields |
|---|---|
| `Identity` | `ID uuid`, `CreatedAt time.Time` |
| `Credential` | `ID`, `IdentityID`, `Type CredentialType`, `Identifier string`, `SecretHash string`, `CreatedAt`, `LastUsedAt *time.Time` |
| `RefreshToken` | `ID`, `IdentityID`, `TokenHash string`, `ExpiresAt`, `RevokedAt *time.Time`, `CreatedAt` |
| `ServiceToken` | `ID`, `IdentityID`, `Token string`, `IssuedAt`, `RevokedAt *time.Time` |

`CredentialType` is a string enum: `"password"` / `"service_token"`.

**Mappers** (`toIdentity`, `toCredential`, `toRefreshToken`, `toServiceToken`) convert sqlc-generated structs to domain types using `helpers.FromNullableTime` for nullable timestamps.

## Store (`store.go`)

```go
store := repository.NewStore(pool)
store.Identities    // IdentityRepository
store.Credentials   // CredentialRepository
store.RefreshTokens // RefreshTokenRepository
store.ServiceTokens // ServiceTokenRepository
```

`store.WithTx(ctx, func(tx *Store) error)` — creates a transaction-scoped store where all repositories share the same `pgx.Tx`.

## Repository Interfaces

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
| `GetByHash(ctx, hash)` | Lookup by token hash; only returns non-revoked, non-expired |
| `GetActiveByIdentity(ctx, identityID)` | All non-revoked, non-expired tokens |
| `Revoke(ctx, tokenID)` | Sets `revoked_at = now()` |
| `RevokeAllByIdentity(ctx, identityID)` | Logout-all / password change |
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

## Error Mapping

All repositories use `helpers.MapError(err)`:
- `pgx.ErrNoRows` → `apierror.ErrNotFound`
- Unique violation (23505) → `apierror.ErrConflict`

## SQL Queries

Query files live in `sqlc/queries/` and follow sqlc naming conventions (`-- name: MethodName :one|:many|:exec`). Key patterns:
- Refresh token lookups filter `WHERE revoked_at IS NULL AND expires_at > now()` to exclude invalid tokens at the SQL level
- Service token lookups filter `WHERE revoked_at IS NULL`
- `RevokeAll*` operations update only non-revoked rows (`WHERE revoked_at IS NULL`)
