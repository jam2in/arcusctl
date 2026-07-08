package zk

import (
	"fmt"

	"github.com/jam2in/arcusctl/internal/store"
	"github.com/jam2in/arcusctl/internal/topology"
)

func loadEnsemble(name string) (*store.ZKMeta, *topology.ZKTopology, error) {
	meta, err := store.LoadZKMeta(name)
	if err != nil {
		return nil, nil, fmt.Errorf("load ZooKeeper metadata %q: %w", name, err)
	}

	topo, err := store.LoadZKTopology(name)
	if err != nil {
		return nil, nil, fmt.Errorf("load ZooKeeper topology %q: %w", name, err)
	}

	if topo.Name != name {
		return nil, nil, fmt.Errorf("ZooKeeper topology name mismatch: got %q, want %q", topo.Name, name)
	}

	mergeServerConfigs(topo)

	if err := topo.Validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid ZooKeeper topology %q: %w", name, err)
	}

	return meta, topo, nil
}

func mergeServerConfigs(topo *topology.ZKTopology) {
	for i := range topo.Servers {
		merged := mergeConfig(topo.GlobalConfig, topo.Servers[i].Config)
		topo.Servers[i].Config = &merged
	}
}
