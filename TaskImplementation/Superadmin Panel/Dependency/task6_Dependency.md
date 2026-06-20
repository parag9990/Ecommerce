# Superadmin Panel Task 6 - Dependency & Setup Guide

## 1. Project Overview

Yeh handbook `task6.md` ke **Session Oversight** module ko locally run, test, debug, aur future full-stack environment me connect karne ke liye dependency and setup requirements explain karti hai. Business logic aur implementation snippets yahan repeat nahi kiye gaye hain.

Actual frontend implementation:

```text
frontend/superadmin-panel/src/features/sessions/
frontend/superadmin-panel/tests/sessions/
```

Module me live traffic polling, session filters/table, UI-side risk triage, suspicious/high-risk views, journey drawer, privacy-safe event rendering, role checks, aur automated tests present hain.

### Current repository reality

| Capability | Current status | Beginner meaning |
|---|---|---|
| Task 6 frontend source | Ready | Route, menu, components, API client, Query hooks, permissions, and tests present hain |
| Frontend install/test/build | Ready | Existing pnpm workspace and lockfile use hote hain |
| New Task 6 npm dependency | None | Koi package add karne ki zarurat nahi hai |
| New frontend environment variable | None | Existing frontend template reuse hota hai |
| Session REST contract | Present | Three required routes `api/master-api.json` me defined hain |
| Session backend runtime | Missing | `backend/services/session-service/` me ignored local `.env` hai, but source/module/server nahi hai |
| API Gateway/Auth runtime | Missing | Env/design contract mila; runnable source nahi mila |
| MongoDB/Redis integration | Designed/configured only | Schema and env names mile; Session Service code nahi mila |
| Docker/Compose stack | Missing | Supported Dockerfile or Compose manifest nahi mila |

> **Important:** Task 6 tests, typecheck, build, aur frontend dev server backend ke bina run ho sakte hain. Real live metrics, session rows, and journeys ke liye runnable Auth, Gateway, and Session Service plus MongoDB and Redis required hain.

### Task-specific API dependency

| Method | Frontend/contract path | Purpose | Current status |
|---|---|---|---|
| `GET` | `/api/v1/analytics/live` | Active users, sessions, events/minute | Contract only |
| `GET` | `/api/v1/analytics/sessions` | Filtered/paginated session list | Contract only |
| `GET` | `/api/v1/analytics/sessions/{session_id}/journey` | Session and ordered events | Contract only |

Intended safe flow:

```text
Admin browser -> API Gateway -> JWT/RBAC -> Session Service
                                             |-> Redis (live counters/cache)
                                             `-> MongoDB (sessions/events)
