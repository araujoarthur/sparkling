package domain

import (
	"context"

	"github.com/araujoarthur/intranetbackend/services/auth/internal/repository"
)

type SessionService interface {
	Register(ctx context.Context, username, password string) (repository.Identity, error)
}
