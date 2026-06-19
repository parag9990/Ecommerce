# Project Dependency & Setup Guide

## 1. Variables used by this guide

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Seller Dashboard (CMS)` |
| `TASK_FILE_NAME` | `task4.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task4_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |

This guide is for the Offers and Coupons work documented in `INPUT_FILE_PATH`.

Simple Hinglish goal: Task 4 me seller dashboard ke andar coupons create/edit karna, campaign calendar dekhna, campaign create karna, aur usage stats UI show karna hai. Ye file business logic repeat nahi karti. Ye sirf setup, dependencies, environment, database, Docker/DevOps, and debugging requirements explain karti hai.

## 2. What changed from previous tasks

Task 1, Task 2, and Task 3 already explain most shared setup. Task 4 reuses the same frontend app, API Gateway, CMS Service, MySQL, Redis, Go/gRPC assumptions, and local tooling.

| New / Reused | Dependency | Why it matters for Task 4 |
|---|---|---|
| Reused | React, TypeScript, Vite, pnpm workspace | Offers module same seller dashboard frontend app ke andar implemented hai. |
| Reused | React Router | `/seller/offers`, coupon editor, and campaign create routes registered hain. |
| Reused | TanStack React Query | Coupon/campaign list fetch, cache, and mutation invalidation ke liye. |
| Reused | React Hook Form, Zod, `@hookform/resolvers` | Coupon and campaign forms validate karne ke liye. These were already introduced earlier. |
| Reused | Zustand seller store | Active seller context required hai before API calls run. |
| Reused | Lucide React | Offers buttons/icons ke liye. |
| Reused | API Gateway on `8080` | Browser REST calls gateway ko hit karti hain. |
| Reused | Redis, if rate limiting enabled | Gateway rate limiting ke liye. Task 4 direct Redis use nahi karta. |
| Reused | CMS Service boundary | Coupons and campaigns ka backend owner hai. |
| Reused with Task 4 focus | MySQL `cms_db` | `coupons`, `coupon_rules`, `coupon_redemptions`, and `campaigns` tables required hain. |
| Task-specific existing env | `CMS_CAMPAIGN_MAX_DURATION_DAYS` | Campaign duration guardrail backend config me present hai. |
| Not added | `date-fns`, MSW, Kafka, RabbitMQ, S3, Stripe | Actual Task 4 code/package list me ye required nahi hain. |

Runtime flow:

```text
Seller browser
  -> Vite Seller Dashboard frontend
  -> API Gateway REST API
  -> CMSService gRPC methods
  -> MySQL cms_db
```

Important boundary:

- Frontend directly MySQL, Redis, ya CMS gRPC ko call nahi karta.
- Frontend only REST calls karta hai through API Gateway.
- Backend must enforce seller ownership and offer permissions. Frontend checks UX ke liye hain, security ke liye enough nahi.
- Current repo me CMS and API Gateway `.env` files hain, but full Go service source/entrypoints clearly available nahi hain. Backend run commands tabhi work karenge jab service source added ho.

## 3. Reuse previous dependency documentation first

Do not duplicate setup that is already documented. Follow these existing files before using this task-specific guide:

| Setup topic | Already explained in |
|---|---|
| Frontend install, Vite ports, pnpm commands | `TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md` |
| General Task 1 local setup and frontend env | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` |
| API Gateway, JWT, Redis rate limit, gRPC target basics | `TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md` |
| Environment variable locations and safety | `TaskImplementation/${SERVICE_NAME}/Dependency/Environment.md` |
| CMS backend config and current backend gaps | `TaskImplementation/${SERVICE_NAME}/Dependency/CMS_Backend.md` |
| MySQL installation and `cms_db` generic setup | `TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md` |
| Migration expectations and current raw SQL fallback | `TaskImplementation/${SERVICE_NAME}/Dependency/Migrations.md` |
| Redis setup, if gateway rate limiting stays enabled | `TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md` |
| Go modules and backend dependency flow | `TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md` |
| gRPC and Protobuf setup | `TaskImplementation/${SERVICE_NAME}/Dependency/gRPC.md`, `TaskImplementation/${SERVICE_NAME}/Dependency/Protobuf.md` |

This file only adds Task 4-specific details: offers routes, CMS coupon/campaign APIs, CMS DB tables, offer permissions, and setup checks.

## 4. Tech stack analysis

### Reused frontend stack

The following technologies are already explained in earlier docs, so do not reinstall or re-explain them from scratch:

```text
React
TypeScript
Vite
pnpm workspace
Tailwind CSS v4
React Router
TanStack React Query
React Hook Form
Zod
@hookform/resolvers
Zustand
Lucide React
Vitest + Testing Library
```

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md`

