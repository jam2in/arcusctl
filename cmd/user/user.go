package user

import (
	"fmt"

	"github.com/spf13/cobra"
)

const propName = "authPassword"

type UserInfo struct {
	Username string
	Roles    []string
}

func (i UserInfo) String() string {
	return fmt.Sprintf("Username: %s, Role: %s", i.Username, i.Roles)
}

var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage SASL users within an ACL group",
	Long:  "Manages users and their credentials within a specific ACL group.",
}

func init() {
	UserCmd.AddCommand(listCmd)
	UserCmd.AddCommand(addCmd)
	UserCmd.AddCommand(removeCmd)
}
