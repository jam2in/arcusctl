package cmd

import (
	"fmt"
	"os"

	"github.com/jam2in/arcus-sasl-passwd/cmd/group"
	"github.com/jam2in/arcus-sasl-passwd/cmd/user"
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
}

func init() {
	rootCmd.AddCommand(group.GroupCmd)
	rootCmd.AddCommand(user.UserCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
