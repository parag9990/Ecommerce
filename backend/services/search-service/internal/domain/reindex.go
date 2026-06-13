package domain

import (
	"fmt"
	"strings"
	"time"
)

const (
	ReindexModeAlias   = "alias"
	ReindexModeInPlace = "in_place"
)

type ReindexRequest struct {
	JobID            string
	Mode             string
	BatchSize        int
	Reason           string
	ActorID          string
	DryRun           bool
	TargetCollection string
}

type ReindexProgress struct {
	JobID            string
	Mode             string
	TargetCollection string
	SourceCursor     string
	ProductsRead     int
	ProductsIndexed  int
	ProductsSkipped  int
	FailedDocuments  int
	StartedAt        time.Time
	LastProgressAt   time.Time
}

type ReindexResult struct {
	JobID              string
	Mode               string
	TargetCollection   string
	PreviousCollection string
	ProductsRead       int
	ProductsIndexed    int
	ProductsSkipped    int
	FailedDocuments    int
	SourceTotal        int
	Duration           time.Duration
	AliasSwapped       bool
	DryRun             bool
}

type ReindexAccepted struct {
	JobID  string
	Status string
}

type ProductExportRequest struct {
	Cursor string
	Limit  int
}

type ProductExportPage struct {
	Items      []ProductIndexPayload
	NextCursor string
	HasMore    bool
	Total      int
}

func NormalizeReindexRequest(req ReindexRequest, defaultMode string, defaultBatchSize int, now time.Time) ReindexRequest {
	req.JobID = strings.TrimSpace(req.JobID)
	if req.JobID == "" {
		req.JobID = NewReindexJobID(now)
	}
	req.Mode = strings.ToLower(strings.TrimSpace(req.Mode))
	if req.Mode == "" {
		req.Mode = strings.ToLower(strings.TrimSpace(defaultMode))
	}
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" {
		req.Reason = "manual"
	}
	req.ActorID = strings.TrimSpace(req.ActorID)
	req.TargetCollection = strings.TrimSpace(req.TargetCollection)
	if req.BatchSize <= 0 {
		req.BatchSize = defaultBatchSize
	}
	return req
}

func (r ReindexRequest) Validate(maxBatchSize int) error {
	switch r.Mode {
	case ReindexModeAlias, ReindexModeInPlace:
	default:
		return fmt.Errorf("%w: mode must be alias or in_place", ErrInvalidReindexRequest)
	}
	if r.BatchSize <= 0 {
		return fmt.Errorf("%w: batch_size must be greater than zero", ErrInvalidReindexRequest)
	}
	if maxBatchSize > 0 && r.BatchSize > maxBatchSize {
		return fmt.Errorf("%w: batch_size must be less than or equal to %d", ErrInvalidReindexRequest, maxBatchSize)
	}
	if strings.TrimSpace(r.Reason) == "" {
		return fmt.Errorf("%w: reason is required", ErrInvalidReindexRequest)
	}
	if strings.TrimSpace(r.JobID) == "" {
		return fmt.Errorf("%w: job_id is required", ErrInvalidReindexRequest)
	}
	return nil
}

func NewReindexJobID(now time.Time) string {
	if now.IsZero() {
		now = time.Now()
	}
	return "search_reindex_" + now.UTC().Format("20060102_150405")
}

func IsSupportedReindexMode(mode string) bool {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case ReindexModeAlias, ReindexModeInPlace:
		return true
	default:
		return false
	}
}
