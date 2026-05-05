// permissions.go implements HTTP handlers for the IAM permissions API.
package rest

import (
	"encoding/json"
	"net/http"

	"github.com/araujoarthur/intranetbackend/services/iam/contract"
	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/araujoarthur/intranetbackend/shared/pkg/response"
	"github.com/araujoarthur/intranetbackend/shared/pkg/token"
)

func (s *Server) listPermissions(w http.ResponseWriter, r *http.Request) {
	list, err := s.permissions.List(r.Context())
	if err != nil {
		response.Error(w, err, "failed to list permissions")
		return
	}

	res := make([]contract.PermissionResponse, len(list))
	for i, p := range list {
		res[i] = toPermissionResponse(p)
	}

	response.JSON(w, http.StatusOK, res)
}

func (s *Server) createPermission(w http.ResponseWriter, r *http.Request) {
	callerID, err := token.ActorFromContext(r.Context())
	if err != nil {
		response.Error(w, apierror.ErrUnauthorized, err.Error())
		return
	}

	var req contract.CreatePermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.ErrInvalidArgument, "invalid request format")
		return
	}

	permission, err := s.permissions.Create(r.Context(), callerID, req.Name, req.Description)
	if err != nil {
		response.Error(w, err, "failed to create permission")
		return
	}

	response.JSON(w, http.StatusCreated, toPermissionResponse(permission))
}

func (s *Server) getPermissionByID(w http.ResponseWriter, r *http.Request) {
	parsed, ok := token.ParseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	permission, err := s.permissions.GetByID(r.Context(), parsed)
	if err != nil {
		response.Error(w, err, "failed to get permission")
		return
	}

	response.JSON(w, http.StatusOK, toPermissionResponse(permission))
}

func (s *Server) deletePermission(w http.ResponseWriter, r *http.Request) {
	callerID, err := token.ActorFromContext(r.Context())
	if err != nil {
		response.Error(w, apierror.ErrUnauthorized, err.Error())
		return
	}

	parsed, ok := token.ParseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	err = s.permissions.Delete(r.Context(), callerID, parsed)
	if err != nil {
		response.Error(w, err, "failed to delete permission")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listRolesByPermission(w http.ResponseWriter, r *http.Request) {
	parsed, ok := token.ParseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	list, err := s.rolePermissions.ListRolesByPermission(r.Context(), parsed)
	if err != nil {
		response.Error(w, err, "failed to list roles by permission")
		return
	}

	res := make([]contract.RoleResponse, len(list))
	for i, role := range list {
		res[i] = toRoleResponse(role)
	}

	response.JSON(w, http.StatusOK, res)
}
