# Project Dependency & Setup Guide

## Variables Used

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Recommendation Service` |
| `TASK_FILE_NAME` | `task4.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task4_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |
| Backend code path | `backend/services/recommendation-service` |

Beginner note: Ye file sirf dependency, setup, environment, database, Docker, DevOps, and troubleshooting guide hai. Business logic ya original implementation file modify nahi kiya gaya.

---

## 1. Project Overview

`TASK_FILE_NAME` ka focus hai: `SERVICE_NAME` ke raw `user_interactions` events se derived feature-store collections banana.

Simple Hinglish:

- Task 3 me events MongoDB `user_interactions` collection me store hote hain.
- Task 4 un events ko compact feature documents me convert karta hai.
- Ye features later ranking/personalization tasks ke input banenge.
- Current backend me feature builder already wired hai: event ingestion ke baad feature collections update hote hain, reconciliation background job missed events repair karta hai, and window rebuild job 24h/7d/30d counters refresh karta hai.

Read these previous dependency files first:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`1. Project Tech Stack Analysis`
`2. Go Dependency System`
`3. Databases`
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
`11. Common Errors & Fixes`
```

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task3_Dependency.md`

Sections:
`5. Database Setup`
`6. Redis / Queue / External Services`
`7. Environment Variables`
`8. Docker Setup`
`9. Local Development Setup`
`11. Common Errors & Fixes`
```

`TASK_FILE_NAME` new setup scope:

- MongoDB migration: `backend/services/recommendation-service/migrations/003_create_feature_store_schema.up.js`
- New MongoDB collections:
  - `user_feature_profiles`
  - `user_product_counters`
  - `product_cooccurrence_features`
  - `feature_job_runs`
  - `feature_processed_events`
- Updated `product_features` indexes for feature-store reads.
- New feature-builder environment variables.
- MongoDB replica set requirement for feature writes, because feature updates use MongoDB transactions.

---

## 2. Tech Stack

Most base technologies are already explained in previous dependency files.

| Technology | Required? | New or Reused | Beginner Explanation | Why Current Task Uses It |
|---|---:|---|---|---|
| Go `1.26.3` | Yes | Reused | Go backend language hai jo compiled service banata hai. | Feature builder, repositories, Kafka consumer, HTTP/gRPC servers run karne ke liye. |
| Go modules | Yes | Reused | `go.mod`/`go.sum` dependency system hai. | Existing dependencies version lock karne ke liye. |
| MongoDB | Yes for full Task 4 | Reused service, new schema | MongoDB document database hai. Simple English: JSON-like documents collections me store hote hain. | Derived recommendation features and job metadata store karne ke liye. |
| MongoDB transactions | Yes for feature writes | New operational requirement | Transaction ka matlab multiple collection updates ek atomic unit me karna. | `product_features`, `user_feature_profiles`, `user_product_counters`, processed-event marker, and job-run write ko safe rakhne ke liye. |
| `mongosh` | Yes for manual migration | Reused tool, new migration | MongoDB shell CLI. | `003_create_feature_store_schema.up.js` run/verify karne ke liye. |
| Kafka | Required when event ingestion is enabled | Reused from Task 3 | Kafka event streaming platform hai. | New incoming events feature builder tak pahunchte hain through Task 3 ingestion flow. |
| Redis | Optional but recommended | Reused | Redis fast in-memory cache hai. | Interaction ke baad cache dirty markers/invalidation ke liye; Task 4 itself new Redis service add nahi karta. |
| Prometheus client | Recommended | Reused, new metrics categories | Metrics expose karne ki Go library. | Feature jobs processed events, update errors, duration, and lag observe karne ke liye. |

No new MySQL, PostgreSQL, RabbitMQ, NATS, MinIO, Elasticsearch, AWS S3, SMTP, Stripe, Twilio, OAuth, Kubernetes, or Nginx dependency is introduced by `TASK_FILE_NAME`.

---

## 3. Required Software

Base installation is already documented in:

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
| Go `1.26.3` or compatible newer | Build/test/run service | Reused |
| MongoDB `7.x` | Feature-store collections | Reused service, new schema |
| MongoDB replica set mode | Feature update transactions | New Task 4 requirement |
| `mongosh` | Run migration `003` and verify indexes | Reused tool |
| Kafka `3.x` | Full event-to-feature verification | Reused from Task 3 |
| Redis `7.x` | Cache invalidation side effects | Reused, optional for Task 4 |
| `curl` | HTTP health/metrics checks | Reused |

Important: Sirf migration `003` run karne ke liye replica set mandatory nahi hai. Lekin feature builder actual events process karega to MongoDB transactions use honge, isliye full local verification ke liye replica set chahiye.

---

## 4. Dependency Management

This is still the same Go module project. Go dependency setup is already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `2. Go Dependency System`
```

