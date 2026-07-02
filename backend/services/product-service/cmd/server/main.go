package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"product-service/internal/app"
	"product-service/internal/config"
	eventing "product-service/internal/events"
	"product-service/internal/repository"
	"product-service/internal/transport/dto"
	transportgrpc "product-service/internal/transport/grpc"
	"product-service/internal/usecase"

	productv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/product/v1"
	platformmiddleware "github.com/parag/ecommerce/backend/shared/platform/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const maxBodyBytes = 2 << 20

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "product-service failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel})).With("service", cfg.ServiceName)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mongoClient, database, err := repository.NewMongoDatabase(ctx, cfg.Mongo.URI, cfg.Mongo.DatabaseName)
	if err != nil {
		return err
	}
	defer mongoClient.Disconnect(context.Background())

	products, err := repository.NewMongoProductRepository(database, logger)
	if err != nil {
		return err
	}
	collections, err := repository.NewMongoCollectionManager(database, logger)
	if err != nil {
		return err
	}
	var eventPublisher eventing.ProductEventPublisher
	var kafkaPublisher *eventing.KafkaPublisher
	if cfg.Events.Enabled && cfg.Events.OutboxWorkerEnabled && strings.EqualFold(cfg.Events.Broker, "kafka") {
		kafkaPublisher, err = eventing.NewKafkaPublisher(cfg.Events.KafkaBrokers, time.Duration(cfg.Events.PublishTimeoutMS)*time.Millisecond)
		if err != nil {
			return err
		}
		defer kafkaPublisher.Close()
		eventPublisher = kafkaPublisher
	}
	application, err := app.New(ctx, cfg, app.Dependencies{
		ProductRepository:              products,
		CollectionSchemaManager:        collections,
		InventoryStockRepository:       products,
		InventoryReservationRepository: products,
		InventorySnapshotRepository:    products,
		ProductEventOutboxRepository:   products,
		ProductEventPublisher:          eventPublisher,
		Logger:                         logger,
	})
	if err != nil {
		return err
	}
	application.StartBackgroundWorkers(ctx)
	grpcHandler, err := transportgrpc.NewServer(application.ProductReadHandler, application.SellerProductHandler, application.InventoryHandler)
	if err != nil {
		return err
	}
	grpcListener, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		return fmt.Errorf("listen for product grpc on %s: %w", cfg.GRPCAddress, err)
	}
	defer grpcListener.Close()
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(internalGRPCAuth(cfg.InternalServiceToken)))
	productv1.RegisterProductServiceServer(grpcServer, grpcHandler)
	healthServer := health.NewServer()
	healthv1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthv1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(productv1.ProductService_ServiceDesc.ServiceName, healthv1.HealthCheckResponse_SERVING)

	mux := routes(application, func(probeCtx context.Context) error { return mongoClient.Ping(probeCtx, nil) })
	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           platformmiddleware.CORS(platformmiddleware.DefaultCORSConfig())(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	errCh := make(chan error, 1)
	grpcErrCh := make(chan error, 1)
	go func() {
		logger.Info("product.http.started", "addr", server.Addr)
		errCh <- server.ListenAndServe()
	}()
	go func() {
		logger.Info("product.grpc.started", "addr", cfg.GRPCAddress)
		grpcErrCh <- grpcServer.Serve(grpcListener)
	}()

	select {
	case <-ctx.Done():
		healthServer.SetServingStatus("", healthv1.HealthCheckResponse_NOT_SERVING)
		healthServer.SetServingStatus(productv1.ProductService_ServiceDesc.ServiceName, healthv1.HealthCheckResponse_NOT_SERVING)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		httpErr := server.Shutdown(shutdownCtx)
		grpcServer.GracefulStop()
		return httpErr
	case err := <-errCh:
		grpcServer.Stop()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case err := <-grpcErrCh:
		_ = server.Close()
		if errors.Is(err, grpc.ErrServerStopped) {
			return nil
		}
		return err
	}
}

