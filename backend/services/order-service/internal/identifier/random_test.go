package identifier

import (
	"strings"
	"testing"
)

func TestCryptoGeneratorProducesPrefixedDistinctIDs(t *testing.T) {
	generator := NewCryptoGenerator()
	first, err := generator.NewID("ord")
	if err != nil {
		t.Fatalf("NewID() error = %v", err)
	}
	second, err := generator.NewID("ord")
	if err != nil {
		t.Fatalf("NewID() error = %v", err)
	}
	if !strings.HasPrefix(first, "ord_") || first == second {
		t.Fatalf("generated IDs %q/%q do not have prefix or are identical", first, second)
	}
}

func TestCryptoGeneratorRejectsUnsafePrefix(t *testing.T) {
	if _, err := NewCryptoGenerator().NewID("Order ID"); err == nil {
		t.Fatal("NewID() error = nil, want invalid prefix error")
	}
}
