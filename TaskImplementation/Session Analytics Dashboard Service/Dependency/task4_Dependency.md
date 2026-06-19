# Project Dependency & Setup Guide

> **Scope:** Ye guide Task 4 ke product-view -> cart -> checkout -> paid funnel ko run aur verify karne ke liye sirf new ya task-specific setup explain karti hai. Shared installation, MongoDB, Redis, Docker, Node/npm, Go, aur base environment setup repeat nahi kiya gaya.
>
> **Current repository status:** Funnel frontend, `/funnels` route, API client, Recharts visualization, aur focused tests implemented hain. Lekin `GET /api/v1/analytics/funnels`, `SessionService.GetFunnelReport`, event ingestion, funnel aggregation worker, admin-auth Gateway, aur funnel MongoDB migrations implemented nahi hain. Isliye mocked frontend tests/build run ho sakte hain; real end-to-end funnel data abhi nahi.

## 1. Project Overview

Task 4 aggregate conversion funnel dikhata hai:

```text
Product viewed -> Added to cart -> Checkout started -> Paid
```

Frontend date range aur segment filters ke saath ye request banata hai:

```text
Browser (:5174)
  -> GET /api/v1/analytics/funnels
  -> Vite proxy
  -> intended API Gateway (:8080)
  -> intended admin authorization
  -> intended Session Service funnel aggregation
  -> intended analytics_aggregates / session_events in MongoDB
```

### What can run today?

| Capability | Status | Beginner note |
|---|---|---|
| `/funnels` UI route | Available | Frontend dependencies installed honi chahiye |
| Funnel math unit tests | Available | Backend/database ki need nahi |
| Funnel page tests with MSW | Available | MSW fake API response deta hai |
| Production frontend build | Available after dependency install | Build sab dashboard routes compile karta hai |
| Real funnel API | Missing | Contract docs me hai, runnable handler nahi |
| Real MongoDB aggregation | Missing | Collections/indexes documented hain, implementation/migration nahi |
| Full admin-auth flow | Missing | Gateway executable aur auth wiring absent hain |

## 2. Tech Stack

### Reused technologies

React, TypeScript, Vite, Tailwind CSS, React Router, TanStack Query, Lucide, date-fns, Vitest, Testing Library, MSW, Node/npm, Go, MongoDB, Redis, Gateway intent, aur Docker ka beginner explanation pehle se available hai.

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Sections:
`2. Tech Stack`, `3. Required Software`, `4. Dependency Management`, `5. Database Setup`, and `8. Docker Setup`

### Task 4-specific technology usage

| Technology | Simple Hinglish explanation | Task 4 me use | Required? |
|---|---|---|---|
| Recharts | Recharts React ke liye chart library hai. Data ko responsive funnel shape aur tooltip me render karta hai. | `Funnel`, `FunnelChart`, `ResponsiveContainer`, `Tooltip`, `Cell`, aur `LabelList` | **Yes** for current package build and funnel chart |
| TanStack Query | Server data fetch/cache manage karta hai. | Filter-based query key, request cancellation, 60-second stale time, manual refresh | Yes, reused |
| MSW | Test ke andar fake HTTP server provide karta hai. | Real backend ke bina success, empty, aur error states verify hote hain | Development only, reused |

> Recharts package Task 4 ka functional addition hai, but current `package.json` me already declared hai. Isko dobara `npm install recharts` se mutate karne ki need nahi; normal package install enough hai.

## 3. Required Software

Task 4 koi new OS-level software introduce nahi karta. Git, Node.js 22 LTS, npm, browser, optional Go `1.26.3`, Docker, `mongosh`, aur `redis-cli` requirements ke liye refer:

`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `3. Required Software`

Frontend-only funnel work ke liye minimum Git + Node/npm + browser hai. MongoDB, Redis, Go, aur Docker sirf backend/full integration work me required hain.

## 4. Dependency Management

### Recharts verification

Current declaration:

```json
"recharts": "^2.15.0"
```

Fresh clone me common install process follow karo:

`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `4. Dependency Management -> Node.js/npm system`

Install ke baad verify:

```bash
cd frontend/session-analytics-dashboard
npm ls recharts
```

Expected: dependency tree me `recharts@2.x` show hoga. `node_modules/` present dikhne bhar se setup complete assume mat karo; package-specific symlinks/modules resolve hone chahiye.

### No new backend module

