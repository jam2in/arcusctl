package cluster

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jam2in/arcusctl/internal"
	"github.com/jam2in/arcusctl/internal/store"
	"github.com/jam2in/arcusctl/internal/topology"
)

func Deploy(version string, topoPath string) error {
	topo, topologyBytes, edition, err := prepareTopology(topoPath)
	if err != nil {
		return err
	}

	printPlan(topo, version, edition)
	if !internal.Confirm("Proceed with deployment? (y/N): ") {
		fmt.Println("Aborted.")
		return nil
	}

	localTarPath, err := ensureDownloaded(version, edition)
	if err != nil {
		return err
	}

	installed, err := installServers(topo, version, localTarPath, edition)
	if err != nil {
		printRecoveryGuide(topo, installed, version, edition)
		return err
	}

	if err := registerZNodes(topo, edition); err != nil {
		printRecoveryGuide(topo, installed, version, edition)
		return fmt.Errorf("register ZNodes in ZooKeeper: %w", err)
	}

	if err := store.SaveCluster(topo.ServiceCode, version, topologyBytes); err != nil {
		printRecoveryGuide(topo, installed, version, edition)
		return fmt.Errorf("save metadata: %w", err)
	}

	fmt.Printf(
		"Arcus cluster %q deployed. Run 'cluster start %s' to launch.\n",
		topo.ServiceCode, topo.ServiceCode,
	)
	return nil
}

func prepareTopology(topoPath string) (*topology.ClusterTopology, []byte, topology.ClusterEdition, error) {
	topo, rawData, err := topology.LoadCluster(topoPath)
	if err != nil {
		return nil, nil, "", err
	}

	if err := topo.Validate(); err != nil {
		return nil, nil, "", err
	}

	if store.ClusterExists(topo.ServiceCode) {
		return nil, nil, "", fmt.Errorf(
			"arcus cluster %q already exists",
			topo.ServiceCode,
		)
	}

	edition, err := topo.Edition()
	if err != nil {
		return nil, nil, "", err
	}

	exists, err := clusterExistsInZK(topo.ZooKeeper, topo.ServiceCode, edition)
	if err != nil {
		return nil, nil, "", err
	}

	if exists {
		return nil, nil, "", fmt.Errorf(
			"arcus cluster %q already exists in ZooKeeper",
			topo.ServiceCode,
		)
	}

	return topo, rawData, edition, nil
}

func printPlan(
	topo *topology.ClusterTopology,
	version string,
	edition topology.ClusterEdition,
) {
	fmt.Printf(
		"Arcus cluster %q will be deployed (edition: %s, version: %s)\n\n",
		topo.ServiceCode, edition, version,
	)

	installPath := memcachedInstallPath(topo.Path, version)

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if edition == topology.EnterpriseEdition {
		fmt.Fprintln(writer, "GROUP\tROLE\tADDRESS\tDIRECTORY")
		fmt.Fprintln(writer, "\t\t\t")
		for _, server := range topo.Servers {
			fmt.Fprintf(writer, "%s\t%s\t%s\t%s\n",
				server.Group.Name, server.Group.Role, server.Address, installPath)
		}
	} else {
		fmt.Fprintln(writer, "ADDRESS\tDIRECTORY")
		fmt.Fprintln(writer, "\t")
		for _, server := range topo.Servers {
			fmt.Fprintf(writer, "%s\t%s\n",
				server.Address, installPath)
		}
	}

	writer.Flush()

	fmt.Println()
	fmt.Println("Attention:")
	fmt.Println("  1. If the topology is not what you expected, check your yaml file.")
	fmt.Println("  2. Please confirm there is no port/directory conflicts in same host.")
}

func printRecoveryGuide(
	topo *topology.ClusterTopology,
	installed []topology.CacheServer,
	version string,
	edition topology.ClusterEdition,
) {
	fmt.Println("\nDeployment failed. Manual recovery required.")

	if len(installed) > 0 {
		fmt.Println("The following servers have been partially installed.")
		for _, s := range installed {
			fmt.Printf("  - %s\n", s.Host())
		}
		cleanupPath := memcachedInstallPath(topo.Path, version)
		fmt.Println("\nTo clean up, manually run on each server:")
		fmt.Printf("  $ rm -rf %s\n", cleanupPath)
	}

	root := rootForEdition(edition)
	fmt.Println("\nZooKeeper nodes may have been created under:")
	fmt.Printf("  %s/{cache_list, client_list, cache_server_mapping}/%s ...\n", root, topo.ServiceCode)
	fmt.Println("They are left in place. To retry from a clean state, remove them.")
}
