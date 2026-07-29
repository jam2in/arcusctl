package cluster

import (
	"github.com/jam2in/arcusctl/internal/cluster"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy <version> <topology.yml>",
	Short: "Deploy a new Arcus cluster",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		version := args[0]
		topologyPath := args[1]

		if err := cluster.Deploy(version, topologyPath); err != nil {
			panic(err)
		}
	},
}
