# Project Dependency & Setup Guide

Source implementation file:

```text
TaskImplementation/Search Service/task3.md
```

Generated dependency file:

```text
TaskImplementation/Search Service/task3_Dependency.md
```

## 1. Project Overview

Task 3 ka focus Product Indexer hai. Simple Hinglish me: Product Service product change event publish karega, Search Service RabbitMQ se event consume karega, payload validate karega, Redis me duplicate event check karega, aur Typesense ke `products` collection me document upsert/delete karega.

Current implementation already exists under:

```text
backend/services/search-service
```

Task 3 introduces no new SQL database, no migration, and no new Docker container. It activates and depends on the existing queue, Redis, and Typesense setup.

| Area | Task 3 Decision |
|---|---|
| Input file | `TaskImplementation/Search Service/task3.md` |
| Runtime service | `backend/services/search-service` |
| Main external dependency | RabbitMQ product event queue |
| Search index dependency | Typesense `products` collection |
| Idempotency dependency | Redis processed-event keys |
| Default queue provider | RabbitMQ |
| Kafka support | Design mention only. Current code validates `QUEUE_PROVIDER=rabbitmq` |
| SQL migrations | Not required |

Reuse rule: previous dependency guides already explain common setup. This file only documents Task 3 specific setup and verification.

## 2. Tech Stack

### Reused tech from previous dependency files

| Technology | Required? | Status | Where already explained |
|---|---:|---|---|
| Go | Yes | Reused | `TaskImplementation/Search Service/task1_Dependency.md`, Section `2. Go Dependency System` |
| Go modules | Yes | Reused | `TaskImplementation/Search Service/task1_Dependency.md`, Section `2. Go Dependency System` |
| Typesense | Yes | Reused | `TaskImplementation/Search Service/task2_Dependency.md`, Section `4.1 Typesense` |
| Redis | Yes when indexer enabled | Reused, Task 3 uses processed-event keys | `TaskImplementation/Search Service/task1_Dependency.md`, Sections `3.2 Redis` and `5.2 Redis` |
| RabbitMQ | Yes by default | Reused, Task 3 uses product event topology | `TaskImplementation/Search Service/task1_Dependency.md`, Section `5.3 RabbitMQ` |
| Docker Compose | Recommended local infra | Reused | `TaskImplementation/Search Service/task1_Dependency.md`, Section `7. Docker and DevOps Setup` |
| Kubernetes ConfigMap/Secret pattern | Optional/prod-like | Reused | `TaskImplementation/Search Service/task2_Dependency.md`, Section `8.2 Kubernetes Typesense Setup` |

### Task 3 specific tech details

| Technology / Package | Required? | Why Used | Beginner Explanation |
|---|---:|---|---|
| `github.com/rabbitmq/amqp091-go` | Yes | RabbitMQ AMQP consumer banane ke liye | Ye Go library Search Service ko RabbitMQ queue se messages receive, ack, nack, retry, aur DLQ publish karne me help karti hai. |
| `github.com/typesense/typesense-go/v2` | Yes | Product document upsert/delete ke liye | Search Service manually HTTP calls nahi banata; Typesense Go client se document write clean hota hai. |
| Custom Redis client | Yes | Processed event id store karne ke liye | Is repo me Redis ke liye external Go package nahi use hua. Code simple Redis protocol commands use karta hai. |
| RabbitMQ Management UI | Optional but helpful | Local queue/exchange/DLQ debugging | Browser me exchanges, queues, bindings, message count dekh sakte ho. Beginners ke liye debugging easy hoti hai. |

Task 3 me new Go package install karne ki zarurat nahi hai, because required packages already exist in:

```text
backend/services/search-service/go.mod
```

## 3. Required Software

Follow the base setup first:

```text
TaskImplementation/Search Service/task1_Dependency.md
```

Sections:

```text
2. Go Dependency System
3. Database and Storage Analysis
5. External Services Analysis
7. Docker and DevOps Setup
8. Project Run Instructions
```

Task 3 only needs you to confirm these software/services are available:

| Software / Service | Required for Task 3? | Version / Default in Repo | Verification |
|---|---:|---|---|
| Go | Yes | `go 1.26.3` in `go.mod` | `go version` |
| Docker + Docker Compose | Recommended | Local infra runner | `docker version` |
| Typesense | Yes | `typesense/typesense:27.1` | `curl http://localhost:8108/health` |
| Redis | Yes if `SEARCH_INDEXER_ENABLED=true` | `redis:7.4-alpine` | `docker exec ecommerce-redis redis-cli -a dev_redis_password ping` |
| RabbitMQ | Yes if `SEARCH_INDEXER_ENABLED=true` | `rabbitmq:3.13-management-alpine` | `docker exec ecommerce-rabbitmq rabbitmq-diagnostics -q ping` |

## 4. Dependency Management

This is a Go project. Go module setup is already fully documented in:

```text
TaskImplementation/Search Service/task1_Dependency.md
```

Section:

```text
2. Go Dependency System
```

Task 3 specific dependency audit:

| File | Current Finding |
|---|---|
| `backend/services/search-service/go.mod` | Already includes RabbitMQ and Typesense Go clients |
| `backend/services/search-service/go.sum` | Already contains checksums for downloaded packages |
| Redis dependency | No external `go-redis` dependency. Custom Redis client is implemented in `internal/repository/redis_client.go` |
| Node/Python dependencies | Not used for this task |

Use the same commands from Task 1:

```bash
cd backend/services/search-service
go mod download
go test ./...
```

Run `go mod tidy` only when imports or dependency versions intentionally change.

## 5. Database Setup

Task 3 does not introduce MySQL, PostgreSQL, MongoDB, SQLite, Cassandra, or SQL migrations.

| Storage / DB | Used in Task 3? | Required? | Task 3 Role | Setup Decision |
|---|---:|---:|---|---|
| Typesense | Yes | Yes | Product document upsert/delete | Reuse Task 2 setup |
| Redis | Yes | Yes if indexer enabled | Processed event idempotency store | Reuse Task 1 setup |
| RabbitMQ disk volume | Yes indirectly | Yes if indexer enabled | Durable queue/DLQ state | Reuse Task 1 setup |
| MySQL/PostgreSQL/MongoDB | No | No | Product data source belongs to Product Service | No setup for this task |

### Typesense usage in Task 3

Typesense setup is already explained in:

```text
TaskImplementation/Search Service/task2_Dependency.md
```

Section:

```text
4.1 Typesense
```

Task 3 uses the existing `products` collection:

| Operation | When It Happens |
|---|---|
| Upsert document | `ProductPublished`, searchable `ProductUpdated`, `ProductPriceChanged`, `ProductInventoryChanged` |
| Delete document | `ProductDeleted`, `ProductUnpublished`, `ProductBlocked`, or non-searchable payload |

Connection variables are reused:

```env
TYPESENSE_HOST=localhost
TYPESENSE_PORT=8108
TYPESENSE_PROTOCOL=http
TYPESENSE_API_KEY=dev-typesense-key
TYPESENSE_PRODUCTS_COLLECTION=products
```

### Redis usage in Task 3

Redis setup is already explained in:

```text
TaskImplementation/Search Service/task1_Dependency.md
```

Sections:

```text
3.2 Redis
5.2 Redis
```

Task 3 specific Redis keys:

| Key Pattern | Example | Purpose | TTL |
|---|---|---|---|
| `search:indexer:processed:{event_id}` | `search:indexer:processed:evt_123` | Duplicate product event avoid karna | `PRODUCT_INDEXER_PROCESSED_EVENT_TTL`, default `720h` |

Important behavior:

- Consumer pehle Redis me check karta hai ki `event_id` already processed hai ya nahi.
- Duplicate event aane par message ack hota hai, Typesense me repeat write nahi hoti.
- Redis unavailable hua to consumer startup fail kar sakta hai, because Task 3 idempotency ke liye Redis required hai.

## 6. Redis / Queue / External Services

### RabbitMQ product event topology

RabbitMQ basic setup is already explained in:

```text
TaskImplementation/Search Service/task1_Dependency.md
```

Section:

```text
5.3 RabbitMQ
```

Task 3 specific topology:

| RabbitMQ Item | Value | Meaning |
|---|---|---|
| Exchange | `product.events` | Product Service yahan product lifecycle events publish karega |
| Exchange type | `topic` | Routing key ke basis par queues bind hoti hain |
| Consumer queue | `search-service.product-indexer` | Search Service ka isolated product indexer queue |
| Dead-letter exchange | `product.events.dlx` | Failed messages ke liye exchange |
| Dead-letter queue | `search-service.product-indexer.dlq` | Invalid/max-retry messages yahan inspect honge |
| Consumer tag | `search-service-product-indexer` | RabbitMQ UI me consumer identify karne ke liye |
| Prefetch | `20` | Ek time par max 20 unacked messages |
| Max retries | `5` | Iske baad DLQ |

