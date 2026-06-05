# Project Dependency & Setup Guide

## Variable Values

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Wishlist Service` |
| `TASK_FILE_NAME` | `task4.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task4_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

Note: Ye guide `INPUT_FILE_PATH` ko analyze karke banaya gaya hai. Original `TASK_FILE_NAME` modify nahi kiya gaya. Is file ka focus sirf setup, dependency, environment, DevOps, verification, and beginner onboarding hai.

---

## 1. Project Overview

`TASK_FILE_NAME` ka main scope:

```text
Add wishlist item API
Remove wishlist item API
Product id validation through Product Service
Duplicate product block
MongoDB duplicate-safe add/remove writes
```

Simple Hinglish me: buyer apni wishlist me product add/remove kar sakta hai. Add karte time `SERVICE_NAME` Product Service ko call karke confirm karta hai ki product exist karta hai, `published` hai, aur optional `variant_id` valid hai. Remove ke time Product Service call required nahi hai because deleted/stale product ko bhi user wishlist se remove kar sake.

### Runtime code paths checked

| Runtime file | Why checked |
|---|---|
| `backend/services/wishlist-service/internal/config/config.go` | Env vars, defaults, validation, ports, downstream URLs. |
| `backend/services/wishlist-service/internal/clients/product_http_validator.go` | Task 4 Product Service dependency and response contract. |
| `backend/services/wishlist-service/internal/usecase/wishlist_service.go` | Add/remove flow, duplicate handling, product validation. |
| `backend/services/wishlist-service/internal/repository/mongo_wishlist_repository.go` | MongoDB `$push`, `$pull`, duplicate guard, indexes. |
| `backend/services/wishlist-service/internal/transport/http/wishlist_handler.go` | REST routes, headers, request validation, error mapping. |
| `backend/services/wishlist-service/cmd/server/main.go` | Startup dependencies and background runner behavior. |
| `api/master-api.json` | API contract for `AddWishlistItem` and `RemoveWishlistItem`. |

Important reuse note:

Most base setup is already documented in previous dependency files. Is file me duplicate installation guide repeat nahi kiya gaya. Sirf Task 4 ke new/incremental setup points explain kiye gaye hain.

```md
Read first:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`
`TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`
`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`
```

---

## 2. Tech Stack

Most technology explanation already covered hai. Task 4 ka real new focus Product Service validation and authenticated add/remove API testing hai.

| Technology / Service | Task 4 Status | Required? | Setup Documentation |
|---|---|---:|---|
| Go | Reused | Yes | `task1_Dependency.md` -> `4. Dependency Management` |
| Go modules | Reused | Yes | `task1_Dependency.md` -> `4. Dependency Management` |
| `net/http` | Reused transport | Yes | `task1_Dependency.md` -> `2. Tech Stack` |
| MongoDB | Reused persistence | Yes | `task1_Dependency.md` -> `5. Database Setup`; `task3_Dependency.md` -> `5. Database Setup` |
| MongoDB Go Driver v2 | Reused dependency | Yes | `task1_Dependency.md` -> `4. Dependency Management` |
| Product Service HTTP API | Task 4 core dependency | Required for add item | Explained below |
| Gateway/auth headers | Task 4 core runtime contract | Required for add/remove | Explained below |
| Kafka-compatible broker | Reused optional runtime feature | No for Task 4 core flow | `task1_Dependency.md` -> `6. Redis / Queue / External Services` |
| Docker | Reused local dependency runner | Recommended | `task1_Dependency.md` -> `8. Docker Setup` |
| Redis | Not introduced | No | No setup needed |
| RabbitMQ | Not introduced | No | No setup needed |
| Cart Service | Runtime config reused, not Task 4 flow | Not needed for add/remove | `task1_Dependency.md` -> `6. Redis / Queue / External Services` |

### Task 4 new technical concept: Product Service validation

Product Service product catalog ka source of truth hai. Simple Hinglish me: `SERVICE_NAME` khud Product DB read nahi karega. Jab buyer product wishlist me add karega, service Product Service se poochega:

- product exist karta hai?
- product `published` hai?
- requested `variant_id` exist karta hai?
- price and stock snapshot response me available hai?

Current runtime call:

```text
GET {WISHLIST_PRODUCT_SERVICE_BASE_URL}/api/v1/products/{product_id}
```

Default:

```text
WISHLIST_PRODUCT_SERVICE_BASE_URL=http://localhost:8082
```

### Task 4 auth concept

Add/remove APIs `user_id` request body se nahi leti. Current HTTP service trusted headers use karta hai:

| Header | Default | Purpose |
|---|---|---|
| `X-User-ID` | `WISHLIST_AUTH_USER_ID_HEADER` | Authenticated buyer id. |
| `X-User-Roles` | `WISHLIST_AUTH_ROLES_HEADER` | Must include `buyer` by default. |
| `X-Request-ID` | `WISHLIST_REQUEST_ID_HEADER` | Trace id for logs and response. |

Production me ye headers public user se directly trust nahi karne chahiye. API Gateway/JWT middleware should validate token and inject trusted headers.

---

## 3. Required Software

No new language runtime, package manager, database server, or queue is introduced by `TASK_FILE_NAME`.

Base software setup is reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`3. Required Software`
`4. Dependency Management`
`5. Database Setup`
`8. Docker Setup`
```

