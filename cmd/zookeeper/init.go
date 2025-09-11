package zookeeper

import (
	"fmt"
	"os"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

var arcusBasicPaths = []string{
	"/arcus",
	"/arcus/cache_list",
	"/arcus/cache_server_mapping",
	"/arcus/client_list",
	"/arcus_repl",
	"/arcus_repl/cache_list",
	"/arcus_repl/cache_server_mapping",
	"/arcus_repl/client_list",
	"/arcus_repl/group_list",
	"/arcus_repl/cloud_stat",
	"/arcus_acl",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the basic Arcus znode structure in Zookeeper.",
	Run: func(cmd *cobra.Command, args []string) {
		zkConn := cmd.Context().Value(internal.CtxZkConnKey{}).(*zk.Conn)

		ops := make([]interface{}, 0)
		for _, p := range arcusBasicPaths {
			exist, _, err := zkConn.Exists(p)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			if !exist {
				ops = append(ops, &zk.CreateRequest{
					Path:  p,
					Data:  []byte{},
					Acl:   zk.WorldACL(zk.PermAll),
					Flags: 0,
				})
			}
		}

		if len(ops) > 0 {
			if _, err := zkConn.Multi(ops...); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("Already initialize the basic Arcus znode structure.\n")
		}
	},
}
