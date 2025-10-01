package memcached

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal/memcached"
	"github.com/jam2in/arcus-cli/internal/types"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config <serviceCode> [options...]",
	Short: "Create or update the global configuration for a service code",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceCode := args[0]
		options := strings.Join(args[1:], " ")
		zkConn := cmd.Context().Value(types.CtxZkConnKey{}).(*zk.Conn)

		if err := memcached.SetClusterConfig(zkConn, serviceCode, options); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("Global config for service '%s' has been updated\n", serviceCode)
	},
}

func init() {
	configCmd.Flags().SetInterspersed(false)
}
