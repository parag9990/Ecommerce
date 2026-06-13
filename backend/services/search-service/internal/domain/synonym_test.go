package domain

import (
	"errors"
	"testing"
)

func TestNormalizeSearchSynonymInput(t *testing.T) {
	got, err := NormalizeSearchSynonymInput(SearchSynonymInput{
		Root:     "  Smart   Phone ",
		Synonyms: []string{" Phone ", "CELL Phone", "mobile-phone"},
	})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if got.Root != "smart phone" {
		t.Fatalf("root = %q", got.Root)
	}
	want := []string{"phone", "cell phone", "mobile-phone"}
	if len(got.Synonyms) != len(want) {
		t.Fatalf("synonyms = %#v", got.Synonyms)
	}
	for i := range want {
		if got.Synonyms[i] != want[i] {
			t.Fatalf("synonyms = %#v, want %#v", got.Synonyms, want)
		}
	}
}

func TestNormalizeSearchSynonymInputRejectsInvalidTerms(t *testing.T) {
	tests := map[string]SearchSynonymInput{
		"empty root": {
			Root:     " ",
			Synonyms: []string{"phone"},
		},
		"short root": {
			Root:     "a",
			Synonyms: []string{"phone"},
		},
		"empty synonyms": {
			Root:     "mobile",
			Synonyms: []string{},
		},
		"blank synonym": {
			Root:     "mobile",
			Synonyms: []string{" "},
		},
		"duplicate synonym": {
			Root:     "mobile",
			Synonyms: []string{"phone", "Phone"},
		},
		"root repeated": {
			Root:     "mobile",
			Synonyms: []string{"mobile"},
		},
		"unsupported characters": {
			Root:     "mobile",
			Synonyms: []string{"phone<script>"},
		},
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := NormalizeSearchSynonymInput(input)
			if !errors.Is(err, ErrInvalidSynonym) {
				t.Fatalf("expected invalid synonym, got %v", err)
			}
		})
	}
}

func TestNormalizeSearchSynonymInputRejectsTooManySynonyms(t *testing.T) {
	synonyms := make([]string, 0, MaxSearchSynonymTerms+1)
	for i := 0; i < MaxSearchSynonymTerms+1; i++ {
		synonyms = append(synonyms, "term"+string(rune('a'+i)))
	}
	_, err := NormalizeSearchSynonymInput(SearchSynonymInput{
		Root:     "mobile",
		Synonyms: synonyms,
	})
	if !errors.Is(err, ErrInvalidSynonym) {
		t.Fatalf("expected invalid synonym, got %v", err)
	}
}

func TestBuildSearchSynonymID(t *testing.T) {
	tests := map[string]string{
		"mobile":      "syn_mobile",
		"smart phone": "syn_smart_phone",
		"t-shirt":     "syn_t_shirt",
		"TV & Audio":  "syn_tv_audio",
	}
	for input, want := range tests {
		if got := BuildSearchSynonymID(input); got != want {
			t.Fatalf("BuildSearchSynonymID(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNormalizeSearchSynonymPageRequest(t *testing.T) {
	got := NormalizeSearchSynonymPageRequest(SearchSynonymPageRequest{Page: -1, PageSize: 1000})
	if got.Page != DefaultSynonymPage {
		t.Fatalf("page = %d", got.Page)
	}
	if got.PageSize != MaxSynonymPageSize {
		t.Fatalf("page_size = %d", got.PageSize)
	}
}
