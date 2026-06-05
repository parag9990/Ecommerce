# ${SERVICE_NAME} ${TASK_FILE_NAME} - Dependency and Setup Documentation

```text
SERVICE_NAME = "Cart Service"
TASK_FILE_NAME = "task1.md"
OUTPUT_FILE_NAME = "task1_Dependency.md"
INPUT_PATH = "TaskImplementation/Cart Service/task1.md"
OUTPUT_PATH = "TaskImplementation/Cart Service/task1_Dependency.md"
```

## 1. Purpose

Ye document `${SERVICE_NAME}` ke `${TASK_FILE_NAME}` ke liye dependency, environment, database, Docker, run, migration, aur troubleshooting guide hai.

Beginner developer is guide ko follow karke:

- repository clone kar sakta hai
- Go dependencies install kar sakta hai
- MongoDB aur Redis local run kar sakta hai
- required environment variables configure kar sakta hai
- MongoDB cart collection migration/bootstrap run kar sakta hai
- backend HTTP service aur expiry worker start kar sakta hai
- common setup issues debug kar sakta hai

Important scope note:

- `${TASK_FILE_NAME}` khud documentation-only cart rules task tha. Usme business rules define hue the, runtime code nahi.
- Current repository me `backend/services/cart-service` Go service already present hai. Isliye ye dependency doc real implemented service files bhi inspect karke banaya gaya hai.
- Business logic rewrite nahi kiya gaya. Sirf setup, dependency, environment, database, external service, DevOps, aur debugging information document ki gayi hai.

## 2. Files Inspected

| File | Why inspected |
|---|---|
| `${INPUT_PATH}` | Task boundary, cart rules, architecture dependency hints |
| `backend/services/cart-service/go.mod` | Go version and Go module dependencies |
| `backend/services/cart-service/go.sum` | Locked dependency checksums |
| `backend/go.work` | Workspace configuration |
| `backend/services/cart-service/.env.example` | Environment variable source |
| `backend/services/cart-service/internal/config/config.go` | Required env validation and defaults |
| `backend/services/cart-service/cmd/server/main.go` | Runtime dependencies, startup flow, ports |
| `backend/services/cart-service/cmd/cart-expiry-worker/main.go` | Background worker requirements |
| `backend/services/cart-service/internal/repository/*.go` | MongoDB and Redis usage |
| `backend/services/cart-service/internal/client/*.go` | Product Service and CMS HTTP dependencies |
| `backend/services/cart-service/migrations/mongo/*.js` | MongoDB schema migration scripts |
| `docs/13-developer-guide.md` | General local setup guidance |
| `docs/11-devops-external-services.md` | External service and DevOps assumptions |

## 3. High-Level Tech Stack

| Technology | Required? | What it is | Why used here |
|---|---:|---|---|
| Go `1.26.3` | Yes | Go backend programming language hai. Fast, compiled, aur microservices ke liye common choice hai. | Cart HTTP API, domain logic, repositories, clients, worker |
| Go modules | Yes | Go ka dependency manager. `go.mod` dependencies declare karta hai aur `go.sum` checksum lock karta hai. | MongoDB/Redis libraries manage karne ke liye |
| Go workspace | Yes for repo flow | `go.work` multiple Go modules ko ek workspace me connect karta hai. | `backend/services/cart-service` ko workspace service ke roop me include karta hai |
| Standard `net/http` | Yes | Go ka built-in HTTP server library. | Cart REST endpoints, health checks, schema bootstrap |
| `log/slog` | Yes | Go structured logging package. | JSON logs emit karne ke liye |
| MongoDB | Yes | Document database. Data JSON-like documents me store hota hai. | Durable cart store: `cart_db.carts` |
| MongoDB Go Driver | Yes | Official Go client for MongoDB. | Mongo connect, ping, collection, indexes, transactions |
| Redis | Yes | In-memory cache/store. Fast key-value storage. | Active cart cache, summary cache, expiry worker lock |
| `go-redis/v9` | Yes | Go Redis client library. | Redis ping, SET, DEL, lock operations |
| Product Service HTTP API | Required for add item | Product details and variant stock/price dene wali service. | Add item ke time product snapshot and inventory validate hota hai |
| CMS Service HTTP API | Required for coupon preview | Coupon validate karne wali service. | Coupon preview endpoint CMS ko call karta hai |
| Docker | Optional but recommended | Containers run karne ka tool. | Beginner local MongoDB/Redis setup easiest banata hai |
| `mongosh` | Optional but useful | MongoDB shell. | Migration, ping, index verify |
| `redis-cli` | Optional but useful | Redis command-line client. | Redis PING, key TTL, cache debug |
| `curl` | Optional but useful | HTTP request tool. | Health check and API verify |

Simple Hinglish:

- MongoDB cart ka permanent storage hai. Agar Redis restart ho jaye to data MongoDB me safe rehta hai.
- Redis fast cache hai. `${SERVICE_NAME}` quick active cart summary store/read kar sakta hai.
- Product Service ke bina Add Item flow complete nahi hoga, kyunki cart ko product price, stock aur variant info chahiye.
- CMS Service coupon preview ke liye chahiye, warna coupon validation temporarily unavailable hoga.

## 4. Go Dependency System

This is a Go project.

### 4.1 Important Go Files

