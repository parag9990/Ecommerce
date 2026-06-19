# Project Dependency & Setup Guide

> **Scope:** Task 5 - Heatmap View dependency, environment, database, and DevOps setup only. Business logic and implementation code are intentionally not repeated here.
>
> **Read first:** Common clone, Node/npm, Go, MongoDB, Redis, Docker, environment, and troubleshooting setup already exists in `task1_Dependency.md`. This file documents only Heatmap-specific differences and verification.

## 1. Project Overview

Task 5 dashboard me `/heatmaps` page add karta hai. Admin page path, device type, date range, aur `click`/`scroll` mode select karke privacy-safe aggregate points dekhta hai.

Current flow:

```text
Browser /heatmaps
  -> React + TanStack Query
  -> GET /api/v1/analytics/heatmaps
  -> intended API Gateway :8080
  -> intended Session Service GetHeatmap
  -> MongoDB heatmap_points / bounded session_events fallback
```

### Current repository reality

| Area | Current status |
|---|---|
| Heatmap React page, route, controls, query hook, Canvas renderer | Implemented |
| API client validation and response parsing | Implemented |
| Unit/component tests with Vitest, Testing Library, and MSW | Implemented |
| `GET /api/v1/analytics/heatmaps` master contract | Documented |
| Session Service heatmap handler/use case/repository | **Missing** |
| Heatmap aggregation worker/checkpoint | **Missing** |
| Executable `heatmap_points` migration | **Missing** |
| Runnable API Gateway | **Missing**; directory contains ignored environment config only |
| Admin auth/RBAC integration | **Missing** |

Isliye frontend ko independently test/build kiya ja sakta hai. Real heatmap data tabhi aayega jab ingestion, aggregation, database migration, Session API, Gateway, and admin authorization complete hon.

## 2. Tech Stack

### Reused technologies

React, TypeScript, Vite, Tailwind CSS, React Router, TanStack Query, `lucide-react`, Vitest, Testing Library, MSW, Go modules, MongoDB, Redis, and Docker ka beginner setup already documented hai.

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Sections: `2. Tech Stack`, `3. Required Software`, and `4. Dependency Management`

### Task 5-specific usage

| Technology | Simple explanation | Task 5 purpose | Required? |
|---|---|---|---|
| Canvas 2D API | Browser ka built-in drawing surface hai. Iske liye npm package install nahi hota. | Click blobs aur scroll-depth color bands render karta hai. | Yes, supported browser me |
| TanStack Query | API loading, cache, cancellation, aur retry state manage karta hai. | Filter-based heatmap request; data `60s` stale rehta hai. | Yes, already installed |
| MSW | Tests me HTTP endpoint ko mock karta hai. | Real Gateway ke bina loading/data/empty/error states test karta hai. | Dev/test only, already installed |
| MongoDB aggregates | Flexible analytics buckets documents ke form me store hote hain. | `heatmap_points` se fast page/device/date reads. | Real data ke liye mandatory |
| Redis | In-memory store/cache hai. | Future short-lived response cache/checkpoint me useful; current heatmap frontend use nahi karta. | Heatmap UI ke liye no; current Go process startup ke liye yes |

Task 5 **Recharts use nahi karta**. Heatmap drawing native Canvas se hoti hai, so `heatmap.js`, D3, ECharts, ya another rendering package install mat karo unless implementation intentionally changes.

## 3. Required Software

Task 5 koi new OS-level software introduce nahi karta.

| Setup mode | Minimum requirement | MongoDB/Redis/Go needed? |
|---|---|---|
| Frontend tests/build | Git, Node.js `22.x`, npm, modern browser | No |
| Frontend with mocked API | Same frontend tools | No |
| Current Session Service startup | Go `1.26.3`, MongoDB, Redis | Yes |
| Full real-data heatmap | Above plus future Gateway, ingestion, migration, worker, and heatmap backend | Yes; missing code must first be implemented |

Installation steps and version checks:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `3. Required Software`

Use Chrome, Edge, Firefox, or Safari ka current version. Canvas needs JavaScript enabled. GPU acceleration helpful hai but correctness ke liye mandatory nahi.

## 4. Dependency Management

### No new npm dependency

