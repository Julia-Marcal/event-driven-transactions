// Package server wires dependencies and manages HTTP server lifecycle.
package server

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Julia-Marcal/event-driven-transactions/internal/core/repository"
	"github.com/Julia-Marcal/event-driven-transactions/internal/core/service"
	"github.com/Julia-Marcal/event-driven-transactions/internal/handler"
	"github.com/Julia-Marcal/event-driven-transactions/internal/infrastructure/config"
	rabbit "github.com/Julia-Marcal/event-driven-transactions/internal/infrastructure/rabbitmq"
	"github.com/Julia-Marcal/event-driven-transactions/internal/router"
)

var (
	_rabbitmqOnce   sync.Once
	rabbitPublisher repository.Publisher
	rabbitInitErr   error
)

func Start() (*http.Server, *log.Logger) {
	logger := buildLogger()

	publisher, err := NewRabbitMQ(logger)
	if err != nil {
		logger.Fatalf("failed to create rabbitmq publisher: %v", err)
	}

	defer func() {
		if publisher != nil {
			_ = publisher.Close()
		}
	}()

	svc := service.NewTransactionService(publisher, logger)
	handler.Init(svc)
	srv := router.Router()

	startHTTPServer(srv, logger)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	return srv, logger
}

func buildLogger() *log.Logger {
	return log.New(os.Stdout, "api: ", log.LstdFlags|log.Lmsgprefix)
}

func startHTTPServer(srv *http.Server, logger *log.Logger) {
	go func() {
		logger.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %v", err)
		}
	}()
}

// NewRabbitMQ returns a singleton publisher instance. It stores initialization
// error and publisher in package-level variables so multiple callers get the
// same instance.
func NewRabbitMQ(logger *log.Logger) (repository.Publisher, error) {
	cfg := config.Load()
	_rabbitmqOnce.Do(func() {
		rabbitPublisher, rabbitInitErr = rabbit.NewRabbitMQPublisher(cfg.RabbitMQURL, logger)
	})
	return rabbitPublisher, rabbitInitErr
}
