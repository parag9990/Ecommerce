# Project Dependency & Setup Guide

> **Scope:** Task 7 - CSV report export aur scheduled reports ke dependency, environment, database, aur DevOps requirements. Business logic aur implementation snippets yahan repeat nahi kiye gaye.
>
> **Read first:** Clone, Node/npm, Go, MongoDB, Redis, Docker, shared environment, ports, aur generic troubleshooting already `task1_Dependency.md` me documented hain. Ye guide sirf Task 7 ke new setup contract, current repository state, aur missing runtime pieces explain karti hai.

## 1. Project Overview

Task 7 dashboard me `/reports` route provide karta hai. Admin six analytics report types me se ek CSV download kar sakta hai aur daily, weekly, ya monthly delivery schedule configure kar sakta hai.

Frontend flow:

```text
Browser /reports (:5174)
  -> React page + TanStack Query
  -> immediate: GET /api/v1/analytics/reports/export
  -> schedules: GET/POST/PATCH/DELETE /api/v1/analytics/reports/schedules
  -> Vite /api proxy
  -> intended API Gateway (:8080) + admin authorization
  -> intended Session Service report handlers
  -> intended MongoDB aggregates + report_schedules
  -> future scheduler worker -> notification delivery
```

### Current repository reality

| Area | Status | Beginner note |
|---|---|---|
| `/reports` route, navigation, export form, schedule form/table | Implemented | Frontend page locally open ho sakta hai |
| CSV `Blob` download, safe filename, API validation/parser | Implemented | Extra CSV/file-saver package required nahi |
| Schedule list/create/pause/resume/delete frontend calls | Implemented | Real server response ke bina UI error dikhayegi |
| Unit/component/API tests | Implemented | Mocked browser/API behavior test hota hai |
| Report routes in `api/master-api.json` | Documented | Catalog auth ko `admin` mark karta hai |
| Session Service report HTTP handlers/use cases/repositories | **Missing** | Current server report routes par `404` deta hai |
| `report_schedules` collection migration/indexes | **Missing** | Schedule persistence ready nahi hai |
| Scheduler worker and Notification Service integration | **Missing** | Recurring report actually run/send nahi hoga |
| Runnable API Gateway and dashboard admin auth | **Missing** | Intended `:8080` browser flow end-to-end available nahi |

Isliye Task 7 ka frontend development aur most mocked verification possible hai. Real CSV aur scheduled delivery ke liye report backend, database migration, worker, notification integration, Gateway, aur authorization implement karna mandatory hai.

## 2. Tech Stack

### Reused technologies

React, TypeScript, Vite, Tailwind CSS, React Router, TanStack Query, Lucide, Vitest, Testing Library, MSW, Node/npm, Go modules, MongoDB, Redis, aur Docker pehle explain ho chuke hain.

Refer:
`task1_Dependency.md`

Sections: `2. Tech Stack`, `3. Required Software`, and `4. Dependency Management`

Analytics aggregate and operational background:

Refer:
`task4_Dependency.md`, `task5_Dependency.md`, and `task6_Dependency.md`

Sections: each file's `5. Database Setup` and `12. Security & Best Practices`

### Task 7-specific usage

| Technology/service | Simple Hinglish explanation | Task 7 purpose | Required? |
|---|---|---|---|
| Browser `Blob` + object URL | `Blob` browser ka file-like binary object hai. | Server CSV ko download karna without extra library | Yes; browser built-in, no install |
| TanStack Query mutations | Async create/update/delete/download state manage karta hai. | Export loading plus schedule CRUD and cache refresh | Yes; already installed |
| Native `Intl` timezone | Browser ka built-in timezone resolver/formatter hai. | Schedule ko IANA timezone, e.g. `Asia/Kolkata`, ke saath send karna | Yes; no install |
| MongoDB | Document database hai jisme flexible schedule documents store ho sakte hain. | Future `report_schedules` persistence and aggregate reads | Real backend ke liye mandatory |
| Scheduler worker | Background process hai jo due jobs automatically run karta hai. | `next_run_at` schedules ko daily/weekly/monthly execute karna | Scheduled delivery ke liye mandatory; absent |
| Notification integration | Generated report/link recipient tak pahunchata hai. | Schedule recipients ko delivery | Scheduled delivery ke liye mandatory; absent |
| Object storage | Generated file ko temporary durable location me store karta hai. | Large async report link delivery | Optional design choice; absent |

