# Project Dependency & Setup Guide

```text
SERVICE_NAME = "Cart Service"
TASK_FILE_NAME = "task8.md"
INPUT_FILE_PATH = TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}
OUTPUT_FILE_NAME = ${TASK_FILE_NAME without ".md"}_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}

PREVIOUS_TASK1_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
PREVIOUS_TASK2_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
PREVIOUS_TASK3_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task3_Dependency.md
PREVIOUS_TASK4_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task4_Dependency.md
PREVIOUS_TASK5_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task5_Dependency.md
PREVIOUS_TASK6_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task6_Dependency.md
PREVIOUS_TASK7_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task7_Dependency.md
```

## 1. Project Overview

Ye document `${SERVICE_NAME}` ke `${TASK_FILE_NAME}` ke liye dependency, setup, environment, database, Docker, DevOps, aur troubleshooting guide hai.

Important scope:

- `INPUT_FILE_PATH` original implementation guide hai. Is file ko modify nahi kiya gaya.
- `OUTPUT_FILE_PATH` new dependency/setup guide hai.
- Previous dependency docs already Go, MongoDB, Redis, Docker, `.env`, migrations, Product Service, CMS Service, and generic troubleshooting explain karte hain.
- Is file me duplicate installation guide repeat nahi kiya gaya. Sirf `${TASK_FILE_NAME}` ke cart expiry worker, scheduler, lock, cleanup config, verification, and operational risks explain kiye gaye hain.

Beginner-friendly summary:

`${TASK_FILE_NAME}` ka purpose inactive carts ko cleanup karna hai. App `expires_at <= now` wale active carts ko expired mark karta hai, Redis active/summary cache delete karta hai, and MongoDB TTL index physical storage cleanup ka safety net banata hai. Simple words me: cart stale ho gaya to backend usko active cart ki tarah treat nahi karega.

Current repo reality:

- `INPUT_FILE_PATH` me kuch sections future design ki tarah likhe hain.
- Inspected backend code me expiry worker and cleanup usecase already present hain at `backend/services/cart-service/cmd/cart-expiry-worker/main.go`, `internal/usecase/cleanup_expired_carts.go`, and `internal/scheduler/*`.
- Isliye ye dependency guide current implemented code ke basis par written hai, not only future design ke basis par.

### Files Inspected

| File | Why inspected |
|---|---|
| `INPUT_FILE_PATH` | `${TASK_FILE_NAME}` scope: inactive cart expiry and cleanup scheduler |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | Base setup: Go, MongoDB, Redis, Docker, `.env`, run commands, expiry worker command |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | MongoDB source of truth, Redis cache responsibility, Redis key patterns |
| `PREVIOUS_TASK3_DEPENDENCY_FILE` | `cart_db.carts` schema, indexes, TTL index, migration/bootstrap |
| `PREVIOUS_TASK4_DEPENDENCY_FILE` | Product Service setup used when preparing carts through add-item API |
| `PREVIOUS_TASK5_DEPENDENCY_FILE` | Expired cart mutation guard and cache invalidation behavior |
| `PREVIOUS_TASK6_DEPENDENCY_FILE` | CMS env/security reused by server startup config |
| `PREVIOUS_TASK7_DEPENDENCY_FILE` | Guest/user cart TTL and expired guest merge behavior |
| `backend/services/cart-service/go.mod` | Confirm no new Go dependency for `${TASK_FILE_NAME}` |
| `backend/services/cart-service/.env.example` | Confirm actual expiry and cleanup env variable names |
| `backend/services/cart-service/internal/config/config.go` | Confirm cleanup config defaults and validation |
| `backend/services/cart-service/cmd/cart-expiry-worker/main.go` | Confirm worker startup dependencies and one-shot behavior |
| `backend/services/cart-service/internal/usecase/cleanup_expired_carts.go` | Confirm batch cleanup, report fields, cache invalidation |
| `backend/services/cart-service/internal/scheduler/expiry_scheduler.go` | Confirm Redis lock, run timeout, and scheduler behavior |
| `backend/services/cart-service/internal/scheduler/redis_lock.go` | Confirm lock uses Redis `SET NX` and owner-safe release |
| `backend/services/cart-service/internal/repository/mongo_cart_repository.go` | Confirm expired cart query/update filters |
| `backend/services/cart-service/internal/repository/redis_cart_cache.go` | Confirm active cart and summary cache key deletion |
| `backend/services/cart-service/migrations/mongo/001_create_carts_collection.up.js` | Confirm TTL index and schema migration |

