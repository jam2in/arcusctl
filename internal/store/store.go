package store

import (
	"os"
	"path/filepath"
	"time"

	"github.com/jam2in/arcusctl/internal"
	"github.com/jam2in/arcusctl/internal/topology"
	"go.yaml.in/yaml/v3"
)

const (
	metaYML     = "meta.yml"
	topologyYML = "topology.yml"
)

type ZKMeta struct {
	Name       string    `yaml:"name"`
	Version    string    `yaml:"version"`
	DeployedAt time.Time `yaml:"deployed_at"`
}

type ClusterMeta struct {
	ServiceCode string    `yaml:"servicecode"`
	Version     string    `yaml:"version"`
	DeployedAt  time.Time `yaml:"deployed_at"`
}

func SaveZK(ensembleName string, version string, topologyData []byte) error {
	dir := zkDir(ensembleName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	meta := ZKMeta{
		Name:       ensembleName,
		Version:    version,
		DeployedAt: time.Now(),
	}
	metaData, err := yaml.Marshal(meta)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, metaYML), metaData, 0644); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, topologyYML), topologyData, 0644)
}

func LoadZKMeta(ensembleName string) (*ZKMeta, error) {
	data, err := os.ReadFile(filepath.Join(zkDir(ensembleName), metaYML))
	if err != nil {
		return nil, err
	}

	var meta ZKMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func LoadZKTopology(ensembleName string) (*topology.ZKTopology, error) {
	data, err := os.ReadFile(filepath.Join(zkDir(ensembleName), topologyYML))
	if err != nil {
		return nil, err
	}

	var topo topology.ZKTopology
	if err := yaml.Unmarshal(data, &topo); err != nil {
		return nil, err
	}

	return &topo, nil
}

func ZKExists(ensembleName string) bool {
	_, err := os.Stat(filepath.Join(zkDir(ensembleName), metaYML))
	return err == nil
}

func DeleteZK(ensembleName string) error {
	return os.RemoveAll(zkDir(ensembleName))
}

func ListZK() ([]string, error) {
	dir := zkBaseDir()
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() && ZKExists(e.Name()) {
			names = append(names, e.Name())
		}
	}

	return names, nil
}

func SaveCluster(serviceCode string, version string, topologyData []byte) error {
	dir := clusterDir(serviceCode)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	meta := ClusterMeta{
		ServiceCode: serviceCode,
		Version:     version,
		DeployedAt:  time.Now(),
	}
	metaData, err := yaml.Marshal(meta)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, metaYML), metaData, 0644); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, topologyYML), topologyData, 0644)
}

func LoadClusterMeta(serviceCode string) (*ClusterMeta, error) {
	data, err := os.ReadFile(filepath.Join(clusterDir(serviceCode), metaYML))
	if err != nil {
		return nil, err
	}

	var meta ClusterMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func ClusterExists(serviceCode string) bool {
	_, err := os.Stat(filepath.Join(clusterDir(serviceCode), metaYML))
	return err == nil
}

func ListCluster() ([]string, error) {
	dir := clusterBaseDir()
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() && ClusterExists(e.Name()) {
			names = append(names, e.Name())
		}
	}

	return names, nil
}

func zkBaseDir() string {
	return filepath.Join(internal.Config.Home, "clusters", "zk")
}

func zkDir(ensembleName string) string {
	return filepath.Join(zkBaseDir(), ensembleName)
}

func clusterBaseDir() string {
	return filepath.Join(internal.Config.Home, "clusters", "arcus")
}

func clusterDir(serviceCode string) string {
	return filepath.Join(clusterBaseDir(), serviceCode)
}
