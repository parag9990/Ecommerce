# Project Dependency & Setup Guide

## Variables Used

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Recommendation Service` |
| `TASK_FILE_NAME` | `task3.md` |
| `OUTPUT_FILE_NAME` | `task3_Dependency.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |
| Backend code path | `backend/services/recommendation-service` |

Beginner note: Ye file sirf dependency, setup, environment, queue, database, Docker, and DevOps onboarding ke liye hai. Business logic ya original implementation file modify nahi kiya gaya.

---

## 1. Project Overview

`TASK_FILE_NAME` ka focus hai: `SERVICE_NAME` me event ingestion enable karna.

Simple Hinglish:

- Source services user actions publish karenge, jaise product view, add-to-cart, wishlist, aur purchase.
- `SERVICE_NAME` Kafka se events consume karega.
- Valid event MongoDB `recommendation_db.user_interactions` collection me idempotently store hoga.
- Redis cache invalidation best-effort side effect hai. Redis down ho to Mongo write successful event fail nahi hona chahiye.
- Invalid events DLQ topic me jayenge.

Read these previous dependency files first:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`1. Project Tech Stack Analysis`
`2. Go Dependency System`
`5. External Services` -> `Kafka`
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

`TASK_FILE_NAME` new setup scope:

- Kafka event consumer using `github.com/segmentio/kafka-go`
- Kafka topic: `recommendation.events`
- Kafka DLQ topic: `recommendation.events.dlq`
- Consumer group: `recommendation-service-v1`
- MongoDB migration: `backend/services/recommendation-service/migrations/002_add_event_ingestion_indexes.up.js`
- New event-ingestion env variables
- Verification by publishing a sample Kafka event and checking MongoDB plus `/metrics`

---

## 2. Tech Stack

Most base technologies are already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `1. Project Tech Stack Analysis`
```

MongoDB and Redis storage/cache choice is already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
Sections: `5. Database Setup`, `6. Redis / Queue / External Services`
```

`TASK_FILE_NAME`-specific technologies:

| Technology | Required? | Beginner Explanation | Why Current Task Uses It |
|---|---:|---|---|
| Kafka | Yes for full `TASK_FILE_NAME` mode | Kafka ek event streaming platform hai. Simple English: producers events bhejte hain, consumers later read karte hain. | Product view/cart/wishlist/purchase events ko reliably consume karne ke liye. |
| `kafka-go` | Yes | Go library jo Kafka se message read/write karne me help karti hai. | Service `recommendation.events` consume karta hai and DLQ me failed events publish karta hai. |
| MongoDB | Yes when events are enabled | Document database. JSON-like documents collections me store hote hain. | Normalized `user_interactions` durable store karne ke liye. |
| Redis | Optional for `TASK_FILE_NAME` | Fast in-memory cache. | New interactions ke baad personalized/trending cache keys invalidate/dirty mark karne ke liye. |
| Prometheus client | Reused, recommended | Metrics expose karne ki library. | Event consumed, duplicate, DLQ, processing duration, and lag metrics expose karta hai. |
| RabbitMQ | Not supported by current code | RabbitMQ ek message broker hai. | `TASK_FILE_NAME` guide me fallback mention hai, but current backend validation only allows Kafka when ingestion is active. |

Important implementation fact:

```text
QUEUE_PROVIDER=rabbitmq
```

current code me active event ingestion ke saath fail karega, because `config.EventConfig.Validate()` only supports `kafka`.

---

## 3. Required Software

Base software installation is already covered in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Sections:
`2. Go Dependency System`
`3. Databases`
`5. External Services`
`7. Docker and DevOps Setup`
```

For `TASK_FILE_NAME`, make sure these are ready:

| Software | Needed For | New or Reused |
|---|---|---|
| Go `1.26.3` or compatible newer | Build and run backend service | Reused |
| MongoDB `7.x` | Store `user_interactions` | Reused, mandatory when events enabled |
| `mongosh` | Run migration `002_add_event_ingestion_indexes.up.js` | Reused, task-specific migration |
| Redis `7.x` | Cache invalidation side effect | Reused, optional for `TASK_FILE_NAME` |
| Kafka `3.x` | Event ingestion queue | Reused setup, mandatory for `TASK_FILE_NAME` |
| Kafka CLI or container shell | Create topics and publish test event | Current-task verification |
| `curl` | Check HTTP health and metrics | Reused |

No new MySQL, PostgreSQL, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, OAuth, Kubernetes, or Nginx dependency is introduced by `TASK_FILE_NAME`.

---

## 4. Dependency Management

This is a Go module project. Go modules, `go.mod`, `go.sum`, `go mod download`, and common Go module errors are already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `2. Go Dependency System`
```

