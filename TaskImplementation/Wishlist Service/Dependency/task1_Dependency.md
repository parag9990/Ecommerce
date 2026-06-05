# Project Dependency & Setup Guide

## Variable Values

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Wishlist Service` |
| `TASK_FILE_NAME` | `task1.md` |
| `OUTPUT_FILE_NAME` | `task1_Dependency.md` |
| Input task path | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| Output dependency path | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

Note: Is guide ka primary source `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` hai. Us task me model decision documented hai: one authenticated buyer ke paas one private wishlist hogi. Repo me runnable Go service bhi present hai at `backend/services/wishlist-service`, isliye setup guide us runtime service ke dependencies bhi explain karta hai.

---

## 1. Project Overview

`SERVICE_NAME` ka Task 1 documentation-only model task hai. Is task ne decide kiya:

- Har buyer ke paas MVP me single private wishlist hogi.
- Same `product_id` duplicate add nahi hoga.
- `variant_id` optional hai.
- `last_known_price` aur `availability` snapshot fields hain, source of truth Product Service rahega.
- Sharing, public wishlist, multiple named lists, API implementation, events, and database setup Task 1 ke direct scope me nahi the.

Current repo me `backend/services/wishlist-service` already available hai. Ye service HTTP APIs, MongoDB repository, optional Kafka consumers/publishers, product validation, cart integration, health checks, and Mongo migrations use karta hai.

Beginner translation: Task 1 sirf model ka decision document tha, but agar aap actual backend service run karna chahte ho to Go, MongoDB, aur optional Kafka setup karna padega.

---

## 2. Tech Stack

| Technology | Required? | Where used | Simple explanation |
|---|---:|---|---|
| Go | Required for runnable service | `backend/services/wishlist-service/go.mod` | Go ek compiled backend language hai. Is project me HTTP service, Mongo repository, and event workers Go me likhe gaye hain. |
| Go modules | Required | `go.mod`, `go.sum`, `backend/go.work` | Go modules dependency manager hai. Ye batata hai kaunsi external libraries exact version ke saath install hongi. |
| `net/http` | Required | HTTP server and handlers | Go ka built-in HTTP package hai. Service APIs expose karne ke liye use hota hai. |
| `log/slog` | Required | service logging | Go ka structured logging package hai. Logs JSON format me useful fields ke saath nikalte hain. |
| MongoDB | Required for runtime | `wishlists`, `wishlist_events` collections | MongoDB document database hai. Wishlist item list flexible array/document shape me store hoti hai, isliye fit hai. |
| MongoDB Go Driver v2 | Required | `go.mongodb.org/mongo-driver/v2` | Go code ko MongoDB se connect karne ke liye official driver. |
| Kafka protocol | Optional but configured | `github.com/segmentio/kafka-go` | Kafka event streaming system hai. Product events consume karne aur analytics events publish karne ke liye use hota hai. |
| Product Service HTTP API | Required for add item flow | `GET /api/v1/products/{product_id}` | Wishlist me product add karne se pehle product published hai ya nahi validate hota hai. |
| Cart Service HTTP API | Required for move-to-cart flow | `POST /api/v1/cart/items` | Wishlist item ko cart me move karne ke liye downstream Cart Service call hota hai. |
| Docker | Optional but recommended | local Mongo/Kafka setup | Docker se database and queue containers easy run hote hain without manual install. |
| Mermaid | Optional docs tool | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` | Markdown diagrams render karne ke liye use hua. App runtime dependency nahi hai. |
| Shields.io | Optional docs tool | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` | Badges display ke liye external image URLs. Runtime me koi role nahi. |

Not used by this service directly:

- MySQL is not required for `SERVICE_NAME`.
- Redis is not required for `SERVICE_NAME`.
- RabbitMQ is not implemented in current code.
- Node.js, npm, Python, and frontend tooling are not required for this backend service.

---

## 3. Required Software

| Software | Required? | Recommended check command | Why needed |
|---|---:|---|---|
| Git | Required | `git --version` | Repo clone karne ke liye. |
| Go matching `go.mod` | Required | `go version` | Service build, test, and run karne ke liye. `go.mod` currently says `go 1.26.3`. |
| MongoDB server | Required for runtime | `mongosh --eval "db.runCommand({ ping: 1 })"` | Wishlist documents store karne ke liye. |
| `mongosh` | Recommended | `mongosh --version` | Mongo migrations run/verify karne ke liye. |
| Docker | Recommended | `docker --version` | MongoDB/Kafka local containers run karne ke liye. |
| Docker Compose | Optional | `docker compose version` | Multiple services ek saath run karne ke liye. |
| Kafka-compatible broker | Optional | depends on broker | Product events and analytics event publishing ke liye. |
| curl | Recommended | `curl --version` | Health check and API test karne ke liye. |

### Windows Installation Notes

- Git: install Git for Windows.
- Go: install Go version matching `go.mod`, then reopen terminal.
- MongoDB: use MongoDB Community Server installer or Docker Desktop.
- Docker: Docker Desktop install karo and WSL2 backend enable rakho.

### Linux Installation Notes

```bash
git --version
go version
docker --version
mongosh --version
```

If command missing hai, OS package manager se install karo. Example Ubuntu style:

```bash
sudo apt update
sudo apt install -y git curl
```

Go and MongoDB ke liye official installer or Docker route beginner ke liye safer hota hai because package manager versions old ho sakte hain.

### macOS Installation Notes

```bash
git --version
go version
docker --version
mongosh --version
```

Homebrew users usually ye install karte hain:

```bash
brew install git go mongosh
```

Docker ke liye Docker Desktop install karna easiest hai.

---

## 4. Dependency Management

This is a Go project. Dependency files:

| File | Meaning |
|---|---|
| `backend/services/wishlist-service/go.mod` | Module name, Go version, and direct/indirect dependency list. |
| `backend/services/wishlist-service/go.sum` | Dependency checksums. Ye security/integrity ke liye hai. |
| `backend/go.work` | Go workspace file. Is repo me workspace `./services/wishlist-service` ko include karta hai. |

### Main Go dependencies

| Dependency | Version | Why used |
|---|---:|---|
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | MongoDB connect, CRUD, indexes, validators. |
| `github.com/segmentio/kafka-go` | `v0.4.49` | Kafka consumer/publisher for product, notification, and analytics events. |

### Common Commands

From service folder:

```bash
cd backend/services/wishlist-service
go mod download
go test ./...
go build ./cmd/server
go run ./cmd/server
```

From backend workspace folder:

```bash
cd backend
go test ./services/wishlist-service/...
go run ./services/wishlist-service/cmd/server
```

### `go mod tidy` kab run karein?

`go mod tidy` tab run karo jab:

- New import add hua ho.
- Unused dependency remove karni ho.
- `go.mod`/`go.sum` inconsistent lag rahe ho.

Command:

```bash
cd backend/services/wishlist-service
go mod tidy
```

### Dependency failure common reasons

| Problem | Cause | Fix |
|---|---|---|
| `go: go.mod requires go >= ...` | Installed Go version old hai. | Go version update karo to match `go.mod`. |
| `module lookup disabled by GOPROXY` | Go proxy config wrong hai. | `go env GOPROXY` check karo; usually `https://proxy.golang.org,direct` useful hota hai. |
| Checksum mismatch | `go.sum` ya module cache issue ho sakta hai. | Pehle team se verify karo, then `go clean -modcache` and `go mod download`. |
| Private module auth issue | Private repo dependency hoti to credentials chahiye hote. | Current service me private module dependency nahi dikhi. |
| Network timeout | Internet/proxy/DNS issue. | Network check karo, corporate proxy config karo. |

