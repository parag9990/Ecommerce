package repository

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func NewMongoDatabase(ctx context.Context, uri string, databaseName string) (*mongo.Client, *mongo.Database, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(uri) == "" {
		return nil, nil, fmt.Errorf("mongo uri is required")
	}
	if strings.TrimSpace(databaseName) == "" {
		return nil, nil, fmt.Errorf("mongo database name is required")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, fmt.Errorf("connect mongo client: %w", err)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)
		return nil, nil, fmt.Errorf("ping mongo database %s: %w", databaseName, err)
	}
	return client, client.Database(databaseName), nil
}
