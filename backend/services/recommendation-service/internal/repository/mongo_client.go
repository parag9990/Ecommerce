package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type MongoClientConfig struct {
	URI            string
	ConnectTimeout time.Duration
	PingTimeout    time.Duration
}

func NewMongoClient(ctx context.Context, cfg MongoClientConfig, logger *slog.Logger) (*mongo.Client, error) {
	if strings.TrimSpace(cfg.URI) == "" {
		return nil, errors.New("mongo uri is required")
	}
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = 10 * time.Second
	}
	if cfg.PingTimeout <= 0 {
		cfg.PingTimeout = 5 * time.Second
	}
	if logger == nil {
		logger = slog.Default()
	}

	connectCtx, cancelConnect := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancelConnect()

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.URI).SetConnectTimeout(cfg.ConnectTimeout))
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}

	pingCtx, cancelPing := context.WithTimeout(connectCtx, cfg.PingTimeout)
	defer cancelPing()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	logger.InfoContext(ctx, "recommendation.mongo.connected")
	return client, nil
}