Papa Parse, `file-saver`, PDF/XLSX library, cron npm package, Kafka client, ya RabbitMQ client Task 7 implementation me use nahi hota. Inhe current checkout me install mat karo.

## 3. Required Software

Task 7 koi new OS-level software add nahi karta.

| Setup mode | Required software | Infrastructure needed? |
|---|---|---|
| Focused tests | Git, Node.js `22.x`, npm | No |
| Frontend dev/build | Above plus modern browser | No; live API error expected |
| Current Session Service startup | Go `1.26.3`, MongoDB, Redis | Yes; report routes phir bhi missing |
| Real immediate CSV | Above plus report backend, aggregate data, Gateway, admin auth | Yes |
| Real scheduled delivery | Above plus migration, scheduler worker, and notification provider/service | Yes; code/config absent |

Installation and version commands repeat nahi kiye gaye.

Refer:
`task1_Dependency.md`

Section: `3. Required Software`

## 4. Dependency Management

### No new npm dependency

Task 7 required packages already `frontend/session-analytics-dashboard/package.json` me declared hain:

```text
@tanstack/react-query, react, react-dom, react-router-dom, lucide-react
vitest, jsdom, @testing-library/react, @testing-library/user-event, msw
```

Native browser APIs CSV save karte hain, so task guide ka optional `papaparse` install current implementation ke liye required nahi hai.

Fresh clone install process:

Refer:
`task1_Dependency.md`

Section: `4. Dependency Management -> Node.js/npm system`

Task-specific check:

```bash
cd frontend/session-analytics-dashboard
npm ls @tanstack/react-query react react-router-dom vitest jsdom
```

### No new Go module

Task 7 ne `backend/services/session-service/go.mod` ya `go.sum` me dependency add nahi ki. MongoDB and Redis clients already declared hain, but report backend itself absent hai.

Go modules, `go mod download`, `go mod tidy`, build, proxy/cache, workspace, aur version mismatch:

Refer:
`task1_Dependency.md`

Section: `4. Dependency Management -> Go modules`

### Reproducibility warning

Dashboard package me checked-in `package-lock.json`, `pnpm-lock.yaml`, ya `yarn.lock` nahi hai. Clean install par transitive versions drift kar sakte hain. Team ko one package manager choose karke reviewed lockfile commit karna chahiye; multiple lockfiles mix mat karo.

## 5. Database Setup

### Reused MongoDB

MongoDB installation, local/Docker startup, named volume, port `27017`, `session_db`, URI, credentials, aur health verification unchanged hain.

Refer:
`task1_Dependency.md`

Section: `5. Database Setup -> MongoDB`

Immediate exports ko existing intended data paths chahiye:

| Report type | Intended source | Current real reader |
|---|---|---|
| Overview / active sessions | Redis live state plus session summaries | Missing |
| Journey summary | `journey_summaries` | Missing |
| Funnel / retention | `analytics_aggregates` | Missing |
| Heatmap | `heatmap_points` or its aggregates | Missing |

MongoDB server start karna alone export endpoint create nahi karta.

### New schedule persistence requirement

Real schedules ke liye a new `report_schedules` collection required hai. Minimum document contract ko organization/tenant scope, creator, report type, filters snapshot, cadence, timezone, recipients, status, `next_run_at`, `last_run_at`, timestamps, and schema version preserve karna chahiye.

Required query patterns ke liye planned indexes:

```javascript
db.report_schedules.createIndex({ status: 1, next_run_at: 1 })
db.report_schedules.createIndex({ created_by: 1, created_at: -1 })
```

Multi-tenant implementation me tenant/store key index prefix me include karo. Ye snippets design direction hain, current executable migration nahi. Final schema and worker query review ke baad versioned, idempotent up/down migration add karo.

Large async files persist karne par optional `report_exports` metadata collection and TTL index useful ho sakta hai. Synchronous CSV-only implementation me ye mandatory nahi.

### Migration status

Current only migration:

```text
backend/services/session-service/migrations/001_privacy_controls.mongodb.js
```

