package zookeeper

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var statCmd = &cobra.Command{
	Use:   "stat",
	Short: "Show the status of the Zookeeper ensemble servers",
	Run: func(cmd *cobra.Command, args []string) {
		zkAddr := os.Getenv("ZK_ADDR")
		if zkAddr == "" {
			fmt.Fprintln(os.Stderr, "Environment variable is not provided. \nPlease set the ZK_ADDR environment variable")
			os.Exit(1)
		}
		serversToCheck := strings.Split(zkAddr, ",")
		for _, server := range serversToCheck {
			status, err := getStatusFromZK(server)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			fmt.Println(status)
		}
	},
}

func getStatusFromZK(server string) (string, error) {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		return "", fmt.Errorf("could not connect to server %s: %v", server, err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte("stat"))
	if err != nil {
		return "", fmt.Errorf("could not send command to server %s: %v", server, err)
	}

	response, err := io.ReadAll(conn)
	if err != nil {
		return "", fmt.Errorf("could not read response from server %s: %v", server, err)
	}

	return string(response), nil
}
