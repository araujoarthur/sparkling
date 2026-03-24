package contract

import "github.com/google/uuid"

// AssignPermissionRequest is the request body for POST /roles/{id}/permissions.
type AssignPermissionRequest struct {
	PermissionID uuid.UUID `json:"permission_id"`
}
