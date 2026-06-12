# ${SERVICE_NAME} ${TASK_FILE_NAME} - Dependency and Setup Documentation

```text
SERVICE_NAME = "Cart Service"
TASK_FILE_NAME = "task2.md"
INPUT_FILE_PATH = TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}
OUTPUT_FILE_NAME = task2_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}
PREVIOUS_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
```

## 1. Purpose

Ye document `${SERVICE_NAME}` ke `${TASK_FILE_NAME}` ke liye dependency, setup, environment, database/cache, Docker, run, aur troubleshooting notes explain karta hai.

Important scope:

- `${TASK_FILE_NAME}` ka main output MongoDB plus Redis storage decision guide hai.
- Is task ne koi new backend package, container, port, ya database engine introduce nahi kiya.
- Existing setup already detail me documented hai in `PREVIOUS_DEPENDENCY_FILE`.
- Is file me sirf Task 2 ke new/incremental setup points detail me explain kiye gaye hain.

Beginner-friendly short version:

MongoDB original cart data rakhega. Redis fast copy/cache rakhega. Agar Redis empty ho jaye to MongoDB se data wapas mil sakta hai. Agar MongoDB down hai to cart ka real data unavailable hai.

## 2. Previous Dependency File Reuse

Task 2 same `${SERVICE_NAME}` folder ke previous setup ko reuse karta hai. Duplicate installation steps intentionally repeat nahi kiye gaye.

| Topic | Status | Refer |
|---|---|---|
| Repository clone and basic tool setup | Reused | `PREVIOUS_DEPENDENCY_FILE`, section `10. Project Run Instructions` |
| Go version, `go.mod`, `go.sum`, `go work` | Reused | `PREVIOUS_DEPENDENCY_FILE`, section `4. Go Dependency System` |
| Go dependency commands | Reused | `PREVIOUS_DEPENDENCY_FILE`, section `4.4 Dependency Commands` |
| MongoDB installation and Docker run | Reused | `PREVIOUS_DEPENDENCY_FILE`, section `5.1 MongoDB` |
| MongoDB migration/bootstrap | Reused | `PREVIOUS_DEPENDENCY_FILE`, section `5.1 MongoDB` -> `MongoDB Migration` |
| Redis installation and Docker run | Reused | `PREVIOUS_DEPENDENCY_FILE`, section `5.2 Redis` |
| Product Service and CMS Service setup | Reused | `PREVIOUS_DEPENDENCY_FILE`, section `6. External Services` |
| Complete `.env` example | Reused | `PREVIOUS_DEPENDENCY_FILE`, section `8. Complete Environment Variables` |
| Docker Compose local dependency setup | Reused | `PREVIOUS_DEPENDENCY_FILE`, section `9. Docker and DevOps Setup` |
| Backend start commands and health checks | Reused | `PREVIOUS_DEPENDENCY_FILE`, section `10. Project Run Instructions` |
| Common setup errors | Reused | `PREVIOUS_DEPENDENCY_FILE`, section `12. Common Errors and Fixes` |

New in this file:

- MongoDB vs Redis responsibility split for Task 2.
- Actual Redis key patterns and TTL variables used by the current service.
- Mapping between Task 2's future config names and current implemented env names.
- Task-specific verification commands for MongoDB durable data and Redis cache keys.

## 3. Files Inspected

| File | Why inspected |
|---|---|
| `INPUT_FILE_PATH` | Task 2 scope: choose MongoDB plus Redis |
| `PREVIOUS_DEPENDENCY_FILE` | Reuse existing setup docs and avoid duplicate instructions |
| `backend/services/cart-service/go.mod` | Confirm Go dependencies did not change for Task 2 |
| `backend/services/cart-service/.env.example` | Confirm actual env variable names |
| `backend/services/cart-service/internal/config/config.go` | Confirm required env validation, defaults, and fallbacks |
| `backend/services/cart-service/internal/repository/mongo_cart_repository.go` | Confirm MongoDB is durable cart store |
| `backend/services/cart-service/internal/repository/redis_cart_cache.go` | Confirm Redis cache keys and TTL usage |
| `backend/services/cart-service/internal/repository/mongo_schema.go` | Confirm collection/index requirements |
| `backend/services/cart-service/migrations/mongo/001_create_carts_collection.up.js` | Confirm MongoDB migration details |
| `backend/services/cart-service/cmd/server/main.go` | Confirm startup dependencies and service port |
| `backend/services/cart-service/cmd/cart-expiry-worker/main.go` | Confirm Redis lock and cleanup worker dependencies |

## 4. Project Tech Stack Analysis

