# Project Dependency & Setup Guide

> **Scope:** Task 6 - Retention Reports ke dependency, environment, database, aur DevOps requirements. Business logic aur implementation snippets yahan repeat nahi kiye gaye.
>
> **Read first:** Clone, Node/npm, Go, MongoDB, Redis, Docker, shared environment, ports, aur generic troubleshooting already `task1_Dependency.md` me documented hain. Ye guide sirf retention-report-specific setup aur repository gaps explain karti hai.

## 1. Project Overview

Task 6 dashboard me `/cohorts` route add karta hai. Admin date range, segment, interval, aur window choose karke new-vs-returning users aur privacy-safe cohort retention matrix dekhta hai.

Intended runtime flow:

```text
Browser /cohorts (:5174)
  -> React + TanStack Query
  -> GET /api/v1/analytics/retention
  -> Vite proxy
  -> intended API Gateway (:8080) + admin authorization
  -> intended SessionService.GetRetentionReport
  -> intended MongoDB analytics_aggregates
  -> bounded sessions/session_events fallback, if explicitly implemented
```

### Current repository reality

| Area | Status | Beginner note |
|---|---|---|
| `/cohorts` page, route, controls, chart, matrix, and privacy states | Implemented | Frontend locally run ho sakta hai |
| API request validation and defensive response parser | Implemented | Snake-case/camel-case aggregate response handle hota hai |
| Vitest, Testing Library, and MSW tests | Implemented | Real database/backend required nahi |
| `GET /api/v1/analytics/retention` master API contract | Documented | Route catalog me available hai |
| Session Service retention report handler/use case/repository | **Missing** | Current `:8086` server report return nahi karta |
| Cohort aggregation worker/read model | **Missing** | Real aggregate generate nahi hota |
| Executable `analytics_aggregates` retention migration | **Missing** | Design-doc index migration nahi hota |
| Runnable API Gateway and dashboard admin-auth wiring | **Missing** | Intended `:8080` flow end-to-end available nahi |

Isliye frontend tests, typecheck, build, aur mocked UI development possible hain. Real retention report ke liye backend implementation, data ingestion/identity linking, aggregation, migration, Gateway, aur authorization pehle complete karne honge.

## 2. Tech Stack

### Reused technologies

React, TypeScript, Vite, Tailwind CSS, React Router, TanStack Query, Recharts, Lucide, Vitest, Testing Library, MSW, Node/npm, Go modules, MongoDB, Redis, aur Docker pehle explain ho chuke hain.

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Sections: `2. Tech Stack`, `3. Required Software`, and `4. Dependency Management`

