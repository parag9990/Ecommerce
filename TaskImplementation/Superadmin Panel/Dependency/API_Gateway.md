# API Gateway / REST Dependency - Superadmin Panel

## 1. What is this dependency?

API Gateway public REST entry point hai. Browser direct gRPC nahi call karta; Superadmin Panel `VITE_API_BASE_URL` ke through REST endpoints call karta hai.

## 2. Why this service uses it

Superadmin Panel ko users, sellers, orders, payments, sessions, settings, and audit data chahiye. Ye data API Gateway ke `/api/v1/...` endpoints se aata hai. Gateway auth, RBAC, validation, rate limiting, and REST to gRPC mapping handle karta hai.

## 3. Required or optional

Required for real frontend data. Without API Gateway, UI shell open ho sakta hai but admin data load nahi hoga.

## 4. Where it is used in project

| Path | Use |
|------|-----|
| `frontend/superadmin-panel/src/lib/http.ts` | Builds API URL and attaches auth headers |
| `frontend/superadmin-panel/src/features/*/api/*.ts` | Feature-specific API calls |
| `frontend/superadmin-panel/.env.example` | `VITE_API_BASE_URL=http://localhost:8080/api/v1` |
| `backend/services/api-gateway/.env` | Gateway env names and ports |
| `api/master-api.json` | REST route to service/gRPC contract |

## 5. Admin endpoints detected

| Method | Path | Service |
|--------|------|---------|
| `POST` | `/auth/login` | Auth flow used by frontend login |
| `GET` | `/api/v1/admin/users` | `superadmin-service` |
| `PATCH` | `/api/v1/admin/users/{user_id}/status` | `superadmin-service` |
| `GET` | `/api/v1/admin/sellers` | `superadmin-service` |
| `PATCH` | `/api/v1/admin/sellers/{seller_id}/status` | `superadmin-service` |
| `GET` | `/api/v1/admin/orders` | `order-service` |
| `GET` | `/api/v1/admin/payments` | `payment-service` |
| `POST` | `/api/v1/admin/refunds/{refund_id}/review` | `superadmin-service` |
| `GET` | `/api/v1/admin/audit-logs` | `superadmin-service` |
| `GET` | `/api/v1/admin/settings` | `superadmin-service` |
| `PATCH` | `/api/v1/admin/settings/{key}` | `superadmin-service` |
| `GET` | `/api/v1/analytics/live` | `session-service` |
| `GET` | `/api/v1/analytics/sessions` | `session-service` |
| `GET` | `/api/v1/analytics/sessions/{session_id}/journey` | `session-service` |

## 6. Installation steps

API Gateway source code was not clearly found. Only `.env` and API contract were found.

Suggested command based on project structure:

```bash
cd backend/services/api-gateway
go mod tidy
go run ./cmd/api-gateway
```

This is suggested only. It requires actual gateway source files.

## 7. Docker setup, if possible

No docker-compose service found.

Suggested command based on common setup:

```bash
docker compose up api-gateway redis
```

Mark as suggested only because compose file is missing.

## 8. Local setup without Docker

Required pieces:

1. API Gateway binary/source.
2. `.env` loaded from `backend/services/api-gateway/.env`.
3. Redis running if rate limit is enabled.
4. gRPC target services reachable.
5. JWT/JWKS config valid.

Suggested command based on project structure:

```bash
cd backend/services/api-gateway
go run ./cmd/api-gateway
```

## 9. Required environment variables

| Variable | Purpose |
|----------|---------|
| `HTTP_ADDR` | Gateway listen address, found `:8080` |
| `API_BASE_PATH` | Base path, found `/api/v1` |
| `API_CONTRACT_PATH` | Contract file path |
| `SUPERADMIN_GRPC_ADDR` | Superadmin backend gRPC address |
| `SESSION_GRPC_ADDR` | Session backend gRPC address |
| `REDIS_ADDR` | Redis address for rate limiting |
| `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_JWKS_URL` | Auth validation |
| `REQUEST_VALIDATION_ENABLED` | Request validation toggle |

## 10. Start commands

Not clearly found in project files.

Suggested command based on project structure:

```bash
cd backend/services/api-gateway
go run ./cmd/api-gateway
```

## 11. Verify running commands

Suggested command based on env/config:

```bash
curl http://localhost:8080/api/v1/admin/users
```

Expected without token: `401 Unauthorized` or similar auth error. Agar connection refused aaye, gateway running nahi hai.

## 12. Common errors and fixes

| Error | Reason | Fix |
|-------|--------|-----|
| Frontend gets network error | Gateway not running on `:8080` | Start gateway or change `VITE_API_BASE_URL` |
| 401 | Missing/invalid JWT | Login again and check token |
| 403 | Admin role missing | Use account with valid admin role |
| 502/Unavailable | Gateway cannot reach gRPC service | Check `SUPERADMIN_GRPC_ADDR` and service startup |
| Rate limit errors | Redis/rate limit config active | Check Redis and rate limit env |

## 13. Security notes

- Gateway must verify JWT on every admin route.
- Frontend role hiding is not enough; route-level RBAC must run in gateway/backend.
- Rate limit admin mutations separately.
- Keep request IDs and audit IDs in logs.
- Do not expose stack traces or internal gRPC errors to browser.

## 14. Final checklist

- [x] Frontend uses API Gateway base URL.
- [x] REST contract found in `api/master-api.json`.
- [x] Gateway env file found.
- [ ] Gateway source code clearly found.
- [ ] Gateway routes/gRPC clients verified in code.

