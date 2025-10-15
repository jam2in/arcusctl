package cmd

import (
	"github.com/jam2in/arcus-cli/cmd/apply"
	"github.com/jam2in/arcus-cli/cmd/show"
	"github.com/jam2in/arcus-cli/config"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "arcus-cli",
		Short: "Arcus CLI",
		Long:  "Arcus CLI is a command line interface for managing Arcus",
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&config.Verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(apply.RootCmd)
	rootCmd.AddCommand(show.RootCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
