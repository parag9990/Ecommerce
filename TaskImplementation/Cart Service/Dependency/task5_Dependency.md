# Project Dependency & Setup Guide

```text
SERVICE_NAME = "Cart Service"
TASK_FILE_NAME = "task5.md"
INPUT_FILE_PATH = TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}
OUTPUT_FILE_NAME = ${TASK_FILE_NAME without ".md"}_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}

PREVIOUS_TASK1_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
PREVIOUS_TASK2_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
PREVIOUS_TASK3_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task3_Dependency.md
PREVIOUS_TASK4_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task4_Dependency.md
```

## 1. Project Overview

Ye document `${SERVICE_NAME}` ke `${TASK_FILE_NAME}` ke liye dependency, setup, environment, database, Docker, DevOps, aur troubleshooting guide hai.

`${TASK_FILE_NAME}` ka focus Remove Item flow hai:

- Shopper `DELETE /api/v1/cart/items/{item_id}` call karta hai.
- `${SERVICE_NAME}` current owner ka active cart MongoDB se load karta hai.
- Matching `item_id` embedded `items[]` array se remove hota hai.
- Quantity `0` ya negative items defensive cleanup me remove hote hain.
- Coupon preview stale ho sakta hai, isliye mutation ke baad clear hota hai.
- Totals recalculate hote hain.
- MongoDB me active cart optimistic version guard ke saath save hota hai.
- Redis active cart and summary cache refresh hota hai.

Important scope:

- `${INPUT_FILE_PATH}` modify nahi kiya gaya.
- `${OUTPUT_FILE_PATH}` new dependency/setup guide hai.
- No new Go library, database engine, queue, Dockerfile, Docker container, volume, network, or service port was introduced by `${TASK_FILE_NAME}`.
- The main task-specific setup impact is Remove Item API verification and correct existing cart mutation configuration.
- Product Service is not called by Remove Item, but current server config still requires Product Service and CMS base URLs at startup.

### Files Inspected

| File | Why inspected |
|---|---|
| `INPUT_FILE_PATH` | `${TASK_FILE_NAME}` scope: remove item, zero quantity cleanup, totals recalculation |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | Base setup: Go, MongoDB, Redis, Docker, `.env`, run commands, troubleshooting |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | MongoDB source of truth plus Redis cache responsibility |
| `PREVIOUS_TASK3_DEPENDENCY_FILE` | `cart_db.carts` schema, indexes, validator, migration/bootstrap |
| `PREVIOUS_TASK4_DEPENDENCY_FILE` | Add Item setup, Product Service dependency, cache refresh, version retry |
| `backend/services/cart-service/go.mod` | Confirm no new Go dependency for Remove Item |
| `backend/services/cart-service/.env.example` | Confirm actual env names and defaults |
| `backend/services/cart-service/internal/config/config.go` | Confirm validation, defaults, and required startup config |
| `backend/services/cart-service/cmd/server/main.go` | Confirm startup dependency checks and usecase wiring |
| `backend/services/cart-service/internal/usecase/remove_item.go` | Confirm Remove Item runtime behavior |
| `backend/services/cart-service/internal/usecase/remove_item_test.go` | Confirm tested behavior and error cases |
| `backend/services/cart-service/internal/domain/cart_mutation.go` | Confirm domain remove, cleanup, coupon clear, totals logic |
| `backend/services/cart-service/internal/repository/mongo_cart_repository.go` | Confirm active cart lookup and `SaveWithVersion` update |
| `backend/services/cart-service/internal/repository/redis_cart_cache.go` | Confirm Redis key patterns and cache refresh |
| `backend/services/cart-service/internal/transport/http/handler.go` | Confirm REST route, headers, and error mapping |

## 2. Tech Stack

Most technologies are reused from previous dependency files. Full beginner installation steps are intentionally not duplicated.

