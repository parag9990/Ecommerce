# Project Dependency & Setup Guide

## Variable Values Used

```text
SERVICE_NAME=Product Service
TASK_FILE_NAME=task5.md
INPUT_FILE_PATH=TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME=task5_Dependency.md
OUTPUT_FILE_PATH=TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

This file is generated for `INPUT_FILE_PATH` and saved as `OUTPUT_FILE_PATH`.

Beginner note:

- `TASK_FILE_NAME` covers product read APIs: public product list/detail, category browse, seller catalog list/detail, and internal batch product reads.
- Shared Go, MongoDB, Docker, RabbitMQ, common `.env`, ports, and generic troubleshooting are already documented in previous dependency files.
- This guide explains only the setup impact that is new or important for `TASK_FILE_NAME`: read service wiring, read pagination env values, read index migration `0005`, and read API verification.
- The original `INPUT_FILE_PATH` file is not modified.

## 1. Project Overview

`TASK_FILE_NAME` ka goal read side ko ready karna hai.

| Read flow | Current implementation area | Setup impact |
|---|---|---|
| Public product list | `internal/transport/productread`, `internal/usecase/product_read_service.go` | Needs MongoDB collections + read indexes for fast query |
| Public product detail | `ProductReadService.GetPublicProduct` | Only `published` products visible |
| Category browse | `ProductReadService.ListCategories` | Needs `categories` collection and active/sort indexes |
| Seller catalog list/detail | `ProductReadService.ListSellerProducts`, `GetSellerProduct` | Needs authenticated actor context in caller/gateway |
| Internal batch product read | `BatchGetProducts` | Controlled by max batch-size config |

Simple Hinglish:

Read APIs ka matlab data fetch karna. Public buyer ko sirf live `published` products dikhne chahiye. Seller dashboard ko apne `draft`, `submitted`, `published`, `unpublished`, and `rejected` products dikhne chahiye. Isliye setup ka main focus MongoDB indexes, pagination limits, and safe visibility rules par hai.

### Current implementation reality

| Area | Status | Beginner meaning |
|---|---|---|
| Go read usecase | Present | Read visibility and pagination logic implemented hai |
| Product read handler | Present | In-process handler exists, but no HTTP/gRPC server entrypoint yet |
| MongoDB read repository | Present | `MongoProductRepository` implements read filters and projections |
| Migration `0005` | Present | Read/listing indexes add karta hai |
| `.env.example` read vars | Present | `PRODUCT_READ_*` values available hain |
| API Gateway route binding | Not present in this service module | Routes documented hain, runnable gateway wiring verify separately |
| Product service Dockerfile | Not present | App container build abhi possible nahi |
| docker-compose | Not present for this service | MongoDB/RabbitMQ local setup previous docs se reuse karo |

Important:

Do not assume `go run .` starts these APIs. Current module still has no `cmd/server/main.go`. For now, verify with `go test`, `go build`, MongoDB migrations, and repository/usecase tests.

## 2. Tech Stack

Full explanation for common technologies is already available here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
3. Project Tech Stack Analysis
4. Language-Specific Dependency System: Go
5. Database Analysis: MongoDB
7. External Services Analysis
```

### Task-specific tech stack summary

| Technology / component | Required? | Why used in `TASK_FILE_NAME` | New explanation needed? |
|---|---:|---|---|
| Go | Yes | Read usecase, DTO, handler, repository code Go me hai | No, reused |
| Go modules | Yes | Existing `go.mod` / `go.sum` manage dependencies | No, reused |
| MongoDB | Yes for real DB-backed verification | Products/categories read source of truth | Setup reused, read indexes explained below |
| MongoDB Go Driver v2 | Yes | Read repository uses official MongoDB driver | No, reused |
| `mongosh` | Yes for manual migration verification | `0005_product_read_indexes.up.js` run/check karne ke liye | No install duplication |
| Docker | Optional | Local MongoDB/RabbitMQ run karne ke liye | Reused |
| `log/slog` | Yes | Read usecase structured debug logs likhta hai | Built-in Go package |
| RabbitMQ | No new requirement | Read APIs do not publish events | Reused only if full app/events enabled |
| Redis | No | Current read APIs Redis cache use nahi karte | No setup |
| Kafka | No | Config allows broker choice for events, but read APIs do not need Kafka | No setup |
| Typesense/Search | No direct dependency | Search Service owns full-text search boundary | Do not install for `TASK_FILE_NAME` |

Hinglish:

MongoDB yahan source of truth hai. Read APIs direct MongoDB se indexed query chalati hain. Redis cache ya search engine abhi is task ke local setup me required nahi hai.

## 3. Required Software

No brand-new software is introduced by `TASK_FILE_NAME`.

Follow previous setup first:

```text
This setup is already explained in:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
4. Language-Specific Dependency System: Go
5. Database Analysis: MongoDB
6. Environment Variables
8. Ports and Networking
9. Docker and DevOps Setup
10. Migration Setup
```

For task-specific verification, you need:

| Software | Purpose | Status |
|---|---|---|
| Git | Clone repository | Reused |
| Go `1.26.3` or compatible configured toolchain | Build/test read code | Reused |
| MongoDB | Store `product_db.products` and `product_db.categories` | Reused, mandatory for DB-backed checks |
| `mongosh` | Run/check read index migration | Reused |
| Docker | Optional local MongoDB/RabbitMQ containers | Reused |

## 4. Dependency Management

Go dependency setup is already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
4. Language-Specific Dependency System: Go
```

### What changed in `TASK_FILE_NAME`?

No new Go module dependency is required only for read APIs.

Current service module remains:

```text
backend/services/product-service/go.mod
backend/services/product-service/go.sum
backend/go.work
```

Current direct dependencies still include:

| Dependency | Version | Used for |
|---|---:|---|
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | MongoDB collection reads, filters, projections, indexes verification |
| `github.com/rabbitmq/amqp091-go` | `v1.10.0` | Product event flow from other tasks, not read-specific |

Do not run `go get` for `TASK_FILE_NAME` unless you intentionally add a new package.

Task-specific verification commands:

```bash
cd backend/services/product-service
go test ./internal/usecase -run 'ListPublicProducts|GetPublicProduct|BatchGetProducts|ListCategories|ListSellerProducts'
go test ./internal/transport/productread ./internal/transport/dto
go test ./...
go build ./...
```

Note:

If any dependency download or Go version issue occurs, use the fixes from `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, section `4. Language-Specific Dependency System: Go`.

## 5. Database Setup

MongoDB install, Docker setup, connection strings, and generic DB troubleshooting are already explained here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
5. Database Analysis: MongoDB
```

### Database used

```text
product_db
```

### Collections used by `TASK_FILE_NAME`

| Collection | Required? | Why |
|---|---:|---|
| `products` | Yes | Public product list/detail, seller catalog, batch reads |
| `categories` | Yes | Public category browse and category filter support |
| `brands` | No direct read API in this task | Existing catalog collection, not newly configured here |
| `inventory_snapshots` | No | Later inventory/reporting scope |
| `price_books` | No | Pricing rules are existing catalog scope |
| `product_event_outbox` | No | Event publishing, not read API setup |

### Task-specific migration

Read API performance depends on this migration:

```text
backend/services/product-service/migrations/mongo/0005_product_read_indexes.up.js
```

It expects base collections to already exist. At minimum, run `0003` before `0005`. For the full current read/seller lifecycle, run migrations in order:

```text
0003_create_product_collections.up.js
0004_seller_product_workflow.up.js
0005_product_read_indexes.up.js
```

Full migration commands are already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
10. Migration Setup
```

### What `0005` adds

| Index / schema change | Why it matters |
|---|---|
| `idx_products_status_published` | Public newest listing by `status + published_at` fast hoti hai |
| `idx_products_category_status_published` | Category page listing fast hoti hai |
| `idx_products_category_path_ids_status_updated` | Ancestor category browsing works with string-array `category_path` |
| `idx_products_status_price_amount` | `price_asc` / `price_desc` sorting ke liye support |
| `idx_products_status_rating` | `rating_desc` sorting ke liye support |
| `idx_categories_active_parent_sort_name` | Active category browse and child category sorting fast hoti hai |
| `category_path` validator changed to array of strings | Current Go repository writes `CategoryPath []string` |

### Verify read indexes

Use this only after MongoDB is running and migrations are applied.

```bash
mongosh "mongodb://localhost:27017/product_db" --eval 'db.products.getIndexes().map(i => i.name)'
mongosh "mongodb://localhost:27017/product_db" --eval 'db.categories.getIndexes().map(i => i.name)'
```

Expected task-specific index names:

```text
idx_products_status_published
idx_products_category_status_published
idx_products_category_path_ids_status_updated
idx_products_status_price_amount
idx_products_status_rating
idx_categories_active_parent_sort_name
```