Task 4-specific required readiness:

| Requirement | Required? | Why |
|---|---:|---|
| Product Service or local Product API mock | Yes for add item API | Add item blocks until product validation succeeds. |
| MongoDB with `wishlists` collection ready | Yes | Add/remove writes wishlist documents. |
| curl/Postman/HTTP client | Recommended | Task 4 API verification ke liye. |
| API Gateway or trusted header injection | Required in real deployment | User identity spoofing avoid karne ke liye. |
| Kafka broker | No for Task 4 core add/remove | Only needed if event/analytics publishing enabled. |

Current repo snapshot note:

```text
backend/services/product-service
```

folder exists, but this snapshot only shows `.env` there. End-to-end add-item testing needs a runnable Product Service implementation, a running external Product Service, or a local mock that returns the expected product JSON.

---

## 4. Dependency Management

No new Go module dependency is introduced by `TASK_FILE_NAME`.

Existing runtime module:

```text
backend/services/wishlist-service/go.mod
```

Existing direct dependencies remain:

| Dependency | Version | Task 4 usage |
|---|---:|---|
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | Add/remove persistence in MongoDB. |
| `github.com/segmentio/kafka-go` | `v0.4.49` | Optional event/analytics runtime, not required for Task 4 core API. |

Reuse Go setup commands from:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Dependency Management`
```

Task 4 beginner warning:

- Do not run `go get` for Product Service validation. Current code uses Go standard library `net/http`.
- Do not add Redis/RabbitMQ packages for this task.
- If `go test ./...` fails due dependency download, follow previous Go module troubleshooting instead of changing `go.mod` manually.

---

## 5. Database Setup

Task 4 does not add a new database, collection, or migration.

Database setup is reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`5. Database Setup`
`8. Docker Setup`
```

Collection schema and index setup is reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Section:
`5. Database Setup`
```

### Task 4 database behavior

| Operation | MongoDB behavior | Setup impact |
|---|---|---|
| Add item | `EnsureWishlist`, then atomic `$push` with `items.product_id: { "$ne": product_id }` | Requires `wishlists` collection and user/product index readiness. |
| Duplicate add | `ModifiedCount == 0` maps to duplicate error | No new index required, but service logic must be used. |
| Remove item | `$pull` by `items.product_id` | Product Service not needed for remove. |
| Read response | `FindByUserID` returns updated wishlist | MongoDB must be reachable. |

Task 4 relevant existing values:

| Item | Value | Status |
|---|---|---|
| Database type | MongoDB | Reused |
| Database name | `wishlist_db` | Reused default |
| Collection | `wishlists` | Reused |
| Default Mongo port | `27017` | Reused |
| Main env var | `WISHLIST_MONGO_URI` | Reused |

### Task 4 verification commands

After following previous setup:

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.wishlists.getIndexes()"
curl http://localhost:8084/readyz
```

Expected:

- `idx_wishlists_items_product_id` exists.
- `/readyz` returns `{"status":"ok","service":"wishlist-service"}` when MongoDB is reachable.

Important: Direct MongoDB insert can bypass Product Service validation. For Task 4 testing, prefer API add/remove flow.

---

## 6. Redis / Queue / External Services

### Product Service

Product Service is the only Task 4-specific external dependency.

| Item | Value |
|---|---|
| Purpose | Validate product before wishlist add. |
| Required for add item? | Yes |
| Required for remove item? | No |
| Default base URL | `http://localhost:8082` |
| Env var | `WISHLIST_PRODUCT_SERVICE_BASE_URL` |
| Fallback env var | `PRODUCT_SERVICE_BASE_URL` |
| Timeout env var | `WISHLIST_PRODUCT_SERVICE_TIMEOUT` |
| Fallback timeout alias | `WISHLIST_PRODUCT_CALL_TIMEOUT_MS` |
| Runtime endpoint | `GET /api/v1/products/{product_id}` |

