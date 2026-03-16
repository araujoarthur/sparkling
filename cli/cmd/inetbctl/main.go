package main

import (
	"os"

	"github.com/araujoarthur/intranetbackend/cli/internal/commands"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "inetbctl",
		Short: "Management CLI for Intranet Backend",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			_ = godotenv.Load()
			return nil
		},
	}

	rootCmd.AddCommand(commands.DBCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
