# Superadmin Panel Task 5 - Dependency & Setup Guide

## 1. Project Overview

Yeh guide `task5.md` ke **Payment Operations** module ko locally validate aur eventually full stack me run karne ke liye dependency, environment, database, payment-provider, reconciliation, networking, security, aur troubleshooting requirements explain karti hai. Business logic aur implementation snippets yahan repeat nahi kiye gaye hain.

Actual frontend implementation yahan present hai:

```text
frontend/superadmin-panel/src/features/payments/
frontend/superadmin-panel/tests/payments/
```

Module me payment search/filter, payment detail and attempts, refund queue, approve/reject review, reconciliation alerts, role checks, loading/error states, aur automated tests included hain.

### Current repository reality

| Capability | Current status | Beginner meaning |
|---|---|---|
| Task 5 frontend source | Ready | Routes, pages, API client, Query hooks, permissions, and tests present hain |
| Frontend install/test/build | Ready | Existing pnpm workspace se run hota hai |
| New Task 5 npm dependency | None | Koi package add karne ki zarurat nahi hai |
| New frontend environment variable | None | Existing `.env.example` reuse hota hai |
| Admin payment list API | Contract only | Master API me defined hai, but runnable Payment Service absent hai |
| Refund review API | Contract only | Superadmin RPC contract defined hai, but runnable service absent hai |
| Payment detail/refund list/reconciliation APIs | Missing contract/backend | Frontend in endpoints ko call karta hai, master contract define nahi karti |
| Payment MySQL bootstrap schema | Present | `database/draw.sql` me five payment tables hain; versioned migrations nahi hain |
| Provider integration | Design only | Provider SDK, keys, webhook config, and runnable adapter nahi mile |
| Reconciliation worker | Design only | Schedule, report source, credentials, worker, and alert delivery absent hain |
| Docker/Compose stack | Missing | Supported Dockerfile or Compose manifest nahi mila |

> **Important:** Task 5 tests, TypeScript check, production build, aur frontend dev server backend ke bina run ho sakte hain. Real payment/refund data aur finance actions tabhi work karenge jab Gateway, Auth, Payment, Superadmin, provider, migrations, and missing contracts complete hon.

### Task-specific API dependency

| Method | Frontend path | Contract status | Expected owner/purpose |
|---|---|---|---|
| `GET` | `/api/v1/admin/payments` | Present | Payment Service; list, filters, pagination |
| `GET` | `/api/v1/admin/payments/{payment_id}` | Missing | Payment Service; admin-safe detail, attempts, refunds |
| `GET` | `/api/v1/admin/refunds` | Missing | Payment/Superadmin owner; finance queue and pagination |
| `POST` | `/api/v1/admin/refunds/{refund_id}/review` | Present | Superadmin Service; audited approve/reject workflow |
| `GET` | `/api/v1/admin/payment-reconciliations` | Missing | Payment Service; mismatch list |
| `GET` | `/api/v1/admin/payment-reconciliations/{id}` | Missing | Payment Service; mismatch detail |

Safe request flow:

```text
Browser -> API Gateway -> server-side RBAC -> owning service -> private database/provider
```

Browser ko MySQL, Redis, gRPC, internal service, ya payment provider secret API directly call nahi karna chahiye.

## 2. Tech Stack

Task 5 same Superadmin frontend stack reuse karta hai. Full beginner definitions and installation steps dobara repeat nahi kiye gaye hain.

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Sections:
`2. Tech Stack`, `3. Required Software`, and `4. Dependency Management`

Task-specific use:

| Technology/service | Task 5 me use | Required? | New? |
|---|---|---:|---:|
| React 19 + TypeScript | Payment pages, detail panels, refund/reconciliation UI | Yes | No |
| React Router | `/admin/payments` and `/admin/payments/refunds` | Yes | No |
| TanStack React Query | Lists/details, cache, refetch, review mutation invalidation | Yes | No |
| Zustand | Admin access token and roles | Yes | No |
| Vite + Tailwind CSS | Local dev, build, and styling | Yes | No |
| Vitest + Testing Library + jsdom | API, permissions, filters, refund, reconciliation tests | Tests only | No |
| Testing Library user-event | Refund reason interaction test | Tests only | No |
| Lucide React | Payment/refund UI icons | Yes | No |
| API Gateway + Auth/JWKS | Browser REST boundary and authenticated admin identity | Real integration | No |
| MySQL | Financial, refund, reconciliation, review, and audit records | Real integration | New Task 5 schema usage |
| Redis | Gateway rate limiting under current local config | Real integration | No |
| Payment Service | Payment list/detail/refunds/reconciliation and provider adapter | Real integration | New mandatory service |
| Superadmin Service | Audited refund decision and review task | Refund integration | New Task 5 workflow use |
| Payment provider | Actual intent/refund/webhook/settlement operations | Production integration | New, provider undecided |
| Reconciliation worker/scheduler | Provider report comparison and mismatch creation | Production operations | New, missing |

