# Project Dependency & Setup Guide

This guide is for:

```text
TaskImplementation/Session Management Service/task2.md
```

Output file:

```text
TaskImplementation/Session Management Service/task2_Dependency.md
```

Goal simple hai: Task 2 me Session Management Service ke liye **MongoDB plus Redis** storage decision document hua hai. Is dependency guide me sirf setup, dependency, environment, database, Redis, Docker, DevOps, and debugging information hai. Business logic ya API implementation yahan rewrite nahi ki gayi.

Important reuse rule: same service folder me already full setup guide available hai:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
```

Isliye common Go install, Docker install, MongoDB install, Redis install, `.env` loading, run commands, migrations, and common troubleshooting ko yahan repeat nahi kiya gaya. Jahan setup same hai, wahan exact reference diya gaya hai.

---

## 1. Project Overview

Task 2 ka purpose storage choice finalize karna hai:

| Storage | Role | Required? |
|---|---|---:|
| MongoDB | Durable session, raw event, journey/analytics data store | Yes |
| Redis | Fast active session cache, last-seen state, live counters | Yes |

Simple Hinglish:

MongoDB long-term record book jaisa hai. Session history, events, journey, and analytics future me MongoDB me store honge. Redis fast memory cache jaisa hai. Active sessions and live counters quickly read/write karne ke liye Redis use hota hai.

Current repository me `backend/services/session-service` Go service already MongoDB and Redis clients use karta hai. Full current local setup is already explained in:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Sections:
- 1. Project Overview
- 8. Docker Setup
- 9. Local Development Setup
- 10. Running the Project
```

Task 2 dependency focus:

- MongoDB database name and collection expectations
- MongoDB indexes and TTL behavior
- Redis key prefix and active-session TTL
- Task-specific environment variables
- Storage-specific common errors
- Security and production configuration notes

---

## 2. Tech Stack

### Task 2 Detected Technologies

| Technology | What it is | Why used in this project | Required? | Reuse reference |
|---|---|---|---:|---|
| Go | Compiled backend language | Session service runtime is Go | Yes | `task1_Dependency.md`, section `2. Tech Stack` |
| Go `net/http` | Built-in Go HTTP server package | Health/API endpoints run with standard library HTTP | Yes | `task1_Dependency.md`, section `2. Tech Stack` |
| MongoDB | NoSQL document database | Flexible high-volume session/event documents | Yes | Detailed below, install reused |
| Redis | In-memory cache/data store | Active sessions, sliding TTL, hot counters | Yes | Detailed below, install reused |
| MongoDB Go Driver | Go client library for MongoDB | Go service connects to MongoDB collections/indexes | Yes in runtime service | `task1_Dependency.md`, section `4. Dependency Management` |
| go-redis/v9 | Go client library for Redis | Go service reads/writes Redis keys, hashes, sets, TTL | Yes in runtime service | `task1_Dependency.md`, section `4. Dependency Management` |
| mongosh | MongoDB shell | Run migrations and verify collections/indexes | Recommended | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| redis-cli | Redis command line client | Verify Redis and inspect keys/TTL | Recommended | `task1_Dependency.md`, section `5.2 Redis Setup` |
| Docker | Container runtime | Easiest way to run MongoDB and Redis locally | Optional but recommended | `task1_Dependency.md`, section `8. Docker Setup` |
| Kafka/RabbitMQ | Message queue/event streaming | Future `session.events` async flow only | No for Task 2/current local run | `task1_Dependency.md`, section `6. Redis / Queue / External Services` |

### MongoDB Beginner Explanation

MongoDB ek document database hai. SQL tables/rows ke instead ye JSON-like documents store karta hai. Session events ka shape flexible hota hai, jaise `page_view`, `click`, `scroll`, `search`, and `add_to_cart` events me different properties ho sakti hain. Isliye MongoDB useful hai.

### Redis Beginner Explanation

Redis ek in-memory data store hai. Ye RAM me data rakhta hai, isliye read/write bahut fast hota hai. Active session ko baar-baar update karna padta hai, so Redis is perfect for short-lived live state.

### Why Not MySQL For Task 2 Data

