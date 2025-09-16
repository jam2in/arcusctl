package memcached

import (
	"path"
	"strings"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

const (
	memcachedStartCommandTemplate = "%s/bin/memcached -E %s/lib/default_engine.so -X %s/lib/syslog_logger.so -X %s/lib/ascii_scrub.so -P %s/memcached-%s.pid -d -v -r -R5 -U 0 -D: -b 8192 %s -z %s"
)

var MemcachedCmd = &cobra.Command{
	Use:   "memcached",
	Short: "Memcached command",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := internal.ContextWithZkConn(cmd.Context(), "", "")
		if err != nil {
			return err
		}
		zkConn := ctx.Value(internal.CtxZkConnKey{}).(*zk.Conn)
		if err = internal.InitializeZK(zkConn); err != nil {
			return err
		}
		cmd.SetContext(ctx)

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		zkConn := cmd.Context().Value(internal.CtxZkConnKey{}).(*zk.Conn)
		zkConn.Close()
	},
}

func init() {
	MemcachedCmd.AddCommand(addCmd)
	MemcachedCmd.AddCommand(removeCmd)
	MemcachedCmd.AddCommand(configCmd)
	MemcachedCmd.AddCommand(connectCmd)
	MemcachedCmd.AddCommand(listCmd)
	MemcachedCmd.AddCommand(startCmd)
	MemcachedCmd.AddCommand(stopCmd)
}

func getServiceCodeServers(zkConn *zk.Conn, serviceCode string) ([]string, error) {
	servers, _, err := zkConn.Children(internal.ArcusCacheServerMappingPath)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, server := range servers {
		children, _, err := zkConn.Children(path.Join(internal.ArcusCacheServerMappingPath, server))
		if err != nil {
			continue
		}
		for _, child := range children {
			if child == serviceCode {
				result = append(result, server)
				break
			}
		}
	}
	return result, nil
}

func getLiveServers(zkConn *zk.Conn, serviceCode string) (map[string]struct{}, error) {
	liveNodes, _, err := zkConn.Children(path.Join(internal.ArcusCacheListPath, serviceCode))
	if err != nil {
		return nil, err
	}

	liveServers := make(map[string]struct{})
	for _, liveNode := range liveNodes {
		addr, _, _ := strings.Cut(liveNode, "-")
		liveServers[addr] = struct{}{}
	}
	return liveServers, nil
}

func filterServers(allServers []string, targets []string) []string {
	targetSet := make(map[string]struct{})
	for _, t := range targets {
		targetSet[t] = struct{}{}
	}

	result := make([]string, 0)
	for _, s := range allServers {
		if _, ok := targetSet[s]; ok {
			result = append(result, s)
		}
	}
	return result
}
