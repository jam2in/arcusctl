package user

import (
	"bytes"
	"fmt"
	"os"
	"syscall"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/jam2in/arcus-cli/internal/scram"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var addCmd = &cobra.Command{
	Use:   "add <groupName> <userName> <role>",
	Short: "Add a new user to an ACL group.",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		userName := args[1]
		role := args[2]

		password, err := readPassword()
		if err != nil {
			panic(err) // FIXME
		}
		secret := scram.GenerateScramSHA256Secret(password, nil, 0)

		zkConn := cmd.Context().Value(internal.CtxZkConnKey{}).(*zk.Conn)
		acl := cmd.Context().Value(internal.CtxZkAclKey{}).([]zk.ACL)

		if _, err := zkConn.Multi(
			&zk.CreateRequest{
				Path:  internal.AclRootPath + "/" + groupName + "/" + userName,
				Data:  []byte(role),
				Acl:   acl,
				Flags: 0,
			},
			&zk.CreateRequest{
				Path:  internal.AclRootPath + "/" + groupName + "/" + userName + "/" + propName,
				Data:  []byte(secret.EncodeToBase64()),
				Acl:   acl,
				Flags: 0,
			},
		); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Printf("User '%s' added to group '%s' successfully.\n", userName, groupName)
	},
}

func readPassword() (string, error) {
	fmt.Print("Enter Password: ")
	rawPassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()

	fmt.Print("Repeat Password: ")
	repeatPassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()

	if !bytes.Equal(rawPassword, repeatPassword) {
		return "", fmt.Errorf("passwords do not match")
	}
	return string(rawPassword), nil
}
