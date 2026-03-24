// auth_repository.go defines domain types and mappers for the auth repository
// layer. It is the foundation file for this package — all other repository
// files depend on the types defined here.
// Sentinel errors are sourced directly from shared/pkg/apierror.
package repository

import (
	"time"

	"github.com/araujoarthur/intranetbackend/services/auth/internal/repository/sqlc/generated"
	"github.com/araujoarthur/intranetbackend/shared/pkg/helpers"
	"github.com/google/uuid"
)

// --------------------------------
// Domain Types
// --------------------------------

type CredentialType string

const (
	CredentialTypePassword     CredentialType = "password"
	CredentialTypeServiceToken CredentialType = "service_token"
)

type Identity struct {
	ID        uuid.UUID
	CreatedAt time.Time
}

type Credential struct {
	ID         uuid.UUID
	IdentityID uuid.UUID
	Type       CredentialType
	Identifier string
	SecretHash string
	CreatedAt  time.Time
	LastUsedAt *time.Time
}

type RefreshToken struct {
	ID         uuid.UUID
	IdentityID uuid.UUID
	TokenHash  string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
}

type ServiceToken struct {
	ID         uuid.UUID
	IdentityID uuid.UUID
	Token      string
	IssuedAt   time.Time
	RevokedAt  *time.Time
}

// --------------------------------
// Mappers
// --------------------------------

func toIdentity(g *generated.AuthIdentity) Identity {
	return Identity{
		ID:        g.ID,
		CreatedAt: g.CreatedAt,
	}
}

func toCredential(g *generated.AuthCredential) Credential {
	return Credential{
		ID:         g.ID,
		IdentityID: g.IdentityID,
		Type:       CredentialType(g.Type),
		Identifier: g.Identifier.String,
		SecretHash: g.SecretHash.String,
		CreatedAt:  g.CreatedAt,
		LastUsedAt: helpers.FromNullableTime(g.LastUsedAt),
	}
}

func toRefreshToken(g *generated.AuthRefreshToken) RefreshToken {
	return RefreshToken{
		ID:         g.ID,
		IdentityID: g.IdentityID,
		TokenHash:  g.TokenHash,
		ExpiresAt:  g.ExpiresAt,
		RevokedAt:  helpers.FromNullableTime(g.RevokedAt),
		CreatedAt:  g.CreatedAt,
	}
}

func toServiceToken(g *generated.AuthServiceToken) ServiceToken {
	return ServiceToken{
		ID:         g.ID,
		IdentityID: g.IdentityID,
		Token:      g.Token,
		IssuedAt:   g.IssuedAt,
		RevokedAt:  helpers.FromNullableTime(g.RevokedAt),
	}
}
