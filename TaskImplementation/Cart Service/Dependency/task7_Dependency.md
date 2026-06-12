# Project Dependency & Setup Guide

```text
SERVICE_NAME = "Cart Service"
TASK_FILE_NAME = "task7.md"
INPUT_FILE_PATH = TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}
OUTPUT_FILE_NAME = ${TASK_FILE_NAME without ".md"}_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}

PREVIOUS_TASK1_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
PREVIOUS_TASK2_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
PREVIOUS_TASK3_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task3_Dependency.md
PREVIOUS_TASK4_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task4_Dependency.md
PREVIOUS_TASK5_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task5_Dependency.md
PREVIOUS_TASK6_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task6_Dependency.md
```

## 1. Project Overview

Ye document `${SERVICE_NAME}` ke `${TASK_FILE_NAME}` ke liye dependency, setup, environment, database, Docker, DevOps, aur troubleshooting guide hai.

`${TASK_FILE_NAME}` ka focus Cart Merge flow hai:

- Buyer login ke baad frontend/API Gateway `POST /api/v1/cart/merge` call karta hai.
- `${SERVICE_NAME}` logged-in `user_id` and guest cart identity receive karta hai.
- Guest cart `guest_cart_id` se load hota hai.
- Trusted guest session header se ownership verify hoti hai.
- Existing user active cart milta hai to guest items usme merge hote hain.
- User active cart nahi mila to new user cart create hota hai.
- Same `product_id + variant_id` duplicate ho to quantity combine hoti hai.
- Quantity max limit cross ho to quantity cap hoti hai and warning log emit hota hai.
- Guest cart `status = merged` ho jata hai.
- Guest cart me `merged_into_cart_id` set hota hai.
- User cart active rehta hai and totals recalculate hote hain.
- Stale coupon preview clear hota hai.
- MongoDB source of truth update hota hai.
- Redis user cache refresh hota hai and guest cache invalidate hota hai.

Important scope:

- `${INPUT_FILE_PATH}` modify nahi kiya gaya.
- `${OUTPUT_FILE_PATH}` new dependency/setup guide hai.
- No new Go module dependency was introduced by `${TASK_FILE_NAME}`.
- No new MongoDB collection, Redis key family, queue, Dockerfile, Docker network, service port, or third-party service was introduced by `${TASK_FILE_NAME}`.
- Main task-specific setup impact: API Gateway/Auth/Session layer must pass trusted owner headers to `${SERVICE_NAME}`.
- Main DevOps nuance: merge uses MongoDB transactions, so local MongoDB must support transactions for full task verification.
- Product Service and CMS Service are not called by merge, but current server config still validates their URLs at startup.

### Files Inspected

| File | Why inspected |
|---|---|
| `INPUT_FILE_PATH` | `${TASK_FILE_NAME}` scope: guest cart merge after login |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | Base setup: Go, MongoDB, Redis, Product/CMS env, Docker, run commands |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | MongoDB source of truth and Redis cache responsibility |
| `PREVIOUS_TASK3_DEPENDENCY_FILE` | `cart_db.carts` schema, indexes, statuses, migration/bootstrap |
| `PREVIOUS_TASK4_DEPENDENCY_FILE` | Add Item setup used to create active guest/user carts |
| `PREVIOUS_TASK5_DEPENDENCY_FILE` | Existing mutation retry/cache/stale coupon behavior |
| `PREVIOUS_TASK6_DEPENDENCY_FILE` | Coupon preview fields cleared by cart mutations |
| `backend/services/cart-service/go.mod` | Confirmed no new Go dependency for merge |
| `backend/services/cart-service/.env.example` | Confirmed merge-related env names and defaults |
| `backend/services/cart-service/internal/config/config.go` | Confirmed `CART_ALLOW_GUEST_CART_ID_ONLY_MERGE`, TTL, retry config |
| `backend/services/cart-service/cmd/server/main.go` | Confirmed merge usecase wiring and startup dependencies |
| `backend/services/cart-service/internal/usecase/merge_guest_cart.go` | Confirmed runtime validation, transaction call, retry, cache refresh |
| `backend/services/cart-service/internal/domain/cart_merge.go` | Confirmed duplicate item merge and quantity cap warnings |
| `backend/services/cart-service/internal/repository/mongo_cart_repository.go` | Confirmed MongoDB transaction and version-guarded updates |
| `backend/services/cart-service/internal/repository/redis_cart_cache.go` | Confirmed reused Redis cache keys |
| `backend/services/cart-service/internal/transport/http/handler.go` | Confirmed route, headers, request body, and error mapping |

## 2. Tech Stack

Most technologies are reused from previous dependency files. Full installation steps are intentionally not duplicated.

