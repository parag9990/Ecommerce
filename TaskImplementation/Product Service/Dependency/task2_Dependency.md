# Project Dependency & Setup Guide

## Variable Values Used

```text
SERVICE_NAME=Product Service
TASK_FILE_NAME=task2.md
INPUT_FILE_PATH=TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME=task2_Dependency.md
OUTPUT_FILE_PATH=TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

This file is generated for `INPUT_FILE_PATH` and saved as `OUTPUT_FILE_PATH`.

Important beginner note:

- `TASK_FILE_NAME` is a documentation-only MongoDB decision guide.
- It chooses MongoDB as the primary catalog database for `SERVICE_NAME`.
- It does not create a new runnable server, Dockerfile, compose file, or new API endpoint by itself.
- The current repository already contains Go code, MongoDB migrations, MongoDB repositories, and RabbitMQ event code from later implementation work. Those runtime dependencies are documented in the previous dependency file, so this file references them instead of repeating everything.

## 1. Project Overview

`TASK_FILE_NAME` answers one main question:

```text
Which database should be used for the product catalog?
```

Final decision:

```text
MongoDB is the primary database for product catalog data.
```

Simple Hinglish explanation:

MongoDB ek document database hai. Product catalog me category-wise dynamic attributes, nested variants, images, price details, inventory snapshot, aur flexible metadata hota hai. Isliye rigid SQL table design ke comparison me MongoDB yahan natural fit hai.

### Current implementation reality

| Area | Status | Setup impact |
|---|---|---|
| `TASK_FILE_NAME` | Documentation-only decision guide | No new command required only for reading the task |
| Go service code | Present in `backend/services/product-service` | Use Go module setup from previous dependency file |
| MongoDB connection code | Present | MongoDB is required for real DB-backed runtime |
| MongoDB migrations | Present | Run migrations before real local DB usage |
| RabbitMQ event code | Present | Required only when product event worker is enabled |
| Server entrypoint | Not present | `go run .` is not a valid startup command yet |
| Dockerfile / docker-compose | Not present | Use manual local infra commands from previous dependency file |

## 2. Tech Stack

Most technology setup is already explained in:

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

| Technology | Required? | Why it matters for `TASK_FILE_NAME` | New explanation needed? |
|---|---:|---|---|
| Go | Yes for current service code | Backend implementation is Go based | No, reused from previous dependency doc |
| Go modules | Yes | `go.mod` and `go.sum` manage packages | No, reused |
| MongoDB | Yes for real catalog runtime | Chosen as primary product catalog store | Yes, decision-specific notes are below |
| MongoDB Go Driver v2 | Yes in current code | Connects Go repository code to MongoDB | No, setup reused |
| Docker | Optional but recommended | Easy local MongoDB/RabbitMQ containers | No, setup reused |
| mongosh | Required for manual migrations | Runs MongoDB migration JavaScript files | No, setup reused |
| RabbitMQ | Conditional | Product events/outbox from later code | No new setup for this task |
| Redis | No | Not used by current Product Service code | No |
| Kafka | Not wired by default | Config accepts it, but no built-in publisher exists | No local setup recommended |
| Typesense | Not a Product Service runtime dependency | Search Service owns search index | No |

## 3. Required Software

No new software is introduced only by `TASK_FILE_NAME`.

Follow the previous dependency file first:

```text
This setup is already explained in:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
3. Project Tech Stack Analysis
4. Language-Specific Dependency System: Go
5. Database Analysis: MongoDB
7. External Services Analysis
9. Docker and DevOps Setup
```

### Required locally for real Product Service verification

| Software | Purpose | Status for this task |
|---|---|---|
| Go `1.26.3` | Build/test current service packages | Reused |
| Docker | Run MongoDB/RabbitMQ containers easily | Reused |
| MongoDB or `mongo` Docker image | Product catalog database | Reused, but MongoDB choice is task-specific |
| `mongosh` | Run and verify migrations | Reused |
| RabbitMQ Docker image | Event worker local testing | Reused, conditional |

## 4. Dependency Management

Dependency setup is already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
4. Language-Specific Dependency System: Go
```

