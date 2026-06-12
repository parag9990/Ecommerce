# Project Dependency & Setup Guide

```text
SERVICE_NAME = "Cart Service"
TASK_FILE_NAME = "task6.md"
INPUT_FILE_PATH = TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}
OUTPUT_FILE_NAME = ${TASK_FILE_NAME without ".md"}_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}

PREVIOUS_TASK1_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
PREVIOUS_TASK2_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
PREVIOUS_TASK3_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task3_Dependency.md
PREVIOUS_TASK4_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task4_Dependency.md
PREVIOUS_TASK5_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task5_Dependency.md
```

## 1. Project Overview

Ye document `${SERVICE_NAME}` ke `${TASK_FILE_NAME}` ke liye dependency, setup, environment, database, Docker, DevOps, aur troubleshooting guide hai.

`${TASK_FILE_NAME}` ka focus Coupon Preview flow hai:

- Shopper `POST /api/v1/cart/coupons/preview` call karta hai.
- `${SERVICE_NAME}` current owner ka active cart MongoDB se load karta hai.
- Coupon code trim + uppercase normalize hota hai.
- Active non-empty cart ka trusted context CMS Service ko HTTP POST se bheja jata hai.
- CMS Service coupon ko valid/invalid bolta hai and preview discount return karta hai.
- Valid result par `coupon_code`, `coupon_preview`, and `totals.discount` MongoDB me save hote hain.
- Invalid result par previous discount clear hota hai and invalid reason save hota hai.
- Redis active cart and summary cache best-effort refresh hote hain.
- Checkout/final apply abhi bhi Order Service me fresh validation se hona chahiye.

Important scope:

- `${INPUT_FILE_PATH}` modify nahi kiya gaya.
- `${OUTPUT_FILE_PATH}` new dependency/setup guide hai.
- No new Go module dependency was introduced by `${TASK_FILE_NAME}`.
- No new MongoDB collection, Redis key family, queue, Dockerfile, Docker container, volume, network, or exposed port was introduced by `${TASK_FILE_NAME}`.
- Main task-specific setup impact: CMS coupon validation endpoint must be reachable from `${SERVICE_NAME}`.
- The inspected implementation uses HTTP for CMS coupon validation, not gRPC/protobuf.

### Files Inspected

| File | Why inspected |
|---|---|
| `INPUT_FILE_PATH` | `${TASK_FILE_NAME}` scope: coupon preview, CMS validation, cart totals update |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | Base setup: Go, MongoDB, Redis, Product/CMS env, Docker, run commands |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | MongoDB source of truth plus Redis cache behavior |
| `PREVIOUS_TASK3_DEPENDENCY_FILE` | `cart_db.carts` schema, indexes, migration/bootstrap |
| `PREVIOUS_TASK4_DEPENDENCY_FILE` | Product Service setup and Add Item flow used to create a non-empty cart |
| `PREVIOUS_TASK5_DEPENDENCY_FILE` | Remove Item stale coupon clear behavior |
| `backend/services/cart-service/go.mod` | Confirmed no new Go dependency for coupon preview |
| `backend/services/cart-service/.env.example` | Confirmed CMS env names and defaults |
| `backend/services/cart-service/internal/config/config.go` | Confirmed required env validation and timeout fallback |
| `backend/services/cart-service/cmd/server/main.go` | Confirmed CMS HTTP validator wiring |
| `backend/services/cart-service/internal/client/cms_coupon_client.go` | Confirmed CMS request/response contract |
| `backend/services/cart-service/internal/usecase/apply_coupon_preview.go` | Confirmed runtime behavior, retry, timeout, cache refresh |
| `backend/services/cart-service/internal/domain/coupon.go` | Confirmed coupon code and discount validation rules |
| `backend/services/cart-service/internal/transport/http/handler.go` | Confirmed route, owner headers, and error mapping |
| `backend/services/cart-service/internal/repository/mongo_cart_repository.go` | Confirmed `SaveWithVersion` update fields |
| `backend/services/cart-service/internal/repository/redis_cart_cache.go` | Confirmed reused Redis cache keys |

## 2. Tech Stack

Most technologies are reused from previous dependency files. Full installation steps are intentionally not duplicated.

