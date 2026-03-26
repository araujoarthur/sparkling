// auth_domain.go provides shared helpers used across all auth domain services.
// It is the domain layer's equivalent of auth_repository.go — a common
// foundation that every service file in this package depends on.
package domain

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// generateToken generates a cryptographically secure random token string.
// The raw token is returned to the caller for transmission to the client.
// Only the hash should be stored in the database — never the raw token.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generateToken: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
