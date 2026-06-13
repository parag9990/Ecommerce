package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/indexer"
)

const (
	defaultReindexMinIndexedRatio = 0.95
	defaultReindexMaxBatchSize    = 5000
)

type ReindexCatalogOptions struct {
	ProductsAlias          string
	DefaultMode            string
	DefaultBatchSize       int
	MaxBatchSize           int
	CollectionPrefix       string
	LockTTL                time.Duration
	JobTimeout             time.Duration
	ProductPageTimeout     time.Duration
	ImportTimeout          time.Duration
	OldCollectionRetention time.Duration
	MinIndexedRatio        float64
	Metrics                ReindexMetricsRecorder
	Clock                  func() time.Time
}

type ReindexCatalogUsecase struct {
	schemaRepo SchemaRepository
	products   ProductCatalogReader
	index      ReindexRepository
	lock       ReindexLock
	options    ReindexCatalogOptions
	logger     *slog.Logger
}

func NewReindexCatalogUsecase(schemaRepo SchemaRepository, products ProductCatalogReader, index ReindexRepository, lock ReindexLock, options ReindexCatalogOptions, logger *slog.Logger) (*ReindexCatalogUsecase, error) {
	if schemaRepo == nil {
		return nil, errors.New("schema repository is required")
	}
	if products == nil {
		return nil, errors.New("product catalog reader is required")
	}
	if index == nil {
		return nil, errors.New("reindex repository is required")
	}
	if lock == nil {
		return nil, errors.New("reindex lock is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	options = normalizeReindexOptions(options)
	return &ReindexCatalogUsecase{
		schemaRepo: schemaRepo,
		products:   products,
		index:      index,
		lock:       lock,
		options:    options,
		logger:     logger,
	}, nil
}

func (u *ReindexCatalogUsecase) Execute(ctx context.Context, req domain.ReindexRequest) (domain.ReindexResult, error) {
	startedAt := u.now().UTC()
	req = domain.NormalizeReindexRequest(req, u.options.DefaultMode, u.options.DefaultBatchSize, startedAt)
	if err := req.Validate(u.options.MaxBatchSize); err != nil {
		u.record(ctx, req.Mode, "failed", "validate_request", domain.ReindexResult{}, err, startedAt)
		return domain.ReindexResult{}, err
	}

	if u.options.JobTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, u.options.JobTimeout)
		defer cancel()
	}

	locked, err := u.lock.Acquire(ctx, req.JobID, u.options.LockTTL)
	if err != nil {
		u.record(ctx, req.Mode, "failed", "acquire_lock", domain.ReindexResult{}, err, startedAt)
		return domain.ReindexResult{}, err
	}
	if !locked {
		err := domain.ErrReindexAlreadyRunning
		u.record(ctx, req.Mode, "failed", "acquire_lock", domain.ReindexResult{}, err, startedAt)
		return domain.ReindexResult{}, err
	}
	stopHeartbeat := u.startLockHeartbeat(ctx, req.JobID)
	defer stopHeartbeat()
	defer u.releaseLock(req.JobID)

	result := domain.ReindexResult{
		JobID:  req.JobID,
		Mode:   req.Mode,
		DryRun: req.DryRun,
	}

	targetCollection, previousCollection, err := u.prepareTarget(ctx, req, startedAt)
	if err != nil {
		u.record(ctx, req.Mode, "failed", "prepare_target", result, err, startedAt)
		return domain.ReindexResult{}, err
	}
	result.TargetCollection = targetCollection
	result.PreviousCollection = previousCollection

	result, err = u.indexAllPages(ctx, req, result)
	if err != nil {
		u.record(ctx, req.Mode, "failed", "index_pages", result, err, startedAt)
		return domain.ReindexResult{}, err
	}

	if err := u.validateResult(ctx, req, result); err != nil {
		u.record(ctx, req.Mode, "failed", "validate_target", result, err, startedAt)
		return domain.ReindexResult{}, err
	}

	if req.Mode == domain.ReindexModeAlias && !req.DryRun {
		if err := u.index.SwapAlias(ctx, u.options.ProductsAlias, targetCollection); err != nil {
			u.record(ctx, req.Mode, "failed", "swap_alias", result, err, startedAt)
			return domain.ReindexResult{}, err
		}
		result.AliasSwapped = true
		u.logger.InfoContext(ctx, "search.reindex.alias_swapped",
			slog.String("job_id", result.JobID),
			slog.String("alias", u.options.ProductsAlias),
			slog.String("target_collection", targetCollection),
			slog.String("previous_collection", previousCollection),
		)
		u.cleanupOldCollections(ctx, result, startedAt)
	}

	result.Duration = u.now().UTC().Sub(startedAt)
	u.record(ctx, req.Mode, "succeeded", "completed", result, nil, startedAt)
	u.logger.InfoContext(ctx, "search.reindex.completed",
		slog.String("job_id", result.JobID),
		slog.String("mode", result.Mode),
		slog.String("target_collection", result.TargetCollection),
		slog.String("previous_collection", result.PreviousCollection),
		slog.Int("products_read", result.ProductsRead),
		slog.Int("products_indexed", result.ProductsIndexed),
		slog.Int("products_skipped", result.ProductsSkipped),
		slog.Int("failed_documents", result.FailedDocuments),
		slog.Int64("duration_ms", result.Duration.Milliseconds()),
		slog.Bool("alias_swapped", result.AliasSwapped),
		slog.Bool("dry_run", result.DryRun),
	)
	return result, nil
}

