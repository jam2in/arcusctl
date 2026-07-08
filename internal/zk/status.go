package zk

import (
	"fmt"

	"github.com/jam2in/arcusctl/internal/ssh"
	"github.com/jam2in/arcusctl/internal/topology"
)

func Status(ensembleName string) error {
	meta, topo, err := loadEnsemble(ensembleName)
	if err != nil {
		return err
	}

	fmt.Printf("Ensemble: %s (version: %s)\n\n", meta.Name, meta.Version)

	for _, server := range topo.Servers {
		fmt.Printf("=== %s (myid=%d) ===\n", server.Host(), server.MyID)
		if err := statusServer(server, topo.Path); err != nil {
			fmt.Printf(" error: %v\n", err)
		}
		fmt.Println()
	}

	return nil
}

func statusServer(server topology.ZKServer, topoPath string) error {
	cmd := fmt.Sprintf("%s status %s", zkServerScript(topoPath), zkConfigPath(topoPath, server.MyID))
	return ssh.Run(server.Host(), cmd)
}
