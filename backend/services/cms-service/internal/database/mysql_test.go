package database

import (
	"strings"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/config"
	"github.com/go-sql-driver/mysql"
)

func TestBuildMySQLDSNFromComponents(t *testing.T) {
	dsn, err := BuildMySQLDSN(config.DatabaseConfig{
		Host:     "127.0.0.1",
		Port:     3306,
		Name:     "cms_db",
		User:     "cms_user",
		Password: "secret",
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("BuildMySQLDSN: %v", err)
	}

	parsed, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("ParseDSN: %v", err)
	}
	if parsed.User != "cms_user" || parsed.Passwd != "secret" || parsed.DBName != "cms_db" {
		t.Fatalf("unexpected dsn identity fields: %+v", parsed)
	}
	if parsed.Addr != "127.0.0.1:3306" {
		t.Fatalf("unexpected address %q", parsed.Addr)
	}
	if !parsed.ParseTime {
		t.Fatal("expected parseTime=true")
	}
	if parsed.Loc == nil || parsed.Loc.String() != "UTC" {
		t.Fatalf("expected UTC location, got %v", parsed.Loc)
	}
	if parsed.Collation != "utf8mb4_unicode_ci" {
		t.Fatalf("expected utf8mb4_unicode_ci collation, got %q", parsed.Collation)
	}
	if !strings.Contains(dsn, "charset=utf8mb4") {
		t.Fatalf("expected utf8mb4 charset in dsn, got %q", dsn)
	}
	if parsed.Params["time_zone"] != "'+00:00'" {
		t.Fatalf("expected UTC mysql session time zone, got %q", parsed.Params["time_zone"])
	}
}

func TestBuildMySQLDSNReturnsExplicitDSN(t *testing.T) {
	explicit := "cms_user:secret@tcp(db:3306)/cms_db?parseTime=true"

	dsn, err := BuildMySQLDSN(config.DatabaseConfig{DSN: " " + explicit + " "})
	if err != nil {
		t.Fatalf("BuildMySQLDSN: %v", err)
	}
	if dsn != explicit {
		t.Fatalf("expected explicit DSN, got %q", dsn)
	}
}

func TestBuildMySQLDSNRejectsInvalidTimezone(t *testing.T) {
	_, err := BuildMySQLDSN(config.DatabaseConfig{
		Host:     "127.0.0.1",
		Port:     3306,
		Name:     "cms_db",
		User:     "cms_user",
		Timezone: "Not/AZone",
	})
	if err == nil || !strings.Contains(err.Error(), "timezone") {
		t.Fatalf("expected timezone error, got %v", err)
	}
}