Task 4 ke liye `backend/services/session-service/go.mod` ya `go.sum` me new Go dependency nahi hai, kyunki funnel backend implement hi nahi hua. Go modules, `go mod download`, `go build`, `go run`, `go mod tidy`, version mismatch, proxy, aur cache issues already documented hain.

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `4. Dependency Management -> Go modules`

### Lockfile warning

Dashboard package ke paas checked-in npm/pnpm/yarn lockfile abhi nahi hai. Isse clean installs transitive dependency versions me drift kar sakte hain. Reviewed install ke baad ek package-manager lockfile commit karna chahiye; multiple lockfiles mix mat karo.

## 5. Database Setup

### MongoDB: reused service, new funnel data requirement

MongoDB installation, Docker container, port `27017`, `session_db`, credentials, URI, persistence, aur health checks repeat nahi kiye gaye.

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `5. Database Setup -> MongoDB`

Real funnel report ke liye MongoDB mandatory hoga. Task 4 ko ye logical data chahiye:

| Collection | Funnel purpose | Current repository status |
|---|---|---|
| `session_events` | Ordered raw events se unique-session fallback funnel calculate karna | Schema/indexes docs me; ingestion/query implementation absent |
| `analytics_aggregates` | Production me fast precomputed funnel buckets read karna | Index docs me; writer/reader/migration absent |

Expected documented indexes:

```javascript
db.session_events.createIndex({ session_id: 1, occurred_at: 1 })
db.session_events.createIndex({ event_type: 1, occurred_at: -1 })
db.session_events.createIndex({ occurred_at: 1 }, { expireAfterSeconds: 7776000 })
db.analytics_aggregates.createIndex(
  { metric: 1, bucket: 1, segment: 1 },
  { unique: true }
)
```

> **Do not run these blindly in production.** Ye architecture-document examples hain, checked-in Task 4 migration nahi. Index names, existing indexes, query plan, TTL impact, backup, aur rollout review karke idempotent migration create karni hogi.

Current `backend/services/session-service/migrations/001_privacy_controls.mongodb.js` funnel collections/indexes create nahi karti. Privacy migration run karne se funnel API ready nahi hota.

### Required event mapping

Aggregation ko at least ye canonical mapping consistently ingest karni hogi:

| Funnel step | Expected source event |
|---|---|
| `product_view` | product page view event |
| `add_to_cart` | cart addition event |
| `checkout_started` | checkout start event |
| `paid` | successful payment result only |

Counts ordered unique sessions hone chahiye. Raw event count use karne se repeated clicks funnel inflate kar denge. Payment failures ko `paid` me include nahi karna.

### Redis

Redis installation, Docker setup, port `6379`, DB `2`, credentials, aur health check unchanged hain.

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `5. Database Setup -> Redis`

Architecture me Redis recent counters/cache ke liye useful hai, but current code funnel cache read/write nahi karta. Redis run karne se missing funnel handler automatically appear nahi hoga.

## 6. Redis / Queue / External Services

### API Gateway and admin auth

API contract `GET /api/v1/analytics/funnels` ko `admin` protected aur `SessionService.GetFunnelReport` mapped batata hai. Frontend `credentials: "include"` use karta hai, so intended login cookie request ke saath jayegi.

Current blockers:

- API Gateway folder me runnable source/module absent hai;
- admin login/JWKS/RBAC dashboard ke saath wired nahi hai;
- Session Service me funnel HTTP route ya gRPC method absent hai;
- current Session Service HTTP server `:8086` par sirf privacy routes aur `/healthz` expose karta hai.

Gateway/admin-auth common context ke liye refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `6. Redis / Queue / External Services -> API Gateway and admin authentication`

### Kafka/RabbitMQ

Docs high-volume analytics ingestion ke liye broker recommend karte hain, but Task 4 manifests/config/code me Kafka ya RabbitMQ producer, consumer, topic, queue, DLQ, ya health check nahi hai. Current setup ke liye broker **optional and unnecessary** hai. Implementation aane ke baad hi exact broker setup document karo.

### Third-party services

SMTP, Stripe, S3, Firebase, OAuth provider, MinIO, Elasticsearch, NATS, Nginx, aur Kubernetes Task 4 frontend ki direct dependency nahi hain. Payment domain ka successful event source eventually required hai, but dashboard ko payment credentials kabhi nahi milne chahiye.

## 7. Environment Variables

### Frontend: no new Task 4 variable

Task 4 koi new frontend environment variable introduce nahi karta. Existing `.env.local` location and values reuse karo:

`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `7. Environment Variables -> Frontend .env.local`

Relevant existing variables:

| Variable | Funnel effect | Required? |
|---|---|---|
| `VITE_API_BASE_URL` | Funnel request ka browser base URL | No; local proxy ke liye empty recommended |
| `VITE_API_PROXY_TARGET` | `/api` ko intended Gateway `http://localhost:8080` par proxy karta hai | No; Vite default same hai |
| `VITE_REQUEST_TIMEOUT_MS` | Funnel request abort timeout | No; default `10000` |

`VITE_` values browser bundle me public hoti hain. DB password, Redis password, payment secret, JWT signing key, ya admin credential yahan mat rakho.

### Backend `.env` audit

Existing ignored Session Service `.env` me funnel-looking names present hain:

```env
SESSION_ANALYTICS_MAX_FUNNEL_RANGE_DAYS=90
SESSION_ANALYTICS_RAW_FALLBACK_RANGE=24h
SESSION_ANALYTICS_MIN_FUNNEL_STEPS=2
SESSION_ANALYTICS_MAX_FUNNEL_STEPS=8
SESSION_ANALYTICS_DEFAULT_FUNNEL_METRIC=funnel.checkout
SESSION_ANALYTICS_FUNNELS_CACHE_TTL=2m
```

**Important:** current `internal/config/config.go` in variables ko load nahi karta. Ye values change karne se runtime behavior change nahi hoga. Backend implementation ke time typed config fields, validation, tests, safe defaults, aur startup logging add karna required hai.

Query parameters such as `from`, `to`, `device_type`, `channel`, `source`, `user_type`, aur `steps` environment variables nahi hain; frontend har request me bhejta hai.

## 8. Docker Setup

Task 4 koi new container, image, volume, network, health check, port mapping, ya restart policy introduce nahi karta. MongoDB + Redis local Compose example reuse karo:

`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `8. Docker Setup`

Repository me dashboard/Session Service Dockerfile aur checked-in analytics Compose stack abhi absent hai. Application image exist kiye bina `docker compose up` se full funnel flow expect mat karo.

## 9. Local Development Setup

### Step 1: Complete common onboarding

Clone, software verification, dependency installation, `.env.local`, and optional MongoDB/Redis setup:

`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Sections: `3. Required Software`, `4. Dependency Management`, `7. Environment Variables`, and `9. Local Development Setup`

### Step 2: Move to the dashboard package

```bash
cd frontend/session-analytics-dashboard
```

### Step 3: Verify Task 4 dependency

```bash
npm ls recharts
```

Missing/incomplete ho to package directory me normal install run karo:

```bash
npm install
```

### Step 4: Run focused verification

```bash
npm run test -- src/features/funnels/lib/funnel-math.test.ts
npm run test -- src/features/funnels/pages/funnel-analysis-page.test.tsx
npm run typecheck
npm run build
```

Focused tests backend, MongoDB, Redis, ya Gateway require nahi karte. Page test MSW use karta hai aur chart component mock karta hai; isliye manual browser chart check bhi required hai.

### Step 5: Start frontend

```bash
npm run dev
```

Open:

```text
http://localhost:5174/funnels
```

Vite config `strictPort: false` use karta hai. Agar `5174` busy hai, terminal jo fallback URL print kare wahi open karo.

### Step 6: Understand the real-data blocker

Intended API probe:

```bash
curl -i 'http://localhost:8080/api/v1/analytics/funnels?from=2026-06-01&to=2026-06-19&device_type=all&channel=all&source=all&user_type=all&steps=product_view%2Cadd_to_cart%2Ccheckout_started%2Cpaid'
```

Current repository state me Gateway unavailable hoga. Direct current Session Service check:

```bash
curl -i http://localhost:8086/healthz
curl -i 'http://localhost:8086/api/v1/analytics/funnels?from=2026-06-01&to=2026-06-19'
```

Expected current behavior: health endpoint can return `200`; funnel path returns `404`. Proxy target ko `8086` karna missing route ko fix nahi karega.

## 10. Running the Project

### Daily frontend workflow

```bash
cd frontend/session-analytics-dashboard
npm run test -- src/features/funnels
npm run dev
```

Browser DevTools -> Network me verify karo:

- request path `/api/v1/analytics/funnels` hai;
- `from` and `to` valid ISO date values hain;
- `steps=product_view,add_to_cart,checkout_started,paid` encoded form me gaya hai;
- device/channel/source/user filters request me update hote hain;
- request credentials policy `include` hai;
- success response me `steps` array hai.

### Expected response shape

