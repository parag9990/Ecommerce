# Project Dependency & Setup Guide

This guide is for:

```text
TaskImplementation/Session Management Service/task6.md
```

Output file:

```text
TaskImplementation/Session Management Service/task6_Dependency.md
```

Goal simple hai: Task 6 me **Heatmap Concept** ke liye click and scroll events ko aggregate karke dashboard-friendly heatmap points banana hai. Ye document sirf dependency, setup, environment, database, Docker, DevOps, and debugging onboarding ke liye hai. Business logic yahan rewrite nahi ki gayi.

Important reuse rule: same service folder me pehle se common setup docs available hain:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
TaskImplementation/Session Management Service/task2_Dependency.md
TaskImplementation/Session Management Service/task3_Dependency.md
TaskImplementation/Session Management Service/task4_Dependency.md
TaskImplementation/Session Management Service/task5_Dependency.md
```

Isliye Go install, MongoDB install, Redis install, Docker basics, base `.env`, event ingestion setup, admin header behavior, device tracking, and GeoIP setup ko yahan duplicate nahi kiya gaya. Jahan setup same hai, exact reference diya gaya hai.

Note: `task6.md` describes the heatmap work as a concept guide, but the current repository also contains actual Task 6 runtime files such as:

```text
backend/services/session-service/internal/domain/heatmap.go
backend/services/session-service/internal/usecase/aggregate_heatmap.go
backend/services/session-service/internal/usecase/get_heatmap.go
backend/services/session-service/internal/repository/mongo_heatmap_repository.go
backend/services/session-service/migrations/004_heatmap_points.up.js
```

This dependency guide documents the current repo state.

---

## 1. Project Overview

Task 6 ka purpose hai raw `click` and `scroll` events ko heatmap aggregate data me convert karna.

Simple Hinglish:

Frontend SDK user ke clicks aur scroll depth bhejta hai. Task 3 un raw events ko MongoDB ke `session_events` collection me save karta hai. Task 6 ka worker un events ko read karke normalized percentage buckets banata hai, then MongoDB me `heatmap_points` collection me aggregate write karta hai. Dashboard admin API se ye points read karta hai.

Current Task 6 runtime flow:

| Step | Component | Purpose |
|---|---|---|
| 1 | `POST /api/v1/sessions/events` | Click/scroll raw events receive karta hai |
| 2 | MongoDB `session_events` | Raw event source of truth |
| 3 | Heatmap aggregation worker | Background goroutine every interval raw events scan karta hai |
| 4 | MongoDB `heatmap_points` | Dashboard-friendly x/y/weight aggregate points |
| 5 | MongoDB `heatmap_bucket_sessions` | Per bucket unique session counting markers |
| 6 | MongoDB `heatmap_checkpoints` | Last processed worker timestamp |
| 7 | `GET /api/v1/analytics/heatmaps` | Admin dashboard heatmap read API |

Task 6 does not introduce a new standalone microservice. Heatmap worker same Session Service process ke andar start hota hai when `SESSION_HEATMAP_AGGREGATION_ENABLED=true`.

---

## 2. Tech Stack

### Task 6 Detected Technologies

| Technology | What it is | Why used in Task 6 | Required? | Setup reference |
|---|---|---|---:|---|
| Go | Compiled backend language | Session Service, heatmap usecases, worker, HTTP API | Yes | `task1_Dependency.md`, section `2. Tech Stack` |
| Go `net/http` | Built-in HTTP server | `GET /api/v1/analytics/heatmaps` route expose karta hai | Yes | `task1_Dependency.md`, section `2. Tech Stack` |
| Go goroutine + ticker | Go concurrency primitive | Background heatmap aggregation interval run karne ke liye | Yes, built-in | Explained below |
| MongoDB | Document database | Raw events, heatmap aggregates, checkpoints, unique-session markers | Yes | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| Redis | In-memory cache | Service startup/readiness and existing active session flow | Yes for app startup | `task1_Dependency.md`, section `5.2 Redis Setup` |
| MongoDB Go Driver | Go MongoDB client | `session_events` scan and `heatmap_points` upsert | Yes | `task1_Dependency.md`, section `4. Dependency Management` |
| go-redis/v9 | Go Redis client | Existing active session/live metrics dependencies | Yes | `task1_Dependency.md`, section `4. Dependency Management` |
| mongosh | MongoDB shell | Run `004_heatmap_points.up.js` migration and verify collections | Recommended | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| API Gateway / admin headers | Auth context provider | Heatmap API is admin-only | Production required | `task4_Dependency.md`, section `6.3 API Gateway / Admin Auth Context` |
| Device tracking | Enriched event context | Heatmap filtering uses `device.type` and viewport bucket | Recommended | `task5_Dependency.md`, sections `6.1` and `6.2` |
| Docker | Container runtime | Local MongoDB and Redis containers | Optional but recommended | `task1_Dependency.md`, section `8. Docker Setup` |

### Beginner Explanation

Go ek fast backend language hai. Is project me heatmap worker and API Go me hi run hote hain.

`net/http` Go ka built-in web framework hai. Gin/Fiber/Echo install karne ki zarurat nahi hai.

MongoDB ek document database hai. Heatmap data flexible documents me store hota hai, jaise `heatmap_type`, `path`, `device_type`, `day`, `x`, `y`, and `weight`.

Redis Task 6 ka heatmap store nahi hai, but Session Service Redis ke bina start nahi karega because existing active-session/live-metrics dependency app startup me ping hoti hai.

Go goroutine + ticker ka matlab: service ke andar ek background loop run hota hai jo fixed interval par heatmap aggregation karta hai. Iske liye separate cron container required nahi hai in current implementation.

### Frontend Libraries Mentioned In `task6.md`

`task6.md` future dashboard ke liye `heatmap.js`, `D3.js`, and `lodash` mention karta hai. Current backend repo me Node `package.json` ya dashboard implementation detect nahi hua.

| Library | Current backend requirement? | Note |
|---|---:|---|
| `heatmap.js` | No | Future dashboard rendering only |
| `d3` | No | Future visualization only |
| `lodash` | No | Future frontend throttle/debounce helper |

Beginner warning: backend run karne ke liye `npm install heatmap.js` mat chalao. Ye current Session Service backend dependency nahi hai.

---

## 3. Required Software

Base required software same hai as previous dependency docs.

| Software | Required? | Purpose | Setup reference |
|---|---:|---|---|
| Git | Yes | Repo clone karne ke liye | `task1_Dependency.md`, section `3. Required Software` |
| Go 1.26.3 or newer | Yes | Backend run/test/build | `task1_Dependency.md`, section `3. Required Software` |
| MongoDB | Yes | `session_events`, `heatmap_points`, checkpoints | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| mongosh | Recommended | Migrations and verification | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| Redis | Yes | Existing service startup/readiness | `task1_Dependency.md`, section `5.2 Redis Setup` |
| redis-cli | Recommended | Redis debug | `task1_Dependency.md`, section `5.2 Redis Setup` |
| Docker | Optional but recommended | Local MongoDB/Redis containers | `task1_Dependency.md`, section `8. Docker Setup` |
| curl/Postman | Recommended | Event ingestion and heatmap API test | `task1_Dependency.md`, section `9. Local Development Setup` |

Task 6 me koi new OS-level server install nahi hota. No Kafka, RabbitMQ, Elasticsearch, MinIO, S3, SMTP, Stripe, Twilio, or OAuth provider is required for Task 6 local setup.

---

## 4. Dependency Management

This is a Go project. Full beginner explanation for `go.mod`, `go.sum`, Go modules, `go mod tidy`, `go mod download`, `go build`, `go run`, module proxy issues, and version mismatch issues already exists here:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
4. Dependency Management
```

