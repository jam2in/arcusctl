package zk

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jam2in/arcusctl/internal/ssh"
	"github.com/jam2in/arcusctl/internal/topology"
)

func installServers(topo *topology.ZKTopology, version string, localTarPath string) ([]topology.ZKServer, error) {
	var installed []topology.ZKServer
	archiveInstalled := map[string]bool{}

	for _, server := range topo.Servers {
		host := server.Host()
		fmt.Printf("[%d/%d] %s: installing...\n", len(installed)+1, len(topo.Servers), host)

		if !archiveInstalled[host] {
			if err := installArchive(host, topo.Path, version, localTarPath); err != nil {
				return installed, fmt.Errorf("install %s: %w", host, err)
			}
			archiveInstalled[host] = true
		}

		if err := configureServer(server, topo); err != nil {
			return installed, fmt.Errorf("install %s: %w", host, err)
		}
		installed = append(installed, server)
	}

	return installed, nil
}

func installArchive(host string, installPath string, version string, localTarPath string) error {
	if err := ssh.Run(host, fmt.Sprintf("mkdir -p %s", installPath)); err != nil {
		return fmt.Errorf("mkdir base path on %s: %w", host, err)
	}

	remoteTarPath := fmt.Sprintf("%s/zookeeper-%s.tar.gz", installPath, version)
	if err := ssh.Copy(localTarPath, host, remoteTarPath); err != nil {
		return fmt.Errorf("copy file to %s: %w", host, err)
	}

	extractCmd := fmt.Sprintf("tar -xzf %s -C %s --strip-components=1", remoteTarPath, installPath)
	if err := ssh.Run(host, extractCmd); err != nil {
		return fmt.Errorf("extract file on %s: %w", host, err)
	}

	return nil
}

func configureServer(server topology.ZKServer, topo *topology.ZKTopology) error {
	host := server.Host()
	confDir := fmt.Sprintf("%s/conf_myid_%d", topo.Path, server.MyID)
	dataDir := server.Config.DataDir
	dataDirPath := fmt.Sprintf("%s/zk%d", dataDir, server.MyID)

	mkdirCmd := fmt.Sprintf("mkdir -p %s %s", confDir, dataDirPath)
	if err := ssh.Run(host, mkdirCmd); err != nil {
		return fmt.Errorf("mkdir on %s: %w", host, err)
	}

	myidCmd := fmt.Sprintf("echo %d > %s/myid", server.MyID, dataDirPath)
	if err := ssh.Run(host, myidCmd); err != nil {
		return fmt.Errorf("write myid on %s: %w", host, err)
	}

	config := buildConfig(server, topo)
	if err := uploadFile(host, config, filepath.Join(confDir, "zoo.cfg")); err != nil {
		return fmt.Errorf("upload zoo.cfg to %s: %w", host, err)
	}

	dynamicConfig := buildDynamicConfig(topo)
	if err := uploadFile(host, dynamicConfig, filepath.Join(confDir, "zoo.cfg.dynamic")); err != nil {
		return fmt.Errorf("upload zoo.cfg.dynamic to %s: %w", host, err)
	}

	return nil
}

func uploadFile(host string, content string, remotePath string) error {
	tmp, err := os.CreateTemp("", "arcusctl-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(content); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	return ssh.Copy(tmp.Name(), host, remotePath)
}
