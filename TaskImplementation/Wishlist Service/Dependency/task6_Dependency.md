# Project Dependency & Setup Guide

## Variable Values

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Wishlist Service` |
| `TASK_FILE_NAME` | `task6.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task6_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

Note: Ye guide `INPUT_FILE_PATH` ko analyze karke banaya gaya hai. Original `TASK_FILE_NAME` modify nahi kiya gaya. Is file ka focus sirf Task 6 availability-sync setup, dependency, environment, Kafka/DLQ, verification, and DevOps onboarding points par hai.

---

## 1. Project Overview

`TASK_FILE_NAME` ka main scope:

```text
ProductDeleted / ProductOutOfStock event consume karo
-> product_id / variant_id validate karo
-> matching wishlist items ka availability snapshot update karo
-> success ke baad Kafka offset commit karo
-> permanent bad message ko DLQ me bhejo
```

Simple Hinglish me: Product Service jab batata hai ki product delete ho gaya ya out of stock ho gaya, tab `SERVICE_NAME` buyer ke saved wishlist item ko remove nahi karta. Sirf item ka status update karta hai:

| Product event | Wishlist item status |
|---|---|
| `ProductDeleted` | `deleted` |
| `ProductOutOfStock` | `out_of_stock` |

### Runtime code paths checked

| Runtime file | Task 6 relevance |
|---|---|
| `backend/services/wishlist-service/internal/events/product_event.go` | Product event envelope, payload validation, supported event mapping. |
| `backend/services/wishlist-service/internal/events/product_consumer.go` | Message handling, availability sync call, metrics hooks, ignored events. |
| `backend/services/wishlist-service/internal/events/kafka_consumer.go` | Kafka reader, retry, DLQ publish, offset commit. |
| `backend/services/wishlist-service/internal/usecase/wishlist_service.go` | `SyncProductAvailability` business rule. |
| `backend/services/wishlist-service/internal/repository/mongo_wishlist_repository.go` | MongoDB `UpdateMany` with `arrayFilters`. |
| `backend/services/wishlist-service/internal/config/config.go` | Event backend, Kafka, topic, DLQ, retry env vars. |
| `backend/services/wishlist-service/cmd/server/main.go` | Kafka consumer wiring and background runner lifecycle. |
| `backend/services/wishlist-service/.env` | Current local event defaults. |
| `docs/11-devops-external-services.md` | Platform-level event envelope and queue guidance. |

### Important Reuse Note

Base setup already documented hai. Is file me Go install, MongoDB install, generic Docker setup, full `.env`, Product Service HTTP validation, and Cart Service setup repeat nahi kiya gaya.

Read first:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

---

## 2. Tech Stack

Most technology explanation already covered hai. Task 6 ka real new setup focus Kafka product-event consumption and DLQ readiness hai.

| Technology / Service | Task 6 Status | Required? | Setup Documentation |
|---|---|---:|---|
| Go | Reused | Yes | `task1_Dependency.md` -> `4. Dependency Management` |
| Go modules | Reused | Yes | `task1_Dependency.md` -> `4. Dependency Management` |
| MongoDB | Reused persistence | Yes | `task1_Dependency.md` -> `5. Database Setup`; `task3_Dependency.md` -> `5. Database Setup` |
| MongoDB Go Driver v2 | Reused | Yes | `task1_Dependency.md` -> `4. Dependency Management` |
| Kafka-compatible broker | Task 6 core event runtime | Required when availability sync is enabled | Explained below |
| `github.com/segmentio/kafka-go` | Already present | Yes for Kafka mode | Existing dependency in `go.mod` |
| Product Service event producer | Task 6 source dependency | Required for real event flow | Event contract explained below |
| DLQ topic | Task 6 reliability dependency | Required in Kafka mode | Explained below |
| Product Service HTTP API | Reused from Task 4 | Not called by availability consumer | `task4_Dependency.md` -> `6. Redis / Queue / External Services` |
| Cart Service HTTP API | Reused from Task 5 | Not used by Task 6 | `task5_Dependency.md` -> `6. Redis / Queue / External Services` |
| RabbitMQ | Mentioned in design docs, not implemented in runtime code | No | Do not configure `rabbitmq` for current code |
| Redis/NATS/MinIO/SMTP/Stripe/Twilio | Not introduced | No | No setup needed |

