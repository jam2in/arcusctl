package group

import (
	"fmt"
	"path"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/config"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <groupName>",
	Short: "Remove an empty ACL group.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupName := args[0]
		adminUser, _ := cmd.Flags().GetString("userName")
		adminPassword, _ := cmd.Flags().GetString("password")
		zkConn := cmd.Context().Value(config.ZkConnKey{}).(*zk.Conn)

		err := removeGroup(zkConn, groupName, adminUser, adminPassword)
		if err != nil {
			return err
		}

		fmt.Printf("ACL group '%s' removed successfully.\n", groupName)
		return nil
	},
}

func init() {
	removeCmd.Flags().StringP("userName", "u", "", "Administrator user name for a group")
	removeCmd.Flags().StringP("password", "p", "", "Administrator password for a group")
	removeCmd.MarkFlagsRequiredTogether("userName", "password")
}

func removeGroup(zkConn *zk.Conn, groupName, adminUser, adminPassword string) error {
	groupPath := path.Join(config.AclRootPath, groupName)
	groupACL, _, err := zkConn.GetACL(groupPath)
	if err != nil {
		return err
	}
	if isDigest(groupACL) {
		if adminUser == "" || adminPassword == "" {
			return fmt.Errorf("'%s' is private group. Require credentials via -u and -p flags.\n", groupName)
		}
	}
	err = zkConn.Delete(groupPath, -1)
	return err
}

func isDigest(acl []zk.ACL) bool {
	for _, a := range acl {
		if a.Scheme == "digest" {
			return true
		}
	}
	return false
}
