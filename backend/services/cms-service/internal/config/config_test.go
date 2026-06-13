package config

import (
	"strings"
	"testing"
	"time"
)

func TestDatabaseConfigValidateAllowsComponentConfig(t *testing.T) {
	cfg := DatabaseConfig{
		Host:            "localhost",
		Port:            3306,
		Name:            "cms_db",
		User:            "cms_user",
		Password:        "secret",
		Timezone:        "UTC",
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
	}

	if err := cfg.Validate("production"); err != nil {
		t.Fatalf("expected valid config: %v", err)
	}
}

func TestDatabaseConfigValidateRequiresProductionPassword(t *testing.T) {
	cfg := DatabaseConfig{
		Host:            "localhost",
		Port:            3306,
		Name:            "cms_db",
		User:            "cms_user",
		Timezone:        "UTC",
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
	}

	err := cfg.Validate("production")
	if err == nil || !strings.Contains(err.Error(), "CMS_DB_PASSWORD") {
		t.Fatalf("expected production password validation error, got %v", err)
	}
}

func TestConfigValidateRequiresDatabaseWhenStaffSourceIsMySQL(t *testing.T) {
	cfg := Config{
		HTTP: HTTPConfig{
			Address:         ":8087",
			ReadTimeout:     time.Second,
			WriteTimeout:    time.Second,
			IdleTimeout:     time.Second,
			ShutdownTimeout: time.Second,
			MaxBodyBytes:    1024,
		},
		Security: SecurityConfig{
			Environment:        "local",
			InternalAuthHeader: "X-Internal-Token",
		},
		Database: DatabaseConfig{
			Host:            "",
			Port:            3306,
			Name:            "cms_db",
			User:            "cms_user",
			Timezone:        "UTC",
			MaxOpenConns:    25,
			MaxIdleConns:    10,
			ConnMaxLifetime: 5 * time.Minute,
		},
		Authorization: AuthorizationConfig{
			StaffStatusSource: StaffStatusSourceMySQL,
		},
	}

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "CMS_DB_HOST") {
		t.Fatalf("expected database host validation error, got %v", err)
	}
}

func TestAnalyticsConfigValidateBoundsDashboardQueries(t *testing.T) {
	cfg := AnalyticsConfig{
		DefaultCurrency:         "INR",
		DefaultRangeDays:        30,
		MaxRangeDays:            366,
		DefaultTopProductsLimit: 5,
		MaxTopProductsLimit:     20,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid analytics config: %v", err)
	}

	cfg.MaxTopProductsLimit = 21
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "CMS_ANALYTICS_MAX_TOP_PRODUCTS_LIMIT") {
		t.Fatalf("expected top product limit validation error, got %v", err)
	}
}

func TestAuditConfigValidateCapsTimelinePages(t *testing.T) {
	cfg := AuditConfig{
		DefaultRangeDays: 30,
		DefaultPageSize:  20,
		MaxPageSize:      100,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid audit config: %v", err)
	}

	cfg.DefaultPageSize = 101
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "CMS_AUDIT_DEFAULT_PAGE_SIZE") {
		t.Fatalf("expected audit page size validation error, got %v", err)
	}
}
