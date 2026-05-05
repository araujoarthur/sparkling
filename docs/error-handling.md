# Error Handling

This project uses a deliberate error translation flow:

```text
Postgres / pgx / sqlc error
    -> repository maps low-level error to shared sentinel error
    -> domain wraps or creates business-rule errors
    -> handler passes error to response.Error
    -> apierror maps sentinel error to HTTP status + machine-readable code
```

The important rule is that HTTP handlers should not inspect database errors, pgx errors, or sqlc errors. By the time an error reaches the handler, it should either contain one of the shared `apierror` sentinel errors or be treated as an internal error.

## Shared Sentinel Errors

The source of truth is `shared/pkg/apierror`.

Current sentinels:

| Sentinel | HTTP status | Response code |
|---|---:|---|
| `apierror.ErrNotFound` | `404` | `NOT_FOUND` |
| `apierror.ErrConflict` | `409` | `CONFLICT` |
| `apierror.ErrForbidden` | `403` | `FORBIDDEN` |
| `apierror.ErrInvalidArgument` | `400` | `INVALID_ARGUMENT` |
| `apierror.ErrUnauthorized` | `401` | `UNAUTHORIZED` |
| `apierror.ErrInvalidCredentials` | `401` | `INVALID_CREDENTIALS` |
| `apierror.ErrInternal` | `500` | `INTERNAL_ERROR` |

`apierror.HTTPStatus(err)` maps an error to an HTTP status. `apierror.Code(err)` maps an error to the machine-readable response code.

Both functions use `errors.Is`, so wrapping is allowed and expected.

```go
return fmt.Errorf("RoleService.Create: %w", apierror.ErrForbidden)
```

The handler can still classify this as `403 FORBIDDEN`.

## Repository Layer

Repositories are responsible for translating low-level persistence errors.

Use `helpers.MapError(err)` before wrapping repository errors:

```go
row, err := r.q.GetRoleByID(ctx, id)
if err != nil {
	return Role{}, fmt.Errorf("RoleRepository.GetByID: %w", helpers.MapError(err))
}
```

`helpers.MapError` currently translates:

| Low-level error | Project error |
|---|---|
| `pgx.ErrNoRows` | `apierror.ErrNotFound` |
| Postgres unique violation `23505` | `apierror.ErrConflict` |
| anything else | original error |

This means repository methods can document and expose project-level errors, not database-driver details.

Do this:

```go
return Credential{}, fmt.Errorf("CredentialRepository.Create: %w", helpers.MapError(err))
```

Avoid this:

```go
return Credential{}, err
```

Avoid this in handlers or domain services:

```go
if errors.Is(err, pgx.ErrNoRows) { ... }
```

Only repositories should know about pgx/sqlc error shapes.

## Domain Layer

Domain services do two things with errors:

1. Preserve repository classifications by wrapping with `%w`.
2. Create business-rule errors directly using `apierror`.

Example: permission failure is a domain rule, so the domain layer returns `apierror.ErrForbidden`.

```go
if !allowed {
	return repository.Role{}, apierror.ErrForbidden
}
```

Example: invalid input is a domain rule, so validation helpers should return `apierror.ErrInvalidArgument`, usually wrapped by the service:

```go
if err := validateRoleName(name); err != nil {
	return repository.Role{}, fmt.Errorf("RoleService.Create: %w", err)
}
```

Example: unknown login credentials intentionally translate `ErrNotFound` into `ErrInvalidCredentials` so the API does not reveal whether the username exists.

```go
credential, err := s.store.Credentials.GetByTypeAndIdentifier(ctx, repository.CredentialTypePassword, username)
if err != nil {
	if errors.Is(err, apierror.ErrNotFound) {
		return LoginResult{}, apierror.ErrInvalidCredentials
	}
	return LoginResult{}, fmt.Errorf("SessionService.Login [fetch]: %w", err)
}
```

Always wrap with `%w`, not `%v`, when adding context:

```go
return fmt.Errorf("AccountService.Delete: %w", err)
```

Using `%v` would destroy the sentinel chain and cause the handler to return `500 INTERNAL_ERROR`.

## Handler Layer

Handlers should be thin. They decode requests, parse path/query values, call domain services, and write responses.

For domain/repository errors, call:

```go
response.Error(w, err, "failed to create role")
```

`response.Error` handles both translations:

```text
error -> apierror.HTTPStatus(err) -> HTTP status
error -> apierror.Code(err)       -> JSON error.code
```

The JSON error envelope is:

```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "failed to create role"
  }
}
```

Handlers create `apierror` values directly only for HTTP/request parsing problems that happen before the domain layer runs.

Examples:

```go
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	response.Error(w, apierror.ErrInvalidArgument, "invalid request body")
	return
}
```

```go
parsed, err := uuid.Parse(raw)
if err != nil {
	response.Error(w, apierror.ErrInvalidArgument, "invalid id format")
	return
}
```

Authentication-context extraction also maps to `apierror.ErrUnauthorized` in handlers:

```go
callerID, err := extractCallerID(r.Context())
if err != nil {
	response.Error(w, apierror.ErrUnauthorized, err.Error())
	return
}
```

## Response Messages

The `message` passed to `response.Error` is the public client-facing message. The wrapped Go error is used only for status/code classification.

This keeps internal details out of API responses:

```go
response.Error(w, err, "failed to delete role")
```

Even if `err` is:

```text
RoleService.Delete: fetching grant permission: PermissionRepository.GetByName: not found
```

the client sees:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "failed to delete role"
  }
}
```

## Adding A New Error Type

When a new project-level error classification is needed:

1. Add a sentinel in `shared/pkg/apierror`.
2. Add a machine-readable `ErrorCode`.
3. Update `apierror.HTTPStatus`.
4. Update `apierror.Code`.
5. Make repositories or domain services return that sentinel.
6. Keep handlers unchanged; they should continue calling `response.Error`.

Do not define service-local sentinel errors for conditions already covered by `shared/pkg/apierror`.

## Layer Rules

Repository:

- Knows about pgx/sqlc errors.
- Calls `helpers.MapError`.
- Wraps with operation context using `%w`.

Domain:

- Knows about `apierror` sentinels.
- Converts business-rule failures into `apierror` values.
- May intentionally translate one sentinel into another, such as `ErrNotFound` to `ErrInvalidCredentials`.
- Wraps errors with `%w`.

Handler:

- Knows about request parsing and response writing.
- Does not know about pgx/sqlc.
- Calls `response.Error`.
- Creates `apierror.ErrInvalidArgument` or `apierror.ErrUnauthorized` for request-level failures.

## Quick Checklist

Before finishing a repository method:

- Did every sqlc call pass errors through `helpers.MapError`?
- Did every wrapper use `%w`?

Before finishing a domain method:

- Are business-rule failures using `apierror`?
- Are login/auth flows avoiding account enumeration?
- Are repository errors wrapped without losing their sentinel?

Before finishing a handler:

- Does every error response go through `response.Error`?
- Are parse/decode failures mapped to `ErrInvalidArgument`?
- Are auth-context failures mapped to `ErrUnauthorized`?
- Is the response message public-safe?
