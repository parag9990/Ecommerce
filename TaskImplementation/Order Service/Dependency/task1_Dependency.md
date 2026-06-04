# Dependency and Setup Guide for `${SERVICE_NAME}`

## Variables used in this document

```text
SERVICE_NAME=Order Service
TASK_FILE_NAME=task1.md
OUTPUT_FILE_NAME=task1_Dependency.md
TASK_PATH=TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}
OUTPUT_PATH=TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}
SERVICE_CODE_PATH=backend/services/order-service
```

Use `${TASK_PATH}` as the source implementation guide and `${OUTPUT_PATH}` as this generated dependency/setup guide.

## What this file covers

`${TASK_FILE_NAME}` is a documentation-only lifecycle task. Directly, it needs only Markdown, Mermaid diagrams, and a developer/editor that can read `.md` files.

But the current repository also contains real backend code under `${SERVICE_CODE_PATH}`. So this guide explains two levels:

- Task-level setup: kya chahiye to read, review, and maintain the lifecycle guide.
- Runtime-level setup: kya chahiye to run/test the current backend module once service bootstrap is available.

Important note: `${SERVICE_CODE_PATH}` currently does not contain `cmd/server/main.go`. That means package tests can run, migrations can be applied, and service code can be built as packages, but a real long-running backend process needs a server entrypoint that wires config, MySQL, Kafka, repositories, usecases, and the gRPC listener.

## 1. Project Tech Stack Analysis

| Technology | Required? | Why used | Beginner explanation |
|---|---:|---|---|
| Markdown | Required for `${TASK_FILE_NAME}` | Task output is a `.md` guide | Markdown simple text format hai. Isme headings, tables, code blocks easily likh sakte ho. |
| Mermaid | Optional but used in guide | Lifecycle/state diagrams | Mermaid text se diagrams banata hai. Image draw karne ki need nahi hoti. |
| Shields.io badges | Optional | Visual task metadata | Badges docs ko readable banate hain, but project run karne ke liye required nahi hain. |
| Go | Required for backend module | Service code is Go | Go ek compiled backend language hai. Fast APIs, gRPC services, and workers banane ke liye use hota hai. |
| Go modules | Required | Dependency management | `go.mod` batata hai project ko kaunse packages chahiye. `go.sum` versions verify karta hai. |
| gRPC | Required for backend API layer | Internal service API | gRPC high-performance RPC framework hai. Services ek dusre ko typed proto contracts se call karte hain. |
| Protocol Buffers | Required for gRPC contracts | API message definitions | Proto files API request/response ka contract define karte hain. Generated Go code use hota hai. |
| MySQL 8+ | Required for runtime persistence | Orders, items, status history, idempotency, outbox | MySQL relational database hai jisme data tables ke form me store hota hai. |
| Kafka | Required when events enabled | Order lifecycle events/outbox publishing | Kafka event streaming system hai. Service `order.events` topic par events publish karta hai. |
| `github.com/go-sql-driver/mysql` | Required for MySQL access | Go MySQL driver | Go code ko MySQL se connect karne ke liye driver chahiye. |
| `github.com/segmentio/kafka-go` | Required when Kafka publishing is enabled | Kafka producer | Go code se Kafka topic me messages write karne ke liye library. |
| `google.golang.org/grpc` | Required | gRPC server/runtime | Go me gRPC server create karne ke liye official package. |
| `google.golang.org/protobuf` | Required | Proto generated message support | Proto messages serialize/deserialize karne ke liye. |
| `github.com/DATA-DOG/go-sqlmock` | Test dependency | Repository unit tests | Real DB ke bina SQL behavior test karne ke liye mock library. |
| Docker | Optional but recommended for local infra | MySQL/Kafka local setup | Docker se DB/Kafka containers easily run ho jate hain. Beginner ke liye easiest local setup. |

## 2. Go Dependency System

`${SERVICE_CODE_PATH}` is a Go module.

### Important files

| File | Meaning |
|---|---|
| `${SERVICE_CODE_PATH}/go.mod` | Module name, Go version, and direct dependencies |
| `${SERVICE_CODE_PATH}/go.sum` | Dependency checksums for repeatable installs |
| `backend/go.work` | Go workspace linking shared generated code and services |
| `backend/shared/gen/go/go.mod` | Local shared generated proto module used through `replace` |

### Go version

The module declares:

```text
go 1.26.3
```

Install the same Go version when possible. Version mismatch se build/test errors aa sakte hain.

### Dependency commands

From repository root:

```bash
cd backend/services/order-service
go mod download
go mod tidy
go test ./...
```

From `backend/` workspace:

```bash
cd backend
go test ./services/order-service/...
```

### What each command does

| Command | Use |
|---|---|
| `go mod download` | Dependencies cache/download karta hai |
| `go mod tidy` | Missing/unused dependencies clean karta hai |
| `go build ./...` | Packages compile karke check karta hai |
| `go test ./...` | Unit tests run karta hai |
| `go run ./cmd/server` | Service start karega only after `cmd/server/main.go` exists |

### Common Go module issues

| Problem | Cause | Fix |
|---|---|---|
| `go: module requires go 1.26.3` | Local Go old hai | Required Go version install karo |
| `missing go.sum entry` | Dependency checksum missing | `go mod tidy` run karo |
| `cannot find module .../shared/gen/go` | Workspace/replace path issue | Repo root intact rakho, `backend/shared/gen/go` delete mat karo |
| Proxy/download fail | Network/proxy issue | `go env GOPROXY`, internet, VPN/proxy check karo |
| Generated proto package missing | Shared generated files absent | `backend/shared/gen/go` files verify karo |

## 3. Database Analysis

### Database used: MySQL 8+

#### A. What it is

MySQL ek relational database hai. Data tables, rows, columns, indexes, foreign keys ke form me store hota hai.

#### B. Why used here

Current backend module uses MySQL for:

- `orders`
- `order_items`
- `order_status_history`
- `shipments`
- `order_idempotency_keys`
- payment coordination fields
- `order_outbox_events`

Lifecycle task ke concept ko runtime me persist karne ke liye MySQL mandatory hai.

#### C. Required or optional

- For `${TASK_FILE_NAME}` documentation-only review: optional.
- For running backend repositories/usecases with real persistence: required.

#### D. Local installation

Windows:

```text
1. Install MySQL Installer from official MySQL website.
2. Choose MySQL Server 8+.
3. Set root password.
4. Start MySQL service from Services app or MySQL Notifier.
```

Linux Ubuntu/Debian:

```bash
sudo apt update
sudo apt install mysql-server mysql-client
sudo systemctl enable mysql
sudo systemctl start mysql
sudo mysql_secure_installation
```

macOS with Homebrew:

```bash
brew install mysql
brew services start mysql
```

#### E. Docker setup

Simple `docker run`:

```bash
docker run -d \
  --name order-mysql \
  -e MYSQL_ROOT_PASSWORD=rootpassword \
  -e MYSQL_DATABASE=order_db \
  -p 3306:3306 \
  -v order_mysql_data:/var/lib/mysql \
  mysql:8
```

Docker Compose example:

```yaml
services:
  mysql:
    image: mysql:8
    container_name: order-mysql
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: order_db
      MYSQL_USER: order_user
      MYSQL_PASSWORD: order_password
    ports:
      - "3306:3306"
    volumes:
      - order_mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-uroot", "-prootpassword"]
      interval: 10s
      timeout: 5s
      retries: 10

volumes:
  order_mysql_data:
```

#### F. Start commands

System service:

```bash
sudo systemctl start mysql
```

Docker:

```bash
docker start order-mysql
```

Compose:

```bash
docker compose up -d mysql
```

#### G. Verify running

```bash
mysql -u root -p -h 127.0.0.1 -P 3306
```

Inside MySQL:

```sql
SHOW DATABASES;
USE order_db;
SHOW TABLES;
```

#### H. Default port

```text
3306
```

#### I. Connection string format

The Go code reads `ORDER_MYSQL_DSN`.

```env
ORDER_MYSQL_DSN=order_user:order_password@tcp(localhost:3306)/order_db?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci
```

`parseTime=true` important hai because Go ko MySQL `TIMESTAMP` values `time.Time` me parse karni hoti hain.

#### J. Where to place credentials

Recommended local file:

```text
${SERVICE_CODE_PATH}/.env
```

But current Go code uses `os.Getenv` directly and does not load `.env` automatically. So either:

- export variables in terminal, or
- use a shell command that loads `.env`, or
- future server entrypoint me `godotenv`/config loader add karo.

Example manual export:

```bash
export ORDER_MYSQL_DSN='order_user:order_password@tcp(localhost:3306)/order_db?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci'
```

### Migrations

Migration files are present:

```text
${SERVICE_CODE_PATH}/migrations/001_create_order_tables.up.sql
${SERVICE_CODE_PATH}/migrations/002_add_payment_coordination.up.sql
${SERVICE_CODE_PATH}/migrations/003_add_grpc_query_indexes.up.sql
${SERVICE_CODE_PATH}/migrations/004_create_order_outbox_events.up.sql
```

Apply in order:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/002_add_payment_coordination.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/003_add_grpc_query_indexes.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/004_create_order_outbox_events.up.sql
```

Verify:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW TABLES;"
```

## 4. Environment Variables

### Complete `.env` example

Create this file for local reference:

```text
${SERVICE_CODE_PATH}/.env
```

Example:

```env
# Database
ORDER_MYSQL_DSN=order_user:order_password@tcp(localhost:3306)/order_db?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci

# Checkout/idempotency timings
ORDER_INVENTORY_RESERVATION_TTL=15m
ORDER_INVENTORY_RELEASE_TIMEOUT=3s
ORDER_IDEMPOTENCY_TTL=24h

# Payment coordination
ORDER_PAYMENT_RETURN_URL=https://localhost.example/checkout/result
ORDER_PAYMENT_ALLOWED_CURRENCIES=INR,USD
ORDER_PAYMENT_INVENTORY_ACTION_TIMEOUT=3s

# gRPC server
ORDER_GRPC_ADDR=:9094
ORDER_GRPC_TRUSTED_CALLER_TOKEN=change-this-local-token-minimum-32-chars
ORDER_PAGE_TOKEN_SIGNING_KEY=change-this-page-token-key-minimum-32

# Events/outbox
ORDER_EVENTS_ENABLED=true
ORDER_EVENTS_TOPIC=order.events
ORDER_KAFKA_BROKERS=localhost:9092
ORDER_KAFKA_WRITE_TIMEOUT=10s
ORDER_OUTBOX_BATCH_SIZE=100
ORDER_OUTBOX_INTERVAL=1s
ORDER_OUTBOX_MAX_ATTEMPTS=8
ORDER_OUTBOX_INITIAL_BACKOFF=5s
ORDER_OUTBOX_MAX_BACKOFF=10m
ORDER_OUTBOX_STALE_LOCK_TIMEOUT=5m
```

### Variable details

| Variable | Required? | Example | Purpose | Security notes |
|---|---:|---|---|---|
| `ORDER_MYSQL_DSN` | Yes | `order_user:order_password@tcp(localhost:3306)/order_db?...` | MySQL connection string | Password secret hai. Commit mat karo. |
| `ORDER_INVENTORY_RESERVATION_TTL` | Optional | `15m` | Inventory reservation kitni der valid rahegi | Positive duration hona chahiye. |
| `ORDER_INVENTORY_RELEASE_TIMEOUT` | Optional | `3s` | Inventory release call timeout | Too high value requests ko hang kar sakti hai. |
| `ORDER_IDEMPOTENCY_TTL` | Optional | `24h` | Duplicate checkout key expiry | Positive duration required. |
| `ORDER_PAYMENT_RETURN_URL` | Yes | `https://shop.example.com/checkout/result` | Payment complete ke baad frontend URL | Code HTTPS absolute URL enforce karta hai. |
| `ORDER_PAYMENT_ALLOWED_CURRENCIES` | Optional | `INR,USD` | Allowed payment currencies | ISO 3-letter uppercase values use karo. |
| `ORDER_PAYMENT_INVENTORY_ACTION_TIMEOUT` | Optional | `3s` | Payment result ke saath inventory action timeout | Positive duration required. |
| `ORDER_GRPC_ADDR` | Optional | `:9094` | gRPC listen address | Public expose karte time firewall/TLS socho. |
| `ORDER_GRPC_TRUSTED_CALLER_TOKEN` | Yes | `32+ chars` | Internal caller authentication token | Minimum 32 chars. Production me secret manager use karo. |
| `ORDER_PAGE_TOKEN_SIGNING_KEY` | Yes | `32+ chars` | Pagination token signing key | Rotate carefully; old page tokens invalid ho sakte hain. |
| `ORDER_EVENTS_ENABLED` | Optional | `true` | Kafka/outbox publishing enable/disable | Default true hai, so Kafka brokers required honge. |
| `ORDER_EVENTS_TOPIC` | Required if events enabled | `order.events` | Kafka topic name | Topic naming consistent rakho. |
| `ORDER_KAFKA_BROKERS` | Required if events enabled | `localhost:9092` | Kafka broker addresses | Production credentials/TLS separate configure honge. |
| `ORDER_KAFKA_WRITE_TIMEOUT` | Optional | `10s` | Kafka write timeout | Positive duration required. |
| `ORDER_OUTBOX_BATCH_SIZE` | Optional | `100` | Worker ek batch me kitne events publish kare | Too high DB/Kafka load badha sakta hai. |
| `ORDER_OUTBOX_INTERVAL` | Optional | `1s` | Worker polling interval | Too low value DB polling load badha sakti hai. |
| `ORDER_OUTBOX_MAX_ATTEMPTS` | Optional | `8` | Fail event retry count | After max attempts event dead letter ho jata hai. |
| `ORDER_OUTBOX_INITIAL_BACKOFF` | Optional | `5s` | First retry delay | Positive duration required. |
| `ORDER_OUTBOX_MAX_BACKOFF` | Optional | `10m` | Max retry delay | Initial backoff se kam nahi hona chahiye. |
| `ORDER_OUTBOX_STALE_LOCK_TIMEOUT` | Optional | `5m` | Stuck event lock recovery | Worker crash recovery ke liye useful. |

