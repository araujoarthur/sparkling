# iam/internal/domain

Business logic layer for the IAM service. Implements five domain services that enforce validation, permission checks, and transactional consistency on top of the repository layer.

## Files

```
domain/
├── iam_domain.go         shared helpers: validation regexes, hasPermission()
├── builtins.go           built-in permission name constants, GetGrantRolePermission()
├── roles.go              RoleService interface + implementation
├── permissions.go        PermissionService interface + implementation
├── principals.go         PrincipalService interface + implementation
├── role_permissions.go   RolePermissionService interface + implementation
└── principal_roles.go    PrincipalRoleService interface + implementation
```

## Shared Helpers (`iam_domain.go`)

**Validation:**
- `validateRoleName(name)` — regex `^[a-z]+(-[a-z]+)*$` — lowercase, hyphens allowed, no spaces
- `validatePermissionName(name)` — regex `^[a-z]+:[a-z]+:[a-z]+(-[a-z]+)*$` — must be `scope:resource:action`

**Permission check:**
```go
func hasPermission(ctx, store, principalID, permissionName) (bool, error)
```
Fetches all permissions for the principal via `store.Principals.GetPermissions` and scans for an exact name match. Called by every mutating domain operation.

## Built-in Permission Constants (`builtins.go`)

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

`GetGrantRolePermission(roleName string) string` — returns `"iam:role-{rolename}:grant"`. Used for per-role grant permission checks.

## Domain Services

All services receive `*repository.Store` and expose an interface. Read operations require no permissions; write operations check caller permissions first.

### RoleService (`roles.go`)

| Method | Required Permission | Notes |
|---|---|---|
| `GetByID`, `GetByName`, `List` | none | Read operations |
| `Create(ctx, callerID, name, description)` | `iam:roles:write` | Validates name; creates role + `iam:role-{name}:grant` permission in one transaction |
| `Update(ctx, callerID, id, name, description)` | `iam:roles:write` | Validates name; cannot update system roles |
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
| `Create(ctx, externalID, type)` | none | No caller check — called by auth service on registration |
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

## Error Handling

- Permission failure → `apierror.ErrForbidden`
- Validation failure → `fmt.Errorf(...)` (wraps as `ErrInvalidArgument` at handler level via error message pattern)
- All errors are wrapped with `"ServiceName.Method: %w"` for traceability
