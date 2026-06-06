# Project Dependency & Setup Guide

## 1. Project Overview

This document is for:

`TaskImplementation/Session Management Service/task4.md`

Output file:

`TaskImplementation/Session Management Service/task4_Dependency.md`

Task 4 ka main focus **Journey Tracking** hai. Simple words me: ek session ke saare stored events ko time order me read karke admin/dashboard ko user journey dikhani hai.

Example:

```text
page_view -> search -> product_view -> add_to_cart -> checkout_step -> payment_result
```

Important boundary:

- Business logic yahan rewrite nahi ki gayi.
- API implementation yahan rewrite nahi ki gayi.
- Ye file sirf dependency, setup, environment, database, Docker, DevOps, and beginner onboarding guide hai.
- Previous dependency docs ko duplicate nahi kiya gaya. Same setup ke liye references diye gaye hain.

Current repository note:

- `task4.md` originally guide-style document hai.
- Current `backend/services/session-service` codebase me journey runtime pieces present hain:
  - `internal/usecase/get_journey.go`
  - `internal/repository/mongo_journey_summary_repository.go`
  - `internal/domain/journey.go`
  - `migrations/002_journey_summaries.up.js`
  - `GET /api/v1/analytics/sessions/{session_id}/journey`

Task 4 setup ka biggest new dependency change:

| Area | Task 4 change |
|---|---|
| Database | MongoDB me new derived collection: `journey_summaries` |
| Migration | New migration: `002_journey_summaries.up.js` |
| Environment | New journey limit/summary env variables |
| External services | No new service beyond existing MongoDB + Redis |
| Ports | No new port |
| Docker | Same MongoDB + Redis setup as previous tasks |

## 2. Tech Stack

Full base tech stack already explain kiya gaya hai:

Refer:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Sections:

- `2. Tech Stack`
- `3. Required Software`
- `4. Dependency Management`
- `5. Database Setup`
- `6. Redis / Queue / External Services`

Task 4 me detected/reused technologies:

| Technology | What it is | Why Task 4 uses it | Required? | Detailed setup |
|---|---|---|---|---|
| Go | Backend programming language | Journey usecase, repository, HTTP handler run karne ke liye | Yes | `task1_Dependency.md`, section `4. Dependency Management` |
| Go standard `net/http` | Go ka built-in HTTP server package | Journey endpoint expose karne ke liye | Yes | `task1_Dependency.md`, section `2. Tech Stack` |
| MongoDB | Document database | `sessions`, `session_events`, and `journey_summaries` store/read karne ke liye | Yes | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| MongoDB Go Driver | Go library for MongoDB | Timeline events query and summary upsert ke liye | Yes | `task1_Dependency.md`, section `4. Dependency Management` |
| Redis | In-memory cache | Journey read ka primary source nahi, but service startup Redis ping karta hai and previous active-session state Redis me hai | Yes for current service startup | `task1_Dependency.md`, section `5.2 Redis Setup` |
| go-redis/v9 | Go Redis client | Redis readiness and active-session repositories ke liye | Yes in service | `task1_Dependency.md`, section `4. Dependency Management` |
| mongosh | MongoDB shell | Task 4 migration and index verification ke liye | Recommended | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| Docker | Container runtime | Local MongoDB/Redis quickly run karne ke liye | Optional but recommended | `task1_Dependency.md`, section `8. Docker Setup` |
| curl | CLI HTTP client | Journey endpoint manually test karne ke liye | Recommended | Usually preinstalled |

Beginner Hinglish explanation:

- Go backend service ka engine hai. Journey endpoint Go code se MongoDB query karta hai.
- MongoDB long-term record book jaisa hai. Raw events and journey summary yahin persist hote hain.
- Redis fast memory cache hai. Task 4 journey raw MongoDB se banti hai, but current app Redis ke bina ready nahi hota because service dependencies Redis ping karti hain.
- Docker optional hai, but beginner ke liye MongoDB and Redis run karna easiest bana deta hai.

Task 4 me Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, Firebase, S3, Kubernetes ka koi required setup introduce nahi hua.

## 3. Required Software

Same base software previous dependency docs me already covered hai. Install steps repeat nahi kiye gaye.

