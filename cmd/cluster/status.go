package cluster

import (
	"fmt"

	"github.com/jam2in/arcusctl/internal/cluster"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <servicecode>",
	Short: "Show status of an Arcus cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("servicecode is required\nUsage: %s", cmd.UseLine())
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cluster.Status(args[0])
	},
}
