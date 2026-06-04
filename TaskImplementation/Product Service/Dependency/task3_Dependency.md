# Project Dependency & Setup Guide

## Variable Values Used

```text
SERVICE_NAME=Product Service
TASK_FILE_NAME=task3.md
INPUT_FILE_PATH=TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME=task3_Dependency.md
OUTPUT_FILE_PATH=TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

This file is generated for `INPUT_FILE_PATH` and saved as `OUTPUT_FILE_PATH`.

Beginner note:

- `TASK_FILE_NAME` is a documentation/design task for MongoDB collection structure.
- It does not add a new API server, Dockerfile, docker-compose file, or external runtime service.
- Shared Go, MongoDB, Docker, RabbitMQ, and environment setup is already covered in previous dependency docs, so this file references those sections instead of repeating them.

## 1. Project Overview

`TASK_FILE_NAME` defines the core MongoDB collections for the catalog database:

| Collection | Purpose |
|---|---|
| `products` | Main product catalog, variants, media metadata, seller ownership, status |
| `categories` | Category tree and category-wise attribute schema |
| `brands` | Brand master data |
| `inventory_snapshots` | Stock audit/history snapshots |
| `price_books` | Scheduled/current pricing rules |

Simple Hinglish explanation:

Is task ka focus code run karna nahi hai. Is task ka focus yeh decide karna hai ki MongoDB me product catalog ka data kis collection me, kis shape me, aur kaunse indexes ke saath store hoga.

### Current implementation reality

| Area | Status | Setup impact |
|---|---|---|
| `TASK_FILE_NAME` | Collection design guide | No new install command just for reading the doc |
| Core MongoDB collections | Implemented as migration `0003_create_product_collections.up.js` | Run this migration for `TASK_FILE_NAME` DB verification |
| Current Go collection definitions | Present in `internal/repository/mongo_collection_definitions.go` | Can auto-create collections only if app wiring provides a manager |
| Current service entrypoint | Not present | `go run .` is still not a valid service startup command |
| Dockerfile / compose | Not present | Use previous docs for manual local MongoDB setup |
| Later collections | Present from later tasks: `inventory_reservations`, `product_event_outbox` | Do not confuse them with `TASK_FILE_NAME` core scope |

## 2. Tech Stack

Full technology explanation is already available here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
3. Project Tech Stack Analysis
4. Language-Specific Dependency System: Go
5. Database Analysis: MongoDB
7. External Services Analysis
```

Task-specific summary:

| Technology | Required? | Why it matters for `TASK_FILE_NAME` | New explanation needed? |
|---|---:|---|---|
| Go | Yes for current repository code | Collection definitions and repository code are written in Go | No, reused |
| Go modules | Yes | `go.mod` / `go.sum` manage Go dependencies | No, reused |
| MongoDB | Yes for real DB-backed verification | `TASK_FILE_NAME` designs MongoDB collections | Partial, collection-specific details below |
| MongoDB Go Driver v2 | Yes in current implementation | Used by repository and collection manager code | No, setup reused |
| `mongosh` | Required for manual migration execution | Runs `migrations/mongo/*.js` files | No, install/setup reused |
| Docker | Optional but recommended | Easiest way to run local MongoDB | No, setup reused |
| RabbitMQ | No for `TASK_FILE_NAME` | Only relevant for later event/outbox code | No, reused only if full current service is tested |
| Redis/Kafka/Typesense | No direct `TASK_FILE_NAME` requirement | Not needed to create `TASK_FILE_NAME` collections | No |

## 3. Required Software

No brand-new software is introduced by `TASK_FILE_NAME`.

Please follow the earlier setup first:

```text
This setup is already explained in:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
4. Language-Specific Dependency System: Go
5. Database Analysis: MongoDB
6. Environment Variables
9. Docker and DevOps Setup
10. Migration Setup
```

For this task, a beginner needs these tools only if they want to verify the collections locally:

| Software | Purpose | Status |
|---|---|---|
| Go `1.26.3` | Build/test current packages | Reused |
| MongoDB | Holds `product_db` collections | Reused, mandatory for DB verification |
| `mongosh` | Runs the `TASK_FILE_NAME` migration script | Reused |
| Docker | Local MongoDB container option | Reused, optional |