---

## 5. Database Setup

### MongoDB Summary

| Field | Value |
|---|---|
| Database | `wishlist_db` |
| Main collection | `wishlists` |
| Event outbox collection | `wishlist_events` |
| Default URI | `mongodb://localhost:27017` |
| Required for service startup | Yes |
| Default service env var | `WISHLIST_MONGO_URI` |

MongoDB ek document database hai. Simple Hinglish me: SQL tables ki jagah MongoDB JSON-like documents store karta hai. Wishlist me user ke items ek document ke andar array me store ho sakte hain, isliye read-heavy wishlist model ke liye MongoDB practical hai.

### Connection String Format

No-auth local:

```env
WISHLIST_MONGO_URI=mongodb://localhost:27017
```

With username/password:

```env
WISHLIST_MONGO_URI=mongodb://wishlist_user:strong_password@localhost:27017/wishlist_db?authSource=admin
```

Docker internal network example:

```env
WISHLIST_MONGO_URI=mongodb://wishlist-mongo:27017
```

Credentials placement:

- Local shell: `backend/services/wishlist-service/.env`, then manually load it.
- Docker Compose: `env_file` or `environment`.
- Production: secret manager or Kubernetes Secret.
- Never commit real DB passwords.

### Local MongoDB with Docker

Recommended beginner command:

```bash
docker volume create wishlist_mongo_data
docker run -d --name wishlist-mongo \
  -p 27017:27017 \
  -v wishlist_mongo_data:/data/db \
  mongo:7
```

Verify:

```bash
docker ps
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
```

Stop/start:

```bash
docker stop wishlist-mongo
docker start wishlist-mongo
docker logs wishlist-mongo
```

Remove container but keep data volume:

```bash
docker rm -f wishlist-mongo
```

### Manual MongoDB Installation

Windows:

- Install MongoDB Community Server.
- Install MongoDB Shell (`mongosh`).
- Start MongoDB service from Services app or terminal.
- Verify with `mongosh`.

Linux:

- Install MongoDB Community Server using official MongoDB packages or use Docker.
- Start service using systemd if installed locally.
- Verify with `mongosh --eval "db.runCommand({ ping: 1 })"`.

macOS:

- Install MongoDB Community using Homebrew tap or use Docker.
- Start MongoDB service.
- Verify with `mongosh`.

### Migrations

Migration files exist:

```text
backend/services/wishlist-service/migrations/mongo/
```

Current service also calls `EnsureCollection` at startup, which creates/updates collections, validators, and indexes. For local development, startup can prepare MongoDB automatically. For controlled environments, run migrations explicitly.

Run up migrations:

```bash
cd backend/services/wishlist-service
mongosh "mongodb://localhost:27017" migrations/mongo/001_create_wishlists_collection.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/002_add_wishlist_item_variant_index.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/003_add_price_drop_candidate_index.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/004_create_wishlist_events_collection.up.js
```

