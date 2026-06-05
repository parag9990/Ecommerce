# Search Service - Task 1 Dependency and Setup Guide

Source implementation file:

```text
TaskImplementation/Search Service/task1.md
```

Generated dependency file:

```text
TaskImplementation/Search Service/task1_Dependency.md
```

## 0. Important Scope Note

Task 1 ka original output schema documentation hai. Is task me Search Service ka product search schema, Typesense collection fields, facets, sorting, typo tolerance, aur synonym model define kiya gaya tha.

Current repository me `backend/services/search-service` code bhi present hai. Isliye ye guide do cheeze clearly explain karti hai:

| Area | Meaning |
|---|---|
| Task 1 dependency | Schema guide samajhne ke liye kya tools/services required hain |
| Runnable Search Service setup | Local machine par existing Go Search Service run karne ke liye kya dependencies chahiye |

Beginner-friendly rule: Task 1 sirf documentation blueprint tha, but actual service run karne ke liye Typesense, Redis, RabbitMQ, Product Service, aur optional Session Service ka setup samajhna zaruri hai.

## 1. Project Tech Stack Analysis

| Technology | Required? | Why Used | Beginner Explanation |
|---|---:|---|---|
| Go | Yes | Search Service backend code Go me hai | Go ek compiled backend language hai. Isme fast APIs, workers, aur CLI jobs banana easy hota hai. |
| Go modules | Yes | Dependencies manage karne ke liye | `go.mod` project ka dependency list hai, jaise Node me `package.json`. |
| `net/http` | Yes | HTTP API server | Is service me Gin/Fiber use nahi hua. Standard Go HTTP server use hua hai. |
| Typesense | Yes | Product search engine | Typesense ek fast search engine hai jo typo tolerance, facets, filters, sorting, aur synonyms support karta hai. |
| Typesense Go client | Yes | Go service se Typesense API call karne ke liye | Code ko manually HTTP request likhne ki zarurat nahi, client library help karti hai. |
| Redis | Yes for full service | Autocomplete cache, zero-result dedupe, processed event store, reindex lock | Redis in-memory key-value store hai. Fast temporary data, cache, lock, aur dedupe ke liye use hota hai. |
| RabbitMQ | Yes by default | Product events consume karke search index update karne ke liye | RabbitMQ message broker hai. Product Service events bhejega, Search Service consume karega. |
| Product Service | Yes for real search/reindex | Canonical product summary fetch karne aur full catalog export ke liye | Search index fast lookup deta hai, but product ka source of truth Product Service hai. |
| Session Service | Optional, enabled by default | Zero-result search analytics bhejne ke liye | Jab search result zero aaye, analytics event Session Service ko bheja ja sakta hai. |
| Docker Compose | Recommended | Local Typesense, Redis, RabbitMQ start karne ke liye | Docker se infra services one command me run ho jati hain. Beginners ke liye easiest path hai. |
| Kubernetes | Optional | Production/dev cluster deployment | K8s manifests config, secrets, Typesense StatefulSet, aur reindex job ke liye present hain. |
| Mermaid | Optional | Task documentation diagrams render karne ke liye | Markdown diagrams banane ka syntax hai. Code run karne ke liye required nahi. |

## 2. Go Dependency System

Search Service Go project hai.

### Important files

| File | Purpose |
|---|---|
| `backend/services/search-service/go.mod` | Module name, Go version, direct dependencies |
| `backend/services/search-service/go.sum` | Downloaded dependency checksums. Security ke liye exact package hash store hota hai. |
| `backend/go.work` | Multiple backend modules ko workspace me connect karta hai |

### Go version

`go.mod` me:

```text
go 1.26.3
```

Use same or compatible Go version. Version mismatch se build fail ho sakta hai.

### Direct Go dependencies

| Package | Why Used |
|---|---|
| `github.com/typesense/typesense-go/v2` | Typesense collection, document, search, synonym, alias APIs ke liye |
| `github.com/rabbitmq/amqp091-go` | RabbitMQ AMQP consumer ke liye |

Indirect dependencies `go.sum` me aa sakti hain. Unhe manually edit mat karo.

### Common Go commands

Run from:

```bash
cd backend/services/search-service
```

Download dependencies:

```bash
go mod download
```

Clean and sync dependencies:

```bash
go mod tidy
```

Run tests:

```bash
go test ./...
```

Build server binary:

```bash
go build ./cmd/server
```

Run server:

```bash
go run ./cmd/server
```

Run reindex CLI:

```bash
go run ./cmd/reindex --mode=alias --batch-size=500 --reason=local-seed --actor-id=local-dev
```

Dry-run reindex without writing to Typesense:

```bash
go run ./cmd/reindex --dry-run --reason=local-check --actor-id=local-dev
```

### Why Go dependencies fail

| Problem | Cause | Fix |
|---|---|---|
| `go: module requires go >= ...` | Installed Go version old hai | New Go install karo, then `go version` verify karo |
| `go mod download` network error | Internet/proxy issue | Network check karo, `GOPROXY=https://proxy.golang.org,direct` set karo |
| checksum mismatch | `go.sum` package hash mismatch | Package source verify karo, then `go clean -modcache` and retry |
| package not found | Wrong module path or stale import | `go mod tidy` run karo |
| workspace confusion | Root `go.work` and service module mismatch | Service folder se command run karo ya `backend` root se workspace commands run karo |

## 3. Database and Storage Analysis

Search Service directly MySQL, PostgreSQL, MongoDB, SQLite, ya Cassandra use nahi karta.

| Storage/Database | Directly Used? | Required? | Purpose |
|---|---:|---:|---|
| Typesense | Yes | Yes | Search index database for products and popular queries |
| Redis | Yes | Yes for full service | Cache, dedupe, processed event IDs, reindex locks |
| MySQL/PostgreSQL/MongoDB | No | No for this service | Product data ka source Product Service ke paas hai |

