# Product Service - Task 1 Dependency and Setup Documentation

## 1. Purpose

This file explains all setup, dependency, environment, database, and DevOps requirements connected with:

```text
TaskImplementation/Product Service/task1.md
```

Generated output file:

```text
TaskImplementation/Product Service/task1_Dependency.md
```

Important beginner note:

- `task1.md` is mainly a catalog model decision document.
- Task 1 itself did not create a runnable API server, Dockerfile, or database.
- The current repository also contains Product Service Go code, MongoDB migrations, and RabbitMQ event code from later work. Those real dependencies are included here because they matter when a developer tries to run or test the Product Service project.

Simple Hinglish explanation:

Product Service catalog ka source of truth hai. Is service me products, variants, categories, attributes, images, inventory rules, MongoDB collections, and product event outbox ka setup involved hai. Beginner developer ko confuse na ho, isliye yahan local setup, env variables, Docker, migrations, tests, and common errors clearly explain kiye gaye hain.

## 2. Current Implementation Reality

| Area | Current status | Beginner meaning |
|---|---|---|
| Task 1 file | Documentation-only model guide | Is task se direct server run nahi hota |
| Product Service language | Go | Backend code Go modules se manage hota hai |
| Runtime entrypoint | Not present yet | `cmd/server/main.go` ya similar file abhi nahi hai |
| HTTP/gRPC listener | Not present yet | API port currently defined nahi hai |
| Database code | Present | MongoDB repository and schema manager code available hai |
| Migrations | Present | `migrations/mongo/*.js` files se collections/indexes create hote hain |
| Events | Present | RabbitMQ publisher and MongoDB outbox relay available hai |
| Dockerfile | Not present | App container build config abhi nahi hai |
| docker-compose.yml | Not present | Local infra manually run karni padegi ya compose file create karni padegi |

Warning:

Do not assume `go run .` will start Product Service. Current product-service folder has no `main` package. For now, use `go test ./...`, `go build ./...`, MongoDB health checks, and migration verification.

## 3. Project Tech Stack Analysis

| Technology | Required? | Why used | Hinglish + simple English |
|---|---:|---|---|
| Go | Yes | Product Service backend code is written in Go | Go ek compiled backend language hai jo fast APIs and microservices banane ke liye use hoti hai |
| Go modules | Yes | Dependency management through `go.mod` and `go.sum` | Go modules project ke external packages ka version lock karte hain |
| Go workspace | Yes for monorepo workflow | `backend/go.work` includes `./services/product-service` | Workspace se monorepo me multiple Go modules ek saath manage ho sakte hain |
| MongoDB | Yes for real Product Service runtime | Product catalog has flexible attributes and nested variants/images | MongoDB document database hai, dynamic product fields ke liye useful hai |
| MongoDB Go Driver v2 | Yes | Go code connects to MongoDB using official driver | Official driver Go app ko MongoDB se connect/query/update karne deta hai |
| RabbitMQ | Required only when product events are enabled with RabbitMQ | Product event outbox publishes events like product created/updated/published | RabbitMQ message broker hai, services ke beech async events bhejne ke liye use hota hai |
| RabbitMQ AMQP client | Required if RabbitMQ publisher is used | `github.com/rabbitmq/amqp091-go` sends messages to RabbitMQ | Go library jo RabbitMQ AMQP protocol se baat karti hai |
| Docker | Optional but recommended for beginners | Easy local MongoDB and RabbitMQ setup | Docker se DB/broker install kiye bina container me run kar sakte ho |
| mongosh | Required for manual migrations | Runs MongoDB migration JavaScript files | `mongosh` MongoDB shell hai, DB inspect and migration run karne ke liye |
| Structured logging (`log/slog`) | Built-in Go package | JSON logs from app wiring | `slog` Go ka logging package hai, debugging and production logs ke liye |
| Mermaid | Documentation only | Diagrams in task docs | Mermaid markdown me diagrams draw karne ke liye use hota hai |

Not currently used by Product Service runtime:

| Technology | Status |
|---|---|
| Redis | Not used in current Product Service code |
| MySQL/PostgreSQL | Not used by Product Service |
| Kafka | Config validation allows `PRODUCT_EVENT_BROKER=kafka`, but no built-in Kafka publisher is wired |
| Elasticsearch/Typesense | Search is handled by Search Service design, not Product Service runtime |
| MinIO/S3/CDN | Product image URLs are modeled, upload/storage implementation is later scope |
| Kubernetes | Not present in current repo for Product Service |

## 4. Language-Specific Dependency System: Go

### Files involved

| File | Purpose |
|---|---|
| `backend/services/product-service/go.mod` | Defines module name, Go version, and direct dependencies |
| `backend/services/product-service/go.sum` | Locks dependency checksums for reproducible builds |
| `backend/go.work` | Go workspace file that includes Product Service module |

Current module:

```text
module product-service
go 1.26.3
```

Current direct dependencies:

```text
github.com/rabbitmq/amqp091-go v1.10.0
go.mongodb.org/mongo-driver/v2 v2.6.0
```

