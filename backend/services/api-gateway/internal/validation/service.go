package validation

import (
	"bytes"
	"log/slog"
	"net/http"

	"ecommerce/api-gateway/internal/domain"
)

type RequestValidator interface {
	Validate(w http.ResponseWriter, r *http.Request, route domain.RouteDefinition) (*http.Request, *Error)
}

type Service struct {
	opts    Options
	schema  *SchemaValidator
	logger  *slog.Logger
	enabled bool
}

func NewService(schemas map[string]domain.Schema, opts Options, logger *slog.Logger) *Service {
	opts = NormalizeOptions(opts)
	return &Service{
		opts:    opts,
		schema:  NewSchemaValidator(schemas),
		logger:  logger,
		enabled: opts.Enabled,
	}
}

func (s *Service) Validate(w http.ResponseWriter, r *http.Request, route domain.RouteDefinition) (*http.Request, *Error) {
	if s == nil || !s.enabled {
		return r, nil
	}
	policy := PolicyForRoute(route, s.opts)
	if err := ValidateHeaders(r, s.opts.MaxHeaderBytes); err != nil {
		return r, err
	}
	if err := ValidatePathParams(r, route); err != nil {
		return r, err
	}
	queryPayload, err := ValidateQuery(r, route, s.schema, s.opts.MaxQueryBytes)
	if err != nil {
		return r, err
	}
	ctx := WithQueryPayload(r.Context(), queryPayload)
	r = r.WithContext(ctx)

	if !policy.BodyAllowed {
		if HasRequestBody(r) {
			return r, NewError(http.StatusBadRequest, FieldError{
				Field:   "body",
				Reason:  "not_allowed",
				Message: "Request body is not allowed for this route",
			})
		}
		return r, nil
	}

	if err := RequireContentType(r, policy.ContentTypes); err != nil {
		return r, err
	}
	rawBody, err := ReadLimitedBody(w, r, policy.MaxBodyBytes)
	if err != nil {
		return r, err
	}
	if policy.BodyRequired && len(bytes.TrimSpace(rawBody)) == 0 {
		return r, NewError(http.StatusBadRequest, FieldError{
			Field:   "body",
			Reason:  "required",
			Message: "Request body is required",
		})
	}
	r = r.WithContext(WithRawBody(r.Context(), rawBody))
	if policy.Webhook {
		return r, nil
	}

	payload, err := DecodeJSONObject(rawBody, policy.BodyRequired)
	if err != nil {
		return r, err
	}
	idempotencyBody := safeStringValue(payload["idempotency_key"])
	idempotency, err := ValidateIdempotencyKey(r, idempotencyBody, policy.RequireIDKey)
	if err != nil {
		return r, err
	}
	if idempotency != "" && idempotencyBody == "" {
		payload["idempotency_key"] = idempotency
	}
	if policy.ValidateSchema {
		if err := s.schema.Validate(policy.RequestSchema, payload); err != nil {
			return r, err
		}
	}
	r = r.WithContext(WithBodyPayload(r.Context(), payload))
	if idempotency != "" {
		r = r.WithContext(WithIdempotencyKey(r.Context(), idempotency))
	}
	return r, nil
}
