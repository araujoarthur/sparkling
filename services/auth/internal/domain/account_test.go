package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/araujoarthur/intranetbackend/services/auth/internal/domain"
	"github.com/araujoarthur/intranetbackend/services/auth/internal/repository"
	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/araujoarthur/intranetbackend/shared/pkg/types"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

// mockIAMClient implements iamclient.IAMClient.
type mockIAMClient struct {
	hasPermissionFn func(ctx context.Context, principalID uuid.UUID, permission string) (bool, error)
}

func (m *mockIAMClient) Provision(_ context.Context, _ uuid.UUID, _ types.PrincipalType) error {
	return nil
}

func (m *mockIAMClient) HasPermission(ctx context.Context, principalID uuid.UUID, permission string) (bool, error) {
	if m.hasPermissionFn != nil {
		return m.hasPermissionFn(ctx, principalID, permission)
	}
	return false, nil
}

// mockHasher implements hasher.Hasher.
type mockHasher struct {
	hashFn   func(plaintext string) (string, error)
	verifyFn func(plaintext, hash string) (bool, error)
}

func (m *mockHasher) Hash(plaintext string) (string, error) {
	if m.hashFn != nil {
		return m.hashFn(plaintext)
	}
	return "hashed:" + plaintext, nil
}

func (m *mockHasher) Verify(plaintext, hash string) (bool, error) {
	if m.verifyFn != nil {
		return m.verifyFn(plaintext, hash)
	}
	return hash == "hashed:"+plaintext, nil
}

// mockIdentityRepo implements repository.IdentityRepository.
type mockIdentityRepo struct {
	deleteFn func(ctx context.Context, id uuid.UUID) error
}

func (m *mockIdentityRepo) Create(_ context.Context) (repository.Identity, error) {
	return repository.Identity{}, nil
}

func (m *mockIdentityRepo) GetByID(_ context.Context, _ uuid.UUID) (repository.Identity, error) {
	return repository.Identity{}, nil
}

func (m *mockIdentityRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

// mockCredentialRepo implements repository.CredentialRepository.
type mockCredentialRepo struct {
	getByIDFn              func(ctx context.Context, credentialID uuid.UUID) (repository.Credential, error)
	getByIdentityAndTypeFn func(ctx context.Context, identityID uuid.UUID, credentialType repository.CredentialType) (repository.Credential, error)
	updateSecretFn         func(ctx context.Context, credentialID uuid.UUID, secretHash string) error
	deleteFn               func(ctx context.Context, credentialID uuid.UUID) error
}

func (m *mockCredentialRepo) Create(_ context.Context, _ uuid.UUID, _ repository.CredentialType, _ string, _ string) (repository.Credential, error) {
	return repository.Credential{}, nil
}

func (m *mockCredentialRepo) GetByID(ctx context.Context, credentialID uuid.UUID) (repository.Credential, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, credentialID)
	}
	return repository.Credential{}, nil
}

func (m *mockCredentialRepo) GetByTypeAndIdentifier(_ context.Context, _ repository.CredentialType, _ string) (repository.Credential, error) {
	return repository.Credential{}, nil
}

func (m *mockCredentialRepo) GetByIdentity(_ context.Context, _ uuid.UUID) ([]repository.Credential, error) {
	return nil, nil
}

func (m *mockCredentialRepo) GetByIdentityAndType(ctx context.Context, identityID uuid.UUID, credentialType repository.CredentialType) (repository.Credential, error) {
	if m.getByIdentityAndTypeFn != nil {
		return m.getByIdentityAndTypeFn(ctx, identityID, credentialType)
	}
	return repository.Credential{}, nil
}

func (m *mockCredentialRepo) UpdateLastUsed(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockCredentialRepo) UpdateSecret(ctx context.Context, credentialID uuid.UUID, secretHash string) error {
	if m.updateSecretFn != nil {
		return m.updateSecretFn(ctx, credentialID, secretHash)
	}
	return nil
}

