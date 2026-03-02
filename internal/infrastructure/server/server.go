package server

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Julia-Marcal/event-driven-transactions/internal/core/service"
	"github.com/Julia-Marcal/event-driven-transactions/internal/dto"
	"github.com/Julia-Marcal/event-driven-transactions/internal/infrastructure/config"
	rabbit "github.com/Julia-Marcal/event-driven-transactions/internal/infrastructure/rabbitmq"
)

func Start() *log.Logger {
	logger := buildLogger()

	cfg := config.Load()

	consumerCfg := rabbit.ConsumerConfig{
		AmqpURL:    cfg.RabbitMQURL,
		Exchange:   "transactions",
		QueueName:  "transactions_worker",
		Kind:       "topic",
		RoutingKey: "transactions.*",
		Logger:     logger,
	}

	ts := service.TransactionService{Logger: logger}
	consumer, err := rabbit.StartConsumer(consumerCfg, func(body []byte) error {
		logger.Printf("[CONSUMER] received message: %s", string(body))
		var req dto.CreateTransactionRequest
		if err := json.Unmarshal(body, &req); err != nil {
			logger.Printf("[CONSUMER] failed to unmarshal message: %v", err)
			return err
		}
		_, err := ts.CreateAndPublish(nil, req)
		if err != nil {
			logger.Printf("[CONSUMER] failed to process message: %v", err)
			return err
		}
		logger.Printf("[CONSUMER] successfully processed message for account: %s", req.AccountID)
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	return logger
}

func buildLogger() *log.Logger {
	return log.New(os.Stdout, "api: ", log.LstdFlags|log.Lmsgprefix)
}
