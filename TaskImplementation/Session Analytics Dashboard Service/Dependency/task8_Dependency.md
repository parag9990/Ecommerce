# Project Dependency & Setup Guide

> **Scope:** Privacy controls (`/privacy`) ke liye dependency, environment, database, local run, verification, security, aur DevOps guide.
>
> **Important:** Ye file incremental hai. Common clone, Node.js, Go, MongoDB, Redis, Docker, aur generic troubleshooting steps pehle ke dependency guides me already documented hain; unko yahan duplicate nahi kiya gaya.

## 1. Project Overview

Task 8 ab sirf design document nahi hai. Repository me frontend privacy page aur Go backend dono implemented hain:

```text
frontend/session-analytics-dashboard/src/features/privacy/
backend/services/session-service/
```

Implemented runtime flow:

```text
Admin browser
  -> Vite / production frontend
  -> API Gateway + verified admin authentication (required for browser flow)
  -> Session Service HTTP :8086
  -> MongoDB session_db
  -> Redis DB 2
```

Task-specific capabilities:

- masking policy read/update;
- retention policy read/update;
- deletion impact preview;
- user, anonymous ID, or session ID based deletion;
- deletion status list;
- MongoDB audit events and deletion-request records;
- Redis active-session cleanup.

> **Reality check:** Frontend and backend compile-time contracts match, but runnable Gateway/admin-login code repository me absent hai. Isliye protected `/privacy` browser flow end-to-end abhi complete nahi hai. Backend ko controlled local `curl` headers ke saath verify kiya ja sakta hai.

## 2. Tech Stack

### Reused frontend stack

React, TypeScript, Vite, Tailwind CSS, TanStack Query, React Router, Lucide, Vitest, Testing Library, jsdom, aur MSW ka beginner explanation already available hai.

Refer:
`task1_Dependency.md`

Section:
`2. Tech Stack -> Frontend technologies`

Task 8 in existing packages ko settings forms, mutations, cache invalidation, route rendering, icons, aur tests ke liye use karta hai. **Koi new npm package add nahi hua.**

### Task-specific backend and infrastructure

| Technology | Simple Hinglish explanation | Task 8 use | Required? |
|---|---|---|---|
| Go `net/http` | Go standard library ka HTTP server hai. | Four privacy paths aur `/healthz` serve karta hai. | Backend ke liye yes |
| MongoDB Go Driver v2 | Go application ko MongoDB se connect karta hai. | Settings, deletion log, audit, sessions, events, aur journey summaries access karta hai. | Yes |
| go-redis v9 | Go ka Redis client hai. | Active session preview/delete karta hai. | Yes |
| MongoDB | Document database hai. | `privacy_settings`, deletion requests, audit events, aur analytics data store karta hai. | Yes |
| Redis | Fast in-memory data store hai. | Active session keys lookup aur cleanup ke liye use hota hai. | Yes |
| API Gateway/admin auth | Browser identity verify karke trusted actor context backend ko deta hai. | Sensitive read/update/delete authorization boundary hai. | Real browser flow ke liye yes |
| `crypto/hmac` + SHA-256 | Standard Go cryptography hai. | Raw deletion target ko store kiye bina peppered hash banata hai. | Built into Go |
| Docker | Local infrastructure ko isolated containers me chalata hai. | MongoDB/Redis ka recommended local option hai. | Optional |

Kafka, RabbitMQ, NATS, S3, MinIO, SMTP, Elasticsearch, MySQL, aur Kubernetes current privacy implementation me use nahi hote. Deletion request current code me synchronous execute hoti hai; koi queue/worker dependency nahi hai.

## 3. Required Software

Common OS installation aur version commands:

Refer:
`task1_Dependency.md`

Section:
`3. Required Software`

Task 8-specific minimum checklist:

| Software | Current project expectation | Why |
|---|---:|---|
| Go | `1.26.3` | `go.mod` aur `backend/go.work` me pinned |
| Node.js | 22 LTS | Dashboard development/tests |
| MongoDB | Reachable on `27017` by default | Backend startup and privacy persistence |
| Redis | Reachable on `6379` by default | Backend startup and active-session cleanup |
| `mongosh` | Optional locally, recommended in CI/CD | Versioned migration run/inspect karne ke liye |
| OpenSSL | Optional | Strong privacy hash pepper generate karne ke liye |

## 4. Dependency Management

### Frontend: no new install

