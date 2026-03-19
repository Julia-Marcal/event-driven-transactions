package server

import (
	"context"

	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Julia-Marcal/event-driven-transactions/internal/infrastructure/config"
	"github.com/Julia-Marcal/event-driven-transactions/internal/infrastructure/mongodb"
	rabbit "github.com/Julia-Marcal/event-driven-transactions/internal/infrastructure/rabbitmq"
)

func Start(ctx context.Context) *log.Logger {
	logger := buildLogger()

	cfg := config.Load()

	cleanup, err := InitMongo(ctx)
	if err != nil {
		slog.Debug("failed to initialize mongodb", "error", err)
		cleanup = nil
	}

	consumerCfg := rabbit.ConsumerConfig{
		AmqpURL:    cfg.RabbitMQURL,
		Exchange:   "transactions",
		QueueName:  "transactions_worker",
		Kind:       "topic",
		RoutingKey: "transactions.*",
		Logger:     logger,
	}

	consumer, err := rabbit.StartConsumer(consumerCfg)
	if err != nil {
		logger.Fatalf("failed to start topic consumer: %v", err)
	}

	defer func() {
		if consumer != nil {
			_ = consumer.Close()
		}
	}()

	if cleanup != nil {
		cleanup()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	return logger
}

func buildLogger() *log.Logger {
	return log.New(os.Stdout, "api: ", log.LstdFlags|log.Lmsgprefix)
}

func InitMongo(parentCtx context.Context) (func(), error) {
	connCtx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
	client, err := mongodb.Connect(connCtx)
	cancel()
	if err != nil {
		return nil, err
	}

	cleanup := func() {
		if err := mongodb.Disconnect(context.Background(), client); err != nil {
			slog.Debug("error disconnecting mongodb", "error", err)
		}
	}

	return cleanup, nil
}
