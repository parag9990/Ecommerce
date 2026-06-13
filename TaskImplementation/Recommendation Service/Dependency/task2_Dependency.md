# Project Dependency & Setup Guide

## Variables Used

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Recommendation Service` |
| `TASK_FILE_NAME` | `task2.md` |
| `OUTPUT_FILE_NAME` | `task2_Dependency.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |
| Backend code path | `backend/services/recommendation-service` |

> Beginner note: Is file me sirf dependency, setup, environment, database, Redis, Docker, and DevOps onboarding explain hai. Business logic ya original implementation file modify nahi kiya gaya.

---

## 1. Project Overview

`TASK_FILE_NAME` ka focus hai: `SERVICE_NAME` ke liye MongoDB plus Redis storage stack choose karna.

Simple Hinglish:

- MongoDB durable storage hai. Isme recommendation data, user interactions, product feature snapshots, and generated recommendation sets store honge.
- Redis fast cache layer hai. Isme hot recommendation lists short TTL ke saath rakhi jaayengi, so API/gRPC reads fast rahein.
- Current task ka setup mostly storage/cache related hai. Go runtime, basic service run commands, Kafka, ports, and generic Docker setup already previous dependency guide me documented hai.

Read this previous file first:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`1. Project Tech Stack Analysis`
`2. Go Dependency System`
`4. Environment Variables`
`5. External Services`
`6. Ports and Networking`
`7. Docker and DevOps Setup`
`8. Project Run Instructions`
`9. Common Errors and Fixes`
```

`TASK_FILE_NAME` new setup scope:

- MongoDB database: `recommendation_db`
- MongoDB migration: `backend/services/recommendation-service/migrations/001_create_recommendation_storage.up.js`
- Redis cache prefix: `reco:v1`
- Cache TTL defaults: `900` seconds for normal lists, `300` seconds for personalized/guest lists
- Storage verification endpoints:
  - `GET /internal/v1/recommendation/storage`
  - `GET /internal/v1/recommendation/storage/status`

---

## 2. Tech Stack

Most base technologies are already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `1. Project Tech Stack Analysis`
```

`TASK_FILE_NAME`-specific technologies:

| Technology | Required? | Beginner Explanation | Why Current Task Uses It |
|---|---:|---|---|
| MongoDB | Required for current storage path | MongoDB ek document database hai. Data JSON-like documents me store hota hai, tables ke fixed columns jaise strict nahi hota. | Flexible recommendation documents, interactions, product feature snapshots, and generated lists store karne ke liye. |
| Redis | Required for current cache path | Redis ek in-memory cache hai. Simple English: fast temporary storage for quick reads. | Trending, similar, personalized, and FBT recommendation responses fast serve karne ke liye. |
| MongoDB Go Driver v2 | Required by backend code | Go application ko MongoDB se connect/query/upsert karne wali official library. | `go.mongodb.org/mongo-driver/v2` already `go.mod` me present hai. |
| go-redis v9 | Required by backend code | Go application ko Redis se connect/get/set/delete karne wali Redis client library. | `github.com/redis/go-redis/v9` already `go.mod` me present hai. |
| `mongosh` | Required for manual migrations | MongoDB shell CLI. | Current task migration file run/verify karne ke liye. |
| `redis-cli` | Recommended | Redis command-line client. | Redis ping, TTL, and cache key debugging ke liye. |

No new MySQL, PostgreSQL, RabbitMQ, NATS, MinIO, Elasticsearch, AWS S3, SMTP, Stripe, Twilio, OAuth, Kubernetes, or Nginx dependency is introduced by `TASK_FILE_NAME`.

---

## 3. Required Software

