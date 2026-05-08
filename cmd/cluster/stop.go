package cluster

import "github.com/spf13/cobra"

var stopCmd = &cobra.Command{
	Use:   "stop <servicecode> [--node <address>]",
	Short: "Stop an Arcus cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: stop 구현
	},
}

func init() {
	stopCmd.Flags().String("node", "", "address of the specific node to stop")
}
