package internal

import (
	"errors"
	"fmt"
	"log"
	"path"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
)

const zooKeeperSessionTimeout = time.Second

type znodeCreator interface {
	Exists(path string) (bool, *zk.Stat, error)
	Create(path string, data []byte, flags int32, acl []zk.ACL) (string, error)
}

type znodeDeleter interface {
	Children(path string) ([]string, *zk.Stat, error)
	Delete(path string, version int32) error
}

func ConnectZooKeeper(addresses string) (*zk.Conn, error) {
	servers, err := parseZooKeeperAddresses(addresses)
	if err != nil {
		return nil, err
	}

	conn, _, err := zk.Connect(
		servers,
		zooKeeperSessionTimeout,
		zk.WithLogInfo(Flags.Verbose),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to ZooKeeper at %q: %w", addresses, err)
	}

	return conn, nil
}

func parseZooKeeperAddresses(addresses string) ([]string, error) {
	parts := strings.Split(addresses, ",")
	servers := make([]string, 0, len(parts))
	for _, part := range parts {
		server := strings.TrimSpace(part)
		if server == "" {
			return nil, fmt.Errorf("invalid ZooKeeper addresses %q: address must not be empty", addresses)
		}
		servers = append(servers, server)
	}

	return servers, nil
}

// EnsureZNode creates zpath and any missing parents. Existing nodes are left
// unchanged, including when another client creates one concurrently.
func EnsureZNode(conn *zk.Conn, zpath string) error {
	if conn == nil {
		return errors.New("ensure znode: ZooKeeper connection is nil")
	}
	return ensureZNode(conn, zpath)
}

func ensureZNode(client znodeCreator, zpath string) error {
	if err := validateZNodePath(zpath); err != nil {
		return fmt.Errorf("ensure znode: %w", err)
	}
	if zpath == "/" {
		return nil
	}

	current := ""
	for _, segment := range strings.Split(strings.TrimPrefix(zpath, "/"), "/") {
		current += "/" + segment

		exists, _, err := client.Exists(current)
		if err != nil {
			return fmt.Errorf("check znode %q: %w", current, err)
		}
		if exists {
			logZNodeState(current, "exists")
			continue
		}

		if _, err := client.Create(current, []byte{}, 0, zk.WorldACL(zk.PermAll)); err != nil {
			if errors.Is(err, zk.ErrNodeExists) {
				logZNodeState(current, "exists")
				continue
			}
			return fmt.Errorf("create znode %q: %w", current, err)
		}
		logZNodeState(current, "created")
	}

	return nil
}

// DeleteZNode recursively deletes zpath. A node that is already absent is
// treated as successfully deleted.
func DeleteZNode(conn *zk.Conn, zpath string) error {
	if conn == nil {
		return errors.New("delete znode: ZooKeeper connection is nil")
	}
	if err := validateZNodePath(zpath); err != nil {
		return fmt.Errorf("delete znode: %w", err)
	}
	if zpath == "/" {
		return errors.New("delete znode: refusing to delete the ZooKeeper root")
	}

	return deleteZNode(conn, zpath)
}

func deleteZNode(client znodeDeleter, zpath string) error {
	children, _, err := client.Children(zpath)
	if errors.Is(err, zk.ErrNoNode) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("list children of znode %q: %w", zpath, err)
	}

	for _, child := range children {
		if err := deleteZNode(client, path.Join(zpath, child)); err != nil {
			return err
		}
	}

	if err := client.Delete(zpath, -1); err != nil && !errors.Is(err, zk.ErrNoNode) {
		return fmt.Errorf("delete znode %q: %w", zpath, err)
	}

	return nil
}

func validateZNodePath(zpath string) error {
	if zpath == "" {
		return errors.New("path must not be empty")
	}
	if !strings.HasPrefix(zpath, "/") {
		return fmt.Errorf("path %q must be absolute", zpath)
	}
	if cleaned := path.Clean(zpath); cleaned != zpath {
		return fmt.Errorf("path %q is not canonical; use %q", zpath, cleaned)
	}
	return nil
}

func logZNodeState(zpath string, state string) {
	if Flags.Verbose {
		log.Printf("znode %s %s", zpath, state)
	}
}
