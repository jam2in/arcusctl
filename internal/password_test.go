package internal

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestValidatePassword(t *testing.T) {
	testcases := []struct {
		name    string
		src     string
		wantErr bool
	}{
		{name: "all four classes", src: "Abcd1234!@#$", wantErr: false},
		{name: "upper lower digit", src: "Abcdefgh1234", wantErr: false},
		{name: "lower digit symbol", src: "abcd1234!@#$", wantErr: false},
		{name: "upper digit symbol", src: "ABCD1234!@#$", wantErr: false},
		{name: "upper lower symbol", src: "Abcdefgh!@#$", wantErr: false},

		{name: "empty", src: "", wantErr: true},
		{name: "eleven characters", src: "Abcd1234!@#", wantErr: true},
		{name: "upper lower only", src: "Abcdefghijkl", wantErr: true},
		{name: "lower digit only", src: "abcdefgh1234", wantErr: true},
		{name: "upper digit only", src: "ABCDEFGH1234", wantErr: true},
		{name: "lower symbol only", src: "abcdefgh!@#$", wantErr: true},

		// 길이는 룬 개수로 센다. 아래 두 값은 15바이트지만 7자다.
		{name: "multibyte shorter than min", src: "비밀번호Aa1", wantErr: true},
		{name: "multibyte shorter than min 2", src: "가나다라Xy9", wantErr: true},
		// 12자를 넘기면 멀티바이트 문자가 섞여 있어도 통과한다.
		{name: "multibyte long enough", src: "비밀번호Abcd1234", wantErr: false},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePassword(tc.src)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validatePassword(%q) = %v, wantErr %v", tc.src, err, tc.wantErr)
			}
		})
	}
}

func TestValidatePasswordErrorMessage(t *testing.T) {
	t.Run("reports rune count, not byte count", func(t *testing.T) {
		const src = "비밀번호Aa1" // 15 bytes, 7 runes
		if got, want := utf8.RuneCountInString(src), 7; got != want {
			t.Fatalf("test fixture changed: %q has %d runes, want %d", src, got, want)
		}

		err := validatePassword(src)
		if err == nil {
			t.Fatal("validatePassword() = nil, want error")
		}
		if !strings.Contains(err.Error(), "is 7 characters") {
			t.Errorf("error should report the rune count: %v", err)
		}
	})

	t.Run("names the missing classes", func(t *testing.T) {
		err := validatePassword("abcdefghijkl")
		if err == nil {
			t.Fatal("validatePassword() = nil, want error")
		}
		for _, want := range []string{"uppercase", "digits", "symbols"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error should name missing class %q: %v", want, err)
			}
		}
	})
}
