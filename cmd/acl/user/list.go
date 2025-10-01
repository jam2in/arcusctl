package user

import (
	"fmt"
	"os"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal/acl"
	"github.com/jam2in/arcus-cli/internal/types"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <groupName>",
	Short: "List all users in an ACL group.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		zkConn := cmd.Context().Value(types.CtxZkConnKey{}).(*zk.Conn)

		users, err := acl.GetUsers(zkConn, groupName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Printf("Users in group '%s':\n", groupName)
		for i, u := range users {
			fmt.Printf("  %d. %s\n", i+1, u)
		}
		fmt.Printf("Total %d users in group '%s'.\n", len(users), groupName)
	},
}
