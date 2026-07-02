package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/clients"
	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/config"
	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/database"
	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/repository"
	grpctransport "github.com/example/ecommerce-platform/backend/services/cms-service/internal/transport/grpc"
	httptransport "github.com/example/ecommerce-platform/backend/services/cms-service/internal/transport/http"
	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/usecase"
	platformmiddleware "github.com/parag/ecommerce/backend/shared/platform/middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("cms.config.load_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.OpenMySQL(ctx, cfg.Database)
	if err != nil {
		logger.Error("cms.mysql.connect_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	auditRecorder, err := repository.NewMySQLAuditRecorder(db)
	if err != nil {
		logger.Error("cms.audit_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	reviewRepo, err := repository.NewMySQLProductModerationRepository(db)
	if err != nil {
		logger.Error("cms.product_review_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	couponRepo, err := repository.NewMySQLCouponRepository(db)
	if err != nil {
		logger.Error("cms.coupon_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	campaignRepo, err := repository.NewMySQLCampaignRepository(db)
	if err != nil {
		logger.Error("cms.campaign_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	analyticsRepo, err := repository.NewMySQLSellerAnalyticsRepository(db)
	if err != nil {
		logger.Error("cms.analytics_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	settingsRepo, err := repository.NewMySQLSellerSettingsRepository(db)
	if err != nil {
		logger.Error("cms.seller_settings_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var staffRepo usecase.SellerStaffRepository
	if cfg.Authorization.StaffStatusSource == config.StaffStatusSourceMySQL {
		staffRepo, err = repository.NewMySQLSellerStaffRepository(db)
		if err != nil {
			logger.Error("cms.staff_repository.init_failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}
	dbName := cfg.Database.Name
	dbHost := cfg.Database.Host
	if cfg.Database.DSN != "" {
		dbName = "configured_by_dsn"
		dbHost = "configured_by_dsn"
	}
	logger.Info("cms.mysql.connected",
		slog.String("db_name", dbName),
		slog.String("db_host", dbHost),
		slog.Int("max_open_conns", cfg.Database.MaxOpenConns),
		slog.Int("max_idle_conns", cfg.Database.MaxIdleConns),
	)

	authorizer, err := usecase.NewAuthorizer(staffRepo, auditRecorder, logger)
	if err != nil {
		logger.Error("cms.authorizer.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	productClient, err := clients.NewHTTPProductClient(clients.ProductClientConfig{
		BaseURL:            cfg.ProductService.BaseURL,
		Timeout:            cfg.ProductService.Timeout,
		InternalAuthHeader: cfg.ProductService.InternalAuthHeader,
		InternalAuthToken:  cfg.ProductService.InternalAuthToken,
	}, logger)
	if err != nil {
		logger.Error("cms.product_client.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	moderationUsecase, err := usecase.NewProductModerationUsecase(authorizer, reviewRepo, productClient, auditRecorder, logger)
	if err != nil {
		logger.Error("cms.product_moderation.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	couponUsecase, err := usecase.NewCouponUsecase(authorizer, couponRepo, auditRecorder, logger)
	if err != nil {
		logger.Error("cms.coupon_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	campaignUsecase, err := usecase.NewCampaignUsecase(authorizer, campaignRepo, couponRepo, auditRecorder, logger, usecase.CampaignOptions{
		MaxDuration: cfg.Campaign.MaxDuration,
	})
	if err != nil {
		logger.Error("cms.campaign_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	couponUsecase.SetCampaignRepository(campaignRepo)

	settingsUsecase, err := usecase.NewSellerSettingsUsecase(authorizer, settingsRepo, logger)
	if err != nil {
		logger.Error("cms.seller_settings_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	analyticsUsecase, err := usecase.NewSellerAnalyticsUsecase(authorizer, analyticsRepo, logger, usecase.SellerAnalyticsOptions{
		DefaultCurrency:         cfg.Analytics.DefaultCurrency,
		DefaultRangeDays:        cfg.Analytics.DefaultRangeDays,
		MaxRangeDays:            cfg.Analytics.MaxRangeDays,
		DefaultTopProductsLimit: cfg.Analytics.DefaultTopProductsLimit,
		MaxTopProductsLimit:     cfg.Analytics.MaxTopProductsLimit,
	})
	if err != nil {
		logger.Error("cms.analytics_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	auditLogUsecase, err := usecase.NewAuditLogUsecase(authorizer, auditRecorder, logger, usecase.AuditLogOptions{
		DefaultRangeDays: cfg.Audit.DefaultRangeDays,
		DefaultPageSize:  cfg.Audit.DefaultPageSize,
		MaxPageSize:      cfg.Audit.MaxPageSize,
	})
	if err != nil {
		logger.Error("cms.audit_log_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	handler, err := httptransport.NewHandler(authorizer, moderationUsecase, couponUsecase, campaignUsecase, auditLogUsecase, httptransport.HandlerConfig{
		InternalAuthHeader: cfg.Security.InternalAuthHeader,
		InternalAuthToken:  cfg.Security.InternalAuthToken,
		MaxBodyBytes:       cfg.HTTP.MaxBodyBytes,
	}, logger)
	if err != nil {
		logger.Error("cms.http.handler_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := handler.SetSellerAnalyticsService(analyticsUsecase); err != nil {
		logger.Error("cms.http.analytics_handler_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	server := &http.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      platformmiddleware.CORS(platformmiddleware.DefaultCORSConfig())(httptransport.NewRouter(handler)),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	grpcHandler, err := grpctransport.NewServer(couponUsecase, settingsUsecase, campaignUsecase, auditLogUsecase, grpctransport.ServerConfig{
		InternalAuthHeader:     cfg.Security.InternalAuthHeader,
		InternalAuthToken:      cfg.Security.InternalAuthToken,
		AllowedInternalCallers: cfg.GRPC.AllowedInternalCallers,
		MaxRecvMsgBytes:        cfg.GRPC.MaxRecvMsgBytes,
		MaxSendMsgBytes:        cfg.GRPC.MaxSendMsgBytes,
	}, logger)
	if err != nil {
		logger.Error("cms.grpc.handler_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	grpcServer := grpctransport.NewGRPCServer(grpcHandler)
	grpcListener, err := net.Listen("tcp", cfg.GRPC.Address)
	if err != nil {
		logger.Error("cms.grpc.listen_failed", slog.String("addr", cfg.GRPC.Address), slog.String("error", err.Error()))
		os.Exit(1)
	}

	go func() {
		logger.Info("cms.http.started", slog.String("addr", cfg.HTTP.Address))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("cms.http.listen_failed", slog.String("error", err.Error()))
			stop()
		}
	}()

	go func() {
		logger.Info("cms.grpc.started", slog.String("addr", cfg.GRPC.Address))
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Error("cms.grpc.serve_failed", slog.String("error", err.Error()))
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancelShutdown()
	grpcStopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(grpcStopped)
	}()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("cms.http.shutdown_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	select {
	case <-grpcStopped:
	case <-shutdownCtx.Done():
		grpcServer.Stop()
	}

	logger.Info("cms.servers.stopped")
}