## 2. Tech Stack

No brand-new technology is introduced by `${TASK_FILE_NAME}`. Existing stack reused hai, but expiry worker ke liye kuch existing pieces operationally important ho jate hain.

| Technology | Required? | Status | Why used for `${TASK_FILE_NAME}` | Full setup reference |
|---|---:|---|---|---|
| Go `1.26.3` | Yes | Reused | HTTP API, cleanup usecase, and worker command run karne ke liye | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4. Go Dependency System` |
| Go modules | Yes | Reused | MongoDB and Redis libraries manage karne ke liye | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4.4 Dependency Commands` |
| Standard `context`, `time`, `log/slog` | Yes | Reused | Worker timeout, scheduling, structured logs | No install needed |
| MongoDB | Yes | Reused | Expired active carts scan/update and TTL index based physical cleanup | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `5.1 MongoDB`; `PREVIOUS_TASK3_DEPENDENCY_FILE`, section `5. Database Setup` |
| MongoDB Go Driver | Yes | Reused | `Find`, `UpdateMany`, index/schema management | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4.3 Direct Go Libraries` |
| Redis | Yes | Reused | Distributed cleanup lock and cache invalidation | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `5.2 Redis`; `PREVIOUS_TASK2_DEPENDENCY_FILE`, section `6.2 Redis` |
| `go-redis/v9` | Yes | Reused | Redis `PING`, `DEL`, `SETNX`, lock release script | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4.3 Direct Go Libraries` |
| Product Service HTTP API | Startup config required | Reused, not called by worker | Shared config validation requires URL values | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `6.1 Product Service` |
| CMS Service HTTP API | Startup config required | Reused, not called by worker | Shared config validation requires URL values | `PREVIOUS_TASK6_DEPENDENCY_FILE`, section `6.1 CMS Service` |
| Docker / Docker Compose | Optional but recommended | Reused | Local MongoDB and Redis start karne ke liye | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `9. Docker and DevOps Setup` |
| Kubernetes CronJob / host cron | Optional | Deployment choice | Recurring production scheduling ke liye useful, but repo me manifest nahi mila | Explained only as operational note in this file |

Simple Hinglish:

- MongoDB permanent cart data rakhta hai.
- Redis fast cache and cleanup worker lock ke liye use hota hai.
- Worker koi new public API port open nahi karta.
- Worker one-shot command hai: run hota hai, cleanup karta hai, logs print karta hai, then exit.

## 3. Required Software

Base software install repeat nahi kiya gaya. Same tools previous setup se use honge.

| Software | Required? | Status for `${TASK_FILE_NAME}` | Refer |
|---|---:|---|---|
| Git | Yes | Reused | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `10. Project Run Instructions` |
| Go `1.26.3` | Yes | Reused | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4.2 Go Version` |
| MongoDB server | Yes | Reused | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `5.1 MongoDB` |
| Redis server | Yes | Reused | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `5.2 Redis` |
| `mongosh` | Strongly recommended | Expired cart count, TTL index verify, local smoke seed | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `5.1 MongoDB` |
| `redis-cli` | Optional but useful | Lock/cache debug locally | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `5.2 Redis` |
| Docker / Docker Compose | Optional but recommended | Local MongoDB and Redis setup | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `9. Docker and DevOps Setup` |
| `curl` | Optional | API health/schema check if HTTP server is also running | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `10. Project Run Instructions` |

No Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, or OAuth provider is required for `${TASK_FILE_NAME}`.

## 4. Dependency Management

This is a Go project. `${TASK_FILE_NAME}` does not require any new `go get`.

Current direct dependencies in `backend/services/cart-service/go.mod`:

| Dependency | Version | Status | Why |
|---|---:|---|---|
| `go.mongodb.org/mongo-driver` | `v1.17.9` | Reused | Expired cart scan/update, Mongo collection/index access |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Reused | Redis cache delete and cleanup lock |

Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 4. Go Dependency System
Section: 4.4 Dependency Commands
Section: 4.6 Common Go Module Issues
```

Task-specific rule:

- Do not add `robfig/cron`, queue clients, or Kubernetes packages for this task.
- Current worker uses existing Go code and Redis lock. Scheduling can be handled outside the app by cron/Kubernetes/systemd without changing Go dependencies.
- Run `go mod tidy` only if you intentionally changed Go dependencies. For `${TASK_FILE_NAME}`, no dependency change is needed.

Useful verification from service folder:

```bash
cd backend/services/cart-service
go test ./...
go build ./cmd/cart-expiry-worker
```

## 5. Database Setup

MongoDB setup is reused. `${TASK_FILE_NAME}` does not add a new database, collection, migration file, or database port.

Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 5.1 MongoDB

PREVIOUS_TASK3_DEPENDENCY_FILE
Section: 5. Database Setup
```

### 5.1 MongoDB Role in Expiry

MongoDB ek document database hai. Is project me cart documents `cart_db.carts` collection me stored hain.

For `${TASK_FILE_NAME}`, MongoDB ka role:

| MongoDB item | Purpose |
|---|---|
| `cart_db` | Required database name; config any other DB reject karta hai |
| `carts` | Cart documents collection |
| `status` | Cleanup sirf `active` carts ko expire karta hai |
| `expires_at` | Business expiry cutoff and TTL index field |
| `idx_carts_expires_at_ttl` | MongoDB TTL index for eventual physical cleanup |
| `uniq_active_cart_per_user` | One active cart per user guard |
| `uniq_active_cart_per_guest_session` | One active cart per guest session guard |

The cleanup worker scans this filter:

```text
status = "active"
expires_at <= current UTC time
```

Then it updates matching carts to:

```text
status = "expired"
updated_at = current UTC time
version = version + 1
```

Why active-only filter important hai:

- `checked_out` carts accidentally expire nahi honge by worker.
- `merged` carts accidentally active state se modify nahi honge.
- If a cart mutation refreshed `expires_at`, worker usko stale candidate ke basis par expire nahi karega.

### 5.2 TTL Index Behavior

TTL index already exists in migration and repository schema:

```text
Index: idx_carts_expires_at_ttl
Field: expires_at
expireAfterSeconds: 0
```

Important beginner note:

MongoDB TTL monitor exact second par delete guarantee nahi karta. Ye background monitor periodically run hota hai. Business logic ko TTL deletion ka wait nahi karna chahiye. App ke liye `expires_at <= now` means cart expired.

Warning:

TTL index physical delete karta hai for any document whose `expires_at` is in the past. Worker only marks active carts expired, but TTL index status nahi dekhta unless index strategy future me change ho. Agar checked-out/audit carts long retention ke liye chahiye, checkout flow ko `expires_at` retention policy clearly set/extend/clear karni hogi.

### 5.3 Task-Specific MongoDB Verification

From `backend/services/cart-service`, after `.env` is sourced:

```bash
mongosh "$CART_MONGO_URI" --eval '
const db = db.getSiblingDB("cart_db");
printjson(db.carts.getIndexes().filter(i => i.name === "idx_carts_expires_at_ttl"));
'
```

Count current expired active candidates:

```bash
mongosh "$CART_MONGO_URI" --eval '
const db = db.getSiblingDB("cart_db");
printjson(db.carts.countDocuments({
  status: "active",
  expires_at: { $lte: new Date() }
}));
'
```

After worker run, check expired count:

```bash
mongosh "$CART_MONGO_URI" --eval '
const db = db.getSiblingDB("cart_db");
printjson(db.carts.countDocuments({ status: "expired" }));
'
```

If local seed data disappears before worker verification, MongoDB TTL monitor may have already physically deleted it. That is not always a worker bug.

## 6. Redis / Queue / External Services

## 6.1 Redis

Redis setup is reused and required.

Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 5.2 Redis

PREVIOUS_TASK2_DEPENDENCY_FILE
Section: 6.2 Redis
```

Redis ek fast in-memory key-value store hai. `${TASK_FILE_NAME}` me Redis ke two roles hain:

| Role | Details |
|---|---|
| Distributed lock | Worker uses `CART_CLEANUP_LOCK_KEY` so duplicate cleanup workers do not run together |
| Cache invalidation | Expired cart owner ke active cart and summary cache keys delete hote hain |

Relevant Redis keys:

| Key pattern | Purpose | Status |
|---|---|---|
| `cart:lock:expiry-cleanup` | Cleanup worker distributed lock | Reused, task-critical |
| `cart:active:user:{user_id}` | User active cart cache | Reused, deleted on expiry |
| `cart:summary:user:{user_id}` | User summary cache | Reused, deleted on expiry |
| `cart:active:guest:{guest_session_id}` | Guest active cart cache | Reused, deleted on expiry |
| `cart:summary:guest:{guest_session_id}` | Guest summary cache | Reused, deleted on expiry |

Lock behavior:

- Worker calls Redis `SETNX` style acquire through `go-redis`.
- If lock is already present, worker logs `cart.expiry.lock_not_acquired` and exits successfully with `lock_skipped=true`.
- Lock release checks owner value before deleting, so one worker does not release another worker's lock.
- Default lock TTL is longer than default run timeout.

Local-only Redis debug:

```bash
redis-cli GET cart:lock:expiry-cleanup
redis-cli TTL cart:lock:expiry-cleanup
```

For local cache checks, prefer exact known keys. Avoid broad `KEYS` in production because it can block Redis on large datasets.

## 6.2 Product Service and CMS Service

Expiry worker does not call Product Service or CMS Service business endpoints.

However, current shared config validation still requires these env values:

```env
CART_PRODUCT_BASE_URL=http://localhost:8082
CART_CMS_BASE_URL=http://localhost:8089
CART_CMS_VALIDATE_COUPON_PATH=/internal/v1/coupons/validate
```

Why:

- `cmd/cart-expiry-worker/main.go` calls `config.Load()`.
- `config.Load()` validates Product and CMS base URLs even though worker does not instantiate those clients.
- Beginner developer ko worker run karte time bhi full `.env.example` source karna chahiye.

Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 6. External Services

PREVIOUS_TASK6_DEPENDENCY_FILE
Section: 6.1 CMS Service
```

## 6.3 Queues and Other Services

No Kafka, RabbitMQ, NATS, queue consumer, MinIO, SMTP, Stripe, Twilio, OAuth, or Elasticsearch dependency is introduced by `${TASK_FILE_NAME}`.

Scheduling can be done by:

| Scheduler option | Required? | Notes |
|---|---:|---|
| Manual `go run ./cmd/cart-expiry-worker` | Yes for local verification | One-shot cleanup and exit |
| Host cron/systemd timer | Optional | Useful for simple VM deployment |
| Kubernetes CronJob | Optional | Recommended for production Kubernetes, but no manifest exists in inspected repo |
| In-process ticker | Implemented in scheduler package | Current worker command calls `RunOnce`, not the continuous `Run` loop |

## 7. Environment Variables

Full `.env` creation and explanation already exists.

Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 8. Complete Environment Variables
```

`${TASK_FILE_NAME}` does not add brand-new env variable names beyond the existing `.env.example`. But these existing variables are task-critical:

```env
CART_EXPIRY_TTL=2160h
CART_USER_EXPIRY_TTL=2160h
CART_GUEST_EXPIRY_TTL=720h
CART_CLEANUP_INTERVAL=15m
CART_CLEANUP_BATCH_SIZE=500
CART_CLEANUP_LOCK_TTL=10m
CART_CLEANUP_RUN_TIMEOUT=2m
CART_CLEANUP_LOCK_KEY=cart:lock:expiry-cleanup
CART_ACTIVE_CACHE_TTL=15m
CART_GUEST_ACTIVE_CACHE_TTL=30m
CART_SUMMARY_CACHE_TTL=5m
```

Task-specific variable details:

| Variable | Required? | Default | Purpose for `${TASK_FILE_NAME}` | Security note |
|---|---:|---:|---|---|
| `CART_EXPIRY_TTL` | Optional | `2160h` | Legacy/default cart expiry TTL fallback | Not secret |
| `CART_USER_EXPIRY_TTL` | Optional | `2160h` | Logged-in cart expiry extension after mutation | Not secret |
| `CART_GUEST_EXPIRY_TTL` | Optional | `720h` | Guest cart expiry extension after mutation | Not secret |
| `CART_CLEANUP_INTERVAL` | Optional | `15m` | Continuous scheduler interval if `Run` loop is used | Not secret |
| `CART_CLEANUP_BATCH_SIZE` | Optional | `500` | Max expired candidates fetched per batch | Not secret |
| `CART_CLEANUP_LOCK_TTL` | Optional | `10m` | Redis distributed lock expiry | Not secret |
| `CART_CLEANUP_RUN_TIMEOUT` | Optional | `2m` | Max time for one cleanup run | Not secret |
| `CART_CLEANUP_LOCK_KEY` | Optional | `cart:lock:expiry-cleanup` | Redis lock key | Not secret |
| `CART_ACTIVE_CACHE_TTL` | Optional | `15m` | User active cart cache TTL | Not secret |
| `CART_GUEST_ACTIVE_CACHE_TTL` | Optional | `30m` | Guest active cart cache TTL | Not secret |
| `CART_SUMMARY_CACHE_TTL` | Optional | `5m` | Header/badge summary cache TTL | Not secret |