Task 5 required packages already `frontend/session-analytics-dashboard/package.json` me declared hain:

```text
react, react-dom, react-router-dom, @tanstack/react-query, lucide-react
vitest, jsdom, @testing-library/react, @testing-library/user-event, msw
```

Fresh clone ka common `npm install`, package-lock limitation, `npm run build`, and module troubleshooting repeat nahi kiya gaya.

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `4. Dependency Management -> Node.js/npm system`

Task-specific verification:

```bash
cd frontend/session-analytics-dashboard
npm ls @tanstack/react-query msw
```

Canvas verify karne ke liye `npm ls canvas` **mat** run/install karo. Browser Canvas and Node ka optional native `canvas` package different things hain; this project browser API use karta hai.

### No new Go module

Task 5 ne `backend/services/session-service/go.mod` or `go.sum` me dependency add nahi ki. Mongo driver and Redis client pehle se declared hain. Go module commands, workspace issue, proxy/cache failure, and version mismatch setup:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `4. Dependency Management -> Go modules`

### Reproducibility warning

Dashboard package me checked-in `package-lock.json`, `pnpm-lock.yaml`, or `yarn.lock` nahi hai. Reviewed install ke baad one package manager choose karke its lockfile commit karo. Multiple lockfiles mix karna reproducible builds ko hurt karega.

## 5. Database Setup

### MongoDB: reused server, new heatmap data path

MongoDB ek document database hai. Heatmap ke high-volume, flexible event and aggregate documents ke liye suitable hai. Installation, Docker container, volume, credentials, `session_db`, port `27017`, connection string, and health commands same hain.

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `5. Database Setup -> MongoDB`

Real Task 5 flow ke liye MongoDB mandatory hai and these logical collections are needed:

| Collection | Purpose | Current status |
|---|---|---|
| `session_events` | Privacy-filtered raw `click` and `scroll` input events | Design docs only; ingestion/migration absent |
| `heatmap_points` | Page/device/day/mode level pre-aggregated coordinates and weights | Design docs only; migration/writer/reader absent |

Documented index:

```javascript
db.heatmap_points.createIndex({ path: 1, device_type: 1, day: 1 })
```

Production implementation me query `mode` use karti hai, so reviewed migration ko actual query plan ke according `mode` field include karna pad sakta hai:

```javascript
// Design recommendation only; this is not a checked-in migration.
db.heatmap_points.createIndex(
  { path: 1, device_type: 1, mode: 1, day: 1 },
  { name: "heatmap_path_device_mode_day" }
)
```

> **Warning:** Snippet ko production me blindly run mat karo. Existing index inventory, duplicate documents, cardinality, query explain plan, write overhead, retention policy, backup, and rollback review karke versioned idempotent migration banao.

### Expected aggregate contract

At minimum each API point needs numeric percentage coordinates and non-negative weight:

```json
{
  "points": [
    { "x": 42.5, "y": 20, "weight": 18 }
  ],
  "total_events": 18,
  "max_weight": 18,
  "average_scroll_depth": 64.5,
  "generated_at": "2026-06-19T12:00:00Z",
  "min_bucket_size": 5,
  "partial": false,
  "suppressed": false
}
```

Rules:

- `x` and `y` percentage range ideally `0..100` ho; frontend display ke liye clamp karta hai, backend validation still required hai.
- `weight` finite and `>= 0` ho. Negative weight parser reject karta hai.
- Click data page-relative or viewport-relative semantics consistently define karo; mixed coordinate systems misleading visualization banayenge.
- Scroll points me `y` reached depth represent kare; aggregation `25/50/75/90/100` buckets support kare.
- `path`, device, mode, day/range, tenant/store, schema version, and consent policy aggregate identity ka part hone chahiye.

### Migration status

Current `backend/services/session-service/migrations/001_privacy_controls.mongodb.js` only deletion/audit indexes create karti hai. It does **not** create `session_events`, `heatmap_points`, heatmap TTL, or aggregation checkpoint. No Task 5 migration command currently exists.

Do not report setup complete merely because privacy migration ran successfully.

### Redis

Redis installation, Docker setup, credentials, DB `2`, port `6379`, and `PING` verification unchanged hain.

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `5. Database Setup -> Redis`

