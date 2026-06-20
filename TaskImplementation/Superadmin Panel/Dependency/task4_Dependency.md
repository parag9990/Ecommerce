# Superadmin Panel Task 4 - Dependency & Setup Guide

## 1. Project Overview

Yeh guide `task4.md` ke **Order Operations** module ko locally validate aur eventually full stack me run karne ke liye dependency, environment, database, external-service, networking, security, aur troubleshooting requirements explain karti hai. Business logic aur implementation snippets yahan repeat nahi kiye gaye hain.

Actual frontend implementation yahan present hai:

```text
frontend/superadmin-panel/src/features/orders/
frontend/superadmin-panel/tests/orders/
```

Module me admin order search/filter, order detail, item and shipment view, status timeline, read-only payment summary, disputes, manual review actions, permission checks, aur automated tests included hain.

### Current repository reality

| Capability | Current status | Beginner meaning |
|---|---|---|
| Task 4 frontend source | Ready | Routes, UI, hooks, API client, permissions, and tests present hain |
| Frontend install/test/build | Ready | Existing pnpm workspace se run hota hai |
| New Task 4 npm dependency | None | Koi package add karne ki zarurat nahi hai |
| New frontend environment variable | None | Existing `.env.example` reuse hota hai |
| Admin order list API | Contract only | `GET /api/v1/admin/orders` master API file me hai, but runnable Order Service absent hai |
| Order detail API | Missing contract/backend | Frontend route use karta hai, master API file define nahi karti |
| Dispute API/storage | Missing | Frontend path and types present hain; API contract, owner, table, and migration absent hain |
| Manual review API | Missing contract/backend | UI mutation present hai; approved server contract/implementation absent hai |
| Order/Payment backend runtime | Missing | Order and Payment service folders/source/config nahi mile |
| Docker/Compose stack | Missing | Supported Dockerfile or Compose manifest nahi mila |

> **Important:** Task 4 tests, TypeScript check, production build, aur frontend dev server backend ke bina run ho sakte hain. Real order data aur review actions tabhi work karenge jab Order/Superadmin dependencies, contracts, migrations, and runtime complete hon.

### Task-specific API dependency

| Method | Frontend path | Contract status | Expected owner/purpose |
|---|---|---|---|
| `GET` | `/api/v1/admin/orders` | Present | Order Service; list and supported filters |
| `GET` | `/api/v1/admin/orders/{order_id}` | Missing | Admin-safe detail, items, history, shipments, and payment summary |
| `GET` | `/api/v1/admin/orders/{order_id}/disputes` | Missing | Approved dispute owner; dispute investigation data |
| `POST` | `/api/v1/admin/orders/{order_id}/review` | Missing | Superadmin workflow; review decision plus immutable audit |

All requests ka safe flow hona chahiye:

```text
Browser -> API Gateway -> authorized owning service -> private database/downstream service
```

Browser ko MySQL, Redis, gRPC, ya an internal service directly call nahi karna chahiye.

## 2. Tech Stack

Task 4 same Superadmin frontend stack reuse karta hai. Full beginner definitions and installation steps dobara repeat nahi kiye gaye hain.

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Sections:
`2. Tech Stack`, `3. Required Software`, and `4. Dependency Management`

Task-specific use:

| Technology/service | Task 4 me use | Required? | New? |
|---|---|---:|---:|
| React 19 + TypeScript | Order pages, filters, timeline, disputes, review dialog | Yes | No |
| React Router | `/admin/orders` and `/admin/orders/:orderId` | Yes | No |
| TanStack React Query | List/detail/dispute queries and review mutation invalidation | Yes | No |
| Zustand | Admin token and roles | Yes | No |
| Vite + Tailwind CSS | Local dev, build, and styling | Yes | No |
| Vitest + Testing Library + jsdom | Order API, filter, permission, and review tests | Tests only | No |
| Lucide React | Order, shipment, payment, and review icons | Yes | No |
| API Gateway | Authenticated REST boundary | Real integration | No |
| MySQL | Transactional order data plus admin review/audit data | Real integration | New Task 4 schema usage |
| Redis | Gateway rate limiting under current config | Real integration | No |
| Order Service | Order list/detail/history/shipment owner | Real integration | New Task 4 service dependency |
| Superadmin Service | Manual-review task and audit owner | Review integration | New Task 4 workflow dependency |
| Payment Service | Read-only payment/refund summary, if not projected by Order Service | Conditional | New Task 4 read dependency |