Important config notes:

- `.env` should be created at `backend/services/cart-service/.env`.
- Code reads env via `os.Getenv`; it does not automatically load `.env`.
- Before `go run`, source the file:

```bash
set -a
. ./.env
set +a
```

Duration format:

- Use Go duration units like `700ms`, `5s`, `15m`, `2160h`.
- If you write `CART_CLEANUP_INTERVAL=15` without unit, current helper silently falls back to default instead of using `15`.
- If you write `CART_CLEANUP_BATCH_SIZE=0`, validation fails because batch size must be greater than zero.

## 8. Docker Setup

Base Docker setup is reused.

Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 9. Docker and DevOps Setup
```

`${TASK_FILE_NAME}` does not add:

- New Dockerfile
- New docker-compose service
- New Docker volume
- New Docker network
- New exposed port
- New queue container
- New restart policy

### 8.1 Docker Impact Table

| Docker area | Status for `${TASK_FILE_NAME}` | Notes |
|---|---|---|
| MongoDB container | Reused | Must have `cart_db.carts` migration/bootstrap applied |
| Redis container | Reused | Required for startup ping, cache delete, and cleanup lock |
| Backend API container | Not present in inspected setup | Run directly with `go run ./cmd/server` unless future Dockerfile is added |
| Expiry worker container | Not present in inspected setup | Run directly with `go run ./cmd/cart-expiry-worker` locally |
| Full compose stack | Not present in inspected setup | Previous docs provide dependency compose example only |

### 8.2 Ports and Networking

No new port is introduced by `${TASK_FILE_NAME}`.

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend API | `8084` | HTTP routes, health, schema bootstrap | Reused |
| MongoDB | `27017` | `cart_db.carts` durable data | Reused |
| Redis | `6379` | Cache and worker lock | Reused |
| Product Service | `8082` | Add-item product snapshot; URL required by shared config | Reused |
| CMS Service | `8089` | Coupon validation; URL required by shared config | Reused |
| Expiry worker | None | One-shot CLI process | New behavior, no port |

Docker networking reminder:

| Where command runs | MongoDB host style | Redis host style |
|---|---|---|
| Host machine / WSL terminal | `localhost:27017` | `localhost:6379` |
| Docker Compose service container | `mongo:27017` | `redis:6379` |

If worker runs inside a future container, update `CART_MONGO_URI` and `CART_REDIS_ADDR` to container DNS names. If worker runs on host, keep `localhost`.

## 9. Local Development Setup

### Step 1: Read Previous Dependency Docs First

Before following this task-specific guide, read:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Sections: 4, 5, 8, 9, 10, 12

PREVIOUS_TASK2_DEPENDENCY_FILE
Sections: 6.2, 7

PREVIOUS_TASK3_DEPENDENCY_FILE
Sections: 5, 9
```

These docs already explain clone, Go modules, MongoDB, Redis, `.env`, Docker, migrations, and generic errors.

### Step 2: Go to Service Directory

```bash
cd backend/services/cart-service
```

### Step 3: Install Dependencies

No new dependency install is required for `${TASK_FILE_NAME}`.

If this is a fresh clone, use the previous dependency guide:

```bash
go mod download
```

Do not run `go get` for scheduler/cron packages.

### Step 4: Start Reused Databases/Services

Start MongoDB and Redis using previous docs.

Required for worker:

- MongoDB running and reachable
- Redis running and reachable
- Full `.env` values available, including Product/CMS URLs because shared config validates them

Product/CMS services themselves do not need to be alive for worker cleanup, but their URL env values must be syntactically valid.