Verify collections:

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval "show collections"
```

Verify indexes:

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.wishlists.getIndexes()"
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.wishlist_events.getIndexes()"
```

### MongoDB Docker Compose Example

```yaml
services:
  wishlist-mongo:
    image: mongo:7
    container_name: wishlist-mongo
    ports:
      - "27017:27017"
    volumes:
      - wishlist_mongo_data:/data/db
    restart: unless-stopped

volumes:
  wishlist_mongo_data:
```

---

## 6. Redis / Queue / External Services

### Redis

Redis current `SERVICE_NAME` runtime me directly used nahi hai. Platform docs me Redis cart/session/rate-limit use cases ke liye planned hai, but this service ke code me Redis client or env var nahi mila.

Beginner note: Redis fast in-memory cache hota hai. Yahan abhi mandatory nahi hai, so Redis install na ho to `SERVICE_NAME` fail nahi hona chahiye.

### Kafka

Kafka optional hai, but event features ke liye important hai.

| Feature | Env switch | Default | Kafka needed? |
|---|---|---:|---:|
| Product event consumer and price-drop notification publisher | `WISHLIST_EVENTS_BACKEND` | `disabled` | Only if set to `kafka` |
| Wishlist analytics outbox publisher | `WISHLIST_ANALYTICS_EVENTS_ENABLED` + `WISHLIST_EVENT_PUBLISHER` | enabled + `kafka` | Yes unless disabled |

Simple explanation: Kafka ek event streaming platform hai. Product price change, availability change, and wishlist analytics jaise async events ke liye use hota hai. Agar beginner sirf API health check and local DB setup kar raha hai, Kafka disable rakhna easiest hai.

Minimal local mode without Kafka:

```env
WISHLIST_EVENTS_BACKEND=disabled
WISHLIST_ANALYTICS_EVENTS_ENABLED=false
WISHLIST_EVENT_PUBLISHER=disabled
```

Kafka enabled mode:

```env
WISHLIST_EVENTS_BACKEND=kafka
WISHLIST_ANALYTICS_EVENTS_ENABLED=true
WISHLIST_EVENT_PUBLISHER=kafka
WISHLIST_KAFKA_BROKERS=localhost:9092
```

Topics used:

| Topic env var | Default topic | Purpose |
|---|---|---|
| `WISHLIST_PRODUCT_EVENTS_TOPIC` | `product.events` | Product availability/price events consume. |
| `WISHLIST_PRODUCT_EVENTS_DLQ_TOPIC` | `product.events.wishlist.dlq` | Failed product events dead-letter queue. |
| `WISHLIST_NOTIFICATION_COMMANDS_TOPIC` | `notification.commands` | Price drop notification command publish. |
| `WISHLIST_EVENT_TOPIC` | `recommendation.events` | Wishlist analytics event publish. |

Local Kafka-compatible broker with Docker example:

```bash
docker run -d --name wishlist-redpanda \
  -p 9092:9092 \
  docker.redpanda.com/redpandadata/redpanda:v24.1.9 \
  redpanda start \
  --overprovisioned \
  --smp 1 \
  --memory 512M \
  --reserve-memory 0M \
  --node-id 0 \
  --check=false \
  --kafka-addr 0.0.0.0:9092 \
  --advertise-kafka-addr localhost:9092
```

This runs a Kafka API compatible broker. Code uses Kafka protocol, so it can talk to this in local development.

Verify broker port:

```bash
docker ps
docker logs wishlist-redpanda
```

### Product Service

Required for `POST /api/v1/wishlist/items`.

| Env var | Default |
|---|---|
| `WISHLIST_PRODUCT_SERVICE_BASE_URL` | `http://localhost:8082` |
| fallback | `PRODUCT_SERVICE_BASE_URL` |

Expected endpoint:

```text
GET /api/v1/products/{product_id}
```

Expected response can be plain product JSON or envelope with `data`. Service checks:

- Product status must be `published`.
- If `variant_id` is provided, variant must exist.
- Price and stock fields are used to create wishlist snapshot.

### Cart Service

Required for `POST /api/v1/wishlist/items/{product_id}/move-to-cart`.

| Env var | Default |
|---|---|
| `WISHLIST_CART_SERVICE_BASE_URL` | `http://localhost:8083` |
| fallback | `CART_SERVICE_BASE_URL` |

Expected endpoint:

```text
POST /api/v1/cart/items
```

Wishlist Service forwards:

- `X-User-ID`
- `X-User-Roles: buyer`
- `X-Request-ID`
- `X-Idempotency-Key` when provided

### Docker

Docker is not mandatory if MongoDB/Kafka are installed locally, but beginners should use Docker for dependencies because setup repeatable hota hai.

Useful commands:

```bash
docker ps
docker logs wishlist-mongo
docker stop wishlist-mongo
docker start wishlist-mongo
docker compose up -d
docker compose down
docker compose logs -f
```

### Kubernetes

Kubernetes manifests for this service are not present in current repo snapshot. Production deployment ke liye Deployment, Service, ConfigMap, Secret, readiness/liveness probes, and network policies add karne honge.

---

## 7. Environment Variables

