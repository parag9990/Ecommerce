# Project Dependency & Setup Guide

## Variable Values

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Wishlist Service` |
| `TASK_FILE_NAME` | `task7.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task7_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

Note: Ye guide `INPUT_FILE_PATH` ko analyze karke banaya gaya hai. Original `TASK_FILE_NAME` modify nahi kiya gaya. Is file ka focus sirf Task 7 price-drop event dependency, setup, environment, Kafka, Notification Service handoff, migration/index verification, and DevOps onboarding points par hai.

---

## 1. Project Overview

`TASK_FILE_NAME` ka scope:

```text
ProductPriceChanged event consume karo
-> matching wishlist items find karo
-> last_known_price se price drop detect karo
-> notification.commands me SendNotification command publish karo
-> wishlist item ka last_known_price latest price se update karo
```

Simple Hinglish me: Product ka price kam hota hai, aur same product kisi buyer ki wishlist me saved hai, to `SERVICE_NAME` buyer ke liye price-drop notification command publish karta hai. Actual email/SMS/push delivery Notification Service ka kaam hai.

### Runtime code paths checked

| Runtime file | Task 7 relevance |
|---|---|
| `backend/services/wishlist-service/internal/usecase/price_drop_service.go` | Price-drop detection, notification command build, idempotency key, snapshot update flow. |
| `backend/services/wishlist-service/internal/events/product_event.go` | `ProductPriceChanged` event envelope/payload validation. |
| `backend/services/wishlist-service/internal/events/product_consumer.go` | Product event consumer routes price events to price-drop usecase. |
| `backend/services/wishlist-service/internal/events/kafka_consumer.go` | Kafka reader, retry, DLQ, offset commit behavior. |
| `backend/services/wishlist-service/internal/events/notification_publisher.go` | Kafka publisher for `notification.commands`. |
| `backend/services/wishlist-service/internal/repository/mongo_wishlist_repository.go` | Candidate query and `last_known_price` update. |
| `backend/services/wishlist-service/internal/config/config.go` | Task 7 env vars, Kafka config, notification topic config. |
| `backend/services/wishlist-service/cmd/server/main.go` | Kafka consumer and notification publisher wiring. |
| `backend/services/wishlist-service/migrations/mongo/003_add_price_drop_candidate_index.up.js` | Price-drop candidate index. |
| `backend/services/wishlist-service/.env` | Local Task 7 variables exist but event backend is disabled by default. |
| `backend/services/product-service/.env` | Product events currently point to RabbitMQ in this snapshot, which does not match current `SERVICE_NAME` Kafka consumer. |
| `backend/services/notification-service/.env` | Notification DB/provider config exists, but no Kafka `notification.commands` consumer config was found in this snapshot. |

### Important Reuse Note

Base setup already documented hai. Is file me Go install, MongoDB install, generic Docker setup, full `.env`, Product Service HTTP validation, Cart Service setup, and generic troubleshooting repeat nahi kiya gaya.

Read first:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
TaskImplementation/{SERVICE_NAME}/task6_Dependency.md
```

---

## 2. Tech Stack

Most technology explanation already exists in previous dependency files. Task 7 ka real new setup focus hai: price-change event consumption plus Notification Service command publishing.

| Technology / Service | Task 7 Status | Required? | Setup Documentation |
|---|---|---:|---|
| Go | Reused runtime | Yes | `task1_Dependency.md` -> `4. Dependency Management` |
| Go modules | Reused dependency system | Yes | `task1_Dependency.md` -> `4. Dependency Management` |
| `net/http` | Reused HTTP health/API server | Yes | `task1_Dependency.md` -> `2. Tech Stack` |
| MongoDB | Reused persistence | Yes | `task1_Dependency.md` -> `5. Database Setup`; `task3_Dependency.md` -> `5. Database Setup` |
| MongoDB Go Driver v2 | Reused | Yes | `task1_Dependency.md` -> `4. Dependency Management` |
| Kafka-compatible broker | Task 7 event runtime | Yes when price-drop events are enabled | Explained below; base setup reused from `task1_Dependency.md` and `task6_Dependency.md` |
| `github.com/segmentio/kafka-go` | Already present | Yes for Kafka mode | Existing dependency in `go.mod` |
| Product Service event producer | Task 7 source dependency | Yes for real price-drop flow | Must publish `ProductPriceChanged` to the same broker/topic consumed by `SERVICE_NAME` |
| Notification Service command consumer | Task 7 downstream dependency | Yes for real notification delivery | Must consume `notification.commands` and support template `wishlist_price_drop` |
| Docker / Redpanda | Reused local broker option | Recommended | `task1_Dependency.md` -> `8. Docker Setup`; `task6_Dependency.md` -> topic creation |
| Redis | Not introduced | No | No setup needed |
| RabbitMQ | Present in some platform/env docs, but not supported by current `SERVICE_NAME` runtime | No for current code | Do not set `WISHLIST_EVENTS_BACKEND=rabbitmq` |
| SMTP/SMS/Push providers | Notification Service concern | Not for `SERVICE_NAME` | Configure only inside Notification Service when delivery is tested |

### Task 7 new technical concept: price-drop command flow

Kafka ek message streaming system hai. Simple Hinglish me: Product Service price-change message publish karta hai, `SERVICE_NAME` us message ko consume karta hai, MongoDB me interested wishlist users dhundta hai, aur Notification Service ko command bhejta hai.

Current `SERVICE_NAME` runtime supports only:

```text
WISHLIST_EVENTS_BACKEND=disabled
WISHLIST_EVENTS_BACKEND=kafka
```

Warning: `rabbitmq` current `SERVICE_NAME` code me valid value nahi hai. Agar set kiya, startup validation fail karega.

---

## 3. Required Software

Base software setup reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`3. Required Software`
`4. Dependency Management`
`5. Database Setup`
`8. Docker Setup`
```

