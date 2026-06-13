# Project Dependency & Setup Guide

## Variables Used

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Recommendation Service` |
| `TASK_FILE_NAME` | `task6.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task6_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |
| Backend code path | `backend/services/recommendation-service` |

Beginner note: Ye file sirf dependency, setup, environment, database, Redis, DevOps, verification, and troubleshooting guide hai. Business logic ya original implementation file modify nahi kiya gaya.

---

## 1. Project Overview

`TASK_FILE_NAME` ka focus hai: `${SERVICE_NAME}` ke liye personalized ranking setup.

Simple Hinglish:

- Task 4 ke feature-store documents se user/guest behavior profile read hota hai.
- Task 6 profile ke category, seller, brand, price, direct product, and popularity signals ko score karta hai.
- Agar profile missing, weak, stale, ya candidate empty ho, Task 5 ke category/global popular sets fallback ke roop me use hote hain.
- Generated personalized result MongoDB `recommendation_sets` collection me short expiry ke saath save hota hai.
- Redis me same personalized result fast read ke liye cache hota hai.

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
`TaskImplementation/${SERVICE_NAME}/task4_Dependency.md`

Sections:
`5. Database Setup`
`7. Environment Variables`
`9. Local Development Setup`
`13. Missing or Misconfigured Things`
```

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task5_Dependency.md`

Sections:
`5. Database Setup`
`6. Redis / Queue / External Services`
`7. Environment Variables`
`9. Local Development Setup`
`11. Common Errors & Fixes`
```

`TASK_FILE_NAME` new setup scope:

- MongoDB migration: `backend/services/recommendation-service/migrations/005_add_personalized_ranking_indexes.up.js`
- Personalization environment variables:
  - `RECOMMENDATION_PERSONALIZATION_ENABLED`
  - `RECOMMENDATION_PERSONALIZED_FORMULA_VERSION`
  - `RECOMMENDATION_PERSONALIZED_MIN_POSITIVE_INTERACTIONS`
  - `RECOMMENDATION_PERSONALIZED_PROFILE_MAX_AGE_DAYS`
  - `RECOMMENDATION_PERSONALIZED_MAX_CANDIDATES`
  - `RECOMMENDATION_PERSONALIZED_MAX_ITEMS_PER_SELLER`
  - `RECOMMENDATION_PERSONALIZED_TOP_CATEGORIES`
  - `RECOMMENDATION_PERSONALIZED_TOP_SELLERS`
  - `RECOMMENDATION_PERSONALIZED_TOP_BRANDS`
  - `RECOMMENDATION_PERSONALIZED_DIRECT_PRODUCT_LIMIT`
  - six personalized scoring weight variables
- Implemented code paths:
  - `internal/domain/personalization.go`
  - `internal/usecase/personalized.go`
  - `internal/repository/mongo_personalized_repository.go`
  - `internal/repository/redis_recommendation_cache.go`
  - `internal/usecase/get_recommendations.go`
  - `cmd/server/main.go`
- New personalized metrics and structured logs.

No new database engine, queue engine, Docker container type, third-party SaaS, or public port is introduced by `TASK_FILE_NAME`.

---

## 2. Tech Stack

Base technologies are already explained in previous dependency files. Do not repeat their full installation here.

| Technology | New or Reused | Required? | Beginner Explanation | Current Task Use |
|---|---|---:|---|---|
| Go | Reused | Yes | Go backend language hai. Service, use case, repository, and tests Go me run hote hain. | Personalized ranker, scoring formula, fallback flow, and service wiring. |
| Go modules | Reused | Yes | `go.mod` and `go.sum` dependency versions manage karte hain. | No new module dependency added for `TASK_FILE_NAME`. |
| MongoDB | Reused storage, new index | Yes for full behavior | MongoDB document database hai. JSON-like collections me profile, products, and result sets store hote hain. | Reads `user_feature_profiles`, `user_product_counters`, `product_features`; writes `recommendation_sets`. |
| Redis | Reused cache | Optional but recommended | Redis in-memory cache hai jo fast data access ke liye use hota hai. | Personalized user/guest cache with short TTL. |
| Kafka | Reused from Task 3 | Optional for manual verification, required for live features | Kafka event stream hai. | Live user behavior events se Task 4 profiles fresh rakhne ke liye. |
| Prometheus client | Reused, new metrics | Recommended | Metrics monitoring ke liye Go library. | Personalized build, fallback, cache, profile miss, and latency metrics. |
| gRPC serving | Reused broader service path | Optional for verification | gRPC service-to-service API protocol hai. | Personalized result verify karne ke liye `GetRecommendations` RPC use ho sakta hai. |

Important implementation note:

```text
Personalized ranking raw event history ko request time par scan nahi karta.
It reads Task 4 derived feature collections and Task 5 fallback sets.
```

Not introduced by `TASK_FILE_NAME`:

| Technology | Status |
|---|---|
| MySQL / PostgreSQL / SQLite | Not used |
| RabbitMQ / NATS | Not used |
| Elasticsearch / vector database | Not used |
| TensorFlow / PyTorch / ML model server | Not used |
| AWS S3 / Firebase / SMTP / Stripe / Twilio / OAuth | Not used |
| Kubernetes / Nginx | Not present as task-specific setup |

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
| MongoDB `7.x` | Feature profiles, candidates, and recommendation sets | Reused |
| `mongosh` | Run migration `005` and verify indexes/data | Reused tool, new migration |
| Redis `7.x` | Personalized response cache | Reused |
| `redis-cli` | Verify personalized cache key and TTL | Reused |
| Kafka `3.x` | Live event-to-feature pipeline only | Reused from Task 3 |
| `curl` | Health and metrics checks | Reused |
| `grpcurl` | Optional personalized RPC verification | Reused optional tool |

Beginner decision:

- Sirf Task 6 setup verify karna hai: MongoDB plus optional Redis enough hai, but seed data or existing Task 4/5 data chahiye.
- Real live personalization verify karna hai: Task 3 Kafka ingestion, Task 4 feature builder, and Task 5 fallback ranking setup bhi running hona chahiye.

---

## 4. Dependency Management

This is still a Go module project. Go dependency setup is already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `2. Go Dependency System`
```