`TASK_FILE_NAME` dependency status:

| Dependency | Version In `go.mod` | Purpose |
|---|---:|---|
| `github.com/segmentio/kafka-go` | `v0.4.51` | Kafka consumer and DLQ producer |
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | MongoDB insert, duplicate key handling, indexes |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Best-effort cache invalidation |
| `github.com/prometheus/client_golang` | `v1.23.2` | `/metrics` endpoint |

No new `go get` command is needed because these dependencies already exist.

Recommended verification:

```bash
cd backend/services/recommendation-service
go mod download
go test ./...
```

RabbitMQ note:

```text
github.com/rabbitmq/amqp091-go
```

is not present in `go.mod`, so RabbitMQ setup should not be followed for the current implementation.

---

## 5. Database Setup

### MongoDB Role

MongoDB is mandatory when event ingestion is enabled.

Reason: `cmd/server/main.go` initializes the interaction consumer only when MongoDB repository exists. Agar `RECOMMENDATION_EVENTS_ENABLED=true` hai but MongoDB configured nahi hai, service startup fail karega with:

```text
event ingestion requires configured MongoDB storage
```

### Reused MongoDB Setup

MongoDB installation, Docker setup, connection string format, and base migration are already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `3. Databases` -> `MongoDB`
```

```md
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
Section: `5. Database Setup`
```

Do not recreate MongoDB if it is already running from the previous setup.

### Current Task Migration

`TASK_FILE_NAME` adds task-specific indexes for event ingestion:

```text
backend/services/recommendation-service/migrations/002_add_event_ingestion_indexes.up.js
```

What it does:

| Collection | Index | Why |
|---|---|---|
| `user_interactions` | `ux_user_interactions_dedupe_key` | Duplicate events safe karne ke liye unique `dedupe_key` |
| `user_interactions` | `idx_user_interactions_product_normalized_event_recent` | Product/event recent reads fast karne ke liye |
| `user_interactions` | `idx_user_interactions_category_normalized_event_recent` | Category/event recent reads fast karne ke liye |

Prerequisite: `001_create_recommendation_storage.up.js` should already be run from the previous task. If not, run `001` first, then `002`.

```bash
cd backend/services/recommendation-service
export RECOMMENDATION_MONGO_DATABASE=recommendation_db

mongosh "mongodb://root:secret@localhost:27017/admin" \
  --file migrations/001_create_recommendation_storage.up.js

mongosh "mongodb://root:secret@localhost:27017/admin" \
  --file migrations/002_add_event_ingestion_indexes.up.js
```

If your local MongoDB has no auth:

```bash
mongosh "mongodb://localhost:27017" \
  --file migrations/002_add_event_ingestion_indexes.up.js
```

### Verify MongoDB Indexes

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" \
  --eval 'db.getSiblingDB("recommendation_db").user_interactions.getIndexes().map(i => i.name)'
```

Expected current-task index names:

```text
ux_user_interactions_dedupe_key
idx_user_interactions_product_normalized_event_recent
idx_user_interactions_category_normalized_event_recent
```

Credentials placement:

```text
backend/services/recommendation-service/.env
```

Use local dummy values only for local development. Real passwords should come from environment-specific secret management.

---

## 6. Redis / Queue / External Services

### Kafka Role

Kafka is the new mandatory external service for full `TASK_FILE_NAME` verification.

Simple Hinglish:

Kafka events ka waiting room hai. Source service event publish karta hai, and `SERVICE_NAME` apne consumer group ke through event read karta hai.

### Reused Kafka Setup

Kafka installation and Docker Compose basics are already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `5. External Services` -> `Kafka`
Section: `7. Docker and DevOps Setup`
```

Use the same Kafka container/Compose setup. Current task only needs the exact topics below.

### Current Task Kafka Topics

| Topic | Required? | Purpose |
|---|---:|---|
| `recommendation.events` | Yes | Main event ingestion topic |
| `recommendation.events.dlq` | Yes | Invalid or failed events are published here |
| `recommendation.events.retry` | Configured, future-ready | Env exists, but current code retries in-process and does not publish to retry topic |

Create topics using the Kafka CLI available in your local Kafka container.

For the Apache Kafka container name used in the previous guide:

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
  --create --if-not-exists \
  --topic recommendation.events.retry \
  --partitions 3 \
  --replication-factor 1
```

