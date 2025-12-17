package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Printf("Could not connect to server: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	serverReader := bufio.NewReader(conn)

	fmt.Println("Redis-Lite Client")
	fmt.Println("Type commands (e.g., SET key val, SUBSCRIBE news, PUBLISH news hello)")
	fmt.Println("Type 'exit' to quit.")

	for {
		fmt.Print("redis-lite> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		line := strings.TrimSpace(input)
		if line == "" {
			continue
		}
		if strings.ToLower(line) == "exit" {
			break
		}

		// Note: We add \n because our server uses ReadString('\n')
		_, err = conn.Write([]byte(line + "\n"))
		if err != nil {
			fmt.Printf("Error writing to server: %v\n", err)
			break
		}

		// Handle Subscription Mode
		// If the user typed SUBSCRIBE, we enter a special "Listen" loop
		if strings.HasPrefix(strings.ToUpper(line), "SUBSCRIBE") {
			fmt.Println("Entering subscription mode. Press Ctrl+C to exit client.")
			for {
				// The server will push RESP arrays continuously
				response, err := serverReader.ReadString('\n')
				if err != nil {
					fmt.Printf("\nConnection lost: %v\n", err)
					return
				}
				// Clean up and print the raw RESP
				fmt.Print(strings.ReplaceAll(response, "\r", ""))
			}
		}

		// Handle Standard Response
		// Read the single response from the server
		response, err := serverReader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading from server: %v\n", err)
			break
		}

		// Basic RESP Pretty-printing
		// +OK -> OK
		// -ERR -> (error) ERR
		// :1 -> (integer) 1
		// $5\r\nhello -> hello
		output := strings.TrimSpace(response)
		if strings.HasPrefix(output, "+") {
			fmt.Println(output[1:])
		} else if strings.HasPrefix(output, "-") {
			fmt.Println("(error)", output[1:])
		} else if strings.HasPrefix(output, ":") {
			fmt.Println("(integer)", output[1:])
		} else if strings.HasPrefix(output, "$") {
			// It's a bulk string. Read the next line for the actual content.
			content, _ := serverReader.ReadString('\n')
			fmt.Print(strings.ReplaceAll(content, "\r", ""))
		} else if strings.HasPrefix(output, "*") {
			// It's an array. For simplicity, we'll just print the raw count
			// In a real client, you'd loop and read the items.
			fmt.Printf("(array of %s items)\n", output[1:])
		} else {
			fmt.Println(output)
		}
	}
}
