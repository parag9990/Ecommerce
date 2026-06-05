# Project Dependency & Setup Guide

## 1. Project Overview

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Search Service` |
| `TASK_FILE_NAME` | `task4.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task4_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

Task 4 ka focus Search API hai. Simple Hinglish me: buyer ya frontend `GET /api/v1/search` call karta hai, service query params validate karti hai, Typesense se ranked product ids aur facets leti hai, phir Product Service se product summary hydrate karke JSON response return karti hai.

Important boundary:

- Business logic ya API implementation yahan rewrite nahi ki gayi.
- Original `task4.md` modify nahi kiya gaya.
- Common setup duplicate nahi kiya gaya. Previous dependency files ko reference kiya gaya hai.
- Is task me koi new database, queue, Docker container, ya Go dependency introduce nahi hoti.

Task-specific runtime flow:

```text
Client / API Gateway
  -> GET /api/v1/search
  -> Search Service HTTP server
  -> Typesense products collection
  -> Product Service batch-get endpoint
  -> JSON response with products, facets, total
```

## 2. Tech Stack

### Reused tech from previous dependency files

| Technology | Required for Task 4? | Reuse reference |
|---|---:|---|
| Go | Yes | Refer `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `2. Go Dependency System` |
| Go modules | Yes | Refer `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `2. Go Dependency System` |
| Typesense | Yes | Refer `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`, Section `4.1 Typesense` |
| Docker Compose | Recommended for local infra | Refer `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `7. Docker and DevOps Setup` |
| Redis | Required by current service startup | Refer `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `3.2 Redis` |
| RabbitMQ | Required when `SEARCH_INDEXER_ENABLED=true` | Refer `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`, Section `6. Redis / Queue / External Services` |

### Task 4 specific tech details

| Technology / Component | Required? | What it does in simple English + Hinglish |
|---|---:|---|
| HTTP Search API | Yes | Public search endpoint expose karta hai: `GET /api/v1/search`. Frontend ke liye ye simple REST interface hai. |
| Typesense search query | Yes | Search text, filters, sort, pagination ko Typesense query params me convert karta hai. |
| Product Service hydration | Yes for non-empty results | Typesense sirf ranked ids return karta hai; final product card data Product Service se aata hai. |
| Request/analytics headers | Optional but useful | `X-Request-ID`, `X-Anonymous-ID`, `X-Session-ID`, `X-User-ID`, `X-Client-Path` tracing aur zero-result analytics me help karte hain. |

## 3. Required Software

No new required software is introduced by `TASK_FILE_NAME`.

Follow the already documented base setup:

- Go runtime and module commands: `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `2. Go Dependency System`
- Docker and Docker Compose: `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `7. Docker and DevOps Setup`
- Typesense runtime: `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`, Section `4.1 Typesense`
- Redis and RabbitMQ infra: `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Sections `3.2 Redis` and `5.3 RabbitMQ`

Task 4 additionally expects Product Service to be reachable at the configured internal URL. Beginner note: Product Service ek separate backend service hai jo final product details ka source of truth hai. Search API Typesense result ko product cards me convert karne ke liye Product Service ko call karti hai.

## 4. Dependency Management

### Go dependency status

`backend/services/search-service/go.mod` already contains the dependencies needed by Task 4:

| Dependency | Why Task 4 uses it | New in Task 4? |
|---|---|---:|
| `github.com/typesense/typesense-go/v2` | Typesense product search call karne ke liye | No |
| Go standard `net/http` | HTTP endpoint, handler, client calls ke liye | No |
| Go standard `context`, `time` | Typesense and Product Service timeout control ke liye | No |

No new `go get` command is required for this task.

Use previous Go commands:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`2. Go Dependency System`
```

Task-specific verification command:

```bash
cd backend/services/search-service
go test ./internal/domain ./internal/usecase ./internal/transport/http ./internal/clients
```

## 5. Database Setup

No new SQL or NoSQL database is introduced by Task 4.

| Storage / Service | Used by Task 4? | Purpose | Setup action |
|---|---:|---|---|
| Typesense `products` collection | Yes | Search index query, facets, sort, pagination | Reuse Task 2 setup |
| Product Service database | Indirect | Product Service owns canonical product summaries | Start/configure Product Service separately if testing real hydration |
| Redis | Indirect in current service startup and zero-result dedupe | Existing service initialization and analytics dedupe | Reuse Task 1 setup |
| RabbitMQ | Not needed for API-only query path, but needed if indexer enabled | Product events keep Typesense index fresh | Reuse Task 3 setup |

