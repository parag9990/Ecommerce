# Project Dependency & Setup Guide

> **Scope:** Ye incremental handbook Journey Explorer ko locally run, test, aur future backend ke saath integrate karne ke liye hai. Common clone, Node.js, Go, MongoDB, Redis, Docker, environment, aur networking setup repeat nahi kiya gaya; pehle `task1_Dependency.md` aur phir `task2_Dependency.md` follow karo.
>
> **Repository reality check:** Journey Explorer frontend, routes, typed API client, response parser, filters, privacy sanitization, and tests implemented hain. Real `GET /api/v1/analytics/sessions/{session_id}/journey` backend route, runnable Gateway, admin-auth flow, `GetJourney` transport/use case, event ingestion, and analytics collection migration abhi implemented nahi hain. Frontend tests/build run ho sakte hain, but real session journey end to end abhi nahi chalegi.

## 1. Project Overview

Task 3 existing dashboard me selected session ki ordered event journey dikhata hai:

- routes: `http://localhost:5174/journey` and `http://localhost:5174/journey/{sessionId}`;
- request: `GET /api/v1/analytics/sessions/{session_id}/journey`;
- session id search and Active Sessions se deep link;
- session context, event filters, timeline, and selected-event details;
- page view, product view, search, click, scroll, cart, checkout, and payment events;
- one retry and `20`-second React Query stale time; automatic polling nahi hai;
- browser cookies through `credentials: "include"`;
- event property and visible-path sanitization before UI display.

### Intended runtime flow

```text
Browser /journey/{sessionId} (Vite :5174)
  -> GET /api/v1/analytics/sessions/{sessionId}/journey
  -> Vite /api proxy
  -> intended API Gateway (:8080, currently missing)
  -> intended admin authorization (currently missing)
  -> intended SessionService.GetJourney (currently missing)
  -> MongoDB sessions + ordered session_events
```

### What can run today?

| Mode | Result | Current status |
|---|---|---|
| Focused frontend tests | Mocked journey response se parser, page, timeline, filtering, and masking verify hote hain | Available after dependency install |
| Frontend dev/build | `/journey` renders; session id request backend absent hone par error state dikhayegi | Available |
| Current Session Service | Privacy routes and `/healthz` run hote hain | Available with reused MongoDB, Redis, and backend env |
| Real journey data | Gateway, admin auth, ingestion, collections/indexes, and `GetJourney` required | **Blocked** |

## 2. Tech Stack

### Reused technologies

React, TypeScript, Vite, Tailwind CSS, TanStack Query, React Router, Lucide, Vitest, Testing Library, MSW, Go, MongoDB, Redis, and Docker already explained hain.

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Sections:
`2. Tech Stack`, `3. Required Software`, and `4. Dependency Management`

### Task 3-specific usage

| Technology | Is task me role | Required? | New installation? |
|---|---|---|---|
| React Router | `/journey` and `/journey/:sessionId` routes plus Active Sessions deep link | Yes | No |
| TanStack Query | Session-id based cache, cancellation, one retry, and `20s` stale time | Yes | No |
| Browser Fetch + AbortController | Cookie request, request ID, external cancellation, and `10s` default timeout | Yes | No |
| TypeScript response parser | snake_case/camelCase normalize karta aur invalid payload reject karta hai | Yes | No |
| Lucide + Tailwind | Timeline, filters, detail panel, loading/error states | Yes | No |
| Vitest + Testing Library | API client, hook, page, timeline, filter, and privacy tests | Development only | No |
| MongoDB | Future `sessions` and `session_events` persistence/query | Real backend ke liye yes | Setup reused |
| Redis | Current Go service startup and broader live-session architecture | Current service ke liye yes; persisted journey read ke liye direct requirement nahi | Setup reused |

> **Test reality:** Task implementation guide MSW example dikhata hai, and `msw` package manifest me declared hai, but checked-in Task 3 tests direct `fetch` stubs use karte hain. Koi mock server ya external test service start nahi karni.

## 3. Required Software

Task 3 koi new operating-system software introduce nahi karta.

Refer:
`TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md`

Section:
`3. Required Software`

Beginner rule:

- frontend page/tests ke liye Git, Node.js, npm, and browser enough hain;
- current Go Session Service ke liye Go, MongoDB, Redis, and exported backend env chahiye;
- MongoDB GUI, Kafka, RabbitMQ, map SDK, replay recorder, or browser extension Task 3 ke liye required nahi hai;
- real journey flow software install se nahi, missing application implementation complete karne se unblock hoga.

## 4. Dependency Management

### Node.js dependencies

Task 3 ne `frontend/session-analytics-dashboard/package.json` me koi new package add nahi kiya. Router, React Query, Lucide, Testing Library, Vitest, and MSW existing manifest me declared hain. Task guide ke standalone `npm install ...` examples dobara run mat karo.

Use the existing install instructions:
`task1_Dependency.md` -> `4. Dependency Management` and `9. Local Development Setup`

Current repository state:

- `node_modules/` absent hai, so first run se pehle one-time `npm install` required hai;
- frontend lockfile absent hai, so `npm ci` currently work nahi karega;
- Task 3 ke liye no extra runtime SDK, date library, or API client install required hai;
- source uses native `fetch`; Axios required nahi hai.

### Go modules

Journey backend implemented nahi hai, so Task 3 koi new Go dependency introduce nahi karta. Existing MongoDB driver, Redis client, Go version, `GOWORK=off`, proxy, checksum, and cache guidance unchanged hai.

Refer:
`task1_Dependency.md` -> `4. Dependency Management` -> `Go modules`

## 5. Database Setup

### MongoDB

MongoDB ek document database hai. Intended Journey API `sessions` collection se context and `session_events` collection se ordered timeline read karegi. Real backend ke liye MongoDB mandatory hoga; frontend mocked tests ke liye nahi.

Installation, Docker container, port `27017`, credentials, connection string, persistence, and verification unchanged hain.

Refer:
`task1_Dependency.md` -> `5. Database Setup` -> `MongoDB`

### Task 3 data requirements

| Collection / index | Purpose | Repository status |
|---|---|---|
| `sessions` | Session identity, timestamps, device, geo, entry/exit context | Design documented; repository/use case absent |
| `session_events` | Timeline event type, path, properties, and occurrence time | Design documented; repository/use case absent |
| `{ session_id: 1, occurred_at: 1 }` | One session ke events chronological order me efficiently read karna | Design doc only; migration absent |
| `{ occurred_at: 1 }` TTL index | Raw event retention, documented as `7776000` seconds (90 days) | Design doc only; migration absent |
| `journey_summaries` | Optional longer-lived summary data | Architecture mention only; schema/migration absent |

> **Do not run design snippets blindly in production.** First versioned, reviewed, idempotent migration banao. Retention value legal/privacy policy ke saath align honi chahiye.

Existing `001_privacy_controls.mongodb.js` Journey migration nahi hai. Woh `sessions`, `session_events`, journey indexes, or sample events create nahi karta.

### Ordering and retention contract

- API events ko `occurred_at ASC` return kare; frontend array ko sort nahi karta.
- Same timestamp ho to stable tie-breaker such as `_id ASC` define karo.
- TTL deletion immediate guarantee nahi hoti; old session mil sakta hai but raw events empty ho sakte hain.
- Session and event queries tenant/store boundary and admin authorization ke andar run honi chahiye.
- Empty event list valid response ho sakti hai; missing session should return `404`.

### Redis

Redis setup, port `6379`, logical DB `2`, credentials, and health check reused hain.

Refer:
`task1_Dependency.md` -> `5. Database Setup` -> `Redis`

Task 3 frontend Redis ko directly access nahi karta. Future Journey API durable events MongoDB se read kar sakti hai. Current Go process nevertheless startup par Redis connect karta hai, so existing service run karte time Redis reachable hona mandatory hai.

## 6. Redis / Queue / External Services

### API Gateway and admin authentication

Journey endpoint contract `auth: admin` declare karta hai. Intended Gateway ko:

- `GET /api/v1/analytics/sessions/{session_id}/journey` register karna;
- admin session/token authenticate and role/tenant authorize karna;
- safe session id downstream `SessionService.GetJourney` ko pass karna;
- `401`, `403`, `404`, timeout, and structured error envelopes preserve karna;
- request ID propagate karna;
- cookie-based browser calls ke liye same-origin or correct CORS/credential policy use karna hoga.

