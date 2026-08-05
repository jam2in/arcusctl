package cluster

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jam2in/arcusctl/internal/ssh"
	"github.com/jam2in/arcusctl/internal/topology"
)

func Status(serviceCode string) error {
	meta, topo, edition, err := loadCluster(serviceCode)
	if err != nil {
		return err
	}

	registered, err := registeredAddresses(topo, edition)
	if err != nil {
		return fmt.Errorf("read cache_server_mapping from ZooKeeper: %w", err)
	}

	fmt.Printf(
		"Arcus cluster %q (edition: %s, version: %s)\n\n",
		meta.ServiceCode, edition, meta.Version,
	)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if edition == topology.EnterpriseEdition {
		fmt.Fprintln(w, "GROUP\tROLE\tADDRESS\tPROCESS_STATUS\tZK_REGISTERED")
		for _, s := range topo.Servers {
			fmt.Fprintf(
				w, "%s\t%s\t%s\t%s\t%s\n",
				s.Group.Name, s.Group.Role, s.Address,
				processStatus(s, topo.Path, meta.Version),
				yesNo(registered[s.Address]),
			)
		}
	} else {
		fmt.Fprintln(w, "ADDRESS\tPROCESS_STATUS\tZK_REGISTERED")
		for _, s := range topo.Servers {
			fmt.Fprintf(
				w, "%s\t%s\t%s\n",
				s.Address, processStatus(s, topo.Path, meta.Version),
				yesNo(registered[s.Address]),
			)
		}
	}

	return w.Flush()
}

func processStatus(
	server topology.CacheServer,
	topoPath string,
	version string,
) string {
	pidFile := pidFilePath(server.Address, topoPath, version)
	cmd := fmt.Sprintf("pgrep -f %q > /dev/null 2>&1", pidFile)
	if err := ssh.Run(server.Host(), cmd); err == nil {
		return "running"
	}
	return "stopped"
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