Task 7-specific readiness:

| Requirement | Required? | Why |
|---|---:|---|
| Kafka-compatible broker on `localhost:9092` or configured broker address | Yes for price-drop event mode | Consumer reads `product.events`; publisher writes `notification.commands`. |
| `product.events` topic | Yes | Product price events arrive here. |
| `notification.commands` topic | Yes | Price-drop notification commands are published here. |
| Task 7 DLQ topic | Yes for Kafka mode | Failed price messages are written to DLQ. |
| Product Service event producer | Yes for end-to-end flow | Must publish `ProductPriceChanged`. |
| Notification Service or test consumer | Yes for delivery, optional for local publisher verification | Without it, commands can be verified in Kafka but no email/SMS/push goes out. |
| MongoDB with `wishlists` collection | Yes | Candidate lookup and snapshot update happen in MongoDB. |
| `mongosh` | Recommended | Verify index and price snapshot updates. |
| `rpk` inside Redpanda container or equivalent Kafka CLI | Recommended | Create topics and produce/consume test events. |
| curl/Postman | Recommended | Verify `/healthz` and `/readyz`. |

Beginner note: Agar aap sirf HTTP wishlist APIs test kar rahe ho, Kafka disabled rakho. Task 7 test karne ke liye Kafka enable karna padega.

---

## 4. Dependency Management

No new Go module dependency is introduced by `TASK_FILE_NAME`.

Existing module:

```text
backend/services/wishlist-service/go.mod
```

Existing direct dependencies:

| Dependency | Version | Task 7 usage |
|---|---:|---|
| `github.com/segmentio/kafka-go` | `v0.4.49` | Consume `product.events`, publish DLQ messages, publish `notification.commands`. |
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | Query price-drop candidates and update `last_known_price`. |
| Go standard library `encoding/json` | Built in | Decode event envelope and encode notification command envelope. |
| Go standard library `log/slog` | Built in | Structured logs for price-drop processing. |

Reuse Go setup commands from:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Dependency Management`
```

Task 7 beginner warnings:

- Do not run `go get github.com/segmentio/kafka-go`; it is already present.
- Do not add RabbitMQ package for current `SERVICE_NAME` runtime.
- Do not manually edit `go.sum`.
- If `go mod download` fails, follow Go module troubleshooting from `task1_Dependency.md`.

Useful Task 7 verification tests:

```bash
cd backend/services/wishlist-service
go test ./internal/usecase ./internal/events ./internal/repository ./internal/config
```

---

## 5. Database Setup

MongoDB installation, credentials, Docker setup, and basic collection setup are reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`5. Database Setup`
`8. Docker Setup`
```

Collection schema and base indexes are reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Section:
`5. Database Setup`
```

### Task 7 Database Behavior

| Operation | MongoDB behavior | Setup impact |
|---|---|---|
| Find candidates | Finds wishlist items with same `product_id`, optional same `variant_id`, same currency, old snapshot amount greater than new amount, and not `deleted`. | Needs `wishlists` collection and price-drop candidate index. |
| Publish notification command | No Mongo write in Notification DB from `SERVICE_NAME`. | Notification Service owns its own DB. |
| Update snapshot | Sets matching `items.$[item].last_known_price` to latest product price. | MongoDB must be reachable after command publish succeeds. |
| Price increase or same price | No notification, but snapshot still updates. | Future drop detection uses latest baseline. |
| Currency mismatch | No notification candidate. | Snapshot update still writes latest currency/amount for matching product/variant. |

### Task 7 Index / Migration

Task 7 uses this migration:

```text
backend/services/wishlist-service/migrations/mongo/003_add_price_drop_candidate_index.up.js
```

Index name:

```text
idx_wishlists_price_drop_candidates
```

Index keys:

```javascript
{
  "items.product_id": 1,
  "items.variant_id": 1,
  "items.last_known_price.currency": 1,
  "items.last_known_price.amount": 1,
  "items.availability": 1
}
```