Task 6 does not add a new external Go module beyond the current Session Service dependencies.

Current direct dependencies in `backend/services/session-service/go.mod`:

| Dependency | Version | Why relevant for Task 6 |
|---|---:|---|
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | Read raw events, write heatmap collections, indexes |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Existing Redis startup/readiness and active session flow |
| `github.com/mileusna/useragent` | `v1.3.5` | Device context from Task 5, useful for heatmap device segmentation |
| `github.com/oschwald/geoip2-golang` | `v1.13.0` | Optional GeoIP enrichment from Task 5, not heatmap-specific |

Useful commands:

```bash
cd backend/services/session-service
go mod download
go mod tidy
go test ./...
go run ./cmd/server
```

Hinglish reminder:

`go.mod` dependency list hai. `go.sum` checksum/security lock file hai. `go mod tidy` unused packages hata deta hai and missing packages add karta hai.

---

## 5. Database Setup

Detected databases for Task 6:

| Database | Required? | Task 6 role |
|---|---:|---|
| MongoDB | Yes | Raw events source, heatmap aggregate store, checkpoints |
| Redis | Yes for service startup | Existing active-session/live metrics infra, not heatmap storage |
| MySQL | No | Not used by Session Management Service Task 6 |
| PostgreSQL | No | Not detected |
| SQLite | No | Not detected |

### 5.1 MongoDB

#### A. What MongoDB Is

MongoDB ek NoSQL document database hai. Ye JSON-like documents store karta hai. Heatmap ke liye ye useful hai because click/scroll aggregate records flexible fields rakhte hain.

#### B. Why Task 6 Uses MongoDB

Task 6 MongoDB me ye data use karta hai:

| Collection | Purpose |
|---|---|
| `session_events` | Raw `click` and `scroll` event source of truth |
| `heatmap_points` | Aggregated x/y/weight dashboard points |
| `heatmap_bucket_sessions` | Unique sessions count karne ke liye bucket-session marker |
| `heatmap_checkpoints` | Background worker ka last processed timestamp |

#### C. Required Or Optional

MongoDB required hai. MongoDB down hoga to:

- Service startup dependencies fail ho sakti hain.
- `/readyz` degraded ho sakta hai.
- Heatmap worker raw events read nahi kar payega.
- Heatmap API `INTERNAL_ERROR` return kar sakta hai.

#### D. Local Installation

Do not duplicate install steps. Follow:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
5.1 MongoDB Setup
```

#### E. Docker Setup

Same MongoDB Docker setup as previous tasks. Follow:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Sections:
- 5.1 MongoDB Setup
- 8. Docker Setup
```

#### F. Start Commands

If using previous Docker container:

```bash
docker start ecommerce-session-mongo
```

If using Docker Compose from `task1_Dependency.md`:

```bash
docker compose -f docker-compose.session.local.yml up -d mongo
```

#### G. Verify Running

```bash
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
```

Expected idea:

```text
ok: 1
```

#### H. Default Port

| Setting | Value |
|---|---|
| MongoDB port | `27017` |
| Local URI | `mongodb://localhost:27017` |
| Default database | `session_db` |