| Technology | Required? | Status for `${TASK_FILE_NAME}` | What beginner should know | Setup reference |
|---|---:|---|---|---|
| Go `1.26.3` | Yes | Reused | Backend service Go me implemented hai. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4. Go Dependency System` |
| Go modules | Yes | Reused | `go.mod` and `go.sum` dependencies manage karte hain. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4. Go Dependency System` |
| Standard `net/http` | Yes | Reused, task-specific use | Coupon preview public route and CMS client HTTP use karte hain. | Base explained in `PREVIOUS_TASK1_DEPENDENCY_FILE`; CMS contract explained here |
| MongoDB | Yes | Reused | Active cart durable source of truth hai. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `5.1 MongoDB` |
| MongoDB schema/indexes | Yes | Reused | Coupon fields already `cart_db.carts` schema me exist karte hain. | `PREVIOUS_TASK3_DEPENDENCY_FILE`, section `5. Database Setup` |
| Redis | Yes | Reused | Coupon preview ke baad active cart and summary cache refresh hota hai. | `PREVIOUS_TASK2_DEPENDENCY_FILE`, section `6.2 Redis` |
| CMS Service HTTP API | Yes | Task-specific focus | Coupon rules validate karne ke liye mandatory hai. | Base config in `PREVIOUS_TASK1_DEPENDENCY_FILE`; exact contract in this file |
| Product Service HTTP API | Needed to create carts through Add Item | Reused | Coupon preview existing non-empty cart par run hota hai; Add Item Product Service use karta hai. | `PREVIOUS_TASK4_DEPENDENCY_FILE`, section `6.1 Product Service` |
| Docker | Optional but recommended | Reused | Local MongoDB/Redis run karne ke liye useful hai. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `9. Docker and DevOps Setup` |
| MySQL for CMS | Required by real CMS stack, not by this service directly | External to `${SERVICE_NAME}` | CMS coupon rules MySQL me ho sakte hain, but `${SERVICE_NAME}` direct MySQL connect nahi karta. | CMS Service docs/future implementation |

Simple Hinglish:

- MongoDB cart ka original data rakhta hai.
- Redis fast copy/cache rakhta hai.
- CMS Service coupon rules ka owner hai.
- `${SERVICE_NAME}` discount formula khud calculate nahi karta. CMS final preview discount amount return karta hai.
- Frontend se discount amount kabhi trust nahi hota.

## 3. Required Software

Base software setup already documented hai. Yahan sirf incremental view diya gaya hai.

| Software / Service | Required? | Status | Notes |
|---|---:|---|---|
| Git | Yes | Reused | Clone/setup steps previous docs me hain. |
| Go `1.26.3` | Yes | Reused | Service build/run/test ke liye. |
| MongoDB | Yes | Reused | Active cart load/save ke liye mandatory. |
| Redis | Yes | Reused | Startup ping and cache refresh ke liye mandatory in current service. |
| CMS Service | Yes for `${TASK_FILE_NAME}` | Task-specific focus | Must expose coupon validation HTTP endpoint. |
| Product Service | Needed for Add Item based setup | Reused | Non-empty cart create karne ke liye useful. |
| `mongosh` | Recommended | Reused | Active cart seed/verify karne ke liye useful. |
| `redis-cli` | Recommended | Reused | Cache key verify karne ke liye useful. |
| `curl` | Recommended | Reused | Health and coupon preview API verify karne ke liye. |
| Docker / Docker Compose | Optional | Reused | MongoDB/Redis local setup easy banata hai. |

Current repo snapshot note:

- `backend/services/cms-service` me runnable CMS source file nahi mila; only `.env` file present hai.
- Full end-to-end coupon validation ke liye real CMS Service kisi actual implementation repo/module se run karo.
- Local smoke testing ke liye temporary mock CMS use kar sakte ho, but production me real CMS Service hi use hona chahiye.

## 4. Dependency Management

This is a Go project. `${TASK_FILE_NAME}` does not add any new Go module dependency.

Current direct dependencies remain unchanged:

| Dependency | Version | Status | Why used |
|---|---:|---|---|
| `go.mongodb.org/mongo-driver` | `v1.17.9` | Reused | MongoDB active cart lookup and save |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Reused | Redis active cart and summary cache refresh |

CMS coupon validation uses Go standard library:

```text
net/http
encoding/json
context
time
```

No `go get` is needed for `${TASK_FILE_NAME}`.

Important beginner note:

- `${INPUT_FILE_PATH}` discusses future/internal gRPC ideas, but inspected runtime code uses HTTP.
- Do not add `google.golang.org/grpc` only for this task.
- Do not generate protobuf files only for this task.
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

Coupon preview database behavior:

- Active cart must already exist.
- Cart must be active and non-empty.
- Lookup is scoped by owner: `user_id` or `guest_session_id`.
- Optional `cart_id` is checked against the loaded active cart.
- Valid coupon saves `coupon_code`, `coupon_preview`, and discounted totals.
- Invalid coupon saves attempted `coupon_code`, invalid `coupon_preview`, zero discount, and reason.
- CMS unavailable does not mutate cart.
- Bad CMS discount result does not mutate cart.
- Save uses `_id`, `status: active`, and `version` guard.
- Version conflict retries up to `CART_MAX_SAVE_ATTEMPTS`.

### MongoDB Fields Reused

These fields were already documented in previous tasks and are reused here:

| Field | Purpose in `${TASK_FILE_NAME}` | Setup status |
|---|---|---|
| `coupon_code` | Latest attempted/previewed coupon code | Reused from schema |
| `coupon_preview.coupon_id` | CMS coupon ID for valid previews | Reused from schema |
| `coupon_preview.code` | Normalized code stored with preview | Reused from schema |
| `coupon_preview.valid` | Valid/invalid CMS decision | Reused from schema |
| `coupon_preview.discount` | Preview discount in minor units | Reused from schema |
| `coupon_preview.reason` | Invalid reason such as `COUPON_EXPIRED` | Reused from schema |
| `totals.discount` | Discount applied to cart total | Reused from schema |
| `version` | Optimistic concurrency guard | Reused from schema |

### Task-Specific MongoDB Verification

After coupon preview, inspect active cart:

```bash
mongosh "mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin" \
  --quiet \
  --eval "db.getSiblingDB('cart_db').carts.findOne({ user_id: 'user_123', status: 'active' }, { coupon_code: 1, coupon_preview: 1, totals: 1, version: 1, updated_at: 1 })"
```

Expected for valid coupon:

- `coupon_code` is normalized, for example `SAVE10`.
- `coupon_preview.valid` is `true`.
- `coupon_preview.coupon_id` is not empty.
- `coupon_preview.discount.amount` matches CMS response.
- `totals.discount.amount` matches `coupon_preview.discount.amount`.
- `totals.total.amount = totals.subtotal.amount - totals.discount.amount`.
- `version` increments after save.

Expected for invalid coupon:

- `coupon_preview.valid` is `false`.
- `coupon_preview.reason` explains the business reason.
- `coupon_preview.discount.amount` is `0`.
- `totals.discount.amount` is `0`.
- Old valid discount is not kept.

## 6. Redis / Queue / External Services

### 6.1 CMS Service

Base CMS Service setup was already introduced earlier.

Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 6.2 CMS Service
Section: 8. Complete Environment Variables
```

`${TASK_FILE_NAME}` is different because coupon preview now actively calls CMS at runtime.

Simple Hinglish:

CMS Service coupon rules ka owner hai. `${SERVICE_NAME}` cart ka trusted snapshot bhejta hai. CMS check karta hai ki coupon active hai ya nahi, min cart amount match ho raha hai ya nahi, product/seller scope valid hai ya nahi, usage limit/time window valid hai ya nahi, and final preview discount amount return karta hai.

#### Required CMS Endpoint

Current implementation calls:

```text
POST {CART_CMS_BASE_URL}{CART_CMS_VALIDATE_COUPON_PATH}
```

Default local values:

```env
CART_CMS_BASE_URL=http://localhost:8089
CART_CMS_VALIDATE_COUPON_PATH=/internal/v1/coupons/validate
```

Important:

- `CART_CMS_BASE_URL` must include `http://` or `https://`.
- Do not put the path inside `CART_CMS_BASE_URL`; path is appended separately.
- Host-run service can use `http://localhost:8089`.
- Docker Compose service should use service DNS, for example `http://cms-service:8089`.

#### Request Sent to CMS

Current HTTP client sends JSON like this:

```json
{
  "coupon_code": "SAVE10",
  "code": "SAVE10",
  "user_id": "user_123",
  "guest_session_id": "",
  "cart_id": "cart_123",
  "currency": "INR",
  "subtotal": {
    "amount": 2000,
    "currency": "INR"
  },
  "items": [
    {
      "product_id": "prod_123",
      "variant_id": "var_1",
      "seller_id": "seller_456",
      "unit_price": {
        "amount": 1000,
        "currency": "INR"
      },
      "quantity": 2,
      "line_subtotal": {
        "amount": 2000,
        "currency": "INR"
      }
    }
  ],
  "requested_at": "2026-05-24T10:00:00Z"
}
```

Header sent by implementation:

```text
Accept: application/json
Content-Type: application/json
X-Internal-Service: cart-service
```

Security note:

- `X-Internal-Service` is an identity hint, not strong authentication by itself.
- Production should protect this endpoint through private networking, service auth, mTLS, signed internal tokens, or gateway policy.

#### CMS Response Shape

CMS can return direct response:

```json
{
  "valid": true,
  "coupon_id": "coupon_save10",
  "code": "SAVE10",
  "discount": {
    "amount": 200,
    "currency": "INR"
  },
  "reason": ""
}
```

The client also accepts nested envelope shapes:

```json
{
  "data": {
    "valid": false,
    "coupon_id": "",
    "discount": {
      "amount": 0,
      "currency": "INR"
    },
    "reason": "COUPON_EXPIRED"
  }
}
```

Accepted wrapper keys:

```text
data
result
preview
```

Required response rules:

| Rule | Why |
|---|---|
| `valid` must be present | Client rejects response if valid flag missing |
| Valid coupon must include non-empty `coupon_id` | Cart domain rejects valid preview without CMS coupon ID |
| Valid discount currency must match cart currency | Prevents corrupt totals |
| Valid discount amount cannot exceed cart subtotal | Prevents negative payable total |
| Invalid coupon should include `reason` | Better buyer-facing and debugging behavior |
| CMS should return non-2xx only for technical failures | Business-invalid coupon should be HTTP `200` with `valid=false` |

#### Verify Real CMS

Use this after real CMS Service is running:

```bash
curl -i -X POST http://localhost:8089/internal/v1/coupons/validate \
  -H "Content-Type: application/json" \
  -H "X-Internal-Service: cart-service" \
  -d '{
    "coupon_code": "SAVE10",
    "code": "SAVE10",
    "user_id": "user_123",
    "cart_id": "cart_123",
    "currency": "INR",
    "subtotal": {"amount": 2000, "currency": "INR"},
    "items": [
      {
        "product_id": "prod_123",
        "variant_id": "var_1",
        "seller_id": "seller_456",
        "unit_price": {"amount": 1000, "currency": "INR"},
        "quantity": 2,
        "line_subtotal": {"amount": 2000, "currency": "INR"}
      }
    ],
    "requested_at": "2026-05-24T10:00:00Z"
  }'
```

Expected:

- HTTP `2xx`.
- JSON has `valid`.
- If `valid=true`, discount currency is `INR` and amount is not greater than `2000`.

#### Optional Local Mock CMS for Smoke Testing

Use only if real CMS is not available locally.

```bash
python3 - <<'PY'
from http.server import BaseHTTPRequestHandler, HTTPServer
import json

class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path != "/internal/v1/coupons/validate":
            self.send_response(404)
            self.end_headers()
            return
        length = int(self.headers.get("content-length", "0"))
        payload = json.loads(self.rfile.read(length) or b"{}")
        code = (payload.get("coupon_code") or payload.get("code") or "").strip().upper()
        if code == "SAVE10":
            body = {
                "valid": True,
                "coupon_id": "coupon_save10",
                "code": "SAVE10",
                "discount": {"amount": 200, "currency": payload.get("currency", "INR")},
                "reason": ""
            }
        else:
            body = {
                "valid": False,
                "coupon_id": "",
                "code": code,
                "discount": {"amount": 0, "currency": payload.get("currency", "INR")},
                "reason": "COUPON_NOT_FOUND"
            }
        data = json.dumps(body).encode()
        self.send_response(200)
        self.send_header("content-type", "application/json")
        self.send_header("content-length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

HTTPServer(("0.0.0.0", 8089), Handler).serve_forever()
PY
```

### 6.2 Redis

Redis setup is reused. `${TASK_FILE_NAME}` does not introduce new Redis key patterns.

Refer:

```text
PREVIOUS_TASK2_DEPENDENCY_FILE
Section: 6.2 Redis
```

Coupon preview refreshes the same keys:

| Key Pattern | Purpose | Status |
|---|---|---|
| `cart:active:user:<user_id>` | Active user cart cache | Reused |
| `cart:active:guest:<guest_session_id>` | Active guest cart cache | Reused |
| `cart:summary:user:<user_id>` | User cart summary cache | Reused |
| `cart:summary:guest:<guest_session_id>` | Guest cart summary cache | Reused |

Important behavior:

- MongoDB save failure fails coupon preview.
- CMS unavailable fails coupon preview and cart is not mutated.
- Redis cache refresh failure is logged as warning, but successful MongoDB save still returns preview.

### 6.3 Queues and Other Services

Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, OAuth provider setup is not required by `${TASK_FILE_NAME}`.

## 7. Environment Variables

Full `.env` setup is already documented.

Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 8. Complete Environment Variables
```

`${TASK_FILE_NAME}` does not add brand-new env variable names compared with previous dependency docs. It makes the CMS coupon variables operationally important because coupon preview now calls CMS at runtime.

### Task-Specific Env Snippet

Place these in:

```text
backend/services/cart-service/.env
```

```env
CART_CMS_BASE_URL=http://localhost:8089
CART_CMS_VALIDATE_COUPON_PATH=/internal/v1/coupons/validate
CART_COUPON_PREVIEW_TIMEOUT=700ms
```

Keep existing MongoDB, Redis, Product Service, HTTP, TTL, and cache env from previous setup.

### Variable Details

| Variable | Required? | Example | Purpose | Security note |
|---|---:|---|---|---|
| `CART_CMS_BASE_URL` | Yes if `CMS_SERVICE_BASE_URL` absent | `http://localhost:8089` | CMS coupon validation base URL | Keep internal; do not expose private CMS URL publicly |
| `CMS_SERVICE_BASE_URL` | Fallback only | `http://localhost:8089` | Legacy/shared fallback if `CART_CMS_BASE_URL` missing | Prefer service-specific env |
| `CART_CMS_VALIDATE_COUPON_PATH` | Yes | `/internal/v1/coupons/validate` | Path appended to CMS base URL | Not secret |
| `CART_COUPON_PREVIEW_TIMEOUT` | Optional but recommended | `700ms` | Timeout used by coupon preview usecase and CMS HTTP client | Too high can slow cart page; too low can cause false failures |
| `CART_CMS_REQUEST_TIMEOUT` | Fallback only | `700ms` | Legacy timeout fallback if preview timeout missing | Prefer `CART_COUPON_PREVIEW_TIMEOUT` |