`TASK_FILE_NAME` dependency status:

| Dependency | Version In `go.mod` | Purpose | New Install Needed? |
|---|---:|---|---:|
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | MongoDB CRUD, sessions, transactions, indexes | No |
| `github.com/segmentio/kafka-go` | `v0.4.51` | Event consumption that triggers feature updates | No |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Cache dirty markers/invalidation | No |
| `github.com/prometheus/client_golang` | `v1.23.2` | Metrics for consumer and feature jobs | No |

No new `go get` command is required for `TASK_FILE_NAME`.

Recommended verification:

```bash
cd backend/services/recommendation-service
go mod download
go test ./...
```

---

## 5. Database Setup

### MongoDB Role

MongoDB is mandatory for full feature-store behavior.

Simple Hinglish:

MongoDB me raw events already `user_interactions` me aate hain. Task 4 un raw events ka compressed version alag collections me store karta hai, taaki recommendation ranking fast ho aur har request par raw history scan na karni pade.

### Reused MongoDB Setup

MongoDB installation, normal Docker setup, connection string format, and base credentials placement are already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `3. Databases` -> `MongoDB`
```

```md
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
Section: `5. Database Setup`
```

Do not duplicate or recreate MongoDB if it is already running.

### Replica Set Requirement

Task 4 feature writes use MongoDB transactions through the Go driver. MongoDB transactions need a replica set or sharded cluster.

For local development, use the single-node replica set approach already introduced in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `3. Databases` -> `MongoDB` -> `E. Docker setup`
```

Expected connection style:

```env
RECOMMENDATION_MONGO_URI=mongodb://localhost:27017/?replicaSet=rs0
RECOMMENDATION_MONGO_DATABASE=recommendation_db
```

If you keep using a plain standalone MongoDB container, schema migration can succeed but feature application may fail during runtime transaction execution.

### Current Task Migration

Prerequisite migrations:

1. `001_create_recommendation_storage.up.js`
2. `002_add_event_ingestion_indexes.up.js`
3. `003_create_feature_store_schema.up.js`

Run from service folder:

```bash
cd backend/services/recommendation-service
export RECOMMENDATION_MONGO_DATABASE=recommendation_db

mongosh "mongodb://localhost:27017/?replicaSet=rs0" \
  --file migrations/001_create_recommendation_storage.up.js

mongosh "mongodb://localhost:27017/?replicaSet=rs0" \
  --file migrations/002_add_event_ingestion_indexes.up.js

mongosh "mongodb://localhost:27017/?replicaSet=rs0" \
  --file migrations/003_create_feature_store_schema.up.js
```

If your local MongoDB uses auth, use the URI from your `.env`, for example:

```bash
mongosh "mongodb://root:secret@localhost:27017/admin?replicaSet=rs0" \
  --file migrations/003_create_feature_store_schema.up.js
```

### Collections Introduced Or Updated By Current Task

| Collection | Status | Purpose |
|---|---|---|
| `product_features` | Updated | Product-level counters, metadata snapshot, quality flags, embedding refs |
| `user_feature_profiles` | New | User/guest category, seller, brand, price, and recent product preference summary |
| `user_product_counters` | New | Per profile/product interaction counters and weighted affinity score |
| `product_cooccurrence_features` | New | Bought-together / related product pair signals |
| `feature_job_runs` | New | Incremental and window rebuild job status, stats, watermark, errors |
| `feature_processed_events` | New | Idempotency marker so same event features are not applied twice |

