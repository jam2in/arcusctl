package topology

import (
	"os"

	"go.yaml.in/yaml/v3"
)

func LoadZK(path string) (*ZKTopology, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	var topo ZKTopology
	if err := yaml.Unmarshal(data, &topo); err != nil {
		return nil, nil, err
	}

	return &topo, data, nil
}

func LoadCluster(path string) (*ClusterTopology, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	var topo ClusterTopology
	if err := yaml.Unmarshal(data, &topo); err != nil {
		return nil, nil, err
	}

	return &topo, data, nil
}
