package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	host := "localhost"
	port := "6379"

	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--host":
			if i+1 < len(args) {
				host = args[i+1]
				i++ // skip next item since we consumed it as value
			} else {
				fmt.Println("Error: --host requires a value")
				os.Exit(1)
			}
		case arg == "--port":
			if i+1 < len(args) {
				port = args[i+1]
				i++
			} else {
				fmt.Println("Error: --host requires a value")
				os.Exit(1)
			}
		case arg == "--help":
			fmt.Println("Usage: client --host <ip> --port <port>")
			return
		default:
			fmt.Printf("Unknown argument: %s\n", arg)
			os.Exit(1)
		}
	}

	addr := net.JoinHostPort(host, port)
	fmt.Printf("Connecting to %s...", addr)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", addr)
		os.Exit(1)
	}
	defer conn.Close()

	runInteractiveLoop(conn, addr)
}

func runInteractiveLoop(conn net.Conn, prefix string) {
	reader := bufio.NewReader(os.Stdin)
	serverReader := bufio.NewReader(conn)

	fmt.Println("Connected! Enter a command: ")

	for {
		fmt.Print(prefix + "> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if text == "exit" || text == "quit" {
			break
		}

		if text == "" {
			continue
		}

		fmt.Fprintf(conn, "%s\n", text)
		line1, err := serverReader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection lost: ", err)
			return
		}

		fmt.Print(line1)

		if strings.HasPrefix(line1, "$") {
			// special case: "$-1" means NULL (Key not found), so don't read more.
			if strings.TrimSpace(line1) == "$-1" {
				continue
			}

			line2, err := serverReader.ReadString('\n')
			if err != nil {
				fmt.Println("connection lost: ", err)
				return
			}
			fmt.Println(line2)
		}
	}
}
