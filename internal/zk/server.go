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

func zkServerScript(topoPath string, version string) string {
	return path.Join(zkInstallPath(topoPath, version), "bin", "zkServer.sh")
}

func zkInstallPath(topoPath string, version string) string {
	return path.Join(topoPath, version)
}

func zkConfigDir(topoPath string, ensembleName string, myID int) string {
	return path.Join(topoPath, "conf", ensembleName, fmt.Sprintf("zk%d", myID))
}

func zkConfigPath(topoPath string, ensembleName string, myID int) string {
	return path.Join(zkConfigDir(topoPath, ensembleName, myID), "zoo.cfg")
}

func zkDynamicConfigPath(topoPath string, ensembleName string, myID int) string {
	return path.Join(zkConfigDir(topoPath, ensembleName, myID), "zoo.cfg.dynamic")
}

func zkNodeDataPath(dataDir string, myID int) string {
	return path.Join(dataDir, fmt.Sprintf("zk%d", myID))
}