| Software | Required for Task 4? | Purpose | Where setup is explained |
|---|---:|---|---|
| Go `1.26.3` or compatible newer version | Yes | Build/run session service | `task1_Dependency.md`, section `3. Required Software` |
| MongoDB | Yes | Store sessions, events, journey summaries | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| mongosh | Recommended | Run `002_journey_summaries.up.js`, verify indexes | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| Redis | Yes for current service startup | Existing active-session/cache dependency | `task1_Dependency.md`, section `5.2 Redis Setup` |
| redis-cli | Recommended | Verify Redis health | `task1_Dependency.md`, section `5.2 Redis Setup` |
| Docker / Docker Compose | Optional | Local MongoDB/Redis containers | `task1_Dependency.md`, section `8. Docker Setup` |
| curl | Recommended | Manual endpoint testing | OS package manager if missing |
| Git | Yes | Clone repository | `task1_Dependency.md`, section `9. Local Development Setup` |

Task 4-specific extra requirement:

| Requirement | Why |
|---|---|
| Existing Task 1-3 setup completed | Journey reads `sessions` and `session_events`, so earlier storage/event setup must exist |
| Migration `001_session_storage_indexes.up.js` already applied | It creates/validates base `sessions` and `session_events` collections/indexes |
| Migration `002_journey_summaries.up.js` applied | It creates/validates `journey_summaries` collection and indexes |

## 4. Dependency Management

This is a Go project. Full explanation of `go.mod`, `go.sum`, Go modules, `go mod tidy`, `go mod download`, `go build`, and common Go module issues already exists.

Refer:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`4. Dependency Management`

Current module:

`backend/services/session-service/go.mod`

Task 4 uses already-present dependencies:

| Dependency | Version in current `go.mod` | Task 4 purpose |
|---|---:|---|
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | MongoDB timeline query, journey summary upsert, index creation |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Service Redis client and readiness ping |

No new Go package is required only for Task 4 journey tracking.

Quick commands:

```bash
cd backend/services/session-service
go mod download
go mod tidy
go test ./...
go run ./cmd/server
```

Common Go dependency problems like `missing go.sum entry`, old Go version, proxy/cache issues, and version mismatch are already documented here:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`12. Common Errors & Fixes`

Task 4-specific dependency note:

- If `go test ./...` fails because MongoDB/Redis are not running, first start MongoDB and Redis.
- If config validation fails before tests/run, check Task 4 journey env variables in section `7. Environment Variables`.

## 5. Database Setup

### 5.1 Database Summary

| Database | Required? | Task 4 usage | Setup reuse |
|---|---:|---|---|
| MongoDB | Yes | Read `sessions`, read `session_events`, write/read `journey_summaries` | Reuse previous MongoDB setup |
| Redis | Yes for service startup | Not primary journey source, but current app initializes Redis | Reuse previous Redis setup |

### 5.2 MongoDB

#### A. What MongoDB Is

MongoDB ek NoSQL document database hai. Isme data JSON-like documents ke form me collections ke andar store hota hai.

Simple English: MongoDB stores flexible documents instead of fixed SQL rows.

#### B. Why Task 4 Uses MongoDB

Task 4 journey timeline MongoDB se banti hai:

| Collection | Purpose |
|---|---|
| `sessions` | Session metadata: session id, anonymous id, user id, start time, last seen, device context |
| `session_events` | Raw user activity events: page views, clicks, searches, add to cart, checkout events |
| `journey_summaries` | Derived compact summary for dashboard and faster future reads |

Hinglish: Raw events source-of-truth hain. `journey_summaries` sirf calculated summary hai. Agar summary delete bhi ho jaye, raw events se journey rebuild ho sakti hai.

#### C. Required Or Optional

MongoDB required hai. Journey endpoint MongoDB se session and events read karta hai. MongoDB unavailable hua to journey API `503 JOURNEY_STORAGE_UNAVAILABLE` return kar sakti hai.

#### D. Local Installation

MongoDB local install already detail me explained hai.

Refer:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`5.1 MongoDB Setup`

#### E. Docker Setup

Same Docker MongoDB setup previous file me already present hai.

Refer:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Sections:

- `5.1 MongoDB Setup`
- `8. Docker Setup`

#### F. Start Commands

Use same MongoDB start flow from previous docs.

Refer:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`9. Local Development Setup`, Step `4. Start MongoDB and Redis`

#### G. Verify Running

Basic MongoDB ping:

```bash
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
```

Verify Task 4 collection after migration:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.journey_summaries.getIndexes()"
```

#### H. Default Port

| MongoDB default port | `27017` |
|---|---:|
| Local URI | `mongodb://localhost:27017` |
| Docker Compose internal URI | `mongodb://mongo:27017` |

#### I. Connection String Format

