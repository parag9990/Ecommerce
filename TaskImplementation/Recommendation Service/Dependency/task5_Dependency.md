# Project Dependency & Setup Guide

## Variables Used

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Recommendation Service` |
| `TASK_FILE_NAME` | `task5.md` |
| `OUTPUT_FILE_NAME` | `task5_Dependency.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |
| Backend code path | `backend/services/recommendation-service` |

Beginner note: Ye file sirf dependency, setup, environment, database, Redis, DevOps, and troubleshooting guide hai. Business logic ya original implementation file modify nahi kiya gaya.

---

## 1. Project Overview

`TASK_FILE_NAME` ka focus hai: `${SERVICE_NAME}` ke liye rule-based MVP top-list generation.

Simple Hinglish:

- Task 4 ke `product_features` documents se eligible products read hote hain.
- Rule-based score calculate hota hai using recent views, wishlist adds, cart adds, and purchases.
- Generated lists MongoDB `recommendation_sets` collection me save hoti hain.
- Redis me same generated lists fast read ke liye cache hoti hain.
- Scope support: global trending, category popular, seller popular.

Read these previous dependency files first:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`1. Project Tech Stack Analysis`
`2. Go Dependency System`
`3. Databases`
`4. Environment Variables`
`5. External Services`
`6. Ports and Networking`
`7. Docker and DevOps Setup`
`8. Project Run Instructions`
`9. Common Errors and Fixes`
```

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`

Sections:
`5. Database Setup`
`6. Redis / Queue / External Services`
`7. Environment Variables`
`8. Docker Setup`
`9. Local Development Setup`
`11. Common Errors & Fixes`
```

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task3_Dependency.md`

Sections:
`6. Redis / Queue / External Services`
`7. Environment Variables`
`8. Docker Setup`
`9. Local Development Setup`
```

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task4_Dependency.md`

Sections:
`5. Database Setup`
`7. Environment Variables`
`8. Docker Setup`
`9. Local Development Setup`
`13. Missing or Misconfigured Things`
```

`TASK_FILE_NAME` new setup scope:

- MongoDB migration: `backend/services/recommendation-service/migrations/004_add_rule_based_ranking_indexes.up.js`
- Ranking env variables:
  - `RECOMMENDATION_RANKING_ENABLED`
  - `RECOMMENDATION_RANKING_REBUILD_INTERVAL`
  - `RECOMMENDATION_RANKING_FORMULA_VERSION`
  - `RECOMMENDATION_RANKING_VIEW_24H_WEIGHT`
  - `RECOMMENDATION_RANKING_WISHLIST_7D_WEIGHT`
  - `RECOMMENDATION_RANKING_CART_7D_WEIGHT`
  - `RECOMMENDATION_RANKING_PURCHASE_7D_WEIGHT`
- Implemented code paths:
  - `internal/domain/ranking.go`
  - `internal/usecase/trending.go`
  - `internal/repository/mongo_feature_repository.go`
  - `internal/repository/redis_recommendation_cache.go`
  - `cmd/server/main.go`
- New ranking metrics and logs.

No new database engine, queue engine, Docker container type, or public port is introduced by `TASK_FILE_NAME`.

---

## 2. Tech Stack

Base technologies are already explained in previous dependency files. Do not repeat their full installation here.

| Technology | New or Reused | Required? | Beginner Explanation | Current Task Use |
|---|---|---:|---|---|
| Go | Reused | Yes | Go backend language hai. Is project me service, ranking job, repositories, and gRPC server Go me run hote hain. | Rule-based ranking service code run karne ke liye. |
| Go modules | Reused | Yes | `go.mod` and `go.sum` dependencies ko version ke saath manage karte hain. | No new module dependency added for `TASK_FILE_NAME`. |
| MongoDB | Reused storage, new indexes | Yes for full task | MongoDB document database hai. Data JSON-like collections me store hota hai. | `product_features` read and `recommendation_sets` write. |
| Redis | Reused cache | Optional but recommended | Redis in-memory cache hai. Fast read ke liye use hota hai. | Generated trending lists cache karne ke liye. |
| Prometheus metrics | Reused observability | Recommended | Metrics expose karne ka system. | Ranking job count, duration, empty set, fallback, cache errors observe karne ke liye. |
| Kafka | Reused from Task 3 | Not required for manual Task 5 verification | Kafka event stream hai. | Live product interaction events se `product_features` counters fresh rakhne ke liye previous flow me use hota hai. |
| MongoDB replica set | Reused from Task 4 | Required for live feature builder transactions | Replica set transaction support deta hai. | Task 5 migration itself ko transaction nahi chahiye, but live Task 4 feature updates need it. |

Important implementation note:

```text
Task 5 ranking directly reads `product_features`.
It does not scan raw `user_interactions` during every generation run.
```

No new MySQL, PostgreSQL, RabbitMQ, NATS, MinIO, Elasticsearch, AWS S3, SMTP, Stripe, Twilio, OAuth, Kubernetes, or Nginx dependency is introduced by `TASK_FILE_NAME`.

---

## 3. Required Software

Installation of base tools is already documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Sections:
`2. Go Dependency System`
`3. Databases`
`5. External Services`
`7. Docker and DevOps Setup`
```