`package.json` me Task 8 ke required packages already declared hain. Implementation manual validation use karta hai; `task8.md` me suggested **Zod installed nahi hai aur current code ko required nahi hai**. Sirf guide ka old install block dekhkar `npm install zod` mat run karo.

Fresh clone dependency setup:

Refer:
`task1_Dependency.md`

Section:
`4. Dependency Management -> Node.js/npm system`

Task-specific package check:

```bash
cd frontend/session-analytics-dashboard
npm ls @tanstack/react-query react react-router-dom lucide-react vitest jsdom
```

Repository me frontend lockfile abhi missing hai. Reviewed lockfile commit hone tak `npm ci` reliable onboarding command nahi hai.

### Backend Go modules

Direct non-standard modules:

```text
github.com/redis/go-redis/v9 v9.20.0
go.mongodb.org/mongo-driver/v2 v2.6.0
```

Current Mongo import v2 hai. `task8.md` ka old `go get go.mongodb.org/mongo-driver/mongo` example wrong major path use karta hai; us command ko run mat karo.

Go module concepts, download/build/tidy, proxy/cache, version mismatch, aur broken workspace explanation:

Refer:
`task1_Dependency.md`

Section:
`4. Dependency Management -> Go modules`

Task-specific verification:

```bash
cd backend/services/session-service
GOWORK=off go mod download
GOWORK=off go test ./...
GOWORK=off go vet ./...
```

`go mod tidy` onboarding ke liye required nahi hai. Imports deliberately change karne par hi run karo aur `go.mod`/`go.sum` diff review karo.

## 5. Database Setup

### Reused MongoDB installation

MongoDB installation, Docker container, named volume, credentials, URI format, port `27017`, database `session_db`, aur health checks unchanged hain.

Refer:
`task1_Dependency.md`

Section:
`5. Database Setup -> MongoDB`

MongoDB backend ke liye mandatory hai. Server startup par connect/ping fail hua to process exit ho jata hai.

### New privacy collections

| Collection | Purpose | Created when |
|---|---|---|
| `privacy_settings` | Global masking and retention policy | First successful update/upsert |
| `analytics_deletion_requests` | Masked/hashed deletion request status | First deletion request or migration/index creation |
| `admin_audit_events` | Masking, retention, and deletion audit | First mutation or migration/index creation |
| `sessions` | Existing session metadata; deletion anonymizes matching rows | Existing analytics ingestion required |
| `session_events` | Existing raw events; deletion removes matching rows | Existing analytics ingestion required |
| `journey_summaries` | Existing journey summaries; deletion removes matching rows | Existing aggregation required |

Empty `privacy_settings` collection se GET fail nahi hota. Backend in-memory safe defaults return karta hai, but document first successful update tak persist nahi hota.

### New migration

Up migration:

```text
backend/services/session-service/migrations/001_privacy_controls.mongodb.js
```

MongoDB running hone ke baad:

```bash
cd backend/services/session-service
mongosh mongodb://localhost:27017/session_db migrations/001_privacy_controls.mongodb.js
```

Verify indexes:

```bash
mongosh mongodb://localhost:27017/session_db --quiet --eval \
  'printjson(db.analytics_deletion_requests.getIndexes()); printjson(db.admin_audit_events.getIndexes())'
```

Expected named indexes:

```text
analytics_deletion_requests_status_created_at
analytics_deletion_requests_requested_by_created_at
analytics_deletion_requests_target_hash_created_at
analytics_deletion_requests_created_at_ttl
admin_audit_actor_created_at
admin_audit_action_created_at
```

Server startup `EnsureIndexes` bhi same indexes create karta hai. Local development me migration miss hone par startup is gap ko cover kar sakta hai, but production me reviewed migration ko explicit CI/CD step rakho.

Rollback sirf deliberately index removal ke liye:

```bash
mongosh mongodb://localhost:27017/session_db migrations/001_privacy_controls.down.mongodb.js
```

Rollback data collections/documents delete nahi karta.

> **Retention warning:** Deletion-request TTL index hard-coded `730` days hai. UI me `deletionRequestLogDays` change karne se existing TTL index automatically update nahi hota.

## 6. Redis / Queue / External Services

### Reused Redis setup

Redis installation, container, volume, password, port `6379`, logical DB `2`, aur `PING` verification:

Refer:
`task1_Dependency.md`

Section:
`5. Database Setup -> Redis`

Task 8 `SESSION_REDIS_KEY_PREFIX` ke under these active-session patterns scan karta hai:

```text
<prefix>:active_session:*
<prefix>:active:*
```