Local:

```env
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_MONGO_DATABASE=session_db
```

Docker Compose internal service-to-service:

```env
SESSION_MONGO_URI=mongodb://mongo:27017
SESSION_MONGO_DATABASE=session_db
```

With username/password:

```env
SESSION_MONGO_URI=mongodb://session_user:strong_password@localhost:27017/?authSource=admin
SESSION_MONGO_DATABASE=session_db
```

#### J. Where To Place Credentials

Local dev:

```text
backend/services/session-service/.env
```

Production:

- Secret manager
- Kubernetes Secret
- CI/CD secret variables
- Managed platform environment variables

Never commit real MongoDB passwords in Git.

### 5.3 Task 4 MongoDB Collection: `journey_summaries`

Task 4 introduces this derived collection:

```text
journey_summaries
```

Purpose:

- Store compact journey result per session.
- Speed up dashboard summary views.
- Keep raw events as source-of-truth.

Important fields:

| Field | Meaning |
|---|---|
| `session_id` | Session whose journey summary this belongs to |
| `anonymous_id` | Anonymous visitor id |
| `user_id` | Optional logged-in user id |
| `entry_page` | First page/path in journey |
| `exit_page` | Last page/path in journey |
| `first_event_at` | First event timestamp |
| `last_event_at` | Last event timestamp |
| `duration_seconds` | Journey duration |
| `total_events` | Events included in summary |
| `products_viewed` | Count of product view events |
| `searches` | Count of search events |
| `cart_actions` | Count of add-to-cart actions |
| `checkout_started` | Whether checkout started |
| `payment_completed` | Whether successful payment event exists |
| `milestones` | Important first events like first page view, first product view, payment completed |
| `top_paths` | Most visited paths in that session |

### 5.4 Required MongoDB Indexes

Task 4 depends on one existing index from Task 1/2:

```javascript
db.session_events.createIndex(
  { session_id: 1, occurred_at: 1 },
  { name: "idx_session_timeline" }
)
```

This is already part of:

`backend/services/session-service/migrations/001_session_storage_indexes.up.js`

Task 4 adds these indexes:

```javascript
db.journey_summaries.createIndex(
  { session_id: 1 },
  { unique: true, name: "uniq_journey_session" }
)

db.journey_summaries.createIndex(
  { user_id: 1, last_event_at: -1 },
  {
    name: "idx_journey_user_recent",
    partialFilterExpression: { user_id: { $type: "string" } }
  }
)

db.journey_summaries.createIndex(
  { last_event_at: -1 },
  { name: "idx_journey_recent" }
)
```

These are created by:

`backend/services/session-service/migrations/002_journey_summaries.up.js`

The Go repository also calls `EnsureIndexes()` during startup, but running migrations is still recommended because migration also applies MongoDB collection validation.

### 5.5 Task 4 Migration

Previous migration process is already explained here:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`9. Local Development Setup`, Step `8. Run MongoDB migrations`

Task 4-specific migration file:

```text
backend/services/session-service/migrations/002_journey_summaries.up.js
```

Run only Task 4 migration:

```bash
cd backend/services/session-service
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/002_journey_summaries.up.js
```

Recommended order on fresh database:

```bash
cd backend/services/session-service
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/001_session_storage_indexes.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/002_journey_summaries.up.js
```

Rollback Task 4 migration in local dev only:

```bash
cd backend/services/session-service
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/002_journey_summaries.down.js
```

Warning:

`002_journey_summaries.down.js` drops the `journey_summaries` collection. Local dev me okay ho sakta hai, but production me carefully backup/approval ke bina run mat karo.

### 5.6 Verify Task 4 Database Setup

Check collection exists:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.getCollectionInfos({ name: 'journey_summaries' })"
```

Check indexes:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.journey_summaries.getIndexes()"
```

Expected index names:

```text
_id_
uniq_journey_session
idx_journey_user_recent
idx_journey_recent
```