Recharts ka beginner explanation aur chart dependency verification:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task4_Dependency.md`

Sections: `2. Tech Stack -> Task 4-specific technology usage` and `4. Dependency Management -> Recharts verification`

### Task 6-specific usage

| Technology/service | Simple Hinglish explanation | Task 6 purpose | Required? |
|---|---|---|---|
| Recharts | React chart library hai jo structured data ko responsive chart me dikhati hai. | New aur returning users ka stacked bar chart | Frontend build ke liye yes; already installed |
| TanStack Query | API loading, cache, cancellation, aur refresh state manage karta hai. | Retention filters ke hisaab se query; `60s` stale time | Yes; already installed |
| Native HTML table/CSS | Browser ka built-in table layout hai; extra grid package nahi. | Accessible cohort matrix aur masked cells | Yes; no install |
| MSW | Tests me fake HTTP endpoint provide karta hai. | Backend ke bina success, empty, error, and filter behavior test | Dev/test only; already installed |
| MongoDB aggregates | Precomputed analytics documents ko fast read karne ka intended store hai. | Cohort sizes, offset buckets, new/returning totals | Real data ke liye mandatory |
| Redis | Fast in-memory store hai. | Task 6 code me retention cache/worker use absent; current Go process startup par Redis ping karta hai | Frontend ke liye no; current Go server ke liye yes |

Task 6 ke liye D3, a heatmap package, Kafka client, RabbitMQ client, ya another cohort library install mat karo. Current matrix native React/table markup use karti hai.

## 3. Required Software

Task 6 koi new OS-level software add nahi karta.

| Setup mode | Required software | Backend infrastructure needed? |
|---|---|---|
| Focused unit/component tests | Git, Node.js `22.x`, npm | No |
| Frontend dev/build | Above plus modern browser | No; API error expected without mock/backend |
| Current Session Service startup | Go `1.26.3`, MongoDB, Redis | Yes |
| Full real-data retention | Above plus future Gateway, ingestion, identity linking, worker, migration, report backend | Yes; required code is currently missing |

Install/version commands repeat nahi kiye gaye.

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `3. Required Software`

## 4. Dependency Management

### No new npm dependency

Task 6 ki required packages already `frontend/session-analytics-dashboard/package.json` me declared hain:

```text
@tanstack/react-query, recharts, react, react-dom, react-router-dom, lucide-react
vitest, jsdom, @testing-library/react, @testing-library/user-event, msw
```

Fresh clone install process:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `4. Dependency Management -> Node.js/npm system`

Task-specific dependency check:

```bash
cd frontend/session-analytics-dashboard
npm ls @tanstack/react-query recharts msw
```

Task guide me shown individual `npm install ...` commands ko already configured checkout par rerun karna unnecessary manifest churn create kar sakta hai. Normal package install enough hai.

### No new Go module

Task 6 ne `backend/services/session-service/go.mod` or `go.sum` me dependency add nahi ki. MongoDB driver and Redis client already declared hain; retention report backend itself absent hai.

Go modules, `go mod download`, `go mod tidy`, `go build`, version mismatch, workspace, proxy/cache issues:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `4. Dependency Management -> Go modules`

### Reproducibility warning

Dashboard package me checked-in `package-lock.json`, `pnpm-lock.yaml`, or `yarn.lock` nahi hai. Clean install par transitive versions drift kar sakte hain. Team ko one package manager select karke reviewed lockfile commit karna chahiye; multiple lockfiles mix mat karo.

## 5. Database Setup

### MongoDB: reused server, new retention read model

MongoDB ek document database hai. Session/event volume aur flexible aggregate buckets ke liye intended storage hai. Installation, Docker container, named volume, port `27017`, `session_db`, URI, credentials placement, and health verification unchanged hain.

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `5. Database Setup -> MongoDB`

Real Task 6 flow ko logically ye data chahiye:

| Collection | Retention purpose | Current status |
|---|---|---|
| `sessions` / identity source | First-seen time aur logged-in/anonymous identity link establish karna | Design context exists; report reader absent |
| `session_events` | Later interval me identity active thi ya nahi derive karna | Design docs only; complete ingestion/report path absent |
| `analytics_aggregates` | Precomputed cohort sizes and retained offset buckets fast serve karna | Generic design index only; writer/reader/migration absent |

Current database design documents only this generic index:

```javascript
db.analytics_aggregates.createIndex(
  { metric: 1, bucket: 1, segment: 1 },
  { unique: true }
)
```

Retention implementation me aggregate identity ko at least tenant/store scope, metric/schema version, cohort bucket, interval, offset/window, and normalized segment filters distinguish karne honge. Exact schema/query finalize kiye bina extra index blindly create mat karo. Query `explain`, cardinality, uniqueness, write cost, backup, and rollback review ke baad versioned idempotent migration banao.

### Required data rules

- Logged-in identity ke liye stable `user_id`; anonymous traffic ke liye privacy-safe `anonymous_id` use ho.
- Anonymous-to-user linking same person ko double count na kare.
- One identity ko one bucket me multiple sessions ke baad bhi once count karo.
- Cohort size and retained counts non-negative integers hon; rate server-side safe divide se `0..100` me ho.
- Consent-disabled/deleted identities aggregation input se exclude hon.
- Small cohorts API response banne se pehle suppress/coarsen hon; frontend flag alone security control nahi hai.
- Timezone and interval boundary canonical hon, warna day/week cohort different environments me shift ho sakta hai.

### Migration status

Current Session Service migration only privacy/deletion indexes manage karti hai:

```text
backend/services/session-service/migrations/001_privacy_controls.mongodb.js
```

Ye retention aggregate collection/index/worker checkpoint create nahi karti. Task 6-specific migration command currently **exist nahi karta**, so design-doc JavaScript ko completed migration samajhkar run mat karo.

### Redis

Redis install, Docker, password, logical DB `2`, port `6379`, and `PING` verification same hain.

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `5. Database Setup -> Redis`

Current Go server startup par Redis ping mandatory hai. Lekin Task 6 retention cache, distributed lock, aggregation checkpoint, TTL, or invalidation implementation repository me nahi hai. Redis start karne se missing retention API ready nahi hogi.

## 6. Redis / Queue / External Services

### API Gateway and admin authentication

Master API catalog intended contract:

```text
GET /api/v1/analytics/retention
service: session-service
downstream: SessionService.GetRetentionReport
auth: admin
```

Frontend `credentials: "include"` ke saath request bhejta hai. Production Gateway ko admin session/JWT validate, RBAC enforce, tenant/store scope inject, rate limit, request ID propagate, and response-size limits enforce karne honge.

Current Gateway executable and dashboard auth integration missing hain. Shared Gateway limitation:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task2_Dependency.md`

