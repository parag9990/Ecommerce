package httptransport

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"ecommerce/api-gateway/internal/authctx"
	"ecommerce/api-gateway/internal/clients"
	"ecommerce/api-gateway/internal/config"
	"ecommerce/api-gateway/internal/domain"
	"ecommerce/api-gateway/internal/handlers"
	"ecommerce/api-gateway/internal/observability"
	"ecommerce/api-gateway/internal/ratelimit"
	"ecommerce/api-gateway/internal/usecase"
	"ecommerce/api-gateway/internal/validation"
)

func NewRouter(ctx context.Context, cfg config.Config, catalog usecase.RouteCatalog, logger *slog.Logger, downstreamHealth DownstreamHealthChecker) (http.Handler, error) {
	return NewRouterWithOptions(ctx, cfg, catalog, logger, downstreamHealth, RouterOptions{})
}

type RouterOptions struct {
	TokenVerifier     TokenVerifier
	RateLimiter       ratelimit.Limiter
	RateLimitPolicies []ratelimit.Policy
	RequestValidator  RequestValidator
	Metrics           *observability.Metrics
	UserClient        clients.UserClient
	SearchClient      clients.SearchServiceClient
	SessionHTTPClient *http.Client
	AuthHTTPClient    *http.Client
}

func NewRouterWithOptions(ctx context.Context, cfg config.Config, catalog usecase.RouteCatalog, logger *slog.Logger, downstreamHealth DownstreamHealthChecker, opts RouterOptions) (http.Handler, error) {
	routes, err := catalog.ListRoutes(ctx)
	if err != nil {
		return nil, fmt.Errorf("load route catalog: %w", err)
	}
	verifier, err := routeTokenVerifier(cfg, logger, routes, opts.TokenVerifier)
	if err != nil {
		return nil, err
	}
	rateLimiter, err := newRateLimitMiddleware(cfg, logger, opts)
	if err != nil {
		return nil, err
	}
	requestValidator, err := routeRequestValidator(ctx, cfg, catalog, logger, opts.RequestValidator)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	handler := NewHandler(catalog, logger, cfg.ServiceName, downstreamHealth)
	labeler := observability.NewRouteLabeler(routes,
		observability.StaticRoute{Method: http.MethodGet, Path: "/health/live"},
		observability.StaticRoute{Method: http.MethodGet, Path: "/health/ready"},
	)

	mux.HandleFunc("GET /health/live", handler.HealthLive)
	mux.HandleFunc("GET /health/ready", handler.HealthReady)
	var userHandler *handlers.UserHandler
	if opts.UserClient != nil {
		userHandler = handlers.NewUserHandler(opts.UserClient, logger)
	}
	var searchHandler *handlers.SearchHandler
	if opts.SearchClient != nil {
		searchHandler = handlers.NewSearchHandler(opts.SearchClient, logger)
	}
	var sessionProxy *SessionProxy
	if strings.TrimSpace(cfg.SessionHTTPURL) != "" {
		client := opts.SessionHTTPClient
		if client == nil {
			client = &http.Client{Timeout: cfg.SessionHTTPTimeout}
		}
		sessionProxy, err = NewSessionProxy(cfg.SessionHTTPURL, client, logger)
		if err != nil {
			return nil, fmt.Errorf("configure session HTTP proxy: %w", err)
		}
	}
	var authProxy *SessionProxy
	if strings.TrimSpace(cfg.AuthHTTPURL) != "" {
		client := opts.AuthHTTPClient
		if client == nil {
			client = &http.Client{Timeout: cfg.AuthHTTPTimeout}
		}
		authProxy, err = NewSessionProxy(cfg.AuthHTTPURL, client, logger)
		if err != nil {
			return nil, fmt.Errorf("configure auth HTTP proxy: %w", err)
		}
	}
	for _, route := range routes {
		route := route
		endpoint := handler.RouteDefined(route)
		if userHandler != nil {
			endpoint = userRouteEndpoint(route, userHandler, endpoint)
		}
		if searchHandler != nil {
			endpoint = searchRouteEndpoint(route, searchHandler, endpoint)
		}
		if sessionProxy != nil && route.Service == "session-service" {
			endpoint = sessionProxy.ServeHTTP
		}
		if authProxy != nil && route.Service == "auth-service" {
			endpoint = authProxy.ServeHTTP
		}
		baseHandler := RequestValidationMiddleware(route, requestValidator, logger, opts.Metrics)(http.HandlerFunc(endpoint))
		routeHandler, err := secureRoute(route, baseHandler, verifier, cfg.WebhookSignatureHeader, logger, rateLimiter, opts.Metrics, cfg.Observability.Normalize(cfg.ServiceName, cfg.Environment).UserHashSalt)
		if err != nil {
			return nil, fmt.Errorf("configure route %s: %w", route.Key(), err)
		}
		if rateLimiter != nil {
			routeHandler = rateLimiter.PreAuth(route)(routeHandler)
		}
		mux.Handle(route.Pattern(), routeHandler)
	}

	var wrapped http.Handler = mux
	wrapped = TracingMiddleware(cfg.Observability.Normalize(cfg.ServiceName, cfg.Environment), labeler)(wrapped)
	wrapped = MetricsMiddleware(opts.Metrics, labeler)(wrapped)
	wrapped = AccessLogMiddlewareWithConfig(cfg.Observability.Normalize(cfg.ServiceName, cfg.Environment), logger, labeler)(wrapped)
	wrapped = RecoveryMiddleware(logger)(wrapped)
	wrapped = RequestIDMiddleware(wrapped)
	return wrapped, nil
}