Important: Current Go code `os.LookupEnv` use karta hai. It does not automatically read `.env` files. Matlab `.env` file create karne ke baad usko shell me load bhi karna padega.

Recommended local file:

```text
backend/services/wishlist-service/.env
```

Load on Linux/macOS/Git Bash:

```bash
cd backend/services/wishlist-service
set -a
source .env
set +a
```

PowerShell example:

```powershell
$env:WISHLIST_HTTP_ADDR=":8084"
$env:WISHLIST_MONGO_URI="mongodb://localhost:27017"
```

### Complete `.env` Example

This example is beginner-friendly local mode. Kafka is disabled so service can start with only MongoDB.

```env
SERVICE_NAME=wishlist-service
APP_ENV=local
WISHLIST_LOG_LEVEL=info

WISHLIST_HTTP_ADDR=:8084
WISHLIST_HTTP_READ_TIMEOUT=5s
WISHLIST_HTTP_WRITE_TIMEOUT=10s
WISHLIST_HTTP_IDLE_TIMEOUT=60s

WISHLIST_MONGO_URI=mongodb://localhost:27017
WISHLIST_MONGO_DATABASE=wishlist_db
WISHLIST_MONGO_COLLECTION=wishlists
WISHLIST_MONGO_CONNECT_TIMEOUT=5s
WISHLIST_MONGO_SERVER_SELECTION_TIMEOUT=5s

WISHLIST_PRODUCT_SERVICE_BASE_URL=http://localhost:8082
WISHLIST_PRODUCT_SERVICE_TIMEOUT=1s

WISHLIST_CART_SERVICE_BASE_URL=http://localhost:8083
WISHLIST_CART_SERVICE_TIMEOUT=1s

WISHLIST_AUTH_USER_ID_HEADER=X-User-ID
WISHLIST_AUTH_ROLES_HEADER=X-User-Roles
WISHLIST_REQUEST_ID_HEADER=X-Request-ID
WISHLIST_AUTH_REQUIRE_BUYER_ROLE=true

WISHLIST_EVENTS_BACKEND=disabled
WISHLIST_KAFKA_BROKERS=localhost:9092
WISHLIST_PRODUCT_EVENTS_TOPIC=product.events
WISHLIST_PRODUCT_EVENTS_GROUP_ID=wishlist-service
WISHLIST_PRODUCT_EVENTS_DLQ_TOPIC=product.events.wishlist.dlq
WISHLIST_PRODUCT_EVENTS_MAX_ATTEMPTS=3
WISHLIST_PRODUCT_EVENTS_RETRY_BACKOFF=500ms

WISHLIST_NOTIFICATION_COMMANDS_TOPIC=notification.commands
WISHLIST_PRICE_DROP_TEMPLATE_KEY=wishlist_price_drop
WISHLIST_PRICE_DROP_BATCH_SIZE=500
WISHLIST_PRICE_DROP_MIN_DELTA_AMOUNT=1

WISHLIST_ANALYTICS_EVENTS_ENABLED=false
WISHLIST_EVENT_OUTBOX_COLLECTION=wishlist_events
WISHLIST_EVENT_TOPIC=recommendation.events
WISHLIST_EVENT_PUBLISHER=disabled
WISHLIST_EVENT_POLL_INTERVAL=5s
WISHLIST_EVENT_BATCH_SIZE=50
WISHLIST_EVENT_MAX_ATTEMPTS=5

WISHLIST_STARTUP_TIMEOUT=10s
WISHLIST_SHUTDOWN_TIMEOUT=10s
```

### Environment Variable Reference

