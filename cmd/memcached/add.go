package memcached

import (
	"fmt"
	"os"
	"path"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <serviceCode> <ip:port>",
	Short: "Add a memcached server for a service code",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		serviceCode := args[0]
		address := args[1]

		zkConn := cmd.Context().Value(internal.CtxZkConnKey{}).(*zk.Conn)

		if err := addServiceCodePath(zkConn, serviceCode); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := addServerPath(zkConn, serviceCode, address); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Printf("Successfully added server %s to service code %s\n", address, serviceCode)
	},
}

func addServiceCodePath(zkConn *zk.Conn, serviceCode string) error {
	ops := []any{
		&zk.CreateRequest{
			Path:  path.Join(internal.ArcusCacheListPath, serviceCode),
			Data:  nil,
			Acl:   zk.WorldACL(zk.PermAll),
			Flags: 0,
		},
		&zk.CreateRequest{
			Path:  path.Join(internal.ArcusClientListPath, serviceCode),
			Data:  nil,
			Acl:   zk.WorldACL(zk.PermAll),
			Flags: 0,
		},
	}
	if _, err := zkConn.Multi(ops...); err != nil && err != zk.ErrNodeExists {
		return err
	}
	return nil
}

func addServerPath(zkConn *zk.Conn, serviceCode, address string) error {
	ops := []any{
		&zk.CreateRequest{
			Path:  path.Join(internal.ArcusCacheServerMappingPath, address),
			Data:  nil,
			Acl:   zk.WorldACL(zk.PermAll),
			Flags: 0,
		},
		&zk.CreateRequest{
			Path:  path.Join(internal.ArcusCacheServerMappingPath, address, serviceCode),
			Data:  nil,
			Acl:   zk.WorldACL(zk.PermAll),
			Flags: 0,
		},
	}
	if _, err := zkConn.Multi(ops...); err != nil {
		return err
	}

	return nil
}
