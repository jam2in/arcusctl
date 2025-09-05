package group

import (
	"fmt"
	"os"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ACL groups.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		zkConn := cmd.Context().Value(internal.CtxZkConnKey{}).(*zk.Conn)

		groups, _, err := zkConn.Children(internal.AclRootPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Println("ACL Groups:")
		for i, g := range groups {
			fmt.Printf("  %d. %s\n", i+1, g)
		}
		fmt.Printf("Total %d groups.\n", len(groups))
	},
}