func routes(application *app.App, ping func(context.Context) error) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("GET /api/v1/products", func(w http.ResponseWriter, r *http.Request) {
		if application.ProductReadHandler == nil {
			writeServiceError(w, errors.New("product read service is unavailable"))
			return
		}
		result, err := application.ProductReadHandler.ListProducts(r.Context(), dto.ListProductsRequestDTO{
			CategoryID: r.URL.Query().Get("category_id"), SellerID: r.URL.Query().Get("seller_id"),
			Status: r.URL.Query().Get("status"), Page: intQuery(r, "page", 1),
			PageSize: intQuery(r, "page_size", 20), Sort: r.URL.Query().Get("sort"),
		})
		respond(w, result, err)
	})
	mux.HandleFunc("GET /api/v1/products/{product_id}", func(w http.ResponseWriter, r *http.Request) {
		result, err := application.ProductReadHandler.GetProduct(r.Context(), dto.GetProductRequestDTO{ProductID: r.PathValue("product_id")})
		respond(w, result, err)
	})
	mux.HandleFunc("GET /internal/v1/products/{product_id}", func(w http.ResponseWriter, r *http.Request) {
		result, err := application.ProductReadHandler.GetSellerProduct(r.Context(), dto.GetSellerProductRequestDTO{
			Actor: internalActor(r, "", ""), ProductID: r.PathValue("product_id"),
		})
		respond(w, result, err)
	})
	mux.HandleFunc("GET /api/v1/categories", func(w http.ResponseWriter, r *http.Request) {
		result, err := application.ProductReadHandler.ListCategories(r.Context(), dto.ListCategoriesRequestDTO{ParentID: r.URL.Query().Get("parent_id")})
		respond(w, result, err)
	})
	mux.HandleFunc("POST /internal/v1/products/batch", func(w http.ResponseWriter, r *http.Request) {
		var input dto.BatchGetProductsRequestDTO
		if !decode(w, r, &input) {
			return
		}
		result, err := application.ProductReadHandler.BatchGetProducts(r.Context(), input)
		respond(w, result, err)
	})
	mux.HandleFunc("GET /internal/v1/products/search-export", func(w http.ResponseWriter, r *http.Request) {
		result, err := application.ProductReadHandler.ExportSearchProducts(r.Context(), dto.SearchProductExportRequestDTO{
			Cursor: r.URL.Query().Get("cursor"),
			Limit:  intQuery(r, "limit", 500),
		})
		respond(w, result, err)
	})
	mux.HandleFunc("GET /api/v1/seller/products", func(w http.ResponseWriter, r *http.Request) {
		result, err := application.ProductReadHandler.ListSellerProducts(r.Context(), dto.ListSellerProductsRequestDTO{
			Actor: actor(r), CategoryID: r.URL.Query().Get("category_id"), Status: r.URL.Query().Get("status"),
			Page: intQuery(r, "page", 1), PageSize: intQuery(r, "page_size", 20), Sort: r.URL.Query().Get("sort"),
		})
		respond(w, result, err)
	})
	mux.HandleFunc("GET /api/v1/seller/products/{product_id}", func(w http.ResponseWriter, r *http.Request) {
		result, err := application.ProductReadHandler.GetSellerProduct(r.Context(), dto.GetSellerProductRequestDTO{
			Actor: actor(r), ProductID: r.PathValue("product_id"),
		})
		respond(w, result, err)
	})
	mux.HandleFunc("POST /api/v1/seller/products", func(w http.ResponseWriter, r *http.Request) {
		var input dto.ProductInputDTO
		if !decode(w, r, &input) {
			return
		}
		result, err := application.SellerProductHandler.CreateProduct(r.Context(), dto.CreateSellerProductRequestDTO{Actor: actor(r), Product: input})
		respondStatus(w, http.StatusCreated, result, err)
	})
	mux.HandleFunc("PATCH /api/v1/seller/products/{product_id}", func(w http.ResponseWriter, r *http.Request) {
		var input dto.ProductInputDTO
		if !decode(w, r, &input) {
			return
		}
		result, err := application.SellerProductHandler.UpdateProduct(r.Context(), dto.UpdateSellerProductRequestDTO{
			Actor: actor(r), ProductID: r.PathValue("product_id"), Product: input,
		})
		respond(w, result, err)
	})
	mux.HandleFunc("POST /api/v1/seller/products/{product_id}/publish", func(w http.ResponseWriter, r *http.Request) {
		result, err := application.SellerProductHandler.PublishProduct(r.Context(), dto.ProductLifecycleRequestDTO{Actor: actor(r), ProductID: r.PathValue("product_id")})
		respond(w, result, err)
	})
	mux.HandleFunc("POST /internal/v1/products/{product_id}/publish", func(w http.ResponseWriter, r *http.Request) {
		var input internalStatusRequest
		if !decode(w, r, &input) {
			return
		}
		result, err := application.SellerProductHandler.ModerateProduct(r.Context(), dto.ProductModerationRequestDTO{
			Actor: internalActor(r, input.ActorUserID, input.SellerID), ProductID: r.PathValue("product_id"), ExpectedStatus: input.ExpectedStatus, Status: "published",
		})
		respond(w, result, err)
	})
	mux.HandleFunc("POST /internal/v1/products/{product_id}/unpublish", func(w http.ResponseWriter, r *http.Request) {
		var input internalStatusRequest
		if !decode(w, r, &input) {
			return
		}
		result, err := application.SellerProductHandler.ModerateProduct(r.Context(), dto.ProductModerationRequestDTO{
			Actor: internalActor(r, input.ActorUserID, input.SellerID), ProductID: r.PathValue("product_id"), ExpectedStatus: input.ExpectedStatus, Status: "unpublished",
		})
		respond(w, result, err)
	})
	mux.HandleFunc("PATCH /internal/v1/products/{product_id}/status", func(w http.ResponseWriter, r *http.Request) {
		var input internalStatusRequest
		if !decode(w, r, &input) {
			return
		}
		result, err := application.SellerProductHandler.ModerateProduct(r.Context(), dto.ProductModerationRequestDTO{
			Actor: internalActor(r, input.ActorUserID, input.SellerID), ProductID: r.PathValue("product_id"), ExpectedStatus: input.ExpectedStatus, Status: input.Status,
		})
		respond(w, result, err)
	})
	mux.HandleFunc("POST /internal/v1/inventory/reservations", func(w http.ResponseWriter, r *http.Request) {
		var input dto.InventoryReservationRequestDTO
		if !decode(w, r, &input) {
			return
		}
		result, err := application.InventoryHandler.ReserveInventory(r.Context(), input)
		respondStatus(w, http.StatusCreated, result, err)
	})
	mux.HandleFunc("POST /internal/v1/inventory/reservations/{reservation_id}/release", func(w http.ResponseWriter, r *http.Request) {
		var input dto.InventoryReservationActionRequestDTO
		if r.ContentLength > 0 && !decode(w, r, &input) {
			return
		}
		input.ReservationID = r.PathValue("reservation_id")
		result, err := application.InventoryHandler.ReleaseInventory(r.Context(), input)
		respond(w, result, err)
	})
	mux.HandleFunc("POST /internal/v1/inventory/reservations/{reservation_id}/commit", func(w http.ResponseWriter, r *http.Request) {
		input := dto.InventoryReservationActionRequestDTO{ReservationID: r.PathValue("reservation_id")}
		result, err := application.InventoryHandler.CommitInventory(r.Context(), input)
		respond(w, result, err)
	})
	return internalServiceAuth(mux, application.Config.InternalServiceToken)
}

