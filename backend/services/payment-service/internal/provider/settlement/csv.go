package settlement

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const DefaultMaxReportBytes int64 = 64 << 20

var requiredCSVColumns = []string{
	"settlement_id",
	"provider_payment_id",
	"outcome",
	"currency",
	"settled_amount",
	"fee_amount",
	"settled_at",
}

type CSVFileSource struct {
	path     string
	maxBytes int64
}

func NewCSVFileSource(path string, maxBytes int64) (*CSVFileSource, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("settlement report file path is required")
	}
	if maxBytes <= 0 {
		maxBytes = DefaultMaxReportBytes
	}
	return &CSVFileSource{path: path, maxBytes: maxBytes}, nil
}

func (s *CSVFileSource) Fetch(ctx context.Context, provider string, reportDate time.Time) (Report, error) {
	if err := ctx.Err(); err != nil {
		return Report{}, err
	}
	file, err := os.Open(s.path)
	if err != nil {
		return Report{}, fmt.Errorf("open settlement report: %w", err)
	}
	defer file.Close()

	limited := &io.LimitedReader{R: file, N: s.maxBytes + 1}
	report, err := ParseCSV(ctx, limited, provider, reportDate)
	if err != nil {
		return Report{}, err
	}
	if limited.N == 0 {
		return Report{}, fmt.Errorf("settlement report exceeds %d bytes", s.maxBytes)
	}
	return report, nil
}

func ParseCSV(ctx context.Context, source io.Reader, provider string, reportDate time.Time) (Report, error) {
	if err := ctx.Err(); err != nil {
		return Report{}, err
	}
	if source == nil {
		return Report{}, errors.New("settlement CSV source is required")
	}
	reader := csv.NewReader(contextReader{ctx: ctx, reader: source})
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return Report{}, fmt.Errorf("read settlement report header: %w", err)
	}
	columns, err := requiredColumns(header)
	if err != nil {
		return Report{}, err
	}

	rows := make([]Row, 0)
	for line := 2; ; line++ {
		values, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return Report{}, fmt.Errorf("read settlement report line %d: %w", line, readErr)
		}
		row, err := normalizeCSVRow(values, columns)
		if err != nil {
			return Report{}, fmt.Errorf("invalid settlement report line %d: %w", line, err)
		}
		rows = append(rows, row)
	}
	return NewReport(provider, reportDate, rows)
}

func requiredColumns(header []string) (map[string]int, error) {
	index := make(map[string]int, len(header))
	for position, value := range header {
		name := strings.ToLower(strings.TrimSpace(value))
		if name == "" {
			return nil, errors.New("settlement report contains a blank column name")
		}
		if _, exists := index[name]; exists {
			return nil, fmt.Errorf("settlement report contains duplicate column %q", name)
		}
		index[name] = position
	}
	for _, required := range requiredCSVColumns {
		if _, exists := index[required]; !exists {
			return nil, fmt.Errorf("settlement report is missing required column %q", required)
		}
	}
	return index, nil
}

func normalizeCSVRow(values []string, columns map[string]int) (Row, error) {
	value := func(name string) (string, error) {
		position := columns[name]
		if position >= len(values) {
			return "", fmt.Errorf("column %q has no value", name)
		}
		return strings.TrimSpace(values[position]), nil
	}
	settlementID, err := value("settlement_id")
	if err != nil {
		return Row{}, err
	}
	providerPaymentID, err := value("provider_payment_id")
	if err != nil {
		return Row{}, err
	}
	rawOutcome, err := value("outcome")
	if err != nil {
		return Row{}, err
	}
	outcome, err := NormalizeOutcome(rawOutcome)
	if err != nil {
		return Row{}, err
	}
	currency, err := value("currency")
	if err != nil {
		return Row{}, err
	}
	settledAmount, err := parseMinorAmount(values, columns, "settled_amount")
	if err != nil {
		return Row{}, err
	}
	feeAmount, err := parseMinorAmount(values, columns, "fee_amount")
	if err != nil {
		return Row{}, err
	}
	rawSettledAt, err := value("settled_at")
	if err != nil {
		return Row{}, err
	}
	settledAt, err := time.Parse(time.RFC3339, rawSettledAt)
	if err != nil {
		return Row{}, fmt.Errorf("settled_at must be RFC3339: %w", err)
	}
	return Row{
		SettlementID:      settlementID,
		ProviderPaymentID: providerPaymentID,
		Outcome:           outcome,
		Currency:          currency,
		SettledAmount:     settledAmount,
		FeeAmount:         feeAmount,
		SettledAt:         settledAt,
	}, nil
}

func parseMinorAmount(values []string, columns map[string]int, name string) (int64, error) {
	position := columns[name]
	if position >= len(values) {
		return 0, fmt.Errorf("column %q has no value", name)
	}
	value, err := strconv.ParseInt(strings.TrimSpace(values[position]), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer minor-unit amount: %w", name, err)
	}
	return value, nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(data []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(data)
}
