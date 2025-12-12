package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"time"
)

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		message, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				slog.ErrorContext(ctx, "Read error", "error", err)
			}
			return
		}

		args := strings.Fields(strings.TrimSpace(message))
		if len(args) == 0 {
			continue
		}

		command := strings.ToUpper(args[0])
		var response string

		slog.Info("command: " + command)

		switch command {
		case "SET":
			// syntax: SET key value [ttl_seconds]
			if len(args) < 3 {
				response = "-ERR wrong number of arguments for `set` command\r\n"
			} else {
				key := args[1]
				val := args[2]
				var ttl time.Duration = 0 // def no expiration

				if len(args) > 3 {
					if d, err := time.ParseDuration(args[3]); err == nil {
						ttl = d
					}
				}

				s.DB.Set(key, val, ttl)

				response = "+OK\r\n"
			}
		case "GET":
			// syntax: GET key
			if len(args) < 2 {
				response = "-ERR wrong number of arguments for `get` command\r\n"
			} else {
				key := args[1]

				val, exists := s.DB.Get(key)
				if !exists {
					response = "$-1\r\n" // redis standart for not found
				} else {
					fmt.Println("THE VAL: ", val)
					response = fmt.Sprintf("$%d\r\n%v\r\n", len(fmt.Sprint(val)), val)
				}
			}
		case "DEL":
			// syntax: DEL key
			if len(args) < 2 {
				response = "-ERR wrong number of arguments for `get` command\r\n"
			}
			key := args[1]

			s.DB.Delete(key)
			response = "+OK\r\n"
		case "PING":
			response = "+PONG\r\n"
		default:
			response = "-ERR unknown command\r\n"
		}

		conn.Write([]byte(response))
	}
}
