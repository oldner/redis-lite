package server

import (
	"bufio"
	"context"
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

		response := core.Eval(s.DB, args)

		conn.Write(response)

		if core.IsWriteOp(args[0]) && len(response) > 0 && response[0] != '-' {
			s.Aof.Write(rawMessage)
		}
	}
}