### How environment loading works

Current config code uses:

```go
os.Getenv("VARIABLE_NAME")
```

Matlab `.env` file automatically load nahi hota. Beginner-friendly options:

Option 1: export manually:

```bash
export ORDER_EVENTS_ENABLED=false
export ORDER_MYSQL_DSN='order_user:order_password@tcp(localhost:3306)/order_db?parseTime=true'
```

Option 2: load `.env` in shell:

```bash
set -a
. backend/services/order-service/.env
set +a
```

Option 3: future server bootstrap me `.env` loader add karo.

### Common `.env` mistakes

| Mistake | Result | Fix |
|---|---|---|
| `.env` created but not loaded | Config still missing | Export variables or source file |
| Short trusted token | Config validation fails | 32+ chars token use karo |
| HTTP return URL | Validation fails | `https://...` use karo |
| Missing Kafka brokers with events enabled | Startup config fails | Set `ORDER_KAFKA_BROKERS` or `ORDER_EVENTS_ENABLED=false` |
| Bad duration like `15 minutes` | Validation fails | Go duration format use karo: `15m`, `3s`, `500ms` |

## 5. External Services Analysis

### MySQL

- What: relational DB.
- Why: orders, item snapshots, status history, idempotency, outbox.
- Mandatory: yes for real runtime.
- Health check:

```bash
mysqladmin ping -h 127.0.0.1 -P 3306 -u root -p
```

### Kafka

#### What it is

Kafka event streaming platform hai. Producer topic me event publish karta hai, consumers baad me read karte hain.

#### Why used here

Current code has outbox publishing for order lifecycle events such as:

- `OrderCreated`
- `OrderPaid`
- `OrderCancelled`
- `OrderDelivered`

Default topic:

```text
order.events
```

#### Mandatory or optional

- Required if `ORDER_EVENTS_ENABLED=true`.
- Optional for local tests if you set `ORDER_EVENTS_ENABLED=false`.

#### Docker setup

Simple Kafka with KRaft mode:

```yaml
services:
  kafka:
    image: bitnami/kafka:latest
    container_name: order-kafka
    ports:
      - "9092:9092"
    environment:
      KAFKA_CFG_NODE_ID: 1
      KAFKA_CFG_PROCESS_ROLES: broker,controller
      KAFKA_CFG_CONTROLLER_QUORUM_VOTERS: 1@kafka:9093
      KAFKA_CFG_LISTENERS: PLAINTEXT://:9092,CONTROLLER://:9093
      KAFKA_CFG_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_CFG_CONTROLLER_LISTENER_NAMES: CONTROLLER
      KAFKA_CFG_AUTO_CREATE_TOPICS_ENABLE: "true"
      ALLOW_PLAINTEXT_LISTENER: "yes"
    volumes:
      - order_kafka_data:/bitnami/kafka

volumes:
  order_kafka_data:
```

Start:

```bash
docker compose up -d kafka
```