Routing keys:

```text
product.published
product.updated
product.price_changed
product.inventory_changed
product.unpublished
product.deleted
product.blocked
```

The Go consumer declares the exchange, queue, DLX, DLQ, and bindings automatically on startup. Beginner rule: agar queue RabbitMQ UI me nahi dikh rahi, pehle Search Service start karo.

### Retry and DLQ behavior

| Failure Type | Example | Behavior |
|---|---|---|
| Permanent | Invalid JSON, missing `event_id`, unsupported event type, bad payload | Publish to DLQ and ack original message |
| Transient | Typesense timeout, Redis temporary issue, RabbitMQ publish retry issue | Retry with backoff |
| Duplicate | Same `event_id` already in Redis | Ack without re-indexing |
| Max retry crossed | Same transient error after retry limit | Publish to DLQ and ack original message |

Task 3 retry headers:

| Header | Purpose |
|---|---|
| `x-search-indexer-retry-count` | Current retry attempt count |
| `x-search-indexer-last-error` | Last retry error summary |
| `x-search-indexer-retried-at` | Retry timestamp |
| `x-search-indexer-failure-reason` | DLQ failure reason |
| `x-search-indexer-error` | DLQ error summary |
| `x-original-exchange` | Original source exchange |
| `x-original-routing-key` | Original routing key |

### Kafka note

`task3.md` mentions Kafka as an alternate design, but current code is RabbitMQ only. Do not set:

```env
QUEUE_PROVIDER=kafka
```

Current config validation will fail with:

```text
QUEUE_PROVIDER must be rabbitmq for the product indexer
```

## 7. Environment Variables

Full Search Service `.env` setup is already documented in:

```text
TaskImplementation/Search Service/task1_Dependency.md
```

Section:

```text
4. Environment Variables
```

Task 3 introduces no new env variable names beyond the already documented Search Service config. For Product Indexer, confirm this subset exists in:

```text
backend/services/search-service/.env
```

For local host `go run`:

```env
SEARCH_INDEXER_ENABLED=true
PRODUCT_INDEXER_MESSAGE_TIMEOUT=30s
PRODUCT_INDEXER_PROCESSED_EVENT_TTL=720h

QUEUE_PROVIDER=rabbitmq
RABBITMQ_URL=amqp://ecommerce:ecommerce_password@localhost:5672/ecommerce
PRODUCT_EVENTS_EXCHANGE=product.events
PRODUCT_INDEXER_QUEUE=search-service.product-indexer
PRODUCT_INDEXER_DLX=product.events.dlx
PRODUCT_INDEXER_DLQ=search-service.product-indexer.dlq
PRODUCT_INDEXER_CONSUMER_TAG=search-service-product-indexer
PRODUCT_INDEXER_ROUTING_KEYS=product.published,product.updated,product.price_changed,product.inventory_changed,product.unpublished,product.deleted,product.blocked
PRODUCT_INDEXER_PREFETCH=20
PRODUCT_INDEXER_MAX_RETRIES=5
PRODUCT_INDEXER_RECONNECT_DELAY=5s
PRODUCT_INDEXER_RETRY_BASE_DELAY=1s
PRODUCT_INDEXER_RETRY_MAX_DELAY=5m

SEARCH_REDIS_ADDR=localhost:6379
SEARCH_REDIS_PASSWORD=dev_redis_password
SEARCH_REDIS_DB=0

TYPESENSE_PRODUCTS_COLLECTION=products
TYPESENSE_API_KEY=dev-typesense-key
```

Variable notes:

| Variable | Required? | Task 3 Purpose | Security Notes |
|---|---:|---|---|
| `SEARCH_INDEXER_ENABLED` | Optional | Enables product event consumer | In prod this should normally be `true` |
| `RABBITMQ_URL` | Yes if indexer enabled | Connects to RabbitMQ | Secret because password is inside URL |
| `PRODUCT_EVENTS_EXCHANGE` | Yes if indexer enabled | Product event exchange name | No secret |
| `PRODUCT_INDEXER_QUEUE` | Yes if indexer enabled | Consumer queue name | No secret |
| `PRODUCT_INDEXER_DLX` | Yes if indexer enabled | Dead-letter exchange | No secret |
| `PRODUCT_INDEXER_DLQ` | Yes if indexer enabled | Dead-letter queue | DLQ can contain product data, restrict access |
| `PRODUCT_INDEXER_ROUTING_KEYS` | Yes if indexer enabled | Which product events this consumer receives | Must match Product Service publisher |
| `PRODUCT_INDEXER_PROCESSED_EVENT_TTL` | Yes if indexer enabled | Redis idempotency TTL | Longer TTL reduces duplicate writes |
| `SEARCH_REDIS_PASSWORD` | Required if Redis auth enabled | Auth for Redis processed-event store | Secret |
| `TYPESENSE_API_KEY` | Yes | Typesense write access | Secret, never expose to frontend |

Where credentials go:

| Environment | Place |
|---|---|
| Local Go run | `backend/services/search-service/.env` or exported shell env |
| Local Compose infra | `infra/compose/.env.local` for container passwords/API keys |
| Kubernetes | `infra/k8s/core/search-service/secret.example.yaml` shape, real values from secret manager |

Important: Go code uses `os.Getenv`. It does not auto-load `.env`. Load it before starting:

```bash
cd backend/services/search-service
set -a
. ./.env
set +a
go run ./cmd/server
```

## 8. Docker Setup

Docker Compose setup is reused from:

```text
TaskImplementation/Search Service/task1_Dependency.md
```

Section:

```text
7. Docker and DevOps Setup
```

Task 3 needs these existing Compose services:

| Compose Service | Image | Why Required |
|---|---|---|
| `typesense` | `typesense/typesense:27.1` | Product search index writes |
| `redis` | `redis:7.4-alpine` | Processed event idempotency |
| `rabbitmq` | `rabbitmq:3.13-management-alpine` | Product event consumer queue and DLQ |

Start only the Task 3 infra:

```bash
cp infra/compose/.env.local.example infra/compose/.env.local
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d typesense redis rabbitmq
```

Check containers:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml ps typesense redis rabbitmq
```

View logs:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml logs -f rabbitmq
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml logs -f redis
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml logs -f typesense
```

Do not use `docker compose down -v` casually. It deletes local queue/index/cache volumes.

### Ports and networking

Task 3 introduces no new port. It uses existing ports:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Search Service HTTP | `8085` | Health/API server | Reused |
| Typesense | `8108` | Product document write/read | Reused |
| Redis | `6379` | Processed-event key store | Reused, required for indexer |
| RabbitMQ AMQP | `5672` | Product event consumer connection | Reused, required for indexer |
| RabbitMQ Management UI | `15672` | Queue/exchange/DLQ debugging | Reused |
| Product Service | `8082` | Source of truth, not called during normal event upsert if event has full snapshot | Reused by broader Search Service |

Host vs Docker network values:

| Where Search Service Runs | RabbitMQ URL | Redis Addr | Typesense Host |
|---|---|---|---|
| Host machine with `go run` | `amqp://ecommerce:ecommerce_password@localhost:5672/ecommerce` | `localhost:6379` | `localhost` |
| Inside Docker Compose network | `amqp://ecommerce:ecommerce_password@rabbitmq:5672/ecommerce` | `redis:6379` | `typesense` |
| Kubernetes | `amqp://...@rabbitmq.data.svc.cluster.local:5672/ecommerce` | `redis.data.svc.cluster.local:6379` | `typesense.data.svc.cluster.local` |

## 9. Local Development Setup

### Step 1: Read previous dependency documentation first

Follow these before this file:

```text
TaskImplementation/Search Service/task1_Dependency.md
TaskImplementation/Search Service/task2_Dependency.md
```

### Step 2: Go to project directory

```bash
cd /home/parag/Ecommerce
```

### Step 3: Start Task 3 infrastructure

```bash
cp infra/compose/.env.local.example infra/compose/.env.local
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d typesense redis rabbitmq
```

### Step 4: Verify infrastructure

Typesense:

```bash
curl http://localhost:8108/health
```

Redis:

```bash
docker exec ecommerce-redis redis-cli -a dev_redis_password ping
```

RabbitMQ:

```bash
docker exec ecommerce-rabbitmq rabbitmq-diagnostics -q ping
curl -u ecommerce:ecommerce_password http://localhost:15672/api/overview
```

### Step 5: Install Go dependencies

```bash
cd backend/services/search-service
go mod download
```

### Step 6: Configure local `.env`

Use the complete `.env` from:

```text
TaskImplementation/Search Service/task1_Dependency.md
```

Section:

```text
4. Environment Variables
```

Confirm Task 3 values:

```env
SEARCH_INDEXER_ENABLED=true
RABBITMQ_URL=amqp://ecommerce:ecommerce_password@localhost:5672/ecommerce
SEARCH_REDIS_ADDR=localhost:6379
SEARCH_REDIS_PASSWORD=dev_redis_password
TYPESENSE_HOST=localhost
TYPESENSE_API_KEY=dev-typesense-key
```

### Step 7: Run tests

```bash
go test ./...
```

### Step 8: Start backend service

```bash
set -a
. ./.env
set +a
go run ./cmd/server
```

Expected useful log lines:

```text
search.http.started
search.product_consumer.started
```

If you only want API boot without Task 3 consumer:

```env
SEARCH_INDEXER_ENABLED=false
```

That is useful for local debugging, but it does not verify Task 3.

## 10. Running the Project

### Verify health

```bash
curl http://localhost:8085/healthz
```

### Verify RabbitMQ topology

After Search Service starts, open:

```text
http://localhost:15672
```

Login:

```text
username: ecommerce
password: ecommerce_password
```

Check:

| RabbitMQ UI Area | Expected |
|---|---|
| Exchanges | `product.events`, `product.events.dlx` |
| Queues | `search-service.product-indexer`, `search-service.product-indexer.dlq` |
| Bindings | Product routing keys bound to `search-service.product-indexer` |
| Consumers | `search-service-product-indexer` visible on main queue |

### Publish one product event for smoke test

Beginner-friendly path: RabbitMQ UI me `Exchanges` -> `product.events` -> `Publish message` open karo.

Routing key:

```text
product.published
```

Payload:

```json
{
  "event_id": "evt_task3_local_001",
  "event_type": "ProductPublished",
  "version": 1,
  "producer": "product-service",
  "request_id": "req_task3_local_001",
  "trace_id": "trace_task3_local_001",
  "correlation_id": "prod_task3_001",
  "occurred_at": "2026-06-05T00:00:00Z",
  "payload": {
    "product_id": "prod_task3_001",
    "title": "Task 3 Test Shoes",
    "description": "Local event indexing smoke test",
    "brand": "Demo",
    "category_ids": ["cat_shoes"],
    "seller_id": "seller_demo",
    "price": 2499,
    "rating": 4.5,
    "popularity_score": 10,
    "in_stock": true,
    "status": "published",
    "is_deleted": false,
    "created_at": "2026-06-05T00:00:00Z",
    "updated_at": "2026-06-05T00:00:00Z"
  }
}
```

### Verify Typesense document

```bash
curl -H "X-TYPESENSE-API-KEY: dev-typesense-key" \
  http://localhost:8108/collections/products/documents/prod_task3_001
```

Expected: response contains `prod_task3_001`, title, brand, price, and category ids.

### Verify Redis idempotency key

```bash
docker exec ecommerce-redis redis-cli -a dev_redis_password \
  EXISTS search:indexer:processed:evt_task3_local_001
```

Expected:

```text
1
```

### Verify delete path

Publish another event with routing key:

```text
product.deleted
```

Payload:

```json
{
  "event_id": "evt_task3_local_002",
  "event_type": "ProductDeleted",
  "version": 1,
  "producer": "product-service",
  "request_id": "req_task3_local_002",
  "trace_id": "trace_task3_local_002",
  "correlation_id": "prod_task3_001",
  "occurred_at": "2026-06-05T00:01:00Z",
  "payload": {
    "product_id": "prod_task3_001",
    "updated_at": "2026-06-05T00:01:00Z"
  }
}
```

Then check Typesense document again. A `404` means delete path worked.

## 11. Common Errors & Fixes

Only Task 3 specific troubleshooting is listed here. Generic setup errors are already covered in:

```text
TaskImplementation/Search Service/task1_Dependency.md
```

Section:

```text
9. Common Errors and Fixes
```

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `search.product_consumer.init_failed` | Redis, RabbitMQ, or Typesense config failed during consumer setup | Check logs for exact nested error, then verify infra health | Start `typesense redis rabbitmq` before `go run` |
| `connect rabbitmq: connection refused` | RabbitMQ not running or wrong host/port | Start RabbitMQ and use `localhost:5672` for host run | Keep host/Docker network values separate |
| RabbitMQ `ACCESS_REFUSED` | Wrong username, password, or vhost in `RABBITMQ_URL` | Match `.env` with `RABBITMQ_DEFAULT_USER`, `RABBITMQ_DEFAULT_PASS`, `RABBITMQ_DEFAULT_VHOST` | Do not change RabbitMQ creds after volume already created without resetting user/vhost |
| `QUEUE_PROVIDER must be rabbitmq for the product indexer` | Config set to `kafka` | Set `QUEUE_PROVIDER=rabbitmq` | Current code is RabbitMQ-only |
| Queue exists but no messages consumed | Routing key mismatch or consumer not running | Check `PRODUCT_INDEXER_ROUTING_KEYS` and RabbitMQ UI Consumers tab | Keep Product Service publisher keys aligned |
| Message lands in DLQ | Invalid JSON, missing metadata, unsupported version/event type, invalid payload | Open DLQ message headers and body, fix publisher payload | Validate event contract in Product Service tests |
| `event_id is required` | Event envelope missing id | Add stable unique `event_id` | Product Service should generate ids before publish |
| `version 2 is not supported` | Publisher sent unsupported schema version | Send version `1` or update consumer contract intentionally | Use event schema version compatibility checks |
| `product_id is required` | Payload missing product id | Include `payload.product_id` | Add publisher-side validation |
| `title is required for searchable products` | Published/searchable product payload incomplete | Send full searchable snapshot | Product Service should publish full search projection |
| `search.product_consumer.duplicate_acked` | Same `event_id` already processed | Use a new `event_id` for a new change | Event ids must be unique per event, not per product |
| Typesense document not created | Collection missing, Typesense down, wrong API key, or event went DLQ | Check Typesense health, Search Service logs, and DLQ | Verify schema ensure before smoke event |
| Redis processed key missing | Consumer did not complete or Redis auth failed | Check Redis password and consumer logs | Keep `SEARCH_REDIS_PASSWORD` aligned with Compose `REDIS_PASSWORD` |
| DLQ keeps growing | Poison messages or downstream dependency failures | Inspect DLQ, fix publisher/Typesense/Redis, replay carefully | Add DLQ alerting |

## 12. Security & Best Practices

Task 3 specific security rules:

- Never expose `TYPESENSE_API_KEY`, `RABBITMQ_URL`, or `SEARCH_REDIS_PASSWORD` to frontend code.
- DLQ access should be restricted. DLQ messages can contain product titles, seller ids, category ids, and operational metadata.
- Product event payload must not include supplier cost, seller private notes, internal moderation comments, user PII, or secrets.
- Use strong RabbitMQ user/password/vhost per environment.
- Keep `SEARCH_INDEXER_ENABLED=true` in environments where the search index must stay fresh.
- Keep `PRODUCT_INDEXER_PROCESSED_EVENT_TTL` long enough for realistic duplicate delivery windows.
- Use durable queues and persistent messages for product index events.
- Monitor DLQ depth and queue lag. Search index stale hona user experience directly hurt karta hai.
- Keep Typesense internal/private. Browser clients should call API Gateway/Search Service, not Typesense directly.
- Treat `SEARCH_ADMIN_AUTH_ENABLED=false` and local dev credentials as local-only settings.

Task 3 operational best practices:

- Product Service should publish a full searchable snapshot so Search Service does not need to call Product Service for each event.
- `event_id` unique per event hona chahiye. Same product update retry same event id use kare, new product update new event id use kare.
- `trace_id`, `request_id`, and `correlation_id` always send karo. Debugging and distributed tracing easy hoti hai.
- Invalid events ko requeue mat karo. Poison message loop queue block kar sakta hai.
- Replaying DLQ messages carefully karo. Pehle payload fix karo, then re-publish with clear audit note.

## 13. Missing or Misconfigured Things

Current audit based on implementation and infra files:

| Finding | Impact | Suggested Fix |
|---|---|---|
| No Search Service Dockerfile found | Service cannot be built as its own container from current repo files | Add service Dockerfile or central build pipeline when container deployment is needed |
| No Search Service Compose service found | Local infra runs in Docker, but Search Service runs on host with `go run` | Add Compose service later if team wants full containerized local backend |
| `infra/k8s/core/search-service` has ConfigMap/Secret examples but no Deployment/Service manifests | K8s runtime is incomplete for Search Service | Add Deployment, Service, probes, resources, and HPA |
| No Redis/RabbitMQ K8s manifests found under `infra/k8s/data` | Search Service K8s config references `redis.data.svc.cluster.local` and `rabbitmq.data.svc.cluster.local`, but manifests are not present here | Add manifests or use managed Redis/RabbitMQ with matching DNS/secrets |
| Kafka is documented as alternate design, but code rejects it | Setting `QUEUE_PROVIDER=kafka` fails startup | Keep RabbitMQ default or implement a real Kafka consumer before enabling |
| Product Service publisher contract not verified in this task | Search index freshness depends on matching event envelope/routing keys | Add Product Service producer tests for Task 7 product events |
| No Prometheus metrics endpoint found for queue lag/DLQ in current Search Service files | Operations may rely on logs/RabbitMQ UI only | Add metrics for consumed, retried, dead-lettered, lag, and indexing duration |
| Readiness endpoint for RabbitMQ/Redis consumer health is not clearly present | `/healthz` may not prove indexer dependencies are healthy | Add readiness check that validates Typesense, Redis, and RabbitMQ consumer status |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/Search Service/task1_Dependency.md` | `2. Go Dependency System` | Go module commands and dependency troubleshooting already documented |
| `TaskImplementation/Search Service/task1_Dependency.md` | `3.2 Redis` | Redis install, Docker, port, password, and verification already documented |
| `TaskImplementation/Search Service/task1_Dependency.md` | `4. Environment Variables` | Full Search Service `.env` already documented, including indexer variables |
| `TaskImplementation/Search Service/task1_Dependency.md` | `5.3 RabbitMQ` | RabbitMQ installation, Docker run, UI, and credential basics already documented |
| `TaskImplementation/Search Service/task1_Dependency.md` | `7. Docker and DevOps Setup` | Compose startup, logs, volumes, and common Docker flow already documented |
| `TaskImplementation/Search Service/task1_Dependency.md` | `8. Project Run Instructions` | Full local Search Service onboarding flow already documented |
| `TaskImplementation/Search Service/task1_Dependency.md` | `9. Common Errors and Fixes` | Generic Typesense/Redis/RabbitMQ/Go/Docker errors already documented |
| `TaskImplementation/Search Service/task2_Dependency.md` | `4.1 Typesense` | Typesense runtime, ports, credentials, and K8s setup already documented |
| `TaskImplementation/Search Service/task2_Dependency.md` | `8. Docker and DevOps Setup` | Typesense-specific Compose/K8s setup already documented |

## 15. Final Checklist

- [ ] Previous dependency documentation checked first.
- [ ] No duplicate Go/Typesense/Redis/RabbitMQ installation docs added.
- [ ] `infra/compose/.env.local` created from `infra/compose/.env.local.example`.
- [ ] Typesense, Redis, and RabbitMQ containers are running.
- [ ] `backend/services/search-service/.env` has `SEARCH_INDEXER_ENABLED=true`.
- [ ] `RABBITMQ_URL` matches local RabbitMQ credentials and vhost.
- [ ] `SEARCH_REDIS_PASSWORD` matches Compose `REDIS_PASSWORD`.
- [ ] `TYPESENSE_API_KEY` matches Compose `TYPESENSE_API_KEY`.
- [ ] `PRODUCT_INDEXER_ROUTING_KEYS` matches Product Service publisher routing keys.
- [ ] `go mod download` completed.
- [ ] `go test ./...` passes.
- [ ] Search Service starts and logs `search.product_consumer.started`.
- [ ] RabbitMQ UI shows `product.events`, `search-service.product-indexer`, and DLQ.
- [ ] Valid `ProductPublished` event creates/updates a Typesense document.
- [ ] Redis processed-event key is created for the event id.
- [ ] Duplicate event id is acked without duplicate side effects.
- [ ] `ProductDeleted` or non-searchable payload deletes the Typesense document.
- [ ] Invalid event goes to DLQ and does not loop forever.
- [ ] DLQ is monitored and not ignored.
- [ ] No real secrets committed.