Task 4 actual dependency status from `frontend/seller-dashboard/package.json`:

| Package | Used for Task 4 | Status |
|---|---|---|
| `@tanstack/react-query` | Coupon/campaign queries and mutations | Already installed |
| `react-hook-form` | Coupon and campaign form state | Already installed |
| `zod` | Form validation | Already installed |
| `@hookform/resolvers` | Zod resolver for React Hook Form | Already installed |
| `react-router-dom` | Offers routes and navigation | Already installed |
| `zustand` | Active seller context | Already installed |
| `lucide-react` | UI icons | Already installed |
| `tailwindcss` / `@tailwindcss/vite` | Dashboard UI styling | Already installed |

No new frontend package installation is required for Task 4.

Important correction: Task 4 implementation does not use `date-fns`. Calendar/date helpers use native `Date` and `Intl.DateTimeFormat` in:

```text
frontend/seller-dashboard/src/features/offers/utils/offer-date-rules.ts
```

### Task 4 backend stack

| Technology | Required? | Beginner-friendly explanation | Why Task 4 uses it |
|---|---:|---|---|
| CMS Service | Yes for real data | CMS Service seller dashboard ka backend owner hai. Simple words me, offers, coupons, campaigns, team, analytics, audit jaise dashboard data yahi manage karta hai. | `GET/POST/PATCH /api/v1/seller/coupons` and `GET/POST /api/v1/seller/campaigns` isi service boundary se aate hain. |
| API Gateway | Yes | Gateway browser se REST request leta hai aur internal service ko gRPC call forward karta hai. | Frontend `http.ts` default `http://localhost:8080` par calls bhejta hai. |
| MySQL `cms_db` | Yes for real CMS APIs | MySQL relational database hai jisme data tables me store hota hai. | Coupons, coupon redemptions, coupon rules, campaigns, seller staff, and audit data structured hai. |
| Redis | Required only if gateway rate limit enabled | Redis fast in-memory store hai. | Gateway env me `RATE_LIMIT_ENABLED=true` hai, so Redis unavailable hoga to gateway fail/429 issues aa sakte hain. |
| gRPC / Protobuf | Yes for full backend | Typed service-to-service communication. | Gateway CMS REST routes ko `CMSService.*` methods tak forward karega. |

## 5. Dependency management

### Frontend dependency system

Frontend setup is already explained in:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md
```

Use the normal install:

```bash
cd frontend
pnpm install
```

Then verify Task 4 code with existing scripts:

```bash
cd frontend
pnpm --filter seller-dashboard typecheck
pnpm --filter seller-dashboard test
```

Do not run `pnpm add` for Task 4 unless a package is genuinely missing from `frontend/seller-dashboard/package.json`.

### Backend dependency system

Go module setup is reused from:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md
```

Current repo observation:

```text
backend/services/cms-service/.env
backend/services/api-gateway/.env
```

were found, but CMS/API Gateway Go source entrypoints were not clearly found. So these suggested commands are blocked until backend source exists:

```bash
cd backend/services/cms-service
go mod tidy
go run ./cmd/server

cd backend/services/api-gateway
go mod tidy
go run ./cmd/server
```

When the CMS module exists, `backend/go.work` should include it:

```bash
cd backend
go work use ./services/cms-service
go work sync
```

## 6. Database analysis

### Reused database technology: MySQL