| Technology | Required? | Status for `${TASK_FILE_NAME}` | What beginner should know | Setup reference |
|---|---:|---|---|---|
| Go `1.26.3` | Yes | Reused | Backend service Go me implemented hai. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4. Go Dependency System` |
| Go modules | Yes | Reused | `go.mod` and `go.sum` dependencies manage karte hain. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4. Go Dependency System` |
| Standard `net/http` | Yes | Reused, task-specific route | Merge route current HTTP server me expose hota hai. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `3. High-Level Tech Stack` |
| MongoDB | Yes | Reused with transaction nuance | Durable cart source of truth hai. Merge me two cart documents transaction me update hote hain. | Base setup in `PREVIOUS_TASK1_DEPENDENCY_FILE`; transaction note in this file |
| MongoDB Go Driver | Yes | Reused | `WithTransaction` use hota hai, so MongoDB deployment transactions support kare. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4.3 Direct Go Libraries` |
| Redis | Yes | Reused | Merge ke baad user cache update and guest cache delete hota hai. | `PREVIOUS_TASK2_DEPENDENCY_FILE`, section `6.2 Redis` |
| Product Service HTTP API | Startup config required | Reused, not called by merge | Server startup Product client config validate karta hai. Merge flow Product Service ko call nahi karta. | `PREVIOUS_TASK4_DEPENDENCY_FILE`, section `6.1 Product Service` |
| CMS Service HTTP API | Startup config required | Reused, not called by merge | Server startup coupon validator config validate karta hai. Merge flow CMS ko call nahi karta. | `PREVIOUS_TASK6_DEPENDENCY_FILE`, section `6.1 CMS Service` |
| Auth Service | Required upstream | External integration | Real login Auth Service karega. `${SERVICE_NAME}` trusted `user_id` header/context expect karta hai. | Auth/Platform docs |
| Session Management / API Gateway | Required upstream | Task-specific integration | Guest session cookie ko trusted `X-Guest-Session-ID` header me convert karna gateway ka kaam hai. | Session/Gateway docs |
| Docker | Optional but recommended | Reused with MongoDB replica-set note | Local MongoDB/Redis run karne ke liye useful hai. | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `9. Docker and DevOps Setup` |

Simple Hinglish:

- MongoDB permanent cart data rakhta hai.
- Redis fast cache hai; source of truth nahi hai.
- Auth/Gateway trusted identity provide karte hain.
- Browser se direct `X-User-ID` trust nahi karna chahiye in production.
- Merge ka sabse important setup piece code dependency nahi, trusted identity forwarding hai.

## 3. Required Software

Base software setup already documented hai. Yahan only task-specific view diya gaya hai.

| Software / Service | Required? | Status | Why needed |
|---|---:|---|---|
| Git | Yes | Reused | Repository clone karne ke liye |
| Go `1.26.3` | Yes | Reused | `${SERVICE_NAME}` build/run/test karne ke liye |
| MongoDB | Yes | Reused, transaction-capable required for merge | Guest and user carts transactionally update karne ke liye |
| Redis | Yes | Reused | Active cart and summary cache update/delete karne ke liye |
| Product Service | URL config required at startup | Reused | Merge does not call it, but server config validates URL |
| CMS Service | URL config required at startup | Reused | Merge does not call it, but server config validates URL |
| Auth Service / API Gateway | Required in real environment | External | `user_id` and guest session trusted context pass karne ke liye |
| `mongosh` | Recommended | Reused | Merge result and transaction support verify karne ke liye |
| `redis-cli` | Recommended | Reused | Guest cache invalidation verify karne ke liye |
| `curl` | Recommended | Reused | `POST /api/v1/cart/merge` verify karne ke liye |
| Docker / Docker Compose | Optional | Reused with new replica-set note | Local MongoDB/Redis easiest way se run karne ke liye |

Refer:
`PREVIOUS_TASK1_DEPENDENCY_FILE`

Sections:
`4. Go Dependency System`, `5.1 MongoDB`, `5.2 Redis`, `8. Complete Environment Variables`, `9. Docker and DevOps Setup`

## 4. Dependency Management

This is a Go project. `${TASK_FILE_NAME}` does not add any new Go module dependency.

### Direct Go Libraries Reused

| Library | Version | Status | Used by merge for |
|---|---:|---|---|
| `go.mongodb.org/mongo-driver` | `v1.17.9` | Reused | Loading carts, MongoDB session/transaction, version-guarded writes |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Reused | Refreshing user cache and deleting guest cache |

No command like `go get ...` is needed for `${TASK_FILE_NAME}`.

Refer for Go setup and commands:

```md
This setup is already explained in:
`PREVIOUS_TASK1_DEPENDENCY_FILE`

Section:
`4. Go Dependency System`
```

If dependencies fail, follow previous troubleshooting first:

```md
Refer:
`PREVIOUS_TASK1_DEPENDENCY_FILE`