| Variable | Required? | Default | Purpose | Security/config note |
|---|---:|---|---|---|
| `SERVICE_NAME` | Optional | `wishlist-service` | Log/service name. | Not secret. |
| `APP_ENV` | Optional | `local` | Environment label. | Use `prod`, `staging`, `local` clearly. |
| `WISHLIST_LOG_LEVEL` | Optional | `info` | Log verbosity. | Avoid `debug` in prod if logs may include sensitive data. |
| `LOG_LEVEL` | Optional fallback | `info` | Shared fallback log level. | `WISHLIST_LOG_LEVEL` wins. |
| `WISHLIST_HTTP_ADDR` | Required by validation | `:8084` | HTTP bind address. | Use private network in prod. |
| `WISHLIST_HTTP_READ_TIMEOUT` | Optional | `5s` | Request read timeout. | Keep positive. |
| `WISHLIST_HTTP_WRITE_TIMEOUT` | Optional | `10s` | Response write timeout. | Keep positive. |
| `WISHLIST_HTTP_IDLE_TIMEOUT` | Optional | `60s` | Keep-alive idle timeout. | Keep positive. |
| `WISHLIST_MONGO_URI` | Required | `mongodb://localhost:27017` | MongoDB connection. | May contain password. Never commit real credentials. |
| `WISHLIST_MONGO_DATABASE` | Required | `wishlist_db` | Mongo database name. | Use separate DB per env. |
| `WISHLIST_MONGO_COLLECTION` | Required | `wishlists` | Main collection name. | Usually do not change after data exists. |
| `WISHLIST_MONGO_CONNECT_TIMEOUT` | Optional | `5s` | Mongo connect timeout. | Keep positive. |
| `WISHLIST_MONGO_SERVER_SELECTION_TIMEOUT` | Optional | `5s` | Mongo server selection timeout. | Keep positive. |
| `WISHLIST_PRODUCT_SERVICE_BASE_URL` | Required | `http://localhost:8082` | Product Service URL. | Internal URL in prod. |
| `PRODUCT_SERVICE_BASE_URL` | Optional fallback | `http://localhost:8082` | Fallback Product Service URL. | Prefer service-specific var. |
| `WISHLIST_PRODUCT_SERVICE_TIMEOUT` | Optional | `1s` | Product call timeout. | Keep realistic for env. |
| `WISHLIST_PRODUCT_CALL_TIMEOUT_MS` | Optional fallback | none | Old millisecond timeout fallback. | Use one timeout var, not both. |
| `WISHLIST_CART_SERVICE_BASE_URL` | Required | `http://localhost:8083` | Cart Service URL. | Internal URL in prod. |
| `CART_SERVICE_BASE_URL` | Optional fallback | `http://localhost:8083` | Fallback Cart Service URL. | Prefer service-specific var. |
| `WISHLIST_CART_SERVICE_TIMEOUT` | Optional | `1s` | Cart call timeout. | Keep positive. |
| `WISHLIST_CART_CALL_TIMEOUT_MS` | Optional fallback | none | Old millisecond timeout fallback. | Use one timeout var, not both. |
| `WISHLIST_AUTH_USER_ID_HEADER` | Required | `X-User-ID` | Header containing authenticated user id. | Must be injected by trusted gateway. |
| `WISHLIST_AUTH_ROLES_HEADER` | Required | `X-User-Roles` | Header containing roles. | Do not expose service directly to public internet. |
| `WISHLIST_REQUEST_ID_HEADER` | Required | `X-Request-ID` | Trace/request id header. | Not secret. |
| `WISHLIST_AUTH_REQUIRE_BUYER_ROLE` | Optional | `true` | Requires `buyer` role. | Keep `true` outside tests. |
| `WISHLIST_EVENTS_BACKEND` | Optional | `disabled` | Product event consumer backend. | Set `kafka` only when Kafka is running. |
| `WISHLIST_KAFKA_BROKERS` | Required if Kafka enabled | `localhost:9092` | Kafka brokers CSV. | Internal broker addresses in prod. |
| `WISHLIST_PRODUCT_EVENTS_TOPIC` | Required if product events enabled | `product.events` | Product event source topic. | Topic must exist or broker must auto-create. |
| `WISHLIST_PRODUCT_EVENTS_GROUP_ID` | Required if product events enabled | `wishlist-service` | Consumer group id. | Stable per environment. |
| `WISHLIST_PRICE_EVENTS_GROUP_ID` | Optional fallback alias | none | Alternative group id var. | Avoid setting both aliases differently. |
| `WISHLIST_PRODUCT_EVENTS_DLQ_TOPIC` | Required if product events enabled | `product.events.wishlist.dlq` | Dead-letter topic. | Monitor DLQ. |
| `WISHLIST_PRICE_EVENTS_DLQ_TOPIC` | Optional fallback alias | none | Alternative DLQ topic var. | Avoid conflicting aliases. |
| `WISHLIST_PRODUCT_EVENTS_MAX_ATTEMPTS` | Optional | `3` | Retry count for product events. | Keep positive. |
| `WISHLIST_PRODUCT_EVENTS_RETRY_BACKOFF` | Optional | `500ms` | Retry wait time. | Keep positive. |
| `WISHLIST_NOTIFICATION_COMMANDS_TOPIC` | Required if product events enabled | `notification.commands` | Price drop notification commands. | Downstream notification service consumes this. |
| `WISHLIST_PRICE_DROP_TEMPLATE_KEY` | Required if product events enabled | `wishlist_price_drop` | Notification template key. | Must match Notification Service templates. |
| `WISHLIST_PRICE_DROP_BATCH_SIZE` | Optional | `500` | Mongo batch size for price-drop candidates. | Keep positive. |
| `WISHLIST_PRICE_DROP_MIN_DELTA_AMOUNT` | Optional | `1` | Minimum price drop in minor units. | Keep positive. |
| `WISHLIST_ANALYTICS_EVENTS_ENABLED` | Optional | `true` | Writes wishlist add/remove events to outbox. | Disable for simplest local setup. |
| `WISHLIST_EVENT_OUTBOX_COLLECTION` | Required if analytics enabled | `wishlist_events` | Mongo outbox collection. | Collection has TTL/indexes. |
| `WISHLIST_EVENT_TOPIC` | Required if analytics enabled | `recommendation.events` | Analytics destination topic. | Kafka needed if publisher is `kafka`. |
| `WISHLIST_EVENT_PUBLISHER` | Optional | `kafka` | Analytics publisher backend. | Set `disabled` if Kafka not running. |
| `WISHLIST_EVENT_POLL_INTERVAL` | Optional | `5s` | Outbox polling interval. | Keep positive. |
| `WISHLIST_EVENT_BATCH_SIZE` | Optional | `50` | Outbox publish batch. | Keep positive. |
| `WISHLIST_EVENT_MAX_ATTEMPTS` | Optional | `5` | Max publish attempts. | Keep positive. |
| `WISHLIST_STARTUP_TIMEOUT` | Optional | `10s` | Startup dependency timeout. | Increase if Mongo slow. |
| `WISHLIST_SHUTDOWN_TIMEOUT` | Optional | `10s` | Graceful shutdown timeout. | Keep positive. |

