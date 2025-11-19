package user

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcusctl/internal"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:  "remove <group_name> <user_name>",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		userName := args[1]

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

		if err := CheckClientList(conn, groupName, userName); err != nil {
			panic(err)
		}

		if _, err := conn.Multi(
			&zk.DeleteRequest{
				Path:    internal.ZPATH_ACL_ROOT + "/" + groupName + "/" + userName + "/" + propName,
				Version: -1,
			},
			&zk.DeleteRequest{
				Path:    internal.ZPATH_ACL_ROOT + "/" + groupName + "/" + userName,
				Version: -1,
			},
		); err != nil {
			panic(err)
		}

		fmt.Println("OK")
	},
}

func CheckClientList(conn *zk.Conn, groupName string, username string) error {
	zpath := internal.ZPATH_ACL_MAPPING_ROOT
	serviceCodes, _, err := conn.Children(zpath)
	if err != nil {
		return fmt.Errorf("Children(\"%s\"): %w", zpath, err)
	}

	for _, serviceCode := range serviceCodes {
		zpath := internal.ZPATH_ACL_MAPPING_ROOT + "/" + serviceCode
		group, _, err := conn.Get(zpath)
		if err != nil {
			return fmt.Errorf("Get(\"%s\"): %w", zpath, err)
		}

		if groupName != string(group) {
			continue
		}

		zpath = internal.ZPATH_ROOT + internal.ZPATH_CLIENT_LIST + "/" + serviceCode
		clientList, _, err := conn.Children(zpath)
		if err != nil && !errors.Is(err, zk.ErrNoNode) {
			return fmt.Errorf("Children(\"%s\"): %w", zpath, err)
		}
		for _, client := range clientList {
			if strings.Contains(client, fmt.Sprintf("_sasluser=%s", username)) {
				return fmt.Errorf("found %s in %s/%s", username, zpath, client)
			}
		}

		zpath = internal.ZPATH_REPL_ROOT + internal.ZPATH_CLIENT_LIST + "/" + serviceCode
		clientList, _, err = conn.Children(zpath)
		if err != nil && !errors.Is(err, zk.ErrNoNode) {
			return fmt.Errorf("Children(\"%s\"): %w", zpath, err)
		}
		for _, client := range clientList {
			if strings.Contains(client, fmt.Sprintf("_sasluser=%s", username)) {
				return fmt.Errorf("found %s in %s/%s", username, zpath, client)
			}
		}
	}

	return nil
}