### Task 6 New Technical Concept: Product Event Consumer

Kafka ek event streaming system hai. Simple Hinglish me: Product Service message publish karta hai, aur `SERVICE_NAME` us message ko consume karke apni MongoDB snapshot update karta hai.

Current runtime supports only:

```text
WISHLIST_EVENTS_BACKEND=disabled
WISHLIST_EVENTS_BACKEND=kafka
```

Warning: `rabbitmq` value current code me valid nahi hai. Agar set kiya, startup validation fail karega with `WISHLIST_EVENTS_BACKEND must be disabled or kafka`.

---

## 3. Required Software

No new language runtime, package manager, database server, or service Dockerfile is introduced by `TASK_FILE_NAME`.

Base software setup:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`3. Required Software`
`4. Dependency Management`
`5. Database Setup`
`8. Docker Setup`
```

Task 6-specific readiness:

| Requirement | Required? | Why |
|---|---:|---|
| Kafka-compatible broker on `localhost:9092` or configured broker address | Yes for event sync | Consumer reads `product.events`. |
| Product Service or producer script publishing valid events | Yes for end-to-end sync | Without events, availability will not change. |
| DLQ topic | Yes for Kafka mode | Bad or failed messages are published there. |
| MongoDB with `wishlists` collection | Yes | Availability field is updated in MongoDB. |
| `mongosh` | Recommended | Verify before/after item status. |
| curl/Postman | Recommended | Verify service health after enabling Kafka. |

---

## 4. Dependency Management

No new Go module dependency is introduced by `TASK_FILE_NAME`.

Existing module:

```text
backend/services/wishlist-service/go.mod
```

Existing direct dependencies:

| Dependency | Version | Task 6 usage |
|---|---:|---|
| `github.com/segmentio/kafka-go` | `v0.4.49` | Kafka product-event consumer, DLQ writer, notification publisher wiring. |
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | Availability update in `wishlists.items`. |
| Go standard library `encoding/json` | Built in | Decode event envelope and payload. |
| Go standard library `log/slog` | Built in | Structured event logs. |

Reuse Go setup commands from:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Dependency Management`
```

Task 6 beginner warning:

- Do not run `go get github.com/segmentio/kafka-go`; it is already present.
- Do not add RabbitMQ packages for current runtime.
- Do not manually edit `go.sum`.
- If dependency download fails, follow Go module troubleshooting from `task1_Dependency.md`.

Useful verification:

```bash
cd backend/services/wishlist-service
go test ./internal/events ./internal/usecase ./internal/repository
```

---

## 5. Database Setup

Task 6 does not add a new database, collection, or migration.

Database setup is reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`5. Database Setup`
`8. Docker Setup`
```

Collection schema and indexes are reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Section:
`5. Database Setup`
```

### Task 6 Database Behavior

| Operation | MongoDB behavior | Setup impact |
|---|---|---|
| Product-level deleted event | Updates every matching `items.product_id`. | Existing `idx_wishlists_items_product_id` helps. |
| Variant-level out-of-stock event | Updates matching `items.product_id` + `items.variant_id`. | Existing `idx_wishlists_items_product_variant` helps. |
| Event older than wishlist item | Item is not updated because update filter uses `item.added_at <= occurred_at`. | Product events need accurate `occurred_at`. |
| Duplicate event delivery | Same `$set` runs again; final status remains same. | No new receipt collection required for Task 6. |
| No matching wishlist item | Success with `matched_count=0`; message can be committed. | No manual action needed. |

Task 6 relevant existing values:

| Item | Value | Status |
|---|---|---|
| Database type | MongoDB | Reused |
| Database name | `wishlist_db` | Reused default |
| Collection | `wishlists` | Reused |
| Availability field | `items.availability` | Reused from Task 3 |
| Allowed availability values | `unknown`, `in_stock`, `out_of_stock`, `deleted` | Reused |
| Default Mongo port | `27017` | Reused |
| Main env var | `WISHLIST_MONGO_URI` | Reused |

### Verify Database Before Event Test

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.wishlists.getIndexes()"
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.wishlists.findOne({}, { items: 1 })"
curl http://localhost:8084/readyz
```

Expected:

- `/readyz` returns healthy when MongoDB is reachable.
- `idx_wishlists_items_product_id` exists.
- Test wishlist item has `product_id`, optional `variant_id`, `added_at`, and `availability`.

---

## 6. Redis / Queue / External Services

### Kafka Product Events

Kafka is the only new Task 6 runtime dependency when availability sync is enabled.

| Item | Value |
|---|---|
| Event backend env var | `WISHLIST_EVENTS_BACKEND` |
| Enabled value | `kafka` |
| Disabled value | `disabled` |
| Broker env var | `WISHLIST_KAFKA_BROKERS` |
| Default broker | `localhost:9092` |
| Source topic env var | `WISHLIST_PRODUCT_EVENTS_TOPIC` |
| Default source topic | `product.events` |
| Consumer group env var | `WISHLIST_PRODUCT_EVENTS_GROUP_ID` |
| Default consumer group | `wishlist-service` |
| DLQ env var | `WISHLIST_PRODUCT_EVENTS_DLQ_TOPIC` |
| Default DLQ topic | `product.events.wishlist.dlq` |

The code also supports older price-event aliases:

| Alias | Current meaning | Recommendation |
|---|---|---|
| `WISHLIST_PRICE_EVENTS_GROUP_ID` | Overrides product event group id if set. | Avoid setting both aliases differently. |
| `WISHLIST_PRICE_EVENTS_DLQ_TOPIC` | Overrides product event DLQ topic if set. | Avoid setting both aliases differently. |

Important current `.env` note:

```text
backend/services/wishlist-service/.env
```

currently contains price-event aliases:

```env
WISHLIST_PRICE_EVENTS_GROUP_ID=wishlist-price-drop-service
WISHLIST_PRICE_EVENTS_DLQ_TOPIC=product.events.wishlist.price.dlq
```

Because `config.go` reads these aliases first, they override `WISHLIST_PRODUCT_EVENTS_GROUP_ID` and `WISHLIST_PRODUCT_EVENTS_DLQ_TOPIC`. For pure Task 6 availability sync, prefer explicit product-event names and remove/align conflicting aliases.

### Product Service Event Producer

Product Service ka source of truth product/inventory status hai. Task 6 me `SERVICE_NAME` Product Service HTTP API call nahi karta. Instead, Product Service must publish valid Kafka messages.

Supported event types:

| Event type | Required? | Wishlist result |
|---|---:|---|
| `ProductDeleted` | Yes | Matching items become `deleted`. |
| `ProductOutOfStock` | Yes | Matching items become `out_of_stock`. |
| `ProductPriceChanged` | Also decoded by current consumer if price-drop usecase is wired | Not Task 6; Task 7 concern. |
| Other event types | No | Ignored and committed. |

Required envelope:

```json
{
  "event_id": "evt_123",
  "event_type": "ProductOutOfStock",
  "version": 1,
  "occurred_at": "2026-05-26T10:30:00Z",
  "producer": "product-service",
  "trace_id": "trace_123",
  "payload": {
    "product_id": "prod_123",
    "variant_id": "var_1",
    "reason": "inventory_zero"
  }
}
```

Required validation rules:

| Field | Required? | Notes |
|---|---:|---|
| `event_id` | Yes | Missing value goes to DLQ. |
| `event_type` | Yes | Must be supported for Task 6 sync. |
| `version` | Yes | Must be `1`. |
| `producer` | Yes | Must be `product-service`. |
| `occurred_at` | Recommended | If missing, service uses processing time and logs warning. |
| `payload.product_id` | Yes | Missing value goes to DLQ. |
| `payload.variant_id` | Optional | If present, only matching variant item updates. |

### DLQ Behavior

DLQ means dead-letter queue/topic. Simple Hinglish me: message repeatedly fail ho ya payload permanently bad ho, to usko side topic me bhej dete hain so normal consumer block na ho.

| Scenario | Runtime behavior |
|---|---|
| Invalid JSON | DLQ, then offset commit. |
| Unsupported version | DLQ, then offset commit. |
| Unexpected producer | DLQ, then offset commit. |
| Missing event id | DLQ, then offset commit. |
| Missing product id | DLQ, then offset commit. |
| Mongo temporary failure | Retry up to `WISHLIST_PRODUCT_EVENTS_MAX_ATTEMPTS`, then DLQ. |
| Unsupported event type | Ignore and commit. |
| Successful update or no matching item | Commit. |

