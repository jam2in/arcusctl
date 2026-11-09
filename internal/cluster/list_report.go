package cluster

import (
	"fmt"
	"strconv"

	"github.com/jam2in/arcusctl/internal/store"
	"github.com/jam2in/arcusctl/internal/topology"
)

type clusterListRow struct {
	ServiceCode string
	Edition     string
	Joined      string
	Topology    string
	Groups      string
	DeployedAt  string
}

type listWarning struct {
	ServiceCode string
	Reasons     []string
}

type registeredService struct {
	ServiceCode string
	Edition     topology.ClusterEdition
	Joined      int
}

func buildListRow(
	meta *store.ClusterMeta,
	topo *topology.ClusterTopology,
	edition topology.ClusterEdition,
	list cacheList,
	cacheListErr error,
) (clusterListRow, []string) {
	row := newListRow(meta, topo, edition)

	if cacheListErr != nil {
		return row, []string{cacheListErr.Error()}
	}

	if list.Missing {
		row.setMembership(edition, 0, 0)
		return row, []string{"not registered in cache_list"}
	}

	joined, groups := countJoined(topo, list)
	row.setMembership(edition, joined, groups)

	return row, extraNodeReasons(topo, list)
}

func newListRow(
	meta *store.ClusterMeta,
	topo *topology.ClusterTopology,
	edition topology.ClusterEdition,
) clusterListRow {
	row := clusterListRow{
		ServiceCode: meta.ServiceCode,
		Edition:     string(edition),
		Joined:      "unknown",
		Topology:    strconv.Itoa(len(topo.Servers)),
		Groups:      "-",
		DeployedAt:  meta.DeployedAt.Format("2006-01-02 15:04:05"),
	}
	if edition == topology.EnterpriseEdition {
		row.Groups = "unknown"
	}

	return row
}

func (row *clusterListRow) setMembership(
	edition topology.ClusterEdition,
	joined int,
	groups int,
) {
	row.Joined = strconv.Itoa(joined)
	if edition == topology.EnterpriseEdition {
		row.Groups = strconv.Itoa(groups)
	}
}

func countJoined(
	topo *topology.ClusterTopology,
	list cacheList,
) (joined int, groups int) {
	seen := make(map[string]bool)

	for _, server := range topo.Servers {
		node, ok := list.Nodes[server.Address]
		if !ok {
			continue
		}

		joined++
		if node.Group != "" {
			seen[node.Group] = true
		}
	}

	return joined, len(seen)
}

func unmanagedWarnings(
	zkAddress string,
	registered []registeredService,
	managed map[string]bool,
) []listWarning {
	var warnings []listWarning

	for _, service := range registered {
		if managed[service.ServiceCode] {
			continue
		}

		warnings = append(warnings, listWarning{
			ServiceCode: service.ServiceCode,
			Reasons: []string{fmt.Sprintf(
				"registered in ZooKeeper %s (edition: %s, joined: %d) "+
					"but no deployment record on this machine",
				zkAddress, service.Edition, service.Joined,
			)},
		})
	}

	return warnings
}

func extraNodeReasons(
	topo *topology.ClusterTopology,
	list cacheList,
) []string {
	_, extra := statusCacheListWarnings(topo, list, nil)

	var reasons []string
	for _, warning := range extra {
		for _, reason := range warning.Reasons {
			reasons = append(reasons, fmt.Sprintf("%s: %s", warning.Address, reason))
		}
	}

	return reasons
}
