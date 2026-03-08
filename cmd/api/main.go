package main

import (
	"context"

	"github.com/Julia-Marcal/event-driven-transactions/internal/infrastructure/server"
)

func main() {
	ctx := context.Background()
	logger := server.Start(ctx)

	logger.Println("server exiting")
}
