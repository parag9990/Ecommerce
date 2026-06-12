# Project Dependency & Setup Guide

```text
SERVICE_NAME = "Cart Service"
TASK_FILE_NAME = "task4.md"
INPUT_FILE_PATH = TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}
OUTPUT_FILE_NAME = ${TASK_FILE_NAME without ".md"}_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}

PREVIOUS_TASK1_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
PREVIOUS_TASK2_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
PREVIOUS_TASK3_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task3_Dependency.md
```

## 1. Project Overview

Ye document `${SERVICE_NAME}` ke `${TASK_FILE_NAME}` ke liye dependency, setup, environment, external service, DevOps, aur troubleshooting guide hai.

`${TASK_FILE_NAME}` ka focus Add Item flow hai:

- Shopper `POST /api/v1/cart/items` call karta hai.
- Request me `product_id`, `variant_id`, and `quantity` aata hai.
- `${SERVICE_NAME}` Product Service se product and variant validate karta hai.
- Stock, price, product status, and variant status check hota hai.
- MongoDB me active cart create/update hota hai.
- Redis me active cart and summary cache refresh hota hai.

Important scope:

- `${INPUT_FILE_PATH}` modify nahi kiya gaya.
- `${OUTPUT_FILE_PATH}` new dependency/setup guide hai.
- No new Go library, database engine, Redis key family, queue, Dockerfile, or service port was introduced by `${TASK_FILE_NAME}`.
- The main task-specific setup impact is Product Service availability and correct `CART_PRODUCT_BASE_URL` configuration.

### Files Inspected

| File | Why inspected |
|---|---|
| `INPUT_FILE_PATH` | `${TASK_FILE_NAME}` scope: Add Item flow and Product Service dependency |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | Base setup: Go, MongoDB, Redis, Docker, `.env`, run commands |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | MongoDB source of truth plus Redis cache setup |
| `PREVIOUS_TASK3_DEPENDENCY_FILE` | `cart_db.carts` schema, indexes, and migration/bootstrap setup |
| `backend/services/cart-service/go.mod` | Confirm no new Go dependency for Add Item |
| `backend/services/cart-service/.env.example` | Actual env variable names and default local values |
| `backend/services/cart-service/internal/config/config.go` | Required env validation and fallback names |
| `backend/services/cart-service/internal/client/product_http_client.go` | Product Service URL, endpoint path, response shape, timeout |
| `backend/services/cart-service/internal/usecase/add_item.go` | Add Item Product Service, MongoDB, Redis, and retry behavior |
| `backend/services/cart-service/internal/domain/cart_mutation.go` | Quantity, stock, snapshot, totals, and coupon clear rules |
| `backend/services/cart-service/internal/transport/http/handler.go` | Public route, headers, request body, and error mapping |
| `backend/services/cart-service/cmd/server/main.go` | Startup dependency checks and HTTP server setup |
| `api/master-api.json` | API contract for `cart.add_item` and `CartItemInput` |

## 2. Tech Stack

Most technologies are reused from previous dependency files. Full beginner setup is not duplicated here.

| Technology | Required? | Status for `${TASK_FILE_NAME}` | What beginner should know | Setup reference |
|---|---:|---|---|---|
| Go `1.26.3` | Yes | Reused | Backend service Go me implemented hai. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4. Go Dependency System` |
| Go modules | Yes | Reused | `go.mod` and `go.sum` dependencies manage karte hain. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4. Go Dependency System` |
| Standard `net/http` | Yes | Reused | Public Add Item route HTTP se expose hota hai. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `3. High-Level Tech Stack` |
| MongoDB | Yes | Reused | Durable active cart yahin save hota hai. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `5.1 MongoDB` |
| MongoDB schema/indexes | Yes | Reused | Add Item `cart_db.carts` schema and unique active cart indexes use karta hai. | `PREVIOUS_TASK3_DEPENDENCY_FILE`, section `5. Database Setup` |
| Redis | Yes | Reused | Add Item ke baad cart and summary cache refresh hota hai. | `PREVIOUS_TASK2_DEPENDENCY_FILE`, section `6.2 Redis` |
| Product Service HTTP API | Yes | Task-specific focus | Product/variant/price/stock validation ke liye mandatory hai. | Explained in this file |
| CMS Service HTTP API | Yes for current server startup | Reused | Add Item CMS ko call nahi karta, but current server config CMS URL require karta hai. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `6.2 CMS Service` |
| Docker | Optional but recommended | Reused | Local MongoDB/Redis run karne ke liye useful hai. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `9. Docker and DevOps Setup` |

