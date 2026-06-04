package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"product-service/internal/client"
	"product-service/internal/config"
	eventing "product-service/internal/events"
	"product-service/internal/repository"
	"product-service/internal/transport/catalogmodel"
	"product-service/internal/transport/catalogschema"
	"product-service/internal/transport/inventory"
	"product-service/internal/transport/productread"
	"product-service/internal/transport/sellerproduct"
	"product-service/internal/usecase"
)

type Dependencies struct {
	CategoryReader                 repository.CategoryReader
	SKUChecker                     repository.SKUUniquenessChecker
	ProductRepository              repository.ProductRepository
	ProductReadRepository          repository.ProductReadRepository
	CategoryReadRepository         repository.CategoryReadRepository
	CMSClient                      client.CMSClient
	ProductIDGenerator             usecase.IDGenerator
	InventoryIDGenerator           usecase.InventoryIDGenerator
	Clock                          usecase.Clock
	CollectionSchemaManager        repository.CollectionSchemaManager
	InventoryStockRepository       repository.InventoryStockRepository
	InventoryReservationRepository repository.InventoryReservationRepository
	InventorySnapshotRepository    repository.InventorySnapshotRepository
	ProductEventOutboxRepository   repository.ProductEventOutboxRepository
	ProductEventIDGenerator        usecase.ProductEventIDGenerator
	ProductEventPublisher          eventing.ProductEventPublisher
	Logger                         *slog.Logger
}

type App struct {
	Config                  config.Config
	Logger                  *slog.Logger
	CatalogModelUseCase     usecase.CatalogModelUseCase
	CatalogModelHandler     *catalogmodel.Handler
	ProductReadUseCase      usecase.ProductReadUseCase
	ProductReadHandler      *productread.Handler
	SellerProductUseCase    usecase.SellerProductUseCase
	SellerProductHandler    *sellerproduct.Handler
	InventoryUseCase        usecase.InventoryUseCase
	InventoryHandler        *inventory.Handler
	CollectionSchemaUseCase usecase.CollectionSchemaUseCase
	CollectionSchemaHandler *catalogschema.Handler
	ProductEventRecorder    *usecase.ProductEventService
	ProductEventRelay       *eventing.OutboxRelay
}

