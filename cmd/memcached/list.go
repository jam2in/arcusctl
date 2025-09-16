package memcached

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

type serviceStatus struct {
	serviceCode string
	total       int
	online      int
	offline     int
}

var listCmd = &cobra.Command{
	Use:   "list [serviceCode]",
	Short: "list all servers in arcus cache cloud",
	Run: func(cmd *cobra.Command, args []string) {

		zkConn := cmd.Context().Value(internal.CtxZkConnKey{}).(*zk.Conn)

		serviceCodeMap, err := buildServiceCodeMap(zkConn)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		var serviceCodes []string
		if len(args) == 0 {
			for sm := range serviceCodeMap {
				serviceCodes = append(serviceCodes, sm)
			}
			sort.Strings(serviceCodes)
		} else {
			serviceCodes = []string{args[0]}
		}

		liveServerMaps := make(map[string]map[string]struct{})
		var statuses []serviceStatus
		for _, sc := range serviceCodes {
			liveServers, _ := getLiveServers(zkConn, sc)
			liveServerMaps[sc] = liveServers

			onlineCnt := 0
			for _, s := range serviceCodeMap[sc] {
				if _, ok := liveServers[s]; ok {
					onlineCnt++
				}
			}

			statuses = append(statuses, serviceStatus{
				serviceCode: sc,
				total:       len(serviceCodeMap[sc]),
				online:      onlineCnt,
				offline:     len(serviceCodeMap[sc]) - onlineCnt,
			})
		}

		if len(args) == 1 {
			serviceCode := args[0]
			servers := serviceCodeMap[serviceCode]
			liveServers := liveServerMaps[serviceCode]

			fmt.Printf("Servers in service code '%s':\n", serviceCode)
			for _, server := range servers {
				status := "offline"
				if _, isLive := liveServers[server]; isLive {
					status = "online"
				}
				fmt.Printf("  - %-21s %s\n", server, status)
			}
			fmt.Println()
		}

		fmt.Printf("%-25s %-8s %-8s %-8s\n", "SERVICE CODE", "TOTAL", "ONLINE", "OFFLINE")
		fmt.Println(strings.Repeat("-", 60))
		for _, s := range statuses {
			fmt.Printf("%-25s %-8d %-8d %-8d\n", s.serviceCode, s.total, s.online, s.offline)
		}

	},
}

func buildServiceCodeMap(zkConn *zk.Conn) (map[string][]string, error) {
	serviceCodeMap := make(map[string][]string)
	allServers, _, err := zkConn.Children(path.Join(internal.ArcusCacheServerMappingPath))
	if err != nil {
		return nil, err
	}

	for _, s := range allServers {
		serviceCodeTags, _, err := zkConn.Children(path.Join(internal.ArcusCacheServerMappingPath, s))
		if err != nil {
			continue
		}
		for _, sc := range serviceCodeTags {
			serviceCodeMap[sc] = append(serviceCodeMap[sc], s)
		}
	}

	return serviceCodeMap, nil
}
