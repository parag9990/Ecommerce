# Project Dependency & Setup Guide

## 1. Project Overview

### Variables used by this guide

| Variable | Value / derivation |
|---|---|
| `SERVICE_NAME` | Use the service folder name provided in the prompt |
| `TASK_FILE_NAME` | Use the implementation task file name provided in the prompt |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `${TASK_FILE_NAME}` with `.md` replaced by `_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |

This guide documents dependency, environment, database, Docker, and local setup needs for `INPUT_FILE_PATH`.

Simple Hinglish goal: current task me seller dashboard ke andar read-only audit activity timeline run karna hai. Seller owner/manager recent product, order, coupon, campaign, team, and settings actions dekh sakta hai. Is file ka purpose business logic repeat karna nahi hai; sirf setup, runtime dependencies, env variables, database checks, and DevOps gaps explain karna hai.

### What changed from previous tasks

Most setup is reused. Current task adds audit-specific API expectations and CMS audit guardrails.

| New / Reused | Dependency | Why it matters |
|---|---|---|
| Reused | Frontend React/Vite/pnpm app | Audit page same seller dashboard app me route hota hai. |
| Reused | API Gateway on `8080` | Browser `GET /api/v1/seller/audit-logs` Gateway ko call karta hai. |
| Reused | CMS Service | Audit logs CMS-owned data hain. |
| Reused with task-specific table | MySQL `cms_db` | `cms_audit_logs` table audit timeline ka source hai. |
| Reused | Redis, if Gateway rate limit enabled | Audit endpoint Gateway through jayega, so rate limit config apply ho sakta hai. |
| Reused | Auth/session/seller context | Active seller and role/permission required hai. |
| Task-specific | `audit:view` permission | Audit data sensitive hai, only owner/manager style roles ko view karna chahiye. |
| Task-specific | CMS audit env guardrails | Default range/page size/max page size control karte hain. |

Runtime flow:

```text
Seller browser
  -> Seller Dashboard frontend
  -> RequireSeller session guard
  -> audit:view permission check
  -> API Gateway REST API
  -> CMS Service audit list handler
  -> MySQL cms_db.cms_audit_logs
```

Important boundary:

- Frontend directly MySQL, Redis, or CMS gRPC ko call nahi karta.
- Current frontend implementation calls `GET /api/v1/seller/audit-logs`.
- `api/master-api.json` me dedicated seller audit endpoint clearly listed nahi mila.
- `cms_audit_logs` table already exists in `database/draw.sql`; no new database engine is needed.
- Backend audit writer logic is outside this current frontend-focused dependency doc, but real data ke liye required hai.

## 2. Tech Stack

### Reused stack

Do not repeat base setup. Follow these existing docs first:

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md`

