# Project Dependency & Setup Guide

## 1. Project Overview

| Variable | Value |
|---|---|
| `SERVICE_NAME` | Prompt variable from the assignment |
| `TASK_FILE_NAME` | Prompt variable from the assignment |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `{TASK_FILE_NAME without .md}_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

Current task ka focus Autocomplete API hai. Simple Hinglish me: buyer header search box me text type karta hai, backend `GET /api/v1/search/autocomplete` se fast suggestions return karta hai. Suggestions do jagah se aati hain:

- Typesense `popular_queries` collection se popular search terms.
- Typesense `products` collection se product title aur brand prefix matches.

Important boundaries:

- Original `TASK_FILE_NAME` modify nahi kiya gaya.
- Business logic rewrite nahi kiya gaya.
- Common Go, Docker, Typesense, Redis, RabbitMQ, and full `.env` setup repeat nahi kiya gaya.
- Previous dependency guides ko reference kiya gaya hai, and sirf autocomplete-specific setup detail yahan add ki gayi hai.

Current implemented runtime flow:

```text
Client / Gateway
  -> GET /api/v1/search/autocomplete?q=sho&limit=8
  -> backend/services/search-service HTTP route
  -> Redis cache GET autocomplete:v1:sho:8
  -> Typesense popular_queries search
  -> Typesense products prefix search when q length >= 2
  -> merge, dedupe, rank
  -> Redis cache SETEX
  -> JSON response: {"suggestions":[...]}
```

Note: `api/master-api.json` defines an internal `SearchService.Autocomplete` contract, but this repo snapshot does not have a `proto/` directory. The implemented runnable path is the HTTP route in `backend/services/search-service/internal/transport/http/handler.go`.

## 2. Tech Stack

### Reused tech from previous dependency files

| Technology | Required for this task? | Reuse reference |
|---|---:|---|
| Go | Yes | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `2. Go Dependency System` |
| Go modules | Yes | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `2. Go Dependency System` |
| Standard Go `net/http` | Yes | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `1. Project Tech Stack Analysis` |
| Typesense | Yes | `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`, Section `4.1 Typesense` |
| Redis | Yes for cache, optional failure path | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `3.2 Redis` |
| Docker Compose | Recommended for local infra | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `7. Docker and DevOps Setup` |
| RabbitMQ | Not needed by autocomplete read path, but needed if indexer is enabled | `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`, Section `6. Redis / Queue / External Services` |

### Task-specific tech details

| Component | Required? | Beginner-friendly explanation |
|---|---:|---|
| Autocomplete HTTP endpoint | Yes | Public API hai jo typed prefix ke badle suggestions return karta hai. |
| Typesense `popular_queries` collection | Yes | Popular search terms store karta hai. Empty query ya short query me ye useful suggestions deta hai. |
| Typesense product prefix search | Yes when `q` has at least 2 characters | Product title aur brand se suggestions nikalta hai, jaise `sho` -> `shoes`. |
| Redis autocomplete cache | Yes for performance | Repeated keystrokes fast banata hai. Redis down ho to endpoint Typesense se response de sakta hai. |
| Custom Redis client | Yes | Repo me `go-redis` package nahi hai; custom RESP client use ho raha hai. |
| gRPC contract in API catalog | Contract-level | `api/master-api.json` me method defined hai, but generated proto implementation current repo me present nahi hai. |

## 3. Required Software

No brand-new software install is introduced by `TASK_FILE_NAME`.

Follow already documented base setup first:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`2. Go Dependency System`
`3. Database and Storage Analysis`
`4. Environment Variables`
`7. Docker and DevOps Setup`
`8. Project Run Instructions`
```

For autocomplete testing, confirm these existing services are available:

| Software / Service | Required? | Why needed | Quick check |
|---|---:|---|---|
| Go | Yes | Run and test backend code | `go version` |
| Typesense | Yes | Reads `products` and `popular_queries` | `curl http://localhost:8108/health` |
| Redis | Recommended and configured | Cache keys like `autocomplete:v1:sho:8` | `docker exec ecommerce-redis redis-cli -a dev_redis_password ping` |
| Docker Compose | Recommended | Starts local Typesense and Redis | `docker compose version` |
| RabbitMQ | Only if `SEARCH_INDEXER_ENABLED=true` | Keeps `products` index fresh through events | `docker exec ecommerce-rabbitmq rabbitmq-diagnostics -q ping` |
| Product Service | Not directly used by autocomplete | Needed by normal search and reindex, not suggestions endpoint | See Task 4 docs |

## 4. Dependency Management

This is a Go project. Dependency system, `go.mod`, `go.sum`, `go mod download`, `go mod tidy`, and common module errors are already explained in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`2. Go Dependency System`
```

Task-specific dependency audit:

| File / Dependency | Current finding | New install needed? |
|---|---|---:|
| `backend/services/search-service/go.mod` | Already has `github.com/typesense/typesense-go/v2 v2.0.0` | No |
| `github.com/rabbitmq/amqp091-go` | Already exists for product indexer, not autocomplete itself | No |
| Redis client | Implemented locally in `internal/repository/redis_client.go` | No |
| `go-redis` | Not used | No |
| Node/Python packages | Not used for this task | No |

Task-specific test commands:

```bash
cd backend/services/search-service
go test ./internal/domain ./internal/usecase ./internal/transport/http
```

Repository package tests can also be run; live Typesense integration tests are skipped unless `TYPESENSE_INTEGRATION=1` is set.

## 5. Database Setup

No SQL database and no SQL migration is introduced by `TASK_FILE_NAME`.

| Storage / DB | Used? | Purpose | Setup action |
|---|---:|---|---|
| Typesense `popular_queries` | Yes | Popular/autocomplete terms | Existing Typesense runtime, task-specific collection must exist |
| Typesense `products` | Yes | Product title and brand prefix candidates | Reuse Task 2 setup and Task 3/Task 8 indexing flow |
| Redis | Yes | Autocomplete cache | Reuse Task 1 setup |
| MySQL/PostgreSQL/MongoDB | No | Not used by this task | No setup |

### Typesense setup reuse

Do not repeat installation steps. Typesense Docker, health checks, credentials, and port setup are already explained in:

```md
`TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`

Section:
`4.1 Typesense`
```

### Task-specific `popular_queries` collection

Current code ensures this collection on server startup through `schema.MustPopularQueriesCollectionSchema()`.

| Field | Type | Why it exists |
|---|---|---|
| `id` | string | Unique document id |
| `query` | string | Display text returned as suggestion |
| `normalized_query` | string | Lowercase/collapsed form for validation and dedupe |
| `score` | int32, sort | Main ranking signal |
| `count` | int32, sort | Search count/popularity count |
| `last_seen_at` | int64, sort | Freshness signal |
| `is_active` | bool, facet | Hide spam/blocked terms when false |
| `locale` | string, optional facet | Future locale-specific suggestions |

Beginner note: Typesense ek search engine hai, normal SQL table nahi. Collection create hone ke baad documents manually seed, batch job se seed, ya future analytics ingestion se populate karne padenge. Fresh local setup me empty suggestions normal ho sakte hain.

Optional local seed for quick manual testing after the backend has started once and ensured the collection:

```bash
curl -X POST "http://localhost:8108/collections/popular_queries/documents" \
  -H "X-TYPESENSE-API-KEY: dev-typesense-key" \
  -H "Content-Type: application/json" \
  -d '{"id":"query_shoes","query":"shoes","normalized_query":"shoes","score":100,"count":50,"last_seen_at":1710000000,"is_active":true}'
```

### Product prefix requirement

Autocomplete product suggestions need product documents in `products`. Product document indexing is not new here:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Sections:
`5. Database Setup`
`6. Redis / Queue / External Services`
```

If `products` is empty, autocomplete can still return popular query suggestions.

## 6. Redis / Queue / External Services

### Redis autocomplete cache

