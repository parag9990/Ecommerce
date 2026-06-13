# Project Dependency & Setup Guide

## 1. Project Overview

This document is for:

```text
TaskImplementation/Session Management Service/task7.md
```

Output file:

```text
TaskImplementation/Session Management Service/task7_Dependency.md
```

Task 7 adds analytics read APIs for the Session Management Service:

| API | Purpose |
|---|---|
| `GET /api/v1/analytics/live` | Current active users, active sessions, and events-per-minute |
| `GET /api/v1/analytics/sessions` | Admin session list with filters and pagination |
| `GET /api/v1/analytics/sessions/{session_id}/journey` | Session journey / replay metadata |
| `GET /api/v1/analytics/funnels` | Funnel conversion and drop-off report |
| `GET /api/v1/analytics/heatmaps` | Heatmap read endpoint, reused from Task 6 |

Simple Hinglish explanation:

Task 7 ka kaam code logic rewrite karna nahi hai. Iska setup focus hai ki analytics APIs chalane ke liye local machine par kaunse services, env variables, database collections, indexes, migrations, and admin headers required hain. Pehle ke tasks ne base session storage, event ingestion, journey, device tracking, and heatmap setup already explain kar diya hai. Yaha sirf Task 7 ke new setup items detail me explain kiye gaye hain.

Important reuse rule:

Base Go, MongoDB, Redis, Docker, and local setup instructions already exist in older dependency files. Is file me unhe repeat nahi kiya gaya. Repeated setup ke liye exact reference sections diye gaye hain.

## 2. Tech Stack

### 2.1 Detected Technologies In Current Repo

| Technology | Required? | Used In Task 7? | Beginner Explanation |
|---|---:|---:|---|
| Go `1.26.3` | Yes | Yes | Go backend language hai. Is service ka HTTP server, usecase, repository, and config Go me implemented hai. |
| Go modules | Yes | Yes | `go.mod` and `go.sum` dependency versions lock karte hain. Isse sab developers same library versions use karte hain. |
| Standard library `net/http` | Yes | Yes | Ye Go ka built-in HTTP framework hai. Current code Gin/Fiber use nahi karta. |
| MongoDB | Yes | Yes | MongoDB document database hai. Task 7 analytics aggregates, sessions, events, journeys, and heatmaps read karta hai. |
| Redis | Yes | Yes | Redis fast in-memory store hai. Task 7 live active metrics Redis se read karta hai. |
| `go.mongodb.org/mongo-driver/v2` | Yes | Yes | Official MongoDB Go driver. Go code ko MongoDB se connect/query karne ke liye use hota hai. |
| `github.com/redis/go-redis/v9` | Yes | Yes | Redis client library. Active users/sessions and event counters read karne ke liye use hoti hai. |
| `log/slog` | Yes | Yes | Go structured logging. Analytics requests and storage failures logs me readable format me aate hain. |
| `mongosh` | Recommended | Yes | MongoDB shell. `005_analytics_aggregates.up.js` migration run and verify karne ke liye use hota hai. |
| Docker | Optional but recommended | Indirect | MongoDB and Redis ko locally easy run karne ke liye useful hai. Official service Dockerfile abhi committed nahi hai. |
| API Gateway / Admin headers | Required in real deployment | Yes | Analytics APIs admin-only hain. Local service headers se role check karta hai; production me gateway JWT validate karega. |
| gRPC | Not currently local dependency | Mentioned in docs | `task7.md` me internal gRPC method names mention hain, but current `go.mod` me gRPC dependency nahi hai and current runtime HTTP handlers expose karta hai. |
| Kafka/RabbitMQ | No | No | Task 7 current implementation me queue dependency nahi hai. Analytics aggregates precomputed data assume karte hain, but queue setup added nahi hai. |

### 2.2 Important Correction For Beginners

`task7.md` has some design-style examples that mention Gin/gRPC. Actual repo code for Task 7 currently uses:

```text
backend/services/session-service/internal/transport/http/handler.go
backend/services/session-service/internal/usecase/analytics.go
backend/services/session-service/internal/repository/mongo_analytics_repository.go
backend/services/session-service/internal/repository/redis_live_metrics_repository.go
```

So local setup ke liye Gin, Fiber, Echo, Express, NestJS, Django, or FastAPI install karne ki zarurat nahi hai.

## 3. Required Software

Task 7 does not add a new runtime like Node or Python. Required software mostly previous tasks jaisa hi hai.

| Software | Required? | Why Needed | Setup Reference |
|---|---:|---|---|
| Git | Yes | Repository clone karne ke liye | `task1_Dependency.md`, section `9. Local Development Setup` |
| Go `1.26.3` or compatible | Yes | Session service run/test/build karne ke liye | `task1_Dependency.md`, section `3. Required Software` |
| MongoDB | Yes | `analytics_aggregates`, `sessions`, `session_events`, `journey_summaries`, `heatmap_points` store/read karne ke liye | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| `mongosh` | Recommended | Task 7 migration and verification commands ke liye | `task1_Dependency.md`, section `5.1 MongoDB Setup` |
| Redis | Yes | Live analytics counters and active sessions ke liye | `task1_Dependency.md`, section `5.2 Redis Setup` |
| Docker Desktop / Docker Engine | Optional | MongoDB/Redis containers easily run karne ke liye | `task1_Dependency.md`, section `8. Docker Setup` |
| curl/Postman | Recommended | Analytics endpoints verify karne ke liye | Reused from previous task verification flow |

