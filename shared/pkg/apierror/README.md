Package apierror defines standard sentinel errors, machine-readable error codes,
and HTTP status mappings used across all services.

It is the single source of truth for error classification. Repository and domain
layers return sentinel errors from this package. The handler layer maps them to
HTTP status codes and error codes via the helper functions provided here.
No service should define its own sentinel errors for conditions already covered
here.

Folder structure:
  apierror/
  └── apierror.go    sentinel errors, ErrorCode type, HTTPStatus and Code functions

Sentinel errors:
  ErrNotFound        the requested entity does not exist
  ErrConflict        a unique constraint would be violated
  ErrForbidden       the caller lacks permission to perform the operation
  ErrInvalidArgument the request contains invalid or missing arguments
  ErrUnauthorized    the caller has no valid identity
  ErrInternal        an unexpected internal error occurred

Error codes (machine-readable, returned in API responses):
  NOT_FOUND
  CONFLICT
  FORBIDDEN
  INVALID_ARGUMENT
  UNAUTHORIZED
  INTERNAL_ERROR

Helper functions:
  HTTPStatus(err) int        maps a sentinel error to an HTTP status code
  Code(err) ErrorCode        maps a sentinel error to its machine-readable code