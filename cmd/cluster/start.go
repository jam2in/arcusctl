package cluster

import "github.com/spf13/cobra"

var startCmd = &cobra.Command{
	Use:   "start <servicecode> [--node <address>]",
	Short: "Start an Arcus cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: start 구현
	},
}

func init() {
	startCmd.Flags().String("node", "", "address of the specific node to start")
}