Do not reinstall same tools if you already followed Task 1 to Task 6 dependency docs.

## 4. Dependency Management

### 4.1 Go Dependency System

Full beginner explanation for these topics already exists:

| Topic | Existing Doc |
|---|---|
| `go.mod` | `TaskImplementation/Session Management Service/task1_Dependency.md`, section `4. Dependency Management` |
| `go.sum` | `TaskImplementation/Session Management Service/task1_Dependency.md`, section `4. Dependency Management` |
| Go modules | `TaskImplementation/Session Management Service/task1_Dependency.md`, section `4. Dependency Management` |
| `go mod tidy` | `TaskImplementation/Session Management Service/task1_Dependency.md`, section `4. Dependency Management` |
| `go mod download` | `TaskImplementation/Session Management Service/task1_Dependency.md`, section `4. Dependency Management` |
| `go build` | `TaskImplementation/Session Management Service/task1_Dependency.md`, section `4. Dependency Management` |
| `go run` | `TaskImplementation/Session Management Service/task1_Dependency.md`, section `4. Dependency Management` |
| Proxy/cache/version mismatch issues | `TaskImplementation/Session Management Service/task1_Dependency.md`, section `4. Dependency Management` and `12. Common Errors & Fixes` |

### 4.2 Task 7 Dependency Delta

Task 7 does not introduce new Go libraries in `go.mod`. It reuses existing dependencies:

| Dependency | Why Task 7 Uses It |
|---|---|
| `go.mongodb.org/mongo-driver/v2` | `analytics_aggregates` and raw `session_events` queries |
| `github.com/redis/go-redis/v9` | Live active user/session counters and events-per-minute |
| Go standard library | HTTP routes, query parsing, validation, logging, in-memory cache |

Quick commands from the service folder:

```bash
cd backend/services/session-service
go mod download
go test ./...
go run ./cmd/server
```

Note:

`go.mod` currently declares `go 1.26.3`. Agar local Go version older hai, `go test` or `go run` fail ho sakta hai. Pehle `go version` verify karo, then Go upgrade karo.

## 5. Database Setup

### 5.1 Database Summary

| Database | Required? | Task 7 Usage | Port |
|---|---:|---|---:|
| MongoDB | Yes | Sessions list, raw events fallback, journey metadata, heatmap read, analytics aggregate reports | `27017` |
| Redis | Yes | Live active users, active sessions, events-per-minute | `6379` |

### 5.2 MongoDB

#### A. What It Is

MongoDB ek document database hai. Isme rows/tables ke bajay JSON-like documents collections me store hote hain. Session analytics me events and aggregates ka shape flexible hota hai, isliye MongoDB useful hai.

Full MongoDB beginner setup already explained in:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section: 5.1 MongoDB Setup
```

#### B. Why Task 7 Uses MongoDB

Task 7 MongoDB se ye data read karta hai:

| Collection | Purpose |
|---|---|
| `sessions` | Session list API and session metadata |
| `session_events` | Raw fallback for funnel reports when aggregate missing and date range is short |
| `journey_summaries` | Session replay metadata summary |
| `heatmap_points` | Heatmap read API, introduced earlier in Task 6 |
| `analytics_aggregates` | New Task 7 collection for funnels, conversion, retention-style reports |

#### C. Required Or Optional

MongoDB is required. Without MongoDB:

| API | Expected Result |
|---|---|
| `/api/v1/analytics/sessions` | `503 ANALYTICS_STORAGE_UNAVAILABLE` |
| `/api/v1/analytics/funnels` | `503 ANALYTICS_STORAGE_UNAVAILABLE` or `409 AGGREGATE_NOT_READY` depending on state |
| `/api/v1/analytics/sessions/{id}/journey` | Journey storage errors |
| `/readyz` | Degraded or unavailable readiness |

#### D. Local Installation

Do not duplicate installation here.

Refer:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section: 5.1 MongoDB Setup
```

#### E. Docker Setup

Base MongoDB Docker setup is already explained here:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section: 8. Docker Setup
```

Task 7 does not need a new MongoDB container. It uses the same MongoDB instance and same database:

```text
SESSION_MONGO_DATABASE=session_db
```

#### F. Start Commands

Use the previous Docker/native start commands. Quick verification:

```bash
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
```

#### G. Verify Running

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.getCollectionNames()"
```

You should eventually see:

```text
sessions
session_events
journey_summaries
heatmap_points
analytics_aggregates
```

#### H. Default Port

```text
27017
```

#### I. Connection String Format

Local host machine:

```env
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_MONGO_DATABASE=session_db
```

Docker Compose service-to-service:

```env
SESSION_MONGO_URI=mongodb://mongo:27017
SESSION_MONGO_DATABASE=session_db
```

