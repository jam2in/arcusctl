package zk

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jam2in/arcusctl/internal/topology"
)

const (
	mntrDialTimeout = 2 * time.Second
	mntrReadTimeout = 3 * time.Second
)

var errMntrNotWhitelisted = errors.New(
	`mntr is not whitelisted; add "mntr" to 4lw.commands.whitelist in zoo.cfg. and restart`,
)

type Mntr struct {
	Version          string
	ServerState      string
	AliveConnections int
	SyncedFollowers  int
}

type probeResult struct {
	Mntr Mntr
	Err  error
}

func probeAll(targets map[int]string) map[int]probeResult {
	results := make(map[int]probeResult)

	var mutex sync.Mutex
	var group sync.WaitGroup

	for myID, addr := range targets {
		group.Go(func() {
			mntr, err := probe(addr)
			mutex.Lock()
			defer mutex.Unlock()
			results[myID] = probeResult{Mntr: mntr, Err: err}
		})
	}
	group.Wait()
	return results
}

func probe(addr string) (Mntr, error) {
	conn, err := net.DialTimeout("tcp", addr, mntrDialTimeout)
	if err != nil {
		return Mntr{}, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(mntrReadTimeout)); err != nil {
		return Mntr{}, fmt.Errorf("set deadline for %s: %w", addr, err)
	}

	if _, err := conn.Write([]byte("mntr")); err != nil {
		return Mntr{}, fmt.Errorf("send mntr command to %s: %w", addr, err)
	}

	body, err := io.ReadAll(conn)
	if err != nil {
		return Mntr{}, fmt.Errorf("read mntr response from %s: %w", addr, err)
	}

	return parseMntr(string(body))
}

func parseMntr(text string) (Mntr, error) {
	if strings.Contains(text, "not executed") {
		return Mntr{}, errMntrNotWhitelisted
	}

	mntr := Mntr{SyncedFollowers: -1}

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), "\t")
		if !ok {
			continue
		}

		switch key {
		case "zk_version":
			mntr.Version = shortVersion(value)
		case "zk_server_state":
			mntr.ServerState = strings.TrimSpace(value)
		case "zk_num_alive_connections":
			count, err := parseCount(key, value)
			if err != nil {
				return Mntr{}, err
			}
			mntr.AliveConnections = count
		case "zk_synced_followers":
			count, err := parseCount(key, value)
			if err != nil {
				return Mntr{}, err
			}
			mntr.SyncedFollowers = count
		}
	}

	if err := scanner.Err(); err != nil {
		return Mntr{}, fmt.Errorf("scan mntr response: %w", err)
	}

	if mntr.Version == "" || mntr.ServerState == "" {
		return Mntr{}, errors.New("mntr response is missing zk_version or zk_server_state")
	}

	return mntr, nil
}

func shortVersion(value string) string {
	if i := strings.IndexAny(value, "-,"); i >= 0 {
		return strings.TrimSpace(value[:i])
	}
	return strings.TrimSpace(value)
}

func parseCount(key string, value string) (int, error) {
	count, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s from mntr response: %w", key, err)
	}
	return count, nil
}

func topologyTargets(servers []topology.ZKServer) map[int]string {
	targets := make(map[int]string, len(servers))

	for _, server := range servers {
		host, clientPort, _, _ := server.ParseAddress()
		targets[server.MyID] = net.JoinHostPort(host, clientPort)
	}

	return targets
}
