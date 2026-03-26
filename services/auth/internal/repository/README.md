# Auth Repository Layer

Package repository provides the auth data access layer.

It exposes domain-oriented interfaces for querying and mutating auth data
(identities, credentials, refresh tokens, service tokens). Implementations
wrap sqlc-generated code and translate between database types and domain types.

Nothing outside this module may import this package.
All database access must go through Store, which is the single entry point
into this layer. Use Store.WithTx to wrap multiple operations in a single
atomic transaction.

## Folder structure

```
repository/
├── auth_repository.go           domain types and mappers
├── store.go                     Store struct, NewStore, WithTx
├── identity_repository.go       IdentityRepository interface + implementation
├── credential_repository.go     CredentialRepository interface + implementation
├── refresh_token_repository.go  RefreshTokenRepository interface + implementation
├── service_token_repository.go  ServiceTokenRepository interface + implementation
├── generate.go                  go:generate directive for sqlc code generation
└── sqlc/
    ├── sqlc.yaml                sqlc configuration, points at migrations for schema
    ├── schema.sql               explicit schema declaration for sqlc
    ├── queries/                 hand-written SQL queries, one file per table
    └── generated/               sqlc output — never edit manually
```

## Domain types

| Type | Description |
| --- | --- |
| `Identity` | Stable UUID anchor for any entity in the system |
| `CredentialType` | Enum: `password` or `service_token` |
| `Credential` | A single authentication method linked to an identity |
| `RefreshToken` | Long-lived token used to obtain new access tokens |
| `ServiceToken` | Non-expiring token issued to service accounts |

## Sentinel errors

Sentinel errors are sourced directly from `shared/pkg/apierror`:

| Error | Condition |
| --- | --- |
| `ErrNotFound` | The requested entity does not exist |
| `ErrConflict` | A unique constraint would be violated |
| `ErrForbidden` | The operation is not permitted |
| `ErrInvalidArgument` | An argument fails validation |

## To regenerate the database code

```bash
go generate ./...
```