func (m *mockCredentialRepo) Delete(ctx context.Context, credentialID uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, credentialID)
	}
	return nil
}

func (m *mockCredentialRepo) DeleteByIdentity(_ context.Context, _ uuid.UUID) error {
	return nil
}

// mockRefreshTokenRepo implements repository.RefreshTokenRepository.
type mockRefreshTokenRepo struct {
	revokeAllByIdentityFn func(ctx context.Context, identityID uuid.UUID) error
}

func (m *mockRefreshTokenRepo) Create(_ context.Context, _ uuid.UUID, _ string, _ time.Time) (repository.RefreshToken, error) {
	return repository.RefreshToken{}, nil
}

func (m *mockRefreshTokenRepo) GetByID(_ context.Context, _ uuid.UUID) (repository.RefreshToken, error) {
	return repository.RefreshToken{}, nil
}

func (m *mockRefreshTokenRepo) GetByHash(_ context.Context, _ string) (repository.RefreshToken, error) {
	return repository.RefreshToken{}, nil
}

func (m *mockRefreshTokenRepo) GetActiveByIdentity(_ context.Context, _ uuid.UUID) ([]repository.RefreshToken, error) {
	return nil, nil
}

func (m *mockRefreshTokenRepo) Revoke(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockRefreshTokenRepo) RevokeAllByIdentity(ctx context.Context, identityID uuid.UUID) error {
	if m.revokeAllByIdentityFn != nil {
		return m.revokeAllByIdentityFn(ctx, identityID)
	}
	return nil
}

func (m *mockRefreshTokenRepo) DeleteAllExpired(_ context.Context) error {
	return nil
}

func (m *mockRefreshTokenRepo) DeleteAllByIdentity(_ context.Context, _ uuid.UUID) error {
	return nil
}

// mockServiceTokenRepo implements repository.ServiceTokenRepository.
type mockServiceTokenRepo struct{}

func (m *mockServiceTokenRepo) Create(_ context.Context, _ uuid.UUID, _ string) (repository.ServiceToken, error) {
	return repository.ServiceToken{}, nil
}

func (m *mockServiceTokenRepo) GetByID(_ context.Context, _ uuid.UUID) (repository.ServiceToken, error) {
	return repository.ServiceToken{}, nil
}

func (m *mockServiceTokenRepo) GetActiveByIdentity(_ context.Context, _ uuid.UUID) (repository.ServiceToken, error) {
	return repository.ServiceToken{}, nil
}

func (m *mockServiceTokenRepo) GetByToken(_ context.Context, _ string) (repository.ServiceToken, error) {
	return repository.ServiceToken{}, nil
}

