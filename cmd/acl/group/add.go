package group

import (
	"errors"
	"fmt"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcusctl/internal"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <group_name>",
	Short: "Create a new ACL group",
	Long: `Create a new ACL group with the specified name.

You will be prompted to enter an admin name and password for managing
this group. The admin credentials are used for controlling access to
the group and its users.

For password requirements, see: https://github.com/jam2in/arcusctl/blob/main/docs/command-acl.md`,
	Example: `  # Create a new group named 'cache01'
  arcusctl acl group add cache01`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupName := args[0]

		adminName, err := internal.ReadInput("admin name")
		if err != nil {
			return fmt.Errorf("read admin name: %w", err)
		}
		adminPassword, err := internal.ReadPassword("password")
		if err != nil {
			return err
		}
		repeatedPassword, err := internal.ReadPassword("repeat password")
		if err != nil {
			return err
		}
		if adminPassword != repeatedPassword {
			return errors.New("password does not match")
		}

		conn, err := internal.ConnectZooKeeper(internal.Config.ZooKeeper)
		if err != nil {
			return err
		}
		defer conn.Close()

		if err := conn.AddAuth("digest", []byte(adminName+":"+adminPassword)); err != nil {
			return fmt.Errorf("authenticate ZooKeeper admin %q: %w", adminName, err)
		}

		if err := internal.EnsureZNode(conn, internal.ZPATH_ACL_ROOT); err != nil {
			return err
		}
		if _, err := conn.Create(internal.ZPATH_ACL_ROOT+"/"+groupName, nil, 0,
			append(zk.DigestACL(zk.PermAll, adminName, adminPassword),
				zk.WorldACL(zk.PermRead)...)); err != nil {
			return fmt.Errorf("create ACL group %q: %w", groupName, err)
		}

		fmt.Println("OK")
		return nil
	},
}
