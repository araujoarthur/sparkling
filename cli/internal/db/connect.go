package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// Uses PGX to connect to the database and returns the PGX connection and a *sql.DB
func Connect(dsn string) (*pgx.Conn, *sql.DB, error) {
	if dsn == "" {
		return nil, nil, fmt.Errorf("DSN is not set")
	}

	pgxconn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, nil, err
	}

	db := stdlib.OpenDB(*pgxconn.Config())
	return pgxconn, db, nil
}