For `TASK_FILE_NAME`, confirm these are available:

| Software | Needed For | New or Reused |
|---|---|---|
| Go `1.26.3` or compatible newer | Build/test/run backend service | Reused |
| MongoDB `7.x` | `product_features` and `recommendation_sets` | Reused |
| `mongosh` | Run migration `004` and verify generated sets | Reused tool, new migration |
| Redis `7.x` | Cache generated top lists | Reused |
| `redis-cli` | Verify task-specific cache keys and TTL | Reused tool |
| `curl` | Check HTTP health, storage status, and metrics | Reused |
| `grpcurl` | Optional gRPC serving verification | Optional, task-specific verification |
| Kafka `3.x` | Live event-to-feature pipeline only | Reused from Task 3 |

Beginner decision:

- Sirf rule-based ranking verify karna hai: MongoDB plus optional Redis enough hai.
- Live counters Product/Session/Order/Wishlist events se build karne hain: Task 3 and Task 4 setup bhi follow karo, including Kafka and MongoDB replica set.

---

## 4. Dependency Management

This is still a Go module project. Go dependency setup is already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `2. Go Dependency System`
```

`TASK_FILE_NAME` dependency status:

| Dependency | Used For | New Install Needed? |
|---|---|---:|
| Go standard library `sort`, `time`, `context`, `log/slog` | Stable ranking, TTL, scheduled rebuild, logs | No |
| `go.mongodb.org/mongo-driver/v2` | Read `product_features`, upsert/read `recommendation_sets`, ensure indexes | No |
| `github.com/redis/go-redis/v9` | Store generated recommendation cache payloads | No |
| `github.com/prometheus/client_golang` | Expose ranking metrics | No |

No new `go get` command is needed for `TASK_FILE_NAME`.

Task-specific verification:

```bash
cd backend/services/recommendation-service
go test ./internal/domain ./internal/usecase ./internal/repository
```

Full service verification is reused from previous docs:

```bash
cd backend/services/recommendation-service
go test ./...
```

Common Go module errors like old Go version, missing `go.sum`, or local proto module path issues are already documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `2. Go Dependency System` -> `Common Go module issues`
```

---

## 5. Database Setup

### MongoDB Role

MongoDB is mandatory for full `TASK_FILE_NAME` behavior.

Simple Hinglish:

MongoDB me Task 4 ka `product_features` data input ke roop me aata hai. Task 5 generated result ko same database ke `recommendation_sets` collection me save karta hai, taaki Redis miss/restart ke baad bhi last valid list recover ho sake.

### Reused MongoDB Setup

MongoDB installation, Docker setup, connection string format, and credentials placement are already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `3. Databases` -> `MongoDB`
```

```md
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
Section: `5. Database Setup`
```

Task 4 feature-builder transaction requirement is already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task4_Dependency.md`
Section: `5. Database Setup` -> `Replica Set Requirement`
```

Do not recreate MongoDB if it is already running.

### Current Task Migration

Prerequisite migrations:

1. `001_create_recommendation_storage.up.js`
2. `002_add_event_ingestion_indexes.up.js`
3. `003_create_feature_store_schema.up.js`
4. `004_add_rule_based_ranking_indexes.up.js`

Run from the service folder:

```bash
cd backend/services/recommendation-service
export RECOMMENDATION_MONGO_DATABASE=recommendation_db