Current checkout me runnable Gateway and `GetJourney` implementation absent hain.

### Queue or ingestion service

Kafka/RabbitMQ/NATS Task 3 frontend ke mandatory dependencies nahi hain. Real journeys ke liye an event ingestion path zaroor chahiye, but repository me `POST /api/v1/sessions/events`, producer, broker consumer, validation, or writer implementation available nahi hai.

Choose a broker only after throughput, durability, ordering, and retry requirements define ho. Sirf UI run karne ke liye broker install mat karo.

### Third-party services

No payment provider, map API, replay SDK, object storage, SMTP, or GeoIP credential Task 3 add karta hai. Payment events yahan analytics records hain; Journey Explorer Stripe/PayPal ko call nahi karta.

## 7. Environment Variables

### New or changed variables

**None.** Task 3 same frontend variables reuse karta hai:

- `VITE_API_BASE_URL`
- `VITE_API_PROXY_TARGET`
- `VITE_REQUEST_TIMEOUT_MS`

Refer:
`task1_Dependency.md` -> `7. Environment Variables`

Use the same locations:

| File | Purpose | Task 3 action |
|---|---|---|
| `frontend/session-analytics-dashboard/.env.local` | API base, Vite proxy, timeout | Reuse unchanged |
| `frontend/session-analytics-dashboard/.env.example` | Safe tracked template | Already present |
| `backend/services/session-service/.env` | Current Go process environment | Reuse/export unchanged |

Local frontend example remains:

```env
VITE_API_BASE_URL=
VITE_API_PROXY_TARGET=http://localhost:8080
VITE_REQUEST_TIMEOUT_MS=10000
```

No journey session id, admin token, MongoDB credential, event payload, or PII belongs in frontend env. Every `VITE_` value browser bundle me visible hoti hai.

> `20s` journey stale time and retry count source constants hain; inke environment variables exist nahi karte. Go service `.env` automatically load nahi karta; Task 1 ka export flow reuse karo.

## 8. Docker Setup

Task 3 koi new container, image, port, volume, network, health check, or restart policy add nahi karta.

Refer:
`task1_Dependency.md` -> `8. Docker Setup`

Same optional MongoDB + Redis local Compose example reuse karo. Current limitations:

- repository me application Dockerfiles or complete Compose stack absent hai;
- no Gateway, journey worker, or event-ingestion container exists;
- healthy MongoDB/Redis ka matlab Journey API implemented hona nahi hai;
- frontend-only tests ke liye containers start karna unnecessary hai;
- `docker compose down -v` local data delete karta hai, so casually use mat karo.

## 9. Local Development Setup

### Step 1: Read previous setup guides

First follow:

1. `TaskImplementation/Session Analytics Dashboard Service/task1_Dependency.md` for clone, tools, install, env, and common infrastructure.
2. `TaskImplementation/Session Analytics Dashboard Service/task2_Dependency.md` for Active Sessions integration and the Journey deep-link source.

### Step 2: Move to the frontend package

From repository root:

```bash
cd frontend/session-analytics-dashboard
```

No Task 3-only install command is needed. If `node_modules/` is absent, run the reused one-time install from Task 1.

### Step 3: Reuse frontend environment

Ensure `.env.local` has the reused values from Section 7. Dev server already proxies `/api` to `http://localhost:8080` by default.

### Step 4: Run focused verification

```bash
npm run typecheck
npm run test -- \
  src/api/session-api.test.ts \
  src/features/journey/hooks/use-session-journey.test.tsx \
  src/features/journey/pages/journey-explorer-page.test.tsx \
  src/features/journey/components/session-timeline.test.tsx \
  src/features/journey/lib/journey-filter.test.ts \
  src/features/journey/lib/event-privacy.test.ts
npm run build
```

These tests direct `fetch` stubs use karte hain. Real Gateway, MongoDB, Redis, or MSW server required nahi hai.

### Step 5: Start frontend

```bash
npm run dev
```

Open:

```text
http://localhost:5174/journey
```

Valid session id uses maximum `160` characters and only letters, numbers, `.`, `_`, `:`, or `-`. Example:

```text
http://localhost:5174/journey/sess_123
```

Backend absent hone par **Journey unavailable** expected hai. Base `/journey` page ko API request nahi bhejni chahiye.