### What is `go.mod`?

`go.mod` project ka dependency manifest hai. Isme likha hota hai ki project ka module name kya hai, Go version kya chahiye, aur kaunse packages use ho rahe hain.

### What is `go.sum`?

`go.sum` dependency integrity file hai. Isme downloaded packages ke checksums hote hain. Isko commit karna chahiye.

### What is `go.work`?

`go.work` monorepo workspace file hai. Is repo me:

```text
backend/go.work
```

Product Service module ko include karta hai:

```text
use ./services/product-service
```

### Install/download dependencies

From service folder:

```bash
cd backend/services/product-service
go mod download
```

From backend workspace:

```bash
cd backend
go work sync
go test ./services/product-service/...
```

### Tidy dependencies

Use this after adding/removing imports:

```bash
cd backend/services/product-service
go mod tidy
```

Beginner meaning:

`go mod tidy` unused dependencies remove karta hai aur missing dependencies add karta hai. Agar import add kiya but dependency missing hai, tidy usually fix kar deta hai.

### Build

```bash
cd backend/services/product-service
go build ./...
```

Current note:

`go build ./...` library packages compile karega. Since no `main` package exists, executable binary generate nahi hoga unless future `cmd/server/main.go` add hota hai.

### Test

```bash
cd backend/services/product-service
go test ./...
```

### Run

Current state:

```bash
cd backend/services/product-service
go run .
```

This is expected to fail because Product Service currently has no `main` package.

Future expected style after entrypoint is added:

```bash
cd backend/services/product-service
go run ./cmd/server
```

### Common Go dependency issues

| Error | Cause | Fix |
|---|---|---|
| `go: go.mod file not found` | Wrong directory | Run from `backend/services/product-service` or `backend` workspace |
| `module requires Go 1.26.3` | Old Go installed | Install matching Go version or newer compatible patch |
| `missing go.sum entry` | Dependency checksum missing | Run `go mod download` or `go mod tidy` |
| `package product-service/internal/... is not in std` | Wrong module path or command run outside module | Run commands from service folder |
| Proxy/cache error | Go proxy/network issue | Try `go clean -modcache`, set `GOPROXY=https://proxy.golang.org,direct`, then run download again |
| Version mismatch | Dependency version conflict | Run `go mod tidy`, review `go.mod`, avoid random manual edits |

## 5. Database Analysis: MongoDB

### A. What is MongoDB?

MongoDB ek document database hai. MySQL jaise rows/tables ke instead yeh JSON-like documents store karta hai.

Simple example:

```json
{
  "title": "Running Shoes",
  "variants": [
    {"sku": "SHOE-BLK-9", "price": {"amount": 299900, "currency": "INR"}}
  ]
}
```

### B. Why Product Service uses MongoDB

Product catalog flexible hota hai:

- Shoes me `size`, `color`, `material`
- Electronics me `ram`, `storage`, `processor`
- Grocery me `weight`, `expiry`, `pack_size`

MongoDB nested documents and dynamic attributes ke liye suitable hai. Product, variants, images, category path, rating summary, and inventory snapshot ek natural document style me model ho sakte hain.

### C. Required or optional?

| Scope | MongoDB required? |
|---|---:|
| Only reading `task1.md` documentation | No |
| Running Product Service repository code with Mongo repositories | Yes |
| Running migrations | Yes |
| Running only pure unit tests that use in-memory mocks | No |

### D. Collections used

Migrations and collection definition code define these collections:

| Collection | Purpose |
|---|---|
| `products` | Product aggregate, variants, images, status, attributes |
| `categories` | Category tree and attribute schema |
| `brands` | Brand metadata |
| `inventory_reservations` | Temporary stock reservations during checkout |
| `inventory_snapshots` | Inventory audit/history snapshots |
| `price_books` | Seller/platform pricing rules |
| `product_event_outbox` | Pending/published product events for async delivery |

### E. Default port

```text
27017
```

### F. Connection string format

Without username/password for local development:

```env
PRODUCT_MONGO_URI=mongodb://localhost:27017
PRODUCT_MONGO_DATABASE=product_db
```

With username/password:

```env
PRODUCT_MONGO_URI=mongodb://product_root:change-me@localhost:27017/product_db?authSource=admin
PRODUCT_MONGO_DATABASE=product_db
```

MongoDB Atlas style:

```env
PRODUCT_MONGO_URI=mongodb+srv://product_app:<password>@<cluster-host>/product_db?retryWrites=true&w=majority
PRODUCT_MONGO_DATABASE=product_db
```

Security note:

Never commit real MongoDB passwords. Put real credentials only in `.env`, CI secrets, Kubernetes secrets, or a secret manager.

### G. Local installation

#### Windows

Option 1: Docker Desktop recommended.

```powershell
docker --version
docker pull mongo:8
```

Option 2: Native MongoDB.

1. Install MongoDB Community Server from MongoDB official installer.
2. Install MongoDB Shell (`mongosh`).
3. Start MongoDB service from Windows Services.
4. Verify:

```powershell
mongosh "mongodb://localhost:27017"
```

