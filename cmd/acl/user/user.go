package user

import (
	"github.com/spf13/cobra"
)

var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage SASL users within an ACL group",
	Long:  "Manages users and their credentials within a specific ACL group.",
}

func init() {
	UserCmd.AddCommand(listCmd)
	UserCmd.AddCommand(addCmd)
	UserCmd.AddCommand(removeCmd)
}