### Notification Commands Coupling

When `WISHLIST_EVENTS_BACKEND=kafka`, current server wiring also creates a Kafka notification publisher for price-drop flow.

Required config even for availability-sync mode:

```env
WISHLIST_NOTIFICATION_COMMANDS_TOPIC=notification.commands
WISHLIST_PRICE_DROP_TEMPLATE_KEY=wishlist_price_drop
WISHLIST_PRICE_DROP_BATCH_SIZE=500
WISHLIST_PRICE_DROP_MIN_DELTA_AMOUNT=1
```

Beginner note: Availability sync itself does not send notifications. Ye coupling current `cmd/server/main.go` wiring ki wajah se hai.

### RabbitMQ / Redis / Other Services

| Service | Task 6 status |
|---|---|
| RabbitMQ | Not implemented in current runtime. Do not set `WISHLIST_EVENTS_BACKEND=rabbitmq`. |
| Redis | Not used for Task 6. |
| NATS | Not used. |
| MinIO/S3 | Not used. |
| Elasticsearch/Typesense | Not used by Task 6. |
| SMTP/Stripe/Twilio/OAuth | Not used. |

---

## 7. Environment Variables

Full `.env` reference is already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Environment Variables`
```

Task 6 needs only event-sync env delta. Current Go code reads OS environment variables. It does not automatically load `.env`.

Create/update local env here:

```text
backend/services/wishlist-service/.env
```

Load before starting service:

```bash
cd backend/services/wishlist-service
set -a
source .env
set +a
```

### Task 6 Env Delta Only

This is not a full `.env`. Add these values only when testing availability sync with Kafka:

```env
# Enable product event consumer
WISHLIST_EVENTS_BACKEND=kafka
WISHLIST_KAFKA_BROKERS=localhost:9092

# Product availability event source
WISHLIST_PRODUCT_EVENTS_TOPIC=product.events
WISHLIST_PRODUCT_EVENTS_GROUP_ID=wishlist-service
WISHLIST_PRODUCT_EVENTS_DLQ_TOPIC=product.events.wishlist.dlq
WISHLIST_PRODUCT_EVENTS_MAX_ATTEMPTS=3
WISHLIST_PRODUCT_EVENTS_RETRY_BACKOFF=500ms

# Required by current Kafka runner wiring
WISHLIST_NOTIFICATION_COMMANDS_TOPIC=notification.commands
WISHLIST_PRICE_DROP_TEMPLATE_KEY=wishlist_price_drop
WISHLIST_PRICE_DROP_BATCH_SIZE=500
WISHLIST_PRICE_DROP_MIN_DELTA_AMOUNT=1

# Recommended if you only want Task 6 and not wishlist analytics publishing
WISHLIST_ANALYTICS_EVENTS_ENABLED=false
WISHLIST_EVENT_PUBLISHER=disabled
```

For Docker Compose service DNS:

```env
WISHLIST_KAFKA_BROKERS=wishlist-redpanda:9092
WISHLIST_MONGO_URI=mongodb://wishlist-mongo:27017
```

### Variable Explanation