#### Linux

Docker recommended:

```bash
docker --version
docker pull mongo:8
```

Native package install depends on distro. For Ubuntu/Debian, follow MongoDB official package repo steps, then verify:

```bash
mongosh "mongodb://localhost:27017"
```

#### macOS

Docker Desktop recommended:

```bash
docker --version
docker pull mongo:8
```

Homebrew option:

```bash
brew tap mongodb/brew
brew install mongodb-community mongosh
brew services start mongodb-community
mongosh "mongodb://localhost:27017"
```

### H. Docker setup

Simple local MongoDB without auth:

```bash
docker volume create product_mongo_data
docker run -d \
  --name ecommerce-product-mongo \
  -p 27017:27017 \
  -v product_mongo_data:/data/db \
  mongo:8
```

Local MongoDB with root auth:

```bash
docker volume create product_mongo_data
docker run -d \
  --name ecommerce-product-mongo \
  -p 27017:27017 \
  -v product_mongo_data:/data/db \
  -e MONGO_INITDB_ROOT_USERNAME=product_root \
  -e MONGO_INITDB_ROOT_PASSWORD=change-me \
  mongo:8
```

If auth is enabled, use:

```env
PRODUCT_MONGO_URI=mongodb://product_root:change-me@localhost:27017/product_db?authSource=admin
```

### I. Docker Compose example

No `docker-compose.yml` currently exists for Product Service. If the project adds one later, this is a beginner-friendly local infra example:

```yaml
services:
  product-mongo:
    image: mongo:8
    container_name: ecommerce-product-mongo
    ports:
      - "27017:27017"
    volumes:
      - product_mongo_data:/data/db
    restart: unless-stopped

  product-rabbitmq:
    image: rabbitmq:3-management
    container_name: ecommerce-product-rabbitmq
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      RABBITMQ_DEFAULT_USER: ecommerce
      RABBITMQ_DEFAULT_PASS: ecommerce
    volumes:
      - product_rabbitmq_data:/var/lib/rabbitmq
    restart: unless-stopped

volumes:
  product_mongo_data:
  product_rabbitmq_data:
```

Run if/when this file exists:

```bash
docker compose up -d
docker compose ps
docker compose logs -f product-mongo
docker compose logs -f product-rabbitmq
docker compose down
```

### J. Verify MongoDB running

If using local host:

```bash
mongosh "mongodb://localhost:27017" --eval "db.adminCommand({ ping: 1 })"
```

If using Docker:

```bash
docker exec ecommerce-product-mongo mongosh --eval "db.adminCommand({ ping: 1 })"
```

List product DB collections after migration:

```bash
mongosh "mongodb://localhost:27017/product_db" --eval "db.getCollectionNames()"
```

### K. Where credentials go

| Credential | File/location | Variable |
|---|---|---|
| Mongo host/port/user/password | `backend/services/product-service/.env` | `PRODUCT_MONGO_URI` |
| Mongo database name | `backend/services/product-service/.env` | `PRODUCT_MONGO_DATABASE` |
| Example values | `backend/services/product-service/.env.example` | same keys |
| Production secrets | Secret manager/CI/Kubernetes secrets | same keys |

Important:

The migration files currently hardcode `product_db` using `db.getSiblingDB("product_db")`. If you change `PRODUCT_MONGO_DATABASE`, also update migration execution strategy or migration scripts.

## 6. Environment Variables

### Where `.env` should be created

Create local env file here:

```text
backend/services/product-service/.env
```

Use the example:

```bash
cd backend/services/product-service
cp .env.example .env
```

Security note:

`.gitignore` already ignores `.env` and nested `.env` files. Keep it that way.

### How the project loads env variables

Current Product Service config uses:

```go
os.LookupEnv(...)
```

Meaning:

- The code reads environment variables from the running process.
- It does not automatically parse `.env`.
- Copying `.env.example` to `.env` is not enough unless your shell/tool loads it.

Load `.env` in Linux/macOS/Git Bash:

```bash
cd backend/services/product-service
set -a
source .env
set +a
go test ./...
```

PowerShell beginner option:

```powershell
cd backend/services/product-service
$env:PRODUCT_MONGO_URI="mongodb://localhost:27017"
$env:PRODUCT_MONGO_DATABASE="product_db"
go test ./...
```

### Complete `.env` example

