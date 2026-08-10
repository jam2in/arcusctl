package zk

import (
	"fmt"
	"strings"

	"github.com/jam2in/arcusctl/internal"
	"github.com/jam2in/arcusctl/internal/ssh"
	"github.com/jam2in/arcusctl/internal/store"
	"github.com/jam2in/arcusctl/internal/topology"
)

func Delete(ensembleName string) error {
	_, topo, err := loadEnsemble(ensembleName)
	if err != nil {
		return err
	}

	if err := verifyTopology(topo.Servers, topo.Path); err != nil {
		return err
	}

	if err := verifyAllStopped(topo.Servers, topo.Path); err != nil {
		return err
	}

	fmt.Printf("This will remove ZooKeeper ensemble %q from all servers.\n", ensembleName)
	if !internal.Confirm("Are you sure you want to proceed? (y/N): ") {
		fmt.Println("Aborted.")
		return nil
	}

	hostsMap := groupServersByHost(topo.Servers)
	for host, servers := range hostsMap {
		fmt.Printf("Removing files on %s...\n", host)
		if err := removeHostFiles(host, servers, topo.Path); err != nil {
			return fmt.Errorf("remove files on %s: %w", host, err)
		}
	}

	if err := store.DeleteZK(ensembleName); err != nil {
		return fmt.Errorf("delete metadata: %w", err)
	}

	fmt.Printf("ZooKeeper ensemble %q deleted.\n", ensembleName)
	return nil
}

func verifyTopology(servers []topology.ZKServer, topoPath string) error {
	for _, server := range servers {
		confDir := fmt.Sprintf("%s/conf_myid_%d", topoPath, server.MyID)
		if err := ssh.Run(server.Host(), fmt.Sprintf("test -d %s", confDir)); err != nil {
			return fmt.Errorf("topology mismatch: %s not found on %s", confDir, server.Host())
		}
	}
	return nil
}

func verifyAllStopped(servers []topology.ZKServer, topoPath string) error {
	for _, server := range servers {
		confPath := zkConfigPath(topoPath, server.MyID)
		cmd := fmt.Sprintf("pgrep -f '[Q]uorumPeerMain.*%s' > /dev/null 2>&1", confPath)
		if err := ssh.Run(server.Host(), cmd); err == nil {
			return fmt.Errorf("server %s (myid=%d) is still running. stop the ensemble before delete",
				server.Host(), server.MyID)
		}
	}
	return nil
}

func groupServersByHost(servers []topology.ZKServer) map[string][]topology.ZKServer {
	hosts := map[string][]topology.ZKServer{}
	for _, server := range servers {
		hosts[server.Host()] = append(hosts[server.Host()], server)
	}
	return hosts
}

func removeHostFiles(host string, servers []topology.ZKServer, topoPath string) error {
	paths := []string{topoPath}
	for _, server := range servers {
		paths = append(paths, fmt.Sprintf("%s/zk%d", server.Config.DataDir, server.MyID))
	}

	cmd := "rm -rf " + strings.Join(paths, " ")
	return ssh.Run(host, cmd)
}
