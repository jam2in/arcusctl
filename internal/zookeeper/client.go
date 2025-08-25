package zookeeper

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
)

type ConnectKey struct{}

func NewConnect() (*zk.Conn, error) {
	addr := os.Getenv("ZK_ADDR")
	if addr == "" {
		fmt.Print("ZooKeeper address is not set. Please build with -ldflags.")
	}

	silentLogger := log.New(io.Discard, "", 0)
	conn, _, err := zk.Connect(strings.Split(addr, ","), time.Second*5, zk.WithLogger(silentLogger))
	if err != nil {
		fmt.Printf("Failed to connect to ZooKeeper: %v", err)
	}
	return conn, nil
}