`TASK_FILE_NAME` dependency status:

| Dependency | Version In `go.mod` | Used For | New Install Needed? |
|---|---:|---|---:|
| Go standard library `sort`, `time`, `strings`, `context`, `log/slog` | Go built-in | Stable scoring order, TTL, identity cleanup, logs | No |
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | Read feature profiles/candidates and upsert recommendation sets | No |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Personalized cache read/write | No |
| `github.com/prometheus/client_golang` | `v1.23.2` | Personalized metrics | No |
| `google.golang.org/grpc` | `v1.81.1` | Optional serving verification | No |

No new `go get` command is needed for `TASK_FILE_NAME`.

Task-specific verification:

```bash
cd backend/services/recommendation-service
go test ./internal/domain ./internal/usecase ./internal/repository
```

Full service verification:

```bash
cd backend/services/recommendation-service
go test ./...
```

Common Go module errors like old Go version, missing `go.sum`, proxy failure, or local `proto-gen/go` replace path issues are already documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `2. Go Dependency System` -> `Common Go module issues`
```

---

## 5. Database Setup

### MongoDB Role

MongoDB is mandatory for full `TASK_FILE_NAME` behavior.

Simple Hinglish:

MongoDB me user behavior ka compact profile already Task 4 se aata hai. Task 6 us profile ko read karke personalized products rank karta hai, aur generated result ko `recommendation_sets` me store karta hai. Redis miss ho to MongoDB last valid non-expired personalized set recover karne me help karta hai.

### Reused MongoDB Setup

MongoDB installation, Docker setup, default port, connection string format, and credentials placement are already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `3. Databases` -> `MongoDB`
```

```md
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
Section: `5. Database Setup`
```

Feature-store schema and replica set requirement are already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task4_Dependency.md`
Section: `5. Database Setup`
```

Do not recreate MongoDB if it is already running.

### Current Task Migration

Prerequisite migrations:

1. `001_create_recommendation_storage.up.js`
2. `002_add_event_ingestion_indexes.up.js`
3. `003_create_feature_store_schema.up.js`
4. `004_add_rule_based_ranking_indexes.up.js`

Run only the new migration if previous migrations are already applied:

```bash
cd backend/services/recommendation-service
export RECOMMENDATION_MONGO_DATABASE=recommendation_db

mongosh "mongodb://root:secret@localhost:27017/admin" \
  --file migrations/005_add_personalized_ranking_indexes.up.js
```

If your local MongoDB has no auth:

```bash
cd backend/services/recommendation-service
export RECOMMENDATION_MONGO_DATABASE=recommendation_db

mongosh "mongodb://localhost:27017" \
  --file migrations/005_add_personalized_ranking_indexes.up.js
```

Fresh local database flow:

```bash
cd backend/services/recommendation-service
export RECOMMENDATION_MONGO_DATABASE=recommendation_db

mongosh "mongodb://root:secret@localhost:27017/admin" --file migrations/001_create_recommendation_storage.up.js
mongosh "mongodb://root:secret@localhost:27017/admin" --file migrations/002_add_event_ingestion_indexes.up.js
mongosh "mongodb://root:secret@localhost:27017/admin" --file migrations/003_create_feature_store_schema.up.js
mongosh "mongodb://root:secret@localhost:27017/admin" --file migrations/004_add_rule_based_ranking_indexes.up.js
mongosh "mongodb://root:secret@localhost:27017/admin" --file migrations/005_add_personalized_ranking_indexes.up.js
```

### Collections Used By Current Task

| Collection | Status | Purpose |
|---|---|---|
| `user_feature_profiles` | Reused from Task 4 | Broad user/guest affinities: category, seller, brand, price, recent products. |
| `user_product_counters` | Reused from Task 4 | Direct product behavior and negative feedback signals. |
| `product_features` | Reused from Task 4/5, new brand index | Eligible personalized candidates read karne ke liye. |
| `recommendation_sets` | Reused from Task 2/5 | Personalized output and fallback top lists store karne ke liye. |

### Current Task Index

Migration `005` adds this index:

| Collection | Index | Why Needed |
|---|---|---|
| `product_features` | `idx_product_features_personalized_brand` | Brand-affinity candidate query ko active, in-stock, recommendable product filters ke saath fast banane ke liye. |

Index fields:

```text
brand_id
status
stock_status
quality_flags.is_recommendable
quality_flags.is_deleted
counters.purchases_7d
product_id
```

Category and seller candidate filters reuse Task 4/5 indexes:

```md
`TaskImplementation/${SERVICE_NAME}/task4_Dependency.md`
Section: `5. Database Setup` -> `Current Task Indexes`

`TaskImplementation/${SERVICE_NAME}/task5_Dependency.md`
Section: `5. Database Setup` -> `Current Task Indexes`
```

### Required Input Documents

Task 6 cannot create useful personalized ranking unless these inputs exist.

| Input | Required? | Expected Fields |
|---|---:|---|
| `user_feature_profiles` | Yes for true personalization | `profile_key`, `category_scores`, `seller_scores`, `brand_scores`, `price_affinity.preferred_buckets`, `recent_products`, `last_event_at` |
| `user_product_counters` | Strongly recommended | `profile_key`, `product_id`, `weighted_score_input`, `counters.views`, `counters.wishlist_adds`, `counters.wishlist_removes`, `counters.cart_adds`, `counters.purchases`, `last_interaction_at` |
| `product_features` | Yes | `product_id`, `category_id`, `seller_id`, `brand_id`, `price.bucket`, `status`, `stock_status`, `quality_flags.is_recommendable`, `quality_flags.is_deleted`, popularity counters |
| Task 5 fallback `recommendation_sets` | Required for cold-start fallback | `home:global` or `category:{category_id}` generated set with non-expired `expires_at` |

Important: Current implementation expects `stock_status: "in_stock"` and `quality_flags.is_recommendable: true`. Agar product doc me sirf `in_stock: true` ho, personalized candidate eligible nahi banega.

### Verify MongoDB

Check migration `005` index:

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" \
  --eval 'db.getSiblingDB("recommendation_db").product_features.getIndexes().filter(i => i.name === "idx_product_features_personalized_brand")'
```

Check that at least one profile exists:

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" \
  --eval 'db.getSiblingDB("recommendation_db").user_feature_profiles.find({}, { profile_key: 1, last_event_at: 1 }).limit(5).pretty()'
```

Check direct counters:

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" \
  --eval 'db.getSiblingDB("recommendation_db").user_product_counters.find({}, { profile_key: 1, product_id: 1, weighted_score_input: 1 }).limit(5).pretty()'
```

Check personalized generated sets after service verification:

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" \
  --eval 'db.getSiblingDB("recommendation_db").recommendation_sets.find({ recommendation_type: "personalized" }, { _id: 1, context_key: 1, strategy_id: 1, metadata: 1, expires_at: 1 }).pretty()'
```

### Optional Local Seed Data

Use this only when you do not yet have Task 3/4 live events and Task 5 generated fallback data. This is local demo data, not production data.

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" --eval '
const d = db.getSiblingDB("recommendation_db");
const now = new Date();
const expires = new Date(now.getTime() + 5 * 60 * 1000);

d.user_feature_profiles.updateOne(
  { profile_key: "user:user_demo" },
  {
    $set: {
      profile_key: "user:user_demo",
      user_id: "user_demo",
      category_scores: { cat_demo: 10 },
      seller_scores: { seller_demo: 4 },
      brand_scores: { brand_demo: 3 },
      price_affinity: { preferred_buckets: ["1000_1999"] },
      recent_products: [{ product_id: "prod_demo_1", occurred_at: now }],
      last_event_at: now
    }
  },
  { upsert: true }
);

d.user_product_counters.updateOne(
  { profile_key: "user:user_demo", product_id: "prod_demo_1" },
  {
    $set: {
      profile_key: "user:user_demo",
      user_id: "user_demo",
      product_id: "prod_demo_1",
      weighted_score_input: NumberLong(10),
      counters: {
        views: NumberLong(3),
        wishlist_adds: NumberLong(1),
        wishlist_removes: NumberLong(0),
        cart_adds: NumberLong(1),
        purchases: NumberLong(0)
      },
      last_interaction_at: now
    }
  },
  { upsert: true }
);

d.product_features.updateOne(
  { product_id: "prod_demo_1" },
  {
    $set: {
      product_id: "prod_demo_1",
      category_id: "cat_demo",
      seller_id: "seller_demo",
      brand_id: "brand_demo",
      price: { bucket: "1000_1999" },
      status: "active",
      stock_status: "in_stock",
      quality_flags: { is_recommendable: true, is_deleted: false },
      counters: {
        views_24h: NumberLong(5),
        wishlist_adds_7d: NumberLong(2),
        cart_adds_7d: NumberLong(1),
        purchases_7d: NumberLong(1)
      }
    }
  },
  { upsert: true }
);

d.recommendation_sets.updateOne(
  { context_key: "home:global", strategy_id: "trending_v1_recent_activity" },
  {
    $set: {
      context_key: "home:global",
      recommendation_type: "trending",
      strategy_id: "trending_v1_recent_activity",
      items: [{ product_id: "prod_demo_1", score: 50, rank: 1, reason: "demo_global_trending" }],
      generated_at: now,
      expires_at: expires
    },
    $setOnInsert: { _id: "reco_demo_global" }
  },
  { upsert: true }
);
'
```

---

## 6. Redis / Queue / External Services

### Redis

Redis setup is already documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
Section: `6. Redis / Queue / External Services`
```

