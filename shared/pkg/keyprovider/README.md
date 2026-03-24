Package keyprovider defines interfaces and implementations for loading RSA keys
used in token signing and validation.

It is the single source of truth for key loading. Services never load keys
directly — they depend on a KeyProvider implementation chosen at startup.
This allows the key source to be swapped between environments without changing
service code.

Interfaces:
  PublicKeyProvider    satisfied by any service that validates tokens
                       exposes PublicKey() (*rsa.PublicKey, error)

  PrivateKeyProvider   satisfied only by the auth service and inetbctl
                       extends PublicKeyProvider with PrivateKey() (*rsa.PrivateKey, error)

Implementations:
  FileKeyProvider      loads keys from PEM files on disk
                       suitable for development and CI
                       constructed via NewFileKeyProvider(privateKeyPath, publicKeyPath)

  EnvKeyProvider       loads keys from base64-encoded PEM environment variables
                       suitable for production and containerized deployments
                       constructed via NewEnvKeyProvider(privateKeyEnv, publicKeyEnv)

Folder structure:
  keyprovider/
  ├── keyprovider.go    PublicKeyProvider and PrivateKeyProvider interfaces
  ├── file.go           FileKeyProvider implementation
  └── env.go            EnvKeyProvider implementation

To add a new provider (e.g. Vault), implement PrivateKeyProvider or
PublicKeyProvider and add it as a new file in this package.