This index was already mentioned in previous dependency docs, so full MongoDB setup is not repeated here. For Task 7, verify it exists:

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.wishlists.getIndexes()"
```

If missing, run only the Task 7-relevant index migration:

```bash
cd backend/services/wishlist-service
mongosh "mongodb://localhost:27017" migrations/mongo/003_add_price_drop_candidate_index.up.js
```

Beginner note: Service startup also calls collection setup and can ensure indexes. Manual migration is better for controlled environments because DB changes are explicit and auditable.

### Credentials placement

Same as previous setup:

- Local: `backend/services/wishlist-service/.env`, then manually load into shell.
- Docker Compose: `env_file` or `environment`.
- Production: secret manager or Kubernetes Secret.
- Do not put Notification Service Mongo credentials in `SERVICE_NAME` env. `SERVICE_NAME` does not read Notification DB.

---

## 6. Redis / Queue / External Services

### Kafka Product Price Events

Kafka is mandatory for current Task 7 event-enabled runtime.

| Item | Value |
|---|---|
| Event backend env var | `WISHLIST_EVENTS_BACKEND` |
| Enabled value | `kafka` |
| Disabled value | `disabled` |
| Broker env var | `WISHLIST_KAFKA_BROKERS` |
| Default broker | `localhost:9092` |
| Source topic env var | `WISHLIST_PRODUCT_EVENTS_TOPIC` |
| Default source topic | `product.events` |
| Task 7 group alias | `WISHLIST_PRICE_EVENTS_GROUP_ID` |
| Product-event group var | `WISHLIST_PRODUCT_EVENTS_GROUP_ID` |
| Task 7 DLQ alias | `WISHLIST_PRICE_EVENTS_DLQ_TOPIC` |
| Product-event DLQ var | `WISHLIST_PRODUCT_EVENTS_DLQ_TOPIC` |
| Notification command topic env var | `WISHLIST_NOTIFICATION_COMMANDS_TOPIC` |
| Default notification command topic | `notification.commands` |

Current config reads the price aliases first:

```text
WISHLIST_PRICE_EVENTS_GROUP_ID overrides WISHLIST_PRODUCT_EVENTS_GROUP_ID
WISHLIST_PRICE_EVENTS_DLQ_TOPIC overrides WISHLIST_PRODUCT_EVENTS_DLQ_TOPIC
```

Keep aliases aligned with product-event vars if both are present.

### Product Service Event Producer

Product Service must publish this event to the same broker/topic consumed by `SERVICE_NAME`:

```json
{
  "event_id": "evt_price_123",
  "event_type": "ProductPriceChanged",
  "version": 1,
  "occurred_at": "2026-05-26T10:30:00Z",
  "producer": "product-service",
  "trace_id": "trace_123",
  "payload": {
    "product_id": "prod_123",
    "variant_id": "var_1",
    "new_price": {
      "amount": 299900,
      "currency": "INR"
    },
    "title": "Running Shoes",
    "image_url": "https://cdn.example.com/prod_123/main.jpg",
    "product_url": "/products/prod_123"
  }
}
```

Required validation rules:

| Field | Required? | Notes |
|---|---:|---|
| `event_id` | Yes | Missing value goes to DLQ. |
| `event_type` | Yes | Must be `ProductPriceChanged` for Task 7. |
| `version` | Yes | Must be `1`. |
| `producer` | Yes | Must be `product-service`. |
| `payload.product_id` | Yes | Used to find wishlist items. |
| `payload.variant_id` | Optional | If present, only matching variant items are processed. |
| `payload.new_price.amount` | Yes | Must be `>= 0`; use minor units such as paise/cents. |
| `payload.new_price.currency` | Yes | Must be uppercase 3-letter code like `INR` or `USD`. |
| `title`, `image_url`, `product_url` | Optional | Passed as notification template variables. |

Important mismatch in current repo snapshot:

```text
backend/services/product-service/.env
```

uses:

```env
PRODUCT_EVENTS_TOPIC=product.events
PRODUCT_EVENT_BROKER=rabbitmq
```

But current `SERVICE_NAME` runtime consumes Kafka only. For true end-to-end Task 7 testing, Product Service must publish `ProductPriceChanged` to Kafka, or a RabbitMQ-to-Kafka bridge/adapter must exist.

### Notification Service Command Consumer

`SERVICE_NAME` publishes this command envelope to `notification.commands`:

```json
{
  "event_id": "notif_evt_price_123_user_123_prod_123_var_1",
  "event_type": "SendNotification",
  "version": 1,
  "occurred_at": "2026-05-26T10:30:01Z",
  "producer": "wishlist-service",
  "trace_id": "trace_123",
  "payload": {
    "user_id": "user_123",
    "channel": "auto",
    "template_key": "wishlist_price_drop",
    "idempotency_key": "wishlist_price_drop:user_123:prod_123:var_1:INR:299900",
    "variables": {
      "product_id": "prod_123",
      "variant_id": "var_1",
      "title": "Running Shoes",
      "old_price_amount": 349900,
      "new_price_amount": 299900,
      "currency": "INR",
      "savings_amount": 50000,
      "product_url": "/products/prod_123",
      "image_url": "https://cdn.example.com/prod_123/main.jpg"
    }
  }
}
```

Notification Service requirements for end-to-end delivery:

| Requirement | Why |
|---|---|
| Consume `notification.commands` | Otherwise commands stay in Kafka and no user message is delivered. |
| Support `SendNotification` event type | Current publisher uses this command type. |
| Template key `wishlist_price_drop` exists | Notification rendering needs a matching template. |
| Enforce preferences | `channel=auto` means Notification Service chooses allowed email/SMS/push channel. |
| Deduplicate by `idempotency_key` | Product event retries can publish same command again. |
| Provider config only in Notification Service | SMTP/SMS/push credentials do not belong in `SERVICE_NAME`. |

Current snapshot note:

- `backend/services/notification-service/.env` has MongoDB, gRPC, provider, retry, and RabbitMQ-style queue settings.
- No Kafka `notification.commands` consumer setting was found in that `.env`.
- If Notification Service consumer is not implemented yet, local Task 7 verification should consume `notification.commands` directly with a Kafka CLI.

### DLQ Behavior

DLQ means dead-letter queue/topic. Simple Hinglish me: bad ya repeatedly failing message ko side topic me bhej dete hain taaki normal consumer block na ho.

| Scenario | Runtime behavior |
|---|---|
| Invalid JSON | DLQ, then offset commit. |
| Unsupported version | DLQ, then offset commit. |
| Unexpected producer | DLQ, then offset commit. |
| Missing event id | DLQ, then offset commit. |
| Missing product id | DLQ, then offset commit. |
| Invalid new price | DLQ, then offset commit. |
| Unsupported event type | Ignored and committed. |
| Notification publish failure | Retry up to configured attempts, then DLQ. |
| Mongo candidate query/update failure | Retry up to configured attempts, then DLQ. |

### Redis, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, OAuth

| Service | Task 7 status for `SERVICE_NAME` |
|---|---|
| Redis | Not used. |
| RabbitMQ | Not supported by current `SERVICE_NAME` runtime. |
| NATS | Not used. |
| MinIO/S3 | Not used. |
| Elasticsearch/Typesense | Not used by Task 7. |
| SMTP/SMS/Push providers | Notification Service concern only. |
| Stripe/Twilio/OAuth | Not used by Task 7. |

---

## 7. Environment Variables

Full `.env` reference is already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Environment Variables`
```

Task 7 env values are already present in:

```text
backend/services/wishlist-service/.env
```

but event processing is disabled by default:

```env
WISHLIST_EVENTS_BACKEND=disabled
```

Current Go code reads OS environment variables. It does not automatically load `.env`.

Load before starting service:

```bash
cd backend/services/wishlist-service
set -a
source .env
set +a
```

### Task 7 Env Delta Only

This is not a full `.env`. Add or change these values only when testing price-drop events with Kafka:

```env
# Enable product event consumer
WISHLIST_EVENTS_BACKEND=kafka
WISHLIST_KAFKA_BROKERS=localhost:9092

