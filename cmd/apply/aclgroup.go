package apply

import (
	"fmt"
	"strings"
	"syscall"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/config"
	"github.com/jam2in/arcus-cli/internal"
	"github.com/jam2in/arcus-cli/internal/aclgroup"
	"github.com/jam2in/arcus-cli/internal/scram"
	"golang.org/x/term"
)

const (
	propName = "authPassword"
)

func passwordPrompt(verify bool) (string, error) {
	rawPassword, err := term.ReadPassword(syscall.Stdin)
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("ReadPassword: %w", err)
	}
	password := string(rawPassword)

	if verify {
		fmt.Printf("Repeat: ")
		repeatPassword, err := term.ReadPassword(syscall.Stdin)
		fmt.Println()
		if err != nil {
			return "", fmt.Errorf("ReadPassword: %w", err)
		}
		if password != string(repeatPassword) {
			return "", fmt.Errorf("does not match")
		}
	}

	return string(rawPassword), nil
}

type aclgroupApplyTask struct {
	groupExist bool
	exist      []aclgroup.User
	deleted    []aclgroup.User
	added      []aclgroup.User
	address    string
	groupName  string
	adminName  string
}

func (t *aclgroupApplyTask) Description() string {
	desc := ""

	if t.groupExist {
		desc += fmt.Sprintf("- %s/%s unchanged\n", aclgroup.RootZpath, t.groupName)
	} else {
		desc += fmt.Sprintf("- %s/%s created\n", aclgroup.RootZpath, t.groupName)
	}

	for _, u := range t.exist {
		desc += fmt.Sprintf("- %s/%s/%s unchanged\n",
			aclgroup.RootZpath, t.groupName, u.Username)
	}
	for _, u := range t.deleted {
		desc += fmt.Sprintf("- %s/%s/%s deleted\n",
			aclgroup.RootZpath, t.groupName, u.Username)
	}
	for _, u := range t.added {
		desc += fmt.Sprintf("- %s/%s/%s added\n",
			aclgroup.RootZpath, t.groupName, u.Username)
	}

	return desc
}

func (t *aclgroupApplyTask) Execute() error {
	if t.groupExist && len(t.added) == 0 && len(t.deleted) == 0 {
		return nil
	}

	conn, _, err := zk.Connect(strings.Split(t.address, ","), time.Second,
		zk.WithLogInfo(config.Verbose))
	if err != nil {
		return fmt.Errorf("Connect: %w", err)
	}
	defer conn.Close()

	acl := zk.WorldACL(zk.PermAll)
	if t.adminName != "" {
		fmt.Printf("password for %s admin(%s): ", t.groupName, t.adminName)
		password, err := passwordPrompt(!t.groupExist)
		if err != nil {
			return fmt.Errorf("passwordPrompt: %w", err)
		}

		if err := conn.AddAuth("digest", []byte(t.adminName+":"+password)); err != nil {
			return fmt.Errorf("AddAuth: %w", err)
		}

		acl = append(
			zk.DigestACL(zk.PermAll, t.adminName, password),
			zk.WorldACL(zk.PermRead)...)
	}

	var ops []any
	groupZpath := aclgroup.RootZpath + "/" + t.groupName

	if !t.groupExist {
		ops = append(ops,
			&zk.CreateRequest{
				Path:  groupZpath,
				Data:  nil,
				Acl:   acl,
				Flags: 0,
			},
		)
	}

	for _, u := range t.deleted {
		username := u.Username
		if u.AuditLogAll {
			username = "*" + u.Username
		}
		userZpath := groupZpath + "/" + username
		ops = append(ops,
			&zk.DeleteRequest{
				Path:    userZpath + "/" + propName,
				Version: -1,
			},
			&zk.DeleteRequest{
				Path:    userZpath,
				Version: -1,
			},
		)
	}

	for _, u := range t.added {
		fmt.Printf("password for %s/%s: ", t.groupName, u.Username)
		password, err := passwordPrompt(true)
		if err != nil {
			return fmt.Errorf("passwordPrompt: %w", err)
		}

		username := u.Username
		if u.AuditLogAll {
			username = "*" + u.Username
		}
		userZpath := groupZpath + "/" + username
		secret := scram.GenerateScramSHA256Secret(password, nil, 0)
		ops = append(ops,
			&zk.CreateRequest{
				Path:  userZpath,
				Data:  []byte(strings.Join(u.Permissions, ",")),
				Acl:   acl,
				Flags: 0,
			},
			&zk.CreateRequest{
				Path:  userZpath + "/" + propName,
				Data:  []byte(secret.EncodeToBase64()),
				Acl:   acl,
				Flags: 0,
			},
		)
	}

	if responses, err := conn.Multi(ops...); err != nil {
		return fmt.Errorf("Multi: %w (%#v)", err, responses)
	}

	return nil
}

func aclUserString(user aclgroup.User) string {
	return fmt.Sprintf("name=%s, perm=%v, log=%v",
		user.Username, user.Permissions, user.AuditLogAll)
}

func applyAclgroup(r aclgroup.Resource) (internal.Task, error) {
	current, err := aclgroup.GetResource(r.ZooKeeper, r.Group.Name)
	if err != nil {
		return nil, fmt.Errorf("GetResource: %w", err)
	}

	if current.Group.Name != "" && current.Group.AdminUsername != r.Group.AdminUsername {
		return nil, fmt.Errorf("group admin name does not match(%s, %s)",
			current.Group.AdminUsername, r.Group.AdminUsername)
	}

	var addedUser []aclgroup.User
	var existUser []aclgroup.User
	for _, after := range r.Users {
		found := false
		for _, before := range current.Users {
			if aclUserString(before) == aclUserString(after) {
				found = true
				existUser = append(existUser, after)
				break
			}
		}

		if !found {
			addedUser = append(addedUser, after)
		}
	}

	var removedUser []aclgroup.User
	for _, before := range current.Users {
		found := false
		for _, after := range r.Users {
			if aclUserString(before) == aclUserString(after) {
				found = true
				break
			}
		}

		if !found {
			removedUser = append(removedUser, before)
		}
	}

	return &aclgroupApplyTask{
		groupExist: current.Group.Name != "",
		added:      addedUser,
		exist:      existUser,
		deleted:    removedUser,
		address:    r.ZooKeeper,
		groupName:  r.Group.Name,
		adminName:  r.Group.AdminUsername,
	}, nil
}
