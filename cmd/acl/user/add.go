package user

import (
	"bytes"
	"fmt"
	"path"
	"syscall"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/config"
	"github.com/jam2in/arcus-cli/internal/scram"
	"github.com/jam2in/arcus-cli/internal/zookeeper"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var addCmd = &cobra.Command{
	Use:   "add <groupName> <userName> <role>",
	Short: "Add a new user to an ACL group.",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		groupName := args[0]
		userName := args[1]
		role := args[2]
		adminUser, _ := cmd.Flags().GetString("userName")
		adminPassword, _ := cmd.Flags().GetString("password")

		fmt.Print("Enter Password: ")
		rawPassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return err
		}
		fmt.Println()

		fmt.Print("Repeat Password: ")
		repeatPassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return err
		}
		fmt.Println()

		if !bytes.Equal(rawPassword, repeatPassword) {
			return fmt.Errorf("passwords do not match")
		}
		password := string(rawPassword)
		secret := scram.GenerateScramSHA256Secret(password, nil, 0)
		err = addUser(groupName, userName, role, secret.EncodeToBase64(), adminUser, adminPassword)
		if err != nil {
			return err
		}
		fmt.Printf("User '%s' added to group '%s' successfully.\n", userName, groupName)

		return nil
	},
}

func addUser(groupName, userName, role, secret, adminUser, adminPassword string) error {
	zkConn, err := zookeeper.NewConnect()
	if err != nil {
		return err
	}
	defer zkConn.Close()

	acl, err := isAuth(zkConn, groupName, adminUser, adminPassword)
	if err != nil {
		return err
	}

	// ex: /arcus_acl/group/user
	userPath := path.Join(config.AclRootPath, groupName, userName)
	// ex: /arcus_acl/group/user/authPassword
	authPath := path.Join(userPath, propName)
	for _, a := range acl {
		fmt.Printf("%s, %s\n", a.ID, a.Scheme)
	}
	ops := []interface{}{
		&zk.CreateRequest{
			Path:  userPath,
			Data:  []byte(role),
			Acl:   acl,
			Flags: 0,
		},
		&zk.CreateRequest{
			Path:  authPath,
			Data:  []byte(secret),
			Acl:   acl,
			Flags: 0,
		},
	}
	_, err = zkConn.Multi(ops...)

	return err
}
