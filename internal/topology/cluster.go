package topology

import "strings"

type ClusterTopology struct {
	ServiceCode  string        `yaml:"servicecode"`
	Path         string        `yaml:"path"`
	ZooKeeper    string        `yaml:"zookeeper"`
	Servers      []CacheServer `yaml:"servers"`
	GlobalConfig CacheConfig   `yaml:"global_config"`
}

type CacheServer struct {
	Address string       `yaml:"address"`
	Group   *GroupInfo   `yaml:"group,omitempty"`
	Config  *CacheConfig `yaml:"config,omitempty"`
}

type GroupInfo struct {
	Name string `yaml:"name"`
	Role string `yaml:"role"`
}

type CacheConfig struct {
	Threads        int           `yaml:"threads,omitempty"`
	MaxConnections int           `yaml:"max_connections,omitempty"`
	Listen         string        `yaml:"listen,omitempty"`
	Engine         *EngineConfig `yaml:"engine,omitempty"`
}

type EngineConfig struct {
	MemoryLimit int  `yaml:"memory_limit,omitempty"`
	Eviction    bool `yaml:"eviction,omitempty"`
}

func (t *ClusterTopology) IsEnterprise() bool {
	for _, s := range t.Servers {
		if s.Group != nil {
			return true
		}
	}
	return false
}

func (s *CacheServer) Host() string {
	return strings.SplitN(s.Address, ":", 2)[0]
}