### What changed in `TASK_FILE_NAME`?

Nothing new was installed by the task document itself.

Current service module still uses:

```text
backend/services/product-service/go.mod
backend/services/product-service/go.sum
backend/go.work
```

Current direct runtime dependencies include:

| Dependency | Why used |
|---|---|
| `go.mongodb.org/mongo-driver/v2` | Official MongoDB Go driver |
| `github.com/rabbitmq/amqp091-go` | RabbitMQ AMQP client for product events |

Beginner note:

`TASK_FILE_NAME` recommends MongoDB. The current repository has already moved beyond only recommendation and has MongoDB driver usage in repository code. So do not run `go get` again unless you are intentionally changing dependencies.

Use previous doc commands for:

- `go mod download`
- `go mod tidy`
- `go test ./...`
- `go build ./...`
- common Go module fixes

## 5. Database Setup

MongoDB is the key dependency decision for `TASK_FILE_NAME`.

### A. What is MongoDB?

MongoDB ek document database hai. Data tables/rows ke bajay JSON-like documents me store hota hai. Product jaise object me variants, images, attributes, rating summary, aur timestamps naturally ek document ke andar fit ho sakte hain.

### B. Why this service uses MongoDB

Product catalog ka shape fixed nahi hota:

| Category | Example dynamic attributes |
|---|---|
| Shoes | size, color, material, sport type |
| Mobiles | RAM, storage, processor, battery |
| Furniture | material, dimensions, finish |
| Grocery | weight, pack size, expiry rules |

SQL me ye model many nullable columns, JSON columns, ya EAV tables me complex ho sakta hai. MongoDB me category-wise flexible attributes cleaner rahte hain.

### C. Required or optional?

| Scenario | MongoDB required? |
|---|---:|
| Reading `TASK_FILE_NAME` only | No |
| Running unit tests that do not touch real DB | No |
| Running Mongo-backed repositories | Yes |
| Running migrations | Yes |
| Running future API server with real catalog data | Yes |

### D. Local installation

Already documented.

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
5. Database Analysis: MongoDB
```

### E. Docker setup

Already documented.

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
5. Database Analysis: MongoDB -> H. Docker setup
5. Database Analysis: MongoDB -> I. Docker Compose example
9. Docker and DevOps Setup
```

### F. Start commands

Use previous dependency file commands. No new container or changed port is introduced by `TASK_FILE_NAME`.

### G. Verify running

Use previous dependency file MongoDB ping and collection checks:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
5. Database Analysis: MongoDB -> J. Verify MongoDB running
10. Migration Setup -> Verify migrations
```

### H. Default port

| Service | Port | Purpose | Status |
|---|---:|---|---|
| MongoDB | 27017 | Product catalog database | Reused |

### I. Connection string format

Current implementation uses Product Service prefixed variables:

```env
PRODUCT_MONGO_URI=mongodb://localhost:27017
PRODUCT_MONGO_DATABASE=product_db
```

Important:

`TASK_FILE_NAME` contains future-looking examples like `MONGO_URI`. Current Go config does not read `MONGO_URI`; it reads `PRODUCT_MONGO_URI`.

### J. Where credentials should go

Already documented.

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
5. Database Analysis: MongoDB -> K. Where credentials go
```

Task-specific warning:

Do not put MongoDB username/password inside Go files, markdown examples for real production use, or committed config. Use local `.env` for development and secret manager/CI/Kubernetes secrets for real environments.

## 6. Redis / Queue / External Services

No new external service is introduced only by `TASK_FILE_NAME`.

### MongoDB

MongoDB is mandatory for real Product Service catalog runtime. Full setup is reused from previous dependency documentation.

### RabbitMQ

RabbitMQ appears in current repository code because later event/outbox work exists. It is not introduced by the MongoDB decision task.

