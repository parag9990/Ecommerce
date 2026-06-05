# Project Dependency & Setup Guide

## Variable Values

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Wishlist Service` |
| `TASK_FILE_NAME` | `task2.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task2_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

Note: Ye guide `INPUT_FILE_PATH` ke Task 2 content ko analyze karke banaya gaya hai. Original implementation file modify nahi ki gayi.

## 1. Project Overview

Task 2 ka decision:

```text
Primary database for SERVICE_NAME = MongoDB
Database name = wishlist_db
Primary collection = wishlists
```

Simple Hinglish me: wishlist data ek buyer-owned list hai. User mostly apni full wishlist read karta hai, aur item fields future me flexible ho sakte hain, jaise `last_known_price`, `availability`, `variant_id`, display snapshots, etc. Isliye MongoDB ka document model is task ke liye fit hai.

Important reuse note:

Most runtime setup already documented hai:

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

Is file me duplicate installation commands repeat nahi kiye gaye. Sirf Task 2 ke MongoDB decision, config deltas, audit points, and verification focus explain kiya gaya hai.

## 2. Tech Stack

| Technology / Service | Status in Task 2 | Required? | Setup Documentation |
|---|---|---:|---|
| Go | Reused | Yes for runnable backend | `task1_Dependency.md` -> `Dependency Management` |
| Go modules | Reused | Yes | `task1_Dependency.md` -> `Dependency Management` |
| MongoDB | Confirmed as primary DB | Yes | Setup reused, decision-specific notes below |
| MongoDB Go Driver v2 | Reused runtime dependency | Yes | `task1_Dependency.md` -> `Dependency Management` |
| `mongosh` | Reused setup/debug tool | Recommended | `task1_Dependency.md` -> `Database Setup` |
| Docker | Reused local dependency runner | Recommended | `task1_Dependency.md` -> `Docker Setup` |
| Kafka-compatible broker | Reused optional event service | Optional for Task 2 | `task1_Dependency.md` -> `Redis / Queue / External Services` |
| Redis | Not introduced by Task 2 | No | No new setup needed |
| RabbitMQ | Not introduced by Task 2 | No | No new setup needed |

### MongoDB Decision Explanation

MongoDB ek document database hai. SQL tables ke bajay ye JSON-like/BSON documents store karta hai. `SERVICE_NAME` me ek buyer ki wishlist naturally ek document ban sakti hai:

```text
one user -> one private wishlist document -> embedded items array
```

Task 2 ka new responsibility setup install karna nahi, balki ye decision lock karna hai ki `SERVICE_NAME` ka durable source of truth MongoDB rahega.

## 3. Required Software

No new software introduced by `TASK_FILE_NAME`.

Follow existing setup:

```md
This setup is already explained in:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`3. Required Software`
```

Task 2-specific checklist:

| Software | Why Task 2 cares |
|---|---|
| MongoDB server | Decision verify karne ke liye primary DB must be available. |
| `mongosh` | `wishlist_db`, `wishlists`, and indexes inspect karne ke liye useful. |
| Docker | Local MongoDB run karne ka easiest path, already documented. |

## 4. Dependency Management

No new Go dependency add karne ki zarurat nahi hai for Task 2.

Existing runtime module:

```text
backend/services/wishlist-service/go.mod
```

Important actual dependency:

```text
go.mongodb.org/mongo-driver/v2 v2.6.0
```

Warning for beginners:

`TASK_FILE_NAME` me conceptual future example old import path jaisa lag sakta hai:

```text
go.mongodb.org/mongo-driver/mongo
```

Current code already MongoDB Go Driver v2 use karta hai:

```text
go.mongodb.org/mongo-driver/v2
```

Isliye manually old v1 package install mat karo. Agar dependency commands chahiye, reuse:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Dependency Management`
```

## 5. Database Setup

### Task 2 Database Decision

| Item | Value | Status |
|---|---|---|
| Database type | MongoDB | Confirmed by `TASK_FILE_NAME` |
| Runtime DB name | `wishlist_db` | Reused existing code default |
| Main collection | `wishlists` | Reused existing code default |
| Analytics/outbox collection | `wishlist_events` | Existing runtime feature, not new for Task 2 |
| Default port | `27017` | Reused |
| Connection env var | `WISHLIST_MONGO_URI` | Reused |

### Why MongoDB Is Mandatory Here

