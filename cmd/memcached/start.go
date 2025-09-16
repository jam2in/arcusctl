package memcached

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start <serviceCode> [ip:port...]",
	Short: "start all servers or specific servers in service code",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		serviceCode := args[0]
		targetServers := args[1:]
		zkConn := cmd.Context().Value(internal.CtxZkConnKey{}).(*zk.Conn)

		globalConfig, _, err := zkConn.Get(path.Join(internal.ArcusCacheListPath, serviceCode))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		serviceCodeServers, err := getServiceCodeServers(zkConn, serviceCode)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		var serversToStart []string
		if len(targetServers) > 0 {
			serversToStart = filterServers(serviceCodeServers, targetServers)
			if len(serversToStart) == 0 {
				fmt.Fprintln(os.Stderr, "No servers found in service code")
				os.Exit(1)
			}
		} else {
			serversToStart = serviceCodeServers
		}

		for _, serverAddress := range serversToStart {
			ip, port, flag := strings.Cut(serverAddress, ":")
			if !flag {
				fmt.Fprintln(os.Stderr, "Invalid server address:", serverAddress)
				os.Exit(1)
			}

			client, err := internal.NewSSHClient(ip)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer client.Close()

			session, err := client.NewSession()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer session.Close()

			memcachedPath := os.Getenv("ARCUS_PATH")
			command := fmt.Sprintf(memcachedStartCommandTemplate,
				memcachedPath, memcachedPath, memcachedPath, memcachedPath, memcachedPath,
				port, string(globalConfig), os.Getenv("ZK_ADDR"))

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			errChan := make(chan error, 1)
			go func() {
				errChan <- session.Run(command)
			}()
			select {
			case err := <-errChan:
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			case <-ctx.Done():
				continue
			}
			fmt.Printf("  - Start command sent to %s successfully.\n", serverAddress)
		}
	},
}
