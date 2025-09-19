package memcached

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <serviceCode> [ip:port...]",
	Short: "stop all servers or specific servers in service code",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceCode := args[0]
		targetServers := args[1:]
		zkConn := cmd.Context().Value(internal.CtxZkConnKey{}).(*zk.Conn)

		serviceCodeServers, err := getServiceCodeServers(zkConn, serviceCode)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		var serversToStop []string
		if len(targetServers) > 0 {
			serversToStop = filterServers(serviceCodeServers, targetServers)
			if len(serversToStop) == 0 {
				fmt.Fprintln(os.Stderr, "No servers found in service code")
				os.Exit(1)
			}
		} else {
			serversToStop = serviceCodeServers
		}

		for _, serverAddress := range serversToStop {
			ip, port, flag := strings.Cut(serverAddress, ":")
			if !flag {
				fmt.Fprintln(os.Stderr, "Invalid server address:", serverAddress)
				os.Exit(1)
			}

			session, close, err := internal.NewSSHSession(ip)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer close()

			pidFilePath := fmt.Sprintf("%s/memcached-%s.pid", os.Getenv("ARCUS_PATH"), port)
			command := fmt.Sprintf("kill -INT $(cat %s)", pidFilePath)
			if err := session.Run(command); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Printf("  - Stop command sent to %s successfully.\n", serverAddress)
		}
	},
}