Simple Hinglish:

Product Service ke bina Add Item flow complete nahi hoga. Client se price trust nahi hota. `${SERVICE_NAME}` Product Service se latest product snapshot leta hai, phir MongoDB me cart save karta hai, phir Redis cache update karta hai.

## 3. Required Software

Base software install already documented hai. Yahan sirf `${TASK_FILE_NAME}` ka incremental view diya gaya hai.

| Software / Service | Required? | Status | Notes |
|---|---:|---|---|
| Git | Yes | Reused | Clone/setup steps previous docs me hain. |
| Go `1.26.3` | Yes | Reused | Service build/run/test ke liye. |
| MongoDB | Yes | Reused | Active cart create/update ke liye. |
| Redis | Yes | Reused | Cache refresh and startup ping ke liye. |
| Product Service | Yes for Add Item | Task-specific focus | Must expose `GET /api/v1/products/{product_id}`. |
| CMS Service | Yes for current server startup | Reused | Add Item nahi use karta, but config validation requires URL. |
| `curl` | Recommended | Reused | Health check and Add Item API verify karne ke liye. |
| Docker / Docker Compose | Optional | Reused | MongoDB/Redis local setup easy banata hai. |

Important repo inspection note:

- `backend/services/product-service` currently contains only `.env` during inspection.
- No runnable Product Service Go module was found in this repository snapshot.
- End-to-end Add Item local verification ke liye Product Service separately run karo, API Gateway ke through reachable banao, ya local mock server use karo jo required product JSON return kare.

## 4. Dependency Management

This is a Go project. `${TASK_FILE_NAME}` does not add any new Go module dependency.

Current direct dependencies are unchanged:

| Dependency | Version | Status | Why used |
|---|---:|---|---|
| `go.mongodb.org/mongo-driver` | `v1.17.9` | Reused | MongoDB repository, schema, indexes |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Reused | Redis active cart and summary cache |

Product Service HTTP integration uses Go standard library:

```text
net/http
net/url
encoding/json
```

No `go get` is needed for `${TASK_FILE_NAME}`.

Refer for dependency commands:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 4.4 Dependency Commands
```

Useful check from service folder:

```bash
cd backend/services/cart-service
go test ./...
```

Beginner note:

`go.mod` me unnecessary dependency add mat karo. Add Item flow already standard HTTP client se Product Service call karta hai.

## 5. Database Setup

MongoDB setup is reused. No new database, collection, migration, or index was added by `${TASK_FILE_NAME}`.

Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 5.1 MongoDB

PREVIOUS_TASK3_DEPENDENCY_FILE
Section: 5. Database Setup
```

`${TASK_FILE_NAME}` database behavior:

| Topic | Value |
|---|---|
| Database | `cart_db` |
| Collection | `carts` |
| Source of truth | MongoDB |
| Main repository methods used | `FindActiveByOwner`, `CreateActiveCart`, `SaveWithVersion`, `ExpireActiveCart` |
| Required schema/index setup | Same schema/index setup already documented in `PREVIOUS_TASK3_DEPENDENCY_FILE` |
| New migration needed | No |

Add Item writes:

- If active cart nahi hai, new active cart create hota hai.
- If active cart exists, same `product_id + variant_id` line item increment hota hai.
- Cart version optimistic concurrency ke liye use hota hai.
- Stale coupon preview clear hota hai, because cart contents changed.
- Totals server side recalculate hote hain.

Task-specific database verification:

```bash
mongosh "mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin" \
  --quiet \
  --eval "db.getSiblingDB('cart_db').carts.find({ status: 'active' }).limit(3).pretty()"
```