| Technology | Required? | Status for `${TASK_FILE_NAME}` | What beginner should know | Setup reference |
|---|---:|---|---|---|
| Go `1.26.3` | Yes | Reused | Backend service Go me implemented hai. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4. Go Dependency System` |
| Go modules | Yes | Reused | `go.mod` and `go.sum` dependencies manage karte hain. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4. Go Dependency System` |
| Standard `net/http` | Yes | Reused | Current Remove Item route HTTP API se expose hota hai. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `3. High-Level Tech Stack` |
| MongoDB | Yes | Reused | Active cart durable source of truth hai. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `5.1 MongoDB` |
| MongoDB schema/indexes | Yes | Reused | Remove Item same `cart_db.carts` schema and active cart indexes use karta hai. | `PREVIOUS_TASK3_DEPENDENCY_FILE`, section `5. Database Setup` |
| Redis | Yes | Reused | Remove Item ke baad active cart and summary cache refresh hota hai. | `PREVIOUS_TASK2_DEPENDENCY_FILE`, section `6.2 Redis` |
| Product Service HTTP API | Config required at startup | Reused, not called by Remove Item | Add Item ke liye required hai; Remove Item me product lookup nahi hota. | `PREVIOUS_TASK4_DEPENDENCY_FILE`, section `6.1 Product Service` |
| CMS Service HTTP API | Config required at startup | Reused, not called by Remove Item | Coupon preview flow ke liye required hai; Remove Item only stale coupon preview clear karta hai. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `6.2 CMS Service` |
| Docker | Optional but recommended | Reused | Local MongoDB and Redis run karne ke liye useful hai. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `9. Docker and DevOps Setup` |

Simple Hinglish:

- MongoDB cart ka original data rakhta hai.
- Redis fast copy/cache rakhta hai.
- Remove Item ko Product Service se price/stock check karne ki zarurat nahi hoti.
- Remove Item direct global delete nahi karta. Pehle current owner ka active cart load hota hai, phir us cart ke andar item remove hota hai.

## 3. Required Software

Base software setup already documented hai. Yahan sirf incremental view diya gaya hai.

| Software / Service | Required? | Status | Notes |
|---|---:|---|---|
| Git | Yes | Reused | Clone/setup steps previous docs me hain. |
| Go `1.26.3` | Yes | Reused | Service build/run/test ke liye. |
| MongoDB | Yes | Reused | Active cart load/save ke liye mandatory. |
| Redis | Yes | Reused | Startup ping and cache refresh ke liye mandatory in current service. |
| Product Service | Not used by Remove Item | Startup URL config required | Server config validates Product URL even if Remove Item does not call it. |
| CMS Service | Not used by Remove Item | Startup URL config required | Server config validates CMS URL even if Remove Item only clears stale coupon fields. |
| `mongosh` | Recommended | Reused | Remove verification me active cart inspect/seed karne ke liye useful. |
| `redis-cli` | Recommended | Reused | Cache key verify karne ke liye useful. |
| `curl` | Recommended | Reused | API health and Remove Item verify karne ke liye. |
| Docker / Docker Compose | Optional | Reused | MongoDB/Redis local setup easy banata hai. |

## 4. Dependency Management

This is a Go project. `${TASK_FILE_NAME}` does not add any new Go module dependency.

Current direct dependencies are unchanged:

| Dependency | Version | Status | Why used |
|---|---:|---|---|
| `go.mongodb.org/mongo-driver` | `v1.17.9` | Reused | MongoDB active cart lookup and save |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Reused | Redis active cart and summary cache refresh |

No `go get` is needed for `${TASK_FILE_NAME}`.

Important beginner note:

- `${TASK_FILE_NAME}` contains future gRPC/protobuf examples, but current implemented runtime is standard HTTP with `net/http`.
- Do not add `google.golang.org/grpc` only for this task unless a separate gRPC implementation task actually adds gRPC source files.
- Do not run unnecessary `go get`; it will dirty `go.mod` and confuse future setup.

Refer for Go setup and commands:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 4. Go Dependency System
Section: 4.4 Dependency Commands
```

Useful check from service folder:

```bash
cd backend/services/cart-service
go test ./...
```

## 5. Database Setup

MongoDB setup is reused. No new database, collection, migration, or index was added by `${TASK_FILE_NAME}`.

Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 5.1 MongoDB

PREVIOUS_TASK3_DEPENDENCY_FILE
Section: 5. Database Setup
```

### Database Usage for `${TASK_FILE_NAME}`

| Topic | Value |
|---|---|
| Database | `cart_db` |
| Collection | `carts` |
| Source of truth | MongoDB |
| Active cart lookup | `FindActiveByOwner` |
| Save method | `SaveWithVersion` |
| New migration needed | No |
| Schema/index setup | Same as `PREVIOUS_TASK3_DEPENDENCY_FILE` |

Remove Item database behavior:

- Active cart must already exist. Remove Item does not create a new cart.
- Cart lookup is scoped by current owner, either `user_id` or `guest_session_id`.
- Matching `item_id` is removed from embedded `items[]`.
- Missing `item_id` is idempotent: current cart returns without a MongoDB save if no cleanup is needed.
- Existing invalid `quantity <= 0` items are cleaned and saved.
- Last item remove keeps an active empty cart with zero totals.
- `coupon_code` and `coupon_preview` are cleared when cart content mutates.
- `updated_at`, `expires_at`, totals, items, and version are updated on mutation.
- MongoDB update uses `_id`, `status: active`, and `version` as guard.

### MongoDB Update Shape

Current repository saves active cart changes with this effective behavior:

```text
Filter:
  _id = cart.ID
  status = active
  version = expectedVersion

Update:
  set items, coupon_code, coupon_preview, totals, updated_at, expires_at
  increment version by 1
```

Beginner explanation:

Version guard ka matlab hai agar same cart ko do requests ek saath update kar rahe hain, stale request blindly overwrite nahi karegi. Conflict aayega, usecase reload/retry karega up to `CART_MAX_SAVE_ATTEMPTS`.

### Task-Specific MongoDB Verification

After Remove Item, active cart inspect karo:

```bash
mongosh "mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin" \
  --quiet \
  --eval "db.getSiblingDB('cart_db').carts.findOne({ user_id: 'user_123', status: 'active' }, { items: 1, totals: 1, coupon_code: 1, coupon_preview: 1, version: 1, expires_at: 1 })"
```

Expected:

- Removed `item_id` no longer appears in `items`.
- Empty cart has `items: []`.
- `totals.item_count` and `totals.unique_item_count` match remaining items.
- `coupon_code` and `coupon_preview` are `null` when a real mutation happened.
- `version` increments when MongoDB save happened.

## 6. Redis / Queue / External Services

### 6.1 Redis

Redis setup is reused. `${TASK_FILE_NAME}` does not introduce new Redis key patterns.

Refer:

```text
PREVIOUS_TASK2_DEPENDENCY_FILE
Section: 6.2 Redis
```

Remove Item cache behavior:

| Key pattern | Status | Purpose |
|---|---|---|
| `cart:active:user:{user_id}` | Reused | Updated full cart cache for logged-in user |
| `cart:active:guest:{guest_session_id}` | Reused | Updated full cart cache for guest |
| `cart:summary:user:{user_id}` | Reused | Header/badge summary for logged-in user |
| `cart:summary:guest:{guest_session_id}` | Reused | Header/badge summary for guest |

Important:

- MongoDB save failure fails Remove Item.
- Redis cache refresh failure is logged as warning, but updated cart can still be returned.
- Idempotent missing-item delete also refreshes cache from the current MongoDB cart.
- If active cart is expired during Remove Item, current implementation marks it expired and deletes active/summary cache.

Verify Redis after Remove Item:

```bash
redis-cli GET cart:summary:user:user_123
redis-cli TTL cart:active:user:user_123
```

For guest cart:

```bash
redis-cli GET cart:summary:guest:guest_sess_123
redis-cli TTL cart:active:guest:guest_sess_123
```

### 6.2 Product Service

Product Service is not called by Remove Item.

Still, current server creates Product Service client during startup, so URL config must exist:

```env
CART_PRODUCT_BASE_URL=http://localhost:8082
CART_PRODUCT_REQUEST_TIMEOUT=2s
```

Refer for full Product Service explanation:

```text
PREVIOUS_TASK4_DEPENDENCY_FILE
Section: 6.1 Product Service
```

Beginner note:

Remove Item does not need product price, stock, title, image, or variant validation. Item already exists in the cart snapshot, so deleting it only needs current owner cart access.

### 6.3 CMS Service

CMS Service is not called by Remove Item.

Still, current server creates Coupon Validator config during startup, so CMS URL config must exist:

```env
CART_CMS_BASE_URL=http://localhost:8089
CART_CMS_VALIDATE_COUPON_PATH=/internal/v1/coupons/validate
```

Remove Item task-specific CMS behavior:

- If cart content mutates, stale `coupon_code` and `coupon_preview` are cleared.
- Discount becomes `0` after totals recalculate.
- Coupon revalidation belongs to coupon preview flow, not this task.

Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 6.2 CMS Service
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
| Kubernetes | No local setup needed | Only future deployment concern |

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