Check base timeline index:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.session_events.getIndexes().map(i => i.name)"
```

Expected:

```text
idx_session_timeline
```

## 6. Redis / Queue / External Services

### 6.1 Redis

Redis setup already explained:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`5.2 Redis Setup`

Task 4-specific Redis note:

- Journey timeline itself is read from MongoDB.
- Current app still requires Redis at startup because dependencies initialize Redis client and ping it.
- If Redis is down, service startup/readiness can fail before you even call journey API.

Required local env:

```env
SESSION_REDIS_ADDR=localhost:6379
SESSION_REDIS_PASSWORD=
SESSION_REDIS_DB=2
```

Verify:

```bash
redis-cli ping
```

Expected:

```text
PONG
```

### 6.2 Kafka / RabbitMQ / Queue

No Kafka/RabbitMQ queue is required for Task 4.

Journey is currently built by read-time MongoDB query plus optional summary upsert. A background worker could be added later for scale, but Task 4 local setup does not need queue infrastructure.

### 6.3 API Gateway / Admin Auth Context

Journey endpoint is admin/dashboard API:

```http
GET /api/v1/analytics/sessions/{session_id}/journey
```

Current service checks admin access through trusted headers:

```http
X-User-Roles: admin
```

Allowed admin roles:

```text
admin
superadmin
operations_admin
```

Local manual testing can use:

```bash
curl -i \
  -H "X-User-Roles: admin" \
  "http://localhost:8086/api/v1/analytics/sessions/sess_123/journey"
```

Production security requirement:

The service should sit behind an API Gateway/Auth service that validates JWT/session auth, strips spoofed incoming auth headers, and injects trusted `X-User-Roles` headers. Direct public exposure of this service is unsafe.

### 6.4 Optional Services Not Introduced By Task 4

Current `.env.example` also contains device tracking, GeoIP, heatmap, and analytics variables because later tasks exist in this repository.

For Task 4 only:

| Service/Feature | Task 4 required? | Note |
|---|---:|---|
| MaxMind GeoIP database | No | Device tracking related, not journey setup |
| Heatmap aggregation worker | No for Task 4 | Later heatmap task |
| Analytics aggregate cache | No for Task 4 | Later analytics task |
| Kafka/RabbitMQ | No | Future async processing only |

## 7. Environment Variables

Full `.env` creation/loading process already exists.

Refer:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`7. Environment Variables`

Very important:

The Go service does not automatically read `.env` files by itself. You must export/source variables in your shell, use a tool that loads `.env`, or configure them in Docker/CI/Kubernetes.

Task 4-specific `.env` subset:

```env
# Journey timeline read limits
SESSION_JOURNEY_DEFAULT_LIMIT=1000
SESSION_JOURNEY_MAX_LIMIT=1000

# Journey summary settings
SESSION_JOURNEY_SUMMARY_TOP_PATHS_LIMIT=10
SESSION_JOURNEY_SUMMARY_UPSERT_ENABLED=true
```

Variable table:

| Variable | Required? | Default | Purpose | Beginner note |
|---|---:|---:|---|---|
| `SESSION_JOURNEY_DEFAULT_LIMIT` | Optional | `1000` | Default number of events returned if query `limit` is absent | Zyada high value API slow kar sakti hai |
| `SESSION_JOURNEY_MAX_LIMIT` | Optional | `1000` | Maximum allowed journey events per request | Must be greater than or equal to default |
| `SESSION_JOURNEY_SUMMARY_TOP_PATHS_LIMIT` | Optional | `10` | Summary me top paths kitne store/return karne hain | Positive integer required |
| `SESSION_JOURNEY_SUMMARY_UPSERT_ENABLED` | Optional | `true` | GET journey ke baad `journey_summaries` upsert enable/disable | Read-only DB user ho to `false` kar sakte ho |

Validation rules from current config:

| Rule | Error if broken |
|---|---|
| `SESSION_JOURNEY_DEFAULT_LIMIT > 0` | `SESSION_JOURNEY_DEFAULT_LIMIT must be greater than zero` |
| `SESSION_JOURNEY_MAX_LIMIT >= SESSION_JOURNEY_DEFAULT_LIMIT` | `SESSION_JOURNEY_MAX_LIMIT must be greater than or equal to default journey limit` |
| `SESSION_JOURNEY_SUMMARY_TOP_PATHS_LIMIT > 0` | `SESSION_JOURNEY_SUMMARY_TOP_PATHS_LIMIT must be greater than zero` |

Credentials placement:

| Secret/config | Local file | Production placement |
|---|---|---|
| Mongo URI/password | `backend/services/session-service/.env` | Secret manager / Kubernetes Secret |
| Redis password | `backend/services/session-service/.env` | Secret manager / Kubernetes Secret |
| Journey limits | `.env` or environment | ConfigMap/environment variables |
| Admin role/JWT validation | API Gateway/Auth service config | Gateway/Auth service, not plain client headers |

Minimum local env for Task 4 journey work:

```env
SESSION_HTTP_ADDR=:8086
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_MONGO_DATABASE=session_db
SESSION_REDIS_ADDR=localhost:6379
SESSION_REDIS_PASSWORD=
SESSION_REDIS_DB=2
SESSION_JOURNEY_DEFAULT_LIMIT=1000
SESSION_JOURNEY_MAX_LIMIT=1000
SESSION_JOURNEY_SUMMARY_TOP_PATHS_LIMIT=10
SESSION_JOURNEY_SUMMARY_UPSERT_ENABLED=true
```

Do not repeat the full `.env` here. The complete environment example already exists in:

`backend/services/session-service/.env.example`

and is explained in:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`7. Environment Variables`

## 8. Docker Setup

No new Docker container is introduced by Task 4.

Reuse previous Docker setup:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`8. Docker Setup`

Task 4 Docker-specific notes:

| Scenario | MongoDB value | Redis value | Journey impact |
|---|---|---|---|
| Go service runs on host | `SESSION_MONGO_URI=mongodb://localhost:27017` | `SESSION_REDIS_ADDR=localhost:6379` | Common beginner setup |
| Go service runs inside Docker Compose | `SESSION_MONGO_URI=mongodb://mongo:27017` | `SESSION_REDIS_ADDR=redis:6379` | Use Compose service names, not localhost |
| MongoDB Atlas / managed Redis | Managed URI/host | Managed Redis host | Use secrets, TLS/private networking |

