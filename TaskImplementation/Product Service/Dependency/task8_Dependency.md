# Project Dependency & Setup Guide

## Variable Values Used

```text
SERVICE_NAME=Product Service
TASK_FILE_NAME=task8.md
INPUT_FILE_PATH=TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME=task8_Dependency.md
OUTPUT_FILE_PATH=TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

Important:

- `INPUT_FILE_PATH` was analyzed only for dependency, setup, environment, database, external service, and DevOps requirements.
- The original `TASK_FILE_NAME` was not modified.
- Shared Go, MongoDB, Docker, RabbitMQ, `.env`, ports, migrations, and generic troubleshooting are already documented in earlier dependency files.
- This file explains only the current task-specific media metadata setup and references previous files wherever setup is already covered.

## 1. Project Overview

`TASK_FILE_NAME` covers product media metadata.

Simple Hinglish:

Is task ka goal image file upload karna nahi hai. Iska goal product document ke andar image metadata maintain karna hai: CDN/public URL, alt text, display position, primary image flag, variant link, dimensions, and status. Actual image binary kisi CDN/object storage layer me rahegi. This service stores and validates the metadata that frontend gallery and search thumbnail flows need.

### Current implementation reality

| Area | Current status | Setup impact |
|---|---|---|
| Structured image domain model | Present in `internal/domain/image.go` | No new install needed |
| Product image validation | Present in `internal/domain/product.go` and `internal/domain/image.go` | Uses Go standard library `net/url` |
| Seller input DTO | Present in `internal/transport/dto/seller_product.go` | Supports legacy `images[]` URL strings plus structured `image_details[]` |
| Catalog/read DTOs | Present in `internal/transport/dto/catalog.go` and `internal/transport/dto/product_read.go` | Public/read responses can return structured image fields |
| MongoDB image schema | Present through migrations `0003` and `0004` | No separate `0008` migration exists |
| Media-specific env variables | No new env variable created only for `TASK_FILE_NAME` | Reuse existing image limit/publish variables |
| CDN/object storage credentials | Not present | Product metadata flow does not upload binaries |
| Dedicated media endpoint | Not present | Current seller create/update payload carries image metadata |
| Server entrypoint | Still not present | Use tests/build verification, not `go run .` |
| Dockerfile / compose | Still not present for this module | Reuse manual local infra setup from previous docs |

Task-specific reality:

Current code supports these image fields:

```text
image_id, url, alt_text, position, is_primary, variant_ids, width, height, status
```

Current code does not support these future fields from the task guide yet:

```text
object_key, content_type, size_bytes, deleted status, CDN allowlist env, upload bucket credentials
```

## 2. Tech Stack

Common technology explanation is already documented here:

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

| Technology / component | Required for `TASK_FILE_NAME`? | Why used | New setup? |
|---|---:|---|---:|
| Go | Yes | Image metadata domain, DTO mapping, validation, repository mapping, tests | No |
| Go modules | Yes | Existing `go.mod` manages dependencies | No |
| Go `net/url` | Yes | Validates image URL is absolute HTTP/HTTPS | No, standard library |
| MongoDB | Yes for DB-backed catalog runtime | Stores `products.images[]` embedded metadata | No new DB, reused |
| MongoDB Go Driver v2 | Yes | Reads/writes product documents with image metadata | No |
| RabbitMQ | Conditional | Product image changes can lead to `ProductUpdated` event via existing Task 7 flow | No new setup |
| CDN/Object Storage | Conceptual upstream dependency | Seller dashboard or another upload layer provides public image URL | No Product Service credentials now |
| Docker | Optional but recommended | Local MongoDB/RabbitMQ containers | No new container |
| `mongosh` | Required for manual migrations | Applies/verifies MongoDB schema files | No |
| Redis | No | Not used by current media metadata implementation | No |
| Kafka | No for beginner setup | Config accepts Kafka, but no default Kafka publisher is wired | No |
| Typesense/Search Service | Consumer-side only | Search thumbnail can be updated from product events | No local setup here |

Beginner note:

CDN ka matlab hai image ko fast public URL se serve karna. Is service ke liye CDN setup mandatory nahi hai jab tak aap real upload pipeline test nahi kar rahe. Local tests me `https://cdn.example.com/products/prod_123/main.jpg` jaisa sample URL enough hai.