MySQL platform ke auth/relational data ke liye useful ho sakta hai, but raw analytics/session events ke liye Task 2 me MongoDB choose hua because:

- event payload flexible hai
- writes high-volume ho sakte hain
- TTL cleanup useful hai
- analytics queries document/event shape ke saath evolve kar sakti hain

MySQL setup is not required for this Session Management Service task.

---

## 3. Required Software

Task 2 ke liye required software same hai jo previous dependency guide me explain hai.

| Software | Required? | Purpose | Setup reference |
|---|---:|---|---|
| Git | Yes | Repository clone karne ke liye | `task1_Dependency.md`, section `3. Required Software` |
| Go 1.26.3 or newer | Yes | Session service run/test/build | `task1_Dependency.md`, section `3. Required Software` |
| MongoDB | Yes | Durable session/event store | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| mongosh | Recommended | Migrations/index verification | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| Redis | Yes | Active session store | `task1_Dependency.md`, section `5.2 Redis Setup` |
| redis-cli | Recommended | Redis verification/debugging | `task1_Dependency.md`, section `5.2 Redis Setup` |
| Docker | Optional | Local MongoDB/Redis containers | `task1_Dependency.md`, section `8. Docker Setup` |
| curl/Postman | Recommended | API/health testing | `task1_Dependency.md`, section `9. Local Development Setup` |

Do not repeat install steps here. Follow:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Sections:
- 3. Required Software
- 5. Database Setup
- 8. Docker Setup
```

---

## 4. Dependency Management

This is a Go project. Dependency management is already fully explained in:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
4. Dependency Management
```

Task 2 storage-related Go dependencies currently detected:

| Dependency | Version | Purpose |
|---|---:|---|
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | MongoDB connection, collection access, indexes |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Redis connection, hashes, sets, counters, TTL |

Important files:

| File | Meaning |
|---|---|
| `backend/services/session-service/go.mod` | Direct dependencies and Go version |
| `backend/services/session-service/go.sum` | Dependency checksum lock file |
| `backend/go.work` | Go workspace file for backend services |

Use these commands from the service folder:

```bash
cd backend/services/session-service
go mod download
go mod tidy
go test ./...
go run ./cmd/server
```

Hinglish note:

`go.mod` dependency list hai. `go.sum` checksum/security proof jaisa hai. `go mod tidy` missing dependencies add karta hai and unused dependencies remove karta hai.

Common Go module errors are already covered in:

```text
task1_Dependency.md
Sections:
- 4. Dependency Management
- 12. Common Errors & Fixes
```

---

## 5. Database Setup

Detected databases for Task 2:

| Database | Required? | Current role |
|---|---:|---|
| MongoDB | Yes | Durable sessions, raw events, journey summaries, future analytics |
| Redis | Yes | Active sessions, last seen state, user/anonymous active-session sets, live counters |
| MySQL | No | Not used by Session Service Task 2 |
| PostgreSQL | No | Not detected |
| SQLite | No | Not detected |

### 5.1 MongoDB Setup

#### A. What MongoDB Is

MongoDB ek NoSQL document database hai. Isme data collections ke andar JSON-like documents ke form me store hota hai.

Simple English: MongoDB stores flexible documents instead of fixed SQL rows.

#### B. Why This Project Uses MongoDB

Session events flexible and high-volume hote hain. Example:

- `page_view` has page/path data
- `click` has element and coordinate data
- `search` has query/filter data
- `add_to_cart` has product/quantity data

In sabko fixed SQL columns me force karna messy ho sakta hai. MongoDB me `properties` object flexible reh sakta hai.

#### C. Required Or Optional

MongoDB is required. Current service startup MongoDB ping karta hai. MongoDB down hoga to backend ready/start behavior fail ho sakta hai.

#### D. Local Installation

MongoDB local installation is already explained in detail here:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
5.1 MongoDB Setup
```

Follow that section for:

- Windows setup
- Linux setup
- macOS setup
- Docker setup
- `mongosh` verification

#### E. Docker Setup

Same Docker setup as previous guide. Do not duplicate.

Refer:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Sections:
- 5.1 MongoDB Setup
- 8. Docker Setup
- 9. Local Development Setup, Step 4
```

Quick reference only:

```bash
docker start ecommerce-session-mongo
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
```

