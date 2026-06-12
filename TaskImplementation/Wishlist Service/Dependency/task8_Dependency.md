# Project Dependency & Setup Guide

## Variable Values

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Wishlist Service` |
| `TASK_FILE_NAME` | `task8.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task8_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

Note: Ye guide `INPUT_FILE_PATH` ko analyze karke banaya gaya hai. Original `TASK_FILE_NAME` modify nahi kiya gaya. Is file ka focus sirf Task 8 analytics events ke dependencies, environment, MongoDB outbox, Kafka publishing, verification, and setup gotchas par hai.

---

## 1. Project Overview

`TASK_FILE_NAME` ka scope:

```text
Wishlist add/remove success
-> wishlist analytics event create
-> wishlist_events outbox collection me pending save
-> background worker event publish
-> recommendation.events topic
-> Recommendation Service personalization improve karega
```

Simple Hinglish me: Jab buyer product wishlist me add ya remove karta hai, `SERVICE_NAME` ek analytics event publish karta hai. Recommendation Service is event ko consume karke user interest samajh sakta hai.

### Runtime code paths checked

| Runtime file | Task 8 relevance |
|---|---|
| `backend/services/wishlist-service/internal/usecase/wishlist_service.go` | `AddItem` and `RemoveItem` success ke baad analytics event enqueue hota hai. |
| `backend/services/wishlist-service/internal/usecase/wishlist_event_builder.go` | `wishlist_item_added` / `wishlist_item_removed` event payload build karta hai. |
| `backend/services/wishlist-service/internal/domain/wishlist_event.go` | Event type, action, status, payload validation. |
| `backend/services/wishlist-service/internal/repository/mongo_wishlist_event_repository.go` | `wishlist_events` outbox collection, claim, retry, publish status, indexes. |
| `backend/services/wishlist-service/internal/repository/wishlist_event_document.go` | Mongo document mapping for outbox events. |
| `backend/services/wishlist-service/internal/events/outbox_worker.go` | Pending events poll karke publish/retry/failed status manage karta hai. |
| `backend/services/wishlist-service/internal/events/kafka_publisher.go` | Kafka publisher for `recommendation.events`. |
| `backend/services/wishlist-service/internal/config/config.go` | Task 8 env vars and defaults. |
| `backend/services/wishlist-service/cmd/server/main.go` | Outbox repository, Kafka publisher, and worker wiring. |
| `backend/services/wishlist-service/migrations/mongo/004_create_wishlist_events_collection.up.js` | `wishlist_events` validator and indexes. |
| `backend/services/wishlist-service/.env` | Current local env exists, but Task 8 analytics vars are not explicitly listed. Add them manually for clarity. |

### Important Reuse Note

Base setup already documented hai. Is file me Go install, MongoDB install, Docker install, generic Kafka explanation, Product Service validation setup, Cart Service setup, and generic troubleshooting repeat nahi kiya gaya.

Read first:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task6_Dependency.md
TaskImplementation/{SERVICE_NAME}/task7_Dependency.md
```

---

## 2. Tech Stack

Most base technology explanation already exists in previous dependency files. Task 8 ka new setup focus hai analytics outbox plus Kafka publishing to Recommendation Service.

| Technology / Service | Task 8 Status | Required? | Setup Documentation |
|---|---|---:|---|
| Go | Reused runtime | Yes | `task1_Dependency.md` -> `4. Dependency Management` |
| Go modules | Reused dependency system | Yes | `task1_Dependency.md` -> `4. Dependency Management` |
| `net/http` | Reused API server | Yes | `task1_Dependency.md` -> `2. Tech Stack` |
| MongoDB | Reused database plus new outbox collection | Yes | Base setup in `task1_Dependency.md`; Task 8 delta below |
| MongoDB Go Driver v2 | Reused dependency | Yes | Already present in `go.mod` |
| Kafka-compatible broker | Required for full Task 8 publishing mode | Yes for end-to-end analytics | Base setup in `task1_Dependency.md`; topic delta below |
| `github.com/segmentio/kafka-go` | Reused Go Kafka client | Yes for Kafka mode | Already present in `go.mod`; do not run `go get` again |
| Recommendation Service consumer | Downstream dependency | Required for real personalization | Not implemented in this snapshot; it must consume `recommendation.events` |
| Product Service HTTP API | Reused from Task 4 | Required for add-item API | `task4_Dependency.md` -> `Product Service` |
| Cart Service HTTP API | Reused service startup dependency | Required by current constructor/config | `task5_Dependency.md` -> `Cart Service` |
| Docker / Redpanda | Reused local broker option | Recommended | `task1_Dependency.md` and `task7_Dependency.md` |
| Redis | Not introduced | No | No setup needed |
| RabbitMQ | Not supported by current `SERVICE_NAME` runtime | No | Do not set Task 8 publisher to `rabbitmq` |

### Task 8 new technical concept: analytics outbox

Outbox ka meaning: API request ke andar event directly Kafka par bhejne ke bajay pehle MongoDB collection me save hota hai. Background worker baad me publish karta hai.

Why useful:

| Problem | Outbox benefit |
|---|---|
| Kafka temporarily down | Event `pending` rahega, worker retry karega. |
| API latency | Buyer ko add/remove response Kafka wait ke bina mil sakta hai. |
| Debugging | `wishlist_events` collection me event status visible rahega. |
| Duplicate publish | Same `event_id` se Recommendation Service idempotency kar sakta hai. |

---

## 3. Required Software

Base required software setup reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`3. Required Software`
`4. Dependency Management`
`5. Database Setup`
`8. Docker Setup`
```