mongosh "mongodb://root:secret@localhost:27017/admin" \
  --file migrations/001_create_recommendation_storage.up.js

mongosh "mongodb://root:secret@localhost:27017/admin" \
  --file migrations/002_add_event_ingestion_indexes.up.js

mongosh "mongodb://root:secret@localhost:27017/admin" \
  --file migrations/003_create_feature_store_schema.up.js

mongosh "mongodb://root:secret@localhost:27017/admin" \
  --file migrations/004_add_rule_based_ranking_indexes.up.js
```

If your MongoDB has no auth, use the no-auth URI from the previous docs:

```bash
mongosh "mongodb://localhost:27017" \
  --file migrations/004_add_rule_based_ranking_indexes.up.js
```

### Collections Used By Current Task

| Collection | Status | Purpose |
|---|---|---|
| `product_features` | Reused from Task 4, new task-specific indexes | Ranking candidates read karne ke liye. |
| `recommendation_sets` | Reused from Task 2 | Generated top lists durable store karne ke liye. |

### Current Task Indexes

Migration `004` adds these indexes:

| Collection | Index | Why Needed |
|---|---|---|
| `product_features` | `idx_product_features_rule_based_global` | Global eligible candidates quickly scan/filter karne ke liye. |
| `product_features` | `idx_product_features_rule_based_category` | Category-specific popular lists fast generate karne ke liye. |
| `product_features` | `idx_product_features_rule_based_seller` | Seller-specific popular lists fast generate karne ke liye. |

These indexes include fields used by the implemented ranking filter:

```text
status
stock_status
quality_flags.is_recommendable
quality_flags.is_deleted
counters.purchases_7d
product_id
category_id
seller_id
```

### Required Candidate Fields

Task 5 expects `product_features` documents with these fields:

| Field | Required? | Expected Value / Purpose |
|---|---:|---|
| `product_id` | Yes | Product identity. Empty product IDs are skipped. |
| `status` | Yes | Must be `active`. |
| `stock_status` | Yes | Must be `in_stock`. |
| `quality_flags.is_recommendable` | Yes | Must be `true`. |
| `quality_flags.is_deleted` | Yes | Must not be `true`. |
| `category_id` | Needed for category lists | Used to group category popular recommendations. |
| `seller_id` | Needed for seller lists | Used to group seller popular recommendations. |
| `counters.views_24h` | Scoring input | Recent attention signal. |
| `counters.wishlist_adds_7d` | Scoring input | Preference signal. |
| `counters.cart_adds_7d` | Scoring input | Intent signal. |
| `counters.purchases_7d` | Scoring input and tie-breaker | Strong conversion signal. |

Important: Current implementation checks `stock_status: "in_stock"`. Agar older sample me sirf `in_stock: true` field ho, Task 5 us product ko eligible nahi maanega.

### Scoring Rule

Default formula:

```text
score =
  views_24h * 1
  + wishlist_adds_7d * 3
  + cart_adds_7d * 4
  + purchases_7d * 8
```

Tie-break order:

1. Higher score first.
2. Higher `purchases_7d` first.
3. Smaller `product_id` lexicographically first.

Negative counter values are treated as zero in scoring.

### Optional Local Seed Data

Use this only when you do not yet have Product Service/event pipeline data but want to verify Task 5 locally.

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" --eval '
const d = db.getSiblingDB("recommendation_db");
const now = new Date();
d.product_features.bulkWrite([
  {
    updateOne: {
      filter: { product_id: "prod_demo_1" },
      update: {
        $set: {
          product_id: "prod_demo_1",
          category_id: "cat_demo",
          seller_id: "seller_demo",
          status: "active",
          stock_status: "in_stock",
          quality_flags: { is_recommendable: true, is_adult: false, is_deleted: false },
          counters: {
            views_24h: 100,
            wishlist_adds_7d: 10,
            cart_adds_7d: 8,
            purchases_7d: 5
          },
          feature_version: 1,
          last_event_at: now,
          created_at: now,
          updated_at: now
        }
      },
      upsert: true
    }
  },
  {
    updateOne: {
      filter: { product_id: "prod_demo_2" },
      update: {
        $set: {
          product_id: "prod_demo_2",
          category_id: "cat_demo",
          seller_id: "seller_demo",
          status: "active",
          stock_status: "in_stock",
          quality_flags: { is_recommendable: true, is_adult: false, is_deleted: false },
          counters: {
            views_24h: 60,
            wishlist_adds_7d: 4,
            cart_adds_7d: 3,
            purchases_7d: 2
          },
          feature_version: 1,
          last_event_at: now,
          created_at: now,
          updated_at: now
        }
      },
      upsert: true
    }
  }
]);
'
```