### Current Task Indexes

| Collection | Index | Why Needed |
|---|---|---|
| `product_features` | `ux_product_features_product_id` | One feature document per product |
| `product_features` | `idx_product_features_category_recommendable_purchases` | Category-level top products fast read |
| `product_features` | `idx_product_features_seller_recommendable_purchases` | Seller storefront popular products fast read |
| `product_features` | `idx_product_features_brand_recommendable` | Brand/similar candidate filtering |
| `product_features` | `idx_product_features_embedding_vector_id` | Future embedding debug/lookup |
| `user_feature_profiles` | `ux_user_feature_profiles_profile_key` | One profile per `user:*` or `anon:*` key |
| `user_feature_profiles` | `idx_user_feature_profiles_user_id` | Logged-in user profile lookup |
| `user_feature_profiles` | `idx_user_feature_profiles_anonymous_id` | Guest profile lookup |
| `user_feature_profiles` | `idx_user_feature_profiles_expires_at_ttl` | Guest profiles auto-cleanup |
| `user_product_counters` | `ux_user_product_counters_profile_product` | Duplicate user-product rows avoid |
| `user_product_counters` | `idx_user_product_counters_profile_recent` | Recent affinities for personalization |
| `user_product_counters` | `idx_user_product_counters_product_score` | Product affinity/debug reads |
| `user_product_counters` | `idx_user_product_counters_category_recent` | Category recent behavior reads |
| `user_product_counters` | `idx_user_product_counters_expires_at_ttl` | Old pair counters auto-cleanup |
| `product_cooccurrence_features` | `idx_product_cooccurrence_source_type_score` | Related product candidates by source |
| `product_cooccurrence_features` | `idx_product_cooccurrence_related` | Reverse lookup/debug |
| `product_cooccurrence_features` | `idx_product_cooccurrence_category_score` | Category-level co-occurrence reads |
| `product_cooccurrence_features` | `idx_product_cooccurrence_last_seen` | Recently updated pairs |
| `feature_job_runs` | `idx_feature_job_runs_type_started` | Job history by job type |
| `feature_job_runs` | `idx_feature_job_runs_status_started` | Failed/running job debug |
| `feature_processed_events` | `ux_feature_processed_events_event_id` | Idempotency for feature processing |
| `feature_processed_events` | `idx_feature_processed_events_expires_at_ttl` | Old processed markers auto-cleanup |

### Verify MongoDB Collections And Indexes

```bash
mongosh "mongodb://localhost:27017/?replicaSet=rs0" \
  --eval 'db.getSiblingDB("recommendation_db").getCollectionNames().sort()'
```

Expected Task 4 collections:

```text
feature_job_runs
feature_processed_events
product_cooccurrence_features
product_features
user_feature_profiles
user_product_counters
```

Verify indexes:

```bash
mongosh "mongodb://localhost:27017/?replicaSet=rs0" \
  --eval 'db.getSiblingDB("recommendation_db").user_product_counters.getIndexes().map(i => i.name)'
```

Expected Task 4 index example:

```text
ux_user_product_counters_profile_product
idx_user_product_counters_profile_recent
idx_user_product_counters_product_score
idx_user_product_counters_category_recent
idx_user_product_counters_expires_at_ttl
```

---

## 6. Redis / Queue / External Services

### Redis

Redis setup is already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `5. External Services` -> `Redis`
```

```md
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
Section: `6. Redis / Queue / External Services`
```

Task 4 does not add a new Redis container or port. Redis remains optional for feature writing, but recommended because event ingestion uses best-effort cache invalidation/dirty markers.

### Kafka

Kafka setup and topics are already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task3_Dependency.md`
Section: `6. Redis / Queue / External Services` -> `Kafka Role`
```

Task 4 uses the same `recommendation.events` flow. No new Kafka topic is required.

### Prometheus Metrics

Prometheus setup is reused from:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `5. External Services` -> `Prometheus`
```

