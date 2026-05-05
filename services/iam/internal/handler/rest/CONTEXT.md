# iam/internal/handler/rest

HTTP handler layer for the IAM service. Registers routes on a Chi router, parses requests, delegates to domain services, and writes JSON responses.

## Files

```
rest/
├── server.go        Server struct, NewServer constructor, route registration, health check
├── roles.go         role + role-permission HTTP handlers
├── permissions.go   permission HTTP handlers
├── principals.go    principal + principal-role HTTP handlers
└── mappers.go       domain→contract mappers, parseUUIDParam, extractCallerID
```

## Server (`server.go`)

**Constructor:** `NewServer(publicKey, roles, permissions, rolePermissions, principals, principalRoles) *Server`

Wraps a `chi.Mux`. Implements `http.Handler` via `ServeHTTP`.

**Global middleware** (applied to all routes): `Logger`, `Recoverer`, `RequestID`, `RealIP`.

## Route Table

| Method | Path | Auth | Handler |
|---|---|---|---|
| GET | `/api/v1/health` | none | `health` |
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
| POST | `/api/v1/principals` | service token | `createPrincipal` |
| GET | `/api/v1/principals/by-external-id/{externalID}` | service token | `getPrincipalByExternalID` |
| GET | `/api/v1/principals/{id}` | service token | `getPrincipalByID` |
| DELETE | `/api/v1/principals/{id}` | service token | `deletePrincipal` |
| POST | `/api/v1/principals/{id}/activate` | service token | `activatePrincipal` |
| POST | `/api/v1/principals/{id}/deactivate` | service token | `deactivatePrincipal` |
| GET | `/api/v1/principals/{id}/permissions` | service token | `getPrincipalPermissions` |
| GET | `/api/v1/principals/{id}/roles` | service token | `listRolesByPrincipal` |
| POST | `/api/v1/principals/{id}/roles` | service token | `assignRoleToPrincipal` |
| DELETE | `/api/v1/principals/{id}/roles/{roleID}` | service token | `removeRoleFromPrincipal` |

`GET /principals/by-external-id/{externalID}` requires a `?type=user|service` query parameter.

## Helpers (`mappers.go`)

**Mappers:**
- `toRoleResponse(repository.Role) contract.RoleResponse`
- `toPermissionResponse(repository.Permission) contract.PermissionResponse`
- `toPrincipalResponse(repository.Principal) contract.PrincipalResponse`

**`parseUUIDParam(w, r, param) (uuid.UUID, bool)`** — extracts and validates a UUID path parameter. Writes `400` and returns `false` on failure.

**`extractCallerID(ctx) (uuid.UUID, error)`** — resolves the acting principal:
1. If `X-Principal-ID` header is present → parse and return it
2. Otherwise fall back to `token.Claims.Subject` (the calling service's own principal)

## Handler Pattern

Every handler follows the same structure:
```
1. Parse path params (parseUUIDParam)
2. Extract caller ID if write operation (extractCallerID)
3. Decode request body if present (json.NewDecoder)
4. Call domain service method
5. Map result to contract type
6. Write response (response.JSON or response.Error)
```
On error at any step, `response.Error` is called and the handler returns immediately.
