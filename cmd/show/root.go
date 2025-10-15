package show

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/jam2in/arcus-cli/internal/aclgroup"
	"github.com/spf13/cobra"
)

var (
	RootCmd = &cobra.Command{
		Use:   "show RESOURCE",
		Short: "TODO",
	}

	aclgroupCmd = &cobra.Command{
		Use:   "aclgroup ZOOKEEPER_ADDR GROUP_NAME",
		Short: "TODO",
		Run: func(cmd *cobra.Command, args []string) {
			addr := args[0]
			group := args[1]

			r, err := aclgroup.GetResource(addr, group)
			if err != nil {
				fmt.Printf("Failed to get aclgroup: %v\n", err)
				os.Exit(1)
			}

			raw, err := yaml.Marshal(r)
			if err != nil {
				fmt.Printf("yaml.Marshal: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("%s\n", raw)
		},
	}
)

func init() {
	RootCmd.AddCommand(aclgroupCmd)
}