**Simple Hinglish:** Payment Service financial domain ka backend owner hoga. MySQL strongly consistent records store karega, provider external money movement karega, aur Superadmin Service finance admin ka reviewed decision plus audit trail manage karega. Frontend sirf approved Gateway APIs call karega.

### Planning guide versus actual manifest

`task5.md` me generic package-add commands `clsx`, MSW, and jest-dom bhi suggest karte hain. Actual implementation and manifest ke according:

- `clsx` import nahi hota; existing project styling/helpers enough hain.
- MSW direct dependency/import nahi hai; payment API tests Vitest `fetch` mocks use karte hain.
- jest-dom import nahi hota.
- `@testing-library/user-event` already installed hai aur refund interaction test use karta hai.
- React, Query, Router, Zustand, Lucide, Vitest, Testing Library, and jsdom already manifest plus lockfile me hain.

> `task5.md` ke package-add commands current repository par mat run karein. `package.json`, actual imports, and `pnpm-lock.yaml` source of truth hain.

## 3. Required Software

Git, Node.js 22.12+, pnpm 10, aur modern browser setup unchanged hai.

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`3. Required Software`

Validation mode ke hisaab se requirements:

| Mode | Required software/services |
|---|---|
| Payment frontend tests/typecheck/build | Node.js, pnpm, existing locked packages |
| Frontend browser and error/permission states | Above plus modern browser |
| Real payment read APIs | API Gateway, Auth/JWKS, Payment Service, MySQL `payment_db`, current Gateway Redis dependency |
| Real refund review | Above plus Superadmin Service and MySQL `superadmin_db` |
| Real provider refund | Above plus provider sandbox account, backend-only keys, verified webhook, idempotency |
| Real reconciliation | Above plus scheduler/worker, provider settlement report access, and alert persistence |

No runnable Payment Service, Gateway, Auth, or Superadmin implementation was found for this flow. Exact backend build/start commands cannot honestly be supplied until those services exist.

## 4. Dependency Management

### New dependency delta

| Check | Result |
|---|---|
| New runtime npm package | None |
| New development npm package | None |
| `package.json` update needed | No |
| `pnpm-lock.yaml` update needed | No |
| Payment provider npm SDK | Not required in admin browser; provider SDK belongs in backend |
| New Go module available | No; Payment source/module is absent and required backend folders are env-only |

Existing pnpm explanation, locked installs, scripts, cache/proxy issues, and Node version troubleshooting are already documented:

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
- `pnpm add` commands tabhi run karein when implementation genuinely imports a new package.
- Browser bundle me Stripe/Razorpay secret SDK configuration add mat karein. Hosted client SDK ki publishable key future checkout scope hai, admin operations scope nahi.
- Missing backend folder me `go mod tidy`, `go build`, or `go run` run karne se service create nahi hogi.
- Backend Go workspace limitations ke liye `Dependency/Go_Modules.md` refer karein.

## 5. Database Setup

Frontend directly database access nahi karta. DB credentials sirf owning backend service me rahenge.

### Task 5 data ownership

```text
Browser
  -> API Gateway
     -> Payment Service     -> MySQL payment_db
        -> payment provider (server-to-server)
     -> Superadmin Service  -> MySQL superadmin_db
        -> Payment Service  -> approved/rejected refund workflow
```

| Store | Task 5 data | Required when | Default port | Current status |
|---|---|---|---:|---|
| MySQL `payment_db` | Payments, attempts, refunds, webhooks, reconciliations | Real read/write flow | `3306` | Bootstrap schema present; service/migrations absent |
| MySQL `superadmin_db` | Review tasks and immutable admin audit | Refund review | `3306` | Bootstrap schema present; service/migrations absent |

MySQL ek relational database hai. Financial records ke liye transactions, unique constraints, indexes, and durable auditability useful hote hain, isliye architecture Payment domain ke liye MySQL choose karti hai. Frontend-only testing ke liye optional hai; real integration ke liye mandatory hai.

### MySQL setup reuse

Local installation, Docker container, persistent volume, credentials, start, verification, security, and generic errors already documented hain.

Refer:
`TaskImplementation/Superadmin Panel/Dependency/MySQL.md`

Sections:
`6. Installation steps` through `13. Security notes`

Task-specific schema verification sirf reviewed local migration/bootstrap apply karne ke baad run karein:

```bash
mysql -u root -p -e "USE payment_db; SHOW TABLES LIKE 'payments'; SHOW TABLES LIKE 'payment_attempts'; SHOW TABLES LIKE 'refunds'; SHOW TABLES LIKE 'payment_webhook_events'; SHOW TABLES LIKE 'payment_reconciliations';"
mysql -u root -p -e "USE superadmin_db; SHOW TABLES LIKE 'admin_review_tasks'; SHOW TABLES LIKE 'admin_audit_logs';"
```

Expected: five Payment tables and two Superadmin tables return hon.

Index and constraint verification:

```bash
mysql -u root -p -e "USE payment_db; SHOW INDEX FROM payments; SHOW INDEX FROM payment_attempts; SHOW INDEX FROM refunds; SHOW INDEX FROM payment_webhook_events; SHOW INDEX FROM payment_reconciliations;"
mysql -u root -p -e "USE superadmin_db; SHOW INDEX FROM admin_review_tasks; SHOW INDEX FROM admin_audit_logs;"
```

Important existing protections include payment/refund idempotency unique keys and unique provider webhook event IDs. Query-plan testing is still required for provider/status/date pagination and reconciliation queues.

### Schema and API alignment gaps

| Area | Frontend expectation | Bootstrap schema | Required resolution |
|---|---|---|---|
| Payment status | Includes `retry_allowed` | `payments.status` does not | Approve one canonical state model/migration |
| Attempt status | Reuses broad payment statuses | DB allows `initiated`, `succeeded`, `failed` | Add separate API/TS attempt enum or map explicitly |
| Refund queue | Defaults to `pending_review` | `refunds.status` has no `pending_review` | Define whether review task or refund row owns pending state |
| Reconciliation time | UI expects `detected_at` | Table stores `created_at` | Map/rename in approved response schema |
| Reconciliation detail | UI expects local/provider amount and statuses | Table stores generic `details` JSON | Version and validate a typed API projection |
| Review linkage | UI reviews a refund | Review table has generic resource fields, no FK | Define resource mapping and transactional consistency |
| Money | UI formats integer amount as minor units | DB uses `BIGINT` | Contract must state integer minor units and ISO currency |

Do not silently rely on frontend normalizers to hide invalid backend records. Unknown financial states should be observable and contract-tested.

### Migration requirements

- Versioned Payment/Superadmin up/down migrations and migration runner are absent.
- `database/draw.sql` repository-wide bootstrap hai, safe production migration command nahi.
- Provider webhook payload and raw response JSON need encryption/redaction/retention decisions.
- Foreign keys currently cover attempts/refunds to payments, but cross-service Superadmin data ko direct cross-database FK se couple nahi karna chahiye.
- Large reconciliation batches need stable pagination, indexes, retention, and archival policy.

Migration setup and rollback practices already explained hain:
`TaskImplementation/Superadmin Panel/Dependency/Migrations.md`

## 6. Redis, Queue, and External Services

### API Gateway, Auth, and Redis

Gateway/JWT/JWKS/CORS/Redis setup reused hai. Full installation and health guidance:

- `TaskImplementation/Superadmin Panel/Dependency/API_Gateway.md`
- `TaskImplementation/Superadmin Panel/Dependency/Redis.md`
- `TaskImplementation/Superadmin Panel/task2_Dependency.md`, sections `6` and `7`

Task 5 specifically expects:

- Browser REST through Gateway port `8080`.
- Gateway Payment gRPC target at `localhost:50057` in current local env.
- Gateway Superadmin gRPC target at `localhost:50062` for refund review.
- JWT/JWKS authentication and server-side finance permissions before returning data or accepting decisions.
- Redis rate limiting at `localhost:6379`; current `RATE_LIMIT_FAIL_OPEN=false` makes Redis availability significant for the Gateway.

### Payment Service

Payment Service real Task 5 integration ke liye mandatory hai. Yeh payment records, attempts, refunds, provider adapter, webhook processing, and reconciliation own karega. Repository me `backend/services/payment-service/`, module, config loader, server, migrations, provider adapter, worker, and health endpoints nahi mile.

Backend implementation ko at minimum provide karna hoga:

- Admin list/detail/refund/reconciliation contracts matching the UI.
- MySQL transaction and connection-pool configuration.
- Provider-independent interface plus selected sandbox adapter.
- Raw-body webhook signature verification and replay/idempotency protection.
- Refund concurrency rules and provider result reconciliation.
- Separate liveness/readiness endpoints and metrics.

### Superadmin Service

Refund approve/reject endpoint master contract me Superadmin Service own karta hai. Current folder only ignored local `.env` provide karta hai; runnable code absent hai.

