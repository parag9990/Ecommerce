package validation

import "context"

type contextKey string

const (
	bodyPayloadKey  contextKey = "validated_body_payload"
	queryPayloadKey contextKey = "validated_query_payload"
	rawBodyKey      contextKey = "validated_raw_body"
	idempotencyKey  contextKey = "validated_idempotency_key"
)

func WithBodyPayload(ctx context.Context, payload map[string]any) context.Context {
	return context.WithValue(ctx, bodyPayloadKey, payload)
}

func BodyPayloadFromContext(ctx context.Context) (map[string]any, bool) {
	payload, ok := ctx.Value(bodyPayloadKey).(map[string]any)
	return payload, ok
}

func WithQueryPayload(ctx context.Context, payload map[string]any) context.Context {
	return context.WithValue(ctx, queryPayloadKey, payload)
}

func QueryPayloadFromContext(ctx context.Context) (map[string]any, bool) {
	payload, ok := ctx.Value(queryPayloadKey).(map[string]any)
	return payload, ok
}

func WithRawBody(ctx context.Context, body []byte) context.Context {
	return context.WithValue(ctx, rawBodyKey, append([]byte(nil), body...))
}

func RawBodyFromContext(ctx context.Context) ([]byte, bool) {
	body, ok := ctx.Value(rawBodyKey).([]byte)
	if !ok {
		return nil, false
	}
	return append([]byte(nil), body...), true
}

func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idempotencyKey, key)
}

func IdempotencyKeyFromContext(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(idempotencyKey).(string)
	return key, ok
}
