package group

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage ARCUS ACL groups",
	Long:  `Manage ARCUS ACL groups, which act as containers for users.`,
}

func init() {
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(addCmd)
	RootCmd.AddCommand(removeCmd)
}
