package clients

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
)

type HTTPSessionServiceClient struct {
	baseURL *url.URL
	token   string
	client  *http.Client
	logger  logging.Logger
}

func NewHTTPSessionServiceClient(baseURL, token string, timeout time.Duration, logger logging.Logger) (*HTTPSessionServiceClient, error) {
	parsed, err := parseServiceBaseURL(baseURL, "session")
	if err != nil {
		return nil, err
	}
	token = strings.TrimSpace(token)
	if len(token) < 32 {
		return nil, errors.New("session service admin token must be at least 32 characters")
	}
	return &HTTPSessionServiceClient{baseURL: parsed, token: token, client: &http.Client{Timeout: serviceTimeout(timeout)}, logger: loggerOrNop(logger)}, nil
}

func (c *HTTPSessionServiceClient) ForwardAnalytics(ctx context.Context, incoming *http.Request, actor domain.AdminActor, downstreamHeaders map[string]string) (*http.Response, error) {
	target := *c.baseURL
	target.Path = singleJoiningSlash(c.baseURL.Path, incoming.URL.Path)
	target.RawQuery = incoming.URL.RawQuery
	request, err := http.NewRequestWithContext(ctx, incoming.Method, target.String(), incoming.Body)
	if err != nil {
		return nil, domain.NewInternal("build session service request failed", err)
	}
	request.Header = incoming.Header.Clone()
	removeHopByHop(request.Header)
	removeForwardedIdentity(request.Header)
	attachAdminHeaders(request.Header, actor, "")
	request.Header.Set("Authorization", "Bearer "+c.token)
	for key, value := range downstreamHeaders {
		request.Header.Set(key, value)
	}
	response, err := c.client.Do(request)
	if err != nil {
		c.logger.Warn(ctx, "session service request failed", "request_id", actor.RequestID, "error", err)
		return nil, domain.NewDownstreamUnavailable("session service is unavailable", err)
	}
	return response, nil
}

func singleJoiningSlash(left, right string) string {
	return strings.TrimRight(left, "/") + "/" + strings.TrimLeft(right, "/")
}
func removeHopByHop(header http.Header) {
	for _, name := range []string{"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding", "Upgrade"} {
		header.Del(name)
	}
}
func removeForwardedIdentity(header http.Header) {
	for _, name := range []string{"X-Admin-ID", "X-Actor-ID", "X-User-ID", "X-Admin-Roles", "X-Actor-Role", "X-Roles", "X-User-Roles", "X-Admin-Mask-PII", "X-Admin-Risk-Allowed"} {
		header.Del(name)
	}
}

type UnavailableSessionServiceClient struct{ message string }

func NewUnavailableSessionServiceClient(message string) *UnavailableSessionServiceClient {
	if strings.TrimSpace(message) == "" {
		message = "session service is not configured"
	}
	return &UnavailableSessionServiceClient{message: message}
}
func (c *UnavailableSessionServiceClient) ForwardAnalytics(context.Context, *http.Request, domain.AdminActor, map[string]string) (*http.Response, error) {
	return nil, domain.NewDownstreamUnavailable(c.message, nil)
}