Redis setup itself is already covered in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`3.2 Redis`
`5.2 Redis`
```

Task-specific cache behavior:

| Item | Value |
|---|---|
| Cache key pattern | `autocomplete:v1:{normalized_query}:{limit}` |
| Empty query example | `autocomplete:v1::8` |
| Prefix query example | `autocomplete:v1:nike shoes:8` |
| Non-empty prefix TTL | `SEARCH_AUTOCOMPLETE_PREFIX_CACHE_TTL`, default `1m` |
| Empty query TTL | `SEARCH_AUTOCOMPLETE_EMPTY_CACHE_TTL`, default `5m` |
| Cache timeout | `SEARCH_AUTOCOMPLETE_CACHE_TIMEOUT_MS`, default `50` |

Important behavior:

- Redis cache get/set errors are logged but autocomplete can still return suggestions from Typesense.
- If `SEARCH_INDEXER_ENABLED=true`, startup also needs Redis for product event idempotency.
- Cache value is JSON array of strings, not product objects.

### Typesense as required search backend

Autocomplete cannot work correctly without Typesense because both `popular_queries` and `products` are queried there. If Typesense is down and Redis cache misses, endpoint returns `503 SEARCH_BACKEND_UNAVAILABLE`.

### RabbitMQ / Kafka / queues

`TASK_FILE_NAME` does not introduce a new queue. RabbitMQ is only relevant when the product indexer is enabled, because indexed product documents improve product prefix suggestions.

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Section:
`6. Redis / Queue / External Services`
```

Kafka is not supported by current code for this service. Keep:

```env
QUEUE_PROVIDER=rabbitmq
```

## 7. Environment Variables

Full `.env` setup is already documented in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Environment Variables`
```

Do not copy the full `.env` again. For `TASK_FILE_NAME`, verify this task-specific subset in:

```text
backend/services/search-service/.env
```

```env
# Autocomplete-specific Typesense collection
TYPESENSE_POPULAR_QUERIES_COLLECTION=popular_queries

# Autocomplete behavior
SEARCH_AUTOCOMPLETE_DEFAULT_LIMIT=8
SEARCH_AUTOCOMPLETE_MAX_LIMIT=10
SEARCH_AUTOCOMPLETE_PREFIX_CACHE_TTL=1m
SEARCH_AUTOCOMPLETE_EMPTY_CACHE_TTL=5m
SEARCH_AUTOCOMPLETE_TYPESENSE_TIMEOUT_MS=150
SEARCH_AUTOCOMPLETE_CACHE_TIMEOUT_MS=50

# Required shared runtime values
TYPESENSE_HOST=localhost
TYPESENSE_PORT=8108
TYPESENSE_PROTOCOL=http
TYPESENSE_API_KEY=dev-typesense-key
TYPESENSE_PRODUCTS_COLLECTION=products
SEARCH_REDIS_ADDR=localhost:6379
SEARCH_REDIS_PASSWORD=dev_redis_password
SEARCH_REDIS_DB=0
```

### Variable explanation

| Variable | Required? | Purpose | Example | Security / mistake note |
|---|---:|---|---|---|
| `TYPESENSE_POPULAR_QUERIES_COLLECTION` | Yes | Collection name for popular suggestions | `popular_queries` | Must match collection ensured on startup |
| `SEARCH_AUTOCOMPLETE_DEFAULT_LIMIT` | Yes | Default suggestion count when client omits `limit` | `8` | Must be `<= SEARCH_AUTOCOMPLETE_MAX_LIMIT` |
| `SEARCH_AUTOCOMPLETE_MAX_LIMIT` | Yes | Hard max suggestions to protect public endpoint | `10` | Code rejects values over `10` |
| `SEARCH_AUTOCOMPLETE_PREFIX_CACHE_TTL` | Yes | TTL for non-empty query cache | `1m` | Too long means stale product/popular data |
| `SEARCH_AUTOCOMPLETE_EMPTY_CACHE_TTL` | Yes | TTL for empty query popular suggestions | `5m` | Usually longer than prefix TTL |
| `SEARCH_AUTOCOMPLETE_TYPESENSE_TIMEOUT_MS` | Yes | Timeout for autocomplete Typesense calls | `150` | Too low can create flaky `503` |
| `SEARCH_AUTOCOMPLETE_CACHE_TIMEOUT_MS` | Yes | Timeout for Redis cache get/set | `50` | Keep short because cache is performance helper |
| `TYPESENSE_API_KEY` | Yes | Auth for Typesense | `dev-typesense-key` | Secret. Never expose in browser |
| `SEARCH_REDIS_PASSWORD` | Required if Redis auth is enabled | Auth for Redis | `dev_redis_password` | Secret. Keep in local env/secret manager |

