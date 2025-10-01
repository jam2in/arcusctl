package types

import (
	"fmt"
	"strings"
)

type CtxZkConnKey struct{}
type CtxZkAclKey struct{}
type CtxZkCloseKey struct{}

type Status struct {
	ServiceCode    string
	Total          int
	OnlineServers  []string
	OfflineServers []string
}

func (s *Status) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("%-25s %-8s %-8s %-8s\n", "SERVICE CODE", "TOTAL", "ONLINE", "OFFLINE"))
	b.WriteString(strings.Repeat("-", 60) + "\n")
	b.WriteString(fmt.Sprintf("%-25s %-8d %-8d %-8d\n", s.ServiceCode, s.Total, len(s.OnlineServers), len(s.OfflineServers)))

	b.WriteString(fmt.Sprintf("Servers in serviceCode: '%s':\n", s.ServiceCode))
	if s.Total > 0 {
		b.WriteString("\n[Online Servers]\n")
		for _, server := range s.OnlineServers {
			b.WriteString(fmt.Sprintf("  - %s\n", server))
		}

		b.WriteString("\n[Offline Servers]\n")
		for _, server := range s.OfflineServers {
			b.WriteString(fmt.Sprintf("  - %s\n", server))
		}
	}
	return b.String()
}

type UserInfo struct {
	Username string
	Roles    []string
}

func (i UserInfo) String() string {
	return fmt.Sprintf("Username: %s, Role: %s", i.Username, i.Roles)
}
