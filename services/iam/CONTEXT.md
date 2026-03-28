# IAM Service

Centralised role-based access control (RBAC) system. Manages roles, permissions, and principals. Enforces access control on all write operations. Fully implemented across all three layers.

**Module:** `github.com/araujoarthur/intranetbackend/services/iam`
**Default port:** `:8081`
**Entry point:** `cmd/iamd/main.go`

## Directory Structure

```
services/iam/
├── cmd/iamd/main.go              composition root
├── contract/                     HTTP request/response types (public API)
│   ├── roles.go
│   ├── permissions.go
│   ├── principals.go
│   ├── role_permissions.go
│   └── principal_roles.go
├── client/
│   └── client.go                 IAM HTTP client (used by auth service)
└── internal/
    ├── domain/
    │   ├── iam_domain.go         shared helpers + validation functions
    │   ├── builtins.go           built-in permission name constants
    │   ├── roles.go              RoleService interface + implementation
    │   ├── permissions.go        PermissionService interface + implementation
    │   ├── principals.go         PrincipalService interface + implementation
    │   ├── role_permissions.go   RolePermissionService interface + implementation
    │   └── principal_roles.go    PrincipalRoleService interface + implementation
    ├── handler/rest/
    │   ├── server.go             chi router + middleware + route registration
    │   ├── roles.go              role HTTP handlers
    │   ├── permissions.go        permission HTTP handlers
    │   ├── principals.go         principal HTTP handlers
    │   └── mappers.go            domain→contract mappers + request helpers
    └── repository/
        ├── iam_repository.go     domain types + mappers
        ├── store.go              Store — single entry point to all repos
        ├── roles_repository.go
        ├── permission_repository.go
        ├── principal_repository.go
        ├── role_permission_repository.go
        ├── principal_role_repository.go
        └── sqlc/generated/       sqlc-generated query code
```

---

## Database Schema (`iam` schema)

| Table | Primary Key | Notable Columns |
|---|---|---|
| `iam.roles` | `id uuid` | `name` (unique), `description`, `is_system bool`, `created_at`, `updated_at` |
| `iam.permissions` | `id uuid` | `name` (unique), `description`, `created_at` |
| `iam.principals` | `id uuid` | `external_id uuid`, `principal_type` (enum), `is_active bool`, `created_at`, `updated_at` |
| `iam.role_permissions` | `(role_id, permission_id)` | composite PK, no extra columns |
| `iam.principal_roles` | `(principal_id, role_id)` | `granted_by uuid FK → principals.id`, `created_at` |

System roles (`is_system = true`) cannot be updated or deleted — the SQL queries exclude them with a `WHERE is_system = false` filter so they return `ErrNotFound` rather than modifying them.

---

## Repository Layer

### Domain Types (`iam_repository.go`)

| Type | Key Fields |
|---|---|
| `Role` | `ID`, `Name`, `Description`, `IsSystem bool`, `CreatedAt`, `UpdatedAt` |
| `Permission` | `ID`, `Name`, `Description`, `CreatedAt` |
| `Principal` | `ID`, `ExternalID uuid`, `PrincipalType`, `IsActive bool`, `CreatedAt`, `UpdatedAt` |
| `PrincipalRole` | `PrincipalID`, `RoleID`, `GrantedBy uuid`, `CreatedAt` |

`PrincipalType` is a type alias for `types.PrincipalType` (`"user"` / `"service"`).

Mappers (`toRole`, `toPermission`, `toPrincipal`, `toPrincipalRole`) unwrap nullable `pgtype` fields into plain Go types.

### Store

```go
store := repository.NewStore(pool)
store.Roles           // RoleRepository
store.Permissions     // PermissionRepository
store.RolePermissions // RolePermissionRepository
store.Principals      // PrincipalRepository
store.PrincipalRoles  // PrincipalRoleRepository
```

`store.WithTx(ctx, func(tx *Store) error)` — atomic multi-step operations (e.g. creating a role + its grant permission).

### RoleRepository

