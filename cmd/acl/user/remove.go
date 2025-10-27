package user

import (
	"fmt"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcusctl/internal"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:  "remove <group_name> <user_name>",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		userName := args[1]

		adminName := internal.ReadStdin("admin name", false)
		adminPassword := internal.ReadStdin("admin password", true)

		conn, err := internal.ConnectZooKeeper(internal.Config.ZooKeeper)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		if err := conn.AddAuth("digest", []byte(adminName+":"+adminPassword)); err != nil {
			panic(err)
		}

		if _, err := conn.Multi(
			&zk.DeleteRequest{
				Path:    internal.ZPATH_ACL_ROOT + "/" + groupName + "/" + userName + "/" + propName,
				Version: -1,
			},
			&zk.DeleteRequest{
				Path:    internal.ZPATH_ACL_ROOT + "/" + groupName + "/" + userName,
				Version: -1,
			},
		); err != nil {
			panic(err)
		}

		fmt.Println("OK")
	},
}
