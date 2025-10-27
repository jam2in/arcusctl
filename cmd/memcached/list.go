package memcached

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcusctl/internal/memcached"
	"github.com/jam2in/arcusctl/internal/types"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [serviceCode]",
	Short: "list all servers in arcus cache cloud",
	Run: func(cmd *cobra.Command, args []string) {
		zkConn := cmd.Context().Value(types.CtxZkConnKey{}).(*zk.Conn)

		if len(args) > 0 {
			serviceCode := args[0]
			status, err := memcached.GetServiceCodeStatus(zkConn, serviceCode)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Println(status.String())
			return
		}

		statuses, err := memcached.GetAllServiceCodeStatus(zkConn)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Printf("%-25s %-8s %-8s %-8s\n", "SERVICE CODE", "TOTAL", "ONLINE", "OFFLINE")
		fmt.Println(strings.Repeat("-", 60))
		for _, s := range statuses {
			fmt.Printf("%-25s %-8d %-8d %-8d\n", s.ServiceCode, s.Total, len(s.OnlineServers), len(s.OfflineServers))
		}

	},
}
