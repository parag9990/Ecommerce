# Project Dependency & Setup Guide

## Variable Values

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Wishlist Service` |
| `TASK_FILE_NAME` | `task5.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task5_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

Note: Ye guide `INPUT_FILE_PATH` ko analyze karke banaya gaya hai. Original `TASK_FILE_NAME` modify nahi kiya gaya. Is file ka focus sirf Task 5 ke setup, dependency, environment, DevOps, verification, and beginner onboarding points par hai.

---

## 1. Project Overview

`TASK_FILE_NAME` ka main scope:

```text
POST /api/v1/wishlist/items/{product_id}/move-to-cart
-> authenticated buyer validate
-> wishlist item find
-> Cart Service AddItem call
-> cart add success ke baad wishlist item remove
-> updated Cart response return
```

Simple Hinglish me: buyer wishlist page se saved product ko direct cart me move kar sakta hai. Important safety rule hai: pehle Cart Service me item add hoga. Agar Cart Service fail hota hai, wishlist item remove nahi hoga.

### Runtime code paths checked

| Runtime file | Task 5 relevance |
|---|---|
| `backend/services/wishlist-service/internal/usecase/wishlist_service.go` | `MoveToCart`, Cart client interface, idempotency key, cart-first cleanup flow. |
| `backend/services/wishlist-service/internal/clients/cart_http_client.go` | HTTP Cart Service client, request headers, error mapping, response decode. |
| `backend/services/wishlist-service/internal/config/config.go` | `WISHLIST_CART_SERVICE_BASE_URL`, fallback, timeout validation. |
| `backend/services/wishlist-service/internal/transport/http/wishlist_handler.go` | `POST /move-to-cart` route, buyer auth headers, `X-Idempotency-Key`. |
| `backend/services/wishlist-service/cmd/server/main.go` | Cart client wiring into Wishlist usecase. |
| `backend/services/wishlist-service/.env` | Current local env exists, but Cart Service variables are not explicitly present. |
| `api/master-api.json` | `wishlist.move_to_cart` and `CartService.AddItem` contract. |

### Important Reuse Note

Base setup already documented hai. Is file me Go install, MongoDB install, Docker basics, full `.env`, Product Service add/remove setup, and generic troubleshooting repeat nahi kiya gaya.

