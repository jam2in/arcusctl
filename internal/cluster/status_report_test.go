package cluster

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"syscall"
	"testing"

	"github.com/jam2in/arcusctl/internal/topology"
)

func TestEvaluateClusterStatus(t *testing.T) {
	const (
		first  = "192.0.2.21:11211"
		second = "192.0.2.22:11211"
		third  = "192.0.2.23:11211"
	)
	cacheListErr := errors.New("ZooKeeper unavailable")

	tests := []struct {
		name         string
		servers      []topology.CacheServer
		list         cacheList
		cacheListErr error
		probes       []cacheProbe
		want         clusterStatusReport
	}{
		{
			name:    "healthy nodes preserve topology order and actual versions",
			servers: []topology.CacheServer{{Address: second}, {Address: first}},
			list: cacheList{Nodes: map[string]cacheNode{
				first: {Address: first}, second: {Address: second},
			}},
			probes: []cacheProbe{{Version: "1.17.0"}, {Version: "1.16.1"}},
			want: clusterStatusReport{
				CacheListKnown: true,
				Healthy:        2,
				Rows: []clusterStatusRow{
					{Address: second, Group: "-", Role: "-", Version: "1.17.0", Responding: "yes", ZK: "joined"},
					{Address: first, Group: "-", Role: "-", Version: "1.16.1", Responding: "yes", ZK: "joined"},
				},
			},
		},
		{
			name:    "responding but unjoined node is not healthy",
			servers: []topology.CacheServer{{Address: first}},
			probes:  []cacheProbe{{Version: "1.16.1"}},
			want: clusterStatusReport{
				CacheListKnown: true,
				Rows: []clusterStatusRow{
					{Address: first, Group: "-", Role: "-", Version: "1.16.1", Responding: "yes", ZK: "not joined"},
				},
				NodeWarnings: []nodeWarning{
					{Address: first, Reasons: []string{"not joined in cache_list"}},
				},
			},
		},
		{
			name:    "joined but unresponsive node is not healthy",
			servers: []topology.CacheServer{{Address: first}},
			list: cacheList{Nodes: map[string]cacheNode{
				first: {Address: first},
			}},
			probes: []cacheProbe{{Err: io.EOF}},
			want: clusterStatusReport{
				CacheListKnown: true,
				Rows: []clusterStatusRow{
					{Address: first, Group: "-", Role: "-", Version: "-", Responding: "no", ZK: "joined"},
				},
				NodeWarnings: []nodeWarning{
					{Address: first, Reasons: []string{"not responding (EOF)"}},
				},
			},
		},
		{
			name:    "unresponsive and unjoined node remains in the table",
			servers: []topology.CacheServer{{Address: first}},
			probes:  []cacheProbe{{Err: io.EOF}},
			want: clusterStatusReport{
				CacheListKnown: true,
				Rows: []clusterStatusRow{
					{Address: first, Group: "-", Role: "-", Version: "-", Responding: "no", ZK: "not joined"},
				},
				NodeWarnings: []nodeWarning{
					{Address: first, Reasons: []string{"not responding (EOF)", "not joined in cache_list"}},
				},
			},
		},
		{
			name:    "missing service znode is known absence rather than a query failure",
			servers: []topology.CacheServer{{Address: first}},
			list:    cacheList{Missing: true},
			probes:  []cacheProbe{{Version: "1.16.1"}},
			want: clusterStatusReport{
				CacheListKnown: true,
				Rows: []clusterStatusRow{
					{Address: first, Group: "-", Role: "-", Version: "1.16.1", Responding: "yes", ZK: "not joined"},
				},
				Warnings: []string{
					`service code "test" is not registered in cache_list`,
				},
				NodeWarnings: []nodeWarning{
					{Address: first, Reasons: []string{"not joined in cache_list"}},
				},
			},
		},
		{
			name:         "cache list query failure preserves community probe results",
			servers:      []topology.CacheServer{{Address: first}},
			cacheListErr: cacheListErr,
			probes:       []cacheProbe{{Version: "1.16.1"}},
			want: clusterStatusReport{
				Rows: []clusterStatusRow{
					{Address: first, Group: "-", Role: "-", Version: "1.16.1", Responding: "yes", ZK: "unknown"},
				},
			},
		},
		{
			name: "cache list query failure does not use partial enterprise membership information",
			servers: []topology.CacheServer{
				{Address: first, Group: &topology.GroupInfo{Name: "g1", Role: "master"}},
			},
			list: cacheList{Nodes: map[string]cacheNode{
				first: {Address: first, Group: "g1", Role: "slave"},
			}},
			cacheListErr: cacheListErr,
			probes:       []cacheProbe{{Version: "0.12.1-E"}},
			want: clusterStatusReport{
				Rows: []clusterStatusRow{
					{Address: first, Group: "unknown", Role: "unknown", Version: "0.12.1-E", Responding: "yes", ZK: "unknown"},
				},
			},
		},
		{
			name: "unjoined enterprise node shows no membership and reports topology group/role in warnings",
			servers: []topology.CacheServer{
				{Address: first, Group: &topology.GroupInfo{Name: "g1", Role: "slave"}},
			},
			list:   cacheList{Nodes: map[string]cacheNode{}},
			probes: []cacheProbe{{Err: io.EOF}},
			want: clusterStatusReport{
				CacheListKnown: true,
				Rows: []clusterStatusRow{
					{Address: first, Group: "-", Role: "-", Version: "-", Responding: "no", ZK: "not joined"},
				},
				NodeWarnings: []nodeWarning{
					{Address: first, Reasons: []string{"not responding (EOF)", "not joined in cache_list (topology=g1/slave)"}},
				},
			},
		},
		{
			name:    "extra nodes only produce sorted warnings and do not affect rows or healthy count",
			servers: []topology.CacheServer{{Address: first}},
			list: cacheList{Nodes: map[string]cacheNode{
				third: {Address: third}, first: {Address: first}, second: {Address: second},
			}},
			probes: []cacheProbe{{Version: "1.16.1"}},
			want: clusterStatusReport{
				CacheListKnown: true,
				Healthy:        1,
				Rows: []clusterStatusRow{
					{Address: first, Group: "-", Role: "-", Version: "1.16.1", Responding: "yes", ZK: "joined"},
				},
				NodeWarnings: []nodeWarning{
					{Address: second, Reasons: []string{"joined in cache_list but not in the stored topology"}},
					{Address: third, Reasons: []string{"joined in cache_list but not in the stored topology"}},
				},
			},
		},
		{
			name: "switchover displays actual roles and keeps both nodes healthy",
			servers: []topology.CacheServer{
				{Address: first, Group: &topology.GroupInfo{Name: "g2", Role: "master"}},
				{Address: second, Group: &topology.GroupInfo{Name: "g2", Role: "slave"}},
			},
			list: cacheList{Nodes: map[string]cacheNode{
				first:  {Address: first, Group: "g2", Role: "slave"},
				second: {Address: second, Group: "g2", Role: "master"},
			}},
			probes: []cacheProbe{{Version: "0.12.1-E"}, {Version: "0.12.1-E"}},
			want: clusterStatusReport{
				CacheListKnown: true,
				Healthy:        2,
				Rows: []clusterStatusRow{
					{Address: first, Group: "g2", Role: "slave", Version: "0.12.1-E", Responding: "yes", ZK: "joined"},
					{Address: second, Group: "g2", Role: "master", Version: "0.12.1-E", Responding: "yes", ZK: "joined"},
				},
				NodeWarnings: []nodeWarning{
					{Address: first, Reasons: []string{"group/role differs (topology=g2/master, actual=g2/slave)"}},
					{Address: second, Reasons: []string{"group/role differs (topology=g2/slave, actual=g2/master)"}},
				},
			},
		},
		{
			name:    "declared node warnings follow topology order before extra nodes",
			servers: []topology.CacheServer{{Address: second}, {Address: first}},
			list: cacheList{Nodes: map[string]cacheNode{
				third: {Address: third}, first: {Address: first}, second: {Address: second},
			}},
			probes: []cacheProbe{{Err: io.EOF}, {Err: io.EOF}},
			want: clusterStatusReport{
				CacheListKnown: true,
				Rows: []clusterStatusRow{
					{Address: second, Group: "-", Role: "-", Version: "-", Responding: "no", ZK: "joined"},
					{Address: first, Group: "-", Role: "-", Version: "-", Responding: "no", ZK: "joined"},
				},
				NodeWarnings: []nodeWarning{
					{Address: second, Reasons: []string{"not responding (EOF)"}},
					{Address: first, Reasons: []string{"not responding (EOF)"}},
					{Address: third, Reasons: []string{"joined in cache_list but not in the stored topology"}},
				},
			},
		},
		{
			name: "extra enterprise node keeps group and role in its warning",
			servers: []topology.CacheServer{
				{Address: first, Group: &topology.GroupInfo{Name: "g1", Role: "master"}},
			},
			list: cacheList{Nodes: map[string]cacheNode{
				first:  {Address: first, Group: "g1", Role: "master"},
				second: {Address: second, Group: "g2", Role: "slave"},
			}},
			probes: []cacheProbe{{Version: "0.12.1-E"}},
			want: clusterStatusReport{
				CacheListKnown: true,
				Healthy:        1,
				Rows: []clusterStatusRow{
					{Address: first, Group: "g1", Role: "master", Version: "0.12.1-E", Responding: "yes", ZK: "joined"},
				},
				NodeWarnings: []nodeWarning{
					{Address: second, Reasons: []string{"joined in cache_list but not in the stored topology (group=g2, role=slave)"}},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			topo := &topology.ClusterTopology{ServiceCode: "test", Servers: test.servers}
			got := evaluateClusterStatus(topo, test.list, test.cacheListErr, test.probes)
			if got.CacheListKnown != test.want.CacheListKnown {
				t.Errorf("CacheListKnown = %t, want %t", got.CacheListKnown, test.want.CacheListKnown)
			}
			if got.Healthy != test.want.Healthy {
				t.Errorf("Healthy = %d, want %d", got.Healthy, test.want.Healthy)
			}
			if !slices.Equal(got.Rows, test.want.Rows) {
				t.Errorf("Rows = %+v, want %+v", got.Rows, test.want.Rows)
			}
			if !slices.Equal(got.Warnings, test.want.Warnings) {
				t.Errorf("Warnings = %q, want %q", got.Warnings, test.want.Warnings)
			}
			if !slices.EqualFunc(got.NodeWarnings, test.want.NodeWarnings, func(a, b nodeWarning) bool {
				return a.Address == b.Address && slices.Equal(a.Reasons, b.Reasons)
			}) {
				t.Errorf("NodeWarnings = %+v, want %+v", got.NodeWarnings, test.want.NodeWarnings)
			}
		})
	}
}

func TestStatusServerWarnings(t *testing.T) {
	const address = "127.0.0.1:11211"
	connectionErr := fmt.Errorf("connect to %s: %w", address, &net.OpError{
		Op:   "dial",
		Net:  "tcp",
		Addr: &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 11211},
		Err:  &os.SyscallError{Syscall: "connect", Err: syscall.ECONNREFUSED},
	})

	tests := []struct {
		name     string
		group    *topology.GroupInfo
		row      clusterStatusRow
		probeErr error
		want     []string
	}{
		{
			name: "healthy server has no warnings",
			row:  clusterStatusRow{ZK: "joined"},
		},
		{
			name:     "unresponsive and unjoined server retains separate reasons",
			row:      clusterStatusRow{ZK: "not joined"},
			probeErr: connectionErr,
			want: []string{
				"not responding (connection refused)",
				"not joined in cache_list",
			},
		},
		{
			name: "responding but unjoined server",
			row:  clusterStatusRow{ZK: "not joined"},
			want: []string{
				"not joined in cache_list",
			},
		},
		{
			name:  "unjoined enterprise server retains topology group and role",
			group: &topology.GroupInfo{Name: "g1", Role: "slave"},
			row:   clusterStatusRow{ZK: "not joined"},
			want: []string{
				"not joined in cache_list (topology=g1/slave)",
			},
		},
		{
			name:     "joined but unresponsive server",
			row:      clusterStatusRow{ZK: "joined"},
			probeErr: connectionErr,
			want: []string{
				"not responding (connection refused)",
			},
		},
		{
			name:  "role mismatch retains topology and actual roles",
			group: &topology.GroupInfo{Name: "g1", Role: "master"},
			row:   clusterStatusRow{ZK: "joined", Group: "g1", Role: "slave"},
			want: []string{
				"group/role differs (topology=g1/master, actual=g1/slave)",
			},
		},
		{
			name:     "probe failure and role mismatch retain separate reasons",
			group:    &topology.GroupInfo{Name: "g1", Role: "master"},
			row:      clusterStatusRow{ZK: "joined", Group: "g1", Role: "slave"},
			probeErr: connectionErr,
			want: []string{
				"not responding (connection refused)",
				"group/role differs (topology=g1/master, actual=g1/slave)",
			},
		},
		{
			name:     "unknown cache list is not reported as unjoined",
			row:      clusterStatusRow{ZK: "unknown"},
			probeErr: connectionErr,
			want: []string{
				"not responding (connection refused)",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := topology.CacheServer{Address: address, Group: test.group}
			got := statusServerWarnings(server, test.row, cacheProbe{Err: test.probeErr})
			if !slices.Equal(got, test.want) {
				t.Errorf("statusServerWarnings() = %q, want %q", got, test.want)
			}
		})
	}
}
