package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/cybergarage/go-sasl/sasl"
	"github.com/cybergarage/go-sasl/sasl/mech"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect <addr> [username:password]",
	Short: "Connect to an ARCUS cache server interactively",
	Long: `Connect to an ARCUS cache server and execute commands interactively.

This command establishes a TCP connection to the specified server address
and allows you to send memcached protocol commands directly. If authentication
credentials are provided, SASL SCRAM-SHA-256 authentication will be performed.`,
	Example: `  # Connect without authentication
  arcusctl connect localhost:11211

  # Connect with SASL authentication
  arcusctl connect localhost:11211 myuser:mypassword

  # After connection, you can execute memcached commands:
  get mykey
  set mykey 0 0 5
  hello`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := net.Dial("tcp", args[0])
		if err != nil {
			panic(err)
		}
		defer func() { _ = conn.Close() }()

		if len(args) == 2 {
			username, password, _ := strings.Cut(args[1], ":")
			if err := authentication(conn, username, password); err != nil {
				panic(err)
			}
		}

		log.Printf("connected to %s\n", conn.RemoteAddr())

		go func() {
			for {
				buf := make([]byte, 1024)
				n, err := conn.Read(buf)
				if err != nil {
					panic(err)
				}
				fmt.Print(string(buf[:n]))
			}
		}()

		reader := bufio.NewReader(os.Stdin)
		for {
			input, err := reader.ReadBytes('\n')
			if err != nil {
				panic(err)
			}
			if len(input) > 1 && input[len(input)-2] != '\r' {
				input = append(input[:len(input)-1], '\r', '\n')
			}
			if _, err := conn.Write(input); err != nil {
				panic(err)
			}
		}
	},
}

func authentication(conn net.Conn, username, password string) error {
	reader := bufio.NewReader(conn)

	command := "sasl mech\r\n"
	if _, err := conn.Write([]byte(command)); err != nil {
		return err
	}

	line, err := reader.ReadBytes('\n')
	if err != nil {
		return err
	}

	line = bytes.TrimSpace(line)
	responseCode := strings.Split(string(line), " ")[0]
	switch responseCode {
	case "SASL_MECH":
		if !bytes.Contains(line, []byte("SCRAM-SHA-256")) {
			return fmt.Errorf("SCRAM-SHA-256 not supported")
		}
	case "ERROR", "NOT_SUPPORTED":
		return nil
	default:
		return fmt.Errorf("unexpected response: %s", line)
	}

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
	command = fmt.Sprintf("sasl auth SCRAM-SHA-256 %d\r\n%s\r\n",
		len(clientFirstMsgString), clientFirstMsgString)

	for {
		if _, err := conn.Write([]byte(command)); err != nil {
			return err
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			return err
		}

		var serverMsg string
		line = bytes.TrimSpace(line)
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
