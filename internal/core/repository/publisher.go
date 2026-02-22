package repository

import (
	"context"

	"github.com/Julia-Marcal/event-driven-transactions/internal/core/model"
)

// Publisher is the interface used by the application to publish events.
type Publisher interface {
	Publish(ctx context.Context, event model.TransactionEvent) error
	Close() error
}
