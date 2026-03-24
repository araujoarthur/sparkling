// keyprovider.go defines interfaces for loading RSA keys used in token signing
// and validation. Services that only validate tokens depend on PublicKeyProvider.
// Only the auth service depends on PrivateKeyProvider.
package keyprovider

import "crypto/rsa"

// PublicKeyProvider is satisfied by any service that validates tokens.
type PublicKeyProvider interface {
	PublicKey() (*rsa.PublicKey, error)
}

// PrivateKeyProvider is satisfied only by the auth service which issues tokens.
type PrivateKeyProvider interface {
	PublicKey() (*rsa.PublicKey, error)
	PrivateKey() (*rsa.PrivateKey, error)
}