How env loads:

- Current Go code uses `os.Getenv`.
- There is no automatic `.env` loader.
- Creating `.env` is not enough; source/export it before `go run`.

Linux/macOS:

```bash
cd backend/services/cart-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Common env mistakes for `${TASK_FILE_NAME}`:

- `CART_CMS_BASE_URL=localhost:8089`: invalid because scheme is missing.
- `CART_CMS_BASE_URL=http://localhost:8089/internal/v1/coupons/validate`: wrong because client appends path again.
- `CART_CMS_VALIDATE_COUPON_PATH=internal/v1/coupons/validate`: accepted by client constructor, but keep leading `/` for clarity.
- `${SERVICE_NAME}` runs inside Docker but env points to host-only `localhost`.
- `CART_COUPON_PREVIEW_TIMEOUT=700`: invalid duration format; use `700ms`.
- Real CMS requires auth/network policy but local env points to public URL.

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
| `${SERVICE_NAME}` Dockerfile | Not present in inspected setup |
| CMS Service container | Required only if you run real CMS in Docker |
| CMS MySQL container | Belongs to CMS stack, not `${SERVICE_NAME}` stack |
| New volume | Not added |
| New network | Not added |
| New port mapping | Not added by this task |

Docker networking reminder:

| Where `${SERVICE_NAME}` runs | CMS base URL example |
|---|---|
| Host machine with `go run` | `CART_CMS_BASE_URL=http://localhost:8089` |
| Docker Compose container | `CART_CMS_BASE_URL=http://cms-service:8089` |
| Kubernetes | `CART_CMS_BASE_URL=http://cms-service.core.svc.cluster.local:8089` or platform service DNS |

Do not put CMS DB credentials in `${SERVICE_NAME}` env. `${SERVICE_NAME}` only needs CMS HTTP URL/path.

## 9. Local Development Setup

### Step 1: Read Previous Dependency Docs First

Follow these before this file:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
PREVIOUS_TASK2_DEPENDENCY_FILE
PREVIOUS_TASK3_DEPENDENCY_FILE
PREVIOUS_TASK4_DEPENDENCY_FILE
PREVIOUS_TASK5_DEPENDENCY_FILE
```

Why:

- `PREVIOUS_TASK1_DEPENDENCY_FILE` has base Go, MongoDB, Redis, Docker, env, run commands.
- `PREVIOUS_TASK2_DEPENDENCY_FILE` explains MongoDB vs Redis responsibility.
- `PREVIOUS_TASK3_DEPENDENCY_FILE` explains `cart_db.carts` schema and indexes.
- `PREVIOUS_TASK4_DEPENDENCY_FILE` explains Add Item and Product Service setup.
- `PREVIOUS_TASK5_DEPENDENCY_FILE` explains stale coupon clearing on item removal.
- `${TASK_FILE_NAME}` only adds the actual CMS coupon preview verification surface.

### Step 2: Go to Service Directory

```bash
cd backend/services/cart-service
```

### Step 3: Install Dependencies

No new dependency install is needed for `${TASK_FILE_NAME}`.

If this is a fresh clone:

```bash
go mod download
```

### Step 4: Start Reused Databases

Start MongoDB and Redis using previous dependency docs.

Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 5.1 MongoDB
Section: 5.2 Redis
```

Apply MongoDB schema/indexes if not already done:

```bash
curl -X POST http://localhost:8084/internal/v1/cart/schema/bootstrap
```

If service is not running yet, use the migration file from previous docs:

```text
backend/services/cart-service/migrations/mongo/001_create_carts_collection.up.js
```

### Step 5: Make CMS Service Reachable

Preferred:

- Run real CMS Service on `http://localhost:8089`.
- Ensure it handles `POST /internal/v1/coupons/validate`.
- Seed CMS coupon data such as `SAVE10`.

If real CMS is not available in this repo snapshot:

- Use the temporary mock from section `6.1 CMS Service`.
- Keep this mock only for local smoke testing.

Verify CMS:

```bash
curl -i -X POST http://localhost:8089/internal/v1/coupons/validate \
  -H "Content-Type: application/json" \
  -d '{"coupon_code":"SAVE10","code":"SAVE10","cart_id":"cart_123","currency":"INR","subtotal":{"amount":2000,"currency":"INR"},"items":[]}'
```

### Step 6: Add Task-Specific Env

In `backend/services/cart-service/.env`, make sure these values exist:

```env
CART_CMS_BASE_URL=http://localhost:8089
CART_CMS_VALIDATE_COUPON_PATH=/internal/v1/coupons/validate
CART_COUPON_PREVIEW_TIMEOUT=700ms
```

Keep existing required values:

```text
CART_MONGO_URI
CART_MONGO_DATABASE
CART_REDIS_ADDR
CART_PRODUCT_BASE_URL
CART_HTTP_ADDR
```

### Step 7: Source Env

```bash
cd backend/services/cart-service
set -a
. ./.env
set +a
```

### Step 8: Ensure a Non-Empty Active Cart Exists

Preferred flow:

```text
Use Add Item from PREVIOUS_TASK4_DEPENDENCY_FILE.
```

This needs Product Service or a Product Service mock.

Local-only seed option:

```bash
mongosh "mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin" <<'JS'
const now = new Date();
const expires = new Date(now.getTime() + 90 * 24 * 60 * 60 * 1000);
db = db.getSiblingDB("cart_db");
db.carts.deleteMany({ user_id: "user_123", status: "active" });
db.carts.insertOne({
  _id: "cart_123",
  user_id: "user_123",
  status: "active",
  items: [
    {
      item_id: "item_1",
      product_id: "prod_123",
      variant_id: "var_1",
      seller_id: "seller_456",
      title_snapshot: "Local Test Product",
      unit_price: { amount: NumberLong(1000), currency: "INR" },
      quantity: 2,
      line_subtotal: { amount: NumberLong(2000), currency: "INR" },
      price_snapshot_at: now,
      added_at: now,
      updated_at: now
    }
  ],
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
});
JS
```

Warning:

- Manual seed is only for local development.
- Real app flow should create carts through service APIs so domain validation is not bypassed.

### Step 9: Start Backend Service

```bash
cd backend/services/cart-service
set -a
. ./.env
set +a
go run ./cmd/server
```

### Step 10: Verify Health and Readiness

```bash
curl http://localhost:8084/healthz
curl http://localhost:8084/readyz
```

Expected:

- `/healthz` returns process health.
- `/readyz` returns ready if MongoDB is reachable.

### Step 11: Verify `${TASK_FILE_NAME}` API

Valid coupon:

```bash
curl -X POST http://localhost:8084/api/v1/cart/coupons/preview \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_123" \
  -d '{"coupon_code":" save10 "}'
```

Expected:

```json
{
  "valid": true,
  "coupon_id": "coupon_save10",
  "discount": {
    "amount": 200,
    "currency": "INR"
  },
  "reason": ""
}
```

Invalid coupon:

```bash
curl -X POST http://localhost:8084/api/v1/cart/coupons/preview \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_123" \
  -d '{"coupon_code":"NOPE"}'
```

Expected:

- HTTP `200`.
- `valid=false`.
- `reason` explains why coupon is invalid.
- Cart discount is zero.

Guest example:

```bash
curl -X POST http://localhost:8084/api/v1/cart/coupons/preview \
  -H "Content-Type: application/json" \
  -H "X-Guest-Session-ID: guest_sess_123" \
  -d '{"coupon_code":"SAVE10"}'
```

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
4. Start or mock Product Service if you need Add Item to create cart.
5. Start or mock CMS Service.
6. Configure `CART_CMS_BASE_URL`, `CART_CMS_VALIDATE_COUPON_PATH`, and timeout.
7. Source `.env`.
8. Run `${SERVICE_NAME}` HTTP server.
9. Create or seed a non-empty active cart.
10. Call `POST /api/v1/cart/coupons/preview`.
11. Verify response, MongoDB document, and Redis cache/logs.

### Ports and Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| `${SERVICE_NAME}` HTTP API | `8084` | Coupon preview route, health, readiness | Reused |
| MongoDB | `27017` | Durable `cart_db.carts` store | Reused |
| Redis | `6379` | Active cart and summary cache | Reused |
| Product Service | `8082` | Add Item product lookup for creating carts | Reused |
| CMS Service | `8089` | Coupon validation endpoint | Task-specific focus |
| CMS MySQL | Usually `3306` | CMS coupon storage, not directly used by `${SERVICE_NAME}` | External CMS stack |
| Kafka / RabbitMQ | N/A | Not used by coupon preview | Not required |

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
| `/api/v1/cart/coupons/preview` | `POST` | Apply coupon preview to active cart |
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

