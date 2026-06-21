package config

import (
	"testing"
	"time"
)

func TestEnvParsesTypedValues(t *testing.T) {
	values := map[string]string{"ENABLED": "true", "LIMIT": "12", "TIMEOUT": "3s", "RATIO": "0.25", "ROLES": "buyer, seller"}
	env := Env{Lookup: func(key string) (string, bool) { value, ok := values[key]; return value, ok }}
	if value, err := env.Bool("ENABLED", false); err != nil || !value {
		t.Fatalf("Bool() = %v, %v", value, err)
	}
	if value, err := env.Int("LIMIT", 1, 1); err != nil || value != 12 {
		t.Fatalf("Int() = %v, %v", value, err)
	}
	if value, err := env.Duration("TIMEOUT", time.Second); err != nil || value != 3*time.Second {
		t.Fatalf("Duration() = %v, %v", value, err)
	}
	if value, err := env.Float64("RATIO", 1, 0, 1); err != nil || value != 0.25 {
		t.Fatalf("Float64() = %v, %v", value, err)
	}
	if value := env.CSV("ROLES", nil); len(value) != 2 || value[1] != "seller" {
		t.Fatalf("CSV() = %#v", value)
	}
}

func TestEnvRejectsInvalidValuesAndRequiredMissing(t *testing.T) {
	env := Env{Lookup: func(key string) (string, bool) {
		values := map[string]string{"BOOL": "sometimes", "INT": "0", "DURATION": "-1s"}
		value, ok := values[key]
		return value, ok
	}}
	if _, err := env.Required("MISSING"); err == nil {
		t.Fatal("Required() expected error")
	}
	if _, err := env.Bool("BOOL", false); err == nil {
		t.Fatal("Bool() expected error")
	}
	if _, err := env.Int("INT", 1, 1); err == nil {
		t.Fatal("Int() expected error")
	}
	if _, err := env.Duration("DURATION", time.Second); err == nil {
		t.Fatal("Duration() expected error")
	}
}