### How env is loaded

The Go code uses `os.Getenv`. It does not auto-load `.env`.

Local Bash run:

```bash
cd backend/services/search-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Common mistake: `infra/compose/.env.local.example` uses Docker service DNS like `typesense`. If the Go service runs on host with `go run`, use `localhost` values in `backend/services/search-service/.env`.

## 8. Docker Setup

No new Docker container, volume, network, or image is added by `TASK_FILE_NAME`.

Reuse Docker setup:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`7. Docker and DevOps Setup`
`8. Project Run Instructions`
```

Task-specific local infra command:

```bash
cp infra/compose/.env.local.example infra/compose/.env.local
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d typesense redis
```

If you keep `SEARCH_INDEXER_ENABLED=true`, also start RabbitMQ:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d rabbitmq
```

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend HTTP API | `8085` | Serves `/api/v1/search/autocomplete` through `SEARCH_HTTP_ADDR=:8085` | Reused |
| Typesense | `8108` | Search backend for `products` and `popular_queries` | Reused |
| Redis | `6379` | Autocomplete cache | Reused |
| RabbitMQ AMQP | `5672` | Product indexer queue when enabled | Reused |
| RabbitMQ Management | `15672` | Local queue debugging UI | Reused |

Networking rules:

- Host `go run`: use `localhost:8108`, `localhost:6379`, and `localhost:5672`.
- Docker network: use service names like `typesense`, `redis`, `rabbitmq`.
- Kubernetes: use DNS values from `infra/k8s/core/search-service/configmap.yaml`.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation first

Follow these before this file:

1. `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`
2. `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`
3. `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` if product indexer is enabled
4. `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` if you also test main search

### Step 2: Go to project directory

```bash
cd backend/services/search-service
```

### Step 3: Install only new dependencies if any

No new Go dependency install is needed for `TASK_FILE_NAME`.

Fresh clone command:

```bash
go mod download
```

### Step 4: Setup only new databases/services if any

No new Docker service is added. Start existing infra:

```bash
cd ../../..
cp infra/compose/.env.local.example infra/compose/.env.local
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d typesense redis
```

For index-freshness testing with product events:

```bash
docker compose --env-file infra/compose/.env.local -f infra/compose/docker-compose.local.yml up -d rabbitmq
```

### Step 5: Add only new or changed environment variables

No brand-new names beyond earlier docs. Verify Section `7. Environment Variables`.

For autocomplete-only local testing without RabbitMQ:

```bash
export SEARCH_INDEXER_ENABLED=false
```

### Step 6: Run migrations if needed

No SQL migrations are required.

Typesense collection ensure happens on service startup:

- `products`
- `popular_queries`

### Step 7: Start backend service

```bash
cd backend/services/search-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Expected startup behavior:

- Loads config from environment.
- Ensures Typesense `products` collection.
- Ensures Typesense `popular_queries` collection.
- Creates Redis cache client.
- Starts HTTP server on `SEARCH_HTTP_ADDR`, default `:8085`.
- Starts RabbitMQ product consumer only if `SEARCH_INDEXER_ENABLED=true`.

### Step 8: Verify API related to `TASK_FILE_NAME`

Health:

```bash
curl -i http://localhost:8085/healthz
```

Autocomplete with prefix:

```bash
curl -s "http://localhost:8085/api/v1/search/autocomplete?q=sho&limit=8"
```

Autocomplete empty query:

```bash
curl -s "http://localhost:8085/api/v1/search/autocomplete?limit=5"
```

Expected response shape:

```json
{
  "suggestions": []
}
```