#### I. Connection String Format

Same as previous docs:

```env
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_MONGO_DATABASE=session_db
```

If service runs inside Docker Compose:

```env
SESSION_MONGO_URI=mongodb://mongo:27017
SESSION_MONGO_DATABASE=session_db
```

#### J. Where To Place Credentials

Local development:

```text
backend/services/session-service/.env
```

Production:

Use Kubernetes Secret, cloud secret manager, CI/CD secret variables, or platform environment variables. Real MongoDB username/password Git me commit mat karo.

Example authenticated URI:

```env
SESSION_MONGO_URI=mongodb://session_user:strong_password@mongo:27017/session_db?authSource=admin
SESSION_MONGO_DATABASE=session_db
```

### 5.2 Task 6 MongoDB Migration

Task 6 introduces this migration:

```text
backend/services/session-service/migrations/004_heatmap_points.up.js
```

It creates or updates validators for:

| Collection | Created by migration? | Purpose |
|---|---:|---|
| `heatmap_points` | Yes | Aggregate heatmap points |
| `heatmap_bucket_sessions` | Yes | Unique session markers |
| `heatmap_checkpoints` | Yes | Worker checkpoint state |

It also creates these indexes:

| Collection | Index name | Why needed |
|---|---|---|
| `heatmap_points` | `idx_heatmap_route_device_day` | Fast dashboard query by path/device/day |
| `heatmap_points` | `uniq_heatmap_click_bucket` | One click aggregate doc per bucket |
| `heatmap_points` | `uniq_heatmap_scroll_bucket` | One scroll aggregate doc per depth bucket |
| `heatmap_bucket_sessions` | `uniq_heatmap_bucket_session` | Count each session once per bucket |
| `heatmap_bucket_sessions` | `idx_heatmap_bucket_sessions_point` | Debug/lookup session markers by point |
| `heatmap_checkpoints` | `uniq_heatmap_checkpoint_worker` | One checkpoint per worker |
| `session_events` | `idx_events_heatmap_scan` | Fast worker scan for click/scroll events |

Run migrations in order from service folder:

```bash
cd backend/services/session-service
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/001_session_storage_indexes.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/002_journey_summaries.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/003_device_tracking_indexes.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/004_heatmap_points.up.js
```

Important:

- App startup also calls `EnsureIndexes`, but migrations add JSON schema validators. For clean local/prod setup, run migration `004_heatmap_points.up.js`.
- Do not run the `.down.js` file unless you intentionally want to delete heatmap collections and indexes.

Verify indexes:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval 'db.heatmap_points.getIndexes().map(i => i.name)'
mongosh "mongodb://localhost:27017/session_db" --eval 'db.heatmap_bucket_sessions.getIndexes().map(i => i.name)'
mongosh "mongodb://localhost:27017/session_db" --eval 'db.heatmap_checkpoints.getIndexes().map(i => i.name)'
```

Verify collections:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval 'db.getCollectionNames().filter(n => n.startsWith("heatmap"))'
```

Expected:

```text
heatmap_points,heatmap_bucket_sessions,heatmap_checkpoints
```

---

## 6. Redis / Queue / External Services

### 6.1 Redis

Redis setup is already fully explained in:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
5.2 Redis Setup

TaskImplementation/Session Management Service/task2_Dependency.md
Section:
6.1 Redis
```

Task 6 does not introduce a new Redis database, key prefix, or heatmap-specific Redis key.

Why Redis is still required:

- Current Session Service initializes Redis on startup.
- `/readyz` checks Redis.
- Event ingestion still refreshes active session state.
- Live analytics from later tasks also use Redis.

Default local values:

```env
SESSION_REDIS_ADDR=localhost:6379
SESSION_REDIS_PASSWORD=
SESSION_REDIS_DB=2
SESSION_REDIS_KEY_PREFIX=session
```

If service runs inside Docker Compose:

```env
SESSION_REDIS_ADDR=redis:6379
```

### 6.2 Background Heatmap Worker

Task 6 introduces an in-process background worker. Ye separate service/container nahi hai.

Worker behavior:

| Setting | Meaning |
|---|---|
| `SESSION_HEATMAP_AGGREGATION_ENABLED` | Worker on/off switch |
| `SESSION_HEATMAP_AGGREGATION_INTERVAL` | Kitni der me worker repeat hota hai |
| `SESSION_HEATMAP_AGGREGATION_BATCH_SIZE` | Ek batch me max raw events scan |
| `SESSION_HEATMAP_AGGREGATION_INITIAL_LOOKBACK` | No checkpoint ho to kitna past scan kare |
| `SESSION_HEATMAP_AGGREGATION_CHECKPOINT_LOOKBACK` | Checkpoint se thoda overlap reprocess safety ke liye |
| `SESSION_HEATMAP_AGGREGATION_WORKER_NAME` | Checkpoint worker name |

Important beginner note:

Service start karte hi worker immediately ek run karta hai, then ticker ke through interval par run karta hai. Agar event ka `occurred_at` initial lookback se bahar hai, worker usko process nahi karega.

### 6.3 Kafka / RabbitMQ / Queue

Kafka/RabbitMQ current Task 6 local setup ke liye required nahi hai.

`task6.md` stream consumer ko future high-traffic option mention karta hai, but current implementation MongoDB raw events se batch aggregation karta hai.

Refer:

```text
TaskImplementation/Session Management Service/task3_Dependency.md
Section:
6.3 Kafka / RabbitMQ / session.events
```

### 6.4 API Gateway / Admin Auth

Heatmap read API admin-only hai:

```http
GET /api/v1/analytics/heatmaps
```

Current local implementation trusted role headers read karta hai:

```http
X-User-Roles: admin
```

Allowed admin roles:

```text
admin
superadmin
operations_admin
```

Production me ye headers public client se directly trust nahi karne chahiye. API Gateway/Auth service should authenticate request and then inject trusted headers.

Full admin header behavior already documented in:

```text
TaskImplementation/Session Management Service/task4_Dependency.md
Section:
6.3 API Gateway / Admin Auth Context
```

### 6.5 Device Tracking / GeoIP

Heatmap aggregation works without GeoIP, but good device data improves dashboard filters like `device_type=mobile`.

Task 5 setup already explains:

```text
TaskImplementation/Session Management Service/task5_Dependency.md
Sections:
- 6.1 MaxMind GeoIP
- 6.2 API Gateway / Trusted Proxy Headers
- 7. Environment Variables
```

Task 6-specific note:

- `device.type` influences heatmap segmentation.
- If device tracking is disabled or headers are missing, heatmap points may use `device_type=unknown`.
- Dashboard queries must match the stored device type.

---

## 7. Environment Variables

Base `.env` setup, `.env` location, and loading behavior are already documented in:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
7. Environment Variables
```

