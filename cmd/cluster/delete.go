package cluster

import "github.com/spf13/cobra"

var deleteCmd = &cobra.Command{
	Use:   "delete <servicecode>",
	Short: "Delete an Arcus cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: delete 구현
	},
}
