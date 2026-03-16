package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/araujoarthur/intranetbackend/cli/internal/db"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func bootstrapCmd() *cobra.Command {
	var down bool

	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Creates schemas, roles and default privileges",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if down {
				if err := goose.Down(database, "db/migrations/global"); err != nil {
					return fmt.Errorf("bootstrap down failed: %w", err)
				}
				fmt.Println("Bootstrap reversed.")
				return nil
			}

			if err := goose.Up(database, "db/migrations/global"); err != nil {
				return fmt.Errorf("bootstrap failed: %w", err)
			}

			fmt.Println("Bootstrap Complete.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&down, "down", false, "Reverse the bootstrap")
	return cmd
}
