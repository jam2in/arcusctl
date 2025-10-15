package aclgroup

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/config"
)

const (
	RootZpath = "/arcus_acl"
)

type User struct {
	Username    string   `yaml:"username"`
	Permissions []string `yaml:"permissions"`
	AuditLogAll bool     `yaml:"auditLogAll,omitempty"`
}

type Resource struct {
	ZooKeeper string `yaml:"zookeeper"`
	Group     struct {
		Name          string `yaml:"name"`
		AdminUsername string `yaml:"adminUsername,omitempty"`
	} `yaml:"group"`
	Users []User `yaml:"users"`
}

func GetResource(addr string, group string) (Resource, error) {
	r := Resource{}

	conn, _, err := zk.Connect(strings.Split(addr, ","), time.Second,
		zk.WithLogInfo(config.Verbose))
	if err != nil {
		return r, fmt.Errorf("Connect(%s): %w", addr, err)
	}
	defer conn.Close()

	r.ZooKeeper = addr

	groupZpath := RootZpath + "/" + group
	acls, _, err := conn.GetACL(groupZpath)
	if errors.Is(err, zk.ErrNoNode) {
		return r, nil
	} else if err != nil {
		return r, fmt.Errorf("GetACL(%s): %w", groupZpath, err)
	}

	r.Group.Name = group
	for _, acl := range acls {
		if acl.Scheme == "digest" && acl.Perms == zk.PermAll {
			// If there is more than one digest with PermAll, it will not work as expected
			adminUsername, _, _ := strings.Cut(acl.ID, ":")
			r.Group.AdminUsername = adminUsername
			break
		}
	}

	rawUsernames, _, err := conn.Children(groupZpath)
	if err != nil {
		return r, fmt.Errorf("Children(%s): %w", groupZpath, err)
	}

	for _, rawUsername := range rawUsernames {
		userZpath := groupZpath + "/" + rawUsername
		value, _, err := conn.Get(userZpath)
		if err != nil {
			return r, fmt.Errorf("Get(%s): %w", userZpath, err)
		}

		username, auditLogAll := strings.CutPrefix(rawUsername, "*")
		r.Users = append(r.Users, User{
			Username:    username,
			Permissions: strings.Split(string(value), ","),
			AuditLogAll: auditLogAll,
		})
	}

	return r, nil
}