**Simple Hinglish:** MySQL ek relational database hai jo transactional order records ke liye suitable hai. Order Service order lifecycle own karega; Superadmin Service admin review/audit workflow own karega. Frontend sirf Gateway se JSON APIs call karega.

### Planning guide versus actual manifest

`task4.md` package-add examples me `clsx`, MSW, `@testing-library/user-event`, aur `@testing-library/jest-dom` mention karta hai. Actual implementation:

- `clsx` import nahi karti; project-local `cn` helper use hota hai.
- MSW import nahi karti; Vitest fetch mocks use hote hain.
- `user-event` and `jest-dom` current Task 4 tests ke liye required nahi hain.
- Required React, Query, Router, Zustand, Lucide, Vitest, and Testing Library packages already manifest/lockfile me hain.

> `task4.md` ke package-add commands current implementation ke liye mat run karein. Actual `package.json`, imports, and `pnpm-lock.yaml` source of truth hain.

## 3. Required Software

Git, Node.js 22.12+, pnpm 10, and browser setup unchanged hai.

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`3. Required Software`

Validation mode ke hisaab se requirements:

| Mode | Required software/services |
|---|---|
| Frontend tests/typecheck/build | Node.js, pnpm, existing locked packages |
| Frontend browser with mocked/unavailable API states | Above plus a modern browser |
| Real order list/detail | API Gateway, Auth/JWKS, Order Service, MySQL `order_db`, Redis for current Gateway config |
| Real manual review | Above plus Superadmin Service and MySQL `superadmin_db` |
| Read-only payment summary | Order projection, or a working Payment Service integration chosen by backend owners |

No runnable Order Service, Payment Service, Gateway, Auth, or Superadmin implementation was found for this flow. Therefore exact backend binary/build/start commands cannot honestly be supplied yet.

## 4. Dependency Management

### New dependency delta

| Check | Result |
|---|---|
| New runtime npm package | None |
| New development npm package | None |
| `package.json` update needed | No |
| `pnpm-lock.yaml` update needed | No |
| New Go module available | No; Order/Payment source is absent and other required backend folders are env-only |

Existing pnpm explanation, clean install, scripts, cache/proxy issues, and Node version troubleshooting are already documented:

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`4. Dependency Management`

Correct locked install remains:

```bash
cd frontend
pnpm install --frozen-lockfile
```

Beginner rules:

- npm, yarn, and pnpm lockfiles mix mat karein.
- `node_modules/` commit mat karein.
- Random package-add commands run karke a passing lockfile change mat karein.
- Empty/missing backend service path me `go mod tidy`, `go build`, or `go run` mat chalayein.
- Backend Go workspace limitations ke liye `TaskImplementation/Superadmin Panel/Dependency/Go_Modules.md` refer karein.

## 5. Database Setup

Frontend directly kisi database ko access nahi karta. Database credentials browser `.env` me rakhna security issue hai.

### Task 4 data ownership

```text
Browser
  -> API Gateway
     -> Order Service       -> MySQL order_db
     -> Superadmin Service  -> MySQL superadmin_db
        -> Order Service    -> approved order snapshot/state transition
        -> Payment Service  -> optional read-only payment summary
```

The exact dispute owner and Superadmin-to-Order orchestration are not yet defined by the API contract. Backend team ko one explicit ownership flow approve karna hoga; services ko each other's tables directly query nahi karne chahiye.

| Store | Task 4 data | Required when | Default port | Current status |
|---|---|---|---:|---|
| MySQL `order_db` | `orders`, items, status history, shipments | Real list/detail | `3306` | Bootstrap schema present; service/migrations absent |
| MySQL `superadmin_db` | `admin_review_tasks`, `admin_audit_logs` | Manual-review mutation | `3306` | Bootstrap schema present; service/migrations absent |
| Dispute store | Dispute type/status/summary/opened-by | Dispute panel | TBD | No table, collection, owner, or migration found |

### MySQL setup reuse

MySQL kya hai, local installation, Docker, persistent volume, credentials, start, verification, security, and generic errors already documented hain.

Refer:
`TaskImplementation/Superadmin Panel/Dependency/MySQL.md`

Sections:
`5. Installation steps` through `12. Security notes`

Task-specific schema verification, only after an approved local migration/bootstrap has been applied:

```bash
mysql -u root -p -e "USE order_db; SHOW TABLES LIKE 'orders'; SHOW TABLES LIKE 'order_items'; SHOW TABLES LIKE 'order_status_history'; SHOW TABLES LIKE 'shipments';"
mysql -u root -p -e "USE superadmin_db; SHOW TABLES LIKE 'admin_review_tasks'; SHOW TABLES LIKE 'admin_audit_logs';"
```

