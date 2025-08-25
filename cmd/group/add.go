package group

import (
	"fmt"
	"path"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-sasl-passwd/config"
	"github.com/jam2in/arcus-sasl-passwd/internal/zookeeper"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <groupName>",
	Short: "Add a new ACL group.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupName := args[0]
		err := addGroup(groupName)
		if err != nil {
			return err
		}

		fmt.Printf("ACL group '%s' created successfully.\n", groupName)
		return nil
	},
}

func addGroup(groupName string) error {
	zkConn, err := zookeeper.NewConnect()
	if err != nil {
		return err
	}
	defer zkConn.Close()

	groupPath := path.Join(config.AclRootPath, groupName)
	_, err = zkConn.Create(groupPath, nil, 0, zk.WorldACL(zk.PermAll))

	return err
}
