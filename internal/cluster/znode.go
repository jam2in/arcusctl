package cluster

import (
	"fmt"
	"path"

	"github.com/jam2in/arcusctl/internal"
	"github.com/jam2in/arcusctl/internal/topology"
)

func clusterExistsInZK(zkAddress string, serviceCode string, edition topology.ClusterEdition) (bool, error) {
	conn, err := internal.ConnectZooKeeper(zkAddress)
	if err != nil {
		return false, fmt.Errorf("connect to ZooKeeper: %w", err)
	}
	defer conn.Close()

	serviceCodePath := path.Join(rootForEdition(edition), internal.ZPATH_CACHE_LIST, serviceCode)
	exists, _, err := conn.Exists(serviceCodePath)
	if err != nil {
		return false, fmt.Errorf("check service code %q in ZooKeeper: %w", serviceCode, err)
	}

	return exists, nil
}

func rootForEdition(edition topology.ClusterEdition) string {
	switch edition {
	case topology.CommunityEdition:
		return internal.ZPATH_ROOT

	case topology.EnterpriseEdition:
		return internal.ZPATH_REPL_ROOT

	default:
		panic(fmt.Sprintf("unsupported cluster edition: %q", edition))
	}
}

func registerZNodes(
	topo *topology.ClusterTopology,
	edition topology.ClusterEdition,
) error {
	conn, err := internal.ConnectZooKeeper(topo.ZooKeeper)
	if err != nil {
		return err
	}
	defer conn.Close()

	root := rootForEdition(edition)

	// service code based path for cache_list and client_list
	for _, base := range []string{internal.ZPATH_CACHE_LIST, internal.ZPATH_CLIENT_LIST} {
		zpath := path.Join(root, base, topo.ServiceCode)
		if err := internal.EnsureZNode(conn, zpath); err != nil {
			return fmt.Errorf("create znode %s: %w", zpath, err)
		}
	}

	// cache server mapping path
	for _, server := range topo.Servers {
		leaf := mappingNode(topo, server, edition)
		zpath := path.Join(root, internal.ZPATH_CACHE_SERVER_MAPPING, server.Address, leaf)
		if err := internal.EnsureZNode(conn, zpath); err != nil {
			return fmt.Errorf("create znode %s: %w", zpath, err)
		}
	}

	// enterprise: group list path
	if edition == topology.EnterpriseEdition {
		for _, server := range topo.Servers {
			zpath := path.Join(root, internal.ZPATH_GROUP_LIST, topo.ServiceCode, server.Group.Name)
			if err := internal.EnsureZNode(conn, zpath); err != nil {
				return fmt.Errorf("create znode %s: %w", zpath, err)
			}
		}
	}

	return nil
}

func mappingNode(
	topo *topology.ClusterTopology,
	server topology.CacheServer,
	edition topology.ClusterEdition,
) string {
	if edition == topology.CommunityEdition {
		return topo.ServiceCode
	}
	return fmt.Sprintf("%s^%s^%s:%d",
		topo.ServiceCode, server.Group.Name, server.Host(), server.Group.Port)
}