With username/password:

```env
SESSION_MONGO_URI=mongodb://session_user:strong_password@mongo:27017/session_db?authSource=admin
SESSION_MONGO_DATABASE=session_db
```

Credentials placement:

```text
backend/services/session-service/.env
```

Do not commit real database usernames/passwords.

#### J. Task 7 MongoDB Migration

Task 7 introduces:

```text
backend/services/session-service/migrations/005_analytics_aggregates.up.js
backend/services/session-service/migrations/005_analytics_aggregates.down.js
```

This migration creates/updates:

| Object | Why Needed |
|---|---|
| `analytics_aggregates` collection | Precomputed funnel, conversion, and retention aggregate docs |
| `uniq_metric_bucket_time_segment` index | Prevent duplicate metric/bucket/segment aggregate documents |
| `idx_metric_time` index | Fast metric + time range lookup |
| `idx_segment_time` index | Fast channel/device/date dashboard filters |
| `idx_retention_cohort` index | Retention/cohort queries |
| `idx_sessions_channel_started` index | Faster session list/report filtering by channel and start time |

Run all migrations required up to Task 7 from the service folder:

```bash
cd backend/services/session-service
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/001_session_storage_indexes.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/002_journey_summaries.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/003_device_tracking_indexes.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/004_heatmap_points.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/005_analytics_aggregates.up.js
```

If MongoDB is running inside Docker and `mongosh` is not installed on host, use the Mongo container method from:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section: 9. Local Development Setup -> Step 8: Run MongoDB migrations
```

Verify Task 7 collection and indexes:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.analytics_aggregates.getIndexes().map(i => i.name)"
mongosh "mongodb://localhost:27017/session_db" --eval "db.analytics_aggregates.findOne()"
```

Expected index names include:

```text
uniq_metric_bucket_time_segment
idx_metric_time
idx_segment_time
idx_retention_cohort
```

Important:

The service also calls `EnsureIndexes()` at startup. Migrations are still recommended because they define collection validators and give beginners an explicit repeatable setup path.

### 5.3 Task 7 Aggregate Data Requirement

Task 7 can read funnel reports from two places:

| Source | When Used |
|---|---|
| `analytics_aggregates` | Normal dashboard path |
| Raw `session_events` fallback | Only when aggregate is missing and requested range is within `SESSION_ANALYTICS_RAW_FALLBACK_RANGE` |

If you request a large funnel date range and no aggregate exists, API returns:

```text
409 AGGREGATE_NOT_READY
```

Simple explanation:

Analytics dashboard ko fast rakhne ke liye purane data par raw events scan nahi karna chahiye. Isliye historical dashboards ke liye aggregate documents precompute/seed karne honge.

Dev-only sample aggregate insert:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval '
db.analytics_aggregates.insertOne({
  _id: "dev_funnel_checkout_daily_2026_05_22_all",
  metric: "funnel.checkout",
  bucket: "daily",
  bucket_start: ISODate("2026-05-22T00:00:00Z"),
  bucket_end: ISODate("2026-05-23T00:00:00Z"),
  segment: {},
  steps: [
    { name: "product_view", event_type: "product_view", count: 100, unique_sessions: 80, unique_users: 70 },
    { name: "add_to_cart", event_type: "add_to_cart", count: 40, unique_sessions: 30, unique_users: 28 },
    { name: "checkout_step", event_type: "checkout_step", count: 20, unique_sessions: 15, unique_users: 14 },
    { name: "payment_result", event_type: "payment_result", count: 8, unique_sessions: 7, unique_users: 7 }
  ],
  schema_version: 1,
  calculated_at: new Date(),
  created_at: new Date(),
  updated_at: new Date()
})
'
```

Warning:

This is only for local dev testing. Production aggregates should be generated by a controlled analytics aggregator job.

## 6. Redis / Queue / External Services

### 6.1 Redis

Redis setup and Docker commands are already explained in:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section: 5.2 Redis Setup
```

Task 7 Redis usage:

| Redis Data | Purpose |
|---|---|
| `session:active_sessions` sorted set | Live active session count |
| `session:active_users` sorted set | Live active user count |
| `session:active:*` hashes | Fallback active session/user scan |
| `session:counter:events:YYYYMMDDHHMM` keys | Events-per-minute calculation |

The prefix comes from:

```env
SESSION_REDIS_KEY_PREFIX=session
```

Verify Redis:

```bash
redis-cli -h localhost -p 6379 ping
redis-cli -h localhost -p 6379 -n 2 keys 'session:*'
```

Task 7 does not need a new Redis instance or new Redis port.

### 6.2 API Gateway / Admin Auth Headers

Analytics APIs are admin-only. Current service checks request headers:

```text
X-User-Roles: admin
X-User-Role: admin
X-Authenticated-Roles: admin
X-Auth-Roles: admin
```

Allowed admin roles:

```text
admin
superadmin
operations_admin
```

Local curl example:

```bash
curl -H "X-User-Roles: admin" "http://localhost:8086/api/v1/analytics/live"
```

Production warning:

Header-based role trust is only safe behind a trusted API Gateway that validates JWT/session and overwrites identity headers. Direct public exposure of the session service is unsafe.

### 6.3 Kafka / RabbitMQ / Queue

Task 7 implementation does not include Kafka, RabbitMQ, NATS, or another queue.

`task7.md` mentions that aggregates can be built by a periodic worker or event consumer, but current repo does not provide a dedicated funnel/retention analytics aggregator worker. For local setup:

| Need | Current Status |
|---|---|
| Queue install | Not required |
| Queue env variables | Not required |
| Funnel aggregate worker | Not implemented in current service |
| Manual/dev aggregate seeding | Optional for testing large date ranges |

### 6.4 Heatmap Read Dependency

Task 7 includes `/api/v1/analytics/heatmaps`, but heatmap setup belongs to Task 6.

Refer:

```text
TaskImplementation/Session Management Service/task6_Dependency.md
Section: 5. Database Setup
Section: 6. Redis / Queue / External Services
Section: 10. Running the Project
```

Task 7 only reuses the heatmap read endpoint.

### 6.5 Device / GeoIP Dependency

Session list and journey response can include device and geo fields. Device and GeoIP setup already exists.

Refer:

```text
TaskImplementation/Session Management Service/task5_Dependency.md
Section: 6.1 MaxMind GeoIP
Section: 7. Environment Variables
```

Task 7 does not introduce new GeoIP dependencies.

## 7. Environment Variables

### 7.1 Where To Put Environment Variables

Use:

```text
backend/services/session-service/.env
```

There is already:

```text
backend/services/session-service/.env.example
```

Important:

The Go code reads environment variables from the process environment. It does not automatically load `.env`. For local shell:

```bash
cd backend/services/session-service
set -a
. ./.env
set +a
go run ./cmd/server
```

This `.env` loading behavior is already explained in:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section: 7. Environment Variables
```

### 7.2 Task 7 `.env` Example Only

Do not duplicate the full `.env` here. Add/verify only these Task 7 analytics variables in `backend/services/session-service/.env`:

```env
# Analytics API limits and cache windows
SESSION_ANALYTICS_LIVE_WINDOW=5m
SESSION_ANALYTICS_SESSION_LOOKBACK=24h
SESSION_ANALYTICS_MAX_SESSION_RANGE_DAYS=90
SESSION_ANALYTICS_DEFAULT_PAGE_SIZE=50
SESSION_ANALYTICS_MAX_PAGE_SIZE=100
SESSION_ANALYTICS_MAX_FUNNEL_RANGE_DAYS=90
SESSION_ANALYTICS_RAW_FALLBACK_RANGE=24h
SESSION_ANALYTICS_MIN_FUNNEL_STEPS=2
SESSION_ANALYTICS_MAX_FUNNEL_STEPS=8
SESSION_ANALYTICS_DEFAULT_FUNNEL_METRIC=funnel.checkout
SESSION_ANALYTICS_LIVE_CACHE_TTL=10s
SESSION_ANALYTICS_SESSIONS_CACHE_TTL=30s
SESSION_ANALYTICS_FUNNELS_CACHE_TTL=2m
```

Required inherited variables are documented in previous files:

| Variable Group | Reference |
|---|---|
| HTTP server `SESSION_HTTP_*` | `task1_Dependency.md`, section `7. Environment Variables` |
| MongoDB `SESSION_MONGO_*` | `task1_Dependency.md`, section `5.1 MongoDB Setup` and `7. Environment Variables` |
| Redis `SESSION_REDIS_*` | `task1_Dependency.md`, section `5.2 Redis Setup` and `7. Environment Variables` |
| Ingest/event limits | `task3_Dependency.md`, section `7. Environment Variables` |
| Journey limits | `task4_Dependency.md`, section `7. Environment Variables` |
| Device/GeoIP settings | `task5_Dependency.md`, section `7. Environment Variables` |
| Heatmap settings | `task6_Dependency.md`, section `7. Environment Variables` |

### 7.3 Task 7 Environment Variable Table

| Variable | Default | Required? | Meaning |
|---|---|---:|---|
| `SESSION_ANALYTICS_LIVE_WINDOW` | `5m` | Yes | Live metrics kitne recent window me count honge |
| `SESSION_ANALYTICS_SESSION_LOOKBACK` | `24h` | Yes | Sessions list me default date range jab `from/to` missing ho |
| `SESSION_ANALYTICS_MAX_SESSION_RANGE_DAYS` | `90` | Yes | Sessions list max allowed date range |
| `SESSION_ANALYTICS_DEFAULT_PAGE_SIZE` | `50` | Yes | Session list default page size |
| `SESSION_ANALYTICS_MAX_PAGE_SIZE` | `100` | Yes | Session list max page size |
| `SESSION_ANALYTICS_MAX_FUNNEL_RANGE_DAYS` | `90` | Yes | Funnel API max allowed date range |
| `SESSION_ANALYTICS_RAW_FALLBACK_RANGE` | `24h` | Yes | Aggregate missing ho to raw events fallback max range |
| `SESSION_ANALYTICS_MIN_FUNNEL_STEPS` | `2` | Yes | Funnel query me minimum steps |
| `SESSION_ANALYTICS_MAX_FUNNEL_STEPS` | `8` | Yes | Funnel query me maximum steps |
| `SESSION_ANALYTICS_DEFAULT_FUNNEL_METRIC` | `funnel.checkout` | Yes | Default aggregate metric name |
| `SESSION_ANALYTICS_LIVE_CACHE_TTL` | `10s` | Optional | In-process live response cache TTL |
| `SESSION_ANALYTICS_SESSIONS_CACHE_TTL` | `30s` | Optional | In-process sessions list cache TTL |
| `SESSION_ANALYTICS_FUNNELS_CACHE_TTL` | `2m` | Optional | In-process funnel report cache TTL |

### 7.4 Validation Rules

Startup fails if:

| Bad Config | Error |
|---|---|
| `SESSION_ANALYTICS_LIVE_WINDOW=0s` | `SESSION_ANALYTICS_LIVE_WINDOW must be greater than zero` |
| `SESSION_ANALYTICS_SESSION_LOOKBACK=0s` | `SESSION_ANALYTICS_SESSION_LOOKBACK must be greater than zero` |
| `SESSION_ANALYTICS_MAX_SESSION_RANGE_DAYS=0` | `SESSION_ANALYTICS_MAX_SESSION_RANGE_DAYS must be greater than zero` |
| `SESSION_ANALYTICS_DEFAULT_PAGE_SIZE=0` | `SESSION_ANALYTICS_DEFAULT_PAGE_SIZE must be greater than zero` |
| `SESSION_ANALYTICS_MAX_PAGE_SIZE` less than default page size | `SESSION_ANALYTICS_MAX_PAGE_SIZE must be greater than or equal to default page size` |
| `SESSION_ANALYTICS_MAX_FUNNEL_RANGE_DAYS=0` | `SESSION_ANALYTICS_MAX_FUNNEL_RANGE_DAYS must be greater than zero` |
| `SESSION_ANALYTICS_RAW_FALLBACK_RANGE=0s` | `SESSION_ANALYTICS_RAW_FALLBACK_RANGE must be greater than zero` |
| `SESSION_ANALYTICS_MIN_FUNNEL_STEPS=0` | `SESSION_ANALYTICS_MIN_FUNNEL_STEPS must be greater than zero` |
| `SESSION_ANALYTICS_MAX_FUNNEL_STEPS` less than min steps | `SESSION_ANALYTICS_MAX_FUNNEL_STEPS must be greater than or equal to min funnel steps` |
| empty `SESSION_ANALYTICS_DEFAULT_FUNNEL_METRIC` | `SESSION_ANALYTICS_DEFAULT_FUNNEL_METRIC cannot be empty` |
| negative cache TTL | `SESSION_ANALYTICS_CACHE_TTL values cannot be negative` |

## 8. Docker Setup

### 8.1 Current Repo Status

No official Session Service Dockerfile or docker-compose file is committed in the current service folder.

This is already documented in:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section: 8. Docker Setup

TaskImplementation/Session Management Service/task6_Dependency.md
Section: 8. Docker Setup
```

