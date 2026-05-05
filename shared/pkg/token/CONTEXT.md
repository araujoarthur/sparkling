# token

Package `token` owns JWT issuance, JWT validation, service-token middleware, and token-related request/context helpers.

**Package:** `github.com/araujoarthur/intranetbackend/shared/pkg/token`

## Purpose

This package is the shared authentication boundary for internal services.

It provides:

- RS256 JWT issuance and parsing.
- Convenience factories for user and service tokens.
- Chi middleware for service-to-service authentication.
- Request context helpers for claims, acting principal, and effective actor resolution.
- UUID URL parameter parsing for handlers that already depend on token/request authentication helpers.

Auth is the only service that should issue JWTs at runtime. Other services validate service tokens through `Middleware`.

## Files

```text
token/
├── actor.go       ActingPrincipalFromContext, ActorFromContext, ParseUUIDParam
├── context.go     contextKey, FromContext
├── factories.go   IssueUserToken, IssueServiceToken
├── middleware.go  service-token-only HTTP middleware
└── token.go       Claims, Issue, Parse, token errors
```

## Claims

```go
type Claims struct {
    Subject       uuid.UUID
    PrincipalType types.PrincipalType
    IssuedAt      time.Time
    ExpiresAt     time.Time
}
```

`Claims` is the internal representation of JWT claims.

JWT payload fields:

- `sub` — principal UUID as a string.
- `principal_type` — `"user"` or `"service"`.
- `iat` — issued-at timestamp.
- `exp` — expiry timestamp, present only for user tokens.

Service tokens intentionally do not expire through JWT `exp`; their validity is controlled by auth-owned revocation state.

## Token Issuance

### `Issue(claims, privateKey) (string, error)`

Signs claims with RS256.

- User tokens receive `exp = now + UserTokenDuration`.
- Service tokens omit `exp`.

Call this only from auth-domain code or tightly controlled administrative tooling.

### `IssueUserToken(identityID, privateKey) (string, error)`

Factory for a short-lived human user token.

Current lifetime:

```go
const UserTokenDuration = 15 * time.Minute
```

### `IssueServiceToken(principalID, privateKey) (string, error)`

Factory for a non-expiring service token.

The `principalID` should identify the service account principal/identity.

## Token Parsing

### `Parse(tokenString, publicKey) (Claims, error)`

Validates and parses a JWT using the public key.

Security behavior:

- rejects non-RSA signing methods
- validates the RSA signature
- validates `iat`
- uses a 5-second clock leeway
- maps expired tokens to `ErrExpiredToken`
- maps malformed, invalid, or bad-signature tokens to `ErrInvalidToken`

Sentinel errors:

| Error | Meaning |
|---|---|
| `ErrExpiredToken` | Token is past its `exp` claim |
| `ErrInvalidToken` | Malformed token, invalid signature, invalid claims, or algorithm mismatch |

Use `errors.Is` when checking these errors.

## Middleware

### `Middleware(publicKey) func(http.Handler) http.Handler`

Chi-compatible middleware for protected internal service routes.

Behavior:

1. Reads `Authorization`.
2. Requires `Bearer <token>`.
3. Parses and validates the JWT.
4. Rejects user tokens.
5. Accepts service tokens only.
6. Reads `X-Principal-ID`.
7. Stores validated `Claims` and raw acting principal ID in request context.

This is why auth and IAM route groups should protect all non-health routes with `token.Middleware`.

## Context Helpers

### `FromContext(ctx) (Claims, bool)`

Returns the validated service-token claims stored by `Middleware`.

Returns `false` when middleware was not applied.

### `ActingPrincipalFromContext(ctx) string`

Returns the raw acting principal value stored by `Middleware`.

This value comes from the `X-Principal-ID` header and may be empty. It is intentionally returned as a string because callers should not trust it as a UUID until it is parsed.

### `ActorFromContext(ctx) (uuid.UUID, error)`

Returns the effective actor for an authenticated request:

1. If `X-Principal-ID` was present, parse and return it.
2. Otherwise, return the service-token subject from `Claims`.

Use this in handlers before calling domain services that require a `callerID`.

Examples:

- IAM role/permission/principal mutations.
- Auth account deletion.
- Auth password changes.
- Logout-all flows that can be self-service or admin-triggered.

If `X-Principal-ID` is malformed, the function returns an error mentioning the header. If claims are missing, it returns an error indicating middleware was not applied.

## URL Helpers

### `ParseUUIDParam(w, r, param) (uuid.UUID, bool)`

Extracts a chi route parameter and parses it as a UUID.

On failure, it writes the error response itself through `response.Error`:

- missing param -> `apierror.ErrInvalidArgument`
- invalid UUID -> `apierror.ErrInvalidArgument`

Handlers should stop immediately when `ok == false`:

```go
id, ok := token.ParseUUIDParam(w, r, "id")
if !ok {
    return
}
```

The raw `return` is correct because `ParseUUIDParam` has already written the HTTP response.

## Service-to-Service Flow

```text
calling service
  -> sends Authorization: Bearer <service-token>
  -> optionally sends X-Principal-ID: <user-principal-uuid>

receiving service
  -> token.Middleware validates service token
  -> rejects user tokens
  -> stores Claims in context
  -> stores X-Principal-ID in context
  -> handler calls ActorFromContext for effective caller
```

If `X-Principal-ID` is absent, the service token subject is the actor.

## Layer Rules

Handlers may import this package for:

- middleware
- actor/caller extraction
- claims extraction
- URL UUID parsing

Domain code may use token issuance/parsing only when token business logic belongs there, as in auth session and service-token services.

Repository code should not import this package.
