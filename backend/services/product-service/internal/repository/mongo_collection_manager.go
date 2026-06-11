package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoCollectionManager struct {
	database     *mongo.Database
	databaseName string
	logger       *slog.Logger
	definitions  []MongoCollectionDefinition
}

func NewMongoCollectionManager(database *mongo.Database, logger *slog.Logger) (*MongoCollectionManager, error) {
	if database == nil {
		return nil, fmt.Errorf("mongo database is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &MongoCollectionManager{
		database:     database,
		databaseName: database.Name(),
		logger:       logger,
		definitions:  ProductMongoCollectionDefinitions(),
	}, nil
}

func (m *MongoCollectionManager) EnsureCollections(ctx context.Context) (CollectionSetupResult, error) {
	if err := ctx.Err(); err != nil {
		return CollectionSetupResult{}, err
	}

	existingNames, err := m.database.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return CollectionSetupResult{}, fmt.Errorf("list mongo collections for %s: %w", m.databaseName, err)
	}
	existing := make(map[string]struct{}, len(existingNames))
	for _, name := range existingNames {
		existing[name] = struct{}{}
	}

	result := CollectionSetupResult{
		Database:    m.databaseName,
		Collections: make([]CollectionSetupItem, 0, len(m.definitions)),
	}

	for _, definition := range m.definitions {
		item, err := m.ensureCollection(ctx, definition, existing)
		if err != nil {
			return CollectionSetupResult{}, err
		}
		result.Collections = append(result.Collections, item)
	}
	return result, nil
}

func (m *MongoCollectionManager) DescribeCollections() []CollectionDescription {
	descriptions := make([]CollectionDescription, 0, len(m.definitions))
	for _, definition := range m.definitions {
		descriptions = append(descriptions, CollectionDescription{
			Name:         definition.Name,
			IndexNames:   definition.IndexNames(),
			HasValidator: len(definition.Validator) > 0,
		})
	}
	return descriptions
}

func (m *MongoCollectionManager) ensureCollection(
	ctx context.Context,
	definition MongoCollectionDefinition,
	existing map[string]struct{},
) (CollectionSetupItem, error) {
	item := CollectionSetupItem{
		Name:       definition.Name,
		IndexNames: definition.IndexNames(),
	}

	if _, exists := existing[definition.Name]; exists {
		if err := m.applyValidator(ctx, definition); err != nil {
			return item, err
		}
		item.ValidatorApplied = true
	} else {
		created, err := m.createCollection(ctx, definition)
		if err != nil {
			return item, err
		}
		item.Created = created
		item.ValidatorApplied = true
		existing[definition.Name] = struct{}{}
	}

	if len(definition.Indexes) > 0 {
		if _, err := m.database.Collection(definition.Name).Indexes().CreateMany(ctx, definition.IndexModels()); err != nil {
			return item, fmt.Errorf("create indexes for %s.%s: %w", m.databaseName, definition.Name, err)
		}
	}

	m.logger.Info(
		"mongo collection schema ensured",
		"database", m.databaseName,
		"collection", definition.Name,
		"created", item.Created,
		"index_count", len(item.IndexNames),
	)
	return item, nil
}

func (m *MongoCollectionManager) createCollection(ctx context.Context, definition MongoCollectionDefinition) (bool, error) {
	err := m.database.CreateCollection(
		ctx,
		definition.Name,
		options.CreateCollection().
			SetValidator(definition.Validator).
			SetValidationLevel("moderate").
			SetValidationAction("error"),
	)
	if err == nil {
		return true, nil
	}
	if isNamespaceExists(err) {
		if applyErr := m.applyValidator(ctx, definition); applyErr != nil {
			return false, applyErr
		}
		return false, nil
	}
	return false, fmt.Errorf("create mongo collection %s.%s: %w", m.databaseName, definition.Name, err)
}

func (m *MongoCollectionManager) applyValidator(ctx context.Context, definition MongoCollectionDefinition) error {
	command := bson.D{
		e("collMod", definition.Name),
		e("validator", definition.Validator),
		e("validationLevel", "moderate"),
		e("validationAction", "error"),
	}
	if err := m.database.RunCommand(ctx, command).Err(); err != nil {
		return fmt.Errorf("apply validator for %s.%s: %w", m.databaseName, definition.Name, err)
	}
	return nil
}

func isNamespaceExists(err error) bool {
	var commandErr mongo.CommandError
	if errors.As(err, &commandErr) {
		return commandErr.HasErrorCode(48) || commandErr.Name == "NamespaceExists"
	}
	return false
}