Current Go server Redis ko startup par ping karta hai, so that process ke liye Redis reachable hona mandatory hai. Lekin no heatmap cache, checkpoint, lock, or invalidation code exists; Redis start karne se missing heatmap API ready nahi hogi.

## 6. Redis / Queue / External Services

### API Gateway and admin authentication

Master contract endpoint:

```text
GET /api/v1/analytics/heatmaps
auth: admin
service: session-service
gRPC: SessionService.GetHeatmap
```

Frontend request `credentials: "include"` bhejti hai. Intended Gateway ko admin credential authenticate, role and tenant/store authorize, query validate, request ID propagate, and safe error envelope return karna hoga.

Current `backend/services/api-gateway/` me executable source/module nahi hai. Its ignored `.env` declares `:8080` and downstream `SESSION_GRPC_ADDR=localhost:50060`, but config alone Gateway start nahi karti. Current Session Service only HTTP `:8086` privacy routes and `/healthz` expose karta hai; heatmap route/gRPC server absent hai.

Shared Gateway/auth explanation:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `6. Redis / Queue / External Services -> API Gateway and admin authentication`

### Kafka/RabbitMQ

Architecture docs high-volume downstream processing ke liye Kafka/RabbitMQ suggest karte hain. Current Task 5 code/manifests me producer, consumer, topic, queue, DLQ, credentials, or health check nahi hai. Therefore broker **optional and not installable from the current repository**.

Future broker choose karte waqt event partition key, at-least-once duplicates, idempotent aggregate upsert, retry/backoff, DLQ replay, checkpoint, and worker lag define karo. UI-only setup ke liye broker install mat karo.

### Other services

No screenshot provider, replay SDK, S3, MinIO, Elasticsearch, SMTP, Stripe, Firebase, NATS, Nginx, or Kubernetes dependency Task 5 adds. Current preview is a local schematic DOM frame, external website screenshot nahi.

## 7. Environment Variables

### Frontend: no new variable

Task 5 existing variables unchanged reuse karta hai:

| Variable | Heatmap effect | Required? |
|---|---|---|
| `VITE_API_BASE_URL` | Heatmap fetch ka base URL; empty means same-origin | No; local proxy ke liye empty recommended |
| `VITE_API_PROXY_TARGET` | `/api` proxy destination; default `http://localhost:8080` | No |
| `VITE_REQUEST_TIMEOUT_MS` | Slow/unreachable heatmap request abort time; default `10000` ms | No |

Create/load rules, `.env.local` example, restart requirement, common mistakes, and security:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `7. Environment Variables -> Frontend .env.local`

`VITE_` variables browser bundle me visible hoti hain. Mongo URI, Redis password, admin password/token, JWT key, pepper, or broker credential kabhi frontend env me mat rakho.

### Heatmap backend names: documented but inactive

Ignored local file `backend/services/session-service/.env` currently these Task 5-related names lists karti hai:

```env
SESSION_HEATMAP_RETENTION_DAYS=730
SESSION_HEATMAP_CLICK_BUCKET_SIZE=5
SESSION_HEATMAP_MAX_DATE_RANGE_DAYS=31
SESSION_HEATMAP_MAX_POINTS=5000
SESSION_HEATMAP_AGGREGATION_ENABLED=true
SESSION_HEATMAP_AGGREGATION_INTERVAL=1m
SESSION_HEATMAP_AGGREGATION_BATCH_SIZE=1000
SESSION_HEATMAP_AGGREGATION_INITIAL_LOOKBACK=1h
SESSION_HEATMAP_AGGREGATION_CHECKPOINT_LOOKBACK=5m
SESSION_HEATMAP_AGGREGATION_WORKER_NAME=heatmap_aggregator
```

| Group | Intended purpose | Required now? | Security/validation |
|---|---|---|---|
| Retention | Aggregate delete/expiry window | No; loader ignores it | Legal/privacy policy and stored settings se align karo |
| Bucket size | Click percentage grid size | No; loader ignores it | Positive bounded number validate karo |
| Query limits | Maximum date range and response points | No; loader ignores them | Server-side enforcement required; UI value trust mat karo |
| Worker switch/timing | Aggregator enable, interval, batch, lookback, checkpoint overlap | No; worker absent and loader ignores all | Duration/batch validation and bounded overlap required |
| Worker name | Checkpoint/metric identity | No; worker absent | Stable allowlisted identifier use karo |

