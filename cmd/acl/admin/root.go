package admin

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "admin",
	Short: "Manage ARCUS ACL group administrators",
	Long:  `Manage administrative credentials for ARCUS ACL groups.`,
}

func init() {
	RootCmd.AddCommand(passwdCmd)
}
