package cluster

import (
	"github.com/jam2in/arcusctl/internal/cluster"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <servicecode>",
	Short: "Delete an Arcus cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceCode := args[0]
		purge, _ := cmd.Flags().GetBool("purge")
		if err := cluster.Delete(serviceCode, purge); err != nil {
			panic(err)
		}
	},
}

func init() {
	deleteCmd.Flags().Bool("purge", false, "also remove the installation directory on each server")
}