Session-ID deletion additionally `<prefix>:session:<id>` check karti hai. User/anonymous deletion ke liye matching Redis hashes me `user_id` ya `anonymous_id` field hona chahiye. Ingestion naming alag hui to cleanup zero matches report karega.

> `SCAN` use hua hai, blocking `KEYS` nahi; phir bhi large keyspace me per-key `HGET` expensive ho sakta hai. Production scale par reverse indexes ya background job consider karo.

### Gateway and authentication

Shared missing Gateway/admin-auth setup:

Refer:
`task2_Dependency.md`

Section:
`6. Redis / Queue / External Services -> API Gateway and admin auth`

Current backend internally these headers read karta hai:

```text
X-Admin-ID
X-Admin-Roles
X-Request-ID
```

Alternative actor header names bhi accepted hain. Ye **public authentication mechanism nahi** hain. Gateway ko browser cookie/JWT verify karna, client-supplied actor headers strip karna, aur verified internal headers overwrite karna mandatory hai. Session Service ko private network par rakho.

Role matrix:

| Capability | Allowed roles |
|---|---|
| View settings/retention/deletion list | `admin`, `operations_admin`, `superadmin` |
| Update masking | `operations_admin`, `superadmin` |
| Preview/create deletion | `operations_admin`, `superadmin` |
| Update retention | `superadmin` |

## 7. Environment Variables

### Frontend

Task 8 koi new frontend variable introduce nahi karta. `.env.local` location, existing three variables, Vite loading/restart, CORS, aur browser-secret warning:

Refer:
`task1_Dependency.md`

Section:
`7. Environment Variables -> Frontend .env.local`

Normal intended value remains:

```env
VITE_API_PROXY_TARGET=http://localhost:8080
```

`http://localhost:8086` direct target karne se privacy routes mil jayenge, but browser requests verified actor headers nahi bhejte; result normally `401` hoga. Isko auth workaround mat samjho.

### Backend privacy variables

Complete backend `.env` example and all active HTTP/Mongo/Redis/privacy values already documented hain.

Refer:
`task1_Dependency.md`

Section:
`7. Environment Variables -> Current Session Service environment`

Task 8 ke do privacy-specific values:

| Variable | Requirement | Task-specific note |
|---|---|---|
| `SESSION_PRIVACY_HASH_PEPPER` | **Required** | Empty hua to startup fail; deletion target HMAC ke liye long, random, stable secret use karo |
| `SESSION_PRIVACY_DELETION_LIST_LIMIT` | Optional, default `50` | Valid range `1..200`; recent request list cap karta hai |

Generate a local pepper without printing it into shell history manually:

```bash
openssl rand -hex 32
```

Value ignored `backend/services/session-service/.env` ya secret manager me rakho. Pepper rotate karne se future target hashes old hashes se comparable nahi rahenge; rotation plan ke bina production value change mat karo.

The Go process `.env` automatically load nahi karta. Bash/zsh start se pehle:

```bash
cd backend/services/session-service
set -a
source .env
set +a
GOWORK=off go run ./cmd/server
```

Existing `.env` me many future names hain jo current `config.go` read nahi karta. File me value present hona feature activation ka proof nahi hai.

## 8. Docker Setup

Task 8 koi new Dockerfile, Compose file, application image, network, volume, health check, ya restart policy add nahi karta.

MongoDB + Redis local containers:

Refer:
`task1_Dependency.md`

Section:
`8. Docker Setup`

That setup infrastructure only run karta hai; Session Service aur Gateway containerize nahi hote.

Port already occupied ho to same container name reuse ya unrelated container delete blindly mat karo. Isolated alternatives:

```bash
# Mongo host port 27018 -> container port 27017
docker run -d --name ecommerce-session-mongo-alt \
  -p 127.0.0.1:27018:27017 mongo:8

# Redis host port 6380 -> container port 6379
docker run -d --name ecommerce-session-redis-alt \
  -p 127.0.0.1:6380:6379 redis:7-alpine redis-server --appendonly yes
```

Then backend values change karo:

```env
SESSION_MONGO_URI=mongodb://localhost:27018
SESSION_REDIS_ADDR=localhost:6380
```

Production additions still needed: non-root application image, private network, secret injection, resource limits, `/readyz`, dependency health, and reviewed deployment manifests.

## 9. Local Development Setup

### Step 1: Read shared onboarding

Follow first:

```text
task1_Dependency.md -> 9. Local Development Setup
task2_Dependency.md -> 6. Redis / Queue / External Services
```

### Step 2: Start MongoDB and Redis