Section:
`Installation steps`, `Required environment variables`, `Start commands`

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/main_dependency.md`

Section:
`All detected dependencies`, `Setup order`, `Ports table`

Refer:
`TaskImplementation/${SERVICE_NAME}/task6_Dependency.md`

Section:
`Tech Stack`, `Environment Variables`, `Ports and networking`

### Current task technology usage

| Technology | Required? | Simple explanation | Why used here | New setup? |
|---|---:|---|---|---|
| React | Yes | React UI components banane ki library hai. | Audit page, timeline, filters, empty/error states render karne ke liye. | No, reused |
| TypeScript | Yes | TypeScript JavaScript me types add karta hai. | Audit response, filters, pagination, and event fields safe rakhne ke liye. | No, reused |
| Vite | Yes | Fast frontend dev server/build tool hai. | Seller dashboard local run/build ke liye. | No, reused |
| React Router | Yes | Browser routes manage karta hai. | `/seller/audit` route activate karne ke liye. | No, reused |
| TanStack React Query | Yes | Server data fetch/cache library hai. | Infinite audit log pagination and refetch ke liye. | No, reused |
| Zustand | Yes | Lightweight frontend store hai. | Active seller context read karne ke liye. | No, reused |
| Lucide React | Yes | Icon library hai. | Audit nav/header/action icons ke liye. | No, reused |
| Tailwind CSS | Yes | Utility-first CSS framework hai. | Dense operational audit timeline style karne ke liye. | No, reused |
| Vitest + Testing Library | Recommended | Frontend tests run karne ke tools hain. | Audit normalizers, labels, formatters, cards test karne ke liye. | No, reused |
| API Gateway | Yes for real data | Browser REST requests receive karta hai. | Seller audit endpoint expose karna hoga. | Reused, route missing/needs confirmation |
| CMS Service | Yes for real data | CMS backend coupons/team/analytics/audit own karta hai. | `cms_audit_logs` query and authorization yahin honi chahiye. | Reused, audit handler missing/needs confirmation |
| MySQL | Yes for real data | Relational DB hai, tables me structured data store karta hai. | `cms_audit_logs` table read karne ke liye. | Reused |
| Redis | Conditional | Fast in-memory store/cache hai. | Gateway rate limiting enabled ho to required. | Reused |

No new frontend package installation is required because audit implementation uses dependencies already present in `frontend/seller-dashboard/package.json`.

## 3. Required Software

Common software setup is already documented. Do not reinstall everything if previous task setup is already done.

| Software | Follow this existing doc | Current task note |
|---|---|---|
| Git | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | Clone/pull repo ke liye. |
| Node.js + Corepack + pnpm | `TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md` | Frontend install/run/test ke liye. |
| Browser | `TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md` | `/seller/audit` manual verification ke liye. |
| Docker / Docker Compose | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | Optional but useful for MySQL/Redis. |
| Go toolchain | `TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md` | Backend services run karne ke liye, once source exists. |
| MySQL client/server | `TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md` | `cms_audit_logs` table verify karne ke liye. |
| Redis | `TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md` | Only if `RATE_LIMIT_ENABLED=true`. |
| curl | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | Gateway/API smoke checks ke liye. |

Beginner note: Agar sirf frontend page open/test karna hai, Node + pnpm enough hai. Agar real audit logs dekhne hain, then Auth/session, API Gateway, CMS Service, MySQL, and maybe Redis running hone chahiye.

## 4. Dependency Management

### Frontend dependencies

The current task uses the existing pnpm workspace.

This setup is already explained in:
`TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md`

Section:
`Installation steps`

Current task dependency status:

| Package | Used by | Current status |
|---|---|---|
| `react` / `react-dom` | Audit page and components | Already installed |
| `react-router-dom` | `/seller/audit` route | Already installed |
| `@tanstack/react-query` | `useSellerAuditLogs` infinite query | Already installed |
| `zustand` | Active seller store | Already installed |
| `lucide-react` | Audit nav and UI icons | Already installed |
| `tailwindcss` / `@tailwindcss/vite` | Styling | Already installed |
| `vitest` | Audit unit tests | Already installed |
| `@testing-library/react` | Component tests | Already installed |

Install only if local `node_modules` is missing:

```bash
cd frontend
pnpm install
```

Task-specific frontend checks:

```bash
cd frontend
pnpm --filter seller-dashboard typecheck
pnpm --filter seller-dashboard test
```

### Backend dependencies

Backend setup remains reused and partly incomplete in this checkout.

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md`

Section:
`Where it is used in project`, `Common errors and fixes`

Current backend reality:

| Finding | Impact |
|---|---|
| `backend/services/cms-service/.env` exists | Audit config values are visible. |
| `backend/services/api-gateway/.env` exists | Gateway ports, Redis, JWT, and CMS target are visible. |
| CMS/Gateway Go source entrypoints were not clearly found | Real audit API cannot be started from this checkout as-is. |
| `api/master-api.json` lists admin audit but not dedicated seller audit endpoint | Gateway contract likely needs update. |

## 5. Database Setup

### Reused database: MySQL

