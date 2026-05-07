package rest

import (
	"net/http"

	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/araujoarthur/intranetbackend/shared/pkg/response"
	"github.com/araujoarthur/intranetbackend/shared/pkg/token"
)

func (s *Server) deleteIdentity(w http.ResponseWriter, r *http.Request) {
	callerID, err := token.ActorFromContext(r.Context())
	if err != nil {
		response.Error(w, apierror.ErrUnauthorized, err.Error())
		return
	}

	identityID, ok := token.ParseUUIDParam(w, r, "id")
	if !ok {
		return
	}

	if err := s.accounts.Delete(r.Context(), callerID, identityID); err != nil {
		response.Error(w, err, "failed to delete identity")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) changePassword(w http.ResponseWriter, r *http.Request) {
	callerID, err := token.ActorFromContext(r.Context())
	if err != nil {
		response.Error(w, apierror.ErrUnauthorized, err.Error())
		return
	}

}

func (s *Server) removeCredential(w http.ResponseWriter, r *http.Request) {

}