Generic setup errors are already documented in previous files. This section only covers `${TASK_FILE_NAME}`-specific or coupon-preview-specific errors.

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `CART_CMS_BASE_URL or CMS_SERVICE_BASE_URL is required` | CMS URL missing in env | Add `CART_CMS_BASE_URL=http://localhost:8089` | Keep `.env.example` copied completely |
| `CART_CMS_BASE_URL is invalid` | URL missing scheme or malformed | Use `http://localhost:8089` | Always include `http://` or `https://` |
| `CART_CMS_BASE_URL must use http or https` | Unsupported URL scheme | Use HTTP/HTTPS service URL | Do not use custom schemes |
| `CART_CMS_VALIDATE_COUPON_PATH is required` | Coupon path missing | Set `/internal/v1/coupons/validate` | Keep path separate from base URL |
| `CART_COUPON_PREVIEW_TIMEOUT must be greater than zero` | Timeout env parsed as zero/negative | Use `700ms` or `1s` | Use Go duration format |
| `COUPON_CODE_REQUIRED` | Empty coupon code | Send non-empty `coupon_code` | Trim input in frontend before submit |
| `COUPON_CODE_TOO_LONG` | Coupon code exceeds 64 characters | Use valid coupon code length | Add frontend validation |
| `EMPTY_CART` | Cart has no items or subtotal is zero | Add item before applying coupon | Coupon preview only makes sense for non-empty cart |
| `CART_OWNER_REQUIRED` | No trusted owner header | Send `X-User-ID` or guest session header | API Gateway should inject owner context |
| `CART_NOT_FOUND` | Active cart does not exist for owner | Create cart through Add Item or seed local cart | Preview does not create cart |
| `CART_OWNERSHIP_MISMATCH` | Optional `cart_id` does not match owner active cart | Send correct cart id or omit it | Frontend should use cart id from latest cart response |
| `COUPON_PREVIEW_UNAVAILABLE` | CMS down, CMS timeout, non-2xx response, invalid JSON, or missing `valid` flag | Start/fix CMS and verify endpoint directly | Add CMS health check/smoke test before local preview |
| `COUPON_PREVIEW_INVALID` | CMS returned unsafe result: currency mismatch, too large discount, missing valid coupon id | Fix CMS response data | Add CMS contract tests |
| `CART_VERSION_CONFLICT` | Concurrent cart mutations exceeded retry attempts | Retry request | Keep `CART_MAX_SAVE_ATTEMPTS=3` unless measured need says otherwise |
| Valid CMS response but cart not discounted | CMS discount amount is zero or invalid coupon branch used | Inspect CMS response and MongoDB `coupon_preview` | Contract-test CMS response |
| Invalid coupon returns HTTP `503` | CMS returned non-2xx for business invalid coupon | CMS should return HTTP `200` with `valid=false` | Separate business invalid from technical failure |
| Server starts but `/readyz` ignores CMS | Readiness currently checks MongoDB only | Verify CMS separately with curl | Add CMS readiness in future if production needs it |
| Task guide mentions gRPC but code has HTTP | Documentation/design vs current implementation mismatch | Use current HTTP env and client | Do not add gRPC dependency until code changes |
| Coupon preview succeeds but Redis key stale/missing | Cache refresh failed after Mongo save | Check logs for `cart.coupon_preview.cache_*_failed`, verify Redis | Monitor Redis and keep it reachable |

## 12. Security & Best Practices

Task-specific best practices:

- Do not trust frontend discount amount. Frontend sends only coupon code.
- Keep CMS coupon validation endpoint internal; do not expose it directly to browsers.
- Protect service-to-service CMS calls with private network and real service authentication in production.
- Keep `CART_COUPON_PREVIEW_TIMEOUT` short. Cart page should not hang because CMS is slow.
- Treat CMS non-2xx as technical unavailable, but business-invalid coupons should return `valid=false`.
- Do not connect `${SERVICE_NAME}` directly to CMS MySQL. CMS owns coupon rules and redemptions.
- Validate CMS result defensively. Current code rejects currency mismatch, discount greater than subtotal, and missing valid coupon ID.
- Keep checkout final coupon validation in Order Service. Preview is not a payment guarantee.
- Rate-limit coupon preview at API Gateway. Coupon endpoints are easy to brute-force.
- Do not log sensitive user/session data with full payloads. Log coupon result and reason carefully.
- Keep owner headers trusted. Public clients should not be able to spoof `X-User-ID`.
- Keep MongoDB source of truth. Redis is only cache.
- Clear or revalidate coupon preview when cart contents change. Previous tasks already clear stale coupon preview on add/remove/merge.
- Use integer minor units for money. Do not use floats for discount calculation.

## 13. Missing or Misconfigured Things