#### F. Start Commands

Use the same start commands from previous guide:

```text
task1_Dependency.md
Section:
9. Local Development Setup, Step 4: Start MongoDB and Redis
```

#### G. Verify Running

```bash
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
```

Expected successful ping contains:

```text
ok: 1
```

#### H. Default Port

| Setting | Value |
|---|---|
| MongoDB default port | `27017` |
| Local URI | `mongodb://localhost:27017` |
| Docker Compose URI from another container | `mongodb://mongo:27017` |

#### I. Connection String Format

Local development:

```env
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_MONGO_DATABASE=session_db
```

Docker Compose internal networking:

```env
SESSION_MONGO_URI=mongodb://mongo:27017
SESSION_MONGO_DATABASE=session_db
```

Production style with credentials:

```env
SESSION_MONGO_URI=mongodb://<username>:<password>@<host>:27017/session_db?authSource=admin
SESSION_MONGO_DATABASE=session_db
```

Warning: Real username/password Git me commit nahi karna.

#### J. Where To Place Credentials

Credentials/config current service me environment variables se aate hain.

Create local file:

```text
backend/services/session-service/.env
```

Use example file:

```text
backend/services/session-service/.env.example
```

MongoDB variables:

| Variable | Example | Purpose |
|---|---|---|
| `SESSION_MONGO_URI` | `mongodb://localhost:27017` | MongoDB host/port/auth connection string |
| `SESSION_MONGO_DATABASE` | `session_db` | Database name |
| `SESSION_MONGO_CONNECT_TIMEOUT` | `5s` | Mongo connect/ping timeout |
| `SESSION_RAW_EVENT_TTL_DAYS` | `90` | Raw event TTL retention window |

Note: `.env` is not auto-loaded by current Go code. Load it before running. Refer:

```text
task1_Dependency.md
Section:
7. Environment Variables
```

### 5.2 MongoDB Database And Collections

Task 2 storage decision expects this database:

```text
session_db
```

Current/future collections:

| Collection | Purpose | Task 2 status |
|---|---|---|
| `sessions` | One durable session summary per visit/session | Required |
| `session_events` | Raw tracked events for timeline/analytics | Required |
| `journey_summaries` | Derived journey summary | Implemented later, already present in current repo migrations |
| `heatmap_points` | Future heatmap aggregate buckets | Reserved |
| `analytics_aggregates` | Future dashboard aggregates | Reserved |

Beginner note:

MongoDB me database ke andar collections hoti hain. Collections roughly SQL tables jaisi lag sakti hain, but documents flexible hote hain.

### 5.3 Required MongoDB Indexes

Task 2 documented these important indexes.

For `sessions`:

| Index | Purpose |
|---|---|
| `uniq_session_id` on `{ session_id: 1 }` | Same session duplicate na ho |
| `idx_anonymous_sessions` on `{ anonymous_id: 1, started_at: -1 }` | Guest visitor session history fast |
| `idx_user_sessions` on `{ user_id: 1, started_at: -1 }` | Logged-in user sessions fast |
| `idx_status_last_seen` on `{ status: 1, last_seen_at: -1 }` | Active/stale sessions identify karna |

For `session_events`:

| Index | Purpose |
|---|---|
| `idx_session_timeline` on `{ session_id: 1, occurred_at: 1 }` | Session journey timeline |
| `idx_user_event_history` on `{ user_id: 1, occurred_at: -1 }` | User event history |
| `idx_event_type_time` on `{ event_type: 1, occurred_at: -1 }` | Event type analytics |
| `ttl_raw_events_90_days` on `{ occurred_at: 1 }` | Old raw events cleanup |
| `idx_session_events_request_id` on `{ request_id: 1 }` | Current repo adds request tracing/idempotency lookup support |

TTL detail:

```text
90 days = 7776000 seconds
```

MongoDB TTL indexes run in background and deletion is not instant. Dashboard queries should still filter by time/status.

### 5.4 Migrations

Current migration files:

| File | Purpose |
|---|---|
| `backend/services/session-service/migrations/001_session_storage_indexes.up.js` | Create/update `sessions` and `session_events` validators and indexes |
| `backend/services/session-service/migrations/001_session_storage_indexes.down.js` | Rollback storage indexes/validators |
| `backend/services/session-service/migrations/002_journey_summaries.up.js` | Create `journey_summaries` collection/indexes |
| `backend/services/session-service/migrations/002_journey_summaries.down.js` | Rollback journey summaries |

Migration run steps are already explained:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
9. Local Development Setup, Step 8: Run MongoDB migrations
```

Task 2 specific verification:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.sessions.getIndexes()"
mongosh "mongodb://localhost:27017/session_db" --eval "db.session_events.getIndexes()"
```

Look for:

```text
uniq_session_id
idx_session_timeline
ttl_raw_events_90_days
```

---

## 6. Redis / Queue / External Services

### 6.1 Redis

#### A. What Redis Is

Redis ek fast in-memory data store hai. Ye keys, hashes, sets, counters, and expiry/TTL support karta hai.

Simple English: Redis keeps hot data in memory so reads/writes are very fast.

#### B. Why This Project Uses Redis

Task 2 me Redis active sessions ke liye choose hua:

- active session snapshot fast update karna
- `last_seen_at` quickly store karna
- user/anonymous ID ke active session IDs track karna
- live counters maintain karna
- inactivity ke baad keys auto-expire karna

#### C. Required Or Optional

Redis is required for current local service startup because app dependencies ping Redis.

#### D. Local Installation

Redis install/setup already documented:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
5.2 Redis Setup
```

#### E. Docker Setup

Same as previous guide:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Sections:
- 5.2 Redis Setup
- 8. Docker Setup
- 9. Local Development Setup, Step 4
```

#### F. Start Commands

```bash
docker start ecommerce-session-redis
redis-cli ping
```

Expected:

```text
PONG
```

#### G. Default Port

| Setting | Value |
|---|---|
| Redis default port | `6379` |
| Current logical DB | `2` |
| Default key prefix | `session` |

#### H. Connection String / Address Format

Local development:

```env
SESSION_REDIS_ADDR=localhost:6379
SESSION_REDIS_PASSWORD=
SESSION_REDIS_DB=2
```

Docker Compose internal networking:

```env
SESSION_REDIS_ADDR=redis:6379
SESSION_REDIS_PASSWORD=
SESSION_REDIS_DB=2
```

Production:

```env
SESSION_REDIS_ADDR=<private-redis-host>:6379
SESSION_REDIS_PASSWORD=<strong-secret>
SESSION_REDIS_DB=2
```

#### I. Where To Place Credentials

Use:

```text
backend/services/session-service/.env
```

Redis variables:

| Variable | Example | Purpose |
|---|---|---|
| `SESSION_REDIS_ADDR` | `localhost:6379` | Redis host and port |
| `SESSION_REDIS_PASSWORD` | empty locally | Redis auth password |
| `SESSION_REDIS_DB` | `2` | Logical Redis DB |
| `SESSION_REDIS_KEY_PREFIX` | `session` | Prefix for all session keys |
| `SESSION_ACTIVE_TTL` or `SESSION_ACTIVE_TTL_SECONDS` | `35m` or `2100` | Active session key TTL |
| `SESSION_ACTIVE_COUNTER_TTL` | `48h` | Live counter key retention |

### 6.2 Redis Key Design

Task 2 documented this key pattern. Current code uses the same pattern with configurable prefix.

If:

```env
SESSION_REDIS_KEY_PREFIX=session
```

Then keys look like:

| Key pattern | Redis type | Purpose |
|---|---|---|
| `session:active:{session_id}` | Hash | Active session snapshot |
| `session:user:{user_id}` | Set | Active session IDs for a logged-in user |
| `session:anon:{anonymous_id}` | Set | Active session IDs for anonymous visitor |
| `session:last_seen:{session_id}` | String | Last heartbeat/activity timestamp |
| `session:counter:active:{yyyyMMddHH}` | String counter | Hourly active event/session counter |

Example debug commands:

```bash
redis-cli -n 2 keys 'session:*'
redis-cli -n 2 hgetall session:active:sess_local_001
redis-cli -n 2 ttl session:active:sess_local_001
redis-cli -n 2 smembers session:anon:anon_local_001
```

Warning:

Do not use `KEYS session:*` on production Redis with large data. Use `SCAN` in production.

### 6.3 TTL Rules

| Setting | Recommended/current value | Meaning |
|---|---:|---|
| Session inactivity timeout | `30m` | No activity after this means session can become inactive |
| Active Redis TTL | `35m` / `2100s` | Slight grace above inactivity timeout |
| Active counter TTL | `48h` | Short-lived live metrics |
| Raw event TTL in MongoDB | `90` days | Raw analytics event retention |

Hinglish:

Redis active key ka TTL sliding hai. Har valid activity/event pe TTL refresh hota hai. Redis expiry permanent history nahi hai; durable history MongoDB me rehni chahiye.

### 6.4 Kafka / RabbitMQ / Other Services

Kafka/RabbitMQ Task 2/current local setup ke liye required nahi hai.

Refer:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
6. Redis / Queue / External Services
```

Current status:

| Service | Required now? | Note |
|---|---:|---|
| Kafka | No | Future `session.events` stream |
| RabbitMQ | No | Alternative future queue |
| NATS | No | Not detected |
| MinIO/S3 | No | Not detected for Task 2 |
| SMTP/Twilio/Stripe/Firebase | No | Not detected |
| Nginx/API Gateway | Not local required | Production should route through gateway |

---

## 7. Environment Variables

Full `.env` documentation already exists:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
7. Environment Variables
```

Task 2 storage-specific variables are below.

### 7.1 Task 2 `.env` Example

Only MongoDB/Redis/storage variables:

```env
# MongoDB durable store
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_MONGO_DATABASE=session_db
SESSION_MONGO_CONNECT_TIMEOUT=5s
SESSION_RAW_EVENT_TTL_DAYS=90

# Redis active session store
SESSION_REDIS_ADDR=localhost:6379
SESSION_REDIS_PASSWORD=
SESSION_REDIS_DB=2
SESSION_REDIS_KEY_PREFIX=session
SESSION_ACTIVE_TTL=35m
SESSION_ACTIVE_COUNTER_TTL=48h

# Older compatibility variables also supported by code
SESSION_ACTIVE_TTL_SECONDS=2100
SESSION_ACTIVE_COUNTER_TTL_SECONDS=172800
```

Use the complete `.env` from previous guide for actual local run:

```text
task1_Dependency.md
Section:
7. Environment Variables -> Complete `.env` example
```

### 7.2 Important `.env` Loading Note

Current Go code uses `os.Getenv`. It does not automatically read `.env`.

Refer:

```text
task1_Dependency.md
Section:
7. Environment Variables -> Very important: `.env` is not auto-loaded
```

Linux/macOS quick reminder:

```bash
cd backend/services/session-service
set -a
source .env
set +a
go run ./cmd/server
```

### 7.3 Credentials Placement

| Secret/config | Place locally | Place in production |
|---|---|---|
| MongoDB URI/password | `backend/services/session-service/.env` | Secret manager / Kubernetes Secret |
| Redis password | `backend/services/session-service/.env` | Secret manager / Kubernetes Secret |
| IP hash salt | `backend/services/session-service/.env` | Secret manager / Kubernetes Secret |
| Database name | `.env` or ConfigMap | ConfigMap/environment |
| Redis key prefix | `.env` or ConfigMap | ConfigMap/environment |

Warning:

Do not commit real `.env` values. `.env.example` can contain safe examples only.

---

## 8. Docker Setup

Docker setup is same as previous dependency file. Do not create duplicate Docker instructions here.

Refer:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Sections:
- 8. Docker Setup
- 9. Local Development Setup, Step 4
```

Task 2 Docker-specific notes:

| Area | Local host Go run | Go inside Docker Compose |
|---|---|---|
| Mongo URI | `mongodb://localhost:27017` | `mongodb://mongo:27017` |
| Redis addr | `localhost:6379` | `redis:6379` |
| Mongo volume | `session_mongo_data` | Compose volume |
| Redis volume | `session_redis_data` | Compose volume |

Beginner warning:

Container ke andar `localhost` ka matlab same container hota hai. Agar Go service bhi container me run ho rahi hai, MongoDB/Redis ke Compose service names use karo.