No new technology was added by `${TASK_FILE_NAME}`. The same Cart Service runtime stack is reused.

| Technology | Required? | Task 2 role | Full setup reference |
|---|---:|---|---|
| Go `1.26.3` | Yes | Cart Service runtime language | `PREVIOUS_DEPENDENCY_FILE`, section `4. Go Dependency System` |
| Go modules | Yes | MongoDB and Redis clients are managed through `go.mod` | `PREVIOUS_DEPENDENCY_FILE`, section `4. Go Dependency System` |
| MongoDB | Yes | Durable source of truth for cart documents | `PREVIOUS_DEPENDENCY_FILE`, section `5.1 MongoDB` |
| Redis | Yes | Hot cart cache, summary cache, and expiry worker lock | `PREVIOUS_DEPENDENCY_FILE`, section `5.2 Redis` |
| MongoDB Go Driver | Yes | Go service talks to MongoDB | `PREVIOUS_DEPENDENCY_FILE`, section `4.3 Direct Go Libraries` |
| `go-redis/v9` | Yes | Go service talks to Redis | `PREVIOUS_DEPENDENCY_FILE`, section `4.3 Direct Go Libraries` |
| Product Service HTTP API | Required for running current API | Product snapshot and stock validation in add-item flow | `PREVIOUS_DEPENDENCY_FILE`, section `6.1 Product Service` |
| CMS Service HTTP API | Required for running current API | Coupon preview validation | `PREVIOUS_DEPENDENCY_FILE`, section `6.2 CMS Service` |
| Docker | Optional but recommended | Local MongoDB/Redis setup | `PREVIOUS_DEPENDENCY_FILE`, section `9. Docker and DevOps Setup` |

Simple Hinglish:

- MongoDB ek document database hai. Cart ka permanent snapshot yahin rahega.
- Redis ek fast in-memory store hai. Cart ka quick cache yahan rahega.
- Redis speed ke liye hai, truth ke liye nahi.
- MongoDB truth ke liye hai, cache ke liye nahi.

## 5. Language-Specific Dependency System

This is a Go project. Task 2 does not introduce a new Go module or new Go library.

Current direct dependencies are already present in `backend/services/cart-service/go.mod`:

| Dependency | Version | Status |
|---|---:|---|
| `go.mongodb.org/mongo-driver` | `v1.17.9` | Reused |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Reused |

Do not run new `go get` commands for Task 2 unless the dependency is missing locally. For install/download/build/test commands, refer:

```text
PREVIOUS_DEPENDENCY_FILE
Section: 4.4 Dependency Commands
Section: 10. Project Run Instructions
```

## 6. Database and Cache Analysis

## 6.1 MongoDB

MongoDB is required.

What it is:

MongoDB ek document database hai. Isme cart JSON-like document ke form me store hota hai, jaise owner, items, totals, coupon preview, status, version, and expiry.

Why this project uses it:

- Cart ka durable source of truth chahiye.
- Cart items naturally embedded array me fit hote hain.
- Guest cart and logged-in user cart dono support karne hain.
- `expires_at` field and TTL index se expired carts cleanup support hota hai.
- Unique partial indexes one active cart per user/session enforce karte hain.

Current required database and collection:

```text
Database: cart_db
Collection: carts
```

Local install, Docker run, migration, and verify commands are already explained in:

```text
PREVIOUS_DEPENDENCY_FILE
Section: 5.1 MongoDB
```

Task 2-specific MongoDB notes:

| Topic | Value |
|---|---|
| Default port | `27017` |
| Actual URI env | `CART_MONGO_URI` |
| Legacy URI fallback | `MONGO_URI` |
| Actual database env | `CART_MONGO_DATABASE` |
| Required database value | `cart_db` |
| Migration file | `backend/services/cart-service/migrations/mongo/001_create_carts_collection.up.js` |
| API bootstrap endpoint | `POST /internal/v1/cart/schema/bootstrap` |

Important:

- `CART_MONGO_DATABASE` must be `cart_db`. Current config rejects any other database name.
- `MONGO_DATABASE` from the future Task 2 config example is not read by the current implementation.
- MongoDB failure is a user-facing failure because cart truth is unavailable.
- Redis data must be rebuildable from MongoDB.

## 6.2 Redis

Redis is required by the current service startup.

What it is:

Redis ek fast in-memory key-value store hai. Cart page, header badge, and summary jaise frequent reads ke liye useful hai.

Why this project uses it:

- Active cart cache fast serve karne ke liye.
- Cart summary/header badge fast serve karne ke liye.
- Expiry cleanup worker me distributed lock ke liye.
- MongoDB repeated reads reduce karne ke liye.

