package topology

import (
	"errors"
	"fmt"
	"strings"
)

type ClusterEdition string

const (
	CommunityEdition  ClusterEdition = "community"
	EnterpriseEdition ClusterEdition = "enterprise"
)

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
	Port int    `yaml:"port"`
}

type CacheConfig struct {
	Options string `yaml:"options,omitempty"`
}

func (topo *ClusterTopology) Edition() (ClusterEdition, error) {
	groupCount := 0

	for _, server := range topo.Servers {
		if server.Group != nil {
			groupCount++
		}
	}

	switch {
	case groupCount == 0:
		return CommunityEdition, nil

	case groupCount == len(topo.Servers):
		return EnterpriseEdition, nil

	default:
		return "", fmt.Errorf(
			"%w: grouped=%d, ungrouped=%d",
			errors.New("community and enterprise servers cannot be mixed"),
			groupCount,
			len(topo.Servers)-groupCount,
		)
	}
}

func (topo *ClusterTopology) Validate() error {
	if strings.TrimSpace(topo.ServiceCode) == "" {
		return fmt.Errorf("cluster servicecode is required")
	}

	if strings.TrimSpace(topo.Path) == "" {
		return fmt.Errorf("cluster installation path is required")
	}

	if strings.TrimSpace(topo.ZooKeeper) == "" {
		return fmt.Errorf("ZooKeeper address is required")
	}

	if len(topo.Servers) == 0 {
		return fmt.Errorf("no servers defined in topology")
	}

	seenAddress := make(map[string]struct{}, len(topo.Servers))
	for _, s := range topo.Servers {
		if strings.TrimSpace(s.Address) == "" {
			return fmt.Errorf("server address is required")
		}

		if _, ok := seenAddress[s.Address]; ok {
			return fmt.Errorf("duplicate address: %s", s.Address)
		}
		seenAddress[s.Address] = struct{}{}
	}

	edition, err := topo.Edition()
	if err != nil {
		return err
	}

	if edition == EnterpriseEdition {
		if err := topo.validateGroups(); err != nil {
			return err
		}
	}

	return nil
}

func (topo *ClusterTopology) validateGroups() error {
	type groupCount struct{ masters, slaves int }
	groups := map[string]*groupCount{}

	for _, server := range topo.Servers {
		if server.Group.Port <= 0 || server.Group.Port > 65535 {
			return fmt.Errorf("server %s: invalid group port %d (must be between 1 and 65535)",
				server.Address, server.Group.Port)
		}

		group := groups[server.Group.Name]
		if group == nil {
			group = &groupCount{}
			groups[server.Group.Name] = group
		}

		switch server.Group.Role {
		case "master":
			group.masters++
		case "slave":
			group.slaves++
		default:
			return fmt.Errorf("server %s: invalid group role %q (must be master or slave)",
				server.Address, server.Group.Role)
		}
	}

	for name, count := range groups {
		if count.masters != 1 {
			return fmt.Errorf("group %q: must have exactly 1 master, got %d", name, count.masters)
		}
		if count.slaves > 1 {
			return fmt.Errorf("group %q: must have at most 1 slave, got %d", name, count.slaves)
		}
	}

	return nil
}

func (s *CacheServer) IsMaster() bool {
	return s.Group != nil && s.Group.Role == "master"
}

func (s *CacheServer) Host() string {
	return strings.SplitN(s.Address, ":", 2)[0]
}