| File | Meaning |
|---|---|
| `backend/services/cart-service/go.mod` | Module name, Go version, direct and indirect dependencies |
| `backend/services/cart-service/go.sum` | Dependency checksum lock file. Isse supply-chain tampering detect hoti hai |
| `backend/go.work` | Workspace file. Multiple modules ko local development me connect karta hai |
| `backend/go.work.sum` | Workspace-level dependency checksums |

### 4.2 Go Version

`go.mod` and `go.work` both declare:

```text
go 1.26.3
```

Required action:

```bash
go version
```

Expected:

```text
go version go1.26.3 ...
```

If exact version available nahi hai, closest compatible Go version install karo. But best practice hai ki repo ke `go.mod` version ke saath match rakho.

### 4.3 Direct Go Libraries

| Dependency | Version | Required? | Why |
|---|---:|---:|---|
| `go.mongodb.org/mongo-driver` | `v1.17.9` | Yes | MongoDB connect, ping, collection schema, transactions, indexes |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Yes | Redis client, cache operations, distributed lock |

Indirect dependencies `go.sum` me MongoDB/Redis ke supporting packages hain. Beginner ko manually install nahi karna padta; Go automatically download karta hai.

### 4.4 Dependency Commands

Run from service folder:

```bash
cd backend/services/cart-service
go mod download
go mod tidy
go test ./...
go build ./cmd/server
go build ./cmd/cart-expiry-worker
```

Run from backend workspace:

```bash
cd backend
go test ./services/cart-service/...
```

### 4.5 Run Commands

Start HTTP API:

```bash
cd backend/services/cart-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Run expiry worker once:

```bash
cd backend/services/cart-service
set -a
. ./.env
set +a
go run ./cmd/cart-expiry-worker
```

### 4.6 Common Go Module Issues

| Problem | Cause | Fix |
|---|---|---|
| `go: go.mod file not found` | Wrong folder me command run hua | `cd backend/services/cart-service` |
| `module requires Go 1.26.3` | Local Go version old hai | Correct Go install karo, then `go version` verify |
| `missing go.sum entry` | Dependency checksum missing | `go mod tidy` run karo |
| `connection refused` during tests | Integration dependency running nahi hai | MongoDB/Redis start karo ya unit-only tests run karo |
| Proxy download failed | Network/proxy issue | `go env GOPROXY`, internet/proxy check karo, then `go mod download` |
| Cache corrupt | Go module cache issue | `go clean -modcache`, then `go mod download` |

## 5. Database Analysis

### 5.1 MongoDB

MongoDB required hai.

Simple Hinglish:

MongoDB ek document database hai. Isme rows/tables ki jagah JSON-like documents store hote hain. Cart ke items, totals, coupon preview, owner, status, expiry date jaise flexible fields ke liye MongoDB useful hai.

This service uses:

```text
Database: cart_db
Collection: carts
```

Why used:

- Active carts durable store me save karne ke liye
- Guest/user cart ownership maintain karne ke liye
- Cart items array as embedded document store karne ke liye
- TTL index se expired cart cleanup support karne ke liye
- Unique partial indexes se one active cart per user/session enforce karne ke liye

Required:

- HTTP API startup MongoDB connect and ping karta hai.
- `/readyz` MongoDB ping karta hai.
- Add/remove/merge/coupon flows MongoDB repository use karte hain.
- Expiry worker MongoDB me expired cart scan/update karta hai.

#### MongoDB Local Install

Windows:

1. MongoDB Community Server download karo: `https://www.mongodb.com/try/download/community`
2. Installer run karo.
3. MongoDB Compass optional install karo.
4. Service start verify:

```powershell
mongosh --eval "db.adminCommand({ ping: 1 })"
```

Linux Ubuntu/Debian:

```bash
sudo apt-get update
sudo apt-get install -y gnupg curl
# MongoDB official repo setup version-specific hota hai, latest instructions MongoDB docs se follow karo.
mongosh --eval "db.adminCommand({ ping: 1 })"
```

macOS:

```bash
brew tap mongodb/brew
brew install mongodb-community
brew services start mongodb-community
mongosh --eval "db.adminCommand({ ping: 1 })"
```

Beginner recommended approach: Docker use karo. Native install me OS-specific steps change ho sakte hain.

#### MongoDB Docker Run

```bash
docker volume create ecommerce_mongo_data

docker run -d \
  --name ecommerce-mongo \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=ecommerce_root \
  -e MONGO_INITDB_ROOT_PASSWORD=ecommerce_password \
  -v ecommerce_mongo_data:/data/db \
  mongo:7
```

Verify:

```bash
docker exec ecommerce-mongo mongosh --quiet \
  -u ecommerce_root \
  -p ecommerce_password \
  --authenticationDatabase admin \
  --eval "db.adminCommand({ ping: 1 })"
```

Expected:

```text
{ ok: 1 }
```

#### MongoDB Connection String

For service running on host machine:

```env
CART_MONGO_URI=mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin
CART_MONGO_DATABASE=cart_db
```

For service running inside same Docker Compose network:

```env
CART_MONGO_URI=mongodb://ecommerce_root:ecommerce_password@mongo:27017/cart_db?authSource=admin
CART_MONGO_DATABASE=cart_db
```

Important:

- `CART_MONGO_DATABASE` must be exactly `cart_db`.
- `config.Validate()` rejects any other database name.
- Username/password go inside `CART_MONGO_URI`.
- Host and port also go inside `CART_MONGO_URI`.

#### MongoDB Migration

Migration files:

```text
backend/services/cart-service/migrations/mongo/001_create_carts_collection.up.js
backend/services/cart-service/migrations/mongo/001_create_carts_collection.down.js
```

Run migration:

```bash
cd backend/services/cart-service
mongosh "mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin" \
  migrations/mongo/001_create_carts_collection.up.js
```

Alternative bootstrap through API:

```bash
curl -X POST http://localhost:8084/internal/v1/cart/schema/bootstrap
```

Alternative auto-bootstrap at startup:

```env
CART_BOOTSTRAP_COLLECTIONS_ON_STARTUP=true
```

Warning:

- Auto-bootstrap is helpful in local development.
- In production, migrations should be controlled through deployment/migration jobs, not random app startup.

#### MongoDB Verify Collection and Indexes

```bash
mongosh "mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin" \
  --eval "db.carts.getIndexes().map(i => i.name)"
```

Expected important indexes:

```text
idx_carts_user_status
idx_carts_guest_status
idx_carts_expires_at_ttl
uniq_active_cart_per_user
uniq_active_cart_per_guest_session
idx_carts_item_product_variant
idx_carts_updated_at
```

### 5.2 Redis

Redis required hai.

Simple Hinglish:

Redis ek very fast in-memory key-value store hai. Is project me Redis cache aur distributed lock ke liye use hota hai. Data memory me hota hai, isliye fast hai, but durable source of truth MongoDB hi hai.

Why used:

- Active user cart cache: `cart:active:user:<user_id>`
- Active guest cart cache: `cart:active:guest:<guest_session_id>`
- Cart summary cache: `cart:summary:user:<user_id>` and `cart:summary:guest:<guest_session_id>`
- Expiry worker distributed lock: default key `cart:lock:expiry-cleanup`

Required:

- HTTP API startup Redis ping karta hai.
- Expiry worker Redis ping karta hai and lock use karta hai.
- Cache refresh failures may affect performance and consistency of cached summary.

#### Redis Local Install

Windows:

- Recommended: Docker Desktop or WSL2.
- Native Redis on Windows official path simple nahi hai. WSL2 Ubuntu me install/run karna best hai.

Linux Ubuntu/Debian:

```bash
sudo apt-get update
sudo apt-get install -y redis-server
sudo systemctl enable redis-server
sudo systemctl start redis-server
redis-cli ping
```

macOS:

```bash
brew install redis
brew services start redis
redis-cli ping
```

Expected:

```text
PONG
```

#### Redis Docker Run Without Password

Matches `.env.example` where `CART_REDIS_PASSWORD=` is empty:

```bash
docker volume create ecommerce_redis_data

docker run -d \
  --name ecommerce-redis \
  -p 6379:6379 \
  -v ecommerce_redis_data:/data \
  redis:7.2-alpine \
  redis-server --appendonly yes
```

Verify:

```bash
docker exec ecommerce-redis redis-cli ping
```

#### Redis Docker Run With Password

```bash
docker run -d \
  --name ecommerce-redis \
  -p 6379:6379 \
  -v ecommerce_redis_data:/data \
  redis:7.2-alpine \
  redis-server --appendonly yes --requirepass dev_redis_password
```

Then set:

```env
CART_REDIS_PASSWORD=dev_redis_password
```

Verify:

```bash
docker exec ecommerce-redis redis-cli -a dev_redis_password ping
```

## 6. External Services

### 6.1 Product Service

Required for Add Item flow.

Simple Hinglish:

Product Service se `${SERVICE_NAME}` product title, seller, image, variant, price, currency, stock, availability data leta hai. Cart me item add karte time fresh product snapshot create hota hai.

Config:

```env
CART_PRODUCT_BASE_URL=http://localhost:8082
CART_PRODUCT_REQUEST_TIMEOUT=2s
```

Runtime call:

```text
GET {CART_PRODUCT_BASE_URL}/api/v1/products/{product_id}
```

Required response should include one of these shapes:

- root product object
- `{ "product": {...} }`
- `{ "data": {...} }`
- `{ "item": {...} }`

Important fields:

- product `id` or `product_id`
- `seller_id`
- `title` or `name`
- `status`
- `variants`
- variant `id` or `variant_id`
- variant `price.amount` or `unit_price.amount` or `price_amount`
- variant `currency`
- variant `stock_quantity` or `inventory_quantity`

If Product Service is down:

- Add item returns `PRODUCT_UNAVAILABLE` or related 503 behavior.
- Health `/healthz` may still be OK because Product Service is not pinged at startup.

### 6.2 CMS Service

Required for Coupon Preview flow.

Simple Hinglish:

CMS Service coupon rules validate karta hai. `${SERVICE_NAME}` cart subtotal and items bhejta hai, CMS valid/invalid coupon and discount return karta hai.

Config:

```env
CART_CMS_BASE_URL=http://localhost:8089
CART_CMS_VALIDATE_COUPON_PATH=/internal/v1/coupons/validate
CART_COUPON_PREVIEW_TIMEOUT=700ms
```

Runtime call:

```text
POST {CART_CMS_BASE_URL}{CART_CMS_VALIDATE_COUPON_PATH}
```

Header sent by `${SERVICE_NAME}`:

```text
X-Internal-Service: cart-service
```

If CMS Service is down:

- Coupon preview returns unavailable.
- Normal add/remove cart operations can still work.

### 6.3 Kafka, RabbitMQ, NATS, MinIO, SMTP, Stripe, Twilio, OAuth

Current `${SERVICE_NAME}` implementation inspected for `${TASK_FILE_NAME}` does not use these services directly.

Do not configure them for this task unless a future task adds them.

### 6.4 Docker

Docker optional hai, but beginners ke liye strongly recommended hai because MongoDB and Redis one-command start ho jate hain.

Current cart service folder does not include a dedicated Dockerfile. Use Docker for dependencies first. Service itself can be run directly with `go run`.

## 7. Ports and Networking

| Service | Default Port | Purpose | Config |
|---|---:|---|---|
| Cart HTTP API | `8084` | Main HTTP API, health, schema endpoints | `CART_HTTP_ADDR=:8084` |
| MongoDB | `27017` | Cart database | `CART_MONGO_URI` |
| Redis | `6379` | Cache and worker lock | `CART_REDIS_ADDR=localhost:6379` |
| Product Service | `8082` | Product lookup for add item | `CART_PRODUCT_BASE_URL=http://localhost:8082` |
| CMS Service | `8089` | Coupon validation | `CART_CMS_BASE_URL=http://localhost:8089` |

Port conflict check:

```bash
lsof -i :8084
lsof -i :27017
lsof -i :6379
```

If `lsof` missing:

```bash
netstat -tulpn | grep 8084
```

Change Cart API port:

```env
CART_HTTP_ADDR=:8094
```

Then verify:

```bash
curl http://localhost:8094/healthz
```

Docker networking note:

- Host machine se MongoDB: `localhost:27017`
- Container-to-container network se MongoDB: `mongo:27017`
- Host machine se Redis: `localhost:6379`
- Container-to-container network se Redis: `redis:6379`

## 8. Complete Environment Variables

### 8.1 Where to Create `.env`

Create:

```text
backend/services/cart-service/.env
```

Use existing example:

```bash
cd backend/services/cart-service
cp .env.example .env
```

Very important:

The Go code uses `os.Getenv`. There is no `godotenv` loader in the implementation. Sirf `.env` file create karna enough nahi hai. You must export/source it before `go run`.

Linux/macOS:

```bash
cd backend/services/cart-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Windows PowerShell:

```powershell
cd backend/services/cart-service
Get-Content .env | ForEach-Object {
  if ($_ -match '^\s*([^#=]+)=(.*)$') {
    [Environment]::SetEnvironmentVariable($matches[1], $matches[2], 'Process')
  }
}
go run ./cmd/server
```

### 8.2 Complete `.env` Example

```env
CART_SERVICE_NAME=cart-service
ENVIRONMENT=local
CART_HTTP_ADDR=:8084

CART_MONGO_URI=mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin
CART_MONGO_DATABASE=cart_db
CART_MONGO_CONNECT_TIMEOUT=5s
CART_MONGO_PING_TIMEOUT=5s

CART_REDIS_ADDR=localhost:6379
CART_REDIS_PASSWORD=
CART_REDIS_DB=0
CART_REDIS_DIAL_TIMEOUT=5s
CART_REDIS_READ_TIMEOUT=3s
CART_REDIS_WRITE_TIMEOUT=3s
CART_REDIS_PING_TIMEOUT=3s

CART_PRODUCT_BASE_URL=http://localhost:8082
CART_PRODUCT_REQUEST_TIMEOUT=2s

CART_CMS_BASE_URL=http://localhost:8089
CART_CMS_VALIDATE_COUPON_PATH=/internal/v1/coupons/validate
CART_COUPON_PREVIEW_TIMEOUT=700ms
CART_CMS_REQUEST_TIMEOUT=700ms

CART_EXPIRY_TTL=2160h
CART_USER_EXPIRY_TTL=2160h
CART_GUEST_EXPIRY_TTL=720h
CART_DEFAULT_CURRENCY=INR
CART_MAX_SAVE_ATTEMPTS=3
CART_ALLOW_GUEST_CART_ID_ONLY_MERGE=false

CART_CLEANUP_INTERVAL=15m
CART_CLEANUP_BATCH_SIZE=500
CART_CLEANUP_LOCK_TTL=10m
CART_CLEANUP_RUN_TIMEOUT=2m
CART_CLEANUP_LOCK_KEY=cart:lock:expiry-cleanup

CART_ACTIVE_CACHE_TTL=15m
CART_GUEST_ACTIVE_CACHE_TTL=30m
GUEST_CART_CACHE_TTL=30m
CART_SUMMARY_CACHE_TTL=5m

CART_BOOTSTRAP_COLLECTIONS_ON_STARTUP=false

