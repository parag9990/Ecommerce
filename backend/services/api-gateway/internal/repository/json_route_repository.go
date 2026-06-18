package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"ecommerce/api-gateway/internal/domain"
)

type JSONRouteRepository struct {
	path string
}

func NewJSONRouteRepository(path string) *JSONRouteRepository {
	return &JSONRouteRepository{path: path}
}

func (r *JSONRouteRepository) LoadContract(ctx context.Context) (domain.Contract, error) {
	if err := ctx.Err(); err != nil {
		return domain.Contract{}, err
	}

	data, err := os.ReadFile(r.path)
	if err != nil {
		return domain.Contract{}, fmt.Errorf("read api route contract %q: %w", r.path, err)
	}
	if err := ctx.Err(); err != nil {
		return domain.Contract{}, err
	}

	var contract domain.Contract
	if err := json.Unmarshal(data, &contract); err != nil {
		return domain.Contract{}, fmt.Errorf("decode api route contract %q: %w", r.path, err)
	}
	return contract, nil
}