# Product price event source
WISHLIST_PRODUCT_EVENTS_TOPIC=product.events
WISHLIST_PRICE_EVENTS_GROUP_ID=wishlist-price-drop-service
WISHLIST_PRICE_EVENTS_DLQ_TOPIC=product.events.wishlist.price.dlq
WISHLIST_PRODUCT_EVENTS_MAX_ATTEMPTS=3
WISHLIST_PRODUCT_EVENTS_RETRY_BACKOFF=500ms

# Notification command publishing
WISHLIST_NOTIFICATION_COMMANDS_TOPIC=notification.commands
WISHLIST_PRICE_DROP_TEMPLATE_KEY=wishlist_price_drop
WISHLIST_PRICE_DROP_BATCH_SIZE=500
WISHLIST_PRICE_DROP_MIN_DELTA_AMOUNT=1

# Recommended if you only want Task 7 and not wishlist analytics publishing
WISHLIST_ANALYTICS_EVENTS_ENABLED=false
WISHLIST_EVENT_PUBLISHER=disabled
```

For Docker Compose service DNS:

```env
WISHLIST_KAFKA_BROKERS=wishlist-redpanda:9092
WISHLIST_MONGO_URI=mongodb://wishlist-mongo:27017
```

### Variable Explanation

| Variable | Required? | Example | Task 7 purpose | Security/config note |
|---|---:|---|---|---|
| `WISHLIST_EVENTS_BACKEND` | Required to enable Task 7 | `kafka` | Starts product event consumer and notification publisher wiring. | Allowed values: `disabled`, `kafka`. |
| `WISHLIST_KAFKA_BROKERS` | Required when Kafka enabled | `localhost:9092` | Broker address CSV. | Use internal broker addresses in prod. |
| `WISHLIST_PRODUCT_EVENTS_TOPIC` | Required when Kafka enabled | `product.events` | Source topic for Product events. | Product Service must publish here. |
| `WISHLIST_PRICE_EVENTS_GROUP_ID` | Optional alias but used in current `.env` | `wishlist-price-drop-service` | Consumer group id for price-drop/event consumer. | Avoid conflict with `WISHLIST_PRODUCT_EVENTS_GROUP_ID`. |
| `WISHLIST_PRICE_EVENTS_DLQ_TOPIC` | Optional alias but used in current `.env` | `product.events.wishlist.price.dlq` | DLQ topic for failed price events. | Monitor this topic. |
| `WISHLIST_PRODUCT_EVENTS_MAX_ATTEMPTS` | Optional | `3` | Retry attempts before DLQ. | Must be positive. |
| `WISHLIST_PRODUCT_EVENTS_RETRY_BACKOFF` | Optional | `500ms` | Delay between retries. | Must be positive duration. |
| `WISHLIST_NOTIFICATION_COMMANDS_TOPIC` | Required when Kafka enabled | `notification.commands` | Destination topic for notification commands. | Notification Service must consume same topic. |
| `WISHLIST_PRICE_DROP_TEMPLATE_KEY` | Required when Kafka enabled | `wishlist_price_drop` | Template key sent to Notification Service. | Must exist in Notification Service templates. |
| `WISHLIST_PRICE_DROP_BATCH_SIZE` | Optional | `500` | Mongo cursor batch size for candidates. | Keep positive; too high can increase memory pressure. |
| `WISHLIST_PRICE_DROP_MIN_DELTA_AMOUNT` | Optional | `1` | Minimum price drop in minor units. | Use `1` for any drop, larger value to suppress tiny drops. |
| `WISHLIST_ANALYTICS_EVENTS_ENABLED` | Optional | `false` for Task 7-only test | Avoids unrelated outbox setup/publisher. | Enable only when analytics setup is ready. |
| `WISHLIST_EVENT_PUBLISHER` | Optional | `disabled` for Task 7-only test | Avoids unrelated analytics Kafka publisher. | Use `kafka` only with broker ready. |

### Common Task 7 Env Mistakes

| Mistake | Result | Fix |
|---|---|---|
| `.env` edited but not sourced | Service keeps old/default values. | Run `set -a; source .env; set +a`. |
| `WISHLIST_EVENTS_BACKEND=disabled` | Price events are not consumed. | Set `WISHLIST_EVENTS_BACKEND=kafka` for Task 7 testing. |
| `WISHLIST_EVENTS_BACKEND=rabbitmq` | Startup validation fails. | Use `kafka` or `disabled`. |
| Kafka broker uses HTTP scheme | Kafka connection fails. | Use `localhost:9092`, not `http://localhost:9092`. |
| Product Service publishes to RabbitMQ | `SERVICE_NAME` never receives event. | Align Product event broker to Kafka or add a bridge. |
| `notification.commands` topic missing | Publish can fail or topic auto-create hides typo. | Create topic explicitly. |
| Template key differs | Notification Service cannot render expected message. | Keep `WISHLIST_PRICE_DROP_TEMPLATE_KEY` aligned with Notification Service. |

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

