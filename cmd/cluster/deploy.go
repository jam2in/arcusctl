package cluster

import "github.com/spf13/cobra"

var deployCmd = &cobra.Command{
	Use:   "deploy <version> <topology.yml>",
	Short: "Deploy a new Arcus cluster",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: deploy 구현
	},
}
