package user

import (
	"fmt"
	"strings"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcusctl/internal"
	"github.com/jam2in/arcusctl/internal/scram"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <group_name> <user_name> <permissions> [logAll]",
	Short: "Add a new user to a group",
	Long: `Add a new user to the specified ACL group with the given permissions.

Permissions should be specified as a comma-separated list.
Optionally, you can enable full logging by adding 'logAll' as the fourth argument.

For password requirements, see: https://github.com/jam2in/arcusctl/blob/main/docs/command-acl.md`,
	Example: `  # Add a user 'myapp' for application
  arcusctl acl user add cache01 myapp kv,list,set,map,btree,attr,scan,flush

  # Add a user 'john' for operator with logging enabled
  arcusctl acl user add cache01 john attr,scan,flush,admin logAll`,
	Args: cobra.RangeArgs(3, 4),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		userName := args[1]
		permissions := args[2]
		if len(args) == 4 {
			if args[3] == "logAll" {
				permissions += ",logall"
			} else {
				panic("invalid arguments")
			}
		}

		adminName := internal.ReadStdin("admin name", false)
		adminPassword := internal.ReadStdin("admin password", true)
		userPassword := internal.ReadStdin("user password", true)
		if userPassword != internal.ReadStdin("repeat user password", true) {
			panic("password does not match")
		}

		secret := scram.GenerateScramSHA256Secret(userPassword, nil, 0)

		conn, err := internal.ConnectZooKeeper(internal.Config.ZooKeeper)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		if err := conn.AddAuth("digest", []byte(adminName+":"+adminPassword)); err != nil {
			panic(err)
		}

		acls := append(zk.DigestACL(zk.PermAll, adminName, adminPassword),
			zk.WorldACL(zk.PermRead)...)

		if _, err := conn.Multi(
			&zk.CreateRequest{
				Path:  internal.ZPATH_ACL_ROOT + "/" + groupName + "/" + userName,
				Data:  []byte(permissions),
				Acl:   acls,
				Flags: 0,
			},
			&zk.CreateRequest{
				Path:  internal.ZPATH_ACL_ROOT + "/" + groupName + "/" + userName + "/" + propName,
				Data:  []byte(secret.EncodeToBase64()),
				Acl:   acls,
				Flags: 0,
			},
		); err != nil {
			panic(err)
		}

		fmt.Println("OK")
	},
}

var passwdCmd = &cobra.Command{
	Use:   "passwd <group_name> <user_name>",
	Short: "Change a user's password",
	Long: `Change the password for the specified user in the given group.

For password requirements, see: https://github.com/jam2in/arcusctl/blob/main/docs/command-acl.md`,
	Example: `  # Change password for user 'john' in group 'cache01'
  arcusctl acl user passwd cache01 john`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		userName := args[1]

		adminName := internal.ReadStdin("admin name", false)
		adminPassword := internal.ReadStdin("admin password", true)
		userPassword := internal.ReadStdin("user password", true)
		if userPassword != internal.ReadStdin("repeat user password", true) {
			panic("password does not match")
		}

		secret := scram.GenerateScramSHA256Secret(userPassword, nil, 0)

		conn, err := internal.ConnectZooKeeper(internal.Config.ZooKeeper)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		if err := conn.AddAuth("digest", []byte(adminName+":"+adminPassword)); err != nil {
			panic(err)
		}

		if _, err := conn.Set(internal.ZPATH_ACL_ROOT+"/"+groupName+"/"+userName+"/"+propName,
			[]byte(secret.EncodeToBase64()), -1); err != nil {
			panic(err)
		}

		fmt.Println("OK")
	},
}

var permissionsCmd = &cobra.Command{
	Use:   "permissions <group_name> <user_name> <permissions>",
	Short: "Update a user's permissions",
	Long: `Update the permissions for the specified user in the given group.

Permissions should be specified as a comma-separated list.
The logAll flag will be preserved from the user's existing configuration.`,
	Example: `  # Update user permissions to key-value only
  arcusctl acl user permissions cache01 john kv`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		userName := args[1]
		permissions := args[2]

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

		// It's pretty stupid
		beforePerm, _, err := conn.Get(internal.ZPATH_ACL_ROOT + "/" + groupName + "/" + userName)
		if err != nil {
			panic(err)
		}
		beforePermList := strings.Split(string(beforePerm), ",")
		if beforePermList[len(beforePermList)-1] == "logall" {
			permissions += ",logall"
		}

		if _, err := conn.Set(internal.ZPATH_ACL_ROOT+"/"+groupName+"/"+userName,
			[]byte(permissions), -1); err != nil {
			panic(err)
		}

		fmt.Println("OK")
	},
}
