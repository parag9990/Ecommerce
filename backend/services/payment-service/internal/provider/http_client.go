package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const maxProviderResponseBytes int64 = 1 << 20

func providerEndpoint(configuredBaseURL string, fallbackBaseURL string, path string) (string, error) {
	baseURL := strings.TrimSpace(configuredBaseURL)
	if baseURL == "" {
		baseURL = fallbackBaseURL
	}
	base, err := url.Parse(baseURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return "", fmt.Errorf("%w: invalid provider base URL", ErrInvalidProviderRequest)
	}
	endpoint, err := base.Parse(path)
	if err != nil {
		return "", fmt.Errorf("%w: invalid provider endpoint", ErrInvalidProviderRequest)
	}
	return endpoint.String(), nil
}

func executeProviderRequest(ctx context.Context, client *http.Client, req *http.Request, providerName string, operation string) ([]byte, error) {
	res, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, NewError(
			providerName,
			operation,
			ErrorCodeUnavailable,
			"payment provider request failed",
			WithRetryable(true),
			WithStatusCode(http.StatusServiceUnavailable),
			WithCause(err),
		)
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
		return nil, providerHTTPError(providerName, operation, res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, maxProviderResponseBytes+1))
	if err != nil {
		return nil, NewError(providerName, operation, ErrorCodeUnavailable, "payment provider response could not be read", WithRetryable(true), WithCause(err))
	}
	if int64(len(body)) > maxProviderResponseBytes {
		return nil, NewError(providerName, operation, ErrorCodeUnavailable, "payment provider response exceeds size limit")
	}
	return body, nil
}

func providerHTTPError(providerName string, operation string, statusCode int) error {
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return NewError(providerName, operation, ErrorCodeAuthentication, "payment provider authentication failed", WithStatusCode(http.StatusBadGateway))
	case http.StatusTooManyRequests:
		return NewError(providerName, operation, ErrorCodeRateLimited, "payment provider rate limited the request", WithRetryable(true), WithStatusCode(http.StatusServiceUnavailable))
	default:
		if statusCode >= http.StatusInternalServerError {
			return NewError(providerName, operation, ErrorCodeUnavailable, "payment provider is unavailable", WithRetryable(true), WithStatusCode(http.StatusServiceUnavailable))
		}
		return NewError(providerName, operation, ErrorCodeValidation, "payment provider rejected the intent request", WithStatusCode(http.StatusBadGateway))
	}
}

func sanitizedAuditPayload(values map[string]any) (json.RawMessage, error) {
	payload, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("%w: create sanitized provider audit payload: %v", ErrInvalidProviderRequest, err)
	}
	if err := ValidateSanitizedJSONPayload(payload, false, "raw_provider_response"); err != nil {
		return nil, err
	}
	return json.RawMessage(payload), nil
}
