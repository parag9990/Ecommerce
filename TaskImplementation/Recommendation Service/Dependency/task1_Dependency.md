# Dependency and Setup Guide for `SERVICE_NAME`

## Variables Used

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Recommendation Service` |
| `TASK_FILE_NAME` | `task1.md` |
| `OUTPUT_FILE_NAME` | `task1_Dependency.md` |
| Input path | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| Output path | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |
| Backend code path | `backend/services/recommendation-service` |

This document explains dependencies, setup, environment variables, databases, external services, Docker, migrations, run commands, debugging, and security notes for the task implementation and the current backend service code.

Important beginner note: Task 1 is mainly recommendation type-definition work. The repo also contains later backend implementation for storage, ranking, event ingestion, personalization, and A/B testing. So this guide separates:

- Minimal Task 1 run: type definitions, HTTP endpoints, and gRPC server can start without MongoDB, Redis, or Kafka if event ingestion is disabled.
- Full service run: MongoDB, Redis, and Kafka are needed when storage, ranking, personalization, and event ingestion are enabled.

---

## 1. Project Tech Stack Analysis

| Technology | Required? | What it is | Why this project uses it |
|---|---:|---|---|
| Go `1.26.3` | Yes | Go ek compiled backend language hai. Simple English: it builds fast APIs and services. | Service code, domain logic, HTTP server, gRPC server, Kafka consumer, Mongo/Redis repositories. |
| Go modules | Yes | `go.mod`/`go.sum` dependency system hai. | External libraries ko version ke saath manage karta hai. |
| HTTP `net/http` | Yes | Go standard HTTP server. | Internal endpoints expose karta hai: health, metrics, type definitions, storage status. |
| gRPC | Yes | High-performance service-to-service RPC framework. | `GetRecommendations` endpoint future/frontend/backend services ke liye. |
| Protocol Buffers | Yes for gRPC contract | API contract define karne ka strongly typed format. | `proto/ecommerce/recommendation/v1/recommendation.proto` se Go stubs generate hote hain. |
| MongoDB | Conditional | Document database hai. Data JSON-like documents me store hota hai. | Interactions, product features, recommendation sets, feature profiles, A/B assignments. Full service ke liye required. |
| Redis | Optional but recommended | In-memory cache hai. Fast reads ke liye use hota hai. | Recommendation result cache, dirty keys, rebuild locks. |
| Kafka | Optional, required if events enabled | Distributed message queue/streaming platform. | Session/order/wishlist events consume karke recommendation features build karta hai. |
| Prometheus client | Recommended | Metrics expose karne ki library. | `/metrics` endpoint se request, consumer, ranking, and gRPC metrics scrape ho sakte hain. |
| Docker | Optional but recommended | Containers run karne ka tool. | Local MongoDB, Redis, Kafka setup beginner ke liye easy ho jata hai. |
| `mongosh` | Required for manual migrations | MongoDB shell. | `migrations/*.up.js` files run karne ke liye. |
| Buf/protoc plugins | Optional | Proto code generation tools. | Sirf tab required jab `.proto` files change karke Go code regenerate karna ho. |

Services/data producers:

| External/internal producer | Required? | Why |
|---|---:|---|
| Product Service | Conceptually required | Product attributes, category, brand, seller, price, stock status recommendation signals dete hain. |
| Session Management Service | Required for behavior events | Product view, clicks, anonymous session behavior, user journeys Kafka me publish kar sakte hain. |
| Order Service | Required for purchase signals | Purchases aur frequently bought together signals ke liye. |
| Wishlist Service | Required for wishlist signals | Wishlist add/remove personalization strength ke liye. |

---

## 2. Go Dependency System

This is a Go project.

### Important files

| File | Meaning |
|---|---|
| `backend/services/recommendation-service/go.mod` | Service module name, Go version, and direct dependencies. |
| `backend/services/recommendation-service/go.sum` | Dependency checksum lock file. Isse dependency tampering detect hoti hai. |
| `backend/go.work` | Workspace file. Multiple local Go modules ko ek saath link karta hai. |
| `backend/proto-gen/go/go.mod` | Generated protobuf Go module. Recommendation service uses it through a local `replace`. |

### Main Go dependencies

| Dependency | What it does |
|---|---|
| `google.golang.org/grpc` | gRPC server and client support. |
| `google.golang.org/protobuf` | Protocol Buffer messages. |
| `github.com/prometheus/client_golang` | Prometheus metrics endpoint. |
| `go.mongodb.org/mongo-driver/v2` | MongoDB connection, indexes, CRUD. |
| `github.com/redis/go-redis/v9` | Redis client, cache operations, locks. |
| `github.com/segmentio/kafka-go` | Kafka consumer and DLQ producer. |

### Practical commands

Run from the service folder:

```bash
cd backend/services/recommendation-service

go version
go mod download
go mod tidy
go test ./...
go build ./cmd/server
go run ./cmd/server
```

Run from the backend workspace:

```bash
cd backend
go test ./services/recommendation-service/...
```

### Common Go module issues

| Error | Cause | Fix |
|---|---|---|
| `go: go.mod requires go >= 1.26.3` | Installed Go version old hai. | Go `1.26.3` or compatible newer version install karo. |
| `module ... proto-gen/go not found` | Local generated proto module missing ya wrong working directory. | Repo root intact rakho. `replace` path `../../proto-gen/go` service folder se resolve hota hai. |
| `missing go.sum entry` | Dependency checksum absent hai. | `go mod tidy` run karo. |
| `GOPROXY` download failed | Network/proxy issue. | `go env GOPROXY`, network check karo; corporate proxy config karo. |
| Version mismatch | Go workspace aur module Go versions alag hain. | `backend/go.work`, service `go.mod`, and proto-gen `go.mod` versions align rakho. |

---

## 3. Databases

### MongoDB

#### A. What it is

MongoDB ek NoSQL document database hai. Simple English: tables ki jagah collections hoti hain, aur records JSON-like documents hote hain.

#### B. Why this project uses it

This service uses MongoDB for:

- `user_interactions`
- `product_features`
- `user_feature_profiles`
- `user_product_counters`
- `product_cooccurrence_features`
- `feature_job_runs`
- `feature_processed_events`
- `recommendation_sets`
- `ab_test_assignments`

#### C. Required or optional

| Scenario | MongoDB required? |
|---|---:|
| Minimal Task 1 type-definition endpoints | No |
| Storage status and storage plan only | No, but status will show unconfigured/unavailable. |
| Persisted recommendations | Yes |
| Feature builder enabled | Yes |
| Ranking/personalization from product features | Yes |
| Kafka event ingestion enabled | Yes |
| A/B assignment persistence | Yes for stored assignments; otherwise fail-open behavior can return default strategy. |

#### D. Local installation

Windows:

1. Install MongoDB Community Server using the MSI installer.
2. Install MongoDB Shell (`mongosh`).
3. Start the MongoDB Windows service from Services, or run `mongod` manually.
4. Verify:

```powershell
mongosh --version
mongosh "mongodb://localhost:27017"
```

Linux:

1. Install MongoDB Community Edition for your distro.
2. Start the service:

```bash
sudo systemctl start mongod
sudo systemctl enable mongod
mongosh "mongodb://localhost:27017"
```

macOS:

```bash
brew tap mongodb/brew
brew install mongodb-community mongosh
brew services start mongodb/brew/mongodb-community
mongosh "mongodb://localhost:27017"
```

#### E. Docker setup

Simple local MongoDB with username/password matching the existing sample env:

```bash
docker volume create reco_mongo_data

docker run -d \
  --name reco-mongo \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=root \
  -e MONGO_INITDB_ROOT_PASSWORD=secret \
  -v reco_mongo_data:/data/db \
  mongo:7
```

Verify:

```bash
mongosh "mongodb://root:secret@localhost:27017/admin"
```

Replica set note: feature builder repository code uses MongoDB transactions. Transactions need MongoDB replica set support. For easiest beginner setup, start minimal mode first by disabling events/features. For full feature jobs, use a single-node replica set or a proper dev replica set.

Single-node local replica set without auth for dev-only:

```bash
docker volume create reco_mongo_rs_data

docker run -d \
  --name reco-mongo-rs \
  -p 27017:27017 \
  -v reco_mongo_rs_data:/data/db \
  mongo:7 \
  --replSet rs0 --bind_ip_all

docker exec reco-mongo-rs mongosh --eval 'rs.initiate({_id:"rs0",members:[{_id:0,host:"localhost:27017"}]})'
```

Then use:

```env
RECOMMENDATION_MONGO_URI=mongodb://localhost:27017/?replicaSet=rs0
```

#### F. Start commands

Native:

```bash
sudo systemctl start mongod
```

Docker:

```bash
docker start reco-mongo
```

#### G. Verify running

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" --eval 'db.runCommand({ ping: 1 })'
```

#### H. Default port

MongoDB default port: `27017`.

#### I. Connection string format

With username/password:

```env
RECOMMENDATION_MONGO_URI=mongodb://root:secret@localhost:27017
RECOMMENDATION_MONGO_DATABASE=recommendation_db
```

Without auth:

```env
RECOMMENDATION_MONGO_URI=mongodb://localhost:27017
RECOMMENDATION_MONGO_DATABASE=recommendation_db
```

Replica set:

```env
RECOMMENDATION_MONGO_URI=mongodb://localhost:27017/?replicaSet=rs0
RECOMMENDATION_MONGO_DATABASE=recommendation_db
```

#### J. Where to place credentials

Put local credentials in:

```text
backend/services/recommendation-service/.env
```

But do not commit real credentials. `.gitignore` already ignores `.env` files. For production, use a secret manager or deployment environment variables.

---

## 4. Environment Variables

### Important `.env` loading behavior

The Go code uses `os.Getenv`. It does not automatically load `.env`.

That means creating `.env` is not enough. You must export it before running:

```bash
cd backend/services/recommendation-service
set -a
source .env
set +a
go run ./cmd/server
```

### Complete `.env` example

Create or update:

```text
backend/services/recommendation-service/.env
```

Minimal Task 1 beginner mode:

```env
SERVICE_NAME=recommendation-service
APP_ENV=local
RECOMMENDATION_HTTP_ADDR=:8088
RECOMMENDATION_GRPC_ADDR=:9088
RECOMMENDATION_LOG_LEVEL=debug

RECOMMENDATION_EVENTS_ENABLED=false
RECOMMENDATION_FEATURES_ENABLED=false
RECOMMENDATION_RANKING_ENABLED=false
RECOMMENDATION_PERSONALIZATION_ENABLED=false
RECOMMENDATION_AB_TESTING_ENABLED=false

RECOMMENDATION_DEFAULT_LIMIT=12
RECOMMENDATION_MAX_LIMIT=100
RECOMMENDATION_MAX_IDENTIFIER_LENGTH=128
```

Full local mode:

```env
# Basic app config
SERVICE_NAME=recommendation-service
APP_ENV=local
RECOMMENDATION_HTTP_ADDR=:8088
RECOMMENDATION_GRPC_ADDR=:9088
RECOMMENDATION_LOG_LEVEL=debug

# HTTP lifecycle
RECOMMENDATION_HTTP_READ_TIMEOUT=5s
RECOMMENDATION_HTTP_WRITE_TIMEOUT=10s
RECOMMENDATION_HTTP_IDLE_TIMEOUT=60s
RECOMMENDATION_SHUTDOWN_TIMEOUT=10s
RECOMMENDATION_MAX_BODY_BYTES=65536

# gRPC serving endpoint
RECOMMENDATION_GRPC_MAX_RECV_BYTES=65536
RECOMMENDATION_GRPC_MAX_SEND_BYTES=262144
RECOMMENDATION_GRPC_DEFAULT_DEADLINE=500ms

# Recommendation guardrails
RECOMMENDATION_DEFAULT_LIMIT=12
RECOMMENDATION_MAX_LIMIT=100
RECOMMENDATION_MAX_IDENTIFIER_LENGTH=128

# MongoDB durable store
RECOMMENDATION_STORAGE_FAIL_FAST=false
RECOMMENDATION_MONGO_URI=mongodb://root:secret@localhost:27017
RECOMMENDATION_MONGO_DATABASE=recommendation_db
RECOMMENDATION_MONGO_CONNECT_TIMEOUT=10s
RECOMMENDATION_MONGO_PING_TIMEOUT=5s
RECOMMENDATION_MONGO_ENSURE_INDEXES=true
RECOMMENDATION_INTERACTION_RETENTION_SECONDS=15552000

# Redis cache
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

# Kafka event ingestion
RECOMMENDATION_EVENTS_ENABLED=true
QUEUE_PROVIDER=kafka
KAFKA_BROKERS=localhost:9092
RECOMMENDATION_EVENTS_TOPIC=recommendation.events
RECOMMENDATION_EVENTS_RETRY_TOPIC=recommendation.events.retry
RECOMMENDATION_EVENTS_DLQ_TOPIC=recommendation.events.dlq
RECOMMENDATION_EVENTS_GROUP=recommendation-service-v1
RECOMMENDATION_EVENTS_MAX_MESSAGE_BYTES=262144
RECOMMENDATION_EVENTS_UNKNOWN_POLICY=dlq
EVENT_MAX_RETRY_ATTEMPTS=4
EVENT_RETRY_BACKOFF_SECONDS=5,30,120
EVENT_SUPPORTED_VERSION=1
RECOMMENDATION_KAFKA_MIN_BYTES=1
RECOMMENDATION_KAFKA_MAX_BYTES=10485760

# Feature builder
RECOMMENDATION_FEATURES_ENABLED=true
RECOMMENDATION_GUEST_PROFILE_RETENTION_SECONDS=2592000
RECOMMENDATION_USER_PRODUCT_RETENTION_SECONDS=15552000
RECOMMENDATION_FEATURE_EVENT_RETENTION_SECONDS=15552000
RECOMMENDATION_FEATURE_RECENT_PRODUCTS_LIMIT=20
RECOMMENDATION_FEATURE_RECONCILE_BATCH_SIZE=200
RECOMMENDATION_FEATURE_RECONCILE_INTERVAL=1m
RECOMMENDATION_FEATURE_WINDOW_REBUILD_INTERVAL=5m

# Rule-based ranking
RECOMMENDATION_RANKING_ENABLED=true
RECOMMENDATION_RANKING_REBUILD_INTERVAL=15m
RECOMMENDATION_RANKING_FORMULA_VERSION=popularity_v1
RECOMMENDATION_RANKING_VIEW_24H_WEIGHT=1
RECOMMENDATION_RANKING_WISHLIST_7D_WEIGHT=3
RECOMMENDATION_RANKING_CART_7D_WEIGHT=4
RECOMMENDATION_RANKING_PURCHASE_7D_WEIGHT=8

# Personalized ranking
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

# A/B testing
RECOMMENDATION_AB_TESTING_ENABLED=false
RECOMMENDATION_AB_ASSIGNMENT_TTL_DAYS=90
RECOMMENDATION_AB_DEFAULT_SALT=reco-ab-v1
RECOMMENDATION_AB_FAIL_OPEN=true
RECOMMENDATION_AB_EXPERIMENTS_JSON=[]
```

### Variable reference

| Variable | Purpose | Required? | Example | Security/config note |
|---|---|---:|---|---|
| `SERVICE_NAME` | Logs/metrics service name. | Optional, must not be empty | `recommendation-service` | Do not put secrets here. |
| `APP_ENV` | Runtime environment. | Optional, must not be empty | `local` | Use separate values for `local`, `staging`, `prod`. |
| `RECOMMENDATION_LOG_LEVEL` | JSON log level. | Optional | `debug` | Use `info` or `warn` in production. |
| `RECOMMENDATION_HTTP_ADDR` | HTTP listen address. | Optional | `:8088` | Bind to localhost/private network in local/staging if needed. |
| `RECOMMENDATION_HTTP_READ_TIMEOUT` | HTTP read timeout. | Optional | `5s` | Prevents slow-client resource abuse. |
| `RECOMMENDATION_HTTP_WRITE_TIMEOUT` | HTTP write timeout. | Optional | `10s` | Keep reasonable for API latency. |
| `RECOMMENDATION_HTTP_IDLE_TIMEOUT` | Keep-alive idle timeout. | Optional | `60s` | Tune behind load balancer. |
| `RECOMMENDATION_SHUTDOWN_TIMEOUT` | Graceful shutdown timeout. | Optional | `10s` | Increase if requests/jobs need longer. |
| `RECOMMENDATION_MAX_BODY_BYTES` | Max HTTP JSON request body. | Optional | `65536` | Prevents huge request payloads. |
| `RECOMMENDATION_GRPC_ADDR` | gRPC listen address. | Optional | `:9088` | Expose only inside trusted network unless gateway/auth exists. |
| `RECOMMENDATION_GRPC_MAX_RECV_BYTES` | Max gRPC request size. | Optional | `65536` | Protects against large payloads. |
| `RECOMMENDATION_GRPC_MAX_SEND_BYTES` | Max gRPC response size. | Optional | `262144` | Increase carefully if large recommendation lists are needed. |
| `RECOMMENDATION_GRPC_DEFAULT_DEADLINE` | Default gRPC deadline if client does not send one. | Optional | `500ms` | Avoid infinite backend waits. |
| `RECOMMENDATION_DEFAULT_LIMIT` | Default number of recommendation items. | Optional | `12` | Must be less than or equal to max limit. |
| `RECOMMENDATION_MAX_LIMIT` | Maximum allowed recommendation item count. | Optional | `100` | Prevents expensive queries. |
| `RECOMMENDATION_MAX_IDENTIFIER_LENGTH` | Max length for ids. | Optional | `128` | Protects cache keys and Mongo fields. |
| `RECOMMENDATION_STORAGE_FAIL_FAST` | Exit service if Mongo/Redis init fails. | Optional | `false` | Set `true` in prod when storage is mandatory. |
| `RECOMMENDATION_MONGO_URI` | MongoDB connection URI. | Conditional | `mongodb://root:secret@localhost:27017` | Secret value; never commit real password. |
| `RECOMMENDATION_MONGO_DATABASE` | Mongo database name. | Optional | `recommendation_db` | Keep separate per environment. |
| `RECOMMENDATION_MONGO_CONNECT_TIMEOUT` | Mongo connect timeout. | Optional | `10s` | Keep finite. |
| `RECOMMENDATION_MONGO_PING_TIMEOUT` | Mongo health ping timeout. | Optional | `5s` | Helps startup fail/warn quickly. |
| `RECOMMENDATION_MONGO_ENSURE_INDEXES` | Ensure indexes on startup. | Optional | `true` | In prod, migrations may be safer than app-time index creation. |
| `RECOMMENDATION_INTERACTION_RETENTION_SECONDS` | TTL for raw interactions. | Optional | `15552000` | 180 days; align with privacy policy. |
| `RECOMMENDATION_REDIS_ADDR` | Redis host:port. | Conditional | `localhost:6379` | Optional unless cache/locks are required. |
| `RECOMMENDATION_REDIS_PASSWORD` | Redis password. | Optional | empty or secret | Secret value; use auth in shared environments. |
| `RECOMMENDATION_REDIS_DB` | Redis logical DB number. | Optional | `0` | Separate envs by DB or instance. |
| `RECOMMENDATION_REDIS_DIAL_TIMEOUT` | Redis connect timeout. | Optional | `5s` | Keep finite. |
| `RECOMMENDATION_REDIS_READ_TIMEOUT` | Redis read timeout. | Optional | `3s` | Prevents stuck reads. |
| `RECOMMENDATION_REDIS_WRITE_TIMEOUT` | Redis write timeout. | Optional | `3s` | Prevents stuck writes. |
| `RECOMMENDATION_REDIS_PING_TIMEOUT` | Redis ping timeout. | Optional | `3s` | Startup verification. |
| `RECOMMENDATION_CACHE_KEY_PREFIX` | Redis key prefix. | Optional | `reco:v1` | Must not start/end with `:` or contain empty segments. |
| `RECOMMENDATION_CACHE_TTL_SECONDS` | Default cache TTL. | Optional | `900` | Short TTL keeps recommendations fresh. |
| `RECOMMENDATION_PERSONALIZED_CACHE_TTL_SECONDS` | Personalized cache TTL. | Optional | `300` | Keep low because user behavior changes quickly. |
| `RECOMMENDATION_GUEST_CACHE_TTL_SECONDS` | Guest personalized cache TTL. | Optional | `300` | Guest behavior changes quickly. |
| `RECOMMENDATION_CACHE_REBUILD_LOCK_TTL` | Redis lock TTL for rebuilds. | Optional | `1m` | Prevents duplicate rebuild jobs. |
| `RECOMMENDATION_CACHE_DIRTY_TTL_SECONDS` | Dirty marker TTL. | Optional | `900` | Helps invalidate changed product/category caches. |
| `RECOMMENDATION_EVENTS_ENABLED` | Enables Kafka consumer. | Optional | `true` or `false` | If true, Kafka brokers and MongoDB are required. |
| `QUEUE_PROVIDER` | Queue provider selector. | Conditional | `kafka` | Code supports only `kafka` for active ingestion. |
| `RECOMMENDATION_QUEUE_PROVIDER` | Alternate queue provider fallback. | Optional | `kafka` | `QUEUE_PROVIDER` has priority. |
| `KAFKA_BROKERS` | Kafka broker list. | Conditional | `localhost:9092` | Required only when events are enabled. |
| `RECOMMENDATION_EVENTS_TOPIC` | Kafka source topic. | Conditional | `recommendation.events` | Producers must publish to same topic. |
| `RECOMMENDATION_EVENTS_RETRY_TOPIC` | Retry topic name. | Optional currently | `recommendation.events.retry` | Keep for retry pipeline consistency. |
| `RECOMMENDATION_EVENTS_DLQ_TOPIC` | Dead letter queue topic. | Conditional | `recommendation.events.dlq` | Required when events enabled. |
| `RECOMMENDATION_EVENTS_GROUP` | Kafka consumer group id. | Conditional | `recommendation-service-v1` | Change carefully; it affects offsets. |
| `RECOMMENDATION_EVENTS_MAX_MESSAGE_BYTES` | Max Kafka message payload. | Optional | `262144` | Prevents oversized events. |
| `RECOMMENDATION_EVENTS_UNKNOWN_POLICY` | Unknown event behavior. | Optional | `dlq` | Allowed: `dlq`, `skip`. |
| `EVENT_MAX_RETRY_ATTEMPTS` | Handler retry attempts. | Optional | `4` | Must be greater than zero. |
| `EVENT_RETRY_BACKOFF_SECONDS` | Retry backoff list. | Optional | `5,30,120` | Comma-separated seconds or durations. |
| `EVENT_SUPPORTED_VERSION` | Event schema version. | Optional | `1` | Producers must send matching version. |
| `RECOMMENDATION_KAFKA_MIN_BYTES` | Kafka fetch min bytes. | Optional | `1` | Consumer tuning. |
| `RECOMMENDATION_KAFKA_MAX_BYTES` | Kafka fetch max bytes. | Optional | `10485760` | Consumer tuning. |
| `RECOMMENDATION_FEATURES_ENABLED` | Enables feature builder background jobs. | Optional | `true` | Requires MongoDB; transactions require replica set. |
| `RECOMMENDATION_GUEST_PROFILE_RETENTION_SECONDS` | Guest profile TTL. | Optional | `2592000` | 30 days; privacy-sensitive. |
| `RECOMMENDATION_USER_PRODUCT_RETENTION_SECONDS` | User product counter TTL. | Optional | `15552000` | 180 days. |
| `RECOMMENDATION_FEATURE_EVENT_RETENTION_SECONDS` | Processed event TTL. | Optional | `15552000` | Must be >= interaction retention. |
| `RECOMMENDATION_FEATURE_RECENT_PRODUCTS_LIMIT` | Recent products per profile. | Optional | `20` | Controls profile size. |
| `RECOMMENDATION_FEATURE_RECONCILE_BATCH_SIZE` | Reconcile batch size. | Optional | `200` | Bigger batch means more DB load. |
| `RECOMMENDATION_FEATURE_RECONCILE_INTERVAL` | Reconcile job interval. | Optional | `1m` | Avoid too frequent in local dev. |
| `RECOMMENDATION_FEATURE_WINDOW_REBUILD_INTERVAL` | Feature window rebuild interval. | Optional | `5m` | Background job load. |
| `RECOMMENDATION_RANKING_ENABLED` | Enables rule-based ranking job. | Optional | `true` | Requires MongoDB product features. |
| `RECOMMENDATION_RANKING_REBUILD_INTERVAL` | Ranking rebuild interval. | Optional | `15m` | Controls freshness/load. |
| `RECOMMENDATION_RANKING_FORMULA_VERSION` | Ranking formula label. | Optional | `popularity_v1` | If weights change, version must change. |
| `RECOMMENDATION_RANKING_VIEW_24H_WEIGHT` | View weight. | Optional | `1` | Non-negative. |
| `RECOMMENDATION_RANKING_WISHLIST_7D_WEIGHT` | Wishlist add weight. | Optional | `3` | Non-negative. |
| `RECOMMENDATION_RANKING_CART_7D_WEIGHT` | Cart add weight. | Optional | `4` | Non-negative. |
| `RECOMMENDATION_RANKING_PURCHASE_7D_WEIGHT` | Purchase weight. | Optional | `8` | Non-negative; at least one ranking weight must be positive. |
| `RECOMMENDATION_PERSONALIZATION_ENABLED` | Enables personalized ranking. | Optional | `true` | Requires MongoDB profiles/features. |
| `RECOMMENDATION_PERSONALIZED_FORMULA_VERSION` | Personalization formula label. | Optional | `behavior_v1` | If weights change, version must change. |
| `RECOMMENDATION_PERSONALIZED_MIN_POSITIVE_INTERACTIONS` | Minimum interactions before personalization. | Optional | `3` | Prevents noisy personalization. |
| `RECOMMENDATION_PERSONALIZED_PROFILE_MAX_AGE_DAYS` | Max profile age. | Optional | `30` | Privacy and freshness. |
| `RECOMMENDATION_PERSONALIZED_MAX_CANDIDATES` | Candidate query size. | Optional | `250` | Must be >= default limit. |
| `RECOMMENDATION_PERSONALIZED_MAX_ITEMS_PER_SELLER` | Seller diversity cap. | Optional | `3` | Prevents one seller dominating. |
| `RECOMMENDATION_PERSONALIZED_TOP_CATEGORIES` | Top categories used from profile. | Optional | `3` | Candidate shaping. |
| `RECOMMENDATION_PERSONALIZED_TOP_SELLERS` | Top sellers used from profile. | Optional | `2` | Candidate shaping. |
| `RECOMMENDATION_PERSONALIZED_TOP_BRANDS` | Top brands used from profile. | Optional | `2` | Candidate shaping. |
| `RECOMMENDATION_PERSONALIZED_DIRECT_PRODUCT_LIMIT` | Direct product affinity limit. | Optional | `20` | Candidate shaping. |
| `RECOMMENDATION_PERSONALIZED_CATEGORY_WEIGHT` | Category signal weight. | Optional | `0.35` | All personalization weights must total `1.00`. |
| `RECOMMENDATION_PERSONALIZED_SELLER_WEIGHT` | Seller signal weight. | Optional | `0.15` | Non-negative. |
| `RECOMMENDATION_PERSONALIZED_BRAND_WEIGHT` | Brand signal weight. | Optional | `0.10` | Non-negative. |
| `RECOMMENDATION_PERSONALIZED_PRICE_WEIGHT` | Price signal weight. | Optional | `0.10` | Non-negative. |
| `RECOMMENDATION_PERSONALIZED_DIRECT_AFFINITY_WEIGHT` | Direct product affinity weight. | Optional | `0.20` | Non-negative. |
| `RECOMMENDATION_PERSONALIZED_POPULARITY_WEIGHT` | Popularity baseline weight. | Optional | `0.10` | Non-negative. |
| `RECOMMENDATION_AB_TESTING_ENABLED` | Enables A/B strategy assignment. | Optional | `false` | Keep off until experiments are configured. |
| `RECOMMENDATION_AB_ASSIGNMENT_TTL_DAYS` | A/B assignment TTL. | Optional | `90` | Controls assignment persistence duration. |
| `RECOMMENDATION_AB_DEFAULT_SALT` | Hash salt for bucketing. | Optional | `reco-ab-v1` | Treat as config; rotate carefully. |
| `RECOMMENDATION_AB_FAIL_OPEN` | Return default strategy if assignment fails. | Optional | `true` | Good for availability; monitor errors. |
| `RECOMMENDATION_AB_EXPERIMENTS_JSON` | JSON experiment definitions. | Optional | `[]` | Must be valid JSON; do not include secrets. |

Common `.env` mistakes:

- File created but not exported before `go run`.
- Mongo URI has password but missing `authSource=admin` when database auth setup needs it.
- `RECOMMENDATION_EVENTS_ENABLED=true` but Kafka is not running.
- `RECOMMENDATION_FEATURES_ENABLED=true` with standalone MongoDB, then transaction-related errors appear.
- Personalized weights do not total `1.00`.
- Ranking weights changed but formula version still `popularity_v1`.

---

## 5. External Services

### Redis

What it is: Redis ek fast in-memory key-value store hai.

Why used: recommendation cache, personalized cache, dirty markers, and rebuild locks.

Required or optional:

- Optional for minimal Task 1.
- Recommended for production/full service.
- If not configured, service can still start and logs `recommendation.redis.not_configured`.

Install:

Windows:

- Recommended: run Redis using Docker or WSL.

Linux:

```bash
sudo apt-get update
sudo apt-get install -y redis-server
sudo systemctl start redis-server
redis-cli ping
```

macOS:

```bash
brew install redis
brew services start redis
redis-cli ping
```

Docker:

```bash
docker volume create reco_redis_data

docker run -d \
  --name reco-redis \
  -p 6379:6379 \
  -v reco_redis_data:/data \
  redis:7 redis-server --appendonly yes
```

Verify:

```bash
redis-cli -h localhost -p 6379 ping
```

Credential placement:

```env
RECOMMENDATION_REDIS_ADDR=localhost:6379
RECOMMENDATION_REDIS_PASSWORD=
RECOMMENDATION_REDIS_DB=0
```

Common issues:

- `connection refused`: Redis container/service not running.
- Password mismatch: set `RECOMMENDATION_REDIS_PASSWORD`.
- Docker container app cannot reach `localhost`: use Docker service name like `redis:6379`.

### Kafka

What it is: Kafka ek message streaming platform hai. Simple English: services events publish karte hain, consumers events read karte hain.

Why used: Product/session/order/wishlist events se user interaction data create hota hai.

Required or optional:

- Optional for minimal Task 1.
- Required only when `RECOMMENDATION_EVENTS_ENABLED=true`.
- If enabled, `QUEUE_PROVIDER` must be `kafka` and `KAFKA_BROKERS` must be set.

Docker setup:

```yaml
services:
  kafka:
    image: apache/kafka:3.8.0
    container_name: reco-kafka
    ports:
      - "9092:9092"
    environment:
      KAFKA_NODE_ID: 1
      KAFKA_PROCESS_ROLES: broker,controller
      KAFKA_CONTROLLER_QUORUM_VOTERS: 1@localhost:9093
      KAFKA_LISTENERS: PLAINTEXT://:9092,CONTROLLER://:9093
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_CONTROLLER_LISTENER_NAMES: CONTROLLER
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR: 1
      KAFKA_TRANSACTION_STATE_LOG_MIN_ISR: 1
      KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS: 0
```

Create topics:

```bash
docker exec reco-kafka /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create --if-not-exists \
  --topic recommendation.events \
  --partitions 3 \
  --replication-factor 1

docker exec reco-kafka /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create --if-not-exists \
  --topic recommendation.events.dlq \
  --partitions 3 \
  --replication-factor 1

docker exec reco-kafka /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --list
```

Credential/config placement:

```env
RECOMMENDATION_EVENTS_ENABLED=true
QUEUE_PROVIDER=kafka
KAFKA_BROKERS=localhost:9092
RECOMMENDATION_EVENTS_TOPIC=recommendation.events
RECOMMENDATION_EVENTS_DLQ_TOPIC=recommendation.events.dlq
```

Expected event envelope:

```json
{
  "event_id": "evt_123",
  "event_type": "ProductViewed",
  "version": 1,
  "occurred_at": "2026-06-05T10:00:00Z",
  "producer": "session-service",
  "trace_id": "trace_123",
  "payload": {
    "user_id": "user_1",
    "anonymous_id": "",
    "session_id": "sess_1",
    "product_id": "prod_1",
    "category_id": "cat_1",
    "seller_id": "seller_1",
    "brand_id": "brand_1",
    "quantity": 1,
    "page": "product_detail"
  }
}
```

Supported event types:

- `ProductViewed`
- `CartItemAdded`
- `WishlistItemAdded`
- `WishlistItemRemoved`
- `OrderPaid`
- `PurchaseCompleted`

Common issues:

- `KAFKA_BROKERS cannot be empty`: events enabled but broker env missing.
- Topic not found: create `recommendation.events` and `recommendation.events.dlq`.
- App in Docker cannot reach `localhost:9092`: use container network broker name and advertised listeners.

### Prometheus

What it is: Prometheus metrics scrape system hai.

Why used: `/metrics` endpoint exposes service metrics.

Required or optional: Optional, but recommended for observability.

Health check:

```bash
curl http://localhost:8088/metrics
```

### Docker

What it is: Docker containers run karta hai so local infra easy ho jati hai.

Why used: MongoDB, Redis, and Kafka local setup ke liye.

Current repo status: no service Dockerfile or root docker-compose file was found for this service. So the examples below run dependencies only.

### Not detected

No direct usage detected for MySQL, PostgreSQL, RabbitMQ, NATS, MinIO, Elasticsearch, Nginx, AWS S3, Firebase, SMTP, Stripe, Twilio, OAuth providers, or Kubernetes in this task/service code.

---

## 6. Ports and Networking

| Service | Port | Purpose |
|---|---:|---|
| HTTP API | `8088` | Health, metrics, type definitions, storage endpoints. |
| gRPC API | `9088` | `GetRecommendations` RPC and gRPC health service. |
| MongoDB | `27017` | Durable recommendation storage. |
| Redis | `6379` | Cache and locks. |
| Kafka | `9092` | Event ingestion broker. |

How to change ports:

```env
RECOMMENDATION_HTTP_ADDR=:8090
RECOMMENDATION_GRPC_ADDR=:9090
RECOMMENDATION_MONGO_URI=mongodb://localhost:27018
RECOMMENDATION_REDIS_ADDR=localhost:6380
KAFKA_BROKERS=localhost:9094
```

Port conflict checks:

Linux/macOS:

```bash
lsof -i :8088
lsof -i :9088
lsof -i :27017
lsof -i :6379
lsof -i :9092
```

Windows PowerShell:

```powershell
netstat -ano | findstr :8088
netstat -ano | findstr :9088
```

Docker networking notes:

- App running on host should use `localhost:27017`, `localhost:6379`, `localhost:9092`.
- App running inside Docker should use compose service names like `mongo:27017`, `redis:6379`, `kafka:9092`.
- If Kafka is in Docker, advertised listeners must match how the app connects.

Firewall notes:

- Local firewall can block `8088`, `9088`, `27017`, `6379`, or `9092`.
- Production should not publicly expose MongoDB/Redis/Kafka.
- gRPC should normally stay inside private service network.

---

## 7. Docker and DevOps Setup

### Dependency-only `docker-compose.yml` example

Create this locally if you want one-command infra:

```yaml
services:
  mongo:
    image: mongo:7
    container_name: reco-mongo
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: secret
    volumes:
      - reco_mongo_data:/data/db
    restart: unless-stopped

  redis:
    image: redis:7
    container_name: reco-redis
    ports:
      - "6379:6379"
    command: ["redis-server", "--appendonly", "yes"]
    volumes:
      - reco_redis_data:/data
    restart: unless-stopped

  kafka:
    image: apache/kafka:3.8.0
    container_name: reco-kafka
    ports:
      - "9092:9092"
    environment:
      KAFKA_NODE_ID: 1
      KAFKA_PROCESS_ROLES: broker,controller
      KAFKA_CONTROLLER_QUORUM_VOTERS: 1@localhost:9093
      KAFKA_LISTENERS: PLAINTEXT://:9092,CONTROLLER://:9093
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_CONTROLLER_LISTENER_NAMES: CONTROLLER
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR: 1
      KAFKA_TRANSACTION_STATE_LOG_MIN_ISR: 1
      KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS: 0
    restart: unless-stopped

volumes:
  reco_mongo_data:
  reco_redis_data:
```

Commands:

```bash
docker compose up -d
docker compose ps
docker compose logs mongo
docker compose logs redis
docker compose logs kafka
docker compose down
```

Persistent volumes:

- `reco_mongo_data` keeps MongoDB data.
- `reco_redis_data` keeps Redis append-only data.

Networks:

- Docker Compose creates a private default network.
- Containers can talk using service names: `mongo`, `redis`, `kafka`.

Recommended approach for beginners:

1. Start minimal mode without Mongo/Redis/Kafka.
2. Add MongoDB.
3. Add Redis.
4. Add Kafka only when event ingestion is needed.

---

## 8. Project Run Instructions

### Step 1: Clone repository

```bash
git clone <repository-url>
cd Ecommerce
```

### Step 2: Go to project directory

```bash
cd backend/services/recommendation-service
```

### Step 3: Install Go dependencies

```bash
go version
go mod download
```

### Step 4: Minimal Task 1 run

Use this when you only want recommendation types/contexts/resolution docs via API.

```bash
cat > .env.local-minimal <<'EOF'
SERVICE_NAME=recommendation-service
APP_ENV=local
RECOMMENDATION_HTTP_ADDR=:8088
RECOMMENDATION_GRPC_ADDR=:9088
RECOMMENDATION_LOG_LEVEL=debug
RECOMMENDATION_EVENTS_ENABLED=false
RECOMMENDATION_FEATURES_ENABLED=false
RECOMMENDATION_RANKING_ENABLED=false
RECOMMENDATION_PERSONALIZATION_ENABLED=false
RECOMMENDATION_AB_TESTING_ENABLED=false
RECOMMENDATION_DEFAULT_LIMIT=12
RECOMMENDATION_MAX_LIMIT=100
RECOMMENDATION_MAX_IDENTIFIER_LENGTH=128
EOF

set -a
source .env.local-minimal
set +a

go run ./cmd/server
```

Verify in another terminal:

```bash
curl http://localhost:8088/healthz
curl http://localhost:8088/internal/v1/recommendation/types
curl http://localhost:8088/internal/v1/recommendation/contexts
curl http://localhost:8088/internal/v1/recommendation/storage/status
```

Resolve type example:

```bash
curl -X POST http://localhost:8088/internal/v1/recommendation/resolve-type \
  -H 'Content-Type: application/json' \
  -d '{
    "context": "product_detail",
    "product_id": "prod_123",
    "category_id": "cat_shoes",
    "limit": 12
  }'
```

### Step 5: Full local dependency setup

Start MongoDB:

```bash
docker run -d \
  --name reco-mongo \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=root \
  -e MONGO_INITDB_ROOT_PASSWORD=secret \
  -v reco_mongo_data:/data/db \
  mongo:7
```

Start Redis:

```bash
docker run -d \
  --name reco-redis \
  -p 6379:6379 \
  -v reco_redis_data:/data \
  redis:7 redis-server --appendonly yes
```

Start Kafka using the compose example above if event ingestion is needed.

### Step 6: Create `.env`

Use the full `.env` example above, or copy the existing local `.env` and adjust credentials:

```bash
cd backend/services/recommendation-service
set -a
source .env
set +a
```

### Step 7: Run migrations

Migrations are Mongo shell JavaScript files:

```text
backend/services/recommendation-service/migrations/*.up.js
```

Run all up migrations:

```bash
cd backend/services/recommendation-service

export RECOMMENDATION_MONGO_DATABASE=recommendation_db

mongosh "mongodb://root:secret@localhost:27017/admin" --file migrations/001_create_recommendation_storage.up.js
mongosh "mongodb://root:secret@localhost:27017/admin" --file migrations/002_add_event_ingestion_indexes.up.js
mongosh "mongodb://root:secret@localhost:27017/admin" --file migrations/003_create_feature_store_schema.up.js
mongosh "mongodb://root:secret@localhost:27017/admin" --file migrations/004_add_rule_based_ranking_indexes.up.js
mongosh "mongodb://root:secret@localhost:27017/admin" --file migrations/005_add_personalized_ranking_indexes.up.js
mongosh "mongodb://root:secret@localhost:27017/admin" --file migrations/006_add_ab_test_assignment_indexes.up.js
```

Verify collections:

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" \
  --eval 'db.getSiblingDB("recommendation_db").getCollectionNames()'
```

### Step 8: Start backend service

```bash
cd backend/services/recommendation-service

set -a
source .env
set +a

go run ./cmd/server
```

### Step 9: Verify APIs

HTTP:

```bash
curl http://localhost:8088/healthz
curl http://localhost:8088/metrics
curl http://localhost:8088/internal/v1/recommendation/types
curl http://localhost:8088/internal/v1/recommendation/contexts
curl http://localhost:8088/internal/v1/recommendation/storage/status
```

gRPC health, if `grpcurl` is installed:

```bash
grpcurl -plaintext localhost:9088 grpc.health.v1.Health/Check
```

Get recommendations, if `grpcurl` and server reflection/client proto setup are available:

```bash
grpcurl -plaintext \
  -import-path ../../../proto \
  -proto ecommerce/recommendation/v1/recommendation.proto \
  -d '{"context":"RECOMMENDATION_CONTEXT_HOME_FEED","anonymous_id":"anon_1","limit":12}' \
  localhost:9088 \
  ecommerce.recommendation.v1.RecommendationService/GetRecommendations
```

---

## 9. Common Errors and Fixes

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `address already in use` | Port `8088` or `9088` already occupied. | Change `RECOMMENDATION_HTTP_ADDR` or `RECOMMENDATION_GRPC_ADDR`, or stop old process. | Keep service ports documented. |
| `recommendation.config.load_failed` | Invalid env values. | Check logs for exact variable. | Start with minimal env first. |
| `.env created but app ignores it` | App does not auto-load `.env`. | Run `set -a; source .env; set +a`. | Add this step to local run scripts. |
| `mongo uri is required` or Mongo unconfigured logs | `RECOMMENDATION_MONGO_URI` empty. | Set Mongo URI or disable full features/events. | Use minimal mode for Task 1. |
| `connect mongo` / `ping mongo` failed | MongoDB not running, wrong URI, wrong auth. | Start Mongo, verify with `mongosh`, fix credentials. | Use Docker command and matching env. |
| `event ingestion requires configured MongoDB storage` | `RECOMMENDATION_EVENTS_ENABLED=true` but Mongo repo unavailable. | Set valid Mongo URI or set events false. | Events require durable storage. |
| Mongo transaction error | Standalone MongoDB used with feature jobs. | Use replica set or disable feature jobs locally. | Use replica set for full mode. |
| `redis: connection refused` | Redis not running or wrong addr. | Start Redis, verify `redis-cli ping`. | Keep Redis optional for minimal run. |
| `KAFKA_BROKERS cannot be empty` | Events enabled without brokers. | Set `KAFKA_BROKERS=localhost:9092` or disable events. | Keep events false until Kafka is ready. |
| Kafka topic missing | Topic was not created. | Create `recommendation.events` and DLQ topic. | Use compose/bootstrap script. |
| `QUEUE_PROVIDER ... not supported` | Provider not `kafka` while events active. | Use `QUEUE_PROVIDER=kafka` or disable events. | Do not set unsupported queue names. |
| `go mod download failed` | Network, proxy, or Go version issue. | Check Go version and proxy/network. | Install required Go version and keep `go.sum`. |
| `go: module proto-gen/go not found` | Local replace path broken. | Run from intact repo; do not move service folder alone. | Clone whole repository. |
| `Permission denied` | Script/file/service permission issue. | Use correct shell permissions; avoid running Docker without group access. | Add user to docker group on Linux if needed. |
| `RECOMMENDATION_FEATURE_EVENT_RETENTION_SECONDS must be greater than or equal...` | Feature processed-event TTL less than interaction TTL. | Increase feature event retention or reduce interaction retention. | Keep both `15552000` for default. |
| `personalization weights must total 1.00` | Weight envs changed incorrectly. | Make six weights sum exactly `1.00`. | Change formula version when tuning. |
| `ranking formula version must change...` | Ranking weights changed but version stayed `popularity_v1`. | Set `RECOMMENDATION_RANKING_FORMULA_VERSION=popularity_v2`. | Version every ranking formula change. |
| `Docker daemon not running` | Docker Desktop/service stopped. | Start Docker Desktop or `sudo systemctl start docker`. | Verify with `docker ps`. |
| `Migration failed` | Mongo auth/URI/database issue or existing conflicting indexes. | Verify `mongosh`, database name, and indexes. | Run migrations in order and keep DB backups. |

---

## 10. Security and Configuration Audit

| Finding | Risk | Suggested fix |
|---|---|---|
| Local `.env` contains `root:secret` example credentials. | Fine for local only, unsafe if copied to staging/prod. | Keep `.env` ignored, create `.env.example`, use secret manager in real deployment. |
| `.env` is not auto-loaded. | Beginners may think config is active when it is not. | Add a local run script or document `set -a; source .env; set +a`. |
| MongoDB can be optional in minimal mode. | Full features silently degrade if `RECOMMENDATION_STORAGE_FAIL_FAST=false`. | Set `RECOMMENDATION_STORAGE_FAIL_FAST=true` in production. |
| App can ensure Mongo indexes on startup. | Startup index creation may slow or surprise production deploys. | Use migrations in production and consider `RECOMMENDATION_MONGO_ENSURE_INDEXES=false`. |
| Kafka events have max payload and schema validation. | Good, but producer contracts must match. | Document event envelope with producer teams. |
| No service Dockerfile detected. | Deployment packaging is incomplete. | Add Dockerfile and deployment health checks before production. |
| No root docker-compose detected. | Beginners need manual dependency setup. | Add dependency compose file or local dev script. |
| gRPC/HTTP ports expose internal APIs. | Unsafe if public without gateway/auth. | Keep behind private network or add auth/API gateway. |
| Redis password empty in local example. | Unsafe for shared hosts. | Use Redis auth/TLS or private network in non-local environments. |
| A/B salt default is public config. | Predictable bucketing if treated as secret. | Use environment-specific salt if needed; rotate carefully. |
| Feature builder uses Mongo transactions. | Standalone MongoDB can fail. | Use Mongo replica set for full local/prod mode. |
| No explicit Kubernetes manifests detected. | Production orchestration assumptions not codified. | Add readiness/liveness probes for `/healthz` and gRPC health. |

---

## 11. Best Practices

- Never commit real `.env` files.
- Keep a safe `.env.example` with dummy credentials.
- Start minimal mode first; add Mongo, Redis, Kafka one by one.
- Use Docker volumes for MongoDB and Redis so data does not disappear on container restart.
- Use MongoDB backups before running down migrations or index changes.
- Use strong, environment-specific database and Redis passwords.
- Keep MongoDB, Redis, and Kafka private; do not expose them publicly.
- Set `RECOMMENDATION_STORAGE_FAIL_FAST=true` in production when the service must not run degraded.
- Keep `RECOMMENDATION_EVENTS_ENABLED=false` until Kafka topics and producers are ready.
- Version ranking and personalization formulas when changing weights.
- Keep personalization data retention aligned with privacy requirements.
- Use `/healthz`, gRPC health, `/metrics`, and structured logs for debugging.
- Document producer event contracts with Product, Session, Order, and Wishlist teams.
- Run `go test ./...` before deployment.
- Keep dependencies updated carefully and review `go.mod`/`go.sum` diffs.

---

## 12. Final Checklist

- [ ] Go `1.26.3` or compatible version installed.
- [ ] Repository cloned completely, not only the service folder.
- [ ] `go mod download` completed.
- [ ] Minimal env or full `.env` created.
- [ ] `.env` exported before running the service.
- [ ] MongoDB installed/running if full storage/features are enabled.
- [ ] MongoDB migrations executed in order.
- [ ] Redis running if cache is configured.
- [ ] Kafka running if event ingestion is enabled.
- [ ] Kafka topics created.
- [ ] HTTP port `8088` available.
- [ ] gRPC port `9088` available.
- [ ] Backend service started with `go run ./cmd/server`.
- [ ] `curl http://localhost:8088/healthz` returns OK.
- [ ] `/internal/v1/recommendation/types` returns type definitions.
- [ ] `/internal/v1/recommendation/storage/status` checked.
- [ ] `/metrics` checked.
- [ ] Logs checked for Mongo/Redis/Kafka degraded mode warnings.
- [ ] Common errors reviewed.
- [ ] Real credentials kept out of git.
