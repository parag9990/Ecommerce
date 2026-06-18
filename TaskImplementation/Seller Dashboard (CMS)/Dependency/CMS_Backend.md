# CMS Backend Dependency - Seller Dashboard (CMS)

## 1. What is this dependency?

CMS Backend seller dashboard ka business backend hai. Ye coupons, campaigns, seller settings, seller analytics, seller staff permissions, aur audit logs jaise CMS concerns own karta hai.

## 2. Why this service uses it

Seller Dashboard frontend offers, analytics, team, aur audit modules ke liye CMS backend par depend karta hai. Product aur order flows separate Product/Order services se aate hain, lekin CMS dashboard summary and CMS-owned data CMS service ka responsibility hai.

## 3. Required or optional

Required for full Seller Dashboard. Without CMS Backend:

- coupons/campaigns work nahi karenge
- analytics summary unavailable rahegi
- seller staff/team backend missing rahega
- audit activity API missing rahegi

## 4. Where it is used in project

| Path | Use |
|---|---|
| `backend/services/cms-service/.env` | CMS service config |
| `docs/04-microservice-design.md` | CMS service design |
| `docs/09-cms-superadmin.md` | Seller CMS modules |
| `database/draw.sql` | CMS MySQL schema |
| `frontend/seller-dashboard/src/features/offers/api/seller-offers-api.ts` | Coupon/campaign API client |
| `frontend/seller-dashboard/src/features/analytics/api/seller-analytics-api.ts` | Analytics API client |
| `frontend/seller-dashboard/src/features/team/api/seller-team-api.ts` | Team API client |
| `frontend/seller-dashboard/src/features/audit/api/seller-audit-api.ts` | Audit API client |

## 5. Installation steps

Actual CMS Go module was not clearly found. Only `.env` exists under `backend/services/cms-service/`.

Suggested command based on project structure:

```bash
cd backend/services/cms-service
go mod tidy
go run ./cmd/server
```

This requires backend source files first:

- `go.mod`
- `cmd/server/main.go`
- `internal/config`
- `internal/domain`
- `internal/usecase`
- `internal/repository`
- `internal/transport/grpc`
- migrations

## 6. Docker setup, if possible

Not clearly found in project files.

Suggested Docker requirement later:

- Expose CMS HTTP port `8087`.
- Expose CMS gRPC port `9098` or whichever port Gateway uses after alignment.
- Inject env through `.env` or secret manager.
- Do not bake secrets into image.

## 7. Local setup without Docker

Current local backend startup is blocked because CMS source files are missing.

Suggested local flow after implementation:

1. Start MySQL.
2. Apply CMS schema/migrations.
3. Set CMS env variables.
4. Start Product Service if CMS product boundary is needed.
5. Start CMS Service.
6. Start API Gateway.
7. Start Seller Dashboard frontend.

## 8. Required environment variables

| Variable | Required | Purpose |
|---|---|---|
| `CMS_HTTP_ADDR` | Yes | CMS HTTP listen address |
| `CMS_GRPC_ADDR` | Yes | CMS gRPC listen address |
| `CMS_ENV` | Yes | Runtime env |
| `CMS_INTERNAL_AUTH_HEADER` | Yes | Internal auth header name |
| `CMS_INTERNAL_AUTH_TOKEN` | Yes | Internal auth secret |
| `CMS_MYSQL_DSN` | Optional | Full DSN override |
| `CMS_DB_HOST` | Yes | MySQL host |
| `CMS_DB_PORT` | Yes | MySQL port |
| `CMS_DB_NAME` | Yes | MySQL database |
| `CMS_DB_USER` | Yes | MySQL username |
| `CMS_DB_PASSWORD` | Yes | MySQL password |
| `CMS_PRODUCT_SERVICE_BASE_URL` | Yes | Product service internal boundary |
| `CMS_CAMPAIGN_MAX_DURATION_DAYS` | Yes | Campaign guardrail |
| `CMS_ANALYTICS_DEFAULT_RANGE_DAYS` | Yes | Analytics default date range |
| `CMS_AUDIT_DEFAULT_PAGE_SIZE` | Yes | Audit pagination default |

## 9. Start commands

Suggested command based on project structure:

```bash
cd backend/services/cms-service
go run ./cmd/server
```

Current status: Not clearly runnable.

## 10. Verify running commands

```bash
curl http://localhost:8087/health/live
```

Health endpoint was not clearly found in project files.

For gRPC after proto/server exists:

```bash
# Suggested command based on project structure
grpcurl -plaintext localhost:9098 list
```

## 11. Common errors and fixes

| Error | Reason | Fix |
|---|---|---|
| Cannot start CMS | `go.mod`/entrypoint missing | Add backend module and service source |
| DB connection fail | MySQL not running or wrong credentials | Start MySQL and verify `CMS_DB_*` |
| Gateway cannot reach CMS | Port mismatch | Align `api-gateway/.env` `CMS_GRPC_ADDR` with CMS `CMS_GRPC_ADDR` |
| Coupons API returns 404 | Route/gRPC handler not wired | Add API contract, gateway route, CMS handler |
| Analytics empty | Aggregate/read model not implemented | Return honest unavailable data; do not fake metrics |

## 12. Security notes

- CMS must enforce seller ownership for every seller-scoped resource.
- Team role changes must be backend-authorized, frontend checks are not enough.
- Internal auth token must be secret and rotated outside git.
- Coupon validation and redemption must be idempotent and protected from race conditions.
- Audit logs should not store sensitive raw payloads.

## 13. Final checklist

- [x] CMS env file found.
- [x] CMS MySQL schema found in `database/draw.sql`.
- [x] CMS frontend API usage found.
- [ ] CMS Go source not clearly found.
- [ ] CMS service `go.mod` not found.
- [ ] CMS migrations not found.
- [ ] CMS proto/server registration not found.
- [ ] CMS Dockerfile not found.