```env
SERVICE_NAME=product-service
ENVIRONMENT=local
LOG_LEVEL=info

PRODUCT_DEFAULT_CURRENCY=INR
PRODUCT_STRICT_ATTRIBUTE_SCHEMA=true
PRODUCT_REQUIRE_PRIMARY_IMAGE_FOR_PUBLISH=false
PRODUCT_MAX_IMAGES_PER_PRODUCT=50
PRODUCT_MAX_VARIANTS_PER_PRODUCT=250

PRODUCT_READ_DEFAULT_PAGE_SIZE=20
PRODUCT_READ_MAX_PAGE_SIZE=100
PRODUCT_READ_MAX_BATCH_SIZE=100

PRODUCT_CMS_CATALOG_MANAGEMENT_ALLOWED=true
PRODUCT_CMS_MODERATION_DECISION=auto_publish
PRODUCT_CMS_ALLOW_DRAFT_WRITES_WHEN_UNAVAILABLE=true

PRODUCT_MONGO_URI=mongodb://localhost:27017
PRODUCT_MONGO_DATABASE=product_db
PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=false

PRODUCT_INVENTORY_DEFAULT_TTL_SECONDS=900
PRODUCT_INVENTORY_MIN_TTL_SECONDS=30
PRODUCT_INVENTORY_MAX_TTL_SECONDS=3600
PRODUCT_INVENTORY_EXPIRY_BATCH_LIMIT=100

PRODUCT_EVENTS_ENABLED=true
PRODUCT_EVENTS_TOPIC=product.events
PRODUCT_EVENT_BROKER=rabbitmq
RABBITMQ_URL=amqp://ecommerce:ecommerce@localhost:5672/

PRODUCT_OUTBOX_POLL_INTERVAL_MS=1000
PRODUCT_OUTBOX_BATCH_SIZE=100
PRODUCT_OUTBOX_MAX_ATTEMPTS=5
PRODUCT_OUTBOX_WORKER_ENABLED=true
PRODUCT_OUTBOX_PUBLISH_TIMEOUT_MS=10000
```

### Variable reference

| Variable | Required? | Example | Purpose | Security notes |
|---|---:|---|---|---|
| `SERVICE_NAME` | Yes | `product-service` | Log/event source name | Not secret |
| `ENVIRONMENT` | Yes | `local` | Runtime environment label | Not secret |
| `LOG_LEVEL` | Optional | `info` | Logging verbosity: `debug`, `info`, `warn`, `error` | Avoid `debug` in production if logs may contain sensitive data |
| `PRODUCT_DEFAULT_CURRENCY` | Yes | `INR` | Default money currency for product prices | Not secret, must be 3-letter code |
| `PRODUCT_STRICT_ATTRIBUTE_SCHEMA` | Optional | `true` | Reject unknown category attributes when strict | Not secret |
| `PRODUCT_REQUIRE_PRIMARY_IMAGE_FOR_PUBLISH` | Optional | `false` | Require primary image before publish | Not secret |
| `PRODUCT_MAX_IMAGES_PER_PRODUCT` | Optional | `50` | Limit product gallery size | Not secret |
| `PRODUCT_MAX_VARIANTS_PER_PRODUCT` | Optional | `250` | Limit variants per product | Not secret |
| `PRODUCT_READ_DEFAULT_PAGE_SIZE` | Optional | `20` | Default list page size | Not secret |
| `PRODUCT_READ_MAX_PAGE_SIZE` | Optional | `100` | Maximum allowed list page size | Not secret |
| `PRODUCT_READ_MAX_BATCH_SIZE` | Optional | `100` | Maximum batch product reads | Not secret |
| `PRODUCT_CMS_CATALOG_MANAGEMENT_ALLOWED` | Optional | `true` | Static CMS permission fallback | Not secret, but affects admin/seller behavior |
| `PRODUCT_CMS_MODERATION_DECISION` | Optional | `auto_publish` | Static moderation fallback: `auto_publish` or `review_required` | Not secret |
| `PRODUCT_CMS_ALLOW_DRAFT_WRITES_WHEN_UNAVAILABLE` | Optional | `true` | Allow draft writes if CMS unavailable | Not secret, production should review carefully |
| `PRODUCT_MONGO_URI` | Yes for DB runtime | `mongodb://localhost:27017` | MongoDB server connection | Secret if username/password included |
| `PRODUCT_MONGO_DATABASE` | Yes for DB runtime | `product_db` | MongoDB database name | Not secret |
| `PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS` | Optional | `false` | App can create collections if collection manager is wired | In production prefer controlled migrations |
| `PRODUCT_INVENTORY_DEFAULT_TTL_SECONDS` | Optional | `900` | Default reservation hold time | Not secret |
| `PRODUCT_INVENTORY_MIN_TTL_SECONDS` | Optional | `30` | Minimum reservation TTL | Not secret |
| `PRODUCT_INVENTORY_MAX_TTL_SECONDS` | Optional | `3600` | Maximum reservation TTL | Not secret |
| `PRODUCT_INVENTORY_EXPIRY_BATCH_LIMIT` | Optional | `100` | Expired reservation batch size | Not secret |
| `PRODUCT_EVENTS_ENABLED` | Optional | `true` | Enables product event recording/outbox | Not secret |
| `PRODUCT_EVENTS_TOPIC` | Required if events enabled | `product.events` | RabbitMQ queue/topic name | Not secret |
| `PRODUCT_EVENT_BROKER` | Required if events enabled | `rabbitmq` | Event broker selection | Kafka not wired by default |
| `RABBITMQ_URL` | Required for RabbitMQ | `amqp://ecommerce:ecommerce@localhost:5672/` | RabbitMQ AMQP connection string | Secret because it contains password |
| `PRODUCT_OUTBOX_POLL_INTERVAL_MS` | Optional | `1000` | Outbox worker polling interval | Not secret |
| `PRODUCT_OUTBOX_BATCH_SIZE` | Optional | `100` | Events processed per batch | Not secret |
| `PRODUCT_OUTBOX_MAX_ATTEMPTS` | Optional | `5` | Retry attempts before dead-letter | Not secret |
| `PRODUCT_OUTBOX_WORKER_ENABLED` | Optional | `true` | Starts relay when app entrypoint calls workers | Not secret |
| `PRODUCT_OUTBOX_PUBLISH_TIMEOUT_MS` | Optional | `10000` | Publish timeout per event | Not secret |

