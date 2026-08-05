package internal

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/go-zookeeper/zk"
	"golang.org/x/term"
)

func ConnectZooKeeper(addr string) (*zk.Conn, error) {
	conn, _, err := zk.Connect(strings.Split(addr, ","), time.Second,
		zk.WithLogInfo(Flags.Verbose))
	return conn, err
}

// FXIME: EnsureZPath와 통합 가능 여부 검토 필요
func EnsureZNode(conn *zk.Conn, zpath string) error {
	if zpath == "" || zpath == "/" {
		return nil
	}

	exists, _, err := conn.Exists(zpath)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	parent := zpath[:strings.LastIndex(zpath, "/")]
	if parent != "" {
		if err := EnsureZNode(conn, parent); err != nil {
			return err
		}
	}

	_, err = conn.Create(zpath, []byte{}, 0, zk.WorldACL(zk.PermAll))
	if err == zk.ErrNodeExists {
		return nil
	}
	return err
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

func isValidPassword(p string) bool {
	if len(p) < 12 {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial int
	for _, c := range p {
		if unicode.IsUpper(c) {
			hasUpper = 1
		} else if unicode.IsLower(c) {
			hasLower = 1
		} else if unicode.IsDigit(c) {
			hasDigit = 1
		} else if unicode.IsPunct(c) || unicode.IsSymbol(c) {
			hasSpecial = 1
		}
	}

	return hasUpper+hasLower+hasDigit+hasSpecial >= 3
}

func ReadStdin(msg string, isPassword bool) string {
	fmt.Print(msg + ": ")
	if isPassword {
		raw, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			panic(err)
		}
		pw := string(raw)
		if !isValidPassword(pw) {
			panic("Password must be 12+ characters and include 3+ of: uppercase, lowercase, digits, symbols")
		}
		fmt.Println()
		return pw
	} else {
		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			panic(err)
		}
		return input
	}
}

// FIXME: ReadStdin와 통합 가능 여부 검토 필요
func Confirm(prompt string) bool {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

func DeleteZNode(conn *zk.Conn, zpath string) error {
	children, _, err := conn.Children(zpath)
	if err == zk.ErrNoNode {
		return nil
	}
	if err != nil {
		return err
	}

	for _, child := range children {
		if err := DeleteZNode(conn, zpath+"/"+child); err != nil {
			return err
		}
	}

	err = conn.Delete(zpath, -1)
	if err == zk.ErrNoNode {
		return nil
	}
	return err
}
