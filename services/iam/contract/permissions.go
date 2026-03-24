package contract

import (
	"time"

	"github.com/google/uuid"
)

type CreatePermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PermissionResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