Verify:

```bash
docker exec reco-kafka /opt/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --list
```

### RabbitMQ Status

`TASK_FILE_NAME` implementation guide mentions Kafka/RabbitMQ conceptually, but current backend code supports Kafka only.

Do not add RabbitMQ env vars for this task. If you set:

```env
QUEUE_PROVIDER=rabbitmq
```

the service will fail while loading config.

### Redis Role

Redis setup is reused from:

```md
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
Section: `6. Redis / Queue / External Services`
```

For `TASK_FILE_NAME`:

- Redis is optional.
- If Redis is configured, ingested interactions invalidate related cache keys.
- If Redis fails after MongoDB insert, service logs a warning and continues.

### Ports and Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Recommendation HTTP | `8088` | `/healthz`, `/metrics`, storage/type endpoints | Reused |
| Recommendation gRPC | `9088` | Internal recommendation gRPC service | Reused |
| MongoDB | `27017` | `user_interactions` storage | Reused |
| Redis | `6379` | Best-effort cache invalidation | Reused |
| Kafka broker | `9092` | Event ingestion topic and DLQ | Current-task full mode required |
| Kafka controller | `9093` | Local single-node Kafka controller | Reused Compose detail |
| RabbitMQ AMQP | `5672` | Not used by current implementation | Not required |
| RabbitMQ UI | `15672` | Not used by current implementation | Not required |

Networking rules:

- App running on host: use `KAFKA_BROKERS=localhost:9092`.
- App running inside Docker Compose: use `KAFKA_BROKERS=kafka:9092`.
- Kafka advertised listeners must match the client location.
- MongoDB, Redis, Kafka, and gRPC should not be publicly exposed in production.

---

## 7. Environment Variables

`.env` loading behavior is already explained in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `4. Environment Variables`
```

Important: The Go code uses `os.Getenv`. `.env` is not auto-loaded. Export the file before `go run`.

### Current Task `.env` Delta

Add or update only these current-task variables:

```env
# Current task event ingestion mode
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

# Required prerequisite when events are enabled
RECOMMENDATION_MONGO_URI=mongodb://root:secret@localhost:27017
RECOMMENDATION_MONGO_DATABASE=recommendation_db

# Optional but useful side effect
RECOMMENDATION_REDIS_ADDR=localhost:6379

# Keep later-task workers off for current-task-only debugging
RECOMMENDATION_FEATURES_ENABLED=false
RECOMMENDATION_RANKING_ENABLED=false
RECOMMENDATION_PERSONALIZATION_ENABLED=false
RECOMMENDATION_AB_TESTING_ENABLED=false
```

Variable details:

| Variable | Purpose | Required? | Security / Config Note |
|---|---|---:|---|
| `RECOMMENDATION_EVENTS_ENABLED` | Starts the interaction consumer | Yes for current-task full mode | Keep `false` for minimal mode |
| `QUEUE_PROVIDER` | Queue backend selection | Yes | Must be `kafka` when events are active |
| `KAFKA_BROKERS` | Kafka bootstrap broker list | Yes | Use `localhost:9092` on host, `kafka:9092` inside Compose |
| `RECOMMENDATION_EVENTS_TOPIC` | Main topic consumed by service | Yes | Default is `recommendation.events` |
| `RECOMMENDATION_EVENTS_RETRY_TOPIC` | Future retry topic name | Optional in current code | Config exists but current code does not publish retries to it |
| `RECOMMENDATION_EVENTS_DLQ_TOPIC` | DLQ topic for bad/failed events | Yes | Treat DLQ messages as sensitive |
| `RECOMMENDATION_EVENTS_GROUP` | Kafka consumer group | Yes | Keep stable for normal processing; change only for replay |
| `RECOMMENDATION_EVENTS_MAX_MESSAGE_BYTES` | Max event payload size | Yes | Default `262144`; prevents huge payloads |
| `RECOMMENDATION_EVENTS_UNKNOWN_POLICY` | Unknown event action | Yes | Allowed values: `dlq`, `skip` |
| `EVENT_MAX_RETRY_ATTEMPTS` | Mongo retry attempts before DLQ | Yes | Default `4` |
| `EVENT_RETRY_BACKOFF_SECONDS` | Retry wait pattern | Yes | Comma-separated seconds, example `5,30,120` |
| `EVENT_SUPPORTED_VERSION` | Accepted event schema version | Yes | Current supported version is `1` |
| `RECOMMENDATION_KAFKA_MIN_BYTES` | Kafka reader min batch bytes | Optional | Keep `1` locally |
| `RECOMMENDATION_KAFKA_MAX_BYTES` | Kafka reader max batch bytes | Optional | Must be positive |

Common mistakes:

- `.env` updated but not sourced.
- `RECOMMENDATION_EVENTS_ENABLED=true` without MongoDB URI.
- `QUEUE_PROVIDER=rabbitmq` while current code only supports Kafka.
- Kafka running in Docker but `KAFKA_ADVERTISED_LISTENERS` points to an unreachable host.
- `EVENT_RETRY_BACKOFF_SECONDS` has negative values.
- Unknown extra fields in event JSON. The decoder uses strict JSON parsing and can DLQ the event.

---

## 8. Docker Setup

Generic Docker and Compose setup is already documented in:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `7. Docker and DevOps Setup`
```