```

Browser ko MongoDB, Redis, internal gRPC, hash salts, ya backend credentials directly access nahi karne chahiye.

## 2. Tech Stack

Task 6 same frontend foundation reuse karta hai. Full beginner definitions and installation dobara repeat nahi ki gayi hain.

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Sections:
`2. Tech Stack`, `3. Required Software`, and `4. Dependency Management`

Task-specific mapping:

| Technology/service | Task 6 me use | Required? | New? |
|---|---|---:|---:|
| React 19 + TypeScript | Oversight page, tables, filters, drawer, typed payloads | Yes | No |
| React Router | Protected `/admin/sessions` route | Yes | No |
| TanStack React Query | Live polling, list query, journey query, retry/cache states | Yes | No |
| Zustand | Admin access token and roles | Yes | No |
| Vite + Tailwind CSS | Local server, build, and styling | Yes | No |
| Lucide React | Operational UI icons | Yes | No |
| Vitest + Testing Library + jsdom | Session API, risk, permission, and component tests | Tests only | No |
| API Gateway + Auth/JWKS | Authenticated REST boundary | Real integration | Reused; runtime missing |
| Session Service | Analytics and journey owner | Real integration | Task 6 core; runtime missing |
| MongoDB | Durable sessions and event journeys | Real integration | Already documented; Task 6 core usage |
| Redis | Active-session counters/cache; Gateway rate limiting | Real integration | Already documented; Task 6 core usage |
| gRPC/Protobuf | Intended Gateway-to-Session call | Real integration | Contract only; implementation missing |

**Simple Hinglish:** MongoDB flexible/high-volume event documents store karta hai. Redis memory me short-lived active-session counters fast rakhta hai. React frontend dono ko directly call nahi karta; Session Service approved result Gateway ke through return karta hai.

### Planning guide versus actual manifest

`task6.md` generic install examples me `clsx`, MSW, and jest-dom mention karta hai. Current code and `package.json` ke according:

- `clsx` import nahi hota; project-local styling helpers enough hain.
- MSW installed/imported nahi hai; tests Vitest `fetch` mocks use karte hain.
- jest-dom current assertions ke liye imported nahi hai.
- React, Query, Router, Zustand, Lucide, Vitest, Testing Library, and jsdom already locked hain.

> `task6.md` ke package-add commands current repository par mat run karein. Actual imports, `package.json`, and `pnpm-lock.yaml` source of truth hain.

## 3. Required Software

Git, Node.js 22.12+, pnpm 10, aur modern browser setup unchanged hai.

Refer:
`TaskImplementation/Superadmin Panel/task1_Dependency.md`

Section:
`3. Required Software`

| Validation mode | Required software/services |
|---|---|
| Session tests/typecheck/build | Node.js, pnpm, locked workspace packages |
| Frontend dev page/error/permission states | Above plus modern browser |
| Real session list/journey | API Gateway, Auth/JWKS, Session Service, MongoDB, and Gateway Redis dependency |
| Real live traffic | Above plus Session Redis counters and active ingestion |
| Production oversight | Above plus TLS, granular RBAC, retention worker, monitoring, backups, and private networking |

No Kafka, RabbitMQ, NATS, MySQL, object storage, SMTP, or payment-provider setup is directly required for the current Task 6 read path.

## 4. Dependency Management

### New dependency delta

| Check | Result |
|---|---|
| New runtime npm package | None |
| New development npm package | None |
| `package.json` change needed | No |
| `pnpm-lock.yaml` change needed | No |
| New frontend install command | No |
| Session backend Go module | Missing; cannot install/build yet |

Existing pnpm files, frozen installs, scripts, Node-version issues, cache/proxy fixes, and package rules are documented in:

`TaskImplementation/Superadmin Panel/task1_Dependency.md`, section `4. Dependency Management`

Correct install remains:

```bash
cd frontend
pnpm install --frozen-lockfile
```

Beginner rules:

- npm/yarn lockfiles add mat karein; this workspace pnpm use karta hai.
- `node_modules/` commit mat karein.
- Sirf guide me package named hai isliye `pnpm add` mat chalayein; imports/manifest verify karein.
- Missing Session Service directory me `go mod tidy` ya `go run` service create nahi karega.
- Future Go module/workspace setup ke liye `Dependency/Go_Modules.md` refer karein.

## 5. Database Setup

### MongoDB requirement

MongoDB ek document database hai. Task 6 ke real list and journey data ke liye **indirectly mandatory** hai; frontend-only tests/build ke liye optional hai.

Generic local installation, one-off Docker command, start, ping, and common errors already documented hain.

Refer:
`TaskImplementation/Superadmin Panel/Dependency/MongoDB.md`

Sections:
`5. Installation steps` through `12. Security notes`

Task-specific data:

| Collection | Task 6 use | Required indexes/effect |
|---|---|---|
| `sessions` | Session identity, status, device, start/last-seen time | user/anonymous/status timeline queries |
| `session_events` | Ordered journey events | session journey and event/date queries |

The repository design defines these indexes:

```javascript
db.sessions.createIndex({ anonymous_id: 1, started_at: -1 })
db.sessions.createIndex({ user_id: 1, started_at: -1 })
db.sessions.createIndex({ status: 1, last_seen_at: -1 })
db.session_events.createIndex({ session_id: 1, occurred_at: 1 })
db.session_events.createIndex({ user_id: 1, occurred_at: -1 })
db.session_events.createIndex({ event_type: 1, occurred_at: -1 })
db.session_events.createIndex({ occurred_at: 1 }, { expireAfterSeconds: 7776000 })
```

After an approved local schema/index bootstrap, verify:

```bash
mongosh session_db --eval 'db.runCommand({ ping: 1 }); db.sessions.getIndexes(); db.session_events.getIndexes();'
```

`7776000` seconds means 90 days, which matches the detected local `SESSION_RAW_EVENT_TTL_DAYS=90`. Mongo TTL cleanup is asynchronous, so expiry instant par document immediately disappear hona guaranteed nahi hai.

### Credentials placement

- Mongo URI and database name Session Service ke backend environment/secret store me rahenge.
- Frontend `.env` me Mongo credentials kabhi mat rakhein.
- Detected exact names are `SESSION_MONGO_URI` and `SESSION_MONGO_DATABASE`.
- Local `.env` Git se ignored hai; committed sanitized Session Service `.env.example` missing hai.
- Production me authenticated TLS URI, least-privilege DB user, private network, and secret manager use karein.

### Migration/bootstrap status

Mongo schema design document exists, but executable/versioned Session Service migration or index runner nahi mila. `database/mongodb-schema-design.md` reference design hai, automatic startup migration proof nahi.

Task 6 does not use Superadmin MySQL directly. MySQL start/migration sirf unrelated admin modules ke liye run karna unnecessary hai.

## 6. Redis, Queue, and External Services

### Redis

Redis ek in-memory data store hai. Task 6 me do logical uses detected hain:

| Owner | Address/database | Purpose | Status |
|---|---|---|---|
| API Gateway | `localhost:6379`, DB `0` | Rate limiting | Env only; Gateway runtime missing |
| Session Service | `localhost:6379`, DB `2`, prefix `session` | Active sessions, live counters, analytics cache | Env only; Session runtime missing |

Same local Redis server different logical DB numbers use kar sakta hai. Lekin logical DB isolation security boundary nahi hai; production me ACLs/separate instances or clusters policy ke according use karein.

Installation, one-off Docker setup, start, `PONG` verification, and generic fixes already documented hain:

`TaskImplementation/Superadmin Panel/Dependency/Redis.md`, sections `5` through `12`

Task-specific verification, only after the future Session Service writes test data:

```bash
redis-cli -h localhost -p 6379 -n 2 PING
redis-cli -h localhost -p 6379 -n 2 --scan --pattern 'session:*' | head
```

Production par `KEYS session:*` mat run karein; large keyspace block ho sakta hai. `SCAN` bhi controlled diagnostics ke liye use karein, data export ke liye nahi.

### API Gateway, authentication, and gRPC

Generic Gateway/JWT/Redis/gRPC guidance reuse hoti hai:

- `Dependency/API_Gateway.md`
- `Dependency/gRPC.md`
- `task2_Dependency.md`, sections `6` and `7`

Task 6 specifically expects Gateway HTTP `8080` and Gateway target `SESSION_GRPC_ADDR=localhost:50060`. Contract `SessionService.GetLiveMetrics`, `ListSessions`, and `GetJourney` methods declare karta hai.

Current mismatch: Session local env only `SESSION_HTTP_ADDR=:8086` provide karta hai. Session gRPC listen address, proto/generated code, server registration, and runtime source nahi mile. Isliye `50060` par actual server start karne ka honest command available nahi hai.

### Kafka/RabbitMQ

Architecture docs high-volume event downstream processing ke liye Kafka/RabbitMQ suggest karte hain, but current Task 6 frontend, Session env, and runnable code me selected broker URL/topic/producer/consumer nahi mila.

Therefore:

- Frontend validation ke liye broker install mat karein.
- Current read API setup ke liye invented Kafka/RabbitMQ container mat start karein.
- Broker tab mandatory hoga jab one approved implementation, topic contract, retry/dead-letter policy, and env template repository me add hon.

## 7. Environment Variables

### Frontend delta

Task 6 introduces **no new frontend environment variable**.

Create/load/security rules already documented hain:

`TaskImplementation/Superadmin Panel/task1_Dependency.md`, section `7. Environment Variables`

Existing template contains:

```dotenv
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_APP_NAME=Superadmin Panel
VITE_ADMIN_SESSION_WARNING_MINUTES=2
```

Task 6 only reads shared `VITE_API_BASE_URL`. Mongo URIs, Redis passwords, JWT secrets, IP salts, and privacy peppers must never use a `VITE_*` name because Vite values browser bundle me visible hote hain.

### Existing API base-path blocker

This shared issue was first documented in `task2_Dependency.md`, section `7. Environment Variables` -> `Critical base-path mismatch`.

Current base already `/api/v1` par end hota hai, while Task 6 API constants bhi `/api/v1/...` se start hote hain. Result:

```text
http://localhost:8080/api/v1/api/v1/analytics/live
http://localhost:8080/api/v1/api/v1/analytics/sessions
```

Tests isko catch nahi karte because they stub base as origin-only. Sirf base URL ko origin par change karna Task 6 fix karega but current login `/auth/login` composition break karega.

> **Required alignment:** One convention choose karein. Recommended: base `http://localhost:8080/api/v1`, feature paths `/analytics/live` and `/analytics/sessions`, login `/auth/login`. Any change ke baad Vite restart and production-like URL tests add karein.