Section: `6. Redis / Queue / External Services -> API Gateway and admin auth`

### Kafka or RabbitMQ

Architecture me high-volume analytics aggregation ke liye queue/broker useful ho sakta hai, but Task 6 code/config/manifests me Kafka, RabbitMQ, or NATS client nahi hai. Current setup ke liye broker install/start **mat** karo. Broker tab mandatory hoga jab producer, consumer, topic/queue, retry/DLQ, credentials, lag metrics, and health checks actually implemented hon.

No third-party SaaS credential Task 6 introduce karta hai.

## 7. Environment Variables

### Frontend: no new variable

Task 6 existing frontend values reuse karta hai:

| Variable | Retention effect | Required? |
|---|---|---|
| `VITE_API_BASE_URL` | Retention fetch base; empty means same-origin `/api` | No; local proxy ke liye empty recommended |
| `VITE_API_PROXY_TARGET` | Local `/api` proxy target; default `http://localhost:8080` | No |
| `VITE_REQUEST_TIMEOUT_MS` | Unreachable/slow report request timeout; default `10000` ms | No |

`.env.local` location, complete shared example, Vite loading behavior, restart rule, direct-origin CORS, and security:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `7. Environment Variables -> Frontend .env.local`

No new Task 6 `.env` block is needed. `VITE_` values browser bundle me public hoti hain; Mongo URI, Redis password, admin token, hash pepper, or identity salt kabhi frontend env me mat rakho.

### Retention backend names are currently inactive

Ignored local file `backend/services/session-service/.env` retention-related names list karti hai:

```env
SESSION_AGGREGATE_RETENTION_YEARS=5
SESSION_RETENTION_WORKER_BATCH_SIZE=1000
SESSION_RETENTION_WORKER_INTERVAL=60m
SESSION_RETENTION_WORKER_INTERVAL_MINUTES=60
SESSION_RETENTION_DRY_RUN=false
SESSION_RETENTION_WORKER_RUN_ONCE=false
SESSION_SMALL_COHORT_THRESHOLD=5
```

| Variable/group | Intended purpose | Required now? | Audit note |
|---|---|---|---|
| Aggregate retention | Long-term aggregate expiry policy | No | Current config loader ignores it; legal/privacy setting must be authoritative |
| Worker batch/interval | Scheduled cohort aggregation cadence and load | No | Worker absent; duplicate interval names create ambiguity |
| Dry-run/run-once | Operational worker modes | No | Loader/worker absent; production startup mode validate karo |
| Small cohort threshold | Low-count suppression cutoff | No active effect | Frontend only API metadata display karta hai; backend enforcement missing |

