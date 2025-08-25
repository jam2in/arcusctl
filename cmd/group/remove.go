package group

import (
	"fmt"
	"path"

	"github.com/jam2in/arcus-sasl-passwd/config"
	"github.com/jam2in/arcus-sasl-passwd/internal/zookeeper"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <groupName>",
	Short: "Remove an empty ACL group.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupName := args[0]
		err := removeGroup(groupName)
		if err != nil {
			return err
		}

		fmt.Printf("ACL group '%s' removed successfully.\n", groupName)
		return nil
	},
}

func removeGroup(groupName string) error {
	zkConn, err := zookeeper.NewConnect()
	if err != nil {
		return err
	}
	defer zkConn.Close()

	groupPath := path.Join(config.AclRootPath, groupName)
	err = zkConn.Delete(groupPath, -1)

	return err
}