Reuse event-mode topic setup pattern:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task6_Dependency.md`

Section:
`8. Docker Setup`
```

Task 7 Docker-specific notes:

| Docker item | Status | Meaning |
|---|---|---|
| MongoDB container | Reused | Required if MongoDB is not installed locally. |
| Redpanda/Kafka container | Reused and required for Task 7 event mode | Needed when `WISHLIST_EVENTS_BACKEND=kafka`. |
| Product Service container | External dependency | Must publish `ProductPriceChanged` to Kafka for end-to-end flow. |
| Notification Service container | External dependency | Must consume `notification.commands` for actual delivery. |
| RabbitMQ container | Not used by current `SERVICE_NAME` code | Only relevant if another service or bridge uses it. |
| Service Dockerfile | Not found for this service in current snapshot | Local service is run with `go run`. |

### Topic Creation With Redpanda Container

If you used the previous guide's Redpanda container named `wishlist-redpanda`, create Task 7 topics:

```bash
docker exec wishlist-redpanda rpk topic create product.events
docker exec wishlist-redpanda rpk topic create product.events.wishlist.price.dlq
docker exec wishlist-redpanda rpk topic create notification.commands
```

Verify:

```bash
docker exec wishlist-redpanda rpk topic list
```

Beginner note: Some local brokers auto-create topics, but explicit topic creation is better because typo issues become visible early.

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
TaskImplementation/{SERVICE_NAME}/task6_Dependency.md
```

Minimum sections:

- `task1_Dependency.md` -> `3. Required Software`
- `task1_Dependency.md` -> `4. Dependency Management`
- `task1_Dependency.md` -> `5. Database Setup`
- `task1_Dependency.md` -> `6. Redis / Queue / External Services`
- `task1_Dependency.md` -> `7. Environment Variables`
- `task1_Dependency.md` -> `8. Docker Setup`
- `task3_Dependency.md` -> `5. Database Setup`
- `task6_Dependency.md` -> `6. Redis / Queue / External Services`
- `task6_Dependency.md` -> `8. Docker Setup`

### Step 2: Go to Project Directory

```bash
cd Ecommerce
cd backend/services/wishlist-service
```

### Step 3: Install Only New Dependencies

None for Task 7.

If Go modules were never downloaded:

```bash
go mod download
```

### Step 4: Setup Only New Databases/Services

No new database server. Use MongoDB setup from previous dependency docs.

Task 7 needs Kafka readiness:

```bash
docker ps
docker exec wishlist-redpanda rpk cluster info
docker exec wishlist-redpanda rpk topic list
```

Expected topics:

```text
product.events
product.events.wishlist.price.dlq
notification.commands
```

### Step 5: Add Only New or Changed Environment Variables

Update:

```text
backend/services/wishlist-service/.env
```

Use the Task 7 env delta from section 7.

Load env:

```bash
set -a
source .env
set +a
```

### Step 6: Run Migrations If Needed

If this is a fresh local DB, use previous full migration setup:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`9. Local Development Setup`
```

