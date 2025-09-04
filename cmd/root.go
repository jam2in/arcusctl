package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-sasl-passwd/cmd/group"
	"github.com/jam2in/arcus-sasl-passwd/cmd/user"
	"github.com/jam2in/arcus-sasl-passwd/config"
	"github.com/jam2in/arcus-sasl-passwd/internal/zookeeper"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "arcus-acl",
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
}

func init() {
	rootCmd.PersistentFlags().StringP("userName", "u", "", "Administrator user name for a group")
	rootCmd.PersistentFlags().StringP("password", "p", "", "Administrator password for a group")
	rootCmd.MarkFlagsRequiredTogether("userName", "password")

	rootCmd.AddCommand(group.GroupCmd)
	rootCmd.AddCommand(user.UserCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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