| Method | Notes |
|---|---|
| `GetByID`, `GetByName` | `ErrNotFound` if missing |
| `List` | Ordered by name ascending |
| `Create(ctx, name, description, isSystem)` | `ErrConflict` on duplicate name |
| `Update(ctx, id, name, description)` | Silently excludes system roles (`ErrNotFound`) |
| `Delete(ctx, id)` | Cascades to `role_permissions`; silently excludes system roles |

### PermissionRepository

| Method | Notes |
|---|---|
| `GetByID`, `GetByName` | `ErrNotFound` if missing |
| `List` | Ordered by name ascending |
| `ListByRole(ctx, roleID)` | Permissions assigned to a role |
| `Create(ctx, name, description)` | `ErrConflict` on duplicate name |
| `Delete(ctx, id)` | Cascades to `role_permissions` |

### PrincipalRepository

| Method | Notes |
|---|---|
| `GetByID` | By IAM internal UUID |
| `GetByExternalID(ctx, externalID, type)` | By auth service identity ID + type |
| `List`, `ListByType` | Ordered by `created_at` ascending |
| `Create(ctx, externalID, type)` | `ErrConflict` if same externalID+type exists |
| `Activate`, `Deactivate` | Toggle `is_active` |
| `Delete` | Cascades to `principal_roles` |
| `GetPermissions(ctx, id)` | Full flat permission list across all roles; only returns results for active principals |

### RolePermissionRepository

| Method | Notes |
|---|---|
| `Assign(ctx, roleID, permissionID)` | Idempotent — duplicate is silently ignored |
| `Remove` | `ErrNotFound` if assignment doesn't exist |
| `RoleHasPermission` | Returns `bool` |
| `ListRolesByPermission` | Roles that have a given permission |

### PrincipalRoleRepository

| Method | Notes |
|---|---|
| `Assign(ctx, principalID, roleID, grantedBy)` | `ErrConflict` if already assigned |
| `Remove` | `ErrNotFound` if not assigned |
| `ListRolesByPrincipal` | Ordered by name ascending |
| `ListPrincipalsByRole` | Active principals only |
| `PrincipalHasRole` | Returns `bool` |

---

## Domain Layer

### Shared Helpers (`iam_domain.go`)

**Validation:**
- `validateRoleName(name)` — regex `^[a-z]+(-[a-z]+)*$` — lowercase, hyphens allowed, no spaces
- `validatePermissionName(name)` — regex `^[a-z]+:[a-z]+:[a-z]+(-[a-z]+)*$` — must be `scope:resource:action`

**Permission check:**
```go
func hasPermission(ctx, store, principalID, permissionName) (bool, error)
```
Calls `store.Principals.GetPermissions` and scans for an exact name match. Used by every mutating domain operation.

### Built-in Permission Names (`builtins.go`)

| Constant | Value |
|---|---|
| `permissionIAMRolesWrite` | `"iam:roles:write"` |
| `permissionIAMRolesDelete` | `"iam:roles:delete"` |
| `permissionIAMPermissionsWrite` | `"iam:permissions:write"` |
| `permissionIAMPermissionsDelete` | `"iam:permissions:delete"` |
| `permissionIAMRolePermissionsAssign` | `"iam:permissions:assign"` |
| `permissionIAMRolePermissionsRevoke` | `"iam:permissions:revoke"` |
| `permissionIAMPrincipalsWrite` | `"iam:principals:write"` |
| `permissionIAMPrincipalsDelete` | `"iam:principals:delete"` |

`GetGrantRolePermission(roleName string) string` — returns `"iam:role-{rolename}:grant"` for a given role name. Used to check/assign per-role grant permissions.

### RoleService (`roles.go`)

| Method | Required Permission | Notes |
|---|---|---|
| `GetByID`, `GetByName`, `List` | none | Read operations |
| `Create(ctx, callerID, name, description)` | `iam:roles:write` | Validates name format; creates role + `iam:role-{name}:grant` permission in one transaction |
| `Update(ctx, callerID, id, name, description)` | `iam:roles:write` | Validates name format; cannot update system roles |
| `Delete(ctx, callerID, id)` | `iam:roles:delete` | Deletes role + its grant permission in one transaction; cannot delete system roles |

