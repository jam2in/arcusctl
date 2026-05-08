package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	gossh "golang.org/x/crypto/ssh"
)

func currentUser() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	return u.Username, nil
}

func newClient(host string) (*gossh.Client, error) {
	username, err := currentUser()
	if err != nil {
		return nil, err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	keyPath := filepath.Join(homeDir, ".ssh", "id_rsa")
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	signer, err := gossh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	config := &gossh.ClientConfig{
		User: username,
		Auth: []gossh.AuthMethod{
			gossh.PublicKeys(signer),
		},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	}

	return gossh.Dial("tcp", host+":22", config)
}

func Run(host string, command string) error {
	client, err := newClient(host)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer func() { _ = session.Close() }()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	return session.Run(command)
}

func Copy(localPath string, host string, remotePath string) error {
	username, err := currentUser()
	if err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dest := fmt.Sprintf("%s@%s:%s", username, host, remotePath)
	cmd := exec.Command("scp",
		"-i", filepath.Join(homeDir, ".ssh", "id_rsa"),
		"-o", "StrictHostKeyChecking=no",
		localPath,
		dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