### 8.1 Persistent Volumes

MongoDB volume is important because session/event data should survive container restart:

```text
session_mongo_data -> /data/db
```

Redis AOF volume is useful for local debugging, but Redis is still not source of truth:

```text
session_redis_data -> /data
```

Warning:

```bash
docker compose down -v
```

ye local MongoDB and Redis data delete kar sakta hai.

### 8.2 Missing Docker Artifacts

Current repo status from previous guide:

| Artifact | Status | Impact |
|---|---|---|
| Session service Dockerfile | Not committed | Official service image cannot be built yet |
| Session local docker-compose file | Not committed | Beginners must use documented local commands |
| MongoDB/Redis containers | Can be run manually | Enough for local development |

Refer:

```text
task1_Dependency.md
Section:
8. Docker Setup -> Current repo status
```

---

## 9. Local Development Setup

Use the full onboarding flow from:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
9. Local Development Setup
```

Task 2 incremental checklist:

1. Start MongoDB.
2. Start Redis.
3. Set MongoDB env variables.
4. Set Redis env variables.
5. Run MongoDB migrations.
6. Verify MongoDB indexes.
7. Verify Redis ping and key prefix.
8. Start Go service.
9. Check `/readyz` includes MongoDB and Redis as `ok`.

Commands:

```bash
cd backend/services/session-service
go mod download

set -a
source .env
set +a

mongosh "$SESSION_MONGO_URI" migrations/001_session_storage_indexes.up.js
mongosh "$SESSION_MONGO_URI" migrations/002_journey_summaries.up.js

go test ./...
go run ./cmd/server
```

Readiness check:

```bash
curl http://localhost:8086/readyz
```

Expected:

```json
{"checks":{"mongo":"ok","redis":"ok"},"status":"ok"}
```

---

## 10. Running the Project

Daily run flow is already documented:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
10. Running the Project
```

Task 2 storage-focused run flow:

```bash
docker start ecommerce-session-mongo ecommerce-session-redis
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
redis-cli ping

cd backend/services/session-service
set -a
source .env
set +a
go run ./cmd/server
```

Verify storage after sending a test event:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.sessions.findOne()"
mongosh "mongodb://localhost:27017/session_db" --eval "db.session_events.findOne()"
redis-cli -n 2 keys 'session:*'
```

Note:

Actual API test event example is already present in:

```text
task1_Dependency.md
Section:
9. Local Development Setup, Step 12: Send test event
```

---

### 10.1 Ports & Networking

| Service | Port | Purpose | Task 2 status |
|---|---:|---|---|
| Session Backend API | `8086` | HTTP APIs, health/readiness | Reused from previous setup |
| MongoDB | `27017` | Durable session/event DB | Required by Task 2 |
| Redis | `6379` | Active session cache | Required by Task 2 |
| Kafka/RabbitMQ | N/A | Future async `session.events` | Not required |
| API Gateway | Usually `8080` in platform docs | Production/front-door routing | Not required for local Task 2 |

Port conflict troubleshooting already exists:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
11. Ports & Networking
```

Task 2 connection rules:

| Scenario | MongoDB value | Redis value |
|---|---|---|
| Go runs on host | `SESSION_MONGO_URI=mongodb://localhost:27017` | `SESSION_REDIS_ADDR=localhost:6379` |
| Go runs in Compose | `SESSION_MONGO_URI=mongodb://mongo:27017` | `SESSION_REDIS_ADDR=redis:6379` |
| Production | private MongoDB host/URI | private Redis host with auth/TLS |

---

## 11. Common Errors & Fixes

Common setup errors are already documented:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section:
12. Common Errors & Fixes
```

Task 2 specific additions:

### 1. TTL Index Still Shows Old Retention

Cause:

`SESSION_RAW_EVENT_TTL_DAYS` changed, but existing MongoDB TTL index still has old `expireAfterSeconds`.

Check:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.session_events.getIndexes()"
```

