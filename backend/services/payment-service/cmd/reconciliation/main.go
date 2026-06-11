package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/config"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/events"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider/settlement"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/repository"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := run(logger, os.Args[1:], time.Now().UTC()); err != nil {
		logger.Error("payment.reconciliation.failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(logger *slog.Logger, args []string, now time.Time) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	if !cfg.Reconciliation.Enabled {
		logger.Info("payment.reconciliation.disabled")
		return nil
	}

	flags := flag.NewFlagSet("reconciliation", flag.ContinueOnError)
	providerName := flags.String("provider", cfg.Reconciliation.Provider, "normalized provider name")
	reportFile := flags.String("report-file", cfg.Reconciliation.ReportFile, "path to the securely delivered settlement CSV")
	reportDateText := flags.String("report-date", "", "provider report date in YYYY-MM-DD")
	if err := flags.Parse(args); err != nil {
		return err
	}
	*providerName = strings.ToLower(strings.TrimSpace(*providerName))
	*reportFile = strings.TrimSpace(*reportFile)
	if *providerName == "" {
		return fmt.Errorf("provider is required")
	}
	if *reportFile == "" {
		return fmt.Errorf("report-file is required")
	}
	reportDate, err := selectedReportDate(*reportDateText, now, cfg.Reconciliation.ReportLag)
	if err != nil {
		return err
	}
	source, err := settlement.NewCSVFileSource(*reportFile, cfg.Reconciliation.MaxReportFileBytes)
	if err != nil {
		return err
	}

	db, err := repository.OpenMySQL(cfg.Database.DSN)
	if err != nil {
		return fmt.Errorf("open mysql: %w", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)
	paymentRepository, err := repository.NewMySQLPaymentRepository(db, logger)
	if err != nil {
		return fmt.Errorf("initialize repository: %w", err)
	}
	alertPublisher, err := events.NewHTTPPublisher(cfg.EventPublisher)
	if err != nil {
		return fmt.Errorf("initialize reconciliation alert publisher: %w", err)
	}
	reconcile, err := usecase.NewReconcileSettlementsUsecase(
		paymentRepository,
		source,
		alertPublisher,
		usecase.ReconcileSettlementsConfig{
			AlertTopic: cfg.Reconciliation.AlertTopic,
			BatchSize:  cfg.Reconciliation.BatchSize,
		},
		logger,
	)
	if err != nil {
		return fmt.Errorf("initialize reconciliation usecase: %w", err)
	}

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(signalCtx, cfg.Reconciliation.Timeout)
	defer cancel()
	startedAt := time.Now().UTC()
	summary, err := reconcile.Execute(ctx, *providerName, reportDate)
	if err != nil {
		return err
	}
	logger.InfoContext(ctx, "payment.reconciliation.completed",
		slog.String("provider", summary.Provider),
		slog.String("settlement_id", summary.SettlementID),
		slog.String("report_date", summary.ReportDate),
		slog.Int("matched_count", summary.MatchedCount),
		slog.Int("mismatch_count", summary.MismatchCount),
		slog.Int("missing_local_count", summary.MissingLocalCount),
		slog.Int("missing_provider_count", summary.MissingProviderCount),
		slog.Int("new_result_count", summary.CreatedCount),
		slog.Int("alerted_count", summary.AlertedCount),
		slog.Int("alert_failure_count", summary.AlertFailureCount),
		slog.Int64("duration_ms", time.Since(startedAt).Milliseconds()),
	)
	return nil
}

func selectedReportDate(raw string, now time.Time, reportLag time.Duration) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw != "" {
		reportDate, err := time.Parse(settlement.ReportDateLayout, raw)
		if err != nil {
			return time.Time{}, fmt.Errorf("report-date must be YYYY-MM-DD: %w", err)
		}
		return reportDate.UTC(), nil
	}
	eligible := now.UTC().Add(-reportLag)
	reportDate, err := time.Parse(settlement.ReportDateLayout, eligible.Format(settlement.ReportDateLayout))
	if err != nil {
		return time.Time{}, err
	}
	return reportDate.UTC(), nil
}
