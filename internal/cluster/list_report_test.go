package cluster

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/jam2in/arcusctl/internal/store"
	"github.com/jam2in/arcusctl/internal/topology"
)

func TestBuildListRow(t *testing.T) {
	const (
		first  = "192.0.2.21:11211"
		second = "192.0.2.22:11211"
		third  = "192.0.2.23:11211"
		extra  = "192.0.2.29:11211"
	)

	meta := &store.ClusterMeta{
		ServiceCode: "test",
		DeployedAt:  time.Date(2026, 9, 1, 18, 30, 31, 0, time.UTC),
	}

	communityServers := []topology.CacheServer{
		{Address: first}, {Address: second}, {Address: third},
	}
	enterpriseServers := []topology.CacheServer{
		{Address: first, Group: &topology.GroupInfo{Name: "g1", Role: "master"}},
		{Address: second, Group: &topology.GroupInfo{Name: "g1", Role: "slave"}},
		{Address: third, Group: &topology.GroupInfo{Name: "g2", Role: "master"}},
	}

	row := func(edition, joined, groups string) clusterListRow {
		return clusterListRow{
			ServiceCode: "test",
			Edition:     edition,
			Joined:      joined,
			Topology:    "3",
			Groups:      groups,
			DeployedAt:  "2026-09-01 18:30:31",
		}
	}

	cacheListErr := errors.New("connect to ZooKeeper 192.0.2.11:2181")

	tests := []struct {
		name         string
		servers      []topology.CacheServer
		edition      topology.ClusterEdition
		list         cacheList
		cacheListErr error
		want         clusterListRow
		wantReasons  []string
	}{
		{
			name:    "community: all declared nodes joined",
			servers: communityServers,
			edition: topology.CommunityEdition,
			list: cacheList{Nodes: map[string]cacheNode{
				first: {Address: first}, second: {Address: second}, third: {Address: third},
			}},
			want: row("community", "3", "-"),
		},
		{
			name:    "community: stopped nodes lower JOINED without warnings",
			servers: communityServers,
			edition: topology.CommunityEdition,
			list: cacheList{Nodes: map[string]cacheNode{
				first: {Address: first},
			}},
			want: row("community", "1", "-"),
		},
		{
			name:    "community: undeclared node is not counted and is warned",
			servers: communityServers,
			edition: topology.CommunityEdition,
			list: cacheList{Nodes: map[string]cacheNode{
				first: {Address: first}, extra: {Address: extra},
			}},
			want: row("community", "1", "-"),
			wantReasons: []string{
				extra + ": joined in cache_list but not in the stored topology",
			},
		},
		{
			name:        "community: service code missing in cache_list",
			servers:     communityServers,
			edition:     topology.CommunityEdition,
			list:        cacheList{Missing: true},
			want:        row("community", "0", "-"),
			wantReasons: []string{"not registered in cache_list"},
		},
		{
			name:         "community: unreadable cache_list leaves JOINED unknown",
			servers:      communityServers,
			edition:      topology.CommunityEdition,
			cacheListErr: cacheListErr,
			want:         row("community", "unknown", "-"),
			wantReasons:  []string{cacheListErr.Error()},
		},
		{
			name:    "enterprise: GROUPS counts groups of joined nodes",
			servers: enterpriseServers,
			edition: topology.EnterpriseEdition,
			list: cacheList{Nodes: map[string]cacheNode{
				first:  {Address: first, Group: "g1", Role: "master"},
				second: {Address: second, Group: "g1", Role: "slave"},
				third:  {Address: third, Group: "g2", Role: "master"},
			}},
			want: row("enterprise", "3", "2"),
		},
		{
			name:    "enterprise: a group with no joined node is not counted",
			servers: enterpriseServers,
			edition: topology.EnterpriseEdition,
			list: cacheList{Nodes: map[string]cacheNode{
				first:  {Address: first, Group: "g1", Role: "master"},
				second: {Address: second, Group: "g1", Role: "slave"},
			}},
			want: row("enterprise", "2", "1"),
		},
		{
			name:    "enterprise: undeclared node's group is not counted",
			servers: enterpriseServers,
			edition: topology.EnterpriseEdition,
			list: cacheList{Nodes: map[string]cacheNode{
				first: {Address: first, Group: "g1", Role: "master"},
				extra: {Address: extra, Group: "g3", Role: "master"},
			}},
			want: row("enterprise", "1", "1"),
			wantReasons: []string{
				extra + ": joined in cache_list but not in the stored topology (group=g3, role=master)",
			},
		},
		{
			name:        "enterprise: service code missing in cache_list",
			servers:     enterpriseServers,
			edition:     topology.EnterpriseEdition,
			list:        cacheList{Missing: true},
			want:        row("enterprise", "0", "0"),
			wantReasons: []string{"not registered in cache_list"},
		},
		{
			name:         "enterprise: unreadable cache_list leaves GROUPS unknown",
			servers:      enterpriseServers,
			edition:      topology.EnterpriseEdition,
			cacheListErr: cacheListErr,
			want:         row("enterprise", "unknown", "unknown"),
			wantReasons:  []string{cacheListErr.Error()},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			topo := &topology.ClusterTopology{ServiceCode: "test", Servers: test.servers}
			got, reasons := buildListRow(meta, topo, test.edition, test.list, test.cacheListErr)
			if got != test.want {
				t.Errorf("row = %+v, want %+v", got, test.want)
			}
			if !slices.Equal(reasons, test.wantReasons) {
				t.Errorf("reasons = %q, want %q", reasons, test.wantReasons)
			}
		})
	}
}

func TestUnmanagedWarnings(t *testing.T) {
	const zkAddress = "192.0.2.11:2181"
	managed := map[string]bool{"test-community": true, "test-enterprise": true}

	tests := []struct {
		name       string
		registered []registeredService
		want       []listWarning
	}{
		{
			name: "no service code in cache_list",
		},
		{
			name: "managed service codes are not reported",
			registered: []registeredService{
				{ServiceCode: "test-community", Edition: topology.CommunityEdition, Joined: 3},
				{ServiceCode: "test-enterprise", Edition: topology.EnterpriseEdition, Joined: 2},
			},
		},
		{
			name: "unmanaged service codes are reported in input order",
			registered: []registeredService{
				{ServiceCode: "legacy-svc", Edition: topology.EnterpriseEdition, Joined: 4},
				{ServiceCode: "test-community", Edition: topology.CommunityEdition, Joined: 3},
				{ServiceCode: "orphan", Edition: topology.CommunityEdition, Joined: 0},
			},
			want: []listWarning{
				{ServiceCode: "legacy-svc", Reasons: []string{
					"registered in ZooKeeper 192.0.2.11:2181 (edition: enterprise, joined: 4) " +
						"but no deployment record on this machine",
				}},
				{ServiceCode: "orphan", Reasons: []string{
					"registered in ZooKeeper 192.0.2.11:2181 (edition: community, joined: 0) " +
						"but no deployment record on this machine",
				}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := unmanagedWarnings(zkAddress, test.registered, managed)
			if !slices.EqualFunc(got, test.want, func(a, b listWarning) bool {
				return a.ServiceCode == b.ServiceCode && slices.Equal(a.Reasons, b.Reasons)
			}) {
				t.Errorf("warnings = %+v, want %+v", got, test.want)
			}
		})
	}
}
