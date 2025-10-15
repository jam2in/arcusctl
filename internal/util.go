package internal

import (
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
)

func ConnectZooKeeper(addr string) (*zk.Conn, error) {
	conn, _, err := zk.Connect(strings.Split(addr, ","), time.Second,
		zk.WithLogInfo(Flags.Verbose))
	return conn, err
}
