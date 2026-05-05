package domain

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/araujoarthur/intranetbackend/services/auth/internal/repository"
	iamclient "github.com/araujoarthur/intranetbackend/services/iam/client"
	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/araujoarthur/intranetbackend/shared/pkg/hasher"
	"github.com/araujoarthur/intranetbackend/shared/pkg/token"
	"github.com/araujoarthur/intranetbackend/shared/pkg/types"
	"github.com/google/uuid"
)

type SessionService interface {
	// Register creates a new identity and password credential.
	// Provisions an IAM principal asynchronously — failure is non-fatal.
	Register(ctx context.Context, username, password string) (repository.Identity, error)

	// Login validates credentials and issues an access token + refresh token.
	// Returns ErrInvalidCredentials if the username or password is wrong.
	Login(ctx context.Context, username, password string) (LoginResult, error)

	// Refresh issues a new access token from a valid refresh token.
	// Returns ErrInvalidToken if the refresh token is invalid or revoked.
	Refresh(ctx context.Context, rawRefreshToken string) (LoginResult, error)

	// Logout revokes the given refresh token.
	Logout(ctx context.Context, rawRefreshToken string) error

	// LogoutAll revokes all refresh tokens for the given identity.
	LogoutAll(ctx context.Context, identityID uuid.UUID) error
}

type sessionService struct {
	store      *repository.Store
	hasher     hasher.Hasher
	iamClient  iamclient.IAMClient
	privateKey *rsa.PrivateKey
}

func NewSessionService(store *repository.Store, hasher hasher.Hasher, iamClient iamclient.IAMClient, privateKey *rsa.PrivateKey) SessionService {
	return &sessionService{
		store:      store,
		hasher:     hasher,
		iamClient:  iamClient,
		privateKey: privateKey,
	}
}

//----------------------------
// Concrete Implementations  -
//----------------------------

func (s *sessionService) Register(ctx context.Context, username, password string) (repository.Identity, error) {

	if strings.TrimSpace(username) == "" {
		return repository.Identity{}, fmt.Errorf("SessionService.Register [username]: %w", apierror.ErrInvalidArgument)
	}

	if strings.TrimSpace(password) == "" {
		return repository.Identity{}, fmt.Errorf("SessionService.Register [password]: %w", apierror.ErrInvalidArgument)
	}

	hashed, err := s.hasher.Hash(password)
	if err != nil {
		return repository.Identity{}, fmt.Errorf("SessionService.Register [hash]: %w", err)
	}

	var identity repository.Identity

	err = s.store.WithTx(ctx, func(store *repository.Store) error {
		authIdentity, err := store.Identities.Create(ctx) // The identity here is regarding authentication entity.
		if err != nil {
			return fmt.Errorf("creating identity: %w", err)
		}

		_, err = store.Credentials.Create(
			ctx,
			authIdentity.ID,
			repository.CredentialTypePassword,
			username,
			hashed,
		)

		if err != nil {
			return fmt.Errorf("creating credential: %w", err)
		}

		identity = authIdentity
		return nil
	})

	if err != nil {
		return repository.Identity{}, fmt.Errorf("SessionService.Register: %w", err)
	}

	// partial failure acceptance
	if err := s.iamClient.Provision(ctx, identity.ID, types.PrincipalTypeUser); err != nil {
		// non fatal error.
		// in the future it will be replaced with an async enqueue call to perform eventual consistency
		fmt.Printf("SessionService.Register: failed to provision IAM principal for %s: %v\n", identity.ID, err)
	}

	return identity, nil
}

