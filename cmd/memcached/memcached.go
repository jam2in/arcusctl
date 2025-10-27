package memcached

import (
	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcusctl/internal/types"
	"github.com/jam2in/arcusctl/internal/zookeeper"
	"github.com/spf13/cobra"
)

var MemcachedCmd = &cobra.Command{
	Use:   "memcached",
	Short: "Memcached command",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := zookeeper.ContextWithZkConn(cmd.Context(), "", "")
		if err != nil {
			return err
		}
		zkConn := ctx.Value(types.CtxZkConnKey{}).(*zk.Conn)
		if err = zookeeper.InitializeZK(zkConn); err != nil {
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
	MemcachedCmd.AddCommand(addCmd)
	MemcachedCmd.AddCommand(removeCmd)
	MemcachedCmd.AddCommand(configCmd)
	MemcachedCmd.AddCommand(listCmd)
	MemcachedCmd.AddCommand(startCmd)
	MemcachedCmd.AddCommand(stopCmd)
}

func filterServers(allServers []string, targets []string) []string {
	targetSet := make(map[string]struct{})
	for _, t := range targets {
		targetSet[t] = struct{}{}
	}

	result := make([]string, 0)
	for _, s := range allServers {
		if _, ok := targetSet[s]; ok {
			result = append(result, s)
		}
	}
	return result
}