Read first:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
```

---

## 2. Tech Stack

Most technology explanation already covered hai. Task 5 ka real new setup focus Cart Service HTTP dependency and idempotent move-to-cart testing hai.

| Technology / Service | Task 5 Status | Required? | Setup Documentation |
|---|---|---:|---|
| Go | Reused | Yes | `task1_Dependency.md` -> `4. Dependency Management` |
| Go modules | Reused | Yes | `task1_Dependency.md` -> `4. Dependency Management` |
| `net/http` | Reused, Task 5 uses it for Cart client | Yes | `task1_Dependency.md` -> `2. Tech Stack` |
| MongoDB | Reused persistence | Yes | `task1_Dependency.md` -> `5. Database Setup`; `task3_Dependency.md` -> `5. Database Setup` |
| MongoDB Go Driver v2 | Reused | Yes | `task1_Dependency.md` -> `4. Dependency Management` |
| Product Service HTTP API | Reused from Task 4 for add-item setup | Required before item exists in wishlist | `task4_Dependency.md` -> `6. Redis / Queue / External Services` |
| Cart Service HTTP API | Task 5 core external dependency | Yes for move-to-cart | Explained below |
| Gateway/auth headers | Reused from Task 4 | Yes | `task4_Dependency.md` -> `2. Tech Stack` and `7. Environment Variables` |
| Kafka-compatible broker | Reused optional runtime feature | No for Task 5 core flow | `task1_Dependency.md` -> `6. Redis / Queue / External Services` |
| Docker | Reused local dependency runner | Recommended | `task1_Dependency.md` -> `8. Docker Setup` |
| Redis | Not introduced by Wishlist Task 5 | No | No new setup needed |
| RabbitMQ/NATS/MinIO/SMTP/Stripe/Twilio | Not introduced | No | No setup needed |

### Task 5 New Technical Concept: Cart Service Client

Cart Service cart ka owner hai. `SERVICE_NAME` direct Cart DB me write nahi karega. Move-to-cart ke time Wishlist Service HTTP call karta hai:

```text
POST {WISHLIST_CART_SERVICE_BASE_URL}/api/v1/cart/items
```

Request body:

```json
{
  "product_id": "prod_123",
  "variant_id": "var_1",
  "quantity": 1
}
```

Beginner meaning: Wishlist Service sirf saved item ko cart add request me convert karta hai. Cart totals, duplicate/merge behavior, stock validation, and final cart shape Cart Service decide karega.

---

## 3. Required Software

No new language runtime, database server, package manager, or queue is introduced by `TASK_FILE_NAME`.

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

Task 5-specific readiness:

| Requirement | Required? | Why |
|---|---:|---|
| Running Cart Service or local Cart API mock | Yes for end-to-end move-to-cart | Wishlist calls `POST /api/v1/cart/items`. |
| MongoDB with `wishlists` collection | Yes | Wishlist item lookup and cleanup happen in MongoDB. |
| Product Service or Product API mock | Required to create wishlist item through add API first | Task 4 add flow validates product before saving item. |
| curl/Postman/HTTP client | Recommended | Manual endpoint verification. |
| API Gateway or trusted headers | Required in real deployment | `X-User-ID` and role headers must come from trusted auth layer. |

Current repo snapshot note:

```text
backend/services/cart-service
```

was not found. Isliye full end-to-end Task 5 testing needs one of these:

- Real Cart Service running from another repo/service.
- A local mock that implements `POST /api/v1/cart/items`.
- A future Cart Service implementation added to this repository.

Unit tests already use fake clients and `httptest`, so code-level Task 5 verification can run without a real Cart Service.

---

## 4. Dependency Management

No new Go module dependency is introduced by `TASK_FILE_NAME`.

Existing module:

```text
backend/services/wishlist-service/go.mod
```

Existing direct dependencies:

| Dependency | Version | Task 5 usage |
|---|---:|---|
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | Existing wishlist lookup and cleanup writes. |
| `github.com/segmentio/kafka-go` | `v0.4.49` | Optional event runtime, not required for move-to-cart core flow. |
| Go standard library `net/http` | Built in | Cart Service HTTP client and Wishlist HTTP handler. |
| Go standard library `encoding/json` | Built in | Cart request/response JSON. |

Reuse Go setup commands from:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Dependency Management`
```

Task 5 beginner warning:

- Do not run `go get` for Cart HTTP client. Current implementation uses Go standard library.
- Do not add Redis/Kafka/RabbitMQ dependency for this task.
- If `go test ./...` fails due dependency download, follow Go module troubleshooting from `task1_Dependency.md`.

Useful verification:

```bash
cd backend/services/wishlist-service
go test ./...
```

---

## 5. Database Setup

Task 5 does not add a new database, collection, index, or migration.

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

### Task 5 Database Behavior

| Operation | MongoDB behavior | Setup impact |
|---|---|---|
| Find wishlist item | `FindByUserID(ctx, userID)` loads user's wishlist document. | Requires `wishlists` collection and `user_id` index readiness. |
| Validate item exists | In-memory `FindItem(productID)` on loaded wishlist. | No new DB setup. |
| Cleanup after Cart success | Existing `RemoveItem` uses `$pull` by `items.product_id`. | No new migration. |
| Cart failure | Wishlist item remains unchanged. | No rollback setup needed. |

Task 5 relevant existing values:

| Item | Value | Status |
|---|---|---|
| Database type | MongoDB | Reused |
| Database name | `wishlist_db` | Reused default |
| Collection | `wishlists` | Reused |
| Default Mongo port | `27017` | Reused |
| Main env var | `WISHLIST_MONGO_URI` | Reused |

### Verify Database Before Move Test

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.wishlists.getIndexes()"
curl http://localhost:8084/readyz
```

