package group

import "github.com/spf13/cobra"

var GroupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage groups in Arcus ACL.",
	Long:  "Manages ACL groups, which act as containers for users in Arcus.",
}

func init() {
	GroupCmd.AddCommand(listCmd)
	GroupCmd.AddCommand(addCmd)
	GroupCmd.AddCommand(removeCmd)
}