`internal/config/config.go` in names ko parse nahi karta. File me value edit karne se Task 6 behavior change nahi hoga. Backend add karte waqt one canonical interval variable choose karo, typed parsing/defaults/ranges add karo, config tests likho, aur secret-free startup summary expose karo.

Current Session Service ke actually loaded Mongo/Redis/HTTP/privacy variables and Bash/PowerShell load commands:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `7. Environment Variables -> Current Session Service environment`

Go service `.env` automatically load nahi karta; process start se pehle variables export karne hote hain.

## 8. Docker Setup

Task 6 koi new Dockerfile, Compose service, image, port mapping, volume, network, health check, or restart policy add nahi karta.

MongoDB + Redis local container setup reuse karo:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `8. Docker Setup`

That setup only infrastructure start karta hai. Full retention stack nahi, because repository me application Dockerfiles, runnable Gateway, retention backend, and aggregator worker absent hain.

Future retention worker/container me ye DevOps controls required honge:

- non-root image and pinned dependencies;
- separate liveness and dependency-aware readiness;
- Mongo/Redis/broker service hostname, container-local `localhost` nahi;
- idempotent checkpointing, retry/backoff, graceful shutdown, and bounded batch size;
- aggregate freshness/worker lag/suppression/error metrics;
- resource limits and backpressure;
- secrets runtime injection, image or Compose source me credentials nahi;
- restart policy jo permanent config/schema error ko infinite tight loop me hide na kare.

## 9. Local Development Setup

### Step 1: Read previous setup guides

1. `task1_Dependency.md`: clone, tools, install, frontend env, MongoDB/Redis, Docker, and shared errors.
2. `task2_Dependency.md`: Gateway/admin-auth limitations.
3. `task3_Dependency.md`: identity-safe `session_events`, ordering, retention, and privacy context.
4. `task4_Dependency.md`: Recharts and aggregate analytics operational principles.
5. `task5_Dependency.md`: latest shared backend/Docker gap audit.

### Step 2: Move to dashboard package

```bash
cd frontend/session-analytics-dashboard
```

### Step 3: Install only when needed

No Task 6-only install command hai. Fresh clone or missing `node_modules` case me Task 1 ka normal install follow karo.

### Step 4: Reuse frontend environment

`frontend/session-analytics-dashboard/.env.local` me existing three `VITE_` values reuse karo. New retention value ya secret add mat karo.

### Step 5: Run focused verification

```bash
npm run test -- \
  src/api/session-api.test.ts \
  src/features/cohorts/lib/retention-math.test.ts \
  src/features/cohorts/pages/cohort-retention-page.test.tsx
npm run typecheck
npm run build
```

Tests MSW fake retention response use karte hain; MongoDB, Redis, Gateway, and Session Service required nahi. Page test Recharts component mock karta hai, so actual responsive chart browser me manually verify karo.

> **Current verification note (2026-06-19):** Focused retention API validation/parser test and all four retention-math tests pass. Cohort page suite me error-state test pass hota hai, but remaining four tests current Linux Node environment me MSW response ki jagah `NETWORK_ERROR` receive karte hain. `npm run typecheck` repo-wide existing test/report-hook typing errors par fail hota hai; because `npm run build` pehle `tsc -b` chalata hai, normal build command bhi blocked hai. Direct Vite bundling succeeds, but that TypeScript gate bypass karta hai and completion proof nahi hai. Troubleshooting and audit tables below dekho.

### Step 6: Start frontend

```bash
npm run dev
```

Open:

```text
http://localhost:5174/cohorts
```

Vite `strictPort: false` use karta hai. `5174` occupied ho to terminal me printed fallback URL open karo.

### Step 7: Migration and backend services

Frontend-only development ke liye migration/service needed nahi. Current Go service inspect karna ho to reused MongoDB + Redis start/load/run steps follow karo, but no Task 6 migration/worker exists. Privacy migration retention analytics create nahi karti.

### Step 8: Verify intended API and known blocker

```bash
curl -i \
  'http://localhost:8080/api/v1/analytics/retention?from=2026-06-01&to=2026-06-19&interval=week&window=8&device_type=mobile'
```

