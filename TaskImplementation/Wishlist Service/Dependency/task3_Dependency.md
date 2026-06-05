# Project Dependency & Setup Guide

## Variable Values

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Wishlist Service` |
| `TASK_FILE_NAME` | `task3.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task3_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

Note: Ye guide `INPUT_FILE_PATH` ko analyze karke banaya gaya hai. Original `TASK_FILE_NAME` modify nahi kiya gaya. Is file ka focus sirf setup, dependency, MongoDB collection readiness, verification, and DevOps onboarding hai.

## 1. Project Overview

`TASK_FILE_NAME` ka main output `wishlists` MongoDB collection ka design hai:

```text
Database   -> wishlist_db
Collection -> wishlists
Purpose    -> one buyer ki private wishlist document store karna
```

Simple Hinglish me: `SERVICE_NAME` me user ke saved products ek MongoDB document ke andar `items` array ke form me store honge. Task 3 ne collection fields, validator, timestamps, and indexes ka shape finalize kiya.

Important reuse note:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`2. Tech Stack`
`3. Required Software`
`4. Dependency Management`
`5. Database Setup`
`6. Redis / Queue / External Services`
`7. Environment Variables`
`8. Docker Setup`
`9. Local Development Setup`
`10. Running the Project`
`11. Ports & Networking`
`12. Common Errors & Fixes`
```

Task 3 me new runtime software add nahi hua. New onboarding value ye hai ki MongoDB `wishlists` collection validator and indexes kaise verify karne hain.

## 2. Tech Stack

Most technology explanation already previous dependency docs me covered hai. Duplicate setup avoid karne ke liye yahan sirf Task 3 status diya gaya hai.

| Technology / Service | Task 3 Status | Required? | Setup Documentation |
|---|---|---:|---|
| Go | Reused | Yes for runnable backend | `task1_Dependency.md` -> `4. Dependency Management` |
| Go modules | Reused | Yes | `task1_Dependency.md` -> `4. Dependency Management` |
| MongoDB | Reused, collection design finalized | Yes | `task1_Dependency.md` -> `5. Database Setup` |
| MongoDB collection validator | Task 3-specific focus | Yes for schema safety | Explained below |
| MongoDB indexes | Task 3-specific verification | Yes | Explained below |
| MongoDB Go Driver v2 | Reused | Yes | `task1_Dependency.md` -> `4. Dependency Management` |
| `mongosh` | Reused debug/setup tool | Recommended | `task1_Dependency.md` -> `5. Database Setup` |
| Docker | Reused local dependency runner | Recommended | `task1_Dependency.md` -> `8. Docker Setup` |
| Kafka-compatible broker | Not introduced by Task 3 | Optional | `task1_Dependency.md` -> `6. Redis / Queue / External Services` |
| Redis | Not introduced by Task 3 | No | No new setup needed |
| RabbitMQ | Not introduced by Task 3 | No | No new setup needed |

### Task 3 New Technical Concept

MongoDB collection validator ek DB-level guard hai. Matlab agar koi developer galat document insert kare, jaise missing `user_id`, wrong `visibility`, wrong price currency, to MongoDB insert/update reject kar sakta hai.

Indexes query speed and uniqueness ke liye hote hain. Is collection me important indexes:

| Index | Why needed |
|---|---|
| `uniq_wishlists_user_id` | One user ke liye one wishlist enforce karne me help karta hai. |
| `idx_wishlists_items_product_id` | Product id based wishlist queries fast karta hai. |
| `idx_wishlists_items_product_variant` | Product plus variant based lookup ke liye useful hai. |
| `idx_wishlists_price_drop_candidates` | Later price-drop candidate scan ke liye useful hai. |

## 3. Required Software

No new software introduced by `TASK_FILE_NAME`.

Follow existing setup:

```md
This setup is already explained in:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`3. Required Software`
```

Task 3-specific required tools:

| Software | Why Task 3 needs it |
|---|---|
| MongoDB server | `wishlist_db.wishlists` collection create/modify/verify karne ke liye. |
| `mongosh` | Validator, collection info, sample insert, and indexes inspect karne ke liye. |
| Go | Runnable service startup ke through `EnsureCollection` verify karne ke liye. |
| Docker | Local MongoDB run karne ka easiest repeatable option. |

## 4. Dependency Management

No new Go module or external library add karne ki zarurat nahi hai for `TASK_FILE_NAME`.

Existing runtime module:

```text
backend/services/wishlist-service/go.mod
```

Existing MongoDB dependency:

```text
go.mongodb.org/mongo-driver/v2
```

Reuse Go dependency commands from:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Dependency Management`
```

Task 3 warning for beginners:

- `TASK_FILE_NAME` is a collection design task, so do not run `go get` for a new Mongo package.
- Current code already uses MongoDB Go Driver v2.
- If `go test ./...` fails due dependency download, use previous dependency guide's Go module troubleshooting.

## 5. Database Setup

### Task 3 Database Target

| Item | Value | Status |
|---|---|---|
| Database type | MongoDB | Reused from Task 2 |
| Database name | `wishlist_db` | Reused default |
| Main collection | `wishlists` | Task 3 focus |
| Event/outbox collection | `wishlist_events` | Existing runtime feature, not Task 3 focus |
| Default Mongo port | `27017` | Reused |
| Main env var | `WISHLIST_MONGO_URI` | Reused |
| DB name env var | `WISHLIST_MONGO_DATABASE` | Reused |
| Collection env var | `WISHLIST_MONGO_COLLECTION` | Reused |

MongoDB installation, Docker command, Docker Compose example, connection string format, and generic DB troubleshooting already documented hain:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`5. Database Setup`
`8. Docker Setup`
```

