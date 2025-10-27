package zookeeper

import (
	"fmt"
	"os"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcusctl/internal/types"
	"github.com/jam2in/arcusctl/internal/zookeeper"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the basic Arcus znode structure in Zookeeper.",
	Run: func(cmd *cobra.Command, args []string) {
		zkConn := cmd.Context().Value(types.CtxZkConnKey{}).(*zk.Conn)
		err := zookeeper.InitializeZK(zkConn)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}