Expected:

- `/readyz` returns healthy when MongoDB is reachable.
- `idx_wishlists_items_product_id` exists.
- The user has a wishlist item with both `product_id` and `variant_id` before calling move-to-cart.

Important: Task 5 requires `variant_id` because `CartItemInput` requires `product_id`, `variant_id`, and `quantity`.

---

## 6. Redis / Queue / External Services

### Cart Service

Cart Service is the only Task 5-specific external runtime dependency.

| Item | Value |
|---|---|
| Purpose | Add wishlist item into buyer cart. |
| Required for move-to-cart? | Yes |
| Required for add/remove wishlist? | No |
| Default base URL | `http://localhost:8083` |
| Env var | `WISHLIST_CART_SERVICE_BASE_URL` |
| Fallback env var | `CART_SERVICE_BASE_URL` |
| Timeout env var | `WISHLIST_CART_SERVICE_TIMEOUT` |
| Fallback timeout alias | `WISHLIST_CART_CALL_TIMEOUT_MS` |
| Runtime endpoint | `POST /api/v1/cart/items` |

Expected successful Cart Service response can be plain Cart JSON or wrapped in `{ "data": ... }`:

```json
{
  "data": {
    "cart_id": "cart_123",
    "user_id": "user_123",
    "items": [
      {
        "item_id": "item_1",
        "product_id": "prod_123",
        "variant_id": "var_1",
        "quantity": 1
      }
    ],
    "subtotal": {
      "amount": 299900,
      "currency": "INR"
    },
    "discount": {
      "amount": 0,
      "currency": "INR"
    },
    "total": {
      "amount": 299900,
      "currency": "INR"
    }
  },
  "error": null
}
```

Wishlist Service sends these headers to Cart Service:

| Header | Source | Purpose |
|---|---|---|
| `X-User-ID` | Authenticated buyer header from incoming request | Cart owner. |
| `X-User-Roles` | Hardcoded as `buyer` for downstream call | Cart authorization context. |
| `X-Request-ID` | Incoming request id or generated id | Trace/debug correlation. |
| `X-Idempotency-Key` | Incoming header or generated key | Avoid duplicate cart adds during retry. |

Cart HTTP status mapping:

| Cart response | Wishlist error code | Meaning |
|---:|---|---|
| `400` | `CART_VALIDATION_ERROR` | Cart rejected input. |
| `401` | `CART_UNAUTHENTICATED` | Auth metadata missing/invalid at Cart Service. |
| `403` | `CART_FORBIDDEN` | Buyer role rejected. |
| `404` | `CART_PRODUCT_NOT_FOUND` | Product or variant not found for cart. |
| `409` or `410` | `CART_ITEM_UNAVAILABLE` | Product cannot be added, often stock/status issue. |
| `408`, `429`, or `5xx` | `CART_SERVICE_UNAVAILABLE` | Cart Service unavailable/slow/overloaded. |

### Product Service

Product Service setup is reused from Task 4. It is not called directly by `MoveToCart`, but it is usually needed to create the wishlist item first through `POST /api/v1/wishlist/items`.

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task4_Dependency.md`

Section:
`6. Redis / Queue / External Services`
```

### Kafka / Analytics Events

Kafka setup is reused and optional for Task 5 core flow.

Recommended beginner local setting if Kafka is not running:

```env
WISHLIST_EVENTS_BACKEND=disabled
WISHLIST_ANALYTICS_EVENTS_ENABLED=false
WISHLIST_EVENT_PUBLISHER=disabled
```

### Redis, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, OAuth

No new setup needed for `TASK_FILE_NAME`.

---

## 7. Environment Variables

Full `.env` reference is already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Environment Variables`
```

