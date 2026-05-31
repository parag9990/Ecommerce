# Superadmin Service - HTTP API Dependency

## 1. What Is This Dependency?

HTTP API wo runtime interface hai jisse admin client ya API Gateway Superadmin Service ko call karta hai. Current implementation Go `net/http` server use karti hai.

## 2. Why This Service Uses It

Current codebase me gRPC server nahi mila. Isliye Superadmin Service ka actual exposed interface HTTP routes hain. Admin headers se actor context pass hota hai, then usecase layer permission check karti hai.

## 3. Required Or Optional

Required. HTTP server ke bina service reachable nahi hogi.

## 4. Where It Is Used In Project

| Path | Usage |
|------|-------|
| `cmd/server/main.go` | Creates `http.Server` and listens on `HTTP_ADDR` |
| `internal/transport/http/routes.go` | Builds `http.ServeMux` |
| `internal/transport/http/rbac_handler.go` | `/healthz`, `/readyz`, RBAC routes |
| `internal/transport/http/control_handler.go` | User/seller admin routes |
| `internal/transport/http/order_payment_handler.go` | Order/payment/refund routes |
| `internal/transport/http/session_visibility_handler.go` | Session analytics authorization routes |
| `internal/transport/http/settings_handler.go` | Platform settings/search synonyms routes |
| `internal/transport/http/audit_log_handler.go` | Audit log route |
| `internal/transport/http/middleware.go` | Admin actor headers |

## 5. Installation Steps

No external HTTP framework install needed. Service uses Go standard library `net/http`.

Install Go and module dependencies:

```bash
cd backend/services/superadmin-service
go mod download
```

## 6. Docker Setup, If Possible

No service Dockerfile or docker-compose file was clearly found.

If containerized later, expose `HTTP_ADDR` port. Example:

```text
HTTP_ADDR=:8088
```

Host port mapping example:

```text
8088:8088
```

## 7. Local Setup Without Docker

```bash
cd backend/services/superadmin-service
set -a
source .env
set +a
go run ./cmd/server
```

## 8. Required Environment Variables

| Variable | Example | Required? |
|----------|---------|-----------|
| `HTTP_ADDR` | `:8088` | Yes |
| `HTTP_READ_HEADER_TIMEOUT` | `5s` | Optional but recommended |
| `HTTP_SHUTDOWN_TIMEOUT` | `10s` | Optional |
| `SERVICE_NAME` | `superadmin-service` | Yes |
| `LOG_LEVEL` | `debug` | Optional |

Admin request headers:

| Header | Required? | Purpose |
|--------|-----------|---------|
| `X-Admin-Id` | Yes | Admin actor id |
| `X-User-Id` or `X-Subject-Id` | Yes | User subject id |
| `X-Admin-Roles` or `X-Roles` | Yes | Role list |
| `X-Session-Id` | Yes | Admin session id |
| `X-Request-Id` | Required for high-risk actions | Trace/idempotency context |
| `X-IP-Hash` | Optional | PII-safe IP tracking |
| `X-MFA-Verified` | Required for some critical actions | `true` for refund/settings critical paths |

## 9. Start Commands

Exact command is not clearly found in project files.

Suggested command based on project structure:

```bash
cd backend/services/superadmin-service
go run ./cmd/server
```

## 10. Verify Running Commands

Health:

```bash
curl http://127.0.0.1:8088/healthz
```

Readiness:

```bash
curl http://127.0.0.1:8088/readyz
```

Example admin call:

```bash
curl http://127.0.0.1:8088/api/v1/admin/settings \
  -H "X-Admin-Id: admin_1" \
  -H "X-User-Id: user_1" \
  -H "X-Admin-Roles: superadmin" \
  -H "X-Session-Id: sess_1" \
  -H "X-Request-Id: req_1"
```

## 11. Common Errors And Fixes

| Error | Reason | Fix |
|-------|--------|-----|
| `connection refused` | Service not running or wrong port | Start service and verify `HTTP_ADDR` |
| `METHOD_NOT_ALLOWED` | Wrong HTTP method | Use route-specific method |
| `admin actor is missing` | Missing admin headers | Send required headers |
| `FORBIDDEN` | Admin lacks permission | Check `admin_users` and role permissions |
| `/readyz` returns `not_ready` | DB ping failed | Check MySQL and DSN |
| High-risk request fails | Reason/request ID/MFA missing | Add `reason`, `X-Request-Id`, and `X-MFA-Verified:true` where required |

## 12. Security Notes

- Put this service behind trusted API Gateway in production.
- Do not trust admin headers from public clients directly.
- Gateway should authenticate JWT/session and then set admin headers.
- Use TLS/mTLS between gateway and service.
- Rate limit high-risk mutation routes.
- Log request IDs, but avoid raw PII.

## 13. Final Checklist

- [x] HTTP server exists.
- [x] Health route exists.
- [x] Readiness route exists.
- [x] Admin routes are registered.
- [x] Admin headers are parsed.
- [x] Request validation exists.
- [ ] Gateway authentication/wiring is not clearly found in this service folder.
- [ ] gRPC server is not implemented.