Common `.env` mistakes:

- File banaya but `source .env` nahi kiya.
- Value me extra quotes/spaces daal diye.
- `WISHLIST_HTTP_ADDR=8084` likha instead of `:8084`.
- Mongo container running nahi hai but URI localhost pe point kar raha hai.
- Kafka disabled nahi kiya and broker running nahi hai.
- Product/Cart URLs wrong port pe set hain.

---

## 8. Docker Setup

Current repo snapshot me service-specific Dockerfile or full local docker-compose file present nahi mila. Isliye below setup dependency containers ke liye practical local guide hai.

### Recommended beginner local stack

- Run MongoDB in Docker.
- Disable Kafka until event testing needed.
- Run Go service directly using `go run`.

### Docker Compose Example for Dependencies

```yaml
services:
  wishlist-mongo:
    image: mongo:7
    container_name: wishlist-mongo
    ports:
      - "27017:27017"
    volumes:
      - wishlist_mongo_data:/data/db
    restart: unless-stopped

  wishlist-redpanda:
    image: docker.redpanda.com/redpandadata/redpanda:v24.1.9
    container_name: wishlist-redpanda
    command:
      - redpanda
      - start
      - --overprovisioned
      - --smp
      - "1"
      - --memory
      - 512M
      - --reserve-memory
      - 0M
      - --node-id
      - "0"
      - --check=false
      - --kafka-addr
      - 0.0.0.0:9092
      - --advertise-kafka-addr
      - localhost:9092
    ports:
      - "9092:9092"
    restart: unless-stopped

volumes:
  wishlist_mongo_data:
```

Commands:

```bash
docker compose up -d
docker compose ps
docker compose logs -f wishlist-mongo
docker compose logs -f wishlist-redpanda
docker compose down
```

Volumes:

- `wishlist_mongo_data` stores MongoDB data.
- `docker compose down` keeps named volume by default.
- `docker compose down -v` deletes volume and data. Use carefully.

Networks:

- From host machine, use `localhost:27017` and `localhost:9092`.
- From another container in same compose network, use service names like `wishlist-mongo:27017`.

Restart policy:

- `unless-stopped` is good for local dependencies because containers restart after Docker restarts.

---

## 9. Local Development Setup

### Step 1: Clone repository

```bash
git clone <repo-url>
```

### Step 2: Go to project directory

```bash
cd Ecommerce
```

### Step 3: Check required tools

```bash
git --version
go version
docker --version
curl --version
```

### Step 4: Install Go dependencies

```bash
cd backend/services/wishlist-service
go mod download
```

### Step 5: Start MongoDB

```bash
docker volume create wishlist_mongo_data
docker run -d --name wishlist-mongo \
  -p 27017:27017 \
  -v wishlist_mongo_data:/data/db \
  mongo:7
```

Verify:

```bash
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
```

### Step 6: Create and load `.env`

Create local env at:

```text
backend/services/wishlist-service/.env
```

For minimum local run, use the `.env` example from section 7 with Kafka disabled.

Load:

```bash
cd backend/services/wishlist-service
set -a
source .env
set +a
```

### Step 7: Run migrations

Manual migration route:

```bash
cd backend/services/wishlist-service
mongosh "mongodb://localhost:27017" migrations/mongo/001_create_wishlists_collection.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/002_add_wishlist_item_variant_index.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/003_add_price_drop_candidate_index.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/004_create_wishlist_events_collection.up.js
```

Beginner note: If you skip manual migrations, service startup still calls collection setup. Manual migrations are better when you want controlled DB changes and clear audit.

### Step 8: Run tests

```bash
cd backend/services/wishlist-service
go test ./...
```

### Step 9: Start backend service

```bash
cd backend/services/wishlist-service
go run ./cmd/server
```

Expected log:

```text
wishlist service listening addr=:8084
```

### Step 10: Verify health APIs

```bash
curl http://localhost:8084/healthz
curl http://localhost:8084/readyz
```

Expected response:

```json
{"status":"ok","service":"wishlist-service"}
```

### Step 11: Verify wishlist API shape

Add item requires Product Service to be running and returning a published product:

```bash
curl -X POST http://localhost:8084/api/v1/wishlist/items \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_123" \
  -H "X-User-Roles: buyer" \
  -H "X-Request-ID: req_local_1" \
  -d '{"product_id":"prod_123","variant_id":"var_1"}'
```

Remove item:

```bash
curl -X DELETE http://localhost:8084/api/v1/wishlist/items/prod_123 \
  -H "X-User-ID: user_123" \
  -H "X-User-Roles: buyer" \
  -H "X-Request-ID: req_local_2"
```

Move to cart requires Cart Service running:

```bash
curl -X POST http://localhost:8084/api/v1/wishlist/items/prod_123/move-to-cart \
  -H "X-User-ID: user_123" \
  -H "X-User-Roles: buyer" \
  -H "X-Request-ID: req_local_3" \
  -H "X-Idempotency-Key: move-prod-123-user-123"
```

---

## 10. Running the Project

### Minimum local run

This is best for beginners.

Required:

- Go installed.
- MongoDB running.
- Kafka disabled in env.

Commands:

```bash
cd backend/services/wishlist-service
set -a
source .env
set +a
go run ./cmd/server
```

Verify:

```bash
curl http://localhost:8084/healthz
curl http://localhost:8084/readyz
```

### Full event-enabled run

Required:

- Go installed.
- MongoDB running.
- Kafka-compatible broker running.
- Product Service running if consuming product events.
- Notification Service topic/consumer ready if price drop notifications are expected.

Env switches:

```env
WISHLIST_EVENTS_BACKEND=kafka
WISHLIST_ANALYTICS_EVENTS_ENABLED=true
WISHLIST_EVENT_PUBLISHER=kafka
WISHLIST_KAFKA_BROKERS=localhost:9092
```

Start:

```bash
cd backend/services/wishlist-service
set -a
source .env
set +a
go run ./cmd/server
```

---

## 11. Ports & Networking

| Service | Port | Required? | Purpose |
|---|---:|---:|---|
| `SERVICE_NAME` HTTP server | `8084` | Yes | Health and wishlist APIs. |
| MongoDB | `27017` | Yes | Wishlist persistence. |
| Product Service | `8082` | Required for add item | Product validation. |
| Cart Service | `8083` | Required for move-to-cart | Cart item creation. |
| Kafka broker | `9092` | Optional | Product events and analytics events. |

### Port conflicts

If `:8084` already in use:

```bash
WISHLIST_HTTP_ADDR=:18084 go run ./cmd/server
```

Then verify:

```bash
curl http://localhost:18084/healthz
```

Find port user on Linux/macOS:

```bash
lsof -i :8084
```

Find port user on Windows PowerShell:

```powershell
netstat -ano | findstr :8084
```

### Docker networking

- Service running on host should use `mongodb://localhost:27017`.
- Service running inside Docker Compose should use `mongodb://wishlist-mongo:27017`.
- `localhost` inside a container means that same container, not host machine.

### Firewall issues

Local firewall usually does not block localhost. If service is in VM/WSL/Docker and host cannot connect, check:

- Port mapping `-p host:container`.
- Service bind address. `:8084` listens on all interfaces.
- Docker Desktop/WSL network settings.

---

## 12. Common Errors & Fixes

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `address already in use` | Port `8084` busy hai. | `WISHLIST_HTTP_ADDR=:18084` set karo ya old process stop karo. | Standard ports document karo. |
| `mongo unavailable` on `/readyz` | MongoDB down or wrong URI. | `docker start wishlist-mongo`; check `WISHLIST_MONGO_URI`. | Health check run before API testing. |
| `server selection error` | Mongo driver server find nahi kar pa raha. | Mongo port, container, URI, network check karo. | Use correct host for host vs Docker. |
| `WISHLIST_MONGO_URI is required` | Env var blank set hai. | Remove blank var or set valid URI. | `.env` me empty values avoid karo. |
| `.env: No such file or directory` | Wrong folder se `source .env` run hua. | `cd backend/services/wishlist-service` then source. | `.env` location consistent rakho. |
| Service ignores `.env` | Go app auto-load nahi karta. | `set -a && source .env && set +a` use karo. | README/setup docs me loading step rakho. |
| `unsupported wishlist events backend` | Invalid value, like `rabbitmq`. | `disabled` or `kafka` use karo. | Allowed values table follow karo. |
| Kafka publish failures | Broker down or topic unreachable. | Kafka start karo or `WISHLIST_EVENT_PUBLISHER=disabled`. | Local minimal mode me Kafka disable rakho. |
| Product add returns `PRODUCT_SERVICE_UNAVAILABLE` | Product Service running nahi hai ya URL wrong. | Start Product Service or set `WISHLIST_PRODUCT_SERVICE_BASE_URL`. | Health/dependency check add karo. |
| Product add returns `PRODUCT_NOT_AVAILABLE` | Product status `published` nahi hai. | Product test data update karo. | Use known-good fixture product. |
| API returns `UNAUTHENTICATED` | `X-User-ID` missing. | Header add karo. | API Gateway should inject auth context. |
| API returns `FORBIDDEN` | `buyer` role missing. | `X-User-Roles: buyer` header add karo. | Keep role mapping documented. |
| Move-to-cart returns `VARIANT_REQUIRED_FOR_CART` | Wishlist item me variant missing hai. | Add item with `variant_id` first. | Product variants handle clearly in UI. |
| `go mod download` fails | Network/proxy issue. | Network check, `GOPROXY` config check. | Dependencies download during onboarding. |
| `go test` fails due Go version | Installed Go version mismatch. | Install Go matching `go.mod`. | Toolchain version pin/document. |
| Migration failed with auth error | Mongo credentials missing/wrong. | URI username/password/authSource fix karo. | Use secret manager and verify with `mongosh`. |
| Docker daemon not running | Docker Desktop/Engine stopped. | Start Docker. | Make Docker startup part of onboarding checklist. |
| Permission denied | Binary/file permission issue. | Check file ownership or run from proper user. | Avoid sudo-created files in repo. |

---

## 13. Security & Best Practices

### Security audit from current implementation

