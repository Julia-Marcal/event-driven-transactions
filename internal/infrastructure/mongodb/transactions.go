package mongodb

import (
	"context"

	"github.com/Julia-Marcal/event-driven-transactions/internal/core/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func InsertTransaction(ctx context.Context, tx model.Transaction) error {
	return WithClient(ctx, func(client *mongo.Client) error {
		coll := client.Database("ledger").Collection("transactions")
		doc := bson.M{
			"_id":       tx.ID.String(),
			"accountId": tx.AccountID,
			"type":      string(tx.Type),
			"amount":    int64(tx.Amount),
			"createdAt": tx.CreatedAt,
		}
		_, err := coll.InsertOne(ctx, doc)
		return err
	})
}