Very important reminder:

The Go app reads environment variables using `os.Getenv`. It does not automatically load `.env`. In bash, load it like:

```bash
cd backend/services/session-service
set -a
. ./.env
set +a
```

### 7.1 Task 6 `.env` Example Only

Add or verify only these heatmap-specific variables in:

```text
backend/services/session-service/.env
```

```env
# Heatmap query and aggregation
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

For faster local testing, you can temporarily use:

```env
SESSION_HEATMAP_AGGREGATION_INTERVAL=10s
SESSION_HEATMAP_AGGREGATION_INITIAL_LOOKBACK=2h
```

Restart the service after changing `.env`.

### 7.2 Task 6 Environment Variable Table

| Variable | Required? | Default | Meaning | Beginner note |
|---|---:|---|---|---|
| `SESSION_HEATMAP_CLICK_BUCKET_SIZE` | Optional | `5` | Click x/y percentage bucket size | `5` means points snap to 0,5,10...100 |
| `SESSION_HEATMAP_MAX_DATE_RANGE_DAYS` | Optional | `31` | Max API query date range | Prevents huge Mongo queries |
| `SESSION_HEATMAP_MAX_POINTS` | Optional | `5000` | Max points returned by API | Dashboard safety limit |
| `SESSION_HEATMAP_AGGREGATION_ENABLED` | Optional | `true` | Start background worker | Set `false` to disable aggregation |
| `SESSION_HEATMAP_AGGREGATION_INTERVAL` | Optional | `1m` | Worker repeat interval | Use Go duration style like `10s`, `1m`, `5m` |
| `SESSION_HEATMAP_AGGREGATION_BATCH_SIZE` | Optional | `1000` | Raw events per scan batch | Increase carefully for high traffic |
| `SESSION_HEATMAP_AGGREGATION_INITIAL_LOOKBACK` | Optional | `1h` | First run scan window if no checkpoint | Old test events can be skipped if too small |
| `SESSION_HEATMAP_AGGREGATION_CHECKPOINT_LOOKBACK` | Optional | `5m` | Overlap before last checkpoint | Helps catch late events, but can reprocess |
| `SESSION_HEATMAP_AGGREGATION_WORKER_NAME` | Optional | `heatmap_aggregator` | Checkpoint key name | Keep stable per worker |

Validation errors you may see:

| Bad value | Startup error |
|---|---|
| `SESSION_HEATMAP_CLICK_BUCKET_SIZE=0` | `SESSION_HEATMAP_CLICK_BUCKET_SIZE must be between 1 and 100` |
| `SESSION_HEATMAP_MAX_DATE_RANGE_DAYS=0` | `SESSION_HEATMAP_MAX_DATE_RANGE_DAYS must be greater than zero` |
| `SESSION_HEATMAP_MAX_POINTS=0` | `SESSION_HEATMAP_MAX_POINTS must be greater than zero` |
| `SESSION_HEATMAP_AGGREGATION_INTERVAL=0s` | `SESSION_HEATMAP_AGGREGATION_INTERVAL must be greater than zero` |
| `SESSION_HEATMAP_AGGREGATION_BATCH_SIZE=0` | `SESSION_HEATMAP_AGGREGATION_BATCH_SIZE must be greater than zero` |
| empty worker name | `SESSION_HEATMAP_AGGREGATION_WORKER_NAME cannot be empty` |

### 7.3 Credentials Placement

Task 6 does not add a new secret. Required credential placement remains:

| Credential | Local file | Production placement |
|---|---|---|
| MongoDB URI/password | `backend/services/session-service/.env` | Secret manager / Kubernetes Secret |
| Redis password | `backend/services/session-service/.env` | Secret manager / Kubernetes Secret |
| IP/hash pepper from previous tasks | `backend/services/session-service/.env` | Secret manager / Kubernetes Secret |

Heatmap-specific variables are configuration, not secrets.

---

## 8. Docker Setup

No new Docker container is introduced by Task 6.

Reuse previous Docker setup:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
8. Docker Setup
```