| Variable | Required? | Example | Task 6 purpose | Security/config note |
|---|---:|---|---|---|
| `WISHLIST_EVENTS_BACKEND` | Required to enable sync | `kafka` | Starts product event consumer. | Allowed values: `disabled`, `kafka`. |
| `WISHLIST_KAFKA_BROKERS` | Required when Kafka enabled | `localhost:9092` | Broker addresses CSV. | Use internal broker addresses in prod. |
| `WISHLIST_PRODUCT_EVENTS_TOPIC` | Required when Kafka enabled | `product.events` | Source topic for Product events. | Product Service must publish here. |
| `WISHLIST_PRODUCT_EVENTS_GROUP_ID` | Required when Kafka enabled | `wishlist-service` | Stable consumer group id. | Use one group per environment. |
| `WISHLIST_PRODUCT_EVENTS_DLQ_TOPIC` | Required when Kafka enabled | `product.events.wishlist.dlq` | Failed-message topic. | Monitor and alert on DLQ volume. |
| `WISHLIST_PRODUCT_EVENTS_MAX_ATTEMPTS` | Optional | `3` | Retry attempts before DLQ. | Must be positive. |
| `WISHLIST_PRODUCT_EVENTS_RETRY_BACKOFF` | Optional | `500ms` | Delay between retries. | Must be positive duration. |
| `WISHLIST_PRICE_EVENTS_GROUP_ID` | Optional alias | `wishlist-price-drop-service` | Overrides group id in current config. | Avoid conflict with product group var. |
| `WISHLIST_PRICE_EVENTS_DLQ_TOPIC` | Optional alias | `product.events.wishlist.price.dlq` | Overrides DLQ topic in current config. | Avoid conflict with product DLQ var. |
| `WISHLIST_NOTIFICATION_COMMANDS_TOPIC` | Required by Kafka runner | `notification.commands` | Price-drop notification command topic wiring. | Topic should exist if price events are used. |
| `WISHLIST_PRICE_DROP_TEMPLATE_KEY` | Required by Kafka runner | `wishlist_price_drop` | Notification template key for Task 7 flow. | Keep aligned with Notification Service. |
| `WISHLIST_ANALYTICS_EVENTS_ENABLED` | Optional | `false` for Task 6-only test | Avoids unrelated outbox setup/publisher. | Enable only when analytics setup is ready. |
| `WISHLIST_EVENT_PUBLISHER` | Optional | `disabled` for Task 6-only test | Avoids unrelated analytics Kafka publisher. | Use `kafka` only with broker ready. |

### Common Task 6 Env Mistakes

| Mistake | Result | Fix |
|---|---|---|
| `.env` edited but not sourced | Service keeps old/default values. | Run `set -a; source .env; set +a`. |
| `WISHLIST_EVENTS_BACKEND=rabbitmq` | Startup validation fails. | Use `kafka` or `disabled`. |
| Kafka broker URL uses HTTP scheme | Kafka connection fails. | Use `localhost:9092`, not `http://localhost:9092`. |
| Product and price alias vars conflict | Consumer group or DLQ topic unexpected. | Keep only one naming style or set both to same value. |
| Analytics left enabled while only testing Task 6 | Extra `wishlist_events`/Kafka publisher behavior appears. | Set `WISHLIST_ANALYTICS_EVENTS_ENABLED=false` and `WISHLIST_EVENT_PUBLISHER=disabled`. |

---

## 8. Docker Setup

No new Dockerfile, service container, volume, or network is introduced by `TASK_FILE_NAME`.

Reuse base Docker dependency setup:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`8. Docker Setup`
```

Task 6 Docker-specific notes:

| Docker item | Status | Meaning |
|---|---|---|
| MongoDB container | Reused | Required if MongoDB is not installed locally. |
| Redpanda/Kafka container | Reused, now required for event sync | Needed when `WISHLIST_EVENTS_BACKEND=kafka`. |
| Product Service container | External dependency | Must publish events for full end-to-end sync. |
| RabbitMQ container | Not needed | Current runtime does not support RabbitMQ. |
| Service Dockerfile | Not found for this service in current snapshot | Local service is run with `go run`. |

### Topic Creation With Redpanda Container

If you used the previous guide's Redpanda container named `wishlist-redpanda`, create topics:

```bash
docker exec wishlist-redpanda rpk topic create product.events
docker exec wishlist-redpanda rpk topic create product.events.wishlist.dlq
docker exec wishlist-redpanda rpk topic create notification.commands
```

Verify:

```bash
docker exec wishlist-redpanda rpk topic list
```

Beginner note: Some local brokers auto-create topics, but explicit topic creation is cleaner because typo issues become visible early.

---

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Read these first:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

Minimum sections:

- `task1_Dependency.md` -> `3. Required Software`
- `task1_Dependency.md` -> `4. Dependency Management`
- `task1_Dependency.md` -> `5. Database Setup`
- `task1_Dependency.md` -> `6. Redis / Queue / External Services`
- `task1_Dependency.md` -> `7. Environment Variables`
- `task1_Dependency.md` -> `8. Docker Setup`
- `task1_Dependency.md` -> `9. Local Development Setup`
- `task3_Dependency.md` -> `5. Database Setup`
- `task4_Dependency.md` -> `6. Redis / Queue / External Services`

### Step 2: Go to Project Directory

```bash
cd Ecommerce
cd backend/services/wishlist-service
```

### Step 3: Install Only New Dependencies

None for Task 6.

If Go modules were never downloaded:

```bash
go mod download
```

### Step 4: Setup Only New Databases/Services

No new database. Use MongoDB setup from previous dependency docs.

Task 6 needs Kafka readiness:

```bash
docker ps
docker exec wishlist-redpanda rpk cluster info
docker exec wishlist-redpanda rpk topic list
```

Expected topics:

```text
product.events
product.events.wishlist.dlq
notification.commands
```

### Step 5: Add Only New or Changed Environment Variables

Update:

```text
backend/services/wishlist-service/.env
```

Use the Task 6 env delta from section 7.

Load env:

```bash
set -a
source .env
set +a
```

### Step 6: Run Migrations If Needed

No Task 6 migration.

If this is a fresh local DB, use previous migration setup:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Section:
`5. Database Setup`
```

