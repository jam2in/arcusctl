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

	report := evaluateClusterStatus(
		topo,
		list,
		cacheListErr,
		probes,
	)

	return errors.Join(
		cacheListErr,
		printClusterStatus(os.Stdout, meta.ServiceCode, edition, report),
	)
}

func printClusterStatus(
	out io.Writer,
	serviceCode string,
	edition topology.ClusterEdition,
	report clusterStatusReport,
) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	fmt.Fprintf(
		w,
		"Arcus cluster %q (edition: %s)\n",
		serviceCode,
		edition,
	)

	if report.CacheListKnown {
		fmt.Fprintf(
			w,
			"Nodes: %d/%d healthy\n\n",
			report.Healthy,
			len(report.Rows),
		)
	} else {
		fmt.Fprintf(
			w,
			"Nodes: unknown/%d healthy\n\n",
			len(report.Rows),
		)
	}

	if edition == topology.EnterpriseEdition {
		fmt.Fprint(w, "GROUP\tROLE\t")
	}
	fmt.Fprintln(w, "ADDRESS\tVERSION\tRESPONDING\tZK")

	for _, row := range report.Rows {
		if edition == topology.EnterpriseEdition {
			fmt.Fprintf(w, "%s\t%s\t", row.Group, row.Role)
		}

		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\n",
			row.Address,
			row.Version,
			row.Responding,
			row.ZK,
		)
	}

	if err := w.Flush(); err != nil {
		return err
	}

	if len(report.Warnings) == 0 && len(report.NodeWarnings) == 0 {
		return nil
	}

	fmt.Fprintln(out, "\nWarnings:")

	// cluster-wide warnings
	for _, warning := range report.Warnings {
		fmt.Fprintf(out, "  - %s\n", warning)
	}

	// warnings grouped by server address
	for _, warning := range report.NodeWarnings {
		fmt.Fprintf(out, "  %s:\n", warning.Address)
		for _, reason := range warning.Reasons {
			fmt.Fprintf(out, "    - %s\n", reason)
		}
	}

	return nil
}
