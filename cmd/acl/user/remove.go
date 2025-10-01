package user

import (
	"fmt"
	"os"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal/acl"
	"github.com/jam2in/arcus-cli/internal/types"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <groupName> <userName>",
	Short: "Remove a user from an ACL group.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		userName := args[1]
		zkConn := cmd.Context().Value(types.CtxZkConnKey{}).(*zk.Conn)

		if err := acl.RemoveUser(zkConn, groupName, userName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("User '%s' removed from group '%s' successfully.\n", userName, groupName)
	},
}
