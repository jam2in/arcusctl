package ssh

import "testing"

func TestQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty",
			input: "",
			want:  "''",
		},
		{
			name:  "plain path",
			input: "/opt/zookeeper",
			want:  "'/opt/zookeeper'",
		},
		{
			name:  "spaces",
			input: "/opt/zookeeper data",
			want:  "'/opt/zookeeper data'",
		},
		{
			name:  "single quote",
			input: "/opt/user's data",
			want:  `'/opt/user'\''s data'`,
		},
		{
			name:  "multiple single quotes",
			input: "a'b'c",
			want:  `'a'\''b'\''c'`,
		},
		{
			name:  "only single quote",
			input: "'",
			want:  `''\'''`,
		},
		{
			name:  "double quotes",
			input: `/opt/"data"`,
			want:  `'/opt/"data"'`,
		},
		{
			name:  "shell metacharacters",
			input: "/opt/$HOME/$(whoami);*",
			want:  "'/opt/$HOME/$(whoami);*'",
		},
		{
			name:  "newline",
			input: "/opt/data\nlogs",
			want:  "'/opt/data\nlogs'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Quote(tt.input)
			if got != tt.want {
				t.Errorf("Quote(%q) = %q, want %q",
					tt.input, got, tt.want)
			}
		})
	}
}
