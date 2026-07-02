package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	notificationv1 "github.com/example/ecommerce-platform/backend/services/notification-service/api/notification/v1"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/analytics"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/config"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/events"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/repository"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/retry"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/security"
	transportgrpc "github.com/example/ecommerce-platform/backend/services/notification-service/internal/transport/grpc"
	transporthttp "github.com/example/ecommerce-platform/backend/services/notification-service/internal/transport/http"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/usecase"
	platformmiddleware "github.com/parag/ecommerce/backend/shared/platform/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"google.golang.org/grpc"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, logger); err != nil {
		logger.Error("notification.server.failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	startupCtx, cancelStartup := context.WithTimeout(ctx, cfg.GRPC.StartupTimeout)
	defer cancelStartup()
	notificationRepository, mongoClient, err := repository.OpenMongoNotificationRepository(
		startupCtx,
		cfg.Mongo.URI,
		cfg.Mongo.Database,
		cfg.Mongo.TemplatesCollection,
		cfg.Mongo.DeliveriesCollection,
		cfg.Mongo.PreferencesCollection,
		cfg.Mongo.ProviderEventsCollection,
	)
	if err != nil {
		return err
	}
	defer disconnectMongo(mongoClient, cfg.GRPC.ShutdownTimeout, logger)

	renderer, err := usecase.NewTemplateRenderer(notificationRepository)
	if err != nil {
		return fmt.Errorf("create template renderer: %w", err)
	}
	registry, err := newProviderRegistry(cfg)
	if err != nil {
		return fmt.Errorf("create provider registry: %w", err)
	}
	observer := analytics.Observer(analytics.NoopObserver{})
	var metricsRegistry *prometheus.Registry
	if cfg.Analytics.MetricsEnabled {
		metricsRegistry = prometheus.NewRegistry()
		observer, err = analytics.NewPrometheusObserver(metricsRegistry)
		if err != nil {
			return fmt.Errorf("create notification metrics observer: %w", err)
		}
	}
	analyticsRecorder, err := analytics.NewRecorder(notificationRepository, observer)
	if err != nil {
		return fmt.Errorf("create notification analytics recorder: %w", err)
	}
	sendOTP, err := usecase.NewSendOTPService(renderer, registry, notificationRepository, analyticsRecorder, logger)
	if err != nil {
		return fmt.Errorf("create send OTP service: %w", err)
	}
	preferences, err := usecase.NewManagePreferenceService(notificationRepository)
	if err != nil {
		return fmt.Errorf("create notification preference service: %w", err)
	}
	consent, err := usecase.NewConsentGate(preferences)
	if err != nil {
		return fmt.Errorf("create notification consent gate: %w", err)
	}
	handler, err := transportgrpc.NewNotificationHandler(sendOTP, preferences, logger)
	if err != nil {
		return fmt.Errorf("create notification gRPC handler: %w", err)
	}

	var eventConsumer *events.RabbitConsumer
	var eventQueues events.QueueNames
	var deliveryQueue *retry.RabbitQueue
	var deliveryWorker *retry.Worker
	if cfg.RabbitMQ.Enabled {
		policy, policyErr := retry.NewPolicy(cfg.Retry.MaxAttempts, cfg.Retry.Delays)
		if policyErr != nil {
			return fmt.Errorf("create notification retry policy: %w", policyErr)
		}
		protector, protectorErr := security.NewRecipientProtector(cfg.Retry.DeliveryEncryptionKey)
		if protectorErr != nil {
			return fmt.Errorf("create notification recipient protector: %w", protectorErr)
		}
		deliveryQueue, err = retry.OpenRabbitQueue(
			startupCtx, cfg.RabbitMQ.URL, cfg.Retry.Prefetch, policy,
			cfg.Retry.PublishConfirmTimeout, logger,
		)
		if err != nil {
			return err
		}
		defer func() {
			if err := deliveryQueue.Close(); err != nil {
				logger.Warn("notification.retry.rabbitmq.close_failed", slog.String("error", err.Error()))
			}
		}()
		attemptSender, senderErr := usecase.NewDeliveryAttemptService(renderer, registry, notificationRepository, protector, consent)
		if senderErr != nil {
			return fmt.Errorf("create notification delivery attempt sender: %w", senderErr)
		}
		deliveryWorker, err = retry.NewWorker(
			policy, notificationRepository, attemptSender, deliveryQueue.Publisher(),
			analyticsRecorder,
			cfg.Retry.AttemptLease, logger,
		)
		if err != nil {
			return fmt.Errorf("create notification retry worker: %w", err)
		}
		eventConsumer, eventQueues, err = newEventConsumer(
			startupCtx, cfg, notificationRepository, deliveryQueue.Publisher(), protector, logger,
		)
		if err != nil {
			return err
		}
		defer func() {
			if err := eventConsumer.Close(); err != nil {
				logger.Warn("notification.rabbitmq.close_failed", slog.String("error", err.Error()))
			}
		}()
	}

	listener, err := net.Listen("tcp", cfg.GRPC.Address)
	if err != nil {
		return fmt.Errorf("listen for notification gRPC: %w", err)
	}
	defer listener.Close()

	readinessCheck := func(checkCtx context.Context) error {
		if err := mongoClient.Ping(checkCtx, readpref.Primary()); err != nil {
			return fmt.Errorf("ping notification MongoDB: %w", err)
		}
		if eventConsumer != nil && !eventConsumer.Ready() {
			return errors.New("notification event consumer is not ready")
		}
		if deliveryQueue != nil && !deliveryQueue.Ready() {
			return errors.New("notification retry queue is not ready")
		}
		return nil
	}
	internalServer, internalListener, err := newAnalyticsHTTPServer(cfg, analyticsRecorder, observer, metricsRegistry, readinessCheck, logger)
	if err != nil {
		return err
	}
	if internalListener != nil {
		defer internalListener.Close()
	}

	server := grpc.NewServer()
	notificationv1.RegisterNotificationServiceServer(server, handler)
	runtimeCtx, cancelRuntime := context.WithCancel(ctx)
	defer cancelRuntime()
	serveErr := make(chan error, 1)
	go func() {
		logger.Info("notification.grpc.started", slog.String("addr", cfg.GRPC.Address))
		serveErr <- server.Serve(listener)
	}()
	var internalServeErr chan error
	if internalServer != nil {
		internalServeErr = make(chan error, 1)
		go func() {
			logger.Info("notification.analytics_http.started", slog.String("addr", cfg.Analytics.HTTPAddress))
			internalServeErr <- internalServer.Serve(internalListener)
		}()
	}
	var consumerErr chan error
	consumerCount := 0
	if eventConsumer != nil {
		consumerErr = make(chan error, 2)
		consumerCount = 2
		go func() {
			logger.Info("notification.events.rabbitmq.started")
			consumerErr <- eventConsumer.Run(runtimeCtx, eventQueues)
		}()
		go func() {
			logger.Info("notification.retry.rabbitmq.started")
			consumerErr <- deliveryQueue.Run(runtimeCtx, deliveryWorker)
		}()
	}

	select {
	case <-ctx.Done():
		cancelRuntime()
		gracefulStop(server, cfg.GRPC.ShutdownTimeout)
		shutdownHTTP(internalServer, cfg.GRPC.ShutdownTimeout, logger)
		waitForRabbitConsumers(consumerErr, consumerCount, cfg.GRPC.ShutdownTimeout, logger)
		logger.Info("notification.grpc.stopped")
		return nil
	case err := <-consumerErr:
		cancelRuntime()
		gracefulStop(server, cfg.GRPC.ShutdownTimeout)
		shutdownHTTP(internalServer, cfg.GRPC.ShutdownTimeout, logger)
		waitForRabbitConsumers(consumerErr, consumerCount-1, cfg.GRPC.ShutdownTimeout, logger)
		if err == nil {
			return errors.New("notification RabbitMQ consumer stopped unexpectedly")
		}
		return fmt.Errorf("consume notification RabbitMQ work: %w", err)
	case err := <-serveErr:
		cancelRuntime()
		shutdownHTTP(internalServer, cfg.GRPC.ShutdownTimeout, logger)
		waitForRabbitConsumers(consumerErr, consumerCount, cfg.GRPC.ShutdownTimeout, logger)
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		return fmt.Errorf("serve notification gRPC: %w", err)
	case err := <-internalServeErr:
		cancelRuntime()
		gracefulStop(server, cfg.GRPC.ShutdownTimeout)
		waitForRabbitConsumers(consumerErr, consumerCount, cfg.GRPC.ShutdownTimeout, logger)
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve notification analytics HTTP: %w", err)
	}
}

