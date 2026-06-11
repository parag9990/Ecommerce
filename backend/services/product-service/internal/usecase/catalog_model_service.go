package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"product-service/internal/domain"
	"product-service/internal/repository"
)

type CatalogModelUseCase interface {
	ValidateProduct(ctx context.Context, request ValidateProductRequest) (domain.ValidationReport, error)
	CheckPublishReadiness(ctx context.Context, request ValidateProductRequest) (domain.ValidationReport, error)
}

type ValidateProductRequest struct {
	Product            domain.Product
	Category           *domain.Category
	ExcludeProductID   string
	CheckSKUUniqueness bool
}

type CatalogModelService struct {
	categoryReader repository.CategoryReader
	skuChecker     repository.SKUUniquenessChecker
	logger         *slog.Logger
	options        domain.ValidationOptions
}

func NewCatalogModelService(
	categoryReader repository.CategoryReader,
	skuChecker repository.SKUUniquenessChecker,
	logger *slog.Logger,
	options domain.ValidationOptions,
) *CatalogModelService {
	if logger == nil {
		logger = slog.Default()
	}
	return &CatalogModelService{
		categoryReader: categoryReader,
		skuChecker:     skuChecker,
		logger:         logger,
		options:        options,
	}
}

func (s *CatalogModelService) ValidateProduct(ctx context.Context, request ValidateProductRequest) (domain.ValidationReport, error) {
	if err := ctx.Err(); err != nil {
		return domain.ValidationReport{}, err
	}

	category, report, err := s.resolveCategory(ctx, request)
	if err != nil {
		return domain.ValidationReport{}, err
	}
	report.Merge(request.Product.Validate(s.options, category))

	if request.CheckSKUUniqueness {
		skuReport, err := s.validateSKUUniqueness(ctx, request.Product, request.ExcludeProductID)
		if err != nil {
			return domain.ValidationReport{}, err
		}
		report.Merge(skuReport)
	}

	s.logReport("catalog product validation completed", request.Product.ID, report)
	return report, nil
}

func (s *CatalogModelService) CheckPublishReadiness(ctx context.Context, request ValidateProductRequest) (domain.ValidationReport, error) {
	if err := ctx.Err(); err != nil {
		return domain.ValidationReport{}, err
	}

	category, report, err := s.resolveCategory(ctx, request)
	if err != nil {
		return domain.ValidationReport{}, err
	}
	report.Merge(request.Product.ValidateForPublish(s.options, category))

	skuReport, err := s.validateSKUUniqueness(ctx, request.Product, request.ExcludeProductID)
	if err != nil {
		return domain.ValidationReport{}, err
	}
	report.Merge(skuReport)

	s.logReport("catalog publish readiness checked", request.Product.ID, report)
	return report, nil
}

func (s *CatalogModelService) resolveCategory(ctx context.Context, request ValidateProductRequest) (*domain.Category, domain.ValidationReport, error) {
	var report domain.ValidationReport
	if request.Category != nil {
		return request.Category, report, nil
	}
	if request.Product.CategoryID == "" {
		return nil, report, nil
	}
	if s.categoryReader == nil {
		report.AddWarning(domain.CodeCategoryNotVerified, "category_id", "category existence and schema were not verified")
		return nil, report, nil
	}

	category, err := s.categoryReader.GetCategoryByID(ctx, request.Product.CategoryID)
	if errors.Is(err, repository.ErrNotFound) {
		report.AddError(domain.CodeCategoryNotFound, "category_id", "category does not exist")
		return nil, report, nil
	}
	if err != nil {
		return nil, report, fmt.Errorf("get category %q: %w", request.Product.CategoryID, err)
	}
	return category, report, nil
}

func (s *CatalogModelService) validateSKUUniqueness(ctx context.Context, product domain.Product, excludeProductID string) (domain.ValidationReport, error) {
	var report domain.ValidationReport
	if s.skuChecker == nil {
		report.AddWarning(domain.CodeSKUUniquenessNotVerified, "variants", "platform-wide SKU uniqueness was not verified")
		return report, nil
	}

	checked := make(map[string]struct{}, len(product.Variants))
	for index, variant := range product.Variants {
		if variant.SKU == "" {
			continue
		}
		if _, exists := checked[variant.SKU]; exists {
			continue
		}
		checked[variant.SKU] = struct{}{}
		if err := ctx.Err(); err != nil {
			return domain.ValidationReport{}, err
		}
		unique, err := s.skuChecker.IsSKUUnique(ctx, variant.SKU, excludeProductID)
		if err != nil {
			return domain.ValidationReport{}, fmt.Errorf("check SKU uniqueness for %q: %w", variant.SKU, err)
		}
		if !unique {
			report.AddError(domain.CodeDuplicateSKU, fmt.Sprintf("variants[%d].sku", index), "variant SKU must be unique platform-wide")
		}
	}
	return report, nil
}

func (s *CatalogModelService) logReport(message, productID string, report domain.ValidationReport) {
	if report.HasErrors() {
		s.logger.Warn(message, "product_id", productID, "issue_count", len(report.Issues), "valid", false)
		return
	}
	s.logger.Info(message, "product_id", productID, "issue_count", len(report.Issues), "valid", true)
}
