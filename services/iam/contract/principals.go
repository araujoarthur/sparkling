package contract

import (
	"time"

	"github.com/araujoarthur/intranetbackend/shared/pkg/types"
	"github.com/google/uuid"
)

type PrincipalResponse struct {
	ID            uuid.UUID           `json:"id"`
	ExternalID    uuid.UUID           `json:"external_id"`
	PrincipalType types.PrincipalType `json:"principal_type"`
	IsActive      bool                `json:"is_active"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// CreatePrincipalRequest is the request body for POST /api/v1/principals.
// Called by the auth service when a new identity is registered.
type CreatePrincipalRequest struct {
	ExternalID    uuid.UUID           `json:"external_id"`
	PrincipalType types.PrincipalType `json:"principal_type"`
}
