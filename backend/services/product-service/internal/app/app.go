package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"product-service/internal/client"
	"product-service/internal/config"
	"product-service/internal/repository"
	"product-service/internal/transport/catalogmodel"
	"product-service/internal/transport/catalogschema"
	"product-service/internal/transport/productread"
	"product-service/internal/transport/sellerproduct"
	"product-service/internal/usecase"
)

type Dependencies struct {
	CategoryReader          repository.CategoryReader
	SKUChecker              repository.SKUUniquenessChecker
	ProductRepository       repository.ProductRepository
	ProductReadRepository   repository.ProductReadRepository
	CategoryReadRepository  repository.CategoryReadRepository
	CMSClient               client.CMSClient
	ProductIDGenerator      usecase.IDGenerator
	Clock                   usecase.Clock
	CollectionSchemaManager repository.CollectionSchemaManager
	Logger                  *slog.Logger
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
	CollectionSchemaUseCase usecase.CollectionSchemaUseCase
	CollectionSchemaHandler *catalogschema.Handler
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
		handler, err := sellerproduct.NewHandler(productUseCase, logger)
		if err != nil {
			return nil, fmt.Errorf("wire seller product handler: %w", err)
		}
		sellerUseCase = productUseCase
		sellerHandler = handler
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
		CollectionSchemaUseCase: collectionUseCase,
		CollectionSchemaHandler: collectionHandler,
	}, nil
}

func newLogger(cfg config.Config) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel})
	return slog.New(handler).With("service", cfg.ServiceName, "environment", cfg.Environment)
}
