package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/config"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/events"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/repository"
	httptransport "github.com/example/ecommerce-platform/backend/services/payment-service/internal/transport/http"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("payment.config.load_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	providerRegistry, err := provider.NewRegistryFromConfig(cfg.PaymentGateway, logger)
	if err != nil {
		logger.Error("payment.provider.registry_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("payment.provider.registry_ready",
		slog.String("default_provider", cfg.PaymentGateway.Normalized().DefaultProvider),
		slog.Any("providers", providerRegistry.Names()),
	)

	stateUsecase, err := usecase.NewPaymentStateUsecase(domain.NewStateMachine(), logger)
	if err != nil {
		logger.Error("payment.state_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	db, err := repository.OpenMySQL(cfg.Database.DSN)
	if err != nil {
		logger.Error("payment.mysql.open_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	paymentRepository, err := repository.NewMySQLPaymentRepository(db, logger)
	if err != nil {
		logger.Error("payment.repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	schemaUsecase, err := usecase.NewPaymentSchemaUsecase(paymentRepository, logger)
	if err != nil {
		logger.Error("payment.schema_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	paymentIntentUsecase, err := usecase.NewCreatePaymentIntentUsecase(paymentRepository, providerRegistry, cfg.PaymentGateway, logger)
	if err != nil {
		logger.Error("payment.intent_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var webhookUsecase httptransport.WebhookUsecase
	var refundUsecase httptransport.RefundUsecase
	var retryUsecase httptransport.RetryPaymentUsecase
	if len(providerRegistry.Names()) > 0 {
		eventPublisher, err := events.NewHTTPPublisher(cfg.EventPublisher)
		if err != nil {
			logger.Error("payment.event_publisher.init_failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
		webhookUsecase, err = usecase.NewHandleWebhookUsecase(paymentRepository, providerRegistry, eventPublisher, logger)
		if err != nil {
			logger.Error("payment.webhook_usecase.init_failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
		refundUsecase, err = usecase.NewRefundPaymentUsecase(
			paymentRepository,
			providerRegistry,
			eventPublisher,
			usecase.RefundPolicy{ManualReviewThresholdMinor: cfg.Refund.ManualReviewThresholdMinor},
			logger,
		)
		if err != nil {
			logger.Error("payment.refund_usecase.init_failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
		retryUsecase, err = usecase.NewRetryPaymentIntentUsecase(
			paymentRepository,
			providerRegistry,
			cfg.PaymentGateway,
			domain.RetryPolicy{
				MaxAttempts: uint32(cfg.Retry.MaxAttempts),
				Cooldown:    cfg.Retry.Cooldown,
			},
			logger,
		)
		if err != nil {
			logger.Error("payment.retry_usecase.init_failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	handler, err := httptransport.NewHandler(
		stateUsecase,
		logger,
		httptransport.WithSchemaUsecase(schemaUsecase),
		httptransport.WithPaymentIntentUsecase(paymentIntentUsecase),
		httptransport.WithRetryPaymentUsecase(retryUsecase),
		httptransport.WithPaymentIntentAuthorizationToken(cfg.InternalAPI.Token),
		httptransport.WithWebhookUsecase(webhookUsecase),
		httptransport.WithRefundUsecase(refundUsecase),
		httptransport.WithWebhookMaxBodyBytes(cfg.Webhook.MaxBodyBytes),
	)
	if err != nil {
		logger.Error("payment.http.handler_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	server := &http.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      httptransport.NewRouter(handler),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("payment.http.started", slog.String("addr", cfg.HTTP.Address))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("payment.http.listen_failed", slog.String("error", err.Error()))
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("payment.http.shutdown_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("payment.http.stopped")
}
