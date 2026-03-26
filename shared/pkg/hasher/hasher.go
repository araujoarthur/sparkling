// hasher.go defines the Hasher interface for password hashing and verification.
// The implementation is chosen at startup via the AUTH_HASHER environment variable.
package hasher

// Hasher defines the contract for password hashing and verification.
// Implementations must be safe for concurrent use.
type Hasher interface {
	// Hash returns a hashed representation of the plaintext.
	// The returned hash is safe to store in the database.
	Hash(plaintext string) (string, error)
	// Verify checks whether the plaintext matches the stored hash.
	// Returns true if they match, false otherwise.
	Verify(plaintext string, hash string) (bool, error)
}