Base installation is already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Sections:
`2. Go Dependency System`
`3. Databases`
`5. External Services`
`7. Docker and DevOps Setup`
```

For `TASK_FILE_NAME`, make sure these are available:

| Software | Needed For | New or Reused |
|---|---|---|
| Go `1.26.3` or compatible newer | Build/run backend service | Reused from previous dependency guide |
| Docker / Docker Compose | Easy local MongoDB and Redis containers | Reused |
| MongoDB `7.x` | Durable recommendation database | Reused setup, current-task-specific usage |
| `mongosh` | Run `001_create_recommendation_storage.up.js` | Reused setup, current-task-specific migration |
| Redis `7.x` | Hot recommendation cache | Reused setup, current-task-specific cache plan |
| `redis-cli` | Verify Redis and debug TTL/key values | Reused setup |

Important: Kafka is not required to verify only current storage/cache choice. Kafka becomes required when event ingestion is enabled, which belongs to later task scope.

---

## 4. Dependency Management

This is a Go module project. Go modules, `go.mod`, `go.sum`, `go mod download`, and common Go dependency errors are already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `2. Go Dependency System`
```

`TASK_FILE_NAME` dependency status:

| Dependency | Present In | Purpose |
|---|---|---|
| `go.mongodb.org/mongo-driver/v2 v2.6.0` | `backend/services/recommendation-service/go.mod` | MongoDB client, ping, indexes, CRUD |
| `github.com/redis/go-redis/v9 v9.19.0` | `backend/services/recommendation-service/go.mod` | Redis client, cache read/write, locks |

No new `go get` command is needed for `TASK_FILE_NAME` because these dependencies already exist in the service module.

Recommended verification:

```bash
cd backend/services/recommendation-service
go mod download
go test ./...
```

---

## 5. Database Setup

### MongoDB Role

MongoDB is the durable source for `TASK_FILE_NAME`.

Simple Hinglish:

MongoDB me `SERVICE_NAME` ka long-lived data rahega. Redis temporary cache hai, but MongoDB source-of-truth style durable storage hai.

### Required or Optional

| Scenario | MongoDB Required? |
|---|---:|
| Service minimal startup with storage env empty | No |
| `GET /internal/v1/recommendation/storage` plan endpoint | No |
| `GET /internal/v1/recommendation/storage/status` showing Mongo ready | Yes |
| Persisting generated recommendation sets | Yes |
| Full recommendation storage/cache flow | Yes |

### Installation

MongoDB installation and Docker setup are already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `3. Databases` -> `MongoDB`
```

Do not duplicate those steps here. Use the same local MongoDB setup.

### Current Task Database Name

```env
RECOMMENDATION_MONGO_DATABASE=recommendation_db
```

### Current Task Migration

Run only the current storage migration when you want to create the initial MongoDB storage collections and indexes:

```bash
cd backend/services/recommendation-service
export RECOMMENDATION_MONGO_DATABASE=recommendation_db

mongosh "mongodb://root:secret@localhost:27017/admin" \
  --file migrations/001_create_recommendation_storage.up.js
```

If your MongoDB has no auth:

```bash
mongosh "mongodb://localhost:27017" \
  --file migrations/001_create_recommendation_storage.up.js