### 3.1 Typesense

#### A. What it is

Typesense ek open-source search engine hai. Simple Hinglish me: ye product search ke liye optimized database jaisa hai jisme text search, typo handling, filters, facets, aur sorting fast hoti hai.

#### B. Why this project uses it

Task 1 me `products` collection schema define kiya gaya hai. Existing code startup par Typesense me ye collections ensure karta hai:

| Collection | Purpose |
|---|---|
| `products` | Searchable product documents |
| `popular_queries` | Autocomplete/popular search query support |

#### C. Required or optional

Required. `TYPESENSE_API_KEY` missing hua to service start nahi hoti.

#### D. Local installation

Recommended beginner approach: Docker use karo.

Windows:

1. Docker Desktop install karo.
2. WSL2 backend enable karo.
3. Compose command run karo from repo root.

Linux:

1. Docker Engine install karo.
2. User ko `docker` group me add karo if needed.
3. Compose command run karo.

macOS:

1. Docker Desktop install karo.
2. Compose command run karo.

#### E. Docker setup

Single container:

```bash
docker run --name ecommerce-typesense \
  -p 8108:8108 \
  -v typesense_data:/data \
  typesense/typesense:27.1 \
  --data-dir=/data \
  --api-key=dev-typesense-key \
  --enable-cors
```

Project Compose setup already exists:

```bash
cp infra/compose/.env.local.example infra/compose/.env.local
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d typesense
```

#### F. Start commands

```bash
docker start ecommerce-typesense
```

or:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d typesense
```

#### G. Verify running

```bash
curl http://localhost:8108/health
```

Expected response contains:

```json
{"ok":true}
```

Check collections:

```bash
curl -H "X-TYPESENSE-API-KEY: dev-typesense-key" http://localhost:8108/collections
```

#### H. Default port

| Port | Purpose |
|---:|---|
| `8108` | Typesense API |
| `8107` | Typesense cluster peering in Kubernetes |

#### I. Connection string format

This service supports either full URL:

```env
TYPESENSE_URL=http://localhost:8108
```

or host/port/protocol:

```env
TYPESENSE_HOST=localhost
TYPESENSE_PORT=8108
TYPESENSE_PROTOCOL=http
```

#### J. Where to place credentials

Local Go run:

```env
TYPESENSE_API_KEY=dev-typesense-key
```

Docker Compose:

```text
infra/compose/.env.local
```

Kubernetes:

```text
infra/k8s/core/search-service/secret.example.yaml
infra/k8s/data/typesense/secret.example.yaml
```

Never commit real API keys.

### 3.2 Redis

#### A. What it is

Redis ek fast in-memory key-value store hai. Simple explanation: chhota temporary data, cache, locks, aur duplicate-event tracking ke liye use hota hai.

#### B. Why this project uses it

| Use Case | Redis Purpose |
|---|---|
| Autocomplete | Prefix cache and empty-query cache |
| Zero-result analytics | Duplicate zero-result events avoid karna |
| Product event consumer | Already processed event IDs store karna |
| Reindex | Distributed lock so multiple reindex jobs same time na chale |

#### C. Required or optional

Full Search Service ke liye required. If `SEARCH_INDEXER_ENABLED=true`, startup Redis ping karta hai through product consumer setup.

#### D. Local installation

Recommended: Docker.

Windows:

1. Docker Desktop use karo, or WSL2 me Redis install karo.
2. Native Windows Redis avoid karo unless team approved package use kar rahi ho.

Linux:

```bash
sudo apt update
sudo apt install redis-server redis-tools
redis-server --version
```

macOS:

```bash
brew install redis
brew services start redis
```

#### E. Docker setup

```bash
docker run --name ecommerce-redis \
  -p 6379:6379 \
  -v redis_data:/data \
  redis:7.4-alpine \
  redis-server --appendonly yes --requirepass dev_redis_password
```

Project Compose:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d redis
```

#### F. Start commands

```bash
docker start ecommerce-redis
```

#### G. Verify running

```bash
docker exec -it ecommerce-redis redis-cli -a dev_redis_password ping
```

Expected:

```text
PONG
```

#### H. Default port

```text
6379
```

#### I. Connection string format

This code uses address plus password, not a Redis URL:

```env
SEARCH_REDIS_ADDR=localhost:6379
SEARCH_REDIS_PASSWORD=dev_redis_password
SEARCH_REDIS_DB=0
```

#### J. Where to place credentials

Local:

```text
backend/services/search-service/.env
```

Compose:

```text
infra/compose/.env.local
```

Kubernetes:

```text
SEARCH_REDIS_PASSWORD in search-service-secret
```

### 3.3 RabbitMQ storage note

RabbitMQ is a broker, but it stores queue state on disk through its Docker volume. In Compose this volume is:

```text
rabbitmq_data
```

Do not delete this volume unless you are okay losing local queue messages.

## 4. Environment Variables

### Where `.env` should be created

For running Go locally:

```text
backend/services/search-service/.env
```

Important: The Go code does not auto-load `.env`. It reads OS environment variables using `os.Getenv`.

Load local `.env` in Bash:

```bash
cd backend/services/search-service
set -a
. ./.env
set +a
go run ./cmd/server
```

For Docker Compose infra:

```text
infra/compose/.env.local
```

Start Compose with:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d
```

### Complete local `.env` example

Create this file at `backend/services/search-service/.env` for local Go run:

```env
# HTTP server
SEARCH_HTTP_ADDR=:8085
SEARCH_HTTP_READ_TIMEOUT=5s
SEARCH_HTTP_WRITE_TIMEOUT=10s
SEARCH_HTTP_IDLE_TIMEOUT=60s
SEARCH_SHUTDOWN_TIMEOUT=10s

