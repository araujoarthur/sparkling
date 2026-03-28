// auth_domain.go provides shared helpers used across all auth domain services.
// It is the domain layer's equivalent of auth_repository.go — a common
// foundation that every service file in this package depends on.
package domain

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	iamclient "github.com/araujoarthur/intranetbackend/services/iam/client"
	"github.com/google/uuid"
)

const (
	permissionAuthIdentitiesDelete  = "auth:identities:delete"
	permissionAuthCredentialsEdit   = "auth:credentials:edit"
	permissionAuthCredentialsDelete = "auth:credentials:delete"
)

type LoginResult struct {
	AccessToken  string
	RefreshToken string
}

const RefreshTokenDuration = 7 * 24 * time.Hour

// generateToken generates a cryptographically secure random token string.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generateToken: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

// hashToken returns a SHA256 hex-encoded hash of the raw token string.
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func isOwnerOrHasPermission(ctx context.Context, client iamclient.IAMClient, callerID uuid.UUID, ownerID uuid.UUID, permission string) (bool, error) {
	if callerID == ownerID {
		return true, nil
	}

	return client.HasPermission(ctx, callerID, permission)
}