### Schema audit note

`0003_create_product_collections.up.js` originally defines `products.category_path` as an array of objects. Current Go repository writes `category_path` as an array of strings, and `0005_product_read_indexes.up.js` updates the validator to match that string-array shape.

Beginner meaning:

If old local data was inserted with object-shaped `category_path`, read code can still decode both shapes. But for new local setup, run migrations in order so MongoDB validator and Go write shape match.

## 6. Redis / Queue / External Services

No Redis, Kafka, RabbitMQ, S3, SMTP, Stripe, Twilio, Firebase, Elasticsearch, Typesense, or MinIO setup is newly required for `TASK_FILE_NAME`.

### MongoDB

MongoDB is mandatory for DB-backed read API verification.

| Item | Detail |
|---|---|
| What it is | Document database |
| Why used here | Products/categories flexible catalog data store karne ke liye |
| Default port | `27017` |
| Setup status | Reused from previous dependency docs |

### RabbitMQ

RabbitMQ is not introduced by read APIs.

```text
This setup is already explained in:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
7. External Services Analysis -> RabbitMQ
```

Task-specific recommendation:

- For read-only local tests, RabbitMQ can stay disabled or not started if your app wiring does not start event workers.
- If you run full current app wiring with events enabled, follow previous RabbitMQ setup and `RABBITMQ_URL` rules.

### Search / Typesense

Do not install Typesense for `TASK_FILE_NAME`.

Hinglish:

Product list/detail yahan MongoDB read hai. Full-text search, typo tolerance, facets, and ranking Search Service ka kaam hai. Agar frontend `search?q=shoes` type feature chahiye, woh separate Search Service setup hoga.

## 7. Environment Variables

Complete `.env` behavior and full variable table already exist here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
6. Environment Variables
```

### Task-specific read variables

These variables already exist in `backend/services/product-service/.env.example`, but they matter directly for `TASK_FILE_NAME`.

```env
PRODUCT_READ_DEFAULT_PAGE_SIZE=20
PRODUCT_READ_MAX_PAGE_SIZE=100
PRODUCT_READ_MAX_BATCH_SIZE=100
```

| Variable | Required? | Example | Purpose | Security note |
|---|---:|---|---|---|
| `PRODUCT_READ_DEFAULT_PAGE_SIZE` | Optional | `20` | Client page_size missing/invalid ho to default list size | Not secret |
| `PRODUCT_READ_MAX_PAGE_SIZE` | Optional | `100` | Large list requests ko clamp karta hai | Not secret |
| `PRODUCT_READ_MAX_BATCH_SIZE` | Optional | `100` | Internal batch product read limit | Not secret |

How project loads them:

| Code file | Behavior |
|---|---|
| `internal/config/config.go` | Env vars read into `cfg.Read` |
| `internal/app/app.go` | `cfg.Read` passed into `usecase.NewProductReadService` |
| `internal/usecase/product_read_service.go` | Pagination and batch-size validation applied |

Validation rules:

| Rule | Error if broken |
|---|---|
| `PRODUCT_READ_DEFAULT_PAGE_SIZE > 0` | Config validation fails |
| `PRODUCT_READ_MAX_PAGE_SIZE > 0` | Config validation fails |
| `PRODUCT_READ_DEFAULT_PAGE_SIZE <= PRODUCT_READ_MAX_PAGE_SIZE` | Config validation fails |
| `PRODUCT_READ_MAX_BATCH_SIZE > 0` | Config validation fails |

### Where to create `.env`

Use the same location explained in previous docs:

```text
backend/services/product-service/.env
```

Beginner tip:

Copy `backend/services/product-service/.env.example` to `.env`, then only change values that are needed for your local machine. Do not commit real credentials.

## 8. Docker Setup

Docker setup is reused.

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
| MongoDB container | No | Reuse previous MongoDB container |
| RabbitMQ container | No | Not required for read-only checks |
| Product service app container | No | Dockerfile still missing |
| New volume | No | Reuse MongoDB volume from previous docs |
| New network | No | No task-specific Docker network added |
| New exposed port | No | No service listener exists yet |

Warning:

If you use Docker MongoDB, run `0005` against the same MongoDB container/database that the service reads from. Agar migration ek DB me run hoti hai aur app dusre DB se connect karta hai, indexes missing lag sakte hain.

## 9. Local Development Setup

### Step 1: Read previous dependency docs first

Read these in order:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
```

### Step 2: Clone and enter repository

```bash
git clone <repository-url>
cd <repository-folder>
```