```

### Collections Introduced By Current Task

| Collection | Purpose | Required? |
|---|---|---:|
| `user_interactions` | Product view/click/cart/wishlist/purchase behavior history store karna | Yes for future event-based recommendations |
| `product_features` | Product metadata/features ka recommendation-friendly snapshot | Yes for similar/trending candidates |
| `recommendation_sets` | Generated recommendation lists per context and strategy | Yes for Redis miss recovery |
| `ab_test_assignments` | Future strategy experiment assignment storage | Optional for current task, created as future-ready collection |

### Current Task Indexes

| Collection | Index | Why |
|---|---|---|
| `user_interactions` | `idx_user_interactions_user_recent` | User ke recent behavior reads fast karne ke liye |
| `user_interactions` | `idx_user_interactions_anon_recent` | Guest/anonymous behavior reads fast karne ke liye |
| `user_interactions` | `idx_user_interactions_product_event_recent` | Product popularity and co-occurrence calculations |
| `user_interactions` | `idx_user_interactions_category_event_recent` | Category trending calculations |
| `user_interactions` | `idx_user_interactions_ttl` | Old raw interactions auto-cleanup, default around 180 days |
| `recommendation_sets` | `ux_recommendation_sets_context_strategy` | Same context/strategy duplicate generated list avoid karna |
| `recommendation_sets` | `idx_recommendation_sets_expires_at_ttl` | Expired generated sets auto-delete |
| `recommendation_sets` | `idx_recommendation_sets_type_generated` | Latest sets by type debug/inspect karna |
| `product_features` | `ux_product_features_product_id` | One product snapshot per product |
| `product_features` | `idx_product_features_category_status_stock` | Category candidate queries |
| `product_features` | `idx_product_features_seller_status_stock` | Seller candidate queries |

Note: Backend code can also ensure indexes on startup when:

```env
RECOMMENDATION_MONGO_ENSURE_INDEXES=true
```

For production, migration-driven indexes are cleaner. App-time index creation can slow startup.

### Verify MongoDB

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" \
  --eval 'db.getSiblingDB("recommendation_db").getCollectionNames()'
```

Expected current-task collections should include:

```text
user_interactions
product_features
recommendation_sets
ab_test_assignments
```

Credentials placement:

```text
backend/services/recommendation-service/.env
```

Do not commit real MongoDB passwords. Local sample values like `root:secret` are only for local development.

---

## 6. Redis / Queue / External Services

### Redis Role

Redis is the hot cache for `TASK_FILE_NAME`.

Simple Hinglish:

Redis fast memory cache hai. MongoDB se data recover ho sakta hai, but Redis se page/API response quick milta hai.

### Redis Setup

Redis installation, Docker setup, and `redis-cli ping` are already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `5. External Services` -> `Redis`
```

Use the same setup. Current task adds the cache key and TTL strategy below.

### Redis Cache Keys

Default prefix:

```env
RECOMMENDATION_CACHE_KEY_PREFIX=reco:v1
```

Current task cache patterns:

| Cache | Key Pattern | TTL | Purpose |
|---|---|---:|---|
| Global trending | `reco:v1:trending:global:strategy:{strategy_id}` | `900s` | Home/global top list |
| Category trending | `reco:v1:trending:category:{category_id}:strategy:{strategy_id}` | `900s` | Category page top list |
| Seller popular | `reco:v1:trending:seller:{seller_id}:strategy:{strategy_id}` | `900s` | Seller storefront popular products |
| Similar products | `reco:v1:similar:product:{product_id}:strategy:{strategy_id}` | `900s` | Product detail recommendations |
| Personalized user | `reco:v1:personalized:user:{user_id}:strategy:{strategy_id}` | `300s` | Known-user feed |
| Personalized guest | `reco:v1:personalized:anon:{anonymous_id}:strategy:{strategy_id}` | `300s` | Anonymous-session feed |
| Frequently bought together | `reco:v1:fbt:product:{product_id}:strategy:{strategy_id}` | `900s` | Cart/product bundle suggestions |
| Dirty product marker | `reco:v1:dirty:product:{product_id}` | `900s` | Product-level cache may be stale |
| Dirty category marker | `reco:v1:dirty:category:{category_id}` | `900s` | Category-level cache may be stale |
| Dirty seller marker | `reco:v1:dirty:seller:{seller_id}` | `900s` | Seller-level cache may be stale |
| Rebuild lock | `reco:v1:lock:{cache_key}` | `60s` | Prevent duplicate rebuild workers |

### Redis Value Type

Current task default cache value is JSON string.

Example:

```json
{
  "recommendation_id": "reco_home_global_v1",
  "context_key": "home_feed:global",
  "recommendation_type": "trending",
  "strategy_id": "trending_v1_recent_activity",
  "items": [
    {
      "product_id": "prod_123",
      "score": 98.5,
      "rank": 1,
      "reason": "high_recent_activity"
    }
  ],
  "cached_at": "2026-06-05T00:00:00Z",
  "expires_at": "2026-06-05T00:15:00Z"
}
```

### Verify Redis

```bash
redis-cli -h localhost -p 6379 ping
redis-cli -h localhost -p 6379 SETEX reco:v1:trending:global:strategy:trending_v1_recent_activity 900 '{"recommendation_id":"demo","context_key":"home_feed:global","recommendation_type":"trending","strategy_id":"trending_v1_recent_activity","items":[{"product_id":"prod_123","score":1,"rank":1}],"cached_at":"2026-06-05T00:00:00Z","expires_at":"2026-06-05T00:15:00Z"}'
redis-cli -h localhost -p 6379 TTL reco:v1:trending:global:strategy:trending_v1_recent_activity
```

### Kafka / Queue

Kafka is not newly introduced by `TASK_FILE_NAME`. It is future event-ingestion scope.

Refer:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `5. External Services` -> `Kafka`
```

