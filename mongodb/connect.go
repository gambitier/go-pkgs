package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// Connect opens a MongoDB client and returns the named database after a primary ping.
func Connect(ctx context.Context, uri, database string, serverSelectionTimeout time.Duration) (*mongo.Client, *mongo.Database, error) {
	clientOpts := options.Client().ApplyURI(uri)
	if serverSelectionTimeout > 0 {
		clientOpts.SetServerSelectionTimeout(serverSelectionTimeout)
	}

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("mongo connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, nil, fmt.Errorf("mongo ping: %w", err)
	}

	return client, client.Database(database), nil
}
