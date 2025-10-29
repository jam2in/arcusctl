package cmd

import (
	"fmt"
	"os"

	"github.com/jam2in/arcusctl/cmd/acl"
	"github.com/jam2in/arcusctl/internal"
	"github.com/spf13/cobra"
)

var version = "develop"

var versionCmd = &cobra.Command{
	Use:  "version",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}

var rootCmd = &cobra.Command{
	Use: "arcusctl",
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&internal.Flags.Verbose, "verbose", "v", false, "")
	rootCmd.PersistentFlags().StringVar(&internal.Flags.ConfigFile, "config-file", "", "")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(acl.RootCmd)
	rootCmd.AddCommand(connectCmd)

	// rootCmd.AddCommand(zookeeper.ZookeeperCmd)
	// rootCmd.AddCommand(memcached.MemcachedCmd)

	cobra.OnInitialize(internal.InitConfig)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
