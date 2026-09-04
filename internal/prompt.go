package internal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

var stdinReader = bufio.NewReader(os.Stdin)

func ReadInput(prompt string) (string, error) {
	return readInput(stdinReader, os.Stdout, prompt)
}

func readInput(reader *bufio.Reader, writer io.Writer, prompt string) (string, error) {
	input, err := readLine(reader, writer, prompt+": ")
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}
	if input == "" {
		return "", errors.New("read input: value must not be empty")
	}
	return input, nil
}

func ReadPassword(prompt string) (string, error) {
	if _, err := fmt.Fprint(os.Stdout, prompt+": "); err != nil {
		return "", fmt.Errorf("write password prompt: %w", err)
	}

	raw, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	if _, err := fmt.Fprintln(os.Stdout); err != nil {
		return "", fmt.Errorf("write password prompt: %w", err)
	}

	password := string(raw)
	if err := validatePassword(password); err != nil {
		return "", fmt.Errorf("invalid password: %w", err)
	}
	return password, nil
}

func Confirm(prompt string) (bool, error) {
	return confirm(stdinReader, os.Stdout, prompt)
}

func confirm(reader *bufio.Reader, writer io.Writer, prompt string) (bool, error) {
	input, err := readLine(reader, writer, prompt)
	if err != nil {
		return false, fmt.Errorf("read confirmation: %w", err)
	}
	return strings.EqualFold(input, "y") || strings.EqualFold(input, "yes"), nil
}

func readLine(reader *bufio.Reader, writer io.Writer, prompt string) (string, error) {
	if _, err := fmt.Fprint(writer, prompt); err != nil {
		return "", fmt.Errorf("write prompt: %w", err)
	}

	line, err := reader.ReadString('\n')
	if err != nil && (!errors.Is(err, io.EOF) || len(line) == 0) {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
