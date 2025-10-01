package memcached

import (
	"fmt"
	"os"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal/memcached"
	"github.com/jam2in/arcus-cli/internal/types"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <serviceCode> [<ip:port>]",
	Short: "Remove a service code or specific cache server",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		serviceCode := args[0]
		zkConn := cmd.Context().Value(types.CtxZkConnKey{}).(*zk.Conn)
		switch len(args) {
		case 1:
			if err := memcached.RemoveServiceCode(zkConn, serviceCode); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		case 2:
			address := args[1]
			if err := memcached.RemoveFromServiceCode(zkConn, serviceCode, address); err != nil {
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