# Typesense
TYPESENSE_URL=
TYPESENSE_HOST=localhost
TYPESENSE_PORT=8108
TYPESENSE_PROTOCOL=http
TYPESENSE_API_KEY=dev-typesense-key
TYPESENSE_PRODUCTS_COLLECTION=products
TYPESENSE_POPULAR_QUERIES_COLLECTION=popular_queries
TYPESENSE_TIMEOUT_MS=300

# Product Service
PRODUCT_SERVICE_URL=http://localhost:8082
PRODUCT_SERVICE_BATCH_GET_PATH=/internal/v1/products:batchGet
PRODUCT_SERVICE_SEARCH_EXPORT_PATH=/internal/v1/products/search-export
PRODUCT_SERVICE_TIMEOUT_MS=300
PRODUCT_SERVICE_SEARCH_EXPORT_TIMEOUT_MS=2000

# Search behavior
SEARCH_DEFAULT_PAGE_SIZE=20
SEARCH_MAX_PAGE_SIZE=100
SEARCH_AUTOCOMPLETE_DEFAULT_LIMIT=8
SEARCH_AUTOCOMPLETE_MAX_LIMIT=10
SEARCH_AUTOCOMPLETE_PREFIX_CACHE_TTL=1m
SEARCH_AUTOCOMPLETE_EMPTY_CACHE_TTL=5m
SEARCH_AUTOCOMPLETE_TYPESENSE_TIMEOUT_MS=150
SEARCH_AUTOCOMPLETE_CACHE_TIMEOUT_MS=50
SEARCH_ZERO_RESULT_TRACKING_ENABLED=true
SEARCH_ZERO_RESULT_DEDUPE_TTL=30m
SEARCH_ZERO_RESULT_SEND_TIMEOUT_MS=150
SEARCH_ZERO_RESULT_QUEUE_SIZE=1024
SEARCH_ZERO_RESULT_WORKERS=2

# Session Service for zero-result analytics
SESSION_SERVICE_URL=http://localhost:8086
SESSION_SERVICE_INGEST_PATH=/api/v1/sessions/events
SESSION_SERVICE_TIMEOUT_MS=150

# Admin endpoints
SEARCH_ADMIN_TIMEOUT_MS=500
SEARCH_ADMIN_AUTH_ENABLED=true
SEARCH_ADMIN_MUTATION_RATE_LIMIT=30
SEARCH_ADMIN_MUTATION_RATE_WINDOW=1m

# Product indexer and RabbitMQ
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

# Redis
SEARCH_REDIS_ADDR=localhost:6379
SEARCH_REDIS_PASSWORD=dev_redis_password
SEARCH_REDIS_DB=0
SEARCH_REDIS_DIAL_TIMEOUT=2s
SEARCH_REDIS_READ_TIMEOUT=2s
SEARCH_REDIS_WRITE_TIMEOUT=2s