### Typesense collection requirement

Task 4 requires the `products` collection to exist and contain indexed product documents if you want non-empty search results.

This setup is already explained in:

```md
`TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`

Sections:
`4.1 Typesense`
`9. Project Run Instructions`
`10. Migrations and Data Setup`
```

Fresh local setup me empty result normal ho sakta hai. API healthy ho sakti hai even when no product documents are indexed yet.

### Migrations

No SQL migration is required for Task 4.

Search Service schema setup is Typesense collection ensure, not SQL migration. This is already covered in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`8. Project Run Instructions`
```

## 6. Redis / Queue / External Services

### Product Service hydration

Task 4 ka most important runtime dependency Product Service hydration hai.

| Item | Value |
|---|---|
| Base URL env | `PRODUCT_SERVICE_URL` |
| Batch endpoint env | `PRODUCT_SERVICE_BATCH_GET_PATH` |
| Default base URL | `http://localhost:8082` |
| Default path | `/internal/v1/products:batchGet` |
| Timeout env | `PRODUCT_SERVICE_TIMEOUT_MS` |
| Default timeout | `300` ms |

How it works:

1. Typesense returns ranked product ids.
2. Search Service sends those ids to Product Service.
3. Product Service returns product summaries.
4. Search Service preserves Typesense ranking order in the final response.

Beginner note: Agar Typesense result me `prod_2, prod_1` order aaya, final JSON me bhi same ranking preserve honi chahiye. Product Service sirf details deta hai, ranking decide nahi karta.

### Typesense search backend

Typesense detailed setup duplicate nahi kiya gaya.

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`

Section:
`4.1 Typesense`
```

Task 4 only needs these search API behaviors from Typesense:

| Behavior | Why required |
|---|---|
| Full-text query | `q` search text handle karne ke liye |
| Facets | brand, category, seller, price, rating, stock counts return karne ke liye |
| Filters | buyer-selected filters apply karne ke liye |
| Sorting | relevance, price, rating, newest, popular sort support karne ke liye |
| Pagination | `page` and `page_size` support karne ke liye |

### Redis and RabbitMQ

Task 4 does not add new Redis or RabbitMQ setup.

Reuse:

- Redis setup: `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, Section `3.2 Redis`
- RabbitMQ product indexer setup: `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`, Section `6. Redis / Queue / External Services`

Local API-only shortcut:

```bash
SEARCH_INDEXER_ENABLED=false go run ./cmd/server
```

Use this only when you want Search API boot without RabbitMQ consumer. Redis is still initialized by current server wiring, so keep Redis configured unless code changes later.

## 7. Environment Variables

Task 4 introduces no brand-new environment variable names, but these existing variables become important for Search API behavior.

### Task 4 specific `.env` values

Add or verify these values in:

```text
backend/services/search-service/.env
```

```env
# Search API pagination behavior
SEARCH_DEFAULT_PAGE_SIZE=20
SEARCH_MAX_PAGE_SIZE=100

# Typesense search backend
TYPESENSE_PRODUCTS_COLLECTION=products
TYPESENSE_TIMEOUT_MS=300

# Product hydration dependency
PRODUCT_SERVICE_URL=http://localhost:8082
PRODUCT_SERVICE_BATCH_GET_PATH=/internal/v1/products:batchGet
PRODUCT_SERVICE_TIMEOUT_MS=300