CART_HTTP_READ_TIMEOUT=5s
CART_HTTP_WRITE_TIMEOUT=10s
CART_HTTP_IDLE_TIMEOUT=60s
CART_SHUTDOWN_TIMEOUT=10s
```

### 8.3 Env Variable Explanation

| Variable | Required? | Example | Purpose | Security note |
|---|---:|---|---|---|
| `CART_SERVICE_NAME` | Yes | `cart-service` | Logs and service identity | Not secret |
| `ENVIRONMENT` | Optional | `local` | Runtime environment name | Not secret |
| `APP_ENV` | Optional fallback | `local` | Used if `ENVIRONMENT` missing | Not secret |
| `CART_HTTP_ADDR` | Yes | `:8084` | HTTP bind address | Avoid exposing publicly in local |
| `CART_MONGO_URI` | Yes if `MONGO_URI` absent | `mongodb://...` | MongoDB connection string | Contains DB password, do not commit real prod value |
| `MONGO_URI` | Optional fallback | `mongodb://...` | Legacy fallback for MongoDB | Secret |
| `CART_MONGO_DATABASE` | Yes | `cart_db` | Mongo database name | Must be `cart_db` |
| `CART_MONGO_CONNECT_TIMEOUT` | Optional | `5s` | Mongo connect timeout | Not secret |
| `CART_MONGO_PING_TIMEOUT` | Optional | `5s` | Mongo ping timeout | Not secret |
| `CART_REDIS_ADDR` | Yes if `REDIS_ADDR` absent | `localhost:6379` | Redis host and port | Not secret |
| `REDIS_ADDR` | Optional fallback | `localhost:6379` | Legacy fallback Redis address | Not secret |
| `CART_REDIS_PASSWORD` | Optional | empty or `dev_redis_password` | Redis auth password | Secret if set |
| `REDIS_PASSWORD` | Optional fallback | `dev_redis_password` | Legacy Redis password | Secret |
| `CART_REDIS_DB` | Optional | `0` | Redis logical DB number | Not secret |
| `REDIS_DB` | Optional fallback | `0` | Legacy Redis DB number | Not secret |
| `CART_REDIS_DIAL_TIMEOUT` | Optional | `5s` | Redis connect timeout | Not secret |
| `CART_REDIS_READ_TIMEOUT` | Optional | `3s` | Redis read timeout | Not secret |
| `CART_REDIS_WRITE_TIMEOUT` | Optional | `3s` | Redis write timeout | Not secret |
| `CART_REDIS_PING_TIMEOUT` | Optional | `3s` | Redis startup ping timeout | Not secret |
| `CART_PRODUCT_BASE_URL` | Yes if `PRODUCT_SERVICE_BASE_URL` absent | `http://localhost:8082` | Product lookup base URL | Not secret |
| `PRODUCT_SERVICE_BASE_URL` | Optional fallback | `http://localhost:8082` | Legacy Product URL | Not secret |
| `CART_PRODUCT_REQUEST_TIMEOUT` | Optional | `2s` | Product request timeout | Not secret |
| `CART_CMS_BASE_URL` | Yes if `CMS_SERVICE_BASE_URL` absent | `http://localhost:8089` | CMS coupon validation base URL | Not secret unless private endpoint |
| `CMS_SERVICE_BASE_URL` | Optional fallback | `http://localhost:8089` | Legacy CMS URL | Not secret unless private endpoint |
| `CART_CMS_VALIDATE_COUPON_PATH` | Yes | `/internal/v1/coupons/validate` | CMS coupon validation path | Not secret |
| `CART_COUPON_PREVIEW_TIMEOUT` | Optional | `700ms` | Coupon validation timeout | Not secret |
| `CART_CMS_REQUEST_TIMEOUT` | Optional fallback | `700ms` | Legacy CMS timeout fallback | Not secret |
| `CART_EXPIRY_TTL` | Optional | `2160h` | Legacy/default cart expiry TTL | Not secret |
| `CART_USER_EXPIRY_TTL` | Optional | `2160h` | Logged-in cart expiry TTL | Not secret |
| `CART_GUEST_EXPIRY_TTL` | Optional | `720h` | Guest cart expiry TTL | Not secret |
| `CART_DEFAULT_CURRENCY` | Optional | `INR` | Empty cart/totals currency | Must be 3 uppercase letters |
| `CART_MAX_SAVE_ATTEMPTS` | Optional | `3` | Retry count for version conflicts | Not secret |
| `CART_ALLOW_GUEST_CART_ID_ONLY_MERGE` | Optional | `false` | Allows less strict merge lookup | Keep `false` for safer local/prod behavior |
| `CART_CLEANUP_INTERVAL` | Optional | `15m` | Expiry scheduler interval config | Not secret |
| `CART_CLEANUP_BATCH_SIZE` | Optional | `500` | Expired cart batch size | Not secret |
| `CART_CLEANUP_LOCK_TTL` | Optional | `10m` | Redis lock TTL for worker | Not secret |
| `CART_CLEANUP_RUN_TIMEOUT` | Optional | `2m` | Worker max run timeout | Not secret |
| `CART_CLEANUP_LOCK_KEY` | Optional | `cart:lock:expiry-cleanup` | Redis distributed lock key | Not secret |
| `CART_ACTIVE_CACHE_TTL` | Optional | `15m` | User active cart cache TTL | Not secret |
| `CART_GUEST_ACTIVE_CACHE_TTL` | Optional | `30m` | Guest active cart cache TTL | Not secret |
| `GUEST_CART_CACHE_TTL` | Optional fallback | `30m` | Legacy guest cache TTL fallback | Not secret |
| `CART_SUMMARY_CACHE_TTL` | Optional | `5m` | Cart summary cache TTL | Not secret |
| `CART_BOOTSTRAP_COLLECTIONS_ON_STARTUP` | Optional | `false` | Auto-create Mongo schema/indexes at startup | Keep `false` in production |
| `CART_HTTP_READ_TIMEOUT` | Optional | `5s` | HTTP read timeout | Not secret |
| `CART_HTTP_WRITE_TIMEOUT` | Optional | `10s` | HTTP write timeout | Not secret |
| `CART_HTTP_IDLE_TIMEOUT` | Optional | `60s` | HTTP idle connection timeout | Not secret |
| `CART_SHUTDOWN_TIMEOUT` | Optional | `10s` | Graceful shutdown timeout | Not secret |

