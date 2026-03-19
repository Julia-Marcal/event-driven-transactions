package rabbitmq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Julia-Marcal/event-driven-transactions/internal/core/service"
	"github.com/Julia-Marcal/event-driven-transactions/internal/dto"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ConsumerConfig struct {
	AmqpURL    string
	Exchange   string
	QueueName  string
	Kind       string
	RoutingKey string
	Logger     *log.Logger
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	stopCh  chan struct{}
	doneCh  chan struct{}
	logger  *log.Logger
}

func StartConsumer(cfg ConsumerConfig) (*Consumer, error) {
	conn, err := connectRabbitMQ(cfg.AmqpURL)
	if err != nil {
		return nil, err
	}
	ch, err := createChannel(conn)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	if err := declareExchange(ch, cfg.Exchange, cfg.Kind); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	q, err := declareQueue(ch, cfg.QueueName)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	if err := bindQueueWithKey(ch, q.Name, cfg.Exchange, cfg.RoutingKey); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	consumer := &Consumer{
		conn:    conn,
		channel: ch,
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
		logger:  cfg.Logger,
	}

	go consumer.consumeLoop(msgs)

	return consumer, nil
}

func (c *Consumer) consumeLoop(msgs <-chan amqp.Delivery) {
	defer close(c.doneCh)
	ts := service.TransactionService{Logger: c.logger}
	ctx := context.Background()
	for {
		select {
		case d, ok := <-msgs:
			if !ok {
				return
			}
			err := c.Process(d.Body, &d, ts, ctx)
			if err != nil {
				c.logger.Printf("[CONSUMER] handler error: %v", err)
				_ = d.Nack(false, true)
				continue
			}
		case <-c.stopCh:
			return
		}
	}
}

func (c *Consumer) Process(body []byte, d *amqp.Delivery, ts service.TransactionService, ctx context.Context) error {
	c.logger.Printf("[CONSUMER] received message: %s", string(body))
	var req dto.CreateTransactionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		_ = d.Nack(false, false)
		c.logger.Printf("[CONSUMER] failed to unmarshal message: %v", err)
		return err
	}
	if err := req.Validate(); err != nil {
		_ = d.Nack(false, false)
		c.logger.Printf("[CONSUMER] validation error: %v", err)
		return err
	}
	_, err := ts.CreateAndPublish(ctx, req)
	if err != nil {
		_ = d.Nack(false, true)
		c.logger.Printf("[CONSUMER] failed to process message: %v", err)
		return err
	}
	c.logger.Printf("[CONSUMER] successfully processed message for account: %s", req.AccountID)
	_ = d.Ack(true)
	return nil
}

func (c *Consumer) Close() error {
	close(c.stopCh)
	<-c.doneCh
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
	return nil
}

func connectRabbitMQ(amqpURL string) (*amqp.Connection, error) {
	return amqp.Dial(amqpURL)
}

func createChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	return conn.Channel()
}

func declareExchange(ch *amqp.Channel, exchange string, kind string) error {
	return ch.ExchangeDeclare(
		exchange,
		kind,
		true,
		false,
		false,
		false,
		nil,
	)
}

func declareQueue(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
}

func bindQueueWithKey(ch *amqp.Channel, queueName, exchange, routingKey string) error {
	return ch.QueueBind(
		queueName,
		routingKey,
		exchange,
		false,
		nil,
	)
}

func startConsumer(ch *amqp.Channel, queueName string) (<-chan amqp.Delivery, error) {
	return ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
}

func logMessages(msgs <-chan amqp.Delivery, logger *log.Logger) {
	for d := range msgs {
		logger.Printf("[CONSUMER] received message: %s", string(d.Body))
	}
}
