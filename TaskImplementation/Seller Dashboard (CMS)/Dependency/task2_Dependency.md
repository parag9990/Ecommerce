# Project Dependency & Setup Guide

## 1. Project Overview

### Variables used by this guide

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `Seller Dashboard (CMS)` |
| `TASK_FILE_NAME` | `task2.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task2_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |

This dependency guide is for the product manager work documented in `INPUT_FILE_PATH`.

Simple Hinglish goal: Task 2 me seller ko product list, create/edit form, variants, image URLs, categories, and publish action use karna hai. Frontend code already `frontend/seller-dashboard/src/features/products/` me present hai. Is guide ka kaam business logic repeat karna nahi hai; sirf setup, dependencies, environment, database, services, and DevOps requirements explain karna hai.

### What changed from Task 1

Task 1 dashboard shell mostly frontend + gateway + seller session setup par depend karta tha. Task 2 product manager ke liye extra backend runtime dependencies chahiye:

| New / Reused | Dependency | Why it matters for Task 2 |
|---|---|---|
| Reused | Frontend React/Vite/pnpm setup | Product manager same seller dashboard app me run hota hai. |
| Reused | API Gateway on `8080` | Browser REST calls gateway ko hit karti hain. |
| Reused | gRPC / Protobuf contract | Gateway Product Service methods ko call karega. |
| Reused | Redis | Gateway rate limiting ke liye, if enabled. |
| Reused | MySQL / CMS setup | Seller session/team/CMS side already documented hai. |
| New | Product Service boundary | Product list/create/update/publish and categories ke liye required. |
| New | MongoDB `product_db` | Product catalog dynamic attributes/variants store karne ke liye required. |
| New | RabbitMQ, if product events stay enabled | Product publish/update events `product.events` topic/queue flow ke liye required. |

### Runtime flow

```text
Seller browser
  -> Vite Seller Dashboard frontend
  -> API Gateway REST API
  -> ProductService gRPC methods
  -> MongoDB product_db
  -> RabbitMQ product.events, only if product events/outbox worker enabled