Review flow ko:

1. current refund state lock/read karna,
2. actor permission and maker-checker rule validate karna,
3. immutable review/audit record write karna,
4. Payment Service ko authenticated idempotent command dena,
5. partial failure ko retry/reconcile karna

chahiye. `SUPERADMIN_REQUIRE_DATABASE=false` and `SUPERADMIN_REQUIRE_PAYMENT_SERVICE=false` local defaults production mutation ke liye unsafe honge; money-changing review dependency unavailable ho to fail closed karein.

### Payment provider

Stripe/Razorpay-like provider external payment platform hai. Frontend Task 5 ko provider key ki zarurat nahi; actual refund, webhook, and settlement operations ke liye backend provider sandbox mandatory hai.

Current repository provider select nahi karti aur no SDK/config loader exists. Therefore provider install commands or fake variable names document karna unsafe hoga. Provider choose hone ke baad backend-only configuration must cover:

- sandbox account and API base URL,
- secret API key/account identifier,
- webhook signing secret and exact public callback URL,
- request/connect timeout and safe retry policy,
- supported currencies and amount limits,
- settlement report credentials/location,
- secret rotation and separate test/staging/production accounts.

No secret `VITE_*` variable me rakhna hai.

### Reconciliation worker and report input

Reconciliation provider settlement report ko local records se compare karke mismatch alert banata hai. Task 5 UI ke mismatch panel ke liye yeh production dependency mandatory, but frontend-only mode me optional hai.

Missing operational pieces:

- worker/scheduler implementation and timezone,
- provider report API/SFTP/object-storage source,
- checkpoint and rerun strategy,
- duplicate batch/idempotency key,
- typed mismatch details,
- alert severity/ownership/SLA,
- metrics for report freshness, matched count, mismatch count, and failures.

Worker ko least-privilege read access to reports and controlled write access to reconciliation records dena chahiye. UI ko stale report timestamp clearly expose karna chahiye.

### Queue and downstream services

Architecture document Payment events ke liye `Kafka/RabbitMQ` concept dikhati hai, but inspected Payment implementation/config me broker selection, URL, topic, producer, consumer, or outbox worker nahi mila. Isliye Task 5 ke liye Kafka/RabbitMQ install karna currently justified nahi hai.

Provider refund completion ke baad Order and Notification services future full workflow me consumers ho sakte hain. Broker and delivery semantics approve hone ke baad one supported setup document karein; Kafka and RabbitMQ dono blindly start mat karein.

## 7. Environment Variables

### Frontend delta

Task 5 introduces **no new or changed frontend environment variable**.

Existing `.env` creation, Vite loading, purpose, and security rules:

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`7. Environment Variables`

Never add MySQL DSNs, Redis passwords, JWT keys, provider secret keys, webhook secrets, settlement credentials, or internal-service tokens to a `VITE_*` variable. Every `VITE_*` value browser JavaScript me visible hota hai.

### Existing API base-path blocker

Same shared defect Task 5 ko bhi affect karta hai:

Refer:
`TaskImplementation/Superadmin Panel/task2_Dependency.md`

Section:
`7. Environment Variables` -> `Critical base-path mismatch`

Template base already `/api/v1` par end hota hai:

```dotenv
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

Task 5 constants bhi `/api/v1/...` se start hote hain, so real URL duplicate ho jata hai:

```text
http://localhost:8080/api/v1/api/v1/admin/payments
```

Only `.env` ko origin par change karna payment URLs fix karega but current login `/auth/login` composition change karega. Permanent fix: one convention choose karein—recommended base `/api/v1`, feature paths `/admin/...`, login `/auth/login`—and production-like URL tests add karein.

### Backend-only detected configuration

| Owner | Detected variable/group | Purpose/status |
|---|---|---|
| Gateway | `PAYMENT_GRPC_ADDR`, `SUPERADMIN_GRPC_ADDR`, `AUTH_GRPC_ADDR` | Service routing; local env values exist, runtime absent |
| Gateway | `JWT_*`, `REDIS_*`, `RATE_LIMIT_*` | Authentication and rate limiting |
| Superadmin | `SUPERADMIN_DATABASE_DSN` and DB pool settings | Review/audit MySQL; ignored local secret file only |
| Superadmin | `PAYMENT_SERVICE_ADMIN_BASE_URL`, timeout, require flag | Planned internal Payment dependency |
| Payment Service | None found | Listener, DB, provider, webhook, reconciliation config missing |
| Provider/report source | None found | Backend-only credentials and operational settings missing |

A valid new `.env` example cannot be generated before Payment Service config loader and provider are selected; invented names would not be loadable by any code. Implementation PR must include a sanitized committed `.env.example` defining at least these categories:

- service HTTP/gRPC listen addresses,
- Payment DB DSN and pool/timeouts,
- selected provider name, sandbox endpoint, secret reference, webhook secret reference,
- public webhook URL/trusted proxy settings,
- internal service authentication and TLS,
- reconciliation enabled flag, schedule/timezone, report source, checkpoint,
- retry/timeouts/circuit-breaker settings,
- health/metrics/logging settings.

Real secrets local ignored `.env` or deployment secret manager me rahenge; only placeholders `.env.example` me commit karein. Generic placement and credential rules already documented hain in `Dependency/Environment.md`.

## 8. Docker Setup

Task 5 adds no Dockerfile, Compose service, volume, network, health check, or restart policy. Reuse:

- `task1_Dependency.md`, section `8. Docker Setup`, for frontend limitations.
- `Dependency/MySQL.md` for the existing local MySQL container approach.
- `Dependency/Redis.md` for the existing local Redis container approach.

Do not invent a full-stack Compose file until Payment/Superadmin listeners, migrations, provider sandbox strategy, readiness endpoints, and secret injection exist. A future supported stack should add separate Payment and reconciliation worker containers, migration job, private network, DB volume, health checks, non-root users, restart policy, and resource limits.

### Ports and networking

| Service | Port/address | Purpose | Task 5 status |
|---|---:|---|---|
| Vite frontend | `5173` | Local admin UI | Reused; runnable |
| Vite preview | `4173` commonly | Built bundle preview | Reused; optional |
| API Gateway HTTP | `8080` | Browser REST entry | Reused; runtime missing |
| Auth/JWKS HTTP | `8081` in current Gateway config | Token/JWKS | Reused; runtime missing |
| Payment gRPC | `50057` | Gateway Payment target | Task 5 mandatory; runtime missing |
| Payment internal HTTP | `8085` in Superadmin local env | Planned refund workflow | Task 5 mandatory path; runtime missing/conflicting |
| Superadmin HTTP | `8088` | Env-only service HTTP | Reused; runtime missing |
| Superadmin gRPC | `50062` | Gateway refund review target | Task 5 required; runtime missing |
| MySQL | `3306` | `payment_db` and `superadmin_db` | Task 5 full-stack required |
| Redis | `6379` | Gateway rate limiting | Reused |
| Provider webhook | Public HTTPS route via Gateway/service | Provider callbacks | New; exact host/port missing |
| Provider/report API | HTTPS/SFTP/object storage | Reconciliation input | New; provider-specific |

Networking findings:

- Planned Payment HTTP `8085` collides with current Search Service listener config.
- Gateway gRPC `50057` and Superadmin internal HTTP `8085` imply two Payment listeners; ownership and need must be explicit.
- Docker/Kubernetes me `localhost` current container/pod hota hai. Cross-container calls ke liye service DNS use karein.
- Gateway CORS must allow exact frontend origin, normally `http://localhost:5173`, and `Authorization`, `Content-Type`, `X-Request-ID` headers.
- MySQL, Redis, internal gRPC/HTTP, provider secrets, and reconciliation report endpoints public internet par expose mat karein.
- Provider webhook must use public HTTPS, signature verification, body-size limit, replay protection, and strict route isolation.

## 9. Local Development Setup

### Step 1: Read reused setup first

Follow:

- `task1_Dependency.md` for clone, tool installation, pnpm, frontend `.env`, and normal startup.
- `task2_Dependency.md`, section 7, for unresolved API base-path behavior.
- `task4_Dependency.md` for shared missing Gateway/Auth/Superadmin runtime and Payment port observations.

### Step 2: Go to the frontend workspace

```bash
cd Ecommerce/frontend
```

### Step 3: Install locked dependencies

```bash
pnpm install --frozen-lockfile
```

No Task 5 package-add command is needed.

### Step 4: Configure frontend environment

Create `frontend/superadmin-panel/.env` from `.env.example` exactly as Task 1 explains. Add no Task 5 variable and no secret. Real integration se pehle shared `/api/v1` convention fix karein.

### Step 5: Choose validation mode

Frontend-only mode:

- MySQL, Redis, Payment Service, provider, and reconciliation worker start nahi karne.
- Tests/typecheck/build se implemented module verify karein.
- Browser me backend errors expected hain unless API mocks or a real Gateway are supplied.

Full-stack mode:

- Continue only after missing API contracts, runnable services, migrations, sanitized env templates, unique ports, provider sandbox, and reconciliation source are delivered.
- MySQL/Redis generic setup previous guides se follow karein.
- Approved migrations run karein; raw bootstrap production par run mat karein.
- Auth -> Redis -> Payment/Superadmin -> Gateway -> frontend dependency order health/readiness ke according follow karein.