| Reason | Simple explanation |
|---|---|
| Wishlist shape is document-friendly | User ki list ek document ke andar `items` array ke form me aa sakti hai. |
| Read-heavy access | Common flow user ki full wishlist read karna hai, so `user_id` se fast lookup useful hai. |
| Flexible fields | Price snapshot, availability, variant, future metadata add karna easy hai. |
| Microservice ownership | `SERVICE_NAME` apna `wishlist_db` own karega; Product DB direct query nahi karega. |
| Existing repo alignment | Docs and runtime code already `wishlist_db` + `wishlists` direction use kar rahe hain. |

### Setup Reuse

MongoDB install, Docker run, Docker Compose example, `mongosh` verification, migrations, and connection string formats already documented hain.

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`5. Database Setup`
`8. Docker Setup`
`9. Local Development Setup`
```

### Task 2 Verification Focus

After following previous setup, verify MongoDB decision-specific pieces:

```bash
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.runCommand({ ping: 1 })"
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.wishlists.getIndexes()"
```

Expected important indexes:

| Index name | Purpose |
|---|---|
| `uniq_wishlists_user_id` | One buyer ke liye one wishlist enforce karne me help. |
| `idx_wishlists_items_product_id` | Product id based wishlist checks/removal fast karne ke liye. |
| `idx_wishlists_items_product_variant` | Product + variant combination queries ke liye. |
| `idx_wishlists_price_drop_candidates` | Future price-drop processing ke liye. |

### Migration Caution

Current migration JS files hardcode:

```text
wishlist_db
wishlists
wishlist_events
```

Runtime Go config can use env vars:

```text
WISHLIST_MONGO_DATABASE
WISHLIST_MONGO_COLLECTION
WISHLIST_EVENT_OUTBOX_COLLECTION
```

Beginner note: Agar aap env me DB/collection names change karte ho, manual migration scripts abhi bhi hardcoded `wishlist_db` par run honge. Local development me defaults use karo. Production me migration strategy ko env-aware banana better hoga.

## 6. Redis / Queue / External Services

Task 2 does not introduce Redis, Kafka, RabbitMQ, S3, SMTP, Stripe, Twilio, OAuth, Kubernetes, or any new third-party service.

Reuse existing documentation:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`6. Redis / Queue / External Services`
```

Task 2-specific decision:

| Service | New in Task 2? | Note |
|---|---:|---|
| Redis | No | Not required for MongoDB decision. |
| Kafka/Redpanda | No | Optional runtime events are already documented; DB choice does not require Kafka. |
| RabbitMQ | No | Not implemented in current `SERVICE_NAME` code. |
| Product Service | No | Future validation boundary stays API/gRPC/events, not direct Product DB access. |
| Cart Service | No | Move-to-cart is later runtime flow, not Task 2 setup. |

## 7. Environment Variables

No new environment variable introduced by `TASK_FILE_NAME`.

Do not duplicate the full `.env` here. Reuse:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Environment Variables`
```

### Task 2 Relevant Existing Variables

| Variable | Required? | Default | Task 2 meaning |
|---|---:|---|---|
| `WISHLIST_MONGO_URI` | Yes | `mongodb://localhost:27017` | MongoDB connection string. |
| `WISHLIST_MONGO_DATABASE` | Yes | `wishlist_db` | DB name chosen for `SERVICE_NAME`. |
| `WISHLIST_MONGO_COLLECTION` | Yes | `wishlists` | Primary collection chosen for wishlist documents. |
| `WISHLIST_MONGO_CONNECT_TIMEOUT` | Optional | `5s` | Mongo connect timeout. |
| `WISHLIST_MONGO_SERVER_SELECTION_TIMEOUT` | Optional | `5s` | Mongo server selection timeout. |
| `WISHLIST_EVENT_OUTBOX_COLLECTION` | Required if analytics enabled | `wishlist_events` | Existing outbox collection, not a new Task 2 variable. |

### Env Name Mismatch Warning

`TASK_FILE_NAME` has conceptual future examples like:

```env
WISHLIST_SERVICE_PORT=50054
WISHLIST_MONGO_CONNECT_TIMEOUT_SECONDS=10
```

Current runnable code uses:

```env
WISHLIST_HTTP_ADDR=:8084
WISHLIST_MONGO_CONNECT_TIMEOUT=5s
```

So for actual local run, follow the runtime env names from `task1_Dependency.md`, not the conceptual names in Task 2.

### Credentials Placement

Credentials placement is reused from previous docs:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Environment Variables`
Topic:
`.env loading and security notes`
```

