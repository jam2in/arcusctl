package memcached

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <serviceCode> [<ip:port>]",
	Short: "Remove a service code or specific cache server",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		serviceCode := args[0]
		zkConn := cmd.Context().Value(internal.CtxZkConnKey{}).(*zk.Conn)
		switch len(args) {
		case 1:
			if err := removeServiceCode(zkConn, serviceCode); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		case 2:
			address := args[1]
			if err := removeServer(zkConn, serviceCode, address); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		default:
			fmt.Fprintln(os.Stderr, "Invalid argument")
			os.Exit(1)
		}

		fmt.Printf("\nSuccessfully remove to service code %s\n", serviceCode)
	},
}

func removeServiceCode(zkConn *zk.Conn, serviceCode string) error {
	serverAddress, _, err := zkConn.Children(internal.ArcusCacheServerMappingPath)
	if err != nil {
		return err
	}

	for _, address := range serverAddress {
		mappingPath := path.Join(internal.ArcusCacheServerMappingPath, address, serviceCode)
		exists, _, err := zkConn.Exists(mappingPath)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%s exists", mappingPath)
		}
	}

	ops := []any{
		&zk.DeleteRequest{
			Path:    path.Join(internal.ArcusCacheListPath, serviceCode),
			Version: -1,
		},
		&zk.DeleteRequest{
			Path:    path.Join(internal.ArcusClientListPath, serviceCode),
			Version: -1,
		},
	}
	if _, err = zkConn.Multi(ops...); err != nil {
		return err
	}

	return nil
}

func removeServer(zkConn *zk.Conn, serviceCode, address string) error {
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid address: %s", address)
	}

	ops := []any{
		&zk.DeleteRequest{
			Path:    path.Join(internal.ArcusCacheServerMappingPath, address, serviceCode),
			Version: -1,
		},
		&zk.DeleteRequest{
			Path:    path.Join(internal.ArcusCacheServerMappingPath, address),
			Version: -1,
		},
	}
	if _, err := zkConn.Multi(ops...); err != nil {
		return err
	}

	return nil
}
