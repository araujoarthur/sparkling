# types

Package `types` holds shared type definitions with no business logic. It exists solely to avoid import cycles between services and shared packages.

## Files

```
types/
└── principal.go    PrincipalType string type and constants
```

## Types

### `PrincipalType`

```go
type PrincipalType string

const (
    PrincipalTypeUser    PrincipalType = "user"
    PrincipalTypeService PrincipalType = "service"
)
```

Represents the kind of entity that can be assigned roles in the IAM system.

| Value | Meaning |
|---|---|
| `"user"` | A human user with password credentials |
| `"service"` | A service account authenticating with a service token |

## Why This Package Exists

Several packages need `PrincipalType`:
- `shared/pkg/provisioner` — `Provision(ctx, id, PrincipalType)`
- `shared/pkg/token` — `Claims.PrincipalType`
- `services/iam/repository` — principal records store `PrincipalType`
- `services/iam/client` — passes `PrincipalType` when provisioning

If `PrincipalType` were defined in `token` or `provisioner`, those packages would create circular imports. Placing it here — a package with no dependencies — breaks all cycles.

## Rules

- This package must contain only type definitions and constants.
- No functions, no business logic, no imports other than the Go standard library.
- If a new cross-cutting type is needed to avoid a cycle, add it here.
