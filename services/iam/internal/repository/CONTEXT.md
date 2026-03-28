# iam/internal/repository

Data access layer for the IAM service. Wraps sqlc-generated queries with domain types, error mapping, and transaction support.

## Files

```
repository/
├── iam_repository.go              domain types (Role, Permission, Principal, PrincipalRole) + mappers
├── store.go                       Store — single entry point to all repos + WithTx
├── roles_repository.go            RoleRepository interface + implementation
├── permission_repository.go       PermissionRepository interface + implementation
├── principal_repository.go        PrincipalRepository interface + implementation
├── role_permission_repository.go  RolePermissionRepository interface + implementation
├── principal_role_repository.go   PrincipalRoleRepository interface + implementation
└── sqlc/
    ├── queries/                   hand-written SQL query files
    └── generated/                 sqlc-generated Go code (do not edit)
```

## Domain Types (`iam_repository.go`)

| Type | Key Fields |
|---|---|
| `Role` | `ID uuid`, `Name`, `Description`, `IsSystem bool`, `CreatedAt`, `UpdatedAt` |
| `Permission` | `ID uuid`, `Name`, `Description`, `CreatedAt` |
| `Principal` | `ID uuid`, `ExternalID uuid`, `PrincipalType`, `IsActive bool`, `CreatedAt`, `UpdatedAt` |
| `PrincipalRole` | `PrincipalID`, `RoleID`, `GrantedBy uuid`, `CreatedAt` |

`PrincipalType` is a type alias for `types.PrincipalType` (`"user"` / `"service"`).

**Mappers** (`toRole`, `toPermission`, `toPrincipal`, `toPrincipalRole`) convert sqlc-generated structs to domain types, unwrapping nullable `pgtype` fields.

## Store (`store.go`)

```go
store := repository.NewStore(pool)
store.Roles           // RoleRepository
store.Permissions     // PermissionRepository
store.RolePermissions // RolePermissionRepository
store.Principals      // PrincipalRepository
store.PrincipalRoles  // PrincipalRoleRepository
```

`store.WithTx(ctx, func(tx *Store) error)` — creates a transaction-scoped store where all repositories share the same `pgx.Tx`. Commits on success, rolls back on error.

## Repository Interfaces

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
| `GetByExternalID(ctx, externalID, type)` | By auth identity ID + type |
| `List`, `ListByType` | Ordered by `created_at` ascending |
| `Create(ctx, externalID, type)` | `ErrConflict` if same externalID+type exists |
| `Activate`, `Deactivate` | Toggle `is_active` |
| `Delete` | Cascades to `principal_roles` |
| `GetPermissions(ctx, id)` | Flat permission list across all roles; only active principals |

### RolePermissionRepository

| Method | Notes |
|---|---|
| `Assign(ctx, roleID, permissionID)` | Idempotent — duplicate silently ignored |
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

## Error Mapping

Every repository method wraps errors through `helpers.MapError(err)`:
- `pgx.ErrNoRows` → `apierror.ErrNotFound`
- Unique violation (PostgreSQL 23505) → `apierror.ErrConflict`
- All other errors pass through unchanged

## Known Issues

- `Deactivate` in `principal_repository.go:143` wraps the error as `"PrincipalRepository.Activate"` — copy-paste error in the message.
