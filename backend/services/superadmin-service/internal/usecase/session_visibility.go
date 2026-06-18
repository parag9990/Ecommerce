package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
	"ecommerce/superadmin-service/internal/rbac"
)

const (
	DefaultSessionAnalyticsMaxRange    = 30 * 24 * time.Hour
	DefaultSessionAnalyticsMaxPageSize = 100
)

var sessionIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

type SessionVisibilityAuthorizer interface {
	RequirePermission(ctx context.Context, actor domain.AdminActor, permission domain.Permission) error
	HasPermission(ctx context.Context, actor domain.AdminActor, permission domain.Permission) (bool, error)
}

type SessionVisibilityConfig struct {
	MaxDateRange time.Duration
	MaxPageSize  int
}

func DefaultSessionVisibilityConfig() SessionVisibilityConfig {
	return SessionVisibilityConfig{
		MaxDateRange: DefaultSessionAnalyticsMaxRange,
		MaxPageSize:  DefaultSessionAnalyticsMaxPageSize,
	}
}

func (c SessionVisibilityConfig) normalized() SessionVisibilityConfig {
	if c.MaxDateRange <= 0 {
		c.MaxDateRange = DefaultSessionAnalyticsMaxRange
	}
	if c.MaxPageSize <= 0 {
		c.MaxPageSize = DefaultSessionAnalyticsMaxPageSize
	}
	return c
}

type SessionVisibilityPolicy struct {
	Route            domain.SessionAnalyticsRoute
	Permission       domain.Permission
	RiskPermission   domain.Permission
	MaskPIIByDefault bool
	AllowExport      bool
}

var sessionVisibilityPolicies = map[domain.SessionAnalyticsRoute]SessionVisibilityPolicy{
	domain.SessionAnalyticsRouteLive: {
		Route:            domain.SessionAnalyticsRouteLive,
		Permission:       domain.PermissionSessionsRead,
		RiskPermission:   domain.PermissionSessionRiskRead,
		MaskPIIByDefault: true,
	},
	domain.SessionAnalyticsRouteSessions: {
		Route:            domain.SessionAnalyticsRouteSessions,
		Permission:       domain.PermissionSessionsRead,
		RiskPermission:   domain.PermissionSessionRiskRead,
		MaskPIIByDefault: true,
	},
	domain.SessionAnalyticsRouteJourney: {
		Route:            domain.SessionAnalyticsRouteJourney,
		Permission:       domain.PermissionSessionsRead,
		RiskPermission:   domain.PermissionSessionRiskRead,
		MaskPIIByDefault: true,
	},
	domain.SessionAnalyticsRouteFunnels: {
		Route:            domain.SessionAnalyticsRouteFunnels,
		Permission:       domain.PermissionSessionsRead,
		RiskPermission:   domain.PermissionSessionRiskRead,
		MaskPIIByDefault: true,
	},
	domain.SessionAnalyticsRouteHeatmaps: {
		Route:            domain.SessionAnalyticsRouteHeatmaps,
		Permission:       domain.PermissionSessionsRead,
		RiskPermission:   domain.PermissionSessionRiskRead,
		MaskPIIByDefault: true,
	},
}

func SessionVisibilityPolicyForRoute(route domain.SessionAnalyticsRoute) (SessionVisibilityPolicy, bool) {
	policy, ok := sessionVisibilityPolicies[route]
	return policy, ok
}

func SessionVisibilityPolicies() []SessionVisibilityPolicy {
	routes := domain.KnownSessionAnalyticsRoutes()
	out := make([]SessionVisibilityPolicy, 0, len(routes))
	for _, route := range routes {
		if policy, ok := SessionVisibilityPolicyForRoute(route); ok {
			out = append(out, policy)
		}
	}
	return out
}

type AuthorizeSessionAnalyticsRequest struct {
	Actor       domain.AdminActor
	Route       domain.SessionAnalyticsRoute
	IncludeRisk bool
	Export      bool
	Filters     domain.SessionAnalyticsFilters
}

type AuthorizeSessionAnalyticsResult struct {
	Allowed           bool
	Route             domain.SessionAnalyticsRoute
	Permission        domain.Permission
	RiskPermission    domain.Permission
	RiskAllowed       bool
	MaskPII           bool
	DownstreamHeaders map[string]string
}

type SessionDashboardModule struct {
	Key                string
	Route              domain.SessionAnalyticsRoute
	RequiredPermission domain.Permission
	RequiresRisk       bool
	Visible            bool
}

type SessionDashboardAccess struct {
	CanViewDashboard bool
	RiskAllowed      bool
	MaskPII          bool
	Modules          []SessionDashboardModule
}

type SessionVisibilityService struct {
	authz  SessionVisibilityAuthorizer
	audit  AuditRecorder
	config SessionVisibilityConfig
	logger logging.Logger
}

