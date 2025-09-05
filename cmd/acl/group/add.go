package group

import (
	"fmt"
	"os"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <groupName>",
	Short: "Add a new ACL group.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]

		zkConn := cmd.Context().Value(internal.CtxZkConnKey{}).(*zk.Conn)
		acl := cmd.Context().Value(internal.CtxZkAclKey{}).([]zk.ACL)

		_, err := zkConn.Create(internal.AclRootPath+"/"+groupName, nil, 0, acl)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Printf("ACL group '%s' created successfully.\n", groupName)
	},
}