Create/verify topic:

```bash
docker exec -it order-kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --if-not-exists --topic order.events --partitions 3 --replication-factor 1
docker exec -it order-kafka kafka-topics.sh --bootstrap-server localhost:9092 --list
```

#### Credentials placement

Local plaintext example:

```env
ORDER_KAFKA_BROKERS=localhost:9092
ORDER_EVENTS_TOPIC=order.events
```

Production should use TLS/SASL/secret manager. Current config does not yet expose Kafka username/password/TLS env vars, so that is a future hardening item.

### Redis

Redis is mentioned in broader architecture docs for other services, but this backend module does not currently import or configure Redis.

- Mandatory for `${TASK_FILE_NAME}`: no.
- Mandatory for current `${SERVICE_CODE_PATH}`: no.

### RabbitMQ/NATS

Not used in current backend module. Kafka is the implemented event client.

### Payment provider

No provider SDK is imported in this module. Payment Service is expected to be a separate integration boundary. This module needs `ORDER_PAYMENT_RETURN_URL` and payment/inventory timeout config.

### Docker

Docker is recommended for MySQL and Kafka local setup. Repo currently has no Dockerfile or docker-compose file checked in for this backend module.

### Kubernetes

Broader docs mention Kubernetes deployment shape, but no Kubernetes manifests are present for this backend module.

## 6. Ports and Networking

| Service | Port | Required? | Purpose |
|---|---:|---:|---|
| gRPC API | 9094 | Runtime yes | Internal RPC API from `ORDER_GRPC_ADDR=:9094` |
| MySQL | 3306 | Runtime yes | Order database |
| Kafka | 9092 | If events enabled | Event broker |
| Kafka controller | 9093 | Docker internal | KRaft controller listener |
| HTTP API | N/A | Not present | No HTTP server entrypoint in current module |
| Redis | 6379 | No | Not used by current module |

### Port conflicts

Check port:

```bash
ss -ltnp | grep ':3306'
ss -ltnp | grep ':9092'
ss -ltnp | grep ':9094'
```

Fix options:

- Stop old process/container.
- Change host port in Docker Compose.
- Change `ORDER_GRPC_ADDR`, for example `:19094`.

### Docker networking notes

Inside Docker Compose, service-to-service hostnames are service names:

```env
ORDER_MYSQL_DSN=order_user:order_password@tcp(mysql:3306)/order_db?parseTime=true
ORDER_KAFKA_BROKERS=kafka:9092
```

From host machine, use:

```env
ORDER_MYSQL_DSN=order_user:order_password@tcp(localhost:3306)/order_db?parseTime=true
ORDER_KAFKA_BROKERS=localhost:9092
```

## 7. Docker and DevOps Setup

### Current state

No checked-in Dockerfile or compose file exists for `${SERVICE_CODE_PATH}` right now.

Recommended beginner approach:

1. Run MySQL and Kafka with Docker.
2. Run Go tests locally.
3. Add service entrypoint later.
4. Add Dockerfile after entrypoint exists.

### Full local compose example for dependencies

```yaml
services:
  mysql:
    image: mysql:8
    container_name: order-mysql
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: order_db
      MYSQL_USER: order_user
      MYSQL_PASSWORD: order_password
    ports:
      - "3306:3306"
    volumes:
      - order_mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-uroot", "-prootpassword"]
      interval: 10s
      timeout: 5s
      retries: 10

  kafka:
    image: bitnami/kafka:latest
    container_name: order-kafka
    ports:
      - "9092:9092"
    environment:
      KAFKA_CFG_NODE_ID: 1
      KAFKA_CFG_PROCESS_ROLES: broker,controller
      KAFKA_CFG_CONTROLLER_QUORUM_VOTERS: 1@kafka:9093
      KAFKA_CFG_LISTENERS: PLAINTEXT://:9092,CONTROLLER://:9093
      KAFKA_CFG_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_CFG_CONTROLLER_LISTENER_NAMES: CONTROLLER
      KAFKA_CFG_AUTO_CREATE_TOPICS_ENABLE: "true"
      ALLOW_PLAINTEXT_LISTENER: "yes"
    volumes:
      - order_kafka_data:/bitnami/kafka

volumes:
  order_mysql_data:
  order_kafka_data:
```

Common Docker commands:

```bash
docker compose up -d
docker compose down
docker compose logs -f
docker ps
docker volume ls
```

