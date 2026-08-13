package zk

import (
	"fmt"
	"os"

	"github.com/jam2in/arcusctl/internal/ssh"
	"github.com/jam2in/arcusctl/internal/topology"
)

func installServers(
	topo *topology.ZKTopology,
	version string,
	localTarPath string,
) ([]topology.ZKServer, error) {
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

func installArchive(
	host string,
	topoPath string,
	version string,
	localTarPath string,
) error {
	installPath := zkInstallPath(topoPath, version)
	scriptPath := zkServerScript(topoPath, version)

	exists, err := ssh.FileExists(host, scriptPath)
	if err != nil {
		return fmt.Errorf("check installation on %s: %w", host, err)
	}
	if exists {
		return nil
	}

	if err := ssh.Run(host, fmt.Sprintf("mkdir -p %s", installPath)); err != nil {
		return fmt.Errorf("mkdir base path on %s: %w", host, err)
	}

	remoteTarPath := fmt.Sprintf("%s/zookeeper-%s.tar.gz", installPath, version)
	if err := ssh.Copy(localTarPath, host, remoteTarPath); err != nil {
		return fmt.Errorf("copy file to %s: %w", host, err)
	}

	extractCmd := fmt.Sprintf(
		"tar -xzf %s -C %s --strip-components=1",
		remoteTarPath,
		installPath,
	)
	if err := ssh.Run(host, extractCmd); err != nil {
		return fmt.Errorf("extract file on %s: %w", host, err)
	}

	return nil
}

func configureServer(server topology.ZKServer, topo *topology.ZKTopology) error {
	host := server.Host()

	confDir := zkConfigDir(topo.Path, topo.Name, server.MyID)
	dataDirPath := zkNodeDataPath(server.Config.DataDir, server.MyID)
	dataLogDirPath := zkNodeDataPath(server.Config.DataLogDir, server.MyID)

	mkdirCmd := fmt.Sprintf(
		"mkdir -p %s %s %s",
		confDir,
		dataDirPath,
		dataLogDirPath,
	)
	if err := ssh.Run(host, mkdirCmd); err != nil {
		return fmt.Errorf("mkdir on %s: %w", host, err)
	}

	myidCmd := fmt.Sprintf("echo %d > %s/myid", server.MyID, dataDirPath)
	if err := ssh.Run(host, myidCmd); err != nil {
		return fmt.Errorf("write myid on %s: %w", host, err)
	}

	config := buildConfig(server, topo)
	configPath := zkConfigPath(topo.Path, topo.Name, server.MyID)
	if err := uploadFile(host, config, configPath); err != nil {
		return fmt.Errorf("upload zoo.cfg to %s: %w", host, err)
	}

	dynamicConfig := buildDynamicConfig(topo)
	dynamicConfigPath := zkDynamicConfigPath(topo.Path, topo.Name, server.MyID)
	if err := uploadFile(host, dynamicConfig, dynamicConfigPath); err != nil {
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
