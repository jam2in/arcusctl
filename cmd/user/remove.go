package user

import (
	"fmt"
	"path"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-sasl-passwd/config"
	"github.com/jam2in/arcus-sasl-passwd/internal/zookeeper"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <groupName> <userName>",
	Short: "Remove a user from an ACL group.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupName := args[0]
		userName := args[1]
		adminUser, _ := cmd.Flags().GetString("userName")
		adminPassword, _ := cmd.Flags().GetString("password")

		err := removeUser(groupName, userName, adminUser, adminPassword)
		if err != nil {
			return err
		}

		fmt.Printf("User '%s' removed from group '%s' successfully.\n", userName, groupName)
		return nil
	},
}

func removeUser(groupName, userName, adminUser, adminPassword string) error {
	zkConn, err := zookeeper.NewConnect()
	if err != nil {
		return err
	}
	defer zkConn.Close()

	_, err = isAuth(zkConn, groupName, adminUser, adminPassword)
	if err != nil {
		return err
	}

	userPath := path.Join(config.AclRootPath, groupName, userName)
	authPath := path.Join(userPath, propName)

	ops := []interface{}{
		&zk.DeleteRequest{Path: authPath, Version: -1},
		&zk.DeleteRequest{Path: userPath, Version: -1},
	}
	_, err = zkConn.Multi(ops...)

	return err
}
