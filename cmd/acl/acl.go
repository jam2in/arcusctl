package acl

import (
	"context"
	"fmt"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/cmd/acl/group"
	"github.com/jam2in/arcus-cli/cmd/acl/user"
	"github.com/jam2in/arcus-cli/config"
	"github.com/jam2in/arcus-cli/internal/zookeeper"
	"github.com/spf13/cobra"
)

var AclCmd = &cobra.Command{
	Use:   "acl",
	Short: "A CLI tool for managing Arcus SASL ACL.",
	Long: "A command-line interface to manage Arcus SASL ACLs stored in ZooKeeper.\n" +
		"This tool provides a set of commands to interact with Arcus's access control list,\n" +
		"allowing you to manage user groups and individual user credentials for SASL authentication.\n" +
		"A typical workflow involves creating a group first and then adding users to it.\n" +
		"To manage groups: arcus-acl group [subcommand]\nTo manage users: arcus-acl user [subcommand]",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		aclUser, _ := cmd.Flags().GetString("userName")
		aclPassword, _ := cmd.Flags().GetString("password")

		zkConn, err := authenticate(aclUser, aclPassword)
		if err != nil {
			return err
		}

		ctx := context.WithValue(cmd.Context(), config.ZkConnKey{}, zkConn)
		cmd.SetContext(ctx)

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		zkConn := cmd.Context().Value(config.ZkConnKey{})
		if zkConn != nil {
			zkConn.(*zk.Conn).Close()
		}
		return nil
	},
}

func init() {
	AclCmd.PersistentFlags().StringP("userName", "u", "", "Administrator user name for a group")
	AclCmd.PersistentFlags().StringP("password", "p", "", "Administrator password for a group")
	AclCmd.MarkFlagsRequiredTogether("userName", "password")

	AclCmd.AddCommand(group.GroupCmd)
	AclCmd.AddCommand(user.UserCmd)
}

func authenticate(aclUser, aclPassword string) (*zk.Conn, error) {
	zkConn, err := zookeeper.NewConnect()
	if err != nil {
		return nil, err
	}

	if aclUser != "" && aclPassword != "" {
		auth := []byte(aclUser + ":" + aclPassword)
		if err := zkConn.AddAuth("digest", auth); err != nil {
			return nil, err
		}
		fmt.Printf("Authenticaed as '%s'.\n", aclUser)
	}

	return zkConn, nil
}