```json
{
  "min_segment_size": 5,
  "generated_at": "2026-06-19T12:00:00Z",
  "partial": false,
  "suppressed": false,
  "steps": [
    { "key": "product_view", "count": 1000 },
    { "key": "add_to_cart", "count": 320 },
    { "key": "checkout_started", "count": 180 },
    { "key": "paid", "count": 90 }
  ]
}
```

Frontend aliases like `sessions`, `uniqueSessions`, `users`, aur `uniqueUsers` parse kar sakta hai, but backend contract ko one canonical response shape choose karke document/test karna chahiye.

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Vite dev server | `5174` | `/funnels` local UI | Reused |
| Vite preview | `4174` | Built bundle preview | Reused |
| Intended API Gateway | `8080` | Admin API entry point | Reused, implementation missing |
| Current Session HTTP | `8086` | Privacy APIs and `/healthz` | Reused; no funnel route |
| Intended Session gRPC | `50060` | `GetFunnelReport` downstream | Reused intent, implementation missing |
| MongoDB | `27017` | Raw events and aggregates | Reused |
| Redis | `6379` | Intended recent cache/counters | Reused |

Port conflicts, firewall, CORS, cookie, and Docker networking details:

`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Sections: `8. Docker Setup -> Docker networking mental model` and `10. Running the Project -> Ports and networking`

## 11. Common Errors & Fixes

Shared npm, Docker, MongoDB, Redis, CORS, auth, port, Go, and environment errors are already documented:

`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section: `11. Common Errors & Fixes`

Only Task 4-specific problems:

| Error/symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `Cannot find package 'recharts'` | Dependencies not installed or incomplete `node_modules` | Dashboard directory me `npm install`, then `npm ls recharts` | Lockfile commit karo and CI me clean install use karo |
| Vitest cannot resolve Vite/plugin packages | Test runner kisi unrelated parent install se launch hua, dashboard dependencies linked nahi | Package-local install complete karo and `npm run test` use karo | Root/package manager layout document and lock karo |
| `/funnels` opens but error state shows | Gateway/funnel API missing or down | Network tab, proxy target, Gateway logs check; backend endpoint implement karo | API readiness/health and contract tests add karo |
| Funnel request returns `404` on `:8086` | Current Session Service funnel route expose nahi karta | Port swap mat karo; handler/gRPC/Gateway implementation complete karo | Route integration test add karo |
| Chart empty but API is `200` | `steps` missing, unknown keys, or first `product_view` count zero | Response payload compare; canonical keys return karo | Response schema validation and contract tests |
| Counts increase at a later step | Backend ordered unique-session semantics enforce nahi kar raha | Aggregation ordering/dedup fix karo | Invariant test: every next count `<=` previous count |
| Paid count too high | Failed/pending payment events counted as success | Only confirmed paid/succeeded result map karo | Payment-event contract and idempotency tests |
| Small segment visible | Backend returned raw small buckets; frontend-only masking incomplete defense hai | Backend suppress/coarsen response before sending | Server-side minimum cohort policy enforce karo |
| Funnel changes do not react to `SESSION_ANALYTICS_*` env | Current Go config loader ignores these names | Typed config implementation add karo | Every env variable ke loader/validation test rakho |
| Browser chart differs from component test | Page test chart ko mock karta hai | `/funnels` manually open karke responsive chart/tooltip verify karo | Dedicated Recharts smoke/E2E visual test add karo |

## 12. Security & Best Practices

### Task-specific security

- Funnel endpoint ko server-side `admin` authorization enforce karna hoga. Sidebar hide karna security nahi hai.
- Aggregate response me email, phone, raw IP, exact address, session ID, user ID, payment data, OTP, password, ya private text kabhi return mat karo.
- Small-count protection backend par apply karo. Current frontend counts mask karta hai, but percentages/chart geometry raw values infer kara sakte hain.
- `suppressed` response ko backend me already safe/coarsened data ke saath return karo. Frontend flag alone sensitive rows ko secure nahi karta.
- Consent-disabled events ko ingestion/aggregation me exclude karo.
- Payment success ko trusted server-side payment event se derive karo; browser-provided `paid` event trust mat karo.
- Date range, allowed funnel steps, filter enums, maximum steps, aur maximum response size server par validate karo.
- Cache keys me admin scope/tenant, date range, steps, filters, and schema version include karo to cross-tenant leakage avoid ho.
- Funnel logs me aggregate metadata/request ID rakho, raw identifiers aur secret-bearing query/context nahi.

### Performance and operations

