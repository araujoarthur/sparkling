# helpers

Package `helpers` provides low-level utilities for working with `pgx` — type conversions and database error mapping. Used exclusively in repository layers; nothing above the repository layer should import this package.

## Files

```
helpers/
└── pgx.go    PgxText, MapError, FromNullableTime
```

## Functions

### `PgxText(s string) pgtype.Text`

Converts a Go `string` to `pgtype.Text`.

- Non-empty string → `{String: s, Valid: true}`
- Empty string → `{String: "", Valid: false}` (stored as `NULL` in the database)

Use this for any nullable `text` column. Never pass an empty string to a nullable column directly — use this helper to ensure `NULL` semantics.

### `MapError(err error) error`

Translates low-level pgx errors into `apierror` sentinels.

| Input | Output |
|---|---|
| `nil` | `nil` |
| `pgx.ErrNoRows` | `apierror.ErrNotFound` |
| PostgreSQL error code `23505` (unique violation) | `apierror.ErrConflict` |
| Any other error | returned unchanged |

This function is the boundary between pgx and the apierror system. Every repository method wraps its error with `helpers.MapError` before returning:

```go
row, err := r.q.GetRoleByID(ctx, id)
if err != nil {
    return Role{}, fmt.Errorf("RoleRepository.GetByID: %w", helpers.MapError(err))
}
```

The `fmt.Errorf` with `%w` preserves the sentinel for `errors.Is` checks upstream.

### `FromNullableTime(t pgtype.Timestamptz) *time.Time`

Converts a `pgtype.Timestamptz` to `*time.Time`.

- `t.Valid == true` → returns `&t.Time`
- `t.Valid == false` (database `NULL`) → returns `nil`

Use in repository mappers for nullable timestamp columns (e.g. `revoked_at`, `last_used_at`).

## Usage Pattern

All three functions are used together in repository mapper functions:

```go
func toCredential(g *generated.AuthCredential) Credential {
    return Credential{
        SecretHash: g.SecretHash.String,           // pgtype.Text → string
        LastUsedAt: helpers.FromNullableTime(g.LastUsedAt), // nullable time
    }
}

func (r *repo) Create(ctx context.Context, name, description string) (Role, error) {
    row, err := r.q.CreateRole(ctx, &generated.CreateRoleParams{
        Description: helpers.PgxText(description), // string → pgtype.Text
    })
    if err != nil {
        return Role{}, fmt.Errorf("...: %w", helpers.MapError(err))
    }
    return toRole(row), nil
}
```
