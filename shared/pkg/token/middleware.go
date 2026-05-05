// middleware.go provides chi-compatible HTTP middleware for JWT token validation.
package token

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/araujoarthur/intranetbackend/shared/pkg/response"
	"github.com/araujoarthur/intranetbackend/shared/pkg/types"
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

// contextKey is a private type for storing claims in request context.
// Using a private type prevents collisions with other packages.
type contextKey struct{}

// Middleware returns a chi-compatible middleware that validates the Bearer token
// in the Authorization header and injects the parsed Claims into the request context.
// Only service tokens are accepted — requests bearing user tokens are rejected with 401.
// The acting principal (the user on whose behalf the service is acting) is extracted
// from the X-Principal-ID header and injected into the context separately.
// Requests with missing, invalid, or expired tokens are rejected with 401.
func Middleware(publicKey *rsa.PublicKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, apierror.ErrUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Error(w, apierror.ErrInvalidArgument, "invalid authorization header format")
				return
			}

			claims, err := Parse(parts[1], publicKey)
			if err != nil {
				if errors.Is(err, ErrExpiredToken) {
					response.Error(w, apierror.ErrUnauthorized, "token has expired")
					return
				}
				response.Error(w, apierror.ErrUnauthorized, "invalid token")
				return
			}

			// only service tokens are accepted by internal services
			if claims.PrincipalType != types.PrincipalTypeService {
				response.Error(w, apierror.ErrUnauthorized, "only service tokens are accepted")
				return
			}

			// extract the acting principal from the header
			// this is the user on whose behalf the service is acting
			actingPrincipalID := r.Header.Get("X-Principal-ID")

			ctx := context.WithValue(r.Context(), contextKey{}, claims)
			ctx = context.WithValue(ctx, actingPrincipalKey{}, actingPrincipalID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromContext extracts the Claims from the request context.
// Returns false if no claims are present — i.e. the middleware was not applied.
func FromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(contextKey{}).(Claims)
	return claims, ok
}
