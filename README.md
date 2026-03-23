# Intranet Backend Project (Overhaul)

## Dependencies

- [chi](https://github.com/go-chi/chi): Handles the API routing
- [Goose](https://github.com/pressly/goose): Handles database migrations
- [sqlc](https://sqlc.dev/): Handles type safe SQL-to-Code translation
- [jwt-go](github.com/golang-jwt/jwt/v5): Handles JWT
- [godotenv](github.com/joho/godotenv): Handles .env files
- [Cobra](github.com/spf13/cobra): Handles the command engine within the CLI
- OpenSSL (for key generation)

- PostgreSQL 18.1

## Useful inetbctl commands

- `inetbctl db`: Parent of all database management commands
- `inetbctl db bootstrap`: Initializes the database schemas, roles and default privileges
- `inetbctl db seed up|down [service]`: Executes a seed migration up or down in the specified service or all services
- `inetbctl db migrate up|down [service]`: Executes a db migration in the specified service or in all services

## Environment Variables

- `OWNER_DSN`: The Data Source Name (connection URL/string) for the owner user in the database.
- `MIGRATIONS_DIR`: Root folder for database migrations.
- `SEEDS_DIR`: Root folder for database seeds.
