package zk

import (
	"errors"
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/jam2in/arcusctl/internal"
	"github.com/jam2in/arcusctl/internal/ssh"
	"github.com/jam2in/arcusctl/internal/store"
	"github.com/jam2in/arcusctl/internal/topology"
)

const (
	removeCommandTemplate = "rm -rf %s"
	exitNCFailed          = 1
)

func Delete(ensembleName string, purge bool) error {
	meta, topo, err := loadEnsemble(ensembleName)
	if err != nil {
		return err
	}

	if err := verifyAllStopped(topo); err != nil {
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

func verifyAllStopped(topo *topology.ZKTopology) error {
	for _, server := range topo.Servers {
		output, err := runRuok(server)

		// If the command succeeded, it means the ZooKeeper is still running and accepting connections.
		if err == nil {
			return fmt.Errorf(
				"ZooKeeper myid=%d on %s is accepting connections;\n"+
					"stop the ensemble first: arcusctl zk stop %s",
				server.MyID, server.Host(), topo.Name,
			)
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) &&
			exitErr.ExitCode() == exitNCFailed &&
			strings.Contains(output, "Connection refused") {
			continue
		}

		detail := strings.TrimSpace(output)
		if detail != "" {
			return fmt.Errorf(
				"cannot check ZooKeeper myid=%d on %s: %s;\n"+
					"check SSH access, nc availability, and network connectivity: %w",
				server.MyID, server.Host(), detail, err,
			)
		}

		return fmt.Errorf(
			"cannot check ZooKeeper myid=%d on %s;\n"+
				"check SSH access and network connectivity: %w",
			server.MyID, server.Host(), err,
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
	var removePaths []string

	// Remove configuration and data directories for each server on the host.
	for _, server := range servers {
		confDir := zkConfigDir(topoPath, ensembleName, server.MyID)
		removePaths = append(removePaths, ssh.Quote(confDir))

		for _, dataPath := range nodeDataPaths(server) {
			removePaths = append(removePaths, ssh.Quote(dataPath))
		}
	}

	cmd := fmt.Sprintf(
		removeCommandTemplate,
		strings.Join(removePaths, " "),
	)

	if err := ssh.Run(host, cmd); err != nil {
		return fmt.Errorf("remove files on %s: %w", host, err)
	}

	return nil
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
