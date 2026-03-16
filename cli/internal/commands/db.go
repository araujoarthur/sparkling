package commands

import "github.com/spf13/cobra"

func DBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database management commands",
	}

	cmd.AddCommand(bootstrapCmd())
	cmd.AddCommand(migrateCmd())
	cmd.AddCommand(seedCmd())

	return cmd
}