`TASK_FILE_NAME` does not add:

- A new service Dockerfile
- A new root `docker-compose.yml`
- New Docker networks or volumes beyond previous MongoDB/Redis/Kafka setup

Current task Docker needs:

| Container | Required for Current-Task Full Mode? | Port |
|---|---:|---:|
| MongoDB | Yes | `27017` |
| Kafka | Yes | `9092` |
| Redis | Optional | `6379` |

Task-specific Docker action:

- Start the same local dependency stack from the previous guide.
- Create `recommendation.events`, `recommendation.events.dlq`, and optional future `recommendation.events.retry` topics.
- If the app runs inside Docker, change `localhost` env values to service names.

Example inside Docker network:

```env
RECOMMENDATION_MONGO_URI=mongodb://root:secret@mongo:27017
RECOMMENDATION_REDIS_ADDR=redis:6379
KAFKA_BROKERS=kafka:9092
```

---

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Read these first:

```text
TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
```

### Step 2: Go To Project Directory

```bash
cd backend/services/recommendation-service
```

### Step 3: Install Dependencies

No new Go package install is needed.

```bash
go mod download
go test ./...
```

### Step 4: Start Required Services

Use the previous Docker setup to start:

- MongoDB
- Kafka
- Redis, optional

Verify:

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" --eval 'db.runCommand({ ping: 1 })'
docker exec reco-kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list
redis-cli -h localhost -p 6379 ping
```

### Step 5: Add Current Task Environment Variables

Create/update:

```text
backend/services/recommendation-service/.env
```

Use the current-task `.env` delta from section `7. Environment Variables`.

Then export it:

```bash
set -a
source .env
set +a
```

### Step 6: Run Migrations

Run base storage migration if not already done, then current-task event indexes:

```bash
export RECOMMENDATION_MONGO_DATABASE=recommendation_db

mongosh "mongodb://root:secret@localhost:27017/admin" \
  --file migrations/001_create_recommendation_storage.up.js

mongosh "mongodb://root:secret@localhost:27017/admin" \
  --file migrations/002_add_event_ingestion_indexes.up.js
```

### Step 7: Create Kafka Topics

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
```

### Step 8: Start Backend Service

```bash
go run ./cmd/server
```

Expected logs should include:

```text
recommendation.kafka.consumer_started
recommendation.http.started
recommendation.grpc.started
```

### Step 9: Publish A Sample Event

In another terminal, publish one valid event:

```bash
docker exec -i reco-kafka /opt/kafka/bin/kafka-console-producer.sh \
  --bootstrap-server localhost:9092 \
  --topic recommendation.events <<'EOF'
{"event_id":"evt_view_001","event_type":"ProductViewed","version":1,"occurred_at":"2026-06-05T10:30:00Z","producer":"session-service","trace_id":"trace_demo_001","payload":{"user_id":"user_123","anonymous_id":"anon_456","session_id":"sess_789","product_id":"prod_123","variant_id":"var_1","category_id":"cat_shoes","seller_id":"seller_456","page":"product_detail"}}
EOF
```

### Step 10: Verify Event Was Stored

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" \
  --eval 'db.getSiblingDB("recommendation_db").user_interactions.find({event_id:"evt_view_001"}).pretty()'
