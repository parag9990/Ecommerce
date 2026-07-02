package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"

	"ecommerce/superadmin-service/internal/clients"
	"ecommerce/superadmin-service/internal/config"
	"ecommerce/superadmin-service/internal/logging"
	"ecommerce/superadmin-service/internal/rbac"
	"ecommerce/superadmin-service/internal/repository"
	transporthttp "ecommerce/superadmin-service/internal/transport/http"
	"ecommerce/superadmin-service/internal/usecase"
	platformmiddleware "github.com/parag/ecommerce/backend/shared/platform/middleware"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()
	logger := logging.New(cfg.ServiceName, cfg.LogLevel)

	if err := cfg.Validate(); err != nil {
		logger.Error(ctx, "invalid configuration", "error", err)
		os.Exit(1)
	}

	db, repo, err := buildPermissionRepository(ctx, cfg, logger)
	if err != nil {
		logger.Error(ctx, "permission repository setup failed", "error", err)
		os.Exit(1)
	}
	if db != nil {
		defer db.Close()
	}

	authz, err := usecase.NewAuthorizationService(repo, logger)
	if err != nil {
		logger.Error(ctx, "authorization service setup failed", "error", err)
		os.Exit(1)
	}

	userClient, err := buildUserServiceClient(cfg, logger)
	if err != nil {
		logger.Error(ctx, "user service client setup failed", "error", err)
		os.Exit(1)
	}

	reviewTasks, err := buildReviewTaskRepository(db)
	if err != nil {
		logger.Error(ctx, "review task repository setup failed", "error", err)
		os.Exit(1)
	}
	platformSettingsRepo, err := buildPlatformSettingsRepository(db)
	if err != nil {
		logger.Error(ctx, "platform settings repository setup failed", "error", err)
		os.Exit(1)
	}
	auditRepo, auditRecorder, err := buildAuditDependencies(db, logger)
	if err != nil {
		logger.Error(ctx, "audit log repository setup failed", "error", err)
		os.Exit(1)
	}

	controls, err := usecase.NewControlService(authz, userClient, reviewTasks, auditRecorder, logger)
	if err != nil {
		logger.Error(ctx, "control service setup failed", "error", err)
		os.Exit(1)
	}

	orderClient, err := buildOrderServiceClient(cfg, logger)
	if err != nil {
		logger.Error(ctx, "order service client setup failed", "error", err)
		os.Exit(1)
	}
	paymentClient, err := buildPaymentServiceClient(cfg, logger)
	if err != nil {
		logger.Error(ctx, "payment service client setup failed", "error", err)
		os.Exit(1)
	}
	orderPaymentControls, err := usecase.NewOrderPaymentControlService(authz, orderClient, paymentClient, reviewTasks, auditRecorder, logger)
	if err != nil {
		logger.Error(ctx, "order payment control service setup failed", "error", err)
		os.Exit(1)
	}
	sessionVisibility, err := usecase.NewSessionVisibilityServiceWithAudit(authz, usecase.SessionVisibilityConfig{
		MaxDateRange: cfg.SessionAnalyticsMaxRange,
		MaxPageSize:  cfg.SessionAnalyticsMaxPageSize,
	}, auditRecorder, logger)
	if err != nil {
		logger.Error(ctx, "session visibility service setup failed", "error", err)
		os.Exit(1)
	}
	sessionClient, err := buildSessionServiceClient(cfg, logger)
	if err != nil {
		logger.Error(ctx, "session service client setup failed", "error", err)
		os.Exit(1)
	}
	platformSettings, err := usecase.NewPlatformSettingsService(
		platformSettingsRepo,
		authz,
		clients.NewLoggingSettingsEventPublisher(cfg.PlatformSettingsEventTopic, logger),
		auditRecorder,
		usecase.PlatformSettingsConfig{CacheTTL: cfg.PlatformSettingsCacheTTL},
		logger,
	)
	if err != nil {
		logger.Error(ctx, "platform settings service setup failed", "error", err)
		os.Exit(1)
	}
	auditLogs, err := usecase.NewAuditLogService(authz, auditRepo, logger)
	if err != nil {
		logger.Error(ctx, "audit log service setup failed", "error", err)
		os.Exit(1)
	}

	rbacHandler := transporthttp.NewRBACHandler(authz, db, logger)
	controlHandler := transporthttp.NewControlHandler(controls, logger)
	orderPaymentHandler := transporthttp.NewOrderPaymentHandler(orderPaymentControls, logger)
	sessionVisibilityHandler := transporthttp.NewSessionVisibilityHandler(sessionVisibility, logger)
	sessionAnalyticsProxyHandler := transporthttp.NewSessionAnalyticsProxyHandler(sessionVisibility, sessionClient, logger)
	settingsHandler := transporthttp.NewSettingsHandler(platformSettings, logger)
	auditLogHandler := transporthttp.NewAuditLogHandler(auditLogs, logger)
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           platformmiddleware.CORS(platformmiddleware.DefaultCORSConfig())(transporthttp.NewServeMux(rbacHandler, controlHandler, orderPaymentHandler, sessionVisibilityHandler, sessionAnalyticsProxyHandler, settingsHandler, auditLogHandler)),
		ReadHeaderTimeout: cfg.HTTPReadHeaderTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info(ctx, "superadmin service listening", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		logger.Error(ctx, "http server failed", "error", err)
		os.Exit(1)
	case sig := <-signalCh:
		logger.Info(ctx, "shutdown signal received", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTPShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, "http server shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info(ctx, "superadmin service stopped")
}

func buildPermissionRepository(ctx context.Context, cfg config.Config, logger logging.Logger) (*sql.DB, usecase.PermissionRepository, error) {
	if cfg.DatabaseDSN == "" {
		logger.Warn(ctx, "SUPERADMIN_DATABASE_DSN is empty; RBAC checks will deny every admin except health endpoints")
		return nil, rbac.NewStaticPermissionRepository(nil), nil
	}

	db, err := sql.Open("mysql", cfg.DatabaseDSN)
	if err != nil {
		return nil, nil, err
	}
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, cfg.DBPingTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, nil, err
	}

	repo, err := repository.NewMySQLPermissionRepository(db)
	if err != nil {
		db.Close()
		return nil, nil, err
	}
	return db, repo, nil
}