## 3. Required Software

Do not reinstall shared tools. Follow:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
4. Language-Specific Dependency System: Go
5. Database Analysis: MongoDB
6. Environment Variables
7. External Services Analysis -> RabbitMQ
9. Docker and DevOps Setup
10. Migration Setup
```

For `TASK_FILE_NAME` verification, make sure these are available:

| Software | Purpose | Status |
|---|---|---|
| Go matching `go.mod` | Run validation/unit tests and build packages | Reused |
| MongoDB | Store product documents if repository/migration verification is needed | Reused |
| `mongosh` | Apply/inspect product collection validators | Reused |
| Docker | Optional local MongoDB/RabbitMQ setup | Reused |
| RabbitMQ | Only if testing event publishing after product image changes | Reused and conditional |

No separate media server, CDN emulator, MinIO, S3 CLI, image processor, or upload worker is required by the current implementation.

## 4. Dependency Management

Go module setup, `go mod download`, `go mod tidy`, `go test`, `go build`, and common Go module errors are already explained here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
4. Language-Specific Dependency System: Go
```

### What changed for `TASK_FILE_NAME`?

No new third-party package needs to be installed for current media metadata support.

Current direct dependencies remain:

| Dependency | Version | Why it matters here |
|---|---:|---|
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | Product repository maps `products.images[]` to/from MongoDB |
| `github.com/rabbitmq/amqp091-go` | `v1.10.0` | Existing event publisher can publish product update events |

Current image URL validation uses Go standard library:

```text
net/url
```

Task guide mentions `github.com/go-playground/validator/v10` as optional, but it is not present in current `go.mod`. Do not run `go get` for it unless you are actually implementing a new DTO validation layer.

Task-specific practical commands:

```bash
cd backend/services/product-service
go test ./internal/domain ./internal/transport/dto ./internal/mapper
go build ./...
```

Beginner note:

`go get` dependency add karta hai. Agar sirf existing media metadata code run/test karna hai, `go get` ki zarurat nahi hai. Agar local module cache empty hai, `go mod download` enough hai.

## 5. Database Setup

MongoDB installation, Docker setup, connection string format, credentials, and generic troubleshooting are reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
5. Database Analysis: MongoDB
10. Migration Setup
```

### Database used

| Item | Value |
|---|---|
| Database | `product_db` |
| Connection env | `PRODUCT_MONGO_URI` |
| Database env | `PRODUCT_MONGO_DATABASE` |
| Default MongoDB port | `27017` |

Warning:

Migration scripts use `db.getSiblingDB("product_db")`. Agar local `.env` me `PRODUCT_MONGO_DATABASE` change karte ho, app aur migration alag DB target kar sakte hain. Beginner setup ke liye `product_db` hi rakho.

### Collections used by `TASK_FILE_NAME`

| Collection | Required? | Status | Why |
|---|---:|---|---|
| `products` | Yes | Reused from `0003` and updated by `0004` | Stores embedded `images[]` metadata |
| `product_event_outbox` | Conditional | Reused from `0007` | Product media changes can be published as product update events |
| `categories` | Conditional | Reused | Product validation may use category schema |
| `inventory_reservations` | No for media metadata | Reused from inventory task | Not needed for image metadata verification |

### Task-specific migration status

There is no new migration file named for `TASK_FILE_NAME`.

Current image metadata schema comes from:

```text
backend/services/product-service/migrations/mongo/0003_create_product_collections.up.js
backend/services/product-service/migrations/mongo/0004_seller_product_workflow.up.js
```

What these migrations already support:

| Field | Source | Status |
|---|---|---|
| `url` | `0003` | Required |
| `position` | `0003` | Required |
| `is_primary` | `0003` | Required |
| `status` | `0003` | Required |
| `image_id` | `0004` | Required after workflow migration |
| `alt_text` | `0004` | Optional |
| `variant_ids` | `0004` | Optional |
| `width` | `0004` | Optional |
| `height` | `0004` | Optional |

Task guide future fields not in current migration:

| Future field | Current DB status | Setup implication |
|---|---|---|
| `object_key` | Not in validator | Needs future migration before storing |
| `content_type` | Not in validator | Needs future migration before storing |
| `size_bytes` | Not in validator | Needs future migration before storing |
| `deleted` status | Not in current enum | Needs domain + migration update |

### Run migrations for media metadata verification

Use previous docs for full MongoDB start/install steps. After MongoDB is running:

```bash
cd backend/services/product-service
mongosh "mongodb://localhost:27017" migrations/mongo/0003_create_product_collections.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0004_seller_product_workflow.up.js
```

For full current service verification, run migrations in order through `0007`:

```bash
mongosh "mongodb://localhost:27017" migrations/mongo/0005_product_read_indexes.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0006_inventory_reservations.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0007_product_event_outbox.up.js
```

### Verify media schema

```bash
mongosh "mongodb://localhost:27017/product_db" --eval 'db.getCollectionInfos({name:"products"})[0].options.validator'
```

Check that `products` validator contains:

```text
images.items.properties.image_id
images.items.properties.url
images.items.properties.alt_text
images.items.properties.variant_ids
images.items.properties.width
images.items.properties.height
```

### MongoDB transaction warning

If you test seller product write plus outbox event creation, MongoDB transactions may be used by the existing service flow. This is already explained here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md

Section:
8. Docker Setup -> MongoDB transaction warning
```

