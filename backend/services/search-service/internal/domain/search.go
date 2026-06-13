package domain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	DefaultSearchQuery      = "*"
	MaxSearchQueryLength    = 120
	MaxSearchPageSize       = 100
	DefaultSearchPage       = 1
	DefaultSearchPageSize   = 20
	MaxSearchFilterValues   = 20
	MaxSearchFilterValueLen = 128
)

const (
	SearchFilterBrand      = "brand"
	SearchFilterCategoryID = "category_id"
	SearchFilterSellerID   = "seller_id"
	SearchFilterMinPrice   = "min_price"
	SearchFilterMaxPrice   = "max_price"
	SearchFilterMinRating  = "min_rating"
	SearchFilterInStock    = "in_stock"
)

var idLikePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)

type SearchRequest struct {
	Query    string
	Filters  map[string]string
	Sort     string
	Page     int
	PageSize int
}

type SearchInput struct {
	Query    string
	Filters  SearchFilters
	Sort     string
	Page     int
	PageSize int
}

type SearchFilters struct {
	Brands     []string
	CategoryID string
	SellerID   string
	MinPrice   *float64
	MaxPrice   *float64
	MinRating  *float64
	InStock    *bool
}

type SearchLimits struct {
	DefaultPage     int
	DefaultPageSize int
	MaxPageSize     int
}

type ProductSearchQuery struct {
	Query          string
	QueryBy        string
	QueryByWeights string
	NumTypos       string
	Prefix         string
	FacetBy        string
	FilterBy       string
	SortBy         string
	Page           int
	PageSize       int
}

type ProductSearchResult struct {
	IDs          []string
	Facets       map[string][]FacetValue
	Total        int
	SearchTimeMS int
}

type FacetValue struct {
	Value string
	Count int
}

type SearchResponse struct {
	Products []ProductSummary
	Facets   map[string][]FacetValue
	Total    int
}

type ProductSummary struct {
	ProductID   string
	SellerID    string
	Title       string
	Description string
	Brand       string
	CategoryID  string
	Status      string
	Variants    []ProductVariant
}

type ProductVariant struct {
	SKU           string
	Attributes    map[string]any
	Price         *Money
	StockQuantity int
}

type Money struct {
	Amount   float64
	Currency string
}

func NormalizeSearchRequest(req SearchRequest, policy SearchPolicy, limits SearchLimits) (SearchInput, error) {
	limits = normalizeSearchLimits(policy, limits)

	query := strings.Join(strings.Fields(strings.TrimSpace(req.Query)), " ")
	if query == "" {
		query = DefaultSearchQuery
	}
	if utf8.RuneCountInString(query) > MaxSearchQueryLength {
		return SearchInput{}, fmt.Errorf("%w: query is too long", ErrInvalidSearchRequest)
	}

	page := req.Page
	if page <= 0 {
		page = limits.DefaultPage
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = limits.DefaultPageSize
	}
	if pageSize > limits.MaxPageSize {
		return SearchInput{}, fmt.Errorf("%w: page_size must be less than or equal to %d", ErrInvalidSearchRequest, limits.MaxPageSize)
	}

	sortKey := strings.TrimSpace(req.Sort)
	if sortKey == "" {
		sortKey = policy.DefaultSort
	}
	if _, err := ResolveSearchSort(policy, sortKey); err != nil {
		return SearchInput{}, err
	}

	filters, err := ParseSearchFilters(req.Filters)
	if err != nil {
		return SearchInput{}, err
	}

	return SearchInput{
		Query:    query,
		Filters:  filters,
		Sort:     sortKey,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func ParseSearchFilters(raw map[string]string) (SearchFilters, error) {
	var filters SearchFilters
	for key, value := range raw {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		switch key {
		case SearchFilterBrand:
			brands, err := parseBrandFilter(value)
			if err != nil {
				return SearchFilters{}, err
			}
			filters.Brands = brands
		case SearchFilterCategoryID:
			if err := validateIDFilter(SearchFilterCategoryID, value); err != nil {
				return SearchFilters{}, err
			}
			filters.CategoryID = value
		case SearchFilterSellerID:
			if err := validateIDFilter(SearchFilterSellerID, value); err != nil {
				return SearchFilters{}, err
			}
			filters.SellerID = value
		case SearchFilterMinPrice:
			amount, err := parseNonNegativeFloat(SearchFilterMinPrice, value)
			if err != nil {
				return SearchFilters{}, err
			}
			filters.MinPrice = &amount
		case SearchFilterMaxPrice:
			amount, err := parseNonNegativeFloat(SearchFilterMaxPrice, value)
			if err != nil {
				return SearchFilters{}, err
			}
			filters.MaxPrice = &amount
		case SearchFilterMinRating:
			rating, err := parseNonNegativeFloat(SearchFilterMinRating, value)
			if err != nil {
				return SearchFilters{}, err
			}
			if rating > 5 {
				return SearchFilters{}, fmt.Errorf("%w: min_rating must be between 0 and 5", ErrInvalidSearchFilter)
			}
			filters.MinRating = &rating
		case SearchFilterInStock:
			inStock, err := parseStrictBool(SearchFilterInStock, value)
			if err != nil {
				return SearchFilters{}, err
			}
			filters.InStock = &inStock
		default:
			return SearchFilters{}, fmt.Errorf("%w: unsupported filter %q", ErrInvalidSearchFilter, key)
		}
	}

	if filters.MinPrice != nil && filters.MaxPrice != nil && *filters.MinPrice > *filters.MaxPrice {
		return SearchFilters{}, fmt.Errorf("%w: min_price cannot be greater than max_price", ErrInvalidSearchFilter)
	}
	return filters, nil
}

func BuildProductFilterBy(filters SearchFilters) (string, error) {
	parts := make([]string, 0, 7)

	if len(filters.Brands) > 0 {
		values := make([]string, 0, len(filters.Brands))
		for _, brand := range filters.Brands {
			if err := validateStringFilter(SearchFilterBrand, brand); err != nil {
				return "", err
			}
			values = append(values, quoteFilterValue(brand))
		}
		parts = append(parts, fmt.Sprintf("%s:=[%s]", ProductFieldBrand, strings.Join(values, ",")))
	}
	if filters.CategoryID != "" {
		if err := validateIDFilter(SearchFilterCategoryID, filters.CategoryID); err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("%s:=%s", ProductFieldCategoryIDs, quoteFilterValue(filters.CategoryID)))
	}
	if filters.SellerID != "" {
		if err := validateIDFilter(SearchFilterSellerID, filters.SellerID); err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("%s:=%s", ProductFieldSellerID, quoteFilterValue(filters.SellerID)))
	}
	if filters.MinPrice != nil {
		parts = append(parts, fmt.Sprintf("%s:>=%s", ProductFieldPrice, formatFloat(*filters.MinPrice)))
	}
	if filters.MaxPrice != nil {
		parts = append(parts, fmt.Sprintf("%s:<=%s", ProductFieldPrice, formatFloat(*filters.MaxPrice)))
	}
	if filters.MinRating != nil {
		parts = append(parts, fmt.Sprintf("%s:>=%s", ProductFieldRating, formatFloat(*filters.MinRating)))
	}
	if filters.InStock != nil {
		parts = append(parts, fmt.Sprintf("%s:=%t", ProductFieldInStock, *filters.InStock))
	}

	return strings.Join(parts, " && "), nil
}

