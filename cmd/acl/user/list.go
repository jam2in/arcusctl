package user

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list <groupName>",
	Short: "List all users in an ACL group.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]

		zkConn := cmd.Context().Value(internal.CtxZkConnKey{}).(*zk.Conn)

		users, err := listUsers(zkConn, groupName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Printf("Users in group '%s':\n", groupName)
		for i, u := range users {
			fmt.Printf("  %d. %s\n", i+1, u)
		}
		fmt.Printf("Total %d users in group '%s'.\n", len(users), groupName)
	},
}

func listUsers(zkConn *zk.Conn, groupName string) ([]UserInfo, error) {
	groupPath := internal.AclRootPath + "/" + groupName

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