Common `.env` mistakes:

- `.env` created but not sourced/exported.
- Mongo URI uses `mongo:27017` while service runs on host. Use `localhost:27017`.
- Mongo URI uses `localhost:27017` while service runs inside Docker Compose. Use `mongo:27017`.
- `CART_MONGO_DATABASE` changed from `cart_db`. Service will fail validation.
- Redis password set in Docker but `CART_REDIS_PASSWORD` empty.
- Duration values written as `5` instead of `5s`. Go expects units like `ms`, `s`, `m`, `h`.
- `CART_DEFAULT_CURRENCY=inr` lowercase. Config normalizes to uppercase, but keep uppercase for clarity.

## 9. Docker and DevOps Setup

### 9.1 Docker Compose Example for Local Dependencies

If `infra/compose/docker-compose.local.yml` is not present, beginners can use this example as local dependency stack.

```yaml
services:
  mongo:
    image: mongo:7
    container_name: ecommerce-mongo
    restart: unless-stopped
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: ecommerce_root
      MONGO_INITDB_ROOT_PASSWORD: ecommerce_password
    volumes:
      - mongo_data:/data/db
    healthcheck:
      test: ["CMD-SHELL", "mongosh --quiet -u ecommerce_root -p ecommerce_password --authenticationDatabase admin --eval 'db.adminCommand({ ping: 1 }).ok' | grep 1"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7.2-alpine
    container_name: ecommerce-redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    command: ["redis-server", "--appendonly", "yes"]
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  mongo_data:
  redis_data:
```

Start:

```bash
docker compose up -d
```

Stop:

```bash
docker compose down
```

Stop and delete volumes:

```bash
docker compose down -v
```

Logs:

```bash
docker compose logs -f mongo
docker compose logs -f redis
```

Status:

```bash
docker compose ps
docker ps
```

### 9.2 Docker Volumes

| Volume | Purpose |
|---|---|
| `mongo_data` | MongoDB database files persist after container restart |
| `redis_data` | Redis AOF persistence files |

Without volumes:

- Container delete karne par data delete ho sakta hai.

### 9.3 Docker Network

Docker Compose default private network create karta hai.

Inside this network:

```env
CART_MONGO_URI=mongodb://ecommerce_root:ecommerce_password@mongo:27017/cart_db?authSource=admin
CART_REDIS_ADDR=redis:6379
```

Host machine se:

```env
CART_MONGO_URI=mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin
CART_REDIS_ADDR=localhost:6379
```

## 10. Project Run Instructions

### Step 1: Clone Repository

```bash
git clone <repo-url>
cd Ecommerce
```

### Step 2: Verify Tools

```bash
go version
docker --version
docker compose version
curl --version
```

Optional:

```bash
mongosh --version
redis-cli --version
```

### Step 3: Install Go Dependencies

```bash
cd backend/services/cart-service
go mod download
go mod tidy
```

### Step 4: Start MongoDB and Redis

Using Docker run:

```bash
docker volume create ecommerce_mongo_data
docker volume create ecommerce_redis_data

docker run -d --name ecommerce-mongo -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=ecommerce_root \
  -e MONGO_INITDB_ROOT_PASSWORD=ecommerce_password \
  -v ecommerce_mongo_data:/data/db \
  mongo:7

docker run -d --name ecommerce-redis -p 6379:6379 \
  -v ecommerce_redis_data:/data \
  redis:7.2-alpine redis-server --appendonly yes
```

Verify:

```bash
docker exec ecommerce-mongo mongosh --quiet -u ecommerce_root -p ecommerce_password \
  --authenticationDatabase admin --eval "db.adminCommand({ ping: 1 })"

docker exec ecommerce-redis redis-cli ping
```

### Step 5: Create `.env`

```bash
cd backend/services/cart-service
cp .env.example .env
```

Check these values:

```env
CART_MONGO_URI=mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin
CART_REDIS_ADDR=localhost:6379
CART_PRODUCT_BASE_URL=http://localhost:8082
CART_CMS_BASE_URL=http://localhost:8089
```

### Step 6: Export `.env`

```bash
set -a
. ./.env
set +a
```

### Step 7: Run MongoDB Migration

```bash
mongosh "$CART_MONGO_URI" migrations/mongo/001_create_carts_collection.up.js
```

Verify schema endpoint later:

```bash
curl http://localhost:8084/internal/v1/cart/schema
```

### Step 8: Start Product and CMS Services

If actual Product/CMS services are available:

```bash
# Start Product Service on http://localhost:8082
# Start CMS Service on http://localhost:8089
```

If they are not available:

- Cart API can start because it validates URL format, not remote service health.
- `/api/v1/cart/items` will fail when Product Service is unavailable.
- `/api/v1/cart/coupons/preview` will fail when CMS Service is unavailable.

For only health/schema checks, Product/CMS can be absent.

### Step 9: Start Cart HTTP API

```bash
go run ./cmd/server
```

Expected log includes:

```text
cart.http.starting
```

### Step 10: Verify Health

```bash
curl http://localhost:8084/healthz
curl http://localhost:8084/readyz
```

Expected:

```json
{"status":"ok"}
```

and:

```json
{"status":"ready"}
```

### Step 11: Bootstrap Schema Through API if Migration Was Skipped

```bash
curl -X POST http://localhost:8084/internal/v1/cart/schema/bootstrap
```

### Step 12: Run Expiry Worker

In another terminal:

```bash
cd backend/services/cart-service
set -a
. ./.env
set +a
go run ./cmd/cart-expiry-worker
```

This worker runs cleanup once and exits.

## 11. API Verification Commands

Health:

```bash
curl http://localhost:8084/healthz
```

Readiness:

```bash
curl http://localhost:8084/readyz
```

Schema spec:

```bash
curl http://localhost:8084/internal/v1/cart/schema
```

Bootstrap schema:

```bash
curl -X POST http://localhost:8084/internal/v1/cart/schema/bootstrap
```

Add item example:

```bash
curl -X POST http://localhost:8084/api/v1/cart/items \
  -H "Content-Type: application/json" \
  -H "X-Guest-Session-ID: guest-session-123" \
  -d '{
    "product_id": "prod_123",
    "variant_id": "var_123",
    "quantity": 1
  }'
```

Important:

- Product Service must return product and variant data for this to succeed.
- If Product Service is missing, expect product unavailable error.

Remove item example:

```bash
curl -X DELETE http://localhost:8084/api/v1/cart/items/item_123 \
  -H "X-Guest-Session-ID: guest-session-123"
```

Coupon preview example:

```bash
curl -X POST http://localhost:8084/api/v1/cart/coupons/preview \
  -H "Content-Type: application/json" \
  -H "X-Guest-Session-ID: guest-session-123" \
  -d '{
    "cart_id": "cart_123",
    "coupon_code": "WELCOME10"
  }'
```

Merge guest cart example:

```bash
curl -X POST http://localhost:8084/api/v1/cart/merge \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_123" \
  -H "X-Guest-Session-ID: guest-session-123" \
  -d '{
    "guest_cart_id": "cart_guest_123"
  }'
```

## 12. Common Errors and Fixes

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `CART_MONGO_URI or MONGO_URI is required` | Env not exported or missing Mongo URI | Source `.env` before `go run` | Always run `set -a; . ./.env; set +a` |
| `CART_MONGO_DATABASE must be "cart_db"` | Database name changed | Set `CART_MONGO_DATABASE=cart_db` | Do not rename Cart DB without code change |
| `cart.mongo.ping_failed` | MongoDB not running or wrong URI | Start MongoDB and verify `mongosh` ping | Keep Docker container running |
| `connection refused localhost:27017` | MongoDB container down or port mismatch | `docker start ecommerce-mongo` | `docker ps` before service start |
| `Authentication failed` | Mongo username/password wrong | Fix credentials in `CART_MONGO_URI` | Keep `.env` and Docker env aligned |
| `cart.redis.ping_failed` | Redis down/wrong address/password | Start Redis, fix `CART_REDIS_ADDR`/password | Verify `redis-cli ping` |
| `NOAUTH Authentication required` | Redis has password but env empty | Set `CART_REDIS_PASSWORD` | Decide password/no-password and document it |
| `port already in use :8084` | Another service using port | Change `CART_HTTP_ADDR=:8094` or stop other process | Check ports before start |
| `go: go.mod file not found` | Wrong directory | `cd backend/services/cart-service` | Run commands from service folder |
| `module requires Go 1.26.3` | Old Go version | Install matching Go version | Use toolchain manager if available |
| `go mod download failed` | Network/proxy issue | Check internet/proxy, `go env GOPROXY` | Cache deps in CI |
| `request body must be valid JSON` | Invalid curl body | Validate JSON quotes/braces | Keep sample requests copied carefully |
| `PRODUCT_NOT_FOUND` | Product Service cannot find product | Use real product ID | Seed product data first |
| `CART_TEMPORARILY_UNAVAILABLE` | Mongo/Redis/Product issue | Check logs and dependency health | Monitor dependency status |
| `COUPON_PREVIEW_UNAVAILABLE` | CMS down or invalid response shape | Start CMS and verify endpoint | Add CMS health checks |
| `CART_VERSION_CONFLICT` | Concurrent cart update conflict | Retry request | Client can retry with backoff |
| Migration failed with duplicate/index error | Existing collection/index conflict | Inspect indexes, run down/up carefully | Use migration history in production |
| Docker daemon not running | Docker Desktop/Engine stopped | Start Docker | Verify `docker info` |
| Permission denied Docker socket | User not in docker group | Linux: add user to docker group or use sudo | Configure Docker after install |