**Important:** `internal/config/config.go` in names ko parse nahi karta. `.env` value change karne se Task 5 behavior change nahi hoga. Backend implementation me typed config fields, explicit validation, defaults, config tests, and non-secret startup summary add karna hoga.

There is also retention-source drift: ignored env says `730` days while current privacy domain default is `365` days and accepts `30..730`. One authoritative runtime source choose karo; operator ko conflicting values mat dikhao.

Go `.env` automatically load nahi karta. Existing loaded Session variables and safe export process:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `7. Environment Variables -> Current Session Service environment`

Query fields `path`, `device_type`, `mode`, `from`, and `to` environment variables nahi hain; browser per request bhejta hai.

## 8. Docker Setup

Task 5 koi new Dockerfile, container, image, volume, network, health check, restart policy, or port mapping introduce nahi karta.

Reuse MongoDB + Redis local Compose documentation:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `8. Docker Setup`

That example only infrastructure start karta hai. Full heatmap stack nahi, because repository me dashboard/Session/Gateway application Dockerfiles, Gateway executable, worker, and heatmap backend absent hain.

Future worker/container checklist:

- non-root image and pinned versions;
- separate liveness and dependency-aware readiness;
- Mongo/Redis/broker service hostnames, not container-local `localhost`;
- graceful shutdown and checkpoint flush;
- resource limits and batch backpressure;
- secret injection, no image-baked credentials;
- health status for worker lag/freshness;
- named volume/backup only where durable state really belongs.

## 9. Local Development Setup

### Step 1: Complete common onboarding

First follow:

1. `task1_Dependency.md` for clone, required tools, install, frontend env, and common Docker infrastructure.
2. `task2_Dependency.md` for shared API Gateway/admin-auth and active analytics integration context.
3. `task3_Dependency.md` for `session_events`, ordering, retention, and privacy-safe event handling.
4. `task4_Dependency.md` for aggregate analytics backend gaps and production worker principles.

### Step 2: Move to dashboard package

```bash
cd frontend/session-analytics-dashboard
```

### Step 3: Install only when needed

No Task 5-only package install hai. Fresh clone or missing `node_modules` case me Task 1 ka normal install use karo.

### Step 4: Reuse frontend environment

`frontend/session-analytics-dashboard/.env.local` me existing three `VITE_` values reuse karo. No new secret/value add karna hai.

### Step 5: Run focused verification

```bash
npm run test -- \
  src/api/session-api.test.ts \
  src/features/heatmaps/lib/heatmap-normalize.test.ts \
  src/features/heatmaps/pages/heatmap-page.test.tsx
npm run typecheck
npm run build
```

Tests real MongoDB, Redis, Gateway, or Session Service require nahi karte. Page test MSW use karta hai and Canvas component mock karta hai, so manual browser Canvas check still necessary hai.

### Step 6: Start frontend

```bash
npm run dev
```

Open:

```text
http://localhost:5174/heatmaps
```

`strictPort: false` hai. `5174` busy ho to terminal ka printed fallback URL open karo.

### Step 7: Migration and services

Frontend-only work ke liye no migration/service needed. Full integration ke liye reused MongoDB/Redis start karo, but Task 5-specific migration/worker currently absent hain. Design snippet ko migration samajhkar execute mat karo.

### Step 8: Verify intended API

```bash
curl -i \
  'http://localhost:8080/api/v1/analytics/heatmaps?path=%2Fproducts%2Fprod_123&device_type=desktop&mode=click&from=2026-06-01&to=2026-06-19'
```

Current checkout me Gateway connection failure expected hai. Direct current Session Service checks:

```bash
curl -i http://localhost:8086/healthz
curl -i \
  'http://localhost:8086/api/v1/analytics/heatmaps?path=%2Fproducts%2Fprod_123&device_type=desktop&mode=click&from=2026-06-01&to=2026-06-19'
# Current expectation: /healthz can return 200; heatmap route returns 404.
```