Minimum command check:

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.wishlists.getIndexes()"
```

### Step 7: Start Backend Service

```bash
go run ./cmd/server
```

Expected logs should include:

```text
events_backend=kafka
wishlist product event consumer started
```

Health check:

```bash
curl http://localhost:8084/healthz
curl http://localhost:8084/readyz
```

### Step 8: Verify Task 6 Functionality

1. Make sure a wishlist item exists with `availability: "in_stock"` or `availability: "unknown"`.

2. Publish a Product out-of-stock event.

Important: `occurred_at` should be after the test item's `added_at`, because the repository protects newly added items from older stale events.

```bash
EVENT_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
printf '{"event_id":"evt_task6_001","event_type":"ProductOutOfStock","version":1,"occurred_at":"%s","producer":"product-service","trace_id":"trace_task6_001","payload":{"product_id":"prod_123","variant_id":"var_1","reason":"inventory_zero"}}\n' "$EVENT_TIME" \
  | docker exec -i wishlist-redpanda rpk topic produce product.events
```

3. Verify MongoDB update:

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval 'db.wishlists.findOne({ "items.product_id": "prod_123" }, { "items.product_id": 1, "items.variant_id": 1, "items.availability": 1 })'
```

Expected item status:

```text
out_of_stock
```

4. Publish a deleted event:

```bash
EVENT_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
printf '{"event_id":"evt_task6_002","event_type":"ProductDeleted","version":1,"occurred_at":"%s","producer":"product-service","trace_id":"trace_task6_002","payload":{"product_id":"prod_123","reason":"seller_deleted"}}\n' "$EVENT_TIME" \
  | docker exec -i wishlist-redpanda rpk topic produce product.events
```

Expected item status:

```text
deleted
```

---

## 10. Running the Project

### Minimum Local Run Without Task 6 Events

Use this when you only want basic API health and Mongo setup:

```env
WISHLIST_EVENTS_BACKEND=disabled
WISHLIST_ANALYTICS_EVENTS_ENABLED=false
WISHLIST_EVENT_PUBLISHER=disabled
```

Run:

```bash
cd backend/services/wishlist-service
set -a
source .env
set +a
go run ./cmd/server
```

### Task 6 Event-Enabled Run

Required:

- MongoDB running.
- Kafka-compatible broker running.
- Topics created.
- Product event producer publishing valid messages.
- Env loaded with `WISHLIST_EVENTS_BACKEND=kafka`.

Run:

```bash
cd backend/services/wishlist-service
set -a
source .env
set +a
go run ./cmd/server
```

Stop with `Ctrl+C`. Service shutdown closes HTTP server and Kafka readers/writers.

---

## 11. Ports & Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| `SERVICE_NAME` HTTP server | `8084` | Health and wishlist APIs | Reused |
| MongoDB | `27017` | Wishlist persistence | Reused |
| Kafka / Redpanda | `9092` | Product events, DLQ, optional analytics | Reused but required for Task 6 event mode |
| Product Service HTTP | `8082` | Add-item validation, not Task 6 consumer path | Reused |
| Cart Service HTTP | `8083` | Move-to-cart, not Task 6 consumer path | Reused |

### Port Conflicts

