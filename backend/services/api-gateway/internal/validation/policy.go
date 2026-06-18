package validation

import (
	"net/http"
	"strings"

	"ecommerce/api-gateway/internal/domain"
)

const (
	ContentTypeJSON       = "application/json"
	ContentTypeForm       = "application/x-www-form-urlencoded"
	defaultHeaderMaxBytes = int64(32 * 1024)
	defaultQueryMaxBytes  = 8 * 1024
	defaultBodyMaxBytes   = int64(128 * 1024)
	authBodyMaxBytes      = int64(64 * 1024)
	commandBodyMaxBytes   = int64(128 * 1024)
	sellerWriteMaxBytes   = int64(1024 * 1024)
	adminWriteMaxBytes    = int64(256 * 1024)
	webhookBodyMaxBytes   = int64(512 * 1024)
	sessionBodyMaxBytes   = int64(64 * 1024)
)

type Options struct {
	Enabled             bool
	DefaultMaxBodyBytes int64
	MaxHeaderBytes      int64
	MaxQueryBytes       int
}

type Policy struct {
	RouteID        string
	Method         string
	Path           string
	RequestSchema  string
	BodyRequired   bool
	BodyAllowed    bool
	MaxBodyBytes   int64
	ContentTypes   []string
	RequireIDKey   bool
	ValidateSchema bool
	Webhook        bool
}

func NormalizeOptions(opts Options) Options {
	if opts.DefaultMaxBodyBytes <= 0 {
		opts.DefaultMaxBodyBytes = defaultBodyMaxBytes
	}
	if opts.MaxHeaderBytes <= 0 {
		opts.MaxHeaderBytes = defaultHeaderMaxBytes
	}
	if opts.MaxQueryBytes <= 0 {
		opts.MaxQueryBytes = defaultQueryMaxBytes
	}
	return opts
}

func PolicyForRoute(route domain.RouteDefinition, opts Options) Policy {
	opts = NormalizeOptions(opts)
	method := string(route.Method)
	schemaName := strings.TrimSpace(route.RequestSchema)
	policy := Policy{
		RouteID:        route.ID,
		Method:         method,
		Path:           route.Path,
		RequestSchema:  schemaName,
		MaxBodyBytes:   routeBodyLimit(route, opts.DefaultMaxBodyBytes),
		ValidateSchema: schemaName != "" && schemaName != "Empty",
		Webhook:        route.AuthLevel == domain.AuthWebhook,
	}

	switch method {
	case http.MethodGet:
		policy.BodyAllowed = false
	case http.MethodDelete:
		policy.BodyAllowed = false
	default:
		policy.BodyAllowed = schemaName != "" && schemaName != "Empty" && schemaName != "IdPathRequest"
		policy.BodyRequired = policy.BodyAllowed
		if policy.BodyAllowed {
			policy.ContentTypes = []string{ContentTypeJSON}
		}
	}

	if policy.Webhook {
		policy.BodyAllowed = true
		policy.BodyRequired = true
		policy.ValidateSchema = false
		policy.MaxBodyBytes = webhookBodyMaxBytes
		policy.ContentTypes = []string{ContentTypeJSON, ContentTypeForm}
	}

	policy.RequireIDKey = requiresIdempotencyKey(route)
	return policy
}

func routeBodyLimit(route domain.RouteDefinition, fallback int64) int64 {
	if fallback <= 0 {
		fallback = defaultBodyMaxBytes
	}
	switch {
	case route.AuthLevel == domain.AuthWebhook:
		return webhookBodyMaxBytes
	case route.Service == "auth-service":
		return authBodyMaxBytes
	case route.ID == "session.ingest":
		return sessionBodyMaxBytes
	case strings.HasPrefix(route.ID, "seller.product.") ||
		strings.HasPrefix(route.ID, "cms.coupon_") ||
		strings.HasPrefix(route.ID, "cms.campaign_"):
		return sellerWriteMaxBytes
	case strings.HasPrefix(route.ID, "admin."):
		return adminWriteMaxBytes
	case strings.HasPrefix(route.ID, "cart.") ||
		strings.HasPrefix(route.ID, "wishlist.") ||
		strings.HasPrefix(route.ID, "order.") ||
		strings.HasPrefix(route.ID, "payment."):
		return commandBodyMaxBytes
	default:
		return fallback
	}
}

func requiresIdempotencyKey(route domain.RouteDefinition) bool {
	switch route.ID {
	case "order.checkout", "payment.retry", "payment.refund":
		return true
	default:
		return false
	}
}