type internalStatusRequest struct {
	ExpectedStatus string `json:"expected_status,omitempty"`
	Status         string `json:"status"`
	ActorUserID    string `json:"actor_user_id,omitempty"`
	SellerID       string `json:"seller_id,omitempty"`
	ReviewID       string `json:"review_id,omitempty"`
	Reason         string `json:"reason,omitempty"`
	RequestID      string `json:"request_id,omitempty"`
	Force          bool   `json:"force,omitempty"`
}

func internalActor(r *http.Request, userID, sellerID string) dto.ActorContextDTO {
	if strings.TrimSpace(userID) == "" {
		userID = r.Header.Get("X-User-ID")
	}
	if strings.TrimSpace(sellerID) == "" {
		sellerID = r.Header.Get("X-Seller-ID")
	}
	return dto.ActorContextDTO{UserID: strings.TrimSpace(userID), SellerID: strings.TrimSpace(sellerID), Roles: []string{"superadmin"}}
}

func actor(r *http.Request) dto.ActorContextDTO {
	return dto.ActorContextDTO{
		UserID: strings.TrimSpace(r.Header.Get("X-User-ID")), SellerID: strings.TrimSpace(r.Header.Get("X-Seller-ID")),
		Roles: csvHeader(r, "X-Roles"), Permissions: csvHeader(r, "X-Permissions"),
	}
}

