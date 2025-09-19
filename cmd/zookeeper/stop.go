package zookeeper

import (
	"fmt"
	"os"
	"strings"

	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the zookeeper ensemble servers",
	Run: func(cmd *cobra.Command, args []string) {
		zkAddr, zkPath := os.Getenv("ZK_ADDR"), os.Getenv("ZK_PATH")
		if zkAddr == "" || zkPath == "" {
			fmt.Fprintln(os.Stderr, "Environment variables are not provided. \nPlease set the ZK_ADDR, ZK_PATH environment variables")
			os.Exit(1)
		}
		zkServers := strings.Split(zkAddr, ",")
		for _, server := range zkServers {
			ip := strings.Split(server, ":")[0]
			command := fmt.Sprintf(zookeeperStopCommandTemplate, zkPath)
			session, close, err := internal.NewSSHSession(ip)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer close()

			if err := session.Run(command); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		fmt.Printf("zookeeper server stopped successfully!\n")
	},
}