### Step 5: Apply MongoDB Migration or Bootstrap

If migration already ran in previous tasks, do not repeat unless needed.

Refer:

```text
PREVIOUS_TASK3_DEPENDENCY_FILE
Section: 5. Database Setup
Section: 9. Local Development Setup
```

Manual migration:

```bash
mongosh "$CART_MONGO_URI" migrations/mongo/001_create_carts_collection.up.js
```

Or run the HTTP server and bootstrap endpoint:

```bash
go run ./cmd/server
curl -X POST http://localhost:8084/internal/v1/cart/schema/bootstrap
```

### Step 6: Create and Source `.env`

```bash
cp .env.example .env
set -a
. ./.env
set +a
```

Verify task-critical values:

```bash
printf '%s\n' "$CART_CLEANUP_BATCH_SIZE" "$CART_CLEANUP_RUN_TIMEOUT" "$CART_CLEANUP_LOCK_KEY"
```

### Step 7: Run Expiry Worker Once

```bash
go run ./cmd/cart-expiry-worker
```

Expected behavior:

- Worker loads config.
- Worker pings MongoDB and Redis.
- Worker optionally bootstraps schema if `CART_BOOTSTRAP_COLLECTIONS_ON_STARTUP=true`.
- Worker acquires Redis lock.
- Worker scans expired active carts in batches.
- Worker marks matching carts as `expired`.
- Worker deletes active/summary cache keys for each expired owner.
- Worker logs `cart.expiry.worker_finished`.
- Worker exits.

Important:

Current worker command calls `RunOnce`. It is not a forever-running daemon. For repeated cleanup, schedule this command through cron/systemd/Kubernetes CronJob or create a future entrypoint that calls scheduler `Run`.

### Step 8: Verify Worker Result

Check logs for:

```text
cart.expiry.cleanup_started
cart.expiry.batch_expired
cart.expiry.cleanup_finished
cart.expiry.worker_finished
```

Useful log fields:

| Field | Meaning |
|---|---|
| `scanned_count` | Candidate carts read from MongoDB |
| `expired_count` | Carts actually updated to `expired` |
| `batch_count` | Number of batches processed |
| `cache_delete_failures` | Redis active/summary delete failures |
| `invalid_owner_count` | Candidate carts missing usable owner fields |
| `lock_skipped` | Another worker had the lock |

If `expired_count=0`, it can be valid: maybe no expired active carts exist.

## 10. Running the Project

### Typical Local Flow

1. Follow previous dependency docs for clone, tools, MongoDB, Redis, `.env`, and migration.
2. Go to `backend/services/cart-service`.
3. Source `.env`.
4. Optionally run HTTP API:

```bash
go run ./cmd/server
```

5. Verify health:

```bash
curl http://localhost:8084/healthz
curl http://localhost:8084/readyz
```

6. In another terminal, source `.env` again and run worker:

```bash
cd backend/services/cart-service
set -a
. ./.env
set +a
go run ./cmd/cart-expiry-worker
```

### Task-Specific Smoke Check

Before worker:

```bash
mongosh "$CART_MONGO_URI" --eval '
const db = db.getSiblingDB("cart_db");
print("expired active candidates:");
printjson(db.carts.countDocuments({
  status: "active",
  expires_at: { $lte: new Date() }
}));
'
```

Run worker:

```bash
go run ./cmd/cart-expiry-worker
```

After worker:

```bash
mongosh "$CART_MONGO_URI" --eval '
const db = db.getSiblingDB("cart_db");
print("remaining expired active candidates:");
printjson(db.carts.countDocuments({
  status: "active",
  expires_at: { $lte: new Date() }
}));
print("business-expired carts:");
printjson(db.carts.countDocuments({ status: "expired" }));
'
```

If TTL index already deleted old rows, counts may be zero. That is expected when local data is old enough for MongoDB TTL monitor to purge.

## 11. Common Errors & Fixes

