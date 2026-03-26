# provisioner

Package `provisioner` defines the `PrincipalProvisioner` interface used by the auth service to create IAM principals when a new identity registers. It decouples the auth service from IAM's concrete HTTP client.

## Files

```
provisioner/
└── provisioner.go    PrincipalProvisioner interface
```

## Interface

```go
type PrincipalProvisioner interface {
    Provision(ctx context.Context, externalID uuid.UUID, principalType types.PrincipalType) error
}
```

| Parameter | Description |
|---|---|
| `externalID` | The identity UUID issued by the auth service |
| `principalType` | `types.PrincipalTypeUser` or `types.PrincipalTypeService` |

## Contract

- A failure from `Provision` is **non-fatal** by design. The auth service logs the error and continues. A background reconciliation job handles identities whose IAM principal was not created.
- The interface is intentionally minimal — just enough for auth to tell IAM "this identity exists now."

## Default Implementation

`services/iam/client.Client` satisfies this interface via `Provision`, which calls `POST /api/v1/principals` on the IAM service. (Note: there is currently a route mismatch — the IAM server registers `createPrincipal` at `POST /api/v1/`, not `/api/v1/principals`. See the IAM service CONTEXT.md for details.)

## Why This Package Exists

Without this package, the auth service would have to import the IAM client directly, creating a cross-service dependency. By depending only on this interface, auth remains decoupled from IAM's transport layer. A future message-queue implementation (async provisioning) can satisfy the same interface without any changes to auth's domain code.

## Usage in Auth

```go
// main.go wires the concrete client
iamClient := iamclient.New(os.Getenv("IAM_URL"), serviceToken)

// domain layer receives the interface
authService := domain.NewSessionService(store, iamClient, hasher)

// domain layer calls provisioner after identity creation
if err := p.provisioner.Provision(ctx, identity.ID, types.PrincipalTypeUser); err != nil {
    log.Printf("warn: IAM provisioning failed for %s: %v", identity.ID, err)
    // continue — reconciliation job will retry
}
```
