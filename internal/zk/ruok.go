package zk

import (
	"fmt"

	"github.com/jam2in/arcusctl/internal/ssh"
	"github.com/jam2in/arcusctl/internal/topology"
)

const (
	ruokTimeout         = 3
	ruokCommandTemplate = "echo ruok | nc -v -w %d %s %s 2>&1"
)

func runRuok(server topology.ZKServer) (string, error) {
	host, clientPort, _, _ := server.ParseAddress()

	cmd := fmt.Sprintf(
		ruokCommandTemplate,
		ruokTimeout,
		ssh.Quote(host),
		ssh.Quote(clientPort),
	)

	return ssh.RunOutput(server.Host(), cmd)
}
