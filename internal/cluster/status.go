package cluster

import (
	"errors"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/jam2in/arcusctl/internal/topology"
)

func Status(serviceCode string) error {
	meta, topo, edition, err := loadCluster(serviceCode)
	if err != nil {
		return err
	}

	list, cacheListErr := readCacheList(topo, edition)
	probes := probeCacheServers(topo.Servers)

	clusterStatus := checkCluster(
		topo,
		list,
		cacheListErr,
		probes,
	)

	return errors.Join(
		cacheListErr,
		printClusterStatus(os.Stdout, meta.ServiceCode, edition, clusterStatus),
	)
}

func printClusterStatus(
	out io.Writer,
	serviceCode string,
	edition topology.ClusterEdition,
	status clusterStatus,
) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	fmt.Fprintf(
		w,
		"Arcus cluster %q (edition: %s)\n",
		serviceCode,
		edition,
	)

	if status.CacheListFound {
		fmt.Fprintf(
			w,
			"Nodes: %d/%d healthy\n\n",
			status.Healthy,
			len(status.Nodes),
		)
	} else {
		fmt.Fprintf(
			w,
			"Nodes: unknown/%d healthy\n\n",
			len(status.Nodes),
		)
	}

	if edition == topology.EnterpriseEdition {
		fmt.Fprint(w, "GROUP\tROLE\t")
	}
	fmt.Fprintln(w, "ADDRESS\tVERSION\tZK")

	for _, node := range status.Nodes {
		if edition == topology.EnterpriseEdition {
			fmt.Fprintf(w, "%s\t%s\t", node.Group, node.Role)
		}

		fmt.Fprintf(
			w,
			"%s\t%s\t%s\n",
			node.Address,
			node.Version,
			node.ZK,
		)
	}

	if err := w.Flush(); err != nil {
		return err
	}

	if len(status.ClusterWarnings) == 0 && len(status.NodeWarnings) == 0 {
		return nil
	}

	fmt.Fprintln(out, "\nWarnings:")

	// cluster-wide warnings
	for _, warning := range status.ClusterWarnings {
		fmt.Fprintf(out, "  - %s\n", warning)
	}

	// warnings grouped by server address
	for _, warning := range status.NodeWarnings {
		fmt.Fprintf(out, "  %s:\n", warning.Address)
		for _, message := range warning.Messages {
			fmt.Fprintf(out, "    - %s\n", message)
		}
	}

	return nil
}
