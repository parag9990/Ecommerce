package config

import "testing"

func TestLoadRuntimeAddressesAndExpiryWorker(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	t.Setenv("PRODUCT_HTTP_ADDR", ":18082")
	t.Setenv("PRODUCT_GRPC_ADDR", ":19092")
	t.Setenv("PRODUCT_INVENTORY_EXPIRY_INTERVAL_SECONDS", "15")
	t.Setenv("PRODUCT_INTERNAL_SERVICE_TOKEN", "staging-service-token")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Environment != "staging" || cfg.HTTPAddress != ":18082" || cfg.GRPCAddress != ":19092" {
		t.Fatalf("runtime config not loaded: %+v", cfg)
	}
	if cfg.Inventory.ExpiryIntervalSeconds != 15 {
		t.Fatalf("expiry interval = %d", cfg.Inventory.ExpiryIntervalSeconds)
	}
}

func TestProductionRejectsDefaultInternalToken(t *testing.T) {
	cfg := Default()
	cfg.Environment = "production"
	if err := cfg.Validate(); err == nil {
		t.Fatal("production config accepted the local internal token")
	}
	cfg.InternalServiceToken = "production-secret-from-secret-manager"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("production config rejected explicit token: %v", err)
	}
}
