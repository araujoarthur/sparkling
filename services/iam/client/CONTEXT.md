# iam/client

Package `iamclient` defines the `IAMClient` interface and provides an HTTP implementation. Used by the auth service to communicate with IAM for principal provisioning and permission checks.

## Files

```
client/
├── intf.go      IAMClient interface definition
└── client.go    Client — HTTP implementation
```

## IAMClient Interface (`intf.go`)

```go
type IAMClient interface {
    Provision(ctx context.Context, externalID uuid.UUID, principalType types.PrincipalType) error
    HasPermission(ctx context.Context, principalID uuid.UUID, permission string) (bool, error)
}
```

| Method | Purpose |
|---|---|
| `Provision` | Creates an IAM principal for a newly registered identity. Failure is non-fatal by contract — the auth service logs and continues. |
| `HasPermission` | Checks whether a principal holds a given permission. |

This interface was extracted from the former `shared/pkg/provisioner` package to co-locate it with its implementation and add `HasPermission`.

## HTTP Implementation (`client.go`)

### Constructor

`New(baseURL, token string) *Client`

- `baseURL` — IAM service URL (e.g. `"http://iam:8081"`)
- `token` — service token attached as `Authorization: Bearer <token>` to every request

**Bug:** Timeout is `10 & time.Second` (bitwise AND = 0). Should be `10 * time.Second`.

### Methods

**`Provision(ctx, externalID, principalType)`**
Posts `CreatePrincipalRequest` to `POST /api/v1/principals`. Returns an error if the response is not `201 Created`.

**`HasPermission`** — declared in the interface but not yet implemented in `client.go`.

### Internal

`do(ctx, method, path, body) (*http.Response, error)` — shared HTTP helper. Marshals body to JSON, attaches auth header and content type, executes the request. Caller is responsible for closing `resp.Body`.

## Known Issues

- **Route mismatch:** The client calls `POST /api/v1/principals` but the IAM server registers `createPrincipal` at `POST /api/v1/` (see `services/iam/internal/handler/rest/server.go:69`).
- **Zero timeout:** `10 & time.Second` evaluates to 0 — no HTTP timeout.
- **HasPermission not implemented:** Interface method exists but `client.go` has no corresponding method.
