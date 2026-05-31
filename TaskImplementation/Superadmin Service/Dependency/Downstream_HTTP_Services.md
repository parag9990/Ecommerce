# Superadmin Service - Downstream HTTP Services Dependency

## 1. What Is This Dependency?

Downstream HTTP services wo internal services hain jinko Superadmin Service call karti hai. Current code me User, Order, and Payment service ke admin APIs HTTP ke through call hote hain.

## 2. Why This Service Uses It

Superadmin Service doosre service ke database ko directly update nahi karti. Example: user status User Service own karta hai, refund Payment Service own karta hai, order status Order Service own karta hai. Isliye Superadmin Service permission check, reason, audit context handle karke downstream HTTP API call karti hai.

## 3. Required Or Optional

| Flow | Downstream Required? |
|------|----------------------|
| `/healthz`, `/readyz` | No downstream needed |
| RBAC permission catalog | No downstream, only MySQL |
| User list/status | User Service required |
| Seller list/status | User Service required |
| Order list/dispute/manual review | Order Service required, Payment Service useful for dispute/payment summary |
| Payment list/refund review | Payment Service required |
| Platform settings | No downstream HTTP call in current code; event is logged |
| Audit logs | No downstream, MySQL |

## 4. Where It Is Used In Project

| Path | Dependency |
|------|------------|
| `internal/clients/user_service_http_client.go` | User/Seller admin HTTP client |
| `internal/clients/order_payment_http_clients.go` | Order and Payment admin HTTP clients |
| `cmd/server/main.go` | Builds downstream clients |
| `internal/usecase/user_seller_controls.go` | Calls User Service |
| `internal/usecase/order_payment_controls.go` | Calls Order and Payment services |
| `.env` | Downstream base URLs and timeouts |

## 5. Installation Steps

No Go package install needed beyond current module. Downstream services must be running separately.

Expected local URLs from `.env`:

| Service | Base URL |
|---------|----------|
| User Service | `http://127.0.0.1:8081/internal/admin` |
| Order Service | `http://127.0.0.1:8084/internal/admin` |
| Payment Service | `http://127.0.0.1:8085/internal/admin` |

## 6. Docker Setup, If Possible

No docker-compose service definitions are clearly found in project files.

If Docker Compose is added later, service names should be used inside container network, for example:

```text
USER_SERVICE_ADMIN_BASE_URL=http://user-service:8081/internal/admin
ORDER_SERVICE_ADMIN_BASE_URL=http://order-service:8084/internal/admin
PAYMENT_SERVICE_ADMIN_BASE_URL=http://payment-service:8085/internal/admin
```

## 7. Local Setup Without Docker

Start User, Order, and Payment services locally on expected ports.

Exact downstream start commands are not clearly found in project files.

Suggested verification commands:

```bash
curl http://127.0.0.1:8081/healthz
curl http://127.0.0.1:8084/healthz
curl http://127.0.0.1:8085/healthz
```

## 8. Required Environment Variables

| Variable | Example | Required? |
|----------|---------|-----------|
| `USER_SERVICE_ADMIN_BASE_URL` | `http://127.0.0.1:8081/internal/admin` | Required for user/seller APIs |
| `USER_SERVICE_TIMEOUT` | `5s` | Optional |
| `SUPERADMIN_REQUIRE_USER_SERVICE` | `true` in prod | Recommended |
| `ORDER_SERVICE_ADMIN_BASE_URL` | `http://127.0.0.1:8084/internal/admin` | Required for order APIs |
| `ORDER_SERVICE_TIMEOUT` | `5s` | Optional |
| `SUPERADMIN_REQUIRE_ORDER_SERVICE` | `true` in prod | Recommended |
| `PAYMENT_SERVICE_ADMIN_BASE_URL` | `http://127.0.0.1:8085/internal/admin` | Required for payment APIs |
| `PAYMENT_SERVICE_TIMEOUT` | `5s` | Optional |
| `SUPERADMIN_REQUIRE_PAYMENT_SERVICE` | `true` in prod | Recommended |

## 9. Start Commands

Start Superadmin Service after downstream URLs are exported:

```bash
cd backend/services/superadmin-service
set -a
source .env
set +a
go run ./cmd/server
```

## 10. Verify Running Commands

Health:

```bash
curl http://127.0.0.1:8088/healthz
```

Example protected request shape:

```bash
curl http://127.0.0.1:8088/api/v1/admin/users \
  -H "X-Admin-Id: admin_1" \
  -H "X-User-Id: user_1" \
  -H "X-Admin-Roles: superadmin" \
  -H "X-Session-Id: sess_1" \
  -H "X-Request-Id: req_1"
```

Note: request will still fail if `admin_1` is not present and active in `admin_users` with required role permissions.

## 11. Common Errors And Fixes

| Error | Reason | Fix |
|-------|--------|-----|
| `user service is not configured` | User base URL empty | Set `USER_SERVICE_ADMIN_BASE_URL` |
| `order service is unavailable` | Order service down or wrong URL | Start service or fix `ORDER_SERVICE_ADMIN_BASE_URL` |
| `payment service returned invalid response` | Response schema mismatch | Align downstream admin API JSON |
| `admin actor is missing` | Required headers absent | Send admin headers from gateway/client |
| Forbidden before downstream call | RBAC denied in MySQL | Provision admin and permissions |
| Docs say gRPC but code uses HTTP | Implementation mismatch | Either document HTTP as current runtime or implement gRPC |

## 12. Security Notes

- These are internal admin APIs; do not expose them publicly.
- Always pass admin context headers from a trusted gateway.
- Use request IDs for high-risk actions.
- Use MFA header for critical operations like refund review and settings write.
- Prefer mTLS/service mesh or private network in production.
- Validate downstream response schemas strictly.

## 13. Final Checklist

- [x] User Service HTTP client exists.
- [x] Order Service HTTP client exists.
- [x] Payment Service HTTP client exists.
- [x] Admin context headers are forwarded.
- [x] Downstream URLs are configurable.
- [ ] Downstream service implementations were not fully verified here.
- [ ] gRPC/proto planned by docs is not implemented in current code.