# Optional zero-result analytics context
SEARCH_ZERO_RESULT_TRACKING_ENABLED=true
SESSION_SERVICE_URL=http://localhost:8086
SESSION_SERVICE_INGEST_PATH=/api/v1/sessions/events
SESSION_SERVICE_TIMEOUT_MS=150
```

Do not duplicate the full `.env` from earlier docs. For full local env:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Environment Variables`
```

### Variable explanation

| Variable | Required? | Purpose | Example | Security / mistake note |
|---|---:|---|---|---|
| `SEARCH_DEFAULT_PAGE_SIZE` | Yes | Default `page_size` when client does not send one | `20` | Must be `<= SEARCH_MAX_PAGE_SIZE` |
| `SEARCH_MAX_PAGE_SIZE` | Yes | Hard limit to avoid heavy Typesense/Product Service calls | `100` | Keep bounded; high value can slow API |
| `TYPESENSE_PRODUCTS_COLLECTION` | Yes | Collection searched by Task 4 | `products` | Must match collection created by schema/indexer |
| `TYPESENSE_TIMEOUT_MS` | Yes | Timeout for Typesense search call | `300` | Too low causes flaky `SEARCH_BACKEND_UNAVAILABLE` |
| `PRODUCT_SERVICE_URL` | Yes | Product Service base URL for hydration | `http://localhost:8082` | Use Docker/K8s DNS name when running inside cluster/network |
| `PRODUCT_SERVICE_BATCH_GET_PATH` | Yes | Internal Product Service batch-get route | `/internal/v1/products:batchGet` | Wrong path causes `PRODUCT_SERVICE_UNAVAILABLE` |
| `PRODUCT_SERVICE_TIMEOUT_MS` | Yes | Product hydration timeout | `300` | Too low causes `503` under slow local service |
| `SEARCH_ZERO_RESULT_TRACKING_ENABLED` | Optional | Sends no-result query analytics | `true` | Set `false` if Session Service is unavailable during API-only local testing |

Common Task 4 `.env` mistakes:

| Mistake | Symptom | Fix |
|---|---|---|
| `PRODUCT_SERVICE_URL` points to wrong port | Search API returns `503 PRODUCT_SERVICE_UNAVAILABLE` for non-empty hits | Use the actual Product Service host and port |
| `TYPESENSE_PRODUCTS_COLLECTION` does not match indexed collection | Search returns empty results or backend error | Keep value as `products` unless reindex alias strategy changed it |
| `SEARCH_DEFAULT_PAGE_SIZE` greater than `SEARCH_MAX_PAGE_SIZE` | Service fails during config load | Lower default or raise max carefully |
| Using Docker service DNS from host shell | Host `go run` cannot resolve names like `product-service` | Use `localhost` when running service on host |

## 8. Docker Setup

Task 4 adds no new Docker container, volume, network, or image.

Reuse Docker setup:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`7. Docker and DevOps Setup`
`8. Project Run Instructions`
```

Task 4 Docker networking note:

| Search Service location | Product Service URL style | Typesense URL style |
|---|---|---|
| Running on host with `go run` | `http://localhost:8082` | `http://localhost:8108` |
| Running inside Docker Compose network | `http://product-service:8082` if service exists in Compose | `http://typesense:8108` |
| Running in Kubernetes | Cluster DNS service name | Kubernetes service DNS for Typesense |

No Docker Compose file change is required for this dependency document.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation first

Follow these in order:

1. `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`
2. `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`
3. `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` if indexer/event ingestion is enabled

### Step 2: Go to project directory

```bash
cd backend/services/search-service
```

### Step 3: Install only new dependencies if any

No new dependency install is needed.

If this is a fresh clone, use the previous Go module setup:

```bash
go mod download
```

### Step 4: Setup only new databases/services if any

No new database/service is added.

For Task 4 functional testing, make sure these existing services are available:

| Service | Needed for | Setup reference |
|---|---|---|
| Typesense | Search index query | `task2_Dependency.md`, Section `4.1 Typesense` |
| Product Service | Product hydration for non-empty hits | `task1_Dependency.md`, Section `5.4 Product Service` |
| Redis | Current server startup and zero-result dedupe | `task1_Dependency.md`, Section `3.2 Redis` |
| RabbitMQ | Product index freshness if indexer enabled | `task3_Dependency.md`, Section `6. Redis / Queue / External Services` |

### Step 5: Add only new or changed environment variables

No new env names are introduced. Verify Task 4 values from Section `7. Environment Variables`.

For API-only local testing without RabbitMQ:

```bash
export SEARCH_INDEXER_ENABLED=false
```

For local testing without Session Service:

```bash
export SEARCH_ZERO_RESULT_TRACKING_ENABLED=false
```

### Step 6: Run migrations if needed

No SQL migrations are needed.

