package contract

import (
	"time"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type IdentityResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutAllRequest struct {
	IdentityID uuid.UUID `json:"identity_id"`
}