func (u *ReindexCatalogUsecase) prepareTarget(ctx context.Context, req domain.ReindexRequest, startedAt time.Time) (string, string, error) {
	targetCollection := strings.TrimSpace(req.TargetCollection)
	if targetCollection == "" {
		if req.Mode == domain.ReindexModeAlias {
			targetCollection = VersionedCollectionName(u.options.CollectionPrefix, startedAt)
		} else {
			targetCollection = u.options.ProductsAlias
		}
	}

	previousCollection := ""
	if req.Mode == domain.ReindexModeAlias {
		resolved, err := u.index.ResolveAlias(ctx, u.options.ProductsAlias)
		if err != nil {
			return "", "", err
		}
		previousCollection = resolved
	}
	if req.DryRun {
		return targetCollection, previousCollection, nil
	}

	collection, err := u.schemaRepo.GetProductsCollectionSchema(ctx)
	if err != nil {
		return "", "", errors.Join(domain.ErrSchemaUnavailable, err)
	}
	collection.Name = targetCollection
	if err := u.index.EnsureCollection(ctx, collection); err != nil {
		return "", "", err
	}
	return targetCollection, previousCollection, nil
}

func (u *ReindexCatalogUsecase) indexAllPages(ctx context.Context, req domain.ReindexRequest, result domain.ReindexResult) (domain.ReindexResult, error) {
	cursor := ""
	seenCursors := map[string]struct{}{}

	for {
		pageCtx := ctx
		cancelPage := func() {}
		if u.options.ProductPageTimeout > 0 {
			pageCtx, cancelPage = context.WithTimeout(ctx, u.options.ProductPageTimeout)
		}
		page, err := u.products.ListSearchableProducts(pageCtx, domain.ProductExportRequest{
			Cursor: cursor,
			Limit:  req.BatchSize,
		})
		cancelPage()
		if err != nil {
			return result, err
		}
		if page.Total > result.SourceTotal {
			result.SourceTotal = page.Total
		}
		if page.HasMore && len(page.Items) == 0 {
			return result, fmt.Errorf("%w: product export returned an empty page with has_more=true", domain.ErrProductCatalogExportUnavailable)
		}

		docs := make([]domain.ProductDocument, 0, len(page.Items))
		now := u.now().UTC()
		for _, item := range page.Items {
			result.ProductsRead++
			doc, ok, err := mapReindexExportPayload(item, now)
			if err != nil {
				result.ProductsSkipped++
				result.FailedDocuments++
				u.logger.WarnContext(ctx, "search.reindex.product_skipped",
					slog.String("job_id", result.JobID),
					slog.String("product_id", item.ProductID),
					slog.String("reason", err.Error()),
				)
				continue
			}
			if !ok {
				result.ProductsSkipped++
				continue
			}
			docs = append(docs, doc)
		}

		if len(docs) > 0 {
			if !result.DryRun {
				importCtx := ctx
				cancelImport := func() {}
				if u.options.ImportTimeout > 0 {
					importCtx, cancelImport = context.WithTimeout(ctx, u.options.ImportTimeout)
				}
				err = u.index.ImportProducts(importCtx, result.TargetCollection, docs)
				cancelImport()
				if err != nil {
					return result, err
				}
			}
			result.ProductsIndexed += len(docs)
		}

		u.logProgress(ctx, result, cursor)
		if !page.HasMore {
			break
		}
		nextCursor := strings.TrimSpace(page.NextCursor)
		if nextCursor == "" {
			return result, fmt.Errorf("%w: next_cursor is required when has_more=true", domain.ErrProductCatalogExportUnavailable)
		}
		if _, exists := seenCursors[nextCursor]; exists || nextCursor == cursor {
			return result, fmt.Errorf("%w: product export cursor did not advance", domain.ErrProductCatalogExportUnavailable)
		}
		seenCursors[nextCursor] = struct{}{}
		cursor = nextCursor
	}
	return result, nil
}