Current checkout me Gateway connection failure expected hai. Direct current Session Service check:

```bash
curl -i http://localhost:8086/healthz
curl -i \
  'http://localhost:8086/api/v1/analytics/retention?from=2026-06-01&to=2026-06-19&interval=week&window=8'
# Current expectation: /healthz can return 200; retention route returns 404.
```

Proxy target `8086` karna missing route implement nahi karta. Final Gateway request valid admin authentication require karega.

## 10. Running the Project

### Daily frontend workflow

```bash
cd frontend/session-analytics-dashboard
npm run test -- src/features/cohorts
npm run dev
```

Manual browser checks:

1. `/cohorts` route navigation shell ke andar open ho.
2. Date range, device, channel, source, user type, interval, and window change karo.
3. Network request me `from`, `to`, `interval`, `window`, and selected segment params verify karo.
4. Success fixture/integration par summary cards, stacked chart, legend, and matrix verify karo.
5. Daily windows `7/14/30`, weekly `4/8/12`, monthly `3/6/12` verify karo.
6. Loading, invalid range, empty, API error, retry, refresh, partial, and suppressed states test karo.
7. Masked cohort me raw count/rate visually leak na ho.
8. Narrow screen, keyboard controls, tooltip, and table scroll behavior check karo.

### Expected response essentials

```json
{
  "summary": {
    "new_users": 1200,
    "returning_users": 800,
    "returning_rate": 40,
    "average_retention": 28.5
  },
  "new_vs_returning": [
    { "bucket": "2026-W21", "new_users": 1200, "returning_users": 800 }
  ],
  "cohorts": [
    {
      "cohort_key": "2026-W21",
      "cohort_label": "May 18 - May 24",
      "cohort_size": 1200,
      "buckets": [
        { "offset": 0, "label": "W0", "users": 1200, "rate": 100 },
        { "offset": 1, "label": "W1", "users": 420, "rate": 35 }
      ]
    }
  ],
  "meta": {
    "from": "2026-06-01",
    "to": "2026-06-19",
    "interval": "week",
    "window": 8,
    "small_count_threshold": 5,
    "partial": false,
    "suppressed": false
  }
}
```

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Vite dev server | `5174` | `/cohorts` UI and `/api` proxy | Reused |
| Vite preview | `4174` | Built frontend preview | Reused |
| Intended API Gateway | `8080` | Admin retention API entry | Reused; executable missing |
| Current Session HTTP | `8086` | Privacy routes and `/healthz` | Reused; retention report route missing |
| Intended Session gRPC | `50060` | `GetRetentionReport` downstream | Reused intent; server missing |
| Auth/JWKS intent | `8081` | Admin identity/role | Reused; dashboard integration missing |
| MongoDB | `27017` | Identity/event/retention aggregates | Reused |
| Redis | `6379` | Current server dependency/future cache | Reused |

No new port introduced hua. Port conflict, firewall, credentialed CORS, cookies, and Docker networking:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Sections: `8. Docker Setup -> Docker networking mental model` and `10. Running the Project -> Ports and networking`

## 11. Common Errors & Fixes

Generic npm, Go, MongoDB, Redis, Docker, environment, port, CORS, auth, and permission issues:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `11. Common Errors & Fixes`

Only Task 6-specific issues:

| Error/symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `/cohorts` shows **Retention report unavailable** | Gateway absent/down, wrong proxy, auth failure, timeout, or endpoint missing | Network status/request ID inspect; current repository me failure expected until backend exists | Gateway/route readiness and contract tests add karo |
| Direct `:8086` request returns `404` | Current Session Service registers no report handler | Handler/use case/repository implement karo; port swap mat karo | Route integration test |
| `VALIDATION_ERROR: Invalid retention window` | Interval-window pair invalid, e.g. weekly `30` | UI-supported pair use karo: day `7/14/30`, week `4/8/12`, month `3/6/12` | Shared enum/contract tests |
| `INVALID_RETENTION_RESPONSE` | Cohort/bucket fields wrong types, invalid arrays, or required key absent | Payload ko master contract and parser tests se align karo | Provider-consumer contract validation |
| API `200` but empty state | No cohorts, zero sizes, filters too narrow, ingestion absent, or worker stale | Filters widen; data/worker freshness inspect | Seed fixture and freshness metric |
| Retention above `100%` or negative | Backend dedupe/formula bug | Identity dedupe and `retained/cohort_size` invariant fix karo | Aggregation invariant/property tests |
| Same person counted new and returning | Anonymous-to-user linking/dedup incomplete | Canonical identity graph before aggregation use karo | Identity transition fixtures |
| Cohorts shift by day/week | Backend timezone/week-start semantics differ | One timezone and interval boundary contract enforce karo | Boundary tests around midnight/DST/year change |
| Small cohort values visible | Backend relied on frontend masking only | Server response se raw small values suppress/coarsen karo | Privacy test at API boundary |
| Retention env changes do nothing | Current Go loader/worker in names ko use nahi karta | Typed config and worker implement or dead names remove karo | Env loader validation tests |
| Cohort page MSW tests show `NETWORK_ERROR` | Current test runtime relative `/api` request ko declared MSW handler se successful response nahi de raha | Test setup, fetch implementation, request URL, and MSW lifecycle inspect/align karo; real backend start karna fix nahi hai | CI me supported Node + clean locked install and page test run karo |
| `npm run typecheck` / `npm run build` fails before bundling | Existing test tuple casts, report mutation signatures, URL mocks, and `Error` lib target produce TypeScript errors | Reported compiler errors resolve karo; `vite build` alone ko green build mat declare karo | Lock toolchain and keep `tsc -b` mandatory in CI |
| Browser chart differs from test | Page test Recharts chart mock karta hai | Real `/cohorts` page manually inspect | Browser/E2E visual smoke test |

## 12. Security & Best Practices

- Report endpoint par server-side admin authorization and tenant/store isolation enforce karo. Hidden navigation security boundary nahi hai.
- Raw email, phone, IP, address, session/user IDs, free-text event payload, or credentials response/log me return mat karo.
- Small cohort suppression backend par calculation and serialization se pehle apply karo; chart/table flags alone privacy nahi dete.
- Consent withdrawal and deletion ko future aggregate rebuild/tombstone flow tak propagate karo.
- Anonymous ID ko frontend-visible stable tracking secret mat banao; rotation/linking policy privacy review ke saath define karo.
- Date range, enum filters, interval/window pair, response bytes, and maximum cohort count server par validate karo.
- Cache key me tenant/store, date range, interval, window, filters, schema version, and privacy policy version include karo.
- Production query per request raw events scan na kare. Precomputed aggregate prefer karo; fallback tightly bounded and observable ho.
- Worker idempotent, checkpointed, retry-safe, and late-event-aware ho. Backfill and schema-version rollout runbook rakho.
- Metrics: API latency/error, aggregate freshness, worker lag, processed identities, suppressed cohorts, fallback count, and rebuild failures.
- Mongo/Redis/admin secrets `.env.example`, source code, frontend `VITE_` values, images, logs, or curl history me commit mat karo.

## 13. Missing or Misconfigured Things