Use this reference instead of repeating setup:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
7. External Services Analysis -> RabbitMQ
```

RabbitMQ is required only when:

```env
PRODUCT_EVENTS_ENABLED=true
PRODUCT_OUTBOX_WORKER_ENABLED=true
PRODUCT_EVENT_BROKER=rabbitmq
```

### Kafka

Current config validation allows:

```env
PRODUCT_EVENT_BROKER=kafka
```

But the app wiring needs a custom `ProductEventPublisher` dependency for Kafka. Beginner recommendation: use RabbitMQ locally until Kafka publisher code exists.

### Redis

Redis is not used by current Product Service code. Agar Redis project ke kisi aur service ke liye chal raha hai, `TASK_FILE_NAME` ke setup me uski zarurat nahi hai.

### Typesense

Typesense is not a direct dependency of this service for `TASK_FILE_NAME`. Search Service owns search indexing/search ranking. MongoDB remains canonical product data source.

### Ports and networking

Detailed networking troubleshooting is already explained in:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
8. Ports and Networking
```

Task-specific port summary:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Product Service API | N/A | HTTP/gRPC API listener | Not available yet because no server entrypoint exists |
| MongoDB | 27017 | Product catalog database | Reused |
| RabbitMQ AMQP | 5672 | Product event publishing when RabbitMQ is enabled | Reused, conditional |
| RabbitMQ Management UI | 15672 | Local browser debugging for queues | Reused, optional |
| Typesense | N/A for this service | Search Service dependency, not Product Service runtime | Not required here |

No new port is introduced by `TASK_FILE_NAME`. If port conflict hota hai, previous dependency file ke `lsof`, `docker ps`, and alternate host-port examples follow karo.

## 7. Environment Variables

Complete `.env` setup is already explained here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
6. Environment Variables
```

### Task-specific MongoDB variables

Only these variables are directly connected to the MongoDB choice:

```env
PRODUCT_MONGO_URI=mongodb://localhost:27017
PRODUCT_MONGO_DATABASE=product_db
PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=false
```

| Variable | Required? | Purpose | Security note |
|---|---:|---|---|
| `PRODUCT_MONGO_URI` | Yes for DB runtime | MongoDB server connection string | Secret if username/password is included |
| `PRODUCT_MONGO_DATABASE` | Yes for DB runtime | Database name used by Product Service | Not secret |
| `PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS` | Optional | Allows app-level collection creation when manager is wired | Keep `false` in production; prefer migrations |

### Where `.env` should be created

Already documented, but the important path is:

```text
backend/services/product-service/.env
```

Beginner note:

The Go config uses `os.LookupEnv`. Iska matlab `.env` file automatically load nahi hoti. Shell me variables export/load karne padenge, ya future server entrypoint me dotenv loader add karna padega.

### Current naming mismatch to avoid

`TASK_FILE_NAME` mentions future examples such as:

```env
MONGO_URI=...
MONGO_DATABASE=...
SERVICE_PORT=...
GRPC_PORT=...
```

Current implementation expects:

```env
PRODUCT_MONGO_URI=...
PRODUCT_MONGO_DATABASE=...
```

No API/gRPC port variable is currently available because there is no runnable server entrypoint yet.

## 8. Docker Setup

No Dockerfile or docker-compose file is introduced by `TASK_FILE_NAME`.

Existing Docker guidance is reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
5. Database Analysis: MongoDB -> H. Docker setup
5. Database Analysis: MongoDB -> I. Docker Compose example
9. Docker and DevOps Setup
```

### Docker status table

| Artifact | Exists now? | Task-specific note |
|---|---:|---|
| Product Service Dockerfile | No | Still pending |
| Product Service compose file | No | Still pending |
| MongoDB Docker run guidance | Yes, in previous doc | Reuse it |
| RabbitMQ Docker run guidance | Yes, in previous doc | Conditional, reuse it |
| Persistent volume guidance | Yes, in previous doc | Reuse it |

## 9. Local Development Setup

