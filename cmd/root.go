package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/jam2in/arcusctl/cmd/acl"
	"github.com/jam2in/arcusctl/cmd/cluster"
	"github.com/jam2in/arcusctl/cmd/zk"
	"github.com/jam2in/arcusctl/internal"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of arcusctl",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if info, ok := debug.ReadBuildInfo(); ok {
			fmt.Println(info.Main.Version)
		} else {
			panic("failed to debug.ReadBuildInfo()")
		}
	},
}

var rootCmd = &cobra.Command{
	Use:  "arcusctl",
	Long: `arcusctl is a CLI tool for managing and operating ARCUS cache clusters.`,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&internal.Flags.Verbose, "verbose", "v", false, "")
	rootCmd.PersistentFlags().StringVar(&internal.Flags.ConfigFile, "config-file", "", "")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(acl.RootCmd)
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(zk.ZKCmd)
	rootCmd.AddCommand(cluster.ClusterCmd)

	cobra.OnInitialize(internal.InitConfig)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