func ResolveSearchSort(policy SearchPolicy, sortKey string) (string, error) {
	sortKey = strings.TrimSpace(sortKey)
	for _, option := range policy.SortOptions {
		if option.Key == sortKey {
			if strings.TrimSpace(option.TypesenseBy) == "" {
				return "", fmt.Errorf("%w: sort %q has no backend mapping", ErrUnsupportedSearchSort, sortKey)
			}
			return option.TypesenseBy, nil
		}
	}
	return "", fmt.Errorf("%w: %q", ErrUnsupportedSearchSort, sortKey)
}

func AllowedFacetSet(policy SearchPolicy) map[string]struct{} {
	allowed := make(map[string]struct{}, len(policy.FacetFields))
	for _, field := range policy.FacetFields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		allowed[field] = struct{}{}
	}
	return allowed
}

func normalizeSearchLimits(policy SearchPolicy, limits SearchLimits) SearchLimits {
	if limits.DefaultPage <= 0 {
		limits.DefaultPage = policy.DefaultPage
	}
	if limits.DefaultPage <= 0 {
		limits.DefaultPage = DefaultSearchPage
	}
	if limits.DefaultPageSize <= 0 {
		limits.DefaultPageSize = policy.DefaultPageSize
	}
	if limits.DefaultPageSize <= 0 {
		limits.DefaultPageSize = DefaultSearchPageSize
	}
	if limits.MaxPageSize <= 0 {
		limits.MaxPageSize = MaxSearchPageSize
	}
	return limits
}

func parseBrandFilter(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		value := strings.Join(strings.Fields(strings.TrimSpace(part)), " ")
		if value == "" {
			continue
		}
		if err := validateStringFilter(SearchFilterBrand, value); err != nil {
			return nil, err
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		values = append(values, value)
	}
	if len(values) > MaxSearchFilterValues {
		return nil, fmt.Errorf("%w: brand can contain at most %d values", ErrInvalidSearchFilter, MaxSearchFilterValues)
	}
	return values, nil
}

func validateStringFilter(name string, value string) error {
	if value == "" {
		return fmt.Errorf("%w: %s cannot be empty", ErrInvalidSearchFilter, name)
	}
	if utf8.RuneCountInString(value) > MaxSearchFilterValueLen {
		return fmt.Errorf("%w: %s is too long", ErrInvalidSearchFilter, name)
	}
	if strings.ContainsAny(value, "`\r\n\t") {
		return fmt.Errorf("%w: %s contains unsupported characters", ErrInvalidSearchFilter, name)
	}
	return nil
}

func validateIDFilter(name string, value string) error {
	if err := validateStringFilter(name, value); err != nil {
		return err
	}
	if !idLikePattern.MatchString(value) {
		return fmt.Errorf("%w: %s has invalid format", ErrInvalidSearchFilter, name)
	}
	return nil
}

func parseNonNegativeFloat(name string, raw string) (float64, error) {
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s must be numeric", ErrInvalidSearchFilter, name)
	}
	if value < 0 {
		return 0, fmt.Errorf("%w: %s cannot be negative", ErrInvalidSearchFilter, name)
	}
	return value, nil
}

func parseStrictBool(name string, raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("%w: %s must be true or false", ErrInvalidSearchFilter, name)
	}
}

func quoteFilterValue(value string) string {
	return "`" + value + "`"
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