- Production dashboard reads ke liye `analytics_aggregates` prefer karo; every request par large raw collection scan mat karo.
- Raw fallback ko short, bounded date range tak limit karo.
- Aggregation worker idempotent, checkpointed, observable, and retry-safe hona chahiye.
- Metrics add karo: request latency, aggregate freshness, worker lag, suppressed query count, raw fallback count, error rate.
- Backend monotonicity invariant enforce kare; frontend currently negative drop-off ko zero clamp karta hai, jo bad upstream data hide kar sakta hai.
- `generated_at`, `partial`, aur freshness status UI me visibly surface karna better hai. Current page in fields ko fully communicate nahi karta.

## 13. Missing or Misconfigured Things

| Finding | Impact | Required fix |
|---|---|---|
| Funnel HTTP/gRPC backend absent | Real API impossible | Handler, use case, repository, DTO, auth, tests implement karo |
| API Gateway executable absent | Intended `:8080` proxy target unreachable | Runnable Gateway and route mapping add karo |
| Admin auth integration absent | Protected dashboard cannot authenticate safely | Login/session/JWKS/RBAC wiring complete karo |
| Event ingestion and canonical mapping absent | No reliable funnel source data | Versioned event contract and trusted producers add karo |
| Aggregate worker/read model absent | Production queries slow/impossible | Scheduled/stream worker and repository add karo |
| Funnel index migration absent | Query performance/schema rollout unmanaged | Reviewed idempotent migration + rollback/runbook add karo |
| `SESSION_ANALYTICS_*` funnel env names not loaded | Operators get false configuration confidence | Config struct, parsing, validation, tests add karo or dead names remove karo |
| `partial`/`suppressed` flags not visibly handled | Stale/incomplete/privacy state unclear | Explicit UI states and backend enforcement add karo |
| Frontend-only small-count masking | Values may remain inferable | Server-side suppression/coarsening mandatory |
| No package lockfile | Dependency resolution can drift | Choose package manager and commit reviewed lockfile |
| No application Dockerfiles/Compose | Full container onboarding unavailable | Non-root images, health/readiness, secrets, network config add karo |
| Task guide says `5173` in manual check | Wrong local URL | Use checked-in Vite port `5174` |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `task1_Dependency.md` | Tech Stack and Required Software | Same dashboard stack/runtime |
| `task1_Dependency.md` | Node/npm and Go modules | Same package/module setup |
| `task1_Dependency.md` | MongoDB and Redis setup | Same services, ports, credentials, volumes |
| `task1_Dependency.md` | Environment Variables | Task 4 adds no active frontend/backend variable |
| `task1_Dependency.md` | Docker Setup | No new Task 4 container/network/volume |
| `task1_Dependency.md` | Ports, networking, common errors | Same local infrastructure |
| `task2_Dependency.md` | Gateway/admin auth and polling context | Same missing integration; funnel uses manual refresh instead of polling |
| `task3_Dependency.md` | `session_events`, ordering, retention, privacy | Same event source principles; Task 4 adds aggregate funnel semantics |

Full paths:

```text
TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md
TaskImplementation/Session Analytics Dashboard Service/task2_Dependency.md
TaskImplementation/Session Analytics Dashboard Service/task3_Dependency.md
```

## 15. Final Checklist

### Frontend setup

- [ ] Previous dependency guides read
- [ ] Node/npm versions verified
- [ ] Dashboard dependencies installed
- [ ] `npm ls recharts` succeeds
- [ ] Existing `.env.local` reused; no secrets added to `VITE_` variables
- [ ] Funnel math test passes
- [ ] Funnel page MSW test passes
- [ ] Typecheck and production build pass
- [ ] `http://localhost:5174/funnels` opens
- [ ] Funnel chart, tooltip, cards, table, loading, empty, error, and retry manually checked
- [ ] Filter changes update API query parameters

### Real backend readiness

- [ ] Admin auth and RBAC enforced server-side
- [ ] Gateway `/api/v1/analytics/funnels` route implemented
- [ ] Session Service `GetFunnelReport` implemented
- [ ] Canonical product/cart/checkout/paid events ingested
- [ ] Ordered unique-session aggregation verified
- [ ] `session_events` and `analytics_aggregates` migrations reviewed/applied
- [ ] Aggregate worker, freshness, retry, and observability configured
- [ ] Server-side small-count suppression verified
- [ ] MongoDB/Redis health checks pass where implementation uses them
- [ ] Funnel config variables are actually parsed and validated
- [ ] Real endpoint contract/integration/load tests pass
- [ ] No duplicate setup documentation added