Expected successful Product Service response shape:

```json
{
  "data": {
    "product_id": "prod_123",
    "status": "published",
    "price": {
      "amount": 299900,
      "currency": "INR"
    },
    "variants": [
      {
        "variant_id": "var_1",
        "price": {
          "amount": 299900,
          "currency": "INR"
        },
        "stock_quantity": 10,
        "reserved_quantity": 2
      }
    ]
  },
  "error": null
}
```

Task 4 validation rules:

| Product Service result | Wishlist result |
|---|---|
| HTTP `404` | `PRODUCT_NOT_FOUND` |
| HTTP `409` or `410` | `PRODUCT_NOT_AVAILABLE` |
| HTTP `408`, `429`, or `5xx` | `PRODUCT_SERVICE_UNAVAILABLE` |
| `status` not `published` | `PRODUCT_NOT_AVAILABLE` |
| requested `variant_id` missing | `VARIANT_NOT_FOUND` |
| price invalid/missing | Product can still be added with no price snapshot |

### Kafka / analytics events

Kafka setup is reused from previous dependency docs. It is not required for the core Task 4 add/remove API if local env disables analytics publishing.

Recommended local Task 4 setting:

```env
WISHLIST_ANALYTICS_EVENTS_ENABLED=false
WISHLIST_EVENT_PUBLISHER=disabled
WISHLIST_EVENTS_BACKEND=disabled
```

If analytics stays enabled with Kafka publisher, make sure broker `localhost:9092` is running. Otherwise add/remove may still respond, but logs can show event publish failures.

### Redis, RabbitMQ, NATS, MinIO, SMTP, Stripe, Twilio

No setup needed for `TASK_FILE_NAME`.

---

## 7. Environment Variables

Full `.env` reference is already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Environment Variables`
```

### Task 4 env delta only

This is not a full `.env`. Ye sirf Task 4 add/remove API testing ke liye important variables hain.

```env
# Product Service validation for add item
WISHLIST_PRODUCT_SERVICE_BASE_URL=http://localhost:8082
WISHLIST_PRODUCT_SERVICE_TIMEOUT=1s

# Auth context headers used by local direct HTTP testing
WISHLIST_AUTH_USER_ID_HEADER=X-User-ID
WISHLIST_AUTH_ROLES_HEADER=X-User-Roles
WISHLIST_AUTH_REQUIRE_BUYER_ROLE=true
WISHLIST_REQUEST_ID_HEADER=X-Request-ID

# Recommended for simple local Task 4 testing if Kafka is not running
WISHLIST_EVENTS_BACKEND=disabled
WISHLIST_ANALYTICS_EVENTS_ENABLED=false
WISHLIST_EVENT_PUBLISHER=disabled
```

### Variable explanation

| Variable | Required? | Example | Task 4 purpose | Security/config note |
|---|---:|---|---|---|
| `WISHLIST_PRODUCT_SERVICE_BASE_URL` | Required for add | `http://localhost:8082` | Product validation endpoint base. | Use internal service URL in prod. |
| `PRODUCT_SERVICE_BASE_URL` | Optional fallback | `http://localhost:8082` | Used only if service-specific var missing. | Prefer `WISHLIST_PRODUCT_SERVICE_BASE_URL`. |
| `WISHLIST_PRODUCT_SERVICE_TIMEOUT` | Optional | `1s` | Timeout for Product Service call. | Keep positive and realistic. |
| `WISHLIST_PRODUCT_CALL_TIMEOUT_MS` | Optional fallback | `1000` | Old millisecond timeout alias. | Avoid setting both timeout vars differently. |
| `WISHLIST_AUTH_USER_ID_HEADER` | Required | `X-User-ID` | Authenticated buyer id source. | Must be injected by trusted gateway. |
| `WISHLIST_AUTH_ROLES_HEADER` | Required | `X-User-Roles` | Role source. | Keep role mapping consistent. |
| `WISHLIST_AUTH_REQUIRE_BUYER_ROLE` | Optional | `true` | Requires `buyer` role. | Do not disable outside local tests. |
| `WISHLIST_REQUEST_ID_HEADER` | Required | `X-Request-ID` | Request tracing. | Not secret. |
| `WISHLIST_ANALYTICS_EVENTS_ENABLED` | Optional | `false` for simple local | Avoids analytics outbox during beginner API testing. | Enable only when event setup is ready. |
| `WISHLIST_EVENT_PUBLISHER` | Optional | `disabled` for simple local | Avoids Kafka publishing during beginner API testing. | Use `kafka` only with broker ready. |

