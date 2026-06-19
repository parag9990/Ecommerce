# Project Dependency & Setup Guide

## 1. Variables Used by This Guide

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Seller Dashboard (CMS)` |
| `TASK_FILE_NAME` | `task5.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task5_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |

Beginner note: is file ka purpose Task 5 implementation ko run karne ke liye dependencies, setup, env, database, Docker/DevOps, and debugging points explain karna hai. Business logic ya UI implementation yahan rewrite nahi kiya gaya.

## 2. Project Overview

Task 5 ka feature Revenue Analytics hai. Seller dashboard me seller revenue, GMV, orders, conversion, and top products dekh sakta hai.

Current implementation evidence:

| Area | Path |
|---|---|
| Task guide | `INPUT_FILE_PATH` |
| Route | `frontend/seller-dashboard/src/routes/seller-routes.tsx` |
| Page | `frontend/seller-dashboard/src/features/analytics/pages/revenue-analytics-page.tsx` |
| API client | `frontend/seller-dashboard/src/features/analytics/api/seller-analytics-api.ts` |
| React Query hook | `frontend/seller-dashboard/src/features/analytics/hooks/use-seller-analytics.ts` |
| Types and normalizer | `frontend/seller-dashboard/src/features/analytics/types.ts`, `frontend/seller-dashboard/src/features/analytics/utils/analytics-adapter.ts` |
| Date range helpers | `frontend/seller-dashboard/src/features/analytics/utils/analytics-date-range.ts` |
| Charts/cards/table | `frontend/seller-dashboard/src/features/analytics/components/` |
| API contract | `api/master-api.json` |
| CMS env | `backend/services/cms-service/.env` |
| Gateway env | `backend/services/api-gateway/.env` |

Runtime flow:

```text
Seller browser
  -> Vite Seller Dashboard frontend
  -> API Gateway REST endpoint
  -> CMSService.GetSellerAnalytics
  -> CMS analytics aggregate/read model
  -> seller analytics response
```

Important boundary:

- Task 5 frontend does not directly connect to MySQL, Redis, gRPC, Kafka, RabbitMQ, or CMS internals.
- Frontend only calls API Gateway through `frontend/seller-dashboard/src/lib/http.ts`.
- Backend aggregation, new database tables, scheduled jobs, and event pipelines are not implemented by this task.
- Missing metrics must show honest unavailable/empty states. Fake revenue, GMV, trend, or conversion data should not be generated.

## 3. Required Software

Most required software is already explained in previous dependency files. Follow those first.

| Software | Required for Task 5? | Status | Setup reference |
|---|---:|---|---|
| Node.js + pnpm | Yes | Reused | `TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md` |
| React/Vite frontend | Yes | Reused | `TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md` |
| API Gateway | Yes for real API calls | Reused/partial | `TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md` |
| CMS Backend | Yes for real analytics data | Reused/partial | `TaskImplementation/${SERVICE_NAME}/Dependency/CMS_Backend.md` |
| MySQL | Yes for CMS backend/read model | Reused | `TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md` |
| Redis | Only if gateway rate limiting is enabled | Reused | `TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md` |
| Go modules | Backend only | Reused/partial | `TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md` |
| Docker | Optional local orchestration | Missing/partial | `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` section `Docker and DevOps setup` |

Task 5 does not introduce a new runtime like Python, Kafka, RabbitMQ, MinIO, Elasticsearch, Stripe, Twilio, SMTP, Firebase, or Kubernetes.

## 4. Tech Stack

### Reused frontend stack

Ye technologies already previous dependency docs me explain ki gayi hain, so yahan full install explanation repeat nahi kiya gaya.