Ye `report_schedules`, `report_exports`, ya analytics report indexes create nahi karti. **Task 7 migration command currently exist nahi karta.** Hand-written production index commands blindly run mat karo; backup, uniqueness/cardinality, rollback, and worker compatibility review karo.

### Redis

Redis installation, Docker setup, password, logical DB `2`, port `6379`, and `PING` verification unchanged hain.

Refer:
`task1_Dependency.md`

Section: `5. Database Setup -> Redis`

Current Go server startup par Redis mandatory ping karta hai. Task 7-specific cache, distributed schedule lock, idempotency key, ya worker checkpoint code absent hai. Multi-replica worker implement karte waqt duplicate delivery prevent karne ke liye atomic job claim/lease mandatory hoga.

## 6. Redis / Queue / External Services

### API Gateway and admin authentication

Master API catalog now declares:

```text
GET    /api/v1/analytics/reports/export
GET    /api/v1/analytics/reports/schedules
POST   /api/v1/analytics/reports/schedules
PATCH  /api/v1/analytics/reports/schedules/{schedule_id}
DELETE /api/v1/analytics/reports/schedules/{schedule_id}
auth: admin
service: session-service
```

Frontend every request me `credentials: "include"` use karta hai. Production Gateway ko admin session/JWT validate, report-specific RBAC, tenant/store scope, CSRF policy for state-changing calls, rate limit, request ID, and response limits enforce karne honge.

Shared Gateway limitation:

Refer:
`task2_Dependency.md`

Section: `6. Redis / Queue / External Services -> API Gateway and admin auth`

### Scheduler and notification delivery

Task 7 frontend schedule config save/manage karta hai; repository me schedule runner or report-to-notification integration nahi hai. Real delivery ke liye these responsibilities assign karo:

1. Worker due active schedules atomically claim kare.
2. UTC `next_run_at` and selected IANA timezone correctly calculate kare.
3. Aggregate-based CSV generate/stream kare.
4. Notification Service/provider ko attachment ya short-lived authorized link de.
5. Success/failure, retry count, `last_run_at`, `next_run_at`, and sanitized error update kare.

Notification transport, credentials, sender identity, attachment limit, storage provider, and retention policy choose nahi kiye gaye. Isliye current setup me SMTP/S3 credentials add ya Notification Service start karna Task 7 ko functional nahi banayega.

### Queue/broker and object storage

Kafka, RabbitMQ, NATS, S3, MinIO, or cron library current Task 7 code/config me declared nahi. Immediate CSV and simple single-worker schedules ke liye broker optional ho sakta hai. Large reports/retries ke liye durable queue and object storage later architecture choice ho sakte hain; implementation, topic/queue, DLQ, credentials, health, and cleanup policy ke bina install mat karo.

## 7. Environment Variables

### Frontend: no new variable

Task 7 existing values reuse karta hai:

| Variable | Report behavior | Required? |
|---|---|---|
| `VITE_API_BASE_URL` | Report requests ka base; empty means same-origin `/api` | No; local proxy ke liye empty recommended |
| `VITE_API_PROXY_TARGET` | Local `/api` target; default `http://localhost:8080` | No |
| `VITE_REQUEST_TIMEOUT_MS` | CSV and schedule request timeout; default `10000` ms | No; positive integer |

`.env.local` location, shared example, Vite loading/restart, CORS, and secret rules:

Refer:
`task1_Dependency.md`

Section: `7. Environment Variables -> Frontend .env.local`

No new Task 7 `.env` block is needed. Large export implementation ke liye current 10-second client timeout likely too small ho sakta hai; server latency measure karke bounded value choose karo or async export design use karo. Randomly very large timeout set karke backend performance issue hide mat karo.

### Backend report variables do not exist yet

Current `internal/config/config.go` report size, schedule worker, storage, notification, retention, or retry variable parse nahi karta. Current backend `.env` me bhi active Task 7 variables nahi hain. Future names guess karke copy-paste mat karo.

Backend implementation se pehle a reviewed config contract define karo for:

| Configuration area | Example meaning | Secret? |
|---|---|---|
| Maximum export rows/bytes/range | Resource abuse control | No |
| Worker poll interval, batch, lease, retries | Schedule execution behavior | No |
| Generated-file TTL | Cleanup and privacy window | No |
| Storage bucket/endpoint | Async report location | Usually no |
| Storage access key/secret | Object store authentication | **Yes** |
| Notification sender/provider credentials | Email/delivery authorization | **Yes** |