func mapReindexExportPayload(payload domain.ProductIndexPayload, now time.Time) (domain.ProductDocument, bool, error) {
	normalized := payload.Normalized()
	if !normalized.IsSearchable() {
		return domain.ProductDocument{}, false, nil
	}
	if err := normalized.ValidateForUpsert(); err != nil {
		return domain.ProductDocument{}, false, err
	}
	doc := indexer.MapProductToDocument(normalized, now)
	if err := doc.Validate(); err != nil {
		return domain.ProductDocument{}, false, err
	}
	return doc, true, nil
}

func (u *ReindexCatalogUsecase) validateResult(ctx context.Context, req domain.ReindexRequest, result domain.ReindexResult) error {
	if result.ProductsIndexed <= 0 {
		return fmt.Errorf("%w: target collection would be empty", domain.ErrReindexValidationFailed)
	}
	if result.SourceTotal > 0 {
		minimum := int(math.Ceil(float64(result.SourceTotal) * u.options.MinIndexedRatio))
		if result.ProductsIndexed < minimum {
			return fmt.Errorf("%w: indexed %d products below required minimum %d from source total %d", domain.ErrReindexValidationFailed, result.ProductsIndexed, minimum, result.SourceTotal)
		}
	}
	if req.DryRun {
		return nil
	}

	count, err := u.index.CountDocuments(ctx, result.TargetCollection)
	if err != nil {
		return err
	}
	if count <= 0 {
		return fmt.Errorf("%w: target collection is empty", domain.ErrReindexValidationFailed)
	}
	if count < result.ProductsIndexed {
		return fmt.Errorf("%w: target count %d is below indexed count %d", domain.ErrReindexValidationFailed, count, result.ProductsIndexed)
	}
	if result.SourceTotal > 0 {
		minimum := int(math.Ceil(float64(result.SourceTotal) * u.options.MinIndexedRatio))
		if count < minimum {
			return fmt.Errorf("%w: target count %d is below required minimum %d from source total %d", domain.ErrReindexValidationFailed, count, minimum, result.SourceTotal)
		}
	}
	return u.index.SmokeSearch(ctx, result.TargetCollection)
}

func (u *ReindexCatalogUsecase) cleanupOldCollections(ctx context.Context, result domain.ReindexResult, now time.Time) {
	if u.options.OldCollectionRetention <= 0 {
		return
	}
	deleted, err := u.index.CleanupOldCollections(ctx, u.options.CollectionPrefix, result.TargetCollection, result.PreviousCollection, u.options.OldCollectionRetention, now)
	if err != nil {
		u.logger.WarnContext(ctx, "search.reindex.cleanup_failed",
			slog.String("job_id", result.JobID),
			slog.String("error", err.Error()),
		)
		return
	}
	if len(deleted) == 0 {
		return
	}
	u.logger.InfoContext(ctx, "search.reindex.cleanup_completed",
		slog.String("job_id", result.JobID),
		slog.Int("deleted_collections", len(deleted)),
		slog.String("collections", strings.Join(deleted, ",")),
	)
}

func (u *ReindexCatalogUsecase) logProgress(ctx context.Context, result domain.ReindexResult, cursor string) {
	u.logger.InfoContext(ctx, "search.reindex.progress",
		slog.String("job_id", result.JobID),
		slog.String("mode", result.Mode),
		slog.String("target_collection", result.TargetCollection),
		slog.String("source_cursor", cursor),
		slog.Int("products_read", result.ProductsRead),
		slog.Int("products_indexed", result.ProductsIndexed),
		slog.Int("products_skipped", result.ProductsSkipped),
		slog.Int("failed_documents", result.FailedDocuments),
		slog.Bool("dry_run", result.DryRun),
	)
}