Then go to the service module:

```bash
cd backend/services/product-service
```

### Step 3: Install dependencies

No new dependencies for `TASK_FILE_NAME`.

Use previous Go module commands only if dependencies are missing:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
4. Language-Specific Dependency System: Go
```

### Step 4: Setup database/services

Use previous MongoDB setup, then run migrations through `0005`.

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
5. Database Analysis: MongoDB
10. Migration Setup
```

### Step 5: Add read env values

Make sure these values exist in local `.env` if you override defaults:

```env
PRODUCT_READ_DEFAULT_PAGE_SIZE=20
PRODUCT_READ_MAX_PAGE_SIZE=100
PRODUCT_READ_MAX_BATCH_SIZE=100
```

### Step 6: Run task-specific tests

```bash
cd backend/services/product-service
go test ./internal/usecase -run 'ListPublicProducts|GetPublicProduct|BatchGetProducts|ListCategories|ListSellerProducts'
go test ./...
```

## 10. Running the Project

Current runtime limitation:

```text
No `SERVICE_NAME` server entrypoint exists yet.
```

So this command is not expected to start read APIs:

```bash
cd backend/services/product-service
go run .
```

Use these checks instead:

| Check | Command / action |
|---|---|
| Compile packages | `go build ./...` |
| Run read tests | `go test ./internal/usecase -run 'ListPublicProducts|GetPublicProduct|BatchGetProducts|ListCategories|ListSellerProducts'` |
| Run all service tests | `go test ./...` |
| Verify MongoDB is running | Use previous MongoDB ping commands |
| Verify read indexes | Use `db.products.getIndexes()` and `db.categories.getIndexes()` |

Future expected run style after an entrypoint is added:

```bash
cd backend/services/product-service
go run ./cmd/server
```

### Task-specific API functionality to verify

When a real HTTP/gRPC server is wired later, verify:

| Route / method | Expected setup result |
|---|---|
| `GET /api/v1/products` | Returns only `published` products |
| `GET /api/v1/products/{product_id}` | Draft/unpublished product returns not found |
| `GET /api/v1/categories` | Returns active categories sorted by `sort_order`, `name` |
| `GET /api/v1/seller/products` | Requires seller actor and returns only own products |
| `BatchGetProducts` | Deduplicates IDs, enforces max batch size, returns published products in request order |

## 11. Common Errors & Fixes

