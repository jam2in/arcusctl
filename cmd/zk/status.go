package zk

import (
	"github.com/jam2in/arcusctl/internal/zk"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <ensemble-name>",
	Short: "Show status of a ZooKeeper ensemble",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := zk.Status(args[0]); err != nil {
			panic(err)
		}
	},
}