### Common `.env` mistakes

| Mistake | Result | Fix |
|---|---|---|
| `.env` created in repo root only | Product Service commands may not see it | Put it in `backend/services/product-service/.env` or export vars globally |
| `.env` copied but not loaded | App uses defaults | Use `source .env` or a dotenv loader |
| Spaces around `=` | Shell may fail to export | Use `KEY=value` |
| Wrong boolean value | Config falls back silently for booleans | Use `true` or `false` |
| Wrong integer value | Config falls back silently for numbers | Use numeric values like `100` |
| `LOG_LEVEL=verbose` | Config validation fails | Use `debug`, `info`, `warn`, or `error` |
| `PRODUCT_EVENT_BROKER=kafka` | App wiring fails unless custom publisher provided | Use `rabbitmq` locally |

## 7. External Services Analysis

### MongoDB

| Item | Detail |
|---|---|
| What it is | Document database |
| Why used | Flexible product catalog data |
| Mandatory? | Yes for real DB-backed Product Service runtime |
| Local install | Docker or native MongoDB |
| Docker image | `mongo:8` |
| Health check | `db.adminCommand({ ping: 1 })` |
| Credentials | `PRODUCT_MONGO_URI` |
| Default port | `27017` |

Common issues:

- MongoDB not running: start container/service.
- Auth failed: check username/password and `authSource`.
- Migration failed: run `0003_create_product_collections.up.js` before later migrations.

### RabbitMQ

| Item | Detail |
|---|---|
| What it is | Message broker |
| Why used | Product event outbox publishes async events |
| Mandatory? | Required when `PRODUCT_EVENTS_ENABLED=true`, `PRODUCT_OUTBOX_WORKER_ENABLED=true`, and broker is RabbitMQ |
| Local install | Docker recommended |
| Docker image | `rabbitmq:3-management` |
| AMQP port | `5672` |
| Management UI port | `15672` |
| Credentials | `RABBITMQ_URL` |

Docker run:

```bash
docker volume create product_rabbitmq_data
docker run -d \
  --name ecommerce-product-rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=ecommerce \
  -e RABBITMQ_DEFAULT_PASS=ecommerce \
  -v product_rabbitmq_data:/var/lib/rabbitmq \
  rabbitmq:3-management
```

Verify:

```bash
docker exec ecommerce-product-rabbitmq rabbitmq-diagnostics ping
```

Open management UI:

```text
http://localhost:15672
```

Login:

```text
username: ecommerce
password: ecommerce
```

Security note:

`ecommerce/ecommerce` is okay only for local development. Production me strong password and secret manager use karo.

### Kafka

Current status:

- Config validation allows `PRODUCT_EVENT_BROKER=kafka`.
- `app.New` returns an error unless a custom `ProductEventPublisher` dependency is injected.
- No Kafka client dependency and no Kafka env variables are present in `go.mod` or config.

Beginner recommendation:

Do not set Kafka locally for this service right now. Use RabbitMQ.

### Redis

Redis is not used in current Product Service code. Agar Redis container already project me chal raha hai, Product Service Task 1 ke liye uski zarurat nahi hai.

### Docker

Docker app runtime ke liye mandatory nahi hai, but local MongoDB/RabbitMQ ke liye recommended hai.

Verify Docker:

```bash
docker --version
docker ps
```

If Docker daemon is not running, start Docker Desktop or system service.

## 8. Ports and Networking

| Service | Port | Purpose | Current status |
|---|---:|---|---|
| Product Service API | N/A | HTTP/gRPC API | No server entrypoint/listener yet |
| MongoDB | 27017 | Product catalog database | Required for DB runtime |
| RabbitMQ AMQP | 5672 | Product event publishing | Required when RabbitMQ events enabled |
| RabbitMQ Management | 15672 | Browser admin UI | Optional local debugging |

### Port conflicts

Check which process uses a port:

```bash
lsof -i :27017
lsof -i :5672
lsof -i :15672
```

Docker alternative:

```bash
docker ps
```

Fix:

- Stop the conflicting process/container.
- Or map a different host port.

Example:

```bash
docker run -d --name ecommerce-product-mongo -p 27018:27017 mongo:8
```

Then set:

```env
PRODUCT_MONGO_URI=mongodb://localhost:27018
```

### Docker network issues

From host machine, use:

```env
PRODUCT_MONGO_URI=mongodb://localhost:27017
RABBITMQ_URL=amqp://ecommerce:ecommerce@localhost:5672/
```

