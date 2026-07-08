package zk

import (
	"fmt"
	"path"

	"github.com/jam2in/arcusctl/internal/topology"
)

func pickServer(topoServers []topology.ZKServer, myID int) (*topology.ZKServer, error) {
	for i := range topoServers {
		if topoServers[i].MyID == myID {
			return &topoServers[i], nil
		}
	}

	return nil, fmt.Errorf("ZooKeeper server myid=%d not found", myID)
}

func zkServerScript(topoPath string) string {
	return path.Join(topoPath, "bin", "zkServer.sh")
}

func zkConfigPath(topoPath string, myID int) string {
	configDir := fmt.Sprintf("conf_myid_%d", myID)
	return path.Join(topoPath, configDir, "zoo.cfg")
}
