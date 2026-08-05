package cluster

import (
	"github.com/jam2in/arcusctl/internal/cluster"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <servicecode> [--node <address>] [--group <group-name>]",
	Short: "Stop an Arcus cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceCode := args[0]
		nodeAddress, _ := cmd.Flags().GetString("node")
		groupName, _ := cmd.Flags().GetString("group")
		if err := cluster.Stop(serviceCode, nodeAddress, groupName); err != nil {
			panic(err)
		}
	},
}

func init() {
	stopCmd.Flags().String("node", "", "address of the specific node to stop (community edition only)")
	stopCmd.Flags().String("group", "", "name of the specific group to stop (enterprise edition only)")
	stopCmd.MarkFlagsMutuallyExclusive("node", "group")
}
