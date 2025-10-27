package user

import (
	"github.com/spf13/cobra"
)

const propName = "authPassword"

var RootCmd = &cobra.Command{
	Use: "user",
}

func init() {
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(addCmd)
	RootCmd.AddCommand(passwdCmd)
	RootCmd.AddCommand(permissionsCmd)
	RootCmd.AddCommand(removeCmd)
}