| Finding | Impact | Required fix |
|---|---|---|
| Retention HTTP/gRPC backend absent | Real report impossible | DTO, handler, use case, repository, auth, and integration tests implement karo |
| Aggregation worker/read model absent | `analytics_aggregates` populate nahi hota | Idempotent worker, checkpoint, late-event strategy, and freshness metrics add karo |
| Retention migration absent | Index/schema rollout unmanaged | Versioned idempotent migration plus rollback/runbook create karo |
| Gateway executable/admin-auth wiring absent | Intended `:8080` endpoint unreachable/unprotected | Gateway route, identity validation, RBAC, and tenant context implement karo |
| Identity ingestion/linking path incomplete | New/returning and cohorts unreliable | Canonical identity/first-seen contract and dedupe implementation add karo |
| Retention env names ignored | Operator ko false configuration confidence | Config struct/parser/validation/tests add karo or unused names remove karo |
| Two worker interval variable names | Precedence ambiguous | One canonical duration variable retain and migrate/document karo |
| Timezone not sent by current retention request | Day/week boundary can be ambiguous | Contract me authoritative timezone define karo or explicit query field add karo |
| Planned country filter current request me absent | Guide/API expectations drift | Product decision ke baad frontend/backend contract align karo |
| Frontend accepts suppression metadata but backend absent | Privacy guarantee end-to-end nahi | Server-side policy and contract tests implement karo |
| Cohort page MSW success/empty tests currently fail | Automated page regression coverage unreliable | Relative request interception/test environment fix and CI verification add karo |
| Repo-wide TypeScript gate currently fails | `npm run build` cannot complete despite successful raw Vite bundle | Existing compiler errors fix karo; build gate bypass mat karo |
| Production bundle warns above `500 kB` | Initial dashboard load/cache cost high ho sakta hai | Route-level lazy loading and reviewed chunk strategy add karo |
| No checked-in package lockfile | Dependency resolution drift | One package manager choose and reviewed lockfile commit karo |
| No application Dockerfiles/full Compose | Container onboarding incomplete | Non-root images, health/readiness, secrets, and network config add karo |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `task1_Dependency.md` | Tech Stack and Required Software | Same frontend/backend runtimes |
| `task1_Dependency.md` | Node/npm and Go modules | No Task 6 package/module added |
| `task1_Dependency.md` | MongoDB and Redis setup | Same ports, credentials, volume, and health commands |
| `task1_Dependency.md` | Environment Variables | Same active frontend/current backend variables |
| `task1_Dependency.md` | Docker, ports, networking, and generic errors | No Task 6 container/network change |
| `task2_Dependency.md` | Gateway and admin authentication | Same missing protected API integration |
| `task3_Dependency.md` | Session events, identity safety, ordering, and retention | Same source-data/privacy principles |
| `task4_Dependency.md` | Recharts and aggregate analytics operations | Same chart package and worker/read-model pattern |
| `task5_Dependency.md` | Latest backend/Docker gap audit | Same current Session/Gateway infrastructure state |

Full paths:

```text
TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md
TaskImplementation/Session Analytics Dashboard Service/task2_Dependency.md
TaskImplementation/Session Analytics Dashboard Service/task3_Dependency.md
TaskImplementation/Session Analytics Dashboard Service/task4_Dependency.md
TaskImplementation/Session Analytics Dashboard Service/task5_Dependency.md
```

## 15. Final Checklist

### Frontend readiness

- [ ] Previous dependency documentation read
- [ ] Node/npm versions verified
- [ ] Existing frontend dependencies installed; no Task 6-only package added
- [ ] Existing `.env.local` reused; no secret placed in `VITE_` values
- [ ] Focused API, retention math, and page tests pass
- [ ] Typecheck and production build pass
- [ ] `/cohorts` route opens on the actual Vite URL
- [ ] Interval/window pairs and segment request params verified
- [ ] Loading, empty, error, retry, partial, and suppressed states checked
- [ ] Recharts visualization manually checked in a real browser

### Real backend and DevOps readiness

- [ ] Canonical identity, anonymous linking, consent, and timezone rules approved
- [ ] Retention request/response schema versioned and contract-tested
- [ ] Session report handler/use case/repository implemented
- [ ] `analytics_aggregates` schema and reviewed migration implemented
- [ ] Idempotent aggregator worker, checkpoint, backfill, and late-event handling implemented
- [ ] Small-count suppression enforced server-side
- [ ] Retention config names loaded, validated, tested, and deduplicated
- [ ] Gateway route, admin authentication, RBAC, and tenant isolation verified
- [ ] MongoDB/Redis health and aggregate freshness monitored
- [ ] Docker images/Compose/readiness/secrets added if container deployment is required
- [ ] Real authenticated API returns safe data; current expected `404` is gone
- [ ] No duplicate common setup documentation added