Task 8-specific readiness:

| Requirement | Required? | Why |
|---|---:|---|
| MongoDB running on configured URI | Yes | `wishlist_events` outbox collection stores analytics events. |
| `mongosh` | Recommended | Run/verify migration `004_create_wishlist_events_collection.up.js`. |
| Kafka-compatible broker on `localhost:9092` or configured address | Yes for full analytics publish mode | Worker publishes events to Kafka. |
| `recommendation.events` topic | Yes for full analytics publish mode | Destination topic for Recommendation Service. |
| Redpanda `rpk` or Kafka CLI | Recommended | Create and consume topic locally. |
| Product Service reachable | Yes for add-item event test | `POST /api/v1/wishlist/items` validates product before wishlist mutation. |
| curl/Postman | Recommended | Trigger add/remove API and health checks. |

Beginner note: Agar aap sirf wishlist API run karna chahte ho without analytics, Task 8 env vars disable mode me set karo. Agar outbox verify karna hai but Kafka nahi chalana, outbox-only mode use karo.

---

## 4. Dependency Management

No new Go module dependency is introduced by `TASK_FILE_NAME`.

Existing module:

```text
backend/services/wishlist-service/go.mod
```

Existing direct dependencies:

| Dependency | Version | Task 8 usage |
|---|---:|---|
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | `wishlist_events` collection CRUD, validator, indexes, retry state. |
| `github.com/segmentio/kafka-go` | `v0.4.49` | Publish analytics envelope to `recommendation.events`. |

Reuse Go setup commands from:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Dependency Management`
```

Task 8 beginner warnings:

- Do not run `go get github.com/segmentio/kafka-go`; dependency already exists.
- Do not add RabbitMQ packages for current Task 8 runtime. `WISHLIST_EVENT_PUBLISHER` accepts only `disabled` or `kafka`.
- Do not manually edit `go.sum`.
- If `go mod download` fails, follow Go module troubleshooting from `task1_Dependency.md`.

Useful Task 8 verification tests:

```bash
cd backend/services/wishlist-service
go test ./internal/usecase ./internal/events ./internal/repository ./internal/config
```

---

## 5. Database Setup

MongoDB installation, Docker setup, credentials, and main `wishlists` collection setup are reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`5. Database Setup`
`8. Docker Setup`
```

Main wishlist collection design is reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Section:
`5. Database Setup`
```

### Task 8 Database Delta

Task 8 introduces one service-owned outbox collection:

| Item | Value |
|---|---|
| Database | `wishlist_db` |
| Collection | `wishlist_events` |
| Env var | `WISHLIST_EVENT_OUTBOX_COLLECTION` |
| Purpose | Pending/published/failed wishlist analytics events track karna |
| Owner | `SERVICE_NAME` only |
| Mandatory? | Yes when `WISHLIST_ANALYTICS_EVENTS_ENABLED=true` |

### Event statuses

| Status | Meaning |
|---|---|
| `pending` | Event saved hai and publish ke liye ready hai. |
| `publishing` | Worker ne event claim kiya hai. |
| `published` | Kafka ack ke baad mark hua. |
| `failed` | Max attempts ke baad publish nahi ho paya. |

### Task 8 Migration

Run this migration when analytics outbox is enabled:

```bash
cd backend/services/wishlist-service
mongosh "mongodb://localhost:27017" migrations/mongo/004_create_wishlist_events_collection.up.js
```