```

Expected important fields:

```text
event_type: ProductViewed
normalized_event_type: product_view
weight: 1
dedupe_key: evt_view_001#product_view#prod_123#var_1
```

Check service metrics:

```bash
curl -s http://localhost:8088/metrics | grep recommendation_events_consumed_total
curl -s http://localhost:8088/metrics | grep queue_consumer_lag
```

Health checks:

```bash
curl http://localhost:8088/healthz
curl http://localhost:8088/internal/v1/recommendation/storage/status
```

Important: `/healthz` only confirms HTTP service is alive. It does not prove Kafka events are being consumed. Use MongoDB and `/metrics` for current-task verification.

---

## 10. Running the Project

Use previous run instructions for clone, Go version, base service startup, MongoDB, Redis, and generic Docker setup:

```md
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
Section: `8. Project Run Instructions`
```

Current-task full run:

```bash
cd backend/services/recommendation-service

set -a
source .env
set +a

go run ./cmd/server
```

Current-task successful run means:

- Service starts without config error.
- MongoDB status is ready.
- Kafka consumer starts for `recommendation.events`.
- A valid event is inserted into `recommendation_db.user_interactions`.
- Duplicate event does not create duplicate documents.
- Invalid event appears in `recommendation.events.dlq`.

Duplicate verification:

```bash
mongosh "mongodb://root:secret@localhost:27017/admin" \
  --eval 'db.getSiblingDB("recommendation_db").user_interactions.countDocuments({dedupe_key:"evt_view_001#product_view#prod_123#var_1"})'