Local install and Docker setup are already explained in:

```text
PREVIOUS_DEPENDENCY_FILE
Section: 5.2 Redis
```

Task 2-specific Redis notes:

| Topic | Value |
|---|---|
| Default port | `6379` |
| Actual address env | `CART_REDIS_ADDR` |
| Legacy address fallback | `REDIS_ADDR` |
| Actual password env | `CART_REDIS_PASSWORD` |
| Legacy password fallback | `REDIS_PASSWORD` |
| Actual DB env | `CART_REDIS_DB` |
| Legacy DB fallback | `REDIS_DB` |

Current Redis key patterns:

| Key pattern | Purpose | TTL env | Default |
|---|---|---|---:|
| `cart:active:user:{user_id}` | Logged-in user's active cart JSON | `CART_ACTIVE_CACHE_TTL` | `15m` |
| `cart:active:guest:{guest_session_id}` | Guest active cart JSON | `CART_GUEST_ACTIVE_CACHE_TTL` | `30m` |
| `cart:summary:user:{user_id}` | Logged-in user's cart summary | `CART_SUMMARY_CACHE_TTL` | `5m` |
| `cart:summary:guest:{guest_session_id}` | Guest cart summary | `CART_SUMMARY_CACHE_TTL` | `5m` |
| `cart:lock:expiry-cleanup` | Expiry worker lock | `CART_CLEANUP_LOCK_TTL` | `10m` |

Important:

- Task 2 mentions future `cart:lock:{cart_id}` mutation locks. Current code does not implement per-cart mutation locks yet.
- Current implemented lock is the expiry worker lock, controlled by `CART_CLEANUP_LOCK_KEY`.
- Current code writes/deletes Redis cache keys after cart mutations and expiry cleanup. A Redis-first `GetCart` read route is part of the Task 2 strategy, but it is not exposed in the current HTTP routes yet.
- Current HTTP server pings Redis during startup. If Redis is down, service will not start.
- After service starts, mutation cache refresh failures are logged as warnings and MongoDB remains the source of truth.

## 7. Environment Variables

Full `.env` setup is already explained in:

```text
PREVIOUS_DEPENDENCY_FILE
Section: 8. Complete Environment Variables
```

Create `.env` in:

```text
backend/services/cart-service/.env
```

Task 2-specific minimal DB/cache env example:

```env
CART_MONGO_URI=mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin
CART_MONGO_DATABASE=cart_db

CART_REDIS_ADDR=localhost:6379
CART_REDIS_PASSWORD=
CART_REDIS_DB=0

CART_ACTIVE_CACHE_TTL=15m
CART_GUEST_ACTIVE_CACHE_TTL=30m
CART_SUMMARY_CACHE_TTL=5m

CART_CLEANUP_LOCK_KEY=cart:lock:expiry-cleanup
CART_CLEANUP_LOCK_TTL=10m
CART_BOOTSTRAP_COLLECTIONS_ON_STARTUP=false
```

### Task 2 Config Name Mapping

`${TASK_FILE_NAME}` includes a future config contract. The current code uses these actual names:

| Future/example name in task guide | Current implemented env | Notes |
|---|---|---|
| `CART_GRPC_PORT=9090` | `CART_HTTP_ADDR=:8084` | Current service exposes HTTP, not gRPC |
| `MONGO_URI` | `CART_MONGO_URI` preferred, `MONGO_URI` fallback | Use Cart-specific name for clarity |
| `MONGO_DATABASE` | `CART_MONGO_DATABASE` | `MONGO_DATABASE` is not read by current code |
| `REDIS_ADDR` | `CART_REDIS_ADDR` preferred, `REDIS_ADDR` fallback | Use Cart-specific name for clarity |
| `REDIS_PASSWORD` | `CART_REDIS_PASSWORD` preferred, `REDIS_PASSWORD` fallback | Keep secret out of git |
| `REDIS_DB` | `CART_REDIS_DB` preferred, `REDIS_DB` fallback | Default is `0` |
| `CART_CACHE_TTL` | `CART_ACTIVE_CACHE_TTL` | `CART_CACHE_TTL` is not read by current code |
| `GUEST_CART_CACHE_TTL` | `CART_GUEST_ACTIVE_CACHE_TTL` preferred, fallback to `GUEST_CART_CACHE_TTL` | Prefer Cart-specific env |
| `CART_SUMMARY_CACHE_TTL` | `CART_SUMMARY_CACHE_TTL` | Same name |

Security notes:

