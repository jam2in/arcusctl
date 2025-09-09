package user

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/jam2in/arcus-cli/internal/scram"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var addCmd = &cobra.Command{
	Use:   "add <groupName> <userName[:password]:role>",
	Short: "Add a new user to an ACL group.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		userArgs := args[1:]

		zkConn := cmd.Context().Value(internal.CtxZkConnKey{}).(*zk.Conn)
		acl := cmd.Context().Value(internal.CtxZkAclKey{}).([]zk.ACL)

		requests := make([]any, 0, 2*len(userArgs))
		var err error
		for _, arg := range userArgs {
			requests, err = appendRequests(requests, groupName, arg, acl)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}

		if _, err := zkConn.Multi(requests...); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
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

func appendRequests(requests []any, group, arg string, acl []zk.ACL) ([]any, error) {
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
		return nil, fmt.Errorf("invalid argument format: %s", arg)
	}

	if user == "" || password == "" || role == "" {
		// TODO: validate role
		return nil, fmt.Errorf("invalid argument format: %s", arg)
	}

	secret := scram.GenerateScramSHA256Secret(password, nil, 0)
	return append(requests,
		&zk.CreateRequest{
			Path:  internal.AclRootPath + "/" + group + "/" + user,
			Data:  []byte(role),
			Acl:   acl,
			Flags: 0,
		},
		&zk.CreateRequest{
			Path:  internal.AclRootPath + "/" + group + "/" + user + "/" + propName,
			Data:  []byte(secret.EncodeToBase64()),
			Acl:   acl,
			Flags: 0,
		},
	), nil
}