MySQL installation, Docker setup, credentials, and generic troubleshooting are already explained in:

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md`

Section:
`Installation steps`, `Docker setup, if possible`, `Required environment variables`

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/Migrations.md`

Section:
`Local setup without Docker`, `Verify running commands`

### Current task table requirement: `cms_audit_logs`

`cms_audit_logs` table current task ka main data source hai.

Simple Hinglish: ye table seller dashboard actions ka immutable trail store karti hai. Product update, coupon change, shipment update, team role change jaise events yahan append-only rows ke form me store hone chahiye.

Evidence:

| Source | What it confirms |
|---|---|
| `database/draw.sql` | `cms_audit_logs` table exists with audit fields and indexes. |
| `docs/04-microservice-design.md` | CMS Service owns audit logs. |
| `docs/05-database-design.md` | CMS DB includes audit logs. |
| `frontend/seller-dashboard/src/features/audit/api/seller-audit-api.ts` | Frontend expects audit log fields and pagination. |

Required or optional:

- Optional only for mocked/frontend-only UI tests.
- Required for real audit timeline data.

Default MySQL port:

```text
3306
```

Credentials placement:

| File | Variables |
|---|---|
| `backend/services/cms-service/.env` | `CMS_MYSQL_DSN` or `CMS_DB_HOST`, `CMS_DB_PORT`, `CMS_DB_NAME`, `CMS_DB_USER`, `CMS_DB_PASSWORD` |

Table columns expected from schema:

| Column | Purpose |
|---|---|
| `audit_id` | Public unique id for timeline event. |
| `seller_id` | Seller scope. Backend must never return another seller's rows. |
| `actor_user_id` | User who performed the action. |
| `action` | Event name, for example `product.updated`. |
| `resource_type` | Resource category, for example `product`, `order`, `team`. |
| `resource_id` | Resource id affected by the action. |
| `before_json` | Optional previous values. |
| `after_json` | Optional new values. |
| `created_at` | Event timestamp. |

Important indexes:

| Index | Why useful |
|---|---|
| `uk_cms_audit_logs_audit_id` | Prevents duplicate audit ids. |
| `idx_cms_audit_seller_created` | Fast recent seller timeline query. |
| `idx_cms_audit_resource` | Fast resource-specific history query. |

Task-specific DB verification:

```bash
mysql -u cms_user -p -D cms_db -e "SHOW TABLES LIKE 'cms_audit_logs';"
mysql -u cms_user -p -D cms_db -e "SHOW INDEX FROM cms_audit_logs;"
mysql -u cms_user -p -D cms_db -e "SELECT audit_id, seller_id, actor_user_id, action, resource_type, resource_id, created_at FROM cms_audit_logs ORDER BY created_at DESC LIMIT 20;"
```

If table is missing locally:

```bash
mysql -u root -p < database/draw.sql
```

Production note: `database/draw.sql` is useful for local bootstrap, but production should use versioned up/down migrations.

## 6. Redis / Queue / External Services

### Redis

Redis current task ka direct data source nahi hai. Redis only tab required hai jab API Gateway rate limiting enabled ho.

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md`

Section:
`Installation steps`, `Start commands`, `Common errors and fixes`

Current Gateway env has:

```text
RATE_LIMIT_ENABLED=true
REDIS_ADDR=localhost:6379
RATE_LIMIT_FAIL_OPEN=false
```

Beginner note: Agar Redis down hai and rate limit fail closed hai, audit API request Gateway par fail ho sakti hai even though audit page ka data MySQL me hota hai.

### API Gateway

API Gateway required hai because frontend `http.ts` browser requests ko `VITE_API_BASE_URL` par bhejta hai. Default base URL `http://localhost:8080` hai.

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md`

Section:
`Required environment variables`, `Common errors and fixes`

Current task expected route:

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/api/v1/seller/audit-logs` | Seller-scoped audit timeline |

Supported query params expected by frontend:

| Query param | Example | Purpose |
|---|---|---|
| `page` | `1` | Page number for pagination. |
| `page_size` | `20` | Rows per page. Allowed UI values are `20`, `50`, `100`. |
| `actor_id` | `user_123` | Filter by actor user id. |
| `action` | `product.updated` | Filter by action name. |
| `resource_type` | `product` | Filter by resource type. |
| `resource_id` | `prod_123` | Filter by resource id. |
| `from` | `2026-06-01T00:00:00.000Z` | Inclusive start timestamp. |
| `to` | `2026-06-30T23:59:59.999Z` | Inclusive end timestamp. |

Current gap: this route was not clearly found in `api/master-api.json`. Gateway contract and route registration must be added or confirmed before real API verification can pass.

### CMS Service

CMS Service required hai for real audit logs.

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/CMS_Backend.md`

Section:
`Why this service uses it`, `Required environment variables`, `Common errors and fixes`

Current task CMS responsibilities:

- Validate authenticated seller context.
- Enforce `audit:view` server-side.
- Query only current seller's `cms_audit_logs`.
- Apply filters safely.
- Cap page size at configured max.
- Return stable response shape expected by frontend.
- Avoid exposing sensitive raw data in `before_json` / `after_json`.

### Queues, object storage, payment providers

No new Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, S3, SMTP, Stripe, Twilio, Firebase, or OAuth provider requirement was detected for the current audit viewer.

Audit writers in future backend tasks may publish or consume events, but current page only reads CMS audit logs.

## 7. Environment Variables

### Reused env docs

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/Environment.md`

Section:
`Required environment variables`, `Common errors and fixes`, `Security notes`

### Frontend env

No new frontend env variable is required.

Current task reuses:

```text
frontend/seller-dashboard/.env.local
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT_MS=15000
```

How frontend loads env:

- Vite exposes only `VITE_*` variables to browser code.
- `frontend/seller-dashboard/src/lib/http.ts` reads `VITE_API_BASE_URL` and `VITE_API_TIMEOUT_MS`.
- If missing, frontend defaults to `http://localhost:8080` and `15000` ms timeout.

Security note: `VITE_*` values browser bundle me visible hote hain. DB password, JWT private key, internal auth token, Redis password, or audit secrets yahan kabhi mat rakho.

### Task-specific CMS env values

These audit guardrails already exist in `backend/services/cms-service/.env` and should be documented/validated:

```text
CMS_AUDIT_DEFAULT_RANGE_DAYS=30
CMS_AUDIT_DEFAULT_PAGE_SIZE=20
CMS_AUDIT_MAX_PAGE_SIZE=100
```

| Variable | Required? | Purpose | Example | Security note |
|---|---:|---|---|---|
| `CMS_AUDIT_DEFAULT_RANGE_DAYS` | Yes for CMS audit API | If no date filter is supplied, backend can default to recent days. | `30` | Not secret. Keep bounded for performance. |
| `CMS_AUDIT_DEFAULT_PAGE_SIZE` | Yes for CMS audit API | Default timeline page size. | `20` | Not secret. Should match frontend default. |
| `CMS_AUDIT_MAX_PAGE_SIZE` | Yes for CMS audit API | Prevents huge audit queries. | `100` | Not secret. Do not set unbounded. |

### Reused CMS/Gateway env values

| File | Important reused variables |
|---|---|
| `backend/services/cms-service/.env` | `CMS_GRPC_ADDR`, `CMS_MYSQL_DSN`, `CMS_DB_*`, `CMS_INTERNAL_AUTH_TOKEN`, audit guardrails |
| `backend/services/api-gateway/.env` | `HTTP_ADDR`, `API_BASE_PATH`, `CMS_GRPC_ADDR`, `JWT_JWKS_URL`, `REDIS_ADDR`, rate limit vars |

Important current mismatch:

| File | Value |
|---|---|
| `backend/services/cms-service/.env` | `CMS_GRPC_ADDR=:9098` |
| `backend/services/api-gateway/.env` | `CMS_GRPC_ADDR=localhost:50059` |

Fix one side before testing audit APIs:

```text
# Option A: keep CMS service on 9098 and update gateway target
CMS_GRPC_ADDR=localhost:9098

# Option B: keep gateway target 50059 and make CMS listen on 50059
CMS_GRPC_ADDR=:50059
```

## 8. Docker Setup

Current task adds no new Docker container, volume, network, health check, or restart policy.

Reused Docker/DevOps setup:

Refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/main_dependency.md`

Section:
`Docker setup`

Refer:
`TaskImplementation/${SERVICE_NAME}/task5_Dependency.md`

Section:
`Docker Setup`

Task runtime services:

| Container / service | Need | Status |
|---|---|---|
| Seller Dashboard frontend | Run browser UI | Reused |
| API Gateway | Expose `/api/v1/seller/audit-logs` | Reused, route missing/needs confirmation |
| CMS Service | Query `cms_audit_logs` | Reused, audit handler missing/needs confirmation |
| Auth/User/Session boundary | Active seller and roles | Reused |
| MySQL | Store CMS audit rows | Reused |
| Redis | Gateway rate limit if enabled | Reused |

No Docker command is repeated here. Follow the previous Docker guidance first, then verify the current audit route after backend is available.

Docker networking note:

- Browser-to-Gateway uses host URL like `http://localhost:8080`.
- Container-to-container calls should use compose service names, for example `cms-service:9098`, `mysql:3306`, and `redis:6379`.
- `localhost` inside a container means that container itself, not your host machine.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation first

Follow these in order:

1. `TaskImplementation/${SERVICE_NAME}/Dependency/main_dependency.md`
2. `TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md`
3. `TaskImplementation/${SERVICE_NAME}/Dependency/Environment.md`
4. `TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md`
5. `TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md`
6. `TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md`
7. `TaskImplementation/${SERVICE_NAME}/Dependency/CMS_Backend.md`
8. `TaskImplementation/${SERVICE_NAME}/task6_Dependency.md`

### Step 2: Go to project directory

Documentation for this task lives here:

```bash
cd "TaskImplementation/${SERVICE_NAME}"
```

Frontend app runs from:

```bash
cd frontend
```

### Step 3: Install only new dependencies if any

No new dependency is required for the current task.

If dependencies are not installed locally:

```bash
cd frontend
pnpm install
```

### Step 4: Setup databases/services

Reused services:

- MySQL with CMS schema from `database/draw.sql`
- Redis if Gateway rate limiting is enabled
- Auth/session/seller context backend
- CMS Service
- API Gateway

Current task DB check:

```bash
mysql -u cms_user -p -D cms_db -e "SHOW TABLES LIKE 'cms_audit_logs';"
```

### Step 5: Add only new or changed environment variables

Frontend has no new env variables.

Confirm reused frontend env:

```text
frontend/seller-dashboard/.env.local
VITE_API_BASE_URL=http://localhost:8080
VITE_API_TIMEOUT_MS=15000
```

Confirm CMS audit guardrails:

```text
backend/services/cms-service/.env
CMS_AUDIT_DEFAULT_RANGE_DAYS=30
CMS_AUDIT_DEFAULT_PAGE_SIZE=20
CMS_AUDIT_MAX_PAGE_SIZE=100
```

Then align Gateway and CMS `CMS_GRPC_ADDR` values as explained in Section 7.

### Step 6: Run migrations if needed

Dedicated CMS migrations were not clearly found. Local fallback:

```bash
mysql -u root -p < database/draw.sql
```

Production-ready fix: create proper CMS up/down migrations for `cms_audit_logs` and related CMS tables.

### Step 7: Start backend services

Current backend source is incomplete in this checkout. Once backend code exists, expected startup order:

1. MySQL
2. Redis if `RATE_LIMIT_ENABLED=true`
3. Auth/User/Session service
4. CMS Service
5. API Gateway
6. Seller Dashboard frontend

Backend smoke checks after implementation:

```bash
curl http://localhost:8080/health/live
curl "http://localhost:8080/api/v1/seller/audit-logs?page=1&page_size=20"
curl "http://localhost:8080/api/v1/seller/audit-logs?resource_type=product&page=1&page_size=20"
```

Note: Real request may need valid auth cookies/session. A plain `curl` can return `401`, which is correct if unauthenticated.

### Step 8: Start Seller Dashboard frontend

```bash
cd frontend
pnpm --filter seller-dashboard dev
```

Open:

```text
http://localhost:5174/seller/audit
```

### Step 9: Verify current task functionality

Manual checks:

- Login/session returns active seller.
- Active seller role has `audit:view`.
- `/seller/audit` route renders inside dashboard layout.
- Audit nav item points to `/seller/audit`.
- Empty state appears when no logs exist.
- Permission denied state appears when role lacks `audit:view`.
- Filters send `actor_id`, `action`, `resource_type`, `resource_id`, `from`, `to`.
- Rows selector sends page size `20`, `50`, or `100`.
- Load more uses next page when `pagination.has_next=true`.
- API response with `before_json` / `after_json` aliases normalizes correctly.

Automated checks:

```bash
cd frontend
pnpm --filter seller-dashboard typecheck
pnpm --filter seller-dashboard test
```

## 10. Running the Project

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Seller Dashboard Vite dev server | `5174` | Browser UI | Reused |
| Seller Dashboard preview | `4174` | Built frontend preview | Reused |
| API Gateway HTTP | `8080` | Frontend REST entrypoint | Reused |
| CMS Service HTTP | `8087` | CMS health/HTTP if implemented | Reused |
| CMS Service gRPC | `9098` | CMS service listen address | Reused, must align |
| Gateway CMS target | `50059` currently | Gateway-to-CMS target | Mismatch risk |
| MySQL | `3306` | CMS database and `cms_audit_logs` | Reused |
| Redis | `6379` | Gateway rate limiting | Reused if enabled |

Networking notes:

- Browser calls `VITE_API_BASE_URL`, default `http://localhost:8080`.
- Frontend HTTP client uses `credentials: include`, so cookies and CORS must be configured correctly.
- If `8080` is busy, change Gateway `HTTP_ADDR` and frontend `VITE_API_BASE_URL` together.
- If MySQL `3306` is busy, reuse existing MySQL or remap Docker host port and update `CMS_DB_PORT`.
- If services run inside Docker Compose later, use Docker service names for internal calls.

### Expected API response shape

Frontend can normalize `logs`, `audit_logs`, or `items`, but backend should prefer one stable contract.

Recommended response:

```json
{
  "data": {
    "logs": [
      {
        "audit_id": "audit_123",
        "seller_id": "seller_456",
        "actor_user_id": "user_789",
        "actor_name": "Catalog Manager",
        "actor_email": "catalog@example.com",
        "action": "product.updated",
        "resource_type": "product",
        "resource_id": "prod_101",
        "resource_title": "Cotton Shirt",
        "before_json": {"status": "draft"},
        "after_json": {"status": "submitted"},
        "created_at": "2026-06-01T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 1,
      "has_next": false
    }
  }
}
```

Required fields per log:

```text
audit_id
actor_user_id
action
resource_type
resource_id
created_at
```

Optional but useful fields:

```text
seller_id
actor_name
actor_email
resource_title
before or before_json
after or after_json
```

## 11. Common Errors & Fixes