Expected: all six listed tables return.

Useful index checks:

```bash
mysql -u root -p -e "USE order_db; SHOW INDEX FROM orders; SHOW INDEX FROM order_items; SHOW INDEX FROM order_status_history; SHOW INDEX FROM shipments;"
mysql -u root -p -e "USE superadmin_db; SHOW INDEX FROM admin_review_tasks; SHOW INDEX FROM admin_audit_logs;"
```

Task 4 filters need reviewed indexes. Existing schema covers order status/date, user/date, seller/item/date, and review resource lookup. Backend owners should query-plan-test any new `q`, review-status, and date-range search instead of adding unreviewed indexes blindly.

### Migration and schema gaps

- Versioned Task 4 up/down migrations and a migration runner are absent.
- `database/draw.sql` is repository-wide bootstrap SQL, not a safe production migration command.
- No dispute table/collection exists.
- Frontend review states are `none`, `disputed`, `manual_review`, and `resolved`.
- `admin_review_tasks.status` currently allows `open`, `approved`, `rejected`, and `cancelled`; an explicit mapping/versioned schema is required.
- `admin_review_tasks` has `reason` and JSON `metadata`, but no dedicated `internal_note` field. The API must define whether sensitive internal notes belong in protected metadata or a new audited column.
- Monetary DB fields use integer minor units. Frontend `formatMoney` divides `amount` by 100, so API contracts must clearly guarantee minor-unit integers and ISO currency codes.

Migration limitations and safe practices already documented hain:
`TaskImplementation/Superadmin Panel/Dependency/Migrations.md`

## 6. Redis, Queue, and External Services

### API Gateway, Auth, and Redis

Gateway/JWT/JWKS/CORS/Redis setup is reused. Full installation and health guidance:

- `TaskImplementation/Superadmin Panel/Dependency/API_Gateway.md`
- `TaskImplementation/Superadmin Panel/Dependency/Redis.md`
- `TaskImplementation/Superadmin Panel/task2_Dependency.md`, sections `6` and `7`

Task 4 specifically expects:

- Browser REST on Gateway port `8080`.
- JWT/JWKS validation before any order response.
- Gateway `ORDER_GRPC_ADDR=localhost:50056` for the current admin list contract.
- Gateway Redis rate limiting on `localhost:6379`; current `RATE_LIMIT_FAIL_OPEN=false` means Redis can be a mandatory Gateway dependency.
- Server-side permissions for order view, dispute view, and review mutation. React role checks are only UX.

### Order Service

Order Service real Task 4 data ke liye mandatory hai. Design says it owns transactional order lifecycle in MySQL, but repository me `backend/services/order-service/`, its config loader, module, migrations, health check, and runnable server nahi mile.

Two unaligned integration clues exist:

| Caller | Detected target | Meaning |
|---|---|---|
| API Gateway | `ORDER_GRPC_ADDR=localhost:50056` | Current admin list contract routes to Order gRPC |
| Superadmin env | `ORDER_SERVICE_ADMIN_BASE_URL=http://127.0.0.1:8084/internal/admin` | Planned admin workflow uses Order internal HTTP |

Backend owner must define which calls use gRPC vs internal HTTP, implement auth between services, and publish exact liveness/readiness endpoints.

### Superadmin Service

Manual review ke liye expected dependency hai because review tasks and immutable audit logs Superadmin domain me belong karte hain. Current folder only `.env` provide karta hai; runnable code and gRPC listener are absent.

`SUPERADMIN_REQUIRE_DATABASE=false` and `SUPERADMIN_REQUIRE_ORDER_SERVICE=false` current local defaults hain. Production review mutation ko dependencies unavailable hone par fail closed karna chahiye; audit/order writes silently skip nahi hone chahiye.

### Payment summary

Task 4 payment section read-only hai. Refund approve/reject Task 5 scope hai.

Backend should choose one safe design:

1. Order detail returns a minimal payment projection, or
2. an authorized backend aggregator calls Payment Service.

Frontend ko Payment Service directly call nahi karna chahiye. Response me provider client secret, card data, webhook secret, or raw provider payload kabhi return na karein.

### Queues and other services

Task 4 frontend does not import or connect to Kafka, RabbitMQ, NATS, MongoDB, Elasticsearch, object storage, SMTP, or a payment provider SDK. No new queue installation is justified by the inspected implementation.

Future Order/Payment event consistency may use a broker/outbox, but provider and configuration are not defined for the missing Order Service. Is guide me speculative broker setup intentionally add nahi kiya gaya hai.

