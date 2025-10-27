package group

import (
	"fmt"

	"github.com/jam2in/arcusctl/internal"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:  "remove <group_name>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]

		conn, err := internal.ConnectZooKeeper(internal.Config.ZooKeeper)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		if err := conn.Delete(internal.ZPATH_ACL_ROOT+"/"+groupName, -1); err != nil {
			panic(err)
		}

		fmt.Println("OK")
	},
}
