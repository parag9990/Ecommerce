package domain

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	DefaultAutocompleteLimit       = 8
	MaxAutocompleteLimit           = 10
	MaxAutocompleteQueryLength     = 80
	MaxAutocompleteSuggestionLen   = 120
	MinAutocompleteProductPrefix   = 2
	AutocompleteCacheKeyPrefix     = "autocomplete:v1"
	AutocompleteSourcePopularQuery = "popular_query"
	AutocompleteSourceProductTitle = "product_title"
	AutocompleteSourceProductBrand = "product_brand"
)

const (
	PopularQueriesCollectionName     = "popular_queries"
	PopularQueryFieldID              = "id"
	PopularQueryFieldQuery           = "query"
	PopularQueryFieldNormalizedQuery = "normalized_query"
	PopularQueryFieldScore           = "score"
	PopularQueryFieldCount           = "count"
	PopularQueryFieldLastSeenAt      = "last_seen_at"
	PopularQueryFieldIsActive        = "is_active"
	PopularQueryFieldLocale          = "locale"
	DefaultPopularQuerySortingField  = PopularQueryFieldScore
)

type AutocompleteRequest struct {
	Query string
	Limit int
}

type AutocompleteInput struct {
	Query           string
	NormalizedQuery string
	Limit           int
}

type AutocompleteResponse struct {
	Suggestions []string
}

type SuggestionCandidate struct {
	Text   string
	Source string
	Score  int
}

type PopularQueryDocument struct {
	ID              string  `json:"id"`
	Query           string  `json:"query"`
	NormalizedQuery string  `json:"normalized_query"`
	Score           int32   `json:"score"`
	Count           int32   `json:"count"`
	LastSeenAt      int64   `json:"last_seen_at"`
	IsActive        bool    `json:"is_active"`
	Locale          *string `json:"locale,omitempty"`
}

func (d PopularQueryDocument) Validate() error {
	if strings.TrimSpace(d.ID) == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidPopularQueryDocument)
	}
	if strings.TrimSpace(d.Query) == "" {
		return fmt.Errorf("%w: query is required", ErrInvalidPopularQueryDocument)
	}
	normalized := NormalizeAutocompleteText(d.Query)
	if normalized == "" || strings.TrimSpace(d.NormalizedQuery) == "" {
		return fmt.Errorf("%w: normalized query is required", ErrInvalidPopularQueryDocument)
	}
	if d.NormalizedQuery != normalized {
		return fmt.Errorf("%w: normalized query must match query", ErrInvalidPopularQueryDocument)
	}
	if d.Score < 0 {
		return fmt.Errorf("%w: score cannot be negative", ErrInvalidPopularQueryDocument)
	}
	if d.Count < 0 {
		return fmt.Errorf("%w: count cannot be negative", ErrInvalidPopularQueryDocument)
	}
	if d.LastSeenAt < 0 {
		return fmt.Errorf("%w: last_seen_at cannot be negative", ErrInvalidPopularQueryDocument)
	}
	return nil
}

func NormalizeAutocompleteRequest(req AutocompleteRequest, defaultLimit int, maxLimit int) (AutocompleteInput, error) {
	if defaultLimit <= 0 {
		defaultLimit = DefaultAutocompleteLimit
	}
	if maxLimit <= 0 {
		maxLimit = MaxAutocompleteLimit
	}
	if maxLimit > MaxAutocompleteLimit {
		maxLimit = MaxAutocompleteLimit
	}
	if defaultLimit > maxLimit {
		defaultLimit = maxLimit
	}

	query := strings.Join(strings.Fields(strings.TrimSpace(req.Query)), " ")
	if utf8.RuneCountInString(query) > MaxAutocompleteQueryLength {
		return AutocompleteInput{}, fmt.Errorf("%w: query must be less than or equal to %d characters", ErrInvalidAutocompleteRequest, MaxAutocompleteQueryLength)
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		return AutocompleteInput{}, fmt.Errorf("%w: limit must be less than or equal to %d", ErrInvalidAutocompleteRequest, maxLimit)
	}

	return AutocompleteInput{
		Query:           query,
		NormalizedQuery: NormalizeAutocompleteText(query),
		Limit:           limit,
	}, nil
}