Task 6 reuses the same Redis client and cache config. No new Redis server setup is required.

| Redis Detail | Value |
|---|---|
| Default prefix | `reco:v1` |
| Personalized TTL | `300` seconds / `5m` |
| Guest TTL | `300` seconds / `5m` |
| Value type | JSON string |
| Mandatory? | No, but recommended for fast reads |

Implemented personalized cache keys include strategy suffixes:

| Feed | Redis Key Pattern |
|---|---|
| Logged-in personalized | `reco:v1:personalized:user:{user_id}:strategy:personalized_v1_behavior` |
| Anonymous personalized | `reco:v1:personalized:anon:{anonymous_id}:strategy:personalized_v1_behavior` |

Verify Redis after a personalized request has run:

```bash
redis-cli --scan --pattern 'reco:v1:personalized:*'
redis-cli TTL 'reco:v1:personalized:user:user_demo:strategy:personalized_v1_behavior'
redis-cli GET 'reco:v1:personalized:user:user_demo:strategy:personalized_v1_behavior'
```

If Redis is not configured, the service can still build and save MongoDB `recommendation_sets`. Cache reads/writes are best-effort.

### Kafka

No new Kafka topic is introduced by `TASK_FILE_NAME`.

Kafka is needed only when you want live behavior events to keep Task 4 feature profiles fresh:

```md
`TaskImplementation/${SERVICE_NAME}/task3_Dependency.md`
Section: `6. Redis / Queue / External Services`
```

For manual Task 6 verification with seeded profile/product/fallback data, Kafka is not required.

### Product Data Dependency

Task 6 needs product feature projection data. It does not query Product Service database directly.

Beginner explanation:

Product Service ya event-feature pipeline ko `product_features` me product id, category id, seller id, brand id, price bucket, active status, stock status, quality flags, and counters populate karne honge. Agar ye fields missing hain, personalized ranker fallback par chala jayega ya empty result de sakta hai.

### Ports and Networking

Detailed networking explanation is already documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `6. Ports and Networking`
```

Current task port delta:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| HTTP API | `8088` | Health, metrics, internal storage/status endpoints | Reused |
| gRPC API | `9088` | Optional `GetRecommendations` verification | Reused |
| MongoDB | `27017` | Feature profiles, candidates, generated sets | Reused |
| Redis | `6379` | Personalized cache | Reused |
| Kafka | `9092` | Live event ingestion for feature freshness | Reused, optional |

No new port is introduced by `TASK_FILE_NAME`.

---

## 7. Environment Variables

Base `.env` setup and loading behavior are already documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `4. Environment Variables`
```

