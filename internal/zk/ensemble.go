package zk

import (
	"fmt"
	"path"

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
	globalConfig := topo.GlobalConfig

	if topo.GlobalConfig.DataDir == "" {
		globalConfig.DataDir = path.Join(topo.Path, "data", topo.Name)
	}

	for i := range topo.Servers {
		merged := mergeConfig(globalConfig, topo.Servers[i].Config)
		topo.Servers[i].Config = &merged
	}
}

// sharingEnsemble returns the name of another ensemble
// that shares the same installation path and has a server on the given host.
func sharingEnsemble(
	ensembleName string,
	installPath string,
	host string,
) (string, error) {
	registered, err := store.ListZK()
	if err != nil {
		return "", err
	}

	for _, other := range registered {
		if other == ensembleName {
			continue
		}

		meta, err := store.LoadZKMeta(other)
		if err != nil {
			continue
		}

		topo, err := store.LoadZKTopology(other)
		if err != nil {
			continue
		}

		otherInstallPath := zkInstallPath(topo.Path, meta.Version)
		if otherInstallPath == installPath &&
			hasServerOn(topo.Servers, host) {
			return other, nil
		}
	}

	return "", nil
}

func hasServerOn(
	zkServers []topology.ZKServer,
	host string,
) bool {
	for _, server := range zkServers {
		if server.Host() == host {
			return true
		}
	}

	return false
}
