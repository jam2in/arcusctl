package zookeeper

import (
	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal/types"
	"github.com/jam2in/arcus-cli/internal/zookeeper"
	"github.com/spf13/cobra"
)

var ZookeeperCmd = &cobra.Command{
	Use:   "zookeeper",
	Short: "A CLI tool for zookeeper commands",
	Long: "A command-line interface to manage the ZooKeeper structure for an Arcus cluster.\n" +
		"This includes initializing the required znode directory layout. that Arcus\n" +
		"and controlling the lifecycle of the Zookeeper cluster.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := zookeeper.ContextWithZkConn(cmd.Context(), "", "")
		if err != nil {
			return err
		}
		cmd.SetContext(ctx)

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		zkConn := cmd.Context().Value(types.CtxZkConnKey{}).(*zk.Conn)
		zkConn.Close()
	},
}

func init() {
	ZookeeperCmd.AddCommand(initCmd)
}
