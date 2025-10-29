package user

import (
	"fmt"
	"strings"

	"github.com/jam2in/arcusctl/internal"
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

			// It's pretty stupid
			permList := strings.Split(string(perm), ",")
			if permList[len(permList)-1] == "logall" {
				fmt.Printf("  * %s %+v logAll\n", u, permList[:len(permList)-1])
			} else {
				fmt.Printf("  * %s %+v\n", u, permList)
			}
		}
		fmt.Printf("Total: %d\n", len(users))
	},
}
