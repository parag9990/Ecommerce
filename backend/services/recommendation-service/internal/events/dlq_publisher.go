package events

import (
	"context"
	"encoding/json"
	"time"
)

type DeadLetterPublisher interface {
	Publish(ctx context.Context, message DeadLetterMessage) error
}

type DeadLetterMessage struct {
	FailedAt         time.Time       `json:"failed_at"`
	Reason           string          `json:"reason"`
	Consumer         string          `json:"consumer"`
	OriginalTopic    string          `json:"original_topic"`
	EventID          string          `json:"event_id,omitempty"`
	EventType        string          `json:"event_type,omitempty"`
	Producer         string          `json:"producer,omitempty"`
	TraceID          string          `json:"trace_id,omitempty"`
	Attempts         int             `json:"attempts"`
	OriginalEvent    json.RawMessage `json:"original_event,omitempty"`
	OriginalEventRaw string          `json:"original_event_raw,omitempty"`
}