Fix for local dev only:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.session_events.dropIndex('ttl_raw_events_90_days')"
mongosh "mongodb://localhost:27017/session_db" migrations/001_session_storage_indexes.up.js
```

Production note:

Retention changes should be handled as planned migrations, not random manual commands.

### 2. Redis Keys Not Visible

Cause:

Wrong Redis DB or key prefix.

Fix:

```bash
redis-cli -n 2 keys 'session:*'
redis-cli -n 2 scan 0 match 'session:*' count 20
```

Check `.env`:

```env
SESSION_REDIS_DB=2
SESSION_REDIS_KEY_PREFIX=session
```

### 3. Active Session Expires Too Early

Cause:

`SESSION_ACTIVE_TTL` is less than or equal to inactivity timeout, or env value was not loaded.

Fix:

```env
SESSION_INACTIVITY_TIMEOUT=30m
SESSION_ACTIVE_TTL=35m
```

Then reload `.env` and restart service.

### 4. MongoDB Database Looks Empty

Cause:

Service is writing to a different database name than the one you are checking.

Fix:

```bash
echo "$SESSION_MONGO_DATABASE"
mongosh "mongodb://localhost:27017/session_db" --eval "db.getCollectionNames()"
```

If `.env` was not sourced, service may use default `session_db`.

### 5. Redis Data Disappeared After Restart

Cause:

Redis is cache/hot state. If container has no volume/AOF or key TTL expired, keys can disappear.

Expected behavior:

Redis expiry should not mean durable data loss. MongoDB should still contain `sessions` and `session_events`.

Verify:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.sessions.countDocuments()"
mongosh "mongodb://localhost:27017/session_db" --eval "db.session_events.countDocuments()"
```

### 6. MongoDB `E11000 duplicate key` On `session_id`

Cause:

Same `session_id` inserted twice where unique index exists.

Meaning:

This is usually correct protection. Session upsert should update existing session instead of inserting duplicate session record.

Debug:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.sessions.find({session_id:'sess_local_001'}).toArray()"
```

### 7. `SESSION_REDIS_KEY_PREFIX cannot contain whitespace`

Cause:

Env value has spaces.

Bad:

```env
SESSION_REDIS_KEY_PREFIX=session dev
```

Good:

```env
SESSION_REDIS_KEY_PREFIX=session_dev
```

---

## 12. Security & Best Practices

### 12.1 Security Audit For Task 2

| Finding | Risk | Recommendation |
|---|---|---|
| MongoDB local URI has no username/password | Fine locally, unsafe in production | Use MongoDB auth, TLS, private networking |
| Redis password blank in local example | Fine locally, unsafe in production | Set `SESSION_REDIS_PASSWORD` in prod |
| Redis is cache, not durable DB | Active data can expire | Always write durable records to MongoDB |
| Raw IP must not be stored | Privacy/compliance risk | Store salted `ip_hash` only |
| Event properties can accidentally contain secrets | Password/card/token leakage risk | Sanitize future ingestion payloads |
| TTL retention can delete raw events | Analytics/compliance surprise | Align `SESSION_RAW_EVENT_TTL_DAYS` with Task 8 policy |
| `.env` is manual | Config may silently use defaults | Source `.env` or add a local runner |
| No official Dockerfile/compose committed | Onboarding/deploy inconsistency | Add service Dockerfile and local compose later |
| MongoDB migration tracking is manual | Hard to know which migrations ran | Add migration tracking collection/tool later |

### 12.2 Sensitive Data Rules

Never store these in MongoDB event properties or Redis fields:

```text
password
otp
card_number
cvv
pin
token
authorization
cookie
private_message
raw_ip
```

Hinglish:

Session analytics ka purpose user behavior samajhna hai, secrets capture karna nahi. Forms, passwords, OTP, card details, cookies, and tokens ko analytics payload me kabhi mat bhejo.

### 12.3 Beginner Best Practices

Task 2 specific:

- MongoDB ko source of truth rakho.
- Redis ko active/live state ke liye use karo.
- Redis key prefix environment-specific rakho, for example `session_dev`.
- TTL values carefully set karo.
- Raw event TTL change karne se pehle product/privacy requirement confirm karo.
- MongoDB indexes fresh DB par verify karo.
- Production me MongoDB and Redis ko public internet par expose mat karo.
- Use `/readyz` to confirm MongoDB and Redis both healthy.

General best practices already documented:

```text
task1_Dependency.md
Section:
13. Security & Best Practices
```

---

## 13. Missing or Misconfigured Things

Task 2/current repo ke setup perspective se:

| Missing / misconfigured | Current impact | Suggested action |
|---|---|---|
| Official session-service Dockerfile not committed | Service container build not standardized | Add Dockerfile before production |
| Official local docker-compose file not committed | Beginners depend on manual commands/docs | Add compose with MongoDB, Redis, healthchecks |
| No migration version tracking collection | Manual JS migrations can be rerun without history | Add migration tool or `schema_migrations` collection |
| Raw event TTL index name fixed as `ttl_raw_events_90_days` | Name can become misleading if retention changes | Use migration per retention change or generic index name |
| Redis persistence not production durability | Redis can lose hot state | Keep MongoDB durable writes mandatory |
| Queue publisher not wired | Future downstream analytics not async yet | Add Kafka/RabbitMQ only when consumer tasks exist |
| No production secret manager config shown | Secrets might end up in env files | Use Kubernetes Secrets/Vault/managed platform secrets |
| No metrics endpoint detected in setup docs | Harder production monitoring | Add Prometheus/OpenTelemetry later |

These are not blockers for local Task 2 storage validation.

---

## 14. References to Previous Dependency Files

Previous dependency file reused:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
```

