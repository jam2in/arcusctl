package cmd

import (
	"fmt"
	"os"

	"github.com/jam2in/arcus-cli/cmd/acl"
	"github.com/jam2in/arcus-cli/cmd/memcached"
	"github.com/jam2in/arcus-cli/cmd/zookeeper"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "arcus-cli",
	Short: "Arcus CLI",
	Long:  "Arcus CLI is a command line interface for managing Arcus",
}

func init() {
	rootCmd.AddCommand(acl.AclCmd)
	rootCmd.AddCommand(zookeeper.ZookeeperCmd)
	rootCmd.AddCommand(memcached.MemcachedCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