If collection/schema missing ho:

```text
PREVIOUS_TASK3_DEPENDENCY_FILE
Section: 5. Database Setup
Topic: Migration / Bootstrap Options
```

## 6. Redis / Queue / External Services

### 6.1 Product Service

Product Service mandatory hai for `${TASK_FILE_NAME}`.

Simple Hinglish:

Product Service catalog ka owner hai. `${SERVICE_NAME}` product price, stock, title, image, seller, and variant status khud se assume nahi karta. Add Item ke time Product Service se fresh data lekar cart item snapshot banata hai.

Current implementation calls:

```text
GET ${CART_PRODUCT_BASE_URL}/api/v1/products/{product_id}
Accept: application/json
```

Default local value:

```env
CART_PRODUCT_BASE_URL=http://localhost:8082
CART_PRODUCT_REQUEST_TIMEOUT=2s
```

Fallback env name:

```env
PRODUCT_SERVICE_BASE_URL=http://localhost:8082
```

Use `CART_PRODUCT_BASE_URL` as preferred name.

Required Product Service response fields:

| Field | Required? | Why |
|---|---:|---|
| `product_id` or `id` | Yes | Product identity |
| `seller_id` | Yes | Cart item seller snapshot |
| `title` or `name` | Yes | Cart display snapshot |
| `image_url` or first `images[].url` | Optional but useful | Cart thumbnail snapshot |
| `status` | Yes | Must be `published`, `active`, or `sellable` |
| `variants[]` | Yes | Variant lookup |
| `variant_id` or variant `id` | Yes | Requested variant identity |
| `price.amount` or `unit_price.amount` or `price_amount` | Yes | Unit price in minor units |
| `price.currency`, `unit_price.currency`, or `currency` | Yes | Currency, for example `INR` |
| `stock_quantity` or `inventory_quantity` | Yes | Stock check |
| `available` / `enabled` / variant `status` | Recommended | Variant sellable decision |

Accepted response shapes:

```json
{
  "product_id": "prod_123",
  "seller_id": "seller_456",
  "title": "Running Shoes",
  "image_url": "https://cdn.example.com/prod_123/main.jpg",
  "status": "published",
  "variants": [
    {
      "variant_id": "var_1",
      "sku": "SHOE-BLK-9",
      "attributes": {
        "size": "9",
        "color": "Black"
      },
      "price": {
        "amount": 299900,
        "currency": "INR"
      },
      "stock_quantity": 5,
      "available": true,
      "status": "active"
    }
  ]
}
```

Wrapper responses are also accepted:

```json
{ "product": { "...": "..." } }
```

```json
{ "data": { "...": "..." } }
```

```json
{ "item": { "...": "..." } }
```

Product Service setup rules:

- Product Service must be reachable from where `${SERVICE_NAME}` is running.
- If `${SERVICE_NAME}` runs on host machine, `http://localhost:8082` is okay.
- If `${SERVICE_NAME}` runs inside Docker Compose, `localhost` means the cart container itself. Use service DNS like `http://product-service:8082`.
- Do not point `CART_PRODUCT_BASE_URL` to a frontend URL. It must return product API JSON.
- Do not include `/api/v1/products` in the base URL. Current client appends `/api/v1/products/{product_id}` automatically.

Verify Product Service before Add Item:

```bash
curl http://localhost:8082/api/v1/products/prod_123
```

Expected:

- HTTP `200`
- Product has requested variant.
- Product status is sellable.
- Variant stock is enough.
- Price amount and currency are present.

### 6.2 Redis

Redis setup is reused. Add Item does not introduce new Redis key patterns.

Refer:

```text
PREVIOUS_TASK2_DEPENDENCY_FILE
Section: 6.2 Redis
```

Add Item cache behavior:

| Key pattern | Status | Purpose |
|---|---|---|
| `cart:active:user:{user_id}` | Reused | Updated full cart cache for logged-in user |
| `cart:active:guest:{guest_session_id}` | Reused | Updated full cart cache for guest |
| `cart:summary:user:{user_id}` | Reused | Header/badge summary for logged-in user |
| `cart:summary:guest:{guest_session_id}` | Reused | Header/badge summary for guest |