### Future Dockerfile shape

This will work only after a server entrypoint exists:

```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY backend/go.work backend/go.work.sum ./backend/
COPY backend/shared ./backend/shared
COPY backend/services/order-service ./backend/services/order-service
WORKDIR /app/backend/services/order-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/order-service ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /out/order-service /order-service
EXPOSE 9094
ENTRYPOINT ["/order-service"]
```

### Volumes

Use persistent volumes for:

- MySQL data: `/var/lib/mysql`
- Kafka data: `/bitnami/kafka`

Without volumes, container delete karne par data lose ho sakta hai.

### Restart policies

For local dev:

```yaml
restart: unless-stopped
```

For production, prefer orchestrator-managed restart policies, health checks, and alerts.

## 8. Project Run Instructions

### Step 1: Clone repository

```bash
git clone <repository-url>
cd Ecommerce
```

### Step 2: Go to backend module

```bash
cd backend/services/order-service
```

### Step 3: Install/download dependencies

```bash
go mod download
go mod tidy
```

### Step 4: Start MySQL

```bash
docker run -d \
  --name order-mysql \
  -e MYSQL_ROOT_PASSWORD=rootpassword \
  -e MYSQL_DATABASE=order_db \
  -e MYSQL_USER=order_user \
  -e MYSQL_PASSWORD=order_password \
  -p 3306:3306 \
  -v order_mysql_data:/var/lib/mysql \
  mysql:8
```

### Step 5: Run migrations

From repository root:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/002_add_payment_coordination.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/003_add_grpc_query_indexes.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/004_create_order_outbox_events.up.sql
```

### Step 6: Start Kafka if events are enabled

Use Docker Compose example from section 7, then:

```bash
docker exec -it order-kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --if-not-exists --topic order.events --partitions 3 --replication-factor 1
```

For quick local config without Kafka:

```bash
export ORDER_EVENTS_ENABLED=false
```

### Step 7: Create/load environment variables

```bash
cd /home/parag/Ecommerce
set -a
. backend/services/order-service/.env
set +a
```

Or export variables manually.

### Step 8: Run tests

```bash
cd backend/services/order-service
go test ./...
```

### Step 9: Build packages

```bash
go build ./...
```

### Step 10: Start backend service

Current state:

```text
No cmd/server/main.go exists, so there is no direct service start command yet.
```

Expected future command after entrypoint is added:

```bash
go run ./cmd/server
```

Expected future verification:

```bash
grpcurl -plaintext localhost:9094 list
```

## 9. Common Errors and Fixes

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `ORDER_MYSQL_DSN is required` | Env var missing | Export `ORDER_MYSQL_DSN` | Keep `.env.example` updated |
| `connection refused 127.0.0.1:3306` | MySQL not running or wrong port | Start MySQL/container | Add health check before service start |
| `Unknown database 'order_db'` | Migration 001 not run | Run `001_create_order_tables.up.sql` | Apply migrations in order |
| `Table ... doesn't exist` | Later migration missing | Run all `001` to `004` migrations | Use migration tool in future |
| `ORDER_PAYMENT_RETURN_URL must be an absolute HTTPS URL` | URL missing or HTTP | Use `https://...` | Document env values |
| `trusted caller token must be at least 32 characters` | Secret too short | Generate longer token | Use password manager/secret manager |
| `ORDER_KAFKA_BROKERS is required when order events are enabled` | Events default true | Set brokers or disable events locally | Add local `.env` |
| Kafka publish timeout | Kafka down/wrong advertised listener | Check `ORDER_KAFKA_BROKERS`, logs | Use compose health checks |
| Port already in use | Another DB/Kafka/gRPC process running | Change port or stop process | Reserve standard dev ports |
| `go mod download` failed | Network/proxy issue | Check internet/proxy/GOPROXY | Cache dependencies in CI |
| `missing go.sum entry` | Module metadata stale | Run `go mod tidy` | Run tidy before commit |
| `go: cannot find main module` | Running command in wrong folder | `cd backend/services/order-service` or `cd backend` | Use documented paths |
| Docker daemon not running | Docker service stopped | Start Docker Desktop/system service | Enable Docker on startup |
| Permission denied on MySQL | Wrong user/password | Check DSN and MySQL grants | Create dedicated DB user |
| Migration failed on enum/index | Migration order wrong or already partially applied | Inspect schema, apply from clean DB | Use migration tool with version table |