Task 7 minimum index migration:

```bash
cd backend/services/wishlist-service
mongosh "mongodb://localhost:27017" migrations/mongo/003_add_price_drop_candidate_index.up.js
```

Verify:

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.wishlists.getIndexes()"
```

Expected index:

```text
idx_wishlists_price_drop_candidates
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

### Step 8: Verify Functionality Related to `TASK_FILE_NAME`

1. Seed a local wishlist item with old price:

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval '
db.wishlists.updateOne(
  { user_id: "user_123" },
  {
    $setOnInsert: {
      _id: "wish_123",
      user_id: "user_123",
      visibility: "private",
      created_at: new Date(),
      updated_at: new Date()
    },
    $set: {
      items: [{
        product_id: "prod_123",
        variant_id: "var_1",
        added_at: new Date(),
        last_known_price: { amount: NumberLong(349900), currency: "INR" },
        availability: "in_stock"
      }],
      updated_at: new Date()
    }
  },
  { upsert: true }
)'
```

2. Publish a Product price-drop event:

```bash
EVENT_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
printf '{"event_id":"evt_task7_001","event_type":"ProductPriceChanged","version":1,"occurred_at":"%s","producer":"product-service","trace_id":"trace_task7_001","payload":{"product_id":"prod_123","variant_id":"var_1","new_price":{"amount":299900,"currency":"INR"},"title":"Running Shoes","product_url":"/products/prod_123","image_url":"https://cdn.example.com/prod_123/main.jpg"}}\n' "$EVENT_TIME" \
  | docker exec -i wishlist-redpanda rpk topic produce product.events
```

3. Verify notification command was published:

```bash
docker exec wishlist-redpanda rpk topic consume notification.commands -n 1
```

Expected command fields:

```text
event_type = SendNotification
payload.user_id = user_123
payload.template_key = wishlist_price_drop
payload.idempotency_key contains wishlist_price_drop:user_123:prod_123:var_1:INR:299900
```

4. Verify MongoDB snapshot updated:

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval '
db.wishlists.findOne(
  { user_id: "user_123", "items.product_id": "prod_123" },
  { user_id: 1, "items.$": 1 }
)'
```

Expected:

```text
items[0].last_known_price.amount = 299900
items[0].last_known_price.currency = INR
```

---

## 10. Running the Project

### Minimum Local Run Without Task 7 Events

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

### Task 7 Event-Enabled Run

Required:

- MongoDB running.
- Kafka-compatible broker running.
- Topics created.
- `idx_wishlists_price_drop_candidates` index exists.
- Product event producer publishing valid `ProductPriceChanged` messages to Kafka.
- Notification Service consumer ready, or Kafka CLI ready for local command verification.
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

### Ports & Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| `SERVICE_NAME` HTTP server | `8084` | Health and wishlist APIs | Reused |
| MongoDB | `27017` | `wishlist_db.wishlists` persistence | Reused |
| Kafka / Redpanda | `9092` | Product price events, DLQ, notification commands | Reused but required for Task 7 event mode |
| Product Service HTTP | `8082` | Product APIs; event producer may be separate worker | Reused |
| Notification Service gRPC | `9090` | Internal notification API from current `.env` | External dependency |
| Notification analytics/callback HTTP | `8081` | Notification metrics/webhooks from current `.env` | External, not required for Wishlist publisher verification |
| RabbitMQ | `5672` | Present in Product/Notification env files | Not consumed by current `SERVICE_NAME` runtime |
| Mailpit SMTP | `1025` | Local email provider if Notification Service enables email | Notification Service only |

Port conflict examples:

```bash
lsof -i :8084
lsof -i :9092
```

Change service HTTP port:

```bash
WISHLIST_HTTP_ADDR=:18084 go run ./cmd/server
```

Change Kafka broker address:

```env
WISHLIST_KAFKA_BROKERS=localhost:<new-port>
```

Docker networking:

| Where `SERVICE_NAME` runs | Mongo URI | Kafka broker |
|---|---|---|
| Host machine | `mongodb://localhost:27017` | `localhost:9092` |
| Docker Compose container | `mongodb://wishlist-mongo:27017` | `wishlist-redpanda:9092` |

Beginner note: `localhost` inside a container means same container. Dusre container ko call karna hai to Compose service name use karo.

---

## 11. Common Errors & Fixes

Generic Go, MongoDB, `.env`, Docker daemon, and port `8084` issues are already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`12. Common Errors & Fixes`
```

Kafka availability event errors are already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task6_Dependency.md`

