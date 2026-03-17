Package repository provides the IAM data access layer.

It exposes domain-oriented interfaces for querying and mutating IAM data
(roles, permissions, principals). Implementations wrap sqlc-generated code
and translate between database types and domain types.

Nothing outside this module may import this package.
All database access must go through Store, which is the single entry point
into this layer. Use Store.WithTx to wrap multiple operations in a single
atomic transaction.

Folder structure:
  repository/
  ├── iam_repository.go               domain types, sentinel errors, helpers, mappers
  ├── store.go                        Store struct, NewStore, WithTx
  ├── roles_repository.go             RoleRepository interface + implementation
  ├── permissions_repository.go       PermissionRepository interface + implementation
  ├── role_permissions_repository.go  RolePermissionRepository interface + implementation
  ├── principals_repository.go        PrincipalRepository interface + implementation
  ├── principal_roles_repository.go   PrincipalRoleRepository interface + implementation
  ├── generate.go                     go:generate directive for sqlc code generation
  └── sqlc/
      ├── sqlc.yaml                   sqlc configuration, points at migrations for schema
      ├── schema.sql                  explicit schema declaration for sqlc
      ├── queries/                    hand-written SQL queries, one file per table
      └── generated/                  sqlc output — never edit manually

Sentinel errors:
  ErrNotFound        returned when a queried entity does not exist
  ErrConflict        returned when a unique constraint would be violated
  ErrForbidden       returned when an operation is not permitted
  ErrInvalidArgument returned when an argument fails validation

To regenerate the database code run: go generate ./...