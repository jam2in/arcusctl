package zk

import (
	"github.com/jam2in/arcusctl/internal/zk"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <ensemble-name>",
	Short: "Delete a ZooKeeper ensemble",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		purge, _ := cmd.Flags().GetBool("purge")
		if err := zk.Delete(args[0], purge); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	deleteCmd.Flags().Bool("purge", false, "also remove the installation directory on each server")
}