| Finding | Impact | Suggested fix |
|---|---|---|
| `backend/services/cms-service` has no runnable source in this repo snapshot | Beginners cannot perform full real CMS coupon validation locally | Add runnable CMS implementation, provide external repo/start command, or include local CMS mock for dev |
| `${INPUT_FILE_PATH}` mentions gRPC/protobuf, but current implementation uses HTTP | New developer may install unnecessary gRPC deps | Keep docs aligned with code; add gRPC only when runtime changes |
| CMS endpoint auth is only an internal header in current client | Header can be spoofed if endpoint is public | Use private network, mTLS, signed service token, or gateway/service mesh policy |
| `/readyz` checks MongoDB through schema readiness but not CMS or Redis after startup | Service may appear ready while CMS is down | Add dependency readiness or separate smoke checks if production requires it |
| No dedicated `${SERVICE_NAME}` Dockerfile found in inspected setup | Full service container build flow is incomplete | Add Dockerfile in future DevOps task |
| No complete local compose stack found for all services | Beginners must manually run MongoDB, Redis, Product Service, CMS, and service | Add local compose with MongoDB, Redis, product mock, CMS mock, and `${SERVICE_NAME}` |
| `.env` is not auto-loaded | Service can fail despite `.env` file existing | Source `.env` before `go run`, or add a documented dev runner |
| Product/CMS URLs are required by startup | Testing one route can fail because unrelated URL config is missing | Keep `.env.example` complete or make clients lazy/feature-specific later |
| Manual MongoDB seed can bypass domain validation | Bad local data can cause totals/currency errors | Prefer Add Item flow; use seed only for local smoke tests |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `4. Go Dependency System` | Same Go module, Go version, and dependency workflow |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `5.1 MongoDB` | Same MongoDB installation, Docker setup, URI, migration, and verification |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `5.2 Redis` | Same Redis installation, Docker setup, and verification |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `6.2 CMS Service` | Base CMS Service purpose and env names already documented |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `8. Complete Environment Variables` | Full `.env` location and complete env list already documented |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `9. Docker and DevOps Setup` | No new Docker setup added by `${TASK_FILE_NAME}` |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `10. Project Run Instructions` | Base run flow reused |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `12. Common Errors and Fixes` | Generic Go/MongoDB/Redis/env troubleshooting reused |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | `6.2 Redis` | Redis key patterns and TTL behavior reused |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | `7. Environment Variables` | Cache and DB env variable behavior reused |
| `PREVIOUS_TASK3_DEPENDENCY_FILE` | `5. Database Setup` | `cart_db.carts` schema, coupon fields, and indexes reused |
| `PREVIOUS_TASK3_DEPENDENCY_FILE` | `9. Local Development Setup` | Schema/bootstrap onboarding reused |
| `PREVIOUS_TASK4_DEPENDENCY_FILE` | `6.1 Product Service` | Add Item setup reused when creating a non-empty cart for preview |
| `PREVIOUS_TASK4_DEPENDENCY_FILE` | `11. Common Errors & Fixes` | Product/Add Item setup errors reused for cart creation |
| `PREVIOUS_TASK5_DEPENDENCY_FILE` | `11. Common Errors & Fixes` | Owner/cart mutation errors reused |
| `PREVIOUS_TASK5_DEPENDENCY_FILE` | `12. Security & Best Practices` | Stale coupon clearing after cart mutation reused |

## 15. Final Checklist

- [ ] Previous dependency documentation checked.
- [ ] No duplicate Go/MongoDB/Redis/Docker installation docs added.
- [ ] No unnecessary gRPC/protobuf dependency added for current HTTP implementation.
- [ ] MongoDB running and reachable.
- [ ] Redis running and reachable.
- [ ] `cart_db.carts` schema/indexes applied.
- [ ] `.env` created from previous setup and sourced before `go run`.
- [ ] `CART_CMS_BASE_URL` configured with `http://` or `https://`.
- [ ] `CART_CMS_VALIDATE_COUPON_PATH` configured.
- [ ] `CART_COUPON_PREVIEW_TIMEOUT` configured or default accepted.
- [ ] Real CMS Service or local mock reachable at coupon validation endpoint.
- [ ] CMS returns `valid` flag in response.
- [ ] Valid CMS response includes `coupon_id`, discount amount, and matching currency.
- [ ] Non-empty active cart exists before preview test.
- [ ] `${SERVICE_NAME}` starts on expected `CART_HTTP_ADDR`.
- [ ] `/healthz` and `/readyz` checked.
- [ ] `POST /api/v1/cart/coupons/preview` verified for valid coupon.
- [ ] `POST /api/v1/cart/coupons/preview` verified for invalid coupon.
- [ ] MongoDB `coupon_code`, `coupon_preview`, `totals.discount`, and `version` checked after preview.
- [ ] Redis cache refresh checked through logs or keys.
- [ ] CMS unavailable behavior checked if needed.
- [ ] No real secrets committed.
- [ ] CMS validation endpoint kept internal/private.
