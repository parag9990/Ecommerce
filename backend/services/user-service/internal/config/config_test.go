package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadIncludesAdminHTTPAndSafeWorkerDefaults(t *testing.T) {
	t.Setenv("USER_SERVICE_DATABASE_DSN", "user:password@tcp(localhost:3306)/user_db?parseTime=true")
	t.Setenv("USER_SERVICE_ADMIN_TOKEN", "test_user_admin_token_at_least_32_chars")
	t.Setenv("USER_SERVICE_HTTP_ADDRESS", "")
	t.Setenv("OUTBOX_WORKER_ENABLED", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.HTTP.Address != ":9091" {
		t.Fatalf("HTTP address = %q, want :9091", cfg.HTTP.Address)
	}
	if cfg.HTTP.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("read header timeout = %v", cfg.HTTP.ReadHeaderTimeout)
	}
	if cfg.OutboxWorker.Enabled {
		t.Fatal("outbox worker should be disabled")
	}
	if strings.Contains(cfg.String(), "password") {
		t.Fatal("Config.String must not expose the database DSN")
	}
}

func TestLoadRejectsWorkerWithoutBrokerConfiguration(t *testing.T) {
	t.Setenv("USER_SERVICE_DATABASE_DSN", "user:password@tcp(localhost:3306)/user_db?parseTime=true")
	t.Setenv("USER_SERVICE_ADMIN_TOKEN", "test_user_admin_token_at_least_32_chars")
	t.Setenv("OUTBOX_WORKER_ENABLED", "true")
	t.Setenv("USER_EVENTS_ENABLED", "true")
	t.Setenv("USER_EVENTS_PROVIDER", "rabbitmq")
	t.Setenv("RABBITMQ_URL", "")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "RABBITMQ_URL") {
		t.Fatalf("Load error = %v, want missing RABBITMQ_URL", err)
	}
}