### Where to place credentials/config

Create local env file here:

```text
backend/services/wishlist-service/.env
```

Current Go code reads OS environment variables. It does not automatically load `.env`.

Load before starting service:

```bash
cd backend/services/wishlist-service
set -a
source .env
set +a
```

Common Task 4 env mistakes:

| Mistake | Result | Fix |
|---|---|---|
| Product URL points to wrong port | Add returns `PRODUCT_SERVICE_UNAVAILABLE` | Set `WISHLIST_PRODUCT_SERVICE_BASE_URL` correctly. |
| Product URL has trailing path like `/api/v1/products` | Runtime builds duplicated path | Use base only, for example `http://localhost:8082`. |
| `.env` created but not sourced | App uses defaults or old shell values | Run `set -a; source .env; set +a`. |
| `WISHLIST_AUTH_REQUIRE_BUYER_ROLE=false` in shared env | Buyer role bypass risk | Keep it `true` outside controlled local tests. |
| Kafka not running and analytics enabled | Event publish logs/errors | Disable analytics for simple Task 4 local testing. |

---

## 8. Docker Setup

No new Dockerfile, Docker Compose service, container, volume, or network is introduced by `TASK_FILE_NAME`.

Reuse base Docker dependency setup:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`8. Docker Setup`
```

Task 4 Docker-specific notes:

| Docker item | Status | Meaning |
|---|---|---|
| MongoDB container | Reused | Required if MongoDB is not installed locally. |
| Kafka/Redpanda container | Optional, reused | Only needed if events/analytics are enabled. |
| Product Service container | Not provided by this task | Required for end-to-end add-item testing if Product Service runs via Docker. |
| Service Dockerfile | Not found in current snapshot | Go service is currently run directly for local development. |

### Host vs Docker URL examples

If `SERVICE_NAME` runs on host and Product Service runs on host:

```env
WISHLIST_PRODUCT_SERVICE_BASE_URL=http://localhost:8082
```

If `SERVICE_NAME` later runs inside Docker Compose and Product Service is another compose service:

```env
WISHLIST_PRODUCT_SERVICE_BASE_URL=http://product-service:8082
WISHLIST_MONGO_URI=mongodb://wishlist-mongo:27017
```

Beginner note: `localhost` inside a container means the same container, not your host machine.

---

## 9. Local Development Setup

### Step 1: Read previous dependency documentation

Read these first:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
```

Minimum sections:

- `task1_Dependency.md` -> `3. Required Software`
- `task1_Dependency.md` -> `4. Dependency Management`
- `task1_Dependency.md` -> `5. Database Setup`
- `task1_Dependency.md` -> `7. Environment Variables`
- `task1_Dependency.md` -> `9. Local Development Setup`
- `task3_Dependency.md` -> `5. Database Setup`

### Step 2: Go to project directory

```bash
cd Ecommerce
cd backend/services/wishlist-service
```

### Step 3: Install only new dependencies

None for Task 4.

If Go modules were never downloaded:

```bash
go mod download
```

### Step 4: Setup only new databases/services

No new database. Use MongoDB setup from previous dependency docs.

Task 4 needs Product Service readiness:

```bash
curl http://localhost:8082/api/v1/products/prod_123
```

Expected for add-item success:

- HTTP status is `200`.
- response product id matches `prod_123`.
- `status` is `published`.
- if using `variant_id`, that variant exists in `variants`.

If Product Service is unavailable, only remove flow can be partially tested through HTTP. Add flow will correctly return `PRODUCT_SERVICE_UNAVAILABLE`.

### Step 5: Add only new or changed environment variables

Add or confirm Task 4 env delta in:

```text
backend/services/wishlist-service/.env
```

