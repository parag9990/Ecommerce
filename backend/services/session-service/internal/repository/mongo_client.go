package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	defaultMongoDatabase       = "session_db"
	defaultMongoConnectTimeout = 5 * time.Second
)

type MongoConfig struct {
	URI            string
	Database       string
	ConnectTimeout time.Duration
}

func NewMongoClient(ctx context.Context, cfg MongoConfig) (*mongo.Client, error) {
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	connectCtx := ctx
	cancel := func() {}
	if _, ok := ctx.Deadline(); !ok {
		connectCtx, cancel = context.WithTimeout(ctx, cfg.ConnectTimeout)
	}
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}
	if err := client.Ping(connectCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("ping mongo: %w", err)
	}
	return client, nil
}

func MongoDatabase(client *mongo.Client, database string) (*mongo.Database, error) {
	if client == nil {
		return nil, errors.New("mongo client is required")
	}
	database = strings.TrimSpace(database)
	if database == "" {
		database = defaultMongoDatabase
	}
	return client.Database(database), nil
}

func (c MongoConfig) Validate() error {
	if strings.TrimSpace(c.URI) == "" {
		return errors.New("SESSION_MONGO_URI cannot be empty")
	}
	if strings.TrimSpace(c.Database) == "" {
		return errors.New("SESSION_MONGO_DATABASE cannot be empty")
	}
	if c.ConnectTimeout <= 0 {
		return errors.New("SESSION_MONGO_CONNECT_TIMEOUT must be greater than zero")
	}
	return nil
}

func (c MongoConfig) withDefaults() MongoConfig {
	if strings.TrimSpace(c.Database) == "" {
		c.Database = defaultMongoDatabase
	}
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = defaultMongoConnectTimeout
	}
	return c
}