```

Important: Frontend directly MongoDB, RabbitMQ, Redis, or MySQL ko access nahi karta. Browser sirf API Gateway ko call karta hai.

## 2. Tech Stack

### Reused stack

Frontend, pnpm workspace, API Gateway, Redis, MySQL, Go Modules, gRPC, and Protobuf basics already explain kiye gaye hain.

Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Also refer:
`TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md`
`TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md`
`TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md`
`TaskImplementation/${SERVICE_NAME}/Dependency/gRPC.md`
`TaskImplementation/${SERVICE_NAME}/Dependency/Protobuf.md`
`TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md`
`TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md`

### Task 2 specific stack

| Technology | Required? | Beginner-friendly explanation | Why Task 2 uses it |
|---|---:|---|---|
| React Hook Form | Yes | React Hook Form form state manage karta hai without har input ke saath heavy re-render. | Product create/edit form large hai: title, category, images, attributes, variants. |
| Zod | Yes | Zod validation schema library hai. Simple words me, form data valid hai ya nahi ye check karta hai. | Product title, category, image URL, variant SKU, price, and stock rules validate karne ke liye. |
| `@hookform/resolvers` | Yes | Ye React Hook Form aur Zod ke beech bridge hai. | Zod schema ko product form resolver ke roop me use karne ke liye. |
| Product Service | Yes for real product data | Product Service catalog ka backend owner hai. | Product list, product detail, category list, create, update, publish APIs isi boundary se aati hain. |
| MongoDB | Yes for Product Service | MongoDB document database hai. Tables ke bajay JSON-like documents store karta hai. | Product attributes category-wise dynamic hain, isliye flexible schema useful hai. |
| RabbitMQ | Required if events enabled | RabbitMQ message broker hai. Services async events exchange kar sakti hain. | Product Service env me product events/outbox worker enabled hai. Publish/update ke baad search/index consumers later event consume kar sakte hain. |

No new frontend package installation is required right now because these packages are already present in `frontend/seller-dashboard/package.json`.

## 3. Required Software

Common software is already documented in previous dependency files:

| Software | Follow this existing doc |
|---|---|
| Git | `task1_Dependency.md`, section `Required Software` |
| Node.js, Corepack, pnpm | `Dependency/Frontend.md`, sections `Installation steps` and `Local setup without Docker` |
| Go | `Dependency/Go_Modules.md`, section `Installation steps` |
| Docker / Docker Compose | `task1_Dependency.md`, sections `Required Software` and `Docker Setup` |
| MySQL | `Dependency/MySQL.md` |
| Redis | `Dependency/Redis.md` |
| curl | `task1_Dependency.md`, section `Required Software` |

New software for Task 2:

| Software | Required? | Verify command | Notes |
|---|---:|---|---|
| MongoDB server | Yes for real Product Service | `mongod --version` | Can be local install or Docker container. |
| MongoDB shell `mongosh` | Recommended | `mongosh --version` | Useful for creating collections, indexes, and seed categories. |
| RabbitMQ | Required if product events enabled | `rabbitmqctl status` | Docker is easiest for local setup. |

Beginner note: Agar sirf frontend UI open karna hai, Node + pnpm enough hai. Agar product list/create/publish real API ke saath test karna hai, then API Gateway, Product Service, MongoDB, RabbitMQ if enabled, Auth/User seller session, and Redis if gateway rate limit enabled chahiye.

## 4. Dependency Management

### Frontend dependencies

Task 2 product manager uses the same pnpm workspace already documented.

This setup is already explained in:
`TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md`

Section:
`Installation steps`

Task-specific dependency status:

| Package | Used by | Current status |
|---|---|---|
| `react-hook-form` | `ProductForm` | Already in `frontend/seller-dashboard/package.json` |
| `zod` | `product-validation.ts` | Already in `frontend/seller-dashboard/package.json` |
| `@hookform/resolvers` | `zodResolver(productFormSchema)` | Already in `frontend/seller-dashboard/package.json` |
| `@tanstack/react-query` | Product queries/mutations | Already documented and installed |
| `lucide-react` | Product buttons/icons | Already documented and installed |

Do not run `pnpm add` for Task 2 unless a package is missing from `frontend/seller-dashboard/package.json`.

After normal install from the previous guide, Task 2 relevant checks are:

```bash
cd frontend
pnpm --filter seller-dashboard typecheck
pnpm --filter seller-dashboard test
```

### Backend dependencies

Product Service is expected to be a Go service, but current project inspection found only:

```text
backend/services/product-service/.env
```

Not clearly found:

```text
backend/services/product-service/go.mod
backend/services/product-service/cmd/server/main.go
backend/services/product-service/internal/
```

So Go module installation/start commands are currently blocked until Product Service code exists.

Reuse:
`TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md`

Important Task 2 backend gap:

```text
backend/go.work currently includes only ./services/auth-service
```

When Product Service source is added, it should be added to the workspace:

```bash
cd backend
go work use ./services/product-service
go work sync
```

## 5. Database Setup

### Reused database: MySQL

MySQL setup is already explained in:
`TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md`

Use that for CMS-owned data like seller staff, settings, coupons, campaigns, and audit logs.

Task 2 product manager does not store product catalog data in CMS MySQL. Product catalog belongs to Product Service.

### New database: MongoDB

#### What it is

MongoDB ek document database hai. Isme data JSON-like documents ke form me store hota hai. Relational tables ki tarah fixed columns nahi hote.

#### Why Task 2 uses it

Product attributes dynamic hote hain. Example:

- T-shirt: size, color, fabric
- Mobile: RAM, storage, battery
- Grocery: weight, expiry, pack size

Isliye Product Service ke liye MongoDB better fit hai. Project docs also mention Product Service database as MongoDB.

#### Required or optional

Required for real Product Service APIs:

- `GET /api/v1/products`
- `GET /api/v1/products/{product_id}`
- `GET /api/v1/categories`
- `POST /api/v1/seller/products`
- `PATCH /api/v1/seller/products/{product_id}`
- `POST /api/v1/seller/products/{product_id}/publish`

Optional only if frontend is tested with mocked API responses.

#### Evidence in project

| Path | Evidence |
|---|---|
| `backend/services/product-service/.env` | `PRODUCT_MONGO_URI`, `PRODUCT_MONGO_DATABASE` |
| `docs/04-microservice-design.md` | Product Service database choice is MongoDB |
| `database/mongodb-schema-design.md` | Product collections and indexes |
| `api/master-api.json` | Product and category API contracts |

#### Default port

```text
27017
```

#### Connection string format

Local without auth:

```env
PRODUCT_MONGO_URI=mongodb://localhost:27017
PRODUCT_MONGO_DATABASE=product_db
```

With auth, recommended for shared/dev/prod:

```env
PRODUCT_MONGO_URI=mongodb://product_user:change-this-locally@localhost:27017/product_db?authSource=product_db
PRODUCT_MONGO_DATABASE=product_db
```

Credentials placement:

```text
backend/services/product-service/.env
```

Do not put MongoDB credentials in frontend `VITE_*` variables.

#### Local install

If MongoDB is already installed, just verify:

```bash
mongod --version
mongosh --version
```

Ubuntu/Debian style install varies by MongoDB repository setup. Beginner-friendly local option is Docker, shown below.

#### Docker setup for MongoDB

No project docker-compose file was found, so this is a local helper command, not an existing project file.

```bash
docker volume create ecommerce-product-mongo-data