This migration creates or updates:

| Database object | Name | Why needed |
|---|---|---|
| Collection | `wishlist_events` | Analytics outbox events store karne ke liye. |
| Validator | `WishlistAnalyticsEvent` JSON schema | Invalid event documents block karne ke liye. |
| Index | `idx_wishlist_events_pending_retry` | Worker pending events fast claim karega. |
| Index | `idx_wishlist_events_type_time` | Debug/audit queries ke liye. |
| TTL index | `ttl_wishlist_events_published_at` | Published events 30 days baad auto cleanup. |

Note: Service startup also calls `EnsureCollection` when analytics is enabled, but migration explicitly run karna better hai because local setup visible and repeatable ban jata hai.

### Verify Collection Exists

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.getCollectionNames()"
```

Expected: output should include `wishlist_events`.

### Verify Indexes

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.wishlist_events.getIndexes()"
```

Expected important indexes:

```text
idx_wishlist_events_pending_retry
idx_wishlist_events_type_time
ttl_wishlist_events_published_at
```

### Verify Pending Events

After triggering wishlist add/remove:

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval 'db.wishlist_events.find({}, { event_type: 1, status: 1, topic: 1, payload: 1, attempts: 1, last_error: 1 }).sort({ created_at: -1 }).limit(5).pretty()'
```

Expected:

```text
event_type: wishlist_item_added OR wishlist_item_removed
topic: recommendation.events
status: pending / publishing / published / failed
```

### Credentials placement

Mongo credentials remain reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`5. Database Setup`
`7. Environment Variables`
```

Task 8-specific note: If `WISHLIST_MONGO_URI` contains username/password, keep it in local uncommitted `.env`, Docker/Kubernetes Secret, or secret manager. Do not commit real credentials.

---

## 6. Redis / Queue / External Services

### Kafka Analytics Events

Task 8 full mode needs Kafka-compatible broker.

| Field | Value |
|---|---|
| Publisher env var | `WISHLIST_EVENT_PUBLISHER` |
| Enabled value | `kafka` |
| Disabled value | `disabled` |
| Broker env var | `WISHLIST_KAFKA_BROKERS` |
| Default broker | `localhost:9092` |
| Topic env var | `WISHLIST_EVENT_TOPIC` |
| Default topic | `recommendation.events` |
| Event types | `wishlist_item_added`, `wishlist_item_removed` |
| Message key | `user_id` |
| Producer header | `producer=wishlist-service` |

Kafka basic setup is already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`6. Redis / Queue / External Services`
`8. Docker Setup`
```

Topic creation style is already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task7_Dependency.md`

Section:
`8. Docker Setup` -> `Topic Creation With Redpanda Container`
```

Task 8-specific topic:

```bash
docker exec wishlist-redpanda rpk topic create recommendation.events
docker exec wishlist-redpanda rpk topic list
```

Consume locally:

```bash
docker exec wishlist-redpanda rpk topic consume recommendation.events -n 1
```

If your broker auto-creates topics, explicit creation is still recommended because typo issues become visible early.

### Recommendation Service Consumer

`SERVICE_NAME` only publishes the event. Recommendation Service must consume it.

Expected event envelope:

```json
{
  "event_id": "evt_wish_abc123",
  "event_type": "wishlist_item_added",
  "version": 1,
  "occurred_at": "2026-05-26T12:10:00Z",
  "producer": "wishlist-service",
  "trace_id": "req_abc123",
  "payload": {
    "user_id": "user_123",
    "product_id": "prod_123",
    "variant_id": "var_1",
    "action": "add",
    "source": "wishlist",
    "availability": "in_stock",
    "last_known_price": {
      "amount": 299900,
      "currency": "INR"
    }
  }
}
```

Recommendation consumer should be idempotent. Use `event_id` or `source_event_id` as unique key so retry duplicate events do not create duplicate interactions.

### Services not introduced by Task 8

| Service | Task 8 status | Note |
|---|---|---|
| Redis | Not used | No Redis env var or client exists in current runtime. |
| RabbitMQ | Not supported by current Task 8 publisher | `WISHLIST_EVENT_PUBLISHER=rabbitmq` will fail validation. |
| NATS | Not used | No setup needed. |
| MinIO / S3 | Not used | No setup needed. |
| Elasticsearch / Typesense | Not used by this service | Recommendation/search concerns are separate. |
| SMTP / Twilio / Stripe / OAuth | Not used by Task 8 | Notification/payment/auth services own these. |

---

## 7. Environment Variables

Base `.env` and common variables are already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Environment Variables`
```

How env is loaded is already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task5_Dependency.md`