## 7. Environment Variables

### Frontend delta

Task 4 introduces **no new or changed frontend variable**.

Existing `.env` creation, purpose, Vite loading, and secret rules:

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`7. Environment Variables`

Never add MySQL DSNs, Redis passwords, JWT keys, internal-service tokens, payment credentials, or admin secrets to a `VITE_*` variable. Every `VITE_*` value browser JavaScript me visible hota hai.

### Existing API base-path blocker

Task 2/3 me documented shared defect Task 4 ko bhi affect karta hai:

Refer:
`TaskImplementation/Superadmin Panel/task2_Dependency.md`

Section:
`7. Environment Variables` -> `Critical base-path mismatch`

Existing template ends with `/api/v1`:

```dotenv
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

Task 4 `ORDERS_PATH` also starts with `/api/v1/admin/orders`, so real URL becomes:

```text
http://localhost:8080/api/v1/api/v1/admin/orders
```

Changing only `.env` to the host origin would fix orders but break current login path composition. Permanent fix: one shared convention choose karein—recommended base at `/api/v1`, feature paths at `/admin/...`—then login plus every module ko production-like URL tests me cover karein.

### Backend-only detected configuration

Task 4 ke liye no complete new backend `.env` can be generated because Order/Payment config loaders do not exist. Variable names invent karna unsafe documentation hoga.

Detected existing configuration:

| Owner | Variable/group | Purpose/status |
|---|---|---|
| Gateway | `ORDER_GRPC_ADDR`, `SUPERADMIN_GRPC_ADDR`, `AUTH_GRPC_ADDR`, `JWT_*`, `REDIS_*` | Routing/auth/rate limit; runtime absent |
| Superadmin | `SUPERADMIN_DATABASE_DSN` | Review/audit MySQL; backend-only secret |
| Superadmin | `ORDER_SERVICE_ADMIN_BASE_URL`, `ORDER_SERVICE_TIMEOUT`, `SUPERADMIN_REQUIRE_ORDER_SERVICE` | Planned Order internal dependency |
| Superadmin | `PAYMENT_SERVICE_ADMIN_BASE_URL`, `PAYMENT_SERVICE_TIMEOUT`, `SUPERADMIN_REQUIRE_PAYMENT_SERVICE` | Conditional payment summary/workflow dependency |
| Order Service | None found | DB DSN, listener, pool, migration, health, internal-auth config still need implementation |
| Dispute subsystem | None found | Ownership/storage/retention config undefined |

Generic variable ownership and credential placement already explained in:
`TaskImplementation/Superadmin Panel/Dependency/Environment.md`

When backend code is added:

- Commit sanitized `.env.example`, not real `.env` credentials.
- Validate mandatory variables at startup.
- Use secret manager/workload identity in deployed environments.
- Keep separate local, test, staging, and production values.
- Never copy tracked credential-like examples into production.

## 8. Docker Setup, Ports, and Networking

Task 4 adds no Dockerfile, Compose service, volume, network, health check, or restart policy. Reuse frontend-only Docker limitations from:

`TaskImplementation/Superadmin Panel/task1_Dependency.md`, section `8. Docker Setup`

MySQL and Redis Docker commands already exist in their dedicated dependency guides. Do not create an invented full-stack Compose manifest until Order/Superadmin listeners, health endpoints, migrations, and secret strategy are implemented.

### Ports table

| Service | Port/address | Purpose | Task 4 status |
|---|---:|---|---|
| Vite frontend | `5173` | Local admin UI | Reused; runnable |
| Vite preview | `4173` commonly | Built bundle preview | Reused; optional |
| API Gateway HTTP | `8080` | Browser REST entry | Reused; runtime missing |
| Auth/JWKS HTTP | `8081` in current Gateway config | Token/JWKS | Reused; runtime missing |
| Order gRPC | `50056` | Gateway order target | Task 4 required; runtime missing |
| Order internal HTTP | `8084` in Superadmin env | Planned review/detail dependency | Task 4 new use; runtime missing/conflicting |
| Superadmin HTTP | `8088` | Env-only service HTTP | Reused; runtime missing |
| Superadmin gRPC | `50062` | Gateway admin workflow target | Reused; runtime missing |
| Payment gRPC | `50057` | Gateway Payment target | Conditional; runtime missing |
| Payment internal HTTP | `8085` in Superadmin env | Planned payment dependency | Conditional; runtime missing/conflicting |
| MySQL | `3306` | `order_db` and `superadmin_db` | Task 4 required full stack |
| Redis | `6379` | Gateway rate limiting | Reused |

### Port and network findings

- Planned Order HTTP `8084` collides with current Cart and Wishlist env listeners.
- Planned Payment HTTP `8085` collides with current Search Service listener.
- Port ownership must be resolved before local multi-service startup.
- Docker/Kubernetes me `localhost` means current container/pod, not another service. Use private service DNS names.
- Gateway CORS must allow exact frontend origin, normally `http://localhost:5173`, and required headers including `Authorization`, `Content-Type`, and `X-Request-ID`.
- MySQL/Redis/internal gRPC/HTTP ports public internet par expose mat karein.
- New service manifests should include separate liveness/readiness probes, non-root runtime, restart policy, resource limits, and secret injection.

