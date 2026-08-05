package cluster

import (
	"fmt"
	"path"
	"strings"

	"github.com/jam2in/arcusctl/internal/ssh"
	"github.com/jam2in/arcusctl/internal/topology"
)

func installServers(
	topo *topology.ClusterTopology,
	version string,
	localTarPath string,
	edition topology.ClusterEdition,
) ([]topology.CacheServer, error) {
	var installed []topology.CacheServer
	handledHosts := make(map[string]struct{}, len(topo.Servers))

	installPath := memcachedInstallPath(topo.Path, version)
	binaryPath := path.Join(installPath, "bin", "memcached")

	for i, server := range topo.Servers {
		host := server.Host()

		if _, ok := handledHosts[host]; ok {
			fmt.Printf("[%d/%d] %s: installation already handled, skipping\n",
				i+1, len(topo.Servers), host)
			continue
		}

		exists, err := ssh.FileExists(host, binaryPath)
		if err != nil {
			return installed, fmt.Errorf("check installation on %s: %w", host, err)
		}

		if exists {
			fmt.Printf("[%d/%d] %s: already installed, skipping\n",
				i+1, len(topo.Servers), host)
			handledHosts[host] = struct{}{}
			continue
		}

		fmt.Printf("[%d/%d] %s: installing...\n", i+1, len(topo.Servers), host)
		installed = append(installed, server)

		if err := installArchive(host, topo.Path, version, localTarPath, edition); err != nil {
			return installed, fmt.Errorf("install %s: %w", host, err)
		}

		handledHosts[host] = struct{}{}
	}

	return installed, nil
}

func installArchive(
	host string,
	basePath string,
	version string,
	localTarPath string,
	edition topology.ClusterEdition,
) error {
	installPath := memcachedInstallPath(basePath, version)
	sourcePath := path.Join(installPath, "src")

	mkdirCmd := fmt.Sprintf("mkdir -p %s %s", installPath, sourcePath)
	if err := ssh.Run(host, mkdirCmd); err != nil {
		return fmt.Errorf("mkdir installation paths on %s: %w", host, err)
	}

	remoteTarPath := path.Join(
		installPath,
		fmt.Sprintf("arcus-memcached-%s.tar.gz", version),
	)
	if err := ssh.Copy(localTarPath, host, remoteTarPath); err != nil {
		return fmt.Errorf("copy file to %s: %w", host, err)
	}

	extractCmd := fmt.Sprintf(
		"tar -xzf %s -C %s --strip-components=1",
		remoteTarPath,
		sourcePath,
	)
	if err := ssh.Run(host, extractCmd); err != nil {
		return fmt.Errorf("extract file on %s: %w", host, err)
	}

	depsScriptPath := path.Join(sourcePath, "deps", "install.sh")

	// arcus-memcached's deps/install.sh builds cyrus-sasl, which on macOS defaults
	// to building the macOS SASL2 Framework and fails. On Darwin targets only,
	// inject --disable-macos-framework into the cyrus-sasl configure line.
	patchCmd := fmt.Sprintf(
		`if [ "$(uname -s)" = "Darwin" ]; then `+
			`sed -i '' 's|$cyrussasl $prefix|$cyrussasl "$prefix --disable-macos-framework"|' %s; `+
			`fi`,
		depsScriptPath,
	)
	if err := ssh.Run(host, patchCmd); err != nil {
		return fmt.Errorf("patch deps script on %s: %w", host, err)
	}

	depsCmd := fmt.Sprintf("%s %s", depsScriptPath, installPath)
	if err := ssh.Run(host, depsCmd); err != nil {
		return fmt.Errorf("install dependencies on %s: %w", host, err)
	}

	buildCmd := buildCommand(sourcePath, installPath, edition)
	if err := ssh.Run(host, buildCmd); err != nil {
		return fmt.Errorf("build on %s: %w", host, err)
	}

	return nil
}

func buildCommand(
	sourcePath string,
	installPath string,
	edition topology.ClusterEdition,
) string {
	options := []string{
		fmt.Sprintf("--prefix=%s", installPath),
		"--enable-zk-integration",
		"--with-zk-reconfig",
		fmt.Sprintf("--with-libevent=%s", installPath),
		fmt.Sprintf("--with-zookeeper=%s", installPath),
	}

	if edition == topology.EnterpriseEdition {
		options = append(options, "--enable-replication")
	}

	return fmt.Sprintf(
		"cd %s && ./configure %s && make && make install",
		sourcePath,
		strings.Join(options, " "),
	)
}
