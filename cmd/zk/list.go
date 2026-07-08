package zk

import (
	"github.com/jam2in/arcusctl/internal/zk"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List ZooKeeper ensembles",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := zk.List(); err != nil {
			panic(err)
		}
	},
}
