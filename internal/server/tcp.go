package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"redis-lite/pkg/aof"
	"redis-lite/pkg/database"
)

type Server struct {
	ConfigAddr string
	DB         *database.Store
	Aof        *aof.Aof
}

func NewServer(host, port string, db *database.Store, aof *aof.Aof) *Server {
	addr := fmt.Sprintf("%s:%s", host, port)
	return &Server{
		ConfigAddr: addr,
		DB:         db,
		Aof:        aof,
	}
}

func (s *Server) Run(ctx context.Context) error {
	lc := net.ListenConfig{}
	l, err := lc.Listen(ctx, "tcp", s.ConfigAddr)
	if err != nil {
		panic(err)
	}

	defer l.Close()

	fmt.Printf("Listening off host: %s\n", s.ConfigAddr)

	go func() {
		<-ctx.Done()
		slog.Info("Shutting down server..")
		l.Close()
	}()

	for {
		// Listen for an incoming connection
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		// Handle connections in a new goroutine
		go s.handleConnection(ctx, conn)
	}
}
