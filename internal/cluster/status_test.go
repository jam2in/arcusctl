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
		report clusterStatusReport
		want   string
	}{
		{
			name: "no warning section for a healthy cluster",
		},
		{
			name: "cluster warnings without node warnings",
			report: clusterStatusReport{
				Warnings: []string{"ZooKeeper unavailable"},
			},
			want: "  - ZooKeeper unavailable\n",
		},
		{
			name: "node reasons appear on separate lines in supplied order",
			report: clusterStatusReport{
				NodeWarnings: []nodeWarning{
					{Address: "127.0.0.1:33534", Reasons: []string{
						"not responding (connection refused)",
						"not joined in cache_list",
					}},
					{Address: "127.0.0.1:33533", Reasons: []string{
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
			report: clusterStatusReport{
				Warnings: []string{"ZooKeeper unavailable"},
				NodeWarnings: []nodeWarning{
					{Address: "127.0.0.1:33533", Reasons: []string{
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
			if err := printClusterStatus(&out, "test", topology.CommunityEdition, test.report); err != nil {
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