Recommended local values:

```env
WISHLIST_PRODUCT_SERVICE_BASE_URL=http://localhost:8082
WISHLIST_PRODUCT_SERVICE_TIMEOUT=1s
WISHLIST_AUTH_USER_ID_HEADER=X-User-ID
WISHLIST_AUTH_ROLES_HEADER=X-User-Roles
WISHLIST_AUTH_REQUIRE_BUYER_ROLE=true
WISHLIST_REQUEST_ID_HEADER=X-Request-ID
WISHLIST_ANALYTICS_EVENTS_ENABLED=false
WISHLIST_EVENT_PUBLISHER=disabled
```

Load env:

```bash
set -a
source .env
set +a
```

### Step 6: Run migrations if needed

Task 4 adds no migration.

Reuse previous migration/setup:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Section:
`5. Database Setup`
```

Minimum verification:

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.wishlists.getIndexes()"
```

### Step 7: Start backend service

```bash
go run ./cmd/server
```

Expected log:

```text
wishlist service listening addr=:8084
```

### Step 8: Verify health

```bash
curl http://localhost:8084/healthz
curl http://localhost:8084/readyz
```

Expected:

```json
{"status":"ok","service":"wishlist-service"}
```

### Step 9: Verify APIs related to `TASK_FILE_NAME`

Add item:

```bash
curl -X POST http://localhost:8084/api/v1/wishlist/items \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_123" \
  -H "X-User-Roles: buyer" \
  -H "X-Request-ID: req_task4_add_1" \
  -d '{"product_id":"prod_123","variant_id":"var_1"}'
```

Expected success:

```json
{
  "data": {
    "wishlist_id": "wish_...",
    "user_id": "user_123",
    "items": [
      {
        "product_id": "prod_123",
        "variant_id": "var_1",
        "availability": "in_stock"
      }
    ]
  },
  "request_id": "req_task4_add_1",
  "error": null
}
```

Duplicate add check:

```bash
curl -X POST http://localhost:8084/api/v1/wishlist/items \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_123" \
  -H "X-User-Roles: buyer" \
  -H "X-Request-ID: req_task4_add_duplicate" \
  -d '{"product_id":"prod_123","variant_id":"var_1"}'
```

Expected:

```text
HTTP 409
error.code = WISHLIST_ITEM_ALREADY_EXISTS
```

Remove item:

```bash
curl -X DELETE http://localhost:8084/api/v1/wishlist/items/prod_123 \
  -H "X-User-ID: user_123" \
  -H "X-User-Roles: buyer" \
  -H "X-Request-ID: req_task4_remove_1"
```

Expected:

```text
HTTP 200
Product removed from returned wishlist items.
```

Idempotent remove check:

```bash
curl -X DELETE http://localhost:8084/api/v1/wishlist/items/prod_123 \
  -H "X-User-ID: user_123" \
  -H "X-User-Roles: buyer" \
  -H "X-Request-ID: req_task4_remove_2"
```

Expected:

```text
HTTP 200
Wishlist stays unchanged.
```

---

## 10. Running the Project

No new run command introduced by `TASK_FILE_NAME`.

Reuse:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`10. Running the Project`
```

Task 4 local run mode:

| Mode | MongoDB | Product Service | Kafka | Best for |
|---|---:|---:|---:|---|
| Remove-only smoke test | Required | Not required | Disabled | Handler/auth/Mongo remove behavior. |
| Add/remove Task 4 test | Required | Required | Disabled | Main Task 4 API verification. |
| Event-enabled runtime | Required | Required for add | Required if publishers enabled | Later event/analytics validation. |

Minimum Task 4 command flow:

```bash
cd backend/services/wishlist-service
set -a
source .env
set +a
go test ./...
go run ./cmd/server
```

In another terminal:

```bash
curl http://localhost:8084/readyz
curl http://localhost:8082/api/v1/products/prod_123
```

Then run the add/remove curl commands from section 9.

---

## 11. Common Errors & Fixes

Generic setup errors are already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`12. Common Errors & Fixes`
```

