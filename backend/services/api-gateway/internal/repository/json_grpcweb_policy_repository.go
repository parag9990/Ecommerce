package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"ecommerce/api-gateway/internal/domain"
)

type JSONGRPCWebPolicyRepository struct {
	path string
}

type grpcWebPolicyDocument struct {
	Policies []grpcWebPolicyDTO `json:"policies"`
}

type grpcWebPolicyDTO struct {
	FullMethod string   `json:"full_method"`
	Downstream string   `json:"downstream"`
	AuthMode   string   `json:"auth_mode"`
	Roles      []string `json:"roles,omitempty"`
	Timeout    string   `json:"timeout"`
}

func NewJSONGRPCWebPolicyRepository(path string) *JSONGRPCWebPolicyRepository {
	return &JSONGRPCWebPolicyRepository{path: path}
}

func (r *JSONGRPCWebPolicyRepository) LoadPolicies(ctx context.Context) ([]domain.GRPCWebMethodPolicy, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(r.path)
	if err != nil {
		return nil, fmt.Errorf("read grpc-web policy file %q: %w", r.path, err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var document grpcWebPolicyDocument
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("decode grpc-web policy file %q: %w", r.path, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("decode grpc-web policy file %q: unexpected trailing content", r.path)
	}

	policies := make([]domain.GRPCWebMethodPolicy, 0, len(document.Policies))
	for index, item := range document.Policies {
		timeout, err := time.ParseDuration(strings.TrimSpace(item.Timeout))
		if err != nil {
			return nil, fmt.Errorf("grpc-web policy %d timeout is invalid: %w", index, err)
		}
		policy := domain.GRPCWebMethodPolicy{
			FullMethod: strings.TrimSpace(item.FullMethod),
			Downstream: strings.TrimSpace(item.Downstream),
			AuthMode:   domain.GRPCWebAuthMode(strings.ToLower(strings.TrimSpace(item.AuthMode))),
			Roles:      normalizePolicyRoles(item.Roles),
			Timeout:    timeout,
		}
		if err := policy.Validate(); err != nil {
			return nil, fmt.Errorf("grpc-web policy %d is invalid: %w", index, err)
		}
		policies = append(policies, policy)
	}
	return policies, nil
}

func normalizePolicyRoles(roles []string) []string {
	seen := make(map[string]struct{}, len(roles))
	out := make([]string, 0, len(roles))
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		if _, exists := seen[role]; exists {
			continue
		}
		seen[role] = struct{}{}
		out = append(out, role)
	}
	return out
}