MySQL installation, Docker examples, credentials placement, and common DB troubleshooting are already explained in:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md
```

Task 4 does not need a new MySQL server. It needs the CMS schema inside `cms_db`.

### Task 4 database tables

Schema source currently available:

```text
database/draw.sql
```

Task 4 relevant tables:

| Table | Required? | Why it matters |
|---|---:|---|
| `coupons` | Yes | Coupon code, discount type/value, seller, status, date window, usage limit. |
| `coupon_rules` | Backend/future rules | Advanced rule storage. Current frontend does not submit custom rules directly. |
| `coupon_redemptions` | Backend/internal usage stats | Records order/user coupon redemption. Current frontend shows optional usage fields if API returns them. |
| `campaigns` | Yes | Campaign name, date window, budget, status, seller. |
| `seller_staff` | Yes for permissions | Seller role/status data supports access control and active seller context. |
| `cms_audit_logs` | Recommended | Coupon/campaign create/update should be auditable when backend supports it. |

Default MySQL port:

```text
3306
```

Local schema bootstrap is reused from the migration docs:

```bash
mysql -u root -p < database/draw.sql
```

Verify Task 4 tables:

```bash
mysql -u cms_user -p -D cms_db -e "SHOW TABLES LIKE 'coupons';"
mysql -u cms_user -p -D cms_db -e "SHOW TABLES LIKE 'campaigns';"
mysql -u cms_user -p -D cms_db -e "SHOW INDEX FROM coupons;"
mysql -u cms_user -p -D cms_db -e "SHOW INDEX FROM campaigns;"
```

Expected useful indexes from `database/draw.sql`:

| Table | Index | Why useful |
|---|---|---|
| `coupons` | `uk_coupons_code` | Duplicate coupon code prevent karta hai. |
| `coupons` | `idx_coupons_seller_status` | Seller/status filter fast banata hai. |
| `coupons` | `idx_coupons_window` | Start/end date filter fast banata hai. |
| `campaigns` | `idx_campaigns_seller_window` | Seller calendar window query fast banata hai. |
| `campaigns` | `idx_campaigns_status` | Status filter fast banata hai. |

Connection config placement:

```text
backend/services/cms-service/.env
```

Task 4 uses the same CMS DB credentials:

```env
CMS_MYSQL_DSN=
CMS_DB_HOST=localhost
CMS_DB_PORT=3306
CMS_DB_NAME=cms_db
CMS_DB_USER=cms_user
CMS_DB_PASSWORD=change-me
CMS_DB_TIMEZONE=UTC
```

Do not put DB username/password in frontend `VITE_*` variables.

## 7. Environment variables

Common frontend, gateway, MySQL, Redis, and CMS env behavior is already explained in:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/Environment.md
```

### Frontend env

Task 4 introduces no new frontend env variable.

Reused optional frontend env:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT_MS=15000
VITE_LOGIN_URL=/login
```

Where to place it:

```text
frontend/seller-dashboard/.env.local
```

Beginner note: `VITE_*` values browser bundle me visible hote hain. Secret token, DB password, internal auth token yahan kabhi mat rakho.

### CMS env values required by Task 4

Task 4 uses existing CMS Service env. The task-specific value to know is:

```env
CMS_CAMPAIGN_MAX_DURATION_DAYS=90
```

| Variable | Required? | Purpose | Example | Security notes |
|---|---:|---|---|---|
| `CMS_CAMPAIGN_MAX_DURATION_DAYS` | Yes for campaign guardrail | Backend maximum campaign duration enforce kar sakta hai. | `90` | Not secret. Keep frontend validation aligned if backend changes it. |
| `CMS_GRPC_ADDR` | Yes | CMS gRPC listen address. | `:9098` or `:50059` | Not secret, but must match gateway target. |
| `CMS_INTERNAL_AUTH_TOKEN` | Yes | Internal service auth token. | `change-me` local only | Secret. Change outside local and never expose to frontend. |
| `CMS_DB_*` / `CMS_MYSQL_DSN` | Yes | MySQL credentials and connection. | `cms_db` on `localhost:3306` | Contains DB password. Keep local only. |

### Critical CMS/Gateway address alignment

Current files show a mismatch:

```text
backend/services/cms-service/.env       CMS_GRPC_ADDR=:9098
backend/services/api-gateway/.env       CMS_GRPC_ADDR=localhost:50059
```

Task 4 coupon/campaign APIs will fail through the gateway until these are aligned.

Choose one local option:

```env
# Option A: keep CMS service on 9098 and update gateway
CMS_GRPC_ADDR=localhost:9098
```

or:

```env
# Option B: make CMS service listen on 50059 and keep gateway as-is
CMS_GRPC_ADDR=:50059
```

Do not leave both values different.

### Gateway env needed by Task 4

Where:

```text
backend/services/api-gateway/.env
```

Confirm:

```env
HTTP_ADDR=:8080
API_BASE_PATH=/api/v1
CMS_GRPC_ADDR=localhost:9098
RATE_LIMIT_ENABLED=true
REDIS_ADDR=localhost:6379
```

If `RATE_LIMIT_ENABLED=true`, Redis must be running. If Redis is not available in local development, either start Redis or explicitly set:

```env
RATE_LIMIT_ENABLED=false
```

Use that only for local development.

## 8. External services analysis

### API Gateway

Setup is reused from:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md
```

