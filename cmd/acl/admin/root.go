package admin

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use: "admin",
}

func init() {
	RootCmd.AddCommand(passwdCmd)
}
