package cluster

import (
	"fmt"

	"github.com/jam2in/arcusctl/internal/ssh"
	"github.com/jam2in/arcusctl/internal/topology"
)

const (
	versionTimeout         = 3
	versionCommandTemplate = "echo version | nc -w %d %s %s 2>&1"
)

func runVersion(server topology.CacheServer) (string, error) {

	cmd := fmt.Sprintf(
		versionCommandTemplate,
		versionTimeout,
		ssh.Quote(server.Host()),
		ssh.Quote(listenPort(server.Address)),
	)

	return ssh.RunOutput(server.Host(), cmd)
}
