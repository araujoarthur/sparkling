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
)

// contextKey is a private type for storing claims in request context.
// Using a private type prevents collisions with other packages.
type contextKey struct{}

// Middleware returns a chi-compatible middleware that validates the Bearer token
// in the Authorization header and injects the parsed Claims into the request context.
// Requests with missing, invalid, or expired tokens are rejected with 401.
func Middleware(publicKey *rsa.PublicKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// extract the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, apierror.ErrUnauthorized, "missing authorization header")
				return
			}

			// validate the Bearer format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Error(w, apierror.ErrInvalidArgument, "invalid authorization header format")
				return
			}

			// parse and validate the token
			claims, err := Parse(parts[1], publicKey)
			if err != nil {
				if errors.Is(err, ErrExpiredToken) {
					response.Error(w, apierror.ErrUnauthorized, "token has expired")
					return
				}
				response.Error(w, apierror.ErrUnauthorized, "invalid token")
				return
			}

			// inject claims into context and pass to next handler
			ctx := context.WithValue(r.Context(), contextKey{}, claims)
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