## 9. Local Development Setup

### Step 1: Read reused setup first

Follow:

- `task1_Dependency.md` for clone, tools, pnpm, `.env`, Docker limitation, and standard frontend startup.
- `task2_Dependency.md`, section 7, for the unresolved API-base issue.
- `task3_Dependency.md` only where shared Gateway/backend-runtime limitations are referenced.

### Step 2: Go to the frontend workspace

```bash
cd Ecommerce/frontend
```

### Step 3: Install locked dependencies

```bash
pnpm install --frozen-lockfile
```

No Task 4 package-add command is needed.

### Step 4: Configure frontend environment

Create `.env` from the existing template as documented in Task 1. Add no Task 4 variable. Real integration se pehle shared `/api/v1` path convention fix and test karein.

### Step 5: Choose validation mode

Frontend-only mode:

- MySQL, Redis, migrations, Order, Payment, and Superadmin runtimes skip kar sakte hain.
- Automated Task 4 tests reliable validation boundary hain.

Full-stack mode:

1. Obtain runnable Auth, Gateway, Order, and Superadmin implementations; Payment only if chosen response path needs it.
2. Add sanitized backend env templates and resolve port collisions.
3. Apply reviewed versioned migrations for `order_db`, `superadmin_db`, and the approved dispute model.
4. Start MySQL and Redis using existing dependency guides.
5. Start services using their future owning READMEs; do not guess commands from empty folders.
6. Implement and contract-test missing detail, disputes, and review endpoints.
7. Provision a least-privilege non-production admin and synthetic orders/disputes.

### Step 6: Run Task 4 tests

```bash
cd frontend/superadmin-panel
pnpm exec vitest run tests/orders
```

### Step 7: Run full frontend quality checks

```bash
cd frontend
pnpm run superadmin:typecheck
pnpm run superadmin:test
pnpm run superadmin:build
```

## 10. Running the Project

Start the frontend:

```bash
cd frontend
pnpm run superadmin:dev
```

Open:

```text
http://localhost:5173/admin/orders
http://localhost:5173/admin/orders/<test-order-id>
```

Without a valid admin session, route `/login` par redirect karega. Backend unavailable ho to real data requests fail hona npm install problem nahi hai.

### Verified frontend baseline

On 2026-06-20, current repository produced:

| Check | Result |
|---|---|
| Task 4 order tests | `4` files, `10` tests passed |
| TypeScript project check | Passed |
| Vite production build | Passed |
| Build note | Main JavaScript chunk is about `560 kB`; code-splitting warning only, not a build failure |
| New package required | None |

### Full-stack read-only verification

Only the list endpoint currently exists in the master contract. Use a temporary non-production token through a secure client:

```bash
export ADMIN_TOKEN='<temporary-admin-access-token>'
curl -i \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "X-Request-ID: task4-order-list-check" \
  'http://localhost:8080/api/v1/admin/orders?status=paid&user_id=user_test&page=1&page_size=25'
```

After missing contracts are approved and implemented, verify a synthetic order:

```bash
export TEST_ORDER_ID='<synthetic-test-order-id>'
curl -i \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "X-Request-ID: task4-order-detail-check" \
  "http://localhost:8080/api/v1/admin/orders/$TEST_ORDER_ID"

curl -i \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "X-Request-ID: task4-dispute-check" \
  "http://localhost:8080/api/v1/admin/orders/$TEST_ORDER_ID/disputes"
```

Manual-review mutation ko curl example se casually run na karein. Use an approved disposable order, follow the final server contract, and verify review task, order/review state, and immutable audit row as one acceptance flow.

### Expected frontend permissions

