package memcached

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/cybergarage/go-sasl/sasl"
	"github.com/cybergarage/go-sasl/sasl/mech"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect ip:port [username:password]",
	Short: "Connect to a memcached server",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := net.Dial("tcp", args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer conn.Close()

		if len(args) == 2 {
			username, password, _ := strings.Cut(args[1], ":")
			if err := authentication(conn, username, password); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}

		go func() {
			for {
				buf := make([]byte, 1024)
				n, err := conn.Read(buf)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				fmt.Print(string(buf[:n]))
			}
		}()

		reader := bufio.NewReader(os.Stdin)
		for {
			input, err := reader.ReadBytes('\n')
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if len(input) > 1 && input[len(input)-2] != '\r' {
				input = append(input[:len(input)-1], '\r', '\n')
			}
			if _, err := conn.Write(input); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	},
}

func authentication(conn net.Conn, username, password string) error {
	reader := bufio.NewReader(conn)

	client := sasl.NewClient()
	mechanism, err := client.Mechanism("SCRAM-SHA-256")
	if err != nil {
		return err
	}

	clientOpts := []mech.Option{
		mech.Username(username),
		mech.Password(password),
	}
	ctx, err := mechanism.Start(clientOpts...)
	if err != nil {
		return err
	}

	clientFirstMsg, err := ctx.Next()
	if err != nil {
		return err
	}

	clientFirstMsgString := clientFirstMsg.String()
	command := fmt.Sprintf("sasl auth SCRAM-SHA-256 %d\r\n%s\r\n",
		len(clientFirstMsgString), clientFirstMsgString)

	for {
		if _, err := conn.Write([]byte(command)); err != nil {
			return err
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			return err
		}
		line = bytes.TrimSpace(line)

		var serverMsg string
		responseCode := strings.Split(string(line), " ")[0]
		switch responseCode {
		case "SASL_OK":
			return ctx.Dispose()
		case "SASL_CONTINUE":
			serverFirstMsgRaw, err := reader.ReadBytes('\n')
			if err != nil {
				return err
			} else {
				serverMsg = string(serverFirstMsgRaw)[:len(serverFirstMsgRaw)-2]
			}
		default:
			return fmt.Errorf("unexpected response: %s", line)
		}

		clientMsg, err := ctx.Next(serverMsg)
		if err != nil {
			return err
		}

		var clientMsgString string
		if clientMsg != nil {
			clientMsgString = clientMsg.String()
		} else {
			clientMsgString = ""
		}
		command = fmt.Sprintf("sasl auth %d\r\n%s\r\n",
			len(clientMsgString), clientMsgString)
	}
}
