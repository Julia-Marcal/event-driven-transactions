package rabbitmq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Julia-Marcal/event-driven-transactions/internal/core/model"
	"github.com/Julia-Marcal/event-driven-transactions/internal/core/repository"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
	logger   *log.Logger
}
type NoopPublisher struct {
	logger *log.Logger
}

func (n *NoopPublisher) Publish(ctx context.Context, event model.TransactionEvent) error {
	if n != nil && n.logger != nil {
		n.logger.Printf("noop publish: transaction_id=%s account=%s amount=%v", event.ID.String(), event.AccountID, event.Amount)
	}
	return nil
}

func (n *NoopPublisher) Close() error { return nil }

func (r *RabbitMQPublisher) Publish(ctx context.Context, event model.TransactionEvent) error {
	if r == nil || r.channel == nil {
		return nil
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return r.channel.PublishWithContext(ctx,
		r.exchange, // exchange
		"",         // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (r *RabbitMQPublisher) Close() error {
	if r == nil {
		return nil
	}
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
	return nil
}

func NewRabbitMQPublisher(url string, logger *log.Logger) (repository.Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	exchange := "transactions"
	if err := ch.ExchangeDeclare(
		exchange,
		"fanout",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // args
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	closeCh := make(chan *amqp.Error, 1)
	ch.NotifyClose(closeCh)
	go func() {
		for e := range closeCh {
			if logger != nil {
				logger.Printf("rabbitmq channel closed: %v", e)
			}
		}
	}()

	return &RabbitMQPublisher{conn: conn, channel: ch, exchange: exchange, logger: logger}, nil
}

var _ repository.Publisher = (*NoopPublisher)(nil)
var _ repository.Publisher = (*RabbitMQPublisher)(nil)