Section:
`12. Common Errors and Fixes`
```

## 5. Database Setup

MongoDB setup is mostly reused. No new database, collection, migration file, or index was added by `${TASK_FILE_NAME}`.

Refer:
`PREVIOUS_TASK1_DEPENDENCY_FILE`

Section:
`5.1 MongoDB`

Refer:
`PREVIOUS_TASK3_DEPENDENCY_FILE`

Section:
`5. Database Setup`

### 5.1 Important New MongoDB Requirement for Merge

`${TASK_FILE_NAME}` uses MongoDB transactions through `session.WithTransaction`.

Beginner explanation:

MongoDB transaction ka matlab hai user cart update and guest cart status update ek logical operation ki tarah commit honge. Agar beech me failure aata hai, half-merged state avoid hoti hai.

Important:

- MongoDB transactions require a transaction-capable deployment.
- Managed MongoDB Atlas normally supports this.
- Local MongoDB standalone container may fail for merge.
- Single-node replica set is enough for local development.

If you see an error like this during merge:

```text
Transaction numbers are only allowed on a replica set member or mongos
```

Then base MongoDB is running as standalone. Use the replica-set setup in section `8. Docker Setup`.

### 5.2 Database Usage for `${TASK_FILE_NAME}`

| Area | Value |
|---|---|
| Database | `cart_db` |
| Collection | `carts` |
| Source guest document | `_id = guest_cart_id`, `status = active`, `guest_session_id` present |
| Target user document | Active cart with `user_id`; created if missing |
| Transaction writes | Save/update user cart and mark guest cart as merged |
| Version guards | `version` on both user and guest carts |
| Expiry check | `expires_at` must be in future |
| Coupon behavior | `coupon_code` and `coupon_preview` are cleared on target cart after merge |

### 5.3 MongoDB Fields Used by Merge

| Field | Status | Purpose |
|---|---|---|
| `_id` | Reused | Guest cart lookup and target cart identity |
| `user_id` | Reused | Logged-in user's active cart owner |
| `guest_session_id` | Reused | Guest cart ownership verification |
| `status` | Reused | Source must be `active`; after merge source becomes `merged` |
| `items` | Reused | Guest items copied/combined into user cart |
| `totals` | Reused | Target cart totals recalculated after merge |
| `coupon_code` | Reused | Cleared on target because old coupon preview may be stale |
| `coupon_preview` | Reused | Cleared on target because cart contents changed |
| `version` | Reused | Optimistic concurrency check |
| `expires_at` | Reused | Expired carts are not merged |
| `merged_into_cart_id` | Reused | Guest cart points to final user cart after merge |

### 5.4 Task-Specific MongoDB Verification

After merge, verify guest cart and user cart:

```bash
mongosh "$CART_MONGO_URI" --quiet --eval '
const dbx = db.getSiblingDB("cart_db");
printjson(dbx.carts.findOne(
  { _id: "cart_guest_task7" },
  { status: 1, guest_session_id: 1, merged_into_cart_id: 1, version: 1 }
));
printjson(dbx.carts.findOne(
  { user_id: "user_task7", status: "active" },
  { user_id: 1, status: 1, items: 1, totals: 1, coupon_code: 1, coupon_preview: 1, version: 1 }
));
'
```

Expected:

- Guest cart `status` becomes `merged`.
- Guest cart `merged_into_cart_id` points to the active user cart.
- User cart `status` remains `active`.
- User cart contains guest items.
- If duplicate item existed, quantity is combined up to max `10`.
- User cart `coupon_code` and `coupon_preview` are absent or null after merge.
- Versions increment after transaction commit.

## 6. Redis / Queue / External Services

### 6.1 Redis

Redis setup is reused. `${TASK_FILE_NAME}` does not introduce new Redis server, Redis DB, Redis port, or Redis key family.

Refer:
`PREVIOUS_TASK2_DEPENDENCY_FILE`

Section:
`6.2 Redis`

Merge-specific Redis behavior:

| Redis key | Status after merge |
|---|---|
| `cart:active:user:{user_id}` | Set/refreshed with merged active user cart |
| `cart:summary:user:{user_id}` | Set/refreshed with merged summary |
| `cart:active:guest:{guest_session_id}` | Deleted/invalidated |
| `cart:summary:guest:{guest_session_id}` | Deleted/invalidated |

Important:

- MongoDB transaction success is the source of truth.
- Redis cache refresh/delete is best-effort.
- If Redis delete fails, merge can still return success after MongoDB commit.
- Check logs for `cart.merge.cache_refresh_failed`.

Verify Redis after merge:

```bash
redis-cli EXISTS cart:active:user:user_task7
redis-cli EXISTS cart:summary:user:user_task7
redis-cli EXISTS cart:active:guest:sess_task7
redis-cli EXISTS cart:summary:guest:sess_task7
```

Expected:

- User keys usually return `1`.
- Guest keys should return `0`.

If Redis uses password:

```bash
redis-cli -a "$CART_REDIS_PASSWORD" EXISTS cart:active:user:user_task7
```

### 6.2 Auth Service, Session Service, and API Gateway

`${TASK_FILE_NAME}` depends on trusted identity context, not on a direct Auth/Session SDK.

Real production flow:

| Context | Source | How `${SERVICE_NAME}` receives it |
|---|---|---|
| Logged-in user id | Auth Service / JWT validation | Trusted `X-User-ID` header or equivalent gateway context |
| Guest session id | Session cookie / Session Service | Trusted `X-Guest-Session-ID`, `X-Guest-Session-Id`, or `X-Session-ID` |
| Guest cart id | Request body | JSON field `guest_cart_id` |

Local testing shortcut:

- You can send these headers manually with `curl`.
- Ye only local/dev ke liye hai.
- Production me browser ko direct identity headers set karne ka trust nahi dena chahiye.

### 6.3 Product Service and CMS Service

Merge does not call Product Service or CMS Service.

Still, current server creates Product and CMS clients at startup, so these env values must be configured:

```env
CART_PRODUCT_BASE_URL=http://localhost:8082
CART_CMS_BASE_URL=http://localhost:8089
CART_CMS_VALIDATE_COUPON_PATH=/internal/v1/coupons/validate
```

Full setup already explained in:

```md
Refer:
`PREVIOUS_TASK4_DEPENDENCY_FILE`

