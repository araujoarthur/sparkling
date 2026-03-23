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

func seedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seed",
		Short: "Seed database with initial data",
	}

	cmd.AddCommand(seedUpCmd())
	cmd.AddCommand(seedDownCmd())
	cmd.AddCommand(seedStatusCmd())

	return cmd
}

func seedUpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "up [service]",
		Short: "Run seeds for a service (or all)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSeed(args, goose.Up)
		},
	}
}

func seedDownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "down [service]",
		Short: "Reverse seeds for a service (or all)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSeed(args, goose.Down)
		},
	}
}

func seedStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [service]",
		Short: "Show seed status for a service (or all)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSeed(args, goose.Status)
		},
	}
}

func runSeed(args []string, fn func(*sql.DB, string, ...goose.OptionsFunc) error) error {
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

	seedsRoot := os.Getenv("SEEDS_DIR")
	if seedsRoot == "" {
		seedsRoot = "db/seeds"
	}

	var targets []string
	if len(args) == 1 {
		targets = []string{args[0]} // the list of targets is only the passed value
	} else {
		targets, err = discoverServices(seedsRoot)
		if err != nil {
			return fmt.Errorf("failed to discover services: %w", err)
		}
	}

	for _, service := range targets {
		dir := fmt.Sprintf("%s/%s", seedsRoot, service)

		goose.SetTableName(fmt.Sprintf("%s.goose_seeds_version", service))

		fmt.Printf("--- %s ---\n", service)
		if err := fn(database, dir); err != nil {
			return fmt.Errorf("seed failed for %s: %w", service, err)
		}
	}

	return nil
}
