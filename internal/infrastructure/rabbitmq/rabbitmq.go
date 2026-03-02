package rabbitmq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Julia-Marcal/event-driven-transactions/internal/core/model"
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
