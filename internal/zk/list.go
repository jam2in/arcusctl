package zk

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jam2in/arcusctl/internal/store"
)

func List() error {
	names, err := store.ListZK()
	if err != nil {
		return fmt.Errorf("list ZooKeeper ensembles: %w", err)
	}

	if len(names) == 0 {
		fmt.Println("No ZooKeeper ensemble found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tDEPLOYED AT")
	for _, name := range names {
		meta, err := store.LoadZKMeta(name)
		if err != nil {
			fmt.Fprintf(w, "%s\t<error>\t%v\n", name, err)
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", meta.Name, meta.Version, meta.DeployedAt.Format("2006-01-02 15:04:05"))
	}
	return w.Flush()
}
