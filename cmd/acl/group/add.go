package group

import (
	"fmt"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcusctl/internal"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <group_name>",
	Short: "Create a new ACL group",
	Long: `Create a new ACL group with the specified name.

You will be prompted to enter an admin name and password for managing
this group. The admin credentials are used for controlling access to
the group and its users.

For password requirements, see: https://github.com/jam2in/arcusctl/blob/main/docs/command-acl.md`,
	Example: `  # Create a new group named 'cache01'
  arcusctl acl group add cache01`,
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
