package group

import (
	"fmt"
	"path"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/config"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <groupName>",
	Short: "Add a new ACL group.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupName := args[0]
		aclUser, _ := cmd.Flags().GetString("userName")
		aclPassword, _ := cmd.Flags().GetString("password")
		zkConn := cmd.Context().Value(config.ZkConnKey{}).(*zk.Conn)

		err := addGroup(zkConn, groupName, aclUser, aclPassword)
		if err != nil {
			return err
		}

		fmt.Printf("ACL group '%s' created successfully.\n", groupName)
		return nil
	},
}

func addGroup(zkConn *zk.Conn, groupName, aclUser, aclPassword string) error {
	groupPath := path.Join(config.AclRootPath, groupName)
	var acl []zk.ACL
	if aclUser != "" && aclPassword != "" {
		adminACL := zk.DigestACL(zk.PermAll, aclUser, aclPassword)
		worldReadACL := zk.WorldACL(zk.PermRead)
		acl = append(adminACL, worldReadACL...)
	} else {
		acl = zk.WorldACL(zk.PermAll)
	}

	_, err := zkConn.Create(groupPath, nil, 0, acl)

	return err
}