# Reindex
SEARCH_REINDEX_MODE=alias
SEARCH_REINDEX_BATCH_SIZE=500
SEARCH_REINDEX_MAX_BATCH_SIZE=5000
SEARCH_REINDEX_LOCK_TTL=3h
SEARCH_REINDEX_JOB_TIMEOUT=2h
SEARCH_REINDEX_COLLECTION_PREFIX=products
SEARCH_REINDEX_OLD_COLLECTION_RETENTION=72h
TYPESENSE_REINDEX_IMPORT_TIMEOUT_MS=5000
```

### Environment variable explanation

| Variable | Required? | Purpose | Example | Security Notes |
|---|---:|---|---|---|
| `SEARCH_HTTP_ADDR` | Optional | Search API bind address | `:8085` | Use `127.0.0.1:8085` if only local access needed |
| `SEARCH_HTTP_READ_TIMEOUT` | Optional | Max request read time | `5s` | Keep finite to avoid slow client abuse |
| `SEARCH_HTTP_WRITE_TIMEOUT` | Optional | Max response write time | `10s` | Keep finite |
| `SEARCH_HTTP_IDLE_TIMEOUT` | Optional | Keep-alive idle timeout | `60s` | Useful for connection cleanup |
| `SEARCH_SHUTDOWN_TIMEOUT` | Optional | Graceful shutdown timeout | `10s` | Avoid too low in production |
| `TYPESENSE_URL` | Optional | Full Typesense URL override | `http://localhost:8108` | Use HTTPS in production if available |
| `TYPESENSE_HOST` | Required if URL empty | Typesense host only | `localhost` | Do not include `http://` here |
| `TYPESENSE_PORT` | Required if URL empty | Typesense API port | `8108` | Must be 1 to 65535 |
| `TYPESENSE_PROTOCOL` | Required if URL empty | `http` or `https` | `http` | Prefer `https` outside local |
| `TYPESENSE_API_KEY` | Yes | Auth key for Typesense API | `dev-typesense-key` | Secret. Never commit real value |
| `TYPESENSE_PRODUCTS_COLLECTION` | Optional | Product collection or alias | `products` | Keep stable for API behavior |
| `TYPESENSE_POPULAR_QUERIES_COLLECTION` | Optional | Popular query collection | `popular_queries` | No secret |
| `TYPESENSE_TIMEOUT_MS` | Optional | Typesense request timeout | `300` | Too low causes flaky search |
| `PRODUCT_SERVICE_URL` | Yes for real search | Product Service base URL | `http://localhost:8082` | Internal URL only |
| `PRODUCT_SERVICE_BATCH_GET_PATH` | Optional | Product hydration path | `/internal/v1/products:batchGet` | Internal endpoint |
| `PRODUCT_SERVICE_SEARCH_EXPORT_PATH` | Optional | Reindex export path | `/internal/v1/products/search-export` | Internal endpoint |
| `PRODUCT_SERVICE_TIMEOUT_MS` | Optional | Product hydration timeout | `300` | Too low causes missing results |
| `PRODUCT_SERVICE_SEARCH_EXPORT_TIMEOUT_MS` | Optional | Product export page timeout | `2000` | Reindex may need more time |
| `SEARCH_DEFAULT_PAGE_SIZE` | Optional | Default search page size | `20` | No secret |
| `SEARCH_MAX_PAGE_SIZE` | Optional | Max page size allowed | `100` | Keep bounded to avoid heavy queries |
| `SEARCH_AUTOCOMPLETE_DEFAULT_LIMIT` | Optional | Default autocomplete limit | `8` | No secret |
| `SEARCH_AUTOCOMPLETE_MAX_LIMIT` | Optional | Max autocomplete limit | `10` | Code enforces max 10 |
| `SEARCH_AUTOCOMPLETE_PREFIX_CACHE_TTL` | Optional | Prefix cache TTL | `1m` | No secret |
| `SEARCH_AUTOCOMPLETE_EMPTY_CACHE_TTL` | Optional | Empty query cache TTL | `5m` | No secret |
| `SEARCH_AUTOCOMPLETE_TYPESENSE_TIMEOUT_MS` | Optional | Autocomplete Typesense timeout | `150` | Too low causes fallback/no suggestions |
| `SEARCH_AUTOCOMPLETE_CACHE_TIMEOUT_MS` | Optional | Redis cache timeout | `50` | Keep low for fast autocomplete |
| `SEARCH_ZERO_RESULT_TRACKING_ENABLED` | Optional | Send zero-result analytics | `true` | Disable locally if Session Service unavailable |
| `SEARCH_ZERO_RESULT_DEDUPE_TTL` | Optional | Dedupe window | `30m` | No secret |
| `SEARCH_ZERO_RESULT_SEND_TIMEOUT_MS` | Optional | Session event send timeout | `150` | No secret |
| `SEARCH_ZERO_RESULT_QUEUE_SIZE` | Optional | In-memory event queue size | `1024` | Too high uses more memory |
| `SEARCH_ZERO_RESULT_WORKERS` | Optional | Background event workers | `2` | Tune carefully |
| `SESSION_SERVICE_URL` | Required if zero tracking enabled | Session Service base URL | `http://localhost:8086` | Internal URL only |
| `SESSION_SERVICE_INGEST_PATH` | Required if zero tracking enabled | Session event ingest path | `/api/v1/sessions/events` | Internal endpoint |
| `SESSION_SERVICE_TIMEOUT_MS` | Optional | Session call timeout | `150` | No secret |
| `SEARCH_ADMIN_TIMEOUT_MS` | Optional | Admin mutation timeout | `500` | No secret |
| `SEARCH_ADMIN_AUTH_ENABLED` | Optional | Header-based admin auth on/off | `true` | Set `false` only for local testing |
| `SEARCH_ADMIN_MUTATION_RATE_LIMIT` | Optional | Admin writes per window | `30` | Protects mutation endpoints |
| `SEARCH_ADMIN_MUTATION_RATE_WINDOW` | Optional | Rate limit window | `1m` | No secret |
| `SEARCH_INDEXER_ENABLED` | Optional | RabbitMQ product consumer enabled | `true` | Set `false` for API-only local boot |
| `PRODUCT_INDEXER_MESSAGE_TIMEOUT` | Optional | Per event processing timeout | `30s` | No secret |
| `PRODUCT_INDEXER_PROCESSED_EVENT_TTL` | Optional | Event dedupe TTL | `720h` | No secret |
| `QUEUE_PROVIDER` | Required if indexer enabled | Queue backend | `rabbitmq` | Only `rabbitmq` supported |
| `RABBITMQ_URL` | Required if indexer enabled | AMQP connection URL | `amqp://user:pass@localhost:5672/ecommerce` | Secret because password is inside URL |
| `PRODUCT_EVENTS_EXCHANGE` | Optional | Product event exchange | `product.events` | No secret |
| `PRODUCT_INDEXER_QUEUE` | Optional | Consumer queue name | `search-service.product-indexer` | No secret |
| `PRODUCT_INDEXER_DLX` | Optional | Dead-letter exchange | `product.events.dlx` | No secret |
| `PRODUCT_INDEXER_DLQ` | Optional | Dead-letter queue | `search-service.product-indexer.dlq` | No secret |
| `PRODUCT_INDEXER_CONSUMER_TAG` | Optional | RabbitMQ consumer tag | `search-service-product-indexer` | No secret |
| `PRODUCT_INDEXER_ROUTING_KEYS` | Optional | Product event routing keys | `product.published,...` | No secret |
| `PRODUCT_INDEXER_PREFETCH` | Optional | Unacked message count | `20` | Too high can overload service |
| `PRODUCT_INDEXER_MAX_RETRIES` | Optional | Retry attempts before DLQ | `5` | No secret |
| `PRODUCT_INDEXER_RECONNECT_DELAY` | Optional | RabbitMQ reconnect delay | `5s` | No secret |
| `PRODUCT_INDEXER_RETRY_BASE_DELAY` | Optional | Retry base delay | `1s` | No secret |
| `PRODUCT_INDEXER_RETRY_MAX_DELAY` | Optional | Retry max delay | `5m` | No secret |
| `SEARCH_REDIS_ADDR` | Yes for full service | Redis host and port | `localhost:6379` | Internal address |
| `REDIS_ADDR` | Optional fallback | Fallback Redis address | `localhost:6379` | Prefer `SEARCH_REDIS_ADDR` |
| `SEARCH_REDIS_PASSWORD` | Required if Redis auth enabled | Redis password | `dev_redis_password` | Secret |
| `REDIS_PASSWORD` | Optional fallback | Fallback Redis password | `dev_redis_password` | Secret |
| `SEARCH_REDIS_DB` | Optional | Redis logical DB | `0` | No secret |
| `SEARCH_REDIS_DIAL_TIMEOUT` | Optional | Redis connect timeout | `2s` | No secret |
| `SEARCH_REDIS_READ_TIMEOUT` | Optional | Redis read timeout | `2s` | No secret |
| `SEARCH_REDIS_WRITE_TIMEOUT` | Optional | Redis write timeout | `2s` | No secret |
| `SEARCH_REINDEX_MODE` | Optional | `alias` or `in_place` reindex mode | `alias` | `alias` is safer |
| `SEARCH_REINDEX_BATCH_SIZE` | Optional | Product export batch size | `500` | Too high can stress Product Service |
| `SEARCH_REINDEX_MAX_BATCH_SIZE` | Optional | Max allowed batch size | `5000` | Keep bounded |
| `SEARCH_REINDEX_LOCK_TTL` | Optional | Redis lock TTL | `3h` | Prevents duplicate reindex |
| `SEARCH_REINDEX_JOB_TIMEOUT` | Optional | Full reindex timeout | `2h` | Increase for large catalog |
| `SEARCH_REINDEX_COLLECTION_PREFIX` | Optional | New collection prefix | `products` | No secret |
| `SEARCH_REINDEX_OLD_COLLECTION_RETENTION` | Optional | Keep old collections after alias swap | `72h` | Helps rollback |
| `TYPESENSE_REINDEX_IMPORT_TIMEOUT_MS` | Optional | Typesense import timeout | `5000` | No secret |