### New or Changed Variables

`${TASK_FILE_NAME}` adds no new environment variable and changes no existing variable.

### Task-Specific Existing Variables to Verify

These values already exist in `.env.example`, but Remove Item depends on them:

```env
CART_DEFAULT_CURRENCY=INR
CART_MAX_SAVE_ATTEMPTS=3
CART_USER_EXPIRY_TTL=2160h
CART_GUEST_EXPIRY_TTL=720h
CART_ACTIVE_CACHE_TTL=15m
CART_GUEST_ACTIVE_CACHE_TTL=30m
CART_SUMMARY_CACHE_TTL=5m
```

| Variable | Required? | Purpose for `${TASK_FILE_NAME}` | Example | Security / setup note |
|---|---:|---|---|---|
| `CART_DEFAULT_CURRENCY` | Yes, defaulted | Empty cart totals currency after last item remove | `INR` | Must be 3 uppercase letters |
| `CART_MAX_SAVE_ATTEMPTS` | Yes, defaulted | Version conflict retry count for remove mutations | `3` | High value blindly set mat karo; concurrency issue hide ho sakta hai |
| `CART_USER_EXPIRY_TTL` | Yes, defaulted | User cart expiry extension after mutation | `2160h` | Duration format must include unit |
| `CART_GUEST_EXPIRY_TTL` | Yes, defaulted | Guest cart expiry extension after mutation | `720h` | Duration format must include unit |
| `CART_ACTIVE_CACHE_TTL` | Yes, defaulted | User active cart cache TTL | `15m` | Not secret |
| `CART_GUEST_ACTIVE_CACHE_TTL` | Yes, defaulted | Guest active cart cache TTL | `30m` | Not secret |
| `CART_SUMMARY_CACHE_TTL` | Yes, defaulted | Cart summary cache TTL | `5m` | Not secret |

Existing startup/config env still relevant:

```text
CART_HTTP_ADDR
CART_MONGO_URI
CART_MONGO_DATABASE
CART_REDIS_ADDR
CART_PRODUCT_BASE_URL
CART_CMS_BASE_URL
CART_CMS_VALIDATE_COUPON_PATH
```

Do not duplicate their full explanation here. Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 8. Complete Environment Variables

PREVIOUS_TASK2_DEPENDENCY_FILE
Section: 7. Environment Variables

PREVIOUS_TASK3_DEPENDENCY_FILE
Section: 7. Environment Variables

PREVIOUS_TASK4_DEPENDENCY_FILE
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

- `.env` exists but not sourced before `go run`.
- `CART_DEFAULT_CURRENCY=inr` lowercase. Current validation requires uppercase 3-letter currency.
- `CART_MAX_SAVE_ATTEMPTS=0`. Current validation requires greater than zero.
- `CART_USER_EXPIRY_TTL=2160` without `h`. Duration parsing falls back silently in config helper, so developer may think value changed when it did not.
- Product/CMS URL missing. Remove Item does not use them, but server startup still validates them.

## 8. Docker Setup

Docker setup is reused. `${TASK_FILE_NAME}` does not add a Dockerfile, container, volume, network, exposed port, restart policy, or health check.

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
| Product Service container | Not used by Remove Item, but URL config still required |
| CMS Service container | Not used by Remove Item, but URL config still required |
| `${SERVICE_NAME}` container | Not added by this task |
| New volume | Not added |
| New network | Not added |
| New exposed port | Not added |

Docker networking reminder:

| Where `${SERVICE_NAME}` runs | MongoDB URI host | Redis address | Product/CMS URL style |
|---|---|---|---|
| Host machine with `go run` | `localhost:27017` | `localhost:6379` | `http://localhost:8082`, `http://localhost:8089` |
| Docker Compose container | `mongo:27017` | `redis:6379` | `http://product-service:8082`, `http://cms-service:8089` |

Do not expose internal schema bootstrap or direct service routes publicly. Public clients should go through API Gateway/auth layer.

## 9. Local Development Setup

### Step 1: Read previous dependency docs first