func (m *mockServiceTokenRepo) Revoke(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockServiceTokenRepo) RevokeAllByIdentity(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockServiceTokenRepo) ListActive(_ context.Context) ([]repository.ServiceToken, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newTestStore creates a Store with mock repositories and a passthrough WithTx.
func newTestStore(
	identities *mockIdentityRepo,
	credentials *mockCredentialRepo,
	refreshTokens *mockRefreshTokenRepo,
) *repository.Store {
	s := &repository.Store{
		Identities:    identities,
		Credentials:   credentials,
		RefreshTokens: refreshTokens,
		ServiceTokens: &mockServiceTokenRepo{},
	}
	s.WithTxFn = func(_ context.Context, fn func(store *repository.Store) error) error {
		return fn(s)
	}
	return s
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestDelete_OwnerDeletesSelf(t *testing.T) {
	identityID := uuid.New()
	var deletedID uuid.UUID

	store := newTestStore(
		&mockIdentityRepo{deleteFn: func(_ context.Context, id uuid.UUID) error {
			deletedID = id
			return nil
		}},
		&mockCredentialRepo{},
		&mockRefreshTokenRepo{},
	)

	svc := domain.NewAccountService(store, &mockHasher{}, &mockIAMClient{})
	err := svc.Delete(context.Background(), identityID, identityID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if deletedID != identityID {
		t.Fatalf("expected identity %v to be deleted, got %v", identityID, deletedID)
	}
}

func TestDelete_AdminWithPermission(t *testing.T) {
	callerID := uuid.New()
	identityID := uuid.New()
	var deletedID uuid.UUID

	store := newTestStore(
		&mockIdentityRepo{deleteFn: func(_ context.Context, id uuid.UUID) error {
			deletedID = id
			return nil
		}},
		&mockCredentialRepo{},
		&mockRefreshTokenRepo{},
	)

	iam := &mockIAMClient{
		hasPermissionFn: func(_ context.Context, pid uuid.UUID, perm string) (bool, error) {
			if pid == callerID && perm == "auth:identities:delete" {
				return true, nil
			}
			return false, nil
		},
	}

	svc := domain.NewAccountService(store, &mockHasher{}, iam)
	err := svc.Delete(context.Background(), callerID, identityID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if deletedID != identityID {
		t.Fatalf("expected identity %v to be deleted, got %v", identityID, deletedID)
	}
}

func TestDelete_CallerWithoutPermission(t *testing.T) {
	callerID := uuid.New()
	identityID := uuid.New()

	store := newTestStore(
		&mockIdentityRepo{},
		&mockCredentialRepo{},
		&mockRefreshTokenRepo{},
	)

	iam := &mockIAMClient{
		hasPermissionFn: func(_ context.Context, _ uuid.UUID, _ string) (bool, error) {
			return false, nil
		},
	}

	svc := domain.NewAccountService(store, &mockHasher{}, iam)
	err := svc.Delete(context.Background(), callerID, identityID)

	if !errors.Is(err, apierror.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestDelete_IAMClientError(t *testing.T) {
	callerID := uuid.New()
	identityID := uuid.New()
	iamErr := errors.New("iam unavailable")

	store := newTestStore(
		&mockIdentityRepo{},
		&mockCredentialRepo{},
		&mockRefreshTokenRepo{},
	)

	iam := &mockIAMClient{
		hasPermissionFn: func(_ context.Context, _ uuid.UUID, _ string) (bool, error) {
			return false, iamErr
		},
	}

	svc := domain.NewAccountService(store, &mockHasher{}, iam)
	err := svc.Delete(context.Background(), callerID, identityID)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, iamErr) {
		t.Fatalf("expected wrapped iamErr, got %v", err)
	}
}

func TestDelete_IdentityNotFound(t *testing.T) {
	identityID := uuid.New()

	store := newTestStore(
		&mockIdentityRepo{deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return apierror.ErrNotFound
		}},
		&mockCredentialRepo{},
		&mockRefreshTokenRepo{},
	)

	svc := domain.NewAccountService(store, &mockHasher{}, &mockIAMClient{})
	err := svc.Delete(context.Background(), identityID, identityID)

	if !errors.Is(err, apierror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// ChangePassword
// ---------------------------------------------------------------------------

func TestChangePassword_OwnerSuccess(t *testing.T) {
	identityID := uuid.New()
	credID := uuid.New()
	var updatedHash string

	creds := &mockCredentialRepo{
		getByIdentityAndTypeFn: func(_ context.Context, _ uuid.UUID, _ repository.CredentialType) (repository.Credential, error) {
			return repository.Credential{
				ID:         credID,
				IdentityID: identityID,
				Type:       repository.CredentialTypePassword,
				SecretHash: "hashed:oldpass",
			}, nil
		},
		updateSecretFn: func(_ context.Context, _ uuid.UUID, hash string) error {
			updatedHash = hash
			return nil
		},
	}

	var revokedIdentity uuid.UUID
	refreshTokens := &mockRefreshTokenRepo{
		revokeAllByIdentityFn: func(_ context.Context, id uuid.UUID) error {
			revokedIdentity = id
			return nil
		},
	}

	store := newTestStore(&mockIdentityRepo{}, creds, refreshTokens)
	svc := domain.NewAccountService(store, &mockHasher{}, &mockIAMClient{})

	err := svc.ChangePassword(context.Background(), identityID, identityID, "oldpass", "newpass")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updatedHash != "hashed:newpass" {
		t.Fatalf("expected hashed new password, got %q", updatedHash)
	}
	if revokedIdentity != identityID {
		t.Fatalf("expected refresh tokens revoked for %v, got %v", identityID, revokedIdentity)
	}
}

func TestChangePassword_OwnerWrongOldPassword(t *testing.T) {
	identityID := uuid.New()

	creds := &mockCredentialRepo{
		getByIdentityAndTypeFn: func(_ context.Context, _ uuid.UUID, _ repository.CredentialType) (repository.Credential, error) {
			return repository.Credential{
				ID:         uuid.New(),
				IdentityID: identityID,
				Type:       repository.CredentialTypePassword,
				SecretHash: "hashed:correctpassword",
			}, nil
		},
	}

	store := newTestStore(&mockIdentityRepo{}, creds, &mockRefreshTokenRepo{})
	svc := domain.NewAccountService(store, &mockHasher{}, &mockIAMClient{})

	err := svc.ChangePassword(context.Background(), identityID, identityID, "wrongpassword", "newpass")

	if !errors.Is(err, apierror.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestChangePassword_AdminSkipsOldPasswordCheck(t *testing.T) {
	callerID := uuid.New()
	identityID := uuid.New()
	credID := uuid.New()
	var updatedHash string

	creds := &mockCredentialRepo{
		getByIdentityAndTypeFn: func(_ context.Context, _ uuid.UUID, _ repository.CredentialType) (repository.Credential, error) {
			return repository.Credential{
				ID:         credID,
				IdentityID: identityID,
				Type:       repository.CredentialTypePassword,
				SecretHash: "hashed:whatever",
			}, nil
		},
		updateSecretFn: func(_ context.Context, _ uuid.UUID, hash string) error {
			updatedHash = hash
			return nil
		},
	}

	refreshTokens := &mockRefreshTokenRepo{
		revokeAllByIdentityFn: func(_ context.Context, _ uuid.UUID) error { return nil },
	}

	iam := &mockIAMClient{
		hasPermissionFn: func(_ context.Context, pid uuid.UUID, perm string) (bool, error) {
			if pid == callerID && perm == "auth:credentials:edit" {
				return true, nil
			}
			return false, nil
		},
	}

	store := newTestStore(&mockIdentityRepo{}, creds, refreshTokens)
	svc := domain.NewAccountService(store, &mockHasher{}, iam)

	// Admin passes empty old password — should not matter because the owner check is skipped.
	err := svc.ChangePassword(context.Background(), callerID, identityID, "", "newpass")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updatedHash != "hashed:newpass" {
		t.Fatalf("expected hashed new password, got %q", updatedHash)
	}
}

func TestChangePassword_CallerWithoutPermission(t *testing.T) {
	callerID := uuid.New()
	identityID := uuid.New()

	store := newTestStore(&mockIdentityRepo{}, &mockCredentialRepo{}, &mockRefreshTokenRepo{})

	iam := &mockIAMClient{
		hasPermissionFn: func(_ context.Context, _ uuid.UUID, _ string) (bool, error) {
			return false, nil
		},
	}

	svc := domain.NewAccountService(store, &mockHasher{}, iam)
	err := svc.ChangePassword(context.Background(), callerID, identityID, "old", "new")

	if !errors.Is(err, apierror.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestChangePassword_CredentialNotFound(t *testing.T) {
	identityID := uuid.New()

	creds := &mockCredentialRepo{
		getByIdentityAndTypeFn: func(_ context.Context, _ uuid.UUID, _ repository.CredentialType) (repository.Credential, error) {
			return repository.Credential{}, apierror.ErrNotFound
		},
	}

	store := newTestStore(&mockIdentityRepo{}, creds, &mockRefreshTokenRepo{})
	svc := domain.NewAccountService(store, &mockHasher{}, &mockIAMClient{})

	err := svc.ChangePassword(context.Background(), identityID, identityID, "old", "new")

	if !errors.Is(err, apierror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestChangePassword_VerifyError(t *testing.T) {
	identityID := uuid.New()
	verifyErr := errors.New("hasher broken")

	creds := &mockCredentialRepo{
		getByIdentityAndTypeFn: func(_ context.Context, _ uuid.UUID, _ repository.CredentialType) (repository.Credential, error) {
			return repository.Credential{
				ID:         uuid.New(),
				IdentityID: identityID,
				Type:       repository.CredentialTypePassword,
				SecretHash: "anything",
			}, nil
		},
	}

	h := &mockHasher{
		verifyFn: func(_, _ string) (bool, error) { return false, verifyErr },
	}

	store := newTestStore(&mockIdentityRepo{}, creds, &mockRefreshTokenRepo{})
	svc := domain.NewAccountService(store, h, &mockIAMClient{})

	err := svc.ChangePassword(context.Background(), identityID, identityID, "old", "new")

	if !errors.Is(err, verifyErr) {
		t.Fatalf("expected wrapped verifyErr, got %v", err)
	}
}

func TestChangePassword_HashError(t *testing.T) {
	identityID := uuid.New()
	hashErr := errors.New("hash failed")

	creds := &mockCredentialRepo{
		getByIdentityAndTypeFn: func(_ context.Context, _ uuid.UUID, _ repository.CredentialType) (repository.Credential, error) {
			return repository.Credential{
				ID:         uuid.New(),
				IdentityID: identityID,
				Type:       repository.CredentialTypePassword,
				SecretHash: "hashed:oldpass",
			}, nil
		},
	}

	h := &mockHasher{
		hashFn: func(_ string) (string, error) { return "", hashErr },
	}

	store := newTestStore(&mockIdentityRepo{}, creds, &mockRefreshTokenRepo{})
	svc := domain.NewAccountService(store, h, &mockIAMClient{})

	err := svc.ChangePassword(context.Background(), identityID, identityID, "oldpass", "new")

	if !errors.Is(err, hashErr) {
		t.Fatalf("expected wrapped hashErr, got %v", err)
	}
}

func TestChangePassword_UpdateSecretError(t *testing.T) {
	identityID := uuid.New()
	updateErr := errors.New("db write failed")

	creds := &mockCredentialRepo{
		getByIdentityAndTypeFn: func(_ context.Context, _ uuid.UUID, _ repository.CredentialType) (repository.Credential, error) {
			return repository.Credential{
				ID:         uuid.New(),
				IdentityID: identityID,
				Type:       repository.CredentialTypePassword,
				SecretHash: "hashed:oldpass",
			}, nil
		},
		updateSecretFn: func(_ context.Context, _ uuid.UUID, _ string) error {
			return updateErr
		},
	}

	store := newTestStore(&mockIdentityRepo{}, creds, &mockRefreshTokenRepo{})
	svc := domain.NewAccountService(store, &mockHasher{}, &mockIAMClient{})

	err := svc.ChangePassword(context.Background(), identityID, identityID, "oldpass", "new")

	if !errors.Is(err, updateErr) {
		t.Fatalf("expected wrapped updateErr, got %v", err)
	}
}

func TestChangePassword_RevokeError(t *testing.T) {
	identityID := uuid.New()
	revokeErr := errors.New("revoke failed")

	creds := &mockCredentialRepo{
		getByIdentityAndTypeFn: func(_ context.Context, _ uuid.UUID, _ repository.CredentialType) (repository.Credential, error) {
			return repository.Credential{
				ID:         uuid.New(),
				IdentityID: identityID,
				Type:       repository.CredentialTypePassword,
				SecretHash: "hashed:oldpass",
			}, nil
		},
		updateSecretFn: func(_ context.Context, _ uuid.UUID, _ string) error { return nil },
	}

	refreshTokens := &mockRefreshTokenRepo{
		revokeAllByIdentityFn: func(_ context.Context, _ uuid.UUID) error { return revokeErr },
	}

	store := newTestStore(&mockIdentityRepo{}, creds, refreshTokens)
	svc := domain.NewAccountService(store, &mockHasher{}, &mockIAMClient{})

	err := svc.ChangePassword(context.Background(), identityID, identityID, "oldpass", "new")

	if !errors.Is(err, revokeErr) {
		t.Fatalf("expected wrapped revokeErr, got %v", err)
	}
}

func TestChangePassword_IAMClientError(t *testing.T) {
	callerID := uuid.New()
	identityID := uuid.New()
	iamErr := errors.New("iam unavailable")

	store := newTestStore(&mockIdentityRepo{}, &mockCredentialRepo{}, &mockRefreshTokenRepo{})

	iam := &mockIAMClient{
		hasPermissionFn: func(_ context.Context, _ uuid.UUID, _ string) (bool, error) {
			return false, iamErr
		},
	}

	svc := domain.NewAccountService(store, &mockHasher{}, iam)
	err := svc.ChangePassword(context.Background(), callerID, identityID, "old", "new")

	if !errors.Is(err, iamErr) {
		t.Fatalf("expected wrapped iamErr, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// RemoveCredential
// ---------------------------------------------------------------------------

func TestRemoveCredential_OwnerSuccess(t *testing.T) {
	identityID := uuid.New()
	credID := uuid.New()
	var deletedID uuid.UUID

	creds := &mockCredentialRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (repository.Credential, error) {
			return repository.Credential{
				ID:         id,
				IdentityID: identityID,
				Type:       repository.CredentialTypeServiceToken,
			}, nil
		},
		deleteFn: func(_ context.Context, id uuid.UUID) error {
			deletedID = id
			return nil
		},
	}

	store := newTestStore(&mockIdentityRepo{}, creds, &mockRefreshTokenRepo{})
	svc := domain.NewAccountService(store, &mockHasher{}, &mockIAMClient{})

	err := svc.RemoveCredential(context.Background(), identityID, identityID, credID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if deletedID != credID {
		t.Fatalf("expected credential %v to be deleted, got %v", credID, deletedID)
	}
}

func TestRemoveCredential_AdminWithPermission(t *testing.T) {
	callerID := uuid.New()
	identityID := uuid.New()
	credID := uuid.New()
	var deletedID uuid.UUID

	creds := &mockCredentialRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (repository.Credential, error) {
			return repository.Credential{
				ID:         id,
				IdentityID: identityID,
				Type:       repository.CredentialTypeServiceToken,
			}, nil
		},
		deleteFn: func(_ context.Context, id uuid.UUID) error {
			deletedID = id
			return nil
		},
	}

	iam := &mockIAMClient{
		hasPermissionFn: func(_ context.Context, pid uuid.UUID, perm string) (bool, error) {
			if pid == callerID && perm == "auth:credentials:delete" {
				return true, nil
			}
			return false, nil
		},
	}

	store := newTestStore(&mockIdentityRepo{}, creds, &mockRefreshTokenRepo{})
	svc := domain.NewAccountService(store, &mockHasher{}, iam)

	err := svc.RemoveCredential(context.Background(), callerID, identityID, credID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if deletedID != credID {
		t.Fatalf("expected credential %v to be deleted, got %v", credID, deletedID)
	}
}

func TestRemoveCredential_CallerWithoutPermission(t *testing.T) {
	callerID := uuid.New()
	identityID := uuid.New()
	credID := uuid.New()

	store := newTestStore(&mockIdentityRepo{}, &mockCredentialRepo{}, &mockRefreshTokenRepo{})

	iam := &mockIAMClient{
		hasPermissionFn: func(_ context.Context, _ uuid.UUID, _ string) (bool, error) {
			return false, nil
		},
	}

	svc := domain.NewAccountService(store, &mockHasher{}, iam)
	err := svc.RemoveCredential(context.Background(), callerID, identityID, credID)

	if !errors.Is(err, apierror.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestRemoveCredential_CredentialNotFound(t *testing.T) {
	identityID := uuid.New()
	credID := uuid.New()

	creds := &mockCredentialRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (repository.Credential, error) {
			return repository.Credential{}, apierror.ErrNotFound
		},
	}

	store := newTestStore(&mockIdentityRepo{}, creds, &mockRefreshTokenRepo{})
	svc := domain.NewAccountService(store, &mockHasher{}, &mockIAMClient{})

	err := svc.RemoveCredential(context.Background(), identityID, identityID, credID)

	if !errors.Is(err, apierror.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRemoveCredential_PasswordTypeBlocked(t *testing.T) {
	identityID := uuid.New()
	credID := uuid.New()

	creds := &mockCredentialRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (repository.Credential, error) {
			return repository.Credential{
				ID:         id,
				IdentityID: identityID,
				Type:       repository.CredentialTypePassword,
			}, nil
		},
	}

	store := newTestStore(&mockIdentityRepo{}, creds, &mockRefreshTokenRepo{})
	svc := domain.NewAccountService(store, &mockHasher{}, &mockIAMClient{})

	err := svc.RemoveCredential(context.Background(), identityID, identityID, credID)

	if !errors.Is(err, apierror.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestRemoveCredential_CredentialBelongsToDifferentIdentity(t *testing.T) {
	identityID := uuid.New()
	otherIdentityID := uuid.New()
	credID := uuid.New()

	creds := &mockCredentialRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (repository.Credential, error) {
			return repository.Credential{
				ID:         id,
				IdentityID: otherIdentityID,
				Type:       repository.CredentialTypeServiceToken,
			}, nil
		},
	}

	store := newTestStore(&mockIdentityRepo{}, creds, &mockRefreshTokenRepo{})
	svc := domain.NewAccountService(store, &mockHasher{}, &mockIAMClient{})

	err := svc.RemoveCredential(context.Background(), identityID, identityID, credID)

	if !errors.Is(err, apierror.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestRemoveCredential_DeleteError(t *testing.T) {
	identityID := uuid.New()
	credID := uuid.New()
	deleteErr := errors.New("db error")

	creds := &mockCredentialRepo{
		getByIDFn: func(_ context.Context, id uuid.UUID) (repository.Credential, error) {
			return repository.Credential{
				ID:         id,
				IdentityID: identityID,
				Type:       repository.CredentialTypeServiceToken,
			}, nil
		},
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return deleteErr
		},
	}

	store := newTestStore(&mockIdentityRepo{}, creds, &mockRefreshTokenRepo{})
	svc := domain.NewAccountService(store, &mockHasher{}, &mockIAMClient{})

	err := svc.RemoveCredential(context.Background(), identityID, identityID, credID)

	if !errors.Is(err, deleteErr) {
		t.Fatalf("expected wrapped deleteErr, got %v", err)
	}
}

func TestRemoveCredential_IAMClientError(t *testing.T) {
	callerID := uuid.New()
	identityID := uuid.New()
	credID := uuid.New()
	iamErr := errors.New("iam unavailable")

	store := newTestStore(&mockIdentityRepo{}, &mockCredentialRepo{}, &mockRefreshTokenRepo{})

	iam := &mockIAMClient{
		hasPermissionFn: func(_ context.Context, _ uuid.UUID, _ string) (bool, error) {
			return false, iamErr
		},
	}

	svc := domain.NewAccountService(store, &mockHasher{}, iam)
	err := svc.RemoveCredential(context.Background(), callerID, identityID, credID)

	if !errors.Is(err, iamErr) {
		t.Fatalf("expected wrapped iamErr, got %v", err)
	}
}
