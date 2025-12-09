package main

import (
	"redis-lite/pkg/application"
)

func main() {
	app := application.NewApp()

	app.Run()
}
