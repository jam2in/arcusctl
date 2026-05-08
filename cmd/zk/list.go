package zk

import "github.com/spf13/cobra"

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List ZooKeeper ensembles",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: list 구현
	},
}