Task 5 needs only Cart Service env delta. Current code has defaults, but explicit values are beginner-friendly and avoid hidden config confusion.

Create/update local env here:

```text
backend/services/wishlist-service/.env
```

### Task 5 Env Delta Only

This is not a full `.env`. Add these values to the existing service env when testing move-to-cart:

```env
# Cart Service dependency for move-to-cart
WISHLIST_CART_SERVICE_BASE_URL=http://localhost:8083
WISHLIST_CART_SERVICE_TIMEOUT=1s

# Optional fallback names supported by config.go
# CART_SERVICE_BASE_URL=http://localhost:8083
# WISHLIST_CART_CALL_TIMEOUT_MS=1000
```

For Docker Compose or Kubernetes-style service DNS:

```env
WISHLIST_CART_SERVICE_BASE_URL=http://cart-service:8083
WISHLIST_CART_SERVICE_TIMEOUT=1s
```

Use the actual port exposed by your Cart Service. `8083` is the default assumed by Wishlist config and earlier docs.

### Variable Explanation

| Variable | Required? | Example | Task 5 purpose | Security/config note |
|---|---:|---|---|---|
| `WISHLIST_CART_SERVICE_BASE_URL` | Required by config, default exists | `http://localhost:8083` | Base URL used to call `POST /api/v1/cart/items`. | Use internal service URL in prod, not public internet. |
| `CART_SERVICE_BASE_URL` | Optional fallback | `http://localhost:8083` | Used only if service-specific var is missing. | Prefer `WISHLIST_CART_SERVICE_BASE_URL`. |
| `WISHLIST_CART_SERVICE_TIMEOUT` | Optional, default `1s` | `1s` | Max wait for Cart Service call. | Keep positive; too low causes false timeouts. |
| `WISHLIST_CART_CALL_TIMEOUT_MS` | Optional fallback | `1000` | Millisecond alias for older config style. | Avoid setting both timeout vars differently. |

### How the Project Loads Env

Current Go code reads OS environment variables. It does not automatically load `.env`.

Load before starting service:

```bash
cd backend/services/wishlist-service
set -a
source .env
set +a
```

Common Task 5 env mistakes:

| Mistake | Result | Fix |
|---|---|---|
| Cart URL points to wrong port | `CART_SERVICE_UNAVAILABLE` | Set `WISHLIST_CART_SERVICE_BASE_URL` to real Cart Service host/port. |
| Cart URL includes `/api/v1/cart/items` | Runtime calls duplicated path | Use base only, for example `http://localhost:8083`. |
| `.env` updated but not sourced | Old/default values still used | Run `set -a; source .env; set +a` before `go run`. |
| Timeout written as `0` or invalid text | Startup validation fails | Use `1s`, `2s`, or integer milliseconds like `1000`. |
| Docker container uses `localhost` for another container | Connection refused | Use Compose service DNS like `http://cart-service:8083`. |

---

## 8. Docker Setup

No new Dockerfile, Docker Compose service, volume, or network is introduced by `TASK_FILE_NAME`.

Reuse base Docker dependency setup:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`8. Docker Setup`
```

Task 5 Docker-specific notes:

| Docker item | Status | Meaning |
|---|---|---|
| MongoDB container | Reused | Required if MongoDB is not installed locally. |
| Kafka/Redpanda container | Optional, reused | Only needed if events/analytics are enabled. |
| Cart Service container | New runtime need for Task 5 testing, but not provided in this snapshot | Must expose Cart API reachable from Wishlist Service. |
| Service Dockerfile | Not found for Wishlist Service in current snapshot | Local service is run with `go run`. |
| Docker network | No new network supplied | If using containers, put Wishlist, MongoDB, and Cart Service on same Compose network. |

### Host vs Docker URL Examples

If Wishlist Service and Cart Service both run on host:

```env
WISHLIST_CART_SERVICE_BASE_URL=http://localhost:8083
WISHLIST_MONGO_URI=mongodb://localhost:27017
```

If Wishlist Service runs inside Docker Compose and Cart Service is another Compose service:

```env
WISHLIST_CART_SERVICE_BASE_URL=http://cart-service:8083
WISHLIST_MONGO_URI=mongodb://wishlist-mongo:27017
```

Beginner note: `localhost` inside a container means the same container. Dusre container ko call karna hai to Compose service name use karo.

---

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Read these first:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
```