Persistent volume reminder:

- MongoDB volume matters because `sessions`, `session_events`, and `journey_summaries` should survive container restart.
- Redis can use volume/AOF locally for debugging, but Redis is not the durable source of truth.

Current repo Docker artifact status for session service:

| Artifact | Current status | Impact |
|---|---|---|
| Official session-service Dockerfile | Not committed in current service folder | Full app container build is not standardized yet |
| Official local docker-compose file | Not committed in current service folder | Beginners should use previous documented Docker commands/compose example |
| MongoDB/Redis containers | Documented previously | Enough for local Task 4 development |

## 9. Local Development Setup

This section is incremental for Task 4. For full clone/install flow, follow:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`9. Local Development Setup`

Task 4 onboarding flow:

1. Clone repo and enter project.
2. Install/verify Go.
3. Start MongoDB.
4. Start Redis.
5. Configure base `.env` from previous docs.
6. Add Task 4 journey variables from section `7. Environment Variables`.
7. Run Task 1/2 base migration first.
8. Run Task 4 migration `002_journey_summaries.up.js`.
9. Run tests.
10. Start backend.
11. Ingest or seed sample events.
12. Call journey endpoint with admin role header.

Commands:

```bash
cd backend/services/session-service
go mod download
```

Load env from `.env` in bash:

```bash
set -a
. ./.env
set +a
```

Run migrations:

```bash
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/001_session_storage_indexes.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/002_journey_summaries.up.js
```

Run tests:

```bash
go test ./...
```

Start service:

```bash
go run ./cmd/server
```

Verify health:

```bash
curl -i http://localhost:8086/healthz
curl -i http://localhost:8086/readyz
```

Expected readiness should include MongoDB and Redis as healthy/ok.

## 10. Running the Project

Daily Task 4 run flow:

```bash
cd backend/services/session-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Journey endpoint format:

```http
GET /api/v1/analytics/sessions/{session_id}/journey
```

Query parameters:

| Parameter | Required? | Example | Purpose |
|---|---:|---|---|
| `limit` | No | `200` | Return max N events |
| `cursor` | No | `2026-05-22T10:03:00Z` | Return events after this `occurred_at` |

Manual request:

```bash
curl -i \
  -H "X-User-Roles: admin" \
  "http://localhost:8086/api/v1/analytics/sessions/sess_01HX9ZK6N5M2V4B7S8T9Q0R1P2/journey?limit=200"
```

With cursor:

```bash
curl -i \
  -H "X-User-Roles: admin" \
  "http://localhost:8086/api/v1/analytics/sessions/sess_01HX9ZK6N5M2V4B7S8T9Q0R1P2/journey?limit=200&cursor=2026-05-22T10:03:00Z"