MongoDB, Redis, and cache credential placement are already documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
Section: `7. Environment Variables`
```

Current task `.env` file location:

```text
backend/services/recommendation-service/.env
```

Important: The Go service reads environment variables from the process using `os.Getenv`. It does not automatically load `.env`. Source the file before running the service, or pass variables through Docker/process manager config.

```bash
cd backend/services/recommendation-service
set -a
source .env
set +a
```

### Current Task `.env` Delta

Add or verify only these Task 6 variables:

```dotenv
# Task 6 personalized ranking
RECOMMENDATION_PERSONALIZATION_ENABLED=true
RECOMMENDATION_PERSONALIZED_FORMULA_VERSION=behavior_v1
RECOMMENDATION_PERSONALIZED_MIN_POSITIVE_INTERACTIONS=3
RECOMMENDATION_PERSONALIZED_PROFILE_MAX_AGE_DAYS=30
RECOMMENDATION_PERSONALIZED_MAX_CANDIDATES=250
RECOMMENDATION_PERSONALIZED_MAX_ITEMS_PER_SELLER=3
RECOMMENDATION_PERSONALIZED_TOP_CATEGORIES=3
RECOMMENDATION_PERSONALIZED_TOP_SELLERS=2
RECOMMENDATION_PERSONALIZED_TOP_BRANDS=2
RECOMMENDATION_PERSONALIZED_DIRECT_PRODUCT_LIMIT=20
RECOMMENDATION_PERSONALIZED_CATEGORY_WEIGHT=0.35
RECOMMENDATION_PERSONALIZED_SELLER_WEIGHT=0.15
RECOMMENDATION_PERSONALIZED_BRAND_WEIGHT=0.10
RECOMMENDATION_PERSONALIZED_PRICE_WEIGHT=0.10
RECOMMENDATION_PERSONALIZED_DIRECT_AFFINITY_WEIGHT=0.20
RECOMMENDATION_PERSONALIZED_POPULARITY_WEIGHT=0.10
```

Reused cache TTL variables from Task 2:

```dotenv
RECOMMENDATION_PERSONALIZED_CACHE_TTL_SECONDS=300
RECOMMENDATION_GUEST_CACHE_TTL_SECONDS=300
```

### Variable Reference

| Variable | Required? | Default | Purpose | Security/config note |
|---|---:|---|---|---|
| `RECOMMENDATION_PERSONALIZATION_ENABLED` | Yes for Task 6 | `true` | Personalized ranker on/off switch. | Set `false` only for debugging degraded mode. |
| `RECOMMENDATION_PERSONALIZED_FORMULA_VERSION` | Yes | `behavior_v1` | Formula version stored in metadata. | If any weight changes, version must change, for example `behavior_v2`. |
| `RECOMMENDATION_PERSONALIZED_MIN_POSITIVE_INTERACTIONS` | Yes | `3` | Minimum positive interactions before profile is trusted. | Avoids over-personalizing from one accidental click. |
| `RECOMMENDATION_PERSONALIZED_PROFILE_MAX_AGE_DAYS` | Yes | `30` | Max profile freshness window. | Old profiles fallback instead of serving stale preferences. |
| `RECOMMENDATION_PERSONALIZED_MAX_CANDIDATES` | Yes | `250` | Max product candidates read before scoring. | Must be greater than or equal to `RECOMMENDATION_DEFAULT_LIMIT`. |
| `RECOMMENDATION_PERSONALIZED_MAX_ITEMS_PER_SELLER` | Yes | `3` | Diversity guard per seller. | Prevents one seller from dominating top results. |
| `RECOMMENDATION_PERSONALIZED_TOP_CATEGORIES` | Yes | `3` | Top profile categories used for candidate query. | Must be greater than zero when enabled. |
| `RECOMMENDATION_PERSONALIZED_TOP_SELLERS` | Yes | `2` | Top profile sellers used for candidate query. | Must be greater than zero when enabled. |
| `RECOMMENDATION_PERSONALIZED_TOP_BRANDS` | Yes | `2` | Top profile brands used for candidate query. | Must be greater than zero when enabled. |
| `RECOMMENDATION_PERSONALIZED_DIRECT_PRODUCT_LIMIT` | Yes | `20` | Recent/direct product counters included in candidate query. | Must be greater than zero when enabled. |
| `RECOMMENDATION_PERSONALIZED_CATEGORY_WEIGHT` | Yes | `0.35` | Category affinity weight. | All six weights must total exactly `1.00`. |
| `RECOMMENDATION_PERSONALIZED_SELLER_WEIGHT` | Yes | `0.15` | Seller affinity weight. | Non-negative finite number. |
| `RECOMMENDATION_PERSONALIZED_BRAND_WEIGHT` | Yes | `0.10` | Brand affinity weight. | Non-negative finite number. |
| `RECOMMENDATION_PERSONALIZED_PRICE_WEIGHT` | Yes | `0.10` | Preferred price bucket weight. | Non-negative finite number. |
| `RECOMMENDATION_PERSONALIZED_DIRECT_AFFINITY_WEIGHT` | Yes | `0.20` | Direct product behavior weight. | Non-negative finite number. |
| `RECOMMENDATION_PERSONALIZED_POPULARITY_WEIGHT` | Yes | `0.10` | Product popularity baseline weight. | Non-negative finite number. |

Reused variables that must already be configured:

| Variable | Why Task 6 Needs It | Reuse Reference |
|---|---|---|
| `RECOMMENDATION_MONGO_URI` | Read profiles/candidates/fallback sets and save personalized sets. | `task2_Dependency.md`, section `7. Environment Variables` |
| `RECOMMENDATION_MONGO_DATABASE` | Should point to `recommendation_db`. | `task2_Dependency.md`, section `7. Environment Variables` |
| `RECOMMENDATION_REDIS_ADDR` | Personalized cache endpoint. | `task2_Dependency.md`, section `7. Environment Variables` |
| `RECOMMENDATION_CACHE_KEY_PREFIX` | Prefix for personalized cache keys. | `task2_Dependency.md`, section `7. Environment Variables` |
| `RECOMMENDATION_DEFAULT_LIMIT` | Default personalized item count. | `task1_Dependency.md`, section `4. Environment Variables` |
| `RECOMMENDATION_MAX_LIMIT` | Maximum request limit. | `task1_Dependency.md`, section `4. Environment Variables` |
| `RECOMMENDATION_MAX_IDENTIFIER_LENGTH` | Validates user, anonymous, category, and cache identifiers. | `task1_Dependency.md`, section `4. Environment Variables` |
| `RECOMMENDATION_FEATURES_ENABLED` | Required for live feature building. | `task4_Dependency.md`, section `7. Environment Variables` |
| `RECOMMENDATION_RANKING_ENABLED` | Required to keep Task 5 fallback sets fresh. | `task5_Dependency.md`, section `7. Environment Variables` |

Validation rules:

- `RECOMMENDATION_PERSONALIZED_MAX_CANDIDATES` must be greater than or equal to `RECOMMENDATION_DEFAULT_LIMIT`.
- `RECOMMENDATION_PERSONALIZED_MAX_ITEMS_PER_SELLER` must be greater than zero.
- Top category/seller/brand/direct product limits must be greater than zero.
- All six personalized weights must total exactly `1.00`.
- If weights differ from defaults, `RECOMMENDATION_PERSONALIZED_FORMULA_VERSION` cannot remain `behavior_v1`.
- Personalized and guest cache TTL values must be greater than zero.

Security notes:

- Personalization weights are not secrets.
- MongoDB URI and Redis password are secrets if they contain credentials.
- Do not commit real `.env` files.
- Avoid logging raw user behavior history. Logs should use aggregate counts and `identity_type` where possible.

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

If live feature updates are enabled, reuse Task 4 MongoDB replica set guidance:

```md
`TaskImplementation/${SERVICE_NAME}/task4_Dependency.md`
Section: `5. Database Setup` -> `Replica Set Requirement`
```

Docker delta for Task 6:

| Docker Item | Status |
|---|---|
| New application container | Not introduced |
| New MongoDB container | Not needed |
| New Redis container | Not needed |
| New Kafka container | Not needed |
| New Docker volume | Not needed |
| New network | Not needed |
| New health check | Not introduced |

Recommended local dependency state:

| Mode | MongoDB | Redis | Kafka | Notes |
|---|---|---|---|---|
| Manual seeded verification | Required | Recommended | Not required | Use optional seed data from section `5`. |
| Live personalization | Required, replica set preferred | Recommended | Required if events are enabled | Follow Task 3/4 setup. |
| Minimal no-personalization run | Optional | Optional | Optional | Set `RECOMMENDATION_PERSONALIZATION_ENABLED=false`. |

---

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Follow previous docs first:

```text
TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
TaskImplementation/${SERVICE_NAME}/task4_Dependency.md
TaskImplementation/${SERVICE_NAME}/task5_Dependency.md
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

