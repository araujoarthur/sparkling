// roles.go implements HTTP handlers for the IAM roles API.
package rest

import (
	"encoding/json"
	"net/http"

	"github.com/araujoarthur/intranetbackend/services/iam/contract"
	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/araujoarthur/intranetbackend/shared/pkg/response"
	"github.com/araujoarthur/intranetbackend/shared/pkg/token"
)

func (s *Server) listRoles(w http.ResponseWriter, r *http.Request) {
	list, err := s.roles.List(r.Context())
	if err != nil {
		response.Error(w, err, "failed to list roles")
		return
	}

	res := make([]contract.RoleResponse, len(list))
	for i, p := range list {
		res[i] = toRoleResponse(p)
	}

	response.JSON(w, http.StatusOK, res)
}

func (s *Server) getRoleByID(w http.ResponseWriter, r *http.Request) {
	roleUUID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return // response.Error was already called within parseUUIDParam
	}

	role, err := s.roles.GetByID(r.Context(), roleUUID)
	if err != nil {
		response.Error(w, err, "failed to get role")
		return
	}

	response.JSON(w, http.StatusOK, toRoleResponse(role))
}

func (s *Server) createRole(w http.ResponseWriter, r *http.Request) {
	callerID, err := token.ActorFromContext(r.Context())
	if err != nil {
		response.Error(w, apierror.ErrUnauthorized, err.Error())
		return
	}

	var req contract.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.ErrInvalidArgument, "invalid request body")
		return
	}

	role, err := s.roles.Create(r.Context(), callerID, req.Name, req.Description)
	if err != nil {
		response.Error(w, err, "failed to create role")
		return
	}

	response.JSON(w, http.StatusCreated, toRoleResponse(role))
}

func (s *Server) updateRole(w http.ResponseWriter, r *http.Request) {
	callerID, err := token.ActorFromContext(r.Context())
	if err != nil {
		response.Error(w, apierror.ErrUnauthorized, err.Error())
		return
	}

	roleUUID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return // response.Error was already called within parseUUIDParam
	}

	var req contract.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.ErrInvalidArgument, "invalid request body")
		return
	}

	role, err := s.roles.Update(r.Context(), callerID, roleUUID, req.Name, req.Description)
	if err != nil {
		response.Error(w, err, "failed to update")
		return
	}

	response.JSON(w, http.StatusOK, toRoleResponse(role))
}

func (s *Server) deleteRole(w http.ResponseWriter, r *http.Request) {
	callerID, err := token.ActorFromContext(r.Context())
	if err != nil {
		response.Error(w, apierror.ErrUnauthorized, err.Error())
		return
	}

	roleUUID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return // response.Error was already called within parseUUIDParam
	}

	err = s.roles.Delete(r.Context(), callerID, roleUUID)
	if err != nil {
		response.Error(w, err, "failed to delete")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listPermissionsByRole(w http.ResponseWriter, r *http.Request) {
	parsed, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	list, err := s.permissions.ListByRole(r.Context(), parsed)
	if err != nil {
		response.Error(w, err, "failed to list permissions")
		return
	}

	res := make([]contract.PermissionResponse, len(list))
	for i, permission := range list {
		res[i] = toPermissionResponse(permission)
	}

	response.JSON(w, http.StatusOK, res)
}

func (s *Server) assignPermissionToRole(w http.ResponseWriter, r *http.Request) {
	callerID, err := token.ActorFromContext(r.Context())
	if err != nil {
		response.Error(w, apierror.ErrUnauthorized, err.Error())
		return
	}

	parsed, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	var req contract.AssignPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.ErrInvalidArgument, "invalid request format")
		return
	}

	if err := s.rolePermissions.Assign(r.Context(), callerID, parsed, req.PermissionID); err != nil {
		response.Error(w, err, "failed to assign permission")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) removePermissionFromRole(w http.ResponseWriter, r *http.Request) {
	callerID, err := token.ActorFromContext(r.Context())
	if err != nil {
		response.Error(w, apierror.ErrUnauthorized, err.Error())
		return
	}

	roleID, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	permID, ok := parseUUIDParam(w, r, "permID")
	if !ok {
		return
	}

	err = s.rolePermissions.Remove(r.Context(), callerID, roleID, permID)
	if err != nil {
		response.Error(w, err, "failed to remove permission from role")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listPrincipalsByRole(w http.ResponseWriter, r *http.Request) {
	parsed, ok := parseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	list, err := s.principalRoles.ListPrincipalsByRole(r.Context(), parsed)
	if err != nil {
		response.Error(w, err, "failed to list principals by role")
		return
	}

	res := make([]contract.PrincipalResponse, len(list))
	for i, principal := range list {
		res[i] = toPrincipalResponse(principal)
	}

	response.JSON(w, http.StatusOK, res)
}
