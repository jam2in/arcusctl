package zk

import (
	"errors"
	"testing"
)

func TestParseMntr(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Mntr
		wantErr bool
	}{
		{
			name: "leader",
			input: "zk_version\t3.5.9-83df9301aa5c2a5d284a9940177808c01bc35cef, built on 01/06/2021 19:49 GMT\n" +
				"zk_server_state\tleader\n" +
				"zk_num_alive_connections\t5\n" +
				"zk_synced_followers\t1\n" +
				"zk_avg_latency\t0\n",
			want: Mntr{
				Version:          "3.5.9",
				ServerState:      "leader",
				AliveConnections: 5,
				SyncedFollowers:  1,
			},
		},
		{
			name: "follower without synced followers",
			input: "zk_version\t3.5.9-83df9301aa5c2a5d284a9940177808c01bc35cef, built on 01/06/2021 19:49 GMT\n" +
				"zk_server_state\tfollower\n" +
				"zk_num_alive_connections\t4\n" +
				"zk_avg_latency\t0\n",
			want: Mntr{
				Version:          "3.5.9",
				ServerState:      "follower",
				AliveConnections: 4,
				SyncedFollowers:  -1,
			},
		},
		{
			name: "leader with zero synced followers",
			input: "zk_version\t3.5.9\n" +
				"zk_server_state\tleader\n" +
				"zk_num_alive_connections\t1\n" +
				"zk_synced_followers\t0\n",
			want: Mntr{
				Version:          "3.5.9",
				ServerState:      "leader",
				AliveConnections: 1,
				SyncedFollowers:  0,
			},
		},
		{
			name:    "empty response",
			input:   "",
			wantErr: true,
		},
		{
			name:    "missing version",
			input:   "zk_server_state\tleader\n",
			wantErr: true,
		},
		{
			name:    "missing state",
			input:   "zk_version\t3.5.9\n",
			wantErr: true,
		},
		{
			name: "invalid connections",
			input: "zk_version\t3.5.9\n" +
				"zk_server_state\tleader\n" +
				"zk_num_alive_connections\tinvalid\n",
			wantErr: true,
		},
		{
			name: "invalid synced followers",
			input: "zk_version\t3.5.9\n" +
				"zk_server_state\tleader\n" +
				"zk_synced_followers\tinvalid\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMntr(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseMntr() error: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseMntr() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseMntrNotWhitelisted(t *testing.T) {
	_, err := parseMntr(
		"mntr is not executed because it is not in the whitelist.\n",
	)
	if !errors.Is(err, errMntrNotWhitelisted) {
		t.Fatalf("got error %v, want %v", err, errMntrNotWhitelisted)
	}
}

func TestShortVersion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain",
			input: "3.5.9",
			want:  "3.5.9",
		},
		{
			name:  "build suffix",
			input: "3.5.9-83df9301aa5c2a5d284a9940177808c01bc35cef, built on 01/06/2021 19:49 GMT",
			want:  "3.5.9",
		},
		{
			name:  "comma suffix",
			input: "3.5.9, built on 01/06/2021 19:49 GMT",
			want:  "3.5.9",
		},
		{
			name:  "whitespace",
			input: "  3.5.9  ",
			want:  "3.5.9",
		},
		{
			name:  "empty",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortVersion(tt.input)
			if got != tt.want {
				t.Errorf("shortVersion(%q) = %q, want %q",
					tt.input, got, tt.want)
			}
		})
	}
}

func TestParseCount(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{name: "zero", input: "0", want: 0},
		{name: "integer", input: "10", want: 10},
		{name: "decimal notation", input: "3.0", wantErr: true},
		{name: "fraction", input: "3.7", wantErr: true},
		{name: "invalid", input: "abc", wantErr: true},
		{name: "empty", input: "", wantErr: true},
		{name: "overflow", input: "999999999999999999999999999", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCount("zk_num_alive_connections", tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseCount() error: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseCount(%q) = %d, want %d",
					tt.input, got, tt.want)
			}
		})
	}
}