docker run -d \
  --name ecommerce-product-mongo \
  -p 27017:27017 \
  -v ecommerce-product-mongo-data:/data/db \
  mongo:7
```

Verify:

```bash
docker ps
mongosh "mongodb://localhost:27017/product_db" --eval 'db.runCommand({ ping: 1 })'
```

Expected ping result contains:

```text
ok: 1
```

#### Product collections and indexes

No Product Service migration/seed folder was clearly found. For local bootstrap, create basic collections and indexes manually:

```bash
mongosh "mongodb://localhost:27017/product_db"
```

Inside `mongosh`:

```javascript
db.createCollection("products")
db.createCollection("categories")
db.createCollection("brands")
db.createCollection("inventory_snapshots")
db.createCollection("price_books")

db.products.createIndex({ seller_id: 1, status: 1, updated_at: -1 })
db.products.createIndex({ category_id: 1, status: 1, updated_at: -1 })
db.products.createIndex({ "variants.sku": 1 }, { unique: true })
db.products.createIndex({ title: "text", description: "text", brand: "text" })

db.categories.createIndex({ parent_id: 1, sort_order: 1 })
db.categories.createIndex({ slug: 1 }, { unique: true })
```

#### Seed at least one category

Product form requires `category_id`. Agar category list empty hai, create product form practical use ke liye blocked ho jayega.

Local seed example:

```javascript
db.categories.updateOne(
  { category_id: "cat-shirts" },
  {
    $set: {
      category_id: "cat-shirts",
      name: "Shirts",
      parent_id: null,
      slug: "shirts",
      sort_order: 10
    }
  },
  { upsert: true }
)
```

Verify:

```javascript
db.categories.find({}, { _id: 0, category_id: 1, name: 1 }).toArray()
```

## 6. Redis / Queue / External Services

### Redis

Redis setup is reused from:
`TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md`

Task 2 does not add a direct Redis dependency in frontend. Redis still matters if API Gateway rate limiting remains enabled:

```env
RATE_LIMIT_ENABLED=true
REDIS_ADDR=localhost:6379
```

### API Gateway and Product Service

API Gateway setup is reused from:
`TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md`

Task 2 specifically needs the gateway Product Service target:

```env
PRODUCT_GRPC_ADDR=localhost:50053
```

Current issue: `backend/services/product-service/.env` does not clearly define a Product Service gRPC listen address. Product Service implementation should either:

- listen on `:50053`, matching gateway `PRODUCT_GRPC_ADDR=localhost:50053`, or
- update gateway `PRODUCT_GRPC_ADDR` to the actual Product Service address.

### New queue dependency: RabbitMQ

#### What it is

RabbitMQ ek message broker hai. Simple words me, ek service event publish karti hai aur doosri service later consume kar sakti hai.

#### Why Task 2 may use it

`backend/services/product-service/.env` has:

```env
PRODUCT_EVENTS_ENABLED=true
PRODUCT_EVENTS_TOPIC=product.events
PRODUCT_EVENT_BROKER=rabbitmq
RABBITMQ_URL=amqp://ecommerce:ecommerce@localhost:5672/
PRODUCT_OUTBOX_WORKER_ENABLED=true
```

Iska matlab Product Service product changes/publish events outbox se RabbitMQ par publish karne ke liye configured hai.

#### Required or optional

Required if these stay enabled:

```env
PRODUCT_EVENTS_ENABLED=true
PRODUCT_OUTBOX_WORKER_ENABLED=true
```

Optional for local UI-only Product CRUD if Product Service supports disabling events:

```env
PRODUCT_EVENTS_ENABLED=false
PRODUCT_OUTBOX_WORKER_ENABLED=false
```

Use disable option only for local development. Production me product publish/search-index sync ke liye events important honge.

#### Default ports

| Port | Purpose |
|---:|---|
| `5672` | AMQP application connection |
| `15672` | RabbitMQ management UI/API, only if management image/plugin enabled |

#### Docker setup for RabbitMQ

No project compose file was found, so this is a local helper command.

```bash
docker volume create ecommerce-rabbitmq-data