From another container in the same compose network, use service names:

```env
PRODUCT_MONGO_URI=mongodb://product-mongo:27017
RABBITMQ_URL=amqp://ecommerce:ecommerce@product-rabbitmq:5672/
```

### Firewall issues

If connection works inside Docker but not from host:

- Check Docker Desktop port forwarding.
- Check local firewall.
- Confirm container is published with `-p`.

## 9. Docker and DevOps Setup

### What exists now

| Artifact | Exists? | Notes |
|---|---:|---|
| Product Service Dockerfile | No | Needs future `deploy/Dockerfile` or similar |
| Root/local docker-compose.yml | No | Local infra commands are manual right now |
| MongoDB migrations | Yes | JavaScript files under `migrations/mongo` |
| `.env.example` | Yes | Present under product-service folder |

### Recommended beginner approach

Use Docker for external services only:

```bash
docker run -d --name ecommerce-product-mongo -p 27017:27017 -v product_mongo_data:/data/db mongo:8
docker run -d --name ecommerce-product-rabbitmq -p 5672:5672 -p 15672:15672 -e RABBITMQ_DEFAULT_USER=ecommerce -e RABBITMQ_DEFAULT_PASS=ecommerce rabbitmq:3-management
```

Then run Go commands on host:

```bash
cd backend/services/product-service
go test ./...
```

### Useful Docker commands

```bash
docker ps
docker logs ecommerce-product-mongo
docker logs ecommerce-product-rabbitmq
docker stop ecommerce-product-mongo ecommerce-product-rabbitmq
docker start ecommerce-product-mongo ecommerce-product-rabbitmq
```

Remove containers but keep named volumes:

```bash
docker rm -f ecommerce-product-mongo ecommerce-product-rabbitmq
```

Remove local data volumes only when you intentionally want a clean DB:

```bash
docker volume rm product_mongo_data product_rabbitmq_data
```

Warning:

Removing volumes deletes local database/broker data.

### Volumes

| Volume | Purpose |
|---|---|
| `product_mongo_data` | Keeps MongoDB data after container restart |
| `product_rabbitmq_data` | Keeps RabbitMQ queues/users after container restart |

### Restart policies

For local dev, use:

```bash
--restart unless-stopped
```

For production, restart policies should be managed by orchestrator like Kubernetes/systemd.

## 10. Migration Setup

### Migration files

```text
backend/services/product-service/migrations/mongo/
```

Important files:

| File | Purpose |
|---|---|
| `0003_create_product_collections.up.js` | Creates core product collections and indexes |
| `0004_seller_product_workflow.up.js` | Adds seller workflow fields/validation |
| `0005_product_read_indexes.up.js` | Adds read/listing indexes |
| `0006_inventory_reservations.up.js` | Creates inventory reservation collection/indexes |
| `0007_product_event_outbox.up.js` | Creates product event outbox collection/indexes |

### Run migrations with local mongosh

```bash
cd backend/services/product-service
mongosh "mongodb://localhost:27017" migrations/mongo/0003_create_product_collections.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0004_seller_product_workflow.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0005_product_read_indexes.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0006_inventory_reservations.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0007_product_event_outbox.up.js
```

### Run migrations into Docker MongoDB

```bash
cd backend/services/product-service
docker exec -i ecommerce-product-mongo mongosh < migrations/mongo/0003_create_product_collections.up.js
docker exec -i ecommerce-product-mongo mongosh < migrations/mongo/0004_seller_product_workflow.up.js
docker exec -i ecommerce-product-mongo mongosh < migrations/mongo/0005_product_read_indexes.up.js
docker exec -i ecommerce-product-mongo mongosh < migrations/mongo/0006_inventory_reservations.up.js
docker exec -i ecommerce-product-mongo mongosh < migrations/mongo/0007_product_event_outbox.up.js
```

With Mongo auth:

```bash
docker exec -i ecommerce-product-mongo mongosh \
  -u product_root \
  -p change-me \
  --authenticationDatabase admin \
  < migrations/mongo/0003_create_product_collections.up.js
```

Repeat for each migration.

### Verify migrations

```bash
mongosh "mongodb://localhost:27017/product_db" --eval "db.getCollectionNames()"
```

Expected collections:

```text
products
categories
brands
inventory_reservations
inventory_snapshots
price_books
product_event_outbox
```

Check indexes:

```bash
mongosh "mongodb://localhost:27017/product_db" --eval "db.products.getIndexes().map(i => i.name)"
mongosh "mongodb://localhost:27017/product_db" --eval "db.product_event_outbox.getIndexes().map(i => i.name)"
```

### Auto-create collections option

Config supports:

```env
PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=true
```

But this works only when the app is wired with `CollectionSchemaManager`. Current service has no runnable main entrypoint, so manual migrations are clearer for beginners.

Production recommendation:

Use explicit migration pipeline, not auto-create, so DB changes are reviewed and repeatable.

## 11. Complete Project Run Instructions

### Step 1: Clone repository

