package zookeeper

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/config"
	"github.com/jam2in/arcus-cli/internal/ssh"
	"github.com/jam2in/arcus-cli/internal/types"
)

const (
	zookeeperStartCommandTemplate = "%s/bin/zkServer.sh start"
	zookeeperStopCommandTemplate  = "%s/bin/zkServer.sh stop"
	zookeeperConfigTemplate       = `mkdir -p %[1]s && echo %[2]d > %[1]s/myid \ 
							mkdir -p %[3]s && \
							mv %[4]s %[4]s.bak$(date +%%s) 2>/dev/null || true && \
							cat << 'EOF' > %[4]s
							%[5]s
							EOF`
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

	ctx = context.WithValue(ctx, types.CtxZkConnKey{}, conn)
	ctx = context.WithValue(ctx, types.CtxZkAclKey{}, acl)

	return ctx, nil
}

func InitializeZK(zkConn *zk.Conn) error {
	ops := make([]any, 0)
	for _, p := range config.ArcusBasicPaths {
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

func StartZKProcess(ip, zkPath string) error {
	command := fmt.Sprintf(zookeeperStartCommandTemplate, zkPath)
	return ssh.RunSSHCommand(ip, command)
}

func StopZKProcess(ip, zkPath string) error {
	command := fmt.Sprintf(zookeeperStopCommandTemplate, zkPath)
	return ssh.RunSSHCommand(ip, command)
}

func StatZKProcess(server string) (string, error) {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_, err = conn.Write([]byte("stat"))
	if err != nil {
		return "", err
	}

	response, err := io.ReadAll(conn)
	if err != nil {
		return "", err
	}

	return string(response), nil
}

func ConfigureZKFiles(ip, port, zkPath string, myid int, zkServers []string) error {
	dataDir := path.Join(zkPath, "data")
	confDir := path.Join(zkPath, "conf")
	confPath := path.Join(confDir, "zoo.cfg")
	zooCfgContent := buildZooCfg(zkServers, zkPath, port)
	configZKCmd := fmt.Sprintf(zookeeperConfigTemplate, dataDir, myid, confDir, confPath, zooCfgContent)
	return ssh.RunSSHCommand(ip, configZKCmd)
}

func buildZooCfg(servers []string, zkPath, port string) string {
	var zooCfg strings.Builder
	zooCfg.WriteString("tickTime=2000\n")
	zooCfg.WriteString("initLimit=10\n")
	zooCfg.WriteString("syncLimit=5\n")
	zooCfg.WriteString(fmt.Sprintf("dataDir=%s/data\n", zkPath))
	zooCfg.WriteString(fmt.Sprintf("clientPort=%s\n", port))
	zooCfg.WriteString("standaloneEnabled=false\n")
	zooCfg.WriteString("reconfigEnabled=true\n")
	zooCfg.WriteString("4lw.commands.whitelist=*\n\n")
	zooCfg.WriteString("# Server Lists\n")

	for i, server := range servers {
		ip := strings.Split(server, ":")[0]
		zooCfg.WriteString(fmt.Sprintf("server.%d=%s:2888:3888\n", i+1, ip))
	}

	return zooCfg.String()
}
