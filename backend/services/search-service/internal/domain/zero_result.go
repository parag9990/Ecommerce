package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	ZeroResultEventTypeSearch = "search"
	ZeroResultEventSource     = "search-service"
	DefaultSearchPath         = "/search"

	ZeroResultSkipBrowseQuery           = "browse_query"
	ZeroResultSkipMissingSessionContext = "missing_session_context"
)

const zeroResultDedupePrefix = "search:zero_result:v1:"

type ZeroResultSearchEvent struct {
	Query           string
	NormalizedQuery string
	Filters         map[string]string
	Sort            string
	Page            int
	PageSize        int
	RequestID       string
	AnonymousID     string
	SessionID       string
	UserID          string
	Path            string
	OccurredAt      time.Time
}

func (e ZeroResultSearchEvent) Normalize() ZeroResultSearchEvent {
	e.Query = strings.Join(strings.Fields(strings.TrimSpace(e.Query)), " ")
	e.NormalizedQuery = NormalizeAnalyticsQuery(firstNonEmpty(e.NormalizedQuery, e.Query))
	if e.Query == "" {
		e.Query = e.NormalizedQuery
	}
	e.Sort = strings.TrimSpace(e.Sort)
	e.RequestID = strings.TrimSpace(e.RequestID)
	e.AnonymousID = strings.TrimSpace(e.AnonymousID)
	e.SessionID = strings.TrimSpace(e.SessionID)
	e.UserID = strings.TrimSpace(e.UserID)
	e.Path = strings.TrimSpace(e.Path)
	if e.Path == "" {
		e.Path = DefaultSearchPath
	}
	if e.Page <= 0 {
		e.Page = DefaultSearchPage
	}
	if e.PageSize <= 0 {
		e.PageSize = DefaultSearchPageSize
	}
	if e.Filters == nil {
		e.Filters = map[string]string{}
	} else {
		e.Filters = cleanAnalyticsFilters(e.Filters)
	}
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	} else {
		e.OccurredAt = e.OccurredAt.UTC()
	}
	return e
}

func (e ZeroResultSearchEvent) ShouldTrack() bool {
	return e.SkipReason() == ""
}

func (e ZeroResultSearchEvent) SkipReason() string {
	e = e.Normalize()
	if e.NormalizedQuery == "" || e.NormalizedQuery == DefaultSearchQuery {
		return ZeroResultSkipBrowseQuery
	}
	if e.AnonymousID == "" || e.SessionID == "" {
		return ZeroResultSkipMissingSessionContext
	}
	return ""
}

func (e ZeroResultSearchEvent) DedupeKey() string {
	e = e.Normalize()
	sum := sha256.Sum256([]byte(strings.Join([]string{
		e.SessionID,
		e.NormalizedQuery,
		canonicalFilterString(e.Filters),
	}, "|")))
	return zeroResultDedupePrefix + hex.EncodeToString(sum[:])
}

func (e ZeroResultSearchEvent) QueryHash() string {
	e = e.Normalize()
	return ShortHash(e.NormalizedQuery)
}

func (e ZeroResultSearchEvent) SessionIDHash() string {
	e = e.Normalize()
	return ShortHash(e.SessionID)
}

func (e ZeroResultSearchEvent) QueryLength() int {
	e = e.Normalize()
	return utf8.RuneCountInString(e.NormalizedQuery)
}

func NormalizeAnalyticsQuery(q string) string {
	q = strings.Join(strings.Fields(strings.TrimSpace(q)), " ")
	if utf8.RuneCountInString(q) > MaxSearchQueryLength {
		q = string([]rune(q)[:MaxSearchQueryLength])
	}
	return strings.ToLower(q)
}

func AnalyticsFilters(input SearchInput) map[string]string {
	filters := make(map[string]string, 7)
	if len(input.Filters.Brands) > 0 {
		filters[SearchFilterBrand] = strings.Join(input.Filters.Brands, ",")
	}
	if input.Filters.CategoryID != "" {
		filters[SearchFilterCategoryID] = input.Filters.CategoryID
	}
	if input.Filters.SellerID != "" {
		filters[SearchFilterSellerID] = input.Filters.SellerID
	}
	if input.Filters.MinPrice != nil {
		filters[SearchFilterMinPrice] = fmt.Sprintf("%.2f", *input.Filters.MinPrice)
	}
	if input.Filters.MaxPrice != nil {
		filters[SearchFilterMaxPrice] = fmt.Sprintf("%.2f", *input.Filters.MaxPrice)
	}
	if input.Filters.MinRating != nil {
		filters[SearchFilterMinRating] = fmt.Sprintf("%.1f", *input.Filters.MinRating)
	}
	if input.Filters.InStock != nil {
		filters[SearchFilterInStock] = strconv.FormatBool(*input.Filters.InStock)
	}
	return filters
}

func ShortHash(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:8])
}

func cleanAnalyticsFilters(filters map[string]string) map[string]string {
	cleaned := make(map[string]string, len(filters))
	for key, value := range filters {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		cleaned[key] = value
	}
	return cleaned
}

func canonicalFilterString(filters map[string]string) string {
	if len(filters) == 0 {
		return ""
	}
	keys := make([]string, 0, len(filters))
	for key := range filters {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+filters[key])
	}
	return strings.Join(parts, "\x1f")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
