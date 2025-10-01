package group

import (
	"fmt"
	"os"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal/acl"
	"github.com/jam2in/arcus-cli/internal/types"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <groupName>",
	Short: "Add a new ACL group.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		zkConn := cmd.Context().Value(types.CtxZkConnKey{}).(*zk.Conn)
		zkAcl := cmd.Context().Value(types.CtxZkAclKey{}).([]zk.ACL)

		if err := acl.AddGroup(zkConn, zkAcl, groupName); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("ACL group '%s' created successfully.\n", groupName)
	},
}