docker run -d \
  --name ecommerce-rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=ecommerce \
  -e RABBITMQ_DEFAULT_PASS=ecommerce \
  -v ecommerce-rabbitmq-data:/var/lib/rabbitmq \
  rabbitmq:3-management
```

Verify container:

```bash
docker ps
docker logs ecommerce-rabbitmq
```

Verify management API:

```bash
curl -u ecommerce:ecommerce http://localhost:15672/api/overview
```

RabbitMQ URL for Product Service:

```env
RABBITMQ_URL=amqp://ecommerce:ecommerce@localhost:5672/
```

Security note: `ecommerce:ecommerce` is okay only for local. Real environments must use strong credentials and private networking.

### Image storage services

Task 2 does not introduce S3, MinIO, Firebase Storage, or any upload service.

Reason:

- Product API schema uses `images: string[]`.
- Frontend `ImageUploader` currently accepts image URLs.
- File upload is optional through an `onUpload` adapter, but no upload endpoint is wired.

So do not add S3/MinIO setup for Task 2 unless a real upload API is implemented later.

## 7. Environment Variables

Common frontend, CMS, gateway, MySQL, and Redis env variables are already covered in:

`TaskImplementation/${SERVICE_NAME}/Dependency/Environment.md`

Task 2 adds Product Service env requirements.

### Product Service `.env` location

```text
backend/services/product-service/.env
```

### Task 2 `.env` example for Product Service

Use this as a local reference. Do not commit real production secrets.

```env
SERVICE_NAME=product-service
ENVIRONMENT=local
LOG_LEVEL=info

# Product business defaults
PRODUCT_DEFAULT_CURRENCY=INR
PRODUCT_STRICT_ATTRIBUTE_SCHEMA=true
PRODUCT_REQUIRE_PRIMARY_IMAGE_FOR_PUBLISH=false
PRODUCT_MAX_IMAGES_PER_PRODUCT=50
PRODUCT_MAX_VARIANTS_PER_PRODUCT=250

# Product read pagination
PRODUCT_READ_DEFAULT_PAGE_SIZE=20
PRODUCT_READ_MAX_PAGE_SIZE=100
PRODUCT_READ_MAX_BATCH_SIZE=100

# CMS/catalog behavior
PRODUCT_CMS_CATALOG_MANAGEMENT_ALLOWED=true
PRODUCT_CMS_MODERATION_DECISION=auto_publish
PRODUCT_CMS_ALLOW_DRAFT_WRITES_WHEN_UNAVAILABLE=true

# MongoDB
PRODUCT_MONGO_URI=mongodb://localhost:27017
PRODUCT_MONGO_DATABASE=product_db
PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=false

# Inventory guardrails
PRODUCT_INVENTORY_DEFAULT_TTL_SECONDS=900
PRODUCT_INVENTORY_MIN_TTL_SECONDS=30
PRODUCT_INVENTORY_MAX_TTL_SECONDS=3600
PRODUCT_INVENTORY_EXPIRY_BATCH_LIMIT=100

# Events and outbox
PRODUCT_EVENTS_ENABLED=true
PRODUCT_EVENTS_TOPIC=product.events
PRODUCT_EVENT_BROKER=rabbitmq
RABBITMQ_URL=amqp://ecommerce:ecommerce@localhost:5672/
PRODUCT_OUTBOX_POLL_INTERVAL_MS=1000
PRODUCT_OUTBOX_BATCH_SIZE=100
PRODUCT_OUTBOX_MAX_ATTEMPTS=5
PRODUCT_OUTBOX_WORKER_ENABLED=true
PRODUCT_OUTBOX_PUBLISH_TIMEOUT_MS=10000