### Session Service task-specific backend template

Previous guides install/secret rules cover karte hain, but below exact variable names Task 6 runtime intent me newly relevant hain. Source/config loader absent hone ke karan “required” classification design-based hai, code-verified nahi.

Create a sanitized committed template in the future; real secrets ignored local env or secret manager me inject karein:

```dotenv
# Listener (HTTP listener is detected; gRPC listener configuration is still missing)
SESSION_HTTP_ADDR=:8086

# Mandatory for stored session lists/journeys
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_MONGO_DATABASE=session_db
SESSION_MONGO_CONNECT_TIMEOUT=5s

# Mandatory for live counters/cache
SESSION_REDIS_ADDR=localhost:6379
SESSION_REDIS_PASSWORD=
SESSION_REDIS_DB=2
SESSION_REDIS_KEY_PREFIX=session
SESSION_REDIS_DIAL_TIMEOUT=2s
SESSION_REDIS_READ_TIMEOUT=1s
SESSION_REDIS_WRITE_TIMEOUT=1s

# Privacy secrets: replace with independent random secret-manager values
SESSION_IP_HASH_SALT=replace-with-a-long-random-local-secret
SESSION_PRIVACY_HASH_PEPPER=replace-with-a-different-long-random-local-secret
SESSION_TRUSTED_PROXY_CIDRS=127.0.0.1/32,::1/128
SESSION_STORE_USER_AGENT=false

# Retention and active-session behavior
SESSION_ACTIVE_TTL=35m
SESSION_RAW_EVENT_TTL_DAYS=90
SESSION_METADATA_RETENTION_DAYS=365
SESSION_JOURNEY_RETENTION_DAYS=365
SESSION_RETENTION_DRY_RUN=true

# List/journey limits
SESSION_DEFAULT_LIST_LIMIT=100
SESSION_MAX_LIST_LIMIT=500
SESSION_JOURNEY_DEFAULT_LIMIT=1000
SESSION_JOURNEY_MAX_LIMIT=1000

# Oversight analytics bounds/cache
SESSION_ANALYTICS_LIVE_WINDOW=5m
SESSION_ANALYTICS_SESSION_LOOKBACK=24h
SESSION_ANALYTICS_MAX_SESSION_RANGE_DAYS=90
SESSION_ANALYTICS_DEFAULT_PAGE_SIZE=50
SESSION_ANALYTICS_MAX_PAGE_SIZE=100
SESSION_ANALYTICS_LIVE_CACHE_TTL=10s
SESSION_ANALYTICS_SESSIONS_CACHE_TTL=30s
```

