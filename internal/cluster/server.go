package cluster

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/jam2in/arcusctl/internal/topology"
)

func memcachedInstallPath(basePath string, version string) string {
	return path.Join(basePath, version)
}

func pidFilePath(
	serverAddress string,
	topoPath string,
	version string,
) string {
	installPath := memcachedInstallPath(topoPath, version)
	return path.Join(installPath, fmt.Sprintf("memcached-%s.pid", listenPort(serverAddress)))
}

func processRunningCommand(pidFile string) string {
	pattern := fmt.Sprintf("[m]emcached.*-P %s", regexp.QuoteMeta(pidFile))
	return fmt.Sprintf("pgrep -f %q > /dev/null 2>&1", pattern)
}

func listenPort(address string) string {
	parts := strings.SplitN(address, ":", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

func pickCacheServer(servers []topology.CacheServer, address string) (*topology.CacheServer, error) {
	for i := range servers {
		if servers[i].Address == address {
			return &servers[i], nil
		}
	}
	return nil, fmt.Errorf("cache server %q not found in cluster", address)
}

func serversInGroup(servers []topology.CacheServer, groupName string) []topology.CacheServer {
	var result []topology.CacheServer
	for _, s := range servers {
		if s.Group != nil && s.Group.Name == groupName {
			result = append(result, s)
		}
	}
	return result
}

func masterFirst(servers []topology.CacheServer) []topology.CacheServer {
	return orderByRole(servers, true)
}

func slaveFirst(servers []topology.CacheServer) []topology.CacheServer {
	return orderByRole(servers, false)
}

func orderByRole(servers []topology.CacheServer, masterFirst bool) []topology.CacheServer {
	var master, slave []topology.CacheServer
	for _, server := range servers {
		if server.IsMaster() {
			master = append(master, server)
		} else {
			slave = append(slave, server)
		}
	}

	if masterFirst {
		return append(master, slave...)
	}
	return append(slave, master...)
}

func selectTargets(
	topo *topology.ClusterTopology,
	edition topology.ClusterEdition,
	nodeAddress string,
	groupName string,
	order func([]topology.CacheServer) []topology.CacheServer,
) ([]topology.CacheServer, error) {
	if edition == topology.CommunityEdition {
		if groupName != "" {
			return nil, fmt.Errorf("--group is not allowed for community cluster")
		}
		if nodeAddress == "" {
			return topo.Servers, nil
		}

		server, err := pickCacheServer(topo.Servers, nodeAddress)
		if err != nil {
			return nil, err
		}
		return []topology.CacheServer{*server}, nil
	}

	// Enterprise edition
	if nodeAddress != "" {
		return nil, fmt.Errorf("--node is not allowed for enterprise cluster")
	}

	servers := topo.Servers
	if groupName != "" {
		servers = serversInGroup(topo.Servers, groupName)
		if len(servers) == 0 {
			return nil, fmt.Errorf("group %q not found in cluster", groupName)
		}
	}
	return order(servers), nil
}

func distinctHosts(
	servers []topology.CacheServer,
) []string {
	seen := map[string]bool{}
	var hosts []string
	for _, server := range servers {
		host := server.Host()
		if _, ok := seen[host]; !ok {
			seen[host] = true
			hosts = append(hosts, host)
		}
	}

	return hosts
}
