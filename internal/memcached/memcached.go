package memcached

import (
	"fmt"
	"path"
	"strings"

	"github.com/go-zookeeper/zk"
	"github.com/jam2in/arcus-cli/config"
	"github.com/jam2in/arcus-cli/internal/ssh"
	"github.com/jam2in/arcus-cli/internal/types"
)

const (
	memcachedStartCommandTemplate = "%s/bin/memcached -E %s/lib/default_engine.so -X %s/lib/syslog_logger.so -X %s/lib/ascii_scrub.so -P %s/memcached-%s.pid -d -v -r -R5 -U 0 -D: -b 8192 %s -z %s"
	memcachedStopCommandTemplate  = "kill -INT $(cat %s/memcached-%s.pid)"
)

func GetClusterConfig(zkConn *zk.Conn, serviceCode string) ([]byte, error) {
	globalConfig, _, err := zkConn.Get(path.Join(config.ArcusCacheListPath, serviceCode))
	if err != nil {
		return nil, err
	}
	return globalConfig, nil
}

func SetClusterConfig(zkConn *zk.Conn, serviceCode, globalConfig string) error {
	cacheListPath := path.Join(config.ArcusCacheListPath, serviceCode)
	exists, _, err := zkConn.Exists(cacheListPath)
	if err != nil {
		return err
	}
	if exists {
		if _, err := zkConn.Set(cacheListPath, []byte(globalConfig), -1); err != nil {
			return err
		}
	}
	return nil
}

func AddToServiceCode(zkConn *zk.Conn, serviceCode, serverAddr string) error {
	ops1 := []any{
		&zk.CreateRequest{
			Path:  path.Join(config.ArcusCacheListPath, serviceCode),
			Data:  nil,
			Acl:   zk.WorldACL(zk.PermAll),
			Flags: 0,
		},
		&zk.CreateRequest{
			Path:  path.Join(config.ArcusClientListPath, serviceCode),
			Data:  nil,
			Acl:   zk.WorldACL(zk.PermAll),
			Flags: 0,
		},
	}
	if _, err := zkConn.Multi(ops1...); err != nil && err != zk.ErrNodeExists {
		return err
	}

	ops2 := []any{
		&zk.CreateRequest{
			Path:  path.Join(config.ArcusCacheServerMappingPath, serverAddr),
			Data:  nil,
			Acl:   zk.WorldACL(zk.PermAll),
			Flags: 0,
		},
		&zk.CreateRequest{
			Path:  path.Join(config.ArcusCacheServerMappingPath, serverAddr, serviceCode),
			Data:  nil,
			Acl:   zk.WorldACL(zk.PermAll),
			Flags: 0,
		},
	}
	_, err := zkConn.Multi(ops2...)
	return err
}

func RemoveFromServiceCode(zkConn *zk.Conn, serviceCode, serverAddr string) error {
	parts := strings.Split(serverAddr, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid address: %s", serverAddr)
	}

	ops := []any{
		&zk.DeleteRequest{
			Path:    path.Join(config.ArcusCacheServerMappingPath, serverAddr, serviceCode),
			Version: -1,
		},
		&zk.DeleteRequest{
			Path:    path.Join(config.ArcusCacheServerMappingPath, serverAddr),
			Version: -1,
		},
	}
	_, err := zkConn.Multi(ops...)

	return err
}

func RemoveServiceCode(zkConn *zk.Conn, serviceCode string) error {
	serverAddress, _, err := zkConn.Children(config.ArcusCacheServerMappingPath)
	if err != nil {
		return err
	}

	for _, address := range serverAddress {
		mappingPath := path.Join(config.ArcusCacheServerMappingPath, address, serviceCode)
		exists, _, err := zkConn.Exists(mappingPath)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%s exists", mappingPath)
		}
	}

	ops := []any{
		&zk.DeleteRequest{
			Path:    path.Join(config.ArcusCacheListPath, serviceCode),
			Version: -1,
		},
		&zk.DeleteRequest{
			Path:    path.Join(config.ArcusClientListPath, serviceCode),
			Version: -1,
		},
	}
	_, err = zkConn.Multi(ops...)

	return err
}

func GetServiceCodeStatus(zkConn *zk.Conn, serviceCode string) (*types.Status, error) {
	allServers, err := GetServiceCodeServers(zkConn, serviceCode)
	if err != nil {
		return nil, err
	}
	return buildStatusForServiceCode(zkConn, serviceCode, allServers)
}

func GetAllServiceCodeStatus(zkConn *zk.Conn) ([]*types.Status, error) {
	serviceCodeMap, err := buildServiceCodeMap(zkConn)
	if err != nil {
		return nil, err
	}

	var statuses []*types.Status
	for sc, allServers := range serviceCodeMap {
		status, err := buildStatusForServiceCode(zkConn, sc, allServers)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}

func GetServiceCodeServers(zkConn *zk.Conn, serviceCode string) ([]string, error) {
	servers, _, err := zkConn.Children(config.ArcusCacheServerMappingPath)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, server := range servers {
		children, _, err := zkConn.Children(path.Join(config.ArcusCacheServerMappingPath, server))
		if err != nil {
			continue
		}
		for _, child := range children {
			if child == serviceCode {
				result = append(result, server)
				break
			}
		}
	}
	return result, nil
}

func buildStatusForServiceCode(zkConn *zk.Conn, serviceCode string, allServers []string) (*types.Status, error) {
	liveServersMap, err := getLiveServers(zkConn, serviceCode)
	if err != nil && err != zk.ErrNoNode {
		return nil, err
	}

	var onlineServers []string
	var offlineServers []string
	for _, server := range allServers {
		if _, isLive := liveServersMap[server]; isLive {
			onlineServers = append(onlineServers, server)
		} else {
			offlineServers = append(offlineServers, server)
		}
	}

	return &types.Status{
		ServiceCode:    serviceCode,
		Total:          len(allServers),
		OnlineServers:  onlineServers,
		OfflineServers: offlineServers,
	}, nil
}

func buildServiceCodeMap(zkConn *zk.Conn) (map[string][]string, error) {
	serviceCodeMap := make(map[string][]string)
	allServers, _, err := zkConn.Children(path.Join(config.ArcusCacheServerMappingPath))
	if err != nil {
		return nil, err
	}

	for _, s := range allServers {
		serviceCodeTags, _, err := zkConn.Children(path.Join(config.ArcusCacheServerMappingPath, s))
		if err != nil {
			continue
		}
		for _, sc := range serviceCodeTags {
			serviceCodeMap[sc] = append(serviceCodeMap[sc], s)
		}
	}

	return serviceCodeMap, nil
}

func getLiveServers(zkConn *zk.Conn, serviceCode string) (map[string]struct{}, error) {
	liveNodes, _, err := zkConn.Children(path.Join(config.ArcusCacheListPath, serviceCode))
	if err != nil {
		return nil, err
	}

	liveServers := make(map[string]struct{})
	for _, liveNode := range liveNodes {
		addr, _, _ := strings.Cut(liveNode, "-")
		liveServers[addr] = struct{}{}
	}
	return liveServers, nil
}

func StartMemcachedProcess(zkServers, ip, port, arcusPath, config string) error {
	command := fmt.Sprintf(memcachedStartCommandTemplate,
		arcusPath, arcusPath, arcusPath, arcusPath, arcusPath,
		port, config, zkServers)
	return ssh.RunSSHCommand(ip, command)
}

func StopMemcachedProcess(ip, port, arcusPath string) error {
	command := fmt.Sprintf(memcachedStopCommandTemplate, arcusPath, port)
	return ssh.RunSSHCommand(ip, command)
}