For current-task-only verification, keep:

```env
RECOMMENDATION_EVENTS_ENABLED=false
```

---

## 7. Environment Variables

`.env` loading behavior is already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `4. Environment Variables`
```

Important: The Go code uses `os.Getenv`. `.env` is not auto-loaded. You must export it before `go run`.

`TASK_FILE_NAME`-specific `.env` delta:

```env
# MongoDB durable recommendation store
RECOMMENDATION_STORAGE_FAIL_FAST=false
RECOMMENDATION_MONGO_URI=mongodb://root:secret@localhost:27017
RECOMMENDATION_MONGO_DATABASE=recommendation_db
RECOMMENDATION_MONGO_CONNECT_TIMEOUT=10s
RECOMMENDATION_MONGO_PING_TIMEOUT=5s
RECOMMENDATION_MONGO_ENSURE_INDEXES=true
RECOMMENDATION_INTERACTION_RETENTION_SECONDS=15552000

# Redis hot recommendation cache
RECOMMENDATION_REDIS_ADDR=localhost:6379
RECOMMENDATION_REDIS_PASSWORD=
RECOMMENDATION_REDIS_DB=0
RECOMMENDATION_REDIS_DIAL_TIMEOUT=5s
RECOMMENDATION_REDIS_READ_TIMEOUT=3s
RECOMMENDATION_REDIS_WRITE_TIMEOUT=3s
RECOMMENDATION_REDIS_PING_TIMEOUT=3s
RECOMMENDATION_CACHE_KEY_PREFIX=reco:v1
RECOMMENDATION_CACHE_TTL_SECONDS=900
RECOMMENDATION_PERSONALIZED_CACHE_TTL_SECONDS=300
RECOMMENDATION_GUEST_CACHE_TTL_SECONDS=300
RECOMMENDATION_CACHE_REBUILD_LOCK_TTL=1m
RECOMMENDATION_CACHE_DIRTY_TTL_SECONDS=900