### What Task 3 Adds

Task 3 adds collection-specific readiness:

| Setup item | File / code path | Meaning |
|---|---|---|
| Collection validator | `backend/services/wishlist-service/migrations/mongo/001_create_wishlists_collection.up.js` | MongoDB me allowed document shape enforce karta hai. |
| Startup collection setup | `backend/services/wishlist-service/internal/usecase/collection_setup.go` | Service start hote waqt Mongo ping and collection setup run karta hai. |
| Repository validator/index setup | `backend/services/wishlist-service/internal/repository/mongo_wishlist_repository.go` | Go code se validator and indexes ensure hote hain. |
| BSON document mapping | `backend/services/wishlist-service/internal/repository/wishlist_document.go` | Go structs MongoDB field names ke saath map hote hain. |

### Collection Shape Verification

Task 3 ke according `wishlists` document me ye required top-level fields hone chahiye:

| Field | Type expectation | Required? | Setup meaning |
|---|---|---:|---|
| `_id` | string | Yes | DB primary id, API me `wishlist_id` ban sakta hai. |
| `user_id` | string | Yes | Authenticated buyer owner. |
| `visibility` | enum `private` | Yes | MVP me public/shared allowed nahi hai. |
| `items` | array | Yes | Saved products list. |
| `created_at` | date | Yes | Creation timestamp. |
| `updated_at` | date | Yes | Last update timestamp. |

Each `items` entry ka expected shape:

| Field | Type expectation | Required? | Note |
|---|---|---:|---|
| `product_id` | string | Yes | Saved product id. |
| `variant_id` | string | No | Variant available ho to store hota hai. |
| `added_at` | date | Yes | Item add timestamp. |
| `last_known_price.amount` | long/int64 | Required only if `last_known_price` exists | Minor unit, jaise paise/cents. |
| `last_known_price.currency` | uppercase 3-letter string | Required only if `last_known_price` exists | Example: `INR`, `USD`. |
| `availability` | enum | No | `unknown`, `in_stock`, `out_of_stock`, `deleted`. |

### Migration Commands

Full migration flow already documented in `task1_Dependency.md`. Task 3-specific minimum migration is:

```bash
cd backend/services/wishlist-service
mongosh "mongodb://localhost:27017" migrations/mongo/001_create_wishlists_collection.up.js
```

For current runtime's complete `wishlists` index set, run the later index migrations also:

```bash
cd backend/services/wishlist-service
mongosh "mongodb://localhost:27017" migrations/mongo/002_add_wishlist_item_variant_index.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/003_add_price_drop_candidate_index.up.js
```

Note: `004_create_wishlist_events_collection.up.js` is for analytics/outbox, not the Task 3 `wishlists` collection design. Run it only when following full runtime setup from previous dependency docs.

### Verify Collection Exists

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.getCollectionNames()"
```

Expected: output should include `wishlists`.

### Verify Validator

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval 'db.getCollectionInfos({ name: "wishlists" })'
```

Look for:

```text
validator
validationLevel: strict
validationAction: error
```

