package acl

import (
	"github.com/jam2in/arcusctl/cmd/acl/admin"
	"github.com/jam2in/arcusctl/cmd/acl/group"
	"github.com/jam2in/arcusctl/cmd/acl/user"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "acl",
	Short: "Manage ARCUS Access Control Lists",
	Long:  `Manage ARCUS Access Control Lists including users, groups, and administrators.`,
}

func init() {
	RootCmd.AddCommand(group.RootCmd)
	RootCmd.AddCommand(admin.RootCmd)
	RootCmd.AddCommand(user.RootCmd)
}
