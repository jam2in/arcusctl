package user

import (
	"fmt"

	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:  "list <group_name>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]

		conn, err := internal.ConnectZooKeeper(internal.Config.ZooKeeper)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		users, _, err := conn.Children(internal.ZPATH_ACL_ROOT + "/" + groupName)
		if err != nil {
			panic(err)
		}

		for _, u := range users {
			perm, _, err := conn.Get(internal.ZPATH_ACL_ROOT + "/" + groupName + "/" + u)
			if err != nil {
				panic(err)
			}
			fmt.Printf("  * %s (%s)\n", u, perm)
		}
		fmt.Printf("Total: %d\n", len(users))
	},
}