Important:

- MongoDB save failure fails the Add Item request.
- Redis cache refresh failure is logged as warning, but the updated cart can still be returned because MongoDB is source of truth.

### 6.3 CMS Service

CMS Service is not used by Add Item, but current server startup validates and creates coupon validator config.

Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 6.2 CMS Service
```

For local startup, keep these configured:

```env
CART_CMS_BASE_URL=http://localhost:8089
CART_CMS_VALIDATE_COUPON_PATH=/internal/v1/coupons/validate
```

### 6.4 Queues and Other Services

| Service | Used by `${TASK_FILE_NAME}`? | Setup needed now |
|---|---:|---|
| Kafka | No | No |
| RabbitMQ | No | No |
| NATS | No | No |
| MinIO / S3 | No | No |
| Elasticsearch / Typesense | No | No |
| SMTP / Stripe / Twilio / OAuth | No | No |

## 7. Environment Variables

Full `.env` setup is already documented in:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 8. Complete Environment Variables
```

Create local env file here:

```text
backend/services/cart-service/.env
```

Important beginner note:

Current Go code uses `os.Getenv`. There is no automatic `.env` loader. `.env` file create karna enough nahi hai; run command se pehle source/export karna hoga.

Task-specific env focus:

```env
CART_PRODUCT_BASE_URL=http://localhost:8082
CART_PRODUCT_REQUEST_TIMEOUT=2s
CART_MAX_SAVE_ATTEMPTS=3
```

| Variable | Required? | Purpose | Example | Security / setup note |
|---|---:|---|---|---|
| `CART_PRODUCT_BASE_URL` | Yes | Product Service base URL for product/variant lookup | `http://localhost:8082` | Internal service URL rakho, public frontend URL nahi |
| `PRODUCT_SERVICE_BASE_URL` | Fallback only | Legacy fallback if preferred env missing | `http://localhost:8082` | Prefer `CART_PRODUCT_BASE_URL` |
| `CART_PRODUCT_REQUEST_TIMEOUT` | Optional | Product Service call timeout | `2s` | Too low hoga to false failures; too high hoga to Add Item slow |
| `CART_MAX_SAVE_ATTEMPTS` | Optional | Version conflict retry attempts | `3` | High value blindly set mat karo; concurrency issue hide ho sakta hai |

Existing env still required for service startup:

```text
CART_MONGO_URI
CART_MONGO_DATABASE
CART_REDIS_ADDR
CART_CMS_BASE_URL
CART_CMS_VALIDATE_COUPON_PATH
CART_HTTP_ADDR
```

Do not duplicate their full explanation here. Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 8. Complete Environment Variables

PREVIOUS_TASK2_DEPENDENCY_FILE
Section: 7. Environment Variables

PREVIOUS_TASK3_DEPENDENCY_FILE
Section: 7. Environment Variables
```

Source env before running:

```bash
cd backend/services/cart-service
set -a
. ./.env
set +a
```

Common env mistakes for `${TASK_FILE_NAME}`:

- `CART_PRODUCT_BASE_URL` missing: service startup fails.
- `CART_PRODUCT_BASE_URL=localhost:8082`: invalid because scheme is missing. Use `http://localhost:8082`.
- `CART_PRODUCT_BASE_URL=http://localhost:8082/api/v1/products`: wrong because client appends path again.
- Product Service runs inside Docker but env points to host-only `localhost`.
- Duration set as `2` instead of `2s`.

## 8. Docker Setup

Docker setup is reused. `${TASK_FILE_NAME}` does not add a Dockerfile, container, volume, network, or health check.

Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 9. Docker and DevOps Setup
```

Docker impact table:

| Docker area | Status for `${TASK_FILE_NAME}` |
|---|---|
| MongoDB container | Reused |
| Redis container | Reused |
| Product Service container | Required only if you run Product Service in Docker |
| `${SERVICE_NAME}` container | Not added by this task |
| New volume | Not added |
| New network | Not added |
| New exposed port | Not added |

Docker networking reminder:

| Where `${SERVICE_NAME}` runs | Product Service URL example |
|---|---|
| Host machine with `go run` | `CART_PRODUCT_BASE_URL=http://localhost:8082` |
| Docker Compose container | `CART_PRODUCT_BASE_URL=http://product-service:8082` |
| Kubernetes | `CART_PRODUCT_BASE_URL=http://product-service.core.svc.cluster.local:8082` or platform service DNS |