# Recommended missing variable, align with API Gateway PRODUCT_GRPC_ADDR
PRODUCT_GRPC_ADDR=:50053
```

### New / changed variable explanation

| Variable | Required? | Purpose | Example | Security notes |
|---|---:|---|---|---|
| `PRODUCT_MONGO_URI` | Yes | Mongo server connection | `mongodb://localhost:27017` | Add auth/TLS outside local. |
| `PRODUCT_MONGO_DATABASE` | Yes | Product DB name | `product_db` | Keep one DB per environment. |
| `PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS` | Depends | Auto-create collections toggle | `false` | Production should use reviewed migrations/seed scripts. |
| `PRODUCT_DEFAULT_CURRENCY` | Yes | Default money currency | `INR` | Must match frontend `Money.currency`. |
| `PRODUCT_READ_DEFAULT_PAGE_SIZE` | Yes | Default product list size | `20` | Keep aligned with frontend initial `pageSize`. |
| `PRODUCT_READ_MAX_PAGE_SIZE` | Yes | API abuse guardrail | `100` | Prevent huge reads. |
| `PRODUCT_CMS_MODERATION_DECISION` | Yes | Publish workflow behavior | `auto_publish` | Review before production. |
| `PRODUCT_REQUIRE_PRIMARY_IMAGE_FOR_PUBLISH` | Yes | Publish validation rule | `false` | If set `true`, frontend should enforce/communicate image requirement. |
| `PRODUCT_EVENTS_ENABLED` | Yes | Product event publishing toggle | `true` | Disable only for local if RabbitMQ is unavailable. |
| `PRODUCT_EVENT_BROKER` | If events enabled | Broker choice | `rabbitmq` | Must match available broker. |
| `PRODUCT_EVENTS_TOPIC` | If events enabled | Event topic/routing name | `product.events` | Do not put secrets here. |
| `RABBITMQ_URL` | If RabbitMQ enabled | Broker connection URL | `amqp://ecommerce:ecommerce@localhost:5672/` | Contains credentials, keep secret outside local. |
| `PRODUCT_OUTBOX_WORKER_ENABLED` | If events enabled | Background publisher toggle | `true` | If true, broker must be healthy. |
| `PRODUCT_GRPC_ADDR` | Recommended | Product Service gRPC listen address | `:50053` | Not found in current env, but needed to align with gateway. |

### Gateway variable to confirm

In `backend/services/api-gateway/.env`, confirm:

```env
PRODUCT_GRPC_ADDR=localhost:50053
```

If Product Service listens on a different port, update either Product Service or gateway so both match.

## 8. Docker Setup

Existing Docker gaps are already documented in:

`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
`TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md`
`TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md`

Task 2 new containers are MongoDB and RabbitMQ.

### Task 2 local container list

| Container | Image | Port | Volume | Status |
|---|---|---:|---|---|
| Product MongoDB | `mongo:7` | `27017` | `ecommerce-product-mongo-data` | New |
| RabbitMQ | `rabbitmq:3-management` | `5672`, `15672` | `ecommerce-rabbitmq-data` | New if events enabled |

### Suggested local compose snippet

No compose file was found in the repo. This snippet is only a future reference, not an existing project file.

```yaml
services:
  product-mongo:
    image: mongo:7
    ports:
      - "27017:27017"
    volumes:
      - product_mongo_data:/data/db

  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      RABBITMQ_DEFAULT_USER: ecommerce
      RABBITMQ_DEFAULT_PASS: ecommerce
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq

volumes:
  product_mongo_data:
  rabbitmq_data:
```

### Docker networking note

If backend services run on your host machine, published ports like `localhost:27017` and `localhost:5672` work.

If Product Service runs inside Docker Compose, use service names instead:

```env
PRODUCT_MONGO_URI=mongodb://product-mongo:27017
RABBITMQ_URL=amqp://ecommerce:ecommerce@rabbitmq:5672/
```

## 9. Local Development Setup

### Step 1: Read previous dependency documentation first

Follow these existing docs before Task 2 specific setup:

```text
TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md
TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md
TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md
TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md
TaskImplementation/${SERVICE_NAME}/Dependency/gRPC.md
TaskImplementation/${SERVICE_NAME}/Dependency/Protobuf.md
```

### Step 2: Go to project directory

From repo root:

```bash
cd /home/parag/Ecommerce
```

### Step 3: Install frontend dependencies

No Task 2-specific `pnpm add` is needed. Use the existing frontend setup from `Dependency/Frontend.md`.

Quick command:

```bash
cd frontend
pnpm install
```

### Step 4: Start new Task 2 databases/services

Start MongoDB:

```bash
docker start ecommerce-product-mongo
```

If container does not exist, create it using the command from section `Database Setup`.

Start RabbitMQ if product events are enabled:

```bash
docker start ecommerce-rabbitmq
```

If RabbitMQ is not available and you only need local Product CRUD, set:

```env
PRODUCT_EVENTS_ENABLED=false
PRODUCT_OUTBOX_WORKER_ENABLED=false
```

### Step 5: Configure Product Service env

Review:

```text
backend/services/product-service/.env
```

Minimum local values:

```env
PRODUCT_MONGO_URI=mongodb://localhost:27017
PRODUCT_MONGO_DATABASE=product_db
PRODUCT_DEFAULT_CURRENCY=INR
PRODUCT_READ_DEFAULT_PAGE_SIZE=20
PRODUCT_READ_MAX_PAGE_SIZE=100
```

If using RabbitMQ:

```env
PRODUCT_EVENTS_ENABLED=true
PRODUCT_EVENT_BROKER=rabbitmq
RABBITMQ_URL=amqp://ecommerce:ecommerce@localhost:5672/
PRODUCT_OUTBOX_WORKER_ENABLED=true
```

### Step 6: Seed categories

Product form needs categories from:

```text
GET /api/v1/categories
```

Seed at least one local category in MongoDB using section `Seed at least one category`.

### Step 7: Start backend services

Current repo does not clearly contain runnable Product Service or API Gateway Go source. After those modules exist, expected order:

1. Start MongoDB.
2. Start RabbitMQ if events enabled.
3. Start Redis if gateway rate limiting enabled.
4. Start Auth/User seller session boundary.
5. Start Product Service on the port gateway expects.
6. Start API Gateway on `8080`.
7. Start Seller Dashboard frontend.

Suggested Product Service command after implementation:

```bash
cd backend/services/product-service
go run ./cmd/server
```

Suggested API Gateway command after implementation:

```bash
cd backend/services/api-gateway
go run ./cmd/server
```

### Step 8: Start frontend

Use reused frontend command:

```bash
cd frontend
pnpm --filter seller-dashboard dev
```

Open:

```text
http://localhost:5174/seller/products
```

## 10. Running the Project

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Seller Dashboard Vite dev server | `5174` | Browser app | Reused |
| Seller Dashboard preview | `4174` | Built app preview | Reused |
| API Gateway | `8080` | REST entrypoint | Reused |
| Product Service gRPC | `50053` | Gateway to Product Service | New for Task 2, expected by gateway |
| MongoDB Product DB | `27017` | Product catalog storage | New |
| RabbitMQ AMQP | `5672` | Product events broker | New if events enabled |
| RabbitMQ Management | `15672` | Local broker UI/API | New if using management image |
| Redis | `6379` | Gateway rate limit | Reused |
| MySQL | `3306` | CMS DB | Reused |

Port conflict fixes:

| Problem | Fix |
|---|---|
| `27017` already used | Use existing MongoDB or map Docker to another host port, then update `PRODUCT_MONGO_URI`. |
| `5672` already used | Use existing RabbitMQ or change host mapping and `RABBITMQ_URL`. |
| `50053` already used | Change Product Service listen port and gateway `PRODUCT_GRPC_ADDR` together. |
| `8080` already used | Change gateway `HTTP_ADDR` and frontend `VITE_API_BASE_URL` together. |

### Task 2 API verification

These endpoints come from `api/master-api.json` and `frontend/seller-dashboard/src/features/products/api/seller-product-api.ts`.

Public/read checks:

```bash
curl http://localhost:8080/api/v1/categories

curl "http://localhost:8080/api/v1/products?seller_id=seller_123&page=1&page_size=20"
```

Seller-auth write checks need a valid seller session/JWT/cookie:

```bash
curl -X POST http://localhost:8080/api/v1/seller/products \
  -H "content-type: application/json" \
  --cookie "session=replace-with-local-session-cookie" \
  -d '{
    "title": "Cotton T-shirt",
    "category_id": "cat-shirts",
    "images": ["https://cdn.example.com/products/tshirt.jpg"],
    "attributes": {"material": "cotton"},
    "variants": [
      {
        "sku": "TSHIRT-BLK-M",
        "attributes": {"size": "M", "color": "black"},
        "price": {"amount": 49900, "currency": "INR"},
        "stock_quantity": 25
      }
    ]
  }'