Notes:

- `SESSION_RETENTION_DRY_RUN=true` safer first-run setting hai; deletion counts inspect karke only then approved environment me disable karein.
- Salt and pepper same value mat rakhein; placeholders production secrets nahi hain.
- `SESSION_STORE_USER_AGENT=false` privacy-minimizing default hai. If business need enables it, length limits, redaction, access, and retention approve karein.
- Docker me `localhost` another container ko point nahi karta. Future Compose/Kubernetes me private service DNS use hoga.
- Heatmap, funnel, cohort, and GeoIP variables intentionally omitted hain because those Task 6 scope me nahi hain.

## 8. Docker Setup, Ports, and Networking

### Docker status

Task 6 implementation koi supported Dockerfile, Compose service, network, volume, health check, or restart policy add nahi karti. No root Compose manifest was found.

Reuse:

- `task1_Dependency.md`, section `8. Docker Setup` for frontend limitation.
- `Dependency/MongoDB.md`, section `6` for the existing suggested Mongo one-off command.
- `Dependency/Redis.md`, section `6` for the existing suggested Redis one-off command.

Those commands examples hain, project-supported stack nahi. Persistent Mongo data ke liye an approved named volume/bind mount required hoga before relying on container data. Invented `docker compose up` command mat run karein.

### Ports table

| Service | Port/address | Purpose | Task 6 status |
|---|---:|---|---|
| Vite frontend | `5173` default | Local admin UI | Reused; runnable |
| Vite preview | `4173` commonly | Built bundle preview | Reused; optional |
| API Gateway HTTP | `8080` | Browser REST entry | Reused; runtime missing |
| Auth/JWKS HTTP | `8081` in current Gateway config | Token/JWKS dependency | Reused; runtime missing |
| Session HTTP | `8086` | Detected Session listener | Task 6 env only; purpose/routing unverified |
| Session gRPC | `50060` | Gateway analytics target | Task 6 required; server missing |
| MongoDB | `27017` | `session_db` | Task 6 required full stack |
| Redis | `6379` | Gateway DB 0 and Session DB 2 | Task 6 required full stack |