Follow these in order:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
PREVIOUS_TASK2_DEPENDENCY_FILE
PREVIOUS_TASK3_DEPENDENCY_FILE
PREVIOUS_TASK4_DEPENDENCY_FILE
```

Why:

- `PREVIOUS_TASK1_DEPENDENCY_FILE` has base Go, MongoDB, Redis, Docker, env, run commands.
- `PREVIOUS_TASK2_DEPENDENCY_FILE` explains MongoDB vs Redis responsibility.
- `PREVIOUS_TASK3_DEPENDENCY_FILE` explains `cart_db.carts` schema and indexes.
- `PREVIOUS_TASK4_DEPENDENCY_FILE` explains how to create an active cart through Add Item.

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

### Step 5: Ensure startup env exists

Keep existing MongoDB, Redis, Product, CMS, HTTP, TTL, and cache env from previous setup.

Minimum task-specific values to check:

```env
CART_DEFAULT_CURRENCY=INR
CART_MAX_SAVE_ATTEMPTS=3
CART_USER_EXPIRY_TTL=2160h
CART_GUEST_EXPIRY_TTL=720h
CART_ACTIVE_CACHE_TTL=15m
CART_GUEST_ACTIVE_CACHE_TTL=30m
CART_SUMMARY_CACHE_TTL=5m
```

### Step 6: Source env

```bash
set -a
. ./.env
set +a
```

### Step 7: Start backend service

```bash
go run ./cmd/server
```

### Step 8: Verify health and readiness

```bash
curl http://localhost:8084/healthz
curl http://localhost:8084/readyz
```

Notes:

- Startup pings MongoDB and Redis.
- `/readyz` checks MongoDB through schema manager.
- Product/CMS URLs are validated at startup, but Product/CMS are not pinged by `/readyz`.

### Step 9: Prepare an active cart

Recommended path:

```text
PREVIOUS_TASK4_DEPENDENCY_FILE
Section: 9. Local Development Setup
Topic: Add Item API verification
```

Use Add Item first, then copy the returned `items[].item_id` for Remove Item.

If Product Service is unavailable and you only need local Remove Item testing, you can seed a valid active cart manually in MongoDB. Manual seeding is only for local testing, not production:

```bash
mongosh "mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin" --quiet --eval '
const now = new Date();
const expires = new Date(now.getTime() + 90 * 24 * 60 * 60 * 1000);
db.getSiblingDB("cart_db").carts.updateOne(
  { _id: "cart_user_123" },
  {
    $set: {
      _id: "cart_user_123",
      user_id: "user_123",
      guest_session_id: null,
      status: "active",
      items: [{
        item_id: "item_1",
        product_id: "prod_123",
        variant_id: "var_1",
        seller_id: "seller_123",
        title_snapshot: "Local Test Product",
        variant_snapshot: { size: "M" },
        unit_price: { amount: NumberLong(1000), currency: "INR" },
        quantity: 2,
        line_subtotal: { amount: NumberLong(2000), currency: "INR" },
        price_snapshot_at: now,
        added_at: now,
        updated_at: now
      }],
      coupon_code: null,
      coupon_preview: null,
      totals: {
        subtotal: { amount: NumberLong(2000), currency: "INR" },
        discount: { amount: NumberLong(0), currency: "INR" },
        total: { amount: NumberLong(2000), currency: "INR" },
        currency: "INR",
        item_count: 2,
        unique_item_count: 1
      },
      version: NumberLong(1),
      created_at: now,
      updated_at: now,
      expires_at: expires
    }
  },
  { upsert: true }
)'
```

### Step 10: Verify `${TASK_FILE_NAME}` API

Logged-in user example:

```bash
curl -X DELETE http://localhost:8084/api/v1/cart/items/item_1 \
  -H "X-User-ID: user_123"
```

Guest example:

```bash
curl -X DELETE http://localhost:8084/api/v1/cart/items/item_1 \
  -H "X-Guest-Session-ID: guest_sess_123"