```

Publish check:

```bash
curl -X POST http://localhost:8080/api/v1/seller/products/prod_123/publish \
  --cookie "session=replace-with-local-session-cookie"
```

### Frontend behavior to verify

| Route | Expected behavior |
|---|---|
| `/seller/products` | Product table loads for active seller. |
| `/seller/products/new` | Product create form opens. |
| `/seller/products/:productId/edit` | Existing product loads and seller ownership is checked. |
| Category dropdown | Shows categories from `GET /api/v1/categories`. |
| Publish button | Visible for draft/approved/unpublished products. |
| Image manager | Accepts valid image URLs only. |

## 11. Common Errors & Fixes

Only Task 2-specific errors are listed here. Generic pnpm, MySQL, Redis, Docker, and gateway errors are already documented in previous dependency files.

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| Product list shows network error | API Gateway not running or `VITE_API_BASE_URL` wrong | Start gateway or set frontend env correctly | Reuse `Dependency/API_Gateway.md` checks |
| Gateway returns Product gRPC unavailable | Product Service not running or `PRODUCT_GRPC_ADDR` mismatch | Start Product Service and align port with gateway | Keep Product Service listen env and gateway target in sync |
| Mongo connection refused | MongoDB not running or wrong `PRODUCT_MONGO_URI` | Start MongoDB and verify `mongosh ... ping` | Add compose/health check |
| Category dropdown empty | `categories` collection empty or category API not implemented | Seed categories in `product_db` | Add seed script/migration for local dev |
| Product create fails with category error | `category_id` not found in Product Service | Use valid category from `GET /api/v1/categories` | Do not hardcode random category IDs |
| `Product response was empty or invalid.` | API response shape does not match frontend normalizer | Return product object in gateway envelope/data shape expected by `http.ts` | Contract test Product API responses |
| Image URL validation error | Image field is not a valid URL | Use full URL like `https://.../image.jpg` | Add upload adapter later if real files are required |
| Image shape mismatch | Mongo schema doc shows image objects, API/frontend uses string URLs | Align Product Service mapper with `api/master-api.json` or update contract and frontend together | Keep API contract as source of truth |
| Duplicate SKU error | `variants.sku` unique index conflict | Use unique SKU per variant | Validate SKU uniqueness before submit |
| Publish fails when RabbitMQ is down | Product publish emits event/outbox worker needs broker | Start RabbitMQ or disable events locally | Add broker health check and graceful failure policy |
| Create button missing | Active seller role lacks `products:write` | Use seller/session role `seller`, `seller_manager`, or `seller_catalog_editor` | Backend should return correct roles in seller session |
| Edit page says product unavailable | Product `seller_id` differs from active seller | Switch seller or open correct product | Backend must enforce ownership too |

## 12. Security & Best Practices

### Product ownership

Frontend sends/uses `seller_id` for filtering, but backend must never trust frontend seller IDs blindly.

Best practice:

- Gateway/Auth should resolve seller identity from session/JWT.
- Product Service should validate product ownership on create/update/publish.
- Seller cannot update another seller's product even if they change URL or request body.

### Validation

Frontend Zod validation improves UX, but backend validation is mandatory.

Backend should validate:

- title length
- category existence
- SKU uniqueness
- price amount and currency
- stock quantity non-negative
- max images and variants from env
- publish status transitions

### Images

Task 2 accepts image URLs. That means backend should guard against unsafe URLs.

Recommended checks:

- accept only `https://` URLs in production
- optionally allow trusted CDN domains
- do not fetch arbitrary internal URLs from backend without SSRF protection
- if file uploads are added later, document S3/MinIO and virus/content checks separately

### MongoDB

- Enable auth/TLS outside local.
- Add indexes before large data import.
- Back up product catalog regularly.
- Do not store seller secrets or auth tokens in product documents.
- Keep product document size under MongoDB document limits.

### RabbitMQ

- Change local `ecommerce:ecommerce` credentials outside local.
- Use durable queues/exchanges for important product events.
- Make outbox publishing idempotent.
- Do not publish product event before DB write commits.
- Monitor dead-letter queue when consumers fail.

### Frontend env

Do not place secrets in `VITE_*` variables. Vite variables are bundled into browser JavaScript.

## 13. Missing or Misconfigured Things

