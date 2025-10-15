package internal

import (
	"fmt"
	"log"
	"strings"
	"syscall"
	"time"

	"github.com/go-zookeeper/zk"
	"golang.org/x/term"
)

func ConnectZooKeeper(addr string) (*zk.Conn, error) {
	conn, _, err := zk.Connect(strings.Split(addr, ","), time.Second,
		zk.WithLogInfo(Flags.Verbose))
	return conn, err
}

func EnsureZPath(conn *zk.Conn, zpath string) {
	if zpath == "" || zpath == "/" {
		return
	}

	exists, _, err := conn.Exists(zpath)
	if err != nil {
		panic(err)
	}
	if exists {
		if Flags.Verbose {
			log.Printf("%s exists\n", zpath)
		}
		return
	}

	parent := zpath[:strings.LastIndex(zpath, "/")]
	if parent != "" {
		EnsureZPath(conn, zpath)
	}

	if _, err := conn.Create(zpath, []byte{}, 0, zk.WorldACL(zk.PermAll)); err != nil {
		panic(err)
	}

	if Flags.Verbose {
		log.Printf("%s created\n", zpath)
	}
}

func ReadStdin(msg string, isPassword bool) string {
	fmt.Print(msg + ": ")
	if isPassword {
		raw, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			panic(err)
		}
		fmt.Println()
		return string(raw)
	} else {
		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			panic(err)
		}
		return input
	}
}