Section:
`7. Environment Variables` -> `How the Project Loads Env`
```

Current Go code reads OS environment variables. It does not automatically load `.env`.

Load before start:

```bash
cd backend/services/wishlist-service
set -a
source .env
set +a
```

### Task 8 Env Delta Only

Current `backend/services/wishlist-service/.env` does not explicitly list the Task 8 analytics vars. Add one of the modes below.

#### Mode A: simple HTTP run without analytics

Use this when a beginner only wants APIs/health checks and no Kafka/outbox testing:

```env
WISHLIST_ANALYTICS_EVENTS_ENABLED=false
WISHLIST_EVENT_PUBLISHER=disabled
```

Result: `wishlist_events` repository is not wired, and add/remove does not enqueue analytics events.

#### Mode B: outbox-only local verification

Use this when Mongo outbox should be tested but Kafka is not running:

```env
WISHLIST_ANALYTICS_EVENTS_ENABLED=true
WISHLIST_EVENT_OUTBOX_COLLECTION=wishlist_events
WISHLIST_EVENT_TOPIC=recommendation.events
WISHLIST_EVENT_PUBLISHER=disabled
WISHLIST_EVENT_POLL_INTERVAL=5s
WISHLIST_EVENT_BATCH_SIZE=50
WISHLIST_EVENT_MAX_ATTEMPTS=5
```

Result: add/remove events are stored in `wishlist_events`, but no worker publishes to Kafka. Events remain `pending`.

#### Mode C: full Task 8 analytics publish

Use this when Kafka/Redpanda is running and `recommendation.events` exists:

```env
WISHLIST_ANALYTICS_EVENTS_ENABLED=true
WISHLIST_EVENT_OUTBOX_COLLECTION=wishlist_events
WISHLIST_EVENT_TOPIC=recommendation.events
WISHLIST_EVENT_PUBLISHER=kafka
WISHLIST_KAFKA_BROKERS=localhost:9092
WISHLIST_EVENT_POLL_INTERVAL=5s
WISHLIST_EVENT_BATCH_SIZE=50
WISHLIST_EVENT_MAX_ATTEMPTS=5
```

Keep this independent product-event switch as needed:

```env
WISHLIST_EVENTS_BACKEND=disabled
```

Beginner meaning: Task 8 analytics publishing can run even if Task 6/Task 7 product event consumer is disabled.

### Variable Explanation

| Variable | Required? | Example | Task 8 purpose | Security/config note |
|---|---:|---|---|---|
| `WISHLIST_ANALYTICS_EVENTS_ENABLED` | Required for clear local behavior | `true` | Creates/wires outbox repository and allows add/remove event enqueue. | Set `false` for simple local mode. |
| `WISHLIST_EVENT_OUTBOX_COLLECTION` | Required when analytics enabled | `wishlist_events` | Mongo collection for analytics outbox. | Keep service-owned; do not share with Recommendation DB. |
| `WISHLIST_EVENT_TOPIC` | Required when analytics enabled | `recommendation.events` | Destination topic stored in each event and used by publisher. | Keep aligned with Recommendation Service consumer. |
| `WISHLIST_EVENT_PUBLISHER` | Required for clear local behavior | `kafka` | Starts Kafka publisher/worker when analytics enabled. | Allowed values: `disabled`, `kafka`. |
| `WISHLIST_KAFKA_BROKERS` | Required when publisher is Kafka | `localhost:9092` | Kafka broker CSV. | Use internal broker DNS in Docker/Kubernetes. |
| `WISHLIST_EVENT_POLL_INTERVAL` | Optional | `5s` | Worker polling interval. | Must be positive. Too low can make noisy polling. |
| `WISHLIST_EVENT_BATCH_SIZE` | Optional | `50` | Max events claimed per worker loop. | Must be positive. Increase carefully. |
| `WISHLIST_EVENT_MAX_ATTEMPTS` | Optional | `5` | Publish retries before status `failed`. | Failed docs need manual replay/ops review. |

### Common Task 8 Env Mistakes

| Mistake | Result | Fix |
|---|---|---|
| `.env` changed but not sourced | Service still uses old/default values. | Run `set -a; source .env; set +a`. |
| `WISHLIST_EVENT_PUBLISHER=rabbitmq` | Startup validation fails. | Use `disabled` or `kafka`. |
| Kafka enabled but wrong broker host | Events move to retry/failed with connection errors. | Host run: `localhost:9092`; Docker-to-Docker: Compose service name like `wishlist-redpanda:9092`. |
| `WISHLIST_EVENT_TOPIC` typo | Events publish to wrong topic or consumer sees nothing. | Keep `recommendation.events` unless platform docs change. |
| `WISHLIST_ANALYTICS_EVENTS_ENABLED=false` during Task 8 test | No event is enqueued. | Set it to `true`. |
| `WISHLIST_EVENT_PUBLISHER=disabled` during Kafka verification | Events remain `pending`. | Set publisher to `kafka` for full publish mode. |

---

## 8. Docker Setup

No new Dockerfile, Docker Compose service, volume, or network is introduced by `TASK_FILE_NAME`.

Reuse base Docker setup:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`8. Docker Setup`
```

