# Project Dependency & Setup Guide

> **Scope:** Ye incremental handbook Active Sessions page ko install, run, test, aur troubleshoot karne ke liye hai. Common repository, Node.js, Go, MongoDB, Redis, Docker, aur environment setup ko repeat nahi kiya gaya; pehle `task1_Dependency.md` follow karo.
>
> **Repository reality check:** `/live` frontend, typed API client, 15-second polling, filters, masked table values, breakdown panels, aur component tests implemented hain. Real `GET /api/v1/analytics/sessions` backend route, Gateway, admin-auth flow, active-session writer/list use case, compatible response contract, aur analytics migration abhi implemented nahi hain. Isliye UI tests/build run ho sakte hain, lekin real live-session data end to end nahi chalega.

## 1. Project Overview

Task 2 existing dashboard shell ke andar **Active sessions** view add karta hai:

- route: `http://localhost:5174/live`;
- request: `GET /api/v1/analytics/sessions?status=active&limit=50`;
- optional filters: search, device type, country, aur entry page;
- automatic refresh: enabled hone par har `15` seconds;
- views: summary, session table, device mix, approximate locations, aur top entry pages;
- browser request auth: cookies ke liye `credentials: "include"`;
- privacy display: anonymous, user, aur session identifiers UI me masked.

### Current runtime flow

```text
Browser /live (Vite :5174)
  -> /api/v1/analytics/sessions?status=active&... every 15 seconds
  -> Vite /api proxy
  -> intended API Gateway (:8080, currently missing)
  -> intended admin authorization (currently missing)
  -> intended Session Service ListSessions implementation (currently missing)
  -> intended Redis live state + optional MongoDB enrichment
```

### What can run today?

| Mode | Result | Current status |
|---|---|---|
| Focused frontend tests | Mocked active-session response se loading/data/empty/error/filter behavior verify hota hai | Available after dependency install |
| Frontend dev/build | `/live` page render hota hai; API absent ho to error state dikhegi | Available |
| Current Session Service | Privacy routes aur `/healthz` run hote hain | Available with MongoDB, Redis, and backend env |
| Real live sessions | Gateway, admin auth, active writer/list API, and compatible data required | **Blocked** |

## 2. Tech Stack

### Reused technologies

React, TypeScript, Vite, Tailwind CSS, TanStack Query, React Router, Lucide, Vitest, Testing Library, MSW, Go, MongoDB, Redis, aur Docker ka beginner explanation already available hai.

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Sections:
`2. Tech Stack`, `3. Required Software`, and `4. Dependency Management`

### Task 2-specific usage

| Technology | Is task me role | Required? | New installation? |
|---|---|---|---|
| TanStack Query | Cache key filters ko include karta hai, stale time `10s` rakhta hai, aur auto-refresh `15s` par chalata hai | Yes | No |
| Browser Fetch + AbortSignal | API call, cookie credentials, request ID, timeout, aur cancelled stale requests handle karta hai | Yes | No; browser feature |
| React Router | `/live` page aur selected session ka `/journey/{id}` link | Yes | No |
| Tailwind + Lucide | Responsive table/panels aur icons | Yes | No |
| Vitest + Testing Library | Live page data, empty, error, aur filter requests test karte hain | Development only | No |
| Redis | Intended active-session state and short TTL storage | Real backend ke liye yes | Setup reused |
| MongoDB | Intended durable session summaries/enrichment | Full backend ke liye likely yes | Setup reused |

> **Test reality:** Task guide MSW suggest karta hai aur package me MSW installed hai, but current `live-sessions-page.test.tsx` direct `fetch` stubs use karta hai. Task 2 ke liye koi new test service start nahi karni.

> **Map reality:** `ActiveUserMap` geographical map SDK nahi hai; ye location counts ka ranked visual panel hai. Google Maps, Mapbox, API key, ya map tile service required nahi hai.

## 3. Required Software

Task 2 koi new system software introduce nahi karta. Git, Node.js/npm, browser, optional Go, Docker, `mongosh`, aur `redis-cli` requirements Task 1 jaise hi hain.

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section:
`3. Required Software`

Beginner rule:

- Frontend page/tests ke liye Git + Node.js + npm enough hain.
- Current Go Session Service start karne ke liye Go, MongoDB, Redis, and exported backend env chahiye.
- Real Active Sessions flow ke liye software install se zyada missing application implementation complete hona zaroori hai.

## 4. Dependency Management

### Node.js dependencies

Task 2 ke implementation ne `package.json` me koi separate/new dependency add nahi ki. Required React Query, Router, Lucide, Vitest, Testing Library, and MSW packages existing manifest me already declared hain. Task guide ke standalone `npm install <package...>` commands dobara run karne ki zaroorat nahi hai.

