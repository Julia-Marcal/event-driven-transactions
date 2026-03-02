package main

import (
	"github.com/Julia-Marcal/event-driven-transactions/internal/infrastructure/server"
)

func main() {
	logger := server.Start()

	logger.Println("server exiting")
}
