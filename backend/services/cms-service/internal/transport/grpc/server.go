package grpctransport

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
	cmsv1 "github.com/example/ecommerce-platform/backend/services/cms-service/internal/gen/ecommerce/cms/v1"
	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	defaultMaxRecvMsgBytes = 1 << 20
	defaultMaxSendMsgBytes = 1 << 20
)

type CouponService interface {
	ValidateCoupon(ctx context.Context, input usecase.ValidateCouponInput) (domain.CouponValidationResult, error)
}

type SellerSettingsService interface {
	GetSellerSettings(ctx context.Context, input usecase.GetSellerSettingsInput) (domain.SellerSettings, error)
}

type CampaignService interface {
	GetCampaign(ctx context.Context, input usecase.GetCampaignInput) (usecase.CampaignRead, error)
	ListCampaignReads(ctx context.Context, input usecase.ListCampaignsInput) ([]usecase.CampaignRead, error)
}

type AuditLogService interface {
	ListAuditLogs(ctx context.Context, input usecase.ListAuditLogsInput) (domain.AuditLogPage, error)
}

type ServerConfig struct {
	InternalAuthHeader     string
	InternalAuthToken      string
	AllowedInternalCallers []string
	MaxRecvMsgBytes        int
	MaxSendMsgBytes        int
}

type Server struct {
	cmsv1.UnimplementedCMSServiceServer

	coupons       CouponService
	settings      SellerSettingsService
	campaigns     CampaignService
	auditLogs     AuditLogService
	config        ServerConfig
	allowedCaller map[string]struct{}
	logger        *slog.Logger
}

func NewServer(coupons CouponService, settings SellerSettingsService, campaigns CampaignService, auditLogs AuditLogService, cfg ServerConfig, logger *slog.Logger) (*Server, error) {
	if coupons == nil {
		return nil, domain.ErrCouponRepositoryRequired
	}
	if settings == nil {
		return nil, domain.ErrSellerSettingsRepositoryRequired
	}
	if campaigns == nil {
		return nil, domain.ErrCampaignRepositoryRequired
	}
	if auditLogs == nil {
		return nil, domain.ErrAuditLogRepositoryRequired
	}
	if logger == nil {
		logger = slog.Default()
	}
	cfg.InternalAuthHeader = strings.TrimSpace(cfg.InternalAuthHeader)
	if cfg.InternalAuthHeader == "" {
		cfg.InternalAuthHeader = "X-Internal-Token"
	}
	if cfg.MaxRecvMsgBytes <= 0 {
		cfg.MaxRecvMsgBytes = defaultMaxRecvMsgBytes
	}
	if cfg.MaxSendMsgBytes <= 0 {
		cfg.MaxSendMsgBytes = defaultMaxSendMsgBytes
	}

	return &Server{
		coupons:       coupons,
		settings:      settings,
		campaigns:     campaigns,
		auditLogs:     auditLogs,
		config:        cfg,
		allowedCaller: allowedCallerSet(cfg.AllowedInternalCallers),
		logger:        logger,
	}, nil
}

func NewGRPCServer(service *Server) *grpc.Server {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(service.unaryLoggingInterceptor),
		grpc.MaxRecvMsgSize(service.config.MaxRecvMsgBytes),
		grpc.MaxSendMsgSize(service.config.MaxSendMsgBytes),
	)
	cmsv1.RegisterCMSServiceServer(server, service)
	return server
}

func (s *Server) unaryLoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	md := metadataFromContext(ctx)
	code := status.Code(err)
	s.logger.InfoContext(ctx, "cms.grpc.request",
		slog.String("grpc_method", info.FullMethod),
		slog.String("grpc_code", code.String()),
		slog.String("request_id", md.RequestID),
		slog.String("trace_id", md.TraceID),
		slog.String("caller_service", md.ServiceName),
		slog.Int64("latency_ms", time.Since(start).Milliseconds()),
	)
	return resp, err
}

func allowedCallerSet(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out[value] = struct{}{}
		}
	}
	return out
}
