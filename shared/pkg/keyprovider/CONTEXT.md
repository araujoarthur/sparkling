# keyprovider

Package `keyprovider` provides interface-based RSA key loading. The key source is swappable per environment without changing any call sites.

## Files

```
keyprovider/
├── keyprovider.go    PublicKeyProvider and PrivateKeyProvider interfaces
├── file.go           FileKeyProvider — loads keys from PEM files on disk
├── env.go            EnvKeyProvider — loads keys from base64-encoded env vars
└── vault.go          VaultKeyProvider — placeholder, not implemented
```

## Interfaces

### `PublicKeyProvider`

```go
type PublicKeyProvider interface {
    PublicKey() (*rsa.PublicKey, error)
}
```

Satisfied by any service that validates tokens. IAM and all future services depend on this interface.

### `PrivateKeyProvider`

```go
type PrivateKeyProvider interface {
    PublicKey() (*rsa.PublicKey, error)
    PrivateKey() (*rsa.PrivateKey, error)
}
```

Extends `PublicKeyProvider`. Auth-domain code needs the private key for JWT issuance. The current CLI generates RSA key files directly and does not use this provider to issue tokens.

## Implementations

### `FileKeyProvider`

**Constructor:** `NewFileKeyProvider(privateKeyPath, publicKeyPath string) *FileKeyProvider`

Loads PEM files from the provided absolute or relative paths on disk. The private key file must be PKCS#1 (`RSA PRIVATE KEY` PEM header). The public key file must be PKIX (`PUBLIC KEY` PEM header).

**When to use:** Development and CI environments where PEM files can be placed on the filesystem.

**Error wrapping:** All errors include a `[step]` label (`[read]`, `[decode]`, `[parse]`) so the failure point is immediately identifiable.

**Note:** `FileKeyProvider` implements `PrivateKeyProvider` (has both methods). IAM's `main.go` passes `PRIVATE_KEY_PATH` even though IAM never calls `PrivateKey()` — this is a constructor artefact, not a security concern.

### `EnvKeyProvider`

**Constructor:** `NewEnvKeyProvider(privateKeyEnv, publicKeyEnv string) *EnvKeyProvider`

Reads environment variable names (not values) at construction time. The actual variable values are read at call time via `os.Getenv`. Values must be **base64-encoded PEM** (standard encoding, not URL-safe).

**When to use:** Production / containerised deployments where secrets are injected as environment variables. Avoids mounting files.

**Key format:** `base64( PEM-encoded-key )` — i.e. base64-encode the entire PEM block including the header and footer lines.

**Error wrapping:** Labels include `[decode base64]` and `[decode pem]` to distinguish the two decoding steps.

### `VaultKeyProvider`

Placeholder only (`// Future...`). Not implemented.

## Usage Example

```go
// Development
kp := keyprovider.NewFileKeyProvider(
    os.Getenv("PRIVATE_KEY_PATH"),
    os.Getenv("PUBLIC_KEY_PATH"),
)

// Production
kp := keyprovider.NewEnvKeyProvider("PRIVATE_KEY_B64", "PUBLIC_KEY_B64")

// All services load the public key at startup and pass it to token.Middleware
publicKey, err := kp.PublicKey()
```

## Key Generation

Use `inetbctl keys generate --out ./keys` to generate a 2048-bit RSA key pair as PEM files. For `EnvKeyProvider`, base64-encode the resulting files:

```bash
base64 -w 0 ./keys/private.pem
base64 -w 0 ./keys/public.pem
```