Use the existing one-time install documented in:
`task1_Dependency.md` -> `4. Dependency Management` and `9. Local Development Setup`

Current important state:

- `node_modules/` currently absent hai, so first frontend run se pehle existing install step required hai.
- frontend lockfile currently absent hai; `npm ci` abhi work nahi karega.
- install ke baad manifest me unexpected change aaye to blindly accept mat karo.
- no new Task 2 package version or peer dependency needs separate resolution.

### Go modules

Task 2 ka active-session listing backend implemented nahi hai, aur current Go module me sirf MongoDB driver aur Redis client relevant external packages hain. Go module commands, `GOWORK=off`, version mismatch, proxy, checksum, aur cache troubleshooting unchanged hai.

Refer:
`task1_Dependency.md` -> `4. Dependency Management` -> `Go modules`

## 5. Database Setup

### MongoDB

MongoDB installation, Docker container, port `27017`, connection string, credentials placement, and verification unchanged hain.

Refer:
`task1_Dependency.md` -> `5. Database Setup` -> `MongoDB`

Task 2-specific status:

| Requirement | Intended use | Repository status |
|---|---|---|
| Session documents | entry/current page, device, location, timestamps, event count | No active-session schema/repository available |
| Session indexes | status + last activity + filter fields ko efficiently query karna | Migration absent |
| Enrichment | Redis snapshot ko durable session summary ke saath combine karna | Use case absent |
| Seed/event path | Local live rows create karna | Ingestion route/writer absent |

Existing `001_privacy_controls.mongodb.js` **Task 2 migration nahi hai**. Isko active-session collections/indexes banane ke liye use mat karo.

### Redis

Redis installation, Docker container, port `6379`, logical DB `2`, password placement, persistence, and `PONG` verification unchanged hain.

Refer:
`task1_Dependency.md` -> `5. Database Setup` -> `Redis`

Current repository `session:{...}` prefixed keys ko privacy deletion ke time scan/delete kar sakta hai, but it does not:

- incoming activity se active-session hashes write/update;
- activity par TTL refresh;
- expired sessions remove/count;
- filtered list response build;
- summary/breakdown counters compute.

Isliye manually Redis start karna ya random keys insert karna frontend ko populate nahi karega. Final implementation ko one documented key schema, TTL policy, atomic update strategy, and index/set strategy define karni hogi.

## 6. Redis / Queue / External Services

### API Gateway and admin auth

Gateway/admin-auth architecture, missing runnable source, intended ports, cookies, and RBAC limitations Task 1 me documented hain.

Refer:
`task1_Dependency.md` -> `6. Redis / Queue / External Services` -> `API Gateway and admin authentication`

Task 2 endpoint specifically `admin` protected hai according to `api/master-api.json`. Browser-side route ya masked labels authorization ka replacement nahi hain; Gateway/Session boundary par role enforce hona chahiye.

### Geo-location and device parsing

Task 2 city/country, browser, OS, aur device type display kar sakta hai, but repository me GeoIP database/library, IP resolver, or user-agent parser implemented nahi hai.

- `SESSION_GEOIP_ENABLED`, `SESSION_GEOIP_DB_PATH`, and related names ignored backend `.env` me listed hain.
- Current `config.go` in variables ko load nahi karta.
- No GeoIP database file, license/download step, or container is required **for the current runnable code**.
- Future implementation aaye to database/license source, update schedule, timeout, trusted proxy rules, and coarse location policy separately document karna hoga.

### Queue and map services

Kafka, RabbitMQ, NATS, or a map provider Task 2 current code me used nahi hain. Inko install mat karo. High-volume event ingestion later broker use kar sakta hai, but producer/consumer/topic/DLQ implementation ke bina broker setup false dependency hoga.

### Polling capacity note

One open tab auto-refresh on hone par approximately `4 requests/minute` karta hai. `N` simultaneous tabs roughly `4 x N requests/minute` create karenge, retries separate. Backend/Gateway ko caching, request coalescing, timeouts, rate limits, and observability ke saath size karo.

## 7. Environment Variables

### New or changed variables

**None.** Task 2 frontend same three variables use karta hai jo Task 1 me already documented hain:

- `VITE_API_BASE_URL`
- `VITE_API_PROXY_TARGET`
- `VITE_REQUEST_TIMEOUT_MS`

No polling interval environment variable exists; `15s` interval and `10s` stale time source code constants hain. No Task 2 database credential belongs in frontend env.

Refer:
`task1_Dependency.md` -> `7. Environment Variables`

Use the same file locations:

| File | Purpose | Task 2 action |
|---|---|---|
| `frontend/session-analytics-dashboard/.env.local` | Browser base URL, Vite proxy target, timeout | Reuse unchanged |
| `backend/services/session-service/.env` | Current Go process env | Reuse/export unchanged |
| `frontend/session-analytics-dashboard/.env.example` | Safe tracked frontend template | Already present |

> **Important:** Go service `.env` automatically load nahi karta. Task 1 ka `set -a; source .env; set +a` flow use karo.

### Misleading future names

Ignored backend `.env` me `SESSION_ACTIVE_TTL`, `SESSION_ANALYTICS_SESSION_LOOKBACK`, cache TTLs, GeoIP settings, and list-limit variables present hain, but current `config.Load()` inhe read nahi karta. In values ko change karne se Active Sessions behavior change nahi hoga. Implementation ke time only actively loaded and validated settings ko sanitized `.env.example` me publish karo.

## 8. Docker Setup

Task 2 koi new container, image, port mapping, volume, network, restart policy, or health check add nahi karta.

Refer:
`task1_Dependency.md` -> `8. Docker Setup`

Same optional MongoDB + Redis local Compose example reuse karo. Current limitations:

- repository me application Dockerfiles/Compose stack checked in nahi hain;
- no Gateway container exists;
- no active-session worker/consumer container exists;
- containers healthy hone ka matlab missing analytics route implemented hona nahi hai;
- `down -v` local Mongo/Redis data delete karta hai, so casually run mat karo.

## 9. Local Development Setup

### Step 1: Complete common onboarding

First follow:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Required sections:

- `3. Required Software`
- `4. Dependency Management`
- `7. Environment Variables`
- `9. Local Development Setup`

### Step 2: Move to the frontend package

```bash
cd frontend/session-analytics-dashboard
```

No Task 2-only `npm install` command is needed. Complete the reused general install if `node_modules` is absent.

### Step 3: Run focused verification

```bash
npm run typecheck
npm run test -- src/features/live/pages/live-sessions-page.test.tsx src/api/session-api.test.ts
npm run build
```

Focused page test real backend use nahi karta. It stubs `fetch` and verifies data, empty, error, and filter behavior.

### Step 4: Start the frontend

```bash
npm run dev
```

Terminal me shown URL open karo, then `/live` route visit karo. Default URL:

```text
http://localhost:5174/live
```

Gateway absent ho to **Active sessions unavailable** state expected hai. Ye npm/React installation failure nahi hai.

### Step 5: Optional current backend

MongoDB, Redis, backend env loading, Go run, and `/healthz` verification exactly Task 1 jaise hain. No new migration run karni hai.

### Step 6: Verify the intended API

Final integrated request:

```bash
curl -i \
  'http://localhost:8080/api/v1/analytics/sessions?status=active&device_type=mobile&country=India&limit=50'
```

Current checkout me Gateway connection failure expected hai. Direct current Session Service check:

```bash
curl -i \
  'http://localhost:8086/api/v1/analytics/sessions?status=active&limit=50'
# Expected current behavior: 404
```

Final Gateway request ko valid admin browser session/cookie chahiye. Real credential command history me hardcode mat karo.

## 10. Running the Project

### Daily frontend workflow

General dev, watch-test, build, and preview commands unchanged hain.

Refer:
`task1_Dependency.md` -> `10. Running the Project`

Task 2 manual check:

1. `/live` open karo.
2. Device filter change karke Network tab me `status=active`, `device_type`, and `limit=50` verify karo.
3. Auto refresh on ho to requests roughly 15 seconds apart verify karo.
4. Auto refresh off karke polling stop verify karo.
5. Manual refresh button ek request trigger kare.
6. Mocked/integrated response me table and three breakdown panels verify karo.

### Ports and networking

No new port introduced hua.

| Service | Port | Task 2 purpose | Status |
|---|---:|---|---|
| Vite dev server | `5174` | `/live` UI and `/api` proxy | Reused |
| Vite preview | `4174` | Production bundle preview | Reused |
| Intended API Gateway | `8080` | Admin API entry point | Reused, implementation missing |
| Current Session HTTP | `8086` | Health/privacy routes only | Reused, list route missing |
| Intended Session gRPC | `50060` | `ListSessions` downstream contract | Reused, implementation missing |
| Auth/JWKS intent | `8081` | Admin identity/role | Reused, integration missing |
| MongoDB | `27017` | Intended session documents | Reused |
| Redis | `6379` | Intended live state | Reused |

Port conflicts, firewall, CORS, cookie, and Docker-network troubleshooting:
`task1_Dependency.md` -> `10. Running the Project` -> `Ports and networking`

