package topology

import "testing"

func TestClusterTopology_GroupCount(t *testing.T) {
	tests := []struct {
		name    string
		servers []CacheServer
		want    int
	}{
		{
			name: "no groups",
			servers: []CacheServer{
				{Address: "10.0.0.1:11211"},
				{Address: "10.0.0.2:11211"},
			},
			want: 0,
		},
		{
			name: "duplicate group names are counted once",
			servers: []CacheServer{
				{Address: "10.0.0.1:11211", Group: &GroupInfo{Name: "g1", Role: "master", Port: 22322}},
				{Address: "10.0.0.2:11211", Group: &GroupInfo{Name: "g1", Role: "slave", Port: 22322}},
			},
			want: 1,
		},
		{
			name: "multiple groups",
			servers: []CacheServer{
				{Address: "10.0.0.1:11211", Group: &GroupInfo{Name: "g1", Role: "master"}},
				{Address: "10.0.0.2:11211", Group: &GroupInfo{Name: "g1", Role: "slave"}},
				{Address: "10.0.0.3:11211", Group: &GroupInfo{Name: "g2", Role: "master"}},
				{Address: "10.0.0.4:11211", Group: &GroupInfo{Name: "g2", Role: "slave"}},
			},
			want: 2,
		},
		{
			name: "multiple groups only masters",
			servers: []CacheServer{
				{Address: "10.0.0.1:11211", Group: &GroupInfo{Name: "g1", Role: "master"}},
				{Address: "10.0.0.2:11211", Group: &GroupInfo{Name: "g2", Role: "master"}},
				{Address: "10.0.0.3:11211", Group: &GroupInfo{Name: "g3", Role: "master"}},
			},
			want: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			topo := &ClusterTopology{Servers: test.servers}

			if got := topo.GroupCount(); got != test.want {
				t.Errorf("GroupCount() = %d, want %d", got, test.want)
			}
		})
	}
}
