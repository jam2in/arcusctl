package cluster

import (
	"fmt"
	"io"
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

	rows, warnings, zkAddresses := listManagedClusters(serviceCodes)
	warnings = append(warnings, listUnmanagedServices(zkAddresses, serviceCodes)...)

	return printClusterList(os.Stdout, rows, warnings)
}

func listManagedClusters(serviceCodes []string) (
	rows []clusterListRow,
	warnings []listWarning,
	zkAddresses []string,
) {
	seen := make(map[string]bool)

	for _, serviceCode := range serviceCodes {
		row, reasons, zkAddress := listManagedCluster(serviceCode)

		rows = append(rows, row)
		if len(reasons) > 0 {
			warnings = append(warnings, listWarning{
				ServiceCode: serviceCode,
				Reasons:     reasons,
			})
		}

		if zkAddress != "" && !seen[zkAddress] {
			seen[zkAddress] = true
			zkAddresses = append(zkAddresses, zkAddress)
		}
	}

	return rows, warnings, zkAddresses
}

func listManagedCluster(
	serviceCode string,
) (clusterListRow, []string, string) {
	meta, topo, edition, err := loadCluster(serviceCode)
	if err != nil {
		return errorListRow(serviceCode), []string{
			fmt.Sprintf(
				"cannot load deployment record; check ~/.arcusctl/clusters/arcus/%s: %v",
				serviceCode, err,
			)}, ""
	}

	list, cacheListErr := readCacheList(topo, edition)
	row, reasons := buildListRow(meta, topo, edition, list, cacheListErr)

	return row, reasons, topo.ZooKeeper
}

func listUnmanagedServices(
	zkAddresses []string,
	serviceCodes []string,
) []listWarning {
	managed := make(map[string]bool, len(serviceCodes))
	for _, serviceCode := range serviceCodes {
		managed[serviceCode] = true
	}

	var warnings []listWarning
	for _, zkAddress := range zkAddresses {
		registered, err := readRegisteredServices(zkAddress)
		if err != nil {
			// The managed clusters read the same ZooKeeper,
			// so a connectivity failure is already reported in their warnings.
			continue
		}

		warnings = append(warnings, unmanagedWarnings(zkAddress, registered, managed)...)
	}

	return warnings
}

func errorListRow(serviceCode string) clusterListRow {
	return clusterListRow{
		ServiceCode: serviceCode,
		Edition:     "<error>",
		Joined:      "-",
		Topology:    "-",
		Groups:      "-",
		DeployedAt:  "-",
	}
}

func printClusterList(
	out io.Writer,
	rows []clusterListRow,
	warnings []listWarning,
) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SERVICECODE\tEDITION\tJOINED\tTOPOLOGY\tGROUPS\tDEPLOYED_AT")

	for _, row := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			row.ServiceCode,
			row.Edition,
			row.Joined,
			row.Topology,
			row.Groups,
			row.DeployedAt,
		)
	}

	if err := w.Flush(); err != nil {
		return err
	}

	if len(warnings) == 0 {
		return nil
	}

	fmt.Fprintln(out, "\nWarnings:")
	for _, warning := range warnings {
		fmt.Fprintf(out, "  %s:\n", warning.ServiceCode)
		for _, reason := range warning.Reasons {
			fmt.Fprintf(out, "    - %s\n", reason)
		}
	}

	return nil
}