## 11. Common Errors & Fixes

Generic npm, Go, MongoDB, Redis, Docker, port, auth, CORS, timeout, and permission errors are already covered in:
`task1_Dependency.md` -> `11. Common Errors & Fixes`

Only Task 2-specific failures are listed below.

| Error/symptom | Cause | Fix / prevention |
|---|---|---|
| `/live` shows **Active sessions unavailable** | Gateway absent, proxy wrong, `401/403`, or list endpoint missing | Browser Network status check karo; current repo me `404`/connection failure expected hai |
| `INVALID_ACTIVE_SESSIONS_RESPONSE` | Response me required summary, timestamp, sessions, or breakdown arrays missing/wrong type hain | Backend contract ko frontend parser shape se align karo; contract test add karo |
| Master API response works but UI rejects it | Current `SessionListResponse` only `sessions` declares karta hai, while UI also needs summary/breakdowns | Canonical schema update karo or separate endpoints combine through Gateway/BFF |
| Filter UI changes but rows same hain | Backend/Gateway `status`, `device_type`, `country`, `entry_page`, `q`, or `limit` ignore karta hai | Query schema and server validation implement/document karo |
| Too many repeated requests | Multiple tabs, 15-second polling, and retries | Auto-refresh pause karo; backend cache/rate limit add karo; duplicate dashboard tabs avoid karo |
| Location shows `Unknown location` | Geo enrichment absent or privacy setting suppresses fields | Treat as valid state; raw IP expose karke workaround mat karo |
| Device shows `unknown` | User-agent parsing/enrichment absent or input unavailable | Server-side bounded parser add karo; unknown fallback preserve karo |
| Active list unexpectedly empty | No ingestion/writer exists, TTL expired, wrong Redis DB/prefix, or query time window differs | Writer and TTL logs inspect karo; DB/prefix and clock synchronization verify karo |
| New Redis keys exist but UI remains empty | Current repository only deletion-oriented key scanning has; list route absent | Repository/use case/handler implement karo; key injection alone enough nahi hai |
| Polling appears stopped | Auto refresh disabled, tab throttled, request still pending, or query error retries exhausted | Toggle state and Network tab inspect karo; manual refresh test karo |
| Journey link fails | Task 3 journey endpoint/integration separate hai | Live listing ko journey backend readiness se confuse mat karo |

## 12. Security & Best Practices

Common secrets, database exposure, cookies, CSRF, TLS, logging, container, and dependency advice:
`task1_Dependency.md` -> `12. Security & Best Practices`

Task 2-specific additions:

- **Server-side masking required:** UI masking sirf screen text hide karta hai. Raw `sessionId` and `anonymousId` Network tab/React state me aa sakte hain. API minimum necessary identifiers return kare.
- **Admin authorization on every poll:** cached response bhi user/tenant/role boundary respect kare. Public or shared cache me admin analytics response store mat karo.
- **Approximate location only:** city/country coarse data use karo; raw IP return/log mat karo. Retention and lawful purpose define karo.
- **Bounded user agent:** full user-agent string UI ko normally required nahi hai. Length cap, sanitization, and retention policy enforce karo.
- **Validated filters:** limits, allowed device values, max query/path lengths, date ranges, and pagination server-side enforce karo. Frontend validation security control nahi hai.
- **Efficient Redis access:** request path par broad production `KEYS` or repeated full `SCAN` avoid karo. Active IDs ke liye TTL-backed sorted set/set or another indexed design use karo.
- **Polling protection:** short cache, request coalescing, per-admin rate limits, cancellation, and dependency timeout define karo.
- **Clock consistency:** `lastSeenAt`, TTL, and “active” window servers ke UTC clocks par depend karte hain; NTP/time synchronization monitor karo.
- **Safe observability:** latency, result count, cache hit, and upstream errors log/measure karo; cookies, raw IDs, IPs, and search text avoid karo.
- **Privacy-safe small segments:** location/entry-page breakdowns me small-count suppression consider karo so one user identifiable na ho.

## 13. Missing or Misconfigured Things