Do not expose Product Service internal/admin routes publicly just for Add Item. Public clients should call API Gateway; service-to-service calls should stay on internal network.

## 9. Local Development Setup

### Step 1: Read previous dependency docs first

Follow these in order:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
PREVIOUS_TASK2_DEPENDENCY_FILE
PREVIOUS_TASK3_DEPENDENCY_FILE
```

Why:

- `PREVIOUS_TASK1_DEPENDENCY_FILE` has base Go, MongoDB, Redis, Docker, env, run commands.
- `PREVIOUS_TASK2_DEPENDENCY_FILE` explains MongoDB vs Redis responsibility.
- `PREVIOUS_TASK3_DEPENDENCY_FILE` explains `cart_db.carts` schema and indexes.
- `${TASK_FILE_NAME}` only adds Product Service Add Item integration details.

### Step 2: Go to service directory

```bash
cd backend/services/cart-service
```

### Step 3: Install dependencies

No new dependency install needed for `${TASK_FILE_NAME}`.

If local modules are not downloaded:

```bash
go mod download
```

### Step 4: Start reused databases/services

Start MongoDB and Redis using previous docs.

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 5.1 MongoDB
Section: 5.2 Redis
```

Apply MongoDB schema/indexes if not already done:

```text
PREVIOUS_TASK3_DEPENDENCY_FILE
Section: 5. Database Setup
Topic: Migration / Bootstrap Options
```

### Step 5: Make Product Service reachable

Required for Add Item:

```bash
curl http://localhost:8082/api/v1/products/prod_123
```

If this fails, Add Item will fail with product-related errors. Since a runnable Product Service module was not found in this repo snapshot, use one of these local approaches:

- Run Product Service from its actual implementation repo.
- Route through API Gateway if it exposes product detail locally.
- Use a local mock server that returns the accepted Product Service JSON shape from section `6.1 Product Service`.

### Step 6: Add task-specific env

In `backend/services/cart-service/.env`:

```env
CART_PRODUCT_BASE_URL=http://localhost:8082
CART_PRODUCT_REQUEST_TIMEOUT=2s
```

Keep existing MongoDB, Redis, CMS, and HTTP env from previous setup.

### Step 7: Source env

```bash
set -a
. ./.env
set +a
```

### Step 8: Start backend service

```bash
go run ./cmd/server
```

### Step 9: Verify health and readiness

```bash
curl http://localhost:8084/healthz
curl http://localhost:8084/readyz
```

### Step 10: Verify `${TASK_FILE_NAME}` API

Logged-in user example:

```bash
curl -X POST http://localhost:8084/api/v1/cart/items \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_123" \
  -d '{"product_id":"prod_123","variant_id":"var_1","quantity":2}'
```

Guest example:

```bash
curl -X POST http://localhost:8084/api/v1/cart/items \
  -H "Content-Type: application/json" \
  -H "X-Guest-Session-ID: guest_sess_123" \
  -d '{"product_id":"prod_123","variant_id":"var_1","quantity":1}'
```

Expected:

- HTTP `200`
- Response has `cart_id`
- Response has one matching item
- `quantity` equals requested quantity or incremented final quantity
- `subtotal`, `total`, and `totals.item_count` are updated

## 10. Running the Project

