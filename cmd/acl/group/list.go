package group

import (
	"fmt"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ACL groups.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		zkConn := cmd.Context().Value(config.ZkConnKey{}).(*zk.Conn)
		groups, err := listGroups(zkConn)
		if err != nil {
			return err
		}

		fmt.Println("ACL Groups:")
		for i, g := range groups {
			fmt.Printf("  %d. %s\n", i+1, g)
		}
		fmt.Printf("Total %d groups.\n", len(groups))
		return nil
	},
}

func listGroups(zkConn *zk.Conn) ([]string, error) {
	groups, _, err := zkConn.Children(config.AclRootPath)

	return groups, err
}