Networking rules:

- Browser only Gateway ko call kare.
- Gateway exact frontend origin (normally `http://localhost:5173`) and `Authorization`, `Content-Type`, `X-Request-ID` headers allow kare.
- MongoDB, Redis, Session HTTP/gRPC public internet par expose mat karein.
- Production internal traffic authenticated/encrypted ho; `GRPC_TLS_ENABLED=false` local-only setting hai.
- Port busy ho to `ss -ltnp | grep ':5173\|:8080\|:8086\|:50060\|:27017\|:6379'` se owner inspect karein; random port change se all dependent configs bhi update honge.

## 9. Local Development Setup

### Step 1: Read reused setup first

Follow:

- `task1_Dependency.md` for clone, Node/pnpm, install, frontend `.env`, and generic startup.
- `task2_Dependency.md`, section 7, for unresolved URL composition.
- `Dependency/MongoDB.md` and `Dependency/Redis.md` only for real integration infrastructure.

### Step 2: Go to the workspace

```bash
cd Ecommerce/frontend
```

### Step 3: Install locked dependencies

```bash
pnpm install --frozen-lockfile
```

No Task 6 package-add command is needed.

### Step 4: Configure frontend environment

```bash
cp superadmin-panel/.env.example superadmin-panel/.env
```

No Task 6 variable add karein. Real integration se pehle section 7 ka `/api/v1` convention resolve karein.

### Step 5: Choose validation mode

Frontend-only:

- MongoDB, Redis, Gateway, Auth, Session Service, gRPC, and migrations skip kar sakte hain.
- Existing fetch-mocked tests module ka reliable local validation boundary hain.

Full stack:

1. Obtain runnable Auth, Gateway, and Session Service source/manifests.
2. Add sanitized backend env templates and a Session gRPC listener configuration.
3. Start MongoDB and create/verify approved indexes.
4. Start Redis and verify DB 0/DB 2 access.
5. Start ingestion so synthetic sessions/events actually exist.
6. Start Session Service and verify Mongo/Redis readiness.
7. Start Auth/JWKS, then Gateway, then verify downstream readiness.
8. Provision a least-privilege non-production admin and privacy-safe synthetic sessions.

Exact backend build/start commands cannot honestly be supplied until source, modules, and server entry points exist.

### Step 6: Run Task 6 tests

```bash
cd frontend/superadmin-panel
pnpm exec vitest run tests/sessions
```

### Step 7: Run frontend quality checks

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
http://localhost:5173/admin/sessions
```

Without a valid stored admin session, route `/login` par redirect karega. Backend unavailable ho to live/list requests fail hona npm install problem nahi hai.

### Verified frontend baseline

Repository verification on 2026-06-20:

| Check | Result |
|---|---|
| Task 6 session tests | `7` files, `15` tests passed |
| TypeScript project check | Passed |
| Vite production build | Passed |
| Build note | Main JavaScript chunk `560.24 kB`; optimization warning only |
| New package required | None |

### Full-stack API verification

Only after missing runtimes and URL/contract alignment are fixed, use a temporary non-production token:

```bash
export ADMIN_TOKEN='<temporary-admin-access-token>'

curl -i \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "X-Request-ID: task6-live-check" \
  'http://localhost:8080/api/v1/analytics/live'

curl -i \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "X-Request-ID: task6-list-check" \
  'http://localhost:8080/api/v1/analytics/sessions?page=1&page_size=25'
```

For a synthetic session:

```bash
export TEST_SESSION_ID='<synthetic-session-id>'
curl -i \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "X-Request-ID: task6-journey-check" \
  "http://localhost:8080/api/v1/analytics/sessions/$TEST_SESSION_ID/journey"
