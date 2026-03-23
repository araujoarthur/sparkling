# Introduction

Package domain implements the IAM business logic layer.

It sits between the handler and the repository layers, owning all business
rules — things the database and HTTP layers should never know about.
The domain layer depends on repository interfaces, never on concrete
implementations or database types. Handlers depend on domain interfaces,
never on concrete service implementations.

Folder structure:
  domain/
  ├── iam_domain.go        shared helpers and validators
  ├── builtins.go          built-in permission name constants
  ├── roles.go             RoleService interface + implementation
  ├── permissions.go       PermissionService interface + implementation
  ├── role_permissions.go  RolePermissionService interface + implementation
  ├── principals.go        PrincipalService interface + implementation
  └── principal_roles.go   PrincipalRoleService interface + implementation

# IAM Business Rules

## Permissions

Permission names must follow the `scope:resource:action` format where:

- **scope** maps to a service (`auth`, `iam`, `profile`, `webapp`) or `global`
- **resource** is the entity being acted on within that scope
- **action** is the operation being performed

Each segment must be lowercase. Actions may be hyphenated (e.g. `assign-role`). No spaces are allowed anywhere.

Only principals with `iam:permissions:write` can create permissions.
Only principals with `iam:permissions:delete` can delete permissions.

## Roles

Role names must be lowercase with no spaces.

When a role is created, its grant permission (`iam:role-{rolename}:grant`) is
created automatically in the same transaction. If either operation fails, neither
is committed.

Only principals with `iam:roles:write` can create roles.
Only principals with `iam:roles:delete` can delete roles.
System roles cannot be mutated or deleted.

## Principals

Principals are created automatically when a user registers in the auth service.
A principal can be assigned roles while inactive, but the assignment has no effect
until the principal is reactivated.

## Principal Roles

Assigning a role to a principal requires the `iam:role-{rolename}:grant` permission
for the specific role being assigned.

A principal may revoke their own roles freely.
Revoking another principal's role requires `iam:role-{rolename}:grant` for that role.