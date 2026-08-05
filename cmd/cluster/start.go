package cluster

import (
	"github.com/jam2in/arcusctl/internal/cluster"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start <servicecode> [--node <address>] [--group <group-name>]",
	Short: "Start an Arcus cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceCode := args[0]
		nodeAddress, _ := cmd.Flags().GetString("node")
		groupName, _ := cmd.Flags().GetString("group")
		if err := cluster.Start(serviceCode, nodeAddress, groupName); err != nil {
			panic(err)
		}
	},
}

func init() {
	startCmd.Flags().String("node", "", "address of the specific node to start (community edition only)")
	startCmd.Flags().String("group", "", "name of the specific group to start (enterprise edition only)")
	startCmd.MarkFlagsMutuallyExclusive("node", "group")
}