### Step 6: Optional current backend

MongoDB, Redis, backend env loading, `GOWORK=off go run`, and `/healthz` verification Task 1 jaise hain. Task 3 ke liye no new migration run karni hai because required migration does not yet exist.

### Step 7: Verify the intended API

Final integrated request:

```bash
curl -i \
  'http://localhost:8080/api/v1/analytics/sessions/sess_123/journey'
```

Current checkout me Gateway connection failure expected hai. Current Session Service direct check:

```bash
curl -i \
  'http://localhost:8086/api/v1/analytics/sessions/sess_123/journey'
# Expected current behavior: 404
```

Final Gateway call valid admin authentication require karega. Real cookies/tokens command history, source, or documentation me hardcode mat karo.

## 10. Running the Project

### Daily frontend workflow

General dev, watch-test, build, and preview commands unchanged hain.

Refer:
`task1_Dependency.md` -> `10. Running the Project`

Task 3 manual check:

1. `/journey` open karo and confirm **Select a session** state.
2. `sess_123` submit karke URL encoding and one journey request verify karo.
3. Integrated/mock response me session context and chronological timeline verify karo.
4. Event filters toggle karke visible list verify karo.
5. Event select karke detail panel and recursive masking verify karo.
6. Invalid/missing id, empty events, `404`, timeout, and retry states check karo.
7. Active Sessions row se Journey link correct session route open karta hai verify karo.

### Expected response shape

Frontend snake_case and camelCase accept karta hai. Minimum useful response:

```json
{
  "session": {
    "session_id": "sess_123",
    "started_at": "2026-05-18T00:00:00Z",
    "last_seen_at": "2026-05-18T00:05:00Z",
    "device": { "type": "mobile" }
  },
  "events": [
    {
      "_id": "evt_1",
      "session_id": "sess_123",
      "event_type": "page_view",
      "path": "/products",
      "properties": {},
      "occurred_at": "2026-05-18T00:00:02Z"
    }
  ]
}
```

`summary` optional hai; frontend events se derive kar leta hai. Allowed event types exactly:

```text
page_view, product_view, search, click, scroll,
add_to_cart, checkout_step, payment_result
```

Unknown event type currently `INVALID_SESSION_JOURNEY_RESPONSE` cause karega.

### Ports and networking

No new port introduced hua.

| Service | Port | Task 3 purpose | Status |
|---|---:|---|---|
| Vite dev server | `5174` | Journey UI and `/api` proxy | Reused |
| Vite preview | `4174` | Production bundle preview | Reused |
| Intended API Gateway | `8080` | Admin journey API | Reused, implementation missing |
| Current Session HTTP | `8086` | Health/privacy only | Reused, journey route missing |
| Intended Session gRPC | `50060` | `GetJourney` downstream contract | Reused, implementation missing |
| Auth/JWKS intent | `8081` | Admin identity and role | Reused, integration missing |
| MongoDB | `27017` | Sessions and ordered events | Reused |
| Redis | `6379` | Current service/live-session dependency | Reused |

Port conflict, firewall, CORS, cookie, and Docker-network guidance:
`task1_Dependency.md` -> `10. Running the Project` -> `Ports and networking`

## 11. Common Errors & Fixes

Generic npm, Go, MongoDB, Redis, Docker, port, auth, CORS, timeout, and permission errors:
`task1_Dependency.md` -> `11. Common Errors & Fixes`

Only Journey-specific failures:

| Error / symptom | Cause | Fix / prevention |
|---|---|---|
| `/journey` makes a request | Route supplied an accidental id or hook enable logic changed | Base route must pass no id; verify Network tab and hook test |
| **Journey unavailable** | Gateway absent, proxy wrong, endpoint missing, auth failure, or upstream timeout | Network status inspect karo; current checkout me connection failure/404 expected hai |
| **Session not found** | Gateway returned `404` for unknown/deleted/expired session | Session id and tenant verify karo; do not convert every upstream error to `404` |
| `VALIDATION_ERROR` before network call | Id empty, over 160 chars, or unsupported character such as `/` or space | Supported id format use karo; server-side same or stricter validation add karo |
| `INVALID_SESSION_JOURNEY_RESPONSE` | Required session/timestamp fields missing, bad property type, or unknown event type | Provider-consumer contract test add karo and canonical schema align karo |
| Timeline order wrong | Backend did not sort by `occurred_at ASC`; frontend preserves response order | Compound index use karo and deterministic server sort add karo |
| Session exists but **No journey events** | TTL expired, ingestion absent/failed, filter hides events, or valid zero-event session | Raw event retention and ingestion logs inspect karo; clear filters |
| Summary count unexpected | Backend summary stale/different semantics or frontend derived limited event types | Summary definition version and contract test karo; partial/truncated response metadata add karo |
| Payment marked incomplete | `payment_result.properties.status` expected success aliases se match nahi karta | Canonical payment status enum define karo |
| Active Sessions link fails | List returns malformed/raw id or Journey backend is missing | URL-safe id contract align karo; frontend encoding already enabled hai |
| Sensitive value visible | New property key/pattern sanitizer list me covered nahi, or raw payload Network tab me available hai | Server allowlist/redaction add karo; UI sanitizer ko defense-in-depth rakho |
| Retry causes duplicate load | TanStack Query one retry performs second GET after transient failure | GET idempotent rakho; request ID and upstream logs correlate karo |

## 12. Security & Best Practices

Common secret, cookie, CSRF, TLS, logging, database, container, and dependency guidance:
`task1_Dependency.md` -> `12. Security & Best Practices`

Task 3-specific additions:

- **Authorize every lookup:** Knowing a session id must not grant access. Admin role plus tenant/store scope server-side enforce karo.
- **Redact before response:** UI sanitizer raw response, browser tools, proxy logs, and query cache se secrets remove nahi karta. Backend event-property allowlist preferred hai.
- **Never collect sensitive input:** Password, OTP, payment card, CVV, access token, full address, free-form form values, or raw keystrokes ingestion par reject/redact karo.
- **Sanitize paths:** Query strings may contain email/token/order identifiers. Store/return normalized paths or an explicit safe parameter allowlist.
- **Bound response size:** Maximum events, pagination/cursor, time range, payload bytes, and property depth/length define karo. One huge session browser/backend memory exhaust kar sakta hai.
- **Stable ordering:** `occurred_at` with `_id` tie-breaker use karo; cursor must use same ordering.
- **Retention:** Raw events and journey summaries ke separate, documented TTLs rakho. Deletion request should sessions, events, summaries, cache, and backups policy cover kare.
- **Safe observability:** Log request ID, latency, result count, and error code. Session ids hash/mask karo; event properties, paths with query strings, user IDs, IP, or cookies log mat karo.
- **Response hardening:** JSON content type, no-store/private cache policy for sensitive journey data, CSP, frame restrictions, and TLS apply karo.
- **Contract versioning:** New event types additive rollout se pehle frontend compatibility ensure karo; current strict parser unknown types reject karta hai.

## 13. Missing or Misconfigured Things

| Severity | Finding | Professional fix |
|---|---|---|
| Blocker | Current Session Service does not register Journey route or implement `GetJourney` | Handler/gRPC transport, use case, repositories, validation, and integration tests add karo |
| Blocker | Runnable Gateway and admin authentication flow absent hain | Gateway mapping, admin RBAC, tenant scope, cookies/JWKS, and error propagation implement karo |
| Blocker | Event ingestion/writer absent hai | Validated, privacy-safe event ingestion and durable write path implement karo |
| High | `sessions`/`session_events` migration and local seed path absent hain | Versioned collection/index migration, rollback policy, and safe sample data command add karo |
| High | Master `JourneyResponse` schema is sparse while frontend expects/derives richer session, event, and summary semantics | Canonical contract update karo and provider-consumer contract tests add karo |
| High | Master event schema allows arbitrary `event_type`; frontend accepts only eight values | Shared enum/versioning and unknown-event rollout behavior define karo |
| High | No pagination/truncation metadata; one long journey can return unbounded events | Cursor pagination, max limit, `partial`, and next cursor contract add karo |
| High | Server-side property redaction/allowlist not implemented | Ingestion and response boundaries par recursive size/type validation and redaction add karo |
| High | Documented TTL/indexes are not executable migrations | Reviewed retention config and idempotent migration implement/verify karo |
| Medium | Frontend assumes backend event order and does not validate timestamp format | Server deterministic sort; parser timestamp validation or contract test add karo |
| Medium | Strict parser rejects future event types and may blank the whole journey | Version event enum or safe `unknown` display strategy choose karo |
| Medium | Task guide advertises MSW mock but checked-in tests use direct fetch stubs | One shared testing pattern document/adopt karo |
| Medium | No dependency-aware readiness for analytics data path | `/readyz` should verify required stores and journey repository without leaking secrets |
| Medium | No application Dockerfiles/complete Compose/deployment runbook | Build images, non-root runtime, health checks, secrets, resource limits, and migrations workflow add karo |