### Step 6: Run Task 5 tests

```bash
pnpm --filter superadmin-panel exec vitest run tests/payments
```

### Step 7: Run all frontend quality checks

```bash
pnpm superadmin:typecheck
pnpm superadmin:test
pnpm superadmin:build
```

## 10. Running the Project

### Frontend startup

From `frontend/`:

```bash
pnpm superadmin:dev
```

Open:

```text
http://localhost:5173/admin/payments
http://localhost:5173/admin/payments/refunds
```

Expected frontend permissions:

| Role | Payment/refund/reconciliation view | Refund approve/reject |
|---|---:|---:|
| `superadmin` | Yes | Yes |
| `finance_admin` | Yes | Yes |
| `readonly_admin` | Yes | No |
| `operations_admin` | No | No |
| `catalog_admin` | No | No |

Frontend role checks UX only hain. Backend must independently authorize each route and resource.

### Full-stack verification when dependencies are supplied

After URL convention fix and with a non-production admin token:

```bash
curl -i -H "Authorization: Bearer <finance-admin-test-token>" \
  "http://localhost:8080/api/v1/admin/payments?page=1&page_size=25&status=captured"

curl -i -H "Authorization: Bearer <finance-admin-test-token>" \
  "http://localhost:8080/api/v1/admin/refunds?status=pending_review&page=1&page_size=25"

curl -i -H "Authorization: Bearer <finance-admin-test-token>" \
  "http://localhost:8080/api/v1/admin/payment-reconciliations?status=mismatch&page=1&page_size=5"
```

Use synthetic sandbox records only. Approve/reject verification money-changing operation hai; disposable refund fixture, maker-checker policy, and audit observation ke bina mutation run mat karein.

Verify in browser/network/logs:

- final request URL me `/api/v1` once ho,
- bearer token and request ID Gateway tak jaye,
- page/filter names contract se match karein,
- money minor units correctly display hon,
- payment detail contains no card/provider secrets,
- readonly admin ko mutation controls and server mutation both denied hon,
- approved/rejected review writes actor, reason, resource, request ID, before/after, timestamp,
- reconciliation alert shows report freshness and typed difference,
- logs request ID se correlate hon without sensitive payloads.

## 11. Common Errors & Fixes

Generic pnpm, Node, Vite, Docker, MySQL, Redis, CORS, and Gateway errors previous guides me documented hain. Task-5-specific errors:

| Error/symptom | Likely cause | Fix | Prevention |
|---|---|---|---|
| Request contains `/api/v1/api/v1` | Base and feature path both versioned | Shared URL convention fix karein | Login plus payment URL tests add karein |
| `404` on payment detail/refund list/reconciliation | Endpoint master contract/backend me missing | Approve and implement contracts before integration | Contract-first CI and Gateway route tests |
| Gateway Payment target unavailable | Payment gRPC runtime `50057` absent/down | Implement/start service and check readiness | Dependency-aware readiness and alerts |
| Refund review returns downstream unavailable | Superadmin Payment HTTP target absent or `8085` collision | Assign unique target/service DNS and implement internal auth | Central port registry plus startup validation |
| Refund queue empty for `pending_review` | DB enum has no matching status | Approve status ownership/migration/mapping | Shared schema-generated enums/contract tests |
| Attempts show wrong/default state | UI attempt type reuses payment status while DB uses succeeded/failed | Separate attempt enum and explicit mapping | Schema compatibility tests |
| Amount appears 100x too large/small | Major units sent where UI expects minor units | Contract integer minor units and test boundaries | Shared Money schema and fixtures |
| Reconciliation panel always empty | Worker/report ingestion absent or endpoint missing | Implement schedule, report source, persistence, API | Freshness metric and missing-run alert |
| Reconciliation details are `Unknown` | DB JSON fields not projected to typed response | Version response and mapping | Provider fixture contract tests |
| Duplicate refund or `409` | Repeated decision/idempotency/concurrency conflict | Fetch latest status; retry only with same safe idempotency key | Lock/state transition and idempotency design |
| UI allows action but API gives `403` | Frontend role and backend permission model differ | Align approved permission claims; backend remains source of truth | End-to-end role matrix tests |
| Provider webhook accepted but status stale | Signature/event mapping/worker failed | Inspect safe event ID and retry queue; never manually fake captured | Idempotent webhook processing metrics/DLQ |

## 12. Security & Best Practices

### Task-specific security rules