If you only run pure domain/DTO tests for image metadata, no MongoDB replica set is required.

## 6. Redis / Queue / External Services

### CDN / Object Storage

CDN/object storage is mentioned by `TASK_FILE_NAME`, but current service code does not provision or connect to it.

| Item | Current status |
|---|---|
| Image binary upload | Out of scope |
| S3/MinIO/GCS/R2 bucket | Not configured in this service |
| CDN distribution | Not configured in this service |
| Storage credentials env | Not present |
| Product Service responsibility | Store and validate public image metadata |

Hinglish explanation:

Seller dashboard pe image upload flow alag layer se ho sakta hai. Us layer se public URL milne ke baad this service ko sirf URL and metadata milta hai. Isliye yahan S3 access key, bucket name, ya CDN secret add karne ki zarurat nahi hai.

Recommended current local test URL:

```text
https://cdn.example.com/products/prod_123/main.jpg
```

Current code accepts absolute `http` or `https` URLs. Production best practice is HTTPS only, but current validation does not enforce HTTPS-only or CDN host allowlist yet.

### RabbitMQ

RabbitMQ setup is reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
7. External Services Analysis -> RabbitMQ
```

Task-specific behavior:

| Scenario | RabbitMQ required? | Why |
|---|---:|---|
| Run media domain validation tests | No | Pure Go tests |
| Run DTO mapping tests | No | Pure Go tests |
| Store product with images in MongoDB | No | MongoDB only |
| Verify `ProductUpdated` event publish after image update | Yes, if outbox worker and RabbitMQ publishing are enabled | Uses existing Task 7 event pipeline |

Event setup details are already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task7_Dependency.md

Sections:
6. Redis / Queue / External Services -> RabbitMQ
7. Environment Variables
10. Running the Project -> Inspect outbox state
```

### Search / Typesense

Search Service and Typesense are consumer-side dependencies, not runtime dependencies of this service for `TASK_FILE_NAME`.

Image metadata matters because the existing product search event mapper can choose a primary active image URL for search payloads. But this service does not need Typesense installed locally to validate media metadata.

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md

Section:
6. Redis / Queue / External Services -> Search / Typesense
```

### Redis

Redis is not used by current media metadata code. No Redis install, port, container, or credentials are newly required.

### Kafka

Kafka is not recommended for beginner local setup.

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task7_Dependency.md

Section:
6. Redis / Queue / External Services -> Kafka
```

Current behavior:

| Config | Result |
|---|---|
| `PRODUCT_EVENT_BROKER=rabbitmq` | App can use built-in RabbitMQ publisher |
| `PRODUCT_EVENT_BROKER=kafka` | App requires injected custom publisher |

## 7. Environment Variables

