package repository

import (
	"errors"
	"time"

	"github.com/araujoarthur/intranetbackend/services/iam/internal/repository/sqlc/generated"
	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/araujoarthur/intranetbackend/shared/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// --------------------------------
// Sentinel Errors
// --------------------------------
var (
	ErrNotFound        = apierror.ErrNotFound
	ErrConflict        = apierror.ErrConflict
	ErrForbidden       = apierror.ErrForbidden
	ErrInvalidArgument = apierror.ErrInvalidArgument
)

// --------------------------------
// Domain Types
// --------------------------------
type PrincipalType = types.PrincipalType // Type alias

const (
	PrincipalTypeUser    PrincipalType = types.PrincipalTypeUser
	PrincipalTypeService PrincipalType = types.PrincipalTypeService
)

// Role represents an IAM role that can be assigned to principals.
type Role struct {
	ID          uuid.UUID
	Name        string
	Description string
	IsSystem    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Permission represents a single action a role can perform
type Permission struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
}

// Principal represents any entity that can be assigned roles
type Principal struct {
	ID            uuid.UUID
	ExternalID    uuid.UUID
	PrincipalType PrincipalType
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// PrincipalRole represents the assignment of a role to a principal.
type PrincipalRole struct {
	PrincipalID uuid.UUID
	RoleID      uuid.UUID
	GrantedBy   uuid.UUID
	CreatedAt   time.Time
}

// --------------------------------
// Helpers
// --------------------------------

// pgxText converts a Go string into a pgtype.Text for use with nullable text columns.
// The value is marked valid only when the string is non-empty, so empty strings
// are stored as NULL rather than as empty strings in the database.
func pgxText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

// --------------------------------
// Error Mapper
// --------------------------------

// mapError translates low-level pgx errors into repository sentinel errors
// so callers never need to import or inspect pgx directly.
//
// Mappings:
//   - pgx.ErrNoRows → ErrNotFound
//   - all other errors are returned unchanged
func mapError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return ErrConflict
		}
	}

	return err
}

// --------------------------------
// Domain Mappers
// --------------------------------

// toRole converts a sqlc-generated IamRole into a domain Role.
// It unwraps nullable fields (Description) from their pgtype wrappers
// into plain Go types that the rest of the application can use directly.
func toRole(r *generated.IamRole) Role {
	return Role{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description.String,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// toPermission converts a sqlc-generated IamPermission into a domain Permission.
// It unwraps nullable fields (Description) from their pgtype wrappers
// into plain Go types that the rest of the application can use directly.
func toPermission(p *generated.IamPermission) Permission {
	return Permission{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description.String,
		CreatedAt:   p.CreatedAt,
	}
}

// toPrincipal converts a sqlc-generated IamPrincipal into a domain Principal.
// It casts the database enum PrincipalType into the domain PrincipalType constant
// so callers can compare against PrincipalTypeUser and PrincipalTypeService directly.
func toPrincipal(p *generated.IamPrincipal) Principal {
	return Principal{
		ID:            p.ID,
		ExternalID:    p.ExternalID,
		PrincipalType: PrincipalType(p.PrincipalType),
		IsActive:      p.IsActive,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// toPrincipalRole converts a sqlc-generated AssignRoleToPrincipalRow into a domain PrincipalRole.
func toPrincipalRole(r *generated.IamPrincipalRole) PrincipalRole {
	return PrincipalRole{
		PrincipalID: r.PrincipalID,
		RoleID:      r.RoleID,
		GrantedBy:   r.GrantedBy,
		CreatedAt:   r.CreatedAt,
	}
}