Make sure Typesense collection exists by starting Search Service once or following Task 2 verification.

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
- Ensures Typesense `products` and `popular_queries` collections.
- Creates Redis client.
- Creates Product Service HTTP client.
- Starts HTTP server on `SEARCH_HTTP_ADDR`, default `:8085`.
- Starts RabbitMQ product consumer only if `SEARCH_INDEXER_ENABLED=true`.

### Step 8: Verify API related to `TASK_FILE_NAME`

Health check:

```bash
curl -i http://localhost:8085/healthz
```

Basic search:

```bash
curl -s "http://localhost:8085/api/v1/search?q=running%20shoes&page=1&page_size=20"
```

Search with filters and sort:

```bash
curl -s "http://localhost:8085/api/v1/search?q=shoes&brand=Nike&brand=Puma&category_id=cat_shoes&min_price=1000&max_price=5000&min_rating=4&in_stock=true&sort=price_asc&page=1&page_size=20"
```

Expected response shape:

```json
{
  "products": [],
  "facets": {
    "brand": [],
    "category_ids": [],
    "seller_id": [],
    "price": [],
    "rating": [],
    "in_stock": []
  },
  "total": 0
}
```

Fresh setup me `products` empty ho sakta hai. Iska matlab API broken nahi hai. Non-empty result ke liye Task 3 indexer events ya reindex flow se product documents Typesense me aane chahiye.

## 10. Running the Project

### API contract

| Item | Value |
|---|---|
| Method | `GET` |
| Path | `/api/v1/search` |
| Auth | Public |
| Contract source | `api/master-api.json` |
| Internal contract name | `SearchService.SearchProducts` |

### Supported query params

| Query param | Type | Example | Notes |
|---|---|---|---|
| `q` | string | `running shoes` | Empty query becomes `*` |
| `brand` | repeated or comma style | `brand=Nike&brand=Puma` | Repeated brand is allowed |
| `category_id` | string | `cat_shoes` | ID-like value only |
| `seller_id` | string | `seller_123` | ID-like value only |
| `min_price` | number | `1000` | Must be non-negative |
| `max_price` | number | `5000` | Must be non-negative and `>= min_price` |
| `min_rating` | number | `4` | Must be between `0` and `5` |
| `in_stock` | boolean | `true` | Strict boolean |
| `sort` | string | `price_asc` | Must be supported sort key |
| `page` | integer | `1` | Negative/invalid values rejected |
| `page_size` | integer | `20` | Max controlled by `SEARCH_MAX_PAGE_SIZE` |

Supported sort values:

| Sort | Meaning |
|---|---|
| `relevance` | Text match with popularity tie-breaker |
| `popular` | Popularity score descending |
| `price_asc` | Cheapest first |
| `price_desc` | Most expensive first |
| `rating_desc` | Best rated first |
| `newest` | New arrivals first |

Useful request headers:

| Header | Purpose |
|---|---|
| `X-Request-ID` | Request tracing |
| `X-Correlation-ID` | Fallback request tracing |
| `X-Anonymous-ID` | Anonymous analytics |
| `X-Session-ID` | Session analytics |
| `X-User-ID` | User analytics |
| `X-Client-Path` | Frontend page path for zero-result analytics |

## 11. Common Errors & Fixes

Generic Docker, Go, Typesense, Redis, and RabbitMQ issues are already documented in:

```md
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`9. Common Errors and Fixes`
```

Task 4 specific errors:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `400 INVALID_SEARCH_REQUEST` | Invalid `page`, `page_size`, unsupported query param, or too-long query | Use only supported params and valid integers | Keep frontend query builder allowlisted |
| `400 INVALID_SEARCH_FILTER` | Invalid filter value, repeated non-brand filter, `min_price > max_price`, `min_rating > 5` | Correct filter values | Validate filters in frontend before API call |
| `400 UNSUPPORTED_SEARCH_SORT` | Unknown `sort` value | Use `relevance`, `popular`, `price_asc`, `price_desc`, `rating_desc`, or `newest` | Keep sort dropdown options synced with backend policy |
| `503 SEARCH_BACKEND_UNAVAILABLE` | Typesense down, wrong API key, wrong collection, timeout | Check Typesense health and `.env` values | Start Typesense before Search Service |
| `503 PRODUCT_SERVICE_UNAVAILABLE` | Product Service down, wrong `PRODUCT_SERVICE_URL`, wrong batch path, or slow response | Start Product Service and verify `/internal/v1/products:batchGet` | Keep Product Service health check in local/dev stack |
| Empty `products` with `total=0` | No indexed product documents | Run product indexer flow or reindex | Seed local Typesense before demo/testing |
| Empty `products` but `total>0` | Typesense returned ids but Product Service did not return matching products | Check Product Service data and id mapping | Keep product ids consistent between Typesense and Product Service |
| Service fails with page size config error | `SEARCH_DEFAULT_PAGE_SIZE > SEARCH_MAX_PAGE_SIZE` | Fix `.env` values | Keep default page size below max |

