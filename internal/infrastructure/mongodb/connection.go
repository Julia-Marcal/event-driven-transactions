package mongodb

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect(ctx context.Context) (*mongo.Client, error) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		user := os.Getenv("MONGO_INITDB_ROOT_USERNAME")
		pass := os.Getenv("MONGO_INITDB_ROOT_PASSWORD")
		db := os.Getenv("MONGO_INITDB_DATABASE")
		host := os.Getenv("MONGO_HOST")
		if host == "" {
			host = "mongodb:27017"
		}

		if user != "" && pass != "" {
			cu := url.QueryEscape(user)
			cp := url.QueryEscape(pass)
			if db == "" {
				db = "admin"
			}
			uri = fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=admin", cu, cp, host, db)
		} else {
			uri = "mongodb://localhost:27017"
		}
	}

	clientOpts := options.Client().ApplyURI(uri)

	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(cctx, clientOpts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(cctx, nil); err != nil {
		_ = client.Disconnect(cctx)
		return nil, err
	}

	return client, nil
}

func Disconnect(ctx context.Context, client *mongo.Client) error {
	if client == nil {
		return nil
	}
	dctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return client.Disconnect(dctx)
}

// WithClient centralizes Connect
func WithClient(ctx context.Context, fn func(*mongo.Client) error) error {
	client, err := Connect(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = Disconnect(context.Background(), client)
	}()

	return fn(client)
}
