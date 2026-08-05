package cluster

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jam2in/arcusctl/internal/store"
)

func List() error {
	serviceCodes, err := store.ListCluster()
	if err != nil {
		return fmt.Errorf("list Arcus clusters: %w", err)
	}

	if len(serviceCodes) == 0 {
		fmt.Println("No Arcus cluster found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SERVICECODE\tVERSION\tEDITION\tNODES\tDEPLOYED_AT")

	for _, serviceCode := range serviceCodes {
		meta, err := store.LoadClusterMeta(serviceCode)
		if err != nil {
			fmt.Fprintf(w, "%s\t<error>\t\t\t%v\n", serviceCode, err)
			continue
		}

		edition := "<unknown>"
		servers := 0
		if topo, err := store.LoadClusterTopology(serviceCode); err == nil {
			servers = len(topo.Servers)
			if e, err := topo.Edition(); err == nil {
				edition = string(e)
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
			serviceCode, meta.Version, edition, servers,
			meta.DeployedAt.Format("2006-01-02 15:04:05"),
		)
	}

	return w.Flush()
}