```

Expected:

- HTTP `200`.
- Response has same `cart_id`.
- Removed item is absent.
- Empty cart stays active with `items: []`.
- Totals become zero if last item was removed.
- Redis active cart and summary cache are refreshed best-effort.

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
4. Keep Product/CMS URL env configured for startup.
5. Source `.env`.
6. Run `${SERVICE_NAME}` HTTP server.
7. Create or seed an active cart.
8. Call `DELETE /api/v1/cart/items/{item_id}`.
9. Verify MongoDB document and Redis keys.

### Ports and Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| `${SERVICE_NAME}` HTTP API | `8084` | Remove Item route, health, readiness | Reused |
| MongoDB | `27017` | Durable `cart_db.carts` store | Reused |
| Redis | `6379` | Active cart and summary cache | Reused |
| Product Service | `8082` | Not called by Remove Item, URL config required | Reused |
| CMS Service | `8089` | Not called by Remove Item, URL config required | Reused |
| Kafka / RabbitMQ | N/A | Not used by Remove Item | Not required |

Change `${SERVICE_NAME}` port:

```env
CART_HTTP_ADDR=:8094
```

Then verify:

```bash
curl http://localhost:8094/healthz
```

Route details:

| Route | Method | Purpose |
|---|---|---|
| `/api/v1/cart/items/{item_id}` | `DELETE` | Remove cart item |
| `/healthz` | `GET` | Process health |
| `/readyz` | `GET` | MongoDB readiness |

Owner headers:

| Header | Purpose |
|---|---|
| `X-User-ID` | Logged-in user owner |
| `X-Guest-Session-ID` | Guest cart owner |
| `X-Guest-Session-Id` | Accepted guest header variant |
| `X-Session-ID` | Accepted guest header fallback |

## 11. Common Errors & Fixes

Generic setup errors are already documented in previous files. This section only covers `${TASK_FILE_NAME}`-specific or Remove Item-specific errors.

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `405 Method Not Allowed` | Route called with wrong HTTP method | Use `DELETE /api/v1/cart/items/{item_id}` | Keep frontend/API Gateway route config aligned |
| `ITEM_ID_REQUIRED` | `{item_id}` path segment missing or blank | Pass actual item id from cart response | Do not call `/api/v1/cart/items/` |
| `NOT_FOUND` for remove route | Extra slash/path after item id | Use exactly one path segment after `/items/` | URL-build item route carefully |
| `CART_OWNER_REQUIRED` | No user or guest owner header | Send `X-User-ID` or guest session header | API Gateway should inject trusted owner context |
| `CART_NOT_FOUND` | Active cart does not exist for this owner | Create cart through Add Item or seed local test cart | Remove Item does not create carts |
| `CART_NOT_ACTIVE` | Cart status is merged, checked_out, expired, or abandoned | Use current active cart only | Do not mutate old carts |
| `CART_EXPIRED` | Active cart `expires_at` is in the past | Create a fresh cart or update test seed expiry | Keep TTL config and seeded dates valid |
| `CART_TOTALS_INVALID` | Corrupt cart data, mixed currency, or invalid totals | Inspect MongoDB cart document and fix test data | Do not manually write invalid cart documents |
| `CART_VERSION_CONFLICT` | Concurrent cart mutations exceeded retry attempts | Retry request or investigate parallel updates | Keep `CART_MAX_SAVE_ATTEMPTS=3` unless measured need says otherwise |
| Remove succeeds but Redis key stale/missing | Cache refresh failed after Mongo save | Check logs for `cart.remove_item.cache_*_failed`, verify Redis | Monitor Redis and keep it reachable |
| Server fails before testing Remove Item due Product/CMS env | Current config validates Product/CMS URLs at startup | Add valid `CART_PRODUCT_BASE_URL` and `CART_CMS_BASE_URL` | Keep `.env.example` copied completely |
| Empty cart still has discount | Old/corrupt document or mutation did not save | Check MongoDB `coupon_preview`, `coupon_code`, and totals | Let service mutate cart instead of direct DB writes |

## 12. Security & Best Practices

Task-specific best practices:

- Do not delete by global `item_id` across all carts. Always scope by current owner active cart.
- API Gateway should set trusted `X-User-ID` or guest session headers. Public clients should not be able to spoof these directly.
- Keep `${SERVICE_NAME}` behind API Gateway or private network.
- Keep MongoDB unique active cart indexes enabled to prevent duplicate active carts.
- Keep Redis as cache only. MongoDB remains source of truth.
- Keep Remove Item idempotent. Retry/double-click should not break user experience.
- Clear stale coupon preview when cart content changes. Old discounts can become invalid after remove.
- Use integer minor units for money. Do not use floats for totals.
- Keep `CART_DEFAULT_CURRENCY` uppercase and consistent with Product Service currency.
- Keep `CART_MAX_SAVE_ATTEMPTS` small and observable. High values can hide real contention.
- Do not expose `POST /internal/v1/cart/schema/bootstrap` publicly.
- Do not commit real MongoDB/Redis credentials.

## 13. Missing or Misconfigured Things

| Finding | Impact | Suggested fix |
|---|---|---|
| Product/CMS URLs are required by startup even when only Remove Item is tested | Beginner may think Product/CMS are used by Remove Item | Keep URLs configured; optionally make clients lazy/feature-specific in future |
| Direct service trusts owner headers | If exposed publicly, a user could spoof owner headers | Expose through API Gateway/auth middleware only |
| No dedicated `${SERVICE_NAME}` Dockerfile found in inspected setup | Full service container build flow is incomplete | Add Dockerfile in future DevOps task |
| No complete local compose stack found for all services | Beginners must manually run dependencies and service | Add local compose with MongoDB, Redis, service, and optional mocks |
| `.env` is not auto-loaded | Service can fail despite `.env` file existing | Source `.env` before `go run`, or add a documented dev runner |
| `/readyz` checks MongoDB readiness, while Redis is checked at startup | Redis can fail after startup but `/readyz` may still pass | Add Redis readiness check if production requires it |
| Manual MongoDB seed can bypass domain validation | Bad test data can cause totals/currency errors | Use Add Item when possible; manual seed only for local Remove Item smoke tests |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `4. Go Dependency System` | Same Go module, Go version, and dependency workflow |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `5.1 MongoDB` | Same MongoDB installation, Docker setup, URI, and verification |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `5.2 Redis` | Same Redis installation, Docker setup, and verification |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `6. External Services` | Same Product Service and CMS config requirement at server startup |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `8. Complete Environment Variables` | Full `.env` location and variable explanation already documented |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `9. Docker and DevOps Setup` | No new Docker setup added |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `10. Project Run Instructions` | Base run flow reused |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `12. Common Errors and Fixes` | Generic Go/MongoDB/Redis/Docker/env troubleshooting reused |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | `6.2 Redis` | Redis key patterns and TTL behavior reused |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | `7. Environment Variables` | Cache and DB env variable behavior reused |
| `PREVIOUS_TASK3_DEPENDENCY_FILE` | `5. Database Setup` | `cart_db.carts` schema, indexes, and migration/bootstrap reused |
| `PREVIOUS_TASK3_DEPENDENCY_FILE` | `9. Local Development Setup` | Schema/bootstrap onboarding reused |
| `PREVIOUS_TASK4_DEPENDENCY_FILE` | `6.1 Product Service` | Add Item creates active cart and explains Product Service setup |
| `PREVIOUS_TASK4_DEPENDENCY_FILE` | `9. Local Development Setup` | Active cart creation via Add Item reused before Remove Item verification |
| `PREVIOUS_TASK4_DEPENDENCY_FILE` | `11. Common Errors & Fixes` | Product-related Add Item setup errors reused when preparing active cart |

## 15. Final Checklist

- [ ] Previous dependency documentation checked.
- [ ] No duplicate Go/MongoDB/Redis/Docker installation docs added.
- [ ] No new Go dependency added for `${TASK_FILE_NAME}`.
- [ ] MongoDB and Redis running from previous setup.
- [ ] `cart_db.carts` schema/indexes applied from previous setup.
- [ ] `.env` created in `backend/services/cart-service/.env`.
- [ ] `.env` sourced before `go run`.
- [ ] Product/CMS URLs configured for server startup, even though Remove Item does not call them.
- [ ] `CART_DEFAULT_CURRENCY` is uppercase, for example `INR`.
- [ ] `CART_MAX_SAVE_ATTEMPTS` is greater than zero.
- [ ] Cache TTL values are positive durations.
- [ ] `${SERVICE_NAME}` started on expected `CART_HTTP_ADDR`.
- [ ] `/healthz` returns `ok`.
- [ ] `/readyz` returns `ready`.
- [ ] Active cart exists for logged-in user or guest session.
- [ ] `DELETE /api/v1/cart/items/{item_id}` verified for logged-in user.
- [ ] `DELETE /api/v1/cart/items/{item_id}` verified for guest session if needed.
- [ ] Removed item absent from MongoDB cart document.
- [ ] Empty cart totals verified after last item remove.
- [ ] Stale coupon fields cleared after mutation.
- [ ] Redis active cart and summary cache refresh checked through logs or keys.
- [ ] Idempotent missing-item delete tested if needed.
- [ ] Expired cart behavior tested if needed.
- [ ] No real secrets committed.
- [ ] Internal routes and direct service headers protected behind API Gateway/private network.
