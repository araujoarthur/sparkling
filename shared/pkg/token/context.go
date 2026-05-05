package token

import "context"

// contextKey is a private type for storing claims in request context.
// Using a private type prevents collisions with other packages.
type contextKey struct{}

// FromContext extracts the Claims from the request context.
// Returns false if no claims are present — i.e. the middleware was not applied.
func FromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(contextKey{}).(Claims)
	return claims, ok
}
