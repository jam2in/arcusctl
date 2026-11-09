package cluster

import (
	"errors"
	"fmt"
	"sort"

	"github.com/jam2in/arcusctl/internal/topology"
)

type clusterStatusRow struct {
	Address    string
	Group      string
	Role       string
	Version    string
	Responding string
	ZK         string
}

type nodeWarning struct {
	Address string
	Reasons []string
}

type clusterStatusReport struct {
	Rows           []clusterStatusRow
	Healthy        int
	CacheListKnown bool
	Warnings       []string
	NodeWarnings   []nodeWarning
}

func evaluateClusterStatus(
	topo *topology.ClusterTopology,
	list cacheList,
	cacheListErr error,
	probes []cacheProbe,
) clusterStatusReport {
	warnings, extra := statusCacheListWarnings(topo, list, cacheListErr)
	report := clusterStatusReport{
		CacheListKnown: cacheListErr == nil,
		Warnings:       warnings,
	}

	for i, server := range topo.Servers {
		row := buildStatusRow(
			server, list, report.CacheListKnown, probes[i],
		)

		report.Rows = append(report.Rows, row)
		if reasons := statusServerWarnings(server, row, probes[i]); len(reasons) > 0 {
			report.NodeWarnings = append(report.NodeWarnings, nodeWarning{
				Address: server.Address,
				Reasons: reasons,
			})
		}

		if row.Responding == "yes" && row.ZK == "joined" {
			report.Healthy++
		}
	}

	report.NodeWarnings = append(report.NodeWarnings, extra...)
	return report
}

func buildStatusRow(
	server topology.CacheServer,
	list cacheList,
	cacheListKnown bool,
	probe cacheProbe,
) clusterStatusRow {
	row := clusterStatusRow{
		Address:    server.Address,
		Group:      "-",
		Role:       "-",
		Version:    "-",
		Responding: "no",
		ZK:         "not joined",
	}

	if probe.Err == nil {
		row.Version = probe.Version
		row.Responding = "yes"
	}

	// GROUP and ROLE show only the actual membership in cache_list;
	// topology values appear in warnings.
	if !cacheListKnown {
		row.ZK = "unknown"
		if server.Group != nil {
			row.Group = "unknown"
			row.Role = "unknown"
		}
		return row
	}

	if node, joined := list.Nodes[server.Address]; joined {
		row.ZK = "joined"
		if server.Group != nil {
			row.Group = node.Group
			row.Role = node.Role
		}
	}

	return row
}

func statusServerWarnings(
	server topology.CacheServer,
	row clusterStatusRow,
	probe cacheProbe,
) []string {
	var reasons []string

	if probe.Err != nil {
		reasons = append(reasons, fmt.Sprintf(
			"not responding (%s)",
			probeErrorReason(probe.Err),
		))
	}

	switch row.ZK {
	case "not joined":
		reason := "not joined in cache_list"
		if server.Group != nil {
			reason += fmt.Sprintf(
				" (topology=%s/%s)",
				server.Group.Name, server.Group.Role,
			)
		}
		reasons = append(reasons, reason)

	case "joined":
		if server.Group != nil &&
			(server.Group.Name != row.Group || server.Group.Role != row.Role) {
			reasons = append(reasons, fmt.Sprintf(
				"group/role differs (topology=%s/%s, actual=%s/%s)",
				server.Group.Name, server.Group.Role,
				row.Group, row.Role,
			))
		}
	}

	return reasons
}

func probeErrorReason(err error) string {
	for {
		cause := errors.Unwrap(err)
		if cause == nil {
			return err.Error()
		}
		err = cause
	}
}

func statusCacheListWarnings(
	topo *topology.ClusterTopology,
	list cacheList,
	cacheListErr error,
) ([]string, []nodeWarning) {
	if cacheListErr != nil {
		return nil, nil
	}

	if list.Missing {
		return []string{fmt.Sprintf(
			"service code %q is not registered in cache_list",
			topo.ServiceCode,
		)}, nil
	}

	declared := make(map[string]bool, len(topo.Servers))
	for _, server := range topo.Servers {
		declared[server.Address] = true
	}

	var extra []string
	for address := range list.Nodes {
		if !declared[address] {
			extra = append(extra, address)
		}
	}
	sort.Strings(extra)

	var warnings []nodeWarning
	for _, address := range extra {
		node := list.Nodes[address]
		reason := "joined in cache_list but not in the stored topology"

		if node.Group != "" {
			reason += fmt.Sprintf(
				" (group=%s, role=%s)", node.Group, node.Role,
			)
		}

		warnings = append(warnings, nodeWarning{
			Address: address,
			Reasons: []string{reason},
		})
	}

	return nil, warnings
}