```

Token shared machine shell history me mat paste karein. Customer session IDs and real journey data casual debugging fixtures ke roop me use mat karein.

### Expected frontend permissions

| Role | View sessions/live/journey | Unmasked session PII |
|---|---:|---:|
| `superadmin` | Yes | No |
| `operations_admin` | Yes | No |
| `readonly_admin` | Yes | No |
| `finance_admin` | No | No |
| `catalog_admin` | No | No |

Backend must enforce the approved matrix. Current contract's generic `admin` role definition does not include `readonly_admin` and includes finance/catalog roles, so it is not aligned with the UI.

## 11. Common Errors & Fixes

Generic Node, pnpm, login, Docker, CORS, and frontend port issues already documented hain:

`TaskImplementation/Superadmin Panel/task1_Dependency.md`, section `11. Common Errors & Fixes`

Task 6-specific troubleshooting:

| Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| URL contains `/api/v1/api/v1/` | Base and session constants both include prefix | Normalize one shared path convention; section 7 dekhein | Production-like config URL tests |
| Tests pass but browser gets `404` | Tests use origin-only base, unlike `.env.example` | Inspect final Network-tab URL and fix convention | Test real template composition |
| Pagination ignored | Frontend sends `limit`; contract declares `page_size` | Align client/server field and response metadata | Contract-generated request types/tests |
| `readonly_admin` gets `403` | Frontend allows it, generic contract role does not | Approve endpoint-specific RBAC and align both sides | One authoritative permission matrix |
| Finance/catalog can call direct API | Generic contract permits them while UI denies | Enforce session-specific backend permission | Never rely on hidden menu |
| Live cards show zero/unknown update time | Backend missing/unreachable or contract lacks optional UI fields | Check Gateway/Session/Redis and response schema | Readiness plus response contract tests |
| Live values never refresh | Poll request fails, tab/network paused, or cache data stale | Inspect 15-second requests and Redis writer | Metrics for cache freshness/writer lag |
| List returns `502`/Unavailable | Gateway cannot dial Session gRPC `50060` | Supply/start gRPC server; verify DNS/TLS/readiness | Dependency-aware Gateway readiness |
| Session HTTP `8086` is healthy but Gateway fails | Gateway expects gRPC `50060`, not detected HTTP listener | Implement/configure intended gRPC boundary | Document both listeners and probes |
| Sessions page is empty | No ingestion, retention expired, wrong DB, or no synthetic events | Verify ingest path, Mongo collection/indexes, time range | Seed privacy-safe integration fixtures |
| Every device appears unknown | DB design uses `device.type`; frontend reads `device_type` | Add explicit backend/client mapping | Shared generated response schema |
| Location is unknown | DB design stores `geo`; frontend expects city/country in `device` | Define admin response projection | Contract fixture tests |
| Suspicious/high-risk views stay empty | Public contract lacks risk/aggregate fields and list rows lack events | Add approved backend risk projection or summary fields | Contract versioning and risk fixtures |
| Status/risk filter misses records | Filters apply only to current fetched page | Move filters server-side or label client-page filtering | Indexed backend filters |
| Journey is empty/slow | Events expired, session id mismatch, or missing `{session_id, occurred_at}` index | Check TTL/id, then indexes/query plan | Index/retention monitoring |
| Redis `NOAUTH` | Password/ACL mismatch | Use backend secret and matching ACL | Managed secret rotation and readiness |
| Mongo connection refused | Mongo not running/wrong host; Docker `localhost` mistake | Start Mongo or use service DNS | Supported Compose/K8s network config |
| Old events still exist after TTL | Mongo TTL monitor is asynchronous | Wait and inspect TTL index | Alert on TTL index drift/storage growth |

## 12. Security & Best Practices

### Task-specific security rules

- Backend authorize every live/list/journey request; React guards are UX only.
- Apply least privilege. Finance/catalog roles ko session telemetry accidentally expose mat karein.
- `readonly_admin` policy explicitly decide karein and contract/backend/UI align karein.
- Raw IP store/return karne ke bajaye approved keyed pseudonymization use karein. Salt/pepper logs or frontend me expose na ho.
- Current UI identifiers mask karti hai and event properties allowlist karti hai. Backend still must remove passwords, OTPs, tokens, cookies, card/CVV, email, phone, address, free text, search query, and private form data.
- URLs se query string/fragment strip karein because they can contain tokens or PII; current timeline already display path sanitize karti hai.
- UI-side risk score **triage helper** hai, fraud verdict nahi. Account lock/revoke/payment action is score se automatically trigger mat karein.
- Backend-provided risk score must be explainable, versioned, bounded, monitored for bias/false positives, and access-controlled.
- Public event ingestion endpoint ko body/rate limits, schema validation, timestamp bounds, abuse protection, and server-side property allowlist chahiye.
- MongoDB/Redis/gRPC private rakhein; external TLS and internal authenticated encryption use karein.
- Retention settings privacy policy ke saath align karein. Deletion worker first dry-run, then metrics/alerts/approval ke saath enable karein.
- `X-Request-ID` end-to-end preserve karein; Bearer token, raw event body, device fingerprint, raw IP, salts, and URIs log mat karein.
- Use synthetic sessions for QA. Real user journey viewing access and purpose audit hona chahiye.

### Operational best practices

- Separate liveness from readiness for Gateway, Session Service, MongoDB, and Redis.
- Metrics: ingest rate/failures, active-counter lag, live-cache age, query latency, journey size, risk distribution, TTL/index drift, Redis errors, and downstream timeouts.
- Put hard maximums on date range, page size, journey events, body size, event properties/depth, and clock skew.
- Use stable pagination; large production lists ke liye cursor pagination preferable ho sakti hai.
- Back up durable session metadata only according to approved privacy policy; restore drill retention/deletion guarantees violate nahi karna chahiye.
- Deploy frontend code splitting for large modules; current 560.24 kB main chunk warning release blocker nahi, but lazy route imports improve load time.

## 13. Missing or Misconfigured Things

| Severity | Finding | Impact | Required action |
|---|---|---|---|
| Critical | Session Service, Gateway, and Auth runnable source absent | Full-stack Task 6 cannot start | Add source/modules, entry points, manifests, health checks, and READMEs |
| Critical | Frontend duplicates `/api/v1` with current template | Real Task 6 URLs are wrong | Normalize shared URL convention and integration-test it |
| Critical | Gateway expects Session gRPC `50060`, but Session env only exposes HTTP `8086` | Gateway downstream cannot be started from found config | Add proto/generated server and explicit gRPC listener/config |
| High | `limit` client query conflicts with contract `page_size` | Pagination may be ignored/rejected | Choose one contract field and generate/test client |
| High | Contract role set conflicts with frontend roles | Overexposure or `readonly_admin` denial | Define endpoint-specific backend permission |
| High | Contract `Session` lacks status, risk, aggregate, and active-session fields used by UI | High-risk/suspicious views cannot work reliably | Version admin session response schema |
| High | Contract list response lacks `total`, page, and limit metadata expected by UI | Pagination count/navigation unreliable | Add canonical pagination response |
| High | Contract live response lacks `suspicious_sessions` and `updated_at` used by UI | UI defaults hide freshness/coverage gap | Add fields or remove unsupported UI expectations |
| High | Risk/status filters are client-side over one fetched page | Results can be incomplete/misleading | Add indexed server filters or clearly scope UI |
| High | DB design `device.type`/separate `geo` differ from UI `device_type`/embedded location | Unknown device/location and inflated UI risk | Define explicit backend projection/mapping |
| High | No executable/versioned Mongo index/retention migration | Slow queries and unbounded storage drift | Add idempotent migration/index runner and rollback policy |
| High | No backend sanitized `.env.example` or proven config loader | Onboarding and secret validation are unreliable | Commit placeholders, startup validation, secret-manager docs |
| Medium | Generic API contract uses opaque `device` object | Privacy/compatibility behavior is undefined | Type and version admin-safe device DTO |
| Medium | `IdPathRequest` defines `id` while route uses `session_id` | Gateway binding may fail | Align path parameter schema and mapping tests |
| Medium | No Docker/Compose/Kubernetes manifests or health probes | Runtime is not reproducible | Add after service boundary/runtime is implemented |
| Medium | No selected broker despite architecture suggestion | Scalable downstream processing remains design-only | Select one only when implementation requires it |
| Medium | Existing tests do not exercise real `.env.example` URL or backend contract | Green unit tests can mask integration defects | Add config/contract/end-to-end tests |
| Low | Task guide suggests unused packages | Beginner may churn lockfile | Follow manifest/imports, not generic snippets |
| Low | Production build emits chunk-size warning | Slower first load possible | Lazy-load route modules and remeasure |

Current local backend `.env` files are correctly ignored by `.gitignore`, and only the frontend `.env.example` is tracked. This is safer than committing secrets, but backend sanitized examples are still required.

## 14. References to Previous Dependency Files

All earlier task dependency files in this service folder were reviewed before this incremental guide was created.

| Previous dependency file | Section/topic reused | Why reused |
|---|---|---|
| `task1_Dependency.md` | Tech stack, software, pnpm, clone, frontend env/Docker/startup, generic errors | Same React application and toolchain |
| `task2_Dependency.md` | Session Mongo/Redis flow, Gateway/Auth, URL mismatch, ports, privacy | Task 2 already introduced recent-session/journey infrastructure |
| `task3_Dependency.md` | External-service decision rule and secret handling | Same rule: do not invent unimplemented brokers/services |
| `task4_Dependency.md` | Shared missing-runtime, gRPC, CORS, health, and contract-first practices | Same platform foundation |
| `task5_Dependency.md` | Incremental setup pattern, ignored frontend-only package suggestions, operational security | Same frontend and unresolved backend boundary |
| `Dependency/Frontend.md` | Actual packages and commands | Task 6 adds no npm package |
| `Dependency/Environment.md` | Vite/backend variable ownership | Same frontend env and secret-placement rules |
| `Dependency/API_Gateway.md` | Session routes, auth, Redis, and Gateway responsibility | Same public REST entry point |
| `Dependency/MongoDB.md` | Install, Docker example, start, ping, and generic errors | Same `session_db`; Task 6 adds exact index/retention details |
| `Dependency/Redis.md` | Install, Docker example, start, variables, and `PONG` | Same server; Task 6 adds DB 2/prefix diagnostics |
| `Dependency/gRPC.md` | Intended internal communication and missing implementation | Gateway expects Session target `50060` |
| `Dependency/Protobuf.md` | Generation/versioning gap | Session proto/generated code is also absent |
| `Dependency/Go_Modules.md` | Missing backend module/workspace setup | Session backend cannot currently build |
| `Dependency/Migrations.md` | Versioned migration principles | Task 6 needs a Mongo index/retention runner, not MySQL migration |
| `Dependency/main_dependency.md` | Overall topology, setup order, ports, and missing-runtime audit | Same service-level foundation |

## 15. Final Checklist

### Frontend-only readiness

- [ ] Previous dependency documentation checked
- [ ] Node.js 22.12+ and pnpm 10 available
- [ ] Workspace cloned and installed with `pnpm install --frozen-lockfile`
- [ ] No unnecessary `clsx`, MSW, jest-dom, broker, or duplicate package added
- [ ] Existing frontend `.env` created without secrets
- [ ] Shared API base-path defect understood before real integration
- [ ] Task 6 session tests pass
- [ ] Frontend typecheck passes
- [ ] Full frontend test suite passes
- [ ] Production build passes; chunk warning recorded separately
- [ ] Vite starts and `/admin/sessions` route is registered
- [ ] Allowed/blocked frontend roles verified
- [ ] Backend absence is not mistaken for dependency-install failure

### Full-stack readiness

- [ ] `/api/v1` convention fixed for login and all Session APIs
- [ ] `limit` versus `page_size` and pagination response aligned
- [ ] Runnable Auth, API Gateway, and Session Service supplied
- [ ] Session gRPC listener/proto/generated server available on approved address
- [ ] Sanitized backend `.env.example` tracked; real secrets injected securely
- [ ] MongoDB `session_db` reachable with least-privilege auth/TLS as required
- [ ] Session and event indexes plus 90-day TTL verified through a migration runner
- [ ] Redis DB 0 and DB 2 access/readiness verified with service-specific ACL/prefix policy
- [ ] Ingestion creates privacy-safe synthetic sessions/events
- [ ] Retention dry run reviewed before deletion worker activation
- [ ] Backend admin response maps device/geo fields explicitly
- [ ] Risk, suspicious count, freshness, aggregate, status, and pagination contract approved
- [ ] Status/risk filtering is complete across the dataset, not one page only
- [ ] Backend RBAC matches approved `superadmin`/operations/read-only policy
- [ ] Finance/catalog direct API access is server-side denied
- [ ] JWT/JWKS, exact CORS origin, headers, TLS, and request IDs verified
- [ ] Live metrics refresh and Redis cache freshness verified
- [ ] Session list/date/user filters and pagination verified
- [ ] High-risk/suspicious fixtures produce expected explainable output
- [ ] Journey events are ordered, bounded, and privacy-filtered
- [ ] Raw IP, secrets, credentials, PII, form text, tokens, and card data are absent
- [ ] Mongo/Redis/gRPC/internal HTTP are private and have liveness/readiness probes
- [ ] Logs/metrics/alerts checked without leaking identifiers or secret values
- [ ] Backup/restore, TTL, privacy deletion, and incident procedures tested
- [ ] No duplicate setup documentation added

Frontend Task 6 setup is independently verified. Complete Session Oversight readiness tabhi claim karein jab missing runtimes, gRPC/config, URL/pagination contracts, Mongo indexes/retention, Redis counters, granular RBAC, and privacy-safe response projections end to end pass ho.