Task 4-specific errors:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `PRODUCT_SERVICE_UNAVAILABLE` | Product Service down, wrong URL, timeout, invalid JSON, or Product API returns `5xx`. | Start Product Service, fix `WISHLIST_PRODUCT_SERVICE_BASE_URL`, or use a valid local mock. | Check Product Service health before add-item tests. |
| `PRODUCT_NOT_FOUND` | Product Service returned `404`. | Use an existing product id. | Keep a known local fixture product for Task 4 testing. |
| `PRODUCT_NOT_AVAILABLE` | Product `status` is not `published`, or Product Service returned `409`/`410`. | Publish test product or use a published product. | Test data should include published products. |
| `VARIANT_NOT_FOUND` | Request has `variant_id`, but Product Service response does not include it. | Use valid variant id or omit `variant_id`. | Validate UI/API fixture variants. |
| `WISHLIST_ITEM_ALREADY_EXISTS` | Same user already has same `product_id` in wishlist. | Treat as expected duplicate block or remove first. | UI should disable repeated add after success. |
| `UNAUTHENTICATED` | `X-User-ID` missing or empty. | Add auth header in local curl. | Gateway should inject user id. |
| `FORBIDDEN` | `buyer` role missing while `WISHLIST_AUTH_REQUIRE_BUYER_ROLE=true`. | Add `X-User-Roles: buyer`. | Keep role mapping documented. |
| `VALIDATION_ERROR` for `variant_id` | `variant_id` was provided as empty string. | Omit it or pass a real variant id. | Frontend should not send empty string for optional fields. |
| `VALIDATION_ERROR` for `body` | JSON invalid, unknown field, multiple JSON objects, or body too large. | Send one valid JSON object with `product_id` and optional `variant_id`. | Use API schema examples in tests. |
| Remove route returns `NOT_FOUND` | Product id path contains an extra slash or bad encoding. | URL-encode product id if needed. | Keep product ids slash-free when possible. |
| Add works but price is missing in response | Product Service response price is absent or invalid currency/amount. | Fix Product Service fixture price shape. | Return uppercase 3-letter currency and integer minor amount. |
| Add/remove logs analytics publish issues | Kafka disabled incorrectly or broker unavailable while analytics enabled. | Set `WISHLIST_ANALYTICS_EVENTS_ENABLED=false` and `WISHLIST_EVENT_PUBLISHER=disabled` for simple local testing. | Keep local `.env` explicit. |

---

## 12. Security & Best Practices

### Task 4 security audit

| Area | Current observation | Risk | Suggested fix |
|---|---|---|---|
| Product Service URL | Configurable through env and validated as HTTP/HTTPS URL. | Wrong URL can validate against wrong environment. | Use env-specific internal DNS, not random public URLs. |
| Product Service timeout | Default is `1s`. | Too low can cause false `PRODUCT_SERVICE_UNAVAILABLE`; too high can slow API. | Tune per environment and monitor latency. |
| Product Service trust | Wishlist accepts product snapshots from Product Service response. | Bad downstream data creates stale/wrong wishlist snapshots. | Keep Product Service response contract tested. |
| Auth headers | Service trusts `X-User-ID` and role headers. | Public direct access could spoof user identity. | Run behind trusted API Gateway/private network or validate JWT in-service. |
| Buyer role | Required by default. | Disabling role check can expose wishlist mutation to non-buyers. | Keep `WISHLIST_AUTH_REQUIRE_BUYER_ROLE=true` outside local tests. |
| Duplicate rule | Current duplicate key is product-level `product_id`, not variant-level. | User cannot save multiple variants of same product. | Keep UI/product requirements aligned, or plan variant-level migration later. |
| Remove behavior | Remove is idempotent and does not validate Product Service. | This is intentional, but may surprise testers expecting `404`. | Document API behavior and frontend expectations. |
| `.env` secrets | No hardcoded production secret needed for Task 4. | Real downstream URLs/tokens could leak if committed. | Keep real `.env` untracked and commit only `.env.example`. |
| Analytics defaults | Code defaults analytics to enabled with Kafka publisher. | Local add/remove can create Kafka-related noise if not disabled. | Use explicit local Task 4 env values. |

### Task 4 best practices

- Keep Product Service as source of truth for product status, price, and variants.
- Never directly query Product DB from `SERVICE_NAME`.
- Keep add-item Product Service call short and timeout-protected.
- Use request id headers during testing so logs can be traced.
- Keep `user_id` from trusted auth context, not request body.
- Do not expose this service directly to public internet while it trusts auth headers.
- Treat wishlist `last_known_price` as display snapshot only. Checkout must revalidate price and stock.
- Keep duplicate behavior product-level unless product requirements explicitly change.
- Keep local fixtures stable: one published product, one draft product, one missing variant case.
- Disable Kafka/analytics for beginner Task 4 testing unless event behavior is being tested.

