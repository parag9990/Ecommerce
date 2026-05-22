package domain

import (
	"testing"
	"time"
)

func TestHeatmapPointFromClickEventNormalizesAndBuckets(t *testing.T) {
	path := "/products/prod_123"
	occurredAt := time.Date(2026, 5, 22, 10, 16, 10, 0, time.UTC)
	point, err := HeatmapPointFromEvent(SessionEvent{
		EventID:     "evt_click_1",
		SessionID:   "sess_123",
		AnonymousID: "anon_123",
		EventType:   EventClick,
		Path:        &path,
		Properties: map[string]any{
			"x":               210,
			"y":               600,
			"viewport_width":  390,
			"viewport_height": 844,
		},
		Device:     &Device{Type: DeviceTypeMobile},
		OccurredAt: occurredAt,
		ReceivedAt: occurredAt.Add(time.Second),
	}, occurredAt.Add(time.Minute), 5)
	if err != nil {
		t.Fatalf("build heatmap point: %v", err)
	}
	if point.HeatmapType != HeatmapTypeClick || point.X != 55 || point.Y != 70 {
		t.Fatalf("unexpected click point: %+v", point)
	}
	if point.ViewportBucket != "mobile_360_480" {
		t.Fatalf("unexpected viewport bucket %q", point.ViewportBucket)
	}
	if point.NormalizedPath != "/products/:product_id" {
		t.Fatalf("unexpected normalized path %q", point.NormalizedPath)
	}
	if point.ID == "" {
		t.Fatal("expected deterministic id")
	}
}

func TestHeatmapPointFromScrollEventBucketsDepth(t *testing.T) {
	path := "/products/prod_123"
	occurredAt := time.Date(2026, 5, 22, 10, 16, 45, 0, time.UTC)
	point, err := HeatmapPointFromEvent(SessionEvent{
		EventID:     "evt_scroll_1",
		SessionID:   "sess_123",
		AnonymousID: "anon_123",
		EventType:   EventScroll,
		Path:        &path,
		Properties: map[string]any{
			"depth_percent":   76,
			"viewport_height": 844,
			"document_height": 2200,
		},
		Device:     &Device{Type: DeviceTypeMobile},
		OccurredAt: occurredAt,
		ReceivedAt: occurredAt.Add(time.Second),
	}, occurredAt.Add(time.Minute), 5)
	if err != nil {
		t.Fatalf("build heatmap point: %v", err)
	}
	if point.HeatmapType != HeatmapTypeScroll || point.X != 50 || point.Y != 75 {
		t.Fatalf("unexpected scroll point: %+v", point)
	}
	if point.DepthBucket == nil || *point.DepthBucket != 75 {
		t.Fatalf("unexpected depth bucket: %+v", point.DepthBucket)
	}
	if point.XBucket != nil || point.YBucket != nil {
		t.Fatalf("scroll point should not include click buckets: %+v", point)
	}
}

func TestHeatmapPointRejectsInvalidClickCoordinates(t *testing.T) {
	path := "/products/prod_123"
	_, err := HeatmapPointFromEvent(SessionEvent{
		EventID:     "evt_click_bad",
		SessionID:   "sess_123",
		AnonymousID: "anon_123",
		EventType:   EventClick,
		Path:        &path,
		Properties: map[string]any{
			"x":               999,
			"y":               10,
			"viewport_width":  390,
			"viewport_height": 844,
		},
		OccurredAt: time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC),
	}, time.Date(2026, 5, 22, 10, 1, 0, 0, time.UTC), 5)
	if err == nil {
		t.Fatal("expected invalid coordinate error")
	}
}

func TestBucketScrollDepth(t *testing.T) {
	tests := []struct {
		depth float64
		want  int
	}{
		{depth: 0, want: 0},
		{depth: 27, want: 25},
		{depth: 55, want: 50},
		{depth: 76, want: 75},
		{depth: 92, want: 90},
		{depth: 100, want: 100},
	}
	for _, tt := range tests {
		if got := BucketScrollDepth(tt.depth); got != tt.want {
			t.Fatalf("BucketScrollDepth(%v) = %d, want %d", tt.depth, got, tt.want)
		}
	}
}