Proxy target `8086` karna missing route ko implement nahi karta. Final Gateway request valid admin authentication require karega.

## 10. Running the Project

### Daily frontend workflow

```bash
cd frontend/session-analytics-dashboard
npm run test -- src/features/heatmaps
npm run dev
```

Manual browser verification:

1. `/heatmaps` route shell/navigation ke andar open karo.
2. Page, desktop/tablet/mobile, date range, and click/scroll filters change karo.
3. Network tab me `path`, `device_type`, `mode`, `from`, and `to` query params verify karo.
4. Mock/integrated success response par summary, preview, legend, and privacy notice verify karo.
5. Click mode me high-DPI Canvas blobs and top buckets inspect karo.
6. Scroll mode me depth bands and `25/50/75/90/100` reach inspect karo.
7. Loading, invalid range, empty, API error, retry, refresh, partial, and suppressed states test karo.
8. Narrow screen par horizontal preview behavior and keyboard controls check karo.

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Vite dev server | `5174` | `/heatmaps` UI and `/api` proxy | Reused |
| Vite preview | `4174` | Built frontend preview | Reused |
| Intended API Gateway | `8080` | Admin heatmap API entry | Reused; executable missing |
| Current Session HTTP | `8086` | Privacy routes and `/healthz` | Reused; heatmap route missing |
| Intended Session gRPC | `50060` | `GetHeatmap` downstream | Reused intent; server missing |
| Auth/JWKS intent | `8081` | Admin identity/role | Reused; integration missing |
| MongoDB | `27017` | Raw events and heatmap aggregates | Reused |
| Redis | `6379` | Current service dependency/future cache | Reused |

No new port introduced hua. Port conflicts, firewall, CORS/cookies, and Docker networking:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Sections: `8. Docker Setup -> Docker networking mental model` and `10. Running the Project -> Ports and networking`

## 11. Common Errors & Fixes

Generic npm, Go, MongoDB, Redis, Docker, environment, port, CORS, auth, and permission errors:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `11. Common Errors & Fixes`

Only Task 5-specific issues:

| Error / symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `/heatmaps` shows **Heatmap unavailable** | Gateway absent/down, proxy wrong, auth failure, timeout, or endpoint missing | Network status/request ID inspect; current repository me failure expected until backend exists | Gateway/route integration tests and readiness add karo |
| Direct `:8086` call returns `404` | Current Session Service registers no heatmap handler | Backend handler/use case/repository implement karo; port swap mat karo | Route contract test add karo |
| `INVALID_HEATMAP_RESPONSE` | `points` missing/not array, point values non-numeric, negative weight, or bad optional metadata | Response master schema and frontend parser align karo | Provider-consumer contract tests |
| API `200` but no visualization | `points: []`, all weights zero, or wrong page/device/mode/range | Payload and selected filters inspect; ingestion/worker freshness check | Seed fixture and aggregate freshness metric |
| Points appear in wrong place | Backend sent pixels instead of percentages, mixed viewport/page coordinates, or wrong device bucket | One coordinate contract choose and normalize during ingestion | Schema version and coordinate invariant tests |
| Canvas looks blurry | Pixel ratio/size handling changed or CSS stretches canvas | Existing device-pixel-ratio sizing preserve; actual browser inspect | Visual/E2E regression test |
| Scroll reach is misleading | Backend point `y` does not mean maximum reached depth or duplicate events inflate weight | Canonical max-depth-per-session aggregation implement karo | Aggregation fixtures for depth semantics |
| Too many points freeze browser | API ignored response cap or raw points returned | Server aggregate/downsample and cap; bounded date range enforce | Load test, `max_points`, payload-byte limits |
| Task env changes have no effect | `config.go` does not load `SESSION_HEATMAP_*` | Typed loader/validation/worker implement or remove dead names | Config tests for every documented variable |
| Privacy notice shown but small groups inferable | Backend only set flag without suppressing/coarsening data | Server response se unsafe buckets remove/round karo | Minimum-bucket privacy tests |
| MSW test passes but Canvas is broken | Page test intentionally mocks Canvas | `/heatmaps` real browser manual/E2E check karo | Browser Canvas smoke test add karo |
| `npm` reports WSL 1 unsupported | Windows npm executable is mixed with WSL 1/Linux Node | WSL 2 use karo or Node/npm same OS environment me install karo | One supported toolchain path document karo |