| Technology | Task 5 me use | Required? | Dependency status |
|---|---|---:|---|
| React | Revenue analytics page and components render karne ke liye | Yes | Already installed |
| TypeScript | API response, analytics types, chart row types safe rakhne ke liye | Yes | Already installed |
| Vite | Local dev server and production build | Yes | Already installed |
| React Router | `/seller/analytics` route activate karne ke liye | Yes | Already installed |
| TanStack React Query | Analytics summary fetch, cache, stale data handling | Yes | Already installed |
| Zustand | Active seller context read karne ke liye | Yes | Already installed |
| Lucide React | KPI/filter icons | Yes | Already installed |
| Tailwind CSS | Dense operational dashboard styling | Yes | Already installed |
| Vitest + Testing Library | Analytics utils and UI tests | Yes | Already installed |

Refer:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md
TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
TaskImplementation/${SERVICE_NAME}/task4_Dependency.md
```

### Task 5 actual package status

`INPUT_FILE_PATH` mentions some possible libraries, but the actual implementation has a smaller dependency footprint.

| Package/tool | Is it installed? | Is it needed by current Task 5 code? | Notes |
|---|---:|---:|---|
| `@tanstack/react-query` | Yes | Yes | `useSellerAnalytics` uses `useQuery` and `keepPreviousData`. |
| `lucide-react` | Yes | Yes | Icons in cards and date filter. |
| `react-router-dom` | Yes | Yes | Route registration. |
| `zustand` | Yes | Yes | Active seller store. |
| `tailwindcss` / `@tailwindcss/vite` | Yes | Yes | Styling. |
| `vitest` | Yes | Yes | Unit tests. |
| `@testing-library/react` | Yes | Yes | Existing component tests. |
| `recharts` | No | No | Current charts are custom SVG/CSS, so install mat karo unless implementation changes. |
| `date-fns` | No | No | Current date helpers native `Date` + UTC logic use karte hain. |
| `clsx` | No | No | Current project local `cn` helper use karta hai. |
| `msw` | No | No | Current analytics tests are utility-focused; MSW is not installed. |

Practical rule: agar current code compile ho raha hai, Task 5 ke liye `pnpm add recharts date-fns clsx msw` run karne ki zarurat nahi hai.

## 5. Dependency Management

### Frontend dependency system

Frontend setup pnpm workspace se managed hai.

Reused setup:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md

Section:
Installation steps
```

Current package file:

```text
frontend/seller-dashboard/package.json
```

Install existing workspace dependencies:

```bash
cd frontend
pnpm install
```

Run checks:

```bash
cd frontend
pnpm --filter seller-dashboard typecheck
pnpm --filter seller-dashboard test
pnpm --filter seller-dashboard build
```

Task 5-specific new install command:

```text
None.
```

### Backend dependency system

Backend dependency setup is Go Modules + gRPC/protobuf based, but current repo still does not clearly include runnable CMS/API Gateway Go source under `backend/services/cms-service/` and `backend/services/api-gateway/`.

Do not duplicate full Go setup here. Refer:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md
TaskImplementation/${SERVICE_NAME}/Dependency/gRPC.md
TaskImplementation/${SERVICE_NAME}/Dependency/Protobuf.md
```

Task 5 backend expectation:

- API Gateway should expose `GET /api/v1/seller/dashboard/summary`.
- Gateway should forward to `CMSService.GetSellerAnalytics`.
- CMS should return `SellerAnalyticsResponse`.
- CMS should use aggregate/read-model data, not direct frontend DB access.

## 6. Database Setup

### Reused database: MySQL

MySQL setup, Docker examples, credentials placement, and common DB troubleshooting are already explained in:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md
```

Task 5 does not require a new MySQL server. It depends on CMS backend data.

### Task 5 data requirement

Task 5 UI expects analytics data from CMS:

| Data | Current API support | Frontend behavior |
|---|---:|---|
| Revenue total | Yes | Show KPI card |
| Orders total | Yes | Show KPI card |
| Conversion rate | Yes | Show KPI card/gauge |
| Top products | Yes, loose object array | Normalize and show chart/table |
| GMV | Optional, not in master contract | Show unavailable if missing |
| Revenue/GMV trend series | Optional, not in master contract | Show empty chart state if missing |
| Orders trend series | Optional, not in master contract | Show empty chart state if missing |

Current schema check:

- `database/draw.sql` contains CMS tables like `seller_settings`, `seller_staff`, `coupons`, `coupon_rules`, `coupon_redemptions`, `campaigns`, and `cms_audit_logs`.
- A CMS-specific analytics aggregate table was not clearly found in `database/draw.sql`.
- Docs say seller analytics can query a read model or pre-aggregated data.

Beginner note: "aggregate/read model" ka matlab pehle se calculated summary table or materialized data hota hai. Analytics page ko har request par raw order/payment/session tables scan nahi karna chahiye.

### Credentials placement

Frontend me DB credentials kabhi mat rakho.

CMS DB credentials belong here:

```text
backend/services/cms-service/.env
```

Reused variables:

```env
CMS_MYSQL_DSN=
CMS_DB_HOST=localhost
CMS_DB_PORT=3306
CMS_DB_NAME=cms_db
CMS_DB_USER=cms_user
CMS_DB_PASSWORD=change-me
CMS_DB_TIMEZONE=UTC
```

This setup is already explained in:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md
TaskImplementation/${SERVICE_NAME}/Dependency/Environment.md
```

## 7. Redis / Queue / External Services

### Redis

Task 5 frontend does not directly use Redis.

Redis matters only because all frontend API calls pass through API Gateway, and gateway env currently has:

```env
RATE_LIMIT_ENABLED=true
REDIS_ADDR=localhost:6379
RATE_LIMIT_FAIL_OPEN=false
```

Redis setup is already explained in:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md
```

If Redis is not running and rate limiting remains enabled, Gateway requests for analytics may fail or return rate-limit related errors.

### API Gateway

Task 5 uses this REST endpoint:

| Purpose | Method | Endpoint | Auth | Contract source |
|---|---|---|---|---|
| Seller analytics summary | `GET` | `/api/v1/seller/dashboard/summary` | seller | `api/master-api.json` |

Expected query format:

```text
GET /api/v1/seller/dashboard/summary?from=2026-05-01T00:00:00.000Z&to=2026-06-01T23:59:59.999Z
```

Current contract response:

```json
{
  "revenue": {
    "amount": 125000,
    "currency": "INR"
  },
  "orders": 84,
  "conversion_rate": 3.4,
  "top_products": []
}
```

Important: API may wrap this in `{ "data": ... }`; frontend `http.ts` supports both envelope and direct JSON.

### CMS Service

CMS setup is already documented in:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/CMS_Backend.md
```

Task 5 specifically requires:

```text
CMSService.GetSellerAnalytics
```

The CMS service should enforce seller ownership server-side. Browser se trusted `seller_id` accept karke data expose karna insecure hoga.

### Queues and event systems

No Task 5-specific queue setup was found.

| Service | Task 5 status |
|---|---|
| Kafka | Not used directly |
| RabbitMQ | Not used directly |
| NATS | Not used directly |
| MinIO/S3 | Not used directly |
| Elasticsearch | Not used directly |
| SMTP/Twilio/Stripe/Firebase | Not used directly |

If future backend analytics derives conversion from session events or order/payment events, that backend task should add its own queue/event setup documentation. This Task 5 dependency file should not invent a queue requirement.

## 8. Environment Variables

Common frontend, gateway, CMS, MySQL, and Redis env rules are already explained in:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/Environment.md
```

### Frontend env

Task 5 introduces no new frontend env variable.

