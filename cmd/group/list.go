package group

import (
	"fmt"

	"github.com/jam2in/arcus-sasl-passwd/config"
	"github.com/jam2in/arcus-sasl-passwd/internal/zookeeper"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ACL groups.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		groups, err := listGroups()
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

func listGroups() ([]string, error) {
	zkConn, err := zookeeper.NewConnect()
	if err != nil {
		return nil, err
	}
	defer zkConn.Close()

	groups, _, err := zkConn.Children(config.AclRootPath)

	return groups, err
}