### Docker Compose-only variables

These are used by `infra/compose/docker-compose.local.yml`:

| Variable | Purpose | Example |
|---|---|---|
| `TYPESENSE_PORT` | Host port for Typesense | `8108` |
| `TYPESENSE_API_KEY` | Typesense API key passed to container | `dev-typesense-key` |
| `REDIS_PORT` | Host port for Redis | `6379` |
| `REDIS_PASSWORD` | Redis password passed to container | `dev_redis_password` |
| `RABBITMQ_PORT` | Host AMQP port | `5672` |
| `RABBITMQ_MANAGEMENT_PORT` | RabbitMQ web UI port | `15672` |
| `RABBITMQ_DEFAULT_USER` | RabbitMQ local user | `ecommerce` |
| `RABBITMQ_DEFAULT_PASS` | RabbitMQ local password | `ecommerce_password` |
| `RABBITMQ_DEFAULT_VHOST` | RabbitMQ vhost | `ecommerce` |

### Common `.env` mistakes

| Mistake | Result | Fix |
|---|---|---|
| `.env` banaya but source nahi kiya | Go service env vars nahi dekhega | `set -a; . ./.env; set +a` use karo |
| `TYPESENSE_HOST=http://localhost` | Validation fails | `TYPESENSE_HOST=localhost`, protocol alag rakho |
| `TYPESENSE_API_KEY` empty | Service exits on startup | Key set karo |
| Compose service names host se use kiye | Host Go process `typesense` resolve nahi karega | Host run me `localhost` use karo |
| Real secrets `.env.local.example` me daal diye | Secret leak risk | Real secrets only `.env.local`, K8s Secret, secret manager me rakho |

## 5. External Services Analysis

### 5.1 Typesense

| Item | Detail |
|---|---|
| What | Search engine |
| Why | Product search, facets, sorting, typo tolerance, synonyms |
| Mandatory | Yes |
| Health check | `curl http://localhost:8108/health` |
| Credentials | `TYPESENSE_API_KEY` |
| Common issue | Wrong API key, collection creation fails |

The server creates or ensures the required collections on startup. Isliye first boot ke time Typesense running hona chahiye.

### 5.2 Redis

| Item | Detail |
|---|---|
| What | In-memory cache/store |
| Why | Autocomplete cache, event dedupe, zero-result dedupe, reindex lock |
| Mandatory | Yes for full local setup |
| Health check | `redis-cli -a dev_redis_password ping` |
| Credentials | `SEARCH_REDIS_PASSWORD` |
| Common issue | Password mismatch or Redis not running |

### 5.3 RabbitMQ

| Item | Detail |
|---|---|
| What | Message broker |
| Why | Product events se search index update hota hai |
| Mandatory | Default yes, unless `SEARCH_INDEXER_ENABLED=false` |
| Health check | `curl -u ecommerce:ecommerce_password http://localhost:15672/api/overview` |
| Credentials | `RABBITMQ_URL` or Compose user/password vars |
| Common issue | Vhost missing, wrong password, port `5672` busy |

Docker run:

```bash
docker run --name ecommerce-rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=ecommerce \
  -e RABBITMQ_DEFAULT_PASS=ecommerce_password \
  -e RABBITMQ_DEFAULT_VHOST=ecommerce \
  -v rabbitmq_data:/var/lib/rabbitmq \
  rabbitmq:3.13-management-alpine
```

Verify management UI:

```text
http://localhost:15672
```

Login:

```text
username: ecommerce
password: ecommerce_password
```

### 5.4 Product Service

| Item | Detail |
|---|---|
| What | Product catalog source of truth |
| Why | Search result hydration and full reindex export |
| Mandatory | Yes for real search and reindex |
| Default URL | `http://localhost:8082` |
| Paths | `/internal/v1/products:batchGet`, `/internal/v1/products/search-export` |
| Common issue | Search returns IDs from Typesense but hydration fails |