Reuse Kafka/Redpanda topic creation style:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task7_Dependency.md`

Section:
`8. Docker Setup`
```

Task 8 Docker delta:

| Docker item | Status | Reason |
|---|---|---|
| MongoDB container | Reused | Stores `wishlists` and new `wishlist_events`. |
| MongoDB volume | Reused | Keeps outbox events across restarts. |
| Redpanda/Kafka container | Reused | Publishes `recommendation.events`. |
| New topic | New | Create `recommendation.events`. |
| New container | Not needed | Task 8 uses existing wishlist service plus existing broker. |
| New network | Not needed | Reuse existing local/Compose network. |

### Host vs Docker URL examples

| Running location | Mongo URI | Kafka brokers |
|---|---|---|
| Service runs on host | `mongodb://localhost:27017` | `localhost:9092` |
| Service runs inside Compose | `mongodb://wishlist-mongo:27017` | `wishlist-redpanda:9092` |

### Ports & Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Wishlist backend API | 8084 | HTTP APIs and health checks | Reused |
| MongoDB | 27017 | `wishlists` and `wishlist_events` storage | Reused |
| Kafka / Redpanda | 9092 | `recommendation.events` publishing | Reused |
| Product Service | 8082 | Product validation for add-item flow | Reused |
| Cart Service | 8083 | Move-to-cart runtime dependency | Reused |
| Recommendation Service | TBD | Consumes `recommendation.events` | Downstream, not in this service snapshot |

Port conflict fixes are already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`11. Ports & Networking`
```

Task 8-specific networking note: Kafka broker address is the most common confusion. `localhost:9092` works from host process. From another Docker container, `localhost` points to that container itself, so use the broker service name instead.

---

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Follow these first:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task6_Dependency.md
TaskImplementation/{SERVICE_NAME}/task7_Dependency.md
```

Why:

| Previous file | Reused setup |
|---|---|
| `task1_Dependency.md` | Go, MongoDB, Kafka/Redpanda, Docker, base `.env`, run commands, common errors. |
| `task3_Dependency.md` | `wishlists` collection migration and verification. |
| `task4_Dependency.md` | Product Service validation required for add item. |
| `task6_Dependency.md` | Kafka event consumer/publisher basics and topic mindset. |
| `task7_Dependency.md` | Redpanda topic creation and Kafka env troubleshooting. |

### Step 2: Go to Project Directory

```bash
cd backend/services/wishlist-service
```

### Step 3: Install Only New Dependencies

No new dependency install is needed.

Optional verification:

```bash
go mod download
go test ./internal/usecase ./internal/events ./internal/repository ./internal/config
```

### Step 4: Setup Only New Databases/Services

Run base Mongo setup first from previous docs. Then run Task 8 migration:

```bash
mongosh "mongodb://localhost:27017" migrations/mongo/004_create_wishlist_events_collection.up.js
```

For full Kafka mode, create Task 8 topic:

```bash
docker exec wishlist-redpanda rpk topic create recommendation.events
docker exec wishlist-redpanda rpk topic list
```

### Step 5: Add Only New or Changed Environment Variables

For outbox-only mode:

```env
WISHLIST_ANALYTICS_EVENTS_ENABLED=true
WISHLIST_EVENT_OUTBOX_COLLECTION=wishlist_events
WISHLIST_EVENT_TOPIC=recommendation.events
WISHLIST_EVENT_PUBLISHER=disabled
WISHLIST_EVENT_POLL_INTERVAL=5s
WISHLIST_EVENT_BATCH_SIZE=50
WISHLIST_EVENT_MAX_ATTEMPTS=5
```

