package rabbitmq

import (
	"log"

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
	handler func([]byte) error
	logger  *log.Logger
}

func StartConsumer(cfg ConsumerConfig, handler func([]byte) error) (*Consumer, error) {
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
		handler: handler,
		logger:  cfg.Logger,
	}

	go consumer.consumeLoop(msgs)

	return consumer, nil
}

func (c *Consumer) consumeLoop(msgs <-chan amqp.Delivery) {
	defer close(c.doneCh)
	for {
		select {
		case d, ok := <-msgs:
			if !ok {
				return
			}
			if c.handler != nil {
				err := c.handler(d.Body)
				if err != nil {
					c.logger.Printf("[CONSUMER] handler error: %v", err)
					_ = d.Nack(false, true)
					continue
				}
			}
			_ = d.Ack(false)
		case <-c.stopCh:
			return
		}
	}
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
