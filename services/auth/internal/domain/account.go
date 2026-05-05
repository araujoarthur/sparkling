package domain

import (
	"context"
	"fmt"

	"github.com/araujoarthur/intranetbackend/services/auth/internal/repository"
	iamclient "github.com/araujoarthur/intranetbackend/services/iam/client"
	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/araujoarthur/intranetbackend/shared/pkg/hasher"
	"github.com/google/uuid"
)

type AccountService interface {
	// Delete permanently removes an identity and all associated data.
	// Cascades to credentials, refresh tokens and service tokens.
	Delete(ctx context.Context, callerID uuid.UUID, identityID uuid.UUID) error

	// ChangePassword updates the password credential for the given identity.
	// Verifies the old password before updating.
	// Revokes all active refresh tokens to invalidate existing sessions.
	ChangePassword(ctx context.Context, callerID uuid.UUID, identityID uuid.UUID, oldPassword, newPassword string) error

	// RemoveCredential permanently removes a specific credential from an identity.
	// Caller must be the identity owner or have auth:identities:delete permission.
	// Returns ErrForbidden if the caller lacks permission.
	// Returns ErrNotFound if the credential does not exist.
	RemoveCredential(ctx context.Context, callerID uuid.UUID, identityID uuid.UUID, credentialID uuid.UUID) error
}

type accountService struct {
	store     *repository.Store
	hasher    hasher.Hasher
	iamClient iamclient.IAMClient
}

func NewAccountService(store *repository.Store, hasher hasher.Hasher, iamClient iamclient.IAMClient) AccountService {
	return &accountService{
		store:     store,
		hasher:    hasher,
		iamClient: iamClient,
	}
}

func (s *accountService) Delete(ctx context.Context, callerID uuid.UUID, identityID uuid.UUID) error {
	allowed, err := isOwnerOrHasPermission(ctx, s.iamClient, callerID, identityID, permissionAuthIdentitiesDelete)
	if err != nil {
		return fmt.Errorf("AccountService.Delete: %w", err)
	}

	if !allowed {
		return apierror.ErrForbidden
	}

	err = s.store.Identities.Delete(ctx, identityID)
	if err != nil {
		return fmt.Errorf("AccountService.Delete: %w", err)
	}

	return nil
}

func (s *accountService) ChangePassword(ctx context.Context, callerID uuid.UUID, identityID uuid.UUID, oldPassword, newPassword string) error {
	allowed, err := isOwnerOrHasPermission(ctx, s.iamClient, callerID, identityID, permissionAuthCredentialsEdit)
	if err != nil {
		return fmt.Errorf("AccountService.ChangePassword: %w", err)
	}

	if !allowed {
		return apierror.ErrForbidden
	}

	fetched, err := s.store.Credentials.GetByIdentityAndType(ctx, identityID, repository.CredentialTypePassword)
	if err != nil {
		return fmt.Errorf("AccountService.ChangePassword: %w", err)
	}

	// The conditional below checks specifically if the caller is the owner
	// of the identity. If true, then the verification of the old password is
	// needed to prevent account theft following a session hijacking. If the
	// request has reached this point but the caller is NOT the owner, then it
	// must have the permission to change passwords and the check below is not
	// necessary.
	// If the callerID matches the identityID AND no error is returned from the
	// hasher's Verify function AND the !matched block does not enter, then it's
	// safe to continue.
	if callerID == identityID {
		matched, err := s.hasher.Verify(oldPassword, fetched.SecretHash)
		if err != nil {
			return fmt.Errorf("AccountService.ChangePassword [owner]: %w", err)
		}

		if !matched {
			return apierror.ErrInvalidCredentials
		}
	}

	newHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("AccountService.ChangePassword [hash]: %w", err)
	}

	err = s.store.WithTx(ctx, func(store *repository.Store) error {
		if err := store.Credentials.UpdateSecret(ctx, fetched.ID, newHash); err != nil {
			return fmt.Errorf("AccountService.ChangePassword.tx [update]: %w", err)
		}

		if err := store.RefreshTokens.RevokeAllByIdentity(ctx, identityID); err != nil {
			return fmt.Errorf("AccountService.ChangePassword.tx [revoke]: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("AccountService.ChangePassword: %w", err)
	}

	return nil
}

func (s *accountService) RemoveCredential(ctx context.Context, callerID uuid.UUID, identityID uuid.UUID, credentialID uuid.UUID) error {
	allowed, err := isOwnerOrHasPermission(ctx, s.iamClient, callerID, identityID, permissionAuthCredentialsDelete)
	if err != nil {
		return fmt.Errorf("AccountService.RemoveCredential: %w", err)
	}

	if !allowed {
		return apierror.ErrForbidden
	}

	credential, err := s.store.Credentials.GetByID(ctx, credentialID)
	if err != nil {
		return fmt.Errorf("AccountService.RemoveCredential: %w", err)
	}

	// users cannot remove the password credential.
	if credential.Type == repository.CredentialTypePassword {
		return apierror.ErrForbidden
	}

	if credential.IdentityID != identityID {
		return apierror.ErrForbidden
	}

	if err := s.store.Credentials.Delete(ctx, credentialID); err != nil {
		return fmt.Errorf("AccountService.RemoveCredential: %w", err)
	}

	return nil
}