func New(ctx context.Context, cfg config.Config, deps Dependencies) (*App, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	logger := deps.Logger
	if logger == nil {
		logger = newLogger(cfg)
	}
	var err error

	categoryReader := deps.CategoryReader
	skuChecker := deps.SKUChecker
	if deps.ProductRepository != nil {
		if categoryReader == nil {
			categoryReader = deps.ProductRepository
		}
		if skuChecker == nil {
			skuChecker = deps.ProductRepository
		}
	}

	productEventOutboxRepository := deps.ProductEventOutboxRepository
	if productEventOutboxRepository == nil && deps.ProductRepository != nil {
		if repo, ok := deps.ProductRepository.(repository.ProductEventOutboxRepository); ok {
			productEventOutboxRepository = repo
		}
	}

	var productEventRecorder *usecase.ProductEventService
	if cfg.Events.Enabled {
		if productEventOutboxRepository == nil {
			logger.Warn("product events enabled but no outbox repository is available")
		} else {
			productEventRecorder, err = usecase.NewProductEventService(
				productEventOutboxRepository,
				deps.ProductEventIDGenerator,
				deps.Clock,
				logger,
				cfg.ProductEventOptions(),
			)
			if err != nil {
				return nil, fmt.Errorf("wire product event recorder: %w", err)
			}
		}
	}

	var productEventRelay *eventing.OutboxRelay
	if cfg.Events.Enabled && cfg.Events.OutboxWorkerEnabled && productEventOutboxRepository != nil {
		publisher := deps.ProductEventPublisher
		if publisher == nil {
			switch strings.ToLower(strings.TrimSpace(cfg.Events.Broker)) {
			case "rabbitmq":
				publisher, err = eventing.NewRabbitMQPublisher(cfg.Events.RabbitMQURL, logger)
				if err != nil {
					return nil, fmt.Errorf("wire rabbitmq product event publisher: %w", err)
				}
			case "kafka":
				return nil, fmt.Errorf("PRODUCT_EVENT_BROKER=kafka requires a ProductEventPublisher dependency")
			default:
				return nil, fmt.Errorf("unsupported PRODUCT_EVENT_BROKER %q", cfg.Events.Broker)
			}
		}
		relay, err := eventing.NewOutboxRelay(
			productEventOutboxRepository,
			publisher,
			deps.Clock,
			logger,
			cfg.OutboxRelayOptions(),
		)
		if err != nil {
			return nil, fmt.Errorf("wire product event outbox relay: %w", err)
		}
		productEventRelay = relay
	}

	catalogUseCase := usecase.NewCatalogModelService(
		categoryReader,
		skuChecker,
		logger,
		cfg.ValidationOptions(),
	)
	catalogHandler, err := catalogmodel.NewHandler(catalogUseCase, logger)
	if err != nil {
		return nil, fmt.Errorf("wire catalog model handler: %w", err)
	}

	productReadRepository := deps.ProductReadRepository
	if productReadRepository == nil {
		if readRepo, ok := deps.ProductRepository.(repository.ProductReadRepository); ok {
			productReadRepository = readRepo
		}
	}
	categoryReadRepository := deps.CategoryReadRepository
	if categoryReadRepository == nil {
		if readRepo, ok := deps.ProductRepository.(repository.CategoryReadRepository); ok {
			categoryReadRepository = readRepo
		} else if readRepo, ok := productReadRepository.(repository.CategoryReadRepository); ok {
			categoryReadRepository = readRepo
		}
	}

	var productReadUseCase usecase.ProductReadUseCase
	var productReadHandler *productread.Handler
	if productReadRepository != nil && categoryReadRepository != nil {
		readUseCase, err := usecase.NewProductReadService(
			productReadRepository,
			categoryReadRepository,
			logger,
			usecase.ProductReadServiceOptions{
				DefaultPageSize: cfg.Read.DefaultPageSize,
				MaxPageSize:     cfg.Read.MaxPageSize,
				MaxBatchSize:    cfg.Read.MaxBatchSize,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("wire product read usecase: %w", err)
		}
		readHandler, err := productread.NewHandler(readUseCase, logger)
		if err != nil {
			return nil, fmt.Errorf("wire product read handler: %w", err)
		}
		productReadUseCase = readUseCase
		productReadHandler = readHandler
	}

	var sellerUseCase usecase.SellerProductUseCase
	var sellerHandler *sellerproduct.Handler
	if deps.ProductRepository != nil {
		cmsClient := deps.CMSClient
		if cmsClient == nil {
			cmsClient, err = client.NewStaticCMSPolicyClient(client.StaticCMSPolicy{
				CatalogManagementAllowed: cfg.CMS.CatalogManagementAllowed,
				ModerationDecision:       client.NormalizeModerationDecision(cfg.CMS.ModerationDecision),
			})
			if err != nil {
				return nil, fmt.Errorf("wire cms policy client: %w", err)
			}
		}
		productValidator := usecase.NewCatalogModelService(
			deps.ProductRepository,
			deps.ProductRepository,
			logger,
			cfg.ValidationOptions(),
		)
		productUseCase, err := usecase.NewSellerProductService(
			deps.ProductRepository,
			productValidator,
			cmsClient,
			deps.ProductIDGenerator,
			deps.Clock,
			logger,
			usecase.SellerProductServiceOptions{
				AllowDraftWritesWhenCMSUnavailable: cfg.CMS.AllowDraftWritesWhenUnavailable,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("wire seller product usecase: %w", err)
		}
		if productEventRecorder != nil {
			productUseCase.EnableProductEventRecording(productEventRecorder)
		}
		handler, err := sellerproduct.NewHandler(productUseCase, logger)
		if err != nil {
			return nil, fmt.Errorf("wire seller product handler: %w", err)
		}
		sellerUseCase = productUseCase
		sellerHandler = handler
	}

	inventoryStockRepository := deps.InventoryStockRepository
	inventoryReservationRepository := deps.InventoryReservationRepository
	inventorySnapshotRepository := deps.InventorySnapshotRepository
	if deps.ProductRepository != nil {
		if inventoryStockRepository == nil {
			if repo, ok := deps.ProductRepository.(repository.InventoryStockRepository); ok {
				inventoryStockRepository = repo
			}
		}
		if inventoryReservationRepository == nil {
			if repo, ok := deps.ProductRepository.(repository.InventoryReservationRepository); ok {
				inventoryReservationRepository = repo
			}
		}
		if inventorySnapshotRepository == nil {
			if repo, ok := deps.ProductRepository.(repository.InventorySnapshotRepository); ok {
				inventorySnapshotRepository = repo
			}
		}
	}

	var inventoryUseCase usecase.InventoryUseCase
	var inventoryHandler *inventory.Handler
	inventoryDependencyCount := 0
	if inventoryStockRepository != nil {
		inventoryDependencyCount++
	}
	if inventoryReservationRepository != nil {
		inventoryDependencyCount++
	}
	if inventorySnapshotRepository != nil {
		inventoryDependencyCount++
	}
	if inventoryDependencyCount == 3 {
		service, err := usecase.NewInventoryService(
			inventoryStockRepository,
			inventoryReservationRepository,
			inventorySnapshotRepository,
			deps.InventoryIDGenerator,
			deps.Clock,
			logger,
			cfg.InventoryOptions(),
		)
		if err != nil {
			return nil, fmt.Errorf("wire inventory usecase: %w", err)
		}
		if productEventRecorder != nil {
			var productReader repository.ProductFinder
			if deps.ProductRepository != nil {
				productReader = deps.ProductRepository
			} else if reader, ok := inventoryStockRepository.(repository.ProductFinder); ok {
				productReader = reader
			}
			service.EnableProductEventRecording(productEventRecorder, productReader)
		}
		handler, err := inventory.NewHandler(service, logger)
		if err != nil {
			return nil, fmt.Errorf("wire inventory handler: %w", err)
		}
		inventoryUseCase = service
		inventoryHandler = handler
	} else if inventoryDependencyCount > 0 {
		return nil, fmt.Errorf("inventory stock, reservation, and snapshot repositories are all required to wire inventory")
	}

	var collectionUseCase usecase.CollectionSchemaUseCase
	var collectionHandler *catalogschema.Handler
	if deps.CollectionSchemaManager != nil {
		schemaService := usecase.NewCollectionSchemaService(deps.CollectionSchemaManager, logger)
		schemaHandler, err := catalogschema.NewHandler(schemaService, logger)
		if err != nil {
			return nil, fmt.Errorf("wire collection schema handler: %w", err)
		}
		if cfg.Mongo.AutoCreateCollections {
			if _, err := schemaService.EnsureProductCollections(ctx); err != nil {
				return nil, fmt.Errorf("ensure product mongo collections: %w", err)
			}
		}
		collectionUseCase = schemaService
		collectionHandler = schemaHandler
	} else if cfg.Mongo.AutoCreateCollections {
		return nil, fmt.Errorf("collection schema manager is required when PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS is true")
	}

	return &App{
		Config:                  cfg,
		Logger:                  logger,
		CatalogModelUseCase:     catalogUseCase,
		CatalogModelHandler:     catalogHandler,
		ProductReadUseCase:      productReadUseCase,
		ProductReadHandler:      productReadHandler,
		SellerProductUseCase:    sellerUseCase,
		SellerProductHandler:    sellerHandler,
		InventoryUseCase:        inventoryUseCase,
		InventoryHandler:        inventoryHandler,
		CollectionSchemaUseCase: collectionUseCase,
		CollectionSchemaHandler: collectionHandler,
		ProductEventRecorder:    productEventRecorder,
		ProductEventRelay:       productEventRelay,
	}, nil
}

func (a *App) StartBackgroundWorkers(ctx context.Context) {
	if a == nil || a.ProductEventRelay == nil {
		return
	}
	go a.ProductEventRelay.Run(ctx)
}

func newLogger(cfg config.Config) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel})
	return slog.New(handler).With("service", cfg.ServiceName, "environment", cfg.Environment)
}