func NewSessionVisibilityService(authz SessionVisibilityAuthorizer, config SessionVisibilityConfig, logger logging.Logger) (*SessionVisibilityService, error) {
	return newSessionVisibilityService(authz, config, nil, logger)
}

func NewSessionVisibilityServiceWithAudit(authz SessionVisibilityAuthorizer, config SessionVisibilityConfig, audit AuditRecorder, logger logging.Logger) (*SessionVisibilityService, error) {
	return newSessionVisibilityService(authz, config, audit, logger)
}

func newSessionVisibilityService(authz SessionVisibilityAuthorizer, config SessionVisibilityConfig, audit AuditRecorder, logger logging.Logger) (*SessionVisibilityService, error) {
	if authz == nil {
		return nil, errors.New("session visibility service requires authorizer")
	}
	if logger == nil {
		logger = logging.NewNop()
	}
	return &SessionVisibilityService{
		authz:  authz,
		audit:  audit,
		config: config.normalized(),
		logger: logger,
	}, nil
}

func (s *SessionVisibilityService) AuthorizeSessionAnalytics(ctx context.Context, req AuthorizeSessionAnalyticsRequest) (*AuthorizeSessionAnalyticsResult, error) {
	policy, ok := SessionVisibilityPolicyForRoute(req.Route)
	if !ok {
		return nil, domain.NewValidationError("unknown session analytics route")
	}
	if !policy.AllowExport && req.Export {
		return nil, domain.NewValidationError("session analytics export is not included in this task")
	}
	if err := ValidateSessionAnalyticsFilters(req.Route, req.Filters, s.config); err != nil {
		return nil, err
	}

	result := &AuthorizeSessionAnalyticsResult{
		Allowed:        false,
		Route:          req.Route,
		Permission:     policy.Permission,
		RiskPermission: policy.RiskPermission,
	}

	if err := s.authz.RequirePermission(ctx, req.Actor, policy.Permission); err != nil {
		s.logDecision(ctx, req, result, "denied", err)
		return result, err
	}

	riskAllowed, err := s.authz.HasPermission(ctx, req.Actor, policy.RiskPermission)
	if err != nil {
		s.logDecision(ctx, req, result, "denied", err)
		return result, err
	}
	result.RiskAllowed = riskAllowed

	if req.IncludeRisk && !riskAllowed {
		err := domain.NewForbidden(policy.RiskPermission)
		s.logDecision(ctx, req, result, "denied", err)
		return result, err
	}

	result.Allowed = true
	result.MaskPII = policy.MaskPIIByDefault && !riskAllowed
	result.DownstreamHeaders = rbac.SessionAnalyticsMetadata(req.Actor, policy.Permission, riskAllowed, result.MaskPII)
	s.logDecision(ctx, req, result, "allowed", nil)
	if err := s.recordRiskAccessAudit(ctx, req, result); err != nil {
		return result, err
	}
	return result, nil
}

func (s *SessionVisibilityService) DashboardAccess(ctx context.Context, actor domain.AdminActor) (*SessionDashboardAccess, error) {
	if err := actor.ValidateForAdminRoute(); err != nil {
		return nil, err
	}

	canRead, err := s.authz.HasPermission(ctx, actor, domain.PermissionSessionsRead)
	if err != nil {
		return nil, err
	}
	riskAllowed, err := s.authz.HasPermission(ctx, actor, domain.PermissionSessionRiskRead)
	if err != nil {
		return nil, err
	}

	modules := []SessionDashboardModule{
		{Key: "live_sessions", Route: domain.SessionAnalyticsRouteLive, RequiredPermission: domain.PermissionSessionsRead, Visible: canRead},
		{Key: "session_table", Route: domain.SessionAnalyticsRouteSessions, RequiredPermission: domain.PermissionSessionsRead, Visible: canRead},
		{Key: "journey_explorer", Route: domain.SessionAnalyticsRouteJourney, RequiredPermission: domain.PermissionSessionsRead, Visible: canRead},
		{Key: "funnel_analysis", Route: domain.SessionAnalyticsRouteFunnels, RequiredPermission: domain.PermissionSessionsRead, Visible: canRead},
		{Key: "heatmap", Route: domain.SessionAnalyticsRouteHeatmaps, RequiredPermission: domain.PermissionSessionsRead, Visible: canRead},
		{Key: "suspicious_activity", Route: domain.SessionAnalyticsRouteSessions, RequiredPermission: domain.PermissionSessionRiskRead, RequiresRisk: true, Visible: canRead && riskAllowed},
		{Key: "risk_details", Route: domain.SessionAnalyticsRouteJourney, RequiredPermission: domain.PermissionSessionRiskRead, RequiresRisk: true, Visible: canRead && riskAllowed},
	}

	return &SessionDashboardAccess{
		CanViewDashboard: canRead,
		RiskAllowed:      canRead && riskAllowed,
		MaskPII:          canRead && !riskAllowed,
		Modules:          modules,
	}, nil
}

