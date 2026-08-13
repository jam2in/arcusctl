package zk

import (
	"fmt"
	"path"
	"strings"

	"github.com/jam2in/arcusctl/internal"
	"github.com/jam2in/arcusctl/internal/ssh"
	"github.com/jam2in/arcusctl/internal/store"
	"github.com/jam2in/arcusctl/internal/topology"
)

const removeCommandTemplate = "rm -rf %s"

func Delete(ensembleName string, purge bool) error {
	meta, topo, err := loadEnsemble(ensembleName)
	if err != nil {
		return err
	}

	if err := verifyTopology(topo.Servers, topo.Path, topo.Name); err != nil {
		return err
	}

	if err := verifyAllStopped(topo.Servers, topo.Path, topo.Name); err != nil {
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
		if err := removeHostFiles(host, servers, topo.Path, topo.Name); err != nil {
			return fmt.Errorf("remove files on %s: %w", host, err)
		}
	}

	// If user specified --purge, remove the installation directories on each server.
	// If the installation directory is shared with another ensemble, it will not be removed.
	if purge {
		if err := removeInstallationDirs(ensembleName, topo, meta.Version); err != nil {
			return err
		}
	}

	if err := store.DeleteZK(ensembleName); err != nil {
		return fmt.Errorf("delete metadata: %w", err)
	}

	fmt.Printf("ZooKeeper ensemble %q deleted.\n", ensembleName)
	return nil
}

func verifyTopology(
	servers []topology.ZKServer,
	topoPath string,
	ensembleName string,
) error {
	for _, server := range servers {
		confDir := zkConfigDir(topoPath, ensembleName, server.MyID)

		if err := ssh.Run(server.Host(), fmt.Sprintf("test -d %s", confDir)); err != nil {
			return fmt.Errorf("topology mismatch: %s not found on %s", confDir, server.Host())
		}
	}
	return nil
}

func verifyAllStopped(
	servers []topology.ZKServer,
	topoPath string,
	ensembleName string,
) error {
	for _, server := range servers {
		confDir := zkConfigDir(topoPath, ensembleName, server.MyID)
		cmd := fmt.Sprintf("pgrep -f '[Q]uorumPeerMain.*%s' > /dev/null 2>&1", confDir)
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

func removeHostFiles(
	host string,
	servers []topology.ZKServer,
	topoPath string,
	ensembleName string,
) error {
	var removePaths []string

	for _, server := range servers {
		// Remove conf/<ensembleName>/zk<myid>.cfg and data/log directories
		removePaths = append(
			removePaths,
			zkConfigDir(topoPath, ensembleName, server.MyID),
		)
		removePaths = append(removePaths, nodeDataPaths(server)...)
	}

	return ssh.Run(host, "rm -rf "+strings.Join(removePaths, " "))
}

func removeInstallationDirs(
	ensembleName string,
	topo *topology.ZKTopology,
	version string,
) error {
	installPath := zkInstallPath(topo.Path, version)

	for host := range groupServersByHost(topo.Servers) {
		other, err := sharingEnsemble(ensembleName, installPath, host)
		if err != nil {
			return err
		}

		if other != "" {
			fmt.Printf(
				"Skip removing directory on %s: install path is shared with ensemble %q\n",
				host, other,
			)
			continue
		}

		fmt.Printf("Removing installation directory on %s...\n", host)
		if err := ssh.Run(host, fmt.Sprintf(removeCommandTemplate, installPath)); err != nil {
			return fmt.Errorf("remove installation on %s: %w", host, err)
		}
	}

	return nil
}

func nodeDataPaths(server topology.ZKServer) []string {
	nodeDirName := fmt.Sprintf("zk%d", server.MyID)

	return []string{
		path.Join(server.Config.DataDir, nodeDirName),
		path.Join(server.Config.DataLogDir, nodeDirName),
	}
}