Kafka port busy:

```bash
lsof -i :9092
```

Use a different broker port only if you also update:

```env
WISHLIST_KAFKA_BROKERS=localhost:<new-port>
```

### Docker Networking

| Where `SERVICE_NAME` runs | Mongo URI | Kafka broker |
|---|---|---|
| Host machine | `mongodb://localhost:27017` | `localhost:9092` |
| Docker Compose container | `mongodb://wishlist-mongo:27017` | `wishlist-redpanda:9092` |

Beginner note: `localhost` inside a container means the same container. Dusre container ko call karna hai to Compose service name use karo.

---

## 12. Common Errors & Fixes

Generic Go, MongoDB, `.env`, Docker daemon, and port `8084` issues are already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`12. Common Errors & Fixes`
```

Task 6-specific issues:

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `WISHLIST_EVENTS_BACKEND must be disabled or kafka` | Env value is `rabbitmq` or typo. | Set `WISHLIST_EVENTS_BACKEND=kafka` or `disabled`. | Keep allowed values documented. |
| `WISHLIST_KAFKA_BROKERS is required when WISHLIST_EVENTS_BACKEND=kafka` | Broker env blank. | Set `WISHLIST_KAFKA_BROKERS=localhost:9092`. | Do not leave blank env vars in `.env`. |
| Service starts but no availability changes | No valid event published, wrong topic, or wrong consumer group expectations. | Publish to `product.events`; check service logs and topic list. | Create topics explicitly and log event ids. |
| Event goes to DLQ with `missing_product_id` | Payload does not include `payload.product_id`. | Fix producer payload. | Contract test Product Service event schema. |
| Event goes to DLQ with `unexpected_producer` | `producer` is not `product-service`. | Set producer correctly. | Standardize envelope library. |
| Event ignored | Unsupported `event_type`. | Use `ProductDeleted` or `ProductOutOfStock`. | Keep producer event names exact. |
| Message repeatedly retries then DLQ | Mongo update error or handler error. | Check Mongo connection and service logs. | Monitor retry count and DLQ. |
| Item added after event does not update | `item.added_at` is after event `occurred_at`. | This is expected stale-event protection. Publish newer event if product is still unavailable. | Product Service should publish accurate event time. |
| DLQ topic has messages but no alert | Monitoring not configured. | Add DLQ consumer/alerting in DevOps. | Treat DLQ growth as production incident signal. |
| `notification commands topic is required` | Kafka backend enabled but companion topic env missing. | Set `WISHLIST_NOTIFICATION_COMMANDS_TOPIC`. | Remember current Kafka runner wires price-drop publisher too. |

---

## 13. Security & Best Practices

### Security and Configuration Audit

| Area | Current observation | Risk | Suggested fix |
|---|---|---|---|
| Event producer validation | Consumer requires `producer=product-service`. | Good, but not cryptographic authentication. | In prod, use broker ACLs, TLS/SASL, and topic-level permissions. |
| Kafka transport security | Local defaults use plaintext `localhost:9092`. | Fine local, unsafe in shared/prod. | Use managed Kafka or TLS/SASL where required. |
| DLQ content | DLQ stores original message body. | Body may contain sensitive product metadata later. | Keep payload minimal and restrict DLQ access. |
| Event id requirement | Missing `event_id` goes to DLQ. | Good for traceability. | Product Service must generate stable unique ids. |
| Idempotency | Availability sync uses deterministic `$set`. | Safe for duplicate delivery. | Keep updates idempotent for future event types too. |
| Event ordering | Code relies partly on `occurred_at` and producer keying. | Out-of-order events can confuse status if producer emits wrong timestamps. | Product Service should key Kafka messages by `product_id`. |
| Alias env vars | `WISHLIST_PRICE_EVENTS_*` override product event values. | Confusing local/prod config. | Align aliases or remove old aliases from env. |
| RabbitMQ docs mismatch | Platform docs mention RabbitMQ, runtime only supports Kafka. | New developer may configure unsupported backend. | Document current allowed values and add RabbitMQ only when implemented. |
| Health checks | `/readyz` checks Mongo, not Kafka. | Service can look ready while consumer cannot connect. | Add Kafka readiness or separate consumer health metric if production needs it. |
| Metrics | Interface exists but default metrics are no-op. | No runtime metrics unless wired later. | Wire Prometheus/OpenTelemetry implementation. |

### Task 6 Best Practices

- Use `product_id` as Kafka message key so same product events keep order in one partition.
- Keep event payload small: id, variant, reason, timestamps; no full product document unless needed.
- Keep `event_id`, `trace_id`, and `occurred_at` in every event.
- Monitor `product.events.wishlist.dlq`.
- Never delete wishlist items automatically on Product delete; keep user intent visible with `deleted` status.
- Use `WISHLIST_EVENTS_BACKEND=disabled` for beginner API-only local testing.
- Use `WISHLIST_EVENTS_BACKEND=kafka` only after broker and topics are ready.
- Keep local `.env` and production secrets separate.

---

## 14. Missing or Misconfigured Things

| Missing/misaligned item | Why it matters | Recommended action |
|---|---|---|
| No service Dockerfile found for current runtime | Containerized local/prod run is manual. | Add Dockerfile for `backend/services/wishlist-service`. |
| No local compose file found in repo snapshot | Mongo/Kafka setup depends on manual commands. | Add a repo-level local compose with MongoDB and Redpanda/Kafka. |
| `.env` has price-event aliases overriding product-event values | Task 6 consumer group/DLQ may not match expected names. | Rename to product-event vars or set both aliases consistently. |
| RabbitMQ alternate appears in task guide but runtime does not support it | Beginners may configure unsupported backend. | Keep docs clear or implement RabbitMQ adapter later. |
| `/readyz` does not prove Kafka consumer health | Deployment may route traffic while event consumer is broken. | Add Kafka consumer health/lag metrics. |
| Metrics interface is no-op by default | Task 6 observability is not visible without wiring. | Add Prometheus/OpenTelemetry implementation. |
| Product event contract not in `api/master-api.json` | Producer/consumer schema can drift. | Add event schema docs or shared contract tests. |
| No replay guide for DLQ messages | Operators need recovery path. | Add DLQ replay runbook before production. |

---

## 15. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Required Software` | Git, Go, Docker, MongoDB, curl setup already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Dependency Management` | Go modules and common commands already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Setup` | MongoDB install, connection string, and Docker setup already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Redis / Queue / External Services` | Kafka/Redpanda base explanation and local broker setup already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Environment Variables` | Full service `.env` reference already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Docker Setup` | Base Docker dependency containers already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `12. Common Errors & Fixes` | Generic Go, MongoDB, Docker, and `.env` troubleshooting already documented. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | MongoDB ownership decision | Same database ownership remains unchanged. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `5. Database Setup` | `wishlists` schema, validator, indexes, and `availability` enum already documented. |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | Product Service HTTP validation | Product Service base setup is reused, though Task 6 consumes events instead of HTTP. |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | Cart Service setup | Cart Service remains runtime config but not Task 6-specific. |

---

## 16. Final Checklist

- [ ] Previous dependency documentation checked.
- [ ] No duplicated Go/Mongo/Docker installation guide added.
- [ ] No new Go dependency added for Task 6.
- [ ] MongoDB running and `wishlists` collection ready.
- [ ] Existing `items.availability` field verified.
- [ ] Kafka/Redpanda running for event-enabled mode.
- [ ] `product.events` topic exists.
- [ ] `product.events.wishlist.dlq` topic exists.
- [ ] `notification.commands` topic exists because current Kafka runner requires it.
- [ ] `WISHLIST_EVENTS_BACKEND=kafka` set only when Kafka is ready.
- [ ] `WISHLIST_KAFKA_BROKERS` points to the correct broker address.
- [ ] Product-event group/DLQ aliases are not conflicting.
- [ ] `.env` loaded into shell before running service.
- [ ] `go test ./internal/events ./internal/usecase ./internal/repository` passes.
- [ ] Backend service starts successfully.
- [ ] `/healthz` and `/readyz` checked.
- [ ] Product out-of-stock event updates matching wishlist item to `out_of_stock`.
- [ ] Product deleted event updates matching wishlist item to `deleted`.
- [ ] Invalid event lands in DLQ and normal consumer continues.
- [ ] Logs checked for `wishlist availability synced`.
- [ ] No duplicate setup documentation added.