| Role | View orders/detail/disputes | Submit manual review |
|---|---:|---:|
| `superadmin` | Yes | Yes |
| `operations_admin` | Yes | Yes |
| `readonly_admin` | Yes | No |
| `finance_admin` | No | No |
| `catalog_admin` | No | No |

Backend must enforce the same or a stricter approved matrix. Current master contract only says generic `auth: admin`, which is not granular enough.

## 11. Common Errors & Fixes

Generic Node, pnpm, Docker, port, login, CORS, and dependency-cache issues already documented hain:

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`11. Common Errors & Fixes`

Task 4-specific issues:

| Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| URL contains `/api/v1/api/v1/` | Base and Task 4 path both include prefix | Align shared path convention; section 7 dekhein | Production-like URL tests |
| Unit tests pass but real URL fails | Tests set API base to host origin, unlike `.env.example` | Add config integration test using real template convention | Test login and all feature URLs together |
| List returns `502`/`Unavailable` | Gateway cannot reach Order gRPC `50056` | Supply/start Order runtime; check DNS/TLS/readiness | Dependency-aware Gateway readiness |
| Detail/disputes/review returns `404` | Routes are absent from master API contract/backend | Approve schemas, owner, auth, and implementation | Contract-first CI checks |
| `q`, `review_status`, `from`, or `to` ignored/rejected | `AdminOrderListRequest` currently defines only status/user/seller plus pagination | Extend and version contract or remove unsupported UI filters | Shared generated contract tests |
| Pagination behaves incorrectly | Frontend sends `limit`; contract uses `page_size` | Align request/response pagination names | One shared pagination type |
| Order detail lacks history/shipments/payment/review state | Current `Order`/`OrderListResponse` schemas are much smaller than UI model | Define separate admin detail response | Schema compatibility tests |
| Order appears as `created` or review as `none` unexpectedly | Frontend normalizer falls back on unknown/missing statuses | Return versioned enums and log contract violations | Do not silently expand statuses without coordination |
| Amount is 100x too small/large | Major units vs minor units mismatch | Standardize integer minor units and currency | Contract examples plus money tests |
| Manual review stays pending/fails | Superadmin/Order dependency unavailable or endpoint missing | Check request ID and both service logs; do not retry blindly | Idempotency and readiness checks |
| Review submits twice | Repeat click/network retry without server idempotency | Server must deduplicate and return stable result | Idempotency key and unique workflow constraint |
| Review succeeds but audit row is missing | Mutation and audit are not atomic or dependency was optional | Treat action as failed and investigate | Mandatory atomic audit/outbox policy |
| Review state disagrees with DB task status | Frontend and `admin_review_tasks` use different enums | Define explicit mapping/state machine | Versioned schema and transition tests |
| Dispute panel always empty | No dispute storage/owner/endpoint exists | Implement approved dispute domain and seed synthetic data | End-to-end fixture and retention policy |
| Payment summary unavailable | No projection/Payment Service integration | Select backend aggregation design | Dependency contract and graceful read-only fallback |
| Order HTTP cannot bind `8084` | Cart/Wishlist already use that port | Assign unique reviewed ports/service DNS | Central port registry/Compose validation |
| Payment HTTP cannot bind `8085` | Search Service already uses that port | Assign unique reviewed port | CI port collision check |
| `403` for a documented role | Generic API role policy and frontend matrix disagree | Align endpoint/action-specific permissions | One authoritative permission matrix |

## 12. Security & Best Practices

### Task-specific security rules

- Backend must authorize every order/detail/dispute request and every review mutation; hidden UI controls are not security.
- Use least privilege. `readonly_admin` must not mutate, and finance/catalog roles should not inherit order access accidentally.
- Prevent IDOR: arbitrary `order_id` lookup must still require admin permission and be audited where policy requires.
- Admin detail DTO should return minimum required fields. Avoid full address snapshots, raw phone/email, card data, provider payloads, or payment client secrets.
- Treat dispute summaries and internal notes as sensitive support data. Set length limits, sanitize output, apply retention policy, and restrict logging/access.
- Server must validate decision enum, current state, allowed transition, reason minimum/maximum, note maximum, actor, request ID, and resource existence.
- UI's 10-character reason check is bypassable; server validation is mandatory.
- Manual review mutation, review task update, and immutable audit record should be atomic where one database owns them. Cross-service updates need idempotency, optimistic concurrency, and a reliable outbox/saga design.
- Store audit actor, action, order ID, before/after summary, reason, request ID, outcome, and timestamp—without secrets or unnecessary PII.
- Do not log Bearer tokens, MySQL DSNs, Redis passwords, full order/address snapshots, internal notes, or payment provider secrets.
- Use TLS externally and authenticated/encrypted internal service communication in production.
- Keep MySQL, Redis, gRPC, and internal HTTP endpoints private.
- Use synthetic orders/disputes and a dedicated non-production admin for destructive QA.
- Apply query bounds: maximum date range/page size, indexed filters, timeouts, and rate limits to prevent expensive platform-wide scans.
- Carry `X-Request-ID` end to end for investigation without exposing customer data.