Remove demo records later if they should not be part of local testing:

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" --eval '
db.getSiblingDB("recommendation_db").product_features.deleteMany({
  product_id: { $in: ["prod_demo_1", "prod_demo_2"] }
});
'
```

### Verify MongoDB

Check indexes:

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" \
  --eval 'db.getSiblingDB("recommendation_db").product_features.getIndexes().map(i => i.name)'
```

Expected task-specific index names:

```text
idx_product_features_rule_based_global
idx_product_features_rule_based_category
idx_product_features_rule_based_seller
```

After service startup, check generated sets:

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" \
  --eval 'db.getSiblingDB("recommendation_db").recommendation_sets.find({ recommendation_type: "trending" }, { _id: 1, context_key: 1, strategy_id: 1, items: 1, metadata: 1, expires_at: 1 }).pretty()'
```

Expected context examples:

```text
home:global
category:cat_demo
seller:seller_demo
```

---

## 6. Redis / Queue / External Services

### Redis

Redis setup is already documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
Section: `6. Redis / Queue / External Services`
```

Task 5 reuses the same Redis client and cache TTL config.

| Redis Detail | Value |
|---|---|
| Default prefix | `reco:v1` |
| Default TTL | `900` seconds / `15m` |
| Value type | JSON string |
| Mandatory? | No, but recommended for fast reads |

Implemented task-specific cache keys include strategy suffixes:

| List | Redis Key Pattern |
|---|---|
| Global trending | `reco:v1:trending:global:strategy:trending_v1_recent_activity` |
| Category popular | `reco:v1:trending:category:{category_id}:strategy:trending_v1_fallback_category` |
| Seller popular | `reco:v1:trending:seller:{seller_id}:strategy:trending_v1_recent_activity` |

Verify Redis after the ranking job has run:

```bash
redis-cli --scan --pattern 'reco:v1:trending:*'
redis-cli TTL 'reco:v1:trending:global:strategy:trending_v1_recent_activity'
redis-cli GET 'reco:v1:trending:global:strategy:trending_v1_recent_activity'
```

If Redis is not configured, the service can still save MongoDB `recommendation_sets`. Cache writes are best-effort.

### Kafka

No new Kafka topic is introduced by `TASK_FILE_NAME`.

Kafka setup is reused only if you want live interaction events to update feature counters:

```md
`TaskImplementation/${SERVICE_NAME}/task3_Dependency.md`
Section: `6. Redis / Queue / External Services`
```

For manual Task 5 verification with seeded `product_features`, Kafka is not required.

### Product Data Dependency

Product data is conceptually required because Task 5 cannot rank products without `product_features`.

Important beginner explanation:

Product Service ya event-feature pipeline ko `product_features` me product id, category id, seller id, stock status, active status, quality flags, and counters populate karne honge. Task 5 direct Product Service database query nahi karta.

---

## 7. Environment Variables

Base `.env` setup and loading behavior are already documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `4. Environment Variables`
```

Credentials and reused MongoDB/Redis variables are already documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
Section: `7. Environment Variables`
```

Task 4 feature variables are already documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task4_Dependency.md`
Section: `7. Environment Variables`
```

Current task `.env` file location:

```text
backend/services/recommendation-service/.env
```

Important: The Go service reads environment variables from the process using `os.Getenv`. It does not automatically load `.env`. Source the file before running the service, or configure your process manager/Docker environment to pass these variables.

```bash
cd backend/services/recommendation-service
set -a
source .env
set +a
```

### Current Task `.env` Delta

Add or verify only these Task 5 variables:

```dotenv
# Task 5 rule-based top-list generation
RECOMMENDATION_RANKING_ENABLED=true
RECOMMENDATION_RANKING_REBUILD_INTERVAL=15m
RECOMMENDATION_RANKING_FORMULA_VERSION=popularity_v1
RECOMMENDATION_RANKING_VIEW_24H_WEIGHT=1
RECOMMENDATION_RANKING_WISHLIST_7D_WEIGHT=3
RECOMMENDATION_RANKING_CART_7D_WEIGHT=4
RECOMMENDATION_RANKING_PURCHASE_7D_WEIGHT=8
```

### Variable Reference

| Variable | Required? | Default | Purpose | Notes |
|---|---:|---|---|---|
| `RECOMMENDATION_RANKING_ENABLED` | Yes for Task 5 | `true` | Ranking job enable/disable karta hai. | `false` karoge to generated trending lists nahi banenge. |
| `RECOMMENDATION_RANKING_REBUILD_INTERVAL` | Yes when enabled | `15m` | Initial generation ke baad periodic rebuild interval. | Must be greater than zero when ranking is enabled. |
| `RECOMMENDATION_RANKING_FORMULA_VERSION` | Yes | `popularity_v1` | Ranking formula version metadata me store hota hai. | Weights change karo to version bhi change karo, for example `popularity_v2`. |
| `RECOMMENDATION_RANKING_VIEW_24H_WEIGHT` | Yes | `1` | Recent views ka score weight. | Non-negative integer. |
| `RECOMMENDATION_RANKING_WISHLIST_7D_WEIGHT` | Yes | `3` | Wishlist signal ka score weight. | Non-negative integer. |
| `RECOMMENDATION_RANKING_CART_7D_WEIGHT` | Yes | `4` | Cart signal ka score weight. | Non-negative integer. |
| `RECOMMENDATION_RANKING_PURCHASE_7D_WEIGHT` | Yes | `8` | Purchase signal ka score weight. | Non-negative integer. |

Validation rules:

- All weights cannot be zero.
- No weight can be negative.
- If weights differ from defaults, `RECOMMENDATION_RANKING_FORMULA_VERSION` cannot remain `popularity_v1`.
- `RECOMMENDATION_RANKING_REBUILD_INTERVAL` must be greater than zero when ranking is enabled.

Reused variables that must already be configured:

| Variable | Why Task 5 Needs It | Reuse Reference |
|---|---|---|
| `RECOMMENDATION_MONGO_URI` | MongoDB connection for `product_features` and `recommendation_sets`. | `task2_Dependency.md`, section `7. Environment Variables` |
| `RECOMMENDATION_MONGO_DATABASE` | Should point to `recommendation_db`. | `task2_Dependency.md`, section `7. Environment Variables` |
| `RECOMMENDATION_REDIS_ADDR` | Redis cache endpoint. | `task2_Dependency.md`, section `7. Environment Variables` |
| `RECOMMENDATION_CACHE_KEY_PREFIX` | Prefix for generated cache keys. | `task2_Dependency.md`, section `7. Environment Variables` |
| `RECOMMENDATION_CACHE_TTL_SECONDS` | TTL for generated task 5 lists. | `task2_Dependency.md`, section `7. Environment Variables` |
| `RECOMMENDATION_DEFAULT_LIMIT` | Number of items generated per set. | `task1_Dependency.md`, section `4. Environment Variables` |
| `RECOMMENDATION_MAX_IDENTIFIER_LENGTH` | Validates category/seller/cache identifiers. | `task1_Dependency.md`, section `4. Environment Variables` |

Security notes:

- Ranking weights are not secrets.
- MongoDB and Redis passwords are secrets. Do not commit real passwords.
- Production secrets should come from environment-specific secret management, not plain committed `.env`.

---

## 8. Docker Setup

No new Docker container is introduced by `TASK_FILE_NAME`.