- Card number, CVV, raw payment token, provider secret, webhook secret, and full raw provider payload UI/API/logs me expose mat karein.
- Refund permission backend par enforce karein; hidden button authorization nahi hota.
- Every decision needs actor, reason, request ID, before/after state, and immutable timestamp.
- High-value refund ke liye maker-checker and self-approval prohibition define karein.
- Review and provider refund command idempotent and concurrency-safe hona chahiye.
- Provider webhook raw bytes par signature verify karein before JSON trust; timestamp/replay window enforce karein.
- Server amount/order/refundable balance calculate kare; browser amount trust mat karein.
- Provider and DB secrets secret manager/workload identity me store and rotate karein.
- Use sandbox accounts and synthetic data locally/CI; real refunds developer laptop se test mat karein.
- Reconciliation report credentials least privilege hon; reports encrypted, retained, and access-audited hon.
- Admin endpoints tighter rate limits, MFA, short session timeout, TLS, and no public internal ports use karein.

### Configuration audit

| Finding | Risk | Recommended fix |
|---|---|---|
| No Payment Service runtime/config | No real Task 5 flow | Implement service, health, migrations, and sanitized env template |
| Three read API groups missing | Partial UI integration always 404 | Approve REST/RPC schemas and Gateway routing |
| Frontend/DB enums differ | Hidden/default financial states | Canonical versioned enums and compatibility tests |
| API base path duplicated | All real feature calls misroute | Normalize base/path convention centrally |
| No provider/webhook configuration | Refund/status cannot settle safely | Select provider sandbox and implement verified adapter |
| No reconciliation worker/report source | Mismatch UI has no data | Implement scheduled idempotent ingestion and freshness SLO |
| No versioned migrations | Unsafe/non-repeatable schema deployment | Add migration runner, up/down policy, backup/rollback test |
| Superadmin dependency flags default false | Mutation may operate without audit/downstream | Fail closed for financial mutation environments |
| Local backend `.env` files contain credential-like values but are ignored | Copy/leak/reuse risk; no team template | Rotate reused values and commit sanitized `.env.example` files |
| No Docker/health manifests | Reproducibility/readiness gaps | Add supported manifests after runtimes exist |

### Task-specific operational best practices

- Monitor provider webhook age, reconciliation report age, mismatch count, refund latency, failure rate, and stuck review tasks.
- Alert on missing daily report as well as mismatches; zero alerts can mean broken ingestion.
- Preserve provider/event/refund idempotency keys while redacting secrets.
- Use outbox or an equivalent durable delivery pattern for cross-service state changes.
- Reconcile Superadmin decision, Payment state, provider state, Order state, and notification delivery after partial failures.
- Backup/restore and migration rollback drills financial data par regularly test karein.
- Define retention for raw provider payloads and reports; minimize stored sensitive data.

## 13. Missing or Misconfigured Things

Frontend Task 5 tooling ready hai, but complete production onboarding is blocked by:

1. Payment Service folder, source, module, config loader, provider adapter, worker, and health endpoints absent hain.
2. Required Gateway/Auth/Superadmin runnable implementations is flow ke liye absent hain.
3. Payment detail, refund list, reconciliation list, and reconciliation detail routes/schemas master contract me missing hain.
4. Payment and Superadmin versioned migrations/migration runner absent hain.
5. Frontend payment/refund/attempt enums bootstrap MySQL enums se fully align nahi hote.
6. Reconciliation DB JSON/time fields typed frontend response se map nahi hote.
7. Frontend `.env.example` and payment paths duplicate `/api/v1`.
8. Payment provider, sandbox account, SDK, webhook secret, public callback, and key rotation undefined hain.
9. Reconciliation scheduler, timezone, report source/credentials, checkpoints, reruns, and freshness alerts undefined hain.
10. Refund review task-to-refund mapping, maker-checker, idempotency, locking, and partial-failure consistency undefined hain.
11. Planned Payment internal HTTP port `8085` another service config se collide karta hai.
12. Payment backend sanitized `.env.example`, Docker/Compose/Kubernetes manifests, and health checks absent hain.
13. Queue/outbox provider and downstream Order/Notification delivery semantics undefined hain.
14. Full-stack provisioning for sandbox provider, finance admin/MFA, synthetic payments/refunds/reports absent hai.
15. API response versioning, minor-unit money guarantee, retention, backup, and observability runbooks incomplete hain.

These are backend/platform prerequisites—not reasons to install extra frontend packages.

## 14. References to Previous Dependency Files

All earlier task dependency files in this service folder were reviewed before creating this incremental guide.

