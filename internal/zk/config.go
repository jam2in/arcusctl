package zk

import (
	"fmt"
	"strings"

	"github.com/jam2in/arcusctl/internal/topology"
)

const (
	defaultTickTime  = 2000
	defaultInitLimit = 10
	defaultSyncLimit = 5
)

func buildConfig(server topology.ZKServer, topo *topology.ZKTopology) string {
	var sb strings.Builder

	cfg := server.Config
	dataDir := zkNodeDataPath(cfg.DataDir, server.MyID)
	dataLogDir := zkNodeDataPath(cfg.DataLogDir, server.MyID)
	dynamicConfigPath := zkDynamicConfigPath(topo.Path, topo.Name, server.MyID)

	fmt.Fprintf(&sb, "tickTime=%d\n", cfg.TickTime)
	fmt.Fprintf(&sb, "initLimit=%d\n", cfg.InitLimit)
	fmt.Fprintf(&sb, "syncLimit=%d\n", cfg.SyncLimit)
	fmt.Fprintf(&sb, "dataDir=%s\n", dataDir)
	fmt.Fprintf(&sb, "dataLogDir=%s\n", dataLogDir)
	fmt.Fprintf(&sb, "dynamicConfigFile=%s\n", dynamicConfigPath)

	sb.WriteString("standaloneEnabled=false\n")
	sb.WriteString("reconfigEnabled=true\n")
	sb.WriteString("4lw.commands.whitelist=*\n")

	for k, v := range cfg.Properties {
		fmt.Fprintf(&sb, "%s=%s\n", k, v)
	}

	return sb.String()
}

func buildDynamicConfig(topo *topology.ZKTopology) string {
	var sb strings.Builder

	for _, s := range topo.Servers {
		host, clientPort, quorumPort, electionPort := s.ParseAddress()
		fmt.Fprintf(&sb, "server.%d=%s:%s:%s;%s\n",
			s.MyID, host, quorumPort, electionPort, clientPort)
	}

	return sb.String()
}

func mergeConfig(globalConfig topology.ZKConfig, nodeConfig *topology.ZKConfig) topology.ZKConfig {
	merged := globalConfig
	if nodeConfig != nil {
		if nodeConfig.TickTime > 0 {
			merged.TickTime = nodeConfig.TickTime
		}
		if nodeConfig.InitLimit > 0 {
			merged.InitLimit = nodeConfig.InitLimit
		}
		if nodeConfig.SyncLimit > 0 {
			merged.SyncLimit = nodeConfig.SyncLimit
		}
		if nodeConfig.DataDir != "" {
			merged.DataDir = nodeConfig.DataDir
		}
		if nodeConfig.DataLogDir != "" {
			merged.DataLogDir = nodeConfig.DataLogDir
		}
		if nodeConfig.Properties != nil {
			if merged.Properties == nil {
				merged.Properties = map[string]string{}
			}
			for k, v := range nodeConfig.Properties {
				merged.Properties[k] = v
			}
		}
	}

	if merged.TickTime == 0 {
		merged.TickTime = defaultTickTime
	}
	if merged.InitLimit == 0 {
		merged.InitLimit = defaultInitLimit
	}
	if merged.SyncLimit == 0 {
		merged.SyncLimit = defaultSyncLimit
	}
	if merged.DataLogDir == "" {
		merged.DataLogDir = merged.DataDir
	}

	return merged
}