Task 6 Docker-specific notes:

| Scenario | MongoDB value | Redis value | Heatmap worker impact |
|---|---|---|---|
| Go service runs on host | `SESSION_MONGO_URI=mongodb://localhost:27017` | `SESSION_REDIS_ADDR=localhost:6379` | Worker runs inside host Go process |
| Go service runs inside Docker Compose | `SESSION_MONGO_URI=mongodb://mongo:27017` | `SESSION_REDIS_ADDR=redis:6379` | Worker runs inside service container |
| Aggregation disabled | Same as base | Same as base | API can read old points, but no new points created |

Current repo Docker artifact status:

| Artifact | Current status | Impact |
|---|---|---|
| Official session-service Dockerfile | Not committed in current service folder | Full app image build is not standardized yet |
| Official local docker-compose file | Not committed in current service folder | Use previous MongoDB/Redis Docker commands or compose example |
| MongoDB/Redis containers | Documented previously | Enough for Task 6 local dev |
| Separate heatmap worker container | Not present | Current worker is in-process |

Persistent volume reminder:

- MongoDB volume is important because `heatmap_points`, `heatmap_bucket_sessions`, and `heatmap_checkpoints` should survive restart.
- Redis persistence is not required for heatmap aggregates, but useful for existing active session debug.
- `docker compose down -v` deletes local MongoDB heatmap data.

---

## 9. Local Development Setup

This section is incremental for Task 6. Full clone/install flow is already documented:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
9. Local Development Setup
```

Task 6 onboarding flow:

1. Clone repo and enter project.
2. Install/verify Go.
3. Start MongoDB.
4. Start Redis.
5. Create/load base `.env` from previous docs.
6. Add Task 6 heatmap env variables from section `7. Environment Variables`.
7. Run migrations `001`, `002`, `003`, then `004_heatmap_points.up.js`.
8. Run tests.
9. Start backend.
10. Send one `click` and one `scroll` event.
11. Wait for heatmap worker interval.
12. Verify MongoDB `heatmap_points`.
13. Call heatmap API with admin role header.

Commands:

```bash
cd backend/services/session-service
go mod download
```

Load env in bash:

```bash
set -a
. ./.env
set +a
```

Run migrations:

```bash
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/001_session_storage_indexes.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/002_journey_summaries.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/003_device_tracking_indexes.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/004_heatmap_points.up.js
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

Expected readiness should include MongoDB and Redis as `ok`.

---

## 10. Running the Project

Daily Task 6 run flow:

```bash
cd backend/services/session-service
set -a
. ./.env
set +a
go run ./cmd/server
```

### 10.1 Send Heatmap Test Events

Use current timestamp because event ingestion rejects events older than the configured max age.

```bash
NOW="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

curl -i \
  -X POST "http://localhost:8086/api/v1/sessions/events" \
  -H "Content-Type: application/json" \
  -H "User-Agent: Mozilla/5.0 (Linux; Android 14; Pixel 7) AppleWebKit/537.36 Chrome/125.0 Mobile Safari/537.36" \
  -H "X-Client-Channel: user_app_web" \
  -H "X-Device-Type: mobile" \
  --data "{
    \"event_type\": \"click\",
    \"anonymous_id\": \"anon_heatmap_doc_test\",
    \"session_id\": \"sess_heatmap_doc_test\",
    \"path\": \"/products/prod_heatmap_doc\",
    \"occurred_at\": \"${NOW}\",
    \"properties\": {
      \"element_id\": \"add-to-cart\",
      \"x\": 210,
      \"y\": 600,
      \"viewport_width\": 390,
      \"viewport_height\": 844
    }
  }"
```

Send scroll event:

```bash
NOW="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

curl -i \
  -X POST "http://localhost:8086/api/v1/sessions/events" \
  -H "Content-Type: application/json" \
  -H "User-Agent: Mozilla/5.0 (Linux; Android 14; Pixel 7) AppleWebKit/537.36 Chrome/125.0 Mobile Safari/537.36" \
  -H "X-Client-Channel: user_app_web" \
  -H "X-Device-Type: mobile" \
  --data "{
    \"event_type\": \"scroll\",
    \"anonymous_id\": \"anon_heatmap_doc_test\",
    \"session_id\": \"sess_heatmap_doc_test\",
    \"path\": \"/products/prod_heatmap_doc\",
    \"occurred_at\": \"${NOW}\",
    \"properties\": {
      \"depth_percent\": 76,
      \"viewport_width\": 390,
      \"viewport_height\": 844,
      \"document_height\": 2200
    }
  }"
```

Expected response:

```text
HTTP/1.1 202 Accepted
```

### 10.2 Wait For Worker

Default worker interval:

```env
SESSION_HEATMAP_AGGREGATION_INTERVAL=1m
```

Wait around 60 seconds after sending events. For faster local feedback, temporarily set:

```env
SESSION_HEATMAP_AGGREGATION_INTERVAL=10s
SESSION_HEATMAP_AGGREGATION_INITIAL_LOOKBACK=2h
```

Then restart the service.

### 10.3 Verify MongoDB Heatmap Data

```bash
mongosh "mongodb://localhost:27017/session_db" --eval 'db.heatmap_points.find({path:"/products/prod_heatmap_doc"}).pretty()'
```

Expected idea:

```text
heatmap_type: click or scroll
device_type: mobile
weight: 1 or higher
x/y: percentage bucket values
```