Complete `.env` creation, loading, and full variable table are already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
6. Environment Variables
```

Where `.env` should be created:

```text
backend/services/product-service/.env
```

Reminder:

Current config uses `os.LookupEnv`. Sirf `.env` file create karne se values auto-load nahi hoti. Shell me `source .env`, direnv, dotenv, ya IDE run configuration use karo.

### New variables for `TASK_FILE_NAME`

No new environment variable was added only for `TASK_FILE_NAME`.

Do not add fake CDN/storage credentials unless implementation actually starts uploading images or validating provider-specific data.

### Task-specific variables to verify

These variables already exist and affect media metadata behavior:

```env
PRODUCT_REQUIRE_PRIMARY_IMAGE_FOR_PUBLISH=false
PRODUCT_MAX_IMAGES_PER_PRODUCT=50
PRODUCT_EVENTS_ENABLED=true
PRODUCT_EVENTS_TOPIC=product.events
PRODUCT_EVENT_BROKER=rabbitmq
RABBITMQ_URL=amqp://ecommerce:ecommerce@localhost:5672/
PRODUCT_OUTBOX_WORKER_ENABLED=true
```

| Variable | New? | Required? | Purpose | Security note |
|---|---:|---:|---|---|
| `PRODUCT_REQUIRE_PRIMARY_IMAGE_FOR_PUBLISH` | No | Optional | If `true`, publish validation fails without an active primary image | Not secret |
| `PRODUCT_MAX_IMAGES_PER_PRODUCT` | No | Optional | Caps gallery size, default `50` | Not secret |
| `PRODUCT_EVENTS_ENABLED` | No | Conditional | Enables product event recording after product changes | Not secret |
| `PRODUCT_EVENTS_TOPIC` | No | Conditional | Event topic/queue, default `product.events` | Not secret |
| `PRODUCT_EVENT_BROKER` | No | Conditional | Use `rabbitmq` for beginner setup | Not secret |
| `RABBITMQ_URL` | No | Conditional | Needed only for RabbitMQ publishing | Secret outside local |
| `PRODUCT_OUTBOX_WORKER_ENABLED` | No | Conditional | Publishes queued events when worker is started by runtime | Not secret |

### Missing future env variables to consider

If the team later implements strict CDN validation or upload support, add real config instead of hardcoding values:

| Future variable idea | Why it may be needed | Current status |
|---|---|---|
| `PRODUCT_MEDIA_ALLOWED_CDN_HOSTS` | Allow only trusted image domains | Not implemented |
| `PRODUCT_MEDIA_REQUIRE_HTTPS` | Enforce HTTPS-only active images | Not implemented |
| `PRODUCT_MEDIA_MAX_ALT_TEXT_LENGTH` | Make alt text limit configurable | Not implemented |
| `PRODUCT_MEDIA_STORAGE_BUCKET` | Needed only if this service uploads/deletes objects | Not implemented |
| `PRODUCT_MEDIA_CDN_BASE_URL` | Needed only if this service constructs public URLs | Not implemented |

Beginner warning:

Environment variable doc me value likhne ka matlab code automatically use karega, aisa nahi hota. Pehle config code me variable read karna padega, validation add karni padegi, and tests update karne padenge.

## 8. Docker Setup

Docker basics are reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
5. Database Analysis: MongoDB -> H. Docker setup
5. Database Analysis: MongoDB -> I. Docker Compose example
7. External Services Analysis -> RabbitMQ
9. Docker and DevOps Setup
```

Task-specific Docker impact:

| Docker item | New for `TASK_FILE_NAME`? | Note |
|---|---:|---|
| MongoDB container | No | Reuse existing local MongoDB setup |
| RabbitMQ container | No | Only needed for event publish verification |
| CDN container | No | No local CDN service is required |
| MinIO/S3 container | No | Current service does not upload image binaries |
| Product service container | No | No Dockerfile exists yet |
| Docker Compose file | No | No repo-owned compose file exists yet |
| New volume | No | Reuse MongoDB/RabbitMQ volumes from previous docs |
| New network | No | No media-specific Docker network added |

If you use Docker MongoDB, run migrations against the same MongoDB instance the app uses. Agar migration host MongoDB pe run hui and app Docker MongoDB pe connect kar raha hai, image schema changes missing lagenge.

## 9. Local Development Setup

### Step 1: Read previous dependency docs first

