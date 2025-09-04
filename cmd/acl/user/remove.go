package user

import (
	"fmt"
	"path"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/config"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <groupName> <userName>",
	Short: "Remove a user from an ACL group.",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupName := args[0]
		userName := args[1]
		aclUser, _ := cmd.Flags().GetString("userName")
		aclPassword, _ := cmd.Flags().GetString("password")
		zkConn := cmd.Context().Value(config.ZkConnKey{}).(*zk.Conn)

		err := removeUser(zkConn, groupName, userName, aclUser, aclPassword)
		if err != nil {
			return err
		}

		fmt.Printf("User '%s' removed from group '%s' successfully.\n", userName, groupName)
		return nil
	},
}

func removeUser(zkConn *zk.Conn, groupName, userName, aclUser, aclPassword string) error {
	err := isAuth(zkConn, groupName, aclUser, aclPassword)
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
