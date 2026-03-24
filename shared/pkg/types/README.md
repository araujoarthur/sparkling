Package types defines shared types used across all services and packages.
It exists to avoid import cycles — any type that would otherwise need to
be defined in a service-specific package but is referenced by shared
packages lives here instead.

No business logic belongs in this package. It contains only type definitions
and constants. If you find yourself adding functions here, the logic likely
belongs in a more specific package.

Folder structure:
  types/
  └── principal.go    PrincipalType and its constants

Types:
  PrincipalType    represents the kind of entity that can be assigned roles
                   either "user" or "service"

Constants:
  PrincipalTypeUser       "user"    — a human user registered via the auth service
  PrincipalTypeService    "service" — a service account registered via inetbctl