Minimum sections:

- `task1_Dependency.md` -> `3. Required Software`
- `task1_Dependency.md` -> `4. Dependency Management`
- `task1_Dependency.md` -> `5. Database Setup`
- `task1_Dependency.md` -> `7. Environment Variables`
- `task1_Dependency.md` -> `8. Docker Setup`
- `task1_Dependency.md` -> `9. Local Development Setup`
- `task3_Dependency.md` -> `5. Database Setup`
- `task4_Dependency.md` -> `6. Redis / Queue / External Services`
- `task4_Dependency.md` -> `7. Environment Variables`

### Step 2: Go to Project Directory

```bash
cd Ecommerce
cd backend/services/wishlist-service
```

### Step 3: Install Only New Dependencies

None for Task 5.

If Go modules were never downloaded:

```bash
go mod download
```

### Step 4: Setup Only New Databases/Services

No new database. Use MongoDB setup from previous dependency docs.

Task 5 needs Cart Service readiness:

```bash
curl http://localhost:8083/healthz
```

If Cart Service does not expose `/healthz`, verify using its own docs or a known `POST /api/v1/cart/items` request. Current Wishlist Service does not check Cart readiness at startup; failure appears during move-to-cart request.

### Step 5: Add Only New or Changed Environment Variables

Add Task 5 Cart values in:

```text
backend/services/wishlist-service/.env
```

```env
WISHLIST_CART_SERVICE_BASE_URL=http://localhost:8083
WISHLIST_CART_SERVICE_TIMEOUT=1s
```

Load env:

```bash
set -a
source .env
set +a
```

### Step 6: Run Migrations If Needed

No Task 5 migration is added.

If this is a fresh local database, run previous Mongo migrations or allow startup collection setup as explained in:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`9. Local Development Setup`
```

### Step 7: Start Backend Service

```bash
go run ./cmd/server
```

Expected log includes:

```text
wishlist service listening addr=:8084
cart_service_base_url=http://localhost:8083
```

### Step 8: Verify APIs or Functionality Related to `TASK_FILE_NAME`

First create or confirm wishlist item with `variant_id`. Add-item setup is covered in `task4_Dependency.md`.

Add item:

```bash
curl -X POST http://localhost:8084/api/v1/wishlist/items \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_123" \
  -H "X-User-Roles: buyer" \
  -H "X-Request-ID: req_add_123" \
  -d '{"product_id":"prod_123","variant_id":"var_1"}'
```

Move item to cart:

```bash
curl -X POST http://localhost:8084/api/v1/wishlist/items/prod_123/move-to-cart \
  -H "X-User-ID: user_123" \
  -H "X-User-Roles: buyer" \
  -H "X-Request-ID: req_move_123" \
  -H "X-Idempotency-Key: wishlist-move-user_123-prod_123-var_1"
```

Expected:

```text
HTTP 200
data.cart_id exists
data.items contains product_id=prod_123 and variant_id=var_1
error is null
```

Verify wishlist cleanup:

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval 'db.wishlists.findOne({user_id:"user_123"},{items:1,updated_at:1})'
```

Expected: `prod_123` is no longer present in `items`.

---

## 10. Running the Project

### Minimum Local Run for Task 5

Required:

- Go installed.
- MongoDB running.
- Product Service available if you need to add the wishlist item through API.
- Cart Service available for move-to-cart.
- Kafka disabled for beginner local mode.

Commands:

```bash
cd backend/services/wishlist-service
set -a
source .env
set +a
go test ./...
go run ./cmd/server
```

Health checks:

```bash
curl http://localhost:8084/healthz
curl http://localhost:8084/readyz
```

Task 5 endpoint:

```bash
curl -X POST http://localhost:8084/api/v1/wishlist/items/prod_123/move-to-cart \
  -H "X-User-ID: user_123" \
  -H "X-User-Roles: buyer" \
  -H "X-Request-ID: req_move_123" \
  -H "X-Idempotency-Key: wishlist-move-user_123-prod_123-var_1"
```

### Full Runtime Mode

Required:

- MongoDB running.
- Cart Service reachable.
- Product Service reachable for add-item flow.
- Kafka-compatible broker running only if product events or analytics publisher are enabled.

Keep these env values aligned:

```env
WISHLIST_PRODUCT_SERVICE_BASE_URL=http://localhost:8082
WISHLIST_CART_SERVICE_BASE_URL=http://localhost:8083
WISHLIST_EVENTS_BACKEND=disabled
WISHLIST_ANALYTICS_EVENTS_ENABLED=false
WISHLIST_EVENT_PUBLISHER=disabled
```

---

## 11. Ports & Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| `SERVICE_NAME` HTTP server | `8084` | Wishlist APIs and health checks | Reused |
| MongoDB | `27017` | `wishlist_db.wishlists` persistence | Reused |
| Product Service | `8082` | Add-item product validation before Task 5 test data exists | Reused from Task 4 |
| Cart Service | `8083` | `POST /api/v1/cart/items` for move-to-cart | Task 5 required runtime dependency |
| Kafka broker | `9092` | Optional product events/analytics events | Reused optional |

Networking notes:

- If `:8084` conflicts, change `WISHLIST_HTTP_ADDR`, then call the new port.
- If Cart Service port differs, change `WISHLIST_CART_SERVICE_BASE_URL`.
- Use `localhost` only when both services run on host.
- Use Docker service DNS like `cart-service` when services run in the same Compose network.
- Wishlist `/readyz` currently checks MongoDB readiness, not Cart Service readiness. Cart issues show up during move request.

Example port change:

```bash
WISHLIST_HTTP_ADDR=:18084 go run ./cmd/server
```

Then:

```bash
curl http://localhost:18084/healthz
```

---

## 12. Common Errors & Fixes

Generic errors like Go version mismatch, Docker daemon down, Mongo connection failure, `.env` not sourced, and auth header issues are already covered:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`12. Common Errors & Fixes`
```

Task 5-specific troubleshooting:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `CART_SERVICE_UNAVAILABLE` | Cart Service down, wrong base URL, timeout, DNS issue, or `5xx`. | Start Cart Service and verify `WISHLIST_CART_SERVICE_BASE_URL`. Increase timeout if needed. | Add Cart Service health/dependency check to local stack. |
| `CART_VALIDATION_ERROR` | Cart Service rejected body, often missing/invalid product, variant, or quantity. | Confirm wishlist item has `variant_id`; check Cart API contract. | Add wishlist items with valid variant IDs. |
| `CART_UNAUTHENTICATED` | Cart Service did not accept auth metadata. | Ensure Cart Service accepts `X-User-ID` and trusted service call headers. | Keep auth header contract same across services. |
| `CART_FORBIDDEN` | Buyer role rejected by Cart Service. | Check Cart Service role rules and incoming `X-User-Roles`. | Do not change role naming without updating all services. |
| `CART_PRODUCT_NOT_FOUND` | Product/variant not found from Cart Service perspective. | Use valid product and variant fixtures. | Keep Product and Cart validation data aligned. |
| `CART_ITEM_UNAVAILABLE` | Product out of stock, unavailable, deleted, or Cart conflict. | Use available product/variant or inspect Cart Service response/logs. | Seed known-good inventory for local tests. |
| `VARIANT_REQUIRED_FOR_CART` | Wishlist item has empty `variant_id`. | Add wishlist item again with `variant_id`. | UI should force variant selection before wishlist add when product has variants. |
| `WISHLIST_ITEM_NOT_FOUND` | Product is not in buyer wishlist. | Add item first using Task 4 add API. | Verify user id and product id before move. |
| `MOVE_TO_CART_PARTIAL_FAILURE` | Cart add succeeded, then wishlist cleanup failed. | Check MongoDB availability and logs; manually inspect cart/wishlist state. | Add retry/compensation job or admin repair workflow. |
| Duplicate cart quantity after retry | Missing or changing idempotency key. | Send stable `X-Idempotency-Key` for same move attempt. | Generate deterministic key per user/product/variant. |
| Move works in tests but curl fails | Tests use fake Cart client; local Cart Service is missing. | Run real Cart Service or mock `POST /api/v1/cart/items`. | Document Cart Service setup in local compose. |

---

## 13. Security & Best Practices

### Task 5 Security Audit

| Area | Current observation | Risk | Suggested fix |
|---|---|---|---|
| Cart Service URL | Config supports env and validates HTTP/HTTPS URL. | Wrong/public URL can leak internal auth headers. | Use internal/private service URL in shared/prod env. |
| Auth headers | Wishlist trusts gateway-provided `X-User-ID` and role headers, then forwards user context to Cart. | Direct public exposure allows spoofing. | Keep service behind API Gateway/private network. |
| Idempotency | `X-Idempotency-Key` is forwarded, and fallback key is generated. | Cart Service must honor key or retries may duplicate quantity. | Confirm Cart Service idempotency support and retention period. |
| Timeout | Cart timeout defaults to `1s` and must be positive. | Too low can fail under normal latency; too high ties up Wishlist requests. | Tune per environment and monitor latency. |
| Partial failure | Cart add can succeed while wishlist cleanup fails. | User sees error but cart changed. | Add compensation/retry workflow and alert on partial failure logs. |
| Readiness | `/readyz` checks MongoDB, not Cart Service. | Deployment may look ready while Cart is unreachable. | Add dependency-specific readiness or synthetic checks if strict readiness is needed. |
| `.env` | Current checked-in `.env` does not explicitly include Cart vars. | Beginners may not know Cart URL/timeout exist. | Add `.env.example` with Task 5 Cart variables. |
| Cart Service implementation | `backend/services/cart-service` not found in this snapshot. | End-to-end setup is incomplete locally. | Add Cart Service, local mock, or compose service docs. |

### Task 5 Best Practices

- Always send a stable `X-Idempotency-Key` from frontend/gateway for move-to-cart.
- Keep `WISHLIST_CART_SERVICE_BASE_URL` internal in production.
- Keep `WISHLIST_AUTH_REQUIRE_BUYER_ROLE=true` outside controlled tests.
- Do not write directly to Cart DB from Wishlist Service.
- Do not remove wishlist item before Cart Service success.
- Verify Cart Service logs when Wishlist returns `CART_*` errors.
- Add item with `variant_id` when the product has variants.
- Keep Product, Wishlist, and Cart test fixtures aligned.
- Run `go test ./...` after config/client/handler changes.

---

## 14. Missing or Misconfigured Things

| Missing/misaligned item | Why it matters | Recommended action |
|---|---|---|
| No `backend/services/cart-service` folder found | Task 5 endpoint depends on Cart Service for real move-to-cart. | Add Cart Service implementation, run external service, or provide local mock/compose entry. |
| Current `backend/services/wishlist-service/.env` lacks explicit Cart vars | Config has defaults, but beginners may miss required downstream dependency. | Add `WISHLIST_CART_SERVICE_BASE_URL` and `WISHLIST_CART_SERVICE_TIMEOUT` to `.env.example` or local `.env`. |
| Wishlist readiness does not check Cart Service | `/readyz` can be green while move-to-cart fails. | Add optional dependency readiness or documented smoke test. |
| No service Dockerfile/compose file found for Wishlist | Containerized local stack is manual. | Add Dockerfile and compose with MongoDB plus Cart/Product services. |
| `api/master-api.json` mentions gRPC methods, current code uses HTTP client/server | Architecture contract and runtime transport differ. | Keep `CartClient` interface and add gRPC implementation later if platform standard requires it. |
| Partial failure has no background repair workflow | Cart can update but wishlist cleanup can fail. | Add retry/compensation queue or admin reconciliation task. |

---

## 15. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `task1_Dependency.md` | `2. Tech Stack` | Go, MongoDB, Kafka, Product/Cart high-level tech already explained. |
| `task1_Dependency.md` | `3. Required Software` | Git, Go, MongoDB, Docker, curl setup unchanged. |
| `task1_Dependency.md` | `4. Dependency Management` | Go modules and dependency commands unchanged. |
| `task1_Dependency.md` | `5. Database Setup` | MongoDB installation, URI, credentials, Docker setup unchanged. |
| `task1_Dependency.md` | `7. Environment Variables` | Full env reference already exists, including Cart defaults. |
| `task1_Dependency.md` | `8. Docker Setup` | Base MongoDB/Kafka Docker setup reused. |
| `task1_Dependency.md` | `11. Ports & Networking` | Standard Wishlist, MongoDB, Product, Cart, Kafka ports already listed. |
| `task1_Dependency.md` | `12. Common Errors & Fixes` | Generic Go/Mongo/Docker/env/auth errors already documented. |
| `task2_Dependency.md` | `5. Database Setup` | MongoDB decision and DB name/collection reasoning reused. |
| `task3_Dependency.md` | `5. Database Setup` | `wishlists` collection validator and index verification reused. |
| `task4_Dependency.md` | `6. Redis / Queue / External Services` | Product Service dependency for creating wishlist item reused. |
| `task4_Dependency.md` | `7. Environment Variables` | Auth header and Product Service env guidance reused. |
| `task4_Dependency.md` | `9. Local Development Setup` | Add/remove API setup reused before Task 5 move test. |

---

## 16. Final Checklist

Use this before saying Task 5 local setup is ready:

- [ ] Previous dependency documentation checked.
- [ ] `task5.md` original file not modified.
- [ ] No duplicate base Go/Mongo/Docker setup added.
- [ ] Go dependencies downloaded if needed.
- [ ] `go test ./...` passes in `backend/services/wishlist-service`.
- [ ] MongoDB is running and `/readyz` is healthy.
- [ ] `wishlists` collection and indexes are ready.
- [ ] Product Service or mock is available if adding wishlist item through API.
- [ ] Cart Service or mock is available for `POST /api/v1/cart/items`.
- [ ] `WISHLIST_CART_SERVICE_BASE_URL` is set or confirmed default is correct.
- [ ] `WISHLIST_CART_SERVICE_TIMEOUT` is positive.
- [ ] `.env` is sourced before running service.
- [ ] Kafka disabled for beginner local mode, or Kafka broker running for event mode.
- [ ] Wishlist item exists before move-to-cart.
- [ ] Wishlist item has `variant_id`.
- [ ] Move request includes `X-User-ID`.
- [ ] Move request includes `X-User-Roles: buyer`.
- [ ] Move request includes stable `X-Idempotency-Key`.
- [ ] Response returns updated Cart data.
- [ ] MongoDB wishlist no longer contains moved product after success.
- [ ] Logs checked for `CART_*` errors or partial failure.
- [ ] Real credentials are not committed to git.

Final beginner summary: Task 5 ke liye new install nahi hai. Main kaam hai Cart Service ko reachable banana, Cart env vars clear rakhna, wishlist item me `variant_id` ensure karna, and stable idempotency key ke saath `POST /move-to-cart` test karna.