Verify checkpoint:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval 'db.heatmap_checkpoints.findOne({worker_name:"heatmap_aggregator"})'
```

Verify unique session marker:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval 'db.heatmap_bucket_sessions.findOne({session_id:"sess_heatmap_doc_test"})'
```

### 10.4 Call Heatmap API

Use admin role header:

```bash
TODAY="$(date -u +"%Y-%m-%d")"

curl -i -G "http://localhost:8086/api/v1/analytics/heatmaps" \
  -H "X-User-Roles: admin" \
  --data-urlencode "path=/products/prod_heatmap_doc" \
  --data-urlencode "device_type=mobile" \
  --data-urlencode "from=${TODAY}" \
  --data-urlencode "to=${TODAY}" \
  --data-urlencode "heatmap_type=click"
```

Expected response shape:

```json
{
  "path": "/products/prod_heatmap_doc",
  "device_type": "mobile",
  "heatmap_type": "click",
  "from": "2026-05-22",
  "to": "2026-05-22",
  "points": [
    {
      "x": 55,
      "y": 70,
      "weight": 1
    }
  ],
  "max_weight": 1,
  "total_events": 1
}
```

For scroll heatmap:

```bash
TODAY="$(date -u +"%Y-%m-%d")"

curl -i -G "http://localhost:8086/api/v1/analytics/heatmaps" \
  -H "X-User-Roles: admin" \
  --data-urlencode "path=/products/prod_heatmap_doc" \
  --data-urlencode "device_type=mobile" \
  --data-urlencode "from=${TODAY}" \
  --data-urlencode "to=${TODAY}" \
  --data-urlencode "heatmap_type=scroll"
```

### 10.5 Ports & Networking

| Service | Port | Purpose | New in Task 6? |
|---|---:|---|---:|
| Session Service HTTP | `8086` | Event ingest, heatmap API, health, readiness | No |
| MongoDB | `27017` | Raw events and heatmap aggregate storage | No |
| Redis | `6379` | Existing active session/cache dependency | No |
| Heatmap worker | No separate port | In-process background goroutine | Yes |
| Kafka/RabbitMQ | N/A | Not required for Task 6 local setup | No |
| Heatmap dashboard frontend | N/A | Not implemented in current backend task | No |

Networking notes:

- Host-run service uses `localhost` for MongoDB/Redis.
- Container-run service should use Compose service names like `mongo` and `redis`.
- Heatmap API query path should be URL-encoded. `curl -G --data-urlencode` handles this.
- Admin headers should come from trusted gateway in production.

---

## 11. Common Errors & Fixes

Previous common setup errors are already documented:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
12. Common Errors & Fixes

TaskImplementation/Session Management Service/task3_Dependency.md
Section:
11. Common Errors & Fixes

TaskImplementation/Session Management Service/task4_Dependency.md
Section:
11. Common Errors & Fixes

TaskImplementation/Session Management Service/task5_Dependency.md
Section:
11. Common Errors & Fixes
```

Task 6-specific issues:

### 1. Heatmap API Returns Empty `points`

Common causes:

- No raw `click` or `scroll` events exist for that `path`.
- Query `device_type` does not match stored event device type.
- Worker has not run yet.
- Event `occurred_at` is outside worker initial lookback.
- `SESSION_HEATMAP_AGGREGATION_ENABLED=false`.
- You queried `heatmap_type=click`, but only scroll events exist.

Debug:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval 'db.session_events.find({path:"/products/prod_heatmap_doc", event_type:{$in:["click","scroll"]}}).pretty()'
mongosh "mongodb://localhost:27017/session_db" --eval 'db.heatmap_points.find({path:"/products/prod_heatmap_doc"}).pretty()'
mongosh "mongodb://localhost:27017/session_db" --eval 'db.heatmap_checkpoints.find().pretty()'
```

Fix:

- Send fresh events with current timestamp.
- Wait for worker interval.
- Use correct `device_type`.
- Temporarily increase `SESSION_HEATMAP_AGGREGATION_INITIAL_LOOKBACK=2h`.

### 2. `400 VALIDATION_ERROR` From Heatmap API

Common causes:

| Cause | Fix |
|---|---|
| Missing `path` | Add `path=/products/prod_123` |
| Missing `device_type` | Add `device_type=mobile` or stored type |
| Missing `from` or `to` | Use `YYYY-MM-DD` |
| Invalid date format | Use `2026-05-22`, not full timestamp |
| `from` after `to` | Reverse/correct date range |
| Date range too large | Keep within `SESSION_HEATMAP_MAX_DATE_RANGE_DAYS` |
| Invalid `heatmap_type` | Use `click` or `scroll` |

Example valid request:

```bash
curl -G "http://localhost:8086/api/v1/analytics/heatmaps" \
  -H "X-User-Roles: admin" \
  --data-urlencode "path=/products/prod_123" \
  --data-urlencode "device_type=mobile" \
  --data-urlencode "from=2026-05-22" \
  --data-urlencode "to=2026-05-22"
```

### 3. `401 AUTHENTICATION_REQUIRED`

Heatmap API admin-only hai. Local request me admin role header missing hai.

Fix:

```bash
curl -H "X-User-Roles: admin" ...
```

Reference:

```text
TaskImplementation/Session Management Service/task4_Dependency.md
Section:
11. Common Errors & Fixes -> 1. 401 AUTHENTICATION_REQUIRED
```

### 4. `403 PERMISSION_DENIED`

