# iam/contract

Package `contract` defines HTTP request and response types for the IAM service's public API. These types are used by both the handler layer (to parse requests and serialize responses) and the client library (to construct requests).

## Files

```
contract/
├── roles.go              CreateRoleRequest, UpdateRoleRequest, RoleResponse
├── permissions.go        CreatePermissionRequest, PermissionResponse
├── principals.go         CreatePrincipalRequest, PrincipalResponse
├── role_permissions.go   AssignPermissionRequest
└── principal_roles.go    AssignRoleRequest
```

## Request Types

| Type | Used by | Fields |
|---|---|---|
| `CreateRoleRequest` | `POST /roles` | `Name`, `Description` |
| `UpdateRoleRequest` | `PUT /roles/{id}` | `Name`, `Description` |
| `CreatePermissionRequest` | `POST /permissions` | `Name`, `Description` |
| `CreatePrincipalRequest` | `POST /principals` | `ExternalID uuid`, `PrincipalType` |
| `AssignPermissionRequest` | `POST /roles/{id}/permissions` | `PermissionID uuid` |
| `AssignRoleRequest` | `POST /principals/{id}/roles` | `RoleID uuid` |

All request types use `json` struct tags for deserialization.

## Response Types

| Type | Fields |
|---|---|
| `RoleResponse` | `ID`, `Name`, `Description`, `IsSystem`, `CreatedAt`, `UpdatedAt` |
| `PermissionResponse` | `ID`, `Name`, `Description`, `CreatedAt` |
| `PrincipalResponse` | `ID`, `ExternalID`, `PrincipalType`, `IsActive`, `CreatedAt`, `UpdatedAt` |

All response types use `json` struct tags for serialization.

## Design

- Contract types are intentionally separate from repository domain types. Handlers map between them using mapper functions in `internal/handler/rest/mappers.go`.
- This package is the only one imported by both `internal/handler/rest` and `client/`, making it the shared API surface.
