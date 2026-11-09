package cluster

import (
	"testing"

	"github.com/jam2in/arcusctl/internal/topology"
)

func TestParseCacheNode(t *testing.T) {
	tests := []struct {
		name    string
		node    string
		edition topology.ClusterEdition
		want    cacheNode
		wantErr bool
	}{
		{
			name:    "community",
			node:    "192.0.2.21:11211-cache01",
			edition: topology.CommunityEdition,
			want:    cacheNode{Address: "192.0.2.21:11211"},
		},
		{
			name:    "enterprise master",
			node:    "g1^M^192.0.2.21:11211-cache01",
			edition: topology.EnterpriseEdition,
			want: cacheNode{
				Address: "192.0.2.21:11211",
				Group:   "g1",
				Role:    "master",
			},
		},
		{
			name:    "enterprise slave",
			node:    "g1^S^192.0.2.22:11211-cache02",
			edition: topology.EnterpriseEdition,
			want: cacheNode{
				Address: "192.0.2.22:11211",
				Group:   "g1",
				Role:    "slave",
			},
		},
		{
			name:    "invalid enterprise role",
			node:    "g1^X^192.0.2.21:11211-cache01",
			edition: topology.EnterpriseEdition,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseCacheNode(test.node, test.edition)
			if (err != nil) != test.wantErr {
				t.Fatalf("parseCacheNode(%q) error = %v, wantErr %t",
					test.node, err, test.wantErr)
			}
			if err != nil {
				return
			}
			if got != test.want {
				t.Errorf("parseCacheNode(%q) = %+v, want %+v",
					test.node, got, test.want)
			}
		})
	}
}