Task 1 ka important architecture rule: Product Service owns product data. Search Service only searchable indexed copy rakhta hai.

### 5.5 Session Service

| Item | Detail |
|---|---|
| What | User/session analytics service |
| Why | Zero-result search event ingest |
| Mandatory | Optional, but enabled by default |
| Default URL | `http://localhost:8086` |
| Disable locally | `SEARCH_ZERO_RESULT_TRACKING_ENABLED=false` |

### 5.6 Docker

| Item | Detail |
|---|---|
| What | Container runtime |
| Why | Typesense, Redis, RabbitMQ locally run karne ke liye |
| Mandatory | Not mandatory, but recommended |
| Check | `docker version` |
| Common issue | Docker daemon not running |

### 5.7 Kubernetes

Kubernetes optional hai. Repo me these manifests exist:

| Path | Purpose |
|---|---|
| `infra/k8s/data/typesense` | Typesense namespace, config, secret example, services, StatefulSet, PDB |
| `infra/k8s/core/search-service/configmap.yaml` | Search Service non-secret config |
| `infra/k8s/core/search-service/secret.example.yaml` | Search Service secret keys example |
| `infra/k8s/jobs/search-reindex` | Reindex job config, secret example, job manifest |

Note: Search Service deployment/service manifests are not present in `infra/k8s/core/search-service` right now. ConfigMap and Secret examples exist, but a Deployment/Service manifest or Helm chart still needs to be added for full K8s deploy.

## 6. Ports and Networking

| Service | Port | Purpose |
|---|---:|---|
| Search Service HTTP API | `8085` | Main Search API and health endpoint |
| Typesense API | `8108` | Search engine API |
| Typesense peering | `8107` | Typesense cluster peering in K8s |
| Redis | `6379` | Cache, locks, dedupe |
| RabbitMQ AMQP | `5672` | Message consumer connection |
| RabbitMQ Management UI | `15672` | Browser/admin API |
| Product Service | `8082` | Product hydration and export |
| Session Service | `8086` | Zero-result analytics ingest |

### Host vs Docker network names

If Search Service runs on your host machine with `go run`, use:

```env
TYPESENSE_HOST=localhost
SEARCH_REDIS_ADDR=localhost:6379
RABBITMQ_URL=amqp://ecommerce:ecommerce_password@localhost:5672/ecommerce
```

If Search Service runs inside the same Docker Compose network, use:

```env
TYPESENSE_HOST=typesense
SEARCH_REDIS_ADDR=redis:6379
RABBITMQ_URL=amqp://ecommerce:ecommerce_password@rabbitmq:5672/ecommerce
```

### Port conflict fixes

| Conflict | Fix |
|---|---|
| `8085 already in use` | Change `SEARCH_HTTP_ADDR=:8087` or stop old process |
| `8108 already in use` | Change Compose `TYPESENSE_PORT=8118` and update service env |
| `6379 already in use` | Stop local Redis or change Compose `REDIS_PORT` |
| `5672 already in use` | Stop other RabbitMQ or change `RABBITMQ_PORT` |

Check port usage:

```bash
ss -ltnp | grep ':8085'
```

or:

```bash
lsof -i :8085
```

### Firewall and Docker network issues

- Localhost ports work only if Compose exposes them.
- Docker service name like `typesense` works inside Docker network, not from host shell.
- Corporate firewall/VPN can block downloads and external package access.
- On Windows, WSL2 and Docker Desktop networking can differ. Prefer running commands inside the same WSL distro.

## 7. Docker and DevOps Setup

### Existing Docker Compose

Project local infra file:

```text
infra/compose/docker-compose.local.yml
```

Services included:

| Service | Image | Volume |
|---|---|---|
| Typesense | `typesense/typesense:27.1` | `typesense_data:/data` |
| Redis | `redis:7.4-alpine` | `redis_data:/data` |
| RabbitMQ | `rabbitmq:3.13-management-alpine` | `rabbitmq_data:/var/lib/rabbitmq` |

### Start local infra

```bash
cp infra/compose/.env.local.example infra/compose/.env.local
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d
```

### Stop local infra

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml down
```

### View logs

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml logs -f typesense
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml logs -f redis
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml logs -f rabbitmq
```

### Check containers

```bash
docker ps
```

### Volumes

Volumes persist data between restarts:

| Volume | Stores |
|---|---|
| `typesense_data` | Search collections and documents |
| `redis_data` | Redis append-only data |
| `rabbitmq_data` | RabbitMQ queues, exchanges, vhost metadata |

Warning: `docker compose down -v` deletes volumes and local data.

### Dockerfile status

Repo documentation has a generic Go Dockerfile example, but there is no service-specific Dockerfile under `backend/services/search-service` right now. For production image build, add a Dockerfile or central build pipeline for `cmd/server`, `cmd/reindex`, and `cmd/reindex-alias`.

### Kubernetes setup

Typesense:

```bash
kubectl apply -f infra/k8s/data/typesense/namespace.yaml
kubectl apply -f infra/k8s/data/typesense/configmap.yaml
kubectl apply -f infra/k8s/data/typesense/secret.example.yaml
kubectl apply -f infra/k8s/data/typesense/service-headless.yaml
kubectl apply -f infra/k8s/data/typesense/service.yaml
kubectl apply -f infra/k8s/data/typesense/statefulset.yaml
kubectl apply -f infra/k8s/data/typesense/pdb.yaml
```

Before real deploy, replace `secret.example.yaml` values with real secrets from a secret manager.

Search Service config:

```bash
kubectl apply -f infra/k8s/core/search-service/namespace.yaml
kubectl apply -f infra/k8s/core/search-service/configmap.yaml
kubectl apply -f infra/k8s/core/search-service/secret.example.yaml
```

