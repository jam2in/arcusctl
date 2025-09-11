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

var configCmd = &cobra.Command{
	Use:   "config <serviceCode> [options...]",
	Short: "Create or update the global configuration for a service code",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceCode := args[0]
		newData := []byte(strings.Join(args[1:], " "))
		cacheListPath := path.Join(internal.ArcusCacheListPath, serviceCode)

		zkConn := cmd.Context().Value(internal.CtxZkConnKey{}).(*zk.Conn)
		exists, _, err := zkConn.Exists(cacheListPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if exists {
			if _, err := zkConn.Set(cacheListPath, newData, -1); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		fmt.Printf("Global config for service '%s' has been updated\n", serviceCode)
	},
}

func init() {
	configCmd.Flags().SetInterspersed(false)
}
