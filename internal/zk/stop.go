package zk

import (
	"fmt"

	"github.com/jam2in/arcusctl/internal/ssh"
	"github.com/jam2in/arcusctl/internal/topology"
)

func Stop(ensembleName string, myID int) error {
	_, topo, err := loadEnsemble(ensembleName)
	if err != nil {
		return err
	}

	// Stop a specific server if myID is provided
	if myID != 0 {
		server, err := pickServer(topo.Servers, myID)
		if err != nil {
			return err
		}

		if err := stopServer(*server, topo.Path); err != nil {
			return err
		}

		fmt.Printf("ZooKeeper server %s (myid=%d) stopped.\n", server.Host(), server.MyID)
		return nil
	}

	for _, server := range topo.Servers {
		if err := stopServer(server, topo.Path); err != nil {
			return err
		}
	}

	fmt.Printf("ZooKeeper ensemble %q stopped.\n", ensembleName)
	return nil
}

func stopServer(server topology.ZKServer, topoPath string) error {
	fmt.Printf("Stopping %s (myid=%d)...\n", server.Host(), server.MyID)
	cmd := fmt.Sprintf("%s stop %s", zkServerScript(topoPath), zkConfigPath(topoPath, server.MyID))
	if err := ssh.Run(server.Host(), cmd); err != nil {
		return fmt.Errorf("stop %s: %w", server.Host(), err)
	}

	return nil
}