| Area | Current observation | Risk | Suggested fix |
|---|---|---|---|
| Credentials | No hardcoded DB username/password found. Defaults use localhost no-auth Mongo. | Safe for local, unsafe if copied to prod. | Use Mongo auth in shared/prod env. Store secrets outside git. |
| `.env` loading | Code reads OS env only. | Beginners may create `.env` and think app reads it. | Add `.env.example` and document shell loading or add config loader intentionally. |
| Auth | Service trusts `X-User-ID` and role headers. | If exposed publicly, user can spoof headers. | Keep service behind API Gateway/private network. Validate JWT at gateway or service boundary. |
| Buyer role | `WISHLIST_AUTH_REQUIRE_BUYER_ROLE=true` by default. | Good default, but can be disabled. | Do not disable outside local tests. |
| Mongo defaults | `mongodb://localhost:27017` no auth. | Fine local, weak for shared environments. | Use authenticated MongoDB URI and network restrictions. |
| Kafka defaults | Broker defaults to `localhost:9092`. | Local only. | Use env-specific broker list and TLS/SASL if required. |
| Health checks | `/healthz` and `/readyz` exist. | Good. Readiness checks Mongo only, not Product/Cart/Kafka. | Add dependency-specific readiness if deployment needs strict checks. |
| Docker config | No service Dockerfile/compose found in repo snapshot. | Onboarding and deploy setup manual. | Add service Dockerfile and local compose under infra/deploy. |
| Migrations | JS migrations exist and startup also ensures collections. | Two setup paths can confuse teams. | Decide clear local/prod migration strategy. |
| API contract mismatch | Docs mention `GET /api/v1/wishlist`, code currently registers item routes and health. | Developers may test route that is not implemented. | Update docs or implement missing route. |
| gRPC docs | Architecture docs mention gRPC WishlistService. Current code exposes HTTP handlers. | Contract/runtime mismatch. | Add gRPC transport or update current scope docs. |

### Best practices for beginners

- Never commit real `.env` files.
- Commit only `.env.example` with dummy values.
- Use different database names for local, staging, and production.
- Use Docker volumes for local MongoDB data.
- Run `/healthz` and `/readyz` before API testing.
- Keep `WISHLIST_AUTH_REQUIRE_BUYER_ROLE=true`.
- Do not expose service directly to public internet if it trusts auth headers.
- Keep Product/Cart Service URLs internal in production.
- Monitor Kafka DLQ topics.
- Run `go test ./...` before committing backend changes.
- Keep `go.mod` and `go.sum` committed together.
- Document every new env var when code changes.

---

## 14. Missing or Misconfigured Things

These are not blockers for Task 1 documentation, but they matter for smooth onboarding:

| Missing/misaligned item | Why it matters | Recommended action |
|---|---|---|
| No `.env.example` found for service | New developers do not know required env values. | Add `backend/services/wishlist-service/.env.example`. |
| No local compose file found | Mongo/Kafka setup is manual. | Add `infra/compose/docker-compose.local.yml` or service-specific compose. |
| No service Dockerfile found | Containerized deployment cannot be built directly from repo conventions. | Add Dockerfile for `backend/services/wishlist-service`. |
| Task 1 says model-only, repo has full service implementation | Scope can confuse beginners. | Keep task docs and runtime docs clearly separated. |
| API master lists `GET /api/v1/wishlist`, current handler does not register it | Testers may call missing endpoint. | Implement route or update API contract. |
| Architecture mentions gRPC, current service is HTTP | Platform contract not fully implemented. | Add proto/gRPC transport in later task or update docs. |
| Kafka analytics defaults can surprise local users | Default publisher is `kafka` while broker may not be running. | Local `.env.example` should disable Kafka by default. |
| Product/Cart services are external but no local mocks are included | Add/move API testing needs downstream services. | Add dev mocks or compose full stack. |

---

## 15. Final Checklist

Use this before saying local setup is ready:

- [ ] Git installed.
- [ ] Go version matches `backend/services/wishlist-service/go.mod`.
- [ ] Docker installed and running, if using containers.
- [ ] MongoDB running on expected host/port.
- [ ] `mongosh` can connect to MongoDB.
- [ ] Go dependencies downloaded with `go mod download`.
- [ ] `.env` created at `backend/services/wishlist-service/.env`.
- [ ] `.env` loaded into shell before running service.
- [ ] `WISHLIST_MONGO_URI` points to correct MongoDB.
- [ ] Kafka disabled for beginner local mode, or Kafka broker running for event mode.
- [ ] Product Service URL configured if testing add item.
- [ ] Cart Service URL configured if testing move-to-cart.
- [ ] Mongo migrations run, or startup collection setup verified.
- [ ] `go test ./...` passes.
- [ ] Service starts with `go run ./cmd/server`.
- [ ] `curl http://localhost:8084/healthz` returns `ok`.
- [ ] `curl http://localhost:8084/readyz` returns `ok`.
- [ ] API calls include `X-User-ID` and `X-User-Roles: buyer`.
- [ ] Logs checked for Mongo/Kafka/Product/Cart errors.
- [ ] Real credentials are not committed to git.

Final beginner summary: Pehle MongoDB run karo, phir env load karo, Kafka local me disable rakho, service start karo, `/healthz` and `/readyz` verify karo. Add item and move-to-cart tabhi test karo jab Product Service and Cart Service available ho.
