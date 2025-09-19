package internal

import (
	"os"
	"os/user"

	"golang.org/x/crypto/ssh"
)

func NewSSHSession(ip string) (*ssh.Session, func(), error) {
	current, err := user.Current()
	if err != nil {
		return nil, nil, err
	}
	username := current.Username

	keyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", ip+":22", sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	close := func() {
		client.Close()
		session.Close()
	}

	return session, close, err
}