func csvHeader(r *http.Request, name string) []string {
	var values []string
	for _, value := range strings.Split(r.Header.Get(name), ",") {
		if value = strings.TrimSpace(value); value != "" {
			values = append(values, value)
		}
	}
	return values
}

func decode(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]string{"code": "INVALID_REQUEST", "message": err.Error()}})
		return false
	}
	return true
}

func respond(w http.ResponseWriter, value any, err error) {
	respondStatus(w, http.StatusOK, value, err)
}

func respondStatus(w http.ResponseWriter, status int, value any, err error) {
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, status, value)
}

func writeServiceError(w http.ResponseWriter, err error) {
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "internal product service error"
	var serviceErr *usecase.ServiceError
	if errors.As(err, &serviceErr) {
		code, message = serviceErr.Code, serviceErr.Message
		switch serviceErr.Kind {
		case usecase.ErrorKindInvalidArgument:
			status = http.StatusBadRequest
		case usecase.ErrorKindUnauthenticated:
			status = http.StatusUnauthorized
		case usecase.ErrorKindPermissionDenied:
			status = http.StatusForbidden
		case usecase.ErrorKindNotFound:
			status = http.StatusNotFound
		case usecase.ErrorKindAlreadyExists, usecase.ErrorKindConflict, usecase.ErrorKindFailedPrecondition:
			status = http.StatusConflict
		case usecase.ErrorKindUnavailable:
			status = http.StatusServiceUnavailable
		}
	}
	writeJSON(w, status, map[string]any{"error": map[string]any{"code": code, "message": message}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func intQuery(r *http.Request, name string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get(name)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func authorizedInternalService(r *http.Request, expected string) bool {
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return false
	}
	provided := strings.TrimSpace(r.Header.Get("X-Service-Token"))
	return len(provided) == len(expected) && subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func internalServiceAuth(next http.Handler, expected string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/internal/v1/") && !authorizedInternalService(r, expected) {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": map[string]string{
				"code": "UNAUTHENTICATED", "message": "valid service credentials are required",
			}})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func internalGRPCAuth(expected string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !internalGRPCMethod(info.FullMethod) {
			return handler(ctx, req)
		}
		md, _ := metadata.FromIncomingContext(ctx)
		provided := ""
		if values := md.Get("x-service-token"); len(values) > 0 {
			provided = values[0]
		}
		expected = strings.TrimSpace(expected)
		provided = strings.TrimSpace(provided)
		if expected == "" || len(provided) != len(expected) || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			return nil, status.Error(codes.Unauthenticated, "valid service credentials are required")
		}
		return handler(ctx, req)
	}
}

func internalGRPCMethod(fullMethod string) bool {
	for _, method := range []string{"/BatchGetProducts", "/ReserveInventory", "/ReleaseInventory", "/CommitInventory"} {
		if strings.HasSuffix(fullMethod, method) {
			return true
		}
	}
	return false
}
