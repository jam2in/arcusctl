package cluster

import "github.com/spf13/cobra"

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List managed Arcus clusters",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: list 구현
	},
}
