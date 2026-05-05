package domain

import (
	"context"
	"fmt"

	"github.com/araujoarthur/intranetbackend/services/auth/internal/repository"
)

type RefreshTokenService interface {
	DeleteAllExpired(ctx context.Context) error
}

type refreshTokenService struct {
	store *repository.Store
}

func NewRefreshTokenService(store *repository.Store) RefreshTokenService {
	return &refreshTokenService{
		store: store,
	}
}

func (s *refreshTokenService) DeleteAllExpired(ctx context.Context) error {
	if err := s.store.RefreshTokens.DeleteAllExpired(ctx); err != nil {
		return fmt.Errorf("RefreshTokenService.DeleteAllExpired: %w", err)
	}

	return nil
}