func NormalizeAutocompleteText(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}

func BuildAutocompleteCacheKey(normalizedQuery string, limit int) string {
	return fmt.Sprintf("%s:%s:%d", AutocompleteCacheKeyPrefix, NormalizeAutocompleteText(normalizedQuery), limit)
}

func MergeAutocompleteCandidates(query string, limit int, groups ...[]SuggestionCandidate) []string {
	if limit <= 0 {
		return []string{}
	}

	normalizedQuery := NormalizeAutocompleteText(query)
	byKey := make(map[string]rankedSuggestion)

	for groupIndex, group := range groups {
		for candidateIndex, candidate := range group {
			text := cleanSuggestionText(candidate.Text)
			if text == "" || utf8.RuneCountInString(text) > MaxAutocompleteSuggestionLen || hasControlRune(text) {
				continue
			}

			key := NormalizeAutocompleteText(text)
			if key == "" {
				continue
			}

			ranked := rankedSuggestion{
				Text:        text,
				Key:         key,
				Source:      strings.TrimSpace(candidate.Source),
				Score:       rankAutocompleteCandidate(normalizedQuery, key, candidate),
				GroupIndex:  groupIndex,
				SourceIndex: candidateIndex,
			}
			if existing, ok := byKey[key]; !ok || ranked.betterThan(existing) {
				byKey[key] = ranked
			}
		}
	}

	ranked := make([]rankedSuggestion, 0, len(byKey))
	for _, candidate := range byKey {
		ranked = append(ranked, candidate)
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		left := ranked[i]
		right := ranked[j]
		if left.Score != right.Score {
			return left.Score > right.Score
		}
		if sourcePriority(left.Source) != sourcePriority(right.Source) {
			return sourcePriority(left.Source) > sourcePriority(right.Source)
		}
		if utf8.RuneCountInString(left.Text) != utf8.RuneCountInString(right.Text) {
			return utf8.RuneCountInString(left.Text) < utf8.RuneCountInString(right.Text)
		}
		if left.GroupIndex != right.GroupIndex {
			return left.GroupIndex < right.GroupIndex
		}
		if left.SourceIndex != right.SourceIndex {
			return left.SourceIndex < right.SourceIndex
		}
		return strings.ToLower(left.Text) < strings.ToLower(right.Text)
	})

	if len(ranked) < limit {
		limit = len(ranked)
	}
	suggestions := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		suggestions = append(suggestions, ranked[i].Text)
	}
	return suggestions
}

type rankedSuggestion struct {
	Text        string
	Key         string
	Source      string
	Score       int
	GroupIndex  int
	SourceIndex int
}

func (s rankedSuggestion) betterThan(other rankedSuggestion) bool {
	if s.Score != other.Score {
		return s.Score > other.Score
	}
	if sourcePriority(s.Source) != sourcePriority(other.Source) {
		return sourcePriority(s.Source) > sourcePriority(other.Source)
	}
	return s.GroupIndex < other.GroupIndex || (s.GroupIndex == other.GroupIndex && s.SourceIndex < other.SourceIndex)
}

func rankAutocompleteCandidate(query string, candidateKey string, candidate SuggestionCandidate) int {
	score := candidate.Score + sourcePriority(candidate.Source)
	if query == "" {
		return score + readabilityBoost(candidateKey)
	}
	switch {
	case candidateKey == query:
		score += 800
	case strings.HasPrefix(candidateKey, query):
		score += 500
	case strings.Contains(candidateKey, " "+query):
		score += 150
	}
	return score + readabilityBoost(candidateKey)
}

func sourcePriority(source string) int {
	switch strings.TrimSpace(source) {
	case AutocompleteSourcePopularQuery:
		return 500
	case AutocompleteSourceProductTitle:
		return 300
	case AutocompleteSourceProductBrand:
		return 250
	default:
		return 0
	}
}

func readabilityBoost(text string) int {
	runes := utf8.RuneCountInString(text)
	switch {
	case runes == 0:
		return 0
	case runes <= 24:
		return 60 - runes
	case runes <= 48:
		return 20
	default:
		return 0
	}
}

func cleanSuggestionText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func hasControlRune(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}