| Severity | Finding | Professional fix |
|---|---|---|
| Blocker | Current Session Service does not register `GET /api/v1/analytics/sessions` | Handler, use case, repository, validation, and integration tests implement karo |
| Blocker | Runnable Gateway and admin authentication flow absent hain | Gateway route + `ListSessions` mapping + admin RBAC/session verification add karo |
| Blocker | Active-session ingestion/writer absent hai | Event ingestion se Redis state atomically write/refresh and expiry policy implement karo |
| High | Frontend response needs counts, `refreshedAt`, and three breakdown arrays; master `SessionListResponse` only sessions list declares karta hai | One canonical response schema choose/update karo and consumer-provider contract tests add karo |
| High | Frontend sends six active filters not declared by current `SessionListRequest` | API schema, Gateway mapping, and server validation align karo |
| High | Mongo session schema/index migration and local seed path absent hain | Versioned collections/index migration plus safe sample ingestion/seed command add karo |
| High | Redis key naming writer/read model undefined hai; current repository supports deletion lookup only | One versioned key schema, TTL, set/index membership, and cleanup behavior document/test karo |
| High | Device/location enrichment implementation absent hai | Privacy-reviewed UA parser and optional coarse GeoIP resolver add karo |
| Medium | Backend `.env` advertises active TTL, analytics limits/cache, and GeoIP values which config ignores | Implement typed loading/validation or remove stale variables; create sanitized backend `.env.example` |
| Medium | Poll interval is hardcoded and visibility/background behavior is not explicitly tuned | Configurable bounded interval or documented product constant add karo; load test expected admin concurrency |
| Medium | Task guide says MSW, but current live page test uses direct fetch stubs | Either docs say fetch stub or shared MSW handlers implement karo; avoid two conflicting test patterns |
| Medium | No Task 2 backend readiness check exists | `/readyz` should verify required stores and worker/data-path readiness without leaking secrets |

Unchanged broader gaps such as lockfile, Dockerfiles, `go.work`, shallow health, and deployment runbooks are not repeated here.

Refer:
`task1_Dependency.md` -> `13. Missing or Misconfigured Things`

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `task1_Dependency.md` | `2. Tech Stack` | Same React/Vite/TypeScript/Go stack |
| `task1_Dependency.md` | `3. Required Software` | No new operating-system tool |
| `task1_Dependency.md` | `4. Dependency Management` | Same `package.json`, npm, Go module, and workspace behavior |
| `task1_Dependency.md` | `5. Database Setup` | Same MongoDB and Redis instances, ports, credentials, and Docker setup |
| `task1_Dependency.md` | `6. Redis / Queue / External Services` | Same missing Gateway/admin-auth integration and no mandatory broker |
| `task1_Dependency.md` | `7. Environment Variables` | No new or changed loaded variable |
| `task1_Dependency.md` | `8. Docker Setup` | No new container, volume, network, or image |
| `task1_Dependency.md` | `9. Local Development Setup` | Clone/install/config/start flow unchanged |
| `task1_Dependency.md` | `10. Running the Project` | Commands and ports reused |
| `task1_Dependency.md` | `11. Common Errors & Fixes` | Generic npm/Go/DB/Docker/network troubleshooting unchanged |
| `task1_Dependency.md` | `12. Security & Best Practices` | Shared secret, cookie, database, and deployment rules |
| `task1_Dependency.md` | `13. Missing or Misconfigured Things` | Shared platform blockers already audited |

## 15. Final Checklist

### Frontend verification

- [ ] `task1_Dependency.md` common setup completed
- [ ] No duplicate Task 2 dependency install attempted
- [ ] Existing `.env.local` reused; no secret added to `VITE_` variables
- [ ] `npm run typecheck` passes
- [ ] Active Sessions page and API client focused tests pass
- [ ] `npm run build` passes
- [ ] `/live` opens from the dashboard navigation
- [ ] Data, empty, error, and refresh states checked
- [ ] Filter query parameters checked in Network tab
- [ ] 15-second auto-refresh and pause behavior checked
- [ ] UI-only error state understood as expected while backend is missing

### Backend/infrastructure readiness

- [ ] No unrelated Kafka, RabbitMQ, map SDK, or GeoIP service installed
- [ ] No privacy migration mistaken for an active-session migration
- [ ] MongoDB/Redis reused only when current/future backend is being run
- [ ] `GET /api/v1/analytics/sessions` implemented and tested
- [ ] Active-session writer, TTL, and indexed list strategy implemented
- [ ] Frontend and master API request/response schemas aligned
- [ ] Gateway route and Session transport aligned
- [ ] Admin authentication/RBAC enforced server-side
- [ ] Device/location enrichment is optional and privacy-safe
- [ ] Required Mongo indexes/migration and local data path exist
- [ ] Dependency-aware readiness, logs, metrics, and polling capacity verified
- [ ] No duplicate setup documentation added

---

**Beginner mental model:** Task 2 ka frontend khud test aur build ho sakta hai. MongoDB/Redis containers start karna useful infrastructure deta hai, but live rows tabhi aayengi jab activity writer, compatible list API, Gateway, aur admin authorization sab implemented aur connected hon.
