package httptransport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"ecommerce/api-gateway/internal/config"
	"ecommerce/api-gateway/internal/domain"
	gatewayerrors "ecommerce/api-gateway/internal/errors"
	"ecommerce/api-gateway/internal/observability"
	"ecommerce/api-gateway/internal/ratelimit"
)

type rateLimitMiddleware struct {
	limiter              ratelimit.Limiter
	policies             []ratelimit.Policy
	keyPrefix            string
	failOpen             bool
	targetBodyLimitBytes int64
	clientIP             clientIPResolver
	logger               *slog.Logger
	metrics              *observability.Metrics
}

type clientIPResolver struct {
	trustedProxies []netip.Prefix
}

type preservedReadCloser struct {
	io.Reader
	closer io.Closer
}

func (r preservedReadCloser) Close() error {
	return r.closer.Close()
}

func newRateLimitMiddleware(cfg config.Config, logger *slog.Logger, opts RouterOptions) (*rateLimitMiddleware, error) {
	if !cfg.RateLimit.Enabled {
		return nil, nil
	}
	if opts.RateLimiter == nil {
		return nil, errors.New("rate limiter is required when rate limiting is enabled")
	}

	policies := opts.RateLimitPolicies
	if len(policies) == 0 {
		policies = ratelimit.DefaultPolicies(cfg.RateLimit.DefaultIPLimit, cfg.RateLimit.DefaultIPWindow)
	}
	if err := ratelimit.ValidatePolicies(policies); err != nil {
		return nil, fmt.Errorf("validate rate limit policies: %w", err)
	}
	resolver, err := newClientIPResolver(cfg.RateLimit.TrustedProxyCIDRs)
	if err != nil {
		return nil, err
	}

	return &rateLimitMiddleware{
		limiter:              opts.RateLimiter,
		policies:             cloneRateLimitPolicies(policies),
		keyPrefix:            cfg.RateLimit.KeyPrefix,
		failOpen:             cfg.RateLimit.FailOpen,
		targetBodyLimitBytes: cfg.RateLimit.TargetBodyLimitBytes,
		clientIP:             resolver,
		logger:               logger,
		metrics:              opts.Metrics,
	}, nil
}

