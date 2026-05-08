package zk

import "github.com/spf13/cobra"

var ZKCmd = &cobra.Command{
	Use:   "zk",
	Short: "Manage ZooKeeper ensembles",
}

func init() {
	ZKCmd.AddCommand(deployCmd)
	ZKCmd.AddCommand(startCmd)
	ZKCmd.AddCommand(stopCmd)
	ZKCmd.AddCommand(deleteCmd)
	ZKCmd.AddCommand(listCmd)
	ZKCmd.AddCommand(statusCmd)
}
