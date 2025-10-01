package ssh

import (
	"os"
	"os/user"

	"golang.org/x/crypto/ssh"
)

func NewSSHClient(ip string) (*ssh.Client, error) {
	current, err := user.Current()
	if err != nil {
		return nil, err
	}
	username := current.Username

	keyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return ssh.Dial("tcp", ip+":22", sshConfig)
}

func RunSSHCommand(ip, command string) error {
	client, err := NewSSHClient(ip)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	return session.Run(command)
}