func (u *ReindexCatalogUsecase) startLockHeartbeat(ctx context.Context, jobID string) func() {
	if u.options.LockTTL <= 0 {
		return func() {}
	}
	interval := u.options.LockTTL / 3
	if interval < 100*time.Millisecond {
		interval = 100 * time.Millisecond
	}
	heartbeatCtx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-heartbeatCtx.Done():
				return
			case <-ticker.C:
				ok, err := u.lock.Refresh(heartbeatCtx, jobID, u.options.LockTTL)
				if err != nil {
					u.logger.WarnContext(heartbeatCtx, "search.reindex.lock_refresh_failed",
						slog.String("job_id", jobID),
						slog.String("error", err.Error()),
					)
					continue
				}
				if !ok {
					u.logger.WarnContext(heartbeatCtx, "search.reindex.lock_lost",
						slog.String("job_id", jobID),
					)
					return
				}
			}
		}
	}()
	return cancel
}

func (u *ReindexCatalogUsecase) releaseLock(jobID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := u.lock.Release(ctx, jobID); err != nil {
		u.logger.Warn("search.reindex.lock_release_failed",
			slog.String("job_id", jobID),
			slog.String("error", err.Error()),
		)
	}
}

func (u *ReindexCatalogUsecase) record(ctx context.Context, mode string, status string, stage string, result domain.ReindexResult, err error, startedAt time.Time) {
	errorCode := ""
	if err != nil {
		errorCode = reindexErrorCode(err)
	}
	u.options.Metrics.RecordReindexRun(ctx, ReindexMetrics{
		Mode:            mode,
		Status:          status,
		Stage:           stage,
		ProductsRead:    result.ProductsRead,
		ProductsIndexed: result.ProductsIndexed,
		ProductsSkipped: result.ProductsSkipped,
		DurationMS:      int(u.now().UTC().Sub(startedAt).Milliseconds()),
		ErrorCode:       errorCode,
	})
}

func (u *ReindexCatalogUsecase) now() time.Time {
	if u.options.Clock != nil {
		return u.options.Clock()
	}
	return time.Now()
}

func VersionedCollectionName(prefix string, now time.Time) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = domain.ProductsCollectionName
	}
	if now.IsZero() {
		now = time.Now()
	}
	return fmt.Sprintf("%s_%s", prefix, now.UTC().Format("20060102_150405"))
}

func normalizeReindexOptions(options ReindexCatalogOptions) ReindexCatalogOptions {
	options.ProductsAlias = strings.TrimSpace(options.ProductsAlias)
	if options.ProductsAlias == "" {
		options.ProductsAlias = domain.ProductsCollectionName
	}
	options.DefaultMode = strings.ToLower(strings.TrimSpace(options.DefaultMode))
	if options.DefaultMode == "" {
		options.DefaultMode = domain.ReindexModeAlias
	}
	if options.DefaultBatchSize <= 0 {
		options.DefaultBatchSize = 500
	}
	if options.MaxBatchSize <= 0 {
		options.MaxBatchSize = defaultReindexMaxBatchSize
	}
	if options.CollectionPrefix == "" {
		options.CollectionPrefix = options.ProductsAlias
	}
	if options.LockTTL <= 0 {
		options.LockTTL = 3 * time.Hour
	}
	if options.JobTimeout <= 0 {
		options.JobTimeout = 2 * time.Hour
	}
	if options.ProductPageTimeout <= 0 {
		options.ProductPageTimeout = 2 * time.Second
	}
	if options.ImportTimeout <= 0 {
		options.ImportTimeout = 5 * time.Second
	}
	if options.MinIndexedRatio <= 0 || options.MinIndexedRatio > 1 {
		options.MinIndexedRatio = defaultReindexMinIndexedRatio
	}
	if options.Metrics == nil {
		options.Metrics = NopReindexMetricsRecorder{}
	}
	return options
}

func reindexErrorCode(err error) string {
	switch {
	case errors.Is(err, domain.ErrInvalidReindexRequest):
		return "invalid_request"
	case errors.Is(err, domain.ErrReindexAlreadyRunning):
		return "already_running"
	case errors.Is(err, domain.ErrProductCatalogExportUnavailable):
		return "product_export_unavailable"
	case errors.Is(err, domain.ErrProductIndexUnavailable):
		return "product_index_unavailable"
	case errors.Is(err, domain.ErrReindexValidationFailed):
		return "validation_failed"
	case errors.Is(err, domain.ErrSearchCollectionUnavailable), errors.Is(err, domain.ErrSearchBackendUnavailable):
		return "search_backend_unavailable"
	case errors.Is(err, context.Canceled):
		return "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline_exceeded"
	default:
		return "unknown"
	}
}
