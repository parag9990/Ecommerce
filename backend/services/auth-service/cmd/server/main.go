package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/clients"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/config"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/events"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/repository"
	otpsecurity "github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/otp"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/password"
	tokensecurity "github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/token"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/sessionlink"
	httptransport "github.com/example/ecommerce-platform/backend/services/auth-service/internal/transport/http"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/usecase"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("auth.config.load_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := sql.Open("mysql", cfg.Database.DSN)
	if err != nil {
		logger.Error("auth.mysql.open_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPing()
	if err := db.PingContext(pingCtx); err != nil {
		logger.Error("auth.mysql.ping_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	redisClient, err := repository.NewRedisClient(repository.RedisClientConfig{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})
	if err != nil {
		logger.Error("auth.redis.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	redisCtx, cancelRedis := context.WithTimeout(ctx, 5*time.Second)
	defer cancelRedis()
	if err := redisClient.Ping(redisCtx); err != nil {
		logger.Error("auth.redis.ping_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	passwordHasher, err := password.NewRouter(cfg.PasswordRouterConfig())
	if err != nil {
		logger.Error("auth.password.hasher_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	otpGenerator, err := otpsecurity.NewGenerator(cfg.OTP.Length)
	if err != nil {
		logger.Error("auth.otp.generator_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	otpHasher, err := otpsecurity.NewHasher(cfg.OTP.HashPepper)
	if err != nil {
		logger.Error("auth.otp.hasher_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	notificationClient, err := clients.NewHTTPNotificationClient(cfg.Notification.OTPEndpoint, cfg.Notification.Timeout)
	if err != nil {
		logger.Error("auth.notification_client.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	credentialRepo, err := repository.NewMySQLCredentialRepository(db)
	if err != nil {
		logger.Error("auth.credential_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	accountRepo, err := repository.NewMySQLAccountRepository(db)
	if err != nil {
		logger.Error("auth.account_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	roleRepo, err := repository.NewMySQLRoleRepository(db)
	if err != nil {
		logger.Error("auth.role_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	tokenRepo, err := repository.NewMySQLTokenRepository(db)
	if err != nil {
		logger.Error("auth.token_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	outboxRepo, err := repository.NewMySQLOutboxRepository(db)
	if err != nil {
		logger.Error("auth.outbox_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	otpRepo, err := repository.NewMySQLOTPRepository(db)
	if err != nil {
		logger.Error("auth.otp_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	otpRateRepo, err := repository.NewRedisOTPRateRepository(redisClient, cfg.OTPPolicy(), cfg.OTP.RateLimitPepper)
	if err != nil {
		logger.Error("auth.otp_rate_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	keySet, err := tokensecurity.LoadRSAKeySet(cfg.Token.KeyID, cfg.Token.PrivateKeyPEMPath, cfg.Token.PublicKeyPEMPath)
	if err != nil {
		logger.Error("auth.jwt.keys_load_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	accessIssuer, err := tokensecurity.NewIssuer(tokensecurity.IssuerConfig{
		Issuer:           cfg.Token.Issuer,
		Audience:         cfg.Token.Audience,
		AccessTokenTTL:   cfg.Token.AccessTokenTTL,
		SigningAlgorithm: cfg.Token.SigningAlgorithm,
	}, keySet)
	if err != nil {
		logger.Error("auth.jwt.issuer_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	accessVerifier, err := tokensecurity.NewVerifier(tokensecurity.VerifierConfig{
		Issuer:           cfg.Token.Issuer,
		Audience:         cfg.Token.Audience,
		ClockSkew:        cfg.Token.ClockSkew,
		SigningAlgorithm: cfg.Token.SigningAlgorithm,
	}, keySet)
	if err != nil {
		logger.Error("auth.jwt.verifier_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	sessionLinker, privacyHasher, err := buildSessionLink(cfg, outboxRepo, logger)
	if err != nil {
		logger.Error("auth.session_link.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	passwordUsecase, err := usecase.NewPasswordUsecase(
		credentialRepo,
		passwordHasher,
		usecase.PasswordSecurityConfig{
			MaxFailedAttempts: cfg.Password.MaxFailedAttempts,
			LockoutDuration:   cfg.Password.LockoutDuration,
		},
		logger,
	)
	if err != nil {
		logger.Error("auth.password_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	tokenUsecase, err := usecase.NewTokenUsecase(
		tokenRepo,
		accountRepo,
		roleRepo,
		accessIssuer,
		accessVerifier,
		sessionLinker,
		privacyHasher,
		tokensecurity.BuildJWKS(keySet),
		usecase.TokenUsecaseConfig{
			RefreshTokenTTL:    cfg.Token.RefreshTokenTTL,
			RefreshTokenPepper: cfg.Token.RefreshTokenPepper,
		},
		logger,
	)
	if err != nil {
		logger.Error("auth.token_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	authUsecase, err := usecase.NewAuthUsecase(passwordUsecase, tokenUsecase, sessionLinker, privacyHasher, logger)
	if err != nil {
		logger.Error("auth.auth_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	otpUsecase, err := usecase.NewOTPUsecase(
		otpRepo,
		otpRateRepo,
		notificationClient,
		otpGenerator,
		otpHasher,
		usecase.OTPUsecaseConfig{Policy: cfg.OTPPolicy()},
		logger,
	)
	if err != nil {
		logger.Error("auth.otp_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	roleUsecase, err := usecase.NewRoleUsecase(
		roleRepo,
		usecase.RoleUsecaseConfig{
			ReasonMaxLength: cfg.RBAC.RoleMutationReasonMaxLength,
		},
		logger,
	)
	if err != nil {
		logger.Error("auth.role_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	handler, err := httptransport.NewHandler(passwordUsecase, authUsecase, tokenUsecase, otpUsecase, roleUsecase, logger)
	if err != nil {
		logger.Error("auth.http.handler_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	server := &http.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      httptransport.NewRouter(handler),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	if cfg.SessionLink.Mode == "outbox" && cfg.SessionLink.HTTPPublishEndpoint != "" {
		publisher, err := events.NewHTTPPublisher(cfg.SessionLink.HTTPPublishEndpoint, cfg.SessionLink.Timeout, logger)
		if err != nil {
			logger.Error("auth.outbox.publisher_init_failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
		worker, err := events.NewOutboxWorker(
			outboxRepo,
			publisher,
			events.OutboxWorkerConfig{
				Topic:            cfg.SessionLink.AuthEventsTopic,
				BatchSize:        cfg.SessionLink.OutboxWorker.BatchSize,
				Interval:         cfg.SessionLink.OutboxWorker.Interval,
				MaxAttempts:      cfg.SessionLink.OutboxWorker.MaxAttempts,
				InitialBackoff:   cfg.SessionLink.OutboxWorker.InitialBackoff,
				MaxBackoff:       cfg.SessionLink.OutboxWorker.MaxBackoff,
				StaleLockTimeout: cfg.SessionLink.OutboxWorker.StaleLockTimeout,
			},
			logger,
		)
		if err != nil {
			logger.Error("auth.outbox.worker_init_failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
		go worker.Run(ctx)
	} else if cfg.SessionLink.Mode == "outbox" {
		logger.Warn("auth.outbox.publisher_disabled",
			slog.String("reason", "AUTH_EVENTS_PUBLISH_ENDPOINT is empty"),
			slog.String("topic", cfg.SessionLink.AuthEventsTopic),
		)
	}

	go func() {
		logger.Info("auth.http.started", slog.String("addr", cfg.HTTP.Address))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("auth.http.listen_failed", slog.String("error", err.Error()))
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("auth.http.shutdown_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("auth.http.stopped")
}

func buildSessionLink(cfg config.Config, outboxRepo *repository.MySQLOutboxRepository, logger *slog.Logger) (usecase.SessionLinker, usecase.SessionPrivacyHasher, error) {
	if cfg.SessionLink.Mode == "disabled" {
		return sessionlink.NewDisabledLinker(), sessionlink.NewDisabledPrivacyHasher(), nil
	}

	privacyHasher, err := sessionlink.NewPrivacyHasher(cfg.SessionLink.EventPepper)
	if err != nil {
		return nil, nil, err
	}
	linker, err := sessionlink.NewOutboxLinker(outboxRepo, logger)
	if err != nil {
		return nil, nil, err
	}
	return linker, privacyHasher, nil
}
