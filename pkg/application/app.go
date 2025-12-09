package application

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"redis-lite/pkg/cfg"
	"syscall"
)

type App interface {
	Run() error
}

type app struct {
	ctx    context.Context
	config *cfg.Config
}

func NewApp() App {
	config := cfg.NewConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	return &app{
		ctx:    ctx,
		config: config,
	}

}

func (app *app) Run() error {
	addr := fmt.Sprintf("%s:%s", app.config.Host, app.config.Port)
	l, err := net.Listen(string(app.config.ServerType), addr)
	if err != nil {
		panic(err)
	}

	defer l.Close()

	host, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		panic(err)
	}

	fmt.Printf("Listening on host: %s, port: %s\n", host, port)

	for {
		// Listen for an incoming connection
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		// Handle connections in a new goroutine
		go func(conn net.Conn) {
			buf := make([]byte, 1022)
			len, err := conn.Read(buf)
			if err != nil {
				fmt.Printf("Error reading: %#v\n", err)
				return
			}
			fmt.Printf("Message received: %s\n", string(buf[:len]))

			conn.Write([]byte("Message received.\n"))
			conn.Close()
		}(conn)
	}
}
