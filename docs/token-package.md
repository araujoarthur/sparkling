# Token Package

`shared/pkg/token` is the shared authentication package for internal services.

It handles two related concerns:

- JWT creation and validation.
- HTTP request helpers that depend on token middleware context.

## Service Boundary

Internal services such as auth and IAM should expose only health checks without authentication.

All other routes should be inside:

```go
r.Group(func(r chi.Router) {
    r.Use(token.Middleware(publicKey))
    // protected routes
})
```

`token.Middleware` accepts service tokens only. User tokens are rejected.

This means calls to auth and IAM are expected to come from trusted application services, such as the future webapp, not directly from browsers or users.

## Token Types

User tokens:

- principal type: `user`
- short-lived
- currently expire after 15 minutes
- issued by auth login/refresh flows

Service tokens:

- principal type: `service`
- non-expiring JWTs
- accepted by internal service middleware
- revocation is managed by auth service state, not by JWT expiry

## Middleware Context

When `token.Middleware` accepts a request, it stores two values in context:

- validated service-token `Claims`
- raw `X-Principal-ID`, if the calling service provided one

The service token identifies the calling service. `X-Principal-ID` identifies the human/user principal the service is acting for.

## Effective Actor

Use `token.ActorFromContext` when a handler needs a `callerID`.

```go
callerID, err := token.ActorFromContext(r.Context())
if err != nil {
    response.Error(w, apierror.ErrUnauthorized, err.Error())
    return
}
```

Resolution order:

1. If `X-Principal-ID` is present, parse and return it.
2. Otherwise, return the service token subject.

This supports both forms:

```text
service acting for a user  -> X-Principal-ID is actor
service acting for itself  -> service token subject is actor
```

Use this for domain calls that enforce permissions or ownership, such as:

- IAM role creation/deletion
- IAM permission assignment
- IAM principal activation/deactivation
- auth identity deletion
- auth password change
- logout-all flows

## Acting Principal Only

`token.ActingPrincipalFromContext` returns only the raw `X-Principal-ID` value.

It does not fall back to service-token claims. Most handlers should use `ActorFromContext` instead.

Use `ActingPrincipalFromContext` only when you specifically need to know whether an acting-user header was supplied.

## Claims Only

`token.FromContext` returns service-token claims.

```go
claims, ok := token.FromContext(r.Context())
```

Use it when you specifically need the calling service's own identity, not the effective actor.

## URL UUID Parameters

Use `token.ParseUUIDParam` to parse chi route parameters:

```go
identityID, ok := token.ParseUUIDParam(w, r, "id")
if !ok {
    return
}
```

The raw `return` is intentional. `ParseUUIDParam` has already written a standard `response.Error` when it returns `false`.

Do not continue after `ok == false`, because the parsed UUID will be `uuid.Nil` and a response has already been written.

## Handler Pattern

For a protected mutating route:

```go
func (s *Server) deleteThing(w http.ResponseWriter, r *http.Request) {
    callerID, err := token.ActorFromContext(r.Context())
    if err != nil {
        response.Error(w, apierror.ErrUnauthorized, err.Error())
        return
    }

    id, ok := token.ParseUUIDParam(w, r, "id")
    if !ok {
        return
    }

    if err := s.service.Delete(r.Context(), callerID, id); err != nil {
        response.Error(w, err, "failed to delete thing")
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
```

## Responsibility Split

Handlers:

- apply middleware
- extract actor IDs
- parse URL parameters
- decode request bodies
- call domain services
- write responses

Domain services:

- enforce permissions
- enforce ownership
- validate business inputs
- issue tokens when auth owns the flow

Repositories:

- never import `token`

## Common Pitfalls

Do not use `ActingPrincipalFromContext` when you need a caller ID with service fallback. Use `ActorFromContext`.

Do not parse `X-Principal-ID` manually in every handler. Use `ActorFromContext`.

Do not write another error response after `ParseUUIDParam` returns `false`.

Do not expose internal service routes without `token.Middleware`, except health checks.