Follow these first because they already explain shared setup:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
TaskImplementation/{SERVICE_NAME}/task7_Dependency.md
```

Main reused topics:

- Go installation and module commands
- MongoDB installation or Docker setup
- `.env` location and loading
- MongoDB migrations and validator checks
- RabbitMQ setup and local credentials
- Product event/outbox setup
- Generic Docker/MongoDB/RabbitMQ troubleshooting

### Step 2: Go to project directory

For task documentation:

```bash
cd "TaskImplementation/{SERVICE_NAME}"
```

For actual Go module verification:

```bash
cd backend/services/product-service
```

### Step 3: Install only new dependencies if any

No new dependency is introduced by `TASK_FILE_NAME`.

If dependencies are not downloaded yet, use the reused Go module command:

```bash
go mod download
```

### Step 4: Setup only new databases/services if any

No new database or service is introduced.

Use this decision table:

| Verification goal | Needed services |
|---|---|
| Domain image validation tests | No external service |
| DTO mapping tests | No external service |
| MongoDB schema verification | MongoDB + migrations `0003`, `0004` |
| Full service repository tests | MongoDB if tests use real repository |
| Product event publish after image update | MongoDB + RabbitMQ + outbox migration `0007` |

### Step 5: Add only new or changed environment variables

No new env variable is required.

For media behavior, verify existing values in:

```text
backend/services/product-service/.env
```

Minimum values to check:

```env
PRODUCT_REQUIRE_PRIMARY_IMAGE_FOR_PUBLISH=false
PRODUCT_MAX_IMAGES_PER_PRODUCT=50
```

For event verification after media changes, reuse Task 7 event values:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task7_Dependency.md

Section:
7. Environment Variables -> Task-specific event variables
```

### Step 6: Run migrations if needed

For media schema:

```bash
mongosh "mongodb://localhost:27017" migrations/mongo/0003_create_product_collections.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0004_seller_product_workflow.up.js
```

For full current service setup, continue through:

```bash
mongosh "mongodb://localhost:27017" migrations/mongo/0005_product_read_indexes.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0006_inventory_reservations.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0007_product_event_outbox.up.js
```

### Step 7: Start backend service

Current run reality:

| Command | Status |
|---|---|
| `go test ./...` | Supported |
| `go build ./...` | Supported for packages |
| `go run .` | Not supported because no `main` package exists |
| Start HTTP/gRPC server | Not available from this module yet |
| Start outbox worker through app entrypoint | App support exists, but no server/worker `main` exists yet |

So for now, verify through tests/build and MongoDB inspection.

### Step 8: Verify `TASK_FILE_NAME` functionality

Run focused verification:

```bash
cd backend/services/product-service
go test ./internal/domain ./internal/transport/dto ./internal/mapper
go build ./...
```

What to verify:

| Flow | Expected result |
|---|---|
| Image URL is empty | Validation reports image URL required |
| Image URL is not absolute HTTP/HTTPS | Validation reports invalid image URL |
| Image position is `0` or negative | Validation reports invalid position |
| Two images have same position | Product validation reports duplicate image position |
| Two active images are primary | Product validation reports multiple primary images |
| Publish without primary image | Warning or error depending on `PRODUCT_REQUIRE_PRIMARY_IMAGE_FOR_PUBLISH` |
| Legacy seller `images[]` input | DTO converts strings to active image metadata |
| Structured `image_details[]` input | DTO keeps alt text, position, primary flag, variant IDs, dimensions, status |
| Search event mapper | Uses active primary image URL where available |

## 10. Running the Project

Full shared run instructions are reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
11. Complete Project Run Instructions
```

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend API | N/A | HTTP/gRPC API | Reused limitation: no server listener yet |
| MongoDB | 27017 | Product DB with `products.images[]` metadata | Reused |
| RabbitMQ AMQP | 5672 | Event publishing if enabled | Reused and conditional |
| RabbitMQ Management UI | 15672 | Inspect queues/messages | Reused and optional |
| CDN HTTPS | 443 | Public image URLs served by external CDN | External, not local Product Service port |
| MinIO/S3 | N/A | Not used by current service | Not required |
| Redis | N/A | Not used | Not required |
| Kafka | N/A | Not wired by default | Not recommended |
| Typesense | N/A | Search Service dependency, not this service task | Not required here |

Port conflict and Docker network troubleshooting are already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
8. Ports and Networking
```

### Example local media payload

Current seller DTO supports both legacy URL strings and structured details.

Legacy beginner example:

```json
{
  "images": [
    "https://cdn.example.com/products/prod_123/main.jpg"
  ]
}
```

Structured current example:

