// apierror.go defines standard sentinel errors and error codes used across all services.
// It provides a central mapping from sentinel errors to HTTP status codes so handlers
// never need to implement their own error-to-status logic.
package apierror

import (
	"errors"
	"net/http"
)

// ErrorCode is a machine-readable error identifier returned in API responses.
type ErrorCode string

const (
	CodeNotFound        ErrorCode = "NOT_FOUND"
	CodeConflict        ErrorCode = "CONFLICT"
	CodeForbidden       ErrorCode = "FORBIDDEN"
	CodeInvalidArgument ErrorCode = "INVALID_ARGUMENT"
	CodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	CodeInternal        ErrorCode = "INTERNAL_ERROR"
)

// Sentinel errors returned by the repository and domain layers.
// Handlers should never inspect raw pgx or database errors directly —
// all errors should be mapped to these before reaching the handler.
var (
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("already exists")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInternal        = errors.New("internal error")
)

// HTTPStatus maps a sentinel error to the appropriate HTTP status code.
// Falls back to 500 Internal Server Error for unrecognized errors.
func HTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrInvalidArgument):
		return http.StatusBadRequest
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// Code maps a sentinel error to its machine-readable ErrorCode.
// Falls back to CodeInternal for unrecognized errors.
func Code(err error) ErrorCode {
	switch {
	case errors.Is(err, ErrNotFound):
		return CodeNotFound
	case errors.Is(err, ErrConflict):
		return CodeConflict
	case errors.Is(err, ErrForbidden):
		return CodeForbidden
	case errors.Is(err, ErrInvalidArgument):
		return CodeInvalidArgument
	case errors.Is(err, ErrUnauthorized):
		return CodeUnauthorized
	default:
		return CodeInternal
	}
}
