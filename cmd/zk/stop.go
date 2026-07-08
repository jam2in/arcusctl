package zk

import (
	"github.com/jam2in/arcusctl/internal/zk"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <ensemble-name> [--node <myid>]",
	Short: "Stop a ZooKeeper ensemble",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ensembleName := args[0]
		myID, _ := cmd.Flags().GetInt("node")

		if err := zk.Stop(ensembleName, myID); err != nil {
			panic(err)
		}
	},
}

func init() {
	stopCmd.Flags().Int("node", 0, "myid of the specific node to stop")
}