func (m *rateLimitMiddleware) PreAuth(route domain.RouteDefinition) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if m == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !m.apply(w, r, route, ratelimit.DimensionIP, ratelimit.DimensionTarget, ratelimit.DimensionRoute) {
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (m *rateLimitMiddleware) PostAuth(route domain.RouteDefinition) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if m == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !m.apply(w, r, route, ratelimit.DimensionUser) {
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (m *rateLimitMiddleware) apply(w http.ResponseWriter, r *http.Request, route domain.RouteDefinition, dimensions ...ratelimit.Dimension) bool {
	policies := ratelimit.MatchPolicies(m.policies, route.ID, r.Method, r.URL.Path, dimensions...)
	for _, policy := range policies {
		identity, ok := m.identityForPolicy(r, route, policy)
		if !ok {
			continue
		}
		result, err := m.limiter.Allow(r.Context(), ratelimit.Request{
			Key:    ratelimit.BuildKey(m.keyPrefix, policy.Dimension, policy.Name, identity),
			Limit:  policy.Limit,
			Window: policy.Window,
			Cost:   policy.RequestCost(),
		})
		if err != nil {
			m.logRateLimitError(r, route, policy, err)
			if m.failOpen {
				continue
			}
			writeRateLimitUnavailable(w, r)
			return false
		}

		writeRateLimitHeaders(w, result)
		if !result.Allowed {
			m.metrics.ObserveRateLimited(route.Path, string(policy.Dimension))
			m.logRateLimitBlocked(r, route, policy, result)
			writeRateLimited(w, r, result)
			return false
		}
	}
	return true
}

func (m *rateLimitMiddleware) identityForPolicy(r *http.Request, route domain.RouteDefinition, policy ratelimit.Policy) (string, bool) {
	switch policy.Dimension {
	case ratelimit.DimensionIP:
		return m.clientIP.Resolve(r), true
	case ratelimit.DimensionUser:
		userID, ok := gatewayauth.UserIDFromContext(r.Context())
		return userID, ok
	case ratelimit.DimensionRoute:
		return route.ID, true
	case ratelimit.DimensionTarget:
		target, err := extractRateLimitTarget(r, policy.TargetFields, m.targetBodyLimitBytes)
		if err != nil {
			m.logTargetExtractionError(r, route, policy, err)
		}
		return target, true
	default:
		return "", false
	}
}

func (m *rateLimitMiddleware) logRateLimitError(r *http.Request, route domain.RouteDefinition, policy ratelimit.Policy, err error) {
	if m.logger == nil {
		return
	}
	m.logger.ErrorContext(r.Context(), "rate_limit_check_failed",
		"route_id", route.ID,
		"policy", policy.Name,
		"dimension", policy.Dimension,
		"method", r.Method,
		"route", route.Path,
		"fail_open", m.failOpen,
		"request_id", RequestIDFromContext(r.Context()),
		"error", err,
	)
}

func (m *rateLimitMiddleware) logRateLimitBlocked(r *http.Request, route domain.RouteDefinition, policy ratelimit.Policy, result ratelimit.Result) {
	if m.logger == nil {
		return
	}
	m.logger.WarnContext(r.Context(), "rate_limit_blocked",
		"route_id", route.ID,
		"policy", policy.Name,
		"dimension", policy.Dimension,
		"method", r.Method,
		"route", route.Path,
		"limit", result.Limit,
		"remaining", result.Remaining,
		"retry_after_ms", result.RetryAfter.Milliseconds(),
		"request_id", RequestIDFromContext(r.Context()),
	)
}

func (m *rateLimitMiddleware) logTargetExtractionError(r *http.Request, route domain.RouteDefinition, policy ratelimit.Policy, err error) {
	if m.logger == nil {
		return
	}
	m.logger.WarnContext(r.Context(), "rate_limit_target_extract_failed",
		"route_id", route.ID,
		"policy", policy.Name,
		"method", r.Method,
		"route", route.Path,
		"request_id", RequestIDFromContext(r.Context()),
		"error", err,
	)
}

func newClientIPResolver(cidrs []string) (clientIPResolver, error) {
	resolver := clientIPResolver{}
	for _, cidr := range cidrs {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(cidr)
		if err != nil {
			return clientIPResolver{}, fmt.Errorf("parse trusted proxy CIDR %q: %w", cidr, err)
		}
		resolver.trustedProxies = append(resolver.trustedProxies, prefix)
	}
	return resolver, nil
}

func (r clientIPResolver) Resolve(req *http.Request) string {
	remoteIP := parseRemoteIP(req.RemoteAddr)
	if remoteIP.IsValid() && r.isTrustedProxy(remoteIP) {
		if ip := firstForwardedFor(req.Header.Get("X-Forwarded-For")); ip.IsValid() {
			return ip.String()
		}
		if ip := parseIP(req.Header.Get("X-Real-IP")); ip.IsValid() {
			return ip.String()
		}
	}
	if remoteIP.IsValid() {
		return remoteIP.String()
	}
	return "unknown"
}

func (r clientIPResolver) isTrustedProxy(ip netip.Addr) bool {
	for _, prefix := range r.trustedProxies {
		if prefix.Contains(ip) {
			return true
		}
	}
	return false
}

func firstForwardedFor(value string) netip.Addr {
	for _, part := range strings.Split(value, ",") {
		if ip := parseIP(part); ip.IsValid() {
			return ip
		}
	}
	return netip.Addr{}
}

func parseRemoteIP(remoteAddr string) netip.Addr {
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err == nil {
		return parseIP(host)
	}
	return parseIP(remoteAddr)
}

func parseIP(value string) netip.Addr {
	ip, err := netip.ParseAddr(strings.TrimSpace(value))
	if err != nil {
		return netip.Addr{}
	}
	return ip.Unmap()
}

func extractRateLimitTarget(r *http.Request, fields []string, limitBytes int64) (string, error) {
	if r.Body == nil {
		return "", nil
	}
	if limitBytes <= 0 {
		limitBytes = 64 * 1024
	}

	body := r.Body
	data, err := io.ReadAll(io.LimitReader(body, limitBytes+1))
	if err != nil {
		r.Body = preservedReadCloser{Reader: bytes.NewReader(data), closer: body}
		return "", err
	}
	if int64(len(data)) > limitBytes {
		r.Body = preservedReadCloser{Reader: io.MultiReader(bytes.NewReader(data), body), closer: body}
		return "", fmt.Errorf("request body exceeds target extraction limit")
	}
	_ = body.Close()
	r.Body = io.NopCloser(bytes.NewReader(data))

	var payload map[string]any
	if len(bytes.TrimSpace(data)) == 0 {
		return "", nil
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}
	for _, field := range fields {
		value, ok := payload[field].(string)
		if ok && strings.TrimSpace(value) != "" {
			return value, nil
		}
	}
	return "", nil
}

func writeRateLimitHeaders(w http.ResponseWriter, result ratelimit.Result) {
	w.Header().Set("RateLimit-Limit", strconv.FormatInt(result.Limit, 10))
	w.Header().Set("RateLimit-Remaining", strconv.FormatInt(result.Remaining, 10))
	w.Header().Set("RateLimit-Reset", strconv.FormatInt(durationSecondsCeil(result.ResetAfter), 10))
	if result.RetryAfter > 0 {
		w.Header().Set("Retry-After", strconv.FormatInt(durationSecondsCeil(result.RetryAfter), 10))
	}
}

func writeRateLimited(w http.ResponseWriter, r *http.Request, result ratelimit.Result) {
	retryAfterSeconds := durationSecondsCeil(result.RetryAfter)
	if retryAfterSeconds <= 0 {
		retryAfterSeconds = 1
	}
	w.Header().Set("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
	writeMappedError(w, r, gatewayerrors.New(
		http.StatusTooManyRequests,
		gatewayerrors.CodeRateLimited,
		"Too many requests",
		map[string]int64{"retry_after_seconds": retryAfterSeconds},
		nil,
	))
}

func writeRateLimitUnavailable(w http.ResponseWriter, r *http.Request) {
	writeMappedError(w, r, gatewayerrors.New(
		http.StatusServiceUnavailable,
		gatewayerrors.CodeServiceUnavailable,
		"Service temporarily unavailable",
		nil,
		nil,
	))
}

func durationSecondsCeil(duration time.Duration) int64 {
	if duration <= 0 {
		return 0
	}
	seconds := int64(duration / time.Second)
	if duration%time.Second != 0 {
		seconds++
	}
	if seconds <= 0 {
		return 1
	}
	return seconds
}

func cloneRateLimitPolicies(policies []ratelimit.Policy) []ratelimit.Policy {
	out := make([]ratelimit.Policy, len(policies))
	for i, policy := range policies {
		out[i] = policy
		out[i].RouteIDs = cloneStrings(policy.RouteIDs)
		out[i].Methods = cloneStrings(policy.Methods)
		out[i].PathPrefixes = cloneStrings(policy.PathPrefixes)
		out[i].TargetFields = cloneStrings(policy.TargetFields)
	}
	return out
}

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}