Reused optional frontend env:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT_MS=15000
VITE_LOGIN_URL=/login
```

Where to place:

```text
frontend/seller-dashboard/.env.local
```

Security note: `VITE_*` values browser bundle me visible hote hain. Internal tokens, JWT secrets, DB password, Redis password, or CMS auth token yahan kabhi mat rakho.

### Task 5 CMS analytics env

These values exist in:

```text
backend/services/cms-service/.env
```

Task 5-specific CMS env example:

```env
CMS_ANALYTICS_DEFAULT_CURRENCY=INR
CMS_ANALYTICS_DEFAULT_RANGE_DAYS=30
CMS_ANALYTICS_MAX_RANGE_DAYS=366
CMS_ANALYTICS_DEFAULT_TOP_PRODUCTS_LIMIT=5
CMS_ANALYTICS_MAX_TOP_PRODUCTS_LIMIT=20
```

| Variable | Required? | Purpose | Example | Security note |
|---|---:|---|---|---|
| `CMS_ANALYTICS_DEFAULT_CURRENCY` | Yes for CMS analytics | Missing money currency ka default | `INR` | Not secret |
| `CMS_ANALYTICS_DEFAULT_RANGE_DAYS` | Yes | Backend default range if client does not send range | `30` | Not secret |
| `CMS_ANALYTICS_MAX_RANGE_DAYS` | Yes | Very large analytics query prevent karta hai | `366` | Not secret |
| `CMS_ANALYTICS_DEFAULT_TOP_PRODUCTS_LIMIT` | Yes | Default top product count | `5` | Not secret |
| `CMS_ANALYTICS_MAX_TOP_PRODUCTS_LIMIT` | Yes | Response size guardrail | `20` | Not secret |

Beginner note: ye values frontend env me nahi jaati. Ye backend CMS service ke guardrails hain.

### Gateway/CMS gRPC alignment

This issue is already documented in previous dependency files, but Task 5 cannot call CMS analytics unless it is fixed.

Current evidence:

```text
backend/services/cms-service/.env       CMS_GRPC_ADDR=:9098
backend/services/api-gateway/.env       CMS_GRPC_ADDR=localhost:50059
```

Choose one alignment:

```env
# Option A: keep CMS service on 9098 and update gateway
CMS_GRPC_ADDR=localhost:9098
```

or:

```env
# Option B: make CMS service listen on 50059 and keep gateway target as-is
CMS_GRPC_ADDR=:50059
```

Do not keep both different, otherwise Gateway analytics calls can fail with gRPC unavailable/connection refused.

## 9. Docker Setup

Task 5 adds no new Docker container, volume, network, health check, or restart policy.

Reused Docker/DevOps setup:

```text
TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
TaskImplementation/${SERVICE_NAME}/task4_Dependency.md
TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md
TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md
```

Current project gap:

- No `frontend/seller-dashboard/Dockerfile` clearly found.
- No CMS Service Dockerfile clearly found.
- No API Gateway Dockerfile clearly found.
- No local docker-compose file clearly found.

If Docker Compose is added later, Task 5 full-stack local run would need at least:

| Container/service | Purpose | Status |
|---|---|---|
| MySQL | CMS database / analytics read model | Reused |
| Redis | Gateway rate limiting | Reused if enabled |
| CMS Service | `GetSellerAnalytics` backend | Required for real data |
| API Gateway | Browser REST entrypoint | Required |
| Seller Dashboard frontend | UI | Required |

## 10. Local Development Setup

### Step 1: Read previous dependency docs first

Follow these before Task 5-specific checks:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md
TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md
TaskImplementation/${SERVICE_NAME}/Dependency/CMS_Backend.md
TaskImplementation/${SERVICE_NAME}/Dependency/Environment.md
TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md
TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md
TaskImplementation/${SERVICE_NAME}/task4_Dependency.md
```

### Step 2: Go to project directory

From repository root:

```bash
pwd
```

Expected root shape:

```text
<your-cloned-repo>
```

### Step 3: Install frontend dependencies

Task 5 adds no new package. Use existing workspace install:

```bash
cd frontend
pnpm install
```

### Step 4: Configure frontend API target

Create only if defaults are not enough:

```text
frontend/seller-dashboard/.env.local
```

