package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"ecommerce/api-gateway/internal/domain"
)

var ErrGRPCWebMethodNotAllowed = errors.New("grpc-web method is not allowed")

type GRPCWebPolicyRepository interface {
	LoadPolicies(ctx context.Context) ([]domain.GRPCWebMethodPolicy, error)
}

type GRPCWebPolicyCatalog interface {
	Load(ctx context.Context) error
	FindByMethod(ctx context.Context, fullMethod string) (domain.GRPCWebMethodPolicy, error)
	List(ctx context.Context) ([]domain.GRPCWebMethodPolicy, error)
	RequiresToken(ctx context.Context) (bool, error)
}

type GRPCWebPolicyCatalogService struct {
	repo            GRPCWebPolicyRepository
	exposedServices map[string]struct{}

	mu       sync.RWMutex
	loaded   bool
	policies []domain.GRPCWebMethodPolicy
	byMethod map[string]domain.GRPCWebMethodPolicy
}

func NewGRPCWebPolicyCatalogService(repo GRPCWebPolicyRepository, exposedServices []string) *GRPCWebPolicyCatalogService {
	allowed := make(map[string]struct{}, len(exposedServices))
	for _, service := range exposedServices {
		service = strings.TrimSpace(service)
		if service != "" {
			allowed[service] = struct{}{}
		}
	}
	return &GRPCWebPolicyCatalogService{repo: repo, exposedServices: allowed}
}

func (s *GRPCWebPolicyCatalogService) Load(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.repo == nil {
		return errors.New("grpc-web policy repository is required")
	}
	policies, err := s.repo.LoadPolicies(ctx)
	if err != nil {
		return err
	}
	if len(policies) == 0 {
		return errors.New("at least one grpc-web method policy is required")
	}

	byMethod := make(map[string]domain.GRPCWebMethodPolicy, len(policies))
	for _, policy := range policies {
		if err := policy.Validate(); err != nil {
			return fmt.Errorf("invalid grpc-web policy for %q: %w", policy.FullMethod, err)
		}
		if _, allowed := s.exposedServices[policy.Service()]; !allowed {
			return fmt.Errorf("grpc-web policy service %q is not in the exposed service allowlist", policy.Service())
		}
		if _, duplicate := byMethod[policy.FullMethod]; duplicate {
			return fmt.Errorf("duplicate grpc-web method policy %q", policy.FullMethod)
		}
		policy.Roles = append([]string(nil), policy.Roles...)
		byMethod[policy.FullMethod] = policy
	}

	s.mu.Lock()
	s.policies = cloneGRPCWebPolicies(policies)
	s.byMethod = byMethod
	s.loaded = true
	s.mu.Unlock()
	return nil
}

func (s *GRPCWebPolicyCatalogService) FindByMethod(ctx context.Context, fullMethod string) (domain.GRPCWebMethodPolicy, error) {
	if err := s.ensureLoaded(ctx); err != nil {
		return domain.GRPCWebMethodPolicy{}, err
	}
	s.mu.RLock()
	policy, ok := s.byMethod[strings.TrimSpace(fullMethod)]
	s.mu.RUnlock()
	if !ok {
		return domain.GRPCWebMethodPolicy{}, fmt.Errorf("%w: %s", ErrGRPCWebMethodNotAllowed, fullMethod)
	}
	policy.Roles = append([]string(nil), policy.Roles...)
	return policy, nil
}

func (s *GRPCWebPolicyCatalogService) List(ctx context.Context) ([]domain.GRPCWebMethodPolicy, error) {
	if err := s.ensureLoaded(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneGRPCWebPolicies(s.policies), nil
}

func (s *GRPCWebPolicyCatalogService) RequiresToken(ctx context.Context) (bool, error) {
	policies, err := s.List(ctx)
	if err != nil {
		return false, err
	}
	for _, policy := range policies {
		if policy.RequiresToken() {
			return true, nil
		}
	}
	return false, nil
}

func (s *GRPCWebPolicyCatalogService) ensureLoaded(ctx context.Context) error {
	s.mu.RLock()
	loaded := s.loaded
	s.mu.RUnlock()
	if loaded {
		return ctx.Err()
	}
	return s.Load(ctx)
}

func cloneGRPCWebPolicies(policies []domain.GRPCWebMethodPolicy) []domain.GRPCWebMethodPolicy {
	out := make([]domain.GRPCWebMethodPolicy, len(policies))
	for index, policy := range policies {
		policy.Roles = append([]string(nil), policy.Roles...)
		out[index] = policy
	}
	return out
}