---

## 13. Missing or Misconfigured Things

| Item | Why it matters | Recommended action |
|---|---|---|
| Runnable Product Service code not visible in current snapshot | Task 4 add-item API cannot be verified end-to-end without Product Service. | Add/run Product Service or provide a local mock for `GET /api/v1/products/{product_id}`. |
| Cart Service folder not present in current snapshot | Runtime config still initializes a Cart client, though Task 4 add/remove does not call it. | Keep `WISHLIST_CART_SERVICE_BASE_URL` as a valid URL; Cart Service only matters for move-to-cart. |
| No service Dockerfile found | Containerized local/prod run is not self-contained. | Add a Dockerfile for `backend/services/wishlist-service` in a later DevOps task. |
| No full local compose file found | Mongo/Product/Kafka cannot be started as one documented stack. | Add local compose with MongoDB, Product Service/mock, and optional Kafka. |
| No `.env.example` found for service | Beginners may not know required env vars without reading docs. | Add `backend/services/wishlist-service/.env.example` with safe dummy values. |
| API contract includes `GET /api/v1/wishlist`, current handler registers item routes and health only | Testers may call an unimplemented route while verifying Task 4. | Implement route later or update API contract/docs. |
| Architecture mentions gRPC methods, current runtime exposes HTTP handlers | Contract/runtime mismatch for clients expecting gRPC. | Add gRPC transport later or keep HTTP-only scope explicit. |
| Product dependency has no health/readiness check in `/readyz` | Service can be ready while Product Service is down, then add-item fails. | Add optional downstream readiness or startup dependency check if deployment needs it. |

---

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack` | Go, MongoDB, Kafka, Docker, Product/Cart service basics already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Required Software` | Git, Go, MongoDB, `mongosh`, Docker, curl setup already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Dependency Management` | Go module commands and troubleshooting already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Setup` | MongoDB install, Docker run, connection strings, and migration flow already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Redis / Queue / External Services` | Kafka/Redpanda, Product Service, Cart Service, and Redis/RabbitMQ status already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Environment Variables` | Full env variable reference and `.env` loading behavior already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Docker Setup` | Dependency container setup, volumes, and Docker networking already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Local Development Setup` | Clone, Mongo start, env load, tests, and service start flow already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `10. Running the Project` | Minimum and event-enabled run modes already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `11. Ports & Networking` | Port conflicts and host-vs-container networking already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `12. Common Errors & Fixes` | Generic setup troubleshooting already documented. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | MongoDB decision and `wishlist_db.wishlists` ownership already documented. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `13. Security & Best Practices` | Cross-service DB ownership rule already documented. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `5. Database Setup` | `wishlists` validator/index readiness already documented. |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `12. Common Errors & Fixes` | Collection validator and duplicate-direct-DB-write issues already documented. |

---

## 15. Final Checklist

- [ ] Previous dependency files checked before using this guide.
- [ ] Original `TASK_FILE_NAME` left unchanged.
- [ ] No duplicate Go/Mongo/Docker installation guide added.
- [ ] No new Go dependency added for Task 4.
- [ ] MongoDB running and `/readyz` returns healthy.
- [ ] `wishlists` collection and indexes verified from previous setup.
- [ ] `WISHLIST_PRODUCT_SERVICE_BASE_URL` points to the correct Product Service or mock.
- [ ] Product Service returns a `published` product for add-item testing.
- [ ] Product Service returns valid variant data if `variant_id` is used.
- [ ] `WISHLIST_AUTH_USER_ID_HEADER` and `WISHLIST_AUTH_ROLES_HEADER` confirmed.
- [ ] Local curl requests include `X-User-ID` and `X-User-Roles: buyer`.
- [ ] Kafka/analytics disabled for simple local Task 4 testing, or broker is running.
- [ ] Add item returns `HTTP 200` for valid product.
- [ ] Duplicate add returns `HTTP 409` with `WISHLIST_ITEM_ALREADY_EXISTS`.
- [ ] Remove item returns `HTTP 200`.
- [ ] Repeated remove remains idempotent.
- [ ] Logs checked for Product Service timeout/unavailable errors.
- [ ] No real credentials committed in `.env`.