Task 4 required REST routes from `api/master-api.json`:

| Action | Method | Path | Gateway target |
|---|---|---|---|
| List coupons | `GET` | `/api/v1/seller/coupons` | `CMSService.ListCoupons` |
| Create coupon | `POST` | `/api/v1/seller/coupons` | `CMSService.CreateCoupon` |
| Update coupon | `PATCH` | `/api/v1/seller/coupons/{coupon_id}` | `CMSService.UpdateCoupon` |
| List campaigns | `GET` | `/api/v1/seller/campaigns` | `CMSService.ListCampaigns` |
| Create campaign | `POST` | `/api/v1/seller/campaigns` | `CMSService.CreateCampaign` |

Current Task 4 frontend sends these filters:

| Route | Query params used |
|---|---|
| `/api/v1/seller/coupons` | `status`, `q`, `starts_from`, `ends_before`, `page`, `page_size` |
| `/api/v1/seller/campaigns` | `status`, `month`, `page`, `page_size` |

Backend should ignore unsupported filters safely or implement them.

### CMS Service

Setup is reused from:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/CMS_Backend.md
```

Task 4 specifically requires CMS handlers for:

```text
CMSService.ListCoupons
CMSService.CreateCoupon
CMSService.UpdateCoupon
CMSService.ListCampaigns
CMSService.CreateCampaign
```

Backend should validate:

- authenticated seller context
- seller ownership
- `offers:view` for reads
- `offers:write` for create/update
- coupon code format and uniqueness
- discount type and discount value
- campaign start/end date range
- `CMS_CAMPAIGN_MAX_DURATION_DAYS`

### Redis

Redis setup is reused from:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md
```

Task 4 has no direct Redis dependency. Redis matters only because all frontend API calls pass through API Gateway, and gateway config currently enables rate limiting.

### Queues, object storage, payment providers

Task 4 does not introduce:

```text
Kafka
RabbitMQ
NATS
MinIO
AWS S3
Firebase
SMTP
Stripe/Razorpay
Twilio
Elasticsearch
Kubernetes
```

Coupon redemption is related to Cart/Order flows, but the Task 4 Seller Dashboard frontend does not call cart preview or order redemption APIs.

## 9. Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Seller Dashboard Vite | `5174` | Frontend dev server | Reused |
| Seller Dashboard preview | `4174` | Built frontend preview | Reused |
| API Gateway | `8080` | Browser REST API entrypoint | Reused |
| MySQL | `3306` | CMS `cms_db` data | Reused |
| Redis | `6379` | Gateway rate limiting, if enabled | Reused |
| CMS Service HTTP | `8087` | Suggested CMS health/internal HTTP | Reused config |
| CMS Service gRPC | `9098` or `50059` | Gateway -> CMS calls | Must align |

Port conflict tips:

- If Vite `5174` is busy, update `frontend/seller-dashboard/vite.config.ts` or pass a different Vite port.
- If MySQL `3306` is busy, do not start a second MySQL container. Reuse the existing server and create/verify `cms_db`.
- If CMS gRPC uses Docker Compose, gateway must use service DNS such as `cms-service:9098`, not `localhost:9098`.
- If backend runs on host machine, gateway can use `localhost:<port>`.

## 10. Docker and DevOps setup

Generic Docker setup is already covered in previous dependency docs. Task 4 does not add a new container type.

Task 4 local service status:

| Container/service | Needed for Task 4? | Notes |
|---|---:|---|
| MySQL | Yes for real CMS APIs | Reuse the same MySQL from previous setup. |
| Redis | If gateway rate limit enabled | Reuse previous Redis setup. |
| API Gateway | Yes for real APIs | Dockerfile/compose not clearly found. |
| CMS Service | Yes for real APIs | Dockerfile/compose not clearly found. |
| RabbitMQ/MongoDB/Product Service | No direct Task 4 need | Product Service may still matter to the wider app, but not for coupon/campaign CRUD. |

Suggested future compose wiring:

```yaml
services:
  cms-service:
    build:
      context: ./backend/services/cms-service
    env_file:
      - ./backend/services/cms-service/.env
    depends_on:
      mysql:
        condition: service_healthy
    expose:
      - "9098"

  api-gateway:
    environment:
      CMS_GRPC_ADDR: cms-service:9098
    depends_on:
      - cms-service
      - redis
```

This is a suggested future snippet, not an existing project file.