```json
{
  "image_details": [
    {
      "image_id": "img_001",
      "url": "https://cdn.example.com/products/prod_123/main.jpg",
      "alt_text": "Black running shoes side view",
      "position": 1,
      "is_primary": true,
      "variant_ids": ["var_black_9"],
      "width": 1200,
      "height": 1200,
      "status": "active"
    }
  ]
}
```

Do not include `object_key`, `content_type`, `size_bytes`, or `status=deleted` in current payloads unless code and migrations are updated first.

## 11. Common Errors & Fixes

Generic errors are reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
12. Common Errors and Fixes
```

Task-specific errors:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `IMAGE_URL_REQUIRED` | Image metadata has empty `url` | Send a non-empty public URL | Validate seller payload before save |
| `INVALID_IMAGE_URL` | URL is relative, malformed, or not HTTP/HTTPS | Use absolute URL like `https://cdn.example.com/products/prod_123/main.jpg` | Use Go `net/url` based validation in backend |
| `INVALID_IMAGE_POSITION` | `position` is `0` or negative | Use positions starting from `1` | Normalize positions in seller workflow |
| `DUPLICATE_IMAGE_POSITION` | Two images share same position | Reorder gallery and make positions unique | On drag-drop, rewrite all positions |
| `MULTIPLE_PRIMARY_IMAGES` | More than one active image has `is_primary=true` | Keep only one active primary image | When setting primary, unset others |
| `PRIMARY_IMAGE_REQUIRED` | Publish validation needs primary image | Add active primary image or set env intentionally | Decide `PRODUCT_REQUIRE_PRIMARY_IMAGE_FOR_PUBLISH` per environment |
| `INVALID_IMAGE_STATUS` | Payload used unsupported status like `deleted` | Use `active`, `hidden`, `processing`, or `failed` | Keep frontend status enum synced with backend |
| `image_id is required` in DB or validation | Structured metadata did not include ID and prepare flow did not generate one | Use seller service prepare flow or pass `image_id` | Do not bypass usecase normalization |
| MongoDB rejects `object_key` / `content_type` / `size_bytes` | Current validator does not include these fields | Add future migration and domain fields before storing | Keep task guide examples aligned with current schema |
| Search thumbnail not updated | Event worker not running or product update event not published | Follow Task 7 outbox/RabbitMQ setup | Inspect `product_event_outbox` before checking Search Service |
| CDN image does not load in browser | URL is fake, private, expired, or blocked by CORS/CDN | Use a real public HTTPS image URL for manual UI testing | Upload layer should return stable public URL |

## 12. Security & Best Practices

