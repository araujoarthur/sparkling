# apierror

Package `apierror` is the single source of truth for error classification across all services. It defines sentinel errors, machine-readable error codes, and helper functions that map errors to HTTP responses. No service should define its own sentinel errors for conditions already covered here.

## Design

The intended flow is:

```
repository / domain layer
  → returns apierror sentinel (e.g. apierror.ErrNotFound)
      → handler layer calls apierror.HTTPStatus / apierror.Code
          → writes HTTP status + JSON error code to the response
```

Handlers never inspect raw `pgx` or database errors directly. All errors must be mapped to a sentinel before reaching the handler layer.

## Sentinel Errors

| Variable             | Meaning                                          |
|----------------------|--------------------------------------------------|
| `ErrNotFound`        | The requested entity does not exist              |
| `ErrConflict`        | A unique constraint would be violated            |
| `ErrForbidden`       | The caller lacks permission for the operation    |
| `ErrInvalidArgument` | The request contains invalid or missing fields   |
| `ErrUnauthorized`    | The caller has no valid identity                 |
| `ErrInvalidCredentials` | The supplied credentials are wrong            |
| `ErrInternal`        | An unexpected internal error occurred            |

## Error Codes

`ErrorCode` is a `string` type returned in API responses as a machine-readable identifier.

| Constant              | Value              |
|-----------------------|--------------------|
| `CodeNotFound`        | `NOT_FOUND`        |
| `CodeConflict`        | `CONFLICT`         |
| `CodeForbidden`       | `FORBIDDEN`        |
| `CodeInvalidArgument` | `INVALID_ARGUMENT` |
| `CodeUnauthorized`    | `UNAUTHORIZED`     |
| `CodeInvalidCredentials` | `INVALID_CREDENTIALS` |
| `CodeInternal`        | `INTERNAL_ERROR`   |

## Helper Functions

### `HTTPStatus(err error) int`

Maps a sentinel error to the appropriate HTTP status code. Unrecognized errors fall back to `500 Internal Server Error`.

| Sentinel             | HTTP Status                  |
|----------------------|------------------------------|
| `ErrNotFound`        | `404 Not Found`              |
| `ErrConflict`        | `409 Conflict`               |
| `ErrForbidden`       | `403 Forbidden`              |
| `ErrInvalidArgument` | `400 Bad Request`            |
| `ErrUnauthorized`    | `401 Unauthorized`           |
| `ErrInvalidCredentials` | `401 Unauthorized`        |
| anything else        | `500 Internal Server Error`  |

### `Code(err error) ErrorCode`

Maps a sentinel error to its `ErrorCode`. Unrecognized errors fall back to `CodeInternal`.

## Usage Example

```go
// domain / repository layer
func (r *repo) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
    var u User
    err := r.db.QueryRow(ctx, query, id).Scan(&u.ID, &u.Name)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, apierror.ErrNotFound
    }
    if err != nil {
        return nil, apierror.ErrInternal
    }
    return &u, nil
}

// handler layer
func (h *handler) GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.svc.GetUser(r.Context(), id)
    if err != nil {
        w.WriteHeader(apierror.HTTPStatus(err))
        json.NewEncoder(w).Encode(map[string]string{
            "error": string(apierror.Code(err)),
        })
        return
    }
    json.NewEncoder(w).Encode(user)
}
```

## File Structure

```
apierror/
└── apierror.go    sentinel errors, ErrorCode type, HTTPStatus and Code functions
```