Full run instructions are reused from:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 10. Project Run Instructions
```

Task-specific startup order:

1. Start MongoDB.
2. Start Redis.
3. Ensure `cart_db.carts` schema/indexes.
4. Start or mock Product Service.
5. Ensure CMS URL config exists for server startup.
6. Source `.env`.
7. Run `${SERVICE_NAME}` HTTP server.
8. Call `POST /api/v1/cart/items`.

### Ports and Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| `${SERVICE_NAME}` HTTP API | `8084` | Add Item route, health, readiness | Reused |
| MongoDB | `27017` | Durable `cart_db.carts` store | Reused |
| Redis | `6379` | Active cart and summary cache | Reused |
| Product Service | `8082` | Product detail, variant, price, stock lookup | Task-specific focus |
| CMS Service | `8089` | Coupon preview config required by server startup | Reused |
| Kafka / RabbitMQ | N/A | Not used by Add Item | Not required |

Change `${SERVICE_NAME}` port:

```env
CART_HTTP_ADDR=:8094
```

Then verify:

```bash
curl http://localhost:8094/healthz
```

Change Product Service location:

```env
CART_PRODUCT_BASE_URL=http://product-service:8082
```

Use this only when container/service DNS resolves in your runtime environment.

## 11. Common Errors & Fixes

Generic setup errors are already documented in previous files. This section only covers `${TASK_FILE_NAME}`-specific errors.

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `CART_PRODUCT_BASE_URL or PRODUCT_SERVICE_BASE_URL is required` | Product Service URL missing | Add `CART_PRODUCT_BASE_URL=http://localhost:8082` | Keep `.env.example` synced with local `.env` |
| `CART_PRODUCT_BASE_URL is invalid` | URL missing scheme or malformed | Use `http://localhost:8082` | Always include `http://` or `https://` |
| `CART_PRODUCT_BASE_URL must use http or https` | Unsupported URL scheme | Use HTTP/HTTPS service URL | Do not use file/internal custom schemes |
| `CART_TEMPORARILY_UNAVAILABLE` | Product Service unavailable or non-2xx response | Start Product Service and verify `GET /api/v1/products/{product_id}` | Health check Product Service before Add Item tests |
| `PRODUCT_NOT_FOUND` | Product Service returned `404` or product identity missing | Use valid `product_id` | Seed Product Service test data first |
| `PRODUCT_NOT_SELLABLE` | Missing `seller_id`/`title` or product status not sellable | Return `seller_id`, `title`, and status `published`/`active`/`sellable` | Keep product publishing workflow clear |
| `VARIANT_NOT_FOUND` | Requested `variant_id` not in Product Service response | Use an existing variant id | Verify Product Service response before API test |
| `VARIANT_NOT_SELLABLE` | Variant unavailable, disabled, negative stock, or bad status | Mark variant available/enabled and status active | Product inventory/status should be consistent |
| `PRODUCT_PRICE_INVALID` | Price amount/currency missing or invalid | Return valid minor-unit amount and 3-letter currency | Never rely on frontend price |
| `INSUFFICIENT_STOCK` | Requested final quantity greater than stock | Lower quantity or update stock | Remember existing quantity + request quantity is checked |
| `QUANTITY_TOO_LARGE` | Request quantity or final line quantity exceeds `10` | Use quantity `1` to `10` and watch existing cart quantity | Follow max per-item rule |
| `CART_OWNER_REQUIRED` | No `X-User-ID` or guest session header | Send `X-User-ID` or `X-Guest-Session-ID` | API Gateway should always pass owner context |
| `INVALID_JSON` | Unknown field or invalid body JSON | Send only `product_id`, `variant_id`, `quantity` | Handler uses strict JSON decoder |
| `CART_VERSION_CONFLICT` | Concurrent Add Item updates hit optimistic version conflict | Retry request | Keep `CART_MAX_SAVE_ATTEMPTS=3` unless there is a measured need |
| Add Item succeeds but Redis key missing | Cache refresh failed after MongoDB save | Check logs and Redis connectivity | Monitor `cart.add_item.cache_*_failed` warnings |

## 12. Security & Best Practices

Task-specific best practices:

- Do not trust client-provided price, title, image, seller, stock, or currency. Current implementation correctly fetches these from Product Service.
- Keep `CART_PRODUCT_BASE_URL` internal. Public browser clients should not control or override this value.
- Use short Product Service timeout like `2s` locally. Production value should be based on real latency SLOs.
- Product Service response should avoid sensitive seller/internal inventory metadata that cart does not need.
- API Gateway should set trusted owner headers such as `X-User-ID`; public clients should not be allowed to spoof them directly.
- Keep Product Service and `${SERVICE_NAME}` on private network in Docker/Kubernetes.
- Monitor Product Service errors separately from MongoDB errors. User sees cart unavailable, but root cause differs.
- Do not reserve inventory during Add Item. Stock validation is okay; final reservation belongs to checkout/order flow.
- Keep MongoDB unique active cart indexes enabled to prevent duplicate active carts under concurrent Add Item requests.
- Do not expose `POST /internal/v1/cart/schema/bootstrap` publicly.

## 13. Missing or Misconfigured Things

| Finding | Impact | Suggested fix |
|---|---|---|
| `backend/services/product-service` has no runnable Go module in this repo snapshot | New developers cannot perform full Add Item end-to-end locally unless Product Service is provided elsewhere | Add runnable Product Service implementation, provide a mock service, or document external repo/start command |
| Product Service health endpoint is not configured in `${SERVICE_NAME}` | Startup validates URL format but does not ping Product Service | Add readiness check or smoke test in local onboarding if desired |
| Current server requires CMS config even when testing only Add Item | Beginner may be confused why coupon service URL is needed | Keep CMS env in `.env` or make coupon validator optional behind feature flag in future |
| No Docker Compose file was found for all local services | Beginners must manually run MongoDB, Redis, Product Service, CMS, and `${SERVICE_NAME}` | Add a local compose file with MongoDB, Redis, product mock, CMS mock, and service env wiring |
| Product Service response contract is tolerant but not formally versioned in code | Future Product API changes can break Add Item at runtime | Add contract tests between Product Service and `${SERVICE_NAME}` |
| Owner headers are accepted directly by service | Direct public exposure could allow spoofing | Put service behind API Gateway or internal network; validate auth context before forwarding |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `4. Go Dependency System` | Same Go module, Go version, and dependency workflow |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `5.1 MongoDB` | Same MongoDB installation, Docker setup, URI, and verification |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `5.2 Redis` | Same Redis installation, Docker setup, and verification |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `8. Complete Environment Variables` | Base `.env` already documented |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `9. Docker and DevOps Setup` | No new Docker setup added |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `10. Project Run Instructions` | Base run flow reused |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | `6.2 Redis` | Redis key patterns and TTL behavior reused |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | `7. Environment Variables` | Cache and DB env variable behavior reused |
| `PREVIOUS_TASK3_DEPENDENCY_FILE` | `5. Database Setup` | `cart_db.carts` schema and indexes reused |
| `PREVIOUS_TASK3_DEPENDENCY_FILE` | `9. Local Development Setup` | Schema/bootstrap onboarding reused |
| `PREVIOUS_TASK3_DEPENDENCY_FILE` | `11. Common Errors & Fixes` | Generic schema and MongoDB errors reused |

## 15. Final Checklist

- [ ] Previous dependency documentation checked.
- [ ] No duplicate Go/MongoDB/Redis/Docker installation docs added.
- [ ] `CART_PRODUCT_BASE_URL` configured.
- [ ] `CART_PRODUCT_REQUEST_TIMEOUT` configured or default accepted.
- [ ] Product Service reachable at `GET /api/v1/products/{product_id}`.
- [ ] Product response includes product id, seller id, title, sellable status, variants, price, currency, and stock.
- [ ] MongoDB and Redis running from previous setup.
- [ ] `cart_db.carts` schema/indexes applied from previous setup.
- [ ] `.env` sourced before `go run`.
- [ ] `${SERVICE_NAME}` started on expected `CART_HTTP_ADDR`.
- [ ] `POST /api/v1/cart/items` verified for logged-in user.
- [ ] `POST /api/v1/cart/items` verified for guest session if needed.
- [ ] Product/variant/stock/price error cases checked.
- [ ] Redis active cart and summary cache refresh checked through logs or keys.
- [ ] No real secrets committed.
- [ ] Product Service and internal schema bootstrap endpoints kept off public internet.