Unchanged lockfile, `go.work`, shallow health, and shared deployment gaps repeat nahi kiye gaye.

Refer:
`task1_Dependency.md` -> `13. Missing or Misconfigured Things`

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `task1_Dependency.md` | `2. Tech Stack` | Same React/Vite/TypeScript/Go stack |
| `task1_Dependency.md` | `3. Required Software` | No new OS-level tool |
| `task1_Dependency.md` | `4. Dependency Management` | Same package manifest, npm, Go module, and workspace behavior |
| `task1_Dependency.md` | `5. Database Setup` | Same MongoDB/Redis instances, ports, credentials, and Docker setup |
| `task1_Dependency.md` | `6. Redis / Queue / External Services` | Same missing Gateway/admin-auth platform and no mandatory broker |
| `task1_Dependency.md` | `7. Environment Variables` | No new or changed loaded variable |
| `task1_Dependency.md` | `8. Docker Setup` | No new container, volume, network, or image |
| `task1_Dependency.md` | `9. Local Development Setup` | Clone/install/config/start flow unchanged |
| `task1_Dependency.md` | `10. Running the Project` | Common commands and networking reused |
| `task1_Dependency.md` | `11. Common Errors & Fixes` | Generic tool/infrastructure troubleshooting unchanged |
| `task1_Dependency.md` | `12. Security & Best Practices` | Shared secrets, cookie, database, and deployment rules |
| `task1_Dependency.md` | `13. Missing or Misconfigured Things` | Shared platform gaps already audited |
| `task2_Dependency.md` | Active Sessions flow and Journey link | Same dashboard package; Task 3 is opened from Task 2 rows |
| `task2_Dependency.md` | Gateway, auth, MongoDB, Redis, ports | Integration dependencies remain unchanged |

## 15. Final Checklist

### Frontend verification

- [ ] Previous dependency documentation read first
- [ ] No duplicate Task 3 package installation attempted
- [ ] Existing `.env.local` reused and no secret placed in `VITE_` variables
- [ ] `npm run typecheck` passes
- [ ] Journey API client, hook, page, timeline, filter, and privacy tests pass
- [ ] `npm run build` passes
- [ ] `/journey` shows no-session state without an API request
- [ ] `/journey/{sessionId}` requests the expected encoded path
- [ ] Loading, data, empty, `404`, general error, and retry states checked
- [ ] Timeline ordering, filters, summary, and event details checked
- [ ] Sensitive properties and visible paths checked for masking
- [ ] Active Sessions Journey link checked

### Backend and infrastructure readiness

- [ ] No unrelated broker, replay SDK, map SDK, or payment integration installed
- [ ] Privacy migration not mistaken for a Journey migration
- [ ] MongoDB collections, compound indexes, TTLs, and rollback policy implemented
- [ ] Privacy-safe event ingestion and local sample path implemented
- [ ] `GetJourney` repository, use case, and transport implemented
- [ ] Gateway route and canonical request/response schema aligned
- [ ] Admin RBAC and tenant/store scope enforced server-side
- [ ] Event type compatibility and deterministic ordering tested
- [ ] Response size limits, pagination, retention, and deletion behavior defined
- [ ] Server-side property redaction and safe logging verified
- [ ] Dependency-aware readiness, metrics, traces, and capacity tested
- [ ] No duplicate setup documentation added

---

**Beginner mental model:** Journey Explorer frontend independently test/build ho sakta hai. MongoDB and Redis start karna sirf infrastructure ready karta hai; real timeline tabhi aayegi jab privacy-safe ingestion, required collections/indexes, `GetJourney`, Gateway, and admin authorization sab implemented aur connected hon.