Section:
`12. Common Errors & Fixes`
```

Task 7-specific issues:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| No price-drop processing happens | `WISHLIST_EVENTS_BACKEND=disabled`. | Set `WISHLIST_EVENTS_BACKEND=kafka`, load env, restart service. | Keep separate local env profile for event mode. |
| `WISHLIST_EVENTS_BACKEND must be disabled or kafka` | Env value is `rabbitmq` or typo. | Use `kafka` or `disabled`. | Document allowed values in `.env.example`. |
| Service starts but no events arrive | Product Service publishes to RabbitMQ while `SERVICE_NAME` consumes Kafka. | Align Product event broker to Kafka or add a bridge. | Contract-test broker/topic config between services. |
| Event goes to DLQ with `invalid_new_price` | `new_price.amount` negative or currency invalid. | Send non-negative minor-unit amount and uppercase currency. | Producer schema validation. |
| Event goes to DLQ with `unexpected_producer` | `producer` is not `product-service`. | Set producer exactly to `product-service`. | Use shared event envelope helper. |
| Event ignored | `event_type` is not `ProductPriceChanged`, `ProductDeleted`, or `ProductOutOfStock`. | Publish correct event type. | Keep event names exact and versioned. |
| No notification command for test product | No matching wishlist item, price was not lower, currency mismatch, deleted item, or variant mismatch. | Verify Mongo item and event payload. | Seed known-good fixtures. |
| Snapshot updated but no user delivery | Command published, but Notification Service consumer/template/provider is missing. | Consume `notification.commands` to verify command; configure Notification Service. | Add Notification Service readiness and template checks. |
| `notification commands topic is required` | Kafka backend enabled but `WISHLIST_NOTIFICATION_COMMANDS_TOPIC` blank. | Set `WISHLIST_NOTIFICATION_COMMANDS_TOPIC=notification.commands`. | Do not leave blank env vars. |
| Duplicate notification commands after retry | Snapshot update failed after publish, so event retried. | Notification Service must dedupe by `idempotency_key`. | Treat idempotency key as required downstream contract. |
| Tiny price drops trigger too many messages | `WISHLIST_PRICE_DROP_MIN_DELTA_AMOUNT=1`. | Increase min delta, for example `100` for INR 1 if amount is paise. | Decide business threshold before production. |
| DLQ grows silently | No DLQ monitoring. | Consume/alert on `product.events.wishlist.price.dlq`. | Add DLQ dashboards and runbook. |

---

## 12. Security & Best Practices

### Security and Configuration Audit

| Area | Current observation | Risk | Suggested fix |
|---|---|---|---|
| Cross-service DB access | `SERVICE_NAME` does not read Notification DB or Product DB directly. | Good microservice boundary. | Keep it that way; use events/API contracts. |
| Local `.env` files | Repo contains checked-in local `.env` files, including local DB/RabbitMQ credentials in other services. | Fine for local examples, unsafe for real secrets. | Add `.env.example`, keep real secrets outside git. |
| Wishlist Mongo default | `WISHLIST_MONGO_URI=mongodb://localhost:27017` has no auth. | Fine local, weak shared/prod default. | Use authenticated Mongo URI and network restrictions in non-local env. |
| Kafka transport | Local defaults use plaintext `localhost:9092`. | Unsafe in shared/prod. | Use TLS/SASL/ACLs or managed Kafka. |
| Event producer trust | Consumer validates `producer=product-service`. | Not cryptographic. | Enforce broker ACLs so only Product Service can write `product.events`. |
| DLQ body | DLQ stores original message body. | Product metadata can leak to operators with broad access. | Keep event payload minimal and restrict DLQ access. |
| Notification payload | Command contains user id, product id, title, image URL, product URL, and price values. | Avoid PII exposure. | Do not include email/phone/name/address in Wishlist event payloads. |
| Idempotency | Publisher sends stable `idempotency_key`. | Duplicate delivery still possible if Notification Service ignores it. | Notification Service must enforce dedupe. |
| Template key | `wishlist_price_drop` must exist downstream. | Missing template means command accepted but no useful delivery. | Seed template as part of Notification Service deployment. |
| Health checks | `/readyz` checks MongoDB, not Kafka or Notification Service. | Service can look ready while event flow is broken. | Add Kafka consumer/lag metrics and downstream synthetic checks. |
| Metrics | Product event metrics default to no-op implementation. | Production observability incomplete. | Wire Prometheus/OpenTelemetry implementation. |

### Task 7 Best Practices

- Use `product_id` as Kafka message key so same product events stay ordered within a partition.
- Keep `event_id`, `trace_id`, and `occurred_at` in every Product event.
- Use minor units for money values, such as paise/cents, not floating-point rupees/dollars.
- Keep `WISHLIST_PRICE_DROP_MIN_DELTA_AMOUNT` aligned with business threshold.
- Publish notification command before updating `last_known_price`; current usecase already follows this safer order.
- Monitor `product.events.wishlist.price.dlq`.
- Monitor `notification.commands` lag because Notification Service delay directly affects user notification timing.
- Never put Notification provider credentials inside `SERVICE_NAME`.
- Keep `WISHLIST_EVENTS_BACKEND=disabled` for beginner API-only local testing.
- Use `WISHLIST_EVENTS_BACKEND=kafka` only after broker and topics are ready.

---

## 13. Missing or Misconfigured Things

