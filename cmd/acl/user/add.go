package user

import (
	"bytes"
	"fmt"
	"path"
	"syscall"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/config"
	"github.com/jam2in/arcus-cli/internal/scram"
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
		aclUser, _ := cmd.Flags().GetString("userName")
		aclPassword, _ := cmd.Flags().GetString("password")
		zkConn := cmd.Context().Value(config.ZkConnKey{}).(*zk.Conn)

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
		err = addUser(zkConn, groupName, userName, role, secret.EncodeToBase64(), aclUser, aclPassword)
		if err != nil {
			return err
		}
		fmt.Printf("User '%s' added to group '%s' successfully.\n", userName, groupName)

		return nil
	},
}

func addUser(zkConn *zk.Conn, groupName, userName, role, secret, aclUser, aclPassword string) error {
	err := isAuth(zkConn, groupName, aclUser, aclPassword)
	if err != nil {
		return err
	}

	// ex: /arcus_acl/group/user
	userPath := path.Join(config.AclRootPath, groupName, userName)
	// ex: /arcus_acl/group/user/authPassword
	authPath := path.Join(userPath, propName)

	var acl []zk.ACL
	if aclUser != "" && aclPassword != "" {
		adminACL := zk.DigestACL(zk.PermAll, aclUser, aclPassword)
		worldReadACL := zk.WorldACL(zk.PermRead)
		acl = append(adminACL, worldReadACL...)
	} else {
		acl = zk.WorldACL(zk.PermAll)
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