Section:
`6.1 Product Service`

Refer:
`PREVIOUS_TASK6_DEPENDENCY_FILE`

Section:
`6.1 CMS Service`
```

### 6.4 Queues and Other Services

Kafka, RabbitMQ, NATS, MinIO, SMTP, Stripe, Twilio, OAuth provider SDK, and Kubernetes are not introduced by `${TASK_FILE_NAME}`.

## 7. Environment Variables

Full `.env` setup is already documented.

Refer:
`PREVIOUS_TASK1_DEPENDENCY_FILE`

Section:
`8. Complete Environment Variables`

`${TASK_FILE_NAME}` does not add brand-new env variable names. It makes one existing env variable operationally important and may require a MongoDB URI value that points to a transaction-capable MongoDB deployment.

### 7.1 Task-Specific Env Snippet

Create or update:

```text
backend/services/cart-service/.env
```

Use previous full `.env`, then verify these values:

```env
CART_MONGO_URI=mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin&replicaSet=rs0
CART_MONGO_DATABASE=cart_db
CART_ALLOW_GUEST_CART_ID_ONLY_MERGE=false
CART_MAX_SAVE_ATTEMPTS=3
CART_DEFAULT_CURRENCY=INR
CART_USER_EXPIRY_TTL=2160h
CART_GUEST_EXPIRY_TTL=720h
CART_ACTIVE_CACHE_TTL=15m
CART_GUEST_ACTIVE_CACHE_TTL=30m
CART_SUMMARY_CACHE_TTL=5m
```

Notes:

- Add `replicaSet=rs0` only when your local MongoDB is running as replica set named `rs0`.
- If you are using managed MongoDB, use the connection string given by that provider.
- Keep all other env variables from previous setup, especially MongoDB, Redis, Product, CMS, and HTTP settings.

### 7.2 Variable Details

| Variable | Required? | Purpose for `${TASK_FILE_NAME}` | Example | Security / setup note |
|---|---:|---|---|---|
| `CART_MONGO_URI` | Yes | Connect to MongoDB that supports merge transaction | `mongodb://...&replicaSet=rs0` | Contains DB password; do not commit real prod URI |
| `CART_MONGO_DATABASE` | Yes | Must remain Cart DB owner database | `cart_db` | Config rejects any other DB |
| `CART_ALLOW_GUEST_CART_ID_ONLY_MERGE` | Optional, default false | If false, merge requires trusted guest session header | `false` | Keep `false` in production; `true` weakens ownership check |
| `CART_MAX_SAVE_ATTEMPTS` | Optional, default 3 | Retries version conflicts during merge | `3` | Do not hide concurrency issues with very high value |
| `CART_DEFAULT_CURRENCY` | Optional, default INR | New empty target cart currency fallback | `INR` | Must be uppercase 3-letter currency |
| `CART_USER_EXPIRY_TTL` | Optional | Target user cart expiry extension | `2160h` | Duration needs unit like `h` |
| `CART_GUEST_EXPIRY_TTL` | Optional | Guest cart TTL policy from previous flows | `720h` | Merge rejects expired guest cart |
| `CART_ACTIVE_CACHE_TTL` | Optional | User active cart cache TTL | `15m` | Not secret |
| `CART_GUEST_ACTIVE_CACHE_TTL` | Optional | Guest active cart cache TTL before merge | `30m` | Not secret |
| `CART_SUMMARY_CACHE_TTL` | Optional | Header/cart badge summary TTL | `5m` | Not secret |

How env loads:

- Current Go code uses `os.Getenv`.
- There is no automatic `.env` loader.
- Creating `.env` is not enough.
- Source/export env before `go run`.

```bash
cd backend/services/cart-service
set -a
. ./.env
set +a
```

Common env mistakes for `${TASK_FILE_NAME}`:

- `CART_ALLOW_GUEST_CART_ID_ONLY_MERGE=false` but request has no guest session header.
- `CART_ALLOW_GUEST_CART_ID_ONLY_MERGE=true` used in production. This is unsafe.
- `CART_MONGO_URI` points to standalone MongoDB while merge needs transaction support.
- `CART_MONGO_URI` has `?authSource=admin?replicaSet=rs0` with two question marks. Use `?authSource=admin&replicaSet=rs0`.
- `CART_DEFAULT_CURRENCY=inr` lowercase. Keep `INR`.
- `CART_MAX_SAVE_ATTEMPTS=0`. Config requires greater than zero.
- Duration values like `2160` without `h`; use `2160h`.