| Missing/misaligned item | Why it matters | Recommended action |
|---|---|---|
| Product Service `.env` currently says `PRODUCT_EVENT_BROKER=rabbitmq` | Current `SERVICE_NAME` consumes Kafka only, so Task 7 events will not arrive automatically. | Align Product Service to Kafka or add RabbitMQ-to-Kafka bridge before end-to-end testing. |
| Notification Service `.env` does not show Kafka `notification.commands` consumer config | Commands may be published but never delivered to users. | Implement/configure Notification Service consumer for `notification.commands` or bridge to its supported queue/API. |
| No service Dockerfile found for current `SERVICE_NAME` runtime | Containerized local/prod run is manual. | Add Dockerfile for `backend/services/wishlist-service`. |
| No local compose file found in repo snapshot | Mongo/Kafka setup depends on manual commands. | Add a repo-level local compose with MongoDB, Kafka/Redpanda, and optional Notification Service. |
| No `.env.example` found for `SERVICE_NAME` | Beginners may edit committed `.env` or miss event-mode values. | Add `backend/services/wishlist-service/.env.example` with safe dummy values. |
| JS migrations hardcode `wishlist_db` and `wishlists` | Runtime env can point to another DB/collection without expected index. | Keep defaults aligned locally or make migrations parameterized later. |
| `/readyz` does not prove Kafka/Notification flow | Deployments may look healthy while price-drop flow is broken. | Add event consumer health/lag checks and notification command synthetic test. |
| Metrics are no-op by default | Task 7 processing is not observable without logs. | Wire real metrics backend. |
| Notification template seed not found in `SERVICE_NAME` setup | `wishlist_price_drop` may not exist downstream. | Add Notification Service template migration/seed. |
| Shared event schema not in `api/master-api.json` | Product/Wishlist contract can drift. | Add event schemas or contract tests for `ProductPriceChanged` and `SendNotification`. |

---

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Required Software` | Git, Go, Docker, MongoDB, `mongosh`, curl setup already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Dependency Management` | Go modules, `go mod download`, `go test`, `go run`, and common Go issues already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Setup` | MongoDB install, URI, credentials placement, migrations, and verification already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Redis / Queue / External Services` | Base Kafka explanation and queue/topic overview already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Environment Variables` | Full service `.env` reference already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Docker Setup` | Base MongoDB/Kafka Docker dependency setup already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `11. Ports & Networking` | Standard ports and Docker networking concepts already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `12. Common Errors & Fixes` | Generic Go, MongoDB, Docker, `.env`, and port troubleshooting already documented. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | MongoDB ownership decision | Same database ownership remains unchanged. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `5. Database Setup` | `wishlists` schema, validator, `last_known_price`, `availability`, and price-drop index already documented. |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | Product Service HTTP setup | Product Service base dependency remains reused for add-item flow, though Task 7 uses events. |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | Cart Service setup | Cart Service remains runtime config but is not Task 7-specific. |
| `TaskImplementation/{SERVICE_NAME}/task6_Dependency.md` | Kafka product-event consumer, DLQ, topic creation, event-mode runbook | Task 7 reuses the same Kafka consumer runtime and retry/DLQ behavior. |

---

## 15. Final Checklist

- [ ] Previous dependency documentation checked.
- [ ] No duplicated Go/Mongo/Docker installation guide added.
- [ ] No new Go dependency added for Task 7.
- [ ] MongoDB running and `wishlists` collection ready.
- [ ] `idx_wishlists_price_drop_candidates` index exists.
- [ ] Kafka/Redpanda running for event-enabled mode.
- [ ] `product.events` topic exists.
- [ ] `product.events.wishlist.price.dlq` topic exists.
- [ ] `notification.commands` topic exists.
- [ ] Product Service event broker aligned with `SERVICE_NAME` Kafka consumer, or bridge planned.
- [ ] Notification Service consumer/template support confirmed, or Kafka CLI verification planned.
- [ ] `WISHLIST_EVENTS_BACKEND=kafka` set only when Kafka is ready.
- [ ] `WISHLIST_KAFKA_BROKERS` points to the correct broker address.
- [ ] Price-event group/DLQ aliases are not conflicting.
- [ ] `WISHLIST_NOTIFICATION_COMMANDS_TOPIC` matches downstream consumer.
- [ ] `WISHLIST_PRICE_DROP_TEMPLATE_KEY=wishlist_price_drop` matches Notification Service template.
- [ ] `WISHLIST_PRICE_DROP_MIN_DELTA_AMOUNT` matches business threshold.
- [ ] `.env` loaded into shell before running service.
- [ ] `go test ./internal/usecase ./internal/events ./internal/repository ./internal/config` passes.
- [ ] Backend service starts successfully.
- [ ] `/healthz` and `/readyz` checked.
- [ ] Test wishlist item has old `last_known_price`.
- [ ] `ProductPriceChanged` event published to `product.events`.
- [ ] `SendNotification` command observed on `notification.commands`.
- [ ] `last_known_price` snapshot updated in MongoDB.
- [ ] Invalid event lands in DLQ and normal consumer continues.
- [ ] Logs checked for `wishlist price drop processed`.
- [ ] No duplicate setup documentation added.