Example:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT_MS=15000
VITE_LOGIN_URL=/login
```

### Step 5: Configure CMS analytics guardrails

Confirm these exist in:

```text
backend/services/cms-service/.env
```

```env
CMS_ANALYTICS_DEFAULT_CURRENCY=INR
CMS_ANALYTICS_DEFAULT_RANGE_DAYS=30
CMS_ANALYTICS_MAX_RANGE_DAYS=366
CMS_ANALYTICS_DEFAULT_TOP_PRODUCTS_LIMIT=5
CMS_ANALYTICS_MAX_TOP_PRODUCTS_LIMIT=20
```

### Step 6: Setup database/services

No new Task 5 DB install is needed.

Required for real backend analytics:

1. MySQL running.
2. CMS schema applied.
3. Any CMS analytics aggregate/read-model tables or queries implemented by backend.
4. Redis running if gateway rate limiting stays enabled.

Reused setup docs:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md
TaskImplementation/${SERVICE_NAME}/Dependency/Migrations.md
TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md
```

### Step 7: Start backend services

Current repo status: backend source entrypoints for CMS Service and API Gateway were not clearly found, so exact commands may be blocked.

Expected order after backend code exists:

1. Start MySQL.
2. Start Redis if `RATE_LIMIT_ENABLED=true`.
3. Start CMS Service with aligned `CMS_GRPC_ADDR`.
4. Start API Gateway on `8080`.
5. Start Seller Dashboard frontend.

Suggested commands based on intended structure:

```bash
cd backend/services/cms-service
go run ./cmd/server
```

```bash
cd backend/services/api-gateway
go run ./cmd/server
```

These commands require service `go.mod`, `cmd/server/main.go`, config loaders, route handlers, gRPC clients/servers, and migrations to exist.

### Step 8: Start Seller Dashboard frontend

```bash
cd frontend
pnpm --filter seller-dashboard dev
```

Open:

```text
http://localhost:5174/seller/analytics
```

### Step 9: Verify Task 5 functionality

Frontend checks:

```bash
cd frontend
pnpm --filter seller-dashboard typecheck
pnpm --filter seller-dashboard test
pnpm --filter seller-dashboard build
```

Manual UI checks:

- `/seller/analytics` route opens for seller/manager/order manager roles.
- User without `analytics:view` sees permission denied.
- Date presets `7D`, `30D`, `90D` update query range.
- Custom from/to date inputs produce ISO `from` and `to`.
- Missing GMV or trend series shows unavailable/empty state.
- Top products table/chart does not crash on loose backend object shape.

API check through Gateway:

```bash
curl "http://localhost:8080/api/v1/seller/dashboard/summary?from=2026-05-01T00:00:00.000Z&to=2026-06-01T23:59:59.999Z"
```

Expected local note: without valid seller auth cookie/session, this can return `401` or `403`. That is normal for an auth-protected seller endpoint.

## 11. Running the Project

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Seller Dashboard dev server | `5174` | Frontend local UI | Reused |
| Seller Dashboard preview | `4174` | Frontend preview build | Reused |
| API Gateway | `8080` | REST entrypoint used by frontend | Reused |
| CMS HTTP | `8087` | CMS service HTTP, if implemented | Reused |
| CMS gRPC | `9098` | CMS service gRPC listen from CMS env | Needs gateway alignment |
| Gateway CMS target | `50059` | Current gateway CMS target | Mismatch unless aligned |
| MySQL | `3306` | CMS database | Reused |
| Redis | `6379` | Gateway rate limiting | Reused if enabled |

Port conflict fixes:

| Conflict | Fix |
|---|---|
| `5174` already used | Change Vite dev server config or stop existing process. |
| `8080` already used | Change gateway `HTTP_ADDR` and frontend `VITE_API_BASE_URL` together. |
| `9098` or `50059` already used | Change CMS listen port and gateway `CMS_GRPC_ADDR` together. |
| `6379` unavailable | Start Redis, change `REDIS_ADDR`, or disable rate limit locally. |
| `3306` unavailable | Use existing MySQL or remap Docker host port and update `CMS_DB_PORT`. |

Docker networking note: if backend services run inside Docker Compose, `localhost` from one container means that container itself. Use compose service names like `cms-service:9098`, `mysql:3306`, and `redis:6379` inside compose networks.

## 12. Common Errors & Fixes