## 4. Dependency Management

Go dependency management is already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
4. Language-Specific Dependency System: Go
```

### What changed in `TASK_FILE_NAME`?

No new Go package is required only because of the collection-design document.

Current module files remain:

```text
backend/services/product-service/go.mod
backend/services/product-service/go.sum
backend/go.work
```

Current direct dependencies already include:

| Dependency | Why used |
|---|---|
| `go.mongodb.org/mongo-driver/v2` | Official MongoDB driver used by repository and collection manager code |
| `github.com/rabbitmq/amqp091-go` | Later event/outbox publisher dependency, not `TASK_FILE_NAME`-specific |

Do not run `go get` just for `TASK_FILE_NAME`. Agar dependencies missing hain, previous doc ke `go mod download` / `go mod tidy` commands follow karo.

## 5. Database Setup

MongoDB setup basics are already explained here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
5. Database Analysis: MongoDB
```

Database used by this task:

```text
product_db
```

### What is new for `TASK_FILE_NAME`?

`TASK_FILE_NAME` introduces the core collection schema and indexes. Ye installation nahi hai, ye database structure hai.

| Collection | Required for `TASK_FILE_NAME`? | Main setup requirement |
|---|---:|---|
| `products` | Yes | Validator + SKU/slug/listing indexes |
| `categories` | Yes | Validator + parent/path/slug indexes |
| `brands` | Yes | Validator + slug/status/text indexes |
| `inventory_snapshots` | Yes | Validator + product/seller/SKU/time indexes |
| `price_books` | Yes | Validator + seller/status/effective-window indexes |
| `inventory_reservations` | No, later task | Created by `0006_inventory_reservations.up.js` |
| `product_event_outbox` | No, later task | Created by `0007_product_event_outbox.up.js` |

### Migration file for `TASK_FILE_NAME`

```text
backend/services/product-service/migrations/mongo/0003_create_product_collections.up.js
```

This file creates or updates:

- `products`
- `categories`
- `brands`
- `inventory_snapshots`
- `price_books`
- JSON schema validators
- required indexes for catalog queries

Rollback file:

```text
backend/services/product-service/migrations/mongo/0003_create_product_collections.down.js
```

Warning:

The migration script uses `db.getSiblingDB("product_db")`. Agar `PRODUCT_MONGO_DATABASE` ko change karte ho, then migration target bhi review karna padega. Otherwise app ek DB read karega aur migration dusre DB me run ho sakti hai.

### Run only the `TASK_FILE_NAME` migration

Use this only after MongoDB is already running.

```bash
cd backend/services/product-service
mongosh "mongodb://localhost:27017" migrations/mongo/0003_create_product_collections.up.js
```

If MongoDB is running inside the Docker container from previous docs:

```bash
cd backend/services/product-service
docker exec -i ecommerce-product-mongo mongosh < migrations/mongo/0003_create_product_collections.up.js
```

For full current repository verification, run all migrations in order as explained in the previous dependency file:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
10. Migration Setup
```

### Verify `TASK_FILE_NAME` collections

```bash
mongosh "mongodb://localhost:27017/product_db" --eval "db.getCollectionNames()"
```

Expected after only the `TASK_FILE_NAME` migration:

```text
products
categories
brands
inventory_snapshots
price_books
```

Check important indexes:

```bash
mongosh "mongodb://localhost:27017/product_db" --eval "db.products.getIndexes().map(i => i.name)"
mongosh "mongodb://localhost:27017/product_db" --eval "db.categories.getIndexes().map(i => i.name)"
mongosh "mongodb://localhost:27017/product_db" --eval "db.price_books.getIndexes().map(i => i.name)"
```

Important note:

If you run later migrations too, you may see additional collections and additional indexes. That is expected for the current repository, but those are outside `TASK_FILE_NAME` scope.

## 6. Redis / Queue / External Services

No Redis, Kafka, RabbitMQ, S3, SMTP, or third-party integration is newly required by `TASK_FILE_NAME`.

### MongoDB

MongoDB is mandatory for real `TASK_FILE_NAME` DB verification.

| Item | Detail |
|---|---|
| What it is | Document database |
| Why used here | Product catalog has flexible nested data like variants, images, category attributes, and price rules |
| Default port | `27017` |
| Credentials location | `backend/services/product-service/.env` through `PRODUCT_MONGO_URI` |
| Full setup | Reused from previous dependency docs |

### RabbitMQ

RabbitMQ is not introduced by `TASK_FILE_NAME`.

It appears in current repository code because later product event/outbox work exists. Use it only if you are testing full current service event behavior:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
7. External Services Analysis -> RabbitMQ
```

