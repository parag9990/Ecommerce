package domain

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	MinSynonymTermLength   = 2
	MaxSynonymTermLength   = 50
	MaxSearchSynonymTerms  = 20
	DefaultSynonymPage     = 1
	DefaultSynonymPageSize = 50
	MaxSynonymPageSize     = 100
	searchSynonymIDPrefix  = "syn_"
)

var allowedSynonymTermPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9 +&.-]{0,49}$`)
var synonymIDSeparatorPattern = regexp.MustCompile(`_+`)

type SynonymModel struct {
	Shape    Synonym   `json:"shape"`
	Examples []Synonym `json:"examples"`
}

type Synonym struct {
	Root     string   `json:"root"`
	Synonyms []string `json:"synonyms"`
}

type SearchSynonymInput struct {
	Root     string
	Synonyms []string
}

type SearchSynonym struct {
	ID       string
	Root     string
	Synonyms []string
}

type SearchSynonymPageRequest struct {
	Page     int
	PageSize int
}

func (m SynonymModel) Validate() error {
	if err := m.Shape.Validate(); err != nil {
		return err
	}
	for _, example := range m.Examples {
		if err := example.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (s Synonym) Validate() error {
	root := NormalizeSynonymTerm(s.Root)
	if root == "" {
		return fmt.Errorf("%w: root is required", ErrInvalidSynonym)
	}
	if len(s.Synonyms) == 0 {
		return fmt.Errorf("%w: at least one synonym is required", ErrInvalidSynonym)
	}

	seen := map[string]struct{}{root: {}}
	for i, synonym := range s.Synonyms {
		term := NormalizeSynonymTerm(synonym)
		if term == "" {
			return fmt.Errorf("%w: synonyms[%d] is required", ErrInvalidSynonym, i)
		}
		if _, ok := seen[term]; ok {
			return fmt.Errorf("%w: duplicate synonym term %q", ErrInvalidSynonym, synonym)
		}
		seen[term] = struct{}{}
	}
	return nil
}

func NormalizeSearchSynonymInput(input SearchSynonymInput) (SearchSynonymInput, error) {
	root := NormalizeSynonymTerm(input.Root)
	if err := validateSynonymTerm("root", root); err != nil {
		return SearchSynonymInput{}, err
	}
	if len(input.Synonyms) == 0 {
		return SearchSynonymInput{}, fmt.Errorf("%w: at least one synonym is required", ErrInvalidSynonym)
	}
	if len(input.Synonyms) > MaxSearchSynonymTerms {
		return SearchSynonymInput{}, fmt.Errorf("%w: maximum %d synonyms allowed", ErrInvalidSynonym, MaxSearchSynonymTerms)
	}

	seen := map[string]struct{}{root: {}}
	synonyms := make([]string, 0, len(input.Synonyms))
	for i, raw := range input.Synonyms {
		term := NormalizeSynonymTerm(raw)
		if term == "" {
			return SearchSynonymInput{}, fmt.Errorf("%w: synonyms[%d] is required", ErrInvalidSynonym, i)
		}
		if err := validateSynonymTerm(fmt.Sprintf("synonyms[%d]", i), term); err != nil {
			return SearchSynonymInput{}, err
		}
		if _, exists := seen[term]; exists {
			if term == root {
				return SearchSynonymInput{}, fmt.Errorf("%w: synonym cannot equal root", ErrInvalidSynonym)
			}
			return SearchSynonymInput{}, fmt.Errorf("%w: duplicate synonym term %q", ErrInvalidSynonym, term)
		}
		seen[term] = struct{}{}
		synonyms = append(synonyms, term)
	}

	return SearchSynonymInput{
		Root:     root,
		Synonyms: synonyms,
	}, nil
}

func NewSearchSynonym(input SearchSynonymInput) (SearchSynonym, error) {
	normalized, err := NormalizeSearchSynonymInput(input)
	if err != nil {
		return SearchSynonym{}, err
	}
	return SearchSynonym{
		ID:       BuildSearchSynonymID(normalized.Root),
		Root:     normalized.Root,
		Synonyms: normalized.Synonyms,
	}, nil
}

func (s SearchSynonym) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return fmt.Errorf("%w: synonym_id is required", ErrInvalidSynonym)
	}
	_, err := NormalizeSearchSynonymInput(SearchSynonymInput{
		Root:     s.Root,
		Synonyms: s.Synonyms,
	})
	return err
}

func NormalizeSearchSynonymPageRequest(req SearchSynonymPageRequest) SearchSynonymPageRequest {
	if req.Page <= 0 {
		req.Page = DefaultSynonymPage
	}
	if req.PageSize <= 0 {
		req.PageSize = DefaultSynonymPageSize
	}
	if req.PageSize > MaxSynonymPageSize {
		req.PageSize = MaxSynonymPageSize
	}
	return req
}

func BuildSearchSynonymID(root string) string {
	root = NormalizeSynonymTerm(root)
	var b strings.Builder
	lastWasSeparator := false
	for _, r := range root {
		isAlphaNumeric := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlphaNumeric {
			b.WriteRune(r)
			lastWasSeparator = false
			continue
		}
		if !lastWasSeparator {
			b.WriteRune('_')
			lastWasSeparator = true
		}
	}
	slug := strings.Trim(b.String(), "_")
	slug = synonymIDSeparatorPattern.ReplaceAllString(slug, "_")
	if slug == "" {
		slug = "term"
	}
	return searchSynonymIDPrefix + slug
}

func NormalizeSynonymTerm(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}

func validateSynonymTerm(field string, term string) error {
	if term == "" {
		return fmt.Errorf("%w: %s is required", ErrInvalidSynonym, field)
	}
	if len([]rune(term)) < MinSynonymTermLength || len([]rune(term)) > MaxSynonymTermLength {
		return fmt.Errorf("%w: %s length must be between %d and %d", ErrInvalidSynonym, field, MinSynonymTermLength, MaxSynonymTermLength)
	}
	if !allowedSynonymTermPattern.MatchString(term) {
		return fmt.Errorf("%w: %s contains unsupported characters", ErrInvalidSynonym, field)
	}
	return nil
}