For full Kafka mode:

```env
WISHLIST_ANALYTICS_EVENTS_ENABLED=true
WISHLIST_EVENT_OUTBOX_COLLECTION=wishlist_events
WISHLIST_EVENT_TOPIC=recommendation.events
WISHLIST_EVENT_PUBLISHER=kafka
WISHLIST_KAFKA_BROKERS=localhost:9092
WISHLIST_EVENT_POLL_INTERVAL=5s
WISHLIST_EVENT_BATCH_SIZE=50
WISHLIST_EVENT_MAX_ATTEMPTS=5
```

Then load env:

```bash
set -a
source .env
set +a
```

### Step 6: Run Migrations If Needed

Recommended full local migration sequence:

```bash
mongosh "mongodb://localhost:27017" migrations/mongo/001_create_wishlists_collection.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/002_add_wishlist_item_variant_index.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/003_add_price_drop_candidate_index.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/004_create_wishlist_events_collection.up.js
```

### Step 7: Start Backend Service

```bash
go run ./cmd/server
```

Expected logs should include:

```text
analytics_events_enabled=true
analytics_event_publisher=disabled OR kafka
wishlist service listening
```

If full Kafka mode is enabled, expected worker log:

```text
wishlist analytics outbox worker started
```

### Step 8: Verify Functionality Related to `TASK_FILE_NAME`

Health:

```bash
curl -i http://localhost:8084/healthz
curl -i http://localhost:8084/readyz
```

Trigger add-item event:

```bash
curl -i -X POST http://localhost:8084/api/v1/wishlist/items \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_123" \
  -H "X-User-Roles: buyer" \
  -H "X-Request-ID: req_task8_add_001" \
  -d '{"product_id":"prod_123","variant_id":"var_1"}'
```

Trigger remove-item event:

```bash
curl -i -X DELETE http://localhost:8084/api/v1/wishlist/items/prod_123 \
  -H "X-User-ID: user_123" \
  -H "X-User-Roles: buyer" \
  -H "X-Request-ID: req_task8_remove_001"
```

Important: Add-item validation needs Product Service to return a valid published product. If Product Service is not running, add request may fail before event creation.

Verify outbox:

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval 'db.wishlist_events.find({}, { event_type: 1, status: 1, topic: 1, "payload.user_id": 1, "payload.product_id": 1, attempts: 1, last_error: 1 }).sort({ created_at: -1 }).limit(5).pretty()'
```

Full Kafka mode verification:

```bash
docker exec wishlist-redpanda rpk topic consume recommendation.events -n 1
```

Expected JSON should include:

```text
event_type: wishlist_item_added OR wishlist_item_removed
producer: wishlist-service
payload.source: wishlist
```

---

## 10. Running the Project

### Minimum local run without Task 8 analytics

Use when onboarding and avoiding Kafka:

```bash
cd backend/services/wishlist-service
set -a
source .env
set +a
export WISHLIST_ANALYTICS_EVENTS_ENABLED=false
export WISHLIST_EVENT_PUBLISHER=disabled
go run ./cmd/server
```

### Outbox-only Task 8 run

Use when verifying Mongo outbox but not Kafka:

```bash
cd backend/services/wishlist-service
set -a
source .env
set +a
export WISHLIST_ANALYTICS_EVENTS_ENABLED=true
export WISHLIST_EVENT_OUTBOX_COLLECTION=wishlist_events
export WISHLIST_EVENT_TOPIC=recommendation.events
export WISHLIST_EVENT_PUBLISHER=disabled
go run ./cmd/server
```

Expected: events stay `pending` in MongoDB.

### Full Task 8 Kafka run

Use when Redpanda/Kafka is running:

```bash
cd backend/services/wishlist-service
set -a
source .env
set +a
export WISHLIST_ANALYTICS_EVENTS_ENABLED=true
export WISHLIST_EVENT_OUTBOX_COLLECTION=wishlist_events
export WISHLIST_EVENT_TOPIC=recommendation.events
export WISHLIST_EVENT_PUBLISHER=kafka
export WISHLIST_KAFKA_BROKERS=localhost:9092
go run ./cmd/server
```

Expected: worker publishes pending events and marks them `published`.

---

## 11. Common Errors & Fixes

Generic setup errors are already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`12. Common Errors & Fixes`
```

