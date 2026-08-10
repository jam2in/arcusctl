package ssh

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func withOptions(args ...string) []string {
	return append([]string{
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=10",
	}, args...)
}

func Run(host string, command string) error {
	cmd := exec.Command("ssh", withOptions(host, command)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Copy(localPath string, host string, remotePath string) error {
	dest := fmt.Sprintf("%s:%s", host, remotePath)
	cmd := exec.Command("scp", withOptions(localPath, dest)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func FileExists(host string, remotePath string) (bool, error) {
	cmd := exec.Command("ssh", withOptions(host, fmt.Sprintf("test -f %s", remotePath))...)
	err := cmd.Run()
	if err == nil {
		return true, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}

	return false, err
}
