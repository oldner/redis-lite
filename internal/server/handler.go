package server

import (
	"bufio"
	"context"
	"io"
	"log/slog"
	"net"
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
				slog.Error("Read error: ", err)
			}
			return
		}

		cmd := strings.TrimSpace(message)
		if len(cmd) == 0 {
			continue
		}

		slog.Info("Received command: " + cmd)

		response := ""
		switch cmd {
		case "PING":
			response = "+PONG\r\n"
		default:
			response = "-ERR unknown command\r\n"
		}

		conn.Write([]byte(response))
	}
}