This flow avoids duplicate setup and keeps the beginner path clean.

### Step 1: Read previous dependency documentation

Start here:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Read these sections first:

- `4. Language-Specific Dependency System: Go`
- `5. Database Analysis: MongoDB`
- `6. Environment Variables`
- `9. Docker and DevOps Setup`
- `10. Migration Setup`
- `11. Complete Project Run Instructions`

### Step 2: Go to project directory

```bash
cd backend/services/product-service
```

### Step 3: Install only new dependencies if any

No new dependency is introduced by `TASK_FILE_NAME`.

If dependencies are not downloaded yet, use the previous dependency file command:

```bash
go mod download
```

### Step 4: Setup only new databases/services if any

No new database beyond MongoDB is introduced. MongoDB setup is reused from the previous dependency file.

### Step 5: Add only new or changed environment variables

No new variable is added only by `TASK_FILE_NAME`.

Make sure the MongoDB variables are present:

```env
PRODUCT_MONGO_URI=mongodb://localhost:27017
PRODUCT_MONGO_DATABASE=product_db
PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=false
```

### Step 6: Run migrations if needed

Migration commands are already documented:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
10. Migration Setup
```

Important:

The migration scripts currently use `product_db`. If you change `PRODUCT_MONGO_DATABASE`, migration execution strategy also needs review.

### Step 7: Start backend service

Current state:

There is no `cmd/server/main.go` or equivalent runnable entrypoint. So a beginner should not expect this to work:

```bash
go run .
```

Use these checks instead:

```bash
go test ./...
go build ./...
```

### Step 8: Verify task-specific functionality

For `TASK_FILE_NAME`, verification means:

- MongoDB is documented as primary product catalog database.
- Current env variables use `PRODUCT_MONGO_URI` and `PRODUCT_MONGO_DATABASE`.
- MongoDB collections/migrations are available for real DB usage.
- Search ownership remains outside this service.

For real local DB verification, follow previous migration checks.

## 10. Running the Project

Full run instructions are reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
11. Complete Project Run Instructions
```

### Task-specific run reality

| Action | Command/status |
|---|---|
| Download dependencies | Use previous doc |
| Start MongoDB | Use previous doc |
| Create/load `.env` | Use previous doc |
| Run MongoDB migrations | Use previous doc |
| Run tests | `go test ./...` |
| Build packages | `go build ./...` |
| Start API server | Not available yet |
| Verify HTTP/gRPC API | Not available yet |

## 11. Common Errors & Fixes

Generic setup errors are already documented:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
12. Common Errors and Fixes
```

### New or task-specific confusion points

| Error / confusion | Cause | Fix | Prevention |
|---|---|---|---|
| `MONGO_URI` is set but code still uses default Mongo URI | `TASK_FILE_NAME` has future-style examples, current config reads `PRODUCT_MONGO_URI` | Rename env var to `PRODUCT_MONGO_URI` | Copy from `.env.example` |
| `MONGO_DATABASE` is set but app uses `product_db` | Current config reads `PRODUCT_MONGO_DATABASE` | Use `PRODUCT_MONGO_DATABASE=product_db` | Follow current code, not old future examples |
| Migrations run on `product_db` even after changing env DB name | Migration scripts are written for `product_db` | Keep DB name as `product_db` locally or update migration strategy | Avoid changing DB name casually |
| Developer expects Typesense setup for this task | `TASK_FILE_NAME` discusses search boundary, but Typesense belongs to Search Service | Do not setup Typesense for this service task | Keep MongoDB as source of truth and Search Service as index owner |
| `go run .` fails | No server entrypoint exists yet | Use `go test ./...` and `go build ./...` | Wait for/add `cmd/server/main.go` in later work |
| `PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=true` fails | Collection schema manager must be wired into app dependencies | Use manual migrations for local setup | Keep auto-create false unless app wiring supports it |

## 12. Security & Best Practices

Generic security notes are reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
13. Security and Configuration Audit
14. Best Practices
```

