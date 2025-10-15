package group

import (
	"fmt"

	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:  "list",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := internal.ConnectZooKeeper(internal.Config.ZooKeeper)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		groups, _, err := conn.Children(internal.ZPATH_ACL_ROOT)
		if err != nil {
			panic(err)
		}

		for _, g := range groups {
			fmt.Printf("  * %s\n", g)
		}
		fmt.Printf("Total: %d\n", len(groups))
	},
}
