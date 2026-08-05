package cluster

import (
	"fmt"

	"github.com/jam2in/arcusctl/internal/ssh"
	"github.com/jam2in/arcusctl/internal/topology"
)

const stopCommandTemplate = "[ -f %s ] && kill $(cat %s) 2>/dev/null || true"

func Stop(serviceCode string, nodeAddress string, groupName string) error {
	meta, topo, edition, err := loadCluster(serviceCode)
	if err != nil {
		return err
	}

	targets, err := selectTargets(topo, edition, nodeAddress, groupName, slaveFirst)
	if err != nil {
		return err
	}

	total := len(targets)
	for i, server := range targets {
		fmt.Printf("[%d/%d] %s: stopping...\n", i+1, total, server.Address)
		if err := stopServer(server, topo.Path, meta.Version); err != nil {
			return err
		}
	}

	fmt.Printf("Arcus cluster %q stopped.\n", serviceCode)
	return nil
}

func stopServer(
	server topology.CacheServer,
	topoPath string,
	version string,
) error {
	if err := ssh.Run(server.Host(), stopCommand(server.Address, topoPath, version)); err != nil {
		return fmt.Errorf("stop %s: %w", server.Address, err)
	}
	return nil
}

func stopCommand(
	serverAddress string,
	topoPath string,
	version string,
) string {
	pidPath := pidFilePath(serverAddress, topoPath, version)
	return fmt.Sprintf(stopCommandTemplate, pidPath, pidPath)
}