## 8. Docker Setup

Base Docker setup is reused.

Refer:
`PREVIOUS_TASK1_DEPENDENCY_FILE`

Section:
`9. Docker and DevOps Setup`

`${TASK_FILE_NAME}` does not add a Dockerfile, containerized app build, new volume, new network, new queue, exposed port, restart policy, or health check.

### 8.1 Docker Impact Table

| Docker area | Status for `${TASK_FILE_NAME}` |
|---|---|
| `${SERVICE_NAME}` Dockerfile | Not present in inspected setup |
| MongoDB container | Reused, but must support transactions for merge |
| Redis container | Reused |
| Product Service container | Not called by merge, URL config still required |
| CMS Service container | Not called by merge, URL config still required |
| Queue container | Not required |
| New volume | Not added by task |
| New network | Not added by task |

### 8.2 Task-Specific MongoDB Replica Set Option

If your existing local MongoDB already supports transactions, skip this section.

If you are using the previous standalone `mongo:7` container and merge fails with transaction error, run MongoDB as a single-node replica set for local development.

Important:

- Only one container can bind host port `27017`.
- Stop the old local MongoDB container first if it is already using port `27017`.
- Use a separate volume if you do not want to mix old local data.

Example local setup:

```bash
docker volume create ecommerce_mongo_rs_data

docker run -d \
  --name ecommerce-mongo-rs \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=ecommerce_root \
  -e MONGO_INITDB_ROOT_PASSWORD=ecommerce_password \
  -v ecommerce_mongo_rs_data:/data/db \
  mongo:7 \
  --replSet rs0 \
  --bind_ip_all
```

Initiate replica set:

```bash
docker exec ecommerce-mongo-rs mongosh --quiet \
  -u ecommerce_root \
  -p ecommerce_password \
  --authenticationDatabase admin \
  --eval 'rs.initiate({_id:"rs0",members:[{_id:0,host:"localhost:27017"}]})'
```

Verify:

```bash
docker exec ecommerce-mongo-rs mongosh --quiet \
  -u ecommerce_root \
  -p ecommerce_password \
  --authenticationDatabase admin \
  --eval 'rs.status().ok'
```

Expected:

```text
1
```

Then use:

```env
CART_MONGO_URI=mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin&replicaSet=rs0
```

Docker networking reminder:

| Where `${SERVICE_NAME}` runs | MongoDB URI style | Redis address style |
|---|---|---|
| Host machine with `go run` | `localhost:27017` | `localhost:6379` |
| Docker Compose container | service DNS like `mongo:27017` | service DNS like `redis:6379` |

If `${SERVICE_NAME}` also runs inside Docker Compose, initiate replica set with the hostname that the service can resolve, for example `mongo:27017`, not `localhost:27017`.

## 9. Local Development Setup

### Step 1: Read Previous Dependency Docs First

Follow these in order:

```md
1. `PREVIOUS_TASK1_DEPENDENCY_FILE`
2. `PREVIOUS_TASK2_DEPENDENCY_FILE`
3. `PREVIOUS_TASK3_DEPENDENCY_FILE`
4. `PREVIOUS_TASK4_DEPENDENCY_FILE`
5. `PREVIOUS_TASK5_DEPENDENCY_FILE`
6. `PREVIOUS_TASK6_DEPENDENCY_FILE`
```

Why:

- `PREVIOUS_TASK1_DEPENDENCY_FILE` has base Go, MongoDB, Redis, Docker, env, migration, and run commands.
- `PREVIOUS_TASK2_DEPENDENCY_FILE` explains MongoDB vs Redis responsibility.
- `PREVIOUS_TASK3_DEPENDENCY_FILE` explains cart schema and indexes.
- `PREVIOUS_TASK4_DEPENDENCY_FILE` explains Add Item, useful for creating active carts.
- `PREVIOUS_TASK5_DEPENDENCY_FILE` explains mutation retry/cache behavior.
- `PREVIOUS_TASK6_DEPENDENCY_FILE` explains coupon preview fields that merge clears.
- `${TASK_FILE_NAME}` only adds merge-specific identity, transaction, and cache invalidation verification.

### Step 2: Go to Service Directory

```bash
cd backend/services/cart-service
```

### Step 3: Install Dependencies

No new dependency install is needed for `${TASK_FILE_NAME}`.

If this is a fresh clone, reuse previous setup:

```bash
go mod download
```

### Step 4: Start Reused Databases/Services

Start Redis using previous docs.

Start MongoDB using previous docs, but ensure it supports transactions if you want to verify merge end-to-end.

Refer:
`PREVIOUS_TASK1_DEPENDENCY_FILE`

Sections:
`5.1 MongoDB`, `5.2 Redis`, `9. Docker and DevOps Setup`

