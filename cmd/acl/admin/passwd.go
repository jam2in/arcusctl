package admin

import (
	"errors"
	"fmt"
	"log"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcusctl/internal"
	"github.com/spf13/cobra"
)

var passwdCmd = &cobra.Command{
	Use:   "passwd <group_name>",
	Short: "Change the admin password for a group",
	Long: `Change the administrator password for the specified ACL group.

This command updates the admin password and recursively updates all
ACLs for the group and its users in ZooKeeper.

For password requirements, see: https://github.com/jam2in/arcusctl/blob/main/docs/command-acl.md`,
	Example: `  # Change admin password for group 'cache01'
  arcusctl acl admin passwd cache01`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupName := args[0]

		adminName, err := internal.ReadInput("admin name")
		if err != nil {
			return fmt.Errorf("read admin name: %w", err)
		}
		adminPassword, err := internal.ReadPassword("admin password")
		if err != nil {
			return err
		}
		newPassword, err := internal.ReadPassword("new password")
		if err != nil {
			return err
		}
		repeatedPassword, err := internal.ReadPassword("repeat new password")
		if err != nil {
			return err
		}
		if newPassword != repeatedPassword {
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

		acls := append(zk.DigestACL(zk.PermAll, adminName, newPassword),
			zk.WorldACL(zk.PermRead)...)

		if err := setAclRecursive(conn, internal.ZPATH_ACL_ROOT+"/"+groupName, acls); err != nil {
			return err
		}

		fmt.Println("OK")
		return nil
	},
}

func setAclRecursive(conn *zk.Conn, path string, acls []zk.ACL) error {
	if _, err := conn.SetACL(path, acls, -1); err != nil {
		return fmt.Errorf("SetACL(%s): %w", path, err)
	} else if internal.Flags.Verbose {
		log.Printf("SetACL(%s): OK", path)
	}

	childs, _, err := conn.Children(path)
	if err != nil {
		return fmt.Errorf("Child(%s): %w", path, err)
	}

	var errs []error
	for _, child := range childs {
		if err := setAclRecursive(conn, path+"/"+child, acls); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