# Keep off for current-task-only setup unless Kafka is also ready
RECOMMENDATION_EVENTS_ENABLED=false
```

Variable details:

| Variable | Purpose | Required? | Security / Config Note |
|---|---|---:|---|
| `RECOMMENDATION_STORAGE_FAIL_FAST` | If `true`, service exits when storage init fails | Optional | Use `true` in production when Mongo/Redis are mandatory |
| `RECOMMENDATION_MONGO_URI` | MongoDB connection URI | Required for Mongo path | Secret if it contains username/password |
| `RECOMMENDATION_MONGO_DATABASE` | Database name | Required | Use `recommendation_db` locally |
| `RECOMMENDATION_MONGO_ENSURE_INDEXES` | App creates indexes on startup | Optional | Prefer migrations in production |
| `RECOMMENDATION_INTERACTION_RETENTION_SECONDS` | TTL for raw interactions | Optional | Default `15552000`, around 180 days |
| `RECOMMENDATION_REDIS_ADDR` | Redis host and port | Required for Redis path | Use service name in Docker network, not `localhost` |
| `RECOMMENDATION_REDIS_PASSWORD` | Redis password | Optional | Secret when configured |
| `RECOMMENDATION_REDIS_DB` | Redis logical DB number | Optional | Keep separate per environment |
| `RECOMMENDATION_CACHE_KEY_PREFIX` | Prefix for all recommendation cache keys | Required | Must not start/end with `:` |
| `RECOMMENDATION_CACHE_TTL_SECONDS` | Normal recommendation cache TTL | Optional | Default current-task value `900` |
| `RECOMMENDATION_PERSONALIZED_CACHE_TTL_SECONDS` | Known-user personalized TTL | Optional | Default current-task value `300` |
| `RECOMMENDATION_GUEST_CACHE_TTL_SECONDS` | Anonymous personalized TTL | Optional | Default current-task value `300` |
| `RECOMMENDATION_CACHE_REBUILD_LOCK_TTL` | Redis lock TTL | Optional | Default `1m` |
| `RECOMMENDATION_CACHE_DIRTY_TTL_SECONDS` | Dirty marker TTL | Optional | Default `900` |

Common mistakes:

- `.env` file created but not exported.
- `RECOMMENDATION_MONGO_URI` set with auth but missing correct auth database.
- `RECOMMENDATION_REDIS_ADDR=localhost:6379` used inside Docker container; use `redis:6379` inside Compose network.
- `RECOMMENDATION_CACHE_KEY_PREFIX=:reco:v1` starts with colon; code rejects empty cache key segments.
- Events enabled while Kafka is not running. For current-task-only setup, keep events disabled.

---

## 8. Docker Setup

Docker basics and dependency-only Compose examples are already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `7. Docker and DevOps Setup`
```

`TASK_FILE_NAME` does not add a new service Dockerfile or a new root `docker-compose.yml`.

Current task local dependency need:

| Container | Required for Current Task? | Port |
|---|---:|---:|
| MongoDB | Yes for storage-ready status | `27017` |
| Redis | Yes for cache-ready status | `6379` |
| Kafka | No for current-task-only verification | `9092` |

If you already followed the previous guide, do not recreate containers. Reuse the existing MongoDB and Redis containers.

---

## 9. Local Development Setup

Follow the previous onboarding first:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `8. Project Run Instructions`
```

Current task incremental flow:

### Step 1: Read Previous Dependency Documentation

Read:

```text
TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
```

### Step 2: Go To Project Directory

```bash
cd backend/services/recommendation-service
```

### Step 3: Install Dependencies

No new Go package install is needed for `TASK_FILE_NAME`.

```bash
go mod download
```

### Step 4: Start MongoDB And Redis

Use the MongoDB and Redis setup already documented in `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`.

### Step 5: Add Current Task Environment Variables

Create/update:

```text
backend/services/recommendation-service/.env
```

Use the `TASK_FILE_NAME`-specific `.env` delta from section `7. Environment Variables`.

### Step 6: Run Current Task Migration

```bash
export RECOMMENDATION_MONGO_DATABASE=recommendation_db
mongosh "mongodb://root:secret@localhost:27017/admin" \
  --file migrations/001_create_recommendation_storage.up.js
```

### Step 7: Start Backend Service

```bash
set -a
source .env
set +a