Apply MongoDB schema/indexes if not already done:

```bash
mongosh "$CART_MONGO_URI" migrations/mongo/001_create_carts_collection.up.js
```

### Step 5: Ensure Startup Env Exists

Keep the full previous `.env`.

For merge, verify these values:

```env
CART_MONGO_DATABASE=cart_db
CART_ALLOW_GUEST_CART_ID_ONLY_MERGE=false
CART_MAX_SAVE_ATTEMPTS=3
CART_PRODUCT_BASE_URL=http://localhost:8082
CART_CMS_BASE_URL=http://localhost:8089
CART_CMS_VALIDATE_COUPON_PATH=/internal/v1/coupons/validate
```

If local MongoDB is replica set `rs0`, use:

```env
CART_MONGO_URI=mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin&replicaSet=rs0
```

### Step 6: Source Env

```bash
set -a
. ./.env
set +a
```

### Step 7: Start Backend Service

```bash
go run ./cmd/server
```

Expected log:

```text
cart.http.starting
```

### Step 8: Verify Health and Readiness

```bash
curl -i http://localhost:8084/healthz
curl -i http://localhost:8084/readyz
```

Notes:

- Startup pings MongoDB and Redis.
- `/readyz` checks MongoDB through schema manager.
- Product/CMS URLs are validated at startup, but Product/CMS are not pinged by `/readyz`.

### Step 9: Prepare a Guest Cart

Recommended path:

- Use Add Item flow from `PREVIOUS_TASK4_DEPENDENCY_FILE`.
- Send guest session header while adding item.
- Keep the returned `cart_id` as `guest_cart_id`.

If Product Service is unavailable and you only need a local merge smoke test, seed a valid active guest cart manually.

Warning:

- Manual DB seed is only for local testing.
- Production data should be created through service APIs.
- Bad manual data can fail domain validation.

Local seed:

```bash
mongosh "$CART_MONGO_URI" --quiet --eval '
const dbx = db.getSiblingDB("cart_db");
const now = new Date();
const expires = new Date(Date.now() + 1000 * 60 * 60 * 24);
dbx.carts.deleteMany({ _id: "cart_guest_task7" });
dbx.carts.deleteMany({ user_id: "user_task7", status: "active" });
dbx.carts.insertOne({
  _id: "cart_guest_task7",
  guest_session_id: "sess_task7",
  status: "active",
  items: [{
    item_id: "item_guest_task7",
    product_id: "prod_task7",
    variant_id: "var_task7",
    seller_id: "seller_task7",
    title_snapshot: "Local Test Product",
    variant_snapshot: { size: "M" },
    unit_price: { amount: NumberLong(1200), currency: "INR" },
    quantity: 2,
    line_subtotal: { amount: NumberLong(2400), currency: "INR" },
    price_snapshot_at: now,
    added_at: now,
    updated_at: now
  }],
  totals: {
    subtotal: { amount: NumberLong(2400), currency: "INR" },
    discount: { amount: NumberLong(0), currency: "INR" },
    total: { amount: NumberLong(2400), currency: "INR" },
    currency: "INR",
    item_count: 2,
    unique_item_count: 1
  },
  version: NumberLong(1),
  created_at: now,
  updated_at: now,
  expires_at: expires
});
'
```

### Step 10: Verify `${TASK_FILE_NAME}` API

Call merge:

```bash
curl -i -X POST http://localhost:8084/api/v1/cart/merge \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_task7" \
  -H "X-Guest-Session-ID: sess_task7" \
  -d '{"guest_cart_id":"cart_guest_task7"}'
```

Expected:

- HTTP `200`.
- Response has `user_id = user_task7`.
- Response has `status = active`.
- Response contains guest item.
- Response does not include `guest_session_id` as the owner of final cart.

Then verify MongoDB and Redis using sections `5.4` and `6.1`.

## 10. Running the Project

Recommended complete flow for `${TASK_FILE_NAME}`:

1. Read previous dependency docs.
2. Install Go if needed.
3. Start MongoDB with transaction support.
4. Start Redis.
5. Run MongoDB migration/bootstrap.
6. Keep Product/CMS URL env configured for startup.
7. Keep `CART_ALLOW_GUEST_CART_ID_ONLY_MERGE=false`.
8. Source `.env`.
9. Start `${SERVICE_NAME}`.
10. Create/seed active guest cart.
11. Call `POST /api/v1/cart/merge` with trusted user and guest session headers.
12. Verify response.
13. Verify guest cart `status = merged`.
14. Verify user cart active items/totals.
15. Verify Redis user keys refreshed and guest keys deleted.
16. Check logs for warnings/errors.

### Ports and Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| `${SERVICE_NAME}` HTTP API | `8084` | Merge route, health, readiness | Reused |
| MongoDB | `27017` | Durable `cart_db.carts` store | Reused; must support transactions |
| Redis | `6379` | Active cart and summary cache | Reused |
| Product Service | `8082` | Not called by merge, URL config required | Reused |
| CMS Service | `8089` | Not called by merge, URL config required | Reused |
| Auth Service | Depends on platform | Login/JWT validation upstream | External |
| Session/Gateway | Depends on platform | Trusted owner/session headers upstream | External |
| Kafka / RabbitMQ | N/A | Not used by merge | Not required |

