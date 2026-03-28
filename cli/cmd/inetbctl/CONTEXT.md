# cli/cmd/inetbctl

Composition root for the `inetbctl` management CLI. Registers all top-level commands and loads `.env` on startup.

## Files

```
cmd/inetbctl/
└── main.go    root command, command registration, .env loading
```

## Startup

1. Creates the root Cobra command (`inetbctl`)
2. Registers `PersistentPreRunE` that calls `godotenv.Load()` — silently ignored if `.env` is missing
3. Adds subcommands: `db` (from `commands.DBCmd()`) and `keys` (from `commands.KeysCmd()`)
4. Executes the command tree; exits with code 1 on error

## Command Tree

```
inetbctl
├── db
│   ├── bootstrap [--down]
│   ├── migrate (up|down|status) [service]
│   └── seed (up|down|status) [service]
└── keys
    └── generate [--out dir]
```

## Dependencies

- `cli/internal/commands` — all command implementations
- `github.com/joho/godotenv` — `.env` file loading
- `github.com/spf13/cobra` — CLI framework