func searchRouteEndpoint(route domain.RouteDefinition, handler *handlers.SearchHandler, fallback http.HandlerFunc) http.HandlerFunc {
	switch string(route.Method) + " " + route.Path {
	case "GET /api/v1/search":
		return handler.SearchProducts
	case "GET /api/v1/search/autocomplete":
		return handler.Autocomplete
	case "POST /api/v1/admin/search/synonyms":
		return handler.CreateSynonym
	case "GET /api/v1/admin/search/synonyms":
		return handler.ListSynonyms
	default:
		return fallback
	}
}

func userRouteEndpoint(route domain.RouteDefinition, handler *handlers.UserHandler, fallback http.HandlerFunc) http.HandlerFunc {
	var endpoint http.HandlerFunc
	switch string(route.Method) + " " + route.Path {
	case "GET /api/v1/me":
		endpoint = handler.GetMe
	case "PATCH /api/v1/me":
		endpoint = handler.UpdateMe
	case "GET /api/v1/me/addresses":
		endpoint = handler.ListAddresses
	case "POST /api/v1/me/addresses":
		endpoint = handler.CreateAddress
	case "PATCH /api/v1/me/addresses/{address_id}":
		endpoint = handler.UpdateAddress
	case "DELETE /api/v1/me/addresses/{address_id}":
		endpoint = handler.DeleteAddress
	case "GET /api/v1/sellers/me":
		endpoint = handler.GetSellerMe
	case "PATCH /api/v1/sellers/me":
		endpoint = handler.UpdateSellerMe
	default:
		return fallback
	}

	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := gatewayauth.ClaimsFromContext(r.Context())
		if ok {
			r = r.WithContext(authctx.WithClaims(r.Context(), authctx.Claims{
				UserID:    claims.UserID(),
				SessionID: claims.SessionID,
				SellerID:  claims.SellerID,
				Roles:     append([]string(nil), claims.Roles...),
				TokenType: claims.TokenType,
			}))
		}
		endpoint(w, r)
	}
}

func secureRoute(route domain.RouteDefinition, next http.Handler, verifier TokenVerifier, webhookSignatureHeader string, logger *slog.Logger, rateLimiter *rateLimitMiddleware, metrics *observability.Metrics, userHashSalt string) (http.Handler, error) {
	switch route.AuthLevel {
	case domain.AuthPublic:
		return next, nil
	case domain.AuthWebhook:
		return RequireWebhookSignature(webhookSignatureHeader, logger, route, metrics)(next), nil
	default:
		allowed := route.AllowedRoles()
		if len(allowed) == 0 {
			return nil, fmt.Errorf("no allowed roles configured for auth level %q", route.AuthLevel)
		}
		if verifier == nil {
			return nil, fmt.Errorf("token verifier is required for auth level %q", route.AuthLevel)
		}
		protected := RequireRoles(logger, route, metrics, userHashSalt, allowed...)(next)
		if rateLimiter != nil {
			protected = rateLimiter.PostAuth(route)(protected)
		}
		protected = AuthRequired(verifier, logger, route, metrics, userHashSalt)(protected)
		return protected, nil
	}
}

func routeTokenVerifier(cfg config.Config, logger *slog.Logger, routes []domain.RouteDefinition, injected TokenVerifier) (TokenVerifier, error) {
	if injected != nil {
		return injected, nil
	}
	if !routesRequireToken(routes) {
		return nil, nil
	}
	if strings.TrimSpace(cfg.JWTJWKSURL) == "" {
		return nil, fmt.Errorf("JWT_JWKS_URL is required when protected routes are configured")
	}
	client := &http.Client{Timeout: cfg.JWTJWKSFetchTimeout}
	keys, err := gatewayauth.NewRemoteJWKSKeyProvider(gatewayauth.JWKSConfig{
		URL:      cfg.JWTJWKSURL,
		CacheTTL: cfg.JWTJWKSCacheTTL,
	}, client, logger)
	if err != nil {
		return nil, fmt.Errorf("configure jwks key provider: %w", err)
	}
	verifier, err := gatewayauth.NewVerifier(cfg.JWTIssuer, cfg.JWTAudience, cfg.JWTAllowedAlgs, cfg.JWTClockSkew, keys)
	if err != nil {
		return nil, fmt.Errorf("configure jwt verifier: %w", err)
	}
	return verifier, nil
}

func routesRequireToken(routes []domain.RouteDefinition) bool {
	for _, route := range routes {
		if route.RequiresToken() {
			return true
		}
	}
	return false
}

func routeRequestValidator(ctx context.Context, cfg config.Config, catalog usecase.RouteCatalog, logger *slog.Logger, injected RequestValidator) (RequestValidator, error) {
	if injected != nil {
		return injected, nil
	}
	if !cfg.Validation.Enabled {
		return nil, nil
	}
	schemas, err := catalog.Schemas(ctx)
	if err != nil {
		return nil, fmt.Errorf("load validation schemas: %w", err)
	}
	return validation.NewService(schemas, validation.Options{
		Enabled:             cfg.Validation.Enabled,
		DefaultMaxBodyBytes: cfg.Validation.DefaultMaxBodyBytes,
		MaxHeaderBytes:      cfg.Validation.MaxHeaderBytes,
		MaxQueryBytes:       cfg.Validation.MaxQueryBytes,
	}, logger), nil
}