Port conflict:

```env
CART_HTTP_ADDR=:8094
```

Then call:

```bash
curl -i http://localhost:8094/healthz
```

### Task-Specific Endpoints

| Endpoint | Method | Purpose |
|---|---|---|
| `/api/v1/cart/merge` | `POST` | Merge guest cart into logged-in user cart |
| `/healthz` | `GET` | Process health |
| `/readyz` | `GET` | MongoDB readiness |
| `/internal/v1/cart/schema/bootstrap` | `POST` | Ensure cart schema/indexes if migration skipped |

## 11. Common Errors & Fixes

Generic Go/MongoDB/Redis/Docker/env errors are already documented in previous files. This section only covers merge-specific or task-relevant errors.

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `USER_AUTH_REQUIRED` | `X-User-ID` missing | Send trusted user header in local curl; configure Gateway/Auth in real env | Gateway should inject authenticated user id |
| `GUEST_CART_ID_REQUIRED` | Request body missing `guest_cart_id` | Send `{"guest_cart_id":"..."}` | Frontend should use latest guest cart id |
| `GUEST_CART_ACCESS_DENIED` | Missing/mismatched guest session header, or guest cart belongs to another session | Send correct `X-Guest-Session-ID` matching cart document | Keep session cookie and cart id paired |
| `GUEST_CART_NOT_FOUND` | Cart id does not exist | Create guest cart through Add Item or seed valid test cart | Do not use stale local cart ids |
| `GUEST_CART_NOT_ACTIVE` | Guest cart is already `merged`, `expired`, `checked_out`, or `abandoned` | Use current active guest cart | Frontend should not reuse old guest cart after login |
| `CART_EXPIRED` | `expires_at` is in the past | Create a fresh guest cart | Keep test seed expiry in future |
| `MIXED_CURRENCY_CART` | Guest and user carts have different currencies | Use same currency carts or clear old user cart in local test | Keep Product/Currency config consistent |
| `CART_VERSION_CONFLICT` | Parallel request modified user/guest cart during merge | Retry request; inspect concurrent clients | Keep `CART_MAX_SAVE_ATTEMPTS=3` and monitor conflicts |
| `Transaction numbers are only allowed on a replica set member or mongos` | Local MongoDB is standalone | Run MongoDB as single-node replica set and add `replicaSet=rs0` to URI | Use transaction-capable MongoDB for merge testing |
| Merge returns `200` but no warning in response | Warnings are logged, not returned in response schema | Check logs for `cart.merge.warning` | Add response warning contract later if frontend needs it |
| Merge succeeds but guest Redis key still exists | Redis delete failed after MongoDB commit | Check logs for `cart.merge.cache_refresh_failed`; verify Redis health | Monitor Redis and keep it reachable |
| Server fails before merge due Product/CMS env | Current config validates Product/CMS URLs at startup | Keep `CART_PRODUCT_BASE_URL`, `CART_CMS_BASE_URL`, and coupon path configured | Copy full `.env.example` from previous setup |
| `CART_TOTALS_INVALID` | Manual seed has wrong totals or line subtotal | Fix seed data or create cart through API | Avoid manual DB writes except local smoke tests |

## 12. Security & Best Practices

### Task-Specific Security

- Keep `CART_ALLOW_GUEST_CART_ID_ONLY_MERGE=false` in production.
- API Gateway must strip any browser-supplied `X-User-ID` and `X-Guest-Session-ID`, then inject trusted values after auth/session validation.
- Do not trust `guest_cart_id` alone. Cart ids can leak through logs, browser storage, or screenshots.
- Keep guest session id secret-ish. Implementation logs only a hash through `guest_session_id_hash`; keep that pattern.
- Use HTTPS/TLS at Gateway. Guest session cookies should be `HttpOnly`, `Secure`, and `SameSite` as appropriate.
- Rate-limit login/merge calls. Merge can mutate two cart documents and should not be spammed.
- Keep MongoDB unique active cart indexes enabled to prevent duplicate active carts per owner.
- Keep MongoDB transaction support enabled for production. Half-merged carts are painful to debug.
- Treat Redis as cache only. MongoDB is source of truth.
- Watch logs: `cart.merge.started`, `cart.merge.completed`, `cart.merge.warning`, `cart.merge.version_conflict`, and `cart.merge.cache_refresh_failed`.
- If frontend needs quantity-cap warnings, add an explicit response contract instead of scraping logs.

### Reused Best Practices

Refer:
`PREVIOUS_TASK1_DEPENDENCY_FILE`

Sections:
`13. Security and Configuration Audit`, `14. Best Practices for Beginners`

Refer:
`PREVIOUS_TASK6_DEPENDENCY_FILE`

Section:
`12. Security & Best Practices`

