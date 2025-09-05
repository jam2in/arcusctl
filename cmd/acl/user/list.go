package user

import (
	"fmt"
	"path"
	"strings"

	"github.com/jam2in/arcus-cli/config"
	"github.com/jam2in/arcus-cli/internal/zookeeper"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <groupName>",
	Short: "List all users in an ACL group.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupName := args[0]
		users, err := listUsers(groupName)
		if err != nil {
			return err
		}

		fmt.Printf("Users in group '%s':\n", groupName)
		for i, u := range users {
			fmt.Printf("  %d. %s\n", i+1, u)
		}
		fmt.Printf("Total %d users in group '%s'.\n", len(users), groupName)
		return nil
	},
}

func listUsers(groupName string) ([]UserInfo, error) {
	zkConn, err := zookeeper.NewConnect()
	if err != nil {
		return nil, err
	}
	defer zkConn.Close()

	// ex: /arcus_acl/group
	groupPath := path.Join(config.AclRootPath, groupName)
	userNames, _, err := zkConn.Children(groupPath)
	if err != nil {
		return nil, err
	}
	users := make([]UserInfo, 0, len(userNames))
	for _, userName := range userNames {
		userPath := path.Join(groupPath, userName)
		roleBytes, _, err := zkConn.Get(userPath)
		if err != nil {
			return nil, err
		}
		users = append(users, UserInfo{Username: userName, Roles: strings.Split(string(roleBytes), ",")})
	}

	return users, nil
}
