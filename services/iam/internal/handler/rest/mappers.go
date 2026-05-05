package rest

import (
	"fmt"
	"net/http"

	"github.com/araujoarthur/intranetbackend/services/iam/contract"
	"github.com/araujoarthur/intranetbackend/services/iam/internal/repository"
	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/araujoarthur/intranetbackend/shared/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

// parseUUIDParam extracts and parses a UUID URL parameter from the request.
// Returns false and writes an error response if the param is missing or invalid.
func parseUUIDParam(w http.ResponseWriter, r *http.Request, param string) (uuid.UUID, bool) {
	raw := chi.URLParam(r, param)
	if raw == "" {
		response.Error(w, apierror.ErrInvalidArgument, fmt.Sprintf("missing %s parameter", param))
		return uuid.UUID{}, false
	}

	parsed, err := uuid.Parse(raw)
	if err != nil {
		response.Error(w, apierror.ErrInvalidArgument, fmt.Sprintf("invalid %s format", param))
		return uuid.UUID{}, false
	}

	return parsed, true
}
