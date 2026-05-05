package rest

import (
	"github.com/araujoarthur/intranetbackend/services/iam/contract"
	"github.com/araujoarthur/intranetbackend/services/iam/internal/repository"
)

func toRoleResponse(r repository.Role) contract.RoleResponse {
	return contract.RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func toPermissionResponse(p repository.Permission) contract.PermissionResponse {
	return contract.PermissionResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
	}
}

func toPrincipalResponse(p repository.Principal) contract.PrincipalResponse {
	return contract.PrincipalResponse{
		ID:            p.ID,
		ExternalID:    p.ExternalID,
		PrincipalType: p.PrincipalType,
		IsActive:      p.IsActive,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}