| Previous dependency file | Section/topic reused | Why reused |
|---|---|---|
| `task1_Dependency.md` | Tech stack, required software, pnpm, clone, `.env`, frontend Docker/startup, generic errors | Same React application and toolchain |
| `task2_Dependency.md` | API base-path mismatch, Gateway/auth/Redis model, credential placement | Same unresolved shared HTTP/runtime configuration |
| `task3_Dependency.md` | External-service decision pattern and secret rules | Same Superadmin architecture; no speculative broker setup |
| `task4_Dependency.md` | Shared missing runtimes, Payment targets/ports, service-to-service security | Task 5 turns Payment from conditional read into core dependency |
| `Dependency/Frontend.md` | Package inventory and frontend commands | Task 5 adds no npm package |
| `Dependency/Environment.md` | Frontend/backend variable ownership | Same Vite and secret-placement rules |
| `Dependency/API_Gateway.md` | Gateway responsibility, admin payment route, auth, Redis | Same browser REST entry |
| `Dependency/MySQL.md` | MySQL install, Docker, credentials, start, verification | Same server; Task 5 adds `payment_db` specifics |
| `Dependency/Redis.md` | Redis install, Docker, variables, health | Same Gateway rate-limit dependency |
| `Dependency/Migrations.md` | Bootstrap-versus-versioned-migration warning | No Task 5 migration runner exists |
| `Dependency/gRPC.md` | Internal communication setup | Gateway expects Payment `50057` and Superadmin `50062` |
| `Dependency/Protobuf.md` | Contract generation/versioning | Missing Task 5 APIs require approved schemas/RPCs |
| `Dependency/Go_Modules.md` | Missing backend module/workspace setup | Runnable backend chain remains unavailable |
| `Dependency/main_dependency.md` | Overall topology, setup order, ports, and missing-runtime audit | Same platform foundation |

## 15. Final Checklist

### Frontend-only readiness

- [ ] Previous dependency documentation checked
- [ ] Node.js 22.12+ and pnpm 10 available
- [ ] Workspace cloned and installed with `pnpm install --frozen-lockfile`
- [ ] No unnecessary `clsx`, MSW, jest-dom, provider SDK, or duplicate package added
- [ ] Existing frontend `.env` created without secrets
- [ ] Shared API base-path defect understood before real API testing
- [ ] Task 5 payment tests pass
- [ ] Full frontend typecheck passes
- [ ] Full frontend test suite passes
- [ ] Production build passes
- [ ] Vite starts and payment/refund routes render for approved test sessions
- [ ] Readonly and blocked-role behavior verified
- [ ] Backend absence is not mistaken for a frontend dependency failure

### Full-stack readiness

- [ ] Shared `/api/v1` URL convention fixed and integration-tested
- [ ] Runnable Auth, Gateway, Payment, and Superadmin services supplied
- [ ] Missing payment detail/refund/reconciliation contracts approved and implemented
- [ ] Unique Payment HTTP address/service DNS assigned; `8085` conflict resolved
- [ ] Sanitized backend `.env.example` supplied; real secrets in secret manager/ignored local env
- [ ] MySQL `payment_db` and `superadmin_db` created through reviewed versioned migrations
- [ ] Required tables, indexes, unique keys, and foreign keys verified
- [ ] Payment/attempt/refund/reconciliation enums and field mappings aligned
- [ ] API money guaranteed as integer minor units with ISO currency
- [ ] Redis reachable for current Gateway fail-closed rate limiting
- [ ] Provider sandbox, backend keys, public HTTPS webhook, signature verification, and rotation configured
- [ ] Webhook replay/idempotency and failed-event recovery tested
- [ ] Reconciliation schedule, timezone, report source, checkpoint, rerun, and freshness alert configured
- [ ] Finance admin account, MFA, backend permissions, and short session policy configured
- [ ] Refund review resource mapping, reason validation, maker-checker, locking, and idempotency implemented
- [ ] Review audit and provider command remain recoverable across partial failures
- [ ] Readonly admin mutation receives server-side `403`
- [ ] Synthetic list/filter/detail/attempt/refund fixtures verified
- [ ] Reconciliation matched/mismatch/missing-local/missing-provider fixtures verified
- [ ] No card data, tokens, provider secrets, raw credentials, or unsafe payloads exposed
- [ ] CORS, TLS, internal auth, timeouts, retries, and readiness verified
- [ ] Logs/metrics traced by request/event ID with sensitive values redacted
- [ ] Payment/refund/reconciliation backup, restore, migration rollback, and incident runbooks tested
- [ ] No duplicate setup documentation added

Frontend Task 5 setup is independently verifiable. Complete Payment Operations readiness tabhi claim karein when missing backend contracts/runtimes, migrations, provider integration, reconciliation pipeline, server-side authorization, and audited/idempotent refund consistency end to end pass ho.
