package domain

import (
	"context"
	"crypto/rsa"
	"fmt"

	"github.com/araujoarthur/intranetbackend/services/auth/internal/repository"
	"github.com/araujoarthur/intranetbackend/shared/pkg/token"
	"github.com/google/uuid"
)

// ServiceTokenService defines the business logic contract for service token management.
// Service tokens are non-expiring tokens issued to service accounts.
// Their validity is controlled by revocation rather than expiry.
type ServiceTokenService interface {
	// Issue creates a new service token for the given identity.
	// Revokes any existing active service token before issuing a new one —
	// only one active service token exists per identity at any time.
	// Returns the raw JWT string to be stored by the caller.
	// Called by inetbctl when bootstrapping a new service account.
	Issue(ctx context.Context, identityID uuid.UUID) (string, error)

	// Rotate revokes and reissues service tokens for all active service accounts.
	// Called by the daily background job.
	// Non-fatal per identity — failure for one account is logged and skipped,
	// rotation continues for the remaining accounts.
	Rotate(ctx context.Context) error
}

type serviceTokenService struct {
	store      *repository.Store
	privateKey *rsa.PrivateKey
}

func (s *serviceTokenService) Issue(ctx context.Context, identityID uuid.UUID) (string, error) {
	serviceToken, err := token.IssueServiceToken(identityID, s.privateKey)
	if err != nil {
		return "", fmt.Errorf("ServiceTokenService.Issue: %w", err)
	}

	err = s.store.WithTx(ctx, func(store *repository.Store) error {
		if err := store.ServiceTokens.RevokeAllByIdentity(ctx, identityID); err != nil {
			return fmt.Errorf("ServiceTokenService.Issue.tx: %w", err)
		}

		if _, err := store.ServiceTokens.Create(ctx, identityID, serviceToken); err != nil {
			return fmt.Errorf("ServiceTokenService.Issue.tx: %w", err)
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("ServiceTokenService.Issue: %w", err)
	}

	return serviceToken, nil
}

func (s *serviceTokenService) Rotate(ctx context.Context) error {
	activeTokens, err := s.store.ServiceTokens.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("ServiceTokenService.Rotate: %w", err)
	}

	for _, tkn := range activeTokens {
		_, err := s.Issue(ctx, tkn.IdentityID)
		if err != nil {
			fmt.Printf("ServiceTokenService.Rotate: failed to rotate token for %s: %v\n", tkn.IdentityID, err)
			continue
		}
	}

	return nil
}