Beginner meaning: MongoDB invalid documents ko allow nahi karega.

### Verify Indexes

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.wishlists.getIndexes()"
```

Expected important indexes:

| Index name | Expected after |
|---|---|
| `_id_` | MongoDB default |
| `uniq_wishlists_user_id` | `001_create_wishlists_collection.up.js` or service startup |
| `idx_wishlists_items_product_id` | `001_create_wishlists_collection.up.js` or service startup |
| `idx_wishlists_items_product_variant` | `002_add_wishlist_item_variant_index.up.js` or service startup |
| `idx_wishlists_price_drop_candidates` | `003_add_price_drop_candidate_index.up.js` or service startup |

### Optional Local Validation Smoke Test

Use this only in local/dev database.

```bash
mongosh "mongodb://localhost:27017/wishlist_db"
```

Inside `mongosh`:

```javascript
db.wishlists.insertOne({
  _id: "wish_local_schema_test",
  user_id: "user_local_schema_test",
  visibility: "private",
  items: [
    {
      product_id: "prod_schema_test",
      variant_id: "var_schema_test",
      added_at: new Date(),
      last_known_price: {
        amount: NumberLong("299900"),
        currency: "INR"
      },
      availability: "in_stock"
    }
  ],
  created_at: new Date(),
  updated_at: new Date()
})

db.wishlists.deleteOne({ _id: "wish_local_schema_test" })
```

If insert succeeds, validator accepts the Task 3 shape. If it fails, check field names, date types, `NumberLong`, and uppercase currency.

## 6. Redis / Queue / External Services

Task 3 does not introduce Redis, Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, OAuth, Kubernetes, or any new external service.

Reuse existing documentation:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`6. Redis / Queue / External Services`
```

Task 3-specific external service status:

| Service | New in Task 3? | Required for collection design? | Note |
|---|---:|---:|---|
| MongoDB | No, reused | Yes | Main `wishlists` collection lives here. |
| Kafka/Redpanda | No | No | Not needed to verify schema validator or indexes. |
| Product Service | No | No | Needed for add-item API flow, not for Task 3 collection setup. |
| Cart Service | No | No | Needed for move-to-cart flow, not for Task 3 collection setup. |
| Redis | No | No | Not used by current `SERVICE_NAME` runtime. |
| RabbitMQ | No | No | Not implemented for current `SERVICE_NAME`. |

For simplest Task 3 local verification, keep Kafka-related env disabled as already explained in `task1_Dependency.md`.

## 7. Environment Variables

No new environment variable introduced by `TASK_FILE_NAME`.

Full `.env` explanation is already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Environment Variables`
```

### Task 3 Relevant Existing Variables

| Variable | Required? | Default | Task 3 meaning |
|---|---:|---|---|
| `WISHLIST_MONGO_URI` | Yes | `mongodb://localhost:27017` | MongoDB server connection. |
| `WISHLIST_MONGO_DATABASE` | Yes | `wishlist_db` | DB where `wishlists` collection lives. |
| `WISHLIST_MONGO_COLLECTION` | Yes | `wishlists` | Main collection name for Task 3. |
| `WISHLIST_MONGO_CONNECT_TIMEOUT` | Optional | `5s` | Mongo connection timeout. |
| `WISHLIST_MONGO_SERVER_SELECTION_TIMEOUT` | Optional | `5s` | Mongo server selection timeout. |
| `WISHLIST_ANALYTICS_EVENTS_ENABLED` | Optional | `true` | Can be `false` for schema-only local verification. |
| `WISHLIST_EVENT_PUBLISHER` | Optional | `kafka` | Can be `disabled` for beginner local mode. |

### Task 3 Minimal Local Env Delta

This is not a full `.env`. Use it only as the Mongo/schema-specific part and keep the full env guidance from `task1_Dependency.md`.

```env
WISHLIST_MONGO_URI=mongodb://localhost:27017
WISHLIST_MONGO_DATABASE=wishlist_db
WISHLIST_MONGO_COLLECTION=wishlists

WISHLIST_ANALYTICS_EVENTS_ENABLED=false
WISHLIST_EVENT_PUBLISHER=disabled
```

Important: Current Go code reads OS environment variables. It does not automatically load `.env` files. Create `.env` at:

```text
backend/services/wishlist-service/.env
```

Then load it as described in `task1_Dependency.md`.

### Credentials Placement