Generic pnpm, MySQL, Redis, Docker, and Go module errors are already documented in previous dependency files. Current task-specific issues are below.

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `/seller/audit` redirects to login | Seller session API returned 401 or session service unavailable | Login again, start auth/session backend, verify cookies | Start auth/session before frontend manual testing |
| Audit page shows permission denied | Active role lacks `audit:view` | Use owner/manager role or fix backend role response | Keep Task 6 role matrix and backend RBAC aligned |
| Audit list returns 404 | `/api/v1/seller/audit-logs` not registered in Gateway/API contract | Add Gateway route and CMS handler | Keep frontend-used routes in `api/master-api.json` |
| Audit list returns 403 | Backend denies missing/invalid `audit:view` | Verify seller role and permission claims | Enforce permissions server-side with contract tests |
| Timeline empty but DB has rows | Wrong `seller_id` scope, date range default, or filters too narrow | Check active seller id, `CMS_AUDIT_DEFAULT_RANGE_DAYS`, and filters | Log request id and query filters without leaking secrets |
| Filters do nothing | Backend ignores query params | Implement `actor_id`, `action`, `resource_type`, `resource_id`, `from`, `to` filters | Add API tests for every filter |
| Load more repeats same rows | Backend pagination always returns same page or bad `has_next` | Return correct `page`, `page_size`, `total`, `has_next` | Add pagination tests |
| Response rows disappear in UI | Backend missing required string fields | Return required fields listed in Section 10 | Add response contract tests |
| Gateway returns CMS unavailable | CMS down or `CMS_GRPC_ADDR` mismatch | Align Gateway and CMS gRPC addresses | Keep one port map in docs/env examples |
| MySQL table missing | CMS schema not applied | Apply `database/draw.sql` locally or run migrations | Add versioned migrations |
| Page size too large / slow query | Backend accepts unbounded `page_size` | Enforce `CMS_AUDIT_MAX_PAGE_SIZE=100` | Never allow unlimited audit reads |
| Sensitive data visible in diff | Backend stores raw secrets/PII in `before_json` or `after_json` | Redact audit payload before insert | Maintain audit redaction policy |

## 12. Security & Best Practices

Current task-specific security checklist:

- Frontend permission check is UX only. Backend must enforce `audit:view`.
- Every audit query must be scoped by authenticated seller id. Never trust `seller_id` from query params.
- Audit logs should be append-only. Do not build edit/delete audit APIs for normal users.
- Do not store raw passwords, tokens, OTPs, payment secrets, private keys, or full session cookies in `before_json` / `after_json`.
- Prefer redacted diffs over full request payloads.
- Keep page size capped with `CMS_AUDIT_MAX_PAGE_SIZE`.
- Query newest rows using `(seller_id, created_at)` index.
- Include request id in logs so frontend errors can be traced to backend logs.
- Backend audit writers should write audit rows in the same transaction as critical mutations where possible.
- Empty state is acceptable when no rows exist; do not fake production audit events.
- Define retention/archival policy before production because audit tables grow continuously.

## 13. Missing or Misconfigured Things

