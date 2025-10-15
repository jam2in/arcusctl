package acl

import (
	"path"
	"strings"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/config"
	"github.com/jam2in/arcus-cli/internal/scram"
	"github.com/jam2in/arcus-cli/internal/types"
)

func AddUser(zkConn *zk.Conn, zkAcl []zk.ACL, groupName, userName, password, role string) error {
	secret := scram.GenerateScramSHA256Secret(password, nil, 0)
	ops := []any{
		&zk.CreateRequest{
			Path:  config.AclRootPath + "/" + groupName + "/" + userName,
			Data:  []byte(role),
			Acl:   zkAcl,
			Flags: 0,
		},
		&zk.CreateRequest{
			Path:  config.AclRootPath + "/" + groupName + "/" + userName + "/" + config.PropName,
			Data:  []byte(secret.EncodeToBase64()),
			Acl:   zkAcl,
			Flags: 0,
		},
	}
	_, err := zkConn.Multi(ops...)

	return err
}

func RemoveUser(zkConn *zk.Conn, groupName, userName string) error {
	ops := []any{
		&zk.DeleteRequest{
			Path:    config.AclRootPath + "/" + groupName + "/" + userName + "/" + config.PropName,
			Version: -1,
		},
		&zk.DeleteRequest{
			Path:    config.AclRootPath + "/" + groupName + "/" + userName,
			Version: -1,
		},
	}
	_, err := zkConn.Multi(ops...)

	return err
}

func GetUsers(zkConn *zk.Conn, groupName string) ([]types.UserInfo, error) {
	groupPath := config.AclRootPath + "/" + groupName

	userNames, _, err := zkConn.Children(groupPath)
	if err != nil {
		return nil, err
	}

	users := make([]types.UserInfo, 0, len(userNames))
	for _, userName := range userNames {
		userPath := path.Join(groupPath, userName)
		roleBytes, _, err := zkConn.Get(userPath)
		if err != nil {
			return nil, err
		}
		users = append(users, types.UserInfo{Username: userName, Roles: strings.Split(string(roleBytes), ",")})
	}
	return users, nil
}