Start MongoDB and Redis using reused setup from previous dependency files.

For manual Task 6 verification:

```text
MongoDB: required
Redis: recommended
Kafka: optional
```

For full live behavior-based personalization:

```text
MongoDB replica set: required by Task 4 feature transactions
Kafka: required when event ingestion is enabled
Redis: recommended
Task 5 ranking: required for strong cold-start fallback
```

### Step 5: Add Current Task Environment Variables

Update `backend/services/recommendation-service/.env` with section `7. Environment Variables`.

Then load the file:

```bash
set -a
source .env
set +a
```

### Step 6: Run Migrations

Run migrations `001` to `005` on fresh DB.

If migrations `001` to `004` are already applied, run only:

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" \
  --file migrations/005_add_personalized_ranking_indexes.up.js
```

### Step 7: Ensure Feature and Fallback Data Exists

Task 6 needs:

| Data | How To Get It |
|---|---|
| User/guest feature profile | Run Task 3/4 event pipeline or use local seed data. |
| Direct product counters | Run Task 3/4 event pipeline or use local seed data. |
| Eligible product features | Product projection/event pipeline or local seed data. |
| Category/global fallback sets | Run Task 5 ranking or seed one fallback set. |

### Step 8: Start Backend Service

```bash
go run ./cmd/server
```

Expected startup behavior:

- Service loads personalization config.
- If MongoDB is configured, personalized service initializes.
- If Redis is configured, personalized cache is available.
- If Task 5 ranking is enabled, global/category/seller fallback lists stay fresh.

Expected useful log messages during personalized request:

```text
recommendation.personalized_set_generated
recommendation.personalized.profile_read_failed
recommendation.personalized.fallback_read_failed
recommendation.personalized.cache_set_failed
```

Warnings are not always fatal. For example, Redis cache set failure is best-effort after MongoDB save.

### Step 9: Verify Task 6 Functionality

HTTP health:

```bash
curl http://localhost:8088/healthz
curl http://localhost:8088/internal/v1/recommendation/storage/status
```

Metrics:

```bash
curl http://localhost:8088/metrics | grep recommendation_personalized
```

Expected metric names include:

```text
recommendation_personalized_build_total
recommendation_personalized_build_duration_ms
recommendation_personalized_profile_miss_total
recommendation_personalized_candidates_total
recommendation_personalized_result_items_total
recommendation_personalized_fallback_total
recommendation_personalized_backfill_total
recommendation_personalized_cache_hit_total
recommendation_personalized_cache_error_total
recommendation_personalized_empty_result_total
recommendation_personalized_set_upsert_error_total
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
3. Start Kafka only if live events are needed.
4. Load `.env`.
5. Run migration `005` if not already applied.
6. Ensure Task 4 feature data and Task 5 fallback sets exist.
7. Run service.
8. Verify logs, MongoDB generated set, Redis personalized key, and metrics.

Minimal run:

```bash
cd backend/services/recommendation-service
set -a
source .env
set +a
go run ./cmd/server
```

Optional gRPC personalized verification:

```bash
grpcurl -plaintext \
  -import-path ../../../proto \
  -proto ecommerce/recommendation/v1/recommendation.proto \
  -d '{"context":"RECOMMENDATION_CONTEXT_HOME_FEED","type":"RECOMMENDATION_TYPE_PERSONALIZED","user_id":"user_demo","category_id":"cat_demo","limit":5}' \
  localhost:9088 \
  ecommerce.recommendation.v1.RecommendationService/GetRecommendations
```

Anonymous user verification:

```bash
grpcurl -plaintext \
  -import-path ../../../proto \
  -proto ecommerce/recommendation/v1/recommendation.proto \
  -d '{"context":"RECOMMENDATION_CONTEXT_HOME_FEED","type":"RECOMMENDATION_TYPE_PERSONALIZED","anonymous_id":"anon_demo","limit":5}' \
  localhost:9088 \
  ecommerce.recommendation.v1.RecommendationService/GetRecommendations
```

Verify MongoDB personalized set:

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" \
  --eval 'db.getSiblingDB("recommendation_db").recommendation_sets.find({ context_key: "home:user:user_demo", strategy_id: "personalized_v1_behavior" }).pretty()'
```

Verify Redis personalized cache:

```bash
redis-cli TTL 'reco:v1:personalized:user:user_demo:strategy:personalized_v1_behavior'
redis-cli GET 'reco:v1:personalized:user:user_demo:strategy:personalized_v1_behavior'
```

Disable only Task 6 personalization if needed:

```dotenv
RECOMMENDATION_PERSONALIZATION_ENABLED=false
```

Warning: Disabling personalization means explicit personalized requests will not build fresh personalized sets. Depending on serving path and stored data, results may fall back or be empty.

---

## 11. Common Errors & Fixes

Generic setup errors are already documented in previous dependency files. Below are only Task 6-specific issues.

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `RECOMMENDATION_PERSONALIZED_FORMULA_VERSION cannot be empty` | Formula version env set to empty value. | Set `RECOMMENDATION_PERSONALIZED_FORMULA_VERSION=behavior_v1`. | Keep default unless tuning formula. |
| `personalization weights must total 1.00` | Six weight env values do not sum to `1.00`. | Restore defaults or adjust values carefully. | Use a calculator before changing weights. |
| `RECOMMENDATION_PERSONALIZED_FORMULA_VERSION must change when behavior_v1 weights differ` | Weights changed but version stayed default. | Use a new version like `behavior_v2`. | Version every scoring formula change. |
| `RECOMMENDATION_PERSONALIZED_MAX_CANDIDATES must be greater than or equal to RECOMMENDATION_DEFAULT_LIMIT` | Candidate cap lower than response limit. | Increase max candidates or reduce default limit. | Keep `250` candidate default for local/dev. |
| `RECOMMENDATION_PERSONALIZED_PROFILE_MAX_AGE_DAYS must be greater than zero` | Profile freshness value is `0`, negative, or invalid. | Set `30`. | Use day count, not duration string, for this variable. |
| `personalized feature reader is required` during wiring/tests | Personalized service created without Mongo feature repository. | Configure MongoDB or disable personalization for minimal mode. | Keep `RECOMMENDATION_MONGO_URI` loaded. |
| Log says `recommendation.personalization.not_configured` | MongoDB is not configured, so personalized service cannot initialize. | Set valid `RECOMMENDATION_MONGO_URI` and restart. | Verify `/internal/v1/recommendation/storage/status`. |
| Personalized request returns fallback only | Profile missing, weak, stale, or no eligible candidates. | Create feature profile/counters or run Task 3/4 pipeline. | Monitor `recommendation_personalized_profile_miss_total`. |
| Personalized request returns empty | No personalized candidates and no non-expired Task 5 fallback set. | Run Task 5 ranking or seed `home:global` fallback set. | Keep fallback job enabled and product features healthy. |
| Product exists but not selected | Product fails eligibility: wrong `status`, `stock_status`, or quality flags. | Set `status="active"`, `stock_status="in_stock"`, `quality_flags.is_recommendable=true`, and not deleted. | Keep Product Service projection aligned. |
| Redis personalized key not found | Redis not configured, request not run, cache expired, or checking key without strategy suffix. | Check Mongo set first, then use strategy-suffixed key from section `6`. | Use `redis-cli --scan --pattern 'reco:v1:personalized:*'`. |
| Migration `005` index missing | Migration not run or wrong database selected. | Export `RECOMMENDATION_MONGO_DATABASE=recommendation_db` and rerun migration `005`. | Verify indexes after fresh DB setup. |
| Explicit personalized request on non-home context fails | Current personalized builder supports `home_feed` context only. | Use `RECOMMENDATION_CONTEXT_HOME_FEED` for Task 6 verification. | Let non-home contexts use their proper recommendation type. |
| Profile exists but considered stale | `last_event_at` older than profile max age. | Update profile via event pipeline or increase max age carefully. | Keep Task 3/4 event flow running. |
| Cache write warning after result generated | Redis down or wrong `RECOMMENDATION_REDIS_ADDR`. | Start Redis or clear Redis env for Mongo-only verification. | Treat Redis as recommended but not mandatory. |

---

## 12. Security & Best Practices

Task-specific best practices:

- Do not store email, phone, address, payment data, or raw PII in personalization metadata.
- Keep `profile_key` as an internal identifier. Avoid exposing it in public API responses or broad logs.
- Keep personalized TTL short. Default `5m` is good because user behavior changes quickly.
- Keep guest profile retention aligned with privacy policy.
- Keep Task 5 fallback enabled so cold-start users do not get blank recommendations.
- Version formula changes with `RECOMMENDATION_PERSONALIZED_FORMULA_VERSION`.
- Do not change personalized weights casually in production. Small changes can impact ranking quality.
- Keep all six personalized weights non-negative and totaling `1.00`.
- Keep MongoDB, Redis, and Kafka private. Do not expose them publicly.
- Set `RECOMMENDATION_STORAGE_FAIL_FAST=true` in production if personalized ranking must not run degraded.
- Prefer migration-driven indexes in production. `RECOMMENDATION_MONGO_ENSURE_INDEXES=true` is convenient locally but can slow large production startup.
- Run `go test ./...` after config or dependency changes.
- Track profile miss, fallback, empty result, and cache error metrics.

---

## 13. Missing or Misconfigured Things

Audit notes for `TASK_FILE_NAME` setup:

| Item | Status | Recommendation |
|---|---|---|
| `.env` auto-load | Not present | Source `.env` manually or use Docker/process manager env injection. |
| Dedicated manual personalized rebuild endpoint | Not present | Current path builds personalized result through serving/use case. Add protected admin tooling later if ops needs it. |
| Live feature data | External dependency | Task 6 depends on Task 3/4 pipeline or local seed data. |
| Cold-start fallback data | External dependency | Task 6 depends on Task 5 generated sets for useful fallback. |
| Product feature schema alignment | Important | Ensure `stock_status`, `quality_flags`, `brand_id`, `price.bucket`, and counters exist. |
| Migration `005` coverage | Adds brand candidate index only | Category/seller candidate performance relies on existing Task 4/5 indexes. |
| Service Dockerfile / root compose | Not detected in current repo scan | Use previous dependency-only Docker setup for local infra; add service Dockerfile before production deployment. |
| User deletion operational path | Not shown as a dedicated tool | Production should delete user profile, user-product counters, generated sets, and Redis personalized keys on account deletion. |
| Raw user ID logging policy | Needs team policy | Prefer `identity_type` and aggregate counts in logs. |
| Production secret handling | Environment-specific | Use secret manager or deployment platform secrets, not committed `.env`. |

---

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `1. Project Tech Stack Analysis` | Base Go, MongoDB, Redis, Kafka, Prometheus, HTTP/gRPC, and Docker explanations already documented. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `2. Go Dependency System` | Go modules, `go.mod`, `go.sum`, `go mod download`, and common Go errors are unchanged. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `3. Databases` -> `MongoDB` | MongoDB install, Docker run, port, connection string, and credential placement are reused. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `4. Environment Variables` | General `.env` setup and process loading behavior is reused. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `6. Ports and Networking` | HTTP, gRPC, MongoDB, Redis, and Kafka ports are unchanged. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `7. Docker and DevOps Setup` | No new Docker container is needed by Task 6. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | `recommendation_db` and `recommendation_sets` setup is reused. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `6. Redis / Queue / External Services` | Redis setup, key prefix, TTL, and verification basics are reused. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `7. Environment Variables` | MongoDB, Redis, and cache TTL variables are reused. |
| `TaskImplementation/${SERVICE_NAME}/task3_Dependency.md` | `6. Redis / Queue / External Services` | Kafka setup is reused only for live event-to-feature updates. |
| `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` | `5. Database Setup` | Feature-store collections and replica set requirement are reused. |
| `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` | `7. Environment Variables` | Feature builder variables are reused for live profiles/counters. |
| `TaskImplementation/${SERVICE_NAME}/task5_Dependency.md` | `5. Database Setup` | `product_features` eligibility fields and fallback `recommendation_sets` are reused. |
| `TaskImplementation/${SERVICE_NAME}/task5_Dependency.md` | `6. Redis / Queue / External Services` | Trending/category fallback cache setup is reused. |
| `TaskImplementation/${SERVICE_NAME}/task5_Dependency.md` | `11. Common Errors & Fixes` | Product eligibility and fallback-generation troubleshooting are reused. |

---

## 15. Final Checklist

- [ ] Previous dependency documentation checked before setup.
- [ ] No duplicate base installation instructions added.
- [ ] Go dependencies downloaded with `go mod download`.
- [ ] MongoDB running and reachable through `RECOMMENDATION_MONGO_URI`.
- [ ] Migration `005_add_personalized_ranking_indexes.up.js` applied.
- [ ] `idx_product_features_personalized_brand` index verified.
- [ ] Task 4 feature profiles and product counters exist.
- [ ] Eligible `product_features` exist with `stock_status="in_stock"`.
- [ ] Task 5 global/category fallback sets exist and are not expired.
- [ ] Redis running if personalized cache verification is required.
- [ ] Task 6 personalization env variables added or verified.
- [ ] Personalized weights total exactly `1.00`.
- [ ] `.env` sourced before `go run ./cmd/server`.
- [ ] Backend service started.
- [ ] Personalized request verified through gRPC or service use case.
- [ ] MongoDB personalized `recommendation_sets` verified.
- [ ] Redis personalized cache key and TTL verified if Redis is configured.
- [ ] Metrics checked with `curl http://localhost:8088/metrics | grep recommendation_personalized`.
- [ ] Logs checked for profile miss, fallback, cache, or storage warnings.
- [ ] No real credentials committed.
- [ ] Original implementation file at `INPUT_FILE_PATH` not modified.
