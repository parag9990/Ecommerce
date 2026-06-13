package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/config"
	"github.com/go-sql-driver/mysql"
)

const defaultPingTimeout = 5 * time.Second

func OpenMySQL(ctx context.Context, cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn, err := BuildMySQLDSN(cfg)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}

func BuildMySQLDSN(cfg config.DatabaseConfig) (string, error) {
	if dsn := strings.TrimSpace(cfg.DSN); dsn != "" {
		return dsn, nil
	}

	location, err := time.LoadLocation(strings.TrimSpace(cfg.Timezone))
	if err != nil {
		return "", fmt.Errorf("load mysql timezone: %w", err)
	}

	mysqlCfg := mysql.NewConfig()
	mysqlCfg.User = strings.TrimSpace(cfg.User)
	mysqlCfg.Passwd = cfg.Password
	mysqlCfg.Net = "tcp"
	mysqlCfg.Addr = net.JoinHostPort(strings.TrimSpace(cfg.Host), strconv.Itoa(cfg.Port))
	mysqlCfg.DBName = strings.TrimSpace(cfg.Name)
	mysqlCfg.ParseTime = true
	mysqlCfg.Loc = location
	if err := mysqlCfg.Apply(mysql.Charset("utf8mb4", "utf8mb4_unicode_ci")); err != nil {
		return "", fmt.Errorf("apply mysql charset: %w", err)
	}
	if strings.EqualFold(location.String(), "UTC") {
		mysqlCfg.Params = map[string]string{"time_zone": "'+00:00'"}
	}

	return mysqlCfg.FormatDSN(), nil
}
