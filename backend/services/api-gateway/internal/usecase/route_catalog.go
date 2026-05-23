package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"ecommerce/api-gateway/internal/domain"
)

var ErrRouteNotFound = errors.New("route not found")

type RouteRepository interface {
	LoadContract(ctx context.Context) (domain.Contract, error)
}

type RouteCatalog interface {
	Load(ctx context.Context) error
	ListRoutes(ctx context.Context) ([]domain.RouteDefinition, error)
	FindByID(ctx context.Context, id string) (domain.RouteDefinition, error)
	FindByKey(ctx context.Context, method domain.HTTPMethod, path string) (domain.RouteDefinition, error)
	Groups(ctx context.Context) (map[string][]domain.RouteDefinition, error)
	Schemas(ctx context.Context) (map[string]domain.Schema, error)
}

type RouteCatalogService struct {
	repo     RouteRepository
	basePath string

	mu      sync.RWMutex
	loaded  bool
	routes  []domain.RouteDefinition
	byID    map[string]domain.RouteDefinition
	byKey   map[string]domain.RouteDefinition
	groups  map[string][]domain.RouteDefinition
	schemas map[string]domain.Schema
}

func NewRouteCatalogService(repo RouteRepository, basePath string) *RouteCatalogService {
	return &RouteCatalogService{
		repo:     repo,
		basePath: basePath,
	}
}

func (s *RouteCatalogService) Load(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	contract, err := s.repo.LoadContract(ctx)
	if err != nil {
		return err
	}
	routes, byID, byKey, groups, err := buildIndexes(contract, s.basePath)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.routes = routes
	s.byID = byID
	s.byKey = byKey
	s.groups = groups
	s.schemas = cloneSchemas(contract.Schemas)
	s.loaded = true
	return nil
}

func (s *RouteCatalogService) ListRoutes(ctx context.Context) ([]domain.RouteDefinition, error) {
	if err := s.ensureLoaded(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneRoutes(s.routes), nil
}

func (s *RouteCatalogService) FindByID(ctx context.Context, id string) (domain.RouteDefinition, error) {
	if err := s.ensureLoaded(ctx); err != nil {
		return domain.RouteDefinition{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	route, ok := s.byID[id]
	if !ok {
		return domain.RouteDefinition{}, fmt.Errorf("%w: id %s", ErrRouteNotFound, id)
	}
	return route, nil
}

func (s *RouteCatalogService) FindByKey(ctx context.Context, method domain.HTTPMethod, path string) (domain.RouteDefinition, error) {
	if err := s.ensureLoaded(ctx); err != nil {
		return domain.RouteDefinition{}, err
	}
	key := string(method) + " " + path
	s.mu.RLock()
	defer s.mu.RUnlock()
	route, ok := s.byKey[key]
	if !ok {
		return domain.RouteDefinition{}, fmt.Errorf("%w: key %s", ErrRouteNotFound, key)
	}
	return route, nil
}

func (s *RouteCatalogService) Groups(ctx context.Context) (map[string][]domain.RouteDefinition, error) {
	if err := s.ensureLoaded(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string][]domain.RouteDefinition, len(s.groups))
	for name, routes := range s.groups {
		out[name] = cloneRoutes(routes)
	}
	return out, nil
}

func (s *RouteCatalogService) Schemas(ctx context.Context) (map[string]domain.Schema, error) {
	if err := s.ensureLoaded(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneSchemas(s.schemas), nil
}

func (s *RouteCatalogService) ensureLoaded(ctx context.Context) error {
	s.mu.RLock()
	loaded := s.loaded
	s.mu.RUnlock()
	if loaded {
		return nil
	}
	return s.Load(ctx)
}

func buildIndexes(contract domain.Contract, basePath string) ([]domain.RouteDefinition, map[string]domain.RouteDefinition, map[string]domain.RouteDefinition, map[string][]domain.RouteDefinition, error) {
	var errs []error
	if contract.Project == "" {
		errs = append(errs, domain.ValidationError{Field: "project", Message: "project is required"})
	}
	if contract.Version == "" {
		errs = append(errs, domain.ValidationError{Field: "version", Message: "version is required"})
	}
	if len(contract.RestEndpoints) == 0 {
		errs = append(errs, domain.ValidationError{Field: "rest_endpoints", Message: "at least one route is required"})
	}

	routes := cloneRoutes(contract.RestEndpoints)
	byID := make(map[string]domain.RouteDefinition, len(routes))
	byKey := make(map[string]domain.RouteDefinition, len(routes))
	groups := make(map[string][]domain.RouteDefinition)

	for _, route := range routes {
		if err := route.Validate(basePath); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", route.ID, err))
			continue
		}
		if existing, ok := byID[route.ID]; ok {
			errs = append(errs, domain.ValidationError{Field: "id", Message: fmt.Sprintf("duplicate route id %q also used by %s", route.ID, existing.Key())})
		}
		if existing, ok := byKey[route.Key()]; ok {
			errs = append(errs, domain.ValidationError{Field: "path", Message: fmt.Sprintf("duplicate route key %q also used by %s", route.Key(), existing.ID)})
		}
		byID[route.ID] = route
		byKey[route.Key()] = route
		group := route.Group(basePath)
		groups[group] = append(groups[group], route)
	}

	if len(errs) > 0 {
		return nil, nil, nil, nil, domain.ContractValidationError{Errors: errs}
	}
	return routes, byID, byKey, groups, nil
}

func cloneRoutes(routes []domain.RouteDefinition) []domain.RouteDefinition {
	out := make([]domain.RouteDefinition, len(routes))
	copy(out, routes)
	return out
}

func cloneSchemas(schemas map[string]domain.Schema) map[string]domain.Schema {
	if schemas == nil {
		return nil
	}
	out := make(map[string]domain.Schema, len(schemas))
	for name, schema := range schemas {
		out[name] = cloneSchema(schema)
	}
	return out
}

func cloneSchema(schema domain.Schema) domain.Schema {
	clone := schema
	if schema.Required != nil {
		clone.Required = append([]string(nil), schema.Required...)
	}
	if schema.Properties != nil {
		clone.Properties = make(map[string]domain.Schema, len(schema.Properties))
		for name, property := range schema.Properties {
			clone.Properties[name] = cloneSchema(property)
		}
	}
	if schema.Items != nil {
		item := cloneSchema(*schema.Items)
		clone.Items = &item
	}
	if schema.Enum != nil {
		clone.Enum = append([]string(nil), schema.Enum...)
	}
	if schema.AllOf != nil {
		clone.AllOf = make([]domain.Schema, len(schema.AllOf))
		for i, item := range schema.AllOf {
			clone.AllOf[i] = cloneSchema(item)
		}
	}
	return clone
}
