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

The underlying `http.Client` uses a `10 * time.Second` timeout.

### Methods

**`Provision(ctx, externalID, principalType)`**
Posts `CreatePrincipalRequest` to `POST /api/v1/principals`. Returns an error if the response is not `201 Created`.

**`HasPermission`**
Calls `GET /api/v1/principals/{id}/permissions` with the principal as `X-Principal-ID`, decodes the response envelope, and scans for an exact permission name match.

### Internal

`do(ctx, method, path, body, actingPrincipal) (*http.Response, error)` — shared HTTP helper. Marshals body to JSON, attaches auth header and content type, optionally sets `X-Principal-ID`, and executes the request. Caller is responsible for closing `resp.Body`.

## Known Issues

- `HasPermission` fetches and scans the full effective permission list client-side. This is correct but may become inefficient for large permission sets.