| Topic | Reused section |
|---|---|
| Project overview and current service status | `1. Project Overview` |
| Go, net/http, MongoDB, Redis descriptions | `2. Tech Stack` |
| Required software list | `3. Required Software` |
| Go modules, `go.mod`, `go.sum`, commands | `4. Dependency Management` |
| MongoDB installation and Docker setup | `5.1 MongoDB Setup` |
| Redis installation and Docker setup | `5.2 Redis Setup` |
| Kafka/RabbitMQ not required note | `6. Redis / Queue / External Services` |
| Full `.env` example and loading instructions | `7. Environment Variables` |
| Docker Compose example | `8. Docker Setup` |
| Clone/install/migrate/run onboarding | `9. Local Development Setup` |
| Daily run commands | `10. Running the Project` |
| Ports and networking | `11. Ports & Networking` |
| Common errors | `12. Common Errors & Fixes` |
| Security and general best practices | `13. Security & Best Practices` |
| Missing production/devops items | `14. Missing or Misconfigured Things` |
| Fresh machine checklist | `15. Final Checklist` |

---

## 15. Final Checklist

Task 2 storage-specific checklist:

- [ ] Read `TaskImplementation/Session Management Service/task1_Dependency.md` for shared setup.
- [ ] Install/verify Go.
- [ ] Start MongoDB on `27017`.
- [ ] Start Redis on `6379`.
- [ ] Create `backend/services/session-service/.env` from `.env.example`.
- [ ] Set `SESSION_MONGO_URI`.
- [ ] Set `SESSION_MONGO_DATABASE=session_db`.
- [ ] Set `SESSION_REDIS_ADDR`.
- [ ] Set `SESSION_REDIS_DB=2`.
- [ ] Set `SESSION_REDIS_KEY_PREFIX=session`.
- [ ] Set `SESSION_ACTIVE_TTL=35m` or `SESSION_ACTIVE_TTL_SECONDS=2100`.
- [ ] Set `SESSION_RAW_EVENT_TTL_DAYS=90` unless product/privacy policy says otherwise.
- [ ] Load `.env` before running the Go service.
- [ ] Run MongoDB migration `001_session_storage_indexes.up.js`.
- [ ] Verify `sessions` indexes.
- [ ] Verify `session_events` indexes.
- [ ] Verify `ttl_raw_events_90_days` exists.
- [ ] Verify Redis with `redis-cli ping`.
- [ ] Run `go test ./...`.
- [ ] Start backend with `go run ./cmd/server`.
- [ ] Check `curl http://localhost:8086/readyz`.
- [ ] After test event, verify MongoDB durable records.
- [ ] After test event, verify Redis active keys.
- [ ] Confirm no real secrets are committed to Git.

End result:

A beginner developer should understand that Task 2 adds a clear storage architecture: **MongoDB for durable session/event data** and **Redis for active/live session state**, while the actual repeated installation and run instructions are reused from `task1_Dependency.md`.