Only Task 5-specific or Task 5-relevant errors are listed here. Generic pnpm, MySQL, Redis, Docker, and Go module errors are already covered in previous dependency docs.

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| Analytics page shows permission denied | Active role does not include `analytics:view`, or active seller context missing | Use seller, seller_manager, or seller_order_manager role; verify seller session response | Seed seller_staff/roles correctly |
| Frontend network error | API Gateway not running or `VITE_API_BASE_URL` wrong | Start Gateway or set frontend env to correct URL | Keep `.env.local` aligned with gateway port |
| API returns `401` | Seller auth cookie/session missing | Login through auth flow or create valid local session | Do not bypass backend auth in frontend |
| API returns `403` | Seller role/status not allowed | Verify backend seller ownership/status/RBAC | Backend should return clear auth errors |
| Gateway returns CMS unavailable | CMS service not running or gRPC port mismatch | Align `CMS_GRPC_ADDR` in gateway and CMS env | Keep one source of truth for service ports |
| GMV card says `Not available` | API contract does not return `gmv` | This is expected unless backend adds `gmv` | Do not fake GMV in frontend |
| Trend charts show empty state | API does not return `series` | This is expected with current master contract | Add backend contract/aggregation before expecting charts |
| Top products missing rows | Backend object lacks product id/name or values are malformed | Fix backend response mapper | Add contract tests for top product shape |
| Money values look 100x high/low | Backend and frontend disagree on minor units | Align contract: current formatter treats `amount` as paise/cents | Document money unit in API contract |
| `Cannot find module 'recharts'` | Someone copied old proposed code, but package is not installed | Use current custom chart implementation or intentionally add package | Keep implementation and docs in sync |
| Date range returns too much data | Backend max range not enforced | Enforce `CMS_ANALYTICS_MAX_RANGE_DAYS` in CMS | Validate range server-side |
| CORS/cookie issue | Gateway/auth cookies not configured for frontend origin | Configure CORS and same-site cookie settings | Test with real browser session |

## 13. Security & Best Practices

### Seller ownership and permissions

- Frontend permission checks improve UX, but backend must enforce real security.
- `analytics:view` is required by current UI.
- Roles with analytics access in current frontend permission matrix: `seller`, `seller_manager`, and `seller_order_manager`.
- Backend must never trust a browser-sent seller id without validating ownership/session.

### Analytics data safety

- Revenue and GMV are sensitive business metrics. Do not expose cross-seller data.
- Top products should not include buyer PII.
- Conversion metric definition should come from backend contract, not frontend guesswork.
- "Zero" and "Not available" are different: zero means backend measured zero; unavailable means backend did not provide the metric.

### Environment and secrets

- Do not put `CMS_INTERNAL_AUTH_TOKEN`, DB password, Redis password, JWT keys, or internal service tokens in `VITE_*`.
- Keep real `.env` local/private.
- Add sanitized `.env.example` files before onboarding beginners.
- Rotate placeholder `change-me` values before any shared or deployed environment.

### Backend performance

- Use aggregate/read-model data for dashboard analytics.
- Avoid scanning raw order/payment/session tables on every dashboard page load.
- Enforce max date range and top product limit.
- Add indexes on seller id and time bucket if analytics tables are added.
- Add request IDs to logs; frontend already sends `x-request-id`.

## 14. Missing or Misconfigured Things

