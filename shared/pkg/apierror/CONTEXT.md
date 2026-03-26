# apierror

Package `apierror` is the single source of truth for error classification across all services. It defines sentinel errors, machine-readable error codes, and helper functions that map errors to HTTP responses.

No service should define its own sentinel errors for conditions already covered here.

## File

```
apierror/
└── apierror.go    sentinel errors, ErrorCode type, HTTPStatus and Code functions
```

## Sentinel Errors

Defined with `errors.New`. Returned by repository and domain layers. Handlers never return raw database errors — all errors are mapped to these before reaching the handler.

| Variable | Message | HTTP Status |
|---|---|---|
| `ErrNotFound` | `"not found"` | 404 |
| `ErrConflict` | `"already exists"` | 409 |
| `ErrForbidden` | `"forbidden"` | 403 |
| `ErrInvalidArgument` | `"invalid argument"` | 400 |
| `ErrUnauthorized` | `"unauthorized"` | 401 |
| `ErrInternal` | `"internal error"` | 500 |

## ErrorCode Type

`type ErrorCode string` — machine-readable string returned in API error responses.

| Constant | Value |
|---|---|
| `CodeNotFound` | `"NOT_FOUND"` |
| `CodeConflict` | `"CONFLICT"` |
| `CodeForbidden` | `"FORBIDDEN"` |
| `CodeInvalidArgument` | `"INVALID_ARGUMENT"` |
| `CodeUnauthorized` | `"UNAUTHORIZED"` |
| `CodeInternal` | `"INTERNAL_ERROR"` |

## Functions

### `HTTPStatus(err error) int`

Maps a sentinel to an HTTP status code using `errors.Is`. Falls back to `500` for unrecognised errors.

### `Code(err error) ErrorCode`

Maps a sentinel to its `ErrorCode`. Falls back to `CodeInternal` for unrecognised errors.

## Usage Pattern

```
repository layer
  → helpers.MapError(pgxErr)       — translates pgx errors to sentinels
      → returns apierror.ErrNotFound / ErrConflict / etc.

domain layer
  → checks permission, returns apierror.ErrForbidden if denied

handler layer
  → response.Error(w, err, "message")
      → internally calls apierror.HTTPStatus(err) and apierror.Code(err)
```

Handlers never inspect raw `pgx` or database errors. All errors must be mapped before they reach the handler layer.
