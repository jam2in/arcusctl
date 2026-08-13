package zk

import (
	"fmt"

	"github.com/jam2in/arcusctl/internal/ssh"
	"github.com/jam2in/arcusctl/internal/topology"
)

func Start(ensembleName string, myID int) error {
	meta, topo, err := loadEnsemble(ensembleName)
	if err != nil {
		return err
	}

	// Start a specific server if myID is provided
	if myID != 0 {
		server, err := pickServer(topo.Servers, myID)
		if err != nil {
			return err
		}
		if err := startServer(*server, topo.Path, topo.Name, meta.Version); err != nil {
			return err
		}
		fmt.Printf("ZooKeeper node %s (myid=%d) started successfully.\n", server.Host(), myID)
		return nil
	}

	for _, server := range topo.Servers {
		if err := startServer(server, topo.Path, topo.Name, meta.Version); err != nil {
			return err
		}
	}

	fmt.Printf("ZooKeeper ensemble %q started successfully.\n", ensembleName)
	return nil
}

func startServer(
	server topology.ZKServer,
	topoPath string,
	ensembleName string,
	version string,
) error {
	fmt.Printf("Starting %s (myid=%d)...\n", server.Host(), server.MyID)

	scriptPath := zkServerScript(topoPath, version)
	configPath := zkConfigPath(topoPath, ensembleName, server.MyID)
	cmd := fmt.Sprintf("%s start %s", scriptPath, configPath)

	if err := ssh.Run(server.Host(), cmd); err != nil {
		return fmt.Errorf("start %s: %w", server.Host(), err)
	}

	return nil
}
