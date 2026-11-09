package cluster

import (
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcusctl/internal"
	"github.com/jam2in/arcusctl/internal/topology"
)

type cacheNode struct {
	Address string
	Group   string
	Role    string
}

type cacheList struct {
	Nodes   map[string]cacheNode
	Missing bool
}

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
		leaf := mappingNodeName(topo, server, edition)
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

func unregisterZNodes(
	topo *topology.ClusterTopology,
	edition topology.ClusterEdition,
) error {
	conn, err := internal.ConnectZooKeeper(topo.ZooKeeper)
	if err != nil {
		return err
	}
	defer conn.Close()

	root := rootForEdition(edition)

	// cache_server_mapping
	for _, server := range topo.Servers {
		zpath := path.Join(root, internal.ZPATH_CACHE_SERVER_MAPPING, server.Address)
		if err := internal.DeleteZNode(conn, zpath); err != nil {
			return fmt.Errorf("delete znode %s: %w", zpath, err)
		}
	}

	// cache_list, client_list
	for _, base := range []string{internal.ZPATH_CACHE_LIST, internal.ZPATH_CLIENT_LIST} {
		zpath := path.Join(root, base, topo.ServiceCode)
		if err := internal.DeleteZNode(conn, zpath); err != nil {
			return fmt.Errorf("delete znode %s: %w", zpath, err)
		}
	}

	// enterprise: group_list
	if edition == topology.EnterpriseEdition {
		for _, server := range topo.Servers {
			zpath := path.Join(root, internal.ZPATH_GROUP_LIST, topo.ServiceCode, server.Group.Name)
			if err := internal.DeleteZNode(conn, zpath); err != nil {
				return fmt.Errorf("delete znode %s: %w", zpath, err)
			}
		}
	}

	return nil
}

func mappingNodeName(
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

func readCacheList(
	topo *topology.ClusterTopology,
	edition topology.ClusterEdition,
) (cacheList, error) {

	conn, err := internal.ConnectZooKeeper(topo.ZooKeeper)
	if err != nil {
		return cacheList{}, fmt.Errorf(
			"connect to ZooKeeper %s; check ensemble connectivity: %w",
			topo.ZooKeeper,
			err,
		)
	}
	defer conn.Close()

	list, err := readCacheNodes(conn, topo.ServiceCode, edition)
	if err != nil {
		return cacheList{}, fmt.Errorf(
			"read cache_list for %q from ZooKeeper %s; "+
				"check connectivity, permissions and registration: %w",
			topo.ServiceCode,
			topo.ZooKeeper,
			err,
		)
	}

	return list, nil
}

func readCacheNodes(
	conn *zk.Conn,
	serviceCode string,
	edition topology.ClusterEdition,
) (cacheList, error) {
	result := cacheList{
		Nodes: make(map[string]cacheNode),
	}

	root := rootForEdition(edition)
	zpath := path.Join(
		root,
		internal.ZPATH_CACHE_LIST,
		serviceCode,
	)

	children, _, err := conn.Children(zpath)
	if errors.Is(err, zk.ErrNoNode) {
		result.Missing = true
		return result, nil
	}
	if err != nil {
		return result, fmt.Errorf("read children of %s: %w", zpath, err)
	}

	for _, child := range children {
		node, err := parseCacheNode(child, edition)
		if err != nil {
			return result, err
		}

		result.Nodes[node.Address] = node
	}

	return result, nil
}

func parseCacheNode(
	name string,
	edition topology.ClusterEdition,
) (cacheNode, error) {
	node := cacheNode{}
	address := name

	if edition == topology.EnterpriseEdition {
		parts := strings.Split(name, "^")
		if len(parts) != 3 || parts[0] == "" {
			return node, fmt.Errorf(
				"invalid cache_list node %q; check its naming format",
				name,
			)
		}

		node.Group = parts[0]

		switch parts[1] {
		case "M":
			node.Role = "master"
		case "S":
			node.Role = "slave"
		default:
			return node, fmt.Errorf(
				"invalid role in cache_list node %q; expected M or S",
				name,
			)
		}

		address = parts[2]
	}

	colon := strings.LastIndex(address, ":")
	if colon < 0 {
		return node, fmt.Errorf(
			"invalid address in cache_list node %q; expected host:port",
			name,
		)
	}

	port, _, _ := strings.Cut(address[colon+1:], "-")
	address = address[:colon+1] + port

	node.Address = address

	return node, nil
}

func readRegisteredServices(
	zkAddress string,
) ([]registeredService, error) {
	conn, err := internal.ConnectZooKeeper(zkAddress)
	if err != nil {
		return nil, fmt.Errorf("connect to ZooKeeper %s: %w", zkAddress, err)
	}
	defer conn.Close()

	var services []registeredService

	editions := []topology.ClusterEdition{
		topology.CommunityEdition,
		topology.EnterpriseEdition,
	}

	for _, edition := range editions {
		base := path.Join(rootForEdition(edition), internal.ZPATH_CACHE_LIST)

		serviceCodes, _, err := conn.Children(base)
		if errors.Is(err, zk.ErrNoNode) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf(
				"read children of %s from ZooKeeper %s: %w",
				base, zkAddress, err,
			)
		}
		sort.Strings(serviceCodes)

		for _, serviceCode := range serviceCodes {
			zpath := path.Join(base, serviceCode)

			nodes, _, err := conn.Children(zpath)
			if errors.Is(err, zk.ErrNoNode) {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf(
					"read children of %s from ZooKeeper %s: %w",
					zpath, zkAddress, err,
				)
			}

			services = append(services, registeredService{
				ServiceCode: serviceCode,
				Edition:     edition,
				Joined:      len(nodes),
			})
		}
	}

	return services, nil
}