```

Expected:

```text
1
```

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
| `event ingestion requires configured MongoDB storage` | Events enabled but Mongo URI missing or Mongo init failed | Set `RECOMMENDATION_MONGO_URI`, start MongoDB, verify `mongosh` | Keep events disabled until Mongo is ready |
| `KAFKA_BROKERS cannot be empty when event ingestion is enabled` | `RECOMMENDATION_EVENTS_ENABLED=true` but no broker list | Set `KAFKA_BROKERS=localhost:9092` | Use current-task `.env` delta |
| `QUEUE_PROVIDER "rabbitmq" is not supported` | RabbitMQ selected but current code supports Kafka only | Set `QUEUE_PROVIDER=kafka` | Do not follow RabbitMQ path for current implementation |
| Kafka consumer starts but no events consumed | Wrong topic, wrong advertised listener, or producer using different broker | Verify topics, broker address, and producer command | Keep `KAFKA_BROKERS` aligned with host vs Docker network |
| Kafka topic missing | Topic not created or auto-create disabled | Create `recommendation.events` and `recommendation.events.dlq` | Add topic creation to local bootstrap |
| Event goes to DLQ with `invalid json` | Bad JSON or multiple JSON objects in one message | Publish one valid JSON object per message | Validate payload before publishing |
| Event goes to DLQ with `unsupported event version` | `version` is not `1` | Send `version: 1` | Version producer contract carefully |
| Event goes to DLQ with `unsupported event` | Event type not in supported list | Use `ProductViewed`, `CartItemAdded`, `WishlistItemAdded`, `WishlistItemRemoved`, `OrderPaid`, or `PurchaseCompleted` | Keep producer event names exact |
| Event goes to DLQ with `invalid payload` | Missing required identity/product fields or unknown payload fields | Add `user_id` or `anonymous_id`, add `product_id`, remove unexpected fields | Share strict schema with producer teams |
| `message_too_large` DLQ reason | Event payload exceeds `RECOMMENDATION_EVENTS_MAX_MESSAGE_BYTES` | Reduce payload or increase limit carefully | Never include huge product/user blobs |
| Duplicate event appears in logs | Same `dedupe_key` received again | This is expected; service treats duplicate as success | Keep unique index migration applied |
| DLQ publish fails | DLQ topic missing or broker unavailable | Create DLQ topic, fix Kafka connectivity | Verify DLQ topic before starting service |
| Redis invalidation warning | Redis down or wrong `RECOMMENDATION_REDIS_ADDR` | Start Redis or clear Redis env for current-task raw ingest | Treat Redis as optional side effect |
| Feature builder starts during current-task debugging | Later-task env/defaults enabled | Set `RECOMMENDATION_FEATURES_ENABLED=false` for current-task-only run | Keep task-specific env small |

---

## 12. Security & Best Practices

Current-task-specific best practices:

- Do not publish email, phone, full name, address, auth token, payment card data, or raw IP in recommendation events.
- Treat DLQ as sensitive because it may contain failed original payloads.
- Use Kafka auth/TLS in shared, staging, and production environments.
- Keep Kafka, MongoDB, Redis, HTTP internal endpoints, and gRPC endpoints on private networks.
- Keep `RECOMMENDATION_EVENTS_MAX_MESSAGE_BYTES` bounded so one bad payload cannot overload the service.
- Keep `EVENT_SUPPORTED_VERSION=1` until producer and consumer schema migration is planned.
- Use stable consumer group `recommendation-service-v1` for normal processing.
- Use a temporary different consumer group only for replay/debug, then switch back.
- Apply MongoDB unique `dedupe_key` index before enabling production consumers.
- Keep Redis invalidation best-effort; do not let cache failure drop valid user events.
- Monitor `recommendation_event_dlq_total`, `recommendation_mongo_insert_errors_total`, and `queue_consumer_lag`.

---

## 13. Missing or Misconfigured Things

| Finding | Impact | Suggested Fix |
|---|---|---|
| `TASK_FILE_NAME` mentions RabbitMQ, but current backend supports only Kafka. | Beginners may configure RabbitMQ and hit config failure. | Keep docs and `.env` on `QUEUE_PROVIDER=kafka` until RabbitMQ consumer code is implemented. |
| `RECOMMENDATION_EVENTS_RETRY_TOPIC` exists but current code does in-process retries, then DLQ. | Retry topic may be created but unused. | Document it as future-ready or implement retry-topic publishing later. |
| No service Dockerfile detected. | Production packaging is incomplete. | Add Dockerfile and container health checks before deployment. |
| No root `docker-compose.yml` detected in repo scan. | Beginners rely on manual Docker commands. | Add local dependency Compose file or bootstrap script. |
| Existing local `.env` contains local sample credentials. | Unsafe if copied to shared environments. | Keep real secrets out of git and use `.env.example` with dummy values. |
| `.env` enables later-task features/ranking/personalization. | Current-task-only debugging may require extra migrations/replica-set behavior. | Disable later-task workers while verifying only raw event ingestion. |
| No dedicated ingestion health endpoint exists. | `/healthz` can pass even if no event has been consumed. | Verify with MongoDB document count and `/metrics`. |
| Kafka topics are not created by migrations. | Service can start but fail to consume/publish depending on broker settings. | Create topics during local bootstrap/deployment. |

---

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `1. Project Tech Stack Analysis` | Base Go, HTTP, gRPC, Prometheus, MongoDB, Redis, Kafka stack already explained |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `2. Go Dependency System` | Go modules, `go.mod`, `go.sum`, common Go commands already documented |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `5. External Services` -> `Kafka` | Kafka install/Docker basics already documented |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `6. Ports and Networking` | Generic port conflict and Docker network rules already documented |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `7. Docker and DevOps Setup` | Dependency-only Docker Compose already documented |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `8. Project Run Instructions` | Clone, install, run, health check flow already documented |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `9. Common Errors and Fixes` | Generic Go/Mongo/Redis/Kafka/Docker errors already documented |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | MongoDB database, base migration, collection setup already documented |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `6. Redis / Queue / External Services` | Redis cache key and TTL setup already documented |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `7. Environment Variables` | MongoDB/Redis env variables already documented |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `11. Common Errors & Fixes` | Storage/cache troubleshooting already documented |

---

## 15. Final Checklist

- [ ] Previous dependency documentation checked.
- [ ] `task1_Dependency.md` generic setup followed.
- [ ] `task2_Dependency.md` MongoDB/Redis setup followed.
- [ ] MongoDB is running and reachable.
- [ ] Kafka is running and reachable.
- [ ] Redis is running if cache invalidation is enabled.
- [ ] `001_create_recommendation_storage.up.js` migration executed.
- [ ] `002_add_event_ingestion_indexes.up.js` migration executed.
- [ ] `recommendation.events` topic created.
- [ ] `recommendation.events.dlq` topic created.
- [ ] Current-task event env variables added.
- [ ] Later-task workers disabled for current-task-only debugging if needed.
- [ ] `.env` exported before running service.
- [ ] Backend service started.
- [ ] Logs show `recommendation.kafka.consumer_started`.
- [ ] Sample `ProductViewed` event published.
- [ ] MongoDB `user_interactions` contains the normalized event.
- [ ] Duplicate event count remains `1`.
- [ ] `/metrics` shows event ingestion metrics.
- [ ] DLQ topic checked for invalid-event testing.
- [ ] No duplicate setup documentation added.
