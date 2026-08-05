package cluster

import (
	"github.com/jam2in/arcusctl/internal/cluster"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List managed Arcus clusters",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cluster.List(); err != nil {
			panic(err)
		}
	},
}
