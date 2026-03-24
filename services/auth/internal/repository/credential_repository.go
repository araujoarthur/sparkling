// credentials_repository.go implements CredentialRepository, providing data access
// for auth credentials backed by PostgreSQL via sqlc-generated queries.
package repository

import (
	"context"

	"github.com/araujoarthur/intranetbackend/services/auth/internal/repository/sqlc/generated"
	"github.com/google/uuid"
)

// CredentialRepository defines the data access contract for auth credentials.
// A credential represents a single authentication method linked to an identity.
// Multiple credential types can exist per identity (e.g. password, service_token).
// Consumers should depend on this interface, never on the concrete implementation.
type CredentialRepository interface {
	// Create inserts a new credential for the given identity and returns it.
	// Returns apierror.ErrConflict if a credential with the same type and identifier already exists.
	Create(ctx context.Context, identityID uuid.UUID, credentialType CredentialType, identifier string, secretHash string) (Credential, error)

	// GetByID retrieves a credential by its internal UUID.
	// Returns apierror.ErrNotFound if no credential exists with the given ID.
	GetByID(ctx context.Context, credentialID uuid.UUID) (Credential, error)

	// GetByTypeAndIdentifier retrieves a credential by its type and identifier.
	// This is the primary login lookup — find the password credential by username.
	// Returns apierror.ErrNotFound if no matching credential exists.
	GetByTypeAndIdentifier(ctx context.Context, credentialType CredentialType, identifier string) (Credential, error)

	// GetByIdentity returns all credentials linked to the given identity.
	GetByIdentity(ctx context.Context, identityID uuid.UUID) ([]Credential, error)

	// GetByIdentityAndType retrieves a specific credential type for the given identity.
	// Returns apierror.ErrNotFound if no matching credential exists.
	GetByIdentityAndType(ctx context.Context, identityID uuid.UUID, credentialType CredentialType) (Credential, error)

	// UpdateLastUsed records the current time as the last usage time for the credential.
	// Called after every successful authentication.
	UpdateLastUsed(ctx context.Context, credentialID uuid.UUID) error

	// UpdateSecret replaces the secret hash for the given credential.
	// Used for password changes and service token rotation.
	UpdateSecret(ctx context.Context, credentialID uuid.UUID, secretHash string) error

	// Delete permanently removes a credential by its ID.
	// Returns apierror.ErrNotFound if no credential exists with the given ID.
	Delete(ctx context.Context, credentialID uuid.UUID) error

	// DeleteByIdentity permanently removes all credentials for the given identity.
	// Called when an identity is deleted.
	DeleteByIdentity(ctx context.Context, identityID uuid.UUID) error
}

// credentialRepository is the concrete implementation of CredentialRepository.
// It wraps sqlc-generated queries and translates between database and domain types.
// Instantiated exclusively by NewStore — never directly.
type credentialRepository struct {
	q *generated.Queries
}
