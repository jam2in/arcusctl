package zookeeper

import (
	"github.com/spf13/cobra"
)

const (
	zookeeperConfigTemplate = `mkdir -p %[1]s && echo %[2]d > %[1]s/myid \ 
							mkdir -p %[3]s && \
							mv %[4]s %[4]s.bak$(date +%%s) 2>/dev/null || true && \
							cat << 'EOF' > %[4]s
							%[5]s
							EOF`
)

var ZookeeperCmd = &cobra.Command{
	Use:   "zookeeper",
	Short: "A CLI tool for zookeeper commands",
	Long: "A command-line interface to manage the ZooKeeper ensemble for an Arcus cluster.\n" +
		"This includes remotely initializing server configuration files (myid, zoo.cfg)\n" +
		"and controlling the lifecycle(start, stop, stat) of the Zookeeper processes.",
}

func init() {
	ZookeeperCmd.AddCommand(initCmd)
}