### PermissionService (`permissions.go`)

| Method | Required Permission | Notes |
|---|---|---|
| `GetByID`, `GetByName`, `List`, `ListByRole` | none | Read operations |
| `Create(ctx, callerID, name, description)` | `iam:permissions:write` | Validates `scope:resource:action` format |
| `Delete(ctx, callerID, id)` | `iam:permissions:delete` | Cascades to role assignments |

### PrincipalService (`principals.go`)

| Method | Required Permission | Notes |
|---|---|---|
| `GetByID`, `GetByExternalID`, `List`, `ListByType`, `GetPermissions` | none | Read operations |
| `Create(ctx, externalID, type)` | none | No caller check — called by auth service on identity creation |
| `Activate`, `Deactivate` | `iam:principals:write` | Toggle `is_active` |
| `Delete` | `iam:principals:delete` | Cascades to role assignments |

### RolePermissionService (`role_permissions.go`)

| Method | Required Permission | Notes |
|---|---|---|
| `Assign(ctx, callerID, roleID, permissionID)` | `iam:permissions:assign` | Adds permission to role |
| `Remove(ctx, callerID, roleID, permissionID)` | `iam:permissions:revoke` | Removes permission from role |
| `RoleHasPermission`, `ListRolesByPermission` | none | Read operations |

### PrincipalRoleService (`principal_roles.go`)

| Method | Required Permission | Notes |
|---|---|---|
| `Assign(ctx, callerID, principalID, roleID)` | `iam:role-{rolename}:grant` | Grants role to principal; permission is role-specific |
| `Remove(ctx, callerID, principalID, roleID)` | `iam:role-{rolename}:grant` (unless `callerID == principalID`) | Self-revocation is always allowed |
| `ListRolesByPrincipal`, `ListPrincipalsByRole`, `PrincipalHasRole` | none | Read operations |

---

## Handler Layer (`internal/handler/rest`)

### Server (`server.go`)

Wraps a `chi.Mux`. Applied global middleware: `Logger`, `Recoverer`, `RequestID`, `RealIP`.

**Constructor:**
```go
rest.NewServer(publicKey, roles, permissions, rolePermissions, principals, principalRoles)
```

### Route Table

| Method | Path | Auth | Handler |
|---|---|---|---|
| GET | `/api/v1/health` | none | `health` |
| POST | `/api/v1/` | none | `createPrincipal` (**see note below**) |
| GET | `/api/v1/roles` | service token | `listRoles` |
| POST | `/api/v1/roles` | service token | `createRole` |
| GET | `/api/v1/roles/{id}` | service token | `getRoleByID` |
| PUT | `/api/v1/roles/{id}` | service token | `updateRole` |
| DELETE | `/api/v1/roles/{id}` | service token | `deleteRole` |
| GET | `/api/v1/roles/{id}/permissions` | service token | `listPermissionsByRole` |
| POST | `/api/v1/roles/{id}/permissions` | service token | `assignPermissionToRole` |
| DELETE | `/api/v1/roles/{id}/permissions/{permID}` | service token | `removePermissionFromRole` |
| GET | `/api/v1/roles/{id}/principals` | service token | `listPrincipalsByRole` |
| GET | `/api/v1/permissions` | service token | `listPermissions` |
| POST | `/api/v1/permissions` | service token | `createPermission` |
| GET | `/api/v1/permissions/{id}` | service token | `getPermissionByID` |
| DELETE | `/api/v1/permissions/{id}` | service token | `deletePermission` |
| GET | `/api/v1/permissions/{id}/roles` | service token | `listRolesByPermission` |
| GET | `/api/v1/principals` | service token | `listPrincipals` |
| GET | `/api/v1/principals/by-external-id/{externalID}` | service token | `getPrincipalByExternalID` |
| GET | `/api/v1/principals/{id}` | service token | `getPrincipalByID` |
| DELETE | `/api/v1/principals/{id}` | service token | `deletePrincipal` |
| POST | `/api/v1/principals/{id}/activate` | service token | `activatePrincipal` |
| POST | `/api/v1/principals/{id}/deactivate` | service token | `deactivatePrincipal` |
| GET | `/api/v1/principals/{id}/roles` | service token | `listRolesByPrincipal` |
| POST | `/api/v1/principals/{id}/roles` | service token | `assignRoleToPrincipal` |
| DELETE | `/api/v1/principals/{id}/roles/{roleID}` | service token | `removeRoleFromPrincipal` |

