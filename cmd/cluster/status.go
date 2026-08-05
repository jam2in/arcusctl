package cluster

import (
	"github.com/jam2in/arcusctl/internal/cluster"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <servicecode>",
	Short: "Show status of an Arcus cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceCode := args[0]
		if err := cluster.Status(serviceCode); err != nil {
			panic(err)
		}
	},
}