Current feature builder adds useful runtime signals through existing metrics plumbing:

| Signal | Why Useful |
|---|---|
| Feature events processed | Check events are becoming features |
| Feature documents updated | See product/profile/counter/co-occurrence writes |
| Feature update errors | Detect transaction/index/schema issues |
| Feature job duration | Catch slow rebuild jobs |
| Feature job lag | Understand delay between event time and feature update |

---

## 7. Environment Variables

Base `.env` setup is already documented in previous dependency files. Only Task 4 feature variables are listed here.

Important loading behavior:

- The Go code reads environment variables through `os.Getenv`.
- There is no automatic `.env` loader in the service code.
- For local development, create `backend/services/recommendation-service/.env` for convenience, then export/source it before `go run`.
- Existing local `.env` already contains the Task 4 variables below.

### Current Task `.env` Delta

```env
# Derived feature store builder
RECOMMENDATION_FEATURES_ENABLED=true
RECOMMENDATION_GUEST_PROFILE_RETENTION_SECONDS=2592000
RECOMMENDATION_USER_PRODUCT_RETENTION_SECONDS=15552000
RECOMMENDATION_FEATURE_EVENT_RETENTION_SECONDS=15552000
RECOMMENDATION_FEATURE_RECENT_PRODUCTS_LIMIT=20
RECOMMENDATION_FEATURE_RECONCILE_BATCH_SIZE=200
RECOMMENDATION_FEATURE_RECONCILE_INTERVAL=1m
RECOMMENDATION_FEATURE_WINDOW_REBUILD_INTERVAL=5m

# MongoDB must point to a replica set for feature transactions
RECOMMENDATION_MONGO_URI=mongodb://localhost:27017/?replicaSet=rs0
RECOMMENDATION_MONGO_DATABASE=recommendation_db
```

### Variable Reference

| Variable | Required? | Default | Purpose | Security / Ops Note |
|---|---:|---|---|---|
| `RECOMMENDATION_FEATURES_ENABLED` | Optional | `true` | Feature builder on/off switch | Set `false` only when you want raw event ingestion without derived feature writes. |
| `RECOMMENDATION_GUEST_PROFILE_RETENTION_SECONDS` | Optional | `2592000` | Guest profile TTL, 30 days | Avoid keeping anonymous behavior forever. |
| `RECOMMENDATION_USER_PRODUCT_RETENTION_SECONDS` | Optional | `15552000` | User-product counter TTL, 180 days | Tune for storage size and personalization freshness. |
| `RECOMMENDATION_FEATURE_EVENT_RETENTION_SECONDS` | Optional | `15552000` | Processed event marker TTL, 180 days | Must be greater than or equal to `RECOMMENDATION_INTERACTION_RETENTION_SECONDS`. |
| `RECOMMENDATION_FEATURE_RECENT_PRODUCTS_LIMIT` | Optional | `20` | Max recent products stored in profile | Keep bounded to avoid huge profile docs. |
| `RECOMMENDATION_FEATURE_RECONCILE_BATCH_SIZE` | Optional | `200` | Pending raw event batches processed per loop | Increase carefully; large batches can stress Mongo. |
| `RECOMMENDATION_FEATURE_RECONCILE_INTERVAL` | Optional | `1m` | Background reconciliation frequency | Set `0` to stop periodic reconciliation if needed. |
| `RECOMMENDATION_FEATURE_WINDOW_REBUILD_INTERVAL` | Optional | `5m` | 24h/7d/30d product window rebuild frequency | Set `0` to disable scheduled window rebuilds. |

Common mistakes:

| Mistake | Result | Fix |
|---|---|---|
| `.env` file exists but variables are not exported | Service uses defaults or empty Mongo URI | Run `set -a; source .env; set +a` before `go run`. |
| `RECOMMENDATION_FEATURE_EVENT_RETENTION_SECONDS` less than interaction retention | Config validation fails | Keep both at `15552000` for local default. |
| Mongo URI points to standalone MongoDB | Feature transactions fail at runtime | Use `?replicaSet=rs0` and initialize single-node replica set. |
| `RECOMMENDATION_FEATURE_RECENT_PRODUCTS_LIMIT=0` | Config validation fails | Use positive value like `20`. |

---

## 8. Docker Setup

Docker installation and base MongoDB/Redis/Kafka setup are already documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `7. Docker and DevOps Setup`
```

```md
`TaskImplementation/${SERVICE_NAME}/task3_Dependency.md`
Section: `8. Docker Setup`
```

Task 4 does not introduce a new container, Docker network, Dockerfile, restart policy, or exposed port.

Task 4 operational change:

| Component | Change |
|---|---|
| MongoDB | For full feature processing, local MongoDB should run as a replica set. |
| Volumes | Reuse existing Mongo volume; no new volume required. |
| Kafka | Reuse Task 3 topics; no new topic required. |
| Redis | Reuse existing cache service; no new key prefix required. |

If your local Compose file still starts MongoDB as standalone, update only the Mongo command/URI to replica-set mode. Keep credentials in `.env` or local Compose override, not in committed production config.

---

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Follow shared setup from:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
`TaskImplementation/${SERVICE_NAME}/task3_Dependency.md`
```

### Step 2: Go To Project Directory

```bash
cd backend/services/recommendation-service
```

### Step 3: Install Dependencies

No new dependency install is needed for `TASK_FILE_NAME`.

```bash
go mod download
go test ./...
```

### Step 4: Start Required Services

Use the same services from earlier tasks:

| Service | Needed For Task 4? | Note |
|---|---:|---|
| MongoDB replica set | Yes | Required for transaction-based feature writes |
| Kafka | Yes for event-driven verification | Reused from Task 3 |
| Redis | Optional | Recommended for cache invalidation side effects |

### Step 5: Add Current Task Environment Variables

Add only Task 4 variables from Section `7. Environment Variables` if they are missing from your local `.env`.

Then export them:

```bash
set -a
source .env
set +a
```

### Step 6: Run Migrations

Run `001`, `002`, then current `003` if this is a fresh database. If `001` and `002` already ran, only run `003`.

```bash
mongosh "$RECOMMENDATION_MONGO_URI" \
  --file migrations/003_create_feature_store_schema.up.js
```

### Step 7: Start Backend Service

```bash
go run ./cmd/server
```

Expected startup behavior:

- HTTP starts on `:8088` unless `RECOMMENDATION_HTTP_ADDR` changed.
- gRPC starts on `:9088` unless `RECOMMENDATION_GRPC_ADDR` changed.
- Kafka consumer starts only when `RECOMMENDATION_EVENTS_ENABLED=true` and `QUEUE_PROVIDER=kafka`.
- Feature builder starts when `RECOMMENDATION_FEATURES_ENABLED=true` and MongoDB is configured.
- Reconciliation and window rebuild jobs run in background.

### Step 8: Verify Task 4 Functionality

Check service health and metrics:

```bash
curl http://localhost:8088/health
curl http://localhost:8088/metrics
```

Verify feature collections exist:

```bash
mongosh "$RECOMMENDATION_MONGO_URI" \
  --eval 'db.getSiblingDB("recommendation_db").feature_job_runs.find().sort({started_at:-1}).limit(5).toArray()'
```

After publishing a valid Task 3 event, check derived features:

```bash
mongosh "$RECOMMENDATION_MONGO_URI" \
  --eval 'db.getSiblingDB("recommendation_db").product_features.find().limit(5).toArray()'

mongosh "$RECOMMENDATION_MONGO_URI" \
  --eval 'db.getSiblingDB("recommendation_db").user_feature_profiles.find().limit(5).toArray()'

mongosh "$RECOMMENDATION_MONGO_URI" \
  --eval 'db.getSiblingDB("recommendation_db").user_product_counters.find().limit(5).toArray()'
```

---

## 10. Running the Project

