package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoCollectionManager struct {
	db     *mongo.Database
	logger *slog.Logger
}

func NewMongoCollectionManager(db *mongo.Database, logger *slog.Logger) (*MongoCollectionManager, error) {
	if db == nil {
		return nil, errors.New("mongo database is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &MongoCollectionManager{db: db, logger: logger}, nil
}

func (m *MongoCollectionManager) Ping(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return m.db.Client().Ping(ctx, nil)
}

func (m *MongoCollectionManager) EnsureCartCollections(ctx context.Context) (domain.CollectionReport, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	spec := domain.CartCollectionSpec()
	report := domain.CollectionReport{
		DatabaseName:   m.db.Name(),
		CollectionName: spec.CollectionName,
	}

	exists, err := m.collectionExists(ctx, spec.CollectionName)
	if err != nil {
		return report, fmt.Errorf("list mongo collections: %w", err)
	}

	if exists {
		if err := m.updateValidator(ctx, spec.CollectionName); err != nil {
			return report, fmt.Errorf("update carts validator: %w", err)
		}
		report.ValidatorUpdated = true
	} else {
		if err := m.db.CreateCollection(ctx, spec.CollectionName, createCollectionOptions()); err != nil {
			return report, fmt.Errorf("create carts collection: %w", err)
		}
		report.CollectionCreated = true
	}

	created, err := m.db.Collection(spec.CollectionName).Indexes().CreateMany(ctx, cartIndexModels())
	if err != nil {
		return report, fmt.Errorf("create cart indexes: %w", err)
	}
	report.IndexesEnsured = created
	m.logger.Info(
		"cart.mongo.collections_ensured",
		slog.String("database", report.DatabaseName),
		slog.String("collection", report.CollectionName),
		slog.Bool("collection_created", report.CollectionCreated),
		slog.Bool("validator_updated", report.ValidatorUpdated),
		slog.Int("index_count", len(report.IndexesEnsured)),
	)
	return report, nil
}

func (m *MongoCollectionManager) collectionExists(ctx context.Context, collectionName string) (bool, error) {
	names, err := m.db.ListCollectionNames(ctx, bson.D{{Key: "name", Value: collectionName}})
	if err != nil {
		return false, err
	}
	return len(names) > 0, nil
}

func (m *MongoCollectionManager) updateValidator(ctx context.Context, collectionName string) error {
	cmd := bson.D{
		{Key: "collMod", Value: collectionName},
		{Key: "validator", Value: cartValidator()},
		{Key: "validationLevel", Value: "strict"},
		{Key: "validationAction", Value: "error"},
	}
	return m.db.RunCommand(ctx, cmd).Err()
}
