package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/Julia-Marcal/event-driven-transactions/internal/core/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrDuplicateIdempotencyKey = errors.New("duplicate idempotency key")

func InsertTransaction(ctx context.Context, tx model.Transaction, idempotencyKey string) error {
	return WithClient(ctx, func(client *mongo.Client) error {
		db := client.Database("ledger")

		if idempotencyKey != "" {
			_, err := db.Collection("idempotency_keys").InsertOne(ctx, bson.M{
				"_id":       idempotencyKey,
				"status":    "processed",
				"createdAt": time.Now(),
			})
			if err != nil {
				if mongo.IsDuplicateKeyError(err) {
					return ErrDuplicateIdempotencyKey
				}
				return err
			}
		}

		_, err := db.Collection("transactions").InsertOne(ctx, bson.M{
			"_id":       tx.ID.String(),
			"accountId": tx.AccountID,
			"type":      string(tx.Type),
			"amount":    int64(tx.Amount),
			"createdAt": tx.CreatedAt,
		})
		return err
	})
}
