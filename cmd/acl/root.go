package acl

import (
	"github.com/jam2in/arcusctl/cmd/acl/group"
	"github.com/jam2in/arcusctl/cmd/acl/user"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use: "acl",
}

func init() {
	RootCmd.AddCommand(group.RootCmd)
	RootCmd.AddCommand(user.RootCmd)
}
