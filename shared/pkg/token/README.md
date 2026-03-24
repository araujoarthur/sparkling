Package token provides JWT token issuance, validation, and HTTP middleware
for service-to-service authentication using RS256 asymmetric signing.

Only the auth service and inetbctl should call Issue — all other services
only call Parse or wire the Middleware. This is enforced by convention:
services without a private key cannot issue tokens meaningfully even if
the function is available to them.

Token format:
  Algorithm:  RS256
  Claims:
    sub               principal ID in IAM (uuid)
    principal_type    "user" or "service"
    iat               issued at
    exp               expiry

Token durations:
  User tokens        15 minutes — refreshed via auth service refresh token flow
  Service tokens     non-expiring   — rotated daily by auth or manually via inetbctl token rotate

Authentication model:
  Internal services only accept service tokens in the Authorization header.
  The acting user is passed separately via the X-Principal-ID header.
  This prevents users from calling internal services directly.

  Authorization: Bearer <service-token>
  X-Principal-ID: <user-principal-uuid>

Folder structure:
  token/
  ├── token.go        Claims struct, jwtClaims, Issue, Parse, sentinel errors
  └── middleware.go   Middleware, FromContext, ActingPrincipalFromContext

Sentinel errors:
  ErrExpiredToken    the token has expired — caller should attempt a refresh
  ErrInvalidToken    the token is malformed or the signature is invalid

Helper functions:
  Issue(claims, privateKey)        signs and returns a JWT token
  Parse(tokenString, publicKey)    validates and parses a JWT token
  Middleware(publicKey)            chi middleware, validates token and injects
                                   claims and acting principal into context
  FromContext(ctx)                 extracts Claims from request context
  ActingPrincipalFromContext(ctx)  extracts the acting principal ID from context