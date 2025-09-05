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
)

func ContextWithZkConn(ctx context.Context, user, password string) (context.Context, error) {
	addr := os.Getenv("ZK_ADDR")
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