Non-secret defaults sanitized `.env.example` me, secrets runtime secret manager/env injection me, typed validation `config.go` me, and config tests code ke saath add hon. `VITE_` prefix me backend secret kabhi mat rakho.

Current Session Service loaded Mongo/Redis/HTTP/privacy variables and `.env` source commands:

Refer:
`task1_Dependency.md`

Section: `7. Environment Variables -> Current Session Service environment`

## 8. Docker Setup

Task 7 koi Dockerfile, Compose service, image, port, volume, network, health check, ya restart policy add nahi karta.

MongoDB + Redis local container setup:

Refer:
`task1_Dependency.md`

Section: `8. Docker Setup`

That setup only common infrastructure start karta hai. It does not provide Gateway, report backend, schedule migration, worker, Notification Service integration, or object storage.

Future scheduler container ke DevOps requirements:

- API process se separate worker command/image entrypoint;
- non-root user, pinned image, read-only filesystem where possible;
- dependency-aware readiness and lightweight liveness;
- graceful shutdown that releases/extends job leases safely;
- retry with exponential backoff and dead-letter/manual recovery state;
- one schedule occurrence ki idempotency key to prevent duplicate email;
- memory/CPU limits and maximum CSV size;
- worker lag, duration, success/failure, rows/bytes, and delivery metrics;
- runtime secrets, private network, TLS, and generated-file expiry;
- permanent config/schema errors ko tight restart loop me hide na karne wali policy.

## 9. Local Development Setup

### Step 1: Read previous setup guides

1. `task1_Dependency.md`: clone, tools, install, env, MongoDB/Redis, Docker, ports, generic errors.
2. `task2_Dependency.md`: Gateway/admin-auth limitations.
3. `task3_Dependency.md`: journey data privacy and ordering.
4. `task4_Dependency.md`: funnel aggregates and Recharts context.
5. `task5_Dependency.md`: heatmap aggregate/storage context.
6. `task6_Dependency.md`: latest cohort aggregate and backend gap audit.

### Step 2: Move to dashboard package

```bash
cd frontend/session-analytics-dashboard
```

### Step 3: Install only when needed

No Task 7-only install command hai. Fresh clone or missing `node_modules` case me Task 1 ka normal `npm install` follow karo. Optional Papa Parse install mat karo.

### Step 4: Reuse frontend environment

`frontend/session-analytics-dashboard/.env.local` me existing three `VITE_` values reuse karo. New report secret or provider credential add mat karo.

### Step 5: Run focused verification

```bash
npm run test -- \
  src/api/session-api.test.ts \
  src/features/reports/components/export-report-panel.test.tsx \
  src/features/reports/lib/report-filenames.test.ts \
  src/features/reports/lib/save-csv-blob.test.ts
npm run typecheck
npm run build
```

Tests fake fetch responses use karte hain; MongoDB, Redis, Gateway, worker, and notification service required nahi.

> **Current verification note (2026-06-19):** Correct dashboard working directory and jsdom config ke saath 27/28 focused assertions pass. `session-api.test.ts` CSV assertion fails because this installed jsdom `Blob` lacks `.text()`. `npm run typecheck` equivalent fails on existing test casts, Node `Error` constructor typing, and four Task 7 React Query mutation function signatures. Since normal build runs `tsc -b` first, `npm run build` is blocked until those typing errors are fixed. Ye infrastructure installation failure nahi; code/test compatibility issue hai.

### Step 6: Start frontend

```bash
npm run dev
```

Open:

```text
http://localhost:5174/reports
```

Vite `strictPort: false` use karta hai. `5174` occupied ho to terminal me printed fallback URL open karo. Current real API absence ke karan schedules error state expected hai; download click bhi fail hoga.

### Step 7: Optional current backend

Current Go service inspect karna ho to Task 1 ke MongoDB + Redis start, env load, and `GOWORK=off go run ./cmd/server` steps follow karo. Existing privacy migration run karna reports initialize nahi karta.

### Step 8: Verify API and known blocker

Intended Gateway export check:

```bash
curl -i \
  -H 'Accept: text/csv' \
  'http://localhost:8080/api/v1/analytics/reports/export?report_type=funnel&format=csv&from=2026-06-01&to=2026-06-19&timezone=UTC'
```

Current checkout me Gateway connection failure expected hai. Direct current Session Service:

```bash
curl -i http://localhost:8086/healthz
curl -i \
  -H 'Accept: text/csv' \
  'http://localhost:8086/api/v1/analytics/reports/export?report_type=funnel&format=csv&from=2026-06-01&to=2026-06-19&timezone=UTC'
# Current expectation: /healthz can return 200; report export returns 404.
```

Proxy target `8086` karna missing route implement nahi karta. Final Gateway request valid admin auth require karega; cookie/token command history me hardcode mat karo.

## 10. Running the Project

### Daily frontend workflow

```bash
cd frontend/session-analytics-dashboard
npm run test -- src/features/reports
npm run dev
```

Manual checks:

1. `/reports` route shell and Reports navigation ke through open ho.
2. Six allowed report types and CSV-only badge show hon.
3. Date and segment changes request query me reflect hon; `all` filters omit hon.
4. Browser timezone valid IANA value ke roop me send ho.
5. Export pending state duplicate click block kare and success file download kare.
6. Server filename unsafe characters sanitize hon; fallback filename `.csv` ho.
7. Daily/weekly/monthly fields change hon and invalid recipient block ho.
8. Schedule create ke baad list refresh ho; pause/resume and confirmed delete work kare.
9. Loading, empty, `401`, `403`, `404`, timeout, server error, and failed schedule states verify hon.
10. CSV me raw PII/sensitive payload absent ho and narrow viewport table scroll kare.

### Frontend request contract

Immediate export required query:

```text
report_type, format=csv, from, to, timezone
```

Optional filters:

```text
device, channel, source, country, user_type
```

Successful response should include:

```http
HTTP/1.1 200 OK
Content-Type: text/csv; charset=utf-8
Content-Disposition: attachment; filename="funnel_2026-06-01_2026-06-19.csv"
Cache-Control: no-store
```

Schedule create accepts CSV format, one of six report types, daily/weekly/monthly cadence, `HH:mm`, valid IANA timezone, filters snapshot, and `1..10` unique valid recipient emails. Schedule name maximum is 120 readable characters. Monthly day must be `1..31`; backend must define shorter-month behavior explicitly.

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Vite dev server | `5174` | `/reports` UI and `/api` proxy | Reused |
| Vite preview | `4174` | Built frontend preview | Reused |
| Intended API Gateway | `8080` | Admin report API entry | Reused; executable missing |
| Current Session HTTP | `8086` | Privacy APIs and `/healthz` | Reused; report routes missing |
| Intended Session gRPC | `50060` | Catalog downstream report methods | Reused design; implementation missing |
| MongoDB | `27017` | Aggregates and future schedules | Reused |
| Redis | `6379` | Current startup/live state; future locks possible | Reused |
| Notification/worker/storage | N/A | Scheduled delivery | New capability, no port/config selected |

Port conflict, firewall, CORS, cookie, and Docker networking details:

Refer:
`task1_Dependency.md`

Section: `10. Running the Project -> Ports and networking`

## 11. Common Errors & Fixes

Generic npm, Go, Docker, MongoDB, Redis, CORS, ports, and permission errors:

Refer:
`task1_Dependency.md`

Section: `11. Common Errors & Fixes`

Task 7-specific issues:

| Error/symptom | Cause | Fix and prevention |
|---|---|---|
| Reports page says schedules unavailable | `GET .../schedules` cannot reach implemented backend | Network URL/status check; current repo me expected blocker; route/repository/auth implement karo |
| Export click shows failure/no file | Export endpoint `404`, auth failure, timeout, or non-CSV response | Network response inspect; handler, auth, `Content-Type`, and aggregate source verify |
| Direct Session Service returns `404` | Current router only registers privacy routes | Proxy change nahi; report handler/use case/repository implement and register karo |
| `401` / `403` | Admin session missing or report permission absent | Gateway login, cookie/JWT policy, tenant scope, and RBAC verify |
| CSV request times out at 10 seconds | Full generation/blob download exceeds frontend timeout | Query optimize/stream, bounded limit, or async export; measured timeout configure karo |
| Download filename is generic | Server omitted/invalid `Content-Disposition` | RFC-compatible attachment filename return karo; frontend safe fallback already exists |
| Schedule create returns validation error | Invalid timezone/time/email/cadence/filter | Request payload inspect; unique `1..10` emails, IANA timezone, valid date range use karo |
| Monthly run missing on day 29-31 | Shorter-month policy undefined | Backend contract choose: last day, skip, or reject; tests/docs align karo |
| Same report delivered twice | Multiple workers claimed same occurrence | Atomic lease + occurrence idempotency key + provider idempotency implement karo |
| Schedule stuck `failed` | Retry budget/provider/storage failure | Sanitized `last_error`, retry/backoff, metrics, and manual resume/retry workflow provide karo |
| `Blob.text is not a function` in focused test | Installed jsdom Blob API differs from browser/Node expectation | `await new Response(blob).text()` compatible assertion or test polyfill use karo; production browser manually verify |
| Typecheck rejects mutation functions | React Query passes mutation context as second arg, API function second arg expects request options | Hooks me one-argument wrapper use karo, e.g. `mutationFn: (input) => createReportSchedule(input)` |
| `npm` says WSL 1 unsupported | Shell resolves Windows npm while Linux Node is active | Node/npm same environment me install/select karo; `command -v node npm` verify karo |

## 12. Security & Best Practices

- Export and schedule endpoints server-side admin/RBAC protect karo; hidden UI link authorization nahi hai.
- Every query and schedule tenant/store scope server-derived ho. Client-provided scope blindly trust mat karo.
- CSV formula injection prevent karo: cells starting `=`, `+`, `-`, or `@` neutralize/quote policy ke saath test karo.
- CSV me password, OTP, token, card data, raw IP, email/phone, address, private form content, or raw event payload export mat karo.
- Aggregate-first generation use karo; raw `session_events` unbounded scan avoid karo.
- `Cache-Control: no-store`, attachment response, TLS, size/range/rate limits, and audit event apply karo.
- Recipient allowlist/domain/RBAC policy define karo; schedule creator ko arbitrary external exfiltration permission automatically mat do.
- Scheduled links short-lived, authorized, single-purpose hon. Public permanent object URLs avoid karo.
- Logs/audit me report type, scope, actor, row/byte count, status, and request ID useful hain; CSV content/recipient secrets/cookies log mat karo.
- Worker atomic claim and idempotency enforce kare; retries duplicate delivery nahi banayen.
- Time calculation trusted timezone library/runtime se ho; `next_run_at` UTC store karo and DST edge cases test karo.
- Generated report retention/deletion privacy policy se align ho; backup/object versions bhi lifecycle policy follow karein.
- Spreadsheet consumers ke liye UTF-8, correct escaping, stable columns, schema version, and dangerous formulas test karo.
- Dependency lockfile commit and CI me typecheck, unit, integration, security, migration, and end-to-end download tests run karo.

## 13. Missing or Misconfigured Things

| Severity | Finding | Professional fix |
|---|---|---|
| Blocker | Current Session Service report export/schedule routes implement/register nahi karta | Handlers, use cases, repositories, validation, CSV writer, and contract/integration tests add karo |
| Blocker | Runnable API Gateway/admin-auth integration absent hai | Catalog routes, auth/RBAC/CSRF, tenant context, proxy/transport, and readiness implement karo |
| Blocker | `report_schedules` migration and repository absent hain | Versioned schema/index migration, rollback, repository, and query tests add karo |
| Blocker | Scheduler worker and notification delivery absent hain | Atomic due-job worker, retry/idempotency, provider integration, and runbook implement karo |
| High | Report aggregates/readers for six export types incomplete hain | Canonical aggregate contracts, indexed queries, privacy filters, and realistic seed/ingestion path add karo |
| High | Catalog declares gRPC report methods but current Session Service exposes unrelated HTTP privacy routes only | One explicit Gateway-to-service transport implement and contracts align karo |
| High | Default 10-second client timeout applies while CSV response is fully read into a `Blob` | Bounded sync reports or async job/link flow choose; size/time limits and streaming tests add karo |
| High | CSV injection, row/byte limits, audit, and `no-store` are frontend/task intentions only | Backend enforcement and security tests add karo |
| Medium | Task 7 guide says source/routes were only planned, but source and master catalog now exist | Original task history preserve karo; dependency guide/current docs ko runtime truth ke saath maintain karo |
| Medium | Normal TypeScript/build gate fails, including four Task 7 mutation signatures | React Query wrappers and test typings fix; green `typecheck` and `build` require karo |
| Medium | One CSV API test assumes jsdom Blob `.text()` | Cross-runtime-compatible assertion/polyfill use karo |
| Medium | Frontend dependency lockfile absent hai | One package manager choose; reviewed lockfile commit and CI `npm ci` use karo |
| Medium | No application Dockerfiles/checked-in Compose/report worker deployment | Reproducible non-root images, health checks, secrets, resources, and private networks add karo |
| Medium | No Task 7 backend config contract or sanitized example | Implemented variables only typed-load/validate/document karo; guessed inactive values avoid karo |
| Operational | No delivery metrics, worker lag, report duration/size, alerts, backup, or restore runbook | Observability SLOs, alerts, dashboards, backup/restore drill, and incident procedures add karo |