```

Expected success:

```http
HTTP/1.1 200 OK
Content-Type: application/json
```

Expected response contains:

```json
{
  "session": {
    "session_id": "sess_..."
  },
  "events": [],
  "summary": {
    "duration_seconds": 0,
    "total_events": 0
  }
}
```

### Ports & Networking

| Service | Port | Purpose | Task 4 status |
|---|---:|---|---|
| Session Backend API | `8086` | Journey endpoint, health, readiness | Reused |
| MongoDB | `27017` | `sessions`, `session_events`, `journey_summaries` | Reused |
| Redis | `6379` | Existing active-session/cache dependency | Reused |
| API Gateway | Usually `8080` in platform docs | Should protect admin endpoint | No new port |
| Kafka/RabbitMQ | N/A | Not required for Task 4 | Not used |

Port conflict troubleshooting already exists:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`11. Ports & Networking`

## 11. Common Errors & Fixes

General setup issues are already covered:

`TaskImplementation/Session Management Service/task1_Dependency.md`

Section:

`12. Common Errors & Fixes`

Task 4-specific errors:

### 1. `401 AUTHENTICATION_REQUIRED`

Cause:

Journey endpoint is admin-only and request has no auth context/admin role header.

Fix for local testing:

```bash
curl -i \
  -H "X-User-Roles: admin" \
  "http://localhost:8086/api/v1/analytics/sessions/sess_123/journey"
```

Production fix:

Use API Gateway/Auth service to validate real JWT/session and inject trusted admin headers.

### 2. `403 PERMISSION_DENIED`

Cause:

Request has auth context, but role is not allowed. Example: `buyer`, `seller`, or missing admin role.

Allowed roles:

```text
admin
superadmin
operations_admin
```

Fix:

Use a valid admin role:

```http
X-User-Roles: admin
```

### 3. `404 SESSION_NOT_FOUND`

Cause:

Session does not exist in MongoDB `sessions` collection.

Fix:

- First create/ingest session activity through Task 3 event ingestion.
- Verify session exists:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.sessions.findOne({ session_id: 'sess_123' })"
```

### 4. `400 VALIDATION_ERROR: session_id must start with sess_`

Cause:

Current validation requires journey path session id to start with `sess_`.

Wrong:

```text
/api/v1/analytics/sessions/123/journey
```

Correct:

```text
/api/v1/analytics/sessions/sess_123/journey
```

### 5. `400 VALIDATION_ERROR: limit must be an integer`

Cause:

Query parameter `limit` is not a number.

Wrong:

```text
?limit=abc
```

Correct:

```text
?limit=200
```

### 6. `400 VALIDATION_ERROR: limit cannot exceed 1000`

Cause:

Requested `limit` is higher than `SESSION_JOURNEY_MAX_LIMIT`.

Fix options:

- Use smaller request limit.
- Increase `SESSION_JOURNEY_MAX_LIMIT` carefully.

Example:

```env
SESSION_JOURNEY_DEFAULT_LIMIT=1000
SESSION_JOURNEY_MAX_LIMIT=2000
```

Warning:

Higher limits can make MongoDB query and response heavy for large sessions.

### 7. `400 VALIDATION_ERROR: cursor must be RFC3339`

Cause:

Cursor is not valid RFC3339/RFC3339Nano time.

Wrong:

```text
?cursor=22-05-2026
```

Correct:

```text
?cursor=2026-05-22T10:03:00Z
```

### 8. `503 JOURNEY_STORAGE_UNAVAILABLE`

Cause:

MongoDB query failed while reading session or timeline events.

Fix:

```bash
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
curl -i http://localhost:8086/readyz
```

Also verify env:

```env
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_MONGO_DATABASE=session_db
```

### 9. `journey_summaries` Collection Missing

Cause:

Task 4 migration was not run. Go startup can create indexes, but migration adds collection validation and makes DB state explicit.

Fix:

```bash
cd backend/services/session-service
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/002_journey_summaries.up.js
```

Verify:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.journey_summaries.getIndexes()"
```

### 10. Summary Not Persisted But API Returns `200`

Cause:

Current behavior can still return raw journey timeline even if summary upsert is disabled or fails.

Possible reasons:

- `SESSION_JOURNEY_SUMMARY_UPSERT_ENABLED=false`
- Request used `cursor`, so current usecase skips summary persistence for paginated reads
- MongoDB write failed and warning was logged

Fix:

```env
SESSION_JOURNEY_SUMMARY_UPSERT_ENABLED=true
```

Call without cursor for full summary upsert:

```bash
curl -i \
  -H "X-User-Roles: admin" \
  "http://localhost:8086/api/v1/analytics/sessions/sess_123/journey"
