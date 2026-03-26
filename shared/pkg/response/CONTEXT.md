# response

Package `response` provides standard HTTP response envelope helpers. All handlers write responses through this package — no handler writes directly to `http.ResponseWriter` or calls `json.NewEncoder` by itself.

## Files

```
response/
└── response.go    envelope types and JSON/Paginated/Error helpers
```

## Response Envelopes

All responses follow one of three shapes:

**Success:**
```json
{ "data": <any> }
```

**Paginated success:**
```json
{
  "data": <any>,
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

**Error:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "role not found"
  }
}
```

## Types

### `Meta`

```go
type Meta struct {
    Page       int `json:"page"`
    PerPage    int `json:"per_page"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
}
```

`TotalPages` is computed internally by `Paginated` using ceiling division: `total/perPage + (1 if remainder > 0)`.

## Functions

### `JSON(w http.ResponseWriter, status int, data any)`

Wraps `data` in `{"data": ...}` and writes the given HTTP status code. Used for all non-paginated successful responses.

### `Paginated(w http.ResponseWriter, status int, data any, page, perPage, total int)`

Wraps `data` in `{"data": ..., "meta": {...}}`. Computes `TotalPages` automatically. Use when the endpoint returns a list that the caller may page through.

### `Error(w http.ResponseWriter, err error, message string)`

Maps `err` to the appropriate HTTP status code and machine-readable error code via `apierror.HTTPStatus` and `apierror.Code`, then writes:

```json
{ "error": { "code": "<CODE>", "message": "<message>" } }
```

The `message` parameter is a human-readable string for clients. The `code` is always derived from the sentinel error — never from `message`.

## Internal

`write(w, status, v)` — shared private helper. Sets `Content-Type: application/json`, writes the status code, then streams JSON via `json.NewEncoder(w).Encode(v)`.

## Usage Pattern

```go
// Success
response.JSON(w, http.StatusCreated, toRoleResponse(role))

// Paginated
response.Paginated(w, http.StatusOK, items, page, perPage, total)

// Error — err is an apierror sentinel (or wraps one)
if err != nil {
    response.Error(w, err, "failed to create role")
    return
}
```

Every handler follows the pattern of calling `response.Error` and returning immediately on any error — there is no fallthrough.