### Kafka / Redis / Typesense

| Service | `TASK_FILE_NAME` status | Beginner guidance |
|---|---|---|
| Kafka | Not required | Do not set `PRODUCT_EVENT_BROKER=kafka` for `TASK_FILE_NAME` |
| Redis | Not used | No Redis container needed |
| Typesense | Not this service runtime dependency | Search Service handles search indexing later |

## 7. Environment Variables

Complete `.env` documentation is already available here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
6. Environment Variables
```

### Task-specific variables

No new variable is introduced only by `TASK_FILE_NAME`.

For `TASK_FILE_NAME` MongoDB verification, make sure these existing variables match the migration target:

```env
PRODUCT_MONGO_URI=mongodb://localhost:27017
PRODUCT_MONGO_DATABASE=product_db
PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=false
```

| Variable | Required? | Purpose | Security note |
|---|---:|---|---|
| `PRODUCT_MONGO_URI` | Yes for DB runtime | MongoDB connection string | Secret if username/password is included |
| `PRODUCT_MONGO_DATABASE` | Yes for DB runtime | DB name used by app code | Not secret |
| `PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS` | Optional | Allows app-level collection creation if manager is wired | Prefer `false` in production |

Where `.env` should live:

```text
backend/services/product-service/.env
```

Beginner warning:

Current config reads process environment variables with `os.LookupEnv`. Sirf `.env` file bana dene se values load nahi hoti. Previous dependency doc ke `source .env` steps follow karo.

### Auto-create behavior

`PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=true` is not the normal beginner path.

It works only when:

- app wiring provides `CollectionSchemaManager`
- MongoDB connection is available
- current service startup path calls `app.New(...)`

Current repository has no server entrypoint, so manual migration execution is clearer. Also, current Go collection definitions include later collections too, not just the five `TASK_FILE_NAME` collections.

## 8. Docker Setup

No Dockerfile, docker-compose service, volume, network, or container is newly introduced by `TASK_FILE_NAME`.

Use previous Docker guidance:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
5. Database Analysis: MongoDB -> H. Docker setup
5. Database Analysis: MongoDB -> I. Docker Compose example
9. Docker and DevOps Setup
```

Task-specific Docker impact:

| Docker item | New for `TASK_FILE_NAME`? | Note |
|---|---:|---|
| MongoDB container | No | Reuse existing local MongoDB setup |
| Service app container | No | No Dockerfile exists yet |
| Compose network | No | No compose file exists yet |
| MongoDB volume | No | Reuse `product_mongo_data` from previous docs |
| RabbitMQ container | No | Not needed for `TASK_FILE_NAME` collection verification |

## 9. Local Development Setup

This onboarding flow avoids duplicate setup and focuses only on `TASK_FILE_NAME` additions.

### Step 1: Read previous dependency documentation

Read these first:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
```

Most important reused sections:

- Go dependency setup
- MongoDB install/Docker setup
- `.env` loading
- common Docker/MongoDB troubleshooting
- full migration flow

### Step 2: Go to project directory

```bash
cd backend/services/product-service
```

### Step 3: Install only new dependencies if any

No new dependency is added by `TASK_FILE_NAME`.

If dependencies are not already downloaded, use the previous dependency file:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
4. Language-Specific Dependency System: Go
```

### Step 4: Setup only new databases/services if any

No new database type is added. MongoDB setup is reused.