## 12. Security & Best Practices

Task 4 specific best practices:

- Do not expose Typesense API key to browser clients. Browser should call Search API, not Typesense directly.
- Keep `SEARCH_MAX_PAGE_SIZE` bounded. Large page size can overload Typesense and Product Service hydration.
- Use short timeouts for search, but not unrealistically short. `300ms` is good for local MVP, but tune using real latency.
- Log request ids, query length, result count, and latency, but do not log raw sensitive user data unnecessarily.
- Keep Product Service batch-get endpoint internal-only. Public users should not call internal hydration endpoints directly.
- Validate frontend filters against the same backend allowlist to avoid noisy `400` responses.
- Keep search index ids aligned with Product Service ids. Ranking is useless if hydration cannot find products.

## 13. Missing or Misconfigured Things

| Finding | Impact | Suggested fix |
|---|---|---|
| No new Docker Compose Product Service wiring is documented for Task 4 | Real hydration testing can fail if Product Service is not running | Add Product Service to local compose stack when that service is ready |
| Search API depends on Product Service availability for non-empty results | Product Service outage turns successful Typesense hits into `503` | Add graceful fallback only if product snapshot in Typesense is acceptable for the product team |
| API Gateway integration is contract-level in docs, but current code exposes HTTP directly | Deployment routing can be unclear for beginners | Document whether local clients call Search Service directly or via API Gateway |
| `SEARCH_ZERO_RESULT_TRACKING_ENABLED=true` requires Session Service config validation | API-only local boot may fail if Session Service URL is invalid | Set `SEARCH_ZERO_RESULT_TRACKING_ENABLED=false` for isolated local testing |
| Full readiness endpoint is not present | `/healthz` only confirms process is alive | Add readiness check for Typesense, Redis, Product Service, and optional RabbitMQ |

Previously documented broader security/configuration notes:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`10. Security and Configuration Audit`
```

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Go Dependency System` | Same Go module setup and commands apply |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3.2 Redis` | Redis setup is unchanged |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Environment Variables` | Full `.env` already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5.4 Product Service` | Product Service dependency already introduced; Task 4 only emphasizes hydration |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Docker and DevOps Setup` | Docker Compose infra is unchanged |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Common Errors and Fixes` | Generic troubleshooting already covered |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `4.1 Typesense` | Typesense installation/runtime setup is unchanged |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `10. Migrations and Data Setup` | Typesense collection setup replaces SQL migrations |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `6. Redis / Queue / External Services` | RabbitMQ/indexer setup is unchanged |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `10. Running the Project` | Product event smoke testing can seed Typesense for non-empty search |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before Task 4 setup
- [ ] No duplicate Docker/Go/Typesense/Redis/RabbitMQ setup copied
- [ ] `backend/services/search-service/.env` has Task 4 values verified
- [ ] `SEARCH_DEFAULT_PAGE_SIZE <= SEARCH_MAX_PAGE_SIZE`
- [ ] Typesense is running and `products` collection exists
- [ ] Product Service is reachable at `PRODUCT_SERVICE_URL`
- [ ] `PRODUCT_SERVICE_BATCH_GET_PATH` points to batch-get endpoint
- [ ] Redis is running for current service startup
- [ ] RabbitMQ is running or `SEARCH_INDEXER_ENABLED=false` is set for API-only local boot
- [ ] Search Service starts on `SEARCH_HTTP_ADDR`, default `:8085`
- [ ] `GET /healthz` returns `200`
- [ ] `GET /api/v1/search?q=...` returns valid JSON
- [ ] Invalid filters and unsupported sort values return clean `400` errors
- [ ] Typesense/Product Service failures return clean `503` errors
- [ ] Logs checked for request id, search latency, hydration latency, and dependency errors
- [ ] No original implementation task file was modified