Role header present hai, but role admin allowed list me nahi hai.

Allowed:

```text
admin
superadmin
operations_admin
```

Fix:

```bash
curl -H "X-User-Roles: operations_admin" ...
```

### 5. Service Startup Fails With Heatmap Config Error

Examples:

```text
SESSION_HEATMAP_CLICK_BUCKET_SIZE must be between 1 and 100
SESSION_HEATMAP_AGGREGATION_INTERVAL must be greater than zero
SESSION_HEATMAP_AGGREGATION_WORKER_NAME cannot be empty
```

Fix:

Use safe defaults:

```env
SESSION_HEATMAP_CLICK_BUCKET_SIZE=5
SESSION_HEATMAP_AGGREGATION_INTERVAL=1m
SESSION_HEATMAP_AGGREGATION_BATCH_SIZE=1000
SESSION_HEATMAP_AGGREGATION_INITIAL_LOOKBACK=1h
SESSION_HEATMAP_AGGREGATION_CHECKPOINT_LOOKBACK=5m
SESSION_HEATMAP_AGGREGATION_WORKER_NAME=heatmap_aggregator
```

### 6. `heatmap_points` Collection Missing

Cause:

Migration `004_heatmap_points.up.js` not run and app has not created indexes yet.

Fix:

```bash
cd backend/services/session-service
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/004_heatmap_points.up.js
```

### 7. Worker Logs `session.heatmap.event_skipped`

Cause:

Worker found invalid raw event. Examples:

- Click event missing `x`, `y`, `viewport_width`, or `viewport_height`.
- Click coordinates outside viewport.
- Scroll event missing `depth_percent`.
- Path missing or does not start with `/`.

Fix:

Check raw event:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval 'db.session_events.find({event_type:{$in:["click","scroll"]}}).sort({occurred_at:-1}).limit(5).pretty()'
```

Send valid event payload as shown in section `10.1`.

### 8. Duplicate Or Inflated `weight`

Current worker uses checkpoint lookback overlap. This can catch late events, but repeated processing can increment `weight` and `sample_events` again because there is no processed-event ledger yet.

Short-term local fix:

- For clean manual test, clear heatmap docs for your test path.
- Avoid repeatedly restarting service with a large lookback during same test.

Example cleanup for only test path:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval 'db.heatmap_points.deleteMany({path:"/products/prod_heatmap_doc"}); db.heatmap_bucket_sessions.deleteMany({session_id:"sess_heatmap_doc_test"});'
```

Professional fix suggestion:

Add processed event markers or idempotent per-event aggregation so reprocessing does not double-count `weight`.

### 9. Scroll Heatmap Query Looks Empty But Click Works

Cause:

Default heatmap type is `click`. Scroll data requires explicit query:

```text
heatmap_type=scroll
```

Fix:

```bash
curl -G "http://localhost:8086/api/v1/analytics/heatmaps" \
  -H "X-User-Roles: admin" \
  --data-urlencode "path=/products/prod_heatmap_doc" \
  --data-urlencode "device_type=mobile" \
  --data-urlencode "from=$(date -u +"%Y-%m-%d")" \
  --data-urlencode "to=$(date -u +"%Y-%m-%d")" \
  --data-urlencode "heatmap_type=scroll"
```

---

## 12. Security & Best Practices

### 12.1 Security Audit For Task 6

| Area | Current status | Risk | Suggested fix |
|---|---|---|---|
| Admin API auth | Uses trusted role headers | Unsafe if exposed directly to internet | Put behind API Gateway/Auth service |
| Heatmap aggregate privacy | `heatmap_points` has no raw IP/user IDs | Good | Keep aggregate-only response |
| Unique session markers | `heatmap_bucket_sessions` stores `session_id` | More sensitive than aggregate points | Restrict DB access and define retention |
| SDK click capture | Backend validates coordinates but cannot know sensitive DOM context | Sensitive elements could be tracked if frontend sends them | SDK must skip private fields/buttons |
| Worker idempotency | Checkpoint overlap can reprocess events | Inflated weights after restart/lookback overlap | Add processed-event marker or idempotent event ledger |
| Retention | No heatmap aggregate retention finalized in Task 6 | Long-lived analytics data may violate policy | Define TTL/retention in later retention task |
| Docker artifacts | No official service Dockerfile/compose | New devs may use inconsistent local setup | Promote previous Docker example or add official compose later |
| `.env` loading | App does not auto-load `.env` | Beginners think config is ignored | Use `set -a; . ./.env; set +a` |
| Secrets | `.env.example` has placeholder salts | Bad if copied to production | Generate strong per-env secrets |
| Health checks | `/readyz` checks Mongo/Redis, not heatmap freshness | Worker could be stale while readiness passes | Add heatmap worker lag metric/check later |
| gRPC | `task6.md` mentions `SessionService.GetHeatmap` | Current repo has HTTP only, no gRPC implementation detected | Document HTTP as current runtime contract |

### 12.2 Heatmap Privacy Rules

Follow these rules strictly:

- Do not store password, OTP, card, email, phone, address, or form input values in event properties.
- Do not store visible text from private UI elements.
- Prefer `element_id`, `component`, or hashed labels instead of raw text.
- Add frontend skip rules like `data-analytics-private="true"`.
- Never expose `heatmap_bucket_sessions` through public/admin API.
- Use route/device/day aggregate views, not individual user replay.

### 12.3 Beginner Best Practices

