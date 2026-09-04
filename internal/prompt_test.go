package internal

import (
	"bufio"
	"errors"
	"strings"
	"testing"
)

func TestConfirm(t *testing.T) {
	testcases := []struct {
		name    string
		input   string
		want    bool
		wantErr bool
	}{
		{name: "short yes", input: "y\n", want: true},
		{name: "case insensitive yes", input: " YES \n", want: true},
		{name: "no", input: "n\n", want: false},
		{name: "anything else", input: "sure\n", want: false},
		{name: "yes ending in EOF", input: "yes", want: true},
		{name: "empty EOF", input: "", wantErr: true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var output strings.Builder
			got, err := confirm(bufio.NewReader(strings.NewReader(tc.input)), &output, "Continue? ")
			if (err != nil) != tc.wantErr {
				t.Fatalf("confirm() error = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("confirm() = %v, want %v", got, tc.want)
			}
			if gotPrompt := output.String(); gotPrompt != "Continue? " {
				t.Errorf("prompt = %q, want %q", gotPrompt, "Continue? ")
			}
		})
	}
}

func TestReadInput(t *testing.T) {
	t.Run("trims surrounding whitespace", func(t *testing.T) {
		var output strings.Builder
		got, err := readInput(
			bufio.NewReader(strings.NewReader("  admin  \n")),
			&output,
			"admin name",
		)
		if err != nil {
			t.Fatalf("readInput() error = %v", err)
		}
		if want := "admin"; got != want {
			t.Errorf("readInput() = %q, want %q", got, want)
		}
		if want := "admin name: "; output.String() != want {
			t.Errorf("prompt = %q, want %q", output.String(), want)
		}
	})

	t.Run("rejects empty input", func(t *testing.T) {
		_, err := readInput(
			bufio.NewReader(strings.NewReader("  \n")),
			&strings.Builder{},
			"admin name",
		)
		if err == nil {
			t.Fatal("readInput() error = nil, want error")
		}
	})
}

func TestReadLineReturnsPromptWriteError(t *testing.T) {
	wantErr := errors.New("write failed")
	_, err := readLine(
		bufio.NewReader(strings.NewReader("yes\n")),
		errorWriter{err: wantErr},
		"Continue? ",
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("readLine() error = %v, want wrapped %v", err, wantErr)
	}
}

type errorWriter struct {
	err error
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, w.err
}