Reuse previous Docker setup:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `7. Docker and DevOps Setup`
```

```md
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
Section: `8. Docker Setup`
```

```md
`TaskImplementation/${SERVICE_NAME}/task4_Dependency.md`
Section: `8. Docker Setup`
```

Task-specific Docker checklist:

| Container | Needed For Task 5? | Notes |
|---|---:|---|
| MongoDB | Yes | Must contain `product_features`; run migration `004`. |
| Redis | Recommended | Needed to verify hot cache keys. |
| Kafka | Only for live event pipeline | Not needed if using manual seed data. |

Docker networking reminder:

- Service running on host should use `localhost:27017` and `localhost:6379` if containers publish ports to host.
- Service running inside Docker should use container service names, not `localhost`, unless the DB/cache is inside the same container.
- Detailed Docker network troubleshooting is already covered in previous dependency files.

---

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Follow previous docs first:

```text
TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
TaskImplementation/${SERVICE_NAME}/task3_Dependency.md
TaskImplementation/${SERVICE_NAME}/task4_Dependency.md
```

### Step 2: Go To Project Directory

```bash
cd backend/services/recommendation-service
```

### Step 3: Install Dependencies

No new dependency install is required for `TASK_FILE_NAME`.

```bash
go mod download
```

### Step 4: Start Required Services

Start MongoDB and Redis using the reused setup from previous dependency files.

For manual Task 5 verification:

```text
MongoDB: required
Redis: recommended
Kafka: optional
```

For full live feature generation:

```text
MongoDB replica set: required by Task 4 feature transactions
Kafka: required by Task 3 event ingestion
Redis: recommended
```

### Step 5: Add Current Task Environment Variables

Update `backend/services/recommendation-service/.env` with the Task 5 variables from section `7. Environment Variables`.

Then load the file:

```bash
set -a
source .env
set +a
```

### Step 6: Run Migrations

Run migrations `001` to `004` if this is a fresh local DB.

If migrations `001` to `003` are already applied, run only:

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" \
  --file migrations/004_add_rule_based_ranking_indexes.up.js
```

### Step 7: Ensure Product Features Exist

Task 5 needs eligible `product_features`.

You have two options:

| Option | Use When |
|---|---|
| Live feature pipeline | Task 3/4 are running and events are already creating product features. |
| Manual seed data | You want quick local verification without Kafka/events. |

If needed, use the seed data from section `5. Database Setup` -> `Optional Local Seed Data`.

### Step 8: Start Backend Service

```bash
go run ./cmd/server
```

Expected startup behavior:

- Service loads ranking config.
- If MongoDB is configured, ranking service initializes.
- Ranking service runs one initial generation.
- Then it rebuilds every `RECOMMENDATION_RANKING_REBUILD_INTERVAL`.
- If Redis is configured, generated sets are cached.

Expected useful log messages:

```text
recommendation.rule_based_set_generated
recommendation.ranking.completed
```

### Step 9: Verify HTTP Health And Storage

```bash
curl http://localhost:8088/healthz
curl http://localhost:8088/internal/v1/recommendation/storage/status
```

Check ranking metrics:

```bash
curl http://localhost:8088/metrics | grep recommendation_ranking
```

Expected metric names include:

```text
recommendation_ranking_job_runs_total
recommendation_ranking_job_duration_ms
recommendation_ranking_candidates_total
recommendation_ranking_items_generated_total
```

### Step 10: Verify MongoDB Generated Sets

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" \
  --eval 'db.getSiblingDB("recommendation_db").recommendation_sets.find({ recommendation_type: "trending" }, { _id: 1, context_key: 1, strategy_id: 1, items: 1, expires_at: 1 }).pretty()'
```

Expected:

- `home:global` set exists.
- `category:{category_id}` sets exist when eligible products have `category_id`.
- `seller:{seller_id}` sets exist when eligible products have `seller_id`.

### Step 11: Verify Redis Cache

```bash
redis-cli --scan --pattern 'reco:v1:trending:*'
redis-cli TTL 'reco:v1:trending:global:strategy:trending_v1_recent_activity'
```

Expected TTL should be close to `900` seconds right after generation.

### Step 12: Optional gRPC Serving Verification

Run from repo root if `grpcurl` is installed:

```bash
grpcurl -plaintext \
  -import-path proto \
  -proto ecommerce/recommendation/v1/recommendation.proto \
  -d '{"context":"RECOMMENDATION_CONTEXT_HOME_FEED","type":"RECOMMENDATION_TYPE_TRENDING","limit":5}' \
  localhost:9088 \
  ecommerce.recommendation.v1.RecommendationService/GetRecommendations
```

Category scoped example:

```bash
grpcurl -plaintext \
  -import-path proto \
  -proto ecommerce/recommendation/v1/recommendation.proto \
  -d '{"context":"RECOMMENDATION_CONTEXT_CATEGORY_LISTING","type":"RECOMMENDATION_TYPE_TRENDING","category_id":"cat_demo","limit":5}' \
  localhost:9088 \
  ecommerce.recommendation.v1.RecommendationService/GetRecommendations
