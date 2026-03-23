// response.go provides standard HTTP response envelope types and helper functions
// used across all services. It ensures every API response follows a consistent
// format regardless of which service sent it.
package response

import (
	"encoding/json"
	"net/http"

	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
)

// envelope wraps a success response with a data field.
type envelope struct {
	Data any `json:"data"`
}

// paginatedEnvelope wraps a paginated success response with data and meta fields.
type paginatedEnvelope struct {
	Data any  `json:"data"`
	Meta Meta `json:"meta"`
}

// Meta contains pagination metadata.
type Meta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// errorDetail is the inner error object returned in error responses.
type errorDetail struct {
	Code    apierror.ErrorCode `json:"code"`
	Message string             `json:"message"`
}

// errorEnvelope wraps an error response.
type errorEnvelope struct {
	Error errorDetail `json:"error"`
}

// JSON writes a simple success response with the given status code and data.
func JSON(w http.ResponseWriter, status int, data any) {
	write(w, status, envelope{Data: data})
}

// Paginated writes a paginated success response with data and pagination metadata.
func Paginated(w http.ResponseWriter, status int, data any, page, perPage, total int) {
	totalPages := total / perPage
	if total%perPage != 0 {
		totalPages++
	}
	write(w, status, paginatedEnvelope{
		Data: data,
		Meta: Meta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// Error writes a structured error response, automatically mapping the error
// to the appropriate HTTP status code and machine-readable error code.
func Error(w http.ResponseWriter, err error, message string) {
	write(w, apierror.HTTPStatus(err), errorEnvelope{
		Error: errorDetail{
			Code:    apierror.Code(err),
			Message: message,
		},
	})
}

// write serializes the given value as JSON and writes it to the response.
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
