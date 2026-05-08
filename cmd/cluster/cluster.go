package cluster

import "github.com/spf13/cobra"

var ClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage Arcus clusters",
}

func init() {
	ClusterCmd.AddCommand(deployCmd)
	ClusterCmd.AddCommand(startCmd)
	ClusterCmd.AddCommand(stopCmd)
	ClusterCmd.AddCommand(deleteCmd)
	ClusterCmd.AddCommand(listCmd)
	ClusterCmd.AddCommand(statusCmd)
}