```

### 11. Startup Error: `SESSION_JOURNEY_MAX_LIMIT must be greater than or equal to default journey limit`

Cause:

Invalid env combination.

Wrong:

```env
SESSION_JOURNEY_DEFAULT_LIMIT=1000
SESSION_JOURNEY_MAX_LIMIT=500
```

Correct:

```env
SESSION_JOURNEY_DEFAULT_LIMIT=500
SESSION_JOURNEY_MAX_LIMIT=1000
```

### 12. Events Look Out Of Order

Cause:

`occurred_at` is client-side event time. If frontend sends wrong time or retries old events, journey order can look surprising.

Fix:

- Ensure frontend sends RFC3339 UTC timestamps.
- Compare `occurred_at` and `received_at` in MongoDB:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.session_events.find({ session_id: 'sess_123' }, { event_type: 1, occurred_at: 1, received_at: 1 }).sort({ occurred_at: 1, received_at: 1 })"
```

### 13. Duplicate Events In Journey

Cause:

Frontend retries or missing stable event id/request id.

Current code deduplicates by:

- `event_id`
- retry key based on `session_id`, `event_type`, `occurred_at`, `path`, `request_id`

Fix:

- Send stable `X-Request-ID` during ingestion.
- Prefer stable event ids from SDK where possible.

## 12. Security & Best Practices

### 12.1 Security Audit For Task 4

| Finding | Risk | Recommended fix |
|---|---|---|
| Journey endpoint trusts role headers | Direct public callers could spoof `X-User-Roles: admin` | Put service behind API Gateway that strips incoming auth headers and injects trusted ones |
| Journey exposes behavior timeline | User activity is privacy-sensitive | Keep endpoint admin-only and log access |
| MongoDB local URI has no auth | Fine local, unsafe production | Use MongoDB auth, TLS, private networking |
| Redis password blank locally | Fine local, unsafe production | Use Redis AUTH/TLS/private network in production |
| `.env` contains placeholder secrets | Placeholder salts are unsafe | Replace salts/passwords with strong secrets |
| Summary upsert happens during GET | GET can perform DB write | Keep idempotent, monitor writes, disable with `SESSION_JOURNEY_SUMMARY_UPSERT_ENABLED=false` for read-only deployments |
| No migration tracking collection | Hard to know which migrations ran | Add migration tracking later |
| No official Dockerfile/compose in service folder | Onboarding/deploy inconsistency | Add official service Dockerfile and local compose later |
| Cursor uses only `occurred_at` value | Same-timestamp events can be tricky for pagination | Future improvement: compound cursor using `occurred_at`, `received_at`, and `_id`/`event_id` |
| Timeline index does not include all sort tie-breakers | Mongo may need extra sort work for large same-time sessions | Future improvement: index `{ session_id: 1, occurred_at: 1, received_at: 1, _id: 1 }` |

### 12.2 Privacy Rules

Never expose/store sensitive values in journey properties:

```text
password
otp
token
authorization
cookie
card_number
cvv
raw_ip
secret
```

Task 3 sanitization already protects ingestion side, and Task 4 read side also sanitizes properties. Still, frontend SDK should avoid sending sensitive data in the first place.

### 12.3 Beginner Best Practices

- Run migrations in order: `001` first, then `002`.
- Keep `SESSION_JOURNEY_MAX_LIMIT` reasonable. High value can slow APIs.
- Use `X-User-Roles: admin` only for local testing. Production should use real gateway-authenticated headers.
- Treat `session_events` as source-of-truth and `journey_summaries` as derived/cache-like data.
- Verify indexes before performance testing.
- Do not expose MongoDB, Redis, or session-service directly to the public internet.
- Use `/readyz` before debugging API behavior. If readiness fails, fix MongoDB/Redis first.
- In production, monitor `session.journey.summary_upsert_failed` logs.

## 13. Missing or Misconfigured Things

Task 4-specific gaps and recommendations:

| Item | Current state | Impact | Recommendation |
|---|---|---|---|
| Migration tracking | Not present | Developers may not know whether `002` ran | Add migration tool/tracking collection |
| Official Dockerfile | Not committed for session service | Harder consistent container build | Add service Dockerfile |
| Official local compose | Not committed for session service | Beginners rely on docs/manual containers | Add local compose with MongoDB, Redis, service |
| Admin auth source | Header-based inside service | Unsafe if service public | Enforce through API Gateway |
| Compound pagination cursor | Current cursor is timestamp-only | Same timestamp pagination edge case | Use compound cursor later |
| Compound timeline index | Existing index covers main filter/order but not all tie-breakers | Large sessions can sort less efficiently | Add index with `received_at` and `_id` if needed |
| Summary upsert observability | Warning log exists, metrics not confirmed | Failures can be missed | Add metric/counter for upsert failures |
| Production Mongo/Redis secrets | Local examples are blank/no-auth | Unsafe production default | Use secrets and private networking |