### Task-specific best practices for MongoDB choice

- Use app-specific MongoDB user in production, not root user.
- Keep MongoDB credentials out of committed files.
- Use `PRODUCT_MONGO_URI` from environment/secret manager.
- Keep `product_db` ownership restricted to Product Service.
- Do not let Cart, Order, Search, or CMS connect directly to Product Service MongoDB.
- Validate dynamic attributes at service layer because MongoDB is flexible.
- Keep product documents bounded; do not store image binaries in MongoDB.
- Use indexes for seller listing, category listing, SKU lookup, slug lookup, and published product reads.
- Keep Typesense/Search Service responsible for advanced search, typo tolerance, facets, and ranking.
- Use migrations for reviewed schema/index changes instead of relying on auto-create in production.
- Enable backups before production catalog data goes live.

## 13. Missing or Misconfigured Things

| Area | Current observation | Impact | Suggested fix |
|---|---|---|---|
| Server entrypoint | No `cmd/server/main.go` | Cannot start real service process | Add server entrypoint with config load, dependency wiring, health checks |
| API port config | No current API/gRPC port env var | No runtime network listener to configure | Add port env vars when server exists |
| Dockerfile | Not present | App container cannot be built directly | Add Dockerfile after entrypoint exists |
| docker-compose | Not present | Local infra setup is manual | Add local compose for MongoDB/RabbitMQ/app later |
| Migration DB name | Scripts target `product_db` | Env DB mismatch can confuse developers | Keep documented default or parameterize migrations |
| Env naming | Task doc future examples differ from current code | Beginners may set wrong variables | Use `PRODUCT_MONGO_*` variables from `.env.example` |
| Kafka option | Config allows Kafka but no built-in Kafka publisher is wired | Runtime confusion | Keep local broker as RabbitMQ or implement Kafka publisher |
| Health checks | No app health endpoint yet | Hard to verify runtime readiness | Add `/healthz` or gRPC health service later |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Project Tech Stack Analysis` | Same Go/MongoDB/RabbitMQ technology explanation already exists |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Language-Specific Dependency System: Go` | Same Go module setup applies |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Analysis: MongoDB` | MongoDB install, Docker, connection strings, and verification already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Environment Variables` | Complete `.env` explanation already exists |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. External Services Analysis` | RabbitMQ/Kafka/Redis status already explained |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Ports and Networking` | Same MongoDB/RabbitMQ ports apply |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Docker and DevOps Setup` | Same local infra commands and Docker limitations apply |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `10. Migration Setup` | Same MongoDB migration files and order apply |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `11. Complete Project Run Instructions` | Same onboarding flow applies |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `12. Common Errors and Fixes` | Generic setup troubleshooting already exists |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `13. Security and Configuration Audit` | Most security findings are unchanged |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `14. Best Practices` | General backend/DevOps practices are unchanged |

## 15. Final Checklist

- [ ] Previous dependency documentation checked first.
- [ ] `TASK_FILE_NAME` confirmed as MongoDB decision guide.
- [ ] No duplicate Go installation instructions added.
- [ ] No duplicate Docker setup added.
- [ ] No duplicate full `.env` table added.
- [ ] MongoDB identified as primary catalog database.
- [ ] `PRODUCT_MONGO_URI` and `PRODUCT_MONGO_DATABASE` naming confirmed.
- [ ] `MONGO_URI` future example mismatch understood.
- [ ] `product_db` default database confirmed.
- [ ] MongoDB migration reuse documented.
- [ ] RabbitMQ marked conditional and reused from previous doc.
- [ ] Redis marked not required.
- [ ] Typesense marked outside this service runtime.
- [ ] Current no-server-entrypoint limitation documented.
- [ ] Security notes for MongoDB ownership and credentials reviewed.
- [ ] No business logic rewritten.
- [ ] Original `TASK_FILE_NAME` not modified.
- [ ] This dependency output saved as `OUTPUT_FILE_PATH`.