Task 2-specific security reminder: MongoDB URI may contain username/password. Real credentials should be in local uncommitted `.env`, Docker/Kubernetes secrets, or a secret manager. Never commit production MongoDB credentials.

## 8. Docker Setup

No new Docker container or Dockerfile introduced by Task 2.

Reuse:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`8. Docker Setup`
```

Task 2-specific note:

`TASK_FILE_NAME` mentions conceptual:

```bash
docker compose up -d mongo
```

Current repo snapshot does not show a root compose service named `mongo` for this service. Use the previous dependency guide's local Mongo container or dependency compose example unless a platform compose file is added later.

## 9. Local Development Setup

Complete beginner onboarding flow is already documented in the previous dependency guide. For Task 2, use this incremental flow:

### Step 1: Read previous dependency documentation

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Read these sections first:

- `3. Required Software`
- `4. Dependency Management`
- `5. Database Setup`
- `7. Environment Variables`
- `9. Local Development Setup`
- `10. Running the Project`

### Step 2: Go to project directory

Use the repository root, then the runtime service folder:

```bash
cd Ecommerce
cd backend/services/wishlist-service
```

### Step 3: Install only new dependencies

None for Task 2.

### Step 4: Setup only new databases/services

No new database beyond MongoDB. MongoDB setup is reused from `task1_Dependency.md`.

### Step 5: Add only new or changed environment variables

None for Task 2. Ensure existing Mongo variables point to:

```env
WISHLIST_MONGO_URI=mongodb://localhost:27017
WISHLIST_MONGO_DATABASE=wishlist_db
WISHLIST_MONGO_COLLECTION=wishlists
```

### Step 6: Run migrations if needed

Migration commands are already documented in `task1_Dependency.md`.

Task 2-specific expectation:

```text
wishlists collection must have user_id and items.product_id indexes.
```

### Step 7: Start backend service

Use the run instructions from `task1_Dependency.md`.

### Step 8: Verify functionality related to `TASK_FILE_NAME`

Task 2 is a DB decision, so verify DB readiness and index direction:

```bash
curl http://localhost:8084/readyz
mongosh "mongodb://localhost:27017/wishlist_db" --eval "db.wishlists.getIndexes()"
```

## 10. Running the Project

No new run command introduced by Task 2.

Reuse:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`10. Running the Project`
```

Task 2-specific local mode:

| Mode | MongoDB required? | Kafka required? | Best for |
|---|---:|---:|---|
| Minimum local run | Yes | No | Beginner verification of DB decision and health checks. |
| Event-enabled run | Yes | Yes | Later event tasks, not required for Task 2. |

## 11. Ports & Networking

Detailed networking setup is reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`11. Ports & Networking`
```

Task 2-specific port table:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| `SERVICE_NAME` HTTP server | `8084` | Health and wishlist APIs | Reused |
| MongoDB | `27017` | Primary wishlist database | Reused and required |
| Product Service | `8082` | Product validation in later add-item flow | Reused |
| Cart Service | `8083` | Move-to-cart in later flow | Reused |
| Kafka-compatible broker | `9092` | Optional event processing | Reused, optional |

No port changed by `TASK_FILE_NAME`.

## 12. Common Errors & Fixes

Generic setup issues are already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`12. Common Errors & Fixes`
```

Task 2-specific errors:

| Error / Confusion | Cause | Fix | Prevention |
|---|---|---|---|
| `go get go.mongodb.org/mongo-driver/mongo` adds unexpected old dependency | Task 2 conceptual example uses older path style. | Do not add it. Current module already uses `go.mongodb.org/mongo-driver/v2`. | Check `go.mod` before installing anything. |
| Manual migration ran on `wishlist_db`, but service uses another DB | Migration scripts hardcode `wishlist_db`; env can override runtime DB. | Keep `WISHLIST_MONGO_DATABASE=wishlist_db` locally, or update migration strategy for changed DB names. | Align env values and migration scripts before changing DB name. |
| `docker compose up -d mongo` fails | No compose service named `mongo` exists in current repo snapshot. | Use Docker run / compose example from `task1_Dependency.md`. | Add official local compose file later. |
| Developer expects Redis for wishlist | Other services use Redis, but Task 2 chose MongoDB only for wishlist persistence. | Do not install Redis for Task 2. | Follow service-specific docs, not generic platform assumptions. |
| Developer tries to query Product DB directly from `SERVICE_NAME` | Misunderstanding microservice ownership. | Use Product Service API/gRPC/events, not direct DB access. | Keep database ownership rule documented. |