## 14. References to Previous Dependency Files

Do not duplicate these sections. Follow the existing docs:

| Topic | Previous file | Section |
|---|---|---|
| Base project overview and full setup | `TaskImplementation/Session Management Service/task1_Dependency.md` | `1. Project Overview` |
| Go, net/http, MongoDB, Redis beginner explanations | `task1_Dependency.md` | `2. Tech Stack` |
| Required software install | `task1_Dependency.md` | `3. Required Software` |
| Go modules, `go.mod`, `go.sum`, `go mod tidy`, `go mod download` | `task1_Dependency.md` | `4. Dependency Management` |
| MongoDB install, Docker run, verify | `task1_Dependency.md` | `5.1 MongoDB Setup` |
| Redis install, Docker run, verify | `task1_Dependency.md` | `5.2 Redis Setup` |
| Full `.env` creation/loading | `task1_Dependency.md` | `7. Environment Variables` |
| Docker Compose example | `task1_Dependency.md` | `8. Docker Setup` |
| Full local development setup | `task1_Dependency.md` | `9. Local Development Setup` |
| Running service basics | `task1_Dependency.md` | `10. Running the Project` |
| Ports and networking basics | `task1_Dependency.md` | `11. Ports & Networking` |
| General troubleshooting | `task1_Dependency.md` | `12. Common Errors & Fixes` |
| Storage architecture and MongoDB/Redis collection/key design | `task2_Dependency.md` | `5. Database Setup`, `6. Redis / Queue / External Services` |
| Event ingestion setup and event verification | `task3_Dependency.md` | `9. Local Development Setup`, `10. Running the Project` |
| Event ingestion specific errors | `task3_Dependency.md` | `11. Common Errors & Fixes` |

Task 4 new material in this file:

- `journey_summaries` collection setup
- `002_journey_summaries.up.js` migration
- Journey env variables
- Admin journey endpoint access requirements
- Journey-specific troubleshooting
- Journey-specific security audit

## 15. Final Checklist

Use this checklist before saying Task 4 dependency/setup is ready:

| Check | Done |
|---|---|
| Go installed and `go version` is compatible with `go.mod` |  |
| `go mod download` succeeds in `backend/services/session-service` |  |
| MongoDB is running on `27017` or configured URI |  |
| Redis is running on `6379` or configured address |  |
| `.env` is created or env vars are exported |  |
| Base MongoDB variables are set |  |
| Base Redis variables are set |  |
| `SESSION_JOURNEY_DEFAULT_LIMIT` is set or default accepted |  |
| `SESSION_JOURNEY_MAX_LIMIT >= SESSION_JOURNEY_DEFAULT_LIMIT` |  |
| `SESSION_JOURNEY_SUMMARY_TOP_PATHS_LIMIT > 0` |  |
| `SESSION_JOURNEY_SUMMARY_UPSERT_ENABLED` is intentionally true/false |  |
| Migration `001_session_storage_indexes.up.js` has run |  |
| Migration `002_journey_summaries.up.js` has run |  |
| `idx_session_timeline` exists on `session_events` |  |
| `uniq_journey_session` exists on `journey_summaries` |  |
| `idx_journey_user_recent` exists on `journey_summaries` |  |
| `idx_journey_recent` exists on `journey_summaries` |  |
| `go test ./...` passes or known dependency-related failures are understood |  |
| Service starts with `go run ./cmd/server` |  |
| `/healthz` works |  |
| `/readyz` confirms MongoDB and Redis |  |
| Sample session/events exist in MongoDB |  |
| Journey endpoint called with `X-User-Roles: admin` |  |
| Journey endpoint returns `200 OK` for valid session |  |
| `journey_summaries` gets upserted when summary upsert is enabled and request has no cursor |  |
| Production plan includes API Gateway/Auth protection |  |
| Production MongoDB/Redis credentials are not committed to Git |  |

Final Hinglish summary:

Task 4 ke liye koi naya database server ya queue install nahi karna. Existing MongoDB + Redis setup reuse hota hai. Naya kaam mainly MongoDB me `journey_summaries` migration run karna, journey env variables set karna, and admin-only journey endpoint ko properly test/protect karna hai.