### 8.2 What Task 7 Needs From Docker

Task 7 does not need new containers. Use the same local stack:

| Container | Needed? | Purpose |
|---|---:|---|
| MongoDB | Yes | Store and query analytics/session collections |
| Redis | Yes | Live metrics |
| Session service app container | Not standardized | Run with `go run` on host for now |
| Kafka/RabbitMQ | No | Not used by current Task 7 code |

### 8.3 Persistent Volumes

Use previous persistent volume setup for MongoDB and Redis.

Refer:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section: 8. Docker Setup -> Recommended beginner Docker Compose
```

### 8.4 Docker Networking

If Go service runs on host:

```env
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_REDIS_ADDR=localhost:6379
```

If Go service later runs inside Docker Compose:

```env
SESSION_MONGO_URI=mongodb://mongo:27017
SESSION_REDIS_ADDR=redis:6379
```

Task 7 adds no port changes.

## 9. Local Development Setup

This section is a Task 7-specific onboarding flow. Base explanations are referenced, commands are included for practical execution.

### Step 1: Clone Repository

Clone steps are already explained in:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section: 9. Local Development Setup -> Step 1: Clone repository
```

Then go to service:

```bash
cd backend/services/session-service
```

### Step 2: Install / Verify Go Dependencies

```bash
go version
go mod download
go test ./...
```

If dependency download fails, refer:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section: 12. Common Errors & Fixes -> go mod download failed
```

### Step 3: Start MongoDB And Redis

Use previous setup:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section: 9. Local Development Setup -> Step 4: Start MongoDB and Redis
```

Verify:

```bash
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
redis-cli -h localhost -p 6379 ping
```

### Step 4: Create `.env`

```bash
cp .env.example .env
```

Edit the required inherited variables:

```env
SESSION_HTTP_ADDR=:8086
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_MONGO_DATABASE=session_db
SESSION_REDIS_ADDR=localhost:6379
SESSION_REDIS_DB=2
SESSION_IP_HASH_SALT=replace-with-real-local-secret
SESSION_PRIVACY_HASH_PEPPER=replace-with-real-local-secret
```

Also verify Task 7 analytics variables from section `7.2`.

### Step 5: Load `.env`

```bash
set -a
. ./.env
set +a
```

### Step 6: Run Migrations Through Task 7

```bash
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/001_session_storage_indexes.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/002_journey_summaries.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/003_device_tracking_indexes.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/004_heatmap_points.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/005_analytics_aggregates.up.js
```

### Step 7: Verify Analytics Collection

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.analytics_aggregates.getIndexes().map(i => i.name)"
```

### Step 8: Start Backend

```bash
go run ./cmd/server
```

Expected log:

```text
session.http.started
```

### Step 9: Verify Health And Readiness

```bash
curl "http://localhost:8086/healthz"
curl "http://localhost:8086/readyz"
```

`/readyz` should show both:

```text
mongo: ok
redis: ok
```

### Step 10: Verify Analytics APIs

Live metrics:

```bash
curl -H "X-User-Roles: admin" "http://localhost:8086/api/v1/analytics/live"
```

Sessions list:

```bash
curl -H "X-User-Roles: admin" "http://localhost:8086/api/v1/analytics/sessions?page=1&page_size=10"
```

Funnel report:

```bash
curl -H "X-User-Roles: admin" "http://localhost:8086/api/v1/analytics/funnels?from=2026-05-22T00:00:00Z&to=2026-05-23T00:00:00Z&steps=product_view,add_to_cart,checkout_step,payment_result"
```

Journey metadata:

```bash
curl -H "X-User-Roles: admin" "http://localhost:8086/api/v1/analytics/sessions/sess_example/journey"
```

Heatmap:

```bash
curl -H "X-User-Roles: admin" "http://localhost:8086/api/v1/analytics/heatmaps?path=/products/example&device_type=mobile&heatmap_type=click&from=2026-05-22&to=2026-05-22"
```

## 10. Running the Project

### 10.1 Daily Run Flow

For a developer who already completed previous setup:

```bash
cd backend/services/session-service
set -a
. ./.env
set +a
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
redis-cli -h localhost -p 6379 ping
go test ./...
go run ./cmd/server
```

In another terminal:

```bash
curl "http://localhost:8086/readyz"
curl -H "X-User-Roles: admin" "http://localhost:8086/api/v1/analytics/live"
```

### 10.2 Ports & Networking

| Service | Port | Purpose | New In Task 7? |
|---|---:|---|---:|
| Session Backend API | `8086` | HTTP APIs, health, readiness, analytics endpoints | No |
| MongoDB | `27017` | Session and analytics storage | No |
| Redis | `6379` | Live metrics and active sessions | No |

No new port was introduced by Task 7.

### 10.3 Admin Header Requirement

Analytics endpoints require admin role header in local development:

```bash
-H "X-User-Roles: admin"
```

Without this:

| Situation | Response |
|---|---|
| No auth/role headers | `401 AUTHENTICATION_REQUIRED` |
| Non-admin role like `buyer` | `403 PERMISSION_DENIED` |

## 11. Common Errors & Fixes

### 1. `401 AUTHENTICATION_REQUIRED`

Cause:

Analytics endpoint called without trusted admin headers.

Fix:

```bash
curl -H "X-User-Roles: admin" "http://localhost:8086/api/v1/analytics/live"
```

Production fix:

Route through API Gateway and make sure gateway injects authenticated role headers only after JWT/session validation.

### 2. `403 PERMISSION_DENIED`

Cause:

Role header exists but role is not admin.

Allowed values:

```text
admin
superadmin
operations_admin
```

Fix:

```bash
curl -H "X-User-Roles: operations_admin" "http://localhost:8086/api/v1/analytics/sessions"
```

### 3. `409 AGGREGATE_NOT_READY`

Cause:

Funnel API did not find matching `analytics_aggregates` data and requested range is larger than raw fallback window.

Fix options:

| Fix | When To Use |
|---|---|
| Reduce `from/to` range within `SESSION_ANALYTICS_RAW_FALLBACK_RANGE` | Local quick test |
| Seed a dev aggregate document | Local dashboard testing |
| Build/run analytics aggregate worker | Real production setup |
| Increase `SESSION_ANALYTICS_RAW_FALLBACK_RANGE` carefully | Temporary dev-only workaround |

Example shorter query:

```bash
curl -H "X-User-Roles: admin" "http://localhost:8086/api/v1/analytics/funnels?from=2026-05-22T00:00:00Z&to=2026-05-22T12:00:00Z&steps=product_view,add_to_cart"
```

### 4. `503 ANALYTICS_STORAGE_UNAVAILABLE`

Cause:

MongoDB or Redis query failed.

Check:

```bash
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
redis-cli -h localhost -p 6379 ping
curl "http://localhost:8086/readyz"
```

Also verify:

```env
SESSION_MONGO_URI=mongodb://localhost:27017
SESSION_REDIS_ADDR=localhost:6379
```

### 5. Funnel API Says `from is required` Or `to is required`

Cause:

Funnel API requires both `from` and `to`.

Fix:

Use RFC3339 or date format:

```bash
curl -H "X-User-Roles: admin" "http://localhost:8086/api/v1/analytics/funnels?from=2026-05-22T00:00:00Z&to=2026-05-23T00:00:00Z&steps=product_view,add_to_cart"
```

### 6. `date range is too large`

Cause:

Requested range exceeds:

```env
SESSION_ANALYTICS_MAX_SESSION_RANGE_DAYS=90
SESSION_ANALYTICS_MAX_FUNNEL_RANGE_DAYS=90
```

Fix:

Use a smaller range or adjust env only if product requirements allow it.

### 7. `page_size cannot exceed 100`

Cause:

`SESSION_ANALYTICS_MAX_PAGE_SIZE` default is `100`.

Fix:

```bash
curl -H "X-User-Roles: admin" "http://localhost:8086/api/v1/analytics/sessions?page=1&page_size=100"
```

### 8. `at least 2 funnel steps are required`

Cause:

Funnel API received fewer than `SESSION_ANALYTICS_MIN_FUNNEL_STEPS`.

Fix:

```bash
curl -H "X-User-Roles: admin" "http://localhost:8086/api/v1/analytics/funnels?from=2026-05-22&to=2026-05-23&steps=product_view,add_to_cart"
```

### 9. `funnel steps cannot exceed 8`

Cause:

Too many `steps` values.

Fix:

Keep funnel steps less than or equal to:

```env
SESSION_ANALYTICS_MAX_FUNNEL_STEPS=8
```

### 10. `analytics_aggregates` Collection Missing

Cause:

Task 7 migration not run.

Fix:

```bash
cd backend/services/session-service
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/005_analytics_aggregates.up.js
```

Verify:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.getCollectionNames().includes('analytics_aggregates')"
```