| Item | Current observation | Impact | Suggested fix |
|---|---|---|---|
| Product Service Go source | Not clearly found, only `.env` exists | Product APIs cannot run locally from current files | Add Product Service module, entrypoint, config, repository, gRPC server |
| Product Service `go.mod` | Not found | Go commands cannot run | Create module and add to `backend/go.work` |
| Product Service gRPC listen env | `PRODUCT_GRPC_ADDR` not found in product `.env` | Gateway target may not align | Add `PRODUCT_GRPC_ADDR=:50053` or equivalent config |
| `backend/go.work` | Only `./services/auth-service` is listed | Product Service not part of workspace | Add real service module paths |
| MongoDB migrations/seed scripts | Not found | Manual collection/index/category setup needed | Add versioned Mongo setup or seed command |
| RabbitMQ compose/service config | Not found | Local event broker setup manual | Add local compose service |
| `.env.example` files | Not clearly found | Beginners may copy unsafe real `.env` | Add sanitized examples |
| Docker Compose local stack | Not found | Multi-service startup is manual | Add compose for MongoDB, RabbitMQ, Redis, MySQL, gateway, services |
| Product image schema mismatch | API/frontend use `images: string[]`; Mongo doc shows image objects | Mapper bugs possible | Align contract and persistence model intentionally |
| Product unpublish REST route | gRPC method listed, REST route not listed | UI should not call unpublish yet | Add route later or keep UI publish-only |
| Health checks | Product/Gateway health routes not clearly found | Harder debugging | Add `/health/live`, `/health/ready`, gRPC health |
| Backend ownership enforcement | Not inspectable because source missing | Security risk if not implemented | Enforce seller ownership in Product Service |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | Project overview, required software, frontend install, common run flow | Task 2 runs inside the same dashboard shell. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Frontend.md` | Installation steps, local setup, start commands, frontend env | Same React/Vite/pnpm app is reused. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/API_Gateway.md` | Gateway purpose, env, local flow, common errors | Product APIs are still called through Gateway. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Environment.md` | Frontend, Gateway, CMS env handling | Common env rules already documented. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Go_Modules.md` | Go module/workspace setup | Product Service is expected to be Go, same backend dependency system applies. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/gRPC.md` | gRPC setup, downstream address behavior | Gateway maps Product REST routes to ProductService gRPC methods. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Protobuf.md` | Proto generation and contract rules | ProductService methods require typed backend contracts. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/MySQL.md` | MySQL install, Docker, env, troubleshooting | CMS MySQL setup is unchanged and should not be repeated. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Redis.md` | Redis install, Docker, env, troubleshooting | Gateway rate limiting dependency is unchanged. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/CMS_Backend.md` | CMS backend boundary and existing gaps | CMS setup remains separate from Product Service setup. |
| `TaskImplementation/${SERVICE_NAME}/Dependency/Migrations.md` | Migration expectations and missing migration tooling | Same migration discipline applies; Product Mongo setup lacks scripts. |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before Task 2 setup.
- [ ] Frontend dependencies installed through existing pnpm workspace.
- [ ] No duplicate frontend install documentation added beyond references.
- [ ] MongoDB running on `localhost:27017` or `PRODUCT_MONGO_URI` updated.
- [ ] `product_db` created.
- [ ] Product collections and indexes created or Product Service auto-migration implemented.
- [ ] At least one category seeded for the category selector.
- [ ] RabbitMQ running if `PRODUCT_EVENTS_ENABLED=true`.
- [ ] `RABBITMQ_URL` matches local RabbitMQ credentials and port.
- [ ] Gateway `PRODUCT_GRPC_ADDR` matches Product Service listen address.
- [ ] Product Service source/module exists before trying `go run`.
- [ ] Product Service added to `backend/go.work` after module creation.
- [ ] API Gateway started on the URL used by `VITE_API_BASE_URL`.
- [ ] `/api/v1/categories` verified.
- [ ] `/api/v1/products?seller_id=...` verified.
- [ ] Create/update/publish tested with a valid seller session.
- [ ] Product ownership and role permissions verified backend-side.
- [ ] Logs checked for MongoDB, RabbitMQ, Product Service, and API Gateway errors.
- [ ] No real secrets added to frontend `VITE_*` env values.
- [ ] Missing Docker Compose, health checks, `.env.example`, and migration/seed scripts tracked as follow-up work.
