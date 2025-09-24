package zookeeper

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/jam2in/arcus-cli/internal"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ZooKeeper config files (myid, zoo.cfg) on ensemble servers",
	Run: func(cmd *cobra.Command, args []string) {
		zkList := os.Getenv("ZK_LIST")
		zkPath := os.Getenv("ZK_PATH")
		if zkList == "" || zkPath == "" {
			fmt.Fprintln(os.Stderr, "Environment variables are not provided.\nPlease set the ZK_LIST, ZK_PATH environment variables")
			os.Exit(1)
		}
		zkServers := strings.Split(zkList, ",")

		fmt.Printf("Initializing ZooKeeper config for %d server(s)...\n", len(zkServers))
		for i, server := range zkServers {
			ip, port, _ := strings.Cut(server, ":")
			myid := i + 1

			err := configZK(ip, port, zkPath, myid, zkServers)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		fmt.Println("ZooKeeper configuration complete. You can now start the ensemble using 'zookeeper start'.")
	},
}

func configZK(ip, port, zkPath string, myid int, allServers []string) error {
	dataDir := path.Join(zkPath, "data")
	confDir := path.Join(zkPath, "conf")
	confPath := path.Join(confDir, "zoo.cfg")
	zooCfgContent := buildZooCfg(allServers, zkPath, port)
	configZKCmd := fmt.Sprintf(zookeeperConfigTemplate, dataDir, myid, confDir, confPath, zooCfgContent)

	client, err := internal.NewSSHClient(ip)
	if err != nil {
		return err
	}
	defer client.Close()

	session, _ := client.NewSession()
	defer session.Close()
	if err = session.Run(configZKCmd); err != nil {
		return fmt.Errorf("failed to configure zookeeper: %w", err)
	}

	return nil
}

func buildZooCfg(servers []string, zkPath, port string) string {
	var zooCfg strings.Builder
	zooCfg.WriteString("tickTime=2000\n")
	zooCfg.WriteString("initLimit=10\n")
	zooCfg.WriteString("syncLimit=5\n")
	zooCfg.WriteString(fmt.Sprintf("dataDir=%s/data\n", zkPath))
	zooCfg.WriteString(fmt.Sprintf("clientPort=%s\n", port))
	zooCfg.WriteString("standaloneEnabled=false\n")
	zooCfg.WriteString("reconfigEnabled=true\n")
	zooCfg.WriteString("4lw.commands.whitelist=*\n\n")
	zooCfg.WriteString("# Server Lists\n")

	for i, server := range servers {
		ip := strings.Split(server, ":")[0]
		zooCfg.WriteString(fmt.Sprintf("server.%d=%s:2888:3888\n", i+1, ip))
	}

	return zooCfg.String()
}
