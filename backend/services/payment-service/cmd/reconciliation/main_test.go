package main

import (
	"testing"
	"time"
)

func TestSelectedReportDateUsesConfiguredGracePeriod(t *testing.T) {
	now := time.Date(2026, time.May, 27, 1, 0, 0, 0, time.UTC)

	date, err := selectedReportDate("", now, 24*time.Hour)
	if err != nil {
		t.Fatalf("selectedReportDate() error = %v", err)
	}
	if got := date.Format("2006-01-02"); got != "2026-05-26" {
		t.Fatalf("report date = %q, want 2026-05-26", got)
	}
}

func TestSelectedReportDateRejectsInvalidExplicitDate(t *testing.T) {
	if _, err := selectedReportDate("05/26/2026", time.Now().UTC(), 24*time.Hour); err == nil {
		t.Fatal("selectedReportDate() error = nil, want invalid explicit date rejected")
	}
}
