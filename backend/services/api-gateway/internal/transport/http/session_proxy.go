package httptransport

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	gatewayauth "ecommerce/api-gateway/internal/auth"
)

type SessionProxy struct {
	target *url.URL
	client *http.Client
	logger *slog.Logger
}

func NewSessionProxy(rawTarget string, client *http.Client, logger *slog.Logger) (*SessionProxy, error) {
	target, err := url.Parse(strings.TrimSpace(rawTarget))
	if err != nil || target.Scheme == "" || target.Host == "" {
		return nil, errors.New("valid session HTTP target is required")
	}
	if client == nil {
		client = http.DefaultClient
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &SessionProxy{target: target, client: client, logger: logger}, nil
}

func (p *SessionProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target := *p.target
	target.Path = singleJoiningSlash(p.target.Path, r.URL.Path)
	target.RawQuery = r.URL.RawQuery

	request, err := http.NewRequestWithContext(r.Context(), r.Method, target.String(), r.Body)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, "UPSTREAM_REQUEST_FAILED", "Session service request could not be created")
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
	}
	request.Host = p.target.Host

	response, err := p.client.Do(request)
	if err != nil {
		p.logger.ErrorContext(r.Context(), "session_proxy_failed", "error", err, "request_id", RequestIDFromContext(r.Context()))
		writeError(w, r, http.StatusBadGateway, "SESSION_UPSTREAM_UNAVAILABLE", "Session service is temporarily unavailable")
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
		p.logger.WarnContext(r.Context(), "session_proxy_response_copy_failed", "error", err, "request_id", RequestIDFromContext(r.Context()))
	}
}

func removeIdentityHeaders(header http.Header) {
	for _, name := range []string{"X-User-ID", "X-Authenticated-User-ID", "X-Auth-User-ID", "X-Actor-ID", "X-Admin-ID", "X-User-Roles", "X-User-Role", "X-Authenticated-Roles", "X-Auth-Roles", "X-Roles", "X-Actor-Roles"} {
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