Generic Go/MongoDB/Redis/Docker/env errors are already documented in previous dependency files. This section only covers `${TASK_FILE_NAME}`-specific or worker-relevant errors.

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `cart.config.load_failed` with Product/CMS env error | Worker uses shared config validation | Keep `CART_PRODUCT_BASE_URL`, `CART_CMS_BASE_URL`, and `CART_CMS_VALIDATE_COUPON_PATH` in `.env` | Copy full `.env.example`, not only cleanup vars |
| `CART_CLEANUP_BATCH_SIZE must be greater than zero` | Batch size set to `0` or negative | Set `CART_CLEANUP_BATCH_SIZE=500` | Use positive integer values |
| `CART_CLEANUP_RUN_TIMEOUT must be greater than zero` | Timeout set to `0s` or negative duration | Set `CART_CLEANUP_RUN_TIMEOUT=2m` | Keep duration values positive |
| Worker exits with `cart.mongo.connect_failed` or `cart.mongo.ping_failed` | MongoDB down or URI wrong | Start MongoDB and verify `CART_MONGO_URI` | Run Mongo ping before worker |
| Worker exits with `cart.redis.ping_failed` | Redis down, wrong addr, or password mismatch | Start Redis and fix `CART_REDIS_ADDR` / `CART_REDIS_PASSWORD` | Run `redis-cli ping` before worker |
| `cart.expiry.lock_not_acquired` and `lock_skipped=true` | Another worker holds Redis lock or stale lock has not expired | Wait for `CART_CLEANUP_LOCK_TTL` or inspect lock key locally | Run only one scheduled worker at a time |
| `expired_count=0` | No expired active carts, candidates already TTL-deleted, or carts are not `active` | Check Mongo candidate query | Understand this can be a healthy result |
| Redis cache key still visible after expiry | Cache delete failed, wrong owner key checked, or key recreated by another request | Check `cache_delete_failures` logs and exact key pattern | Monitor Redis failures and keep cache TTLs short |
| TTL index missing | Migration/bootstrap skipped | Run migration or schema bootstrap endpoint | Include migration in local setup and deployments |
| Worker seems to run forever in local terminal | Mongo/Redis call hanging until timeout or huge candidate backlog | Check logs, lower batch size, verify DB health | Keep `CART_CLEANUP_RUN_TIMEOUT` bounded |
| `go: go.mod file not found` | Command run from wrong folder | `cd backend/services/cart-service` | Run worker from service module folder |
| Duration value did not change | Env value missing Go unit, e.g. `15` | Use `15m`, `2m`, `10m`, `720h` | Always include `ms`, `s`, `m`, or `h` |

## 12. Security & Best Practices

### Task-Specific Security

- Never log full cart items, product snapshots, coupon details, email, phone, or address in cleanup logs.
- Current worker logs IDs and counts, which is safer than logging full document payloads.
- Keep Redis private to backend network. Cleanup lock and cache keys should not be accessible from public internet.
- Use Redis auth/TLS in staging and production. Empty Redis password is local-only.
- Keep MongoDB credentials in `.env` locally and a secret manager in production.
- Do not run cleanup with production credentials from a developer laptop.
- If checkout/audit retention matters, review the TTL index policy before production. TTL physical delete is irreversible.

### Task-Specific Best Practices

- Run cleanup as a one-shot scheduled job, not as many always-on replicas doing duplicate work.
- Keep `CART_CLEANUP_LOCK_TTL` greater than `CART_CLEANUP_RUN_TIMEOUT`.
- Keep `CART_CLEANUP_BATCH_SIZE` moderate. `500` is a safe beginner default.
- Alert on `cart.expiry.cleanup_failed`.
- Track `cache_delete_failures`; Redis failures should not silently grow.
- Treat MongoDB as source of truth. Redis is cache only.
- Do not depend on MongoDB TTL timing for user-facing behavior.
- For local Redis debugging, avoid broad `KEYS` commands on shared/prod Redis.
- Keep `CART_BOOTSTRAP_COLLECTIONS_ON_STARTUP=false` in production and run migrations through controlled deployment jobs.
- Until config is split, keep Product/CMS URL env values present for worker startup.

## 13. Missing or Misconfigured Things

Findings from inspected implementation:

| Area | Observation | Risk | Suggested fix |
|---|---|---|---|
| Worker scheduling | `cmd/cart-expiry-worker` runs `RunOnce` and exits | Cleanup will not repeat unless an external scheduler runs it | Add cron/systemd/Kubernetes CronJob in deployment docs/manifests |
| Dockerfile | No dedicated service Dockerfile found | Harder to run worker/API as containers | Add Dockerfile in future DevOps task |
| Compose stack | No complete local compose stack for API plus worker found | Beginners manually run dependencies and worker | Add local compose profile for MongoDB, Redis, API, and one-shot worker |
| Shared config | Worker requires Product/CMS URLs even though it does not use them | Beginner may fail worker startup with unrelated env error | Split worker config validation or make unused dependency config feature-specific |
| Metrics | Metrics interfaces exist but default to no-op | Production observability incomplete | Wire Prometheus/OpenTelemetry metrics in deployment |
| Dry run | `INPUT_FILE_PATH` mentions dry-run idea, but inspected worker has no dry-run flag | Operators cannot preview cleanup impact through CLI | Add `--dry-run` in future if needed |
| TTL retention | TTL index deletes documents with expired `expires_at` regardless of business status | Checked-out/audit retention can be lost if `expires_at` is not managed | Define retention model before production checkout/order integration |
| `.env` loading | Code reads OS env only | `.env` file copied but not sourced causes startup failure | Keep source/export step in docs or add local env loader |
| Readiness | HTTP `/readyz` checks schema/Mongo path, worker has no readiness endpoint | Scheduled job health depends on logs/exit code | Use job exit code, logs, and metrics for worker monitoring |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `4. Go Dependency System` | Same Go version, `go.mod`, `go.sum`, build/test commands |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `5.1 MongoDB` | Same MongoDB installation, Docker setup, URI, migration basics |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `5.2 Redis` | Same Redis installation, Docker setup, verification |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `8. Complete Environment Variables` | Full `.env` creation and complete variable list already documented |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `9. Docker and DevOps Setup` | Base Docker dependency stack reused |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `10. Project Run Instructions` | Clone, source env, run server, health checks reused |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `12. Common Errors and Fixes` | Generic Go/MongoDB/Redis/Docker/env troubleshooting reused |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | `6.2 Redis` | Redis cache key patterns and expiry lock key reused |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | `7. Environment Variables` | DB/cache config mapping reused |
| `PREVIOUS_TASK3_DEPENDENCY_FILE` | `5. Database Setup` | Cart schema, statuses, indexes, and TTL index reused |
| `PREVIOUS_TASK3_DEPENDENCY_FILE` | `9. Local Development Setup` | Migration/bootstrap flow reused |
| `PREVIOUS_TASK4_DEPENDENCY_FILE` | `6.1 Product Service` | Product setup reused only when preparing active carts through API |
| `PREVIOUS_TASK5_DEPENDENCY_FILE` | `11. Common Errors & Fixes` | Mutation-time expired cart/cache behavior reused |
| `PREVIOUS_TASK6_DEPENDENCY_FILE` | `6.1 CMS Service` | CMS URL/env behavior reused by shared config |
| `PREVIOUS_TASK7_DEPENDENCY_FILE` | `7. Environment Variables` | User/guest expiry TTL behavior reused |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before using this file.
- [ ] No duplicate Go/MongoDB/Redis/Docker installation docs added.
- [ ] No new Go dependency added for `${TASK_FILE_NAME}`.
- [ ] MongoDB running and reachable.
- [ ] Redis running and reachable.
- [ ] `.env` created at `backend/services/cart-service/.env`.
- [ ] `.env` sourced/exported before running worker.
- [ ] Product/CMS URL env values present for shared config validation.
- [ ] MongoDB migration or schema bootstrap completed.
- [ ] `idx_carts_expires_at_ttl` index verified.
- [ ] `CART_USER_EXPIRY_TTL` and `CART_GUEST_EXPIRY_TTL` reviewed.
- [ ] `CART_CLEANUP_BATCH_SIZE` positive.
- [ ] `CART_CLEANUP_RUN_TIMEOUT` positive and has duration unit.
- [ ] `CART_CLEANUP_LOCK_TTL` positive and greater than normal run timeout.
- [ ] `CART_CLEANUP_LOCK_KEY` configured.
- [ ] Worker runs with `go run ./cmd/cart-expiry-worker`.
- [ ] Logs show `cart.expiry.worker_finished`.
- [ ] `expired_count`, `scanned_count`, and `cache_delete_failures` reviewed.
- [ ] No new public port expected for worker.
- [ ] Production scheduling gap noted: external cron/CronJob still needed.
- [ ] TTL retention risk reviewed before checkout/audit production use.
- [ ] Real secrets are not committed.
