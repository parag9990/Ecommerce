package main

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/clients"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/config"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/events"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/identifier"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/repository"
	ordergrpc "github.com/example/ecommerce-platform/backend/services/order-service/internal/transport/grpc"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/usecase"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("order.service.stopped", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", cfg.Database.DSN)
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(30)
	pingCtx, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	err = db.PingContext(pingCtx)
	cancelPing()
	if err != nil {
		return err
	}

	httpClient := &http.Client{Timeout: cfg.Downstream.RequestTimeout}
	cartClient, err := clients.NewCartHTTPClient(cfg.Downstream.CartBaseURL, httpClient)
	if err != nil {
		return err
	}
	productClient, err := clients.NewProductHTTPClient(cfg.Downstream.ProductBaseURL, httpClient, cfg.Downstream.ProductServiceToken)
	if err != nil {
		return err
	}
	paymentClient, err := clients.NewPaymentHTTPClient(
		cfg.Downstream.PaymentBaseURL,
		cfg.Downstream.PaymentInternalToken,
		cfg.Downstream.PaymentActionTTL,
		httpClient,
	)
	if err != nil {
		return err
	}

	orders, err := repository.NewMySQLOrderRepository(db)
	if err != nil {
		return err
	}
	idempotency, err := repository.NewMySQLIdempotencyRepository(db)
	if err != nil {
		return err
	}
	ids := identifier.NewCryptoGenerator()
	checkout, err := usecase.NewCreateOrderFromCartUsecase(cartClient, productClient, orders, ids, cfg.CreateOrderFromCartConfig(), logger)
	if err != nil {
		return err
	}
	payment, err := usecase.NewInitiateOrderPaymentUsecase(orders, paymentClient, productClient, ids, cfg.InitiateOrderPaymentConfig(), logger)
	if err != nil {
		return err
	}
	paymentResults, err := usecase.NewApplyPaymentResultUsecase(orders, productClient, ids, cfg.ApplyPaymentResultConfig(), logger)
	if err != nil {
		return err
	}
	grpcServer, err := ordergrpc.NewApplicationServer(ordergrpc.ApplicationDependencies{
		Checkout: checkout, Payment: payment, Orders: orders, Idempotency: idempotency,
		CreateOrderConfig: cfg.CreateOrderConfig(), IDs: ids,
		PageTokenSigningKey: cfg.PageTokenSigningKey(), TrustedCallerToken: cfg.GRPC.TrustedCallerToken,
		Logger: logger,
	})
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if cfg.Events.Enabled {
		outbox, repoErr := repository.NewMySQLOutboxRepository(db)
		if repoErr != nil {
			return repoErr
		}
		publisher, publisherErr := events.NewKafkaPublisher(cfg.KafkaPublisherConfig())
		if publisherErr != nil {
			return publisherErr
		}
		defer publisher.Close()
		worker, workerErr := events.NewOutboxWorker(outbox, publisher, cfg.OrderOutboxWorkerConfig(), logger)
		if workerErr != nil {
			return workerErr
		}
		go worker.Run(ctx)
	}

	healthServer := &http.Server{
		Addr: cfg.HTTP.Address, Handler: serviceMux(db, paymentResults, orders, cfg.HTTP.AdminToken, cfg.Downstream.PaymentEventsToken, logger),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second,
	}
	errCh := make(chan error, 2)
	go func() { errCh <- ordergrpc.ListenAndServe(ctx, cfg.GRPC.Address, grpcServer) }()
	go func() {
		serveErr := healthServer.ListenAndServe()
		if errors.Is(serveErr, http.ErrServerClosed) {
			serveErr = nil
		}
		errCh <- serveErr
	}()
	logger.Info("order.service.started", slog.String("grpc_address", cfg.GRPC.Address), slog.String("http_address", cfg.HTTP.Address))

	select {
	case <-ctx.Done():
		err = nil
	case err = <-errCh:
		stop()
	}
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if shutdownErr := healthServer.Shutdown(shutdownCtx); err == nil && shutdownErr != nil {
		err = shutdownErr
	}
	return err
}

type paymentResultExecutor interface {
	Execute(context.Context, usecase.ApplyPaymentResultCommand) error
}

func serviceMux(db *sql.DB, paymentResults paymentResultExecutor, orders adminOrderRepository, adminToken string, paymentEventsToken string, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("POST /internal/v1/payment-events", func(w http.ResponseWriter, r *http.Request) {
		if !authorizedBearer(r.Header.Get("Authorization"), paymentEventsToken) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		defer r.Body.Close()
		var envelope paymentEventEnvelope
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&envelope); err != nil || decoder.Decode(&struct{}{}) != io.EOF || strings.TrimSpace(envelope.Topic) != "payment.events" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payment event"})
			return
		}
		result := ""
		switch envelope.Event.EventType {
		case "PaymentCaptured", "DuplicateCaptureDetected":
			result = "captured"
		case "PaymentFailed":
			result = "failed"
		default:
			writeJSON(w, http.StatusAccepted, map[string]string{"status": "ignored"})
			return
		}
		if err := paymentResults.Execute(r.Context(), usecase.ApplyPaymentResultCommand{
			PaymentID: envelope.Event.PaymentID, OrderID: envelope.Event.OrderID,
			TraceID: r.Header.Get("X-Trace-ID"), Result: result,
			Amount: envelope.Event.Amount, Currency: envelope.Event.Currency,
			ProviderEventID: envelope.Event.EventID, OccurredAt: envelope.Event.OccurredAt,
		}); err != nil {
			logger.Error("order.payment.event_failed", slog.String("event_id", envelope.Event.EventID), slog.String("error", err.Error()))
			writeJSON(w, http.StatusConflict, map[string]string{"error": "payment event could not be applied"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "applied"})
	})
	(&adminOrderHandler{token: adminToken, orders: orders}).register(mux)
	return mux
}

type paymentEventEnvelope struct {
	Topic string `json:"topic"`
	Event struct {
		EventID           string    `json:"event_id"`
		EventType         string    `json:"event_type"`
		PaymentID         string    `json:"payment_id"`
		OrderID           string    `json:"order_id"`
		Provider          string    `json:"provider"`
		ProviderPaymentID string    `json:"provider_payment_id"`
		Amount            int64     `json:"amount"`
		Currency          string    `json:"currency"`
		Reason            string    `json:"reason"`
		OccurredAt        time.Time `json:"occurred_at"`
	} `json:"event"`
}

func authorizedBearer(header string, expected string) bool {
	provided := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	expected = strings.TrimSpace(expected)
	if provided == "" || expected == "" || len(provided) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
