package cluster

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jam2in/arcusctl/internal/topology"
)

func TestPrintClusterStatusWarnings(t *testing.T) {
	tests := []struct {
		name   string
		status clusterStatus
		want   string
	}{
		{
			name: "no warning section for a healthy cluster",
		},
		{
			name: "cluster warnings without node warnings",
			status: clusterStatus{
				ClusterWarnings: []string{"ZooKeeper unavailable"},
			},
			want: "  - ZooKeeper unavailable\n",
		},
		{
			name: "node reasons appear on separate lines in supplied order",
			status: clusterStatus{
				NodeWarnings: []nodeWarning{
					{Address: "127.0.0.1:33534", Messages: []string{
						"not responding (connection refused)",
						"not joined in cache_list",
					}},
					{Address: "127.0.0.1:33533", Messages: []string{
						"not responding (EOF)",
					}},
				},
			},
			want: "  127.0.0.1:33534:\n" +
				"    - not responding (connection refused)\n" +
				"    - not joined in cache_list\n" +
				"  127.0.0.1:33533:\n" +
				"    - not responding (EOF)\n",
		},
		{
			name: "cluster and node warnings share a single section",
			status: clusterStatus{
				ClusterWarnings: []string{"ZooKeeper unavailable"},
				NodeWarnings: []nodeWarning{
					{Address: "127.0.0.1:33533", Messages: []string{
						"not responding (EOF)",
					}},
				},
			},
			want: "  - ZooKeeper unavailable\n" +
				"  127.0.0.1:33533:\n" +
				"    - not responding (EOF)\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := printClusterStatus(&out, "test", topology.CommunityEdition, test.status); err != nil {
				t.Fatal(err)
			}
			_, got, found := strings.Cut(out.String(), "\nWarnings:\n")
			if found != (test.want != "") {
				t.Fatalf("unexpected warning section in output: %q", out.String())
			}
			if got != test.want {
				t.Errorf("warning section = %q, want %q", got, test.want)
			}
		})
	}
}