```bash
git clone <repository-url>
cd Ecommerce
```

### Step 2: Go to Product Service

```bash
cd backend/services/product-service
```

### Step 3: Install Go

Check version:

```bash
go version
```

Use the version from `go.mod`:

```text
go 1.26.3
```

### Step 4: Install Go dependencies

```bash
go mod download
go mod tidy
```

### Step 5: Start MongoDB

```bash
docker volume create product_mongo_data
docker run -d \
  --name ecommerce-product-mongo \
  -p 27017:27017 \
  -v product_mongo_data:/data/db \
  mongo:8
```

Verify:

```bash
docker exec ecommerce-product-mongo mongosh --eval "db.adminCommand({ ping: 1 })"
```

### Step 6: Start RabbitMQ

Required if product events/outbox worker are enabled:

```bash
docker volume create product_rabbitmq_data
docker run -d \
  --name ecommerce-product-rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=ecommerce \
  -e RABBITMQ_DEFAULT_PASS=ecommerce \
  -v product_rabbitmq_data:/var/lib/rabbitmq \
  rabbitmq:3-management
```

Verify:

```bash
docker exec ecommerce-product-rabbitmq rabbitmq-diagnostics ping
```

### Step 7: Create `.env`

```bash
cp .env.example .env
```

For simple local Docker setup, default values are fine:

```env
PRODUCT_MONGO_URI=mongodb://localhost:27017
PRODUCT_MONGO_DATABASE=product_db
RABBITMQ_URL=amqp://ecommerce:ecommerce@localhost:5672/
```

### Step 8: Load `.env`

```bash
set -a
source .env
set +a
```

### Step 9: Run MongoDB migrations

```bash
mongosh "mongodb://localhost:27017" migrations/mongo/0003_create_product_collections.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0004_seller_product_workflow.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0005_product_read_indexes.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0006_inventory_reservations.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0007_product_event_outbox.up.js
```

### Step 10: Run tests

```bash
go test ./...
```

### Step 11: Build packages

```bash
go build ./...
```

### Step 12: Start backend service

Current state:

Product Service does not yet have a `cmd/server/main.go` or equivalent entrypoint. So there is no real backend process to start from this service folder right now.

Future command once server entrypoint exists:

```bash
go run ./cmd/server
```

### Step 13: Verify APIs

Current state:

No HTTP/gRPC listener is implemented in the current Product Service tree. So API health check is not available yet.

Use these checks for now:

```bash
go test ./...
mongosh "mongodb://localhost:27017/product_db" --eval "db.getCollectionNames()"
docker exec ecommerce-product-rabbitmq rabbitmq-diagnostics ping
```

## 12. Common Errors and Fixes

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `go: go.mod file not found` | Command run from wrong folder | `cd backend/services/product-service` | Run Go commands from module folder |
| `package ... is not a main package` | No server entrypoint yet | Use `go test ./...`; wait/add `cmd/server/main.go` | Do not document `go run .` as startup |
| `connection refused localhost:27017` | MongoDB not running or wrong port | Start Mongo container/service | Verify with `mongosh ping` |
| `Authentication failed` | Wrong Mongo credentials/authSource | Fix `PRODUCT_MONGO_URI` | Keep one local credential source |
| `products collection must exist before applying...` | Migrations run out of order | Run `0003` before `0004/0005` | Apply migrations sequentially |
| Duplicate key error on SKU/slug | Test/sample data uses repeated unique values | Change SKU/slug or clean DB | Use unique fixtures |
| `Docker daemon not running` | Docker Desktop/service stopped | Start Docker | Check `docker ps` before setup |
| `port is already allocated` | Another container/process uses port | Stop old container or change host port | Use consistent container names |
| `connect rabbitmq: dial tcp ... connection refused` | RabbitMQ not running or wrong URL | Start RabbitMQ, check `RABBITMQ_URL` | Verify `rabbitmq-diagnostics ping` |
| RabbitMQ login refused | Wrong user/pass in URL | Match Docker env and `RABBITMQ_URL` | Store local creds in `.env` |
| Kafka broker failure | Kafka selected but publisher not implemented | Use `PRODUCT_EVENT_BROKER=rabbitmq` | Avoid Kafka until implementation exists |
| `.env` changes ignored | `.env` not loaded into process | `source .env` before command | Use direnv/dotenv tool |
| `LOG_LEVEL must be one of...` | Invalid log level | Use `debug`, `info`, `warn`, `error` | Copy `.env.example` |
| `PRODUCT_DEFAULT_CURRENCY must be valid` | Currency not 3-letter code | Use `INR`, `USD`, etc. | Validate env before deploy |
| Permission denied on scripts/files | OS permission issue | Fix file permissions or run terminal as correct user | Avoid root-owned workspace files |
| `go mod download` network/proxy failure | Internet/proxy/cache issue | Retry, set `GOPROXY`, clear mod cache | Keep `go.sum` committed |
| `mongosh: command not found` | Mongo shell not installed | Install `mongosh` or use Docker exec | Prefer Docker for beginners |

## 13. Security and Configuration Audit

