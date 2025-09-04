package user

import (
	"fmt"
	"path"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-sasl-passwd/config"
	"github.com/spf13/cobra"
)

const propName = "authPassword"

type UserInfo struct {
	Username string
	Roles    []string
}

func (i UserInfo) String() string {
	return fmt.Sprintf("Username: %s, Role: %s", i.Username, i.Roles)
}

var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage SASL users within an ACL group",
	Long:  "Manages users and their credentials within a specific ACL group.",
}

func init() {

	UserCmd.AddCommand(listCmd)
	UserCmd.AddCommand(addCmd)
	UserCmd.AddCommand(removeCmd)
}

func isAuth(zkConn *zk.Conn, groupName, adminUser, adminPassword string) ([]zk.ACL, error) {
	groupPath := path.Join(config.AclRootPath, groupName)
	groupACL, _, err := zkConn.GetACL(groupPath)
	if err != nil {
		return nil, err
	}
	if isDigest(groupACL) {
		fmt.Println("fhdahpof")
		if adminUser == "" || adminPassword == "" {
			return nil, fmt.Errorf("'%s' is private group. Require credentials via -u and -p flags.\n", groupName)
		}
		auth := []byte(adminUser + ":" + adminPassword)
		if err := zkConn.AddAuth("digest", auth); err != nil {
			return nil, err
		}
	}
	return groupACL, nil
}

func isDigest(acl []zk.ACL) bool {
	for _, a := range acl {
		if a.Scheme == "digest" {
			return true
		}
	}
	return false
}
