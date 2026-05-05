package token

import (
	"context"
	"fmt"
	"net/http"

	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/araujoarthur/intranetbackend/shared/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// actingPrincipalKey is a private type for storing the acting principal ID in context.
type actingPrincipalKey struct{}

// ActingPrincipalFromContext extracts the acting principal ID from the request context.
// This is the user on whose behalf the service is acting, passed via X-Principal-ID header.
// Returns empty string if no acting principal is present — meaning the service is acting
// on its own behalf.
func ActingPrincipalFromContext(ctx context.Context) string {
	id, _ := ctx.Value(actingPrincipalKey{}).(string)
	return id
}

// ActorFromContext returns the effective principal for an authenticated request.
// If token.Middleware stored an X-Principal-ID value in the context, that acting
// principal is parsed and returned. Otherwise, it falls back to the subject of
// the validated service token, meaning the calling service is acting on its own
// behalf. Returns an error if X-Principal-ID is not a UUID or if token claims
// are missing from the context.
func ActorFromContext(ctx context.Context) (uuid.UUID, error) {
	actingID := ActingPrincipalFromContext(ctx)

	if actingID != "" {
		parsed, err := uuid.Parse(actingID)

		if err != nil {
			return uuid.UUID{}, fmt.Errorf("invalid X-Principal-ID header: %w", err)
		}
		return parsed, nil
	}

	claims, ok := FromContext(ctx)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("missing claims in context")
	}

	return claims.Subject, nil
}

// parseUUIDParam extracts and parses a UUID URL parameter from the request.
// Returns false and writes an error response if the param is missing or invalid.
func ParseUUIDParam(w http.ResponseWriter, r *http.Request, param string) (uuid.UUID, bool) {
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
