package zk

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jam2in/arcusctl/internal"
	"github.com/jam2in/arcusctl/internal/store"
	"github.com/jam2in/arcusctl/internal/topology"
)

func Deploy(version string, topoPath string) error {
	topo, topologyBytes, err := prepareTopology(topoPath)
	if err != nil {
		return err
	}

	printPlan(topo, version)
	if !internal.Confirm("Proceed with deployment? (y/N): ") {
		fmt.Println("Aborted.")
		return nil
	}

	localTarPath, err := ensureDownloaded(version)
	if err != nil {
		return err
	}

	installed, err := installServers(topo, version, localTarPath)
	if err != nil {
		printRecoveryGuide(installed, topo)
		return err
	}

	if err := store.SaveZK(topo.Name, version, topologyBytes); err != nil {
		printRecoveryGuide(installed, topo)
		return fmt.Errorf("save metadata: %w", err)
	}

	fmt.Printf("ZooKeeper ensemble %q deployed successfully.\n", topo.Name)
	return nil
}

func prepareTopology(topoPath string) (*topology.ZKTopology, []byte, error) {
	topo, rawData, err := topology.LoadZK(topoPath)
	if err != nil {
		return nil, nil, err
	}

	mergeServerConfigs(topo)

	if err := topo.Validate(); err != nil {
		return nil, nil, err
	}

	if store.ZKExists(topo.Name) {
		return nil, nil, fmt.Errorf("ZooKeeper ensemble %q already exists", topo.Name)
	}

	return topo, rawData, nil
}

func printPlan(topo *topology.ZKTopology, version string) {
	fmt.Printf("ZooKeeper ensemble %q will be deployed (version: %s)\n\n", topo.Name, version)
	installPath := zkInstallPath(topo.Path, version)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "MYID\tHOST\tPORTS\tDIRECTORIES")
	fmt.Fprintln(w, "\t\t\t\t")
	for _, s := range topo.Servers {
		host, clientPort, quorumPort, electionPort := s.ParseAddress()
		ports := fmt.Sprintf("%s/%s/%s", clientPort, quorumPort, electionPort)
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", s.MyID, host, ports, installPath)
	}
	w.Flush()

	fmt.Println()
	fmt.Println("Attention:")
	fmt.Println("  1. If the topology is not what you expected, check your yaml file.")
	fmt.Println("  2. Please confirm there is no port/directory conflicts in same host.")
}

func printRecoveryGuide(deployed []topology.ZKServer, topo *topology.ZKTopology) {
	if len(deployed) == 0 {
		return
	}

	fmt.Println("\nDeployment failed. Manual recovery required.")
	fmt.Println("The following servers have been partially installed:")
	for _, s := range deployed {
		fmt.Printf("  - %s (myid=%d)\n", s.Host(), s.MyID)
	}
	fmt.Println("\nTo clean up, manually run on each server:")
	for _, server := range deployed {
		confDir := zkConfigDir(topo.Path, topo.Name, server.MyID)
		dataDir := zkNodeDataPath(server.Config.DataDir, server.MyID)
		dataLogDir := zkNodeDataPath(server.Config.DataLogDir, server.MyID)

		fmt.Printf("  %s:\n", server.Host())
		fmt.Printf("      rm -rf %s %s %s\n", confDir, dataDir, dataLogDir)
	}
}
