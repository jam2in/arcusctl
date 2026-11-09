package cluster

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/jam2in/arcusctl/internal/topology"
)

type nodeStatus struct {
	Address string
	Group   string
	Role    string
	Version string
	ZK      string
}

type nodeWarning struct {
	Address  string
	Messages []string
}

type clusterStatus struct {
	Nodes           []nodeStatus
	Healthy         int
	CacheListFound  bool
	ClusterWarnings []string
	NodeWarnings    []nodeWarning
}

func checkCluster(
	topo *topology.ClusterTopology,
	list cacheList,
	cacheListErr error,
	probes []cacheProbe,
) clusterStatus {
	status := clusterStatus{
		CacheListFound: cacheListErr == nil,
	}
	if status.CacheListFound && list.Unregistered {
		status.ClusterWarnings = []string{fmt.Sprintf(
			"service code %q is not registered in cache_list",
			topo.ServiceCode,
		)}
	}

	for i, server := range topo.Servers {
		node := buildNodeStatus(
			server, list, status.CacheListFound, probes[i],
		)

		status.Nodes = append(status.Nodes, node)
		if messages := nodeWarningMessages(server, node, probes[i]); len(messages) > 0 {
			status.NodeWarnings = append(status.NodeWarnings, nodeWarning{
				Address:  server.Address,
				Messages: messages,
			})
		}

		if probes[i].Err == nil && node.ZK == "joined" {
			status.Healthy++
		}
	}

	// A failed cache_list read may leave partial data; do not trust it.
	if status.CacheListFound {
		status.NodeWarnings = append(status.NodeWarnings, unexpectedNodeWarnings(topo, list)...)
	}
	return status
}

func buildNodeStatus(
	server topology.CacheServer,
	list cacheList,
	cacheListFound bool,
	probe cacheProbe,
) nodeStatus {
	node := nodeStatus{
		Address: server.Address,
		Group:   "-",
		Role:    "-",
		Version: "-",
		ZK:      "not joined",
	}

	if probe.Err == nil {
		node.Version = probe.Version
	}

	// GROUP and ROLE show only the actual membership in cache_list;
	// topology values appear in warnings.
	if !cacheListFound {
		node.ZK = "unknown"
		if server.Group != nil {
			node.Group = "unknown"
			node.Role = "unknown"
		}
		return node
	}

	if entry, joined := list.Nodes[server.Address]; joined {
		node.ZK = "joined"
		if server.Group != nil {
			node.Group = entry.Group
			node.Role = entry.Role
		}
	}

	return node
}

func nodeWarningMessages(
	server topology.CacheServer,
	node nodeStatus,
	probe cacheProbe,
) []string {
	var messages []string

	if probe.Err != nil {
		messages = append(messages, fmt.Sprintf(
			"not responding (%s)",
			rootCause(probe.Err),
		))
	}

	switch node.ZK {
	case "not joined":
		message := "not joined in cache_list"
		if server.Group != nil {
			message += fmt.Sprintf(
				" (topology=%s/%s)",
				server.Group.Name, server.Group.Role,
			)
		}
		messages = append(messages, message)

	case "joined":
		if server.Group != nil &&
			(server.Group.Name != node.Group || server.Group.Role != node.Role) {
			messages = append(messages, fmt.Sprintf(
				"group/role differs (topology=%s/%s, actual=%s/%s)",
				server.Group.Name, server.Group.Role,
				node.Group, node.Role,
			))
		}
	}

	return messages
}

func rootCause(err error) string {
	for {
		cause := errors.Unwrap(err)
		if cause == nil {
			return err.Error()
		}
		err = cause
	}
}

// unexpectedNodeWarnings reports nodes joined in cache_list
// that are not declared in the stored topology.
func unexpectedNodeWarnings(
	topo *topology.ClusterTopology,
	list cacheList,
) []nodeWarning {
	expected := make(map[string]bool, len(topo.Servers))
	for _, server := range topo.Servers {
		expected[server.Address] = true
	}

	var warnings []nodeWarning
	for address, entry := range list.Nodes {
		if expected[address] {
			continue
		}

		message := "joined in cache_list but not in the stored topology"
		if entry.Group != "" {
			message += fmt.Sprintf(" (group=%s, role=%s)", entry.Group, entry.Role)
		}

		warnings = append(warnings, nodeWarning{
			Address:  address,
			Messages: []string{message},
		})
	}

	slices.SortFunc(warnings, func(a, b nodeWarning) int {
		return strings.Compare(a.Address, b.Address)
	})
	return warnings
}