- Do not commit real MongoDB or Redis passwords.
- Use `.env.example` for sample values only.
- In Docker Compose network, use service names like `mongo:27017` and `redis:6379`.
- On host machine, use `localhost:27017` and `localhost:6379`.

## 8. External Services Analysis

No new external service is introduced by Task 2.

| Service | Required? | Task 2 status | Setup |
|---|---:|---|---|
| MongoDB | Yes | Reused durable cart store | `PREVIOUS_DEPENDENCY_FILE`, section `5.1 MongoDB` |
| Redis | Yes | Reused hot cache and worker lock | `PREVIOUS_DEPENDENCY_FILE`, section `5.2 Redis` |
| Product Service | Required by current API startup/config | Reused, not newly added by Task 2 | `PREVIOUS_DEPENDENCY_FILE`, section `6.1 Product Service` |
| CMS Service | Required by current API startup/config | Reused, not newly added by Task 2 | `PREVIOUS_DEPENDENCY_FILE`, section `6.2 CMS Service` |
| Docker | Optional but recommended | Reused local infra runner | `PREVIOUS_DEPENDENCY_FILE`, section `9. Docker and DevOps Setup` |
| Kafka/RabbitMQ/NATS/MinIO/SMTP/Stripe/Twilio/OAuth | No | Not used by Task 2 | Not required |

Beginner note:

Product Service and CMS Service are not part of the MongoDB plus Redis decision itself, but current Cart HTTP API validates their base URLs during startup. Isliye `.env` me `CART_PRODUCT_BASE_URL` and `CART_CMS_BASE_URL` bhi set hone chahiye. Full explanation previous dependency file me hai.

## 9. Ports and Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Cart HTTP API | `8084` | Current Cart Service REST API and health checks | Reused |
| MongoDB | `27017` | Durable cart database | Reused |
| Redis | `6379` | Cart cache and expiry lock | Reused |
| Product Service | `8082` | Product/variant snapshot validation | Reused |
| CMS Service | `8089` | Coupon preview validation | Reused |
| Cart gRPC | `9090` | Mentioned in Task 2 future example only | Not active in current code |

How to change values:

| Need | Env/config |
|---|---|
| Change Cart HTTP port | `CART_HTTP_ADDR=:9091` |
| Change Mongo host/port | Update host/port inside `CART_MONGO_URI` |
| Change Redis host/port | `CART_REDIS_ADDR=host:port` |
| Change Product Service URL | `CART_PRODUCT_BASE_URL=http://host:port` |
| Change CMS Service URL | `CART_CMS_BASE_URL=http://host:port` |

Common networking reminder:

- Service running on host machine: use `localhost`.
- Service running inside Docker Compose: use container service names, for example `mongo` and `redis`.
- If `localhost` is used inside a container, it points to that container itself, not your host machine.

## 10. Docker and DevOps Setup

Docker setup is reused. Do not create a separate MongoDB or Redis container only for Task 2 if the previous local dependency containers are already running.

Refer:

```text
PREVIOUS_DEPENDENCY_FILE
Section: 9. Docker and DevOps Setup
```

Task 2-specific DevOps expectations:

| Area | Requirement |
|---|---|
| MongoDB volume | Required for durable cart data. Do not treat MongoDB as disposable cache. |
| MongoDB backups | Needed in real environment because MongoDB is source of truth. |
| Redis persistence | Optional for this task because Redis is cache, but password and network isolation are still important. |
| Redis memory policy | Must not silently evict critical non-cache keys in production. Keep cart cache TTLs bounded. |
| Health checks | Cart `/readyz` must pass after MongoDB and Redis are reachable. |
| Internal schema bootstrap endpoint | Protect `/internal/v1/cart/schema/bootstrap`; do not expose it publicly. |

## 11. Migration and Bootstrap Notes

Migration details are reused from:

```text
PREVIOUS_DEPENDENCY_FILE
Section: 5.1 MongoDB
Topic: MongoDB Migration
```

Task 2 dependency impact:

- MongoDB must have `cart_db.carts`.
- Indexes must exist before real traffic.
- TTL index on `expires_at` manages long-term cart expiry.
- Redis TTLs only manage cache freshness, not durable cart expiry.

Verify indexes after migration:

```bash
mongosh "mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin" \
  --quiet \
  --eval "db.getSiblingDB('cart_db').carts.getIndexes().map(i => i.name)"
```

Expected important index names:

```text
idx_carts_user_status
idx_carts_guest_status
idx_carts_expires_at_ttl
uniq_active_cart_per_user
uniq_active_cart_per_guest_session
idx_carts_item_product_variant
idx_carts_updated_at
```