func (s *sessionService) Login(ctx context.Context, username, password string) (LoginResult, error) {
	requestedCredentials, err := s.store.Credentials.GetByTypeAndIdentifier(ctx, repository.CredentialTypePassword, username)
	if err != nil {
		if errors.Is(err, apierror.ErrNotFound) {
			return LoginResult{}, apierror.ErrInvalidCredentials
		}

		return LoginResult{}, fmt.Errorf("SessionService.Login [fetch]: %w", err)
	}

	match, err := s.hasher.Verify(password, requestedCredentials.SecretHash)
	if err != nil {
		return LoginResult{}, fmt.Errorf("SessionService.Login [verify]: %w", err)
	}

	if !match {
		return LoginResult{}, apierror.ErrInvalidCredentials
	}

	// Issue the AccessToken
	accessToken, err := token.IssueUserToken(requestedCredentials.IdentityID, s.privateKey)
	if err != nil {
		return LoginResult{}, fmt.Errorf("SessionService.Login [access token]: %w", err)
	}

	// Create the refresh token
	refreshToken, err := generateToken()
	if err != nil {
		return LoginResult{}, fmt.Errorf("SessionService.Login [refresh token]: %w", err)
	}

	// Store the refresh token
	_, err = s.store.RefreshTokens.Create(ctx, requestedCredentials.IdentityID, hashToken(refreshToken), time.Now().Add(RefreshTokenDuration))
	if err != nil {
		return LoginResult{}, fmt.Errorf("SessionService.Login [store]: %w", err)
	}

	if err := s.store.Credentials.UpdateLastUsed(ctx, requestedCredentials.ID); err != nil {
		// non fatal. stale metric
		fmt.Printf("SessionService.Login failed to update last_used_at for %s: %v\n", requestedCredentials.IdentityID, err)
	}

	return LoginResult{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *sessionService) Refresh(ctx context.Context, rawRefreshToken string) (LoginResult, error) {
	hashed := hashToken(rawRefreshToken)

	fetched, err := s.store.RefreshTokens.GetByHash(ctx, hashed)
	if err != nil {
		if errors.Is(err, apierror.ErrNotFound) {
			return LoginResult{}, apierror.ErrInvalidCredentials
		}

		return LoginResult{}, fmt.Errorf("SessionService.Refresh [fetch]: %w", err)
	}

	// if fetching returns no error, the token is equal and valid.

	// issue new access token
	newAccessToken, err := token.IssueUserToken(fetched.IdentityID, s.privateKey)
	if err != nil {
		return LoginResult{}, fmt.Errorf("SessionService.Refresh [access token]: %w", err)
	}

	// issue new refresh token
	newRefreshToken, err := generateToken()
	if err != nil {
		return LoginResult{}, fmt.Errorf("SessionService.Refresh [refresh token]: %w", err)
	}

	// hash and store the token
	hashedNewRefreshToken := hashToken(newRefreshToken)

	err = s.store.WithTx(ctx, func(store *repository.Store) error {
		if _, err := store.RefreshTokens.Create(ctx, fetched.IdentityID, hashedNewRefreshToken, time.Now().Add(RefreshTokenDuration)); err != nil {
			return fmt.Errorf("SessionService.Refresh.TX [store]: %w", err)
		}

		if err := store.RefreshTokens.Revoke(ctx, fetched.ID); err != nil {
			return fmt.Errorf("SessionService.Refresh.TX [revoke]: %w", err)
		}

		return nil
	})

	if err != nil {
		return LoginResult{}, fmt.Errorf("SessionService.Refresh [TX]: %w", err)
	}

	return LoginResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil

}

func (s *sessionService) Logout(ctx context.Context, rawRefreshToken string) error {
	hashed := hashToken(rawRefreshToken)

	fetched, err := s.store.RefreshTokens.GetByHash(ctx, hashed)
	if err != nil {
		if errors.Is(err, apierror.ErrNotFound) {
			return apierror.ErrInvalidCredentials
		}
		return fmt.Errorf("SessionService.Logout [fetch]: %w", err)
	}

	if err := s.store.RefreshTokens.Revoke(ctx, fetched.ID); err != nil {
		return fmt.Errorf("SessionService.Logout: %w", err)
	}

	return nil

}

func (s *sessionService) LogoutAll(ctx context.Context, identityID uuid.UUID) error {

	if err := s.store.RefreshTokens.RevokeAllByIdentity(ctx, identityID); err != nil {
		return fmt.Errorf("SessionService.LogoutAll: %w", err)
	}

	return nil
}
