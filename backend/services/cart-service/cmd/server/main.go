package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	productclient "github.com/example/ecommerce-platform/backend/services/cart-service/internal/client"
	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/config"
	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/repository"
	httptransport "github.com/example/ecommerce-platform/backend/services/cart-service/internal/transport/http"
	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/usecase"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("cart.config.load_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger = logger.With(
		slog.String("service", cfg.ServiceName),
		slog.String("environment", cfg.Environment),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	connectCtx, cancelConnect := context.WithTimeout(ctx, cfg.Mongo.ConnectTimeout)
	defer cancelConnect()
	client, err := mongo.Connect(connectCtx, options.Client().ApplyURI(cfg.Mongo.URI).SetConnectTimeout(cfg.Mongo.ConnectTimeout))
	if err != nil {
		logger.Error("cart.mongo.connect_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		disconnectCtx, cancelDisconnect := context.WithTimeout(context.Background(), cfg.Mongo.ConnectTimeout)
		defer cancelDisconnect()
		if err := client.Disconnect(disconnectCtx); err != nil {
			logger.Error("cart.mongo.disconnect_failed", slog.String("error", err.Error()))
		}
	}()

	pingCtx, cancelPing := context.WithTimeout(ctx, cfg.Mongo.PingTimeout)
	defer cancelPing()
	if err := client.Ping(pingCtx, nil); err != nil {
		logger.Error("cart.mongo.ping_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Address,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("cart.redis.close_failed", slog.String("error", err.Error()))
		}
	}()
	redisPingCtx, cancelRedisPing := context.WithTimeout(ctx, cfg.Redis.PingTimeout)
	defer cancelRedisPing()
	if err := redisClient.Ping(redisPingCtx).Err(); err != nil {
		logger.Error("cart.redis.ping_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	db := client.Database(cfg.Mongo.Database)
	collectionManager, err := repository.NewMongoCollectionManager(db, logger)
	if err != nil {
		logger.Error("cart.collection_manager.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	schemaUsecase, err := usecase.NewSchemaUsecase(collectionManager, logger)
	if err != nil {
		logger.Error("cart.schema_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if cfg.BootstrapOnStartup {
		if _, err := schemaUsecase.EnsureCollections(ctx); err != nil {
			logger.Error("cart.schema.bootstrap_startup_failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	cartRepository, err := repository.NewMongoCartRepository(db, logger)
	if err != nil {
		logger.Error("cart.repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	cartCache, err := repository.NewRedisCartCache(redisClient, cfg.Cache.ActiveUserTTL, cfg.Cache.ActiveGuestTTL, cfg.Cache.SummaryTTL)
	if err != nil {
		logger.Error("cart.cache.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	productClient, err := productclient.NewHTTPProductClient(cfg.Product.BaseURL, cfg.Product.RequestTimeout)
	if err != nil {
		logger.Error("cart.product_client.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	couponValidator, err := productclient.NewHTTPCouponValidator(cfg.CMS.BaseURL, cfg.CMS.ValidateCouponPath, cfg.CMS.RequestTimeout)
	if err != nil {
		logger.Error("cart.coupon_validator.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	addItemUsecase, err := usecase.NewAddItemUsecase(usecase.AddItemDependencies{
		Repository:      cartRepository,
		Cache:           cartCache,
		ProductClient:   productClient,
		IDGenerator:     usecase.CryptoIDGenerator{},
		Clock:           usecase.SystemClock{},
		Logger:          logger,
		CartTTL:         cfg.Cart.ExpiryTTL,
		UserCartTTL:     cfg.Cart.UserExpiryTTL,
		GuestCartTTL:    cfg.Cart.GuestExpiryTTL,
		MaxSaveAttempts: cfg.Cart.MaxSaveAttempts,
	})
	if err != nil {
		logger.Error("cart.add_item_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	removeItemUsecase, err := usecase.NewRemoveItemUsecase(usecase.RemoveItemDependencies{
		Repository:      cartRepository,
		Cache:           cartCache,
		Clock:           usecase.SystemClock{},
		Logger:          logger,
		CartTTL:         cfg.Cart.ExpiryTTL,
		UserCartTTL:     cfg.Cart.UserExpiryTTL,
		GuestCartTTL:    cfg.Cart.GuestExpiryTTL,
		DefaultCurrency: cfg.Cart.DefaultCurrency,
		MaxSaveAttempts: cfg.Cart.MaxSaveAttempts,
	})
	if err != nil {
		logger.Error("cart.remove_item_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	applyCouponPreviewUsecase, err := usecase.NewApplyCouponPreviewUsecase(usecase.ApplyCouponPreviewDependencies{
		Repository:      cartRepository,
		Cache:           cartCache,
		CouponValidator: couponValidator,
		Clock:           usecase.SystemClock{},
		Logger:          logger,
		CartTTL:         cfg.Cart.ExpiryTTL,
		UserCartTTL:     cfg.Cart.UserExpiryTTL,
		GuestCartTTL:    cfg.Cart.GuestExpiryTTL,
		RequestTimeout:  cfg.CMS.RequestTimeout,
		MaxSaveAttempts: cfg.Cart.MaxSaveAttempts,
	})
	if err != nil {
		logger.Error("cart.apply_coupon_preview_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	mergeGuestCartUsecase, err := usecase.NewMergeGuestCartUsecase(usecase.MergeGuestCartDependencies{
		Repository:                cartRepository,
		Cache:                     cartCache,
		IDGenerator:               usecase.CryptoIDGenerator{},
		Clock:                     usecase.SystemClock{},
		Logger:                    logger,
		CartTTL:                   cfg.Cart.ExpiryTTL,
		UserCartTTL:               cfg.Cart.UserExpiryTTL,
		GuestCartTTL:              cfg.Cart.GuestExpiryTTL,
		DefaultCurrency:           cfg.Cart.DefaultCurrency,
		MaxSaveAttempts:           cfg.Cart.MaxSaveAttempts,
		AllowGuestCartIDOnlyMerge: cfg.Cart.AllowGuestCartIDOnlyMerge,
	})
	if err != nil {
		logger.Error("cart.merge_guest_cart_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	cartUsecase, err := usecase.NewCartMutationUsecase(addItemUsecase, removeItemUsecase, applyCouponPreviewUsecase, mergeGuestCartUsecase)
	if err != nil {
		logger.Error("cart.mutation_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	handler, err := httptransport.NewHandler(schemaUsecase, cartUsecase, logger, httptransport.WithCartReader(cartRepository))
	if err != nil {
		logger.Error("cart.http_handler.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	server := &http.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      httptransport.NewRouter(handler),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("cart.http.starting", slog.String("addr", cfg.HTTP.Address))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		if err != nil {
			logger.Error("cart.http.failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	case <-ctx.Done():
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancelShutdown()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("cart.http.shutdown_failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
		logger.Info("cart.http.stopped")
	}
}