## 12. Project Run Instructions

Complete onboarding flow is already documented in:

```text
PREVIOUS_DEPENDENCY_FILE
Section: 10. Project Run Instructions
```

Use that full flow, with these Task 2 checks added:

1. Start MongoDB and Redis using the previous dependency doc.
2. Create `.env` using the previous full `.env` example.
3. Confirm DB/cache env names match section `7. Environment Variables` in this file.
4. Run MongoDB migration or enable `CART_BOOTSTRAP_COLLECTIONS_ON_STARTUP=true`.
5. Start the Cart HTTP API on `CART_HTTP_ADDR`, default `:8084`.
6. Verify readiness:

```bash
curl http://localhost:8084/readyz
```

Expected:

```json
{"status":"ready"}
```

7. After a cart mutation request has succeeded, inspect Redis cache keys:

```bash
redis-cli --scan --pattern "cart:*"
```

8. Check TTL for a specific key:

```bash
redis-cli TTL "cart:summary:user:<user_id>"
```

Expected:

- Positive number means key has expiry.
- `-2` means key does not exist.
- `-1` means key exists but no TTL, which is not expected for these cache keys.

## 13. Task-Specific Troubleshooting

| Problem | Likely cause | Fix |
|---|---|---|
| `CART_MONGO_DATABASE must be "cart_db"` | Wrong database env value | Set `CART_MONGO_DATABASE=cart_db` |
| Service starts on `8084`, not `9090` | Task 2 future example mentions gRPC, current code is HTTP | Use `CART_HTTP_ADDR=:8084` or change this env |
| `CART_MONGO_URI or MONGO_URI is required` | Mongo URI missing | Add `CART_MONGO_URI` in `.env` |
| Redis ping fails at startup | Redis container/service not running or wrong host | Start Redis and check `CART_REDIS_ADDR` |
| Redis keys missing after mutation | Mutation may not have run, cache refresh failed, or TTL expired | Check service logs for `cache_*_failed`, rerun API request |
| Redis key has no TTL | Key was created manually or wrong code path used | Delete manual key and let service recreate it |
| Duplicate active cart error | Unique active cart index is enforcing one active cart per owner | Check existing active cart for same `user_id` or `guest_session_id` |
| Works on host but fails inside Docker | `localhost` points to wrong network namespace | Use `mongo:27017` and `redis:6379` inside Compose |
| MongoDB has data but Redis is empty | Redis is cache only | This is acceptable; next cache refresh can rebuild keys |
| Redis has data but MongoDB is missing cart | Redis data cannot be trusted as source | Treat cart as missing and debug MongoDB write path |

## 14. Best Practices for Beginners

- Always think `MongoDB = truth`, `Redis = speed`.
- Do not manually edit Redis cart JSON during debugging unless you delete it afterwards.
- Do not store permanent cart-only data in Redis.
- Do not trust Redis-only data during checkout.
- Keep Redis TTLs short enough to avoid stale UX.
- Keep MongoDB indexes in place before load testing.
- Use Cart-specific env names like `CART_MONGO_URI` and `CART_REDIS_ADDR`.
- Keep `.env` local and never commit secrets.
- In production, protect MongoDB and Redis from public internet access.

## 15. Final Setup Checklist

| Check | Status |
|---|---|
| Previous dependency file reviewed | Required |
| MongoDB running on correct host/port | Required |
| Redis running on correct host/port | Required |
| `.env` created in `backend/services/cart-service/.env` | Required |
| `CART_MONGO_URI` points to `cart_db` | Required |
| `CART_MONGO_DATABASE=cart_db` | Required |
| `CART_REDIS_ADDR` points to running Redis | Required |
| Cache TTL envs set or defaults accepted | Required |
| MongoDB migration/bootstrap completed | Required before real traffic |
| `/readyz` returns ready | Required |
| Redis `cart:*` keys have TTL after mutations | Expected |

## 16. What Is Not Required for Task 2

Task 2 does not require:

- New Go dependency installation.
- New database engine.
- Kafka, RabbitMQ, NATS, MinIO, SMTP, Stripe, Twilio, or OAuth setup.
- A new Docker Compose file.
- A gRPC server port in the current implementation.
- Business logic rewrite.
- Manual Redis cache preloading.

Final beginner summary:

Follow `PREVIOUS_DEPENDENCY_FILE` for full setup. For `${TASK_FILE_NAME}`, the only important new understanding is storage responsibility: MongoDB must be configured as durable cart source of truth, Redis must be configured as short-lived hot cache, and current env names must match the implemented Cart Service config.
