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
	Use:  "passwd <group_name>",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]

		adminName := internal.ReadStdin("admin name", false)
		adminPassword := internal.ReadStdin("admin password", true)
		newPassword := internal.ReadStdin("new password", true)
		if newPassword != internal.ReadStdin("repeat new password", true) {
			panic("password does not match")
		}

		conn, err := internal.ConnectZooKeeper(internal.Config.ZooKeeper)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		if err := conn.AddAuth("digest", []byte(adminName+":"+adminPassword)); err != nil {
			panic(err)
		}

		acls := append(zk.DigestACL(zk.PermAll, adminName, newPassword),
			zk.WorldACL(zk.PermRead)...)

		if err := setAclRecursive(conn, internal.ZPATH_ACL_ROOT+"/"+groupName, acls); err != nil {
			panic(err)
		}

		fmt.Println("OK")
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