Minimal schema verification:

```bash
cd backend/services/recommendation-service
set -a
source .env
set +a
mongosh "$RECOMMENDATION_MONGO_URI" --file migrations/003_create_feature_store_schema.up.js
```

Full feature-processing run:

```bash
cd backend/services/recommendation-service
set -a
source .env
set +a
go run ./cmd/server
```

Full run prerequisites:

| Dependency | Required Value |
|---|---|
| MongoDB | Running and reachable through `RECOMMENDATION_MONGO_URI` |
| MongoDB mode | Replica set for transactions |
| Kafka | Running if `RECOMMENDATION_EVENTS_ENABLED=true` |
| Kafka topics | Same topics from Task 3 |
| Redis | Running if `RECOMMENDATION_REDIS_ADDR` is set |
| Migrations | `001`, `002`, and `003` applied |

---

## 11. Common Errors & Fixes

Generic Go/Mongo/Redis/Kafka issues are already documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `9. Common Errors and Fixes`
```

```md
`TaskImplementation/${SERVICE_NAME}/task3_Dependency.md`
Section: `11. Common Errors & Fixes`
```

Task 4-specific errors:

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `Transaction numbers are only allowed on a replica set member or mongos` | MongoDB is running standalone | Run MongoDB as single-node replica set and use URI with `?replicaSet=rs0` | Use replica-set Docker setup for full local verification |
| `feature building requires configured MongoDB storage` | `RECOMMENDATION_FEATURES_ENABLED=true` and events active, but Mongo URI empty/unreachable | Set `RECOMMENDATION_MONGO_URI` correctly or disable features/events for minimal run | Keep `.env` sourced before starting service |
| `RECOMMENDATION_FEATURE_EVENT_RETENTION_SECONDS must be greater than or equal to RECOMMENDATION_INTERACTION_RETENTION_SECONDS` | Processed-event marker expires before raw interaction TTL | Set both to `15552000` locally, or make feature event retention larger | Tune retention values together |
| `E11000 duplicate key error` on feature processed events | Same event already processed | Usually safe; service treats processed event duplicates as idempotent | Keep event IDs stable and unique |
| Feature collections stay empty | No valid events arrived, features disabled, or migration not run | Check `RECOMMENDATION_FEATURES_ENABLED`, Kafka consumer logs, DLQ, and Mongo collections | Publish one valid Task 3 event and verify `user_interactions` first |
| Window counters stay `0` | Scheduled window rebuild not running or no recent raw events | Check `RECOMMENDATION_FEATURE_WINDOW_REBUILD_INTERVAL` and `feature_job_runs` | Keep interval positive for local verification |
| `cannot be used as a feature map key` | Category/seller/brand ID contains `.` or starts with `$` | Sanitize producer IDs before publishing events | Do not use Mongo operator-like strings as dynamic map keys |

---

## 12. Security & Best Practices

Task 4-specific best practices:

| Area | Recommendation |
|---|---|
| PII | Do not store email, phone, address, or payment data in feature collections. Use only IDs and behavior counters. |
| Guest data | Keep `RECOMMENDATION_GUEST_PROFILE_RETENTION_SECONDS` bounded. Default 30 days is beginner-friendly. |
| User-product counters | Keep TTL enabled so high-cardinality pair data does not grow forever. |
| Processed-event markers | Keep retention at least as long as raw interaction retention to avoid duplicate feature replay. |
| MongoDB transactions | Use replica set in every environment where feature builder is enabled. |
| Dynamic map keys | Validate category/seller/brand IDs. MongoDB field keys must not contain unsafe structures like `$` prefixes. |
| Embedding refs | Store only vector IDs/metadata here. Do not store large raw vectors in MongoDB until vector storage is intentionally designed. |
| Production indexes | Prefer migration-driven index creation in production. `RECOMMENDATION_MONGO_ENSURE_INDEXES=true` is convenient locally but can slow startup with large collections. |
| Logs | Logs should include event/job IDs, not raw sensitive payloads. |

Security audit notes:

| Finding | Status | Suggested Fix |
|---|---|---|
| Local `.env` uses sample `root:secret` Mongo credentials | Local-only risk | Use secret manager or environment-specific secret injection outside local dev. |
| Feature builder silently depends on replica set for transactions | New operational requirement | Documented here; keep local/prod Mongo deployment in replica-set/cluster mode. |
| `.env` is not loaded automatically | Setup footgun | Source env before run or add a documented dev wrapper script later. |
| Product metadata completeness | Partial | Feature builder can derive counters from events, but Product Service should eventually backfill `status`, `stock_status`, `price`, and attributes. |

---

## 13. Missing or Misconfigured Things

| Item | Current Status | Impact | Recommendation |
|---|---|---|---|
| Dedicated migration runner | Not present | Developers run `mongosh --file` manually | Add a small migration command/tool later if migrations grow. |
| Automatic `.env` loading | Not present | Beginners may think `.env` is enough | Use `set -a; source .env; set +a` before commands. |
| Product metadata backfill job | Not part of Task 4 | `product_features` may have counters but missing status/stock/price until source sync exists | Add Product Service projection/backfill in a future task. |
| Vector store | Future-ready only | `embedding_refs` stores references, not actual vectors | Add vector DB/provider only when ML similarity is implemented. |
| Manual feature rebuild endpoint/CLI | Not present | Rebuild currently runs scheduled in service process | Add admin command later for controlled rebuilds. |
| MongoDB standalone local setup | Common misconfig | Transactions fail | Use replica-set mode for Task 4 full verification. |

---

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `1. Project Tech Stack Analysis` | Go, HTTP, gRPC, MongoDB, Redis, Kafka, Prometheus basics already explained |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `2. Go Dependency System` | Same Go module and dependency commands |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `3. Databases` -> `MongoDB` | MongoDB installation, default port, connection string, and basic Docker setup already documented |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `5. External Services` | Redis, Kafka, Prometheus, Docker basics already documented |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `6. Ports and Networking` | HTTP/gRPC/Mongo/Redis/Kafka ports are unchanged |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | Base `recommendation_db`, storage migration, and Mongo credentials placement reused |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `6. Redis / Queue / External Services` | Redis cache key prefix and TTL setup reused |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `7. Environment Variables` | Base storage/cache env variables reused |
| `TaskImplementation/${SERVICE_NAME}/task3_Dependency.md` | `6. Redis / Queue / External Services` -> `Kafka Role` | Same event topic, DLQ topic, and consumer group reused |
| `TaskImplementation/${SERVICE_NAME}/task3_Dependency.md` | `9. Local Development Setup` | Event publishing and raw `user_interactions` verification reused before checking features |
| `TaskImplementation/${SERVICE_NAME}/task3_Dependency.md` | `11. Common Errors & Fixes` | Generic Kafka/event ingestion troubleshooting reused |

---

## 15. Final Checklist

- [ ] Previous dependency documentation checked
- [ ] No duplicate MongoDB/Redis/Kafka installation steps added
- [ ] MongoDB is running as a replica set for full feature processing
- [ ] `001_create_recommendation_storage.up.js` applied if database is fresh
- [ ] `002_add_event_ingestion_indexes.up.js` applied if event ingestion is used
- [ ] `003_create_feature_store_schema.up.js` applied
- [ ] Task 4 feature environment variables added/exported
- [ ] `RECOMMENDATION_FEATURE_EVENT_RETENTION_SECONDS` is greater than or equal to raw interaction retention
- [ ] Backend service starts on HTTP `:8088` and gRPC `:9088`
- [ ] Kafka event ingestion is enabled only when Kafka is running
- [ ] A valid event appears first in `user_interactions`
- [ ] Derived documents appear in `product_features`
- [ ] User/guest documents appear in `user_feature_profiles`
- [ ] Pair counters appear in `user_product_counters`
- [ ] Job history appears in `feature_job_runs`
- [ ] Metrics/logs checked for feature update errors
- [ ] Secrets are not committed with real production values
