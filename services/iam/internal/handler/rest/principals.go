// principals.go implements HTTP handlers for the IAM principals API.
package rest

import (
	"encoding/json"
	"net/http"

	"github.com/araujoarthur/intranetbackend/services/iam/contract"
	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/araujoarthur/intranetbackend/shared/pkg/response"
	"github.com/araujoarthur/intranetbackend/shared/pkg/token"
	"github.com/araujoarthur/intranetbackend/shared/pkg/types"
)

func (s *Server) createPrincipal(w http.ResponseWriter, r *http.Request) {
	var req contract.CreatePrincipalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.ErrInvalidArgument, "invalid request body")
		return
	}

	principal, err := s.principals.Create(r.Context(), req.ExternalID, req.PrincipalType)
	if err != nil {
		response.Error(w, err, "failed to create principal")
		return
	}

	response.JSON(w, http.StatusCreated, toPrincipalResponse(principal))
}

func (s *Server) listPrincipals(w http.ResponseWriter, r *http.Request) {
	list, err := s.principals.List(r.Context())
	if err != nil {
		response.Error(w, err, "failed to list principals")
		return
	}

	res := make([]contract.PrincipalResponse, len(list))
	for i, principal := range list {
		res[i] = toPrincipalResponse(principal)
	}

	response.JSON(w, http.StatusOK, res)
}

func (s *Server) getPrincipalByID(w http.ResponseWriter, r *http.Request) {
	parsed, ok := token.ParseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	principal, err := s.principals.GetByID(r.Context(), parsed)
	if err != nil {
		response.Error(w, err, "failed to get principal by ID")
		return
	}

	response.JSON(w, http.StatusOK, toPrincipalResponse(principal))
}

func (s *Server) getPrincipalByExternalID(w http.ResponseWriter, r *http.Request) {
	parsed, ok := token.ParseUUIDParam(w, r, "externalID")
	if !ok {
		return
	}

	principalType := r.URL.Query().Get("type")
	if principalType == "" {
		response.Error(w, apierror.ErrInvalidArgument, "missing type query parameter")
		return
	}

	principal, err := s.principals.GetByExternalID(r.Context(), parsed, types.PrincipalType(principalType))
	if err != nil {
		response.Error(w, err, "failed to get principal by external ID")
		return
	}

	response.JSON(w, http.StatusOK, toPrincipalResponse(principal))
}

func (s *Server) deletePrincipal(w http.ResponseWriter, r *http.Request) {
	callerID, err := token.ActorFromContext(r.Context())
	if err != nil {
		response.Error(w, apierror.ErrUnauthorized, err.Error())
		return
	}

	parsed, ok := token.ParseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	err = s.principals.Delete(r.Context(), callerID, parsed)
	if err != nil {
		response.Error(w, err, "failed to delete principal")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) activatePrincipal(w http.ResponseWriter, r *http.Request) {
	callerID, err := token.ActorFromContext(r.Context())
	if err != nil {
		response.Error(w, apierror.ErrUnauthorized, err.Error())
		return
	}

	parsed, ok := token.ParseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	principal, err := s.principals.Activate(r.Context(), callerID, parsed)
	if err != nil {
		response.Error(w, err, "failed to activate principal")
		return
	}

	response.JSON(w, http.StatusOK, toPrincipalResponse(principal))
}

func (s *Server) deactivatePrincipal(w http.ResponseWriter, r *http.Request) {
	callerID, err := token.ActorFromContext(r.Context())
	if err != nil {
		response.Error(w, apierror.ErrUnauthorized, err.Error())
		return
	}

	parsed, ok := token.ParseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	principal, err := s.principals.Deactivate(r.Context(), callerID, parsed)
	if err != nil {
		response.Error(w, err, "failed to deactivate principal")
		return
	}

	response.JSON(w, http.StatusOK, toPrincipalResponse(principal))
}

func (s *Server) listRolesByPrincipal(w http.ResponseWriter, r *http.Request) {
	parsed, ok := token.ParseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	list, err := s.principalRoles.ListRolesByPrincipal(r.Context(), parsed)
	if err != nil {
		response.Error(w, err, "failed to list roles by principal")
		return
	}

	res := make([]contract.RoleResponse, len(list))
	for i, role := range list {
		res[i] = toRoleResponse(role)
	}

	response.JSON(w, http.StatusOK, res)
}

func (s *Server) assignRoleToPrincipal(w http.ResponseWriter, r *http.Request) {
	callerID, err := token.ActorFromContext(r.Context())
	if err != nil {
		response.Error(w, apierror.ErrUnauthorized, err.Error())
		return
	}

	parsed, ok := token.ParseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	var req contract.AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.ErrInvalidArgument, "invalid request body")
		return
	}

	if err := s.principalRoles.Assign(r.Context(), callerID, parsed, req.RoleID); err != nil {
		response.Error(w, err, "failed to assign role to principal")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) removeRoleFromPrincipal(w http.ResponseWriter, r *http.Request) {
	callerID, err := token.ActorFromContext(r.Context())
	if err != nil {
		response.Error(w, apierror.ErrUnauthorized, err.Error())
		return
	}

	principalID, ok := token.ParseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	roleID, ok := token.ParseUUIDParam(w, r, "roleID")
	if !ok {
		return
	}

	if err := s.principalRoles.Remove(r.Context(), callerID, principalID, roleID); err != nil {
		response.Error(w, err, "failed to remove role from principal")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) getPrincipalPermissions(w http.ResponseWriter, r *http.Request) {
	parsed, ok := token.ParseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	list, err := s.principals.GetPermissions(r.Context(), parsed)
	if err != nil {
		response.Error(w, err, "failed to get principal permissions")
		return
	}

	res := make([]contract.PermissionResponse, len(list))
	for i, p := range list {
		res[i] = toPermissionResponse(p)
	}

	response.JSON(w, http.StatusOK, res)
}