## 10. Security and Configuration Audit

| Finding | Risk | Recommendation |
|---|---|---|
| No `${SERVICE_CODE_PATH}/.env.example` file | Beginners do not know required vars | Add `.env.example` with safe placeholders |
| `.env` not auto-loaded | Local startup confusion | Either document shell loading or add loader in `cmd/server` |
| `ORDER_EVENTS_ENABLED` defaults to true | Local startup fails if Kafka missing | For local docs, set false when Kafka not needed |
| Kafka config has no TLS/SASL vars | Production plaintext risk | Add secure Kafka config before production |
| No checked-in Docker Compose | Infra setup manual | Add local compose for MySQL/Kafka |
| No Dockerfile | Container build not standardized | Add after `cmd/server/main.go` exists |
| No service entrypoint | Cannot run real backend process | Add `cmd/server/main.go` bootstrap |
| No health check endpoint visible | Harder operations monitoring | Add gRPC health service or HTTP health endpoint |
| Secrets are env-based | Fine for local, risky if committed | Never commit `.env`; use Secret Manager/K8s Secrets |
| MySQL root examples in docs | Bad for production | Use dedicated least-privilege DB user |
| Migrations are raw SQL only | Version tracking manual | Use a migration tool or migration table |
| `payment_failed` exists in later code/migration but `${TASK_FILE_NAME}` did not include it | Lifecycle docs may drift | Keep task docs and implementation decisions synced |

## 11. Best Practices

- Never commit `.env`.
- Commit `.env.example` with fake values only.
- Use strong 32+ character secrets for gRPC caller token and page token signing key.
- Use a dedicated MySQL user, not root, for app runtime.
- Keep migrations ordered and apply them consistently.
- Use Docker volumes for MySQL/Kafka data.
- Use `ORDER_EVENTS_ENABLED=false` for local work when Kafka is not needed.
- Use `ORDER_EVENTS_ENABLED=true` in integration/staging to test event flow.
- Run `go test ./...` before pushing.
- Run `go mod tidy` only when dependency changes are intentional.
- Keep Go version aligned with `go.mod`.
- Add health checks before production deployment.
- Add structured logging and trace IDs for checkout/payment/event flows.
- Keep lifecycle docs synced with actual domain constants and migrations.
- Use separate dev/staging/prod config values.
- Rotate secrets and do not paste production secrets into docs.

## 12. Final Checklist

- [ ] Go `1.26.3` installed
- [ ] Repository cloned
- [ ] `${TASK_PATH}` reviewed
- [ ] `${OUTPUT_PATH}` created
- [ ] Dependencies downloaded with `go mod download`
- [ ] MySQL installed or Docker container running
- [ ] `order_db` created
- [ ] Migrations `001` to `004` applied in order
- [ ] Kafka running if `ORDER_EVENTS_ENABLED=true`
- [ ] Kafka topic `order.events` created/verified
- [ ] `.env` created locally or env vars exported
- [ ] `ORDER_MYSQL_DSN` configured
- [ ] `ORDER_PAYMENT_RETURN_URL` uses HTTPS
- [ ] `ORDER_GRPC_TRUSTED_CALLER_TOKEN` is 32+ chars
- [ ] `ORDER_PAGE_TOKEN_SIGNING_KEY` is 32+ chars
- [ ] `go test ./...` passes
- [ ] `go build ./...` passes
- [ ] Missing `cmd/server/main.go` understood before trying to start service
- [ ] Docker daemon checked if using containers
- [ ] Ports `3306`, `9092`, and `9094` checked for conflicts
- [ ] Common errors section reviewed
- [ ] Security/config audit items tracked for future hardening

## Quick Beginner Summary

For `${TASK_FILE_NAME}`, no database or backend process is required because it is a lifecycle documentation task.

For current backend code under `${SERVICE_CODE_PATH}`, minimum practical setup is:

```bash
cd backend/services/order-service
go mod download
go test ./...
```

For real runtime after server bootstrap is added, you need:

```text
Go + MySQL + migrations + env vars + Kafka if events are enabled + gRPC port
```

Simple Hinglish rule: pehle dependencies install karo, phir database chalao, phir migrations apply karo, phir env vars set karo, phir tests/build run karo. Service start tabhi hoga jab server entrypoint add ho chuka ho.
