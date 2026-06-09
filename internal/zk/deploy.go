package zk

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/jam2in/arcusctl/internal/store"
	"github.com/jam2in/arcusctl/internal/topology"
)

func Deploy(version string, topologyPath string) error {
	topo, topologyBytes, err := prepareTopology(topologyPath)
	if err != nil {
		return err
	}

	printPlan(topo, version)
	if !confirm() {
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

func prepareTopology(topologyPath string) (*topology.ZKTopology, []byte, error) {
	topo, rawData, err := topology.LoadZK(topologyPath)
	if err != nil {
		return nil, nil, err
	}

	for i := range topo.Servers {
		merged := mergeConfig(topo.GlobalConfig, topo.Servers[i].Config)
		topo.Servers[i].Config = &merged
	}

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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ROLE\tHOST\tPORTS\tDIRECTORIES")
	fmt.Fprintln(w, "\t\t\t\t")
	for _, s := range topo.Servers {
		host, clientPort, quorumPort, electionPort := s.ParseAddress()
		ports := fmt.Sprintf("%s/%s/%s", clientPort, quorumPort, electionPort)
		fmt.Fprintf(w, "zookeeper\t%s\t%s\t%s\n", host, ports, topo.Path)
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
	fmt.Printf("  rm -rf %s/conf_myid_<myid>\n", topo.Path)
}

func confirm() bool {
	// FIXME: internal.ReadStdin()으로 변경 필요
	fmt.Print("\nProceed? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}
