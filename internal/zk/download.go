package zk

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jam2in/arcusctl/internal"
)

const zkDownloadURLTemplate = "https://archive.apache.org/dist/zookeeper/zookeeper-%s/apache-zookeeper-%s-bin.tar.gz"

func ensureDownloaded(version string) (string, error) {
	dir := imageDir()

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create image dir: %w", err)
	}

	filename := fmt.Sprintf("apache-zookeeper-%s-bin.tar.gz", version)
	localPath := filepath.Join(dir, filename)

	if _, err := os.Stat(localPath); err == nil {
		fmt.Printf("Using existing file: %s\n", localPath)
		return localPath, nil
	}

	url := fmt.Sprintf(zkDownloadURLTemplate, version, version)
	fmt.Printf("Downloading %s...\n", url)

	cmd := exec.Command("wget", "-q", "-O", localPath, url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.Remove(localPath)
		return "", fmt.Errorf("download %q failed: %w", version, err)
	}

	return localPath, nil
}

func imageDir() string {
	return filepath.Join(internal.Config.Home, "images", "zookeeper")
}
