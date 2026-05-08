package zk

import "github.com/spf13/cobra"

var deployCmd = &cobra.Command{
	Use:   "deploy <version> <topology.yml>",
	Short: "Deploy a new ZooKeeper ensemble",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: deploy 구현
	},
}
