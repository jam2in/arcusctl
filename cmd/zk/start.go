package zk

import (
	"github.com/jam2in/arcusctl/internal/zk"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start <ensemble-name> [--node <myid>]",
	Short: "Start a ZooKeeper ensemble",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ensembleName := args[0]
		myID, _ := cmd.Flags().GetInt("node")

		if err := zk.Start(ensembleName, myID); err != nil {
			panic(err)
		}
	},
}

func init() {
	startCmd.Flags().Int("node", 0, "myid of the specific node to start")
}