func buildUserServiceClient(cfg config.Config, logger logging.Logger) (usecase.UserServiceClient, error) {
	if cfg.UserServiceBaseURL == "" {
		logger.Warn(context.Background(), "USER_SERVICE_ADMIN_BASE_URL is empty; user/seller control endpoints will return downstream unavailable")
		return clients.NewUnavailableUserServiceClient("user service is not configured"), nil
	}
	return clients.NewHTTPUserServiceClient(cfg.UserServiceBaseURL, cfg.UserServiceToken, cfg.UserServiceTimeout, logger)
}

func buildOrderServiceClient(cfg config.Config, logger logging.Logger) (usecase.OrderServiceClient, error) {
	if cfg.OrderServiceBaseURL == "" {
		logger.Warn(context.Background(), "ORDER_SERVICE_ADMIN_BASE_URL is empty; order control endpoints will return downstream unavailable")
		return clients.NewUnavailableOrderServiceClient("order service is not configured"), nil
	}
	return clients.NewHTTPOrderServiceClient(cfg.OrderServiceBaseURL, cfg.OrderServiceToken, cfg.OrderServiceTimeout, logger)
}

func buildPaymentServiceClient(cfg config.Config, logger logging.Logger) (usecase.PaymentServiceClient, error) {
	if cfg.PaymentServiceBaseURL == "" {
		logger.Warn(context.Background(), "PAYMENT_SERVICE_ADMIN_BASE_URL is empty; payment control endpoints will return downstream unavailable")
		return clients.NewUnavailablePaymentServiceClient("payment service is not configured"), nil
	}
	return clients.NewHTTPPaymentServiceClient(cfg.PaymentServiceBaseURL, cfg.PaymentServiceToken, cfg.PaymentServiceTimeout, logger)
}

func buildSessionServiceClient(cfg config.Config, logger logging.Logger) (transporthttp.SessionAnalyticsForwarder, error) {
	if cfg.SessionServiceBaseURL == "" {
		logger.Warn(context.Background(), "SESSION_SERVICE_ADMIN_BASE_URL is empty; session analytics endpoints will return downstream unavailable")
		return clients.NewUnavailableSessionServiceClient("session service is not configured"), nil
	}
	return clients.NewHTTPSessionServiceClient(cfg.SessionServiceBaseURL, cfg.SessionServiceToken, cfg.SessionServiceTimeout, logger)
}

func buildReviewTaskRepository(db *sql.DB) (*repository.MySQLReviewTaskRepository, error) {
	if db == nil {
		return nil, nil
	}
	return repository.NewMySQLReviewTaskRepository(db)
}

func buildPlatformSettingsRepository(db *sql.DB) (usecase.PlatformSettingsRepository, error) {
	if db == nil {
		return repository.NewUnavailablePlatformSettingsRepository("platform settings storage is not configured"), nil
	}
	return repository.NewMySQLPlatformSettingsRepository(db)
}

func buildAuditDependencies(db *sql.DB, logger logging.Logger) (usecase.AuditLogRepository, usecase.AuditRecorder, error) {
	if db == nil {
		return repository.NewUnavailableAuditLogRepository("admin audit log storage is not configured"), usecase.NewLoggingAuditRecorder(logger), nil
	}

	auditRepo, err := repository.NewMySQLAuditLogRepository(db)
	if err != nil {
		return nil, nil, err
	}
	auditRecorder, err := usecase.NewMySQLAuditRecorder(auditRepo)
	if err != nil {
		return nil, nil, err
	}
	return auditRepo, auditRecorder, nil
}