Use native services or referenced Docker commands. Confirm both are reachable before server start.

### Step 3: Run the privacy migration

```bash
cd backend/services/session-service
mongosh mongodb://localhost:27017/session_db migrations/001_privacy_controls.mongodb.js
```

If alternate Mongo host port use kiya hai, URI accordingly update karo.

### Step 4: Load backend environment

Confirm at minimum Mongo, Redis, and strong `SESSION_PRIVACY_HASH_PEPPER` configured hain. Local bind ko safer rakho:

```env
SESSION_HTTP_ADDR=127.0.0.1:8086
```

Then source `.env` as Section 7 shows.

### Step 5: Verify backend before running it

```bash
GOWORK=off go test ./...
GOWORK=off go vet ./...
```

### Step 6: Start Session Service

```bash
GOWORK=off go run ./cmd/server
```

Expected log contains `session.http.starting`. Mongo/Redis connection or required pepper failure process ko stop karega.

### Step 7: Prepare frontend

```bash
cd ../../../frontend/session-analytics-dashboard
npm install
npm run dev
```

Open `http://localhost:5174/privacy` only after Gateway/admin auth available ho. Gateway absent hone par frontend unit/component tests use karo; direct UI-to-`:8086` is not a secure working login flow.

## 10. Running the Project

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Vite dev server | `5174` | `/privacy` UI and `/api` proxy | Reused |
| Vite preview | `4174` | Production bundle preview | Reused |
| Intended API Gateway | `8080` | Verified admin entry point | Reused; executable missing |
| Session Service HTTP | `8086` | Privacy APIs and `/healthz` | Implemented |
| Intended Session gRPC | `50060` | Architecture/catalog transport | Reused design; not implemented here |
| MongoDB | `27017` | Settings, requests, audit, analytics data | Reused |
| Redis | `6379` | Active sessions | Reused |

Generic port conflict, firewall, CORS, cookie, and Docker networking details:

Refer:
`task1_Dependency.md`

Section:
`10. Running the Project -> Ports and networking`

### Backend health and safe read verification

```bash
curl -i http://127.0.0.1:8086/healthz
```

Expected: HTTP `200` and an envelope containing `"status":"ok"`.

Controlled local read with a test actor:

```bash
curl -sS http://127.0.0.1:8086/api/v1/analytics/privacy/settings \
  -H 'X-Admin-ID: local-superadmin' \
  -H 'X-Admin-Roles: superadmin' \
  -H 'X-Request-ID: req-local-privacy-read'
```

Safe deletion preview, which does not mutate data:

```bash
curl -sS -X POST \
  http://127.0.0.1:8086/api/v1/analytics/privacy/deletion-preview \
  -H 'Content-Type: application/json' \
  -H 'X-Admin-ID: local-operations-admin' \
  -H 'X-Admin-Roles: operations_admin' \
  -d '{"targetType":"user_id","targetValue":"nonexistent_local_user"}'
```

Expected counts zero ho sakte hain. Real deletion POST destructive hai; disposable seed database aur approved test identity ke bina run mat karo.

### Test status observed during this audit

```text
Backend: GOWORK=off go test ./...  -> pass
Backend: GOWORK=off go vet ./...   -> pass
Frontend focused run               -> 29 pass, 2 fail
Frontend typecheck                 -> fail
```

Privacy helpers and API privacy tests pass hue. Privacy page test async deletion-list heading ko too early assert karta hai. Dusra focused failure Task 7 Blob test ka hai. Typecheck me shared test typings, Task 7 mutation signatures, aur `Error` cause target mismatch included hain; Task 8 ko production-ready mark karne se pehle full package gate green hona chahiye.

## 11. Common Errors & Fixes

Generic npm, Go module, Docker, MongoDB, Redis, port, and CORS issues:

Refer:
`task1_Dependency.md`

Section:
`11. Common Errors & Fixes`

Task-specific failures:

| Symptom | Cause | Fix / prevention |
|---|---|---|
| Startup: `SESSION_PRIVACY_HASH_PEPPER is required` | Required pepper empty or `.env` not sourced | Strong value set karo, shell me export verify karo, process restart karo |
| Startup: Mongo connect/ping failed | Wrong URI, Mongo down, auth/TLS mismatch | URI/credentials verify, ping command run, container logs inspect |
| Startup: Redis ping failed | Redis down, wrong port/password/DB connectivity | `redis-cli PING` or container health check, config correct karo |
| `/privacy` shows unavailable | Gateway absent, proxy wrong, or one of three initial queries failed | Browser Network tab me settings, retention, deletion-list requests inspect karo |
| Direct browser proxy to `8086` returns `401` | Browser cookie ko backend actor headers me convert karne wala Gateway missing | Gateway/admin auth implement/run karo; public client headers trust mat karo |
| API returns `403` | Actor role capability ke liye insufficient | Verified role mapping check karo; UI role flag security boundary nahi hai |
| Migration says index options conflict | Same name/key different TTL/options ke saath exists | Existing index inspect; reviewed migration se reconcile karo, blind drop mat karo |
| Deletion preview finds Mongo rows but zero Redis sessions | Redis key prefix/hash fields ingestion contract se mismatch | `SESSION_REDIS_KEY_PREFIX` and active-session key schema verify karo |
| Retention setting saves but old data remains | Current code policy store karta hai, cleanup/TTL apply nahi karta | Retention worker/TTL reconciler implement karo; setting ko enforcement proof mat samjho |
| Deletion request returns `500` after partial work | Mongo deletes, session anonymization, Redis delete transactional unit nahi hain | Failed request inspect, idempotent retry/reconciliation design implement karo |
| Page test cannot find “Recent deletion requests” | Independent deletion list query still pending | Test me heading ko `findByRole`/`findByText` se await karo |
| `npm` reports WSL 1 unsupported | Linux Node ke saath Windows npm resolve ho raha hai | WSL 2 use karo; `command -v node npm` same environment confirm karo |

## 12. Security & Best Practices

- Session Service ko internet/LAN par directly expose mat karo. Local me `127.0.0.1:8086`, deployment me private network use karo.
- Gateway client-supplied `X-Admin-*`, `X-Actor-*`, `X-User-*`, aur role headers strip karke verified values overwrite kare.
- Mutation endpoints par CSRF protection, rate limiting, re-authentication for destructive actions, aur tenant/store scoping enforce karo.
- Pepper secret manager me rakho, logs/frontend/committed files me nahi. Backup and rotation procedure define karo.
- Raw deletion target request body ya access logs me record mat karo. Current persisted record hash + masked display value use karta hai.
- Deletion ko idempotent, resumable background job banana production scale ke liye safer hai; queue add karne par retry/DLQ and exactly-once business effect document karo.
- Mongo/Redis TLS and authentication production me mandatory rakho. Database ports localhost/private network only expose karo.
- Privacy policy backend response shaping/ingestion par enforce karo. Frontend hiding alone data protection nahi hai.
- Audit event, request ID, actor, action, status, counts, and duration log karo; raw identifiers, cookies, tokens, payload content, and infrastructure secrets avoid karo.
- Deletion and retention behavior backups, replicas, exports, object storage, logs, and downstream aggregates tak define karo; primary database cleanup alone full erasure claim nahi hai.
- Legal/security approval ke bina “GDPR/DPDP compliant” claim mat karo.

## 13. Missing or Misconfigured Things

| Severity | Finding | Professional fix |
|---|---|---|
| Blocker | Runnable Gateway/admin-login integration absent; frontend credentials cookie backend actor headers nahi banati | Gateway route, session/JWT validation, header sanitization, CSRF, RBAC, and integration tests implement karo |
| Critical if exposed | Backend caller-controlled identity/role headers trust karta hai and default `:8086` all interfaces bind karta hai | Private-only service, verified proxy identity, mTLS/network policy, and safer local bind use karo |
| Critical | Deletion queries tenant/store scope ke bina global IDs match karti hain | Server-derived tenant scope every Mongo filter, Redis key, request record, and audit event me add karo |
| High | Retention UI values persistence-only hain; raw/journey/heatmap/aggregate TTL and Redis TTL dynamically enforce nahi hote | Versioned TTL reconciler/worker, dry-run, metrics, rollback, and policy-to-index tests add karo |
| High | Deletion log TTL always 730 days; editable `deletionRequestLogDays` ignored by index | Validated `collMod`/index reconciliation or cleanup worker implement karo |
| High | Deletion synchronously POST ke andar multi-collection delete + Redis cleanup karti hai; cross-store transaction/idempotency absent | Durable job/outbox, resumable steps, idempotency key, reconciliation, and bounded batches add karo |
| High | Masking policy Tasks 1-7 pages/API responses par dynamically applied nahi hai | Central backend projection policy and shared frontend policy context integrate/test karo |
| High | Heatmap points and other identity-bearing downstream stores deletion path me explicitly handled nahi hain | Data inventory complete karo; every linked store/backups/downstream deletion or irreversible anonymization define karo |
| High | No tenant-aware uniqueness/scope in `privacy_settings` (`_id: global`) | Per-tenant/store policy key and authorization model add karo if platform multi-tenant hai |
| Medium | Audit write settings persistence se pehle hota hai; later write fail ho to misleading “updated” audit exists | Atomic Mongo transaction or outcome-aware audit event design use karo |
| Medium | Failed deletion stores raw internal error and API list returns it to all viewing admins | Sanitized error code/UI message store karo; detailed cause restricted logs me rakho |
| Medium | `/healthz` live process check hai; runtime Mongo/Redis readiness continuously verify nahi karta | Separate `/livez` and dependency-aware `/readyz` add karo |
| Medium | Migration and startup both own indexes; policy ownership/TTL reconciliation unclear | Migration-first ownership define karo; startup only validate or idempotently reconcile documented policy |
| Medium | Full frontend test/typecheck/build gate red hai | Async privacy test fix, shared TypeScript/test issues fix, CI gates enforce karo |
| Medium | No frontend lockfile, application Dockerfile, checked-in Compose, or deployment manifest | Reviewed lockfile and reproducible non-root build/deploy assets add karo |
| Operational | Metrics named in `task8.md` implemented nahi hain | Prometheus/OpenTelemetry counters, durations, failures, queue lag, and alerts add karo |

