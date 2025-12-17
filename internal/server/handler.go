package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"redis-lite/pkg/core"
	"strings"
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

		rawMessage := strings.TrimSpace(message)
		args := strings.Fields(rawMessage)
		if len(args) == 0 {
			continue
		}

		command := strings.ToUpper(args[0])

		slog.Info("command: " + command)

		if command == "SUBSCRIBE" {
			s.handleSubscribe(conn, args)
			return
		}

		response := core.Eval(s.DB, args)

		conn.Write(response)

		if core.IsWriteOp(args[0]) && len(response) > 0 && response[0] != '-' {
			s.Aof.Write(rawMessage)
		}
	}
}

func (s *Server) handleSubscribe(conn net.Conn, args []string) {
	// syntax => subscribe topic
	if len(args) < 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'subscribe' command\r\n"))
		return
	}

	topic := args[1]

	// buffered chan for this specific client to receive messages
	msgChan := make(chan string, 100)

	s.DB.PubSub.Subscribe(topic, msgChan)
	defer s.DB.PubSub.UnSubscribe(topic, msgChan)

	// *{LENGTH_OF_ARRAY}
	// ${LENGTH_OF_STRING}
	// :{LENGTH_OF_SUBSCRIBED_CHANNELS}
	confirm := fmt.Sprintf("*3\r\n$9\r\nsubscribe\r\n$%d\r\n%s\r\n:1\r\n", len(topic), topic)
	conn.Write([]byte(confirm))

	// block here and push messages to the client as they arrive on the channel
	for msg := range msgChan {
		response := fmt.Sprintf("*3\r\n$7\r\nmessage\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
			len(topic), topic, len(msg), msg)

		_, err := conn.Write([]byte(response))
		if err != nil {
			slog.Info("Subscriber disconnected", "topic", topic)
			return
		}
	}
}
