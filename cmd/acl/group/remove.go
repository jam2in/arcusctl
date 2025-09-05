package group

import (
	"fmt"
	"os"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <groupName>",
	Short: "Remove an empty ACL group.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]

		zkConn := cmd.Context().Value(internal.CtxZkConnKey{}).(*zk.Conn)

		if err := zkConn.Delete(internal.AclRootPath+"/"+groupName, -1); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Printf("ACL group '%s' removed successfully.\n", groupName)
	},
}