### Configuration audit

| Severity | Finding | Required action |
|---|---|---|
| Critical | Order detail, dispute, and review endpoints/schemas are absent from master contract | Contract, authorize, implement, and integration-test before release |
| Critical | Runnable Order/Superadmin backend flow is absent | Add source, manifests, health checks, migrations, and service READMEs |
| Critical | Tracked backend `.env` files contain credential-like DSNs/passwords | Rotate real/reused values; replace tracked secrets with sanitized templates |
| High | Frontend base/path duplicates `/api/v1` | Normalize URL construction across login and modules |
| High | No dispute data model, owner, migration, or retention policy | Define one owning service and versioned storage/API design |
| High | Review API model does not map to current review-task enum/schema | Define state machine, `internal_note` handling, and concurrency behavior |
| High | Generic `auth: admin` is broader than Task 4 role matrix | Add endpoint/action-specific backend permissions |
| High | Superadmin required DB/Order flags are false | Fail closed for production mutations and mandatory audit dependencies |
| High | No versioned Order/Superadmin migrations | Add reviewed up/down migrations and deployment runner |
| High | Frontend filters/pagination exceed current contract | Align `q`, review/date filters, `limit` vs `page_size`, and response fields |
| Medium | Order/Payment planned HTTP ports collide with other service envs | Assign unique ports and document service DNS |
| Medium | List contract routes directly to Order while review ownership/orchestration is undefined | Publish one service flow with auth, timeout, and consistency rules |
| Medium | Current response schema lacks detail/history/shipment/payment/review fields | Create admin-specific list/detail schemas |
| Medium | No Docker/Compose/deployment health configuration | Add only after runnable service boundaries are defined |
| Low | `task4.md` suggests packages unused by actual code | Use manifest/imports as source of truth; avoid lockfile churn |

### Task-specific operational best practices

- Separate liveness from readiness for Gateway, Order, Superadmin, MySQL, Redis, and optional Payment Service.
- Emit metrics for list/detail latency, dispute fetch failures, review queue depth, review decision latency, audit-write failure, downstream timeout, and duplicate/idempotent requests.
- Alert on review mutation success without corresponding audit persistence.
- Add structured logs with sanitized order ID, action, status, and request ID.
- Load-test admin search with realistic indexes and maximum ranges; do not scan raw order history for analytics.
- Back up `order_db` and `superadmin_db`; test restore and migration rollback procedures.
- Use a read replica/materialized projection for heavy cross-platform search only after consistency requirements are defined.
- Keep raw order status separate from admin review status in API and monitoring.

## 13. Missing or Misconfigured Things

Frontend Task 4 tooling ready hai, but complete production onboarding is blocked by:

1. Order Service and Payment Service runtime/config folders are absent.
2. Gateway/Auth/Superadmin required runnable implementation is absent for this flow.
3. Order detail, dispute, and manual-review routes/schemas are missing from `api/master-api.json`.
4. Dispute owner, table/collection, indexes, retention, and migrations are undefined.
5. Frontend list filters exceed `AdminOrderListRequest`; `limit` disagrees with `page_size`.
6. Existing `Order` response lacks review status, item detail fields, history, shipments, and payment projection required by the UI.
7. Frontend `.env.example` and Task 4 path duplicate `/api/v1`.
8. Review state enums do not map directly to `admin_review_tasks.status`.
9. `internal_note` storage/access/retention is undefined.
10. Versioned Task 4 migrations and migration runner are absent.
11. Order/Superadmin orchestration, internal auth, idempotency, and atomic audit behavior are not defined.
12. Planned Order `8084` and Payment `8085` HTTP ports collide with other service configs.
13. Backend credential-like values are tracked in `.env` files.
14. No supported Docker/Compose/Kubernetes manifests or health endpoints exist.
15. Full-stack seed/provisioning process for synthetic orders, disputes, and admin accounts is missing.

These are backend/platform prerequisites—not a reason to install extra frontend packages.

## 14. References to Previous Dependency Files

All earlier task dependency files in this service folder were reviewed before creating this incremental guide.

