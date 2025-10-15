package group

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use: "group",
}

func init() {
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(addCmd)
	RootCmd.AddCommand(removeCmd)
}