Fresh local setup may return an empty array until `popular_queries` or `products` has documents.

## 10. Running the Project

### API contract

| Item | Value |
|---|---|
| Method | `GET` |
| Path | `/api/v1/search/autocomplete` |
| Auth | Public |
| Query params | `q`, `limit` |
| Response | `{"suggestions":["..."]}` |
| API catalog | `api/master-api.json` |
| Current runnable transport | Go HTTP handler |

### Request behavior

| Input | Behavior |
|---|---|
| Missing `q` | Allowed. Returns popular suggestions from `popular_queries`. |
| One-character `q` | Popular query lookup only. Product prefix lookup is skipped. |
| `q` length `>= 2` | Popular query lookup plus product title/brand prefix lookup. |
| Missing `limit` | Uses `SEARCH_AUTOCOMPLETE_DEFAULT_LIMIT`, default `8`. |
| `limit > SEARCH_AUTOCOMPLETE_MAX_LIMIT` | Returns `400 INVALID_AUTOCOMPLETE_REQUEST`. |
| Repeated `q` or `limit` | Returns `400 INVALID_AUTOCOMPLETE_REQUEST`. |
| Unsupported query param | Returns `400 INVALID_AUTOCOMPLETE_REQUEST`. |

### Useful manual checks

Validate bad limit:

```bash
curl -i "http://localhost:8085/api/v1/search/autocomplete?q=sho&limit=99"
```

Validate unsupported param:

```bash
curl -i "http://localhost:8085/api/v1/search/autocomplete?q=sho&page=1"
```

Inspect Redis key after a request:

```bash
docker exec ecommerce-redis redis-cli -a dev_redis_password GET "autocomplete:v1:sho:8"
```

Check Typesense collections:

```bash
curl -H "X-TYPESENSE-API-KEY: dev-typesense-key" "http://localhost:8108/collections"
```

## 11. Common Errors & Fixes

Generic Docker, Go, Typesense, Redis, and RabbitMQ troubleshooting is already documented in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`9. Common Errors and Fixes`
```

Task-specific errors:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `400 INVALID_AUTOCOMPLETE_REQUEST` | `limit` invalid, repeated params, unsupported param, or query too long | Use only `q` and `limit`; keep `limit <= 10`; keep query `<= 80` chars | Keep frontend request builder allowlisted |
| `503 SEARCH_BACKEND_UNAVAILABLE` | Typesense down, wrong API key, wrong collection, or timeout | Check `curl http://localhost:8108/health`, API key, and collection names | Start Typesense before backend |
| Response is `{"suggestions":[]}` | No matching docs in `popular_queries` or `products` | Seed `popular_queries` or index products | Add local seed data for demos/tests |
| Product title suggestions missing | `q` has fewer than 2 chars or `products` is empty | Test with 2+ chars and seed product documents | Use Task 3/Task 8 indexing flow |
| Popular suggestions missing for empty query | `popular_queries` has no active docs | Insert docs with `is_active=true` and score | Add future analytics seed/write path |
| Redis key not created | Cache set failed or Redis unavailable | Check Redis password/address and backend logs | Keep `SEARCH_REDIS_*` aligned with run location |
| Host `go run` cannot resolve `typesense` | Docker DNS name used outside Docker network | Use `TYPESENSE_HOST=localhost` for host runs | Keep separate Compose env and backend `.env` |
| Startup fails with `SEARCH_AUTOCOMPLETE_MAX_LIMIT must be less than or equal to 10` | Env value too high | Set max limit to `10` or lower | Do not increase public endpoint fanout without code review |
| Startup fails when indexer enabled | RabbitMQ or Redis unavailable | Start RabbitMQ and Redis, or set `SEARCH_INDEXER_ENABLED=false` for autocomplete-only local testing | Decide API-only vs full-service mode before running |

## 12. Security & Best Practices

Task-specific security notes:

- Do not expose `TYPESENSE_API_KEY` to browser clients. Browser should call backend/gateway only.
- Keep `SEARCH_AUTOCOMPLETE_MAX_LIMIT` small. Public autocomplete runs per keystroke, so large limits can be abused.
- Do not log raw search queries by default. Current autocomplete log stores query length, limit, cache hit, suggestion count, duration, and request id.
- Use `is_active:=true` for `popular_queries` so spam/blocked terms are not suggested.
- Keep Redis TTL short. Prefix cache default `1m` and empty-query cache default `5m` match the documented `1 to 5 min` cache guidance.
- Keep frontend debounce around `250ms` to `350ms` so the API is not called on every key event too aggressively.
- Treat Redis password, RabbitMQ URL, and Typesense API key as secrets. Use `.env` locally and Kubernetes Secret in cluster.
- Keep local Compose CORS setting limited to dev. Typesense should not be directly browser-facing in production.

Previously documented broader audit:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`10. Security and Configuration Audit`
```

## 13. Missing or Misconfigured Things

| Finding | Impact | Suggested fix |
|---|---|---|
| `proto/` directory is absent in current repo snapshot | `api/master-api.json` mentions gRPC, but runnable code is HTTP only | Add proto/gRPC generation when gateway integration is implemented |
| `popular_queries` write/analytics ingestion is not implemented in this task | Empty-query suggestions may be empty unless manually seeded | Add ingestion/batch seed in the later analytics task |
| Compose stack starts infra only, not backend service | Beginners may expect API on `8085` after Compose up | Run Go service separately or add service image later |
| `infra/compose/.env.local.example` uses Docker DNS names | Host `go run` can fail if same file is sourced directly | Use `backend/services/search-service/.env` with localhost values |
| Metrics recorder defaults to no-op | Autocomplete metrics described in task guide are not exported yet | Wire a real Prometheus/OpenTelemetry recorder later |
| Redis cache errors are non-fatal | Good for availability, but cache outage may be missed | Alert on `search.autocomplete.cache_get_failed` and cache set warnings |
| Product suggestions depend on indexed product docs | Fresh local Typesense may return only popular query suggestions | Seed products through Task 3 indexer or future reindex task |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Go Dependency System` | Same Go module setup and commands apply |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3.2 Redis` | Redis installation and credential setup are unchanged |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Environment Variables` | Full `.env` is already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Docker and DevOps Setup` | Same Compose stack is reused |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Common Errors and Fixes` | Generic infra troubleshooting is reused |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `4.1 Typesense` | Typesense runtime, ports, credentials, and Docker setup are unchanged |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Environment Variables` | Typesense collection env pattern is reused |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `6. Redis / Queue / External Services` | RabbitMQ/indexer setup is unchanged and only needed for product index freshness |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `8. Docker Setup` and `9. Local Development Setup` | Same service run and networking pattern applies |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before creating this file
- [ ] No duplicate Go, Docker, Typesense, Redis, or RabbitMQ installation docs copied
- [ ] `backend/services/search-service/.env` has autocomplete variables verified
- [ ] `SEARCH_AUTOCOMPLETE_DEFAULT_LIMIT <= SEARCH_AUTOCOMPLETE_MAX_LIMIT`
- [ ] `SEARCH_AUTOCOMPLETE_MAX_LIMIT <= 10`
- [ ] Typesense is running on port `8108`
- [ ] Redis is running on port `6379`
- [ ] `products` collection exists in Typesense
- [ ] `popular_queries` collection exists in Typesense
- [ ] Optional `popular_queries` seed document added for local demo
- [ ] RabbitMQ is running or `SEARCH_INDEXER_ENABLED=false` is set for API-only local testing
- [ ] Backend starts on `SEARCH_HTTP_ADDR`, default `:8085`
- [ ] `GET /healthz` returns `200`
- [ ] `GET /api/v1/search/autocomplete?q=sho&limit=8` returns valid JSON
- [ ] Invalid limit returns clean `400 INVALID_AUTOCOMPLETE_REQUEST`
- [ ] Typesense failure returns clean `503 SEARCH_BACKEND_UNAVAILABLE`
- [ ] Logs checked for `search.autocomplete` entries
- [ ] No original `TASK_FILE_NAME` file was modified
