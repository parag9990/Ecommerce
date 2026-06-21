package repository

import (
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestHeatmapPointUpdateDoesNotWriteAFieldWithMultipleOperators(t *testing.T) {
	now := time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC)
	xBucket, yBucket := 50, 50
	retainUntil := now.AddDate(2, 0, 0)
	update := heatmapPointUpdate(domain.HeatmapPoint{
		ID:             "hm_test",
		HeatmapType:    domain.HeatmapTypeClick,
		Path:           "/products/test",
		NormalizedPath: "/products/:product_id",
		DeviceType:     domain.DeviceTypeMobile,
		ViewportBucket: "mobile_360_480",
		Day:            "2026-06-21",
		XBucket:        &xBucket,
		YBucket:        &yBucket,
		X:              50,
		Y:              50,
		SchemaVersion:  domain.CurrentHeatmapSchemaVersion,
		FirstSeenAt:    now,
		LastSeenAt:     now,
		CreatedAt:      now,
		UpdatedAt:      now,
		RetainUntil:    &retainUntil,
		RetentionClass: domain.RetentionClassAggregateFine,
	}, 1)

	seen := make(map[string]string)
	for _, operator := range update {
		fields, ok := operator.Value.(bson.D)
		if !ok {
			t.Fatalf("operator %s does not contain bson.D: %T", operator.Key, operator.Value)
		}
		for _, field := range fields {
			if previous, exists := seen[field.Key]; exists {
				t.Fatalf("field %q is written by both %s and %s", field.Key, previous, operator.Key)
			}
			seen[field.Key] = operator.Key
		}
	}
}
