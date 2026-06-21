package tracing

import (
	"context"
	"net/http"
	"testing"
)

func TestSetupWithoutExporterAndHTTPPropagation(t *testing.T) {
	shutdown, err := Setup(context.Background(), Config{ServiceName: "test-service", Environment: "test", SampleRatio: 1})
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	if shutdown == nil {
		t.Fatal("Setup() returned nil shutdown")
	}
	header := make(http.Header)
	InjectHTTP(context.Background(), header)
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown error = %v", err)
	}
}