Shared security and config guidance is reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
13. Security and Configuration Audit
14. Best Practices
```

Task-specific security rules:

- Store image metadata only; do not store binary image bytes in MongoDB.
- Prefer HTTPS image URLs in production even though current validation accepts HTTP too.
- Add trusted CDN host allowlist before accepting arbitrary seller-supplied public URLs in production.
- Keep `object_key` and storage internals out of public read responses unless there is a real requirement.
- Do not put CDN/S3 credentials in markdown, Go code, or committed `.env.example`.
- Treat `alt_text` as user input. Escape/sanitize at render time in frontend templates.
- Keep exactly one active primary image so product cards and search thumbnails stay deterministic.
- Keep gallery positions unique and stable; frontend drag-drop should send rewritten positions.
- Do not expose hidden/processing/failed images in public buyer responses.
- Emit/update product event after material image changes if search thumbnails depend on it.

Best practices specific to `TASK_FILE_NAME`:

| Practice | Why |
|---|---|
| Use structured `image_details[]` for new clients | Keeps alt text, ordering, primary flag, status, and dimensions together |
| Keep legacy `images[]` only for compatibility | URL-only arrays lose metadata |
| Generate missing `image_id` in usecase layer | Avoid frontend-only ID assumptions |
| Keep max image count configurable | Prevent very large product documents |
| Avoid provider-specific fields until needed | Keeps service independent from S3/R2/GCS/CDN choices |
| Add migration before adding new fields | MongoDB validator will reject unsupported fields |
| Test search mapper when primary image rules change | Search thumbnail depends on active primary image selection |

## 13. Missing or Misconfigured Things

| Area | Current observation | Setup risk | Recommended fix |
|---|---|---|---|
| Dedicated media endpoints | Not present | Seller dashboard may need cleaner gallery editor API | Add media-specific handlers/usecases when API surface is implemented |
| Server entrypoint | No `cmd/server/main.go` exists | Cannot start API with `go run .` | Add server/worker entrypoint before writing runtime instructions |
| Dockerfile | No product service Dockerfile exists | Cannot build app container | Add Dockerfile after runtime entrypoint exists |
| docker-compose | No repo-owned local compose exists | Beginners must run MongoDB/RabbitMQ manually | Add local compose with MongoDB and RabbitMQ |
| CDN allowlist config | Not implemented | Any absolute HTTP/HTTPS URL can pass current domain validation | Add `PRODUCT_MEDIA_ALLOWED_CDN_HOSTS` style config and validation |
| HTTPS-only enforcement | Not implemented | HTTP image URLs can pass validation | Enforce HTTPS for active/public images |
| Future fields from task guide | `object_key`, `content_type`, `size_bytes`, `deleted` not in current code/schema | Payload/schema mismatch if frontend sends them | Add domain fields, DTO mapping, migration, and tests before use |
| Alt text length limit | Not enforced in current domain code | Very long alt text can be stored | Add max length validation if product UX requires it |
| Variant existence validation for image links | Current image validation checks non-empty variant IDs, not actual existence | Broken image-to-variant references may pass | Cross-check `variant_ids` against product variants |
| Upload/storage credentials | Not present | Cannot upload/delete binaries from this service | Keep out of scope or implement dedicated storage adapter |
| Health checks | No app health endpoint yet | Hard to verify media/event runtime readiness | Add health/readiness checks with future server entrypoint |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Project Tech Stack Analysis` | Go, MongoDB, RabbitMQ, Docker basics already explained |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Language-Specific Dependency System: Go` | Go module commands and common dependency issues are same |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Analysis: MongoDB` | MongoDB install, Docker run, connection string, credentials are same |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Environment Variables` | `.env` location/loading and common mistakes are same |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. External Services Analysis -> RabbitMQ` | RabbitMQ install, ports, credentials, and verification are same |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Ports and Networking` | Port conflict and Docker networking guidance is same |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Docker and DevOps Setup` | Generic local container setup is same |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `10. Migration Setup` | General migration workflow is same |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | MongoDB decision and `PRODUCT_MONGO_*` naming warnings are reused |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `5. Database Setup` | Base `products` collection and image array validator are reused |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `5. Database Setup` and `8. Docker Setup -> MongoDB transaction warning` | Seller workflow migration adds `image_id`, `alt_text`, `variant_ids`, dimensions; transaction warning remains relevant |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | `6. Redis / Queue / External Services -> Search / Typesense` | Search boundary and read behavior are already clarified |
| `TaskImplementation/{SERVICE_NAME}/task7_Dependency.md` | `6. Redis / Queue / External Services` and `7. Environment Variables` | Event/outbox/RabbitMQ setup is reused for product update events |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before using this file
- [ ] No duplicate Go/MongoDB/Docker/RabbitMQ setup copied into this file
- [ ] `INPUT_FILE_PATH` reviewed for media metadata dependency requirements
- [ ] Original `TASK_FILE_NAME` left unchanged
- [ ] Go dependencies downloaded only if local cache was empty
- [ ] No unnecessary `go get` run for optional validator library
- [ ] `.env` created at `backend/services/product-service/.env` if local config is needed
- [ ] `.env` loaded into shell/IDE before running commands
- [ ] `PRODUCT_REQUIRE_PRIMARY_IMAGE_FOR_PUBLISH` chosen intentionally
- [ ] `PRODUCT_MAX_IMAGES_PER_PRODUCT` verified
- [ ] MongoDB running if schema/repository verification is needed
- [ ] Migrations `0003` and `0004` applied for media schema
- [ ] Migrations through `0007` applied if full event/outbox verification is needed
- [ ] `products.images[]` validator inspected if DB setup is being verified
- [ ] RabbitMQ running only if event publishing is being tested
- [ ] CDN/object storage credentials not added because current service does not upload binaries
- [ ] Payloads avoid unsupported future fields unless code/schema are updated
- [ ] Focused media validation/DTO/mapper tests run
- [ ] `go build ./...` run
- [ ] Common media validation errors understood
- [ ] Missing future items documented instead of hidden
- [ ] No duplicate setup documentation added