Credentials placement is reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Environment Variables`
Topic:
`.env loading and security notes`
```

Task 3-specific security note: If `WISHLIST_MONGO_URI` contains username/password, keep it in local uncommitted `.env`, Docker/Kubernetes Secret, or secret manager. Do not commit real credentials.

## 8. Docker Setup

No new Dockerfile, container, network, volume, or compose service introduced by `TASK_FILE_NAME`.

Reuse:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`8. Docker Setup`
```

Task 3-specific Docker notes:

| Docker item | Status | Meaning |
|---|---|---|
| MongoDB container | Reused | Needed if MongoDB not installed locally. |
| MongoDB volume | Reused | Keeps `wishlist_db` data after container restart. |
| Kafka/Redpanda container | Optional, reused | Not needed for Task 3 schema verification. |
| Service container | Not added by Task 3 | Current beginner path runs Go service directly. |

If service runs on host, use:

```env
WISHLIST_MONGO_URI=mongodb://localhost:27017
```

If service later runs inside the same compose network as MongoDB, use the compose service name:

```env
WISHLIST_MONGO_URI=mongodb://wishlist-mongo:27017
```

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Read these first:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
```

Minimum sections:

- `task1_Dependency.md` -> `3. Required Software`
- `task1_Dependency.md` -> `5. Database Setup`
- `task1_Dependency.md` -> `7. Environment Variables`
- `task1_Dependency.md` -> `9. Local Development Setup`
- `task2_Dependency.md` -> `5. Database Setup`

### Step 2: Go to Project Directory

```bash
cd Ecommerce
cd backend/services/wishlist-service
```

### Step 3: Install Only New Dependencies

None for Task 3.

If dependencies were never downloaded:

```bash
go mod download
```

This command is reused from `task1_Dependency.md`.

### Step 4: Setup Only New Databases/Services

No new database server. Use MongoDB setup from previous dependency guide.

Task 3 requires MongoDB to be reachable:

```bash
mongosh "mongodb://localhost:27017" --eval "db.runCommand({ ping: 1 })"
```

### Step 5: Add Only New or Changed Environment Variables

No new env vars.

Make sure existing Mongo env values point to Task 3 target:

```env
WISHLIST_MONGO_URI=mongodb://localhost:27017
WISHLIST_MONGO_DATABASE=wishlist_db
WISHLIST_MONGO_COLLECTION=wishlists
```

### Step 6: Run Migrations If Needed

Manual Task 3 route:

```bash
mongosh "mongodb://localhost:27017" migrations/mongo/001_create_wishlists_collection.up.js
```

Current runtime route:

```bash
go run ./cmd/server
```

The service startup calls collection setup and ensures validator/indexes. For controlled environments, prefer explicit migrations and deployment runbooks.

### Step 7: Start Backend Service

Reuse run instructions from:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`10. Running the Project`
```

Minimum command:

```bash
go run ./cmd/server
```

### Step 8: Verify APIs or Functionality Related to `TASK_FILE_NAME`

Task 3 is schema-focused, so verify MongoDB rather than Product/Cart API flows:

```bash
curl http://localhost:8084/readyz
mongosh "mongodb://localhost:27017/wishlist_db" --eval 'db.getCollectionInfos({ name: "wishlists" })'
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.wishlists.getIndexes()"
```

Expected:

- `/readyz` returns healthy when MongoDB is reachable.
- `wishlists` collection exists.
- Validator is present with strict/error settings.
- Required indexes exist.

## 10. Running the Project

No new run command introduced by `TASK_FILE_NAME`.

Reuse:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`10. Running the Project`
```

Task 3-specific running modes:

| Mode | MongoDB required? | Kafka required? | Best for |
|---|---:|---:|---|
| Schema-only local verification | Yes | No | Validator and indexes check karna. |
| Full runtime local service | Yes | Optional | Health/readiness plus collection setup verify karna. |
| Event-enabled runtime | Yes | Yes | Later event tasks, not required by Task 3. |

For Task 3, a beginner should start with schema-only local verification, then run the service once to confirm startup setup also works.

## 11. Ports & Networking

