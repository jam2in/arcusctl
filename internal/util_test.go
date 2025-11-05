package internal

import "testing"

func TestIsValidPassword(t *testing.T) {
	testcases := []struct {
		src  string
		want bool
	}{
		{src: "Abcd1234!@#$", want: true},
		{src: "Abcdefgh1234", want: true},
		{src: "abcd1234!@#$", want: true},
		{src: "ABCD1234!@#$", want: true},
		{src: "Abcdefgh!@#$", want: true},
		{src: "", want: false},
		{src: "Abcd1234!@#", want: false},
		{src: "Abcdefghijkl", want: false},
		{src: "abcdefgh1234", want: false},
		{src: "ABCDEFGH1234", want: false},
		{src: "abcdefgh!@#$", want: false},
	}

	for _, tc := range testcases {
		got := isValidPassword(tc.src)
		if got != tc.want {
			t.Errorf("isValidPassword(%s) = %v, want %v", tc.src, got, tc.want)
		}
	}
}
