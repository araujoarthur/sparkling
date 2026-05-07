# Auth REST Handoff

This note captures where development paused while implementing the auth REST layer.

## Current Focus

Building `services/auth/internal/handler/rest`.

Auth and IAM are internal services. Except for `/api/v1/health`, routes are protected with `token.Middleware`, meaning callers must use a service token. User-facing apps call auth on behalf of users.

## Completed In Auth REST

### `server.go`

`Server` is wired with:

- `sessions domain.SessionService`
- `accounts domain.AccountService`
- `serviceTokens domain.ServiceTokenService`
- `publicKey *rsa.PublicKey`

Routes are registered under `/api/v1`.

Current route groups:

```text
GET    /health
POST   /register
POST   /login
POST   /refresh
POST   /logout
POST   /logout-all
DELETE /identities/{id}
PUT    /identities/{id}/password
DELETE /identities/{id}/credentials/{credentialID}
POST   /service-tokens
POST   /service-tokens/rotate
```

Only `/health` is outside `token.Middleware`.

### `sessions.go`

Implemented handlers:

- `register`
- `login`
- `refresh`
- `logout`
- `logoutAll`

Current behavior:

- `register` decodes `contract.RegisterRequest`, calls `sessions.Register`, returns `201` with `contract.IdentityResponse`.
- `login` decodes `contract.LoginRequest`, calls `sessions.Login`, returns token pair.
- `refresh` decodes `contract.RefreshRequest`, calls `sessions.Refresh`, returns token pair.
- `logout` decodes `contract.LogoutRequest`, calls `sessions.Logout`, returns `204`.
- `logoutAll` uses `token.ActorFromContext`, decodes `contract.LogoutAllRequest`, rejects `uuid.Nil`, calls `sessions.LogoutAll`, returns `204`.

Note: `refresh` currently returns `contract.LoginResponse` because the fields match. Future plan is to add `contract.RefreshResponse`, because login may later return a login-attempt/session ID.

### `identities.go`

Implemented:

- `deleteIdentity`

Current behavior:

- uses `token.ActorFromContext`
- parses `{id}` with `token.ParseUUIDParam`
- calls `accounts.Delete`
- returns `204`

Remaining in this file:

- `changePassword`
- `removeCredential`

### `service_tokens.go`

Still empty:

- `issueServiceToken`
- `rotateServiceTokens`

## Shared Token Helpers

`shared/pkg/token` now hosts request/context helpers:

- `ActorFromContext`
- `ActingPrincipalFromContext`
- `FromContext`
- `ParseUUIDParam`

IAM should no longer use local lowercase `parseUUIDParam`; it has been migrated to `token.ParseUUIDParam`.

Important pattern:

```go
id, ok := token.ParseUUIDParam(w, r, "id")
if !ok {
    return
}
```

The raw return is correct because `ParseUUIDParam` already wrote the error response.

## Domain Decisions Made

`SessionService.Register` should validate blank username/password in the domain.

`SessionService.Refresh` should validate blank refresh tokens in the domain.

`SessionService.Logout` should also validate blank refresh tokens in the domain if not already done.

`SessionService.LogoutAll` now accepts:

```go
LogoutAll(ctx, callerID, identityID uuid.UUID) error
```

It supports both:

- self-service: `callerID == identityID`
- admin/service action: requires `auth:sessions:revoke`

The permission string should live as a constant in auth domain helpers.

## Documentation Added

New or updated docs:

- `docs/error-handling.md`
- `docs/response-package.md`
- `docs/token-package.md`
- `shared/pkg/token/CONTEXT.md`
- root `README.md`
- root `CONTEXT.md`

## Known Next Steps

1. Run auth compile/tests:

   ```powershell
   cd services/auth
   go test ./...
   ```

2. Finish `changePassword`.

   Suggested route:

   ```text
   PUT /api/v1/identities/{id}/password
   ```

   Handler shape:

   - actor from `token.ActorFromContext`
   - target identity from `token.ParseUUIDParam(w, r, "id")`
   - body with `old_password` and `new_password`
   - call `accounts.ChangePassword(ctx, callerID, identityID, oldPassword, newPassword)`
   - return `204`

3. Finish `removeCredential`.

   Suggested route:

   ```text
   DELETE /api/v1/identities/{id}/credentials/{credentialID}
   ```

   Handler shape:

   - actor from `token.ActorFromContext`
   - target identity from route `{id}`
   - credential from route `{credentialID}`
   - call `accounts.RemoveCredential`
   - return `204`

4. Finish service-token handlers.

   Current domain service:

   ```go
   Issue(ctx, identityID uuid.UUID) (string, error)
   Rotate(ctx) error
   ```

   Decide whether issuing/rotation should require extra permissions before exposing as HTTP endpoints, or whether these should be CLI/background-only.

5. Add handler tests after the handler shapes settle.

## Last Good Compile Signal

IAM was tested after migrating to `token.ParseUUIDParam` and passed:

```text
cd services/iam
go test ./...
```

Auth should be tested next after finishing or checking the remaining stubs.