Full Task 7 setup complete tab maana jayega jab:

1. Admin authenticates and report RBAC passes.
2. Gateway routes all five method/path combinations correctly.
3. Immediate export returns privacy-safe CSV with correct headers within documented limits.
4. Mongo migration and aggregate readers are deployed and verified.
5. Schedule create/list/pause/resume/delete persists tenant-scoped data.
6. Worker executes each due occurrence once and notification reaches allowed recipient.
7. Failure/retry/audit/metrics/file-expiry behavior is observable.
8. Typecheck, build, unit, integration, and browser download tests pass.

## 14. References to Previous Dependency Files

| Previous dependency file | Section/topic reused | Why reused |
|---|---|---|
| `task1_Dependency.md` | Tech stack and required software | Same React/Node/Go toolchain |
| `task1_Dependency.md` | Node/npm and Go modules | No new package/module |
| `task1_Dependency.md` | MongoDB and Redis setup | Same instances, ports, credentials, and Docker setup |
| `task1_Dependency.md` | Environment, Docker, ports, generic errors | Shared setup unchanged |
| `task2_Dependency.md` | API Gateway and admin auth | Same missing Gateway/auth integration |
| `task3_Dependency.md` | Journey ordering/privacy context | Journey summary export reuses safe session data rules |
| `task4_Dependency.md` | Funnel aggregates and analytics operations | Funnel CSV reuses same source model |
| `task5_Dependency.md` | Heatmap storage/aggregation | Heatmap CSV reuses aggregate data path |
| `task6_Dependency.md` | Cohort aggregates and latest gap audit | Retention CSV reuses cohort read model |

## 15. Final Checklist

### Frontend readiness

- [ ] Previous dependency guides read kiye
- [ ] Node `22.x` and npm same environment se resolve hote hain
- [ ] Existing dependencies installed; no unnecessary CSV library added
- [ ] Existing `.env.local` reused; no secret `VITE_` variable me added
- [ ] Focused Task 7 tests run kiye and known Blob failure understood/fixed
- [ ] Typecheck and normal build green kiye
- [ ] `/reports` route and responsive navigation verified
- [ ] Export query, CSV headers, safe filename, and error/loading states verified
- [ ] Schedule create/list/pause/resume/delete states verified

### Real backend and DevOps readiness

- [ ] Gateway and admin report RBAC/CSRF implemented
- [ ] Six report readers and CSV writer implemented with privacy/size limits
- [ ] `report_schedules` schema/index migration reviewed and applied
- [ ] MongoDB aggregates populated; Redis role explicitly defined
- [ ] Schedule worker atomic claim, retry, idempotency, and timezone tests pass
- [ ] Notification/storage provider contract, secrets, health, and expiry configured
- [ ] Docker/deployment health checks, resources, networking, and secret injection verified
- [ ] `200`, empty, `400`, `401`, `403`, `404`, timeout, provider failure, and duplicate-run tests pass
- [ ] Audit logs and report/worker/delivery metrics checked
- [ ] CSV contains no prohibited PII or formula-injection payload
- [ ] No duplicate shared setup documentation added