`GET /principals/by-external-id/{externalID}` requires a `?type=user|service` query parameter.

**Route mismatch — `createPrincipal`:** The server registers `createPrincipal` at `POST /api/v1/` (`r.Post("/", ...)` inside the `/api/v1` route group), but the IAM client (`client/client.go:78`) sends requests to `POST /api/v1/principals`. These paths do not match. The route is placed before `r.Use(token.Middleware(...))` in the group, so it intentionally has no auth — but the path itself needs to be reconciled with the client.

### Helpers (`mappers.go`)

`parseUUIDParam(w, r, param)` — extracts and validates a UUID path parameter; writes `400` and returns `false` on failure.

`extractCallerID(ctx)` — resolves the acting principal:
1. If `X-Principal-ID` header is present → parse and return it
2. Otherwise fall back to `token.Claims.Subject` (the calling service's own principal)

### Contract Types (`contract/`)

| File | Types |
|---|---|
| `roles.go` | `CreateRoleRequest`, `UpdateRoleRequest`, `RoleResponse` |
| `permissions.go` | `CreatePermissionRequest`, `PermissionResponse` |
| `principals.go` | `CreatePrincipalRequest`, `PrincipalResponse` |
| `role_permissions.go` | `AssignPermissionRequest` |
| `principal_roles.go` | `AssignRoleRequest` |

---

## Client Library (`client/`)

Package `iamclient` defines the `IAMClient` interface and provides an HTTP implementation (`Client`).

### IAMClient Interface (`intf.go`)

```go
type IAMClient interface {
    Provision(ctx context.Context, externalID uuid.UUID, principalType types.PrincipalType) error
    HasPermission(ctx context.Context, principalID uuid.UUID, permission string) (bool, error)
}
```

This interface replaces the former `shared/pkg/provisioner.PrincipalProvisioner`. It lives in the IAM client package so that consumers (e.g. auth service) depend on the interface without importing IAM internals.

- `Provision` — creates an IAM principal for a newly registered identity. Non-fatal failure by contract.
- `HasPermission` — checks whether a principal holds a given permission.

### HTTP Implementation (`client.go`)

**Constructor:** `iamclient.New(baseURL, token string) *Client`
- `baseURL` — e.g. `"http://iam:8081"`
- `token` — service token for the calling service

**Bug:** The HTTP client timeout is set with `10 & time.Second` (bitwise AND). Since `time.Second` is `1_000_000_000` and `10` is `0b1010`, the result is `0` — effectively no timeout. This should be `10 * time.Second`.

**`Provision(ctx, externalID, principalType)`** — posts `{"external_id": "...", "principal_type": "..."}` to `POST /api/v1/principals`. Returns an error if the response status is not `201 Created`. Note: the server currently registers `createPrincipal` at `POST /api/v1/` — see the route mismatch note in the route table above.

**`HasPermission`** — not yet implemented in `client.go`.

All requests attach `Authorization: Bearer <token>` and `Content-Type: application/json`.

---

## Startup Wiring (`cmd/iamd/main.go`)

```
godotenv.Load()
→ keyprovider.NewFileKeyProvider → publicKey
→ database.NewPool(IAM_DSN) → pool
→ repository.NewStore(pool) → store
→ domain.New*Service(store) × 5
→ rest.NewServer(publicKey, services...)
→ http.ListenAndServe(IAM_ADDR || ":8081", server)
```

**Environment variables:**
| Variable | Purpose |
|---|---|
| `IAM_DSN` | PostgreSQL connection string |
| `PUBLIC_KEY_PATH` | Path to RSA public key PEM file |
| `PRIVATE_KEY_PATH` | Passed to `FileKeyProvider` constructor (unused by IAM) |
| `IAM_ADDR` | Listen address; defaults to `:8081` |
