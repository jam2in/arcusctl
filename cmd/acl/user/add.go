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
	Use:  "add <group_name> <user_name> <permissions> [logAll]",
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
	Use:  "passwd <group_name> <user_name>",
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
	Use:  "permissions <group_name> <user_name> <permissions>",
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