go run ./cmd/server
```

### Step 8: Verify Current Task Functionality

In another terminal:

```bash
curl http://localhost:8088/healthz
curl http://localhost:8088/internal/v1/recommendation/storage
curl http://localhost:8088/internal/v1/recommendation/storage/status
```

Expected storage status when MongoDB and Redis are configured correctly:

```json
{
  "ready": true,
  "mongodb": {
    "configured": true,
    "state": "ready"
  },
  "redis": {
    "configured": true,
    "state": "ready"
  }
}
```

If MongoDB/Redis env vars are empty, status can still be logically ready because storage is `not_configured`. That is okay for minimal mode, but it does not verify current storage/cache path.

---

## 10. Running the Project

Use previous run instructions for clone, Go version, base service startup, and common API checks:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `8. Project Run Instructions`
```

Current task-specific run target:

```bash
cd backend/services/recommendation-service
set -a
source .env
set +a
go run ./cmd/server
```

Current task-specific verification:

```bash
curl -s http://localhost:8088/internal/v1/recommendation/storage
curl -s http://localhost:8088/internal/v1/recommendation/storage/status
```

What to check:

- `database_name` should be `recommendation_db`.
- `cache_key_prefix` should be `reco:v1`.
- `mongo_collections` should include current-task collections.
- `redis_caches` should include trending, similar, personalized, FBT, dirty markers, and rebuild lock.
- `mongodb.state` should be `ready` when Mongo env is configured.
- `redis.state` should be `ready` when Redis env is configured.

---

## 11. Common Errors & Fixes

