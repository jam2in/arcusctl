package zk

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/jam2in/arcusctl/internal"
	"github.com/jam2in/arcusctl/internal/ssh"
	"github.com/jam2in/arcusctl/internal/store"
	"github.com/jam2in/arcusctl/internal/topology"
)

const (
	removeCommandTemplate = "rm -rf %s"
	exitStillRunning      = 9
	exitSSHConnFailed     = 255
)

func Delete(ensembleName string, purge bool) error {
	meta, topo, err := loadEnsemble(ensembleName)
	if err != nil {
		return err
	}

	if err := verifyNotResponding(topo); err != nil {
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
			return err
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

func verifyNotResponding(topo *topology.ZKTopology) error {
	targets := topologyTargets(topo.Servers)
	results := probeAll(targets)

	var responding []string
	for _, server := range topo.Servers {
		result, ok := results[server.MyID]
		if !ok {
			continue
		}

		if errors.Is(result.Err, errMntrNotWhitelisted) {
			return fmt.Errorf(
				"ZooKeeper at %s (myid=%d) responded but mntr is not whitelisted. "+
					"stop the ensemble first: arcusctl zk stop %s",
				targets[server.MyID], server.MyID, topo.Name,
			)
		}

		if result.Err == nil {
			responding = append(
				responding,
				fmt.Sprintf("myid=%d (%s)", server.MyID, targets[server.MyID]),
			)
		}
	}

	if len(responding) > 0 {
		return fmt.Errorf(
			"ZooKeeper is still responding %s.\nstop the ensemble first: arcusctl zk stop %s",
			strings.Join(responding, ", "), topo.Name,
		)
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
	var confDirs, removePaths []string

	for _, server := range servers {
		confDir := zkConfigDir(topoPath, ensembleName, server.MyID)
		confDirs = append(confDirs, ssh.Quote(confDir))

		removePaths = append(removePaths, ssh.Quote(confDir))
		for _, dataPath := range nodeDataPaths(server) {
			removePaths = append(removePaths, ssh.Quote(dataPath))
		}
	}

	cmd := fmt.Sprintf(
		`for d in %s; do pgrep -f "[Q]uorumPeerMain.*$d" > /dev/null 2>&1 && exit %d; done; rm -rf %s`,
		strings.Join(confDirs, " "),
		exitStillRunning,
		strings.Join(removePaths, " "),
	)

	code, err := ssh.RunCode(host, cmd)
	if err != nil {
		return fmt.Errorf("run ssh for %s: %w", host, err)
	}

	switch code {
	case 0:
		return nil
	case exitStillRunning:
		return fmt.Errorf(
			"ZooKeeper is still running on %s. stop the ensemble first: arcusctl zk stop %s",
			host, ensembleName,
		)
	case exitSSHConnFailed:
		return fmt.Errorf(
			"cannot connect to %s over ssh. check the host is reachable and your ssh key is authorized",
			host,
		)
	default:
		return fmt.Errorf("remove files on %s: exit code %d", host, code)
	}
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
		cmd := fmt.Sprintf(removeCommandTemplate, ssh.Quote(installPath))
		if err := ssh.Run(host, cmd); err != nil {
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