```

Seller scoped example:

```bash
grpcurl -plaintext \
  -import-path proto \
  -proto ecommerce/recommendation/v1/recommendation.proto \
  -d '{"context":"RECOMMENDATION_CONTEXT_SELLER_STORE","type":"RECOMMENDATION_TYPE_TRENDING","seller_id":"seller_demo","limit":5}' \
  localhost:9088 \
  ecommerce.recommendation.v1.RecommendationService/GetRecommendations
```

---

## 10. Running the Project

Base run instructions are reused from:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `8. Project Run Instructions`
```

Task-specific run flow:

1. Start MongoDB.
2. Start Redis if cache verification is needed.
3. Load `.env`.
4. Run migration `004` if not already applied.
5. Confirm eligible `product_features` exist.
6. Run service.
7. Check logs, MongoDB generated sets, Redis keys, and ranking metrics.

Minimal run:

```bash
cd backend/services/recommendation-service
set -a
source .env
set +a
go run ./cmd/server
```

Disable only Task 5 ranking if needed:

```dotenv
RECOMMENDATION_RANKING_ENABLED=false
```

Warning: Disabling ranking means no new global/category/seller generated top lists will be created by the service.

---

## 11. Common Errors & Fixes

Generic setup errors are already documented in previous dependency files. Below are only Task 5-specific issues.

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `RECOMMENDATION_RANKING_REBUILD_INTERVAL must be greater than zero` | Ranking enabled but interval is `0`, negative, or invalid. | Set `RECOMMENDATION_RANKING_REBUILD_INTERVAL=15m`. | Keep duration format valid, like `15m`, `5m`, `1h`. |
| `RECOMMENDATION_RANKING_FORMULA_VERSION must change when ranking weights differ from popularity_v1` | Weights changed but formula version stayed default. | Change version, for example `RECOMMENDATION_RANKING_FORMULA_VERSION=popularity_v2`. | Version every scoring formula change. |
| `popularity weights cannot be negative` | One ranking weight is below zero. | Set all weight values to `0` or greater. | Treat weights as tuning config, not arbitrary numbers. |
| `at least one popularity weight must be greater than zero` | All four ranking weights are `0`. | Restore defaults or make at least one weight positive. | Keep one meaningful ranking signal enabled. |
| No `recommendation_sets` generated | `product_features` empty or all products fail eligibility. | Seed valid product features or run Task 3/4 pipeline. | Ensure Product Service projection includes required fields. |
| Product exists but not ranked | Current code expects `stock_status: "in_stock"`, not only `in_stock: true`. | Update product feature projection to include `stock_status`. | Keep schema aligned with `internal/domain/feature.go`. |
| Redis key not found | Redis not configured, cache write failed, job not run yet, or checking old key without strategy suffix. | Check Mongo sets first, then check `reco:v1:trending:global:strategy:trending_v1_recent_activity`. | Use implemented key patterns from section `6`. |
| Only global list exists, no category/seller list | Candidate docs do not have `category_id` or `seller_id`. | Add those fields to `product_features`. | Product feature projection should always include category and seller IDs when known. |
| Generated sets expire too quickly | `RECOMMENDATION_CACHE_TTL_SECONDS` too low. | Restore `900` or choose a sane TTL. | Keep rebuild interval and TTL coordinated. |
| Task 5 job logs `mongo storage is not configured` | `RECOMMENDATION_MONGO_URI` empty or Mongo init failed. | Configure MongoDB and check `/internal/v1/recommendation/storage/status`. | Keep storage env loaded before `go run`. |
| Migration `004` indexes missing | Migration not run or wrong database selected. | Set `RECOMMENDATION_MONGO_DATABASE=recommendation_db` and rerun migration `004`. | Verify indexes after every fresh DB setup. |

---

## 12. Security & Best Practices

Task-specific best practices:

- Keep `recommendation_sets` and Redis cache free of raw PII. Product IDs, score, rank, and reason are enough.
- Do not expose Redis to the public internet. Redis cache has product behavior signals and should be private.
- Do not commit real MongoDB or Redis credentials.
- If product is deleted, hidden, adult-restricted, blocked, or unsafe, upstream projection should set `quality_flags.is_recommendable=false` or `quality_flags.is_deleted=true`.
- Keep `status` and `stock_status` fresh. Out-of-stock products should not remain recommendable.
- Version formula changes using `RECOMMENDATION_RANKING_FORMULA_VERSION`.
- Keep TTL and rebuild interval aligned. Default `15m` TTL plus `15m` rebuild interval is simple for local/dev.
- Monitor empty set metrics. Empty trending list usually means missing product features, stale counters, or eligibility mismatch.
- Prefer migration-driven indexes in production. App-time `RECOMMENDATION_MONGO_ENSURE_INDEXES=true` is useful locally, but production teams often control indexes through migrations.
- Do not query another service database directly from Task 5. Use `product_features` projection instead.

---

## 13. Missing or Misconfigured Things

Audit notes for `TASK_FILE_NAME` setup:

| Item | Status | Recommendation |
|---|---|---|
| Manual ranking trigger endpoint | Not present | Ranking runs on startup and interval. For production ops, consider a protected admin trigger later. |
| Ranking freshness health check | Not present as a separate health endpoint | Use `recommendation_sets.generated_at`, `expires_at`, logs, and metrics for now. |
| `.env` auto-load | Not present | Source `.env` manually or use Docker/process manager env injection. |
| Product feature freshness | External dependency | Task 5 depends on Task 3/4 pipeline or Product Service projection keeping `product_features` current. |
| `stock_status` field alignment | Important | Current code requires `stock_status="in_stock"`. Make sure upstream projection does not only write `in_stock: true`. |
| Redis cache source visibility in serving response | Limited | Verify cache via Redis CLI and metrics; gRPC response does not show internal source. |
| Production secret handling | Environment-specific | Use secret manager or deployment platform secrets, not committed `.env`. |

---

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `1. Project Tech Stack Analysis` | Base Go, gRPC, MongoDB, Redis, Kafka, Prometheus, Docker explanation already documented. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `2. Go Dependency System` | Go modules, `go.mod`, `go.sum`, `go mod download`, and common Go errors are unchanged. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `3. Databases` -> `MongoDB` | MongoDB install, Docker run, port, connection string, and credentials placement are reused. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `4. Environment Variables` | General `.env` setup and loading behavior is reused. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `6. Ports and Networking` | HTTP, gRPC, MongoDB, Redis, and Kafka ports are unchanged. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `7. Docker and DevOps Setup` | No new Docker container is needed by Task 5. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | `recommendation_db` and `recommendation_sets` storage setup is reused. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `6. Redis / Queue / External Services` | Redis cache setup, key prefix, TTL, and verification basics are reused. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `7. Environment Variables` | MongoDB and Redis env variables are reused. |
| `TaskImplementation/${SERVICE_NAME}/task3_Dependency.md` | `6. Redis / Queue / External Services` | Kafka event ingestion setup is reused only for live counter updates. |
| `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` | `5. Database Setup` | `product_features` schema, feature-store collections, and replica set requirement are reused. |
| `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` | `7. Environment Variables` | Feature builder env variables are reused when live product features are generated from events. |

---

## 15. Final Checklist

- [ ] Previous dependency documentation checked.
- [ ] No duplicate base install/Docker/MongoDB/Redis/Kafka documentation added.
- [ ] Migration `004_add_rule_based_ranking_indexes.up.js` run.
- [ ] `product_features` contains eligible products with `status="active"` and `stock_status="in_stock"`.
- [ ] `quality_flags.is_recommendable=true` and `quality_flags.is_deleted` is not `true`.
- [ ] Task 5 ranking env variables added or verified.
- [ ] `.env` sourced before running the service.
- [ ] MongoDB running and `recommendation_db` selected correctly.
- [ ] Redis running if cache verification is required.
- [ ] Backend service started successfully.
- [ ] Logs show `recommendation.rule_based_set_generated`.
- [ ] MongoDB `recommendation_sets` contains global, category, and seller trending sets where data exists.
- [ ] Redis keys checked with strategy suffix.
- [ ] `/metrics` contains `recommendation_ranking_*` metrics.
- [ ] Optional gRPC trending request verified.
- [ ] No real credentials committed.
- [ ] Demo seed data removed if not needed.
- [ ] No original implementation task file modified.