## 12. Security & Best Practices

- Heatmap endpoint par server-side admin RBAC and tenant/store scope enforce karo. Navigation visibility authorization nahi hai.
- Password, OTP, card/CVV, authorization tokens, keystrokes, private text, or sensitive form values collect mat karo.
- Consent-disabled sessions ingestion and all downstream aggregates se exclude karo.
- Full URLs/query strings store mat karo; normalized safe path use karo so tokens/email/order identifiers leak na hon.
- Small buckets backend par suppress/coarsen karo. Frontend `partial`, `suppressed`, or `min_bucket_size` notice defense nahi hai.
- Coordinates and metadata logs me raw user/session identifiers join mat karo. Request ID, latency, bucket count, worker lag, and freshness sufficient hain.
- Raw fallback short and bounded rakho. Large request-time Mongo scans admin API ko DoS vector bana sakte hain.
- Aggregator idempotent and checkpointed ho; retry se same event double-count nahi hona chahiye.
- Cache key me tenant, path, device, mode, range, privacy policy version, and schema version include karo.
- Path/date/device/mode, maximum range, response point count, payload bytes, and numeric finiteness server-side validate karo.
- Heatmap retention current Privacy settings/deletion workflow se integrate karo; ignored env and stored setting ko silently conflict mat karne do.
- Aggregate delete/anonymize behavior document karo. A user deletion ke baad irreversible anonymous aggregates may stay only if policy/legal basis allows it.
- Content Security Policy and frame restrictions use karo. Future live screenshots/URLs introduce karne par SSRF, cross-origin image, and Canvas tainting risk review karo.
- Metrics: request latency/error, cache hit, aggregate age, worker lag, scanned events, generated buckets, suppressed buckets, and raw-fallback count.

Shared secrets, cookies/CSRF, TLS, databases, container hardening, and dependency best practices:

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `12. Security & Best Practices`

## 13. Missing or Misconfigured Things

| Severity | Finding | Impact | Professional fix |
|---|---|---|---|
| Blocker | Session Service heatmap route/gRPC, use case, and repository absent | Real frontend request cannot succeed | Implement validated admin API and contract/integration tests |
| Blocker | Runnable API Gateway/admin auth integration absent | Intended `:8080` entry unavailable/unprotected | Add executable Gateway, auth middleware, RBAC, tenant scope, and readiness |
| Blocker | Event ingestion and heatmap aggregation worker absent | No source data or aggregates | Add privacy-safe ingestion, idempotent worker, checkpoint, retry, and observability |
| High | No executable `session_events`/`heatmap_points` migration | Schema/index/retention rollout unmanaged | Add reviewed idempotent migration, rollback/runbook, and explain-plan verification |
| High | Master `HeatmapRequest` omits frontend `mode` query parameter | Contract validation/generation may reject or ignore click/scroll selection | Add enum `click|scroll` or define separate canonical response semantics |
| High | Master response only defines `points`; frontend also consumes total/max/average/generated/privacy metadata | Provider-consumer drift can hide freshness/privacy state | Expand canonical schema and add contract tests |
| High | `SESSION_HEATMAP_*` names are ignored | Operators get false confidence and unsafe defaults may not apply | Parse/validate/use them or remove until implemented |
| High | Retention env says `730`, privacy domain default says `365` | Ambiguous deletion behavior | Make stored privacy policy authoritative or explicitly reconcile precedence |
| High | No server-side small-bucket enforcement | Aggregate re-identification risk | Suppress/coarsen before response and test differencing attacks |
| Medium | Frontend clamps out-of-range coordinates | Bad backend data can be silently hidden | Server reject/measure invalid points; expose data quality metric |
| Medium | Heatmap page test mocks Canvas | Rendering regression can escape unit suite | Add Playwright/browser screenshot or Canvas smoke test |
| Medium | Page paths are hard-coded frontend options | New routes cannot be selected without release | Provide privacy-safe top-pages/options endpoint or config source |
| Medium | Preview is schematic, not actual page layout | Hotspot visual position may not correspond to production DOM | Clearly label conceptual view or add versioned safe screenshot/template pipeline |
| Medium | No package lockfile | Clean install versions can drift | Choose npm/pnpm/yarn and commit one reviewed lockfile |
| Medium | `/healthz` is shallow | Healthy process can have broken heatmap dependencies | Add `/livez`, `/readyz`, worker freshness, and migration readiness |
| Medium | No application Dockerfiles/complete Compose | Full stack cannot be reproduced by one command | Add non-root images, health checks, secrets, resources, and migration job |

