# iam/cmd/iamd

Composition root for the IAM service. Wires together configuration, database, domain services, and the HTTP server.

## Files

```
cmd/iamd/
└── main.go    service entry point and dependency wiring
```

## Startup Sequence

```
godotenv.Load()
→ FileKeyProvider(PRIVATE_KEY_PATH, PUBLIC_KEY_PATH) → publicKey
→ database.NewPool(IAM_DSN) → pool
→ repository.NewStore(pool) → store
→ domain.New*Service(store) × 5
→ rest.NewServer(publicKey, services...)
→ http.ListenAndServe(IAM_ADDR || ":8081", server)
```

## Environment Variables

| Variable | Purpose |
|---|---|
| `IAM_DSN` | PostgreSQL connection string |
| `PUBLIC_KEY_PATH` | RSA public key PEM file (for token validation) |
| `PRIVATE_KEY_PATH` | Passed to FileKeyProvider constructor (unused by IAM) |
| `IAM_ADDR` | Listen address; defaults to `:8081` |

## Notes

- `.env` is loaded via `godotenv.Load()` — failure is silently ignored (file is optional).
- `PRIVATE_KEY_PATH` is required by `FileKeyProvider`'s constructor but IAM never calls `PrivateKey()`. Only the public key is used for token validation.
- No graceful shutdown — `http.ListenAndServe` blocks until error or process kill.
- Pool is closed via `defer pool.Close()` on process exit.