Reindex job:

```bash
kubectl apply -f infra/k8s/jobs/namespace.yaml
kubectl apply -f infra/k8s/jobs/search-reindex/configmap.yaml
kubectl apply -f infra/k8s/jobs/search-reindex/job.yaml
```

## 8. Project Run Instructions

### Step 1: Clone repository

```bash
git clone <repository-url>
cd Ecommerce
```

### Step 2: Install required runtimes

Verify:

```bash
go version
docker version
docker compose version
curl --version
```

Recommended:

- Go compatible with `go 1.26.3`.
- Docker Desktop or Docker Engine.
- `curl` for health checks.
- Optional: `redis-cli` for Redis debugging.

### Step 3: Start infra services

```bash
cp infra/compose/.env.local.example infra/compose/.env.local
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d
```

Verify:

```bash
curl http://localhost:8108/health
docker exec -it ecommerce-redis redis-cli -a dev_redis_password ping
curl -u ecommerce:ecommerce_password http://localhost:15672/api/overview
```

### Step 4: Create local service `.env`

```bash
cd backend/services/search-service
```

Create `backend/services/search-service/.env` using the example in Section 4.

For quick standalone boot without RabbitMQ consumer, you can set:

```env
SEARCH_INDEXER_ENABLED=false
```

But for full service behavior, keep it `true` and make sure RabbitMQ is running.

### Step 5: Install Go dependencies

```bash
go mod download
go mod tidy
```

### Step 6: Run tests

```bash
go test ./...
```

Integration tests that require live Typesense may be skipped unless `TYPESENSE_INTEGRATION=1` is set.

### Step 7: Load env and start backend service

```bash
set -a
. ./.env
set +a
go run ./cmd/server
```

Expected log includes:

```text
search.http.started
```

### Step 8: Verify Search Service health

```bash
curl http://localhost:8085/healthz
```

Expected:

```json
{"status":"ok"}
```

### Step 9: Verify schema endpoint

```bash
curl http://localhost:8085/internal/v1/search/schema/products
```

This should return product schema, searchable fields, facets, sort policy, and synonym model.

### Step 10: Verify Typesense collections

```bash
curl -H "X-TYPESENSE-API-KEY: dev-typesense-key" http://localhost:8108/collections/products
curl -H "X-TYPESENSE-API-KEY: dev-typesense-key" http://localhost:8108/collections/popular_queries
```

### Step 11: Run a sample search

```bash
curl "http://localhost:8085/api/v1/search?q=shoes&page=1&page_size=20"
```

Note: Real product results need indexed product documents and Product Service hydration. Empty results are okay on a fresh setup.

### Step 12: Run autocomplete

```bash
curl "http://localhost:8085/api/v1/search/autocomplete?q=sho&limit=5"
```

### Step 13: Admin synonym example

Admin auth is header-based. With `SEARCH_ADMIN_AUTH_ENABLED=true`, pass admin headers:

```bash
curl -X POST http://localhost:8085/api/v1/admin/search/synonyms \
  -H "Content-Type: application/json" \
  -H "X-User-ID: local-admin" \
  -H "X-Roles: superadmin" \
  -d '{"root":"mobile","synonyms":["phone","smartphone","cellphone"]}'
```

For local-only testing, you can set:

```env
SEARCH_ADMIN_AUTH_ENABLED=false
```

Do not use this in production.

### Step 14: Reindex

Dry run first:

```bash
go run ./cmd/reindex --dry-run --reason=local-check --actor-id=local-dev
```

Real alias reindex:

```bash
go run ./cmd/reindex --mode=alias --batch-size=500 --reason=local-seed --actor-id=local-dev
```

Rollback/move alias manually:

```bash
go run ./cmd/reindex-alias --alias=products --target=products_20260605_120000 --reason=rollback --actor-id=local-dev
```

## 9. Common Errors and Fixes

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `TYPESENSE_API_KEY cannot be empty` | Env missing | Set `TYPESENSE_API_KEY` and reload `.env` | Keep `.env.example` updated |
| `TYPESENSE_HOST must not include a protocol` | Host has `http://` | Use `TYPESENSE_HOST=localhost` and `TYPESENSE_PROTOCOL=http` | Use `TYPESENSE_URL` if you want full URL |
| `collection_ensure_failed` | Typesense down or API key wrong | Check `curl /health`, key, port | Start Typesense before service |
| `address already in use :8085` | Another process uses port | Change `SEARCH_HTTP_ADDR` or stop process | Standardize local ports |
| `connect rabbitmq: connection refused` | RabbitMQ not running | Start RabbitMQ or set `SEARCH_INDEXER_ENABLED=false` | Use Compose health checks |
| RabbitMQ `ACCESS_REFUSED` | Wrong user/pass/vhost | Check `RABBITMQ_URL` and Compose env | Keep URL aligned with vhost |
| Redis `NOAUTH` or auth error | Password missing/wrong | Set `SEARCH_REDIS_PASSWORD=dev_redis_password` | Use same value as Compose `REDIS_PASSWORD` |
| Redis `connection refused` | Redis not running or wrong host | Start Redis and use `localhost:6379` from host | Avoid Docker service names in host Go run |
| Search returns empty products | No documents indexed | Publish product events or run reindex | Seed local Typesense |
| Product hydration fails | Product Service unavailable | Start Product Service or check `PRODUCT_SERVICE_URL` | Verify `/internal/v1/products:batchGet` |
| Reindex fails at export | Product export endpoint unavailable | Check `PRODUCT_SERVICE_SEARCH_EXPORT_PATH` | Dry-run reindex first |
| Zero-result analytics errors | Session Service unavailable | Start Session Service or set `SEARCH_ZERO_RESULT_TRACKING_ENABLED=false` | Disable optional analytics locally |
| Admin endpoint returns `401` | Missing user header | Add `X-User-ID` | Use test admin headers locally |
| Admin endpoint returns `403` | Role not allowed | Use `X-Roles: superadmin` or `catalog_admin` | Keep role mapping documented |
| `go mod download` failed | Network/proxy problem | Check internet, set `GOPROXY` | Avoid editing `go.sum` manually |
| `go test` integration skipped | Live Typesense test not enabled | Set `TYPESENSE_INTEGRATION=1` if needed | Keep unit and integration test modes separate |
| Docker daemon not running | Docker Desktop/Engine stopped | Start Docker | Check `docker version` before setup |
| Compose says env var required | `.env.local` missing | Copy `.env.local.example` | Keep Compose commands using `--env-file` |
| Migration failed | No SQL migration exists for this service | Use collection ensure and reindex flow | Document that Search Service has Typesense schema, not SQL migrations |

