# Response Package

`shared/pkg/response` is the standard way HTTP handlers write JSON responses.

Handlers should use this package instead of hand-writing JSON envelopes. This keeps every service response shaped the same way and keeps error-to-status translation in one place.

## Success Envelope

Use `response.JSON` for normal success responses:

```go
response.JSON(w, http.StatusOK, toRoleResponse(role))
```

The response body is always wrapped in `data`:

```json
{
  "data": {
    "id": "4b59a4da-8b2e-4729-89f2-d8ab1fba5d89",
    "name": "admin"
  }
}
```

For list responses, `data` holds the array:

```go
response.JSON(w, http.StatusOK, roles)
```

```json
{
  "data": [
    { "name": "admin" },
    { "name": "member" }
  ]
}
```

## Paginated Envelope

Use `response.Paginated` when a list endpoint supports pagination:

```go
response.Paginated(w, http.StatusOK, items, page, perPage, total)
```

The response body includes `data` and `meta`:

```json
{
  "data": [],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 45,
    "total_pages": 3
  }
}
```

`total_pages` is calculated by the response package from `total` and `perPage`.

Current IAM handlers mostly use non-paginated list responses. Add `Paginated` when the domain/repository layer supports paginated queries.

## Error Envelope

Use `response.Error` for all structured errors:

```go
if err != nil {
	response.Error(w, err, "failed to create role")
	return
}
```

The body is:

```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "failed to create role"
  }
}
```

`response.Error` calls:

```text
apierror.HTTPStatus(err) -> HTTP status
apierror.Code(err)       -> error.code
```

Because `apierror` uses `errors.Is`, wrapped errors still classify correctly:

```go
return fmt.Errorf("RoleService.Create: %w", apierror.ErrForbidden)
```

See `docs/error-handling.md` for the full repository/domain/handler error flow.

## Public Messages

The `message` argument is public API text. It should be short, stable, and safe to show to clients.

Good:

```go
response.Error(w, err, "failed to delete role")
```

Avoid leaking internal details:

```go
response.Error(w, err, err.Error())
```

One exception already present in IAM is auth-context parsing:

```go
response.Error(w, apierror.ErrUnauthorized, err.Error())
```

That is acceptable only when the error is already a public request problem, such as an invalid `X-Principal-ID` header or missing claims.

## Request Parsing Errors

Handlers should create `apierror.ErrInvalidArgument` for malformed request bodies, path parameters, and query parameters:

```go
var req contract.CreateRoleRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	response.Error(w, apierror.ErrInvalidArgument, "invalid request body")
	return
}
```

```go
id, err := uuid.Parse(raw)
if err != nil {
	response.Error(w, apierror.ErrInvalidArgument, "invalid id format")
	return
}
```

Helpers such as `parseUUIDParam` should write the response themselves and return a boolean so the handler can stop cleanly:

```go
roleID, ok := parseUUIDParam(w, r, "id")
if !ok {
	return
}
```

## No Content Responses

For successful operations that intentionally return no body, write `204 No Content` directly:

```go
w.WriteHeader(http.StatusNoContent)
```

This is appropriate for delete, assignment, revoke, activate/deactivate, and other command-style endpoints where the client does not need a response body.

Do not use `response.JSON` with `nil` for these routes. A `204` response should not include a JSON body.

## Handler Pattern

A typical handler should follow this shape:

```go
func (s *Server) createRole(w http.ResponseWriter, r *http.Request) {
	callerID, err := extractCallerID(r.Context())
	if err != nil {
		response.Error(w, apierror.ErrUnauthorized, err.Error())
		return
	}

	var req contract.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.ErrInvalidArgument, "invalid request body")
		return
	}

	role, err := s.roles.Create(r.Context(), callerID, req.Name, req.Description)
	if err != nil {
		response.Error(w, err, "failed to create role")
		return
	}

	response.JSON(w, http.StatusCreated, toRoleResponse(role))
}
```

The responsibilities stay separated:

- handler parses HTTP details
- domain returns business errors
- `response.Error` translates errors into status and JSON code
- `response.JSON` wraps success data consistently

## Rules Of Thumb

- Use `response.JSON` for success bodies.
- Use `response.Paginated` for paginated success bodies.
- Use `response.Error` for all JSON error responses.
- Use raw `w.WriteHeader(http.StatusNoContent)` for successful empty responses.
- Keep error messages public-safe.
- Do not manually build `{ "data": ... }` or `{ "error": ... }` in handlers.
- Do not call `apierror.HTTPStatus` or `apierror.Code` directly from handlers unless you are extending the response package itself.