## 13. Security & Best Practices

### Task 2 Security Audit

| Area | Observation | Risk | Suggested fix |
|---|---|---|---|
| MongoDB credentials | No hardcoded production credentials found. | Local no-auth URI is unsafe if reused in shared/prod env. | Use authenticated URI and secrets outside git. |
| DB ownership | Task 2 correctly says Product data should not be directly read from Product DB. | Cross-service DB reads create coupling and security gaps. | Keep API/gRPC/events as integration boundary. |
| Migration config | JS migrations hardcode DB/collection names. | Env-driven deployments can drift from migration target. | Make migrations parameterized or keep names fixed per environment. |
| Conceptual vs runtime env names | Task 2 examples differ from actual Go config. | Beginners may set unused variables. | Use actual runtime env names from `internal/config/config.go`. |
| Mongo public exposure | Local Docker maps `27017` to host. | Fine local, risky on public/shared machines. | Bind to local/private networks and enable auth in shared env. |

### Task 2 Best Practices

- Keep MongoDB as the source of truth for wishlist data.
- Keep Product Service as source of truth for product title, price, inventory, and availability.
- Store only safe snapshots in wishlist documents; checkout should revalidate price and stock.
- Keep `user_id` from trusted auth context, not request body.
- Keep one wishlist per buyer enforced through service logic plus `uniq_wishlists_user_id`.
- Keep wishlist item arrays bounded in later business rules to avoid very large documents.
- Do not change `WISHLIST_MONGO_DATABASE` or `WISHLIST_MONGO_COLLECTION` casually once real data exists.
- Document every schema/index change in a migration file and setup doc.

## 14. Missing or Misconfigured Things

| Item | Why it matters | Recommended action |
|---|---|---|
| Task 2 conceptual env names differ from actual runtime env names | Beginners may set `WISHLIST_SERVICE_PORT` but service listens using `WISHLIST_HTTP_ADDR`. | Prefer actual env names from `backend/services/wishlist-service/internal/config/config.go`. |
| Task 2 says schema/index implementation deferred, but repo already has migrations and collection setup | Docs and runtime implementation are at different maturity levels. | Keep Task 2 as decision doc, and runtime setup in dependency docs. |
| Migration scripts hardcode `wishlist_db` | Changing env DB name will not automatically change manual migration target. | Parameterize migrations or keep defaults consistent. |
| No new Docker Compose service added by Task 2 | `docker compose up -d mongo` may not work until platform compose exists. | Use prior dependency file's Docker guidance. |
| No new `.env.example` added by Task 2 | Setup can still be confusing for fresh developers. | Add/update service `.env.example` in a separate implementation task if needed. |

## 15. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack` | Go, MongoDB Go Driver, Kafka, Docker, and runtime stack already explained. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Required Software` | Git, Go, MongoDB, `mongosh`, Docker, and curl setup already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Dependency Management` | Go module commands and dependency failure fixes already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Setup` | MongoDB installation, Docker run, connection strings, migrations, and verification already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Redis / Queue / External Services` | Kafka/Redpanda, Product Service, Cart Service, and Docker external service notes already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Environment Variables` | Full `.env` example and env variable reference already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Docker Setup` | Dependency container compose example already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Local Development Setup` | Clone, install, Mongo start, env load, migration, test, and run flow already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `11. Ports & Networking` | Port conflicts and Docker networking already documented. |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `12. Common Errors & Fixes` | Generic setup troubleshooting already documented. |

## 16. Final Checklist

- [ ] Previous dependency documentation checked.
- [ ] No duplicate MongoDB install guide added.
- [ ] No duplicate full `.env` example added.
- [ ] MongoDB confirmed as primary database for `SERVICE_NAME`.
- [ ] `wishlist_db` confirmed as default database.
- [ ] `wishlists` confirmed as default main collection.
- [ ] MongoDB Go Driver v2 confirmed from `go.mod`.
- [ ] No new Go dependency introduced by `TASK_FILE_NAME`.
- [ ] No new Redis/Kafka/RabbitMQ requirement introduced by `TASK_FILE_NAME`.
- [ ] Existing Mongo env variables verified.
- [ ] Conceptual Task 2 env names checked against actual runtime env names.
- [ ] Migration hardcoding risk documented.
- [ ] Docker setup reused from previous dependency file.
- [ ] Task-specific MongoDB verification commands included.
- [ ] Security and best-practice notes added.
- [ ] Original `TASK_FILE_NAME` left unchanged.