Detailed networking explanation is reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`11. Ports & Networking`
```

Task 3 port table:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| `SERVICE_NAME` HTTP server | `8084` | Readiness endpoint and service startup verification | Reused |
| MongoDB | `27017` | `wishlist_db.wishlists` storage | Reused and required |
| Product Service | `8082` | Add-item validation in later/runtime flow | Reused, not required for Task 3 |
| Cart Service | `8083` | Move-to-cart in later/runtime flow | Reused, not required for Task 3 |
| Kafka-compatible broker | `9092` | Events in later/runtime flow | Reused, optional |

No port changed by `TASK_FILE_NAME`.

Task 3 networking reminder:

- Host-run service should use `mongodb://localhost:27017`.
- Container-run service should use Docker service DNS, for example `mongodb://wishlist-mongo:27017`.
- If `mongosh` can connect but `/readyz` fails, compare shell env `WISHLIST_MONGO_URI` with the URI used in `mongosh`.

## 12. Common Errors & Fixes

Generic setup errors are already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`12. Common Errors & Fixes`
```

Task 3-specific errors:

| Error / Confusion | Cause | Fix | Prevention |
|---|---|---|---|
| `Document failed validation` | Insert/update document does not match MongoDB validator. | Check required fields, date types, `visibility`, `last_known_price`, and `availability`. | Use service APIs or validated sample document. |
| `visibility: "public"` fails | Task 3 MVP validator allows only `private`. | Use `visibility: "private"`. | Do not enable public/shared wishlist until a later schema migration. |
| Price insert fails | `last_known_price.amount` is not BSON long or `currency` is lowercase. | Use `NumberLong("299900")` and uppercase `INR`. | Store money as integer minor unit and normalize currency. |
| `E11000 duplicate key` on `uniq_wishlists_user_id` | Same `user_id` already has a wishlist. | Reuse existing wishlist document instead of inserting another. | Keep one-user-one-wishlist rule in service logic. |
| Expected index missing | Migration not run or startup setup not executed successfully. | Run migration or start service and check logs. | Verify `db.wishlists.getIndexes()` during setup. |
| `IndexOptionsConflict` | Local DB has same index name with different keys/options from old experiments. | In local only, drop the wrong index and rerun migration. In shared/prod, create a controlled migration. | Avoid manual index experiments on shared DB. |
| `collMod` unauthorized | Mongo user lacks permission to modify collection validator. | Use a DB user with `dbAdmin`/migration permissions for setup. | Separate app runtime user and migration/admin user in production. |
| Migration ran on `wishlist_db`, but service uses another DB | JS migrations hardcode `wishlist_db`; env can change runtime DB. | Keep local `WISHLIST_MONGO_DATABASE=wishlist_db`, or make migrations parameterized later. | Do not change DB name casually after docs/migrations are created. |
| `wishlists` collection exists but validator absent | Collection was created manually before migration/startup setup. | Run `001_create_wishlists_collection.up.js` or start service so `collMod` applies validator. | Always create collection through migration or service setup. |
| Duplicate product appears inside `items` | DB validator checks shape, not embedded-array uniqueness. | Use service add-item flow, which blocks duplicate product IDs. | Do not bypass service logic with direct DB writes. |

## 13. Security & Best Practices

### Task 3 Security Audit

| Area | Observation | Risk | Suggested fix |
|---|---|---|---|
| Mongo credentials | No Task 3 hardcoded credential requirement. | Local no-auth URI is unsafe if copied to shared/prod env. | Use authenticated Mongo URI and secrets outside git. |
| Collection validator | Validator enforces basic shape. | It does not enforce all business rules, like max item count or trusted ownership. | Keep application/domain validation in Go code. |
| `user_id` ownership | Collection stores `user_id`. | If service trusts public headers directly, ownership can be spoofed. | Keep service behind trusted API Gateway/private network. |
| Direct DB writes | Manual inserts can bypass product validation and duplicate logic. | Bad data can enter DB if admin writes directly. | Use service APIs or controlled migrations/seed scripts. |
| Price snapshot | `last_known_price` is stored in wishlist item. | It can become stale and must not be checkout source of truth. | Revalidate price through Product/Checkout flow. |
| Migration permissions | Validator/index setup needs elevated DB permissions. | Runtime app user with too many permissions increases blast radius. | Use separate migration/admin credential in production. |
| DB/collection env override | Runtime names can differ from hardcoded JS migration names. | Service may run against a DB without expected validator/indexes. | Keep defaults aligned or parameterize migrations later. |

### Task 3 Best Practices

- Keep `WISHLIST_MONGO_DATABASE=wishlist_db` and `WISHLIST_MONGO_COLLECTION=wishlists` for local development unless there is a clear reason to change.
- Run `db.wishlists.getIndexes()` after migrations and before API testing.
- Store timestamps in UTC.
- Store money in minor units as int64/long, not float.
- Keep currency uppercase 3-letter format.
- Keep `visibility` as `private` until a future migration intentionally supports public/shared wishlist.
- Use service code for add/remove flows instead of direct DB writes.
- Add a max wishlist item limit in business logic before production scale.
- Treat Product Service as source of truth for product status, price, and inventory.
- Keep migration scripts reviewed and versioned.

## 14. Missing or Misconfigured Things

| Item | Why it matters | Recommended action |
|---|---|---|
| Migration JS files hardcode `wishlist_db` and `wishlists` | Env vars can point service to a different DB/collection. | Keep defaults aligned locally; later make migrations env-aware or use a migration tool. |
| DB validator does not enforce embedded item uniqueness | MongoDB validator checks shape only. | Keep duplicate prevention in service logic; add tests around duplicate product behavior. |
| No Task 3-specific `.env.example` added | Fresh developers may not know minimal schema verification env. | Add/update service `.env.example` in a separate implementation task. |
| Analytics collection can be created during full startup | `wishlist_events` is runtime feature, but Task 3 focuses on `wishlists`. | Disable analytics for schema-only local setup if you want less moving parts. |
| No service Dockerfile/compose introduced by Task 3 | Containerized service startup remains manual. | Reuse previous Docker dependency setup; add service Dockerfile later if deployment task requires it. |
| Public/shared visibility not supported by validator | Future sharing feature would fail current schema. | Add a planned migration before implementing shared/public wishlist. |
| Direct DB sample data can bypass Product Service validation | Product id may not actually exist. | Use sample inserts only for local validator testing, then delete them. |

## 15. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack` | Go, MongoDB, Kafka, Docker, Product/Cart dependencies already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Required Software` | Git, Go, MongoDB, `mongosh`, Docker, curl setup already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Dependency Management` | Go module commands and dependency troubleshooting already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Setup` | MongoDB install, Docker run, connection strings, and full migration flow already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Redis / Queue / External Services` | Kafka/Redpanda, Product Service, Cart Service, and Redis status already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Environment Variables` | Full `.env` example, loading behavior, and env variable reference already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Docker Setup` | Dependency container setup, volumes, and Docker networking already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Local Development Setup` | Clone, dependency install, Mongo startup, env load, tests, and service start flow already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `11. Ports & Networking` | Port conflicts and host-vs-container Mongo URI behavior already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `12. Common Errors & Fixes` | Generic setup troubleshooting already documented. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | MongoDB decision, `wishlist_db`, `wishlists`, and index expectation already documented. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `7. Environment Variables` | Task 2 Mongo env relevance and migration hardcoding warning already documented. |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `14. Missing or Misconfigured Things` | Runtime/config mismatch and migration naming risk already documented. |

## 16. Final Checklist

- [ ] Previous dependency files checked before using this guide.
- [ ] No duplicate MongoDB installation guide added.
- [ ] No duplicate full `.env` explanation added.
- [ ] No new Go dependency added for `TASK_FILE_NAME`.
- [ ] MongoDB is running and reachable.
- [ ] `WISHLIST_MONGO_URI` points to the correct MongoDB server.
- [ ] `WISHLIST_MONGO_DATABASE=wishlist_db` confirmed for local setup.
- [ ] `WISHLIST_MONGO_COLLECTION=wishlists` confirmed for local setup.
- [ ] `001_create_wishlists_collection.up.js` run, or service startup collection setup verified.
- [ ] `wishlists` collection exists.
- [ ] Collection validator is present with `strict` and `error` settings.
- [ ] `uniq_wishlists_user_id` index exists.
- [ ] `idx_wishlists_items_product_id` index exists.
- [ ] Variant and price-drop indexes checked if full runtime setup is needed.
- [ ] Optional local validation sample inserted and deleted.
- [ ] `/readyz` checked after starting service.
- [ ] Kafka disabled for schema-only beginner mode, or broker running for event-enabled mode.
- [ ] Product/Cart services not treated as mandatory for Task 3 schema verification.
- [ ] Real MongoDB credentials are not committed.
- [ ] Original `TASK_FILE_NAME` left unchanged.
