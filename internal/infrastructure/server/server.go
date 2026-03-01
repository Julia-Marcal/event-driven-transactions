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

	cfg := config.Load()

	consumerCfg := rabbit.ConsumerConfig{
		AmqpURL:    cfg.RabbitMQURL,
		Exchange:   "transactions",
		QueueName:  "transactions_worker",
		Kind:       "topic",
		RoutingKey: "transactions.*",
		Logger:     logger,
	}
	consumer, err := rabbit.StartConsumer(consumerCfg, func(body []byte) error {
		logger.Printf("[CONSUMER] received message: %s", string(body))
		return nil
	})
	if err != nil {
		logger.Fatalf("failed to start topic consumer: %v", err)
	}

	defer func() {
		if consumer != nil {
			_ = consumer.Close()
		}
	}()

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

func NewRabbitMQ(logger *log.Logger) (repository.Publisher, error) {
	cfg := config.Load()
	_rabbitmqOnce.Do(func() {
		rabbitPublisher, rabbitInitErr = rabbit.NewRabbitMQPublisher(cfg.RabbitMQURL, logger)
	})
	return rabbitPublisher, rabbitInitErr
}
