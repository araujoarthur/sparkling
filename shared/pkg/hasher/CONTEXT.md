# hasher

Package `hasher` provides a password hashing abstraction with two implementations: Argon2id (recommended) and bcrypt (legacy). The implementation is chosen at startup via configuration.

## Files

```
hasher/
├── hasher.go    Hasher interface
├── argon2.go    Argon2Hasher — Argon2id implementation
└── bcrypt.go    BcryptHasher — bcrypt implementation
```

## Interface

```go
type Hasher interface {
    Hash(plaintext string) (string, error)
    Verify(plaintext string, hash string) (bool, error)
}
```

Both methods are safe for concurrent use. The hash returned by `Hash` is self-contained — it includes all parameters needed to verify it.

## Argon2Hasher

**Algorithm:** Argon2id (memory-hard, OWASP-recommended)

**Constructor:** `NewArgon2Hasher() *Argon2Hasher`

**Default parameters:**
| Parameter | Value |
|---|---|
| Memory | 64 MB (`64 * 1024` KiB) |
| Iterations | 3 |
| Parallelism | 2 |
| Salt length | 16 bytes |
| Key length | 32 bytes |

**Output format:** `$argon2id$v=19$m=65536,t=3,p=2$<base64-salt>$<base64-hash>`

The format is self-contained: `Verify` extracts parameters and salt from the stored string, recomputes the hash, and compares using `subtle.ConstantTimeCompare` to prevent timing attacks.

**`Hash` steps:**
1. Generate 16-byte random salt via `crypto/rand`
2. Derive key with `argon2.IDKey`
3. Encode as PHC string

**`Verify` steps:**
1. Split stored hash on `$` (must have 6 parts)
2. Parse `m`, `t`, `p` parameters from part 3
3. Decode base64 salt and stored hash
4. Recompute with same parameters
5. Constant-time compare

## BcryptHasher

**Algorithm:** bcrypt

**Constructor:** `NewBcryptHasher(cost int) *BcryptHasher`
- If `cost < bcrypt.MinCost`, defaults to `12`
- Minimum recommended cost: `12`

**`Hash`:** calls `bcrypt.GenerateFromPassword` with the configured cost. The salt is embedded in the output by bcrypt.

**`Verify`:** calls `bcrypt.CompareHashAndPassword`. Returns `(false, nil)` on mismatch (not an error), error only on unexpected failure.

## Choosing an Implementation

The implementation is selected at service startup (auth service uses `AUTH_HASHER` env var). Inject as the `Hasher` interface — call sites never reference the concrete type.

| Use case | Implementation |
|---|---|
| New deployments | `Argon2Hasher` |
| Existing bcrypt hashes | `BcryptHasher` |
