Package response provides standard HTTP response envelope types and helper
functions used across all services. It ensures every API response follows
a consistent format regardless of which service sent it.

Every handler should use this package to write responses — never write
directly to http.ResponseWriter or use json.NewEncoder outside of this package.

Response formats:

  Simple success:
    {
        "data": { ... }
    }

  Paginated success:
    {
        "data": [ ... ],
        "meta": {
            "page": 1,
            "per_page": 20,
            "total": 100,
            "total_pages": 5
        }
    }

  Error:
    {
        "error": {
            "code": "NOT_FOUND",
            "message": "role with id abc123 does not exist"
        }
    }

Folder structure:
  response/
  └── response.go    envelope types, JSON, Paginated and Error helper functions

Helper functions:
  JSON(w, status, data)                      writes a simple success response
  Paginated(w, status, data, page, perPage, total)  writes a paginated response
  Error(w, err, message)                     writes a structured error response,
                                             automatically mapping the error to
                                             the correct HTTP status code and
                                             machine-readable error code via apierror