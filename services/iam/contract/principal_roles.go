package contract

import "github.com/google/uuid"

type AssignRoleRequest struct {
	RoleID uuid.UUID `json:"role_id"`
}