### Step 5: Add only new or changed environment variables

No new variables are needed. Confirm the MongoDB values:

```env
PRODUCT_MONGO_URI=mongodb://localhost:27017
PRODUCT_MONGO_DATABASE=product_db
PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=false
```

### Step 6: Run `TASK_FILE_NAME` migration

```bash
mongosh "mongodb://localhost:27017" migrations/mongo/0003_create_product_collections.up.js
```

### Step 7: Start backend service

Current state:

There is no runnable server entrypoint yet.

This is expected to fail:

```bash
go run .
```

Use these checks instead:

```bash
go test ./...
go build ./...
```

### Step 8: Verify functionality related to `TASK_FILE_NAME`

For `TASK_FILE_NAME`, verification means MongoDB collection setup is correct:

```bash
mongosh "mongodb://localhost:27017/product_db" --eval "db.getCollectionNames()"
mongosh "mongodb://localhost:27017/product_db" --eval "db.products.getIndexes().map(i => i.name)"
```

Expected core collections:

- `products`
- `categories`
- `brands`
- `inventory_snapshots`
- `price_books`

## 10. Running the Project

Full run instructions are reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
11. Complete Project Run Instructions
```

Task-specific run reality:

| Action | Command/status |
|---|---|
| Start MongoDB | Reuse previous doc |
| Load `.env` | Reuse previous doc |
| Run `TASK_FILE_NAME` migration | `mongosh "mongodb://localhost:27017" migrations/mongo/0003_create_product_collections.up.js` |
| Run all current migrations | Reuse previous doc, section `10. Migration Setup` |
| Run tests | `go test ./...` |
| Build packages | `go build ./...` |
| Start API server | Not available yet |
| Verify HTTP/gRPC API | Not available yet |

## 11. Common Errors & Fixes

Generic setup errors are already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
12. Common Errors and Fixes
```

Task-specific confusion points:

| Error / confusion | Cause | Fix | Prevention |
|---|---|---|---|
| Only default `test` DB changes, `product_db` empty | Migration not run from the file that calls `db.getSiblingDB("product_db")`, or wrong script was used | Run `migrations/mongo/0003_create_product_collections.up.js` exactly | Use the documented command from service folder |
| Expected 5 collections but see 7 | Later migrations or auto-create definitions also created `inventory_reservations` and `product_event_outbox` | This is okay for full current repo; `TASK_FILE_NAME` core is still the original 5 | Know whether you are verifying `TASK_FILE_NAME` only or full current service |
| Validator rejects sample product insert | Sample document is missing required fields like `_id`, `seller_id`, `title`, `category_id`, `status`, `variants`, timestamps | Add required fields or insert through service code later | Copy document shape from `TASK_FILE_NAME` |
| Duplicate key error for SKU or slug | `TASK_FILE_NAME` creates unique indexes for `variants.sku` and product/category/brand slugs | Use unique sample values or clean test data | Generate unique fixtures for every insert |
| `PRODUCT_MONGO_DATABASE` changed but migration still uses `product_db` | Migration script has fixed `db.getSiblingDB("product_db")` | Keep local DB name as `product_db` or update migration strategy carefully | Do not rename DB casually |
| `go run .` fails | Current module has no `main` package | Use `go test ./...` and `go build ./...` | Add server entrypoint in future service work |
| Category path shape seems different after later migrations | Later read migration changes product `category_path` validator to string IDs for read queries | Run migrations in order and follow current code when testing full repo | Treat `TASK_FILE_NAME` as design history, not the only current schema source |
| `PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=true` fails | App dependencies do not include `CollectionSchemaManager` or no entrypoint wires it | Keep it `false` and run migrations manually | Use reviewed migrations for local/prod setup |

## 12. Security & Best Practices

