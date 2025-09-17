package memcached

import (
	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

var MemcachedCmd = &cobra.Command{
	Use:   "memcached",
	Short: "Memcached command",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := internal.ContextWithZkConn(cmd.Context(), "", "")
		if err != nil {
			return err
		}
		zkConn := ctx.Value(internal.CtxZkConnKey{}).(*zk.Conn)
		if err = internal.InitializeZK(zkConn); err != nil {
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
	MemcachedCmd.AddCommand(addCmd)
	MemcachedCmd.AddCommand(removeCmd)
	MemcachedCmd.AddCommand(configCmd)
	MemcachedCmd.AddCommand(connectCmd)
}
