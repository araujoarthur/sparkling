// bcrypt.go implements Hasher using the bcrypt algorithm.
package hasher

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BcryptHasher implements Hasher using bcrypt.
// Cost controls the computational expense of hashing —
// higher cost is more secure but slower.
// The recommended minimum cost is 12.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher constructs a BcryptHasher with the given cost.
// If cost is below bcrypt.MinCost it defaults to 12.
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < bcrypt.MinCost {
		cost = 12
	}
	return &BcryptHasher{cost: cost}
}

// Hash returns a bcrypt hash of the plaintext.
func (h *BcryptHasher) Hash(plaintext string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), h.cost)
	if err != nil {
		return "", fmt.Errorf("BcryptHasher.Hash: %w", err)
	}
	return string(hash), nil
}

// Verify checks whether the plaintext matches the bcrypt hash.
func (h *BcryptHasher) Verify(plaintext, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, fmt.Errorf("BcryptHasher.Verify: %w", err)
	}
	return true, nil
}