func newAnalyticsHTTPServer(
	cfg config.Config,
	recorder *analytics.Recorder,
	observer analytics.Observer,
	metricsRegistry *prometheus.Registry,
	readinessCheck transporthttp.ReadinessCheck,
	logger *slog.Logger,
) (*http.Server, net.Listener, error) {
	mux := http.NewServeMux()
	transporthttp.NewHealthHandler(readinessCheck, 2*time.Second).Register(mux)
	if cfg.Analytics.MetricsEnabled {
		mux.Handle(cfg.Analytics.MetricsPath, transporthttp.MetricsHandler(metricsRegistry))
	}
	if cfg.Analytics.WebhooksEnabled {
		endpoints := make([]transporthttp.WebhookEndpoint, 0, len(domain.SupportedChannels()))
		for _, channel := range domain.SupportedChannels() {
			if secret := cfg.Analytics.SigningSecret(channel); secret != "" {
				endpoints = append(endpoints, transporthttp.WebhookEndpoint{
					Channel: channel, Provider: cfg.ProviderName(channel), SigningSecret: secret,
					AllowOpenTracking: cfg.Analytics.AllowsOpenTracking(channel),
				})
			}
		}
		handler, err := transporthttp.NewProviderWebhookHandler(
			endpoints, recorder, observer, cfg.Analytics.WebhookMaxBodyBytes,
			cfg.Analytics.WebhookReplayWindow, logger,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("create provider webhook handler: %w", err)
		}
		if err := handler.Register(mux, cfg.Analytics.WebhookPathPrefix); err != nil {
			return nil, nil, fmt.Errorf("register provider webhook handler: %w", err)
		}
	}
	listener, err := net.Listen("tcp", strings.TrimSpace(cfg.Analytics.HTTPAddress))
	if err != nil {
		return nil, nil, fmt.Errorf("listen for notification analytics HTTP: %w", err)
	}
	return &http.Server{
		Handler:           platformmiddleware.CORS(platformmiddleware.DefaultCORSConfig())(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}, listener, nil
}

func newEventConsumer(
	ctx context.Context,
	cfg config.Config,
	deliveries usecase.RetryEventRepository,
	publisher usecase.AttemptPublisher,
	protector usecase.RecipientProtector,
	logger *slog.Logger,
) (*events.RabbitConsumer, events.QueueNames, error) {
	sender, err := usecase.NewQueueEventService(deliveries, publisher, protector, cfg.Retry.MaxAttempts, logger)
	if err != nil {
		return nil, events.QueueNames{}, fmt.Errorf("create event notification service: %w", err)
	}
	handler, err := events.NewHandler(sender, logger)
	if err != nil {
		return nil, events.QueueNames{}, fmt.Errorf("create event notification handler: %w", err)
	}
	consumer, err := events.OpenRabbitConsumer(ctx, cfg.RabbitMQ.URL, cfg.RabbitMQ.Prefetch, handler, logger)
	if err != nil {
		return nil, events.QueueNames{}, err
	}
	queues := events.QueueNames{
		OrderEvents:   cfg.RabbitMQ.OrderEventsQueue,
		PaymentEvents: cfg.RabbitMQ.PaymentEventsQueue,
		UserEvents:    cfg.RabbitMQ.UserEventsQueue,
	}
	if err := consumer.DeclareTopology(queues); err != nil {
		_ = consumer.Close()
		return nil, events.QueueNames{}, fmt.Errorf("declare notification event topology: %w", err)
	}
	return consumer, queues, nil
}

func waitForRabbitConsumers(result <-chan error, count int, timeout time.Duration, logger *slog.Logger) {
	if result == nil || count < 1 {
		return
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for remaining := count; remaining > 0; remaining-- {
		select {
		case err := <-result:
			if err != nil {
				logger.Warn("notification.rabbitmq.stopped_with_error", slog.String("error", err.Error()))
			}
		case <-timer.C:
			logger.Warn("notification.rabbitmq.shutdown_timeout")
			return
		}
	}
}

func newProviderRegistry(cfg config.Config) (*provider.Registry, error) {
	var providers []provider.Provider
	if cfg.Email.Enabled {
		client, err := provider.NewSMTPClient(provider.SMTPClientConfig{
			Host:     cfg.EmailSMTP.Host,
			Port:     cfg.EmailSMTP.Port,
			From:     cfg.EmailSMTP.From,
			Username: cfg.EmailSMTP.Username,
			Password: cfg.EmailSMTP.Password,
			TLSMode:  cfg.EmailSMTP.TLSMode,
			Timeout:  cfg.EmailSMTP.Timeout,
		})
		if err != nil {
			return nil, err
		}
		email, err := provider.NewEmailProvider(cfg.Email.ProviderName, client)
		if err != nil {
			return nil, err
		}
		providers = append(providers, email)
	}
	if cfg.SMS.Enabled {
		client, err := provider.NewHTTPSMSClient(provider.HTTPSMSClientConfig{
			Endpoint:    cfg.SMSHTTP.Endpoint,
			BearerToken: cfg.SMSHTTP.BearerToken,
			Timeout:     cfg.SMSHTTP.Timeout,
		})
		if err != nil {
			return nil, err
		}
		sms, err := provider.NewSMSProvider(cfg.SMS.ProviderName, client)
		if err != nil {
			return nil, err
		}
		providers = append(providers, sms)
	}
	return provider.NewConfiguredRegistry(cfg, providers...)
}

func gracefulStop(server *grpc.Server, timeout time.Duration) {
	stopped := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(stopped)
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-stopped:
	case <-timer.C:
		server.Stop()
	}
}

func shutdownHTTP(server *http.Server, timeout time.Duration, logger *slog.Logger) {
	if server == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Warn("notification.analytics_http.shutdown_failed", slog.String("error", err.Error()))
	}
}

func disconnectMongo(client interface{ Disconnect(context.Context) error }, timeout time.Duration, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := client.Disconnect(ctx); err != nil {
		logger.Error("notification.mongo.disconnect_failed", slog.String("error", err.Error()))
	}
}