| Item | Current evidence | Impact | Suggested fix |
|---|---|---|---|
| CMS Service Go source | Only `backend/services/cms-service/.env` clearly found | `GetSellerAnalytics` cannot run locally from current files | Add CMS service module, entrypoint, config, gRPC server, repository, and tests |
| API Gateway Go source | Only `backend/services/api-gateway/.env` clearly found | Frontend cannot call real gateway locally from current files | Add gateway module, REST routes, auth middleware, gRPC clients |
| CMS gRPC port mismatch | CMS env `:9098`, gateway env `localhost:50059` | Analytics route can fail through gateway | Align both values |
| CMS analytics read model/migration | No CMS analytics aggregate table clearly found in `database/draw.sql` | Backend may not have real analytics source | Add versioned migrations/read model if backend owns aggregates |
| Master API contract lacks GMV/series | `SellerAnalyticsResponse` has revenue, orders, conversion_rate, top_products only | GMV/trend charts show unavailable states | Extend API contract intentionally if backend supports these fields |
| Protobuf files | `proto/` directory not clearly found | gRPC codegen blocked | Add `proto/ecommerce/cms/v1/cms.proto` or selected contract source |
| Docker Compose | No compose file clearly found | Beginners must start services manually | Add local compose for MySQL, Redis, CMS, Gateway, frontend |
| `.env.example` files | Not clearly found | Onboarding confusion and secret leakage risk | Add sanitized examples for frontend, gateway, CMS |
| Backend health checks | Not clearly found | Hard to verify CMS/Gateway readiness | Add `/health/live`, `/health/ready`, and gRPC health checks |
| Analytics backend contract tests | Not clearly found | Response shape drift can break UI | Add tests for `SellerAnalyticsResponse` and top product shape |
| Proposed vs actual chart packages | Task guide mentions Recharts/date-fns/MSW possibilities, actual code does not use them | Beginners may install unnecessary packages | Use current package.json as source of truth |

## 15. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | Project overview, frontend install, MySQL, Redis, env, run flow | Task 5 runs inside the same seller dashboard shell. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | Frontend workspace, API Gateway, product/service dependency pattern | Same frontend and gateway architecture is reused. |
| `TaskImplementation/${SERVICE_NAME}/task3_Dependency.md` | Order Service/data dependency pattern, MySQL reuse, backend gaps | Revenue/orders analytics depends conceptually on order/payment aggregates. |
| `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` | CMS Service, gateway/CMS port alignment, Docker gaps, local setup flow | Task 5 uses same CMS backend boundary as offers. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md` | React/Vite/pnpm commands, frontend env, ports | No new Task 5 frontend package setup is needed. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md` | Gateway env, REST-to-gRPC flow, Redis rate limit | Analytics endpoint is called through Gateway. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/CMS_Backend.md` | CMS Service purpose, env, missing backend source | Analytics belongs to CMS backend. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Environment.md` | Env file locations and secret rules | Task 5 only adds/uses CMS analytics guardrails. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md` | MySQL install, credentials, troubleshooting | CMS DB setup is reused. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md` | Redis setup and rate-limit behavior | Redis is reused through Gateway rate limiting. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md` | Go module/workspace setup | Backend services are expected to be Go modules. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/gRPC.md` | gRPC setup and port mismatch troubleshooting | Gateway forwards to `CMSService.GetSellerAnalytics`. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Protobuf.md` | Proto/codegen expectations | CMS gRPC contract should be generated/defined. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Migrations.md` | Migration requirements and current gaps | Analytics read-model schema should use migrations if added. |

## 16. Final Checklist

- [ ] Previous dependency documentation checked before starting Task 5 setup.
- [ ] `frontend/seller-dashboard/package.json` used as source of truth for packages.
- [ ] No unnecessary `recharts`, `date-fns`, `clsx`, or `msw` install added.
- [ ] Frontend dependencies installed with `pnpm install`.
- [ ] `VITE_API_BASE_URL` points to running API Gateway.
- [ ] CMS analytics env guardrails confirmed in `backend/services/cms-service/.env`.
- [ ] CMS and Gateway `CMS_GRPC_ADDR` values aligned.
- [ ] MySQL/CMS schema available for backend, if running real CMS analytics.
- [ ] Redis running or `RATE_LIMIT_ENABLED=false` intentionally set for local.
- [ ] CMS `GetSellerAnalytics` backend implemented before expecting real data.
- [ ] `/seller/analytics` route verified with a seller role that has `analytics:view`.
- [ ] Missing GMV/trend data handled as unavailable, not fake values.
- [ ] Frontend `typecheck`, `test`, and `build` commands pass.
- [ ] No DB credentials or internal tokens added to frontend `VITE_*` env.
- [ ] Docker, migrations, `.env.example`, proto, and health check gaps tracked as follow-up work.