## 10. Security and Configuration Audit

| Finding | Risk | Suggested Fix |
|---|---|---|
| Dev secrets exist in examples | Safe for examples, unsafe if reused in prod | Rotate and store real secrets in secret manager |
| `RABBITMQ_URL` contains username/password | Secret leak risk | Put in K8s Secret or vault, never commit real value |
| `TYPESENSE_API_KEY` required | Empty key stops service, leaked key exposes index | Use strong per-env key |
| `SEARCH_REDIS_PASSWORD` optional in code | Redis could run without auth | Require auth in shared/dev/prod environments |
| Admin auth trusts headers | Direct exposure could allow spoofing | Expose admin endpoints only behind API Gateway/auth middleware |
| `SEARCH_ADMIN_AUTH_ENABLED=false` possible | Useful locally but unsafe in prod | Block false value in production config validation or deployment policy |
| Local HTTP uses no TLS | Fine for local, not for public traffic | Use TLS at ingress/API Gateway |
| Redis client uses plain TCP | Managed Redis with TLS may need code/config support | Add TLS support before using TLS-only Redis |
| `/healthz` is shallow | It only confirms HTTP process, not dependencies | Add readiness checks for Typesense, Redis, RabbitMQ |
| No Search Service Dockerfile present | Build/deploy gap | Add service Dockerfile or central image build config |
| No Search Service K8s Deployment present | K8s runtime gap | Add Deployment, Service, HPA, probes, resources |
| Reindex can stress Product Service | Large batch and timeout can overload dependency | Keep batch size bounded and run off-peak |
| Old Typesense collections retained | Storage can grow | Cleanup old collections after retention window |

## 11. Best Practices

- Never commit `.env`, `.env.local`, real K8s Secret, or production credentials.
- Keep `.env.example` updated whenever config changes.
- Use Docker Compose for local Typesense, Redis, and RabbitMQ.
- Use strong unique `TYPESENSE_API_KEY`, Redis password, and RabbitMQ password per environment.
- Keep `SEARCH_REINDEX_MODE=alias` for safer reindex and rollback.
- Run `go test ./...` before pushing Search Service changes.
- Run `go mod tidy` only when dependency/import changes are intentional.
- Keep Product Service as source of truth. Search index can be stale.
- Use Redis locks for reindex, and do not manually run multiple reindex jobs at once.
- Monitor RabbitMQ DLQ so failed product events are not ignored.
- Add alerts for Typesense health, RabbitMQ queue lag, Redis latency, search latency, and zero-result rate.
- Use separate dev/staging/prod config values.
- Keep admin endpoints behind authenticated gateway.
- Use Docker volumes for local persistence, but clean them intentionally when testing fresh setup.
- Document every new external service and env var in this dependency file or a central runbook.

## 12. Final Checklist

- [ ] Go compatible with `go 1.26.3` installed
- [ ] Docker and Docker Compose installed
- [ ] Repository cloned
- [ ] Search Service folder opened: `backend/services/search-service`
- [ ] Go dependencies downloaded with `go mod download`
- [ ] Tests run with `go test ./...`
- [ ] `infra/compose/.env.local` created from example
- [ ] Typesense container running on `8108`
- [ ] Redis container running on `6379`
- [ ] RabbitMQ container running on `5672` and UI on `15672`
- [ ] Local service `.env` created
- [ ] `TYPESENSE_API_KEY` set
- [ ] `SEARCH_REDIS_PASSWORD` matches Redis container password
- [ ] `RABBITMQ_URL` matches RabbitMQ user/password/vhost
- [ ] Product Service URL configured
- [ ] Session Service URL configured or zero-result tracking disabled locally
- [ ] Search Service started with `go run ./cmd/server`
- [ ] `/healthz` returns `{"status":"ok"}`
- [ ] Typesense `products` collection exists
- [ ] Typesense `popular_queries` collection exists
- [ ] Search endpoint tested
- [ ] Autocomplete endpoint tested
- [ ] Admin headers understood for synonyms/reindex endpoints
- [ ] Reindex dry run tested
- [ ] Logs checked for Redis/RabbitMQ/Typesense errors
- [ ] Common errors reviewed
- [ ] Secrets are not committed

## Quick Beginner Path

If you are setting this up first time, follow this minimum flow:

```bash
git clone <repository-url>
cd Ecommerce
cp infra/compose/.env.local.example infra/compose/.env.local
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d
cd backend/services/search-service
go mod download
```

Create `backend/services/search-service/.env` from Section 4, then:

```bash
set -a
. ./.env
set +a
go run ./cmd/server
```

Verify:

```bash
curl http://localhost:8085/healthz
curl http://localhost:8085/internal/v1/search/schema/products
```
