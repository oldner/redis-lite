package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"redis-lite/pkg/utils"
	"strconv"
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
		case "HSET":
			// syntax HSET key field value expiration
			if len(args) < 4 {
				response = "-ERR wrong number of arguments for `hset` command\r\n"
			} else {
				key := args[1]
				field := args[2]
				value := args[3]
				expiry := args[4]

				var ttl time.Duration
				if len(args) > 4 {
					ttl = utils.ParseDuration(expiry)
				}

				created, err := s.DB.HSet(key, field, value, ttl)
				if err != nil {
					response = "-ERR" + err.Error() + "\r\n"
				} else {
					if created {
						response = ":1\r\n"
					} else {
						response = ":0\r\n"
					}
				}
			}
		case "HGET":
			// syntax HGET key field
			if len(args) < 3 {
				response = "-ERR wrong number of arguments for `hget` command\r\n"
			} else {
				key := args[1]
				field := args[2]

				val, exists := s.DB.HGet(key, field)
				if !exists {
					response = "$-1\r\n"
				} else {
					response = fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
				}
			}
		case "LPUSH":
			// syntax LPUSH key value expiration
			if len(args) < 3 {
				response = "-ERR wrong number of arguments for `hget` command\r\n"
			} else {
				key := args[1]
				value := args[2]
				expiry := args[3]

				var ttl time.Duration
				if len(args) > 3 {
					ttl = utils.ParseDuration(expiry)
				}

				val, err := s.DB.LPush(key, value, ttl)
				if err != nil {
					response = "-ERR" + err.Error() + "\r\n"
				} else {
					response = fmt.Sprintf("$%d\r\n", val)
				}
			}
		case "LPOP":
			// syntax LPOP key
			if len(args) < 2 {
				response = "-ERR wrong number of arguments for `get` command\r\n"
			} else {
				key := args[1]

				val, exists := s.DB.LPop(key)
				if !exists {
					response = "$-1\r\n" // redis standart for not found
				} else {
					response = fmt.Sprintf("$%d\r\n%v\r\n", len(fmt.Sprint(val)), val)
				}
			}
		case "LRANGE":
			// syntax LRANGE key start stop
			if len(args) < 4 {
				response = "-ERR wrong number of arguments for `hget` command\r\n"
			} else {
				key := args[1]
				start, err1 := strconv.Atoi(args[2])
				stop, err2 := strconv.Atoi(args[3])
				if err1 != nil || err2 != nil {
					response = "-ERR value is not an integer or out of range\r\n"
				}

				list, ok := s.DB.LRange(key, start, stop)
				if !ok {
					response = "*0\r\n" // empty array
				} else {
					var sb strings.Builder
					sb.WriteString(fmt.Sprintf("*%d\r\n", len(list)))
					response = sb.String()
				}
			}
		case "DEL":
			// syntax: DEL key
			if len(args) < 2 {
				response = "-ERR wrong number of arguments for `del` command\r\n"
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
