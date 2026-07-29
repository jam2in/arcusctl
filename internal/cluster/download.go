package cluster

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jam2in/arcusctl/internal"
	"github.com/jam2in/arcusctl/internal/topology"
)

const arcusDownloadURLTemplate = "https://github.com/naver/arcus-memcached/releases/download/%s/arcus-memcached-%s.tar.gz"

func ensureDownloaded(version string, edition topology.ClusterEdition) (string, error) {
	dir := imageDir(edition)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create image dir: %w", err)
	}

	filename := fmt.Sprintf("arcus-memcached-%s.tar.gz", version)
	localTarPath := filepath.Join(dir, filename)

	// check if the file already exists
	// enterprise edition may not be downloadable, so we check for existence first
	if _, err := os.Stat(localTarPath); err == nil {
		fmt.Printf("Using existing file: %s\n", localTarPath)
		return localTarPath, nil
	}

	if edition == topology.EnterpriseEdition {
		return "", fmt.Errorf("enterprise archive not found: place %s in %s", filename, dir)
	}

	url := fmt.Sprintf(arcusDownloadURLTemplate, version, version)
	fmt.Printf("Downloading %s...\n", url)

	cmd := exec.Command("wget", "-q", "-O", localTarPath, url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.Remove(localTarPath)
		return "", fmt.Errorf("download %q failed: %w", version, err)
	}

	return localTarPath, nil
}

func imageDir(edition topology.ClusterEdition) string {
	if edition == topology.EnterpriseEdition {
		return filepath.Join(internal.Config.Home, "images", "arcus-enterprise")
	}
	return filepath.Join(internal.Config.Home, "images", "arcus-community")
}
