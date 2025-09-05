package acl

import (
	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/cmd/acl/group"
	"github.com/jam2in/arcus-cli/cmd/acl/user"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

var (
	digestUsername string
	digestPassword string
)

var AclCmd = &cobra.Command{
	Use:   "acl",
	Short: "A CLI tool for managing Arcus SASL ACL.",
	Long: "A command-line interface to manage Arcus SASL ACLs stored in ZooKeeper.\n" +
		"This tool provides a set of commands to interact with Arcus's access control list,\n" +
		"allowing you to manage user groups and individual user credentials for SASL authentication.\n" +
		"A typical workflow involves creating a group first and then adding users to it.\n",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := internal.ContextWithZkConn(cmd.Context(), digestUsername, digestPassword)
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
	AclCmd.PersistentFlags().StringVarP(&digestUsername, "username", "u", "", "Administrator user for a group")
	AclCmd.PersistentFlags().StringVarP(&digestPassword, "password", "p", "", "Administrator password for a group")
	AclCmd.MarkFlagsRequiredTogether("username", "password")

	AclCmd.AddCommand(group.GroupCmd)
	AclCmd.AddCommand(user.UserCmd)
}
