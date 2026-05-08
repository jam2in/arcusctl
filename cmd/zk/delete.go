package zk

import "github.com/spf13/cobra"

var deleteCmd = &cobra.Command{
	Use:   "delete <ensemble-name>",
	Short: "Delete a ZooKeeper ensemble",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: delete 구현
	},
}
