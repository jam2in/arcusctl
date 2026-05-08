package cluster

import "github.com/spf13/cobra"

var statusCmd = &cobra.Command{
	Use:   "status <servicecode>",
	Short: "Show status of an Arcus cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: status 구현
	},
}
