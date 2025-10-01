package user

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal/acl"
	"github.com/jam2in/arcus-cli/internal/types"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var roles = map[string]struct{}{
	"kv":    {},
	"list":  {},
	"set":   {},
	"map":   {},
	"btree": {},
	"attr":  {},
	"scan":  {},
	"flush": {},
	"admin": {},
}

var addCmd = &cobra.Command{
	Use:   "add <groupName> <userName[:password]:role>",
	Short: "Add a new user to an ACL group.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		userArgs := args[1:]
		zkConn := cmd.Context().Value(types.CtxZkConnKey{}).(*zk.Conn)
		zkAcl := cmd.Context().Value(types.CtxZkAclKey{}).([]zk.ACL)

		var err error
		for _, arg := range userArgs {
			err = addUserRequest(zkConn, zkAcl, groupName, arg)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	},
}

func readPassword() string {
	fmt.Print("Enter Password: ")
	rawPassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}
	fmt.Println()

	fmt.Print("Repeat Password: ")
	repeatPassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}
	fmt.Println()

	if !bytes.Equal(rawPassword, repeatPassword) {
		panic("passwords do not match")
	}
	return string(rawPassword)
}

func validateRoles(role string) error {
	if role == "" {
		return fmt.Errorf("role cannot be empty")
	}

	seenRoles := make(map[string]struct{})
	for _, r := range strings.Split(role, ",") {
		if _, ok := roles[r]; !ok {
			return fmt.Errorf("invalid role found: %s", r)
		}
		if _, seen := seenRoles[r]; seen {
			return fmt.Errorf("duplicate role found: %s", r)
		}
		seenRoles[r] = struct{}{}
	}
	return nil
}

func addUserRequest(zkConn *zk.Conn, zkAcl []zk.ACL, group, arg string) error {
	tokens := strings.Split(arg, ":")
	var user, password, role string
	switch len(tokens) {
	case 2:
		user = tokens[0]
		password = readPassword()
		role = tokens[1]
	case 3:
		user = tokens[0]
		password = tokens[1]
		role = tokens[2]
	default:
		return fmt.Errorf("invalid argument format: %s", arg)
	}

	if user == "" || password == "" {
		return fmt.Errorf("user & password cannot be empty: %s", arg)
	} else if err := validateRoles(role); err != nil {
		return err
	}

	return acl.AddUser(zkConn, zkAcl, group, user, password, role)
}