Unchanged shared platform gaps are not repeated. Refer to `task1_Dependency.md`, Section `13. Missing or Misconfigured Things`.

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `task1_Dependency.md` | Tech stack and required software | Same React/Vite/TypeScript/Go toolchain |
| `task1_Dependency.md` | Node/npm and Go modules | Task 5 adds no package/module |
| `task1_Dependency.md` | MongoDB and Redis installation/credentials/health | Same instances, ports, database, and connection setup |
| `task1_Dependency.md` | Frontend and loaded backend environment | Existing variables and file locations unchanged |
| `task1_Dependency.md` | Docker setup and networking | No new Task 5 container/volume/network |
| `task1_Dependency.md` | Common run commands, ports, errors, and security | Shared local/deployment behavior |
| `task2_Dependency.md` | Gateway/admin auth and analytics API integration | Same missing platform path |
| `task3_Dependency.md` | `session_events`, retention, event privacy | Click/scroll input reuses event principles |
| `task4_Dependency.md` | Aggregate analytics worker, API gaps, and performance | Same pre-aggregation operational model |

Full paths:

```text
TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md
TaskImplementation/Session Analytics Dashboard Service/task2_Dependency.md
TaskImplementation/Session Analytics Dashboard Service/task3_Dependency.md
TaskImplementation/Session Analytics Dashboard Service/task4_Dependency.md
```

## 15. Final Checklist

### Frontend setup

- [ ] Previous dependency guides read first
- [ ] Supported Node/npm toolchain verified in one environment
- [ ] Dashboard dependencies installed using the existing manifest
- [ ] No unnecessary Canvas/heatmap/chart package installed
- [ ] Existing `.env.local` reused; no secret stored in `VITE_` values
- [ ] Heatmap API client, helper, and page tests pass
- [ ] Typecheck and production build pass
- [ ] `/heatmaps` opens at the terminal-reported Vite URL
- [ ] Page/device/range/mode requests verified in Network tab
- [ ] Canvas click and scroll rendering manually checked in a real browser
- [ ] Loading, validation, empty, error, retry, partial, and suppressed states checked
- [ ] Responsive and keyboard behavior checked

### Backend and DevOps readiness

- [ ] Broker/screenshot/replay dependencies not installed without an implemented need
- [ ] Privacy migration not mistaken for a heatmap migration
- [ ] Canonical click/scroll coordinate and depth semantics defined
- [ ] Privacy-safe event ingestion implemented and consent enforced
- [ ] `heatmap_points` schema, compound index, retention, and rollback reviewed
- [ ] Idempotent aggregation worker/checkpoint/retry implemented
- [ ] `SESSION_HEATMAP_*` variables parsed, validated, tested, and documented accurately
- [ ] One authoritative heatmap retention value enforced
- [ ] Session Service `GetHeatmap` and Gateway route implemented
- [ ] `mode` and response metadata master contract drift resolved
- [ ] Admin RBAC and tenant/store isolation enforced server-side
- [ ] Server-side small-bucket suppression and deletion behavior verified
- [ ] Query range/point/payload limits and raw-fallback bounds load-tested
- [ ] Readiness, aggregate freshness, worker lag, metrics, and alerts verified
- [ ] Real endpoint `200/400/401/403/404/timeout/upstream-failure` tests pass
- [ ] No duplicate common setup documentation added

---

**Beginner mental model:** Frontend Heatmap page abhi independently runnable and testable hai. MongoDB/Redis start karna infrastructure ready karta hai, feature nahi. Real heatmap ke liye privacy-safe events, executable indexes, aggregation worker, Session API, Gateway, and admin authorization sab connected hone chahiye.