func ValidateSessionAnalyticsFilters(route domain.SessionAnalyticsRoute, filters domain.SessionAnalyticsFilters, config SessionVisibilityConfig) error {
	config = config.normalized()

	if filters.Page < 0 {
		return domain.NewValidationError("page must be a positive integer")
	}
	if filters.PageSize < 0 {
		return domain.NewValidationError("page_size must be a positive integer")
	}
	if filters.PageSize > config.MaxPageSize {
		return domain.NewValidationError(fmt.Sprintf("page_size cannot exceed %d", config.MaxPageSize))
	}

	if filters.From.IsZero() != filters.To.IsZero() {
		return domain.NewValidationError("from and to must be provided together")
	}
	if !filters.From.IsZero() {
		if filters.From.After(filters.To) {
			return domain.NewValidationError("from must be before to")
		}
		if filters.To.Sub(filters.From) > config.MaxDateRange {
			return domain.NewValidationError("date range is too large")
		}
	}

	sessionID := strings.TrimSpace(filters.SessionID)
	if route == domain.SessionAnalyticsRouteJourney && sessionID == "" {
		return domain.NewValidationError("session_id is required for journey analytics")
	}
	if sessionID != "" && !sessionIDPattern.MatchString(sessionID) {
		return domain.NewValidationError("session_id has invalid format")
	}

	path := strings.TrimSpace(filters.Path)
	if route == domain.SessionAnalyticsRouteHeatmaps && path == "" {
		return domain.NewValidationError("path is required for heatmap analytics")
	}
	if path != "" && (!strings.HasPrefix(path, "/") || len(path) > 2048) {
		return domain.NewValidationError("path has invalid format")
	}

	if !filters.DeviceType.Valid() {
		return domain.NewValidationError("device_type must be desktop, mobile, or tablet")
	}

	return nil
}

func (s *SessionVisibilityService) logDecision(ctx context.Context, req AuthorizeSessionAnalyticsRequest, result *AuthorizeSessionAnalyticsResult, status string, err error) {
	args := []any{
		"admin_id", req.Actor.AdminID,
		"route", req.Route,
		"permission", result.Permission,
		"risk_requested", req.IncludeRisk,
		"risk_allowed", result.RiskAllowed,
		"request_id", req.Actor.RequestID,
		"admin_session_id", req.Actor.SessionID,
		"result", status,
	}
	if err != nil {
		args = append(args, "error", err)
		s.logger.Warn(ctx, "admin session analytics access decision", args...)
		return
	}
	s.logger.Info(ctx, "admin session analytics access decision", args...)
}

func (s *SessionVisibilityService) recordRiskAccessAudit(ctx context.Context, req AuthorizeSessionAnalyticsRequest, result *AuthorizeSessionAnalyticsResult) error {
	if s.audit == nil || !req.IncludeRisk {
		return nil
	}

	after := map[string]string{
		"route":          string(req.Route),
		"risk_requested": strconv.FormatBool(req.IncludeRisk),
		"risk_allowed":   strconv.FormatBool(result.RiskAllowed),
		"mask_pii":       strconv.FormatBool(result.MaskPII),
	}
	if sessionID := strings.TrimSpace(req.Filters.SessionID); sessionID != "" {
		after["session_id"] = sessionID
	}
	if !req.Filters.From.IsZero() {
		after["from"] = req.Filters.From.UTC().Format(time.RFC3339)
	}
	if !req.Filters.To.IsZero() {
		after["to"] = req.Filters.To.UTC().Format(time.RFC3339)
	}
	if req.Filters.PageSize > 0 {
		after["page_size"] = strconv.Itoa(req.Filters.PageSize)
	}
	if req.Filters.DeviceType != "" {
		after["device_type"] = string(req.Filters.DeviceType)
	}
	if strings.TrimSpace(req.Filters.Path) != "" {
		after["path_filter_present"] = "true"
	}

	if err := s.audit.RecordAdminMutation(ctx, domain.AuditRecord{
		ActorAdminID: req.Actor.AdminID,
		Action:       "session.analytics.access",
		ResourceType: "session_analytics",
		ResourceID:   string(req.Route),
		RequestID:    req.Actor.RequestID,
		SessionID:    req.Actor.SessionID,
		IPHash:       req.Actor.IPHash,
		Reason:       "risk session analytics access",
		After:        after,
	}); err != nil {
		s.logger.Warn(ctx, "session analytics risk access audit record failed",
			"route", req.Route,
			"request_id", req.Actor.RequestID,
			"error", err,
		)
		return auditRecordFailure(err)
	}
	return nil
}