Only Task 6-specific best practices are listed here. General Go/Mongo/Redis best practices are already in previous dependency files.

| Practice | Why it matters |
|---|---|
| Keep `SESSION_HEATMAP_CLICK_BUCKET_SIZE=5` for MVP | Good balance between detail and MongoDB cardinality |
| Do not change bucket size without rebuild plan | Old and new aggregates become inconsistent |
| Always query with correct `device_type` | Mobile and desktop layouts are different |
| Keep worker batch size reasonable | Avoid large MongoDB memory/CPU spikes |
| Use current timestamps in manual test events | Ingest max event age can reject old samples |
| Run migration `004` before manual verification | Validators and indexes keep data consistent |
| Use admin headers only locally | Production should rely on trusted gateway |
| Monitor skipped events | Skips usually mean SDK payload is bad |
| Keep raw event TTL longer than rebuild window | Aggregates can only be rebuilt while raw events exist |

---

## 13. Missing or Misconfigured Things

Task 6-specific gaps noticed while inspecting current implementation and docs:

| Missing / Misconfigured | Impact | Recommendation |
|---|---|---|
| `task6.md` says backend files were not created, but repo now has heatmap code/migration | Docs can confuse new developers | Update implementation doc later or keep this dependency guide as current-state source |
| No official migration runner | Beginners must run `mongosh` scripts manually | Add Makefile/script later |
| No official service Dockerfile/compose | Containerized setup is not standardized | Add service Dockerfile and compose when deployment scope arrives |
| No one-shot aggregation CLI | Manual local aggregation depends on running server/worker | Add admin CLI or internal endpoint only if needed |
| No event-level processed ledger | Checkpoint overlap can double-count `weight` | Add idempotent event marker design |
| Heatmap retention not finalized | Aggregates and session markers may grow | Add retention/TTL policy in Task 8 |
| No dashboard frontend package | `heatmap.js`/D3 are conceptual only | Add frontend dependency docs when dashboard task is implemented |
| No gRPC heatmap handler detected | `task6.md` mentions internal gRPC method | Keep HTTP as current API, add gRPC docs only after implementation |
| Heatmap freshness not in `/readyz` | Worker can silently lag | Add metric/check for checkpoint lag later |

---

## 14. References to Previous Dependency Files

Use these instead of duplicating setup:

| Topic | Refer |
|---|---|
| Full project overview and base setup | `TaskImplementation/Session Management Service/task1_Dependency.md`, sections `1`, `3`, `9`, `10` |
| Go modules, `go.mod`, `go.sum`, `go mod tidy`, `go run`, `go build` | `task1_Dependency.md`, section `4. Dependency Management` |
| MongoDB install, Docker run/compose, connection string, credentials | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| Redis install, Docker run/compose, connection string, credentials | `task1_Dependency.md`, section `5.2 Redis Setup` |
| Base Docker compose and service Dockerfile suggestion | `task1_Dependency.md`, section `8. Docker Setup` |
| MongoDB/Redis storage strategy and reserved `heatmap_points` note | `task2_Dependency.md`, sections `5` and `6` |
| Event ingestion endpoint and valid event troubleshooting | `task3_Dependency.md`, sections `7`, `10`, `11` |
| Admin API auth headers and journey admin errors | `task4_Dependency.md`, sections `6.3`, `11` |
| Device tracking, GeoIP, trusted proxy, privacy hashing | `task5_Dependency.md`, sections `6`, `7`, `11`, `12` |

What this file adds newly:

- Heatmap-specific env variables
- `004_heatmap_points.up.js` migration
- `heatmap_points`, `heatmap_bucket_sessions`, `heatmap_checkpoints`
- Background heatmap worker setup
- Heatmap API run and debug commands
- Task 6 security gaps and best practices

---

## 15. Final Checklist

Before saying Task 6 setup is complete, verify:

| Check | Done |
|---|---|
| Previous Task 1-5 dependency docs reviewed |  |
| Go dependencies downloaded with `go mod download` |  |
| MongoDB running on `27017` or configured URI |  |
| Redis running on `6379` or configured address |  |
| `.env` exists at `backend/services/session-service/.env` |  |
| Base MongoDB/Redis variables configured |  |
| Task 6 heatmap env variables added |  |
| `.env` loaded before running service |  |
| Migrations `001`, `002`, `003`, and `004_heatmap_points.up.js` applied |  |
| `heatmap_points` collection exists |  |
| `heatmap_bucket_sessions` collection exists |  |
| `heatmap_checkpoints` collection exists |  |
| Heatmap indexes visible in MongoDB |  |
| `go test ./...` passes |  |
| Backend starts on `:8086` |  |
| `/healthz` returns `200` |  |
| `/readyz` shows MongoDB and Redis ok |  |
| Fresh `click` event returns `202 Accepted` |  |
| Fresh `scroll` event returns `202 Accepted` |  |
| Worker has run after configured interval |  |
| `heatmap_points` contains test path data |  |
| `GET /api/v1/analytics/heatmaps` works with `X-User-Roles: admin` |  |
| Empty heatmap debug steps understood |  |
| Production admin headers are planned behind trusted gateway |  |
| Sensitive frontend elements are excluded from click tracking |  |

Final note:

Task 6 ka setup mostly MongoDB aggregation setup hai. New developer ko sabse zyada dhyan in cheezon par dena chahiye: migration `004`, fresh click/scroll events, worker interval/lookback, correct `device_type`, and admin role header.
