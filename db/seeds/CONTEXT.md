# db/seeds

Goose seed data organized by service. Managed via `inetbctl db seed`.

## Structure

```
seeds/
└── iam/
    └── 001_roles.sql    built-in system roles
```

## IAM Seeds

### `001_roles.sql`

Inserts three built-in system roles:

| Name | Description | `is_system` |
|---|---|---|
| `superadmin` | Full system access | `true` |
| `admin` | Administrative access | `true` |
| `user` | Standard user access | `true` |

Uses `ON CONFLICT (name) DO NOTHING` for idempotent re-runs.

System roles (`is_system = true`) cannot be updated or deleted — the IAM repository SQL queries filter `WHERE is_system = false` on mutation operations.

## Versioning

Seed version tables are separate from migration version tables: `{service}.goose_seeds_version`. This allows seeds to be applied and reversed independently of schema migrations.
