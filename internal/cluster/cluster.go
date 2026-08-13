package cluster

import (
	"github.com/jam2in/arcusctl/internal/store"
	"github.com/jam2in/arcusctl/internal/topology"
)

func loadCluster(serviceCode string) (
	*store.ClusterMeta,
	*topology.ClusterTopology,
	topology.ClusterEdition,
	error,
) {
	meta, err := store.LoadClusterMeta(serviceCode)
	if err != nil {
		return nil, nil, "", err
	}

	topo, err := store.LoadClusterTopology(serviceCode)
	if err != nil {
		return nil, nil, "", err
	}

	edition, err := topo.Edition()
	if err != nil {
		return nil, nil, "", err
	}

	return meta, topo, edition, nil
}

// sharingCluster return, for each host, the service code of another cluster
// installed there from same path and version.
func sharingCluster(
	serviceCode string,
	installPath string,
	host string,
) (string, error) {
	registered, err := store.ListCluster()
	if err != nil {
		return "", err
	}

	for _, other := range registered {
		if other == serviceCode {
			continue
		}

		meta, err := store.LoadClusterMeta(other)
		if err != nil {
			continue
		}

		topo, err := store.LoadClusterTopology(other)
		if err != nil {
			continue
		}

		otherInstallPath := memcachedInstallPath(topo.Path, meta.Version)
		if installPath == otherInstallPath &&
			hasServerOn(topo.Servers, host) {
			return other, nil
		}
	}

	return "", nil
}

func hasServerOn(
	cacheServers []topology.CacheServer,
	host string,
) bool {
	for _, server := range cacheServers {
		if server.Host() == host {
			return true
		}
	}
	return false
}
