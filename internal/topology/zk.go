package topology

import "strings"

type ZKTopology struct {
	Name         string     `yaml:"name"`
	Path         string     `yaml:"path"`
	Servers      []ZKServer `yaml:"servers"`
	GlobalConfig ZKConfig   `yaml:"global_config"`
}

type ZKServer struct {
	MyID    int       `yaml:"myid"`
	Address string    `yaml:"address"`
	Config  *ZKConfig `yaml:"config,omitempty"`
}

type ZKConfig struct {
	TickTime   int               `yaml:"tick_time,omitempty"`
	InitLimit  int               `yaml:"init_limit,omitempty"`
	SyncLimit  int               `yaml:"sync_limit,omitempty"`
	DataDir    string            `yaml:"data_dir,omitempty"`
	DataLogDir string            `yaml:"data_log_dir,omitempty"`
	Properties map[string]string `yaml:"properties,omitempty"`
}

func (s *ZKServer) ParseAddress() (host, clientPort, quorumPort, electionPort string) {
	parts := strings.SplitN(s.Address, ":", 4)
	return parts[0], parts[1], parts[2], parts[3]
}

func (s *ZKServer) Host() string {
	host, _, _, _ := s.ParseAddress()
	return host
}
