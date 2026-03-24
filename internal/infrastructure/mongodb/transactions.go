package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/Julia-Marcal/event-driven-transactions/internal/core/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrDuplicateIdempotencyKey = errors.New("duplicate idempotency key")
	ErrAccountNotFound         = errors.New("account not found")
	ErrInsufficientFunds       = errors.New("insufficient funds")
)

func ProcessTransaction(ctx context.Context, tx model.Transaction, idempotencyKey string) error {
	return WithClient(ctx, func(client *mongo.Client) error {
		db := client.Database("ledger")

		session, err := client.StartSession()
		if err != nil {
			return err
		}
		defer session.EndSession(ctx)

		return mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
			if err := session.StartTransaction(); err != nil {
				return err
			}

			if idempotencyKey != "" {
				_, err := db.Collection("idempotency_keys").InsertOne(sc, bson.M{
					"_id":       idempotencyKey,
					"status":    "processed",
					"createdAt": time.Now(),
				})
				if err != nil {
					_ = session.AbortTransaction(sc)
					if mongo.IsDuplicateKeyError(err) {
						return ErrDuplicateIdempotencyKey
					}
					return err
				}
			}

			var account model.Account
			err = db.Collection("accounts").FindOne(sc, bson.M{"_id": tx.AccountID}).Decode(&account)
			if err != nil {
				_ = session.AbortTransaction(sc)
				if errors.Is(err, mongo.ErrNoDocuments) {
					return ErrAccountNotFound
				}
				return err
			}

			if tx.Type == model.TransactionTypeDebit && account.Balance < tx.Amount {
				_ = session.AbortTransaction(sc)
				return ErrInsufficientFunds
			}

			newAmount := tx.Amount
			if tx.Type == model.TransactionTypeDebit {
				newAmount = -tx.Amount
			}

			_, err = db.Collection("accounts").UpdateOne(sc,
				bson.M{"_id": tx.AccountID},
				mongo.Pipeline{bson.D{{Key: "$set", Value: bson.M{
					"balance":   bson.M{"$add": bson.A{"$balance", newAmount}},
					"createdAt": bson.M{"$ifNull": bson.A{"$createdAt", "$$NOW"}},
				}}}},
			)
			if err != nil {
				_ = session.AbortTransaction(sc)
				return err
			}

			_, err = db.Collection("transactions").InsertOne(sc, bson.M{
				"_id":       tx.ID.String(),
				"accountId": tx.AccountID,
				"type":      string(tx.Type),
				"amount":    int64(tx.Amount),
				"createdAt": tx.CreatedAt,
			})
			if err != nil {
				_ = session.AbortTransaction(sc)
				return err
			}

			return session.CommitTransaction(sc)
		})
	})
}
