package zookeeper

import (
	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

const (
	zookeeperStartCommandTemplate = "%s/bin/zkServer.sh start"
	zookeeperStopCommandTemplate  = "%s/bin/zkServer.sh stop"
)

var ZookeeperCmd = &cobra.Command{
	Use:   "zookeeper",
	Short: "A CLI tool for zookeeper commands",
	Long: "A command-line interface to manage the ZooKeeper structure for an Arcus cluster.\n" +
		"This includes initializing the required znode directory layout. that Arcus\n" +
		"and controlling the lifecycle of the Zookeeper cluster.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := internal.ContextWithZkConn(cmd.Context(), "", "")
		if err != nil {
			return err
		}
		cmd.SetContext(ctx)

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		zkConn := cmd.Context().Value(internal.CtxZkConnKey{}).(*zk.Conn)
		zkConn.Close()
	},
}

func init() {
	ZookeeperCmd.AddCommand(initCmd)
	ZookeeperCmd.AddCommand(startCmd)
	ZookeeperCmd.AddCommand(statCmd)
	ZookeeperCmd.AddCommand(stopCmd)
}
