package commands

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/araujoarthur/intranetbackend/cli/internal/db"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migration",
	}

	cmd.AddCommand(migrateUpCmd())
	cmd.AddCommand(migrateDownCmd())
	cmd.AddCommand(migrateStatusCmd())

	return cmd
}

func migrateUpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "up [service]",
		Short: "Run migrations up for a service or for all services",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigration(args, goose.Up)
		},
	}
}

func migrateDownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "down [service]",
		Short: "Run one migration down for a service (or all)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigration(args, goose.Down)
		},
	}
}

func migrateStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [service]",
		Short: "Show migration status for a service (or all)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigration(args, goose.Status)
		},
	}
}

func runMigration(args []string, fn func(*sql.DB, string, ...goose.OptionsFunc) error) error {
	dsn := os.Getenv("OWNER_DSN")
	if dsn == "" {
		return fmt.Errorf("OWNER_DSN environment variable not set")
	}

	conn, database, err := db.Connect(dsn)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close(context.Background())
	defer database.Close()

	migrationsRoot := os.Getenv("MIGRATIONS_DIR")
	if migrationsRoot == "" {
		migrationsRoot = "db/migrations" // sensible default
	}

	var targets []string
	if len(args) == 1 {
		// specific service requested
		targets = []string{args[0]}
	} else {
		// discover services from filesystem
		targets, err = discoverServices(migrationsRoot)
		if err != nil {
			return fmt.Errorf("failed to discover services: %w", err)
		}
	}

	for _, service := range targets {
		dir := fmt.Sprintf("%s/%s", migrationsRoot, service)

		// Schema isolation for goose migration ID.
		goose.SetTableName(fmt.Sprintf("%s.goose_db_version", service))

		fmt.Printf("--- %s ---\n", service)
		if err := fn(database, dir); err != nil {
			return fmt.Errorf("migration failed for %s: %w", service, err)
		}
	}

	return nil
}

// Navigates the migrations folder and discovers all services
func discoverServices(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("cannot read migrations directory %s: %w", root, err)
	}

	var services []string
	for _, e := range entries {
		if e.IsDir() && e.Name() != "global" {
			services = append(services, e.Name())
		}
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no service directories found in %s", root)
	}

	return services, nil
}
