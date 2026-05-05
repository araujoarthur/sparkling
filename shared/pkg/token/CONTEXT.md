# token

Package `token` handles JWT issuance, validation, and HTTP middleware for service-to-service authentication. Algorithm: RS256 (asymmetric RSA).

## Files

```
token/
├── token.go       Claims type, Issue, Parse, sentinel errors
└── middleware.go  Middleware, FromContext, ActingPrincipalFromContext
```

## Types

### `Claims`

```go
type Claims struct {
    Subject       uuid.UUID
    PrincipalType types.PrincipalType
    IssuedAt      time.Time
    ExpiresAt     time.Time  // zero value for service tokens (no expiry)
}
```

Internal domain representation of JWT claims. Handlers and domain services use this type — they never touch `jwt.Claims` directly.

**JWT payload fields:**
- `sub` — principal UUID as string
- `principal_type` — `"user"` or `"service"`
- `iat` — issued at
- `exp` — expiry (omitted for service tokens)

## Sentinel Errors

| Error | Meaning |
|---|---|
| `ErrExpiredToken` | Token is past its `exp` claim |
| `ErrInvalidToken` | Malformed, bad signature, or algorithm mismatch |

Callers can distinguish these with `errors.Is` to decide whether to prompt a refresh or return 401.

## Constants

`UserTokenDuration = 15 * time.Minute` — the expiry window added to user token `exp` claims.

## Functions

### `Issue(claims Claims, privateKey *rsa.PrivateKey) (string, error)`

Signs and returns a JWT using RS256.

- **User tokens** (`PrincipalTypeUser`): `exp` is set to `now + 15m`
- **Service tokens** (`PrincipalTypeService`): `exp` is omitted — validity is controlled by the auth service's revocation table

Should only be called by auth-domain code. The current CLI generates RSA keys but does not issue JWTs. All other services only validate tokens.

### `Parse(tokenString string, publicKey *rsa.PublicKey) (Claims, error)`

Validates and parses a JWT.

**Security measures:**
1. **Algorithm check** — verifies the signing method is `*jwt.SigningMethodRSA` (RS256). Rejects tokens with `alg: none` or any other algorithm (prevents algorithm substitution attacks).
2. **Signature verification** — validates against the RSA public key.
3. **`iat` check** — validates the issued-at claim is not in the future.
4. **5-second leeway** — accommodates small clock drift between services.

**Returns:**
- `(Claims, nil)` on success
- `(Claims{}, ErrExpiredToken)` if `exp` is in the past
- `(Claims{}, ErrInvalidToken)` for any other validation failure

## Middleware

### `Middleware(publicKey *rsa.PublicKey) func(http.Handler) http.Handler`

Chi-compatible middleware. Attaches to all protected routes.

**Request processing:**
1. Reads `Authorization` header — returns `401` if missing or not `Bearer <token>`
2. Calls `Parse` — returns `401` on expired or invalid token
3. **Rejects user tokens** — only `PrincipalTypeService` tokens are accepted on internal service routes; user tokens return `401 "only service tokens are accepted"`
4. Reads `X-Principal-ID` header — the UUID of the human user on whose behalf the service is acting (may be empty if the service is acting for itself)
5. Injects `Claims` and acting principal ID into request context via private context keys

### `FromContext(ctx context.Context) (Claims, bool)`

Extracts `Claims` from context. Returns `false` if the middleware was not applied to the request.

### `ActingPrincipalFromContext(ctx context.Context) string`

Returns the `X-Principal-ID` header value injected by the middleware. Returns empty string if no acting principal was provided (service is acting on its own behalf).

## Context Keys

Both context keys use private unexported struct types (`contextKey{}`, `actingPrincipalKey{}`), preventing collisions with other packages that use `context.WithValue`.

## Service-to-Service Auth Flow

```
Service A (caller)
  → has a service token (non-expiring JWT issued by auth)
  → sets Authorization: Bearer <service-token>
  → sets X-Principal-ID: <user-uuid>  (optional)

Service B (receiver)
  → token.Middleware validates signature with public key
  → rejects if not a service token
  → injects Claims + acting principal into context
  → handler calls token.FromContext / ActingPrincipalFromContext
```