Generic Go, Docker, MongoDB, RabbitMQ, and `.env` errors are already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
12. Common Errors and Fixes
```

Task-specific errors:

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `products collection must exist before applying product read indexes migration` | `0005` ran before `0003` | Run migrations in order: `0003`, `0004`, `0005` | Keep migration checklist in setup docs |
| `categories collection must exist before applying product read indexes migration` | Base collection migration missing | Run `0003_create_product_collections.up.js` first | Verify collections before read index migration |
| `PRODUCT_READ_DEFAULT_PAGE_SIZE cannot exceed PRODUCT_READ_MAX_PAGE_SIZE` | Bad `.env` values | Set default lower than or equal to max | Keep local values close to `.env.example` |
| `unsupported product sort` | Client sent unknown `sort` | Use `newest`, `updated_at_desc`, `price_asc`, `price_desc`, or `rating_desc` | Validate query params in gateway/client |
| `too many product ids requested` | Batch request exceeded `PRODUCT_READ_MAX_BATCH_SIZE` | Split batch request or increase max carefully | Keep batch size bounded |
| Seller list returns permission error | Actor has no `seller_id` or role/permission | Pass authenticated seller actor context | Do not trust seller_id from query/body |
| Public API does not show draft product | Expected behavior | Use seller catalog API for own drafts | Public reads must force `published` |
| Price/rating sort seems slow | `0005` not applied or index not used | Verify read indexes exist | Run explain plans before exposing high-traffic sort |

## 12. Security & Best Practices

### Task-specific security rules

| Rule | Why |
|---|---|
| Public list/detail must force `published` status | Draft/unpublished data leakage avoid hota hai |
| Public product missing/unpublished should return not found | Product existence leak kam hota hai |
| Seller catalog must use auth actor `seller_id` | Cross-seller data exposure avoid hota hai |
| Do not trust request body/query `seller_id` for seller routes | User easily query param tamper kar sakta hai |
| Keep page and batch limits small | DB overload and accidental huge responses avoid hote hain |
| Use projections for list/detail reads | Internal fields and heavy data avoid hote hain |
| Do not expose MongoDB directly to other services | `SERVICE_NAME` owns catalog DB boundary |

### Best practices specific to `TASK_FILE_NAME`

- Keep `PRODUCT_READ_MAX_PAGE_SIZE` conservative. Beginner-friendly default `100` is okay for local, but production traffic ke hisaab se tune karo.
- Keep `PRODUCT_READ_MAX_BATCH_SIZE` bounded because Cart/Order/Search can accidentally send large product ID lists.
- Run `0005` before load testing product list/category pages.
- Add query explain checks for high-traffic category pages.
- Add gateway-level rate limits for public product list/detail routes.
- Log `viewer_type`, `seller_id`, `category_id`, `page_size`, `result_count`, and `sort`, but do not log customer tokens.
- Cache later only after correctness is verified. Redis is not required for this task.

## 13. Missing or Misconfigured Things

| Item | Current status | Setup risk | Suggested fix |
|---|---|---|---|
| `SERVICE_NAME` server entrypoint | Missing | Beginner cannot start actual API server | Add `cmd/server/main.go` when transport is ready |
| gRPC/HTTP route binding | Not present in this module | Handlers exist but are not exposed over network | Wire gateway/proto/server layer |
| Dockerfile | Missing | Cannot build app container | Add Dockerfile after entrypoint exists |
| docker-compose | Missing | Local infra setup manual | Add local compose for MongoDB and optional RabbitMQ |
| Health checks | Missing for app runtime | No standard readiness check | Add MongoDB ping and optional broker check |
| Read index migration target hardcoded to `product_db` | Present in migration files | If `PRODUCT_MONGO_DATABASE` changes, migration and app DB may differ | Keep default DB or update migrations/tooling together |
| `category_path` shape changed across migrations | `0003` object array, `0005` string array | Old local data/index expectations may confuse developers | Run migrations in order and keep Go write shape aligned |
| Search boundary | Full-text search not in this task | Developers may install Typesense unnecessarily | Keep search setup in Search Service docs |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Project Tech Stack Analysis` | Same Go, MongoDB, Docker, RabbitMQ technology explanation already exists |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Language-Specific Dependency System: Go` | Same Go module commands and troubleshooting apply |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Analysis: MongoDB` | MongoDB install, Docker run, connection string, and verification are already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Environment Variables` | Full `.env` table already includes `PRODUCT_READ_*` variables |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. External Services Analysis` | RabbitMQ/Kafka/Redis status already explained |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Ports and Networking` | Same MongoDB/RabbitMQ port guidance applies |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Docker and DevOps Setup` | Dockerfile/compose limitations and local Docker commands are reused |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `10. Migration Setup` | Full migration command flow, including `0005`, already exists |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `12. Common Errors and Fixes` | Generic setup errors are reused |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | MongoDB as primary catalog DB decision is reused |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `5. Database Setup` | Base `products` and `categories` collections are prerequisites for read APIs |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `5. Database Setup` | Seller lifecycle statuses and `0004` migration are prerequisites for seller catalog reads |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `7. Environment Variables` | CMS/event local-mode caveats are reused when full app wiring is tested |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before using this file
- [ ] No duplicate Go/MongoDB/Docker setup copied unnecessarily
- [ ] `backend/services/product-service/.env` created from `.env.example` if local config is needed
- [ ] `PRODUCT_READ_DEFAULT_PAGE_SIZE` reviewed
- [ ] `PRODUCT_READ_MAX_PAGE_SIZE` reviewed
- [ ] `PRODUCT_READ_MAX_BATCH_SIZE` reviewed
- [ ] MongoDB running locally or in Docker
- [ ] Base collections created with `0003_create_product_collections.up.js`
- [ ] Seller workflow migration `0004_seller_product_workflow.up.js` applied if seller catalog statuses are tested
- [ ] Read indexes migration `0005_product_read_indexes.up.js` applied
- [ ] Read indexes verified in `products` and `categories`
- [ ] `go test ./internal/usecase -run 'ListPublicProducts|GetPublicProduct|BatchGetProducts|ListCategories|ListSellerProducts'` passed
- [ ] `go test ./...` passed
- [ ] Public reads verified to show only `published` products
- [ ] Seller reads verified to use authenticated seller context
- [ ] Batch reads verified to enforce max batch size
- [ ] No Redis/Typesense/Kafka setup added for this task
- [ ] Missing server entrypoint and Dockerfile limitations understood
- [ ] Logs checked during read tests/debugging