Task 8-specific errors:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| No document appears in `wishlist_events` | `WISHLIST_ANALYTICS_EVENTS_ENABLED=false` or event repository not wired. | Set analytics enabled true and restart service. | Use explicit Task 8 env delta. |
| Events remain `pending` forever | `WISHLIST_EVENT_PUBLISHER=disabled`, or worker not running. | Use full Kafka mode with `WISHLIST_EVENT_PUBLISHER=kafka`. | Choose correct run mode before testing. |
| Events move to `failed` | Kafka publish failed until max attempts. | Check `last_error`, broker host, topic, and `WISHLIST_KAFKA_BROKERS`. | Verify topic and broker before triggering events. |
| Startup fails with `WISHLIST_EVENT_PUBLISHER must be disabled or kafka` | Unsupported publisher value. | Use `disabled` or `kafka`. | Do not copy RabbitMQ examples into current runtime. |
| Consumer sees no message | Wrong topic or no pending events. | Confirm `WISHLIST_EVENT_TOPIC=recommendation.events` and check Mongo outbox status. | Create topic explicitly and verify with `rpk topic list`. |
| Add-item does not create event | Product validation failed before wishlist mutation. | Start Product Service or use valid product fixture. | Check API response and service logs first. |
| Duplicate add does not create second event | Duplicate add is blocked/no mutation. | This is expected. | Test with a new product id. |
| Remove absent item does not create event | Item was not present before delete. | This is expected. | Add item first, then remove it. |
| Published event appears twice in Recommendation Service | Worker may retry after publish but before mark-published. | Recommendation consumer must upsert by `event_id`. | Add unique index on source event id in Recommendation DB. |
| `publishing` events never retry after crash | Current code has no lease expiry for `publishing` state. | Manually inspect and reset stuck events to `pending` after confirming worker crash. | Add `locked_until` / lease recovery in future hardening. |