Docker networking note:

- Host-run CMS uses `CMS_DB_HOST=localhost`.
- Compose-run CMS should use `CMS_DB_HOST=mysql`.
- Host-run gateway uses `CMS_GRPC_ADDR=localhost:9098`.
- Compose-run gateway should use `CMS_GRPC_ADDR=cms-service:9098`.

## 11. Local development setup

### Step 1: Read reused docs

Read these first:

```text
TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md
TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md
TaskImplementation/${SERVICE_NAME}/Dependency/CMS_Backend.md
TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md
TaskImplementation/${SERVICE_NAME}/Dependency/Environment.md
TaskImplementation/${SERVICE_NAME}/Dependency/Migrations.md
```

### Step 2: Install frontend dependencies

No Task 4-specific package install needed.

```bash
cd frontend
pnpm install
```

### Step 3: Configure frontend API target

Create or update:

```text
frontend/seller-dashboard/.env.local
```

Only if defaults are not enough:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT_MS=15000
```

### Step 4: Setup CMS MySQL schema

If MySQL is already running from previous tasks, only apply/verify schema:

```bash
mysql -u root -p < database/draw.sql
mysql -u root -p -D cms_db -e "SHOW TABLES LIKE 'coupons';"
mysql -u root -p -D cms_db -e "SHOW TABLES LIKE 'campaigns';"
```

### Step 5: Confirm CMS DB credentials

Check:

```text
backend/services/cms-service/.env
```

Required local values:

```env
CMS_DB_HOST=localhost
CMS_DB_PORT=3306
CMS_DB_NAME=cms_db
CMS_DB_USER=cms_user
CMS_DB_PASSWORD=change-me
```

### Step 6: Align CMS gRPC port

Check both files:

```text
backend/services/cms-service/.env
backend/services/api-gateway/.env
```

Make gateway `CMS_GRPC_ADDR` point to the real CMS gRPC listener.

### Step 7: Ensure seller permissions

Task 4 frontend requires:

```text
offers:view
offers:write
```

Current frontend role mapping grants both permissions to:

```text
seller
seller_manager
seller_catalog_editor
```

`seller_order_manager` does not get offers permissions.

For real backend testing, active seller/session response should include one of these roles or permissions. Also ensure related `seller_staff` row is active if the backend checks staff status.

### Step 8: Start external dependencies

MySQL:

```bash
mysqladmin ping -h 127.0.0.1 -P 3306
```

Redis, only if `RATE_LIMIT_ENABLED=true`:

```bash
redis-cli ping
```

Expected Redis output:

```text
PONG
```

### Step 9: Start backend services

Current blocker: backend source entrypoints were not clearly found.

Expected future flow after backend source exists:

```bash
cd backend/services/cms-service
go run ./cmd/server
```

```bash
cd backend/services/api-gateway
go run ./cmd/server
```

Verify after services exist:

```bash
curl http://localhost:8080/health/live
curl http://localhost:8087/health/live
```

If CMS gRPC is running and reflection is enabled:

```bash
grpcurl -plaintext localhost:9098 list
```

### Step 10: Start Seller Dashboard frontend

```bash
cd frontend
pnpm --filter seller-dashboard dev
```

Open:

```text
http://localhost:5174/seller/offers
```

Useful Task 4 routes:

| Route | Purpose |
|---|---|
| `/seller/offers` | Coupons tab and campaigns tab |
| `/seller/offers?tab=campaigns` | Campaign calendar |
| `/seller/offers/coupons/new` | Create coupon |
| `/seller/offers/coupons/:couponId/edit` | Edit coupon |
| `/seller/offers/campaigns/new` | Create campaign |

## 12. API payload setup notes

Current API contract supports these input fields only.

Coupon input:

```json
{
  "code": "WELCOME10",
  "discount_type": "percentage",
  "discount_value": 10,
  "min_cart_amount": {
    "amount": 50000,
    "currency": "INR"
  },
  "starts_at": "2026-06-19T00:00:00.000Z",
  "ends_at": "2026-06-30T23:59:59.000Z",
  "usage_limit": 100
}
```

Campaign input:

```json
{
  "name": "June Sale",
  "starts_at": "2026-06-19T00:00:00.000Z",
  "ends_at": "2026-06-30T23:59:59.000Z",
  "budget": {
    "amount": 1000000,
    "currency": "INR"
  }
}
```

Not currently supported by the Task 4 frontend/API contract:

```text
coupon delete
campaign edit
campaign delete
dedicated coupon redemption list
buyer cart coupon preview
advanced product/category coupon scope fields
per-user limit UI
max discount cap UI
```

Backend should not require unsupported fields from the frontend unless the API contract and UI are updated together.

## 13. Testing and verification

Frontend checks:

```bash
cd frontend
pnpm --filter seller-dashboard typecheck
pnpm --filter seller-dashboard test
pnpm --filter seller-dashboard build
```

Task 4 relevant frontend files:

```text
frontend/seller-dashboard/src/features/offers/
frontend/seller-dashboard/src/routes/seller-routes.tsx
frontend/seller-dashboard/src/lib/http.ts
frontend/seller-dashboard/src/features/team/utils/seller-permissions.ts
```

Manual API checks after backend exists:

```bash
curl -i http://localhost:8080/api/v1/seller/coupons
curl -i http://localhost:8080/api/v1/seller/campaigns
```

These routes require seller authentication, so unauthenticated local calls may correctly return `401`.

Database checks:

```bash
mysql -u cms_user -p -D cms_db -e "SELECT code, status FROM coupons LIMIT 5;"
mysql -u cms_user -p -D cms_db -e "SELECT name, status, starts_at, ends_at FROM campaigns LIMIT 5;"
```

## 14. Common setup issues and fixes

| Problem | Likely reason | Fix |
|---|---|---|
| Offers page shows permission denied | Active seller role lacks `offers:view` | Use role `seller`, `seller_manager`, or `seller_catalog_editor`; verify backend role response. |
| Create/edit buttons missing | Role lacks `offers:write` | Grant write role or test with owner/manager/catalog editor. |
| Frontend network error | API Gateway not running or wrong `VITE_API_BASE_URL` | Start gateway or update `frontend/seller-dashboard/.env.local`. |
| Gateway returns gRPC unavailable | `CMS_GRPC_ADDR` mismatch or CMS down | Align CMS/Gateway addresses and start CMS Service. |
| Gateway fails on startup | Redis down while rate limiting enabled | Start Redis or set `RATE_LIMIT_ENABLED=false` locally. |
| `Unknown database cms_db` | Schema not applied | Run `mysql -u root -p < database/draw.sql`. |
| `Duplicate coupon code` | `coupons.code` unique index | Use a new code or handle backend validation error cleanly. |
| Campaign create fails for date range | End date before start date or duration too long | Fix dates and check `CMS_CAMPAIGN_MAX_DURATION_DAYS`. |
| Usage stats look incomplete | API does not return `used_count`, `redeemed_count`, or `total_discount` | This is expected; UI derives only available stats. |
| Campaign update/delete missing | API contract has list/create only | Do not wire unsupported UI actions until backend contract adds endpoints. |
| Coupon status buttons not working | No explicit activate/pause/delete endpoint | Only use `PATCH` if backend supports status updates in `CouponInput`; otherwise status is read-only. |

## 15. Security and DevOps notes

- Backend must enforce seller ownership for every coupon and campaign.
- Frontend permissions are UX-only. API Gateway/CMS Service must authorize `offers:view` and `offers:write`.
- `CMS_INTERNAL_AUTH_TOKEN`, DB passwords, JWT secrets, and Redis passwords must never be exposed through `VITE_*` env vars.
- Coupon redemption must be transactional and idempotent when Order Service records final usage.
- `coupon_redemptions` should prevent duplicate redemption for the same coupon/order through its unique key.
- Coupon/campaign changes should create audit logs when backend audit pipeline is available.
- Production MySQL should use backups, migrations, least-privilege users, and TLS/private networking.

## 16. Final checklist

- [x] Task 4 frontend Offers module found.
- [x] Task 4 routes found in `seller-routes.tsx`.
- [x] Coupon/campaign API client found.
- [x] No new frontend package required.
- [x] No `date-fns` dependency required by actual code.
- [x] CMS API contract found in `api/master-api.json`.
- [x] CMS MySQL tables found in `database/draw.sql`.
- [x] `CMS_CAMPAIGN_MAX_DURATION_DAYS` found in CMS env.
- [x] Offer permissions found in frontend role mapping.
- [ ] CMS/API Gateway backend source entrypoints not clearly found.
- [ ] CMS/Gateway `CMS_GRPC_ADDR` mismatch still needs runtime alignment.
- [ ] Dedicated CMS migration files not found.
- [ ] Docker Compose services for full local stack not found.
