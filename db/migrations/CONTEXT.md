# db/migrations

Goose SQL migrations organized by service. Managed via `inetbctl db migrate` and `inetbctl db bootstrap`.

## Structure

```
migrations/
├── global/              cross-service bootstrap (schemas, roles, privileges)
│   └── 001_bootstrap.sql
├── auth/                auth service schema
│   ├── 001_create_identities.sql
│   ├── 002_create_credentials.sql
│   ├── 003_create_refresh_tokens.sql
│   └── 004_create_service_tokens.sql
└── iam/                 IAM service schema
    ├── 001_create_roles.sql
    ├── 002_create_permissions.sql
    ├── 003_create_role_permissions.sql
    ├── 004_create_principals.sql
    └── 005_create_principal_roles.sql
```

## Execution Order

1. **`global/`** — run first via `inetbctl db bootstrap`. Creates schemas (`auth`, `iam`, `global`), database roles (`intranetbackend_auth`, `intranetbackend_iam`), and grants default privileges.
2. **`auth/` and `iam/`** — run after bootstrap via `inetbctl db migrate up`. Each service's migrations operate within their own schema.

## Versioning

Each service has its own goose version table: `{service}.goose_db_version` (e.g. `auth.goose_db_version`, `iam.goose_db_version`). This means migration state for one service never affects another.

The global bootstrap uses `db/migrations/global` as a standalone goose directory with its own version table.

## Key Design Decisions

- **Owner DSN** — all migrations run under the `OWNER_DSN` (admin user), not the service-specific DSNs. This user has privileges to create schemas and assign roles.
- **Schema isolation** — each service operates in its own PostgreSQL schema (`auth.*`, `iam.*`). Service database roles only have access to their own schema.
- **Default privileges** — `ALTER DEFAULT PRIVILEGES` in the bootstrap ensures tables created later automatically grant CRUD to the appropriate service role.

## Security Note

The bootstrap migration creates database roles with `PASSWORD 'default'`. These must be changed before any production deployment.
