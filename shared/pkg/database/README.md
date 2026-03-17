Package database provides shared database connectivity primitives for all services.

It is the single source of truth for pool creation and transaction management.
No service should open a database connection outside of this package.

Each service provides its own DSN via config — this package handles the rest.

Folder structure:
  database/
  ├── database.go    pool construction and connectivity validation
  └── tx.go       transaction helper (WithTx)

Usage:
  Instantiate a pool once at service startup via NewPool, pass it down
  through constructors. Use WithTx to wrap multiple repository calls
  in a single atomic transaction.