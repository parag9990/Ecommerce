package events

import "testing"

func TestRabbitMQPublisherUsesTopicExchange(t *testing.T) {
	if rabbitMQExchangeType != "topic" {
		t.Fatalf("rabbitmq exchange type = %q, want topic", rabbitMQExchangeType)
	}
}

func TestRoutingKeyFromPayloadUsesEventType(t *testing.T) {
	payload := []byte(`{"event_id":"evt_1","event_type":"UserCreated"}`)

	if got := routingKeyFromPayload(payload); got != "UserCreated" {
		t.Fatalf("routing key = %q, want UserCreated", got)
	}
}

func TestRoutingKeyFromPayloadIgnoresInvalidPayload(t *testing.T) {
	if got := routingKeyFromPayload([]byte(`not-json`)); got != "" {
		t.Fatalf("routing key = %q, want empty", got)
	}
}
