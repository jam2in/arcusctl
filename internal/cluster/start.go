package cluster

import (
	"fmt"

	"github.com/jam2in/arcusctl/internal/ssh"
	"github.com/jam2in/arcusctl/internal/topology"
)

const startCommandTemplate = "%s/bin/memcached" +
	" -E %s/lib/default_engine.so" +
	" -X %s/lib/ascii_scrub.so" +
	" -X %s/lib/syslog_logger.so" +
	" -p %s -P %s -z %s -d %s"

func Start(serviceCode string, nodeAddress string, groupName string) error {
	meta, topo, edition, err := loadCluster(serviceCode)
	if err != nil {
		return err
	}

	targets, err := selectTargets(topo, edition, nodeAddress, groupName, masterFirst)
	if err != nil {
		return err
	}

	total := len(targets)
	for i, server := range targets {
		fmt.Printf("[%d/%d] %s: starting...\n", i+1, total, server.Address)
		if err := startServer(server, topo, meta.Version); err != nil {
			return err
		}
	}

	fmt.Printf("Arcus cluster %q started.\n", serviceCode)
	return nil
}

func startServer(
	server topology.CacheServer,
	topo *topology.ClusterTopology,
	version string,
) error {
	command := startCommand(server, topo, version)
	if err := ssh.Run(server.Host(), command); err != nil {
		return fmt.Errorf("start %s: %w", server.Address, err)
	}
	return nil
}

func startCommand(
	server topology.CacheServer,
	topo *topology.ClusterTopology,
	version string,
) string {
	installPath := memcachedInstallPath(topo.Path, version)
	port := listenPort(server.Address)
	pidFile := pidFilePath(server.Address, topo.Path, version)
	options := mergedOptions(topo.GlobalConfig, server.Config)

	return fmt.Sprintf(
		startCommandTemplate,
		installPath, // bin/memcached
		installPath, // -E default_engine.so
		installPath, // -X syslog_logger.so
		installPath, // -X ascii_scrub.so
		port,
		pidFile,
		topo.ZooKeeper,
		options,
	)
}

func mergedOptions(global topology.CacheConfig, override *topology.CacheConfig) string {
	if override != nil && override.Options != "" {
		return global.Options + " " + override.Options
	}
	return global.Options
}
