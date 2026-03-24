package repository

import (
	"time"

	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/google/uuid"
)

// --------------------------------
// Sentinel Errors
// --------------------------------
var (
	ErrNotFound        = apierror.ErrNotFound
	ErrConflict        = apierror.ErrConflict
	ErrForbidden       = apierror.ErrForbidden
	ErrInvalidArgument = apierror.ErrInvalidArgument
)

// --------------------------------
// Domain Types
// --------------------------------
type Identity struct {
	ID        uuid.UUID
	CreatedAt time.Time
}

type CredentialType string

const (
	CredentialTypePassword     CredentialType = "password"
	CredentialTypeServiceToken CredentialType = "service_token"
)

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
