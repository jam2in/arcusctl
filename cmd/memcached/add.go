package memcached

import (
	"fmt"
	"os"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal/memcached"
	"github.com/jam2in/arcus-cli/internal/types"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <serviceCode> <ip:port>",
	Short: "Add a memcached server for a service code",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		serviceCode := args[0]
		address := args[1]
		zkConn := cmd.Context().Value(types.CtxZkConnKey{}).(*zk.Conn)

		if err := memcached.AddToServiceCode(zkConn, serviceCode, address); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("Successfully added server %s to service code %s\n", address, serviceCode)
	},
}
