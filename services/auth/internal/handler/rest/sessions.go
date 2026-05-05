package rest

import (
	"encoding/json"
	"net/http"

	"github.com/araujoarthur/intranetbackend/services/auth/contract"
	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/araujoarthur/intranetbackend/shared/pkg/response"
	"github.com/araujoarthur/intranetbackend/shared/pkg/token"
	"github.com/google/uuid"
)

func (s *Server) register(w http.ResponseWriter, r *http.Request) {

	var req contract.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.ErrInvalidArgument, "invalid request body")
		return
	}

	identity, err := s.sessions.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		response.Error(w, err, "failed to register identity")
		return
	}

	response.JSON(w, http.StatusCreated, contract.IdentityResponse{
		ID:        identity.ID,
		CreatedAt: identity.CreatedAt,
	})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req contract.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.ErrInvalidArgument, "invalid request body")
		return
	}

	result, err := s.sessions.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		response.Error(w, err, "failed to login")
		return
	}

	response.JSON(w, http.StatusOK, contract.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {

	var req contract.RefreshRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.ErrInvalidArgument, "invalid request body")
		return
	}

	result, err := s.sessions.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		response.Error(w, err, "failed to refresh session")
		return
	}

	response.JSON(w, http.StatusOK, contract.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})

}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {

	var req contract.LogoutRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.ErrInvalidArgument, "invalid request body")
		return
	}

	if err := s.sessions.Logout(r.Context(), req.RefreshToken); err != nil {
		response.Error(w, err, "failed to logout")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) logoutAll(w http.ResponseWriter, r *http.Request) {

	callerID, err := token.ActorFromContext(r.Context())
	if err != nil {
		response.Error(w, apierror.ErrUnauthorized, err.Error())
		return
	}

	var req contract.LogoutAllRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.ErrInvalidArgument, "invalid request body")
		return
	}

	if req.IdentityID == uuid.Nil {
		response.Error(w, apierror.ErrInvalidArgument, "identity_id is required")
		return
	}

	if err := s.sessions.LogoutAll(r.Context(), callerID, req.IdentityID); err != nil {
		response.Error(w, err, "failed to logout all")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