General security guidance is already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
13. Security and Configuration Audit
14. Best Practices
```

Task-specific best practices:

- Keep `product_db` owned by this service only. Other services should not directly read/write these collections.
- Use MongoDB users with least privilege in staging/production.
- Do not expose MongoDB port `27017` publicly.
- Keep `PRODUCT_MONGO_URI` in `.env` or secret manager, not in Go code or committed markdown with real passwords.
- Use migrations for schema/index changes. Auto-create is useful for local experiments but risky for production discipline.
- Keep SKU and slug uniqueness rules clear for seed/test data.
- Validate dynamic attributes in service logic, not only MongoDB validators.
- Do not store image binaries in MongoDB. Store URLs/metadata only.
- Keep money as integer minor units, for example paise/cents, not floating point.
- Review indexes before production because product listing queries can become slow without the right compound indexes.

## 13. Missing or Misconfigured Things

| Area | Current observation | Impact | Suggested fix |
|---|---|---|---|
| Server entrypoint | No `cmd/server/main.go` or equivalent exists | No API process can be started yet | Add service entrypoint later with config load, Mongo connection, handlers, health checks |
| Migration runner | Migrations are plain `mongosh` scripts | Beginners must run scripts manually and in order | Add a small migration runner or documented make target later |
| DB name in migrations | `0003` uses `product_db` directly | Env DB mismatch can confuse setup | Keep `PRODUCT_MONGO_DATABASE=product_db` locally or parameterize migration process |
| Dockerfile | Not present | App container cannot be built | Add Dockerfile after server entrypoint exists |
| docker-compose | Not present | Local setup is manual | Add local compose for MongoDB and optional RabbitMQ |
| Seed data | No beginner seed script for categories/brands | Hard to test valid products manually | Add seed script for sample categories, brands, and one valid product |
| Current schema drift from `TASK_FILE_NAME` design | Later migrations add/change fields/indexes | A reader may compare docs and current DB and see extra objects | Document migration order and current schema source |
| Health checks | No app health/readiness endpoint yet | Runtime dependency status is hard to verify | Add `/healthz` or gRPC health after server is implemented |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Project Tech Stack Analysis` | Same Go/MongoDB/RabbitMQ technology explanation already exists |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Language-Specific Dependency System: Go` | Same Go module setup applies |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Analysis: MongoDB` | MongoDB install, Docker, connection string, and verification are already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Environment Variables` | Complete `.env` behavior and variable table already exists |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. External Services Analysis` | RabbitMQ/Kafka/Redis status is already explained |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Ports and Networking` | Same MongoDB/RabbitMQ port guidance applies |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Docker and DevOps Setup` | Same manual Docker setup and missing Dockerfile/compose notes apply |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `10. Migration Setup` | Full current migration order is already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `11. Complete Project Run Instructions` | Same onboarding flow applies except this file narrows `TASK_FILE_NAME` verification |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `12. Common Errors and Fixes` | Generic Go/MongoDB/Docker/env issues already exist |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | MongoDB choice and `PRODUCT_MONGO_*` naming notes are already documented |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `11. Common Errors & Fixes` | DB-name and env-name confusion points are already explained |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `13. Missing or Misconfigured Things` | Server entrypoint, Dockerfile, compose, and health-check gaps are unchanged |

## 15. Final Checklist

- [ ] Previous dependency documentation checked
- [ ] `INPUT_FILE_PATH` reviewed
- [ ] No original task file modified
- [ ] No duplicate generic Go setup added
- [ ] No duplicate MongoDB installation guide added
- [ ] No duplicate Docker/RabbitMQ setup added
- [ ] `TASK_FILE_NAME` core collections identified
- [ ] `TASK_FILE_NAME` migration file identified
- [ ] MongoDB is running locally if DB verification is needed
- [ ] `PRODUCT_MONGO_URI` points to the correct MongoDB instance
- [ ] `PRODUCT_MONGO_DATABASE=product_db` confirmed
- [ ] `PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=false` kept for normal local setup
- [ ] `0003_create_product_collections.up.js` run for `TASK_FILE_NAME` verification
- [ ] Core collections verified in `product_db`
- [ ] Important indexes verified
- [ ] `go test ./...` run if current code verification is needed
- [ ] `go build ./...` run if compile verification is needed
- [ ] No real secrets committed
- [ ] Missing entrypoint/Docker/health-check gaps understood
- [ ] New dependency file saved at `OUTPUT_FILE_PATH`
