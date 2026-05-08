package zk

import "github.com/spf13/cobra"

var startCmd = &cobra.Command{
	Use:   "start <ensemble-name> [--node <myid>]",
	Short: "Start a ZooKeeper ensemble",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: start 구현
	},
}

func init() {
	startCmd.Flags().Int("node", 0, "myid of the specific node to start")
}