Generic setup errors are already documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `9. Common Errors and Fixes`
```

Current task-specific errors:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `mongodb.state = not_configured` | `RECOMMENDATION_MONGO_URI` empty | Add Mongo URI and restart service | Keep current-task `.env` delta in service env |
| `redis.state = not_configured` | `RECOMMENDATION_REDIS_ADDR` empty | Add Redis addr and restart service | Keep Redis env values documented |
| `recommendation.storage.init_degraded` | Mongo/Redis env configured but connection failed | Verify container/service is running and credentials match | Run `mongosh` and `redis-cli ping` before `go run` |
| Service exits with `recommendation.storage.init_failed` | `RECOMMENDATION_STORAGE_FAIL_FAST=true` and storage failed | Fix Mongo/Redis connection or set fail-fast false for local | Use fail-fast true only when infra is ready |
| `invalid cache key prefix` | Prefix starts/ends with `:` or has empty segments | Use `RECOMMENDATION_CACHE_KEY_PREFIX=reco:v1` | Avoid values like `:reco:v1`, `reco:v1:`, `reco::v1` |
| `Migration failed` with auth error | Wrong Mongo URI/user/password/auth database | Connect with `mongosh` first, then rerun migration | Keep local URI and migration URI aligned |
| Collections missing after migration | `RECOMMENDATION_MONGO_DATABASE` not exported for `mongosh` | Export database env or verify default `recommendation_db` | Always run `export RECOMMENDATION_MONGO_DATABASE=recommendation_db` before migration |
| Redis TTL is `-2` | Key does not exist | Set key or trigger service cache write path | Use `redis-cli KEYS 'reco:v1:*'` only in local dev |
| Redis TTL is `-1` | Key exists without expiry | Use `SETEX` or service cache writer with TTL | Never set recommendation cache without TTL |
| Cache JSON corrupt error | Redis value is not valid expected JSON | Delete the key; service will fall back to MongoDB | Do not manually write malformed cache values |

---

## 12. Security & Best Practices

Current task-specific best practices:

- MongoDB and Redis credentials ko real production values ke saath git me commit mat karo.
- Local sample `root:secret` sirf local development ke liye acceptable hai.
- Redis cache me raw PII store mat karo. Product IDs, rank, score, reason, strategy metadata enough hai.
- MongoDB `recommendation_db` `SERVICE_NAME` ka owned database hai. Dusri service ke DB ko direct read/write mat karo.
- `RECOMMENDATION_STORAGE_FAIL_FAST=true` production me useful hai, so broken storage ke saath degraded service accidentally run na ho.
- `RECOMMENDATION_MONGO_ENSURE_INDEXES=true` local me convenient hai; production me migrations better hain.
- MongoDB, Redis, and service gRPC port ko public internet par expose mat karo.
- Cache TTL short rakho. Recommendation data stale ho sakta hai, so `900s` and `300s` defaults practical hain.
- Cache key prefix versioned rakho, jaise `reco:v1`. Payload format change ho to `reco:v2` use karo.
- Redis password empty sirf local isolated setup me okay hai. Shared/staging/prod me password/TLS/private network use karo.

---

## 13. Missing or Misconfigured Things

| Finding | Impact | Suggested Fix |
|---|---|---|
| No root Docker Compose file detected for the whole service stack | Beginners manually start MongoDB/Redis/Kafka | Add a checked-in dependency-only Compose file or dev script |
| No service Dockerfile detected for `SERVICE_NAME` | Production packaging not yet codified | Add Dockerfile with health checks before deployment |
| `.env` exists locally but no `.env.example` detected for this service | Beginners may copy real/local secrets accidentally | Add `.env.example` with dummy credentials |
| Existing local `.env` enables Kafka events | Current-task-only local run may fail if Kafka is not running | Set `RECOMMENDATION_EVENTS_ENABLED=false` for current-task-only verification |
| MongoDB startup index creation enabled by default | Production startup can slow if indexes are created at runtime | Use migrations and consider `RECOMMENDATION_MONGO_ENSURE_INDEXES=false` in production |
| Standalone MongoDB is enough for current migration, but later feature jobs use transactions | Later full-service runs can fail on standalone MongoDB | Use a single-node replica set for full feature/ranking/event flows |

---

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `1. Project Tech Stack Analysis` | Go, gRPC, Prometheus, Kafka, Docker basics already explained |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `2. Go Dependency System` | `go.mod`, `go.sum`, `go mod download`, common Go module errors already documented |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `3. Databases` -> `MongoDB` | MongoDB installation, Docker run, verify commands already documented |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `4. Environment Variables` | `.env` location and `os.Getenv` loading behavior already documented |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `5. External Services` -> `Redis` | Redis install, Docker run, and health check already documented |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `5. External Services` -> `Kafka` | Kafka is not new for `TASK_FILE_NAME` and belongs to event ingestion setup |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `6. Ports and Networking` | Ports `8088`, `9088`, `27017`, `6379`, and `9092` already documented |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `7. Docker and DevOps Setup` | Dependency-only Docker examples already documented |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `8. Project Run Instructions` | Clone, base Go run, and generic verification flow already documented |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `9. Common Errors and Fixes` | Generic port, Docker, Go, Mongo, Redis, Kafka errors already documented |

---

## 15. Final Checklist

- [ ] Previous dependency documentation checked first.
- [ ] No duplicate MongoDB/Redis installation guide copied again.
- [ ] Go dependencies verified with `go mod download`.
- [ ] MongoDB running locally or via Docker.
- [ ] Redis running locally or via Docker.
- [ ] Current-task `.env` values added in `backend/services/recommendation-service/.env`.
- [ ] `RECOMMENDATION_EVENTS_ENABLED=false` used for current-task-only run unless Kafka is ready.
- [ ] `migrations/001_create_recommendation_storage.up.js` executed.
- [ ] `recommendation_db` collections verified with `mongosh`.
- [ ] Backend service started with exported environment variables.
- [ ] `/internal/v1/recommendation/storage` returns MongoDB and Redis plan.
- [ ] `/internal/v1/recommendation/storage/status` shows MongoDB `ready`.
- [ ] `/internal/v1/recommendation/storage/status` shows Redis `ready`.
- [ ] Redis cache prefix checked as `reco:v1`.
- [ ] TTL values checked: default `900s`, personalized/guest `300s`.
- [ ] Logs checked for `recommendation.storage.init_degraded`.
- [ ] Real credentials not committed.
- [ ] No original implementation file overwritten.
- [ ] No duplicate setup documentation added beyond task-specific deltas.
