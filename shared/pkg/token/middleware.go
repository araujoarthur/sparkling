// middleware.go provides chi-compatible HTTP middleware for JWT token validation.
package token

import (
	"context"
	"crypto/rsa"
	"errors"
	"net/http"
	"strings"

	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/araujoarthur/intranetbackend/shared/pkg/response"
	"github.com/araujoarthur/intranetbackend/shared/pkg/types"
)

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
