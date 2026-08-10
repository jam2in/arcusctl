package cluster

import (
	"fmt"

	"github.com/jam2in/arcusctl/internal"
	"github.com/jam2in/arcusctl/internal/ssh"
	"github.com/jam2in/arcusctl/internal/store"
	"github.com/jam2in/arcusctl/internal/topology"
)

const removeCommandTemplate = "rm -rf %s"

func Delete(serviceCode string, purge bool) error {
	meta, topo, edition, err := loadCluster(serviceCode)
	if err != nil {
		return err
	}

	if err := verifyAllStopped(topo, meta.Version); err != nil {
		return err
	}

	fmt.Printf("This will remove Arcus cluster %q from all servers.\n", serviceCode)
	if !internal.Confirm("Are you sure you want to proceed? (y/N): ") {
		fmt.Println("Aborted.")
		return nil
	}

	if err := unregisterZNodes(topo, edition); err != nil {
		return err
	}

	// If user specified --purge, remove the installation directories on each server.
	// If the installation directory is shared with another cluster, it will not be removed.
	if purge {
		if err := removeInstallationDirs(serviceCode, topo, meta.Version); err != nil {
			return err
		}
	}

	if err := store.DeleteCluster(serviceCode); err != nil {
		return fmt.Errorf("delete cluster metadata: %w", err)
	}

	fmt.Printf("Arcus cluster %q deleted.\n", serviceCode)
	return nil
}

func removeInstallationDirs(
	serviceCode string,
	topo *topology.ClusterTopology,
	version string,
) error {
	installPath := memcachedInstallPath(topo.Path, version)

	for _, host := range distinctHosts(topo.Servers) {
		other, err := sharingCluster(serviceCode, topo.Path, version, host)
		if err != nil {
			return err
		}

		if other != "" {
			fmt.Printf(
				"Skip removing directory on %s: install path is shared with cluster %q\n",
				host, other,
			)
			continue
		}

		fmt.Printf("Removing files on %s...\n", host)
		if err := ssh.Run(host, fmt.Sprintf(removeCommandTemplate, installPath)); err != nil {
			return fmt.Errorf("remove files on %s: %w", host, err)
		}
	}

	return nil
}

func verifyAllStopped(
	topo *topology.ClusterTopology,
	version string,
) error {
	for _, server := range topo.Servers {
		pidFile := pidFilePath(server.Address, topo.Path, version)
		cmd := processRunningCommand(pidFile)
		if err := ssh.Run(server.Host(), cmd); err == nil {
			return fmt.Errorf(
				"cache server %q is still running; stop the cluster before delete",
				server.Address,
			)
		}
	}

	return nil
}