Manual reset example for stuck local dev events:

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval 'db.wishlist_events.updateMany({ status: "publishing" }, { $set: { status: "pending", next_retry_at: new Date(), updated_at: new Date() } })'
```

Use manual reset carefully in production; first confirm no active worker is still publishing those events.

---

## 12. Security & Best Practices

### Security and Configuration Audit

| Area | Current observation | Risk | Suggested fix |
|---|---|---|---|
| Event payload PII | Payload uses user id, product id, variant id, availability, price. | Low, but user id is still sensitive internal data. | Do not add email, phone, address, payment, or name to event payload. |
| Secrets | No hardcoded broker password found in Task 8 code. | Real broker credentials may be added later. | Store credentials in `.env` locally and Secrets in deployed env. |
| `.env` clarity | Local `.env` does not explicitly list Task 8 analytics vars. | Beginners may unknowingly rely on defaults. | Add explicit Task 8 vars for chosen mode. |
| Publisher values | Current code supports only `disabled` and `kafka`. | RabbitMQ copied from platform docs will fail. | Keep docs and env aligned to current runtime. |
| Recommendation consumer | Not present in this service snapshot. | Events publish but personalization does not update. | Build/enable consumer in Recommendation Service and monitor lag. |
| Idempotency | Event id exists, outbox duplicate insert is ignored. | Consumer duplicates still possible after retry. | Recommendation Service must dedupe by `event_id`. |
| Analytics DLQ | No separate Task 8 Kafka DLQ topic exists. Failed events stay in `wishlist_events`. | Ops must inspect failed collection. | Add replay tooling or DLQ export later. |
| Publishing lock | `publishing` status has no `locked_until` lease. | Worker crash can leave stuck events. | Add lease id + lock expiry in future production hardening. |
| Mongo migration DB | Migration hardcodes `wishlist_db`. | Env DB mismatch can create collection in different DB than runtime uses. | Keep local `WISHLIST_MONGO_DATABASE=wishlist_db`, or parameterize migrations later. |
| Logs | Logs include internal user/product ids and event ids. | Avoid leaking PII in shared logs. | Keep only internal ids; do not log full customer data. |

### Task 8 Best Practices

- Keep `WISHLIST_EVENT_TOPIC=recommendation.events` aligned between producer and consumer.
- Keep analytics event schema versioned. Current version is `1`.
- Treat Kafka delivery as at-least-once. Consumer idempotency is mandatory.
- Use outbox-only mode for beginner debugging before enabling Kafka.
- Monitor `status=failed` count in `wishlist_events`.
- Alert if pending events keep growing for a long time.
- Do not fail buyer wishlist add/remove only because Recommendation Service is down.
- Keep `last_known_price.currency` uppercase ISO-style, for example `INR` or `USD`.
- Use internal Docker/Kubernetes service DNS for broker URL in containerized deployments.

---

## 13. Missing or Misconfigured Things

| Item | Current state | Setup impact | Recommendation |
|---|---|---|---|
| Explicit Task 8 env vars in local `.env` | Not present in checked `.env` | Defaults can surprise beginners. | Add one of the documented env modes. |
| Recommendation Service consumer | Not found in this repo snapshot | Published messages may have no consumer. | Implement/enable consumer for `recommendation.events`. |
| Analytics Kafka DLQ topic | Not implemented for Task 8 | Failed events remain in Mongo outbox. | Add replay/admin tooling or DLQ exporter if needed. |
| `publishing` lease recovery | Not implemented | Worker crash can leave stuck events. | Add `locked_until`, `locked_by`, and recovery query later. |
| Docker Compose file for full local stack | Not found in current repo file list | Beginners may need to use documented `docker run` examples. | Add Compose stack in platform foundation task when available. |
| Migration DB parameterization | Migration uses `wishlist_db` constant | Custom DB env can mismatch migration DB. | Keep default DB locally or parameterize migrations. |

---

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `task1_Dependency.md` | `3. Required Software` | Git, Go, MongoDB, Docker, Kafka-compatible broker basics already explained. |
| `task1_Dependency.md` | `4. Dependency Management` | Go modules, `go mod download`, `go test`, common Go failures already documented. |
| `task1_Dependency.md` | `5. Database Setup` | MongoDB install, connection string, Docker setup, credentials placement already documented. |
| `task1_Dependency.md` | `6. Redis / Queue / External Services` | Kafka/Redpanda and unused Redis/RabbitMQ context already documented. |
| `task1_Dependency.md` | `8. Docker Setup` | Local MongoDB and Redpanda Docker setup already documented. |
| `task1_Dependency.md` | `11. Ports & Networking` | Generic port conflicts and Docker networking already documented. |
| `task3_Dependency.md` | `5. Database Setup` | Main `wishlists` collection and base migrations already documented. |
| `task4_Dependency.md` | `6. Redis / Queue / External Services` -> `Product Service` | Add-item API still depends on Product Service validation. |
| `task5_Dependency.md` | `7. Environment Variables` -> `How the Project Loads Env` | `.env` loading behavior already explained. |
| `task6_Dependency.md` | `6. Redis / Queue / External Services` -> `Kafka Product Events` | Kafka env and event mindset reused. |
| `task7_Dependency.md` | `8. Docker Setup` -> `Topic Creation With Redpanda Container` | Redpanda topic creation command style reused for `recommendation.events`. |
| `task7_Dependency.md` | `11. Common Errors & Fixes` | Kafka broker/topic troubleshooting reused. |

---

## 15. Final Checklist

- [ ] Previous dependency documentation checked.
- [ ] No duplicate Go/Mongo/Docker/Kafka base setup copied.
- [ ] `backend/services/wishlist-service/go.mod` checked; no new dependency required.
- [ ] MongoDB running and reachable through `WISHLIST_MONGO_URI`.
- [ ] Base `wishlists` migrations already applied.
- [ ] Task 8 migration `004_create_wishlist_events_collection.up.js` applied.
- [ ] `wishlist_events` collection exists.
- [ ] `idx_wishlist_events_pending_retry` index exists.
- [ ] `idx_wishlist_events_type_time` index exists.
- [ ] `ttl_wishlist_events_published_at` index exists.
- [ ] Chosen Task 8 env mode added to `.env` or exported in shell.
- [ ] `.env` sourced before `go run`.
- [ ] For simple mode, `WISHLIST_ANALYTICS_EVENTS_ENABLED=false`.
- [ ] For outbox-only mode, `WISHLIST_EVENT_PUBLISHER=disabled`.
- [ ] For full mode, `WISHLIST_EVENT_PUBLISHER=kafka`.
- [ ] For full mode, `WISHLIST_KAFKA_BROKERS` points to correct broker address.
- [ ] For full mode, `recommendation.events` topic created.
- [ ] Product Service available when testing wishlist add event.
- [ ] Backend service starts on port `8084`.
- [ ] Add/remove API tested with buyer headers.
- [ ] Outbox document verified in MongoDB.
- [ ] Kafka message consumed from `recommendation.events` in full mode.
- [ ] Recommendation Service consumer idempotency planned with `event_id`.
- [ ] Failed/pending event monitoring plan documented.
- [ ] No real secrets committed.
- [ ] No PII added to analytics payload.
