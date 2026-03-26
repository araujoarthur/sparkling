package helpers

import (
	"errors"
	"time"

	"github.com/araujoarthur/intranetbackend/shared/pkg/apierror"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// pgxText converts a Go string into a pgtype.Text for use with nullable text columns.
// The value is marked valid only when the string is non-empty, so empty strings
// are stored as NULL rather than as empty strings in the database.
func PgxText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

// MapError translates low-level pgx errors into repository sentinel errors
// so callers never need to import or inspect pgx directly.
//
// Mappings:
//   - pgx.ErrNoRows → apierror.ErrNotFound
//   - unique violation (23505) → apierror.ErrConflict
//   - all other errors are returned unchanged
func MapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return apierror.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return apierror.ErrConflict
		}
	}

	return err
}

// FromNullableTime converts a pgtype.Timestamptz into a *time.Time.
// Returns nil if the value is not valid (i.e. the column is NULL in the database).
// Use this in repository mappers when converting nullable timestamptz columns
// into domain types.
func FromNullableTime(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}

	return &t.Time
}
