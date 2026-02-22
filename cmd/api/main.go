package main

import (
	"context"
	"time"

	"github.com/Julia-Marcal/event-driven-transactions/internal/infrastructure/server"
)

func main() {
	srv, logger := server.Start()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("server forced to shutdown: %v", err)
	}
	logger.Println("server exiting")
}