### 11. Live Metrics Are Always Zero

Possible causes:

| Cause | Fix |
|---|---|
| No session events ingested recently | Send test events using Task 3 flow |
| Redis DB mismatch | Check `SESSION_REDIS_DB=2` and use `redis-cli -n 2` |
| Wrong key prefix | Check `SESSION_REDIS_KEY_PREFIX=session` |
| Active TTL expired | Check `SESSION_ACTIVE_TTL` and ingest a fresh event |

Reference:

```text
TaskImplementation/Session Management Service/task3_Dependency.md
Section: 10. Running the Project
```

### 12. `/readyz` Shows Redis Or Mongo Error

Cause:

Dependency service is down, wrong port, wrong URI, or container still starting.

Fix:

Refer existing troubleshooting:

```text
TaskImplementation/Session Management Service/task1_Dependency.md
Section: 12. Common Errors & Fixes
```

### 13. Heatmap Endpoint Returns Empty Points

Task 7 reuses Task 6 heatmap setup. Refer:

```text
TaskImplementation/Session Management Service/task6_Dependency.md
Section: 11. Common Errors & Fixes
```

## 12. Security & Best Practices

### 12.1 Task 7 Security Audit

| Area | Current Status | Risk | Recommendation |
|---|---|---|---|
| Admin auth | Header-based role check in service | Unsafe if service is directly public | Expose only behind API Gateway; gateway must validate JWT and overwrite role headers |
| Direct service port | `:8086` can be bound locally or network-wide | Analytics data exposure | In production bind to private network, not public internet |
| Analytics PII | Session responses can include `anonymous_id`, `user_id`, geo/device data | Privacy risk | Keep admin-only, mask fields where dashboard role does not need full identity |
| `analytics_aggregates` | Schema allows flexible `segment` and `values` objects | Teams may accidentally store PII in aggregates | Store counts/rates only; never store user lists, raw IPs, emails, phone numbers |
| Secrets | `.env.example` has placeholder salts | Weak hashing if not changed | Set strong random `SESSION_IP_HASH_SALT` and `SESSION_PRIVACY_HASH_PEPPER` |
| Migration runner | Manual `mongosh` scripts | Human error in setup | Add a Makefile/script or migration tool later |
| Docker artifacts | No official service Dockerfile/compose | Inconsistent local/prod setup | Add official Dockerfile and compose before production deployment |
| Aggregation worker | Funnel/retention aggregate builder not implemented in current repo | Historical reports can be missing | Add scheduled aggregator job or controlled seed process |
| Cache | In-process analytics cache only | Cache resets on restart and is per instance | For multi-instance production, consider Redis/shared cache if needed |
| Rate limiting | Not visible in session service | Heavy analytics queries may overload DB | Enforce API Gateway rate limits for admin analytics endpoints |

### 12.2 Beginner Best Practices

Avoid repeating older generic best practices. For Task 7 specifically:

- Keep analytics endpoints admin-only.
- Use small date ranges while testing.
- Prefer precomputed `analytics_aggregates` for dashboard reports.
- Do not scan raw `session_events` for large historical date ranges.
- Do not store personal data inside aggregate documents.
- Run `005_analytics_aggregates.up.js` before testing funnel APIs.
- Keep `SESSION_ANALYTICS_MAX_PAGE_SIZE` reasonable to avoid huge responses.
- Use `SESSION_ANALYTICS_RAW_FALLBACK_RANGE` as a safety limit, not as a production reporting strategy.
- Verify `/readyz` before debugging API logic.
- Use `X-Request-ID` in manual tests when debugging logs.

## 13. Missing or Misconfigured Things

Task 7 setup audit found these gaps:

| Missing / Misconfigured Item | Impact | Recommended Fix |
|---|---|---|
| No official Session Service Dockerfile | Cannot build standard service image | Add Dockerfile in service or deploy folder |
| No official local docker-compose file committed | Beginners must rely on docs/manual commands | Add compose with MongoDB, Redis, service, healthchecks |
| No migration runner | Migrations are manual `mongosh` commands | Add `make migrate-session` or a small migration command |
| No dedicated analytics aggregate worker | Large funnel/retention reports can return `AGGREGATE_NOT_READY` | Build scheduled worker for `analytics_aggregates` |
| gRPC mentioned in docs but not in current module | New developers may search for nonexistent gRPC setup | Document that current runtime is HTTP only unless gRPC is later implemented |
| Header-based admin auth inside service | Unsafe without trusted gateway | Keep service private and validate auth at gateway |
| No production rate-limit config in service | Analytics queries can be expensive | Add gateway-level admin rate limits |
| No committed sample aggregate seed file | Local funnel testing may be confusing | Add optional dev seed under migrations or testdata later |

## 14. References to Previous Dependency Files

Use these instead of duplicating old setup:

| Topic | Previous File / Section |
|---|---|
| Base project overview and clone flow | `TaskImplementation/Session Management Service/task1_Dependency.md`, section `9. Local Development Setup` |
| Go modules, `go.mod`, `go.sum`, `go mod tidy`, `go run`, `go build` | `TaskImplementation/Session Management Service/task1_Dependency.md`, section `4. Dependency Management` |
| MongoDB install, Docker run, connection string, credentials | `TaskImplementation/Session Management Service/task1_Dependency.md`, section `5.1 MongoDB Setup` |
| Redis install, Docker run, connection config | `TaskImplementation/Session Management Service/task1_Dependency.md`, section `5.2 Redis Setup` |
| Base Docker compose example and missing Docker artifacts | `TaskImplementation/Session Management Service/task1_Dependency.md`, section `8. Docker Setup` |
| Base `.env` loading and complete env example | `TaskImplementation/Session Management Service/task1_Dependency.md`, section `7. Environment Variables` |
| Session storage indexes and TTL behavior | `TaskImplementation/Session Management Service/task2_Dependency.md`, section `5. Database Setup` |
| Event ingestion setup and test events | `TaskImplementation/Session Management Service/task3_Dependency.md`, section `10. Running the Project` |
| Journey metadata setup and errors | `TaskImplementation/Session Management Service/task4_Dependency.md`, sections `5. Database Setup` and `11. Common Errors & Fixes` |
| Device, trusted proxy, and GeoIP setup | `TaskImplementation/Session Management Service/task5_Dependency.md`, sections `6. Redis / Queue / External Services` and `7. Environment Variables` |
| Heatmap aggregation and heatmap endpoint setup | `TaskImplementation/Session Management Service/task6_Dependency.md`, sections `5. Database Setup`, `7. Environment Variables`, and `10. Running the Project` |

## 15. Final Checklist

Use this checklist only for Task 7 analytics setup:

- [ ] Previous base setup from `task1_Dependency.md` completed.
- [ ] MongoDB is running on `27017`.
- [ ] Redis is running on `6379`.
- [ ] `backend/services/session-service/.env` exists.
- [ ] `SESSION_MONGO_URI` and `SESSION_MONGO_DATABASE` are correct.
- [ ] `SESSION_REDIS_ADDR`, `SESSION_REDIS_DB`, and `SESSION_REDIS_KEY_PREFIX` are correct.
- [ ] Strong local values are set for `SESSION_IP_HASH_SALT` and `SESSION_PRIVACY_HASH_PEPPER`.
- [ ] Task 7 analytics env variables are present or defaults are acceptable.
- [ ] Migrations `001` through `005` have been run.
- [ ] `analytics_aggregates` collection exists.
- [ ] `uniq_metric_bucket_time_segment`, `idx_metric_time`, `idx_segment_time`, and `idx_retention_cohort` indexes exist.
- [ ] `go mod download` succeeds.
- [ ] `go test ./...` succeeds.
- [ ] Backend starts with `go run ./cmd/server`.
- [ ] `/healthz` returns OK.
- [ ] `/readyz` shows MongoDB and Redis OK.
- [ ] `GET /api/v1/analytics/live` works with `X-User-Roles: admin`.
- [ ] Session list API works with admin header.
- [ ] Funnel API either returns data or expected `409 AGGREGATE_NOT_READY` when aggregates are missing.
- [ ] Heatmap API setup is verified via Task 6 docs.
- [ ] Service is not exposed publicly without API Gateway authentication.