## 13. Security and Configuration Audit

Findings from inspected implementation:

| Area | Observation | Risk | Suggested fix |
|---|---|---|---|
| `.env` loading | Code uses `os.Getenv`; `.env.example` exists but no auto-loader | Beginner may think `.env` is automatically loaded | Document source/export flow or add explicit local loader if desired |
| Mongo credentials | `.env.example` contains dev username/password | Safe for local, unsafe if reused in prod | Use secret manager for prod, never commit real prod secrets |
| Redis password | Example is empty | Local OK, production unsafe | Set Redis auth/TLS in staging/prod |
| Product/CMS URLs | Required URL format but not startup health checked | Add/coupon fail later at runtime | Add dependency health/readiness checks if needed |
| `CART_BOOTSTRAP_COLLECTIONS_ON_STARTUP` | Available but default false | If migration skipped, schema may be missing | Run migration or bootstrap endpoint in local setup |
| `CART_ALLOW_GUEST_CART_ID_ONLY_MERGE` | Defaults false | Safer default | Keep false unless explicitly needed |
| HTTP binding | `:8084` binds all interfaces | In local Docker/VM may expose externally | Bind to `127.0.0.1:8084` for local-only access if needed |
| Dockerfile | No dedicated cart-service Dockerfile found | Service containerization not ready | Add Dockerfile in future DevOps task |
| Health checks | `/healthz` and `/readyz` exist | Good baseline | Extend readiness to Redis/Product/CMS if production requires |
| Logs | Uses structured JSON `slog` | Good baseline | Add request IDs/tracing in gateway integration |
| TLS | Service uses plain HTTP locally | OK for local | Use ingress/API gateway TLS in production |

Professional recommendation:

- Local development can use plain Docker Mongo/Redis with dev credentials.
- Production must use managed secrets, private networking, Redis auth/TLS, Mongo auth/TLS, backups, migration jobs, and observability.

## 14. Best Practices for Beginners

- Never commit real `.env` files.
- Keep `.env.example` safe and local-only.
- Always verify dependency containers with `docker ps`.
- Use `mongosh` ping before blaming Go code.
- Use `redis-cli ping` before debugging cache code.
- Keep MongoDB database name `cart_db`.
- Do not use production credentials in local machine.
- Prefer Docker volumes for MongoDB/Redis data persistence.
- Run migrations before testing cart mutations.
- Keep `CART_BOOTSTRAP_COLLECTIONS_ON_STARTUP=false` in production.
- Use strong Redis password in staging/prod.
- Keep Product Service and CMS URLs environment-specific.
- Use duration units: `700ms`, `5s`, `15m`, `2160h`.
- Run `go test ./...` after dependency changes.
- Run `go mod tidy` only after intentional dependency changes.
- Check service logs first; startup validation errors are explicit.
- For Docker Compose, use service hostnames inside containers and `localhost` from host machine.

## 15. Final Setup Checklist

- [ ] Repository cloned
- [ ] Go `1.26.3` or compatible version installed
- [ ] Docker installed and running
- [ ] `backend/services/cart-service/go.mod` dependencies downloaded
- [ ] MongoDB running on port `27017`
- [ ] Redis running on port `6379`
- [ ] `backend/services/cart-service/.env` created
- [ ] `.env` exported/sourced before running service
- [ ] `CART_MONGO_URI` points to correct Mongo host
- [ ] `CART_MONGO_DATABASE=cart_db`
- [ ] `CART_REDIS_ADDR` points to correct Redis host
- [ ] Redis password config matches actual Redis server
- [ ] Product Service URL configured
- [ ] CMS Service URL configured
- [ ] Mongo migration or schema bootstrap completed
- [ ] Cart HTTP API starts on port `8084`
- [ ] `/healthz` returns OK
- [ ] `/readyz` returns ready
- [ ] Schema endpoint returns cart collection spec
- [ ] Expiry worker can run once
- [ ] Common errors section reviewed
- [ ] Real secrets are not committed

## 16. Quick Command Summary

```bash
cd backend/services/cart-service

go mod download

docker volume create ecommerce_mongo_data
docker volume create ecommerce_redis_data

docker run -d --name ecommerce-mongo -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=ecommerce_root \
  -e MONGO_INITDB_ROOT_PASSWORD=ecommerce_password \
  -v ecommerce_mongo_data:/data/db \
  mongo:7

docker run -d --name ecommerce-redis -p 6379:6379 \
  -v ecommerce_redis_data:/data \
  redis:7.2-alpine redis-server --appendonly yes

cp .env.example .env
set -a
. ./.env
set +a

mongosh "$CART_MONGO_URI" migrations/mongo/001_create_carts_collection.up.js

go run ./cmd/server
```

In another terminal:

```bash
curl http://localhost:8084/healthz
curl http://localhost:8084/readyz
curl http://localhost:8084/internal/v1/cart/schema
```

## 17. What Is Not Required for This Task

Not required for current `${SERVICE_NAME}` Task 1 setup:

- Kafka
- RabbitMQ
- NATS
- MinIO
- Elasticsearch
- Typesense
- SMTP
- Stripe
- Twilio
- OAuth provider setup
- Kubernetes
- Production Docker image build

These may be needed by other services/tasks, but not directly by this inspected cart task/service setup.
