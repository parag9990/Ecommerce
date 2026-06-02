# API Gateway Dependency - Seller Dashboard (CMS)

## 1. What is this dependency?

API Gateway ek backend edge service hai jo browser REST calls receive karta hai aur internal services ko gRPC calls forward karta hai. Simple words me: Seller Dashboard frontend directly Product/CMS/Order DB ko touch nahi karta; Gateway ke through request bhejta hai.

## 2. Why this service uses it

Seller Dashboard ko authenticated seller APIs chahiye:

- seller session
- product list/create/update/publish
- seller orders and fulfillment
- coupons/campaigns
- analytics summary
- team and audit APIs

Ye sab frontend se `http.ts` ke through API Gateway base URL par jaate hain.

## 3. Required or optional

Required. Frontend default API base `http://localhost:8080` hai.

## 4. Where it is used in project

| Path | Use |
|---|---|
| `frontend/seller-dashboard/src/lib/http.ts` | Gateway REST calls |
| `backend/services/api-gateway/.env` | Gateway port, gRPC targets, Redis rate limit, JWT/JWKS config |
| `api/master-api.json` | REST to gRPC route contract |
| `docs/02-system-architecture.md` | Gateway responsibility |

## 5. Installation steps

Backend source for API Gateway was not clearly found in project files, so exact install command is not available.

Suggested command based on project structure:

```bash
cd backend/services/api-gateway
go mod tidy
go run ./cmd/server
```

This will work only after `go.mod`, `cmd/server/main.go`, clients, routes, and config loader exist.

## 6. Docker setup, if possible

Not clearly found in project files.

No API Gateway Dockerfile or docker-compose service was found. Suggested compose dependency list:

- API Gateway service on `8080`
- Redis on `6379`
- downstream gRPC services
- JWKS/Auth service endpoint

## 7. Local setup without Docker

Current local setup is incomplete because API Gateway code is missing.

Suggested local flow:

1. Start Redis if `RATE_LIMIT_ENABLED=true`.
2. Start Auth/JWKS provider.
3. Start Product, Order, CMS, User services.
4. Start API Gateway on `:8080`.
5. Start Seller Dashboard frontend.

## 8. Required environment variables

| Variable | Required | Purpose |
|---|---|---|
| `HTTP_ADDR` | Yes | Gateway HTTP listen address |
| `API_BASE_PATH` | Yes | REST base path |
| `API_CONTRACT_PATH` | Yes | Path to `api/master-api.json` |
| `GRPC_TLS_ENABLED` | Yes | gRPC TLS toggle |
| `GRPC_DIAL_TIMEOUT` | Yes | gRPC dial timeout |
| `PRODUCT_GRPC_ADDR` | Yes | Product Service target |
| `ORDER_GRPC_ADDR` | Yes | Order Service target |
| `CMS_GRPC_ADDR` | Yes | CMS Service target |
| `USER_GRPC_ADDR` | Yes | User Service target |
| `RATE_LIMIT_ENABLED` | Yes | Gateway rate limit toggle |
| `REDIS_ADDR` | If rate limit enabled | Redis address |
| `JWT_JWKS_URL` | Yes | JWT verification key source |

## 9. Start commands

Suggested command based on project structure:

```bash
cd backend/services/api-gateway
go run ./cmd/server
```

Current status: Not clearly runnable because backend source was not found.

## 10. Verify running commands

```bash
curl http://localhost:8080/health/live
```

Health endpoint was not clearly found in project files. If not implemented, use any known public route after gateway exists.

## 11. Common errors and fixes

| Error | Reason | Fix |
|---|---|---|
| Frontend network error | Gateway not running | Start Gateway or update `VITE_API_BASE_URL` |
| 401 | JWT/session missing | Verify auth cookies/JWKS config |
| 403 | Seller role/status missing | Verify RBAC and seller status |
| 429 | Redis rate limit active | Tune route limits or Redis config |
| gRPC unavailable | Downstream service not running or wrong port | Align `*_GRPC_ADDR` with service env |
| CMS calls fail | Gateway CMS target `50059` differs from CMS service `9098` | Align CMS gRPC port |

## 12. Security notes

- Gateway must validate JWT, seller roles, seller ownership, request size, and content type.
- Redis rate limit should fail closed or fail open according to risk. Current env has `RATE_LIMIT_FAIL_OPEN=false`.
- Do not log JWTs, cookies, internal tokens, or raw passwords.
- Internal gRPC in production should use TLS/mTLS or service mesh controls.

## 13. Final checklist

- [x] Gateway env file found.
- [x] Frontend points to Gateway by default.
- [x] Master API contract found.
- [x] Redis rate limit config found.
- [ ] API Gateway Go source not clearly found.
- [ ] Gateway Dockerfile not found.
- [ ] Health route not clearly found.
- [ ] CMS gRPC port mismatch needs fix.

