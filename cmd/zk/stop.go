package zk

import "github.com/spf13/cobra"

var stopCmd = &cobra.Command{
	Use:   "stop <ensemble-name> [--node <myid>]",
	Short: "Stop a ZooKeeper ensemble",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: stop 구현
	},
}

func init() {
	stopCmd.Flags().Int("node", 0, "myid of the specific node to stop")
}
