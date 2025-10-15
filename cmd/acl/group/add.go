package group

import (
	"fmt"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:  "add <group_name>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]

		adminName := internal.ReadStdin("admin name", false)
		adminPassword := internal.ReadStdin("password", true)
		if adminPassword != internal.ReadStdin("repeat password", true) {
			panic("password does not match")
		}

		conn, err := internal.ConnectZooKeeper(internal.Config.ZooKeeper)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		if err := conn.AddAuth("digest", []byte(adminName+":"+adminPassword)); err != nil {
			panic(err)
		}

		internal.EnsureZPath(conn, internal.ZPATH_ACL_ROOT)
		if _, err := conn.Create(internal.ZPATH_ACL_ROOT+"/"+groupName, nil, 0,
			append(zk.DigestACL(zk.PermAll, adminName, adminPassword),
				zk.WorldACL(zk.PermRead)...)); err != nil {
			panic(err)
		}

		fmt.Println("OK")
	},
}