### Findings

| Area | Observation | Risk | Recommended fix |
|---|---|---|---|
| `.env` handling | Code reads process env only | Beginner may think `.env` auto-loads | Add startup docs or a dotenv loader in future entrypoint |
| Local credentials | `.env.example` uses `ecommerce/ecommerce` for RabbitMQ | Unsafe if copied to production | Mark local-only and use secrets in prod |
| MongoDB default URI | Defaults to unauthenticated localhost | Fine for local, unsafe for shared env | Require authenticated URI outside local |
| Auto-create collections | Supported by config | Can hide migration discipline | Keep false in prod and run reviewed migrations |
| Kafka config | Broker value accepted but no built-in Kafka publisher | Runtime confusion | Either implement Kafka publisher or remove/disable option |
| Service startup | No `cmd/server/main.go` | No health checks/API start possible | Add entrypoint, port config, graceful shutdown |
| Docker config | No Dockerfile/compose | Onboarding friction | Add local compose for Mongo/RabbitMQ and app Dockerfile later |
| Health checks | No app health endpoint yet | Hard to verify runtime | Add `/healthz` or gRPC health service when server exists |
| Secrets | `.gitignore` ignores `.env` | Good | Continue never committing real `.env` |
| Migration DB name | Scripts use `product_db` directly | Env DB mismatch possible | Parameterize migrations or document fixed DB name |

### Hardcoded credentials

Local-only defaults found:

```env
RABBITMQ_URL=amqp://ecommerce:ecommerce@localhost:5672/
```

This should not be used in production.

### Missing environment variables

There is no API port variable because no server entrypoint exists yet. When a server is added, add something like:

```env
PRODUCT_SERVICE_PORT=8080
```

or a gRPC-specific port variable.

### Missing Docker config

Recommended future files:

```text
backend/services/product-service/deploy/Dockerfile
infra/compose/docker-compose.local.yml
```

### Missing health checks

Recommended future checks:

- MongoDB ping during startup
- RabbitMQ connection check if events enabled
- App health endpoint
- Readiness endpoint that fails if mandatory dependencies are unavailable

## 14. Best Practices

- Never commit `.env`.
- Keep `.env.example` updated whenever config changes.
- Use Docker volumes for local MongoDB and RabbitMQ data.
- Use strong RabbitMQ and MongoDB passwords outside local development.
- Run migrations in order.
- Keep `go.mod` and `go.sum` committed.
- Run `go test ./...` before pushing.
- Do not use float for money; keep amount in minor units.
- Keep `PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=false` in production.
- Use separate local, staging, and production configuration.
- Add health checks before deploying a real service.
- Keep logs structured, but avoid logging passwords/tokens.
- Prefer explicit event broker implementation instead of config-only support.
- Back up MongoDB before destructive migrations.

## 15. Beginner Mental Model

Think of Product Service setup like this:

```text
Go code
  needs Go modules
  uses config from process env
  connects to MongoDB for product data
  optionally publishes product events through RabbitMQ
  needs migrations before real DB usage
  currently has tests/buildable packages but no server main
```

Hinglish:

Pehle Go dependencies install karo. Phir MongoDB and RabbitMQ local containers start karo. Phir `.env` banao and load karo. Phir migrations run karo. Abhi API server start nahi hoga kyunki entrypoint missing hai, but tests/build se implementation verify ho sakti hai.

## 16. Final Checklist

- [ ] Repository cloned
- [ ] Go version installed according to `go.mod`
- [ ] Product Service folder opened
- [ ] `go mod download` completed
- [ ] MongoDB installed/running
- [ ] RabbitMQ installed/running if events enabled
- [ ] `backend/services/product-service/.env` created
- [ ] `.env` loaded into shell/process
- [ ] `PRODUCT_MONGO_URI` verified
- [ ] `RABBITMQ_URL` verified if RabbitMQ enabled
- [ ] MongoDB migrations run in order
- [ ] Collections verified in `product_db`
- [ ] `go test ./...` passing
- [ ] `go build ./...` passing
- [ ] No real secrets committed
- [ ] Common errors reviewed
- [ ] Future server entrypoint/health check still tracked as pending work

## 17. Quick Command Summary

```bash
cd backend/services/product-service

go mod download

docker volume create product_mongo_data
docker run -d --name ecommerce-product-mongo -p 27017:27017 -v product_mongo_data:/data/db mongo:8

docker volume create product_rabbitmq_data
docker run -d --name ecommerce-product-rabbitmq -p 5672:5672 -p 15672:15672 -e RABBITMQ_DEFAULT_USER=ecommerce -e RABBITMQ_DEFAULT_PASS=ecommerce -v product_rabbitmq_data:/var/lib/rabbitmq rabbitmq:3-management

cp .env.example .env
set -a
source .env
set +a

mongosh "mongodb://localhost:27017" migrations/mongo/0003_create_product_collections.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0004_seller_product_workflow.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0005_product_read_indexes.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0006_inventory_reservations.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0007_product_event_outbox.up.js

go test ./...
go build ./...
```
