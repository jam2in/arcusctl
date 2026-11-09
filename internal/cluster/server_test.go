package cluster

import "testing"

func TestParseCacheVersion(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     string
		wantErr  bool
	}{
		{name: "community", response: "VERSION 1.16.1\r\n", want: "1.16.1"},
		{name: "enterprise suffix", response: "VERSION 0.12.1-E\r\n", want: "0.12.1-E"},
		{name: "LF line ending", response: "VERSION 1.16.1\n", want: "1.16.1"},
		{name: "empty response", response: "", wantErr: true},
		{name: "blank response", response: "\r\n", wantErr: true},
		{name: "protocol error", response: "ERROR\r\n", wantErr: true},
		{name: "wrong prefix", response: "STAT version 1.16.1\r\n", wantErr: true},
		{name: "missing version", response: "VERSION \r\n", wantErr: true},
		{name: "space in version", response: "VERSION 1.16.1 extra\r\n", wantErr: true},
		{name: "tab in version", response: "VERSION 1.16.1\textra\r\n", wantErr: true},
		{name: "multiple lines", response: "VERSION 1.16.1\r\nERROR\r\n", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseCacheVersion(test.response)
			if (err != nil) != test.wantErr {
				t.Fatalf("parseCacheVersion(%q) error = %v, wantErr %t", test.response, err, test.wantErr)
			}
			if got != test.want {
				t.Errorf("parseCacheVersion(%q) = %q, want %q", test.response, got, test.want)
			}
		})
	}
}