| Previous dependency file | Section/topic reused | Why reused |
|---|---|---|
| `task1_Dependency.md` | Tech stack, software installation, pnpm, `.env`, Docker limitation, startup, generic errors | Same frontend application and toolchain |
| `task2_Dependency.md` | API base-path mismatch, Gateway/auth/Redis model, backend credential rules | Same unresolved shared HTTP configuration/runtime |
| `task3_Dependency.md` | Incremental documentation pattern, backend-runtime limitations, service security practices | Same Superadmin architecture; Task 4 adds Order-specific dependencies |
| `Dependency/Frontend.md` | Actual package inventory and commands | Task 4 adds no npm package |
| `Dependency/Environment.md` | Frontend/backend variable ownership and Order/Payment target variables | Same Vite and secret-placement rules |
| `Dependency/API_Gateway.md` | Gateway responsibility, admin order route, auth, and Redis | Same browser REST entry point |
| `Dependency/MySQL.md` | MySQL install, Docker, credentials, start, verification | Same server; Task 4 adds `order_db`/review schema context |
| `Dependency/Redis.md` | Redis install, Docker, variables, and health check | Same Gateway rate-limit dependency |
| `Dependency/Migrations.md` | Bootstrap-versus-versioned-migration warning | No Task 4 migration runner exists |
| `Dependency/gRPC.md` | Internal service communication | Gateway expects Order `50056` and Superadmin `50062` |
| `Dependency/Protobuf.md` | Contract generation/versioning gap | Missing Task 4 APIs need approved schemas/RPCs |
| `Dependency/Go_Modules.md` | Missing backend module/workspace setup | Runnable backend chain remains unavailable |
| `Dependency/main_dependency.md` | Overall topology, downstream Order/Payment variables, and missing-runtime audit | Same platform foundation |

## 15. Final Checklist

### Frontend-only readiness

- [ ] Previous dependency guides checked
- [ ] Node.js 22.12+ and pnpm 10 available
- [ ] Workspace installed with `pnpm install --frozen-lockfile`
- [ ] No unnecessary `clsx`, MSW, jest-dom, or duplicate package added
- [ ] Existing frontend `.env` created without secrets
- [ ] Shared API base-path defect understood before real API testing
- [ ] Task 4 order tests pass
- [ ] Full frontend typecheck passes
- [ ] Full frontend test suite passes
- [ ] Production build passes
- [ ] Vite starts and `/admin/orders` routes render for an approved session
- [ ] Backend absence is not mistaken for a frontend dependency failure

### Full-stack readiness

- [ ] Shared `/api/v1` URL convention fixed and integration-tested
- [ ] Runnable Auth, Gateway, Order, and Superadmin services supplied
- [ ] Payment projection/Service requirement explicitly decided
- [ ] Unique Order and Payment HTTP ports/service DNS assigned
- [ ] Sanitized backend env templates supplied and tracked secrets rotated
- [ ] MySQL `order_db` and `superadmin_db` created through reviewed migrations
- [ ] Required order tables and indexes verified
- [ ] Dispute owner/schema/indexes/retention/migrations implemented
- [ ] Review state mapping and `internal_note` policy approved
- [ ] Redis reachable for Gateway rate limiting
- [ ] Missing detail/dispute/review contracts and RPCs approved and implemented
- [ ] Filter and pagination contract matches the frontend
- [ ] Admin list/detail response contains versioned minor-unit money and required fields
- [ ] JWT/JWKS, exact CORS origin, and request headers verified
- [ ] Backend RBAC matches approved view/dispute/review matrix
- [ ] Production review dependencies fail closed
- [ ] Internal service auth, TLS, timeouts, retries, and readiness configured
- [ ] Review mutation is idempotent and concurrency-safe
- [ ] Review task/state update and immutable audit behavior verified
- [ ] Dedicated non-production admin and synthetic order/dispute fixtures provisioned
- [ ] Order list/search/filter/pagination verified
- [ ] Detail items/history/shipments verified
- [ ] Payment summary exposes no payment secret or card data
- [ ] Dispute data is minimized and permission-protected
- [ ] Manual review tested only on disposable data with a valid reason
- [ ] Logs/metrics checked using request ID without leaking PII, notes, tokens, or DSNs
- [ ] Database backup/restore and migration rollback tested
- [ ] No duplicate setup documentation added

Frontend Task 4 setup is verified. Complete Order Operations readiness tabhi claim karein jab missing backend contracts/runtimes, dispute storage, migrations, granular authorization, review consistency, and immutable audit behavior end to end implement and test ho chuke hon.