| Area | Finding | Impact | Suggested fix |
|---|---|---|---|
| API contract | Dedicated seller audit endpoint not clearly listed in `api/master-api.json` | Frontend may call a route Gateway does not expose | Add `/api/v1/seller/audit-logs` contract entry |
| Gateway route | Seller audit route registration not clearly found | Real API request may return 404 | Add Gateway REST route to CMS audit handler |
| CMS handler | CMS audit list handler/source not clearly found | Real audit data cannot load | Implement list handler/usecase/repository |
| Backend source | CMS/Gateway Go source and service `go.mod` files not clearly found | Backend cannot be run from this checkout as-is | Add service modules and entrypoints |
| gRPC port alignment | CMS listens on `:9098`, Gateway points to `localhost:50059` | Gateway cannot reach CMS | Align `CMS_GRPC_ADDR` values |
| Migrations | Dedicated CMS up/down migrations not found | Local schema bootstrap is manual and rollback unsafe | Add versioned CMS migrations |
| `.env.example` | Sanitized env examples not clearly found | Beginners may copy real `.env` values | Add `.env.example` files with placeholders |
| Docker Compose | No full local compose file clearly found | Beginners must start services manually | Add local compose for MySQL, Redis, CMS, Gateway, frontend |
| Audit writer | Product/order/offers/team mutation audit writes are outside current scope | Timeline can be empty even when UI works | Add backend audit write middleware/usecases |
| Data redaction | Audit payload redaction policy not clearly found | Secrets/PII could leak in timeline | Add redaction allowlist/denylist before inserts |
| Health checks | Gateway/CMS health endpoints not clearly confirmed | Harder local debugging | Add `/health/live` and `/health/ready` |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | Required Software, frontend setup, MySQL, Redis, API Gateway, env basics | Current task runs inside same seller dashboard and infra pattern. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | Product action context, frontend dependency reuse, API Gateway pattern | Product changes can appear in audit timeline. |
| `TaskImplementation/${SERVICE_NAME}/task3_Dependency.md` | Ports, frontend env, Gateway verification, permission/network errors | Order shipment/status changes can appear in audit timeline. |
| `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` | CMS Service, MySQL CMS tables, CMS/Gateway gRPC alignment, Docker gaps | Coupon/campaign actions are CMS-backed audit sources. |
| `TaskImplementation/${SERVICE_NAME}/task5_Dependency.md` | Analytics-era CMS env, Docker notes, ports, missing backend status | Same CMS/Gateway setup state and no new containers. |
| `TaskImplementation/${SERVICE_NAME}/task6_Dependency.md` | `audit:view` permission, role matrix, team action audit linkage, env/port reuse | Current task depends on Task 6 permission model. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/main_dependency.md` | Overall setup order, detected dependencies, ports table, missing implementation audit | Central dependency summary. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md` | pnpm install, Vite ports, frontend env, frontend commands | No new frontend install process needed. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Environment.md` | Env files, required vars, security notes | Current task only adds/uses CMS audit guardrails. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md` | MySQL install/Docker/credentials | `cms_audit_logs` reuses CMS MySQL. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Migrations.md` | Schema bootstrap and migration gaps | Audit table exists in schema but needs proper migrations. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md` | Gateway env, health checks, common errors | Audit API must go through Gateway. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/CMS_Backend.md` | CMS backend purpose, env, missing source status | Audit logs are CMS-owned. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md` | Redis setup | Gateway rate limiting may require Redis. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/gRPC.md` | gRPC env and port mismatch troubleshooting | Gateway-to-CMS calls require gRPC alignment. |

## 15. Final Checklist

- [x] Previous dependency documentation checked.
- [x] `INPUT_FILE_PATH` analyzed.
- [x] Existing frontend audit files inspected.
- [x] Previous setup duplication avoided.
- [x] Original implementation file not modified.
- [x] No business logic rewritten.
- [x] New package requirement checked.
- [x] No new frontend env variable required.
- [x] CMS audit env guardrails documented.
- [x] MySQL `cms_audit_logs` requirement documented.
- [x] Redis/Gateway/CMS/Auth dependencies marked as reused.
- [x] Ports and networking notes added.
- [x] Missing seller audit API contract documented.
- [x] CMS/Gateway gRPC mismatch documented.
- [x] Security and audit redaction notes added.
- [x] References to previous dependency files included.
- [ ] Backend `/api/v1/seller/audit-logs` route implemented and verified.
- [ ] CMS audit repository/usecase/handler implemented.
- [ ] Backend enforces `audit:view` and seller scoping.
- [ ] Audit writers added for product/order/offers/team/settings mutations.
- [ ] CMS migrations added with up/down files.
- [ ] Sanitized `.env.example` files added.
- [ ] Full local Docker Compose added, if desired.
- [ ] `pnpm --filter seller-dashboard typecheck` run successfully.
- [ ] `pnpm --filter seller-dashboard test` run successfully.
- [ ] Manual `/seller/audit` browser verification completed with real or intentionally empty data.
