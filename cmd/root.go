package cmd

import (
	"fmt"
	"os"

	"github.com/jam2in/arcusctl/cmd/acl"
	"github.com/jam2in/arcusctl/cmd/memcached"
	"github.com/jam2in/arcusctl/cmd/zookeeper"
	"github.com/jam2in/arcusctl/internal"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "arcusctl",
	Short: "Arcus CLI",
	Long:  "Arcus CLI is a command line interface for managing Arcus",
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&internal.Flags.Verbose, "verbose", "v", false, "")
	rootCmd.PersistentFlags().StringVar(&internal.Flags.ConfigFile, "config-file", "", "")

	rootCmd.AddCommand(acl.RootCmd)
	rootCmd.AddCommand(zookeeper.ZookeeperCmd)
	rootCmd.AddCommand(memcached.MemcachedCmd)

	cobra.OnInitialize(internal.InitConfig)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