## 13. Missing or Misconfigured Things

| Issue | Why it matters | Suggested fix |
|---|---|---|
| Base local MongoDB docs use a standalone container | Merge transaction can fail on standalone MongoDB | Add an official local replica-set compose/profile for `${SERVICE_NAME}` |
| No dedicated `${SERVICE_NAME}` Dockerfile found | Full service container build flow is incomplete | Add Dockerfile in future DevOps task |
| No complete local compose stack found for all services | Beginners must manually run MongoDB, Redis, optional Product/CMS mocks, and service | Add compose stack with MongoDB replica set, Redis, service, and optional mocks |
| Merge endpoint trusts headers in current local HTTP handler | Safe only behind trusted Gateway; unsafe if public service accepts browser headers directly | Deploy behind Gateway that strips/injects identity headers |
| Auth/Session service implementation is not part of this folder | Beginner cannot test real login-to-merge flow from only `${SERVICE_NAME}` | Provide platform-level Auth/Session runbook or local gateway mock |
| `/readyz` checks MongoDB readiness, while Redis is checked only at startup | Redis can fail after startup but readiness may still pass | Add Redis readiness if production requires cache health signal |
| Product/CMS URLs are required at startup even though merge does not call them | Beginner may think Product/CMS are required for merge business logic | Keep env configured; optionally lazy-init feature clients later |
| Merge warnings are logs only | Frontend cannot display "quantity capped" without reading logs | Add warnings field to API response if buyer UX requires it |
| Manual MongoDB seed can bypass domain rules | Bad seed can produce totals/currency errors | Prefer Add Item flow; manual seed only for local smoke tests |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `4. Go Dependency System` | Same Go module, Go version, and dependency workflow |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `5.1 MongoDB` | Same database, URI pattern, migration, and verification base |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `5.2 Redis` | Same Redis installation, Docker setup, and verification |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `6. External Services` | Product/CMS startup config is reused |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `8. Complete Environment Variables` | Full `.env` location and variable explanations already documented |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `9. Docker and DevOps Setup` | Base Docker setup reused |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `12. Common Errors and Fixes` | Generic Go/MongoDB/Redis/Docker/env troubleshooting reused |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | `6.2 Redis` | Redis key patterns and cache TTL behavior reused |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | `7. Environment Variables` | Cache and DB env behavior reused |
| `PREVIOUS_TASK3_DEPENDENCY_FILE` | `5. Database Setup` | Cart schema, statuses, indexes, and migration reused |
| `PREVIOUS_TASK4_DEPENDENCY_FILE` | `6.1 Product Service` | Add Item setup reused for creating guest carts through API |
| `PREVIOUS_TASK4_DEPENDENCY_FILE` | `11. Common Errors & Fixes` | Product/Add Item setup errors reused when preparing active carts |
| `PREVIOUS_TASK5_DEPENDENCY_FILE` | `11. Common Errors & Fixes` | Mutation owner/version/cache error patterns reused |
| `PREVIOUS_TASK6_DEPENDENCY_FILE` | `5. Database Setup` | Coupon fields reused and cleared by merge |
| `PREVIOUS_TASK6_DEPENDENCY_FILE` | `12. Security & Best Practices` | Stale coupon preview and internal service security reused |

## 15. Final Checklist

- [ ] Previous dependency documentation checked.
- [ ] No duplicate Go/MongoDB/Redis/Docker installation docs added.
- [ ] No new Go dependency added for `${TASK_FILE_NAME}`.
- [ ] MongoDB running and reachable.
- [ ] MongoDB supports transactions for merge verification.
- [ ] `CART_MONGO_URI` includes correct host, credentials, `authSource`, and replica-set option when needed.
- [ ] MongoDB migration/bootstrap completed.
- [ ] Redis running and reachable.
- [ ] Product/CMS URLs configured for server startup.
- [ ] `backend/services/cart-service/.env` created from previous setup.
- [ ] `.env` sourced before `go run`.
- [ ] `CART_ALLOW_GUEST_CART_ID_ONLY_MERGE=false` unless doing a controlled local-only test.
- [ ] API Gateway/Auth/Session plan exists for trusted `user_id` and `guest_session_id`.
- [ ] `${SERVICE_NAME}` starts on expected `CART_HTTP_ADDR`.
- [ ] `/healthz` and `/readyz` verified.
- [ ] Active guest cart exists with future `expires_at`.
- [ ] `POST /api/v1/cart/merge` verified with `X-User-ID` and `X-Guest-Session-ID`.
- [ ] Response returns active user cart.
- [ ] Guest MongoDB cart marked `merged`.
- [ ] Guest MongoDB cart has `merged_into_cart_id`.
- [ ] User MongoDB cart contains merged items and recalculated totals.
- [ ] Stale coupon fields are cleared on target cart.
- [ ] Redis user active/summary keys refreshed.
- [ ] Redis guest active/summary keys deleted.
- [ ] Logs checked for merge warnings, access denied, version conflicts, and cache refresh failures.
- [ ] No real secrets committed to repository.
