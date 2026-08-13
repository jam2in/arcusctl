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
		if err := statusServer(server, topo.Path, topo.Name, meta.Version); err != nil {
			fmt.Printf(" error: %v\n", err)
		}
		fmt.Println()
	}

	return nil
}

func statusServer(
	server topology.ZKServer,
	topoPath string,
	ensembleName string,
	version string,
) error {
	scriptPath := zkServerScript(topoPath, version)
	configPath := zkConfigPath(topoPath, ensembleName, server.MyID)
	cmd := fmt.Sprintf("%s status %s", scriptPath, configPath)
	return ssh.Run(server.Host(), cmd)
}
