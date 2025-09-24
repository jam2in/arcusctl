package internal

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
)

type CtxZkConnKey struct{}
type CtxZkAclKey struct{}
type CtxZkCloseKey struct{}

const (
	AclRootPath = "/arcus_acl"

	ArcusCacheListPath          = "/arcus/cache_list"
	ArcusClientListPath         = "/arcus/client_list"
	ArcusCacheServerMappingPath = "/arcus/cache_server_mapping"
)

var arcusBasicPaths = []string{
	"/arcus",
	"/arcus/cache_list",
	"/arcus/cache_server_mapping",
	"/arcus/client_list",
	"/arcus_repl",
	"/arcus_repl/cache_list",
	"/arcus_repl/cache_server_mapping",
	"/arcus_repl/client_list",
	"/arcus_repl/group_list",
	"/arcus_repl/cloud_stat",
	"/arcus_acl",
}

func ContextWithZkConn(ctx context.Context, user, password string) (context.Context, error) {
	addr := os.Getenv("ZK_LIST")
	if addr == "" {
		return nil, fmt.Errorf("ZooKeeper address is not set")
	}

	conn, _, err := zk.Connect(strings.Split(addr, ","), time.Second*5,
		zk.WithLogger(log.New(io.Discard, "", 0)))
	if err != nil {
		return nil, err
	}

	var acl []zk.ACL
	if user != "" && password != "" {
		if err := conn.AddAuth("digest", []byte(user+":"+password)); err != nil {
			return nil, err
		}
		acl = append(zk.DigestACL(zk.PermAll, user, password), zk.WorldACL(zk.PermRead)...)
	} else {
		acl = zk.WorldACL(zk.PermAll)
	}

	ctx = context.WithValue(ctx, CtxZkConnKey{}, conn)
	ctx = context.WithValue(ctx, CtxZkAclKey{}, acl)

	return ctx, nil
}

func InitializeZK(zkConn *zk.Conn) error {
	ops := make([]any, 0)
	for _, p := range arcusBasicPaths {
		exist, _, err := zkConn.Exists(p)
		if err != nil {
			return err
		}
		if !exist {
			ops = append(ops, &zk.CreateRequest{
				Path:  p,
				Data:  []byte{},
				Acl:   zk.WorldACL(zk.PermAll),
				Flags: 0,
			})
		}
	}
	if len(ops) > 0 {
		if _, err := zkConn.Multi(ops...); err != nil {
			return err
		}
	}
	return nil
}
