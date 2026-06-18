package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
)

const defaultUserServiceAdminPath = "/internal/admin"

type HTTPUserServiceClient struct {
	baseURL *url.URL
	client  *http.Client
	logger  logging.Logger
}

func NewHTTPUserServiceClient(baseURL string, timeout time.Duration, logger logging.Logger) (*HTTPUserServiceClient, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil, errors.New("user service base url is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse user service base url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("user service base url must use http or https")
	}
	if parsed.Host == "" {
		return nil, errors.New("user service base url must include a host")
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if logger == nil {
		logger = logging.NewNop()
	}
	return &HTTPUserServiceClient{
		baseURL: parsed,
		client:  &http.Client{Timeout: timeout},
		logger:  logger,
	}, nil
}

func (c *HTTPUserServiceClient) ListUsersForAdmin(ctx context.Context, req domain.AdminUserListRequest, actor domain.AdminActor) (domain.AdminUserListResponse, error) {
	var out domain.AdminUserListResponse
	values := listQuery(req.Query, string(req.Status), req.Pagination)
	if err := c.do(ctx, http.MethodGet, c.endpoint("/users", values), nil, actor, "", "user", "", &out); err != nil {
		return domain.AdminUserListResponse{}, err
	}
	return out, nil
}

func (c *HTTPUserServiceClient) GetUserForAdmin(ctx context.Context, userID string, actor domain.AdminActor) (domain.UserProfile, error) {
	var out domain.UserProfile
	path := "/users/" + url.PathEscape(userID)
	if err := c.do(ctx, http.MethodGet, c.endpoint(path, nil), nil, actor, "", "user", userID, &out); err != nil {
		return domain.UserProfile{}, err
	}
	return out, nil
}

func (c *HTTPUserServiceClient) UpdateUserStatus(ctx context.Context, userID string, status domain.UserStatus, mutation domain.AdminMutationContext) error {
	body := domain.StatusUpdateRequest{Status: string(status), Reason: mutation.Reason}
	path := "/users/" + url.PathEscape(userID) + "/status"
	return c.do(ctx, http.MethodPatch, c.endpoint(path, nil), body, mutation.Actor, mutation.Reason, "user", userID, nil)
}

func (c *HTTPUserServiceClient) ListSellersForAdmin(ctx context.Context, req domain.AdminSellerListRequest, actor domain.AdminActor) (domain.AdminSellerListResponse, error) {
	var out domain.AdminSellerListResponse
	values := listQuery(req.Query, string(req.Status), req.Pagination)
	if err := c.do(ctx, http.MethodGet, c.endpoint("/sellers", values), nil, actor, "", "seller", "", &out); err != nil {
		return domain.AdminSellerListResponse{}, err
	}
	return out, nil
}

func (c *HTTPUserServiceClient) GetSellerForAdmin(ctx context.Context, sellerID string, actor domain.AdminActor) (domain.SellerProfile, error) {
	var out domain.SellerProfile
	path := "/sellers/" + url.PathEscape(sellerID)
	if err := c.do(ctx, http.MethodGet, c.endpoint(path, nil), nil, actor, "", "seller", sellerID, &out); err != nil {
		return domain.SellerProfile{}, err
	}
	return out, nil
}

func (c *HTTPUserServiceClient) UpdateSellerStatus(ctx context.Context, sellerID string, status domain.SellerStatus, mutation domain.AdminMutationContext) error {
	body := domain.StatusUpdateRequest{Status: string(status), Reason: mutation.Reason}
	path := "/sellers/" + url.PathEscape(sellerID) + "/status"
	return c.do(ctx, http.MethodPatch, c.endpoint(path, nil), body, mutation.Actor, mutation.Reason, "seller", sellerID, nil)
}

func (c *HTTPUserServiceClient) endpoint(path string, values url.Values) string {
	u := *c.baseURL
	basePath := strings.TrimRight(u.Path, "/")
	if basePath == "" {
		basePath = defaultUserServiceAdminPath
	}
	u.Path = strings.TrimRight(basePath, "/") + "/" + strings.TrimLeft(path, "/")
	u.RawQuery = values.Encode()
	return u.String()
}

func (c *HTTPUserServiceClient) do(ctx context.Context, method string, endpoint string, body any, actor domain.AdminActor, reason string, resourceType string, resourceID string, out any) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return domain.NewInternal("marshal user service request failed", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return domain.NewInternal("build user service request failed", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	attachAdminHeaders(req.Header, actor, reason)

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Warn(ctx, "user service request failed", "method", method, "endpoint", endpoint, "request_id", actor.RequestID, "error", err)
		return domain.NewDownstreamUnavailable("user service is unavailable", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if out == nil {
			io.Copy(io.Discard, resp.Body)
			return nil
		}
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return domain.NewDownstreamUnavailable("user service returned an invalid response", err)
		}
		return nil
	}

	return mapUserServiceError(resp, resourceType, resourceID)
}

func mapUserServiceError(resp *http.Response, resourceType string, resourceID string) error {
	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	message := strings.TrimSpace(string(responseBody))

	switch resp.StatusCode {
	case http.StatusNotFound:
		switch resourceType {
		case "seller":
			return domain.NewSellerNotFound(resourceID)
		default:
			return domain.NewUserNotFound(resourceID)
		}
	case http.StatusBadRequest:
		if message == "" {
			message = "user service rejected the request"
		}
		return domain.NewValidationError(message)
	case http.StatusUnauthorized, http.StatusForbidden:
		return domain.NewDownstreamUnavailable("user service denied the admin request", nil)
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return domain.NewDownstreamUnavailable("user service is unavailable", nil)
	default:
		if resp.StatusCode >= 500 {
			return domain.NewDownstreamUnavailable("user service is unavailable", nil)
		}
		return domain.NewInternal(fmt.Sprintf("user service returned unexpected status %d", resp.StatusCode), nil)
	}
}

func listQuery(query string, status string, pagination domain.Pagination) url.Values {
	values := url.Values{}
	if strings.TrimSpace(query) != "" {
		values.Set("q", strings.TrimSpace(query))
	}
	if strings.TrimSpace(status) != "" {
		values.Set("status", strings.TrimSpace(status))
	}
	if pagination.Page > 0 {
		values.Set("page", strconv.Itoa(pagination.Page))
	}
	if pagination.PageSize > 0 {
		values.Set("page_size", strconv.Itoa(pagination.PageSize))
	}
	if strings.TrimSpace(pagination.Cursor) != "" {
		values.Set("cursor", strings.TrimSpace(pagination.Cursor))
	}
	return values
}

func attachAdminHeaders(header http.Header, actor domain.AdminActor, reason string) {
	header.Set("X-Admin-Id", actor.AdminID)
	header.Set("X-User-Id", actor.UserID)
	header.Set("X-Admin-Roles", strings.Join(domain.RolesToStrings(actor.Roles), ","))
	header.Set("X-Session-Id", actor.SessionID)
	header.Set("X-Request-Id", actor.RequestID)
	if actor.IPHash != "" {
		header.Set("X-IP-Hash", actor.IPHash)
	}
	if reason != "" {
		header.Set("X-Admin-Reason", reason)
	}
}

type UnavailableUserServiceClient struct {
	message string
}

func NewUnavailableUserServiceClient(message string) *UnavailableUserServiceClient {
	if strings.TrimSpace(message) == "" {
		message = "user service is not configured"
	}
	return &UnavailableUserServiceClient{message: message}
}

func (c *UnavailableUserServiceClient) ListUsersForAdmin(ctx context.Context, req domain.AdminUserListRequest, actor domain.AdminActor) (domain.AdminUserListResponse, error) {
	return domain.AdminUserListResponse{}, domain.NewDownstreamUnavailable(c.message, nil)
}

func (c *UnavailableUserServiceClient) GetUserForAdmin(ctx context.Context, userID string, actor domain.AdminActor) (domain.UserProfile, error) {
	return domain.UserProfile{}, domain.NewDownstreamUnavailable(c.message, nil)
}

func (c *UnavailableUserServiceClient) UpdateUserStatus(ctx context.Context, userID string, status domain.UserStatus, mutation domain.AdminMutationContext) error {
	return domain.NewDownstreamUnavailable(c.message, nil)
}

func (c *UnavailableUserServiceClient) ListSellersForAdmin(ctx context.Context, req domain.AdminSellerListRequest, actor domain.AdminActor) (domain.AdminSellerListResponse, error) {
	return domain.AdminSellerListResponse{}, domain.NewDownstreamUnavailable(c.message, nil)
}

func (c *UnavailableUserServiceClient) GetSellerForAdmin(ctx context.Context, sellerID string, actor domain.AdminActor) (domain.SellerProfile, error) {
	return domain.SellerProfile{}, domain.NewDownstreamUnavailable(c.message, nil)
}

func (c *UnavailableUserServiceClient) UpdateSellerStatus(ctx context.Context, sellerID string, status domain.SellerStatus, mutation domain.AdminMutationContext) error {
	return domain.NewDownstreamUnavailable(c.message, nil)
}
