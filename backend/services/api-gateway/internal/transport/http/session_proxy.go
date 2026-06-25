package httptransport

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	gatewayauth "ecommerce/api-gateway/internal/auth"
)

type SessionProxy struct {
	target             *url.URL
	client             *http.Client
	logger             *slog.Logger
	service            string
	internalAuthHeader string
	internalAuthToken  string
}

func NewSessionProxy(rawTarget string, client *http.Client, logger *slog.Logger) (*SessionProxy, error) {
	return NewServiceHTTPProxy("session", rawTarget, client, logger)
}

type ServiceHTTPProxyOption func(*SessionProxy)

func WithInternalAuth(header string, token string) ServiceHTTPProxyOption {
	return func(proxy *SessionProxy) {
		proxy.internalAuthHeader = strings.TrimSpace(header)
		proxy.internalAuthToken = strings.TrimSpace(token)
	}
}

func NewServiceHTTPProxy(service, rawTarget string, client *http.Client, logger *slog.Logger, options ...ServiceHTTPProxyOption) (*SessionProxy, error) {
	target, err := url.Parse(strings.TrimSpace(rawTarget))
	if err != nil || target.Scheme == "" || target.Host == "" {
		return nil, errors.New("valid service HTTP target is required")
	}
	service = strings.ToLower(strings.TrimSpace(service))
	if service == "" {
		service = "downstream"
	}
	if client == nil {
		client = http.DefaultClient
	}
	if logger == nil {
		logger = slog.Default()
	}
	proxy := &SessionProxy{target: target, client: client, logger: logger, service: service}
	for _, option := range options {
		option(proxy)
	}
	return proxy, nil
}

func (p *SessionProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target := *p.target
	target.Path = singleJoiningSlash(p.target.Path, r.URL.Path)
	target.RawQuery = r.URL.RawQuery

	request, err := http.NewRequestWithContext(r.Context(), r.Method, target.String(), r.Body)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, "UPSTREAM_REQUEST_FAILED", "Downstream service request could not be created")
		return
	}
	request.Header = r.Header.Clone()
	removeHopByHopHeaders(request.Header)
	removeIdentityHeaders(request.Header)
	if claims, ok := gatewayauth.ClaimsFromContext(r.Context()); ok {
		request.Header.Set("X-User-ID", claims.UserID())
		request.Header.Set("X-Actor-ID", claims.UserID())
		request.Header.Set("X-User-Roles", strings.Join(claims.Roles, ","))
		request.Header.Set("X-Roles", strings.Join(claims.Roles, ","))
		request.Header.Set("X-Seller-ID", claims.SellerID)
		request.Header.Set("X-Session-ID", claims.SessionID)
		request.Header.Set("X-Request-ID", RequestIDFromContext(r.Context()))
		if p.service == "superadmin" {
			request.Header.Set("X-Admin-ID", claims.UserID())
			request.Header.Set("X-Admin-Roles", strings.Join(claims.Roles, ","))
			request.Header.Set("X-MFA-Verified", strconv.FormatBool(claims.MFAVerified()))
		}
	}
	if p.internalAuthHeader != "" && p.internalAuthToken != "" {
		request.Header.Set(p.internalAuthHeader, p.internalAuthToken)
	}
	request.Host = p.target.Host

	response, err := p.client.Do(request)
	if err != nil {
		p.logger.ErrorContext(r.Context(), p.service+"_proxy_failed", "error", err, "request_id", RequestIDFromContext(r.Context()))
		writeError(w, r, http.StatusBadGateway, strings.ToUpper(p.service)+"_UPSTREAM_UNAVAILABLE", "Downstream service is temporarily unavailable")
		return
	}
	defer response.Body.Close()
	removeHopByHopHeaders(response.Header)
	for key, values := range response.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(response.StatusCode)
	if _, err := io.Copy(w, response.Body); err != nil {
		p.logger.WarnContext(r.Context(), p.service+"_proxy_response_copy_failed", "error", err, "request_id", RequestIDFromContext(r.Context()))
	}
}

func removeIdentityHeaders(header http.Header) {
	for _, name := range []string{"X-User-ID", "X-Authenticated-User-ID", "X-Auth-User-ID", "X-Actor-ID", "X-Admin-ID", "X-User-Roles", "X-User-Role", "X-Authenticated-Roles", "X-Auth-Roles", "X-Roles", "X-Actor-Roles", "X-Admin-Roles", "X-MFA-Verified", "X-Session-ID", "X-Seller-ID", "X-Tenant-ID", "X-Permissions", "X-Staff-Status"} {
		header.Del(name)
	}
}

func removeHopByHopHeaders(header http.Header) {
	for _, name := range []string{"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding", "Upgrade"} {
		header.Del(name)
	}
}

func singleJoiningSlash(a, b string) string {
	return strings.TrimRight(a, "/") + "/" + strings.TrimLeft(b, "/")
}