Full setup complete tab maana jayega jab verified admin browser Gateway se page load kare, role matrix enforce ho, policy changes actually all read paths/TTL workers par apply hon, scoped deletion every linked store me safely complete ho, audit/metrics visible hon, and backend/frontend/integration gates pass hon.

## 14. References to Previous Dependency Files

| Previous dependency file | Section / topic reused | Why reused |
|---|---|---|
| `task1_Dependency.md` | Tech stack and required software | Same React/Node/Go toolchain |
| `task1_Dependency.md` | Node/npm and Go modules | Same package/module management; no new frontend package |
| `task1_Dependency.md` | MongoDB and Redis installation | Same instances, credentials, volumes, ports, and health checks |
| `task1_Dependency.md` | Environment and Docker | Common variables and infrastructure containers already complete |
| `task1_Dependency.md` | Local run, ports, generic errors, security | Shared onboarding behavior unchanged |
| `task2_Dependency.md` | Gateway and admin authentication | Same protected browser integration boundary |
| `task3_Dependency.md` | Session events, identity safety, ordering, retention | Deletion reuses session/journey identity contract |
| `task4_Dependency.md` | Aggregate analytics and small-count privacy | Funnel aggregates must stay identity-safe |
| `task5_Dependency.md` | Heatmap storage and privacy | Deletion/data inventory must include heatmap implications |
| `task6_Dependency.md` | Retention read model | Privacy retention settings govern the same data classes |
| `task7_Dependency.md` | Export security and scheduled delivery | Exports must respect masking/deletion/retention policy |

All references are relative to this file's folder.

## 15. Final Checklist

### Common setup

- [ ] Previous dependency guides read; repeated installation not copied
- [ ] Go `1.26.3` and Node 22 available
- [ ] MongoDB reachable with intended database/credentials
- [ ] Redis reachable with intended password, DB `2`, and key prefix
- [ ] Conflicting shared ports identified before starting new containers

### Task 8 setup

- [ ] `SESSION_PRIVACY_HASH_PEPPER` generated, stored securely, and sourced
- [ ] Deletion list limit is within `1..200`
- [ ] Privacy migration applied and named indexes verified
- [ ] `GOWORK=off go test ./...` passes
- [ ] `GOWORK=off go vet ./...` passes
- [ ] Session Service starts on private/local `8086`
- [ ] `/healthz`, settings GET, and safe deletion preview verified
- [ ] Destructive tests use disposable scoped data only

### Integration and security

- [ ] Gateway verifies admin session/JWT and strips spoofed actor headers
- [ ] CSRF, RBAC, rate limit, re-authentication, and tenant/store scope enforced
- [ ] `/privacy` loads through Gateway rather than directly trusting browser headers
- [ ] Masking policy applied across every analytics response/page/export
- [ ] Retention settings connected to actual TTL/cleanup enforcement
- [ ] Deletion covers all linked stores, backups, exports, and downstream data policy
- [ ] Audit logs and operational metrics contain no raw sensitive target
- [ ] Frontend tests, typecheck, build, backend integration, and end-to-end tests green
- [ ] No duplicate setup documentation or original task-file modification